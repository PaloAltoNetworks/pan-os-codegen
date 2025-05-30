package translate

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/version"
)

type structType string

const (
	structXmlType structType = "xml"
	structApiType structType = "api"
)

var xmlNameVariant = &properties.NameVariant{
	Original:       "xml-name",
	LowerCamelCase: "xmlName",
	CamelCase:      "XMLName",
	Underscore:     "xml_name",
}

// LocationType function used in template location.tmpl to generate location type name.
func LocationType(location *properties.Location, pointer bool) string {
	prefix := ""
	if pointer {
		prefix = "*"
	}
	return fmt.Sprintf("%s%sLocation", prefix, location.Name.CamelCase)
}

const importLocationStructTmpl = `
type ImportLocation interface {
	XpathForLocation(version.Number, util.ILocation) ([]string, error)
	MarshalPangoXML([]string) (string, error)
	UnmarshalPangoXML([]byte) ([]string, error)
}

{{- range .Specs }}
  {{- $spec := . }}
  {{- $const := printf "%s%sType" $spec.Variant.CamelCase $spec.Type.CamelCase }}
type {{ $const }} int

const (
  {{- range .Locations }}
	{{ $spec.Variant.LowerCamelCase }}{{ $spec.Type.CamelCase }}{{ .Name.CamelCase }} {{ $const }} = iota
  {{- end }}
)


  {{ $topType := printf "%s%sImportLocation" $spec.Variant.CamelCase $spec.Type.CamelCase }}
type {{ $spec.Variant.CamelCase }}{{ $spec.Type.CamelCase }}ImportLocation struct {
	typ {{ $const }}
  {{- range .Locations }}

    {{- $typeName := printf "%s%s%sImportLocation" $spec.Variant.CamelCase $spec.Type.CamelCase .Name.CamelCase }}
	{{ .Name.LowerCamelCase }} *{{ $typeName }}
  {{- end }}
}

  {{- range .Locations }}
    {{ $location := . }}
    {{- $typeName := printf "%s%s%sImportLocation" $spec.Variant.CamelCase $spec.Type.CamelCase .Name.CamelCase }}
type {{ $typeName }} struct {
	xpath []string
    {{- range .XpathVariables }}
	{{ .Name.LowerCamelCase }} string
    {{- end }}
}

type {{ $typeName }}Spec struct {
    {{- range .XpathVariables }}
	{{ .Name.CamelCase }} string
    {{- end }}
}

func New{{ $typeName }}(spec {{ $typeName }}Spec) *{{ $topType }} {
	location := &{{ $typeName }}{
    {{- range .XpathVariables }}
	{{ .Name.LowerCamelCase }}: spec.{{ .Name.CamelCase }},
    {{- end }}
	}

	return &{{ $topType }}{
		typ: {{ $spec.Variant.LowerCamelCase }}{{ $spec.Type.CamelCase }}{{ .Name.CamelCase }},
		{{ $location.Name.LowerCamelCase }}: location,
	}
}

func (o *{{ $typeName }}) XpathForLocation(vn version.Number, loc util.ILocation) ([]string, error) {
	ans, err := loc.XpathPrefix(vn)
	if err != nil {
		return nil, err
	}

	importAns := []string{
    {{- range .XpathElements }}
		{{ . }},
    {{- end }}
	}

	return append(ans, importAns...), nil
}

func (o *{{ $typeName }}) MarshalPangoXML(interfaces []string) (string, error) {
	type member struct {
		Name string ` + "`" + `xml:",chardata"` + "`" + `
	}

	type request struct {
		XMLName xml.Name ` + "`" + `xml:"{{ .XpathFinalElement }}"` + "`" + `
		Members []member ` + "`" + `xml:"member"` + "`" + `
	}

	var members []member
	for _, elt := range interfaces {
		members = append(members, member{Name: elt})
	}

	expected := request {
		Members: members,
	}
	bytes, err := xml.Marshal(expected)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func (o *{{ $typeName }}) UnmarshalPangoXML(bytes []byte) ([]string, error) {
	type member struct {
		Name string ` + "`" + `xml:",chardata"` + "`" + `
	}

	type response struct {
		Members []member ` + "`" + `xml:"{{ .ResponseXpath }}"` + "`" + `
	}

	var existing response
	err := xml.Unmarshal(bytes, &existing)
	if err != nil {
		return nil, err
	}

	var interfaces []string
	for _, elt := range existing.Members {
		interfaces = append(interfaces, elt.Name)
	}

	return interfaces, nil
}
  {{- end }}

func (o *{{ $topType }}) MarshalPangoXML(interfaces []string) (string, error) {
	switch o.typ {
  {{- range .Locations }}
	case {{ $spec.Variant.LowerCamelCase }}{{ $spec.Type.CamelCase }}{{ .Name.CamelCase }}:
		return o.{{ .Name.LowerCamelCase }}.MarshalPangoXML(interfaces)
  {{- end }}
	default:
		return "", fmt.Errorf("invalid import location")
	}
}

func (o *{{ $topType }}) UnmarshalPangoXML(bytes []byte) ([]string, error) {
	switch o.typ {
  {{- range .Locations }}
	case {{ $spec.Variant.LowerCamelCase }}{{ $spec.Type.CamelCase }}{{ .Name.CamelCase }}:
		return o.{{ .Name.LowerCamelCase }}.UnmarshalPangoXML(bytes)
  {{- end }}
	default:
		return nil, fmt.Errorf("invalid import location")
	}
}

func (o *{{ $topType }}) XpathForLocation(vn version.Number, loc util.ILocation) ([]string, error) {
	switch o.typ {
  {{- range .Locations }}
	case {{ $spec.Variant.LowerCamelCase }}{{ $spec.Type.CamelCase }}{{ .Name.CamelCase }}:
		return o.{{ .Name.LowerCamelCase }}.XpathForLocation(vn, loc)
  {{- end }}
	default:
		return nil, fmt.Errorf("invalid import location")
	}
}
{{- end }}
`

