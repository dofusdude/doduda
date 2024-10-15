package main

import (
	"path/filepath"

	"github.com/dofusdude/ankabuffer"
)

func DownloadItems(hashJson *ankabuffer.Manifest, dir string, indent string, headless bool) error {
	fileNames := []HashFile{
		//{Filename: "Dofus_Data/StreamingAssets/Content/Data/data_assets_itemtypesroot.asset.bundle", FriendlyName: "items.d2o"},
		//{Filename: "zaap.yml", FriendlyName: "zaap.yml"},
		{Filename: "Dofus_Data/StreamingAssets/Content/Picto/Items/catalog_1.0.json", FriendlyName: "items_catalog.json"},
		{Filename: "Dofus_Data/StreamingAssets/Content/Picto/Items/item_.bundle", FriendlyName: "items.bundle"},
	}

	outPath := filepath.Join(dir, "data")

	// configuration
	err := DownloadUnpackFiles("Items", hashJson, "picto", fileNames, dir, outPath, false, indent, headless, false)

	return err
}
