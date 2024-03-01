package generate

import (
	"fmt"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Creator struct {
	GoOutputDir string
	Spec        *properties.Normalization
}

func NewCreator(goOutputDir string, spec *properties.Normalization) *Creator {
	return &Creator{
		GoOutputDir: goOutputDir,
		Spec:        spec,
	}
}

func (c *Creator) RenderTemplate() error {
	templates := make([]string, 0, 100)
	templates, err := c.listOfTemplates(templates)
	if err != nil {
		return err
	}

	for _, templateName := range templates {
		filePath := c.createFullFilePath(c.GoOutputDir, c.Spec, templateName)
		fmt.Printf("Create file %s\n", filePath)

		if err := c.makeAllDirs(filePath, err); err != nil {
			return err
		}

		outputFile, err := c.createFile(filePath)
		if err != nil {
			return err
		}
		defer func(outputFile *os.File) {
			err := outputFile.Close()
			if err != nil {

			}
		}(outputFile)

		tmpl, err := c.parseTemplate(templateName, err)
		if err != nil {
			return err
		}

		err = c.generateOutputFileFromTemplate(err, tmpl, outputFile, c.Spec)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Creator) generateOutputFileFromTemplate(err error, tmpl *template.Template, outputFile *os.File, spec *properties.Normalization) error {
	if err = tmpl.Execute(outputFile, spec); err != nil {
		return err
	}
	return nil
}

func (c *Creator) parseTemplate(templateName string, err error) (*template.Template, error) {
	templatePath := fmt.Sprintf("templates/%s", templateName)
	funcMap := template.FuncMap{
		"packageName": naming.PackageName,
	}
	tmpl, err := template.New(templateName).Funcs(funcMap).ParseFiles(templatePath)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

func (c *Creator) createFile(filePath string) (*os.File, error) {
	outputFile, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	return outputFile, nil
}

func (c *Creator) createFullFilePath(goOutputDir string, spec *properties.Normalization, templateName string) string {
	return fmt.Sprintf("%s/%s/%s.go", goOutputDir, strings.Join(spec.GoSdkPath, "/"), strings.Split(templateName, ".")[0])
}

func (c *Creator) makeAllDirs(filePath string, err error) error {
	dirPath := filepath.Dir(filePath)
	if err = os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func (c *Creator) listOfTemplates(files []string) ([]string, error) {
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
