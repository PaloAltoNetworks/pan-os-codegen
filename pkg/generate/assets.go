package generate

import (
	"bytes"
	"fmt"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func CopyAssets(config *properties.Config) error {
	for _, asset := range config.Assets {
		files, err := listAssets(asset)
		if err != nil {
			return err
		}

		if asset.Target.GoSdk {
			if err = copyAsset(config.Output.GoSdk, asset, files); err != nil {
				return err
			}
		}
		if asset.Target.TerraformProvider {
			if err = copyAsset(config.Output.TerraformProvider, asset, files); err != nil {
				return err
			}
		}
	}

	return nil
}

func listAssets(asset *properties.Asset) ([]string, error) {
	var files []string

	// Walk through directory and get list of all files
	err := filepath.WalkDir(asset.Source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

func copyAsset(target string, asset *properties.Asset, files []string) error {
	// Prepare destination path
	destinationDir := target + "/" + asset.Destination

	// Create the destination directory if it doesn't exist
	if err := os.MkdirAll(destinationDir, os.ModePerm); err != nil {
		return err
	}

	for _, sourceFilePath := range files {
		// Prepare destination path
		destinationFilePath := filepath.Join(destinationDir, filepath.Base(sourceFilePath))
		fmt.Printf("Copy file from %s to %s\n", sourceFilePath, destinationFilePath)

		// Read the contents of the source file
		data, err := os.ReadFile(sourceFilePath)
		if err != nil {
			return err
		}

		// Create the destination file
		destinationFile, err := os.Create(destinationFilePath)
		if err != nil {
			return err
		}
		defer func(destinationFile *os.File) {
			_ = destinationFile.Close()
		}(destinationFile)

		// Write the contents into the destination file
		_, err = io.Copy(destinationFile, bytes.NewReader(data))
		if err != nil {
			return err
		}
	}
	return nil
}
