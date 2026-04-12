package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/kvarenzn/ssm/uni"
	"github.com/xypwn/filediver/dds"
)

func unpackUnityImagesNative(inputDir string, outputDir string) error {
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".imagebundle") {
			continue
		}

		inputPath := filepath.Join(inputDir, entry.Name())
		if err := unpackUnityImageBundleNative(inputPath, outputDir); err != nil {
			return fmt.Errorf("unpack %s: %w", entry.Name(), err)
		}
	}

	return nil
}

func unpackUnityImageBundleNative(inputPath string, outputDir string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}

	assetsManager := uni.NewAssetsManager()
	if err := assetsManager.LoadDataFromHandler(data, filepath.Base(inputPath)); err != nil {
		return err
	}

	for _, assetFile := range assetsManager.AssetFiles {
		for _, objectInfo := range assetFile.ObjectInfos {
			reader := uni.NewObjectReader(assetFile.Reader.BinaryReader, assetFile, objectInfo)
			switch objectInfo.ClassID {
			case uni.ClassIDAssetBundle:
				assetFile.AddObject(uni.NewAssetBundle(reader))
			case uni.ClassIDTexture2D:
				assetFile.AddObject(uni.NewTexture2D(reader))
			case uni.ClassIDSprite:
				assetFile.AddObject(uni.NewSprite(reader))
			default:
				assetFile.AddObject(uni.NewObject(reader))
			}
		}
	}

	targetDir := outputDir
	if subdir := unityImageResolutionSubdir(filepath.Base(inputPath)); subdir != "" {
		targetDir = filepath.Join(outputDir, subdir)
	}
	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return err
	}
	layerOrder := unityBundleLayerDirs(filepath.Base(inputPath))

	textureCache := make(map[*uni.Texture2D]image.Image)
	nameStates := make(map[string]*unityNameState)

	for _, assetFile := range assetsManager.AssetFiles {
		for _, object := range assetFile.Objects {
			sprite, ok := object.(*uni.Sprite)
			if !ok {
				continue
			}

			spriteImage, err := unityDecodeSpriteImage(sprite, textureCache)
			if err != nil {
				return fmt.Errorf("decode sprite %q: %w", sprite.Name, err)
			}

			var textureName string
			if sprite.RenderData != nil && sprite.RenderData.Texture != nil {
				if texture, ok := sprite.RenderData.Texture.Get().(*uni.Texture2D); ok && texture != nil {
					textureName = texture.Name
				}
			}

			outputName := unityOutputImageName(sprite.Name, unityObjectFallbackName(sprite.GetObject()))
			outputPath := unityNextImagePath(targetDir, outputName, spriteImage, nameStates, layerOrder)
			if err := unityWritePNG(outputPath, spriteImage); err != nil {
				return err
			}

			// AssetStudio occasionally names the emitted file after the linked texture
			// instead of the sprite object. Export both aliases when they differ.
			if len(layerOrder) == 0 && textureName != "" {
				textureAliasName := unityOutputImageName(textureName, unityObjectFallbackName(sprite.GetObject()))
				if textureAliasName != outputName {
					textureAliasPath := unityNextImagePath(targetDir, textureAliasName, spriteImage, nameStates, layerOrder)
					if err := unityWritePNG(textureAliasPath, spriteImage); err != nil {
						return err
					}
				}
			}
		}
	}

	if len(layerOrder) > 0 {
		return nil
	}

	for _, assetFile := range assetsManager.AssetFiles {
		for _, object := range assetFile.Objects {
			texture, ok := object.(*uni.Texture2D)
			if !ok {
				continue
			}

			textureImage, err := unityDecodeTextureImage(texture, 0, 0)
			if err != nil {
				return fmt.Errorf("decode texture %q: %w", texture.Name, err)
			}

			outputName := unityOutputImageName(texture.Name, unityObjectFallbackName(texture.GetObject()))
			outputPath := unityNextImagePath(targetDir, outputName, textureImage, nameStates, layerOrder)
			if err := unityWritePNG(outputPath, textureImage); err != nil {
				return err
			}
		}
	}

	return nil
}