type importXpathVariableSpec struct {
	Name        *properties.NameVariant
	Description string
	Default     *string
}

type importLocationSpec struct {
	Name              *properties.NameVariant
	XpathElements     []string
	XpathVariables    []importXpathVariableSpec
	XpathFinalElement string
	ResponseXpath     string
}

type importSpec struct {
	Variant   *properties.NameVariant
	Type      *properties.NameVariant
	Locations []importLocationSpec
}

func createImportLocationSpecsForLocation(location properties.ImportLocation) importLocationSpec {
	var variables []importXpathVariableSpec
	variablesByName := make(map[string]importXpathVariableSpec, len(location.XpathVariables))

	for _, elt := range location.OrderedXpathVariables() {
		var defaultValue *string
		if elt.Default != "" {
			defaultValue = &elt.Default
		}
		variableSpec := importXpathVariableSpec{
			Name:        elt.Name,
			Description: elt.Description,
			Default:     defaultValue,
		}

		variables = append(variables, variableSpec)
		variablesByName[elt.Name.Underscore] = variableSpec
	}

	var elements []string
	for _, elt := range location.XpathElements {
		if strings.HasPrefix(elt, "$") {
			variableName := elt[1:]
			asEntryXpath := fmt.Sprintf("util.AsEntryXpath(o.%s)", variablesByName[variableName].Name.LowerCamelCase)
			elements = append(elements, asEntryXpath)
		} else {
			elements = append(elements, fmt.Sprintf("\"%s\"", elt))
		}
	}

	xpathFinalElement := location.XpathElements[len(location.XpathElements)-1]
	responseXpath := fmt.Sprintf("result>%s>member", xpathFinalElement)

	return importLocationSpec{
		Name:              location.Name,
		XpathElements:     elements,
		XpathVariables:    variables,
		XpathFinalElement: xpathFinalElement,
		ResponseXpath:     responseXpath,
	}
}

func createImportSpecsForNormalization(spec *properties.Normalization) []importSpec {
	var specs []importSpec

	for _, imp := range spec.Imports {
		var locations []importLocationSpec
		for _, location := range imp.OrderedLocations() {
			locations = append(locations, createImportLocationSpecsForLocation(*location))
		}

		specs = append(specs, importSpec{
			Variant:   imp.Variant,
			Type:      imp.Type,
			Locations: locations,
		})
	}

	return specs
}

