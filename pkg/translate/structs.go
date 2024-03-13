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
	prefix := ""
	if !param.Required {
		prefix = "*"
	}
	if param.Type == "list" {
		prefix = "[]"
	}

	calculatedType := ""
	if param.Type == "list" && param.Items != nil {
		if param.Items.Type == "object" && param.Items.Ref != nil {
			calculatedType = "string"
		} else {
			calculatedType = param.Items.Type
		}
	} else if param.Spec != nil {
		calculatedType = fmt.Sprintf("Spec%s%s", parent, naming.CamelCase("", param.Name.CamelCase, "", true))
	} else {
		calculatedType = param.Type
	}

	return prefix + calculatedType
}

// XmlParamType return param type (it can be nested spec) (for struct based on spec from YAML files).
func XmlParamType(parent string, param *properties.SpecParam) string {
	prefix := ""
	if !param.Required {
		prefix = "*"
	}

	calculatedType := ""
	if param.Type == "list" && param.Profiles != nil && len(param.Profiles) > 0 && param.Profiles[0].Type == "member" {
		calculatedType = "util.MemberType"
	} else if param.Spec != nil {
		calculatedType = fmt.Sprintf("Spec%s%sXml", parent, naming.CamelCase("", param.Name.CamelCase, "", true))
	} else {
		calculatedType = param.Type
	}

	return prefix + calculatedType
}

// XmlTag creates a string with xml tag (e.g. `xml:"description,omitempty"`).
func XmlTag(param *properties.SpecParam) string {
	suffix := ""
	if !param.Required {
		suffix = ",omitempty"
	}

	calculatedTag := ""
	if param.Profiles != nil && len(param.Profiles) > 0 {
		calculatedTag = fmt.Sprintf("`xml:\"%s%s\"`", param.Profiles[0].Xpath[len(param.Profiles[0].Xpath)-1], suffix)
	}
	return calculatedTag
}

// OmitEmpty return omitempty in XML tag for location, if there are variables defined.
func OmitEmpty(location *properties.Location) string {
	if location.Vars != nil {
		return ",omitempty"
	} else {
		return ""
	}
}
