package codegen

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/generate"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/load"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

type CommandType string

const (
	CommandTypeSDK       CommandType = "sdk"
	CommandTypeTerraform CommandType = "terraform"
)

type Command struct {
	ctx          context.Context
	args         []string
	specs        []string
	commandType  properties.CommandType
	templatePath string
}

func NewCommand(ctx context.Context, commandType properties.CommandType, args ...string) (*Command, error) {
	var templatePath string
	switch commandType {
	case properties.CommandTypeSDK:
		templatePath = "templates/sdk"
	case properties.CommandTypeTerraform:
		templatePath = "templates/terraform"
	default:
		return nil, fmt.Errorf("unsupported command type: %s", commandType)
	}

	return &Command{
		ctx:          ctx,
		args:         args,
		commandType:  commandType,
		templatePath: templatePath,
	}, nil
}

// deriveSubcategoryFromPath extracts the subcategory from the spec file path.
// It maps directory names to proper subcategory names.
// For example: specs/network/interface.yaml -> "Network"
func deriveSubcategoryFromPath(specPath string) string {
	// Extract the directory name between specs/ and the filename
	dir := filepath.Dir(specPath)
	parts := strings.Split(filepath.ToSlash(dir), "/")

	// Find the part after "specs"
	var category string
	for i, part := range parts {
		if part == "specs" && i+1 < len(parts) {
			category = parts[i+1]
			break
		}
	}

	// Map directory names to subcategory names
	subcategoryMap := map[string]string{
		"network":  "Network",
		"objects":  "Objects",
		"device":   "Device",
		"panorama": "Panorama",
		"policies": "Policies",
		"actions":  "", // empty for actions
		"schema":   "", // empty for schema
	}

	if subcategory, ok := subcategoryMap[category]; ok {
		return subcategory
	}

	// Default to empty string if no mapping found
	return ""
}

// generateTfplugindocsTemplates creates individual documentation templates for each resource/data source
// with the correct subcategory. These templates are used by terraform-plugin-docs when generating documentation.
func generateTfplugindocsTemplates(outputDir string, specMetadata map[string]properties.TerraformProviderSpecMetadata) error {
	templatesDir := filepath.Join(outputDir, "templates")
	resourcesDir := filepath.Join(templatesDir, "resources")
	dataSourcesDir := filepath.Join(templatesDir, "data-sources")

	if err := os.MkdirAll(resourcesDir, 0755); err != nil {
		return fmt.Errorf("error creating resources templates directory: %w", err)
	}
	if err := os.MkdirAll(dataSourcesDir, 0755); err != nil {
		return fmt.Errorf("error creating data sources templates directory: %w", err)
	}

	resourceCount := 0
	dataSourceCount := 0

	for resourceSuffix, metadata := range specMetadata {
		slog.Debug("Processing spec metadata", "resourceSuffix", resourceSuffix, "subcategory", metadata.Subcategory, "flags", metadata.Flags)

		// Generate template for resources
		if metadata.Flags&properties.TerraformSpecResource != 0 {
			subcategory := metadata.Subcategory
			if subcategory == "" {
				subcategory = "Uncategorized"
			}

			resourceTemplate := fmt.Sprintf(`---
page_title: "{{.Name}} {{.Type}} - {{.ProviderName}}"
subcategory: "%s"
description: |-
{{ .Description | plainmarkdown | trimspace | prefixlines "  " }}
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

{{ if .HasExample -}}
## Example Usage

{{ tffile .ExampleFile }}
{{- end }}

{{ .SchemaMarkdown | trimspace }}

{{ if .HasImport -}}
## Import

Import is supported using the following syntax:

{{ codefile "shell" .ImportFile }}
{{- end }}
`, subcategory)

			// Remove leading underscore from resourceSuffix for template filename
			// terraform-plugin-docs automatically prepends the provider name (panos_)
			templateName := strings.TrimPrefix(resourceSuffix, "_")
			resourcePath := filepath.Join(resourcesDir, fmt.Sprintf("%s.md.tmpl", templateName))
			if err := os.WriteFile(resourcePath, []byte(resourceTemplate), 0644); err != nil {
				return fmt.Errorf("error writing resource template %s: %w", resourcePath, err)
			}
			resourceCount++
		}

		// Generate template for data sources
		if metadata.Flags&properties.TerraformSpecDatasource != 0 {
			subcategory := metadata.Subcategory
			if subcategory == "" {
				subcategory = "Uncategorized"
			}

			dataSourceTemplate := fmt.Sprintf(`---
page_title: "{{.Name}} {{.Type}} - {{.ProviderName}}"
subcategory: "%s"
description: |-
{{ .Description | plainmarkdown | trimspace | prefixlines "  " }}
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

{{ if .HasExample -}}
## Example Usage

{{ tffile .ExampleFile }}
{{- end }}

{{ .SchemaMarkdown | trimspace }}
`, subcategory)

			// Remove leading underscore from resourceSuffix for template filename
			// terraform-plugin-docs automatically prepends the provider name (panos_)
			templateName := strings.TrimPrefix(resourceSuffix, "_")
			dataSourcePath := filepath.Join(dataSourcesDir, fmt.Sprintf("%s.md.tmpl", templateName))
			if err := os.WriteFile(dataSourcePath, []byte(dataSourceTemplate), 0644); err != nil {
				return fmt.Errorf("error writing data source template %s: %w", dataSourcePath, err)
			}
			dataSourceCount++
		}
	}

	slog.Info("Generated tfplugindocs templates", "resources", resourceCount, "dataSources", dataSourceCount, "templatesDir", templatesDir)
	return nil
}