func RenderEntryImportStructs(spec *properties.Normalization) (string, error) {
	type renderContext struct {
		Specs []importSpec
	}

	tmpl, err := template.New("render-entry-import-structs").Parse(importLocationStructTmpl)
	if err != nil {
		return "", err
	}

	data := renderContext{
		Specs: createImportSpecsForNormalization(spec),
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// SpecParamType returns param type (it can be a nested spec) for structs based on spec from YAML files.
func SpecParamType(parent string, param *properties.SpecParam) string {
	prefix := determinePrefix(param, false)

	calculatedType := ""
	if param.Spec != nil {
		calculatedType = calculateNestedSpecType(parent, param)
	} else if param.Type == "list" && param.Items != nil {
		calculatedType = determineListType(param)
	} else {
		calculatedType = param.Type
	}

	return fmt.Sprintf("%s%s", prefix, calculatedType)
}

// ParamType return param type (it can be nested spec) (for struct based on spec from YAML files).
func ParamType(structTyp structType, parentName *properties.NameVariant, param *properties.SpecParam, suffix string) string {
	var calculatedType string
	if param.Type == "" || isParamListAndProfileTypeIsExtendedEntry(param) {
		typ := calculateNestedXmlSpecType(structTyp, parentName, param, suffix)
		if structTyp == structXmlType {
			calculatedType = typ.LowerCamelCase
		} else {
			calculatedType = typ.CamelCase
		}
	} else if isParamListAndProfileTypeIsMember(param) {
		if structTyp == structXmlType {
			calculatedType = "util.Member"
		} else {
			calculatedType = param.Items.Type
		}
	} else if isParamListAndProfileTypeIsSingleEntry(param) {
		if structTyp == structXmlType {
			calculatedType = "util.Entry"
		} else {
			calculatedType = calculateNestedXmlSpecType(structTyp, parentName, param, suffix).CamelCase
		}
	} else if param.Type == "bool" && structTyp == structXmlType {
		calculatedType = "string"
	} else {
		calculatedType = param.Type
	}

	return calculatedType
}

func XmlParamType(parent string, param *properties.SpecParam) string {
	return ParamType(structXmlType, properties.NewNameVariant(parent), param, "")
}

func determinePrefix(param *properties.SpecParam, useMemberOrEntryTypeStruct bool) string {
	if param.Type == "list" {
		if useMemberOrEntryTypeStruct && (isParamListAndProfileTypeIsMember(param) || isParamListAndProfileTypeIsSingleEntry(param)) {
			return "*"
		} else {
			return "[]"
		}
	} else if !param.Required {
		return "*"
	}
	return ""
}

func determineListType(param *properties.SpecParam) string {
	if param.Items.Type == "object" && param.Items.Ref != nil {
		return "string"
	}
	return param.Items.Type
}

func calculateNestedSpecType(parent string, param *properties.SpecParam) string {
	return fmt.Sprintf("%s%s", parent, naming.CamelCase("", param.Name.CamelCase, "", true))
}

func calculateNestedXmlSpecType(structTyp structType, parentName *properties.NameVariant, param *properties.SpecParam, suffix string) *properties.NameVariant {
	var typ *properties.NameVariant
	if parentName.IsEmpty() {
		typ = param.Name
	} else {
		typ = parentName.WithSuffix(param.Name)
	}

	if structTyp == structXmlType {
		typ = typ.WithSuffix(properties.NewNameVariant("xml")).WithLiteralSuffix(suffix)
	}

	return typ
}

// XmlName creates a string with xml name (e.g. `description`).
func XmlName(param *properties.SpecParam) string {
	if len(param.Profiles) > 0 {
		// FIXME: lists of objects have an extra "entry" element on their xpath
		if param.Type == "list" && param.Items.Type == "entry" {
			return param.Profiles[0].Xpath[0]
		}
		return strings.Join(param.Profiles[0].Xpath, ">")
	}

	return ""
}

// XmlTag creates a string with xml tag (e.g. `xml:"description,omitempty"`).
func XmlTag(param *properties.SpecParam) string {
	if len(param.Profiles) > 0 {
		suffix := ""

		if param.Name != nil && (param.Name.Underscore == "uuid" || param.Name.Underscore == "name") {
			suffix = suffix + ",attr"
		}

		if !param.Required {
			suffix = suffix + ",omitempty"
		}

		return fmt.Sprintf("`xml:\"%s%s\"`", XmlName(param), suffix)
	}

	return ""
}

// OmitEmpty return omitempty in XML tag for location, if there are variables defined.
func OmitEmpty(location *properties.Location) string {
	if location.Vars != nil {
		return ",omitempty"
	} else {
		return ""
	}
}

func CreateGoSuffixFromVersionTmpl(v any) (string, error) {
	if v != nil {
		typed, ok := v.(version.Version)
		if !ok {
			return "", fmt.Errorf("Failed to cast version to *version.Version: '%T'", v)
		}
		return CreateGoSuffixFromVersion(&typed), nil
	}

	return "", nil
}

// CreateGoSuffixFromVersion convert version into Go suffix e.g. 10.1.1 into _10_1_1
func CreateGoSuffixFromVersion(v *version.Version) string {
	if v == nil {
		return ""
	}

	return fmt.Sprintf("_%s", strings.ReplaceAll(v.String(), ".", "_"))
}

func ParamNotSkippedTmpl(param *properties.SpecParam) bool {
	if param.GoSdkConfig != nil && param.GoSdkConfig.Skip != nil {
		return !*param.GoSdkConfig.Skip
	}

	return true
}

func ParamSupportedInVersionTmpl(param *properties.SpecParam, deviceVersion any) (bool, error) {
	if deviceVersion == nil {
		return true, nil
	}

	typed, ok := deviceVersion.(version.Version)
	if !ok {
		return false, fmt.Errorf("Failed to cast deviceVersion to version.Version, received '%T'", deviceVersion)
	}

	return ParamSupportedInVersion(param, &typed), nil
}

// ParamSupportedInVersion checks if param is supported in specific PAN-OS version
func ParamSupportedInVersion(param *properties.SpecParam, deviceVersion *version.Version) bool {
	if deviceVersion == nil {
		return true
	}

	result := checkIfDeviceVersionSupportedByProfile(param, *deviceVersion)
	return result
}

func checkIfDeviceVersionSupportedByProfile(param *properties.SpecParam, deviceVersion version.Version) bool {
	for _, profile := range param.Profiles {
		if profile.MinVersion == nil && profile.MaxVersion == nil {
			return true
		}

		log.Printf("Param: %s, deviceVersion: %s, MinVersion: %s, MaxVersion: %s", param.Name.CamelCase, deviceVersion, profile.MinVersion.String(), profile.MaxVersion.String())

		if deviceVersion.GreaterThanOrEqualTo(*profile.MinVersion) && deviceVersion.LesserThan(*profile.MaxVersion) {
			return true
		}
	}
	return false
}

type entryStructFieldContext struct {
	Name             *properties.NameVariant
	IsInternal       bool
	Required         bool
	FieldType        string
	Type             string
	ItemsType        string
	XmlType          string
	XmlContainerType string
	ItemsXmlType     string
	Tags             string
	version          *version.Version
}

func (o entryStructFieldContext) FinalType() string {
	switch o.FieldType {
	case "list-entry", "list-member":
		return o.ItemsType
	case "object", "simple":
		if o.Required {
			return o.Type
		} else {
			return "*" + o.Type
		}
	case "internal":
		return o.Type
	default:
		panic(fmt.Sprintf("unreachable FieldType '%s' for '%s'", o.FieldType, o.Name.CamelCase))
	}
}

func (o entryStructFieldContext) FinalXmlType() string {
	switch o.FieldType {
	case "list-entry":
		if o.XmlContainerType != "" {
			return "*" + o.XmlContainerType
		} else {
			return o.ItemsXmlType
		}
	case "list-member":
		return "*" + o.ItemsXmlType
	case "object", "simple":
		if o.Required {
			return o.XmlType
		} else {
			return "*" + o.XmlType
		}
	case "internal":
		return o.XmlType
	default:
		panic(fmt.Sprintf("unreachable FieldType '%s' for '%s'", o.FieldType, o.Name.CamelCase))
	}
}

type entryStructContext struct {
	TopLevel       bool
	IsXmlContainer bool
	Fields         []entryStructFieldContext

	version *version.Version
	name    *properties.NameVariant
}

func (o entryStructContext) versionSuffix() string {
	if o.version == nil {
		return ""
	}

	return fmt.Sprintf("_%s", strings.ReplaceAll(o.version.String(), ".", "_"))
}

func (o entryStructContext) StructName() string {
	return o.name.CamelCase
}

func (o entryStructContext) XmlStructName() string {
	return o.name.LowerCamelCase + "Xml" + o.versionSuffix()
}

func (o entryStructContext) XmlContainerStructName() string {
	return o.name.LowerCamelCase + "XmlContainer" + o.versionSuffix()
}

func (o entryStructContext) SpecifierFuncName(suffix string) string {
	return "specify" + suffix + o.versionSuffix()
}

func getTypesForParam(structTyp structType, parent *properties.NameVariant, param *properties.SpecParam, version *version.Version, overrideForXmlContainer bool) (string, string, string) {
	var versionSuffix string
	if version != nil {
		versionSuffix = fmt.Sprintf("_%s", strings.ReplaceAll(version.String(), ".", "_"))
	}

	if structTyp == structXmlType {
		typ := ParamType(structXmlType, parent, param, versionSuffix)
		var itemsType string
		if param.Type == "list" && param.Items.Type == "entry" {
			itemsType = "[]" + typ
		} else if param.Type == "list" {
			itemsType = "util.MemberType"
		}
		var xmlContainerType string
		if overrideForXmlContainer {
			xmlContainerType = parent.WithSuffix(param.Name).WithSuffix(properties.NewNameVariant("container")).WithSuffix(properties.NewNameVariant("xml")).WithLiteralSuffix(versionSuffix).LowerCamelCase
		}
		return typ, itemsType, xmlContainerType
	} else {
		typ := ParamType(structApiType, parent, param, "")
		var itemsType string
		if param.Type == "list" && param.Items.Type == "string" {
			itemsType = "[]string"
		} else if param.Type == "list" && param.Items.Type == "int64" {
			itemsType = "[]int64"
		} else if param.Type == "list" && param.Items.Type == "entry" {
			itemsType = "[]" + typ
		}
		return typ, itemsType, ""
	}
}

func getFieldTypeForParam(param *properties.SpecParam) string {
	if param.Type == "" {
		return "object"
	}

	if param.Type == "list" && param.Items.Type == "entry" {
		return "list-entry"
	}

	if param.Type == "list" {
		return "list-member"
	}

	return "simple"
}

func createStructSpecForXmlListContainer(prefix *properties.NameVariant, param *properties.SpecParam, version *version.Version) []entryStructContext {
	typ, itemsType, _ := getTypesForParam(structApiType, prefix, param, version, false)
	xmlType, itemsXmlType, _ := getTypesForParam(structXmlType, prefix, param, version, false)
	fieldType := "list-entry"

	fields := []entryStructFieldContext{
		{
			Name:         properties.NewNameVariant("entries"),
			Required:     false,
			FieldType:    fieldType,
			Type:         typ,
			ItemsType:    itemsType,
			XmlType:      xmlType,
			ItemsXmlType: itemsXmlType,
			Tags:         "`xml:\"entry\"`",
			version:      version,
		},
	}

	return []entryStructContext{{
		IsXmlContainer: true,
		Fields:         fields,
		name:           prefix.WithSuffix(param.Name).WithSuffix(properties.NewNameVariant("container")),
		version:        version,
	}}
}

func createEntryXmlStructSpecsForParameter(structTyp structType, parentPrefix *properties.NameVariant, param *properties.SpecParam, version *version.Version) []entryStructContext {
	var fields []entryStructFieldContext
	var entries []entryStructContext

	if param.Type == "list" && param.Items.Type == "entry" {
		if structTyp == structXmlType {
			fields = append(fields, entryStructFieldContext{
				IsInternal: true,
				FieldType:  "internal",
				Name:       xmlNameVariant,
				XmlType:    "xml.Name",
				Tags:       "`xml:\"entry\"`",
			})
		}
		fields = append(fields, entryStructFieldContext{
			Name:      properties.NewNameVariant("name"),
			Required:  true,
			FieldType: "simple",
			Type:      "string",
			XmlType:   "string",
			Tags:      "`xml:\"name,attr\"`",
		})
	}

	processParameter := func(prefix *properties.NameVariant, param *properties.SpecParam) {
		if param.GoSdkConfig != nil && param.GoSdkConfig.Skip != nil && *param.GoSdkConfig.Skip {
			return
		}

		if !ParamSupportedInVersion(param, version) {
			return
		}

		var overrideTypeForXmlContainer bool
		if structTyp == structXmlType && param.Type == "list" && param.Items.Type == "entry" {
			overrideTypeForXmlContainer = true
			entries = append(entries, createStructSpecForXmlListContainer(prefix, param, version)...)
		}

		typ, itemsType, _ := getTypesForParam(structApiType, prefix, param, version, overrideTypeForXmlContainer)
		xmlType, itemsXmlType, xmlContainerType := getTypesForParam(structXmlType, prefix, param, version, overrideTypeForXmlContainer)
		fieldType := getFieldTypeForParam(param)

		fields = append(fields, entryStructFieldContext{
			Name:             param.Name,
			Required:         param.Required,
			FieldType:        fieldType,
			Type:             typ,
			ItemsType:        itemsType,
			XmlType:          xmlType,
			XmlContainerType: xmlContainerType,
			ItemsXmlType:     itemsXmlType,
			Tags:             XmlTag(param),
			version:          version,
		})

		if param.Type == "" || (param.Type == "list" && param.Items.Type == "entry") {
			entries = append(entries, createEntryXmlStructSpecsForParameter(structTyp, prefix, param, version)...)
		}

	}

	prefixName := parentPrefix.WithSuffix(param.Name)
	for _, elt := range param.Spec.SortedParams() {
		processParameter(prefixName, elt)
	}

	for _, elt := range param.Spec.SortedOneOf() {
		processParameter(prefixName, elt)
	}

	fields = append(fields, entryStructFieldContext{
		Name:      properties.NewNameVariant("misc"),
		FieldType: "internal",
		Type:      "[]generic.Xml",
		XmlType:   "[]generic.Xml",
		Tags:      "`xml:\",any\"`",
	})

	name := parentPrefix.WithSuffix(param.Name)
	entries = append([]entryStructContext{{
		Fields:  fields,
		name:    name,
		version: version,
	}}, entries...)

	return entries
}

func creasteStructSpecsForNormalization(structTyp structType, parentPrefix *properties.NameVariant, spec *properties.Normalization, version *version.Version) []entryStructContext {
	var entries []entryStructContext
	var fields []entryStructFieldContext

	if structTyp == structXmlType {
		var xmlTags string
		switch spec.TerraformProviderConfig.ResourceType {
		case properties.TerraformResourceEntry, properties.TerraformResourceUuid:
			xmlTags = "`xml:\"entry\"`"
		case properties.TerraformResourceConfig:
			xmlTags = "`xml:\"system\"`"
		case properties.TerraformResourceCustom:
			fallthrough
		default:
			panic(fmt.Sprintf("unreachable resource type: '%s'", spec.TerraformProviderConfig.ResourceType))
		}

		fields = append(fields, entryStructFieldContext{
			IsInternal: true,
			FieldType:  "internal",
			Name:       xmlNameVariant,
			XmlType:    "xml.Name",
			Tags:       xmlTags,
		})
	}

	switch spec.TerraformProviderConfig.ResourceType {
	case properties.TerraformResourceEntry, properties.TerraformResourceUuid:
		fields = append(fields, entryStructFieldContext{
			Name:      properties.NewNameVariant("name"),
			Required:  true,
			FieldType: "simple",
			Type:      "string",
			XmlType:   "string",
			Tags:      "`xml:\"name,attr\"`",
		})
	case properties.TerraformResourceConfig:
	case properties.TerraformResourceCustom:
		fallthrough
	default:
		panic(fmt.Sprintf("unreachable resource type: '%s'", spec.TerraformProviderConfig.ResourceType))
	}

	processParameter := func(prefix *properties.NameVariant, param *properties.SpecParam) {
		if param.GoSdkConfig != nil && param.GoSdkConfig.Skip != nil && *param.GoSdkConfig.Skip {
			return
		}

		if !ParamSupportedInVersion(param, version) {
			return
		}

		var overrideTypeForXmlContainer bool
		if structTyp == structXmlType && param.Type == "list" && param.Items.Type == "entry" {
			overrideTypeForXmlContainer = true
			entries = append(entries, createStructSpecForXmlListContainer(prefix, param, version)...)
		}

		typ, itemsType, _ := getTypesForParam(structApiType, prefix, param, version, overrideTypeForXmlContainer)
		xmlType, itemsXmlType, xmlContainerType := getTypesForParam(structXmlType, prefix, param, version, overrideTypeForXmlContainer)
		fieldType := getFieldTypeForParam(param)

		fields = append(fields, entryStructFieldContext{
			Name:             param.Name,
			Required:         param.Required,
			FieldType:        fieldType,
			Type:             typ,
			ItemsType:        itemsType,
			XmlType:          xmlType,
			XmlContainerType: xmlContainerType,
			ItemsXmlType:     itemsXmlType,
			Tags:             XmlTag(param),
			version:          version,
		})

		if param.Type == "" || (param.Type == "list" && param.Items.Type == "entry") {
			entries = append(entries, createEntryXmlStructSpecsForParameter(structTyp, properties.NewNameVariant(""), param, version)...)
		}
	}

	for _, elt := range spec.Spec.SortedParams() {
		processParameter(parentPrefix, elt)
	}

	for _, elt := range spec.Spec.SortedOneOf() {
		processParameter(parentPrefix, elt)
	}

	fields = append(fields, entryStructFieldContext{
		Name:      properties.NewNameVariant("misc"),
		FieldType: "internal",
		Type:      "[]generic.Xml",
		XmlType:   "[]generic.Xml",
		Tags:      "`xml:\",any\"`",
	})

	var name *properties.NameVariant
	switch spec.TerraformProviderConfig.ResourceType {
	case properties.TerraformResourceEntry, properties.TerraformResourceUuid:
		name = properties.NewNameVariant("entry")
	case properties.TerraformResourceConfig:
		name = properties.NewNameVariant("config")
	case properties.TerraformResourceCustom:
		fallthrough
	default:
		panic(fmt.Sprintf("unreachable resource type: %v", spec.TerraformProviderConfig.ResourceType))
	}

	entries = append([]entryStructContext{{
		TopLevel: true,
		Fields:   fields,
		name:     name,
		version:  version,
	}}, entries...)

	return entries
}

func createStructSpecs(structTyp structType, spec *properties.Normalization, version *version.Version) []entryStructContext {
	return creasteStructSpecsForNormalization(structTyp, properties.NewNameVariant(""), spec, version)
}

const apiStructsTmpl = `
{{- range .Specs }}
{{- $spec := . }}
type {{ .StructName }} struct{
  {{- range .Fields }}
    {{- if .IsInternal }}{{ continue }}{{ end }}
	{{ .Name.CamelCase }} {{ .FinalType }}
  {{- end }}
}
{{- end }}
`

func RenderEntryApiStructs(spec *properties.Normalization) (string, error) {
	tmpl := template.Must(template.New("render-entry-api-structs").Parse(apiStructsTmpl))

	specs := createStructSpecs(structApiType, spec, nil)
	type context struct {
		Specs []entryStructContext
	}

	data := context{Specs: specs}

	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", err
	}

	return builder.String(), nil
}

const xmlStructsTmpl = `
{{- range .Specs }}
{{- $spec := . }}
type {{ .XmlStructName }} struct{
  {{- range .Fields }}
	{{ .Name.CamelCase }} {{ .FinalXmlType }} {{ .Tags }}
  {{- end }}
}
{{- end }}
`

func RenderEntryXmlStructs(spec *properties.Normalization) (string, error) {
	tmpl := template.Must(template.New("render-entry-xml-structs").Parse(xmlStructsTmpl))

	specs := createStructSpecs(structXmlType, spec, nil)
	for _, elt := range spec.SupportedVersionRanges() {
		specs = append(specs, createStructSpecs(structXmlType, spec, &elt.Minimum)...)
	}

	type context struct {
		Specs []entryStructContext
	}

	data := context{Specs: specs}

	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", err
	}

	return builder.String(), nil
}

