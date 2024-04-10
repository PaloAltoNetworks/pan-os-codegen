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

var templatePaths = map[CommandType]string{
	CommandTypeSDK:       "templates/sdk",
	CommandTypeTerraform: "templates/terraform",
}

type Command struct {
	ctx          context.Context
	args         []string
	specs        []string
	config       string
	commandType  CommandType
	templatePath string
}

func NewCommand(ctx context.Context, commandType CommandType, args ...string) (*Command, error) {
	templatePath, ok := templatePaths[commandType]
	if !ok {
		return nil, fmt.Errorf("unsupported command type: %s", commandType)
	}

	return &Command{
		ctx:          ctx,
		args:         args,
		commandType:  commandType,
		templatePath: templatePath,
	}, nil
}

func (c *Command) Execute() error {
	for _, arg := range []string{"specs", "configs"} {
		if err := c.setup(arg); err != nil {
			log.Fatalf("Setup failed: %s", err)
		}
		log.Printf("Generating %s for %s type \n", c.commandType, arg)
		if len(c.args) == 0 {
			return fmt.Errorf("path to configuration file is required")
		}

		if err := c.processConfig(c.args[0]); err != nil {
			return err
		}

		log.Printf("Finish generating %s for %s type.", c.commandType, arg)
	}
	return nil
}

func (c *Command) setup(localization string) error {
	specs, err := properties.GetNormalizations(localization)
	if err != nil {
		return fmt.Errorf("error getting normalizations: %s", err)
	}
	c.specs = specs
	if localization == "configs" {
		c.templatePath = "templates/custom"
	}
	return nil
}

func (c *Command) processConfig(configPath string) error {
	config, err := c.loadAndParseConfig(configPath)
	if err != nil {
		return err
	}

	for _, specPath := range c.specs {
		if err := c.processSpec(config, specPath); err != nil {
			return err
		}
	}

	if err := generate.CopyAssets(config, string(c.commandType)); err != nil {
		return fmt.Errorf("error copying assets: %s", err)
	}

	return nil
}

func (c *Command) loadAndParseConfig(path string) (*properties.Config, error) {
	content, err := load.File(path)
	if err != nil {
		return nil, fmt.Errorf("error loading %s - %s", path, err)
	}

	config, err := properties.ParseConfig(content)
	if err != nil {
		return nil, fmt.Errorf("error parsing %s - %s", path, err)
	}

	return config, nil
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

	if err := c.validateSpec(spec, specPath); err != nil {
		return err
	}

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

	generator := generate.NewCreator(output, c.templatePath, spec, string(c.commandType))
	if err := generator.RenderTemplate(string(c.commandType)); err != nil {
		return fmt.Errorf("error rendering %s - %s", specPath, err)
	}

	return nil
}
