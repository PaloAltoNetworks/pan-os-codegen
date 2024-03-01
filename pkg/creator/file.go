package creator

import (
	"fmt"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func CreateOutputDirs(config *properties.Config) (bool, error) {
	if err := os.MkdirAll(config.Output.GoSdk, 0755); err != nil && !os.IsExist(err) {
		return false, err
	}

	if err := os.MkdirAll(config.Output.TerraformProvider, 0755); err != nil && !os.IsExist(err) {
		return false, err
	}
	return true, nil
}

func RenderTemplate(goOutputDir string, spec *properties.Normalization) (bool, error) {
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
		filePath := fmt.Sprintf("%s/%s/%s.go", goOutputDir, strings.Join(spec.GoSdkPath, "/"), strings.Split(fileName, ".")[0])
		fmt.Printf("Create file %s\n", filePath)

		dirPath := filepath.Dir(filePath)
		if err = os.MkdirAll(dirPath, os.ModePerm); err != nil {
			return false, err
		}

		outputFile, err := os.Create(filePath)
		if err != nil {
			return false, err
		}
		defer func(outputFile *os.File) {
			err := outputFile.Close()
			if err != nil {

			}
		}(outputFile)

		templatePath := fmt.Sprintf("templates/%s", fileName)
		tmpl, err := template.ParseFiles(templatePath)
		if err != nil {
			return false, err
		}

		if err = tmpl.Execute(outputFile, spec); err != nil {
			return false, err
		}
	}
	return true, nil
}