const structToXmlMarshalersTmpl = `
{{- range .Specs }}
  {{- if .IsXmlContainer }}{{ continue }}{{ end }}
func (o *{{ .XmlStructName }}) MarshalFromObject(s {{ .StructName }}) {
  {{- range .Fields }}
    {{- if .IsInternal }}{{ continue }}{{- end }}
    {{- if eq .FieldType "object" }}
	if s.{{ .Name.CamelCase }} != nil {
		var obj {{ .XmlType }}
		obj.MarshalFromObject(*s.{{ .Name.CamelCase }})
		o.{{ .Name.CamelCase }} = &obj
	}
    {{-  else if eq .FieldType "list-member" }}
      {{- if eq .ItemsType "[]string" }}
	if s.{{ .Name.CamelCase }} != nil {
		o.{{ .Name.CamelCase }} = util.StrToMem(s.{{ .Name.CamelCase }})
	}
      {{- else if eq .ItemsType "int64" }}
	if s.{{ .Name.CamelCase }} != nil {
		o.{{ .Name.CamelCase }} = util.Int64ToMem(s.{{ .Name.CamelCase }})
	}
      {{- end }}
    {{- else if eq .FieldType "list-entry" }}
	if s.{{ .Name.CamelCase }} != nil {
		var objs {{ .ItemsXmlType }}
		for _, elt := range s.{{ .Name.CamelCase }} {
			var obj {{ .XmlType }}
			obj.MarshalFromObject(elt)
			objs = append(objs, obj)
		}
		o.{{ .Name.CamelCase }} = &{{ .XmlContainerType }}{ Entries: objs }
	}
    {{- else if and (eq .FieldType "simple") (eq .Type "bool") }}
	o.{{ .Name.CamelCase }} = util.YesNo(s.{{ .Name.CamelCase }}, nil)
    {{- else }}
	o.{{ .Name.CamelCase }} = s.{{ .Name.CamelCase }}
    {{- end }}
  {{- end }}
}

func (o {{ .XmlStructName }}) UnmarshalToObject() (*{{ .StructName }}, error) {
  {{- range .Fields }}
    {{- if .IsInternal }}{{ continue }}{{- end }}
    {{- if eq .FieldType "object" }}
	var {{ .Name.LowerCamelCase }}Val {{ .FinalType }}
	if o.{{ .Name.CamelCase }} != nil {
		obj, err := o.{{ .Name.CamelCase }}.UnmarshalToObject()
		if err != nil {
			return nil, err
		}
		{{ .Name.LowerCamelCase }}Val = obj
	}
    {{- else if eq .FieldType "list-member" }}
	var {{ .Name.LowerCamelCase }}Val {{ .FinalType }}
      {{- if eq .ItemsType "[]string" }}
	if o.{{ .Name.CamelCase }} != nil {
		{{ .Name.LowerCamelCase }}Val = util.MemToStr(o.{{ .Name.CamelCase }})
	}
      {{- else if eq .ItemsType "[]int64" }}
	if o.{{ .Name.CamelCase }} != nil {
		var err error
		{{ .Name.LowerCamelCase }}Val, err = util.MemToInt64(o.{{ .Name.CamelCase }})
		if err != nil && !errors.Is(err, util.ErrEmptyMemberList) {
			return nil, err
		}
	}
      {{- end }}
    {{- else if eq .FieldType "list-entry" }}
	var {{ .Name.LowerCamelCase }}Val {{ .FinalType }}
	if o.{{ .Name.CamelCase }} != nil {
		for _, elt := range o.{{ .Name.CamelCase }}.Entries {
			obj, err := elt.UnmarshalToObject()
			if err != nil {
				return nil, err
			}
			{{ .Name.LowerCamelCase }}Val = append({{ .Name.LowerCamelCase }}Val, *obj)
		}
	}
    {{- end }}
  {{- end }}

	result := &{{ .StructName }}{
  {{- range .Fields }}
    {{- if .IsInternal }}{{- continue }}{{- end }}
    {{- if or (eq .FieldType "list-member") (eq .FieldType "list-entry") (eq .FieldType "object") }}
		{{ .Name.CamelCase }}: {{ .Name.LowerCamelCase }}Val,
    {{- else if and (eq .FieldType "simple") (eq .Type "bool") }}
		{{ .Name.CamelCase }}: util.AsBool(o.{{ .Name.CamelCase }}, nil),
    {{- else }}
		{{ .Name.CamelCase }}: o.{{ .Name.CamelCase }},
    {{- end }}
  {{- end }}
	}
	return result, nil
}
{{- end }}
`

