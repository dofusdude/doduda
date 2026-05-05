package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/kvarenzn/ssm/uni"
	"github.com/pierrec/lz4/v4"
	"github.com/ulikunitz/xz/lzma"
)

const unityAlignBytesFlag = 0x4000

func unpackUnityBundleNative(inputPath string, outputPath string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}

	assetsManager := uni.NewAssetsManager()
	if err := loadUnityAssetFilesNative(data, inputPath, assetsManager); err != nil {
		return err
	}

	type monoSource struct {
		file *uni.SerializedFile
		info *uni.ObjectInfo
	}

	var monos []monoSource
	for _, assetFile := range assetsManager.AssetFiles {
		for _, objectInfo := range assetFile.ObjectInfos {
			if objectInfo.ClassID == uni.ClassIDMonoBehaviour {
				monos = append(monos, monoSource{file: assetFile, info: objectInfo})
			}
		}
	}

	if len(monos) != 1 {
		return fmt.Errorf("expected exactly 1 MonoBehaviour, found %d", len(monos))
	}

	mono := monos[0]
	if mono.info.SerializedType == nil || mono.info.SerializedType.Type == nil || len(mono.info.SerializedType.Type.Nodes) == 0 {
		return fmt.Errorf("bundle has no type tree for MonoBehaviour")
	}

	reader := uni.NewObjectReader(mono.file.Reader.BinaryReader, mono.file, mono.info)
	if err := reader.SeekTo(mono.info.ByteStart); err != nil {
		return err
	}

	decoded, _, err := decodeUnityTypeTree(newUnityDecodeState(mono.file), reader.BinaryReader, mono.info.SerializedType.Type.Nodes, 0)
	if err != nil {
		return err
	}

	encoded, err := json.Marshal(decoded)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		return err
	}

	return os.WriteFile(outputPath, encoded, os.ModePerm)
}

type unityBundleStorageBlock struct {
	compressedSize   uint32
	uncompressedSize uint32
	flags            uint16
}

type unityBundleNode struct {
	offset int64
	size   int64
	path   string
}

type unityManagedType struct {
	Class string
	NS    string
	Asm   string
}

type unityJSONFloat float64

func (f unityJSONFloat) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatFloat(float64(f), 'f', 1, 64)), nil
}

type unityDecodeState struct {
	refTypeByKey        map[string]*uni.SerializedType
	managedPayloadDepth int
	registryRefDepth    int
}

func newUnityDecodeState(serializedFile *uni.SerializedFile) *unityDecodeState {
	state := &unityDecodeState{
		refTypeByKey: make(map[string]*uni.SerializedType),
	}

	if serializedFile == nil {
		return state
	}

	for _, refType := range serializedFile.RefTypes {
		state.refTypeByKey[unityManagedTypeKey(refType.ClassName, refType.Namespace, refType.AsmName)] = refType
	}

	return state
}

