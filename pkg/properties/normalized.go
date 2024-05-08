package properties

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/content"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
)

type Normalization struct {
	Name                    string                  `json:"name" yaml:"name"`
	TerraformProviderConfig TerraformProviderConfig `json:"terraform_provider_config" yaml:"terraform_provider_config"`
	GoSdkPath               []string                `json:"go_sdk_path" yaml:"go_sdk_path"`
	XpathSuffix             []string                `json:"xpath_suffix" yaml:"xpath_suffix"`
	Locations               map[string]*Location    `json:"locations" yaml:"locations"`
	Entry                   *Entry                  `json:"entry" yaml:"entry"`
	Imports                 map[string]*Import      `json:"imports" yaml:"imports"`
	Version                 string                  `json:"version" yaml:"version"`
	Spec                    *Spec                   `json:"spec" yaml:"spec"`
	Const                   map[string]*Const       `json:"const" yaml:"const"`
}

type TerraformProviderConfig struct {
	SkipResource          bool   `json:"skip_resource" yaml:"skip_resource"`
	SkipDatasource        bool   `json:"skip_datasource" yaml:"skip_datasource"`
	SkipDatasourceListing bool   `json:"skip_datasource_listing" yaml:"skip_datasource_listing"`
	Suffix                string `json:"suffix" yaml:"suffix"`
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

type Import struct {
	Name  *NameVariant
	Xpath []string              `json:"xpath" yaml:"xpath"`
	Vars  map[string]*ImportVar `json:"vars" yaml:"vars"`
}

type ImportVar struct {
	Name        *NameVariant
	Description string `json:"description" yaml:"description"`
	Default     string `json:"default" yaml:"default"`
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

type Const struct {
	Name   *NameVariant
	Values map[string]*ConstValue `json:"values" yaml:"values"`
}

type ConstValue struct {
	Name  *NameVariant
	Value string `json:"value" yaml:"value"`
}

type SpecParam struct {
	Name        *NameVariant
	Description string              `json:"description" yaml:"description"`
	Type        string              `json:"type" yaml:"type"`
	Default     string              `json:"default" yaml:"default"`
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
	Ref    []*string             `json:"ref" yaml:"ref"`
}

type SpecParamItemsLength struct {
	Min *int64 `json:"min" yaml:"min"`
	Max *int64 `json:"max" yaml:"max"`
}

type SpecParamProfile struct {
	Xpath       []string `json:"xpath" yaml:"xpath"`
	Type        string   `json:"type" yaml:"type,omitempty"`
	NotPresent  bool     `json:"not_present" yaml:"not_present"`
	FromVersion string   `json:"from_version" yaml:"from_version"`
}

// GetNormalizations get list of all specs (normalizations).
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

// ParseSpec parse single spec (unmarshal file), add name variants for locations and params, add default types for params.
func ParseSpec(input []byte) (*Normalization, error) {
	var spec Normalization

	err := content.Unmarshal(input, &spec)
	if err != nil {
		return nil, err
	}

	err = spec.AddNameVariantsForLocation()
	if err != nil {
		return nil, err
	}

	err = spec.AddNameVariantsForParams()
	if err != nil {
		return nil, err
	}

	err = spec.AddDefaultTypesForParams()
	if err != nil {
		return nil, err
	}

	err = spec.AddNameVariantsForTypes()
	if err != nil {
		return nil, err
	}

	return &spec, err
}

// AddNameVariantsForLocation add name variants for location (under_score and CamelCase).
func (spec *Normalization) AddNameVariantsForLocation() error {
	for key, location := range spec.Locations {
		location.Name = &NameVariant{
			Underscore: naming.Underscore("", key, ""),
			CamelCase:  naming.CamelCase("", key, "", true),
		}

		for subkey, variable := range location.Vars {
			variable.Name = &NameVariant{
				Underscore: naming.Underscore("", subkey, ""),
				CamelCase:  naming.CamelCase("", subkey, "", true),
			}
		}
	}

	return nil
}

// AddNameVariantsForParams recursively add name variants for params for nested specs.
func AddNameVariantsForParams(name string, param *SpecParam) error {
	param.Name = &NameVariant{
		Underscore: naming.Underscore("", name, ""),
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

// AddNameVariantsForParams add name variants for params (under_score and CamelCase).
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

// AddNameVariantsForTypes add name variants for types (under_score and CamelCase).
func (spec *Normalization) AddNameVariantsForTypes() error {
	if spec.Const != nil {
		for nameType, customType := range spec.Const {
			customType.Name = &NameVariant{
				Underscore: naming.Underscore("", nameType, ""),
				CamelCase:  naming.CamelCase("", nameType, "", true),
			}
			for nameValue, customValue := range customType.Values {
				customValue.Name = &NameVariant{
					Underscore: naming.Underscore("", nameValue, ""),
					CamelCase:  naming.CamelCase("", nameValue, "", true),
				}
			}
		}
	}
	return nil
}

// addDefaultTypesForParams recursively add default types for params for nested specs.
func addDefaultTypesForParams(params map[string]*SpecParam) error {
	for _, param := range params {
		if param.Type == "" && param.Spec == nil {
			param.Type = "string"
		}

		if param.Spec != nil {
			if err := addDefaultTypesForParams(param.Spec.Params); err != nil {
				return err
			}
			if err := addDefaultTypesForParams(param.Spec.OneOf); err != nil {
				return err
			}
		}
	}

	return nil
}

// AddDefaultTypesForParams ensures all params within Spec have a default type if not specified.
func (spec *Normalization) AddDefaultTypesForParams() error {
	if spec.Spec != nil {
		if err := addDefaultTypesForParams(spec.Spec.Params); err != nil {
			return err
		}
		if err := addDefaultTypesForParams(spec.Spec.OneOf); err != nil {
			return err
		}
		return nil
	} else {
		return nil
	}
}

// Sanity basic checks for specification (normalization) e.g. check if at least 1 location is defined.
func (spec *Normalization) Sanity() error {
	if spec.Name == "" {
		return fmt.Errorf("name is required")
	}
	if spec.Locations == nil {
		return fmt.Errorf("at least 1 location is required")
	}
	if spec.GoSdkPath == nil {
		return fmt.Errorf("golang SDK path is required")
	}

	return nil
}

// Validate validations for specification (normalization) e.g. check if XPath contain /.
func (spec *Normalization) Validate() []error {
	var checks []error

	if strings.Contains(spec.TerraformProviderConfig.Suffix, "panos_") {
		checks = append(checks, fmt.Errorf("suffix for Terraform provider cannot contain `panos_`"))
	}
	for _, suffix := range spec.XpathSuffix {
		if strings.Contains(suffix, "/") {
			checks = append(checks, fmt.Errorf("XPath cannot contain /"))
		}
	}
	if len(spec.Locations) < 1 {
		checks = append(checks, fmt.Errorf("at least 1 location is required"))
	}
	if len(spec.GoSdkPath) < 2 {
		checks = append(checks, fmt.Errorf("golang SDK path should contain at least 2 elements of the path"))
	}

	return checks
}

// SupportedVersions provides list of all supported versions in format MAJOR.MINOR.PATCH
func (spec *Normalization) SupportedVersions() []string {
	if spec.Spec != nil {
		versions := supportedVersions(spec.Spec.Params, []string{""})
		versions = supportedVersions(spec.Spec.OneOf, versions)
		return versions
	}
	return nil
}

func supportedVersions(params map[string]*SpecParam, versions []string) []string {
	for _, param := range params {
		for _, profile := range param.Profiles {
			if profile.FromVersion != "" {
				if notExist := listContains(versions, profile.FromVersion); notExist {
					versions = append(versions, profile.FromVersion)
				}
			}
		}
		if param.Spec != nil {
			versions = supportedVersions(param.Spec.Params, versions)
			versions = supportedVersions(param.Spec.OneOf, versions)
		}
	}
	return versions
}

func listContains(versions []string, checkedVersion string) bool {
	for _, version := range versions {
		if version == checkedVersion {
			return false
		}
	}
	return true
}