func RenderToXmlMarshalers(spec *properties.Normalization) (string, error) {
	tmpl := template.Must(template.New("render-to-xml-marsrhallers").Parse(structToXmlMarshalersTmpl))

	specs := createStructSpecs(structXmlType, spec, nil)
	for _, elt := range spec.SupportedVersionRanges() {
		specs = append(specs, createStructSpecs(structXmlType, spec, &elt.Minimum)...)
	}
	type context struct {
		EntryOrConfig string
		Specs         []entryStructContext
	}

	entryOrConfig := "Entry"
	if spec.TerraformProviderConfig.ResourceType == properties.TerraformResourceConfig {
		entryOrConfig = "Config"
	}

	data := context{
		EntryOrConfig: entryOrConfig,
		Specs:         specs,
	}

	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", err
	}

	return builder.String(), nil
}

const xmlContainerNormalizersTmpl = `
{{- range .Specs }}
{{- if not .TopLevel }}{{ continue }}{{ end }}
func (o *{{ .XmlContainerStructName }}) Normalize() ([]*{{ $.EntryOrConfig }}, error) {
	entries := make([]*{{ $.EntryOrConfig }}, 0, len(o.Answer))
	for _, elt := range o.Answer {
		obj, err := elt.UnmarshalToObject()
		if err != nil {
			return nil, err
		}
		entries = append(entries, obj)
	}

	return entries, nil
}
{{- end }}
`

