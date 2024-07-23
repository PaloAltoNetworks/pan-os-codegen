package codegen

import (
	"context"
	"fmt"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/generate"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/load"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"log"
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
			return fmt.Errorf("error getting normalizations: %w", err)
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
		return fmt.Errorf("error loading %s: %w", configPath, err)
	}

	config, err := properties.ParseConfig(content)
	if err != nil {
		return fmt.Errorf("error parsing %s: %w", configPath, err)
	}
	var resourceList []string
	var dataSourceList []string

	for _, specPath := range c.specs {
		log.Printf("Parsing %s...\n", specPath)
		content, err := load.File(specPath)
		if err != nil {
			return fmt.Errorf("error loading %s: %w", specPath, err)
		}

		spec, err := properties.ParseSpec(content)
		if err != nil {
			return fmt.Errorf("error parsing %s: %w", specPath, err)
		}

		if err = spec.Sanity(); err != nil {
			return fmt.Errorf("%s sanity failed: %w", specPath, err)
		}

		if c.commandType == properties.CommandTypeTerraform {

			newProviderObject := properties.NewTerraformProviderFile(spec.Name)
			terraformGenerator := generate.NewCreator(config.Output.TerraformProvider, c.templatePath, spec)
			err = terraformGenerator.RenderTerraformProviderFile(newProviderObject, spec)
			if err != nil {
				return fmt.Errorf("error generating Terraform provider: %w", err)
			}

			resourceList = append(resourceList, newProviderObject.Resources...)
			dataSourceList = append(dataSourceList, newProviderObject.DataSources...)

		} else if c.commandType == properties.CommandTypeSDK {
			generator := generate.NewCreator(config.Output.GoSdk, c.templatePath, spec)
			if err = generator.RenderTemplate(); err != nil {
				return fmt.Errorf("error rendering %s: %w", specPath, err)
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
			return err
		}
	}

	if err = generate.CopyAssets(config, c.commandType); err != nil {
		return fmt.Errorf("error copying assets %w", err)
	}

	log.Println("Generation complete.")

	log.Printf("Generated resources: %s\n Generated dataSources: %s", resourceList, dataSourceList)
	return nil
}
