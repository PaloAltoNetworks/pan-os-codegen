package codegen

import (
	"context"
	"fmt"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/generate"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/load"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"log"
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

// Execute the command and generate the outputs
func (c *Command) Execute() error {
	log.Printf("Generating %s\n", c.commandType)
	if len(c.args) == 0 {
		return fmt.Errorf("path to configuration file is required")
	}
	configPath := c.args[0]

	if err := c.processConfig(configPath); err != nil {
		return err
	}

	log.Printf("Finish generating %s.", c.commandType)
	return nil
}

// processConfig process the configuration and spec files
func (c *Command) processConfig(configPath string) error {
	content, err := load.File(configPath)
	if err != nil {
		return fmt.Errorf("error loading %s - %s", configPath, err)
	}

	config, err := properties.ParseConfig(content)
	if err != nil {
		return fmt.Errorf("error parsing %s - %s", configPath, err)
	}

	for _, specPath := range c.specs {
		log.Printf("[DEBUG] print specPath -> %s \n", specPath)
		if err := c.processSpec(config, specPath); err != nil {
			return err
		}
	}

	if err = generate.CopyAssets(config); err != nil {
		return fmt.Errorf("error copying assets %s", err)
	}

	return nil
}

// processSpec process individual spec
func (c *Command) processSpec(config *properties.Config, specPath string) error {
	log.Printf("Parsing %s...\n", specPath)

	content, err := load.File(specPath)
	if err != nil {
		return fmt.Errorf("error loading %s - %s", specPath, err)
	}

	spec, err := properties.ParseSpec(content)
	if err != nil {
		return fmt.Errorf("error parsing %s - %s", specPath, err)
	}

	// validate the spec
	if err := c.validateSpec(spec, specPath); err != nil {
		return err
	}

	// Generate the output
	return c.generateOutput(config, spec, specPath)
}

// validateSpec validate the spec file with Sanity function
func (c *Command) validateSpec(spec *properties.Normalization, specPath string) error {
	if err := spec.Sanity(string(c.commandType)); err != nil {
		return fmt.Errorf("%s sanity failed: %s", specPath, err)
	}

	return nil
}

// generateOutput generate the output with the spec and config
func (c *Command) generateOutput(config *properties.Config, spec *properties.Normalization, specPath string) error {
	var output string

	switch c.commandType {
	case CommandTypeSDK:
		output = config.Output.GoSdk
	case CommandTypeTerraform:
		output = config.Output.TerraformProvider
	}

	log.Printf("[DEBUG] print output -> %s \n", output)

	generator := generate.NewCreator(output, c.templatePath, spec)
	if err := generator.RenderTemplate(string(c.commandType)); err != nil {
		return fmt.Errorf("error rendering %s - %s", specPath, err)
	}

	return nil
}