func loadUnityAssetFilesNative(data []byte, inputPath string, assetsManager *uni.AssetsManager) error {
	reader := uni.NewBinaryReaderFromBytes(data, true)
	signature := reader.CString()
	if signature != "UnityFS" {
		return fmt.Errorf("unsupported bundle signature %q", signature)
	}

	version := int(reader.U32())
	unityVersion := reader.CString()
	_ = reader.CString() // unity revision

	totalSize := int64(reader.S64())
	if totalSize != int64(len(data)) {
		return fmt.Errorf("bundle size mismatch: header=%d actual=%d", totalSize, len(data))
	}

	compressedBlocksInfoSize := int(reader.U32())
	uncompressedBlocksInfoSize := int(reader.U32())
	flags := int(reader.U32())

	if version >= 7 {
		reader.Align(16)
	}

	var blockInfoCompressed []byte
	if flags&0x80 != 0 {
		position := reader.Position()
		blockInfoOffset := int64(len(data) - compressedBlocksInfoSize)
		if blockInfoOffset < 0 {
			return fmt.Errorf("invalid blocks info offset")
		}
		if err := reader.SeekTo(blockInfoOffset); err != nil {
			return err
		}
		blockInfoCompressed = reader.Bytes(compressedBlocksInfoSize)
		if err := reader.SeekTo(position); err != nil {
			return err
		}
	} else {
		blockInfoCompressed = reader.Bytes(compressedBlocksInfoSize)
	}

	blockInfo, err := decompressUnityData(blockInfoCompressed, flags&0x3f, uncompressedBlocksInfoSize)
	if err != nil {
		return fmt.Errorf("decompress blocks info: %w", err)
	}

	blockReader := uni.NewBinaryReaderFromBytes(blockInfo, true)
	blockReader.Skip(16) // uncompressed data hash

	blocksCount := int(blockReader.S32())
	blocks := make([]unityBundleStorageBlock, 0, blocksCount)
	for range blocksCount {
		blocks = append(blocks, unityBundleStorageBlock{
			uncompressedSize: blockReader.U32(),
			compressedSize:   blockReader.U32(),
			flags:            blockReader.U16(),
		})
	}

	nodesCount := int(blockReader.S32())
	nodes := make([]unityBundleNode, 0, nodesCount)
	for range nodesCount {
		offset := blockReader.S64()
		size := blockReader.S64()
		_ = blockReader.U32() // node flags, currently unused
		path := blockReader.CString()
		nodes = append(nodes, unityBundleNode{
			offset: offset,
			size:   size,
			path:   path,
		})
	}

	if flags&0x200 != 0 {
		reader.Align(16)
	}

	blockStream := bytes.NewBuffer(nil)
	for _, block := range blocks {
		compressed := reader.Bytes(int(block.compressedSize))
		uncompressed, err := decompressUnityData(compressed, int(block.flags&0x3f), int(block.uncompressedSize))
		if err != nil {
			return fmt.Errorf("decompress data block: %w", err)
		}
		blockStream.Write(uncompressed)
	}

	uncompressedBundleData := blockStream.Bytes()
	for _, node := range nodes {
		start := int(node.offset)
		end := start + int(node.size)
		if start < 0 || end < start || end > len(uncompressedBundleData) {
			return fmt.Errorf("invalid bundle node bounds for %s", node.path)
		}

		streamData := uncompressedBundleData[start:end]
		normalizedStream := normalizeUnitySerializedHeader(streamData)
		subReader, err := uni.NewFileReader(streamData, node.path)
		if err != nil {
			return fmt.Errorf("create sub reader for %s: %w", node.path, err)
		}

		if subReader.FileType != uni.FileTypeAssetsFile && !bytes.Equal(normalizedStream, streamData) {
			normalizedReader, err := uni.NewFileReader(normalizedStream, node.path)
			if err == nil && normalizedReader.FileType == uni.FileTypeAssetsFile {
				subReader = &uni.FileReader{
					BinaryReader: uni.NewBinaryReaderFromBytes(normalizedStream, true),
					Path:         node.path,
					FileType:     uni.FileTypeAssetsFile,
				}
			}
		}

		if subReader.FileType != uni.FileTypeAssetsFile {
			assetsManager.ResourceFileReaders[filepath.Base(node.path)] = subReader.BinaryReader
			continue
		}

		if !bytes.Equal(normalizedStream, streamData) {
			subReader = &uni.FileReader{
				BinaryReader: uni.NewBinaryReaderFromBytes(normalizedStream, true),
				Path:         node.path,
				FileType:     uni.FileTypeAssetsFile,
			}
		}
		if err := subReader.SeekTo(0); err != nil {
			return err
		}

		if err := assetsManager.LoadAssets(subReader, inputPath, unityVersion, nil); err != nil {
			return fmt.Errorf("load assets from %s: %w", node.path, err)
		}
	}

	return nil
}

func decompressUnityData(data []byte, compression int, expectedSize int) ([]byte, error) {
	switch compression {
	case 0:
		return data, nil
	case 1:
		decoder, err := lzma.NewReader(bytes.NewReader(prepareUnityLZMAStream(data, expectedSize)))
		if err != nil {
			return nil, err
		}
		out, err := io.ReadAll(decoder)
		if err != nil {
			return nil, err
		}
		if expectedSize > 0 && len(out) != expectedSize {
			return nil, fmt.Errorf("lzma size mismatch: expected %d got %d", expectedSize, len(out))
		}
		return out, nil
	case 2, 3:
		out := make([]byte, expectedSize+0x100)
		n, err := lz4.UncompressBlock(data, out)
		if err != nil {
			return nil, err
		}
		if n != expectedSize {
			return nil, fmt.Errorf("lz4 size mismatch: expected %d got %d", expectedSize, n)
		}
		return out[:expectedSize], nil
	default:
		return nil, fmt.Errorf("unsupported compression type %d", compression)
	}
}

