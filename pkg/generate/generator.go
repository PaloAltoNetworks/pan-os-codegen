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
			if err := c.createAndWriteFile(filePath, &data); err != nil {
				return fmt.Errorf("error creating and writing to file %s: %w", filePath, err)
			}
		}
	}
	return nil
}

// RenderTerraformProvider generates a Go file for a Terraform provider based on the provided TerraformProviderFile and Normalization arguments.
// It calls terraform.GenerateTerraformResource() passing the Normalization specification and the TerraformProviderFile.
func (c *Creator) RenderTerraformProvider(terraformProvider *properties.TerraformProviderFile, spec *properties.Normalization) error {
	tfp := terraform.GenerateTerraformProvider{}

	if err := tfp.GenerateTerraformDataSource(spec, terraformProvider); err != nil {
		return err
	}

	if err := tfp.GenerateTerraformResource(spec, terraformProvider); err != nil {
		return err
	}

	if err := tfp.GenerateTerraformProviderFile(spec, terraformProvider); err != nil {
		return err
	}
	filePath := c.createTerraformProviderFilePath(*spec.TerraformProviderConfig.Suffix)

	content := bytes.NewBufferString(terraformProvider.Code.String())
	if err := c.createFileAndWriteContent(filePath, content); err != nil {
		return err
	}

	return nil
}

// createTerraformProviderFilePath returns a file path for a Terraform provider based on the provided suffix.
func (c *Creator) createTerraformProviderFilePath(terraformProviderSuffix string) string {
	fileName := fmt.Sprintf("%s_object.go", terraformProviderSuffix)
	return filepath.Join(c.GoOutputDir, "internal", fileName)
}

// createFileAndWriteContent creates a new file at the specified filePath and writes the content from the content buffer to the file.
func (c *Creator) createFileAndWriteContent(filePath string, content *bytes.Buffer) error {
	if err := c.makeAllDirs(filePath); err != nil {
		return fmt.Errorf("error creating directories for %s: %w", filePath, err)
	}
	if err := c.createAndWriteFile(filePath, content); err != nil {
		return err
	}
	return nil
}

// createAndWriteFile creates a new file at the specified filePath and writes the content from the content buffer to the file.
// If an error occurs during file creation or content writing, it returns an error. The file is automatically closed after writing.
func (c *Creator) createAndWriteFile(filePath string, content *bytes.Buffer) error {
	outputFile, err := c.createFile(filePath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	return writeContentToFile(content, outputFile)
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
		return nil, fmt.Errorf("error creating file %s: %w", filePath, err)
	}
	return outputFile, nil
}

func writeContentToFile(content *bytes.Buffer, file *os.File) error {
	_, err := io.Copy(file, content)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}
	return nil
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
