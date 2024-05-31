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

	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/translate"
)

type Creator struct {
	GoOutputDir  string
	TemplatesDir string
	Spec         *properties.Normalization
}

// NewCreator initializes a Creator instance.
func NewCreator(goOutputDir, templatesDir string, spec *properties.Normalization) *Creator {
	return &Creator{
		GoOutputDir:  goOutputDir,
		TemplatesDir: templatesDir,
		Spec:         spec,
	}
}

// RenderTemplate loops through all templates, parses them, and renders content, which is saved to the output file.
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

		if err := c.processTemplate(templateName, filePath); err != nil {
			return err
		}
	}
	return nil
}

// RenderTerraformProviderFile generates a Go file for a Terraform provider based on the provided TerraformProviderFile and Normalization arguments.
func (c *Creator) RenderTerraformProviderFile(terraformProvider *properties.TerraformProviderFile, spec *properties.Normalization) error {
	tfp := terraform_provider.GenerateTerraformProvider{}

	if err := tfp.GenerateTerraformDataSource(spec, terraformProvider); err != nil {
		return err
	}

	if err := tfp.GenerateTerraformResource(spec, terraformProvider); err != nil {
		return err
	}

	if err := tfp.GenerateTerraformProviderFile(spec, terraformProvider); err != nil {
		return err
	}
	filePath := c.createTerraformProviderFilePath(spec.TerraformProviderConfig.Suffix)

	return c.writeFormattedContentToFile(filePath, terraformProvider.Code.String())
}

// RenderTerraformProvider generates and writes a Terraform provider file.
func (c *Creator) RenderTerraformProvider(terraformProvider *properties.TerraformProviderFile, spec *properties.Normalization, providerConfig properties.TerraformProvider) error {
	tfp := terraform_provider.GenerateTerraformProvider{}
	if err := tfp.GenerateTerraformProvider(terraformProvider, spec, providerConfig); err != nil {
		return err
	}
	filePath := c.createTerraformProviderFilePath(spec.Name)

	return c.writeFormattedContentToFile(filePath, terraformProvider.Code.String())
}

// processTemplate processes a single template and writes the rendered content to a file.
func (c *Creator) processTemplate(templateName, filePath string) error {
	tmpl, err := c.parseTemplate(templateName)
	if err != nil {
		return fmt.Errorf("error parsing template %s: %w", templateName, err)
	}

	var data bytes.Buffer
	if err := tmpl.Execute(&data, c.Spec); err != nil {
		return fmt.Errorf("error executing template %s: %w", templateName, err)
	}

	// If no data was rendered from the template, skip creating an empty file.
	if data.Len() > 0 {
		//formattedCode, err := format.Source(data.Bytes())
		//if err != nil {
		//	return fmt.Errorf("error formatting code %w", err)
		//}
		formattedBuf := bytes.NewBuffer(data.Bytes())

		if err := c.createAndWriteFile(filePath, formattedBuf); err != nil {
			return fmt.Errorf("error creating and writing to file %s: %w", filePath, err)
		}
	}
	return nil
}

// writeFormattedContentToFile formats the content and writes it to a file.
func (c *Creator) writeFormattedContentToFile(filePath, content string) error {
	//formattedCode, err := format.Source([]byte(content))
	//if err != nil {
	//	return fmt.Errorf("error formatting code %w", err)
	//}
	formattedBuf := bytes.NewBuffer([]byte(content))

	return c.createFileAndWriteContent(filePath, formattedBuf)
}

// createTerraformProviderFilePath returns a file path for a Terraform provider based on the provided suffix.
func (c *Creator) createTerraformProviderFilePath(terraformProviderFileName string) string {
	terraformProviderFileNameSuffix := "_object"

	if terraformProviderFileName == "provider" {
		terraformProviderFileNameSuffix = ""
	}

	fileName := fmt.Sprintf("%s%s.go", terraformProviderFileName, terraformProviderFileNameSuffix)
	return filepath.Join(c.GoOutputDir, "internal/provider", fileName)
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

// createFullFilePath returns a full path for the output file generated from the template passed as an argument.
func (c *Creator) createFullFilePath(templateName string) string {
	fileBaseName := strings.TrimSuffix(templateName, filepath.Ext(templateName))
	return filepath.Join(c.GoOutputDir, filepath.Join(c.Spec.GoSdkPath...), fmt.Sprintf("%s.go", fileBaseName))
}

// listOfTemplates returns a list of templates defined in TemplatesDir.
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

// makeAllDirs creates all required directories in the file path.
func (c *Creator) makeAllDirs(filePath string) error {
	dirPath := filepath.Dir(filePath)
	return os.MkdirAll(dirPath, os.ModePerm)
}

// createFile creates a file and returns it.
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

// parseTemplate parses the template passed as an argument with the function map defined below.
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
		"add": func(a, b int) int {
			return a + b
		},
		"subtract": func(a, b int) int {
			return a - b
		},
		"generateEntryXpath":        translate.GenerateEntryXpath,
		"nestedSpecs":               translate.NestedSpecs,
		"createGoSuffixFromVersion": translate.CreateGoSuffixFromVersion,
		"paramSupportedInVersion":   translate.ParamSupportedInVersion,
		"xmlPathSuffixes":           translate.XmlPathSuffixes,
		"underscore":                naming.Underscore,
		"camelCase":                 naming.CamelCase,
	}
	return template.New(templateName).Funcs(funcMap).ParseFiles(templatePath)
}
