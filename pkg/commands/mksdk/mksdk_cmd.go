package mksdk

import (
	"context"
	"log"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

type Cmd struct {
	ctx     context.Context
	args    []string
	configs []string
}

func Command(ctx context.Context, args ...string) *Cmd {
	return &Cmd{
		ctx:  ctx,
		args: args,
	}
}

func (c *Cmd) Setup() error {
	var err error

	if c.configs == nil {
		c.configs, err = properties.GetNormalizations()
	}

	return err
}

func (c *Cmd) Execute() error {
	log.Print("Making Panos SDK - Pango\n")

	for _, configPath := range c.configs {
		log.Printf("Parsing %s...\n", configPath)
		spec, err := properties.ParseSpec(configPath)
		if err != nil {
			log.Fatalf("error parsing %s - %s", configPath, err)
		}

		// Sanity check.
		if err = spec.Sanity(); err != nil {
			log.Fatalf("%s sanity failed: %s", configPath, err)
		}

		// Output normalization as pango code.
	}

	// Finalize pango code:
	// * make fmt

	// Done.
	log.Printf("Done\n")

	return nil
}
