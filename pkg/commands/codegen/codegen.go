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
	// TODO: add datasource here

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
			log.Printf("[DEBUG] create new provider object: %s", newProviderObject)
			terraformGenerator := generate.NewCreator(config.Output.TerraformProvider, c.templatePath, spec)
			err = terraformGenerator.RenderTerraformProvider(newProviderObject, spec)
			if err != nil {
				return fmt.Errorf("error generating Terraform provider - %s", err)
			}

			resourceList = append(resourceList, newProviderObject.Resources...)

		} else if c.commandType == CommandTypeSDK {
			generator := generate.NewCreator(config.Output.GoSdk, c.templatePath, spec)
			if err = generator.RenderTemplate(); err != nil {
				return fmt.Errorf("error rendering %s - %s", specPath, err)
			}
		}

	}

	if err = generate.CopyAssets(config); err != nil {
		return fmt.Errorf("error copying assets %s", err)
	}

	log.Println("Generation complete.")
	log.Printf("Generated resources: %s", resourceList)
	return nil
}
