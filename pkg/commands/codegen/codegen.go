package codegen

import (
	"context"
	"fmt"
	"log"

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
	log.Printf("Generating %s\n", c.commandType)

	if len(c.args) == 0 {
		return fmt.Errorf("path to configuration file is required")
	}
	configPath := c.args[0]

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

	for _, specPath := range c.specs {
		log.Printf("Parsing %s...\n", specPath)
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
				case properties.TerraformResourceConfig:
					panic("missing implementation for config type resources")
				}

				terraformGenerator := generate.NewCreator(config.Output.TerraformProvider, c.templatePath, spec)
				dataSources, resources, err := terraformGenerator.RenderTerraformProviderFile(spec, resourceTyp)
				if err != nil {
					return fmt.Errorf("error rendering Terraform provider file for %s - %s", specPath, err)
				}

				resourceList = append(resourceList, resources...)
				dataSourceList = append(dataSourceList, dataSources...)
			}

			if pluralVariant {
				var resourceTyp properties.ResourceType
				switch terraformResourceType {
				case properties.TerraformResourceEntry:
					resourceTyp = properties.ResourceEntryPlural
				case properties.TerraformResourceUuid:
					resourceTyp = properties.ResourceUuidPlural
				case properties.TerraformResourceConfig:
					panic("missing implementation for config type resources")
				}

				terraformGenerator := generate.NewCreator(config.Output.TerraformProvider, c.templatePath, spec)
				dataSources, resources, err := terraformGenerator.RenderTerraformProviderFile(spec, resourceTyp)
				if err != nil {
					return fmt.Errorf("error rendering Terraform provider file for %s - %s", specPath, err)
				}

				resourceList = append(resourceList, resources...)
				dataSourceList = append(dataSourceList, dataSources...)
			}
		} else if c.commandType == properties.CommandTypeSDK {
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

		terraformGenerator := generate.NewCreator(config.Output.TerraformProvider, c.templatePath, providerSpec)
		err = terraformGenerator.RenderTerraformProvider(newProviderObject, providerSpec, config.TerraformProviderConfig)
		if err != nil {
			return fmt.Errorf("error generating terraform code: %w", err)
		}
		log.Printf("Generated Terraform resources: %s\n Generated dataSources: %s", resourceList, dataSourceList)
	}

	if err = generate.CopyAssets(config, c.commandType); err != nil {
		return fmt.Errorf("error copying assets %w", err)
	}
	return nil
}
