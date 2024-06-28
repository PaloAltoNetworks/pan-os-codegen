package generate

import (
	"bytes"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

// CopyAssets copy assets (static files) according to configuration.
func CopyAssets(config *properties.Config, commandType properties.CommandType) error {
	for _, asset := range config.Assets {
		files, err := listAssets(asset)
		if err != nil {
			return err
		}

		log.Printf("%v", asset)
		if asset.Target.GoSdk && commandType == properties.CommandTypeSDK {
			if err = copyAsset(config.Output.GoSdk, asset, files); err != nil {
				return err
			}
		}

		if asset.Target.TerraformProvider && commandType == properties.CommandTypeTerraform {
			if err = copyAsset(config.Output.TerraformProvider, asset, files); err != nil {
				return err
			}
		}
	}

	return nil
}

// listAssets walk through directory and get list of all assets (static files).
func listAssets(asset *properties.Asset) ([]string, error) {
	var files []string

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

// copyAsset copy single asset, which may contain multiple files.
func copyAsset(target string, asset *properties.Asset, files []string) error {
	// Prepare destination path
	destinationDir := filepath.Join(target, asset.Destination)

	// Create the destination directory if it doesn't exist
	if err := os.MkdirAll(destinationDir, os.ModePerm); err != nil {
		return err
	}

	for _, sourceFilePath := range files {
		// Prepare destination path
		sourceFileDir := filepath.Dir(sourceFilePath)
		sourceFileDirRelative := strings.TrimPrefix(filepath.Clean(sourceFileDir), filepath.Clean(asset.Source))

		destinationSubDir := filepath.Join(destinationDir, sourceFileDirRelative)
		if _, err := os.Stat(destinationSubDir); os.IsNotExist(err) {
			if err := os.MkdirAll(destinationSubDir, os.ModePerm); err != nil {
				return err
			}
		}

		destinationFilePath := filepath.Join(destinationDir, sourceFileDirRelative, filepath.Base(sourceFilePath))
		log.Printf("Copy file from %s to %s\n", sourceFilePath, destinationFilePath)

		// Read the contents of the source files
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
