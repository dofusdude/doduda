package main

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/dofusdude/ankabuffer"
)

func DownloadImagesLauncher(hashJson *ankabuffer.Manifest, dir string, pythonPath string) error {

	fileNames := []HashFile{
		{Filename: "content/gfx/items/bitmap0.d2p", FriendlyName: "bitmaps_0.d2p"},
		{Filename: "content/gfx/items/bitmap0_1.d2p", FriendlyName: "bitmaps_1.d2p"},
		{Filename: "content/gfx/items/bitmap1.d2p", FriendlyName: "bitmaps_2.d2p"},
		{Filename: "content/gfx/items/bitmap1_1.d2p", FriendlyName: "bitmaps_3.d2p"},
		{Filename: "content/gfx/items/bitmap1_2.d2p", FriendlyName: "bitmaps_4.d2p"},
	}

	inPath := filepath.Join(dir, "data", "tmp")
	outPath := filepath.Join(dir, "data", "img", "item")
	if err := DownloadUnpackFiles(hashJson, "main", fileNames, dir, inPath, false, pythonPath); err != nil {
		return err
	}

	absConvertCmd := fmt.Sprintf("%s/PyDofus/%s_unpack.py", dir, "d2p")
	if err := exec.Command(pythonPath, absConvertCmd, inPath, outPath).Run(); err != nil {
		return err
	}

	fileNames = []HashFile{
		{Filename: "content/gfx/items/vector0.d2p", FriendlyName: "vector_0.d2p"},
		{Filename: "content/gfx/items/vector0_1.d2p", FriendlyName: "vector_1.d2p"},
		{Filename: "content/gfx/items/vector1.d2p", FriendlyName: "vector_2.d2p"},
		{Filename: "content/gfx/items/vector1_1.d2p", FriendlyName: "vector_3.d2p"},
		{Filename: "content/gfx/items/vector1_2.d2p", FriendlyName: "vector_4.d2p"},
	}

	inPath = filepath.Join(dir, "data", "tmp", "vector")
	outPath = filepath.Join(dir, "data", "vector", "item")
	if err := DownloadUnpackFiles(hashJson, "main", fileNames, dir, inPath, false, pythonPath); err != nil {
		return err
	}

	if err := exec.Command(pythonPath, absConvertCmd, inPath, outPath).Run(); err != nil {
		return err
	}

	return nil
}