func unityDecodeSpriteImage(sprite *uni.Sprite, textureCache map[*uni.Texture2D]image.Image) (image.Image, error) {
	if sprite == nil || sprite.RenderData == nil || sprite.RenderData.Texture == nil {
		return nil, fmt.Errorf("sprite has no render data texture")
	}

	textureObject, ok := sprite.RenderData.Texture.Get().(*uni.Texture2D)
	if !ok || textureObject == nil {
		return nil, fmt.Errorf("sprite texture reference is not a Texture2D")
	}

	textureImage, ok := textureCache[textureObject]
	if !ok {
		var err error
		hintWidth, hintHeight := 0, 0
		if sprite.RenderData != nil && sprite.RenderData.TextureRect != nil {
			hintWidth = int(math.Round(float64(sprite.RenderData.TextureRect.Width)))
			hintHeight = int(math.Round(float64(sprite.RenderData.TextureRect.Height)))
		}
		textureImage, err = unityDecodeTextureImage(textureObject, hintWidth, hintHeight)
		if err != nil {
			return nil, err
		}
		textureCache[textureObject] = textureImage
	}

	return textureImage, nil
}

func unityDecodeTextureImage(texture *uni.Texture2D, hintWidth int, hintHeight int) (image.Image, error) {
	raw, err := unityReadTextureData(texture)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("texture data is empty")
	}

	meta := unityNormalizedTextureMeta(texture, len(raw), hintWidth, hintHeight)
	if meta.width <= 0 || meta.height <= 0 {
		return nil, fmt.Errorf("invalid texture dimensions %dx%d", meta.width, meta.height)
	}

	switch meta.format {
	case uni.BC7:
		decoded := image.NewNRGBA(image.Rect(0, 0, meta.width, meta.height))
		size := meta.completeSize
		if size <= 0 || size > len(raw) {
			size = len(raw)
		}
		if err := dds.DecompressBC7(decoded.Pix, bytes.NewReader(raw[:size]), meta.width, meta.height, dds.Info{ColorModel: color.NRGBAModel}); err != nil {
			return nil, err
		}
		return unityFlipVerticalNRGBA(decoded), nil
	default:
		textureCopy := *texture
		textureCopy.Width = int32(meta.width)
		textureCopy.Height = int32(meta.height)
		textureCopy.Format = meta.format
		textureCopy.ImageData = uni.NewResourceReader(uni.NewBinaryReaderFromBytes(raw, true), 0, int64(len(raw)))
		return uni.DecodeTexture2D(&textureCopy)
	}
}

type unityTextureMeta struct {
	width        int
	height       int
	completeSize int
	format       uni.TextureFormat
}

