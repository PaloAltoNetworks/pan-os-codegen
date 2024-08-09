package generate

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/translate"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/translate/terraform_provider"
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
func (c *Creator) RenderTerraformProviderFile(spec *properties.Normalization, typ properties.ResourceType) ([]string, []string, error) {
	var name string
	if typ == properties.ResourceUuidPlural {
		name = fmt.Sprintf("%s_%s", spec.TerraformProviderConfig.Suffix, spec.TerraformProviderConfig.PluralName)
	} else {
		name = spec.Name
	}

	terraformProvider := properties.NewTerraformProviderFile(name)
	tfp := terraform_provider.GenerateTerraformProvider{}

	if err := tfp.GenerateTerraformDataSource(typ, spec, terraformProvider); err != nil {
		return nil, nil, err
	}

	if err := tfp.GenerateTerraformResource(typ, spec, terraformProvider); err != nil {
		return nil, nil, err
	}

	if err := tfp.GenerateCommonCode(typ, spec, terraformProvider); err != nil {
		return nil, nil, err
	}

	if err := tfp.GenerateTerraformProviderFile(spec, terraformProvider); err != nil {
		return nil, nil, err
	}

	var filePath string
	if typ == properties.ResourceUuidPlural {
		name = fmt.Sprintf("%s_%s", spec.TerraformProviderConfig.Suffix, spec.TerraformProviderConfig.PluralName)
		filePath = c.createTerraformProviderFilePath(name)
	} else {
		filePath = c.createTerraformProviderFilePath(spec.TerraformProviderConfig.Suffix)
	}

	if err := c.writeFormattedContentToFile(filePath, terraformProvider.Code.String()); err != nil {
		return nil, nil, err
	}

	return terraformProvider.DataSources, terraformProvider.Resources, nil
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
	dataLength := len(bytes.TrimSpace(data.Bytes()))
	if dataLength > 0 {
		formattedCode, err := format.Source(data.Bytes())
		if err != nil {
			return fmt.Errorf("error formatting code %w", err)
		}
		formattedBuf := bytes.NewBuffer(formattedCode)

		if err := c.createAndWriteFile(filePath, formattedBuf); err != nil {
			return fmt.Errorf("error creating and writing to file %s: %w", filePath, err)
		}
	}
	return nil
}

// writeFormattedContentToFile formats the content and writes it to a file.
func (c *Creator) writeFormattedContentToFile(filePath, content string) error {
	formattedCode, err := format.Source([]byte(content))
	if err != nil {
		log.Printf("provided content: %s", content)
		return fmt.Errorf("error formatting code: %w", err)
	}
	formattedBuf := bytes.NewBuffer(formattedCode)

	return c.createFileAndWriteContent(filePath, formattedBuf)
}

// createTerraformProviderFilePath returns a file path for a Terraform provider based on the provided suffix.
func (c *Creator) createTerraformProviderFilePath(terraformProviderFileName string) string {
	fileName := fmt.Sprintf("%s.go", terraformProviderFileName)
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
	defer func(outputFile *os.File) {
		_ = outputFile.Close()
	}(outputFile)

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
