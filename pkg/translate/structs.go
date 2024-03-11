package translate

import (
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

func LocationType(location *properties.Location, pointer bool) string {
	prefix := ""
	if pointer {
		prefix = "*"
	}
	if location.Vars != nil {
		return prefix + location.Name.CamelCase + "Location"
	} else {
		return "bool"
	}
}

func SpecParamType(param *properties.SpecParam) string {
	prefix := ""
	if !param.Required {
		prefix = "*"
	}
	if param.Type == "list" {
		prefix = "[]"
	}

	calculatedType := ""
	if param.Type == "list" && param.Items != nil {
		calculatedType = param.Items.Type
	} else {
		calculatedType = param.Type
	}

	return prefix + calculatedType
}

func XmlParamType(param *properties.SpecParam) string {
	prefix := ""
	if !param.Required {
		prefix = "*"
	}

	calculatedType := ""
	if param.Type == "list" && param.Profiles != nil && len(param.Profiles) > 0 && param.Profiles[0].Type == "member" {
		calculatedType = "util.MemberType"
	} else {
		calculatedType = param.Type
	}

	return prefix + calculatedType
}

func XmlTag(param *properties.SpecParam) string {
	suffix := ""
	if !param.Required {
		suffix = ",omitempty"
	}

	calculatedTag := ""
	if param.Profiles != nil && len(param.Profiles) > 0 {
		calculatedTag = "`xml:\"" + param.Profiles[0].Xpath[len(param.Profiles[0].Xpath)-1] + suffix + "\"`"
	}
	return calculatedTag
}

func OmitEmpty(location *properties.Location) string {
	if location.Vars != nil {
		return ",omitempty"
	} else {
		return ""
	}
}
