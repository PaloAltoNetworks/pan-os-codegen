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

func OmitEmpty(location *properties.Location) string {
	if location.Vars != nil {
		return ",omitempty"
	} else {
		return ""
	}
}