func RenderXmlContainerNormalizers(spec *properties.Normalization) (string, error) {
	tmpl := template.Must(template.New("render-xml-container-normalizers").Parse(xmlContainerNormalizersTmpl))

	specs := createStructSpecs(structXmlType, spec, nil)
	for _, elt := range spec.SupportedVersionRanges() {
		specs = append(specs, createStructSpecs(structXmlType, spec, &elt.Minimum)...)
	}
	type context struct {
		EntryOrConfig string
		Specs         []entryStructContext
	}

	entryOrConfig := "Entry"
	if spec.TerraformProviderConfig.ResourceType == properties.TerraformResourceConfig {
		entryOrConfig = "Config"
	}

	data := context{
		EntryOrConfig: entryOrConfig,
		Specs:         specs,
	}

	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", err
	}

	return builder.String(), nil
}

const xmlContainerSpecifiersTmpl = `
{{- range .Specs }}
{{- if not .TopLevel }}{{ continue }}{{ end }}
func {{ .SpecifierFuncName $.EntryOrConfig }}(source *{{ $.EntryOrConfig }}) (any, error) {
	var obj {{ .XmlStructName }}
	obj.MarshalFromObject(*source)
	return obj, nil
}
{{- end }}
`

func RenderXmlContainerSpecifiers(spec *properties.Normalization) (string, error) {
	tmpl := template.Must(template.New("render-xml-container-specifiers").Parse(xmlContainerSpecifiersTmpl))

	specs := createStructSpecs(structXmlType, spec, nil)
	for _, elt := range spec.SupportedVersionRanges() {
		specs = append(specs, createStructSpecs(structXmlType, spec, &elt.Minimum)...)
	}
	type context struct {
		EntryOrConfig string
		Specs         []entryStructContext
	}

	entryOrConfig := "Entry"
	if spec.TerraformProviderConfig.ResourceType == properties.TerraformResourceConfig {
		entryOrConfig = "Config"
	}

	data := context{
		EntryOrConfig: entryOrConfig,
		Specs:         specs,
	}

	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", err
	}

	return builder.String(), nil
}

