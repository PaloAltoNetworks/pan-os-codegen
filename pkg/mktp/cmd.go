package mktp

import (
	"context"
	"fmt"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/creator"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/load"
	"io"
	"os"
	_ "os/exec"
	_ "path/filepath"
	_ "sort"
	_ "strings"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

type Cmd struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	ctx    context.Context
	args   []string
	specs  []string
	config string
}

func Command(ctx context.Context, args ...string) *Cmd {
	return &Cmd{
		ctx:  ctx,
		args: args,
	}
}

func (c *Cmd) Setup() error {
	var err error

	if c.Stdin == nil {
		c.Stdin = os.Stdin
	}

	if c.Stdout == nil {
		c.Stdout = os.Stdout
	}

	if c.Stderr == nil {
		c.Stderr = os.Stderr
	}

	if c.specs == nil {
		c.specs, err = properties.GetNormalizations()
	}

	return err
}

func (c *Cmd) Execute() error {
	// TODO(shinmog): everything
	fmt.Fprintf(c.Stdout, "Making pango / panos Terraform provider\n")

	//providerDataSources := make([]string, 0, 200)
	//providerResources := make([]string, 0, 100)

	// Check if path to configuration file is passed as argument
	if len(c.args) == 0 {
		return fmt.Errorf("path to configuration file is required")
	}
	configPath := c.args[0]

	// Load configuration file
	content, err := load.File(configPath)
	if err != nil {
		return fmt.Errorf("error loading %s - %s", configPath, err)
	}

	// Parse configuration file
	config, err := properties.ParseConfig(content)
	if err != nil {
		return fmt.Errorf("error parsing %s - %s", configPath, err)
	}

	// Create output directories
	fmt.Fprintf(c.Stdout, "Creating output directories defined in %s... \n", configPath)
	_, err = creator.CreateOutputDirs(config)
	if err != nil {
		return fmt.Errorf("error config %s - %s", configPath, err)
	}

	for _, specPath := range c.specs {
		fmt.Fprintf(c.Stdout, "Parsing %s...\n", specPath)

		// Load YAML file
		content, err := load.File(specPath)
		if err != nil {
			return fmt.Errorf("error loading %s - %s", specPath, err)
		}

		// Parse content
		spec, err := properties.ParseSpec(content)
		if err != nil {
			return fmt.Errorf("error parsing %s - %s", specPath, err)
		}

		// Sanity check.
		if err = spec.Sanity(); err != nil {
			return fmt.Errorf("%s sanity failed: %s", specPath, err)
		}

		// Prepare files
		_, err = creator.RenderTemplate(config.Output.GoSdk, spec)

		// Output normalization as pango code.

		// Output as Terraform code.
	}

	// Finalize pango code:
	// * make fmt

	// Finalize Terraform code.
	// * output provider.go
	// * write any static files
	// * make fmt
	// * output examples to ./examples
	// * make docs
	// * docs modifications (e.g. - subcategories)

	// Done.
	fmt.Fprintf(c.Stdout, "Done\n")

	return nil
}
