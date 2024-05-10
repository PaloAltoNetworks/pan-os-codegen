package properties

import "github.com/paloaltonetworks/pan-os-codegen/pkg/content"

type Config struct {
	Output                  OutputPaths       `json:"output" yaml:"output"`
	Assets                  map[string]*Asset `json:"assets" yaml:"assets"`
	TerraformProviderConfig TerraformProvider `json:"terraform_provider_config" yaml:"terraform_provider_config"`
}

type OutputPaths struct {
	GoSdk             string `json:"go_sdk" yaml:"go_sdk"`
	TerraformProvider string `json:"terraform_provider" yaml:"terraform_provider"`
}

type Asset struct {
	Source      string `json:"source" yaml:"source"`
	Target      Target `json:"target" yaml:"target"`
	Destination string `json:"destination" yaml:"destination"`
}

type Target struct {
	GoSdk             bool `json:"go_sdk" yaml:"go_sdk"`
	TerraformProvider bool `json:"terraform_provider" yaml:"terraform_provider"`
}

// ParseConfig initialize Config instance using input data from YAML file.
func ParseConfig(input []byte) (*Config, error) {
	var ans Config
	err := content.Unmarshal(input, &ans)
	return &ans, err
}
