package properties

import (
	"bytes"
	"fmt"
	"io/fs"
	"maps"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/content"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/object"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/parameter"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/validator"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/xpath"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/version"
)

type Normalization struct {
	Name                    string                  `json:"name" yaml:"name"`
	TerraformProviderConfig TerraformProviderConfig `json:"terraform_provider_config" yaml:"terraform_provider_config"`
	GoSdkSkip               bool                    `json:"go_sdk_skip" yaml:"go_sdk_skip"`
	GoSdkPath               []string                `json:"go_sdk_path" yaml:"go_sdk_path"`
	PanosXpath              PanosXpath              `json:"panos_xpath" yaml:"panos_xpath"`
	Locations               map[string]*Location    `json:"locations" yaml:"locations"`
	Entry                   *Entry                  `json:"entry" yaml:"entry"`
	Imports                 []Import                `json:"imports" yaml:"imports"`
	Version                 string                  `json:"version" yaml:"version"`
	Spec                    *Spec                   `json:"spec" yaml:"spec"`
	Const                   map[string]*Const       `json:"const" yaml:"const"`
}

type PanosXpath struct {
	Path      []string                    `yaml:"path"`
	Variables []object.PanosXpathVariable `yaml:"vars"`
}

type Import struct {
	Variant   *NameVariant
	Type      *NameVariant
	Locations map[string]ImportLocation
}

type ImportLocation struct {
	Name           *NameVariant
	SpecOrder      int
	Required       bool
	XpathElements  []string
	XpathVariables map[string]ImportXpathVariable
}

type ImportXpathVariable struct {
	Name        *NameVariant
	SpecOrder   int
	Description string
	Default     string
}

type TerraformResourceType string

const (
	TerraformResourceEntry  TerraformResourceType = "entry"
	TerraformResourceUuid   TerraformResourceType = "uuid"
	TerraformResourceConfig TerraformResourceType = "config"
	TerraformResourceCustom TerraformResourceType = "custom"
)

type TerraformResourceVariant string

const (
	TerraformResourceSingular TerraformResourceVariant = "singular"
	TerraformResourcePlural   TerraformResourceVariant = "plural"
)

type TerraformProviderConfig struct {
	Description           string                     `json:"description" yaml:"description"`
	Ephemeral             bool                       `json:"ephemeral" yaml:"ephemeral"`
	SkipResource          bool                       `json:"skip_resource" yaml:"skip_resource"`
	SkipDatasource        bool                       `json:"skip_datasource" yaml:"skip_datasource"`
	SkipDatasourceListing bool                       `json:"skip_datasource_listing" yaml:"skip_datasource_listing"`
	ResourceType          TerraformResourceType      `json:"resource_type" yaml:"resource_type"`
	CustomFuncs           map[string]string          `json:"custom_functions" yaml:"custom_functions"`
	ResourceVariants      []TerraformResourceVariant `json:"resource_variants" yaml:"resource_variants"`
	Suffix                string                     `json:"suffix" yaml:"suffix"`
	PluralSuffix          string                     `json:"plural_suffix" yaml:"plural_suffix"`
	PluralName            string                     `json:"plural_name" yaml:"plural_name"`
	PluralType            object.TerraformPluralType `json:"plural_type" yaml:"plural_type"`
	PluralDescription     string                     `json:"plural_description" yaml:"plural_description"`
}

type Location struct {
	Name        *NameVariant
	SpecOrder   int
	Description string                  `json:"description" yaml:"description"`
	Device      *LocationDevice         `json:"device" yaml:"device"`
	Xpath       []string                `json:"xpath" yaml:"xpath"`
	ReadOnly    bool                    `json:"read_only" yaml:"read_only"`
	Vars        map[string]*LocationVar `json:"vars" yaml:"vars"`
}

func (o Location) ValidatorType() string {
	return "object"
}

func (o *Location) OrderedVars() []*LocationVar {
	elements := make([]*LocationVar, len(o.Vars))

	for _, elt := range o.Vars {
		elements[elt.StateOrder] = elt
	}

	return elements
}

type LocationDevice struct {
	Panorama bool `json:"panorama" yaml:"panorama"`
	Ngfw     bool `json:"ngfw" yaml:"ngfw"`
}

