package properties

import "github.com/paloaltonetworks/pan-os-codegen/pkg/content"

type Config struct {
	Output OutputPaths `json:"output" yaml:"output"`
}

type OutputPaths struct {
	GoSdk             string `json:"go_sdk" yaml:"go_sdk"`
	TerraformProvider string `json:"terraform_provider" yaml:"terraform_provider"`
}

func ParseConfig(input []byte) (*Config, error) {
	var ans Config
	err := content.Unmarshal(input, &ans)
	return &ans, err
}
