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
	config       string
	commandType  CommandType
	templatePath string
}

func NewCommand(ctx context.Context, commandType CommandType, args ...string) (*Command, error) {
	var templatePath string
	switch commandType {
	case CommandTypeSDK:
		templatePath = "templates/sdk"
	case CommandTypeTerraform:
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

		if c.commandType == CommandTypeTerraform {

			newProviderObject := properties.NewTerraformProviderFile(spec.Name)
			terraformGenerator := generate.NewCreator(config.Output.TerraformProvider, c.templatePath, spec)
			err = terraformGenerator.RenderTerraformProviderFile(newProviderObject, spec)
			if err != nil {
				return fmt.Errorf("error generating Terraform provider - %s", err)
			}

			resourceList = append(resourceList, newProviderObject.Resources...)
			dataSourceList = append(dataSourceList, newProviderObject.DataSources...)

		} else if c.commandType == CommandTypeSDK {
			generator := generate.NewCreator(config.Output.GoSdk, c.templatePath, spec)
			if err = generator.RenderTemplate(); err != nil {
				return fmt.Errorf("error rendering %s - %s", specPath, err)
			}
		}

	}

	if c.commandType == CommandTypeTerraform {
		providerSpec := new(properties.Normalization)
		providerSpec.Name = "provider"

		newProviderObject := properties.NewTerraformProviderFile(providerSpec.Name)
		newProviderObject.DataSources = append(newProviderObject.DataSources, dataSourceList...)
		newProviderObject.Resources = append(newProviderObject.Resources, resourceList...)

		terraformGenerator := generate.NewCreator(config.Output.TerraformProvider, c.templatePath, providerSpec)
		err = terraformGenerator.RenderTerraformProvider(newProviderObject, providerSpec, config.TerraformProviderConfig)
	}

	if err = generate.CopyAssets(config); err != nil {
		return fmt.Errorf("error copying assets %s", err)
	}

	log.Println("Generation complete.")

	log.Printf("Generated resources: %s\n Generated dataSources: %s", resourceList, dataSourceList)
	return nil
}