func (c *Command) Setup() error {
	var err error
	if c.specs == nil {
		c.specs, err = properties.GetNormalizations()
		if err != nil {
			return fmt.Errorf("error getting normalizations: %s", err)
		}
	}
	return nil
}

func (c *Command) Execute() error {
	if len(c.args) == 0 {
		return fmt.Errorf("path to configuration file is required")
	}
	configPath := c.args[0]

	slog.Info("Generating code", "type", c.commandType)

	content, err := load.File(configPath)
	if err != nil {
		return fmt.Errorf("error loading %s - %s", configPath, err)
	}

	config, err := properties.ParseConfig(content)
	if err != nil {
		return fmt.Errorf("error parsing %s - %s", configPath, err)
	}
	var resourceList []string
	var dataSourceList []string
	var ephemeralResourceList []string
	var actionsList []string
	specMetadata := make(map[string]properties.TerraformProviderSpecMetadata)

	for _, specPath := range c.specs {
		slog.Info("Parsing YAML spec", "spec", specPath)
		content, err := load.File(specPath)
		if err != nil {
			return fmt.Errorf("error loading %s - %s", specPath, err)
		}

		spec, err := properties.ParseSpec(content)
		if err != nil {
			return fmt.Errorf("error parsing %s - %s", specPath, err)
		}

		if err = spec.Sanity(); err != nil {
			return fmt.Errorf("%s sanity failed: %s", specPath, err)
		}

		// Extract subcategory: use YAML override if present, otherwise derive from path
		if c.commandType == properties.CommandTypeTerraform {
			if spec.TerraformProviderConfig.Subcategory == "" {
				spec.TerraformProviderConfig.Subcategory = deriveSubcategoryFromPath(specPath)
			}
		}

		if c.commandType == properties.CommandTypeTerraform {
			var singularVariant, pluralVariant bool
			// For specs that are missing resource_variants, default to generating
			// just singular variants of entry type.
			if len(spec.TerraformProviderConfig.ResourceVariants) == 0 {
				singularVariant = true
			}
			terraformResourceType := spec.TerraformProviderConfig.ResourceType
			if terraformResourceType == "" {
				terraformResourceType = properties.TerraformResourceEntry
			}

			for _, elt := range spec.TerraformProviderConfig.ResourceVariants {
				switch elt {
				case properties.TerraformResourceSingular:
					singularVariant = true
				case properties.TerraformResourcePlural:
					pluralVariant = true
				}
			}

			if singularVariant {
				var resourceTyp properties.ResourceType
				switch terraformResourceType {
				case properties.TerraformResourceEntry:
					resourceTyp = properties.ResourceEntry
				case properties.TerraformResourceUuid:
					resourceTyp = properties.ResourceUuid
				case properties.TerraformResourceCustom:
					resourceTyp = properties.ResourceCustom
				case properties.TerraformResourceConfig:
					resourceTyp = properties.ResourceConfig
				}

				terraformGenerator := generate.NewCreator(config.Output.TerraformProvider, c.templatePath, spec)
				data, err := terraformGenerator.RenderTerraformProviderFile(spec, resourceTyp)
				if err != nil {
					return fmt.Errorf("error rendering Terraform provider file for %s - %s", specPath, err)
				}

				resourceList = append(resourceList, data.Resources...)
				dataSourceList = append(dataSourceList, data.DataSources...)
				ephemeralResourceList = append(ephemeralResourceList, data.EphemeralResources...)
				actionsList = append(actionsList, data.Actions...)

				for k, v := range data.SpecMetadata {
					specMetadata[k] = v
				}

			}

			if pluralVariant {
				var resourceTyp properties.ResourceType
				switch terraformResourceType {
				case properties.TerraformResourceEntry:
					resourceTyp = properties.ResourceEntryPlural
				case properties.TerraformResourceUuid:
					resourceTyp = properties.ResourceUuidPlural
				case properties.TerraformResourceCustom:
					resourceTyp = properties.ResourceCustom
				case properties.TerraformResourceConfig:
					panic("missing implementation for config type resources")
				}

				terraformGenerator := generate.NewCreator(config.Output.TerraformProvider, c.templatePath, spec)
				data, err := terraformGenerator.RenderTerraformProviderFile(spec, resourceTyp)
				if err != nil {
					return fmt.Errorf("error rendering Terraform provider file for %s - %s", specPath, err)
				}

				resourceList = append(resourceList, data.Resources...)
				dataSourceList = append(dataSourceList, data.DataSources...)
				ephemeralResourceList = append(ephemeralResourceList, data.EphemeralResources...)
				actionsList = append(actionsList, data.Actions...)

				for k, v := range data.SpecMetadata {
					specMetadata[k] = v
				}
			}
		} else if c.commandType == properties.CommandTypeSDK && !spec.GoSdkSkip {
			generator := generate.NewCreator(config.Output.GoSdk, c.templatePath, spec)
			if err = generator.RenderTemplate(); err != nil {
				return fmt.Errorf("error rendering %s - %s", specPath, err)
			}
		}

	}

	if c.commandType == properties.CommandTypeTerraform {
		providerSpec := new(properties.Normalization)
		providerSpec.Name = "provider"

		newProviderObject := properties.NewTerraformProviderFile(providerSpec.Name)
		newProviderObject.DataSources = append(newProviderObject.DataSources, dataSourceList...)
		newProviderObject.Resources = append(newProviderObject.Resources, resourceList...)
		newProviderObject.EphemeralResources = append(newProviderObject.EphemeralResources, ephemeralResourceList...)
		newProviderObject.Actions = append(newProviderObject.Actions, actionsList...)
		newProviderObject.SpecMetadata = specMetadata

		terraformGenerator := generate.NewCreator(config.Output.TerraformProvider, c.templatePath, providerSpec)
		err = terraformGenerator.RenderTerraformProvider(newProviderObject, providerSpec, config.TerraformProviderConfig)
		if err != nil {
			return fmt.Errorf("error generating terraform code: %w", err)
		}

		// Generate tfplugindocs templates with subcategory support
		slog.Debug("Generating tfplugindocs templates", "metadataCount", len(specMetadata))
		if err = generateTfplugindocsTemplates(config.Output.TerraformProvider, specMetadata); err != nil {
			return fmt.Errorf("error generating tfplugindocs templates: %w", err)
		}

		slog.Debug("Generated Terraform resources", "resources", resourceList, "dataSources", dataSourceList)
	}

	if err = generate.CopyAssets(config, c.commandType); err != nil {
		return fmt.Errorf("error copying assets %w", err)
	}
	return nil
}
