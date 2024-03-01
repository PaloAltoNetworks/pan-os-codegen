package creator

import (
	"github.com/paloaltonetworks/pan-os-codegen/pkg/load"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"os"
)

func PrepareOutputDirs(configPath string) (bool, error) {
	content, err := load.File(configPath)
	if err != nil {
		return false, err
	}

	config, err := properties.ParseConfig(content)
	if err != nil {
		return false, err
	}

	if err = os.MkdirAll(config.Output.GoSdk, 0755); err != nil && !os.IsExist(err) {
		return false, err
	}

	if err = os.MkdirAll(config.Output.TerraformProvider, 0755); err != nil && !os.IsExist(err) {
		return false, err
	}
	return true, nil
}
