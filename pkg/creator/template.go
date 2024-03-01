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

func RenderTemplate(goOutputDir string, spec *properties.Normalization) error {
	templates := make([]string, 0, 100)
	templates, err := listOfTemplates(templates)
	if err != nil {
		return err
	}

	for _, templateName := range templates {
		filePath := createFullFilePath(goOutputDir, spec, templateName)
		fmt.Printf("Create file %s\n", filePath)

		if err := makeAllDirs(filePath, err); err != nil {
			return err
		}

		outputFile, err := createFile(filePath)
		if err != nil {
			return err
		}
		defer func(outputFile *os.File) {
			err := outputFile.Close()
			if err != nil {

			}
		}(outputFile)

		tmpl, err := parseTemplate(templateName, err)
		if err != nil {
			return err
		}

		err = generateOutputFileFromTemplate(err, tmpl, outputFile, spec)
		if err != nil {
			return err
		}
	}
	return nil
}

func generateOutputFileFromTemplate(err error, tmpl *template.Template, outputFile *os.File, spec *properties.Normalization) error {
	if err = tmpl.Execute(outputFile, spec); err != nil {
		return err
	}
	return nil
}

func parseTemplate(templateName string, err error) (*template.Template, error) {
	templatePath := fmt.Sprintf("templates/%s", templateName)
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

func createFile(filePath string) (*os.File, error) {
	outputFile, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	return outputFile, nil
}

func createFullFilePath(goOutputDir string, spec *properties.Normalization, templateName string) string {
	return fmt.Sprintf("%s/%s/%s.go", goOutputDir, strings.Join(spec.GoSdkPath, "/"), strings.Split(templateName, ".")[0])
}

func makeAllDirs(filePath string, err error) error {
	dirPath := filepath.Dir(filePath)
	if err = os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func listOfTemplates(files []string) ([]string, error) {
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
		return nil, err
	}
	return files, nil
}