type LocationVar struct {
	Name        *NameVariant
	StateOrder  int
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

type EnumValue struct {
	Name  string
	Value *string
}

type SpecParam struct {
	Name                    *NameVariant
	SpecOrder               int                               `json:"-" yaml:"-"`
	Description             string                            `json:"description" yaml:"description"`
	TerraformProviderConfig *SpecParamTerraformProviderConfig `json:"terraform_provider_config" yaml:"terraform_provider_config"`
	GoSdkConfig             *SpecParamGoSdkConfig             `json:"gosdk_config" yaml:"gosdk_config"`
	IsNil                   bool                              `json:"-" yaml:"-"`
	Type                    string                            `json:"type" yaml:"type"`
	Default                 string                            `json:"default" yaml:"default"`
	Required                bool                              `json:"required" yaml:"required"`
	Length                  *SpecParamLength                  `json:"length" yaml:"length,omitempty"`
	EnumValues              []EnumValue                       `json:"enum_values" yaml:"enum_values,omitempty"`
	Count                   *SpecParamCount                   `json:"count" yaml:"count,omitempty"`
	Hashing                 *SpecParamHashing                 `json:"hashing" yaml:"hashing,omitempty"`
	Items                   *SpecParamItems                   `json:"items" yaml:"items,omitempty"`
	Regex                   string                            `json:"regex" yaml:"regex,omitempty"`
	Profiles                []*SpecParamProfile               `json:"profiles" yaml:"profiles"`
	VariantGroupId          int                               `json:"variant_group_id" yaml:"variant_group_id"`
	Spec                    *Spec                             `json:"spec" yaml:"spec"`
}

type SpecParamGoSdkConfig struct {
	Skip *bool `json:"skip" yaml:"skip"`
}

type SpecParamTerraformProviderConfig struct {
	Name          *string `json:"name" yaml:"name"`
	Type          *string `json:"type" yaml:"type"`
	Private       *bool   `json:"ignored" yaml:"private"`
	Sensitive     *bool   `json:"sensitive" yaml:"sensitive"`
	Computed      *bool   `json:"computed" yaml:"computed"`
	Required      *bool   `json:"required" yaml:"required"`
	VariantCheck  *string `json:"variant_check" yaml:"variant_check"`
	XpathVariable *string `json:"xpath_variable" yaml:"xpath_variable"`
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
	Xpath      []string         `json:"xpath" yaml:"xpath"`
	Type       string           `json:"type" yaml:"type,omitempty"`
	MinVersion *version.Version `json:"not_present" yaml:"min_version"`
	MaxVersion *version.Version `json:"max_version" yaml:"max_version"`
}

func (o *Spec) HackFixInjectedNameSpecOrder() {
	for _, elt := range o.Params {
		if elt.Name.CamelCase != "Name" {
			elt.SpecOrder += 1
		}
	}
}

func (o *Spec) SortedParams() []*SpecParam {
	params := make([]*SpecParam, len(o.Params))

	var idx int
	for _, elt := range o.Params {
		params[elt.SpecOrder] = elt
		idx += 1
	}

	return params
}

func (o *Spec) SortedOneOf() []*SpecParam {
	if len(o.OneOf) == 0 {
		return nil
	}

	params := make([]*SpecParam, len(o.OneOf))
	for _, elt := range o.OneOf {
		params[elt.SpecOrder] = elt
	}

	return params
}

func (o *SpecParam) NameVariant() *NameVariant {
	if o.TerraformProviderConfig != nil && o.TerraformProviderConfig.Name != nil {
		return NewNameVariant(*o.TerraformProviderConfig.Name)
	}

	return o.Name
}

func (o *SpecParam) ComplexType() string {
	var terraformType *string
	if o.TerraformProviderConfig != nil {
		terraformType = o.TerraformProviderConfig.Type
	}

	if terraformType != nil && o.Type == "list" && o.Items.Type == "string" {
		return *terraformType
	}

	return ""
}

func (o *SpecParam) FinalType() string {
	if o.TerraformProviderConfig != nil && o.TerraformProviderConfig.Type != nil {
		return *o.TerraformProviderConfig.Type
	}

	return o.Type
}

func (o *SpecParam) FinalSensitive() bool {
	if o.TerraformProviderConfig != nil && o.TerraformProviderConfig.Sensitive != nil {
		return *o.TerraformProviderConfig.Sensitive
	}

	if o.Hashing != nil {
		return true
	}

	return false
}

func (o *SpecParam) FinalComputed() bool {
	if o.TerraformProviderConfig != nil && o.TerraformProviderConfig.Computed != nil {
		return *o.TerraformProviderConfig.Computed
	}

	if o.Default != "" {
		return true
	}

	return false
}

func (o *SpecParam) FinalRequired() bool {
	if o.TerraformProviderConfig != nil && o.TerraformProviderConfig.Required != nil {
		return *o.TerraformProviderConfig.Required
	}

	return o.Required
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

func (o *SpecParam) DefaultType() string {
	return o.Type
}

func (o *SpecParam) ValidatorType() string {
	if o.Type == "" {
		return "object"
	} else if o.Type == "list" && o.Items.Type == "entry" {
		return "object"
	} else if o.Type == "list" {
		return "list"
	} else {
		return o.Type
	}
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

func (o *SpecParam) HasPrivateParameters() bool {
	if o.TerraformProviderConfig != nil && o.TerraformProviderConfig.Private != nil {
		return *o.TerraformProviderConfig.Private
	}

	for _, elt := range o.Spec.Params {
		if elt.HasPrivateParameters() {
			return true
		}
	}

	for _, elt := range o.Spec.OneOf {
		if elt.HasPrivateParameters() {
			return true
		}
	}

	return false
}

func (o *SpecParam) IsTerraformOnly() bool {
	if o.GoSdkConfig != nil && o.GoSdkConfig.Skip != nil {
		return *o.GoSdkConfig.Skip
	}

	return false
}

func (o *SpecParam) IsPrivateParameter() bool {
	if o.TerraformProviderConfig != nil && o.TerraformProviderConfig.Private != nil {
		return *o.TerraformProviderConfig.Private
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

	var paramTypeIsNil bool
	var innerSpec *Spec
	var itemsSpec SpecParamItems

	generateInnerSpec := func(spec *parameter.StructSpec) (*Spec, error) {
		params := make(map[string]*SpecParam)
		oneofs := make(map[string]*SpecParam)

		for idx, elt := range spec.Parameters {
			param, err := schemaParameterToSpecParameter(elt)
			if err != nil {
				return nil, err
			}
			param.SpecOrder = idx
			params[elt.Name] = param
		}

		for idx, elt := range spec.Variants {
			param, err := schemaParameterToSpecParameter(elt)
			if err != nil {
				return nil, err
			}
			param.SpecOrder = idx
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
		} else if spec.Items.Type == "enum" {
			itemsSpec.Type = "string"
		} else {
			itemsSpec.Type = spec.Items.Type
		}
		for _, v := range schemaSpec.Validators {
			switch spec := v.Spec.(type) {
			case *validator.CountSpec:
				minValue := int64(spec.Min)
				maxValue := int64(spec.Max)
				itemsSpec.Length = &SpecParamItemsLength{
					Min: &minValue,
					Max: &maxValue,
				}
			}
		}
	case *parameter.EnumSpec:
		defaultVal = spec.Default
	case *parameter.NilSpec:
		paramTypeIsNil = true
		specType = ""
	case *parameter.SimpleSpec:
		if typed, ok := spec.Default.(string); ok {
			defaultVal = typed
		} else if typed, ok := spec.Default.(int); ok {
			defaultVal = strconv.Itoa(typed)
		}
	}

	var profiles []*SpecParamProfile
	for _, profile := range schemaSpec.Profiles {
		profiles = append(profiles, &SpecParamProfile{
			Xpath:      profile.Xpath,
			Type:       profile.Type,
			MinVersion: profile.MinimumVersion,
			MaxVersion: profile.MaximumVersion,
		})
	}

	var specHashing *SpecParamHashing
	if schemaSpec.Hashing != nil {
		specHashing = &SpecParamHashing{
			Type: schemaSpec.Hashing.Type,
		}
	}

	var terraformProviderConfig *SpecParamTerraformProviderConfig
	var goSdkConfig *SpecParamGoSdkConfig
	if schemaSpec.CodegenOverrides != nil {
		var variantCheck *string
		if schemaSpec.CodegenOverrides.Terraform.VariantCheck != nil {
			converted := string(*schemaSpec.CodegenOverrides.Terraform.VariantCheck)
			variantCheck = &converted
		}

		terraformProviderConfig = &SpecParamTerraformProviderConfig{
			Name:          schemaSpec.CodegenOverrides.Terraform.Name,
			Type:          schemaSpec.CodegenOverrides.Terraform.Type,
			Private:       schemaSpec.CodegenOverrides.Terraform.Private,
			Sensitive:     schemaSpec.CodegenOverrides.Terraform.Sensitive,
			Computed:      schemaSpec.CodegenOverrides.Terraform.Computed,
			Required:      schemaSpec.CodegenOverrides.Terraform.Required,
			VariantCheck:  variantCheck,
			XpathVariable: schemaSpec.CodegenOverrides.Terraform.XpathVariable,
		}

		goSdkConfig = &SpecParamGoSdkConfig{
			Skip: schemaSpec.CodegenOverrides.GoSdk.Skip,
		}
	}

	specParameter := &SpecParam{
		Description:             schemaSpec.Description,
		Type:                    specType,
		IsNil:                   paramTypeIsNil,
		Default:                 defaultVal,
		Required:                schemaSpec.Required,
		GoSdkConfig:             goSdkConfig,
		TerraformProviderConfig: terraformProviderConfig,
		Hashing:                 specHashing,
		Profiles:                profiles,
		Spec:                    innerSpec,
		VariantGroupId:          schemaSpec.VariantGroupId,
	}

	for _, v := range schemaSpec.Validators {
		switch spec := v.Spec.(type) {
		case *validator.RegexpSpec:
			specParameter.Regex = spec.Expr
		case *validator.StringLengthSpec:
			minValue := int64(spec.Min)
			maxValue := int64(spec.Max)
			specParameter.Length = &SpecParamLength{
				Min: &minValue,
				Max: &maxValue,
			}
		case *validator.CountSpec:
			minValue := int64(spec.Min)
			maxValue := int64(spec.Max)
			specParameter.Count = &SpecParamCount{
				Min: &minValue,
				Max: &maxValue,
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
	for idx, variable := range variables {
		entry := &LocationVar{
			StateOrder:  idx,
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

func schemaToSpec(object object.Object) (*Normalization, error) {
	var resourceVariants []TerraformResourceVariant
	for _, elt := range object.TerraformConfig.ResourceVariants {
		resourceVariants = append(resourceVariants, TerraformResourceVariant(elt))
	}

	panosXpath := PanosXpath{
		Path:      object.PanosXpath.Path,
		Variables: object.PanosXpath.Variables,
	}

	spec := &Normalization{
		Name: object.DisplayName,
		TerraformProviderConfig: TerraformProviderConfig{
			Description:           object.TerraformConfig.Description,
			Ephemeral:             object.TerraformConfig.Epheneral,
			SkipResource:          object.TerraformConfig.SkipResource,
			SkipDatasource:        object.TerraformConfig.SkipDatasource,
			SkipDatasourceListing: object.TerraformConfig.SkipdatasourceListing,
			ResourceType:          TerraformResourceType(object.TerraformConfig.ResourceType),
			CustomFuncs:           object.TerraformConfig.CustomFunctions,
			ResourceVariants:      resourceVariants,
			Suffix:                object.TerraformConfig.Suffix,
			PluralSuffix:          object.TerraformConfig.PluralSuffix,
			PluralName:            object.TerraformConfig.PluralName,
			PluralType:            object.TerraformConfig.PluralType,
			PluralDescription:     object.TerraformConfig.PluralDescription,
		},
		Locations:  make(map[string]*Location),
		GoSdkSkip:  object.GoSdkConfig.Skip,
		GoSdkPath:  object.GoSdkConfig.Package,
		PanosXpath: panosXpath,
		Version:    object.Version,
		Spec: &Spec{
			Params: make(map[string]*SpecParam),
			OneOf:  make(map[string]*SpecParam),
		},
	}

	for idx, location := range object.Locations {
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
			SpecOrder:   idx,
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
					minValue := int64(spec.Min)
					maxValue := int64(spec.Max)
					specEntry.Name.Length = &EntryNameLength{
						Min: &minValue,
						Max: &maxValue,
					}
				}
			}
			spec.Entry = specEntry
		}

	}

	var imports []Import
	for _, elt := range object.Imports {
		locations := make(map[string]ImportLocation, len(elt.Locations))
		for idx, location := range elt.Locations {
			schemaXpathVars := make(map[string]xpathschema.Variable, len(location.Xpath.Variables))
			xpathVars := make(map[string]ImportXpathVariable, len(location.Xpath.Variables))
			for variableIdx, xpathVariable := range location.Xpath.Variables {
				schemaXpathVars[xpathVariable.Name] = xpathVariable

				_, found := xpathVars[xpathVariable.Name]
				if found {
					panic(fmt.Sprintf("Variable duplicate: %s", xpathVariable.Name))
				}

				xpathVars[xpathVariable.Name] = ImportXpathVariable{
					Name:        NewNameVariant(xpathVariable.Name),
					SpecOrder:   variableIdx,
					Description: xpathVariable.Description,
					Default:     xpathVariable.Default,
				}
			}

			var xpath []string
			xpath = append(xpath, location.Xpath.Elements...)

			locations[location.Name] = ImportLocation{
				Name:           NewNameVariant(location.Name),
				SpecOrder:      idx,
				Required:       location.Required,
				XpathVariables: xpathVars,
				XpathElements:  xpath,
			}
		}

		imports = append(imports, Import{
			Type:      NewNameVariant(elt.Type),
			Variant:   NewNameVariant(elt.Variant),
			Locations: locations,
		})
	}

	if len(imports) > 0 {
		spec.Imports = imports
	}

	consts := make(map[string]*Const)
	for idx, param := range object.Spec.Parameters {
		specParam, err := schemaParameterToSpecParameter(param)
		if err != nil {
			return nil, err
		}
		specParam.SpecOrder = idx
		var enumValues []EnumValue
		switch spec := param.Spec.(type) {
		case *parameter.EnumSpec:
			constValues := make(map[string]*ConstValue)
			for _, elt := range spec.Values {
				var constValue *string
				if elt.Const == "" {
					constValue = nil
				} else {
					constValue = &elt.Const
				}

				enumValues = append(enumValues, EnumValue{
					Name:  elt.Value,
					Value: constValue,
				})

				if constValue != nil {
					constValues[*constValue] = &ConstValue{
						Value: elt.Value,
					}
				}
			}
			if len(constValues) > 0 {
				consts[param.Name] = &Const{
					Values: constValues,
				}
			}
		}
		specParam.EnumValues = enumValues
		spec.Spec.Params[param.Name] = specParam
	}

	if len(consts) > 0 {
		spec.Const = consts
	}

	for idx, param := range object.Spec.Variants {
		specParam, err := schemaParameterToSpecParameter(param)
		if err != nil {
			return nil, err
		}
		specParam.SpecOrder = idx
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

func (spec *Normalization) OrderedLocations() []*Location {
	elements := make([]*Location, len(spec.Locations))
	for _, elt := range spec.Locations {
		elements[elt.SpecOrder] = elt
	}

	return elements
}

func (o *Import) OrderedLocations() []*ImportLocation {
	elements := make([]*ImportLocation, len(o.Locations))

	for _, elt := range o.Locations {
		elements[elt.SpecOrder] = &elt
	}

	return elements
}

func (o *ImportLocation) OrderedXpathVariables() []*ImportXpathVariable {
	elements := make([]*ImportXpathVariable, len(o.XpathVariables))

	for _, elt := range o.XpathVariables {
		elements[elt.SpecOrder] = &elt
	}

	return elements
}

// AddNameVariantsForLocation add name variants for location (under_score and CamelCase).
func (spec *Normalization) AddNameVariantsForLocation() error {
	for key, location := range spec.Locations {
		location.Name = NewNameVariant(key)

		for subkey, variable := range location.Vars {
			variable.Name = NewNameVariant(subkey)
		}
	}

	return nil
}

// AddNameVariantsForParams recursively add name variants for params for nested specs.
func AddNameVariantsForParams(name string, param *SpecParam) error {
	param.Name = NewNameVariant(name)
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
			customType.Name = NewNameVariant(nameType)
			for nameValue, customValue := range customType.Values {
				customValue.Name = NewNameVariant(nameValue)
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

func (spec *Normalization) XpathSuffix() []string {
	return spec.PanosXpath.Path
}

func (spec *Normalization) ResourceXpathVariablesWithChecks(checkEntryName bool) bool {
	for _, elt := range spec.PanosXpath.Variables {
		if elt.Name == "name" && !checkEntryName {
			continue
		}

		if elt.Spec.Type != "value" {
			return true
		}
	}

	return false
}

func (spec *Normalization) ResourceXpathVariablesCount() int {
	if spec.HasEntryName() && len(spec.PanosXpath.Variables) == 0 {
		return 1
	}

	return len(spec.PanosXpath.Variables)
}

const attributesFromXpathComponentsTmpl = `
{{- range .Components }}
  {{ if eq .Type "value" }}
	{{ $.Target }}.{{ .Name }} = types.StringValue(components[{{ .Idx }}])
  {{- else if eq .Type "entry" }}
{
	component := components[{{ .Idx }}]
	component = strings.TrimPrefix(component, "entry[@name='")
	component = strings.TrimSuffix(component, "']")
	{{ $.Target }}.{{ .Name.CamelCase }} = types.StringValue(component)
}
  {{- end }}
{{- end }}
`

func (spec *Normalization) AttributesFromXpathComponents(target string) (string, error) {
	type componentCtx struct {
		Type object.PanosXpathVariableType
		Name *NameVariant
		Idx  int
	}

	type componentsCtx struct {
		Target     string
		Components []componentCtx
	}

	var components []componentCtx
	for idx, elt := range spec.PanosXpath.Variables {
		if elt.Name == "name" {
			continue
		}

		param, err := spec.ParameterForPanosXpathVariable(elt)
		if err != nil {
			return "", err
		}

		components = append(components, componentCtx{
			Type: elt.Spec.Type,
			Name: param.Name,
			Idx:  idx,
		})
	}

	data := componentsCtx{
		Target:     target,
		Components: components,
	}

	tmpl := template.Must(template.New("attributes-from-xpath-components").Parse(attributesFromXpathComponentsTmpl))

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, data)
	if err != nil {
		panic(err)
	}

	return buf.String(), nil
}

const resourceXpathAssignmentsTmpl = `
{{- range .Variables }}
  {{- if or (eq .Type "entry") (eq .Type "value") }}
	ans = append(ans, components[{{ .Idx }}])
  {{- else if eq .Type "static" }}
	ans = append(ans, "{{ .Name.Original }}")
  {{- end }}
{{- end }}
`

func (spec *Normalization) ResourceXpathAssignments() (string, error) {
	data := xpathVariablesCtx{
		Variables: createXpathVariablesContext(spec),
	}

	tmpl := template.Must(template.New("resource-xpath-assignments").Parse(resourceXpathAssignmentsTmpl))

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, data)
	if err != nil {
		panic(err)
	}

	return buf.String(), nil
}

const xpathVariableChecksTmpl = `
{{- range .Variables }}
  {{- if eq .Type "entry" }}
{
	component := components[{{ .Idx }}]
  {{- if .AllowEmpty }}
	if component != "entry" {
  {{- end }}
	if !strings.HasPrefix(component, "entry[@name=\"]") && !strings.HasPrefix(component, "entry[@name='") {
		return nil, errors.NewInvalidXpathComponentError(fmt.Sprintf("{{ .Name.CamelCase }} must be formatted as entry: %s", component))
	}

	if !strings.HasSuffix(component, "\"]") && !strings.HasSuffix(component, "']") {
		return nil, errors.NewInvalidXpathComponentError(fmt.Sprintf("{{ .Name.CamelCase }} must be formatted as entry: %s", component))
	}
  {{- if .AllowEmpty }}
	}
  {{- end }}
}
  {{- end }}
{{- end }}
`

type xpathVariableCtx struct {
	Type       object.PanosXpathVariableType
	Name       *NameVariant
	AllowEmpty bool
	Idx        int
}

type xpathVariablesCtx struct {
	Variables []xpathVariableCtx
}

func createXpathVariablesContext(spec *Normalization) []xpathVariableCtx {
	var variables []xpathVariableCtx

	variablesByName := make(map[string]object.PanosXpathVariable)
	for _, elt := range spec.PanosXpath.Variables {
		variablesByName[elt.Name] = elt
	}

	var idx int
	for _, elt := range spec.PanosXpath.Path {
		if strings.HasPrefix(elt, "$") {
			variable, found := variablesByName[elt[1:]]
			if !found {
				panic("couldn't match panos xpath variable to path item")
			}

			var allowEmpty bool
			var name *NameVariant
			if variable.Name == "name" {
				allowEmpty = true
				name = NewNameVariant("name")
			} else {
				param, err := spec.ParameterForPanosXpathVariable(variable)
				if err != nil {
					panic(fmt.Sprintf("couldn't find matching param for xpath variable: %s", err.Error()))
				}
				name = param.Name
			}

			variables = append(variables, xpathVariableCtx{
				Type:       variable.Spec.Type,
				AllowEmpty: allowEmpty,
				Name:       name,
				Idx:        idx,
			})

			idx += 1
		} else {
			variables = append(variables, xpathVariableCtx{
				Type: object.PanosXpathVariableStatic,
				Name: NewNameVariant(elt),
			})
		}
	}

	return variables
}

func (spec *Normalization) ResourceXpathVariableChecks() (string, error) {
	data := xpathVariablesCtx{
		Variables: createXpathVariablesContext(spec),
	}

	var buf bytes.Buffer
	tmpl := template.Must(template.New("xpath-variable-checks").Parse(xpathVariableChecksTmpl))

	err := tmpl.Execute(&buf, data)
	if err != nil {
		panic(err)
	}

	return buf.String(), nil
}

// Validate validations for specification (normalization) e.g. check if XPath contain /.
func (spec *Normalization) Validate() []error {
	var checks []error

	if strings.Contains(spec.TerraformProviderConfig.Suffix, "panos_") {
		checks = append(checks, fmt.Errorf("suffix for Terraform provider cannot contain `panos_`"))
	}
	for _, suffix := range spec.XpathSuffix() {
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

// SupportedVersionDefinitions returns a map of versions keyed by their variable name.
//
// For example, for versions "10.1.0" and "11.0.2" it will return:
// {"version_10_1_0": "10.1.0", "version_11_0_2": "11.0.2"}
func (spec *Normalization) SupportedVersionDefinitions() map[string]string {
	versions := make(map[string]string)

	for _, elt := range spec.Spec.Params {
		maps.Copy(versions, supportedVersionDefinitions(elt))
	}

	for _, elt := range spec.Spec.OneOf {
		maps.Copy(versions, supportedVersionDefinitions(elt))
	}

	return versions
}

func supportedVersionDefinitions(param *SpecParam) map[string]string {
	versionsFromProfile := func(profiles []*SpecParamProfile) map[string]string {
		versions := make(map[string]string)
		for _, elt := range profiles {
			if elt.MinVersion != nil {
				minVer := elt.MinVersion
				minVerKey := fmt.Sprintf("version_%d_%d_%d", minVer.Major, minVer.Minor, minVer.Patch)
				versions[minVerKey] = elt.MinVersion.String()

				maxVer := version.Version{Major: minVer.Major, Minor: minVer.Minor + 1, Patch: 0, Hotfix: ""}
				maxVerKey := fmt.Sprintf("version_%d_%d_%d", maxVer.Major, maxVer.Minor, maxVer.Patch)
				versions[maxVerKey] = maxVer.String()
			}

			// if elt.MaxVersion != nil {
			// 	maxVer := strings.ReplaceAll(elt.MaxVersion.String(), ".", "_")
			// 	maxVerKey := fmt.Sprintf("version_%s", maxVer)
			// 	versions[maxVerKey] = elt.MaxVersion.String()
			// }
		}

		return versions
	}

	versions := versionsFromProfile(param.Profiles)
	if param.Spec != nil {
		if param.Spec.Params != nil {
			for _, elt := range param.Spec.Params {
				maps.Copy(versions, supportedVersionDefinitions(elt))
			}
		}

		if param.Spec.OneOf != nil {
			for _, elt := range param.Spec.OneOf {
				maps.Copy(versions, supportedVersionDefinitions(elt))
			}
		}
	}

	return versions
}

type SupportedVersion struct {
	Minimum version.Version
	Maximum version.Version
}

func (o *SupportedVersion) MinimumVariable() string {
	return fmt.Sprintf("version_%s", o.MinimumSuffix())
}

func (o *SupportedVersion) MaximumVariable() string {
	return fmt.Sprintf("version_%s", o.MaximumSuffix())
}

func (o *SupportedVersion) MinimumSuffix() string {
	return fmt.Sprintf("%d_%d_%d", o.Minimum.Major, o.Minimum.Minor, o.Minimum.Patch)
}

func (o *SupportedVersion) MaximumSuffix() string {
	return fmt.Sprintf("%d_%d_%d", o.Maximum.Major, o.Maximum.Minor, o.Maximum.Patch)
}

func (o *SupportedVersion) SpecifierFunc() string {
	return fmt.Sprintf("specifyEntry_%s", o.MinimumSuffix())
}

func (o *SupportedVersion) EntryXmlContainer() string {
	return fmt.Sprintf("entryXmlContainer_%s", o.MinimumSuffix())
}

// SupportedVersionRanges returns a list of SupportedVersion objects for entry Versioning()
//
// SupportedVersion contains values required by a template render logic used to select
// a correct specifier function and entry xml structure for a given server version.
// Returned list is sorted in a descending order by minimum version.
func (spec *Normalization) SupportedVersionRanges() []SupportedVersion {
	var minimums []version.Version

	minimumsMap := make(map[version.Version]struct{})
	for _, elt := range spec.Spec.Params {
		maps.Copy(minimumsMap, supportedVersionRanges(elt))
	}

	for _, elt := range spec.Spec.OneOf {
		maps.Copy(minimumsMap, supportedVersionRanges(elt))
	}

	for key := range minimumsMap {
		minimums = append(minimums, key)
	}

	slices.SortFunc(minimums, func(a, b version.Version) int {
		if a.LesserThan(b) {
			return -1
		} else if a.GreaterThan(b) {
			return 1
		}

		return 0
	})

	slices.Reverse(minimums)

	var supported []SupportedVersion
	for _, elt := range minimums {
		maxVersion := version.Version{Major: elt.Major, Minor: elt.Minor + 1, Patch: 0}

		supported = append(supported, SupportedVersion{
			Minimum: elt,
			Maximum: maxVersion,
		})

	}

	return supported
}

func supportedVersionRanges(param *SpecParam) map[version.Version]struct{} {
	versions := make(map[version.Version]struct{})

	for _, elt := range param.Profiles {
		if elt.MinVersion == nil {
			continue
		}

		versions[*elt.MinVersion] = struct{}{}
	}

	if param.Spec != nil {
		if param.Spec.Params != nil {
			for _, elt := range param.Spec.Params {
				maps.Copy(versions, supportedVersionRanges(elt))
			}
		}

		if param.Spec.OneOf != nil {
			for _, elt := range param.Spec.OneOf {
				maps.Copy(versions, supportedVersionRanges(elt))
			}
		}
	}

	return versions
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

func (spec *Normalization) HasEntryUuid() bool {
	_, found := spec.Spec.Params["uuid"]
	return found
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

func (spec *Normalization) HasPrivateParameters() bool {
	for _, param := range spec.Spec.Params {
		if param.HasPrivateParameters() {
			return true
		}
	}

	for _, param := range spec.Spec.OneOf {
		if param.HasPrivateParameters() {
			return true
		}
	}

	return false
}

func resolveXpath(xpath []string, spec *Spec) (*SpecParam, error) {
	fmt.Printf("xpath[0]: %s\n", xpath[0])
	elt := xpath[0]
	if elt == "spec" {
		elt = xpath[1]
	}

	var stack map[string]*SpecParam
	if strings.HasPrefix(elt, "params") {
		stack = spec.Params
	} else if strings.HasPrefix(elt, "variants") {
		stack = spec.OneOf
	} else {
		return nil, fmt.Errorf("failed to resolve variable xpath: %s", elt)
	}

	getParamNameFromXpathElt := func(elt string) string {
		nameRegex := regexp.MustCompile(`.+("|')(?P<name>.+)("|')`)
		regexNames := nameRegex.SubexpNames()
		matches := nameRegex.FindStringSubmatch(elt)

		return matches[nameRegex.SubexpIndex(regexNames[2])]
	}

	name := getParamNameFromXpathElt(elt)

	needle, found := stack[name]
	if !found {
		return nil, fmt.Errorf("failed to revolve variable xpath: %s", elt)
	}

	if len(xpath) == 1 {
		return needle, nil
	}

	if needle.Spec == nil {
		return nil, fmt.Errorf("failed to revolve variable xpath: %s", elt)
	}

	return resolveXpath(xpath[1:], needle.Spec)
}

func (spec *Normalization) ParameterForPanosXpathVariable(variable object.PanosXpathVariable) (*SpecParam, error) {
	xpath := strings.Split(variable.Spec.Xpath, "/")
	return resolveXpath(xpath[1:], spec.Spec)
}

func supportedVersions(params map[string]*SpecParam, versions []string) []string {
	for _, param := range params {
		for _, profile := range param.Profiles {
			if profile.MinVersion != nil && profile.MaxVersion != nil {
				profile_versions, err := version.SupportedPatchVersionRange(*profile.MinVersion, *profile.MaxVersion)
				if err != nil {
					panic(fmt.Sprintf("Failed to create a range of supported versions: %s", err))
				}

				for _, ver := range profile_versions {
					if notExist := listContains(versions, ver.String()); notExist {
						versions = append(versions, ver.String())
					}
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