func prepareUnityLZMAStream(compressed []byte, expectedSize int) []byte {
	// Unity LZMA blocks use the first 5 bytes for properties and then the raw
	// payload. Build a classic .lzma header by injecting the expected size.
	if len(compressed) < 5 {
		return compressed
	}

	header := make([]byte, 13)
	copy(header[:5], compressed[:5])
	binary.LittleEndian.PutUint64(header[5:], uint64(expectedSize))

	out := make([]byte, 13+len(compressed)-5)
	copy(out, header)
	copy(out[13:], compressed[5:])
	return out
}

func normalizeUnitySerializedHeader(stream []byte) []byte {
	if len(stream) < 44 {
		return stream
	}

	metadataV1 := binary.BigEndian.Uint32(stream[0:4])
	fileSizeV1 := binary.BigEndian.Uint32(stream[4:8])
	version := binary.BigEndian.Uint32(stream[8:12])
	dataOffsetV1 := binary.BigEndian.Uint32(stream[12:16])
	if version < 22 {
		return stream
	}

	// This variant stores extended v22+ header fields directly at offset 16
	// instead of after big-endian/reserved bytes. Normalize it to the layout
	// expected by the parser.
	if metadataV1 != 0 || fileSizeV1 != 0 || dataOffsetV1 != 0 {
		return stream
	}
	if stream[17] == 0 && stream[18] == 0 && stream[19] == 0 {
		return stream
	}

	metadataSize := binary.BigEndian.Uint32(stream[16:20])
	fileSize := binary.BigEndian.Uint64(stream[20:28])
	dataOffset := binary.BigEndian.Uint64(stream[28:36])
	unknown := stream[36:44]
	if fileSize == 0 || dataOffset == 0 {
		return stream
	}

	normalized := make([]byte, len(stream)+4)
	copy(normalized[0:16], stream[0:16])
	normalized[16] = 0 // big-endian flag (false)
	// normalized[17:20] reserved bytes remain zero
	binary.BigEndian.PutUint32(normalized[20:24], metadataSize)
	binary.BigEndian.PutUint64(normalized[24:32], fileSize)
	binary.BigEndian.PutUint64(normalized[32:40], dataOffset+4)
	copy(normalized[40:48], unknown)
	copy(normalized[48:], stream[44:])

	return normalized
}

func decodeUnityReferencedObjectData(state *unityDecodeState, reader *uni.BinaryReader, node *uni.TypeTreeNode, managedType *unityManagedType) (any, error) {
	if state == nil {
		return nil, fmt.Errorf("decode state is nil")
	}
	if managedType == nil {
		return nil, fmt.Errorf("managed type is nil")
	}

	candidates := state.matchingManagedReferenceTypes(managedType)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("missing managed reference type tree for %s.%s (%s)", managedType.NS, managedType.Class, managedType.Asm)
	}

	startPos := reader.Position()
	var (
		value any
		err   error
	)
	for _, refType := range candidates {
		if err = reader.SeekTo(startPos); err != nil {
			return nil, err
		}

		state.managedPayloadDepth++
		value, _, err = decodeUnityTypeTree(state, reader, refType.Type.Nodes, 0)
		state.managedPayloadDepth--
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}

	unityAlignIfNeeded(reader, node)
	return value, nil
}

func unityExtractManagedType(value any) *unityManagedType {
	asMap, ok := value.(map[string]any)
	if !ok {
		return nil
	}

	className, classOK := asMap["class"].(string)
	nsName, nsOK := asMap["ns"].(string)
	asmName, asmOK := asMap["asm"].(string)
	if !classOK || !nsOK || !asmOK {
		return nil
	}

	return &unityManagedType{
		Class: className,
		NS:    nsName,
		Asm:   asmName,
	}
}

func unityManagedTypeKey(className string, namespace string, assembly string) string {
	return className + "|" + namespace + "|" + assembly
}

