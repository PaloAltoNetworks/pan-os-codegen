package properties

import (
	"errors"
	"fmt"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/content"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"
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

type NameVariant struct {
	Underscore string
	CamelCase  string
}

type Location struct {
	Name        *NameVariant
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
	Name        *NameVariant
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
	Name        *NameVariant
	Description string              `json:"description" yaml:"description"`
	Type        string              `json:"type" yaml:"type"`
	Required    bool                `json:"required" yaml:"required"`
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

func ParseSpec(input []byte) (*Normalization, error) {
	var spec Normalization

	err := content.Unmarshal(input, &spec)

	err = spec.AddNameVariantsForLocation()
	err = spec.AddNameVariantsForParams()
	err = spec.AddDefaultTypesForParams()

	return &spec, err
}

func (spec *Normalization) AddNameVariantsForLocation() error {
	for key, location := range spec.Locations {
		location.Name = &NameVariant{
			Underscore: key,
			CamelCase:  naming.CamelCase("", key, "", true),
		}

		for subkey, variable := range location.Vars {
			variable.Name = &NameVariant{
				Underscore: subkey,
				CamelCase:  naming.CamelCase("", subkey, "", true),
			}
		}
	}

	return nil
}

func AddNameVariantsForParams(name string, param *SpecParam) error {
	param.Name = &NameVariant{
		Underscore: name,
		CamelCase:  naming.CamelCase("", name, "", true),
	}
	if param.Spec != nil {
		for key, childParam := range param.Spec.Params {
			if err := AddNameVariantsForParams(key, childParam); err != nil {
				return err
			}
		}
		for key, childParam := range param.Spec.OneOf {
			if err := AddNameVariantsForParams(key, childParam); err != nil {
				return err
			}
		}
	}
	return nil
}

func (spec *Normalization) AddNameVariantsForParams() error {
	if spec.Spec != nil {
		for key, param := range spec.Spec.Params {
			if err := AddNameVariantsForParams(key, param); err != nil {
				return err
			}
		}
		for key, param := range spec.Spec.OneOf {
			if err := AddNameVariantsForParams(key, param); err != nil {
				return err
			}
		}
	}
	return nil
}

// AddDefaultTypesForParams ensures all SpecParams within Spec have a default type if not specified.
func (spec *Normalization) AddDefaultTypesForParams() error {
	if spec.Spec == nil {
		return nil
	}

	setDefaultParamTypeForMap(spec.Spec.Params)
	setDefaultParamTypeForMap(spec.Spec.OneOf)

	return nil
}

// setDefaultParamTypeForMap iterates over a map of SpecParam pointers, setting their Type to "string" if not specified.
func setDefaultParamTypeForMap(params map[string]*SpecParam) {
	for _, param := range params {
		if param.Type == "" {
			param.Type = "string"
		}
	}
}

func (spec *Normalization) Sanity() error {
	if spec.Name == "" {
		return errors.New("name is required")
	}
	if spec.Locations == nil {
		return errors.New("at least 1 location is required")
	}
	if spec.GoSdkPath == nil {
		return errors.New("golang SDK path is required")
	}

	return nil
}

func (spec *Normalization) Validate() []error {
	var checks []error

	if strings.Contains(spec.TerraformProviderSuffix, "panos_") {
		checks = append(checks, errors.New("suffix for Terraform provider cannot contain `panos_`"))
	}
	for _, suffix := range spec.XpathSuffix {
		if strings.Contains(suffix, "/") {
			checks = append(checks, errors.New("XPath cannot contain /"))
		}
	}
	if len(spec.Locations) < 1 {
		checks = append(checks, errors.New("at least 1 location is required"))
	}
	if len(spec.GoSdkPath) < 2 {
		checks = append(checks, errors.New("golang SDK path should contain at least 2 elements of the path"))
	}

	return checks
}
