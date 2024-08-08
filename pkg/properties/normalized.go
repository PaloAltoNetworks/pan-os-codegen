package properties

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/content"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/object"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/parameter"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/validator"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/xpath"
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
	Underscore     string
	CamelCase      string
	LowerCamelCase string
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
	Default     string                 `json:"default" yaml:"default"`
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
	Name          *NameVariant
	Xpath         []string              `json:"xpath" yaml:"xpath"`
	Vars          map[string]*ImportVar `json:"vars" yaml:"vars"`
	OnlyForParams []string              `json:"only_for_params" yaml:"only_for_params"`
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
	Sensitive   bool                `json:"sensitive" yaml:"sensitive"`
	Length      *SpecParamLength    `json:"length" yaml:"length,omitempty"`
	Count       *SpecParamCount     `json:"count" yaml:"count,omitempty"`
	Hashing     *SpecParamHashing   `json:"hashing" yaml:"hashing,omitempty"`
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

type SpecParamHashing struct {
	Type string `json:"type" yaml:"type"`
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

func hasChildEncryptedResources(param *SpecParam) bool {
	if param.Hashing != nil {
		return true
	}

	if param.Spec == nil {
		return false
	}

	for _, elt := range param.Spec.Params {
		if hasChildEncryptedResources(elt) {
			return true
		}
	}

	for _, elt := range param.Spec.OneOf {
		if hasChildEncryptedResources(elt) {
			return true
		}
	}

	return false
}

func (o *SpecParam) HasEntryName() bool {
	if o.Type != "list" {
		return false
	}

	return o.Items.Type == "entry"
}

func (o *SpecParam) HasEncryptedResources() bool {
	if o.Hashing != nil {
		return true
	}

	if o.Spec == nil {
		return false
	}

	for _, elt := range o.Spec.Params {
		if hasChildEncryptedResources(elt) {
			return true
		}
	}

	for _, elt := range o.Spec.OneOf {
		if hasChildEncryptedResources(elt) {
			return true
		}
	}

	return false
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

func schemaParameterToSpecParameter(schemaSpec *parameter.Parameter) (*SpecParam, error) {
	var specType string
	if schemaSpec.Type == "object" {
		specType = ""
	} else if schemaSpec.Type == "enum" {
		specType = "string"
	} else {
		specType = schemaSpec.Type
	}

	var defaultVal string

	var innerSpec *Spec
	var itemsSpec SpecParamItems

	generateInnerSpec := func(spec *parameter.StructSpec) (*Spec, error) {
		params := make(map[string]*SpecParam)
		oneofs := make(map[string]*SpecParam)

		for _, elt := range spec.Parameters {
			param, err := schemaParameterToSpecParameter(elt)
			if err != nil {
				return nil, err
			}
			params[elt.Name] = param
		}

		for _, elt := range spec.Variants {
			param, err := schemaParameterToSpecParameter(elt)
			if err != nil {
				return nil, err
			}
			oneofs[elt.Name] = param
		}

		return &Spec{
			Params: params,
			OneOf:  oneofs,
		}, nil
	}

	switch spec := schemaSpec.Spec.(type) {
	case *parameter.StructSpec:
		var err error
		innerSpec, err = generateInnerSpec(spec)
		if err != nil {
			return nil, err
		}

	case *parameter.ListSpec:
		if spec.Items.Type == "object" {
			itemsSpec.Type = "entry"
			var err error
			innerSpec, err = generateInnerSpec(&spec.Items.Spec)
			if err != nil {
				return nil, err
			}
		} else {
			itemsSpec.Type = spec.Items.Type
		}
		for _, v := range schemaSpec.Validators {
			switch spec := v.Spec.(type) {
			case *validator.CountSpec:
				min := int64(spec.Min)
				max := int64(spec.Max)
				itemsSpec.Length = &SpecParamItemsLength{
					Min: &min,
					Max: &max,
				}
			}
		}
	case *parameter.EnumSpec:
		defaultVal = spec.Default
	case *parameter.NilSpec:
		specType = "string"
	case *parameter.SimpleSpec:
		if typed, ok := spec.Default.(string); ok {
			defaultVal = typed
		}
	}

	var profiles []*SpecParamProfile
	for _, profile := range schemaSpec.Profiles {
		var notPresent bool
		var version string
		if profile.MaximumVersion != nil {
			notPresent = true
			version = profile.MaximumVersion.String()
		} else if profile.MinimumVersion != nil {
			version = profile.MinimumVersion.String()
		}
		profiles = append(profiles, &SpecParamProfile{
			Xpath:       profile.Xpath,
			Type:        profile.Type,
			NotPresent:  notPresent,
			FromVersion: version,
		})
	}

	specParameter := &SpecParam{
		Description: schemaSpec.Description,
		Type:        specType,
		Default:     defaultVal,
		Required:    schemaSpec.Required,
		Profiles:    profiles,
		Spec:        innerSpec,
	}

	for _, v := range schemaSpec.Validators {
		switch spec := v.Spec.(type) {
		case *validator.RegexpSpec:
			specParameter.Regex = spec.Expr
		case *validator.StringLengthSpec:
			min := int64(spec.Min)
			max := int64(spec.Max)
			specParameter.Length = &SpecParamLength{
				Min: &min,
				Max: &max,
			}
		case *validator.CountSpec:
			min := int64(spec.Min)
			max := int64(spec.Max)
			specParameter.Count = &SpecParamCount{
				Min: &min,
				Max: &max,
			}
		}
	}

	if schemaSpec.Type == "list" {
		specParameter.Items = &itemsSpec
	}

	return specParameter, nil
}

func generateXpathVariables(variables []xpathschema.Variable) map[string]*LocationVar {
	xpathVars := make(map[string]*LocationVar)
	for _, variable := range variables {
		entry := &LocationVar{
			Description: variable.Description,
			Default:     variable.Default,
			Required:    variable.Required,
			Validation:  nil,
		}

		for _, v := range variable.Validators {
			switch spec := v.Spec.(type) {
			case *validator.NotValuesSpec:
				notValues := make(map[string]string)
				for _, value := range spec.Values {
					notValues[value.Value] = value.Error

				}
				entry.Validation = &LocationVarValidation{
					NotValues: notValues,
				}

			}
		}

		xpathVars[variable.Name] = entry
	}

	return xpathVars
}

func generateImportVariables(variables []xpathschema.Variable) map[string]*ImportVar {
	importVars := make(map[string]*ImportVar)
	for _, variable := range variables {
		entry := &ImportVar{
			Description: variable.Description,
			Default:     variable.Default,
		}
		importVars[variable.Name] = entry
	}

	return importVars
}

func schemaToSpec(object object.Object) (*Normalization, error) {
	spec := &Normalization{
		Name: object.DisplayName,
		TerraformProviderConfig: TerraformProviderConfig{
			SkipResource:          object.TerraformConfig.SkipResource,
			SkipDatasource:        object.TerraformConfig.SkipDatasource,
			SkipDatasourceListing: object.TerraformConfig.SkipdatasourceListing,
			Suffix:                object.TerraformConfig.Suffix,
		},
		Locations:   make(map[string]*Location),
		Imports:     make(map[string]*Import),
		GoSdkPath:   object.GoSdkConfig.Package,
		XpathSuffix: object.XpathSuffix,
		Version:     object.Version,
		Spec: &Spec{
			Params: make(map[string]*SpecParam),
			OneOf:  make(map[string]*SpecParam),
		},
	}

	for _, location := range object.Locations {
		var xpath []string

		schemaXpathVars := make(map[string]xpathschema.Variable)
		for _, elt := range location.Xpath.Variables {
			schemaXpathVars[elt.Name] = elt
		}
		for _, elt := range location.Xpath.Elements {
			var eltEntry string
			if xpathVar, ok := schemaXpathVars[elt[1:]]; ok {
				if xpathVar.Type == "entry" {
					eltEntry = fmt.Sprintf("{{ Entry %s }}", elt)
				} else if xpathVar.Type == "object" {
					eltEntry = fmt.Sprintf("{{ Object %s }}", elt)
				}
			} else {
				if strings.HasPrefix(elt, "$") {
					panic(fmt.Sprintf("elt: %s", elt))
				}
				eltEntry = elt
			}
			xpath = append(xpath, eltEntry)
		}

		locationDevice := &LocationDevice{}

		for _, device := range location.Devices {
			if device == "panorama" {
				locationDevice.Panorama = true
			} else if device == "ngfw" {
				locationDevice.Ngfw = true
			}
		}

		xpathVars := generateXpathVariables(location.Xpath.Variables)
		if len(xpathVars) == 0 {
			xpathVars = nil
		}

		entry := &Location{
			Description: location.Description,
			Device:      locationDevice,
			Xpath:       xpath,
			Vars:        xpathVars,
		}
		spec.Locations[location.Name] = entry
	}

	for _, entry := range object.Entries {
		if entry.Name == "name" {
			specEntry := &Entry{
				Name: &EntryName{
					Description: entry.Description,
				},
			}

			for _, v := range entry.Validators {
				switch spec := v.Spec.(type) {
				case *validator.StringLengthSpec:
					min := int64(spec.Min)
					max := int64(spec.Max)
					specEntry.Name.Length = &EntryNameLength{
						Min: &min,
						Max: &max,
					}
				}
			}
			spec.Entry = specEntry
		}

	}

	imports := make(map[string]*Import, len(object.Imports))
	for _, elt := range object.Imports {
		var xpath []string

		schemaXpathVars := make(map[string]xpathschema.Variable)
		for _, xpathVariable := range elt.Xpath.Variables {
			schemaXpathVars[xpathVariable.Name] = xpathVariable
		}

		for _, element := range elt.Xpath.Elements {
			var eltEntry string
			if xpathVar, ok := schemaXpathVars[element[1:]]; ok {
				if xpathVar.Type == "entry" {
					eltEntry = fmt.Sprintf("{{ Entry %s }}", elt.Name)
				} else if xpathVar.Type == "object" {
					eltEntry = fmt.Sprintf("{{ Object %s }}", elt.Name)
				}
			} else {
				eltEntry = element
			}
			xpath = append(xpath, eltEntry)
		}

		importVariables := generateImportVariables(elt.Xpath.Variables)
		if len(importVariables) == 0 {
			importVariables = nil
		}

		imports[elt.Name] = &Import{
			Xpath:         xpath,
			Vars:          importVariables,
			OnlyForParams: elt.OnlyForParams,
		}
	}

	if len(imports) > 0 {
		spec.Imports = imports
	}

	consts := make(map[string]*Const)
	for _, param := range object.Spec.Parameters {
		specParam, err := schemaParameterToSpecParameter(param)
		if err != nil {
			return nil, err
		}

		switch spec := param.Spec.(type) {
		case *parameter.EnumSpec:
			constValues := make(map[string]*ConstValue)
			for _, elt := range spec.Values {
				if elt.Const == "" {
					continue
				}
				constValues[elt.Const] = &ConstValue{
					Value: elt.Value,
				}
			}
			if len(constValues) > 0 {
				consts[param.Name] = &Const{
					Values: constValues,
				}
			}

		}
		spec.Spec.Params[param.Name] = specParam
	}

	if len(consts) > 0 {
		spec.Const = consts
	}

	for _, param := range object.Spec.Variants {
		specParam, err := schemaParameterToSpecParameter(param)
		if err != nil {
			return nil, err
		}
		spec.Spec.OneOf[param.Name] = specParam
	}

	return spec, nil
}

// ParseSpec parse single spec (unmarshal file), add name variants for locations and params, add default types for params.
func ParseSpec(input []byte) (*Normalization, error) {
	var object object.Object
	err := content.Unmarshal(input, &object)
	if err != nil {
		return nil, err
	}

	spec, err := schemaToSpec(object)
	if err != nil {
		return nil, err
	}

	err = spec.AddNameVariantsForLocation()
	if err != nil {
		return nil, err
	}

	err = spec.AddNameVariantsForImports()
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

	return spec, err
}

// AddNameVariantsForLocation add name variants for location (under_score and CamelCase).
func (spec *Normalization) AddNameVariantsForLocation() error {
	for key, location := range spec.Locations {
		location.Name = &NameVariant{
			Underscore:     naming.Underscore("", key, ""),
			CamelCase:      naming.CamelCase("", key, "", true),
			LowerCamelCase: naming.CamelCase("", key, "", false),
		}

		for subkey, variable := range location.Vars {
			variable.Name = &NameVariant{
				Underscore:     naming.Underscore("", subkey, ""),
				CamelCase:      naming.CamelCase("", subkey, "", true),
				LowerCamelCase: naming.CamelCase("", subkey, "", false),
			}
		}
	}

	return nil
}

// AddNameVariantsForImports add name variants for imports (under_score and CamelCase).
func (spec *Normalization) AddNameVariantsForImports() error {
	for key, imp := range spec.Imports {
		imp.Name = &NameVariant{
			Underscore:     naming.Underscore("", key, ""),
			CamelCase:      naming.CamelCase("", key, "", true),
			LowerCamelCase: naming.CamelCase("", key, "", false),
		}

		for subkey, variable := range imp.Vars {
			variable.Name = &NameVariant{
				Underscore:     naming.Underscore("", subkey, ""),
				CamelCase:      naming.CamelCase("", subkey, "", true),
				LowerCamelCase: naming.CamelCase("", subkey, "", false),
			}
		}
	}

	return nil
}

// AddNameVariantsForParams recursively add name variants for params for nested specs.
func AddNameVariantsForParams(name string, param *SpecParam) error {
	param.Name = &NameVariant{
		Underscore:     naming.Underscore("", name, ""),
		CamelCase:      naming.CamelCase("", name, "", true),
		LowerCamelCase: naming.CamelCase("", name, "", false),
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
				Underscore:     naming.Underscore("", nameType, ""),
				CamelCase:      naming.CamelCase("", nameType, "", true),
				LowerCamelCase: naming.CamelCase("", nameType, "", false),
			}
			for nameValue, customValue := range customType.Values {
				customValue.Name = &NameVariant{
					Underscore:     naming.Underscore("", nameValue, ""),
					CamelCase:      naming.CamelCase("", nameValue, "", true),
					LowerCamelCase: naming.CamelCase("", nameValue, "", false),
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

func (spec *Normalization) EntryOrConfig() string {
	if spec.Entry == nil {
		return "Config"
	}

	return "Entry"
}

func (spec *Normalization) HasEntryName() bool {
	return spec.Entry != nil
}

func (spec *Normalization) HasEncryptedResources() bool {
	for _, param := range spec.Spec.Params {
		if param.HasEncryptedResources() {
			return true
		}
	}

	for _, param := range spec.Spec.OneOf {
		if param.HasEncryptedResources() {
			return true
		}
	}

	return false
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