func (s *unityDecodeState) matchingManagedReferenceTypes(managedType *unityManagedType) []*uni.SerializedType {
	if s == nil || managedType == nil {
		return nil
	}

	if refType, ok := s.refTypeByKey[unityManagedTypeKey(managedType.Class, managedType.NS, managedType.Asm)]; ok && refType != nil {
		return []*uni.SerializedType{refType}
	}

	var byNamespaceAsm []*uni.SerializedType
	var byNamespace []*uni.SerializedType
	var byAsm []*uni.SerializedType
	for _, refType := range s.refTypeByKey {
		if refType == nil || refType.Type == nil || len(refType.Type.Nodes) == 0 {
			continue
		}
		if refType.Namespace == managedType.NS && refType.AsmName == managedType.Asm {
			byNamespaceAsm = append(byNamespaceAsm, refType)
			continue
		}
		if refType.Namespace == managedType.NS {
			byNamespace = append(byNamespace, refType)
		}
		if refType.AsmName == managedType.Asm {
			byAsm = append(byAsm, refType)
		}
	}
	if len(byNamespaceAsm) > 0 {
		return byNamespaceAsm
	}
	if len(byNamespace) > 0 {
		return byNamespace
	}
	if len(byAsm) > 0 {
		return byAsm
	}

	return nil
}

