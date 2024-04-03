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

// NewCreator initialize Creator instance.
func NewCreator(goOutputDir, templatesDir string, spec *properties.Normalization) *Creator {
	return &Creator{
		GoOutputDir:  goOutputDir,
		TemplatesDir: templatesDir,
		Spec:         spec,
	}
}

// RenderTemplate loop through all templates, parse them and render content, which is saved to output file.
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
		// If from template no data was rendered (e.g. for DNS spec entry should not be created),
		// then we don't need to create empty file (e.g. `entry.go`) with no content
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

// createFullFilePath returns a full path for output file generated from template passed as argument to function.
func (c *Creator) createFullFilePath(templateName string) string {
	fileBaseName := strings.TrimSuffix(templateName, filepath.Ext(templateName))
	return filepath.Join(c.GoOutputDir, filepath.Join(c.Spec.GoSdkPath...), fmt.Sprintf("%s.go", fileBaseName))
}

// listOfTemplates return list of templates defined in TemplatesDir.
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

// makeAllDirs creates all required directories, which are in the file path.
func (c *Creator) makeAllDirs(filePath string) error {
	dirPath := filepath.Dir(filePath)
	return os.MkdirAll(dirPath, os.ModePerm)
}

// createFile just create a file and return it.
func (c *Creator) createFile(filePath string) (*os.File, error) {
	outputFile, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	return outputFile, nil
}

// parseTemplate parse template passed as argument and with function map defined below.
func (c *Creator) parseTemplate(templateName string) (*template.Template, error) {
	templatePath := filepath.Join(c.TemplatesDir, templateName)
	funcMap := template.FuncMap{
		"renderImports":             translate.RenderImports,
		"packageName":               translate.PackageName,
		"locationType":              translate.LocationType,
		"specParamType":             translate.SpecParamType,
		"xmlParamType":              translate.XmlParamType,
		"xmlName":                   translate.XmlName,
		"xmlTag":                    translate.XmlTag,
		"specifyEntryAssignment":    translate.SpecifyEntryAssignment,
		"normalizeAssignment":       translate.NormalizeAssignment,
		"specMatchesFunction":       translate.SpecMatchesFunction,
		"nestedSpecMatchesFunction": translate.NestedSpecMatchesFunction,
		"omitEmpty":                 translate.OmitEmpty,
		"contains":                  strings.Contains,
		"subtract": func(a, b int) int {
			return a - b
		},
		"generateEntryXpath":        translate.GenerateEntryXpathForLocation,
		"nestedSpecs":               translate.NestedSpecs,
		"createGoSuffixFromVersion": translate.CreateGoSuffixFromVersion,
		"paramSupportedInVersion":   translate.ParamSupportedInVersion,
		"xmlPathSuffixes":           translate.XmlPathSuffixes,
	}
	return template.New(templateName).Funcs(funcMap).ParseFiles(templatePath)
}