const specMatchersTmpl = `
func SpecMatches(a, b *{{ .EntryOrConfig }}) bool {
	if a == nil && b == nil {
		return true
	}

	if (a == nil && b != nil) || (a != nil && b == nil) {
		return false
	}

	return a.matches(b)
}

{{- range .Specs }}
  {{ if .IsXmlContainer }}{{ continue }}{{ end }}
  {{ $spec := . }}
func (o *{{ .StructName }}) matches(other *{{ .StructName }}) bool {
	if o == nil && other == nil {
		return true
	}

	if (o == nil && other != nil) || (o != nil && other == nil) {
		return false
	}

  {{- range .Fields }}
    {{- if .IsInternal }}{{ continue }}{{ end }}
    {{- if and $spec.TopLevel (eq .Name.CamelCase "Name") }}{{ continue }}{{ end }}
    {{- if eq .Name.CamelCase "Misc" }}{{ continue }}{{ end }}
    {{- if eq .FieldType "object" }}
	if !o.{{ .Name.CamelCase }}.matches(other.{{ .Name.CamelCase }}) {
		return false
	}
    {{- else if eq .FieldType "list-entry" }}
	if len(o.{{ .Name.CamelCase }}) != len(other.{{ .Name.CamelCase }}) {
		return false
	}
	for idx := range o.{{ .Name.CamelCase }} {
		if !o.{{ .Name.CamelCase }}[idx].matches(&other.{{ .Name.CamelCase }}[idx]) {
			return false
		}
	}
    {{- else if eq .FieldType "list-member" }}
	if !util.OrderedListsMatch[{{ .Type }}](o.{{ .Name.CamelCase}}, other.{{ .Name.CamelCase }}) {
		return false
	}
    {{- else if and (eq .Type "string") (eq .Required false)}}
	if !util.StringsMatch(o.{{ .Name.CamelCase }}, other.{{ .Name.CamelCase }}) {
		return false
	}
    {{- else if and (eq .Type "int64") (eq .Required false)}}
	if !util.Ints64Match(o.{{ .Name.CamelCase }}, other.{{ .Name.CamelCase }}) {
		return false
	}
    {{- else if and (eq .Type "int64") (eq .Required false)}}
	if !util.Ints64Match(o.{{ .Name.CamelCase }}, other.{{ .Name.CamelCase }}) {
		return false
	}
    {{- else if and (eq .Type "bool") (eq .Required false)}}
	if !util.BoolsMatch(o.{{ .Name.CamelCase }}, other.{{ .Name.CamelCase }}) {
		return false
	}
    {{- else if and (eq .Type "float64") (eq .Required false)}}
	if !util.FloatsMatch(o.{{ .Name.CamelCase }}, other.{{ .Name.CamelCase }}) {
		return false
	}
    {{- else }}
	if o.{{ .Name.CamelCase }} != other.{{ .Name.CamelCase }} {
		return false
	}
    {{- end }}
  {{- end }}

	return true
}
{{- end }}
`

func RenderSpecMatchers(spec *properties.Normalization) (string, error) {
	tmpl := template.Must(template.New("render-spec-matchers").Parse(specMatchersTmpl))

	specs := createStructSpecs(structApiType, spec, nil)
	type context struct {
		EntryOrConfig string
		Specs         []entryStructContext
	}

	entryOrConfig := "Entry"
	if spec.TerraformProviderConfig.ResourceType == properties.TerraformResourceConfig {
		entryOrConfig = "Config"
	}

	data := context{
		EntryOrConfig: entryOrConfig,
		Specs:         specs,
	}

	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", err
	}

	return builder.String(), nil
}