func decodeUnityTypeTree(state *unityDecodeState, reader *uni.BinaryReader, nodes []*uni.TypeTreeNode, idx int) (any, int, error) {
	if idx < 0 || idx >= len(nodes) {
		return nil, idx, fmt.Errorf("type tree node index out of range")
	}

	node := nodes[idx]
	nextIdx := unitySkipSubtree(nodes, idx)
	typeLower := strings.ToLower(node.Type)

	if typeLower == strings.ToLower("ReferencedManagedType") {
		start := reader.Position()
		var lastErr error
		var fallbackManagedType map[string]any
		for _, variant := range []struct {
			align        bool
			skipInt32s   int
			classCString bool
		}{
			{align: true, skipInt32s: 0, classCString: false},
			{align: false, skipInt32s: 0, classCString: false},
			{align: true, skipInt32s: 1, classCString: false},
			{align: false, skipInt32s: 1, classCString: false},
			{align: true, skipInt32s: 0, classCString: true},
			{align: true, skipInt32s: 1, classCString: true},
		} {
			if err := reader.SeekTo(start); err != nil {
				return nil, 0, err
			}
			managedType, err := decodeUnityReferencedManagedType(reader, node, variant.align, variant.skipInt32s, variant.classCString)
			if err == nil && unityManagedTypeLooksValid(managedType) {
				if state != nil {
					if matched := state.matchingManagedReferenceTypes(unityExtractManagedType(managedType)); len(matched) > 0 {
						return managedType, nextIdx, nil
					}
				}
				if fallbackManagedType == nil {
					fallbackManagedType = managedType
				}
			}
			if err != nil {
				lastErr = err
			}
		}
		if fallbackManagedType != nil {
			return fallbackManagedType, nextIdx, nil
		}
		if lastErr != nil {
			return nil, 0, fmt.Errorf("ReferencedManagedType decode failed at offset %d (%s) bytes=%s: %w", start, unityDirectChildSummary(nodes, idx), unityHexPreview(reader, start, 24), lastErr)
		}
		return nil, 0, fmt.Errorf("invalid ReferencedManagedType at offset %d (%s) bytes=%s", start, unityDirectChildSummary(nodes, idx), unityHexPreview(reader, start, 24))
	}

	if strings.EqualFold(node.Type, "ReferencedObject") {
		return decodeUnityReferencedObjectNode(state, reader, nodes, idx)
	}

	switch typeLower {
	case "string":
		value, err := readUnityStringMode(reader, true)
		if err != nil {
			return nil, 0, err
		}
		return value, nextIdx, nil
	case "typelessdata":
		if !unityCanRead(reader, 4) {
			return nil, 0, fmt.Errorf("typelessdata missing length bytes")
		}
		length := int(reader.S32())
		if length < 0 {
			return nil, 0, fmt.Errorf("negative typeless data length")
		}
		if !unityCanRead(reader, int64(length)) {
			return nil, 0, fmt.Errorf("typelessdata length %d exceeds remaining bytes", length)
		}
		raw := reader.Bytes(length)
		out := make([]int, len(raw))
		for i, b := range raw {
			out[i] = int(b)
		}
		reader.Align(4)
		return out, nextIdx, nil
	}

	if typeLower == "array" {
		return decodeUnityArray(state, reader, nodes, idx)
	}

	hasChildren := idx+1 < len(nodes) && nodes[idx+1].Level > node.Level
	if !hasChildren {
		value, err := readUnityPrimitive(reader, node.Type)
		if err != nil {
			return nil, 0, fmt.Errorf("node %q type %q at index %d: %w", node.Name, node.Type, idx, err)
		}
		unityAlignIfNeeded(reader, node)
		return value, nextIdx, nil
	}

	objectValue := make(map[string]any)
	var referencedType *unityManagedType
	for childIdx := idx + 1; childIdx < len(nodes) && nodes[childIdx].Level > node.Level; {
		if nodes[childIdx].Level != node.Level+1 {
			childIdx++
			continue
		}

		childNode := nodes[childIdx]
		var (
			childValue any
			newIdx     int
			err        error
		)

		if strings.EqualFold(childNode.Name, "data") && strings.EqualFold(childNode.Type, "ReferencedObjectData") {
			if referencedType == nil {
				return nil, 0, fmt.Errorf("referenced object data without managed type context")
			}
			childValue, err = decodeUnityReferencedObjectData(state, reader, childNode, referencedType)
			if err != nil {
				return nil, 0, err
			}
			newIdx = unitySkipSubtree(nodes, childIdx)
		} else if state != nil &&
			state.managedPayloadDepth > 0 &&
			strings.EqualFold(childNode.Name, "references") &&
			strings.EqualFold(childNode.Type, "ManagedReferencesRegistry") {
			childIdx = unitySkipSubtree(nodes, childIdx)
			continue
		} else {
			if state != nil && strings.EqualFold(node.Type, "ManagedReferencesRegistry") && strings.EqualFold(childNode.Name, "RefIds") {
				state.registryRefDepth++
				childValue, newIdx, err = decodeUnityTypeTree(state, reader, nodes, childIdx)
				state.registryRefDepth--
			} else {
				childValue, newIdx, err = decodeUnityTypeTree(state, reader, nodes, childIdx)
			}
		}
		if err != nil {
			return nil, 0, fmt.Errorf("node %q(%s){%s} child %q(%s): %w", node.Name, node.Type, unityDirectChildSummary(nodes, idx), childNode.Name, childNode.Type, err)
		}

		key := childNode.Name
		if key == "" {
			key = fmt.Sprintf("field_%d", childIdx)
		}
		childValue = unityNormalizeFieldValue(node.Type, key, childValue)
		if strings.EqualFold(node.Type, "ManagedReferencesRegistry") && key == "RefIds" {
			if asMap, ok := childValue.(map[string]any); ok {
				if arr, ok := asMap["Array"]; ok {
					childValue = arr
				}
			}
		}
		objectValue[key] = childValue
		if strings.EqualFold(key, "type") {
			referencedType = unityExtractManagedType(childValue)
		}
		childIdx = newIdx
	}

	unityAlignIfNeeded(reader, node)
	return objectValue, nextIdx, nil
}

