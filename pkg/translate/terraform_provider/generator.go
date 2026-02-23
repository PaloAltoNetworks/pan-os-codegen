package terraform_provider

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

// NameProvider is a type alias for TerraformNameProvider.
type NameProvider = properties.TerraformNameProvider

// NewNameProvider creates a new NameProvider based on given specifications.
func NewNameProvider(spec *properties.Normalization, resourceTyp properties.ResourceType) *NameProvider {
	return properties.NewTerraformNameProvider(spec, resourceTyp)
}

// GenerateTerraformProvider handles the generation of Terraform resources and data sources.
type GenerateTerraformProvider struct{}

// createTemplate parses the provided template string using the given FuncMap and returns a new template.
func (g *GenerateTerraformProvider) createTemplate(resourceType string, spec *properties.Normalization, templateStr string, funcMap template.FuncMap) (*template.Template, error) {
	templateName := fmt.Sprintf("terraform-%s-%s", resourceType, spec.Name)

	// Try to load from file if templateStr looks like a file path
	var tmplContent string
	if strings.HasSuffix(templateStr, ".tmpl") || strings.Contains(templateStr, "/") {
		templatePath := filepath.Join("templates", "terraform-provider", templateStr)
		if content, err := os.ReadFile(templatePath); err == nil {
			tmplContent = string(content)
		} else {
			// Fallback to embedded string if file doesn't exist
			tmplContent = templateStr
		}
	} else {
		tmplContent = templateStr
	}

	return template.New(templateName).Funcs(funcMap).Parse(tmplContent)
}

// executeTemplate executes the provided resource template using the given spec and returns an error if it fails.
func (g *GenerateTerraformProvider) executeTemplate(template *template.Template, spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile, resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, names *NameProvider) error {
	var renderedTemplate strings.Builder
	if err := template.Execute(&renderedTemplate, spec); err != nil {
		return fmt.Errorf("error executing %v template: %v", resourceTyp, err)
	}
	renderedTemplate.WriteString("\n")
	return g.updateProviderFile(spec, &renderedTemplate, terraformProvider, resourceTyp, schemaTyp, names)
}

// updateProviderFile updates the Terraform provider file by appending the rendered template
// to the appropriate slice in the TerraformProviderFile based on the provided resourceType.
func (g *GenerateTerraformProvider) updateProviderFile(spec *properties.Normalization, renderedTemplate *strings.Builder, terraformProvider *properties.TerraformProviderFile, resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, names *NameProvider) error {
	if schemaTyp == properties.SchemaProvider {
		terraformProvider.Code = renderedTemplate
	} else {
		if _, err := terraformProvider.Code.WriteString(renderedTemplate.String()); err != nil {
			return fmt.Errorf("error writing %v template: %v", resourceTyp, err)
		}
	}
	return g.appendResourceType(spec, terraformProvider, resourceTyp, schemaTyp, names)
}

// appendResourceType appends the given struct name to the appropriate slice in the TerraformProviderFile
// based on the provided resourceType.
func (g *GenerateTerraformProvider) appendResourceType(spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile, resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, names *NameProvider) error {
	var flags properties.TerraformSpecFlags
	switch schemaTyp {
	case properties.SchemaDataSource:
		flags |= properties.TerraformSpecDatasource
		terraformProvider.DataSources = append(terraformProvider.DataSources, names.DataSourceStructName)
	case properties.SchemaResource:
		flags |= properties.TerraformSpecResource
		terraformProvider.Resources = append(terraformProvider.Resources, names.ResourceStructName)
	case properties.SchemaEphemeralResource:
		flags |= properties.TerraformSpecEphemeralResource
		terraformProvider.EphemeralResources = append(terraformProvider.EphemeralResources, names.ResourceStructName)
	case properties.SchemaAction:
		terraformProvider.Actions = append(terraformProvider.Actions, names.ActionStructName())
	case properties.SchemaProvider, properties.SchemaCommon:
	default:
		panic(fmt.Sprintf("unsupported schemaTyp: '%s'", schemaTyp))
	}

	switch resourceTyp {
	case properties.ResourceEntry, properties.ResourceEntryPlural, properties.ResourceUuid, properties.ResourceUuidPlural:
		if !spec.TerraformProviderConfig.SkipResource {
			flags |= properties.TerraformSpecImportable
		}
	case properties.ResourceCustom, properties.ResourceConfig:
	}

	terraformProvider.SpecMetadata[names.MetaName] = properties.TerraformProviderSpecMetadata{
		ResourceSuffix: names.MetaName,
		StructName:     names.StructName,
		Flags:          flags,
	}
	return nil
}

// generateTerraformEntityTemplate is the common logic for generating Terraform resources and data sources.
func (g *GenerateTerraformProvider) generateTerraformEntityTemplate(resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, names *NameProvider, spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile, templateStr string, funcMap template.FuncMap) error {
	if templateStr == "" {
		return nil
	}

	var resourceType string
	switch schemaTyp {
	case properties.SchemaDataSource:
		resourceType = "DataSource"
	case properties.SchemaResource:
		resourceType = "Resource"
	case properties.SchemaEphemeralResource:
		resourceType = "EphemeralResource"
	case properties.SchemaCommon:
		resourceType = "Common"
	case properties.SchemaProvider:
		resourceType = "ProviderFile"
	case properties.SchemaAction:
		resourceType = "Action"
	default:
		panic(fmt.Sprintf("unsupported schemaTyp: '%+v'", schemaTyp))
	}

	template, err := g.createTemplate(resourceType, spec, templateStr, funcMap)
	if err != nil {
		log.Fatalf("Error creating template: %v", err)
		return err
	}
	return g.executeTemplate(template, spec, terraformProvider, resourceTyp, schemaTyp, names)
}
