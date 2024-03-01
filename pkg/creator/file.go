package creator

import (
	"fmt"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func PrepareOutputDirs(config *properties.Config) (bool, error) {
	if err := os.MkdirAll(config.Output.GoSdk, 0755); err != nil && !os.IsExist(err) {
		return false, err
	}

	if err := os.MkdirAll(config.Output.TerraformProvider, 0755); err != nil && !os.IsExist(err) {
		return false, err
	}
	return true, nil
}

func PrepareSpecFiles(goOutputDir string, goSdkPath []string) (bool, error) {
	files := make([]string, 0, 100)

	err := filepath.WalkDir("templates", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(entry.Name(), ".tmpl") {
			files = append(files, filepath.Base(path))
		}

		return nil
	})
	if err != nil {
		return false, err
	}

	for _, fileName := range files {
		filePath := fmt.Sprintf("%s/%s/%s.go", goOutputDir, strings.Join(goSdkPath, "/"), strings.Split(fileName, ".")[0])
		fmt.Printf("Create file %s\n", filePath)

		dirPath := filepath.Dir(filePath)
		if err = os.MkdirAll(dirPath, os.ModePerm); err != nil {
			return false, err
		}
		if err = os.WriteFile(filePath, []byte(""), 0644); err != nil {
			return false, err
		}
	}
	return true, nil
}
