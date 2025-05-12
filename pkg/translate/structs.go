package translate

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"text/template"
	"unicode"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/version"
)

// LocationType function used in template location.tmpl to generate location type name.
func LocationType(location *properties.Location, pointer bool) string {
	prefix := ""
	if pointer {
		prefix = "*"
	}
	return fmt.Sprintf("%s%sLocation", prefix, location.Name.CamelCase)
}

// NestedSpecs goes through all params and one ofs (recursively) and returns map of all nested specs.
func NestedSpecs(spec *properties.Spec) (map[string]NestedSpec, error) {
	nestedSpecs := make(map[string]NestedSpec)

	fmt.Println("HELLO1")
	if spec == nil {
		panic("spec == nil")
	}

	checkNestedSpecs([]string{}, spec, nestedSpecs)
	fmt.Println("END1")

	return nestedSpecs, nil
}

type NestedSpec struct {
	ParentIsList bool
	Spec         *properties.Spec
}

func checkNestedSpecs(parent []string, spec *properties.Spec, nestedSpecs map[string]NestedSpec) {
	for _, param := range spec.SortedParams() {
		paramKey := append(parent, param.Name.CamelCase)
		updateNestedSpecs(paramKey, param, nestedSpecs)
		if len(param.Profiles) > 0 && param.Profiles[0].Type == "entry" && param.Items != nil && param.Items.Type == "entry" {
			nested, modified := addNameAsParamForNestedSpec(paramKey, nestedSpecs)
			nested.ParentIsList = true
			if modified {
				nested.Spec.HackFixInjectedNameSpecOrder()
			}
		}
	}
	for _, param := range spec.SortedOneOf() {
		paramKey := append(parent, param.Name.CamelCase)
		updateNestedSpecs(paramKey, param, nestedSpecs)
		if len(param.Profiles) > 0 && param.Profiles[0].Type == "entry" && param.Items != nil && param.Items.Type == "entry" {
			nested, modified := addNameAsParamForNestedSpec(paramKey, nestedSpecs)
			nested.ParentIsList = true
			if modified {
				nested.Spec.HackFixInjectedNameSpecOrder()
			}
		}
	}
}

func updateNestedSpecs(parent []string, param *properties.SpecParam, nestedSpecs map[string]NestedSpec) {
	if param.Spec != nil {
		nestedSpecs[strings.Join(parent, "")] = NestedSpec{
			Spec: param.Spec,
		}

		checkNestedSpecs(parent, param.Spec, nestedSpecs)
	}
}

func addNameAsParamForNestedSpec(parent []string, nestedSpecs map[string]NestedSpec) (*NestedSpec, bool) {
	nested := nestedSpecs[strings.Join(parent, "")]
	if _, found := nested.Spec.Params["name"]; found {
		return &nested, false
	}

	nested.Spec.Params["name"] = &properties.SpecParam{
		Name: &properties.NameVariant{
			Underscore: "name",
			CamelCase:  "Name",
		},
		SpecOrder: 0,
		Type:      "string",
		Required:  true,
		Profiles: []*properties.SpecParamProfile{
			{
				Xpath: []string{"name"},
			},
		},
	}

	return &nested, true
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
			asEntryXpath := fmt.Sprintf("util.AsEntryXpath([]string{o.%s})", variablesByName[variableName].Name.LowerCamelCase)
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

// XmlParamType return param type (it can be nested spec) (for struct based on spec from YAML files).
func XmlParamType(parent string, param *properties.SpecParam) string {
	prefix := determinePrefix(param, true)

	calculatedType := ""
	if param.Spec != nil {
		calculatedType = calculateNestedXmlSpecType(parent, param)
	} else if isParamListAndProfileTypeIsMember(param) {
		calculatedType = "util.MemberType"
	} else if isParamListAndProfileTypeIsSingleEntry(param) {
		calculatedType = "util.EntryType"
	} else if param.Type == "bool" {
		calculatedType = "string"
	} else {
		calculatedType = param.Type
	}

	return fmt.Sprintf("%s%s", prefix, calculatedType)
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

func calculateNestedXmlSpecType(parent string, param *properties.SpecParam) string {
	xmlType := fmt.Sprintf("%s%sXml", parent, naming.CamelCase("", param.Name.CamelCase, "", true))
	r := []rune(xmlType)
	r[0] = unicode.ToLower(r[0])

	return string(r)
}

// XmlName creates a string with xml name (e.g. `description`).
func XmlName(param *properties.SpecParam) string {
	if len(param.Profiles) > 0 {
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

		log.Printf("Param: %s, Version: %s, MinVersion: %s, MaxVersion: %s", param.Name.CamelCase, deviceVersion, profile.MinVersion.String(), profile.MaxVersion.String())

		if deviceVersion.GreaterThanOrEqualTo(*profile.MinVersion) && deviceVersion.LesserThan(*profile.MaxVersion) {
			return true
		}
	}
	return false
}
