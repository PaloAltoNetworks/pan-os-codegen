package properties

import (
	"github.com/paloaltonetworks/pan-os-codegen/pkg/load"
	"os"
)

type Config struct {
	Output OutputPaths `json:"output" yaml:"output"`
	Name   string      `json:"name" yaml:"name"`
}

type OutputPaths struct {
	GoSdk             string `json:"go_sdk" yaml:"go_sdk"`
	TerraformProvider string `json:"terraform_provider" yaml:"terraform_provider"`
}

func ParseConfig(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var ans Config
	err = load.File(b, &ans)
	return &ans, err
}
