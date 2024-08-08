package imports

import "github.com/paloaltonetworks/pan-os-codegen/pkg/schema/xpath"

type Import struct {
	Name          string            `yaml:"name"`
	Xpath         xpathschema.Xpath `yaml:"xpath"`
	OnlyForParams []string          `yaml:"only_for_params"`
}
