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
	Name                    string               `json:"name" yaml:"name"`
	TerraformProviderSuffix string               `json:"terraform_provider_suffix" yaml:"terraform_provider_suffix"`
	GoSdkPath               []string             `json:"go_sdk_path" yaml:"go_sdk_path"`
	XpathSuffix             []string             `json:"xpath_suffix" yaml:"xpath_suffix"`
	Locations               map[string]*Location `json:"locations" yaml:"locations"`
	Entry                   *Entry               `json:"entry" yaml:"entry"`
	Version                 string               `json:"version" yaml:"version"`
	Spec                    *Spec                `json:"spec" yaml:"spec"`
}

type Location struct {
	Description string                  `json:"description" yaml:"description"`
	Device      *LocationDevice         `json:"device" yaml:"device"`
	Xpath       []string                `json:"xpath" yaml:"xpath"`
	ReadOnly    bool                    `json:"read_only" yaml:"read_only"`
	Vars        map[string]*LocationVar `json:"vars" yaml:"vars"`
}

type LocationDevice struct {
	Panorama bool `json:"panorama" yaml:"panorama"`
	Ngfw     bool `json:"ngfw" yaml:"ngfw"`
}

type LocationVar struct {
	Description string                 `json:"description" yaml:"description"`
	Required    bool                   `json:"required" yaml:"required"`
	Validation  *LocationVarValidation `json:"validation" yaml:"validation"`
}

type LocationVarValidation struct {
	NotValues map[string]string `json:"not_values" yaml:"not_values"`
}

type Entry struct {
	Name *EntryName `json:"name" yaml:"name"`
}

type EntryName struct {
	Description string           `json:"description" yaml:"description"`
	Length      *EntryNameLength `json:"length" yaml:"length"`
}

type EntryNameLength struct {
	Min *int64 `json:"min" yaml:"min"`
	Max *int64 `json:"max" yaml:"max"`
}

type Spec struct {
	Params map[string]*SpecParam `json:"params" yaml:"params"`
	OneOf  map[string]*SpecParam `json:"one_of" yaml:"one_of,omitempty"`
}

type SpecParam struct {
	Description string              `json:"description" yaml:"description"`
	Type        string              `json:"type" yaml:"type"`
	Length      *SpecParamLength    `json:"length" yaml:"length,omitempty"`
	Count       *SpecParamCount     `json:"count" yaml:"count,omitempty"`
	Items       *SpecParamItems     `json:"items" yaml:"items,omitempty"`
	Regex       string              `json:"regex" yaml:"regex,omitempty"`
	Profiles    []*SpecParamProfile `json:"profiles" yaml:"profiles"`
	Spec        *Spec               `json:"spec" yaml:"spec"`
}

type SpecParamLength struct {
	Min *int64 `json:"min" yaml:"min"`
	Max *int64 `json:"max" yaml:"max"`
}

type SpecParamCount struct {
	Min *int64 `json:"min" yaml:"min"`
	Max *int64 `json:"max" yaml:"max"`
}

type SpecParamItems struct {
	Type   string                `json:"type" yaml:"type"`
	Length *SpecParamItemsLength `json:"length" yaml:"length"`
}

type SpecParamItemsLength struct {
	Min *int64 `json:"min" yaml:"min"`
	Max *int64 `json:"max" yaml:"max"`
}

type SpecParamProfile struct {
	Xpath []string `json:"xpath" yaml:"xpath"`
	Type  string   `json:"type" yaml:"type,omitempty"`
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