func decodeUnityArray(state *unityDecodeState, reader *uni.BinaryReader, nodes []*uni.TypeTreeNode, idx int) (any, int, error) {
	node := nodes[idx]
	nextIdx := unitySkipSubtree(nodes, idx)
	if idx+1 >= len(nodes) || nodes[idx+1].Level != node.Level+1 {
		return []any{}, nextIdx, nil
	}

	sizeIdx := idx + 1
	sizeValue, sizeNextIdx, err := decodeUnityTypeTree(state, reader, nodes, sizeIdx)
	if err != nil {
		return nil, 0, err
	}

	count, err := unityToInt(sizeValue)
	if err != nil {
		return nil, 0, fmt.Errorf("array size decode failed: %w", err)
	}
	if count < 0 {
		return nil, 0, fmt.Errorf("negative array size %d", count)
	}
	if count > 10_000_000 {
		return nil, 0, fmt.Errorf("array size %d exceeds safety limit", count)
	}
	remaining := reader.Len() - reader.Position()
	if int64(count) > remaining {
		return nil, 0, fmt.Errorf("array size %d exceeds remaining bytes %d", count, remaining)
	}

	if sizeNextIdx >= nextIdx {
		return []any{}, nextIdx, nil
	}

	dataIdx := sizeNextIdx
	dataNode := nodes[dataIdx]
	dataHasChildren := dataIdx+1 < len(nodes) && nodes[dataIdx+1].Level > dataNode.Level
	if !dataHasChildren && unityIsByteType(dataNode.Type) {
		if !unityCanRead(reader, int64(count)) {
			return nil, 0, fmt.Errorf("array byte payload %d exceeds remaining bytes", count)
		}
		raw := reader.Bytes(count)
		out := make([]int, len(raw))
		for i, b := range raw {
			out[i] = int(b)
		}
		unityAlignIfNeeded(reader, dataNode)
		unityAlignIfNeeded(reader, node)
		return out, nextIdx, nil
	}

	out := make([]any, count)
	for i := range count {
		value, _, err := decodeUnityTypeTree(state, reader, nodes, dataIdx)
		if err != nil {
			return nil, 0, fmt.Errorf("array %q element %d (%q/%s): %w", node.Name, i, dataNode.Name, dataNode.Type, err)
		}
		out[i] = value
	}

	unityAlignIfNeeded(reader, node)
	return out, nextIdx, nil
}

func decodeUnityReferencedObjectNode(state *unityDecodeState, reader *uni.BinaryReader, nodes []*uni.TypeTreeNode, idx int) (any, int, error) {
	node := nodes[idx]
	nextIdx := unitySkipSubtree(nodes, idx)

	if state == nil || state.registryRefDepth == 0 {
		if !unityCanRead(reader, 8) {
			return nil, 0, fmt.Errorf("insufficient bytes for referenced rid")
		}
		out := map[string]any{
			"rid": reader.S64(),
		}
		unityAlignIfNeeded(reader, node)
		return out, nextIdx, nil
	}

	var ridIdx, typeIdx, dataIdx = -1, -1, -1
	for childIdx := idx + 1; childIdx < len(nodes) && nodes[childIdx].Level > node.Level; {
		if nodes[childIdx].Level == node.Level+1 {
			switch {
			case strings.EqualFold(nodes[childIdx].Name, "rid"):
				ridIdx = childIdx
			case strings.EqualFold(nodes[childIdx].Name, "type"):
				typeIdx = childIdx
			case strings.EqualFold(nodes[childIdx].Name, "data"):
				dataIdx = childIdx
			}
		}
		childIdx = unitySkipSubtree(nodes, childIdx)
	}

	if ridIdx == -1 || typeIdx == -1 || dataIdx == -1 {
		return nil, 0, fmt.Errorf("ReferencedObject missing expected fields (%s)", unityDirectChildSummary(nodes, idx))
	}

	start := reader.Position()
	var lastErr error
	for _, ridSize := range []int{8} {
		if err := reader.SeekTo(start); err != nil {
			return nil, 0, err
		}

		ridNode := nodes[ridIdx]
		var ridValue any
		switch ridSize {
		case 8:
			if !unityCanRead(reader, 8) {
				lastErr = fmt.Errorf("insufficient bytes for rid:int64")
				continue
			}
			ridValue = reader.S64()
		case 4:
			if !unityCanRead(reader, 4) {
				lastErr = fmt.Errorf("insufficient bytes for rid:int32")
				continue
			}
			ridValue = int64(reader.S32())
		}
		unityAlignIfNeeded(reader, ridNode)

		typeValue, _, err := decodeUnityTypeTree(state, reader, nodes, typeIdx)
		if err != nil {
			lastErr = fmt.Errorf("type decode (rid=%d bytes): %w", ridSize, err)
			continue
		}
		managedType := unityExtractManagedType(typeValue)
		if managedType == nil {
			lastErr = fmt.Errorf("type decode (rid=%d bytes): missing managed type context", ridSize)
			continue
		}

		dataValue, err := decodeUnityReferencedObjectData(state, reader, nodes[dataIdx], managedType)
		if err != nil {
			lastErr = fmt.Errorf("data decode (rid=%d bytes): %w", ridSize, err)
			continue
		}

		out := map[string]any{
			"rid":  ridValue,
			"type": typeValue,
			"data": dataValue,
		}
		unityAlignIfNeeded(reader, node)
		return out, nextIdx, nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("failed to decode ReferencedObject")
	}
	return nil, 0, lastErr
}

