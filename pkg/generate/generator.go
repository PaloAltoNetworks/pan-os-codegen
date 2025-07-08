package generate

import (
	"bytes"
	"errors"
	"fmt"
	"go/format"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/object"
	codegentmpl "github.com/paloaltonetworks/pan-os-codegen/pkg/template"
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
	templates, err := c.listOfTemplates()
	if err != nil {
		return fmt.Errorf("error listing templates: %w", err)
	}

	for _, templateName := range templates {
		filePath := c.createFullFilePath(templateName)
		slog.Debug("Creating target file", "path", filePath)

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
func (c *Creator) RenderTerraformProviderFile(spec *properties.Normalization, typ properties.ResourceType) ([]string, []string, []string, map[string]properties.TerraformProviderSpecMetadata, error) {
	var name string
	switch typ {
	case properties.ResourceUuidPlural:
		name = fmt.Sprintf("%s_%s", spec.TerraformProviderConfig.Suffix, spec.TerraformProviderConfig.PluralName)
	case properties.ResourceEntryPlural:
		name = spec.TerraformProviderConfig.PluralSuffix
	case properties.ResourceEntry, properties.ResourceUuid, properties.ResourceCustom:
		name = spec.Name
	case properties.ResourceConfig:
	}

	terraformProvider := properties.NewTerraformProviderFile(name)
	tfp := terraform_provider.GenerateTerraformProvider{}

	if err := tfp.GenerateTerraformDataSource(typ, spec, terraformProvider); err != nil {
		return nil, nil, nil, nil, err
	}

	if err := tfp.GenerateTerraformResource(typ, spec, terraformProvider); err != nil {
		return nil, nil, nil, nil, err
	}

	if err := tfp.GenerateCommonCode(typ, spec, terraformProvider); err != nil {
		return nil, nil, nil, nil, err
	}

	if err := tfp.GenerateTerraformProviderFile(spec, terraformProvider); err != nil {
		return nil, nil, nil, nil, err
	}

	var filePath string
	switch typ {
	case properties.ResourceUuidPlural:
		name = fmt.Sprintf("%s_%s", spec.TerraformProviderConfig.Suffix, spec.TerraformProviderConfig.PluralName)
		filePath = c.createTerraformProviderFilePath(name)
	case properties.ResourceEntryPlural:
		name = spec.TerraformProviderConfig.PluralSuffix
		filePath = c.createTerraformProviderFilePath(name)
	case properties.ResourceEntry, properties.ResourceUuid, properties.ResourceCustom, properties.ResourceConfig:
		filePath = c.createTerraformProviderFilePath(spec.TerraformProviderConfig.Suffix)
	}

	if err := c.writeFormattedContentToFile(filePath, terraformProvider.Code.String()); err != nil {
		return nil, nil, nil, nil, err
	}

	return terraformProvider.DataSources, terraformProvider.Resources, terraformProvider.EphemeralResources, terraformProvider.SpecMetadata, nil
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
		var formattedCode []byte
		formattedCode, err = format.Source(data.Bytes())
		if err != nil {
			log.Printf("Failed to format source code: %s, %s", filePath, err.Error())
			formattedCode = data.Bytes()
		}
		formattedBuf := bytes.NewBuffer(formattedCode)

		writeErr := c.createAndWriteFile(filePath, formattedBuf)
		if writeErr != nil {
			return errors.Join(err, writeErr)
		}
	}
	return err
}

// writeFormattedContentToFile formats the content and writes it to a file.
func (c *Creator) writeFormattedContentToFile(filePath, content string) error {
	var formattedCode []byte
	var formatErr error
	formattedCode, formatErr = format.Source([]byte(content))
	if formatErr != nil {
		log.Printf("Failed to format target path: %s", filePath)
		formattedCode = []byte(content)
	}
	formattedBuf := bytes.NewBuffer(formattedCode)

	return errors.Join(formatErr, c.createFileAndWriteContent(filePath, formattedBuf))
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
		"Map": codegentmpl.TemplateMap,
		"renderImports": func(templateTypes ...string) (string, error) {
			return translate.RenderImports(c.Spec, templateTypes...)
		},
		"SupportedMethod":           func(method object.GoSdkMethod) bool { return c.Spec.SupportedMethod(method) },
		"RenderEntryImportStructs":  func() (string, error) { return translate.RenderEntryImportStructs(c.Spec) },
		"packageName":               translate.PackageName,
		"locationType":              translate.LocationType,
		"specParamType":             translate.SpecParamType,
		"xmlName":                   translate.XmlName,
		"xmlParamType":              translate.XmlParamType,
		"xmlTag":                    translate.XmlTag,
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
		"generateEntryXpath": translate.GenerateEntryXpath,
		"RenderApiStructs": func(spec *properties.Normalization) (string, error) {
			return translate.RenderEntryApiStructs(spec)
		},
		"RenderXmlStructs": func(spec *properties.Normalization) (string, error) {
			return translate.RenderEntryXmlStructs(spec)
		},
		"RenderXmlContainerNormalizers": func(spec *properties.Normalization) (string, error) {
			return translate.RenderXmlContainerNormalizers(spec)
		},
		"RenderXmlContainerSpecifiers": func(spec *properties.Normalization) (string, error) {
			return translate.RenderXmlContainerSpecifiers(spec)
		},
		"RenderToXmlMarshallers": func(spec *properties.Normalization) (string, error) {
			return translate.RenderToXmlMarshalers(spec)
		},
		"RenderSpecMatchers": func(spec *properties.Normalization) (string, error) {
			return translate.RenderSpecMatchers(spec)
		},
		"createGoSuffixFromVersion": translate.CreateGoSuffixFromVersionTmpl,
		"paramNotSkipped":           translate.ParamNotSkippedTmpl,
		"paramSupportedInVersion":   translate.ParamSupportedInVersionTmpl,
		"xmlPathSuffixes":           translate.XmlPathSuffixes,
		"underscore":                naming.Underscore,
		"camelCase":                 naming.CamelCase,
		"lowerCamelCase":            naming.LowerCamelCase,
	}
	return template.New(templateName).Funcs(funcMap).ParseFiles(templatePath)
}
