package mktp

import (
	"context"
	"fmt"
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

	fmt.Fprintf(c.Stdout, "Reading configuration file: %s...\n", c.args[0])
	config, err := properties.ParseConfig(c.args[0])
	if err != nil {
		return fmt.Errorf("error parsing %s - %s", c.args[0], err)
	}
	fmt.Fprintf(c.Stdout, "Output directory for Go SDK: %s\n", config.Output.GoSdk)
	fmt.Fprintf(c.Stdout, "Output directory for Terraform provider: %s\n", config.Output.TerraformProvider)

	for _, configPath := range c.specs {
		fmt.Fprintf(c.Stdout, "Parsing %s...\n", configPath)
		spec, err := properties.ParseSpec(configPath)
		if err != nil {
			return fmt.Errorf("error parsing %s - %s", configPath, err)
		}

		// Sanity check.
		if err = spec.Sanity(); err != nil {
			return fmt.Errorf("%s sanity failed: %s", configPath, err)
		}

		// Output normalization as pango code.
		if err = os.MkdirAll(config.Output.GoSdk, 0755); err != nil && !os.IsExist(err) {
			return err
		}

		// Output as Terraform code.
		if err = os.MkdirAll(config.Output.TerraformProvider, 0755); err != nil && !os.IsExist(err) {
			return err
		}
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