func decodeUnityReferencedManagedType(reader *uni.BinaryReader, node *uni.TypeTreeNode, alignStrings bool, skipInt32s int, classCString bool) (map[string]any, error) {
	for i := 0; i < skipInt32s; i++ {
		if !unityCanRead(reader, 4) {
			return nil, fmt.Errorf("insufficient bytes for ReferencedManagedType prefix")
		}
		_ = reader.S32()
	}

	var (
		class string
		err   error
	)
	if classCString {
		class, err = readUnityCStringMode(reader, alignStrings)
	} else {
		class, err = readUnityStringMode(reader, alignStrings)
	}
	if err != nil {
		return nil, err
	}
	ns, err := readUnityStringMode(reader, alignStrings)
	if err != nil {
		return nil, err
	}
	asm, err := readUnityStringMode(reader, alignStrings)
	if err != nil {
		return nil, err
	}
	unityAlignIfNeeded(reader, node)
	return map[string]any{
		"class": class,
		"ns":    ns,
		"asm":   asm,
	}, nil
}

func readUnityStringMode(reader *uni.BinaryReader, align bool) (string, error) {
	if !unityCanRead(reader, 4) {
		return "", fmt.Errorf("insufficient bytes for string length")
	}
	length := int64(reader.S32())
	if length < 0 {
		return "", fmt.Errorf("negative string length")
	}
	if !unityCanRead(reader, length) {
		return "", fmt.Errorf("string length %d exceeds remaining bytes", length)
	}
	value := string(reader.Bytes(int(length)))
	if align {
		reader.Align(4)
	}
	return value, nil
}

func readUnityCStringMode(reader *uni.BinaryReader, align bool) (string, error) {
	const maxLen = 1024
	var out []byte
	for len(out) < maxLen {
		if !unityCanRead(reader, 1) {
			return "", fmt.Errorf("insufficient bytes for cstring")
		}
		b := reader.U8()
		if b == 0 {
			if align {
				reader.Align(4)
			}
			return string(out), nil
		}
		out = append(out, b)
	}
	return "", fmt.Errorf("cstring exceeds maximum length")
}

func unityManagedTypeLooksValid(entry map[string]any) bool {
	class, _ := entry["class"].(string)
	ns, _ := entry["ns"].(string)
	asm, _ := entry["asm"].(string)
	if class == "" || ns == "" || asm == "" {
		return false
	}
	return unityStringLooksSane(class) && unityStringLooksSane(ns) && unityStringLooksSane(asm)
}

func unityStringLooksSane(value string) bool {
	if value == "" || len(value) > 512 || !utf8.ValidString(value) {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return false
		}
	}
	return true
}

