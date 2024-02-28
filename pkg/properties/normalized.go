package properties

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/load"
)

type Normalization struct {
	Name                    string                      `json:"name" yaml:"name"`
	TerraformProviderSuffix string                      `json:"terraform_provider_suffix" yaml:"terraform_provider_suffix"`
	GoSdkPath               []string                    `json:"go_sdk_path" yaml:"go_sdk_path"`
	XpathSuffix             []string                    `json:"xpath_suffix" yaml:"xpath_suffix"`
	Locations               map[interface{}]interface{} `json:"locations" yaml:"locations"`
	Entry                   map[interface{}]interface{} `json:"entry" yaml:"entry"`
	Version                 string                      `json:"version" yaml:"version"`
	Spec                    map[interface{}]interface{} `json:"spec" yaml:"spec"`
}

func GetNormalizations() ([]string, error) {
	_, loc, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("couldn't get caller info")
	}

	basePath := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(loc))), "specs")

	files := make([]string, 0, 100)

	err := filepath.WalkDir(basePath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(entry.Name(), ".yaml") {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

func ParseSpec(path string) (*Normalization, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var ans Normalization
	err = load.File(b, &ans)
	return &ans, err
}

func (spec *Normalization) Sanity() error {
	if strings.Contains(spec.TerraformProviderSuffix, "panos_") {
		return errors.New("suffix for Terraform provider cannot contain `panos_`")
	}
	for _, suffix := range spec.XpathSuffix {
		if strings.Contains(suffix, "/") {
			return errors.New("XPath cannot contain /")
		}
	}
	if len(spec.Locations) < 1 {
		return errors.New("at least 1 location is required")
	}
	if len(spec.GoSdkPath) < 2 {
		return errors.New("golang SDK path should contain at least 2 elements of the path")
	}
	return nil
}

func (spec *Normalization) Validate() []error {
	return nil
}
