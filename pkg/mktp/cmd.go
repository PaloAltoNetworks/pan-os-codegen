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
    Stdin io.Reader
    Stdout io.Writer
    Stderr io.Writer

    ctx context.Context
    args []string
    configs []string
}

func Command(ctx context.Context, args ...string) *Cmd {
    return &Cmd{
        ctx: ctx,
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

    if c.configs == nil {
        c.configs, err = properties.GetNormalizations()
    }

    return err
}

func (c *Cmd) Execute() error {
    // TODO(shinmog): everything
    fmt.Fprintf(c.Stdout, "Making pango / panos Terraform provider\n")

    //providerDataSources := make([]string, 0, 200)
    //providerResources := make([]string, 0, 100)

    for _, configPath := range c.configs {
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
