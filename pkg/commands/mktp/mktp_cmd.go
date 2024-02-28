package mktp

import (
	"context"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"log"
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
	log.Print("Making panos Terraform provider\n")

	//providerDataSources := make([]string, 0, 200)
	//providerResources := make([]string, 0, 100)

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

		// Output as Terraform code.
	}

	// Finalize Terraform code.
	// * output provider.go
	// * write any static files
	// * make fmt
	// * output examples to ./examples
	// * make docs
	// * docs modifications (e.g. - subcategories)

	// Done.
	log.Printf("Done\n")

	return nil
}