func readUnityPrimitive(reader *uni.BinaryReader, typ string) (any, error) {
	switch strings.ToLower(typ) {
	case "bool":
		if !unityCanRead(reader, 1) {
			return nil, fmt.Errorf("insufficient bytes for bool")
		}
		return reader.Bool(), nil
	case "sint8", "int8":
		if !unityCanRead(reader, 1) {
			return nil, fmt.Errorf("insufficient bytes for int8")
		}
		return int(reader.S8()), nil
	case "uint8", "unsigned char", "char":
		if !unityCanRead(reader, 1) {
			return nil, fmt.Errorf("insufficient bytes for uint8")
		}
		return int(reader.U8()), nil
	case "sint16", "int16", "short":
		if !unityCanRead(reader, 2) {
			return nil, fmt.Errorf("insufficient bytes for int16")
		}
		return int(reader.S16()), nil
	case "uint16", "unsigned short":
		if !unityCanRead(reader, 2) {
			return nil, fmt.Errorf("insufficient bytes for uint16")
		}
		return int(reader.U16()), nil
	case "sint32", "int32", "int":
		if !unityCanRead(reader, 4) {
			return nil, fmt.Errorf("insufficient bytes for int32")
		}
		return int(reader.S32()), nil
	case "uint32", "unsigned int":
		if !unityCanRead(reader, 4) {
			return nil, fmt.Errorf("insufficient bytes for uint32")
		}
		return uint32(reader.U32()), nil
	case "sint64", "int64", "long", "long long":
		if !unityCanRead(reader, 8) {
			return nil, fmt.Errorf("insufficient bytes for int64")
		}
		return reader.S64(), nil
	case "uint64", "unsigned long long":
		if !unityCanRead(reader, 8) {
			return nil, fmt.Errorf("insufficient bytes for uint64")
		}
		return reader.U64(), nil
	case "float", "single":
		if !unityCanRead(reader, 4) {
			return nil, fmt.Errorf("insufficient bytes for float")
		}
		return reader.F32(), nil
	case "double":
		if !unityCanRead(reader, 8) {
			return nil, fmt.Errorf("insufficient bytes for double")
		}
		return reader.F64(), nil
	default:
		return nil, fmt.Errorf("unsupported primitive type %q", typ)
	}
}

func unityCanRead(reader *uni.BinaryReader, byteCount int64) bool {
	if reader == nil {
		return false
	}
	if byteCount < 0 {
		return false
	}
	return reader.Position()+byteCount <= reader.Len()
}

func unitySkipSubtree(nodes []*uni.TypeTreeNode, idx int) int {
	base := nodes[idx].Level
	next := idx + 1
	for next < len(nodes) && nodes[next].Level > base {
		next++
	}
	return next
}

func unityAlignIfNeeded(reader *uni.BinaryReader, node *uni.TypeTreeNode) {
	if node.MetaFlag.IsSome() && (node.MetaFlag.Unwrap()&unityAlignBytesFlag) != 0 {
		reader.Align(4)
	}
}

func unityToInt(value any) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int8:
		return int(v), nil
	case int16:
		return int(v), nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case uint8:
		return int(v), nil
	case uint16:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("unsupported integer type %T", value)
	}
}

func unityIsByteType(typ string) bool {
	switch strings.ToLower(typ) {
	case "uint8", "unsigned char", "char", "sint8", "int8":
		return true
	default:
		return false
	}
}

func unityNormalizeFieldValue(ownerType string, fieldName string, value any) any {
	if strings.EqualFold(ownerType, "CharacterXpMappingData") && strings.EqualFold(fieldName, "experiencePoints") {
		switch v := value.(type) {
		case int:
			return unityJSONFloat(float64(v))
		case int8:
			return unityJSONFloat(float64(v))
		case int16:
			return unityJSONFloat(float64(v))
		case int32:
			return unityJSONFloat(float64(v))
		case int64:
			return unityJSONFloat(float64(v))
		case uint8:
			return unityJSONFloat(float64(v))
		case uint16:
			return unityJSONFloat(float64(v))
		case uint32:
			return unityJSONFloat(float64(v))
		case uint64:
			return unityJSONFloat(float64(v))
		case float32:
			return unityJSONFloat(float64(v))
		case float64:
			return unityJSONFloat(v)
		}
	}
	return value
}

func unityDirectChildSummary(nodes []*uni.TypeTreeNode, idx int) string {
	if idx < 0 || idx >= len(nodes) {
		return ""
	}
	base := nodes[idx].Level
	var parts []string
	for i := idx + 1; i < len(nodes) && nodes[i].Level > base; i++ {
		if nodes[i].Level != base+1 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s:%s", nodes[i].Name, nodes[i].Type))
	}
	return strings.Join(parts, ",")
}

func unityHexPreview(reader *uni.BinaryReader, offset int64, length int) string {
	if reader == nil || length <= 0 {
		return ""
	}
	current := reader.Position()
	defer func() {
		_ = reader.SeekTo(current)
	}()
	if err := reader.SeekTo(offset); err != nil {
		return ""
	}
	remaining := reader.Len() - reader.Position()
	if remaining <= 0 {
		return ""
	}
	if int64(length) > remaining {
		length = int(remaining)
	}
	return hex.EncodeToString(reader.Bytes(length))
}
