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

func OmitEmpty(location *properties.Location) string {
	if location.Vars != nil {
		return ",omitempty"
	} else {
		return ""
	}
}
