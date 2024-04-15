package generate

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/load"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/parsing"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/retrieve"
	xgo "github.com/paloaltonetworks/pan-os-codegen/pkg/translate/golang"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/translate/golang/terraform"
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

func (c *Cmd) parseFlags() error {
	var err error
	var pcf string
	var filedata []byte

	fp := flag.NewFlagSet("mktp", flag.ExitOnError)
	fp.StringVar(&pcf, "p", "", "The provider settings config file.")
	fp.Usage = func() {
		fmt.Fprintf(c.Stderr, `Usage: mktp <-p FILENAME> [config...]

This command creates a Terraform provider from an OpenAPI spec file,
merging the provider.go config file with each OpenAPI spec config file given.

`)
		flag.PrintDefaults()
	}
	fp.Parse(c.args)

	if pcf == "" {
		return errors.New("mktp: provider settings file is unspecified")
	} else if len(fp.Args()) == 0 {
		return errors.New("mktp: no input specified")
	}

	filedata, err = retrieve.FileContents(pcf)
	if err != nil {
		return err
	}

	psettings, err := properties.LoadTerraformProviderProperties(filedata)
	if err != nil {
		return err
	}
	c.pconfig = psettings

	c.configs = make([]*properties.OpenApiFile, 0, len(fp.Args()))
	for i := range fp.Args() {
		filedata, err = retrieve.FileContents(fp.Arg(i))
		if err != nil {
			return err
		}

		v, err := properties.LoadOpenApiFileProperties(filedata)
		if err != nil {
			return err
		}
		c.configs = append(c.configs, v)
	}

	return nil
}

func (c *Command) Execute() error {
	log.Printf("Generating %s\n", c.commandType)
	providerDataSources := make([]string, 0, 200)
	providerResources := make([]string, 0, 100)

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

		generator := generate.NewCreator(config.Output.GoSdk, c.templatePath, spec)
		if err = generator.RenderTemplate(); err != nil {
			return fmt.Errorf("error rendering %s - %s", specPath, err)
		}
	}

	if err = generate.CopyAssets(config); err != nil {
		return fmt.Errorf("error copying assets %s", err)
	}

	log.Println("Generation complete.")
	return nil
}

func (c *Cmd) Execute() error {
	// TODO(shinmog): everything
	log.Printf("ok\n")

	// Generate Terraform code for all of the namespaces.
	providerDataSources := make([]string, 0, 200)
	providerResources := make([]string, 0, 100)

	for i, spec := range c.configs {
		log.Printf("Parsing %s (%d)\n", spec.Name, i)
		parser, err := parsing.NewOpenApiSpecParser(spec.FilePath)
		if err != nil {
			return err
		}

		// Sanity checks.
		if spec.Output.Directory == "" {
			return fmt.Errorf("No output directory specified")
		}

		// Get the servers.
		servers, err := parsing.NormalizeServers(parser, spec)
		if err != nil {
			return err
		}

		// Normalize the schemas.
		log.Printf("Normalizing schemas...\n")
		namer := naming.NewNamer()
		schemas, err := parsing.NormalizeSchemas(parser, spec, namer)
		if err != nil {
			return err
		}

		schemaNames := make([]string, 0, len(schemas))
		for key := range schemas {
			schemaNames = append(schemaNames, key)
		}
		sort.Strings(schemaNames)

		// Create SDK code for each schema.
		log.Printf("Outputting schema SDK code...\n")
		for locNum, name := range schemaNames {
			v2, err := xgo.SchemaCode(name, schemas, spec)
			if err != nil {
				return fmt.Errorf("Failed %d to translate %s: %s", locNum, name, err)
			}
			path := fmt.Sprintf("%s/%s/schemas/%s", spec.Output.Directory, spec.Name, strings.Join(schemas[name].GetSdkPath(), "/"))
			if err = os.MkdirAll(path, 0755); err != nil && !os.IsExist(err) {
				return err
			}
			path = path + "/config.go"
			if err = os.WriteFile(path, []byte(v2), 0644); err != nil {
				return err
			}
		}

		// Create namespaces.
		log.Printf("Creating namespaces from paths...\n")
		namespaces, err := parsing.GatherPathFunctions(parser, spec, schemas, namer)
		if err != nil {
			return err
		}

		namespaceNames := make([]string, 0, len(namespaces))
		for key := range namespaces {
			namespaceNames = append(namespaceNames, key)
		}
		sort.Strings(namespaceNames)

		// Output service code for each namespace.
		outputPaths := make(map[string]string)
		log.Printf("Outputting service code...\n")
		for _, name := range namespaceNames {
			ns := namespaces[name]
			code, err := xgo.ServiceCode(name, namespaces, schemas, servers, spec)
			if err != nil {
				return err
			}
			path := ns.Path(spec.Output.Directory, spec.Name)
			if outputPaths[path] != "" {
				return fmt.Errorf("%q and %q both output to %q", outputPaths[path], name, path)
			}
			outputPaths[path] = name
			if err = os.MkdirAll(path, 0755); err != nil && !os.IsExist(err) {
				return err
			}
			path = path + "/service.go"
			if err = os.WriteFile(path, []byte(code), 0644); err != nil {
				return err
			}
		}

		//Format the SDK files.
		//if spec.Output.Directory != "" {
		//	log.Printf("Formatting SDK files...\n")
		//	cmd := exec.Command("go", "fmt")
		//	cmd.Dir = spec.Output.Directory
		//	if err = cmd.Run(); err != nil {
		//		return err
		//	}
		//}

		for _, loc := range namespaceNames {
			dslist, rslist, filename, code, err := terraform.Implementation(loc, namespaces, schemas, spec)
			if err != nil {
				return err
			}
			if code == "" {
				continue
			}
			providerDataSources = append(providerDataSources, dslist...)
			providerResources = append(providerResources, rslist...)
			path := fmt.Sprintf("%s/internal/provider", c.pconfig.Output.Directory)
			if err = os.MkdirAll(path, 0755); err != nil && !os.IsExist(err) {
				log.Printf("Err with %q path:%s - %s", loc, path, err)
				return err
			}
			path = path + "/" + filename
			if err = os.WriteFile(path, []byte(code), 0644); err != nil {
				log.Printf("Err in %q create %q - %s", loc, path, err)
				return err
			}
		}

		// Experimenting with encrypted values.
		//xgo.EncryptedParams(schemas)
	}

	log.Printf("Finalizing Terraform implementation...\n")
	if err := terraform.Provider(providerDataSources, providerResources, c.pconfig); err != nil {
		log.Printf(err.Error())
		return err
	}

	return nil
}
