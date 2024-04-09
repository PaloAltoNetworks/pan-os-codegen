package generate

import (
	"bytes"
	"fmt"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/translate/golang/terraform"
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
	CommandType  string
}

// NewCreator initialize Creator instance.
func NewCreator(goOutputDir, templatesDir string, spec *properties.Normalization, commandType string) *Creator {
	return &Creator{
		GoOutputDir:  goOutputDir,
		TemplatesDir: templatesDir,
		Spec:         spec,
		CommandType:  commandType,
	}
}

// RenderTemplate loop through all templates, parse them and render content, which is saved to output file.
func (c *Creator) RenderTemplate(pathType string) error {
	log.Println("Start rendering templates")
	templates, err := c.listOfTemplates()

	if err != nil {
		return fmt.Errorf("error listing templates: %w", err)
	}
	for _, templateName := range templates {
		if err := c.processTemplate(templateName, pathType); err != nil {
			return err
		}
	}
	return nil
}

func (c *Creator) processTemplate(templateName string, pathType string) error {
	filePath := c.createFullFilePath(templateName, pathType)
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
		if err := c.writeDataToFile(filePath, &data); err != nil {
			return err
		}
	}
	return nil
}

func (c *Creator) writeDataToFile(filePath string, data *bytes.Buffer) error {
	//TODO: We have function createFile to handle creation of the file
	outputFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file %s: %w", filePath, err)
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, data)
	return err
}

// createFullFilePath returns a full path for output file generated from template passed as argument to function.
func (c *Creator) createFullFilePath(templateName string, pathType string) string {
	fileBaseName := strings.TrimSuffix(templateName, filepath.Ext(templateName))

	switch pathType {
	case "sdk":
		return filepath.Join(c.GoOutputDir, filepath.Join(c.Spec.GoSdkPath...), fmt.Sprintf("%s.go", fileBaseName))
	case "terraform":
		return filepath.Join(c.GoOutputDir, filepath.Join(c.Spec.TerraformProviderPath...), fmt.Sprintf("%s.go", fileBaseName))
	}
	return ""
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
		if strings.Contains(path, "exclusive") {
			if strings.ToLower(c.Spec.Name) != strings.TrimSuffix(entry.Name(), ".tmpl") {
				return nil
			}
		}
		if strings.HasSuffix(entry.Name(), ".tmpl") {
			files = append(files, entry.Name())
		}
		if strings.ToLower(c.Spec.Name) == strings.TrimSuffix(entry.Name(), ".tmpl") {
			return filepath.SkipAll
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
	var templatesDir string

	if c.Spec.Exclusive != "" {
		templatesDir = c.TemplatesDir + "/exclusive"
	} else {
		templatesDir = c.TemplatesDir
	}

	templatePath := filepath.Join(templatesDir, templateName)
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
		"createResourceList":        terraform.CreateResourceList,
	}
	return template.New(templateName).Funcs(funcMap).ParseFiles(templatePath)
}