func unityNormalizedTextureMeta(texture *uni.Texture2D, rawLen int, hintWidth int, hintHeight int) unityTextureMeta {
	meta := unityTextureMeta{
		width:        int(texture.Width),
		height:       int(texture.Height),
		completeSize: int(texture.CompleteImageSize),
		format:       texture.Format,
	}

	// Dofus 3 image bundles currently expose a shifted field layout in ssm:
	// m_CompleteImageSize is read into height, and m_TextureFormat into mipsStripped.
	if meta.format == uni.Alpha8 && meta.completeSize == 0 && texture.MipsStripped.IsSome() && meta.height > meta.width {
		meta.completeSize = meta.height
		meta.format = uni.TextureFormat(texture.MipsStripped.Unwrap())
	}

	if meta.completeSize <= 0 {
		meta.completeSize = rawLen
	}
	if hintWidth > 0 && hintHeight > 0 && meta.completeSize > 0 && hintWidth*hintHeight == meta.completeSize {
		meta.width = hintWidth
		meta.height = hintHeight
		return meta
	}
	if hintWidth > 0 && meta.completeSize > 0 && meta.completeSize%hintWidth == 0 {
		derivedHeight := meta.completeSize / hintWidth
		if hintHeight <= 0 || absInt(derivedHeight-hintHeight) <= 2 {
			meta.width = hintWidth
			meta.height = derivedHeight
			return meta
		}
	}
	if hintHeight > 0 && meta.completeSize > 0 && meta.completeSize%hintHeight == 0 {
		derivedWidth := meta.completeSize / hintHeight
		if hintWidth <= 0 || absInt(derivedWidth-hintWidth) <= 2 {
			meta.width = derivedWidth
			meta.height = hintHeight
			return meta
		}
	}

	if meta.width > 0 {
		switch meta.format {
		case uni.BC7:
			if meta.completeSize%meta.width == 0 {
				meta.height = meta.completeSize / meta.width
			}
		case uni.Alpha8:
			if rawLen%meta.width == 0 {
				meta.height = rawLen / meta.width
			}
		case uni.RGB24:
			if rawLen%(meta.width*3) == 0 {
				meta.height = rawLen / (meta.width * 3)
			}
		case uni.RGBA32, uni.ARGB32:
			if rawLen%(meta.width*4) == 0 {
				meta.height = rawLen / (meta.width * 4)
			}
		}
	}

	return meta
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func unityFlipVerticalNRGBA(src *image.NRGBA) *image.NRGBA {
	if src == nil {
		return nil
	}
	bounds := src.Bounds()
	out := image.NewNRGBA(bounds)
	rowSize := bounds.Dx() * 4
	for y := 0; y < bounds.Dy(); y++ {
		srcStart := (bounds.Dy()-1-y)*src.Stride + bounds.Min.X*4
		dstStart := y*out.Stride + bounds.Min.X*4
		copy(out.Pix[dstStart:dstStart+rowSize], src.Pix[srcStart:srcStart+rowSize])
	}
	return out
}

func unityReadTextureData(texture *uni.Texture2D) ([]byte, error) {
	if texture == nil || texture.ImageData == nil {
		return nil, fmt.Errorf("texture has no image data reader")
	}

	resourceReader := texture.ImageData
	reader := resourceReader.GetReader()
	if reader == nil {
		return nil, fmt.Errorf("texture resource reader is unavailable")
	}

	size := resourceReader.Size
	if size <= 0 {
		return nil, nil
	}
	offset := resourceReader.Offset
	if offset < 0 || offset+size > reader.Len() {
		return nil, fmt.Errorf("texture data range is out of bounds (offset=%d size=%d len=%d)", offset, size, reader.Len())
	}
	if size > int64(^uint(0)>>1) {
		return nil, fmt.Errorf("texture data size %d is too large", size)
	}

	if err := reader.SeekTo(offset); err != nil {
		return nil, err
	}
	out := reader.Bytes(int(size))
	return append([]byte(nil), out...), nil
}

func unityCropSpriteImage(src image.Image, sprite *uni.Sprite) image.Image {
	if src == nil || sprite == nil || sprite.RenderData == nil || sprite.RenderData.TextureRect == nil {
		return src
	}

	rect := sprite.RenderData.TextureRect
	rx := int(math.Round(float64(rect.X)))
	ry := int(math.Round(float64(rect.Y)))
	rw := int(math.Round(float64(rect.Width)))
	rh := int(math.Round(float64(rect.Height)))
	if rw <= 0 || rh <= 0 {
		return src
	}

	bounds := src.Bounds()
	if rw == bounds.Dx() && rh == bounds.Dy() && rx == 0 && ry == 0 {
		return src
	}

	candidates := []image.Rectangle{
		image.Rect(rx, ry, rx+rw, ry+rh),
		image.Rect(rx, bounds.Dy()-ry-rh, rx+rw, bounds.Dy()-ry),
	}

	var cropRect image.Rectangle
	found := false
	for _, candidate := range candidates {
		if candidate.Min.X < 0 || candidate.Min.Y < 0 || candidate.Max.X > bounds.Dx() || candidate.Max.Y > bounds.Dy() {
			continue
		}
		cropRect = candidate
		found = true
		break
	}
	if !found {
		return src
	}

	dst := image.NewNRGBA(image.Rect(0, 0, rw, rh))
	draw.Draw(dst, dst.Bounds(), src, cropRect.Min, draw.Src)
	return dst
}

func unityImageResolutionSubdir(bundleName string) string {
	base := strings.TrimSuffix(bundleName, ".imagebundle")
	lastUnderscore := strings.LastIndex(base, "_")
	if lastUnderscore < 0 || lastUnderscore+1 >= len(base) {
		return ""
	}

	resolutionID := base[lastUnderscore+1:]
	if !isOnlyDigits(resolutionID) {
		return ""
	}

	return resolutionID + "x"
}

func unityOutputImageName(assetName string, fallback string) string {
	name := strings.TrimSpace(assetName)
	if name == "" {
		name = fallback
	}
	name = filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	name = strings.TrimSuffix(name, filepath.Ext(name))
	name = strings.TrimSpace(name)
	if name == "" {
		return "unnamed"
	}

	invalid := `<>:"/\|?*`
	name = strings.Map(func(r rune) rune {
		if strings.ContainsRune(invalid, r) {
			return '_'
		}
		return r
	}, name)

	return name
}

func unityObjectFallbackName(object *uni.Object) string {
	if object == nil {
		return "unnamed"
	}
	return fmt.Sprintf("%d", object.PathID)
}

type unityNameState struct {
	total int
	byDim map[string]int
}

func unityBundleLayerDirs(bundleName string) []string {
	if strings.Contains(bundleName, "emblem_images_") {
		// For emblem bundles, the primary sprite is "up", with optional additional layers.
		return []string{"up", "backcontent", "outlinealliance", "outlineguild"}
	}
	return nil
}

func unityNextImagePath(dir string, baseName string, img image.Image, states map[string]*unityNameState, layerOrder []string) string {
	dimKey := "unknown"
	if img != nil {
		bounds := img.Bounds()
		dimKey = fmt.Sprintf("%dx%d", bounds.Dx(), bounds.Dy())
	}

	state, ok := states[baseName]
	if !ok {
		state = &unityNameState{byDim: make(map[string]int)}
		states[baseName] = state
	}

	outDir := dir
	if len(layerOrder) > 0 {
		outDir = filepath.Join(dir, layerOrder[state.total%len(layerOrder)])
	}

	basePath := filepath.Join(outDir, baseName+".png")
	if state.total == 0 {
		state.total++
		state.byDim[dimKey]++
		return basePath
	}

	state.total++
	state.byDim[dimKey]++

	var preferred string
	if state.byDim[dimKey] > 1 {
		// Same-dimension duplicates often represent truncated numeric IDs in AssetStudio output.
		suffix := state.byDim[dimKey] - 1
		if isOnlyDigits(baseName) && baseName != "0" && suffix == 1 {
			preferred = filepath.Join(outDir, fmt.Sprintf("%s_#%02d.png", baseName, suffix))
		} else {
			preferred = filepath.Join(outDir, fmt.Sprintf("%s_#%d.png", baseName, suffix))
		}
	} else {
		// Cross-dimension duplicates should collapse back to the same base name after cleanImages.
		preferred = filepath.Join(outDir, fmt.Sprintf("%s_#%02d.png", baseName, state.total-1))
	}
	if _, err := os.Stat(preferred); os.IsNotExist(err) {
		return preferred
	}

	for i := 1; ; i++ {
		candidate := filepath.Join(outDir, fmt.Sprintf("%s_#%02d_%d.png", baseName, state.total-1, i))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}

func unityWritePNG(path string, img image.Image) error {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}
