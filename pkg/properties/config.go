package properties

import (
	"github.com/paloaltonetworks/pan-os-codegen/pkg/load"
)

type Config struct {
	Output OutputPaths `json:"output" yaml:"output"`
}

type OutputPaths struct {
	GoSdk             string `json:"go_sdk" yaml:"go_sdk"`
	TerraformProvider string `json:"terraform_provider" yaml:"terraform_provider"`
}

func ParseConfig(content []byte) (*Config, error) {
	var ans Config
	err := load.Unmarshall(content, &ans)
	return &ans, err
}
