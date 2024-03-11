package generate

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/translate"
)

type Creator struct {
	GoOutputDir  string
	TemplatesDir string
	Spec         *properties.Normalization
}

func NewCreator(goOutputDir, templatesDir string, spec *properties.Normalization) *Creator {
	return &Creator{
		GoOutputDir:  goOutputDir,
		TemplatesDir: templatesDir,
		Spec:         spec,
	}
}

func (c *Creator) RenderTemplate() error {
	log.Println("Start rendering templates")

	templates, err := c.listOfTemplates()
	if err != nil {
		return fmt.Errorf("error listing templates: %w", err)
	}

	for _, templateName := range templates {
		filePath := c.createFullFilePath(templateName)
		log.Printf("Creating file: %s\n", filePath)

		if err := c.makeAllDirs(filePath); err != nil {
			return fmt.Errorf("error creating directories for %s: %w", filePath, err)
		}

		tmpl, err := c.parseTemplate(templateName)
		if err != nil {
			return fmt.Errorf("error parsing template %s: %w", templateName, err)
		}

		var data bytes.Buffer
		if err := tmpl.Execute(&data, c.Spec); err != nil {
			return fmt.Errorf("error executing template %s: %w", templateName, err)
		}
		if data.Len() > 0 {
			outputFile, err := os.Create(filePath)
			if err != nil {
				return fmt.Errorf("error creating file %s: %w", filePath, err)
			}
			defer outputFile.Close()

			_, err = io.Copy(outputFile, &data)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Creator) createFullFilePath(templateName string) string {
	fileBaseName := strings.TrimSuffix(templateName, filepath.Ext(templateName))
	return filepath.Join(c.GoOutputDir, filepath.Join(c.Spec.GoSdkPath...), fileBaseName+".go")
}

func (c *Creator) listOfTemplates() ([]string, error) {
	var files []string
	err := filepath.WalkDir(c.TemplatesDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if strings.HasSuffix(entry.Name(), ".tmpl") {
			files = append(files, entry.Name())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (c *Creator) makeAllDirs(filePath string) error {
	dirPath := filepath.Dir(filePath)
	return os.MkdirAll(dirPath, os.ModePerm)
}

func (c *Creator) createFile(filePath string) (*os.File, error) {
	outputFile, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	return outputFile, nil
}

func (c *Creator) parseTemplate(templateName string) (*template.Template, error) {
	templatePath := filepath.Join(c.TemplatesDir, templateName)
	funcMap := template.FuncMap{
		"packageName":            translate.PackageName,
		"locationType":           translate.LocationType,
		"specParamType":          translate.SpecParamType,
		"xmlParamType":           translate.XmlParamType,
		"xmlTag":                 translate.XmlTag,
		"specifyEntryAssignment": translate.SpecifyEntryAssignment,
		"specMatchesFunction":    translate.SpecMatchesFunction,
		"omitEmpty":              translate.OmitEmpty,
		"contains": func(full, part string) bool {
			return strings.Contains(full, part)
		},
		"subtract": func(a, b int) int {
			return a - b
		},
		"asEntryXpath": translate.AsEntryXpath,
		"nestedSpecs":  translate.NestedSpecs,
	}
	return template.New(templateName).Funcs(funcMap).ParseFiles(templatePath)
}
