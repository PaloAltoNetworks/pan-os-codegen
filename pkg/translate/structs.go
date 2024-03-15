package translate

import (
	"fmt"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"strings"
)

// LocationType function used in template location.tmpl to generate location type name.
func LocationType(location *properties.Location, pointer bool) string {
	prefix := ""
	if pointer {
		prefix = "*"
	}
	if location.Vars != nil {
		return fmt.Sprintf("%s%sLocation", prefix, location.Name.CamelCase)
	} else {
		return "bool"
	}
}

// NestedSpecs go through all params and one of (recursively) and return map of all nested specs.
func NestedSpecs(spec *properties.Spec) (map[string]*properties.Spec, error) {
	nestedSpecs := make(map[string]*properties.Spec)

	checkNestedSpecs([]string{}, spec, nestedSpecs)

	return nestedSpecs, nil
}

func checkNestedSpecs(parent []string, spec *properties.Spec, nestedSpecs map[string]*properties.Spec) {
	for _, param := range spec.Params {
		updateNestedSpecs(append(parent, param.Name.CamelCase), param, nestedSpecs)
	}
	for _, param := range spec.OneOf {
		updateNestedSpecs(append(parent, param.Name.CamelCase), param, nestedSpecs)
	}
}

func updateNestedSpecs(parent []string, param *properties.SpecParam, nestedSpecs map[string]*properties.Spec) {
	if param.Spec != nil {
		nestedSpecs[strings.Join(parent, "")] = param.Spec
		checkNestedSpecs(parent, param.Spec, nestedSpecs)
	}
}

// SpecParamType return param type (it can be nested spec) (for struct based on spec from YAML files).
func SpecParamType(parent string, param *properties.SpecParam) string {
	prefix := determinePrefix(param)

	calculatedType := ""
	if param.Type == "list" && param.Items != nil {
		calculatedType = determineListType(param)
	} else if param.Spec != nil {
		calculatedType = calculateNestedSpecType(parent, param)
	} else {
		calculatedType = param.Type
	}

	return fmt.Sprintf("%s%s", prefix, calculatedType)
}

// XmlParamType return param type (it can be nested spec) (for struct based on spec from YAML files).
func XmlParamType(parent string, param *properties.SpecParam) string {
	prefix := determinePrefix(param)

	calculatedType := ""
	if isParamListAndProfileTypeIsMember(param) {
		calculatedType = "util.MemberType"
	} else if param.Spec != nil {
		calculatedType = calculateNestedXmlSpecType(parent, param)
	} else {
		calculatedType = param.Type
	}

	return fmt.Sprintf("%s%s", prefix, calculatedType)
}

func determinePrefix(param *properties.SpecParam) string {
	prefix := ""
	if param.Type == "list" {
		prefix = prefix + "[]"
	}
	if !param.Required {
		prefix = prefix + "*"
	}
	return prefix
}

func determineListType(param *properties.SpecParam) string {
	if param.Items.Type == "object" && param.Items.Ref != nil {
		return "string"
	}
	return param.Items.Type
}

func calculateNestedSpecType(parent string, param *properties.SpecParam) string {
	return fmt.Sprintf("Spec%s%s", parent, naming.CamelCase("", param.Name.CamelCase, "", true))
}

func calculateNestedXmlSpecType(parent string, param *properties.SpecParam) string {
	return fmt.Sprintf("Spec%s%sXml", parent, naming.CamelCase("", param.Name.CamelCase, "", true))
}

// XmlTag creates a string with xml tag (e.g. `xml:"description,omitempty"`).
func XmlTag(param *properties.SpecParam) string {
	if param.Profiles != nil && len(param.Profiles) > 0 {
		suffix := ""
		if !param.Required {
			suffix = ",omitempty"
		}

		return fmt.Sprintf("`xml:\"%s%s\"`", param.Profiles[0].Xpath[len(param.Profiles[0].Xpath)-1], suffix)
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
