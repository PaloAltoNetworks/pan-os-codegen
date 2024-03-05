package translate

import (
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"sort"
	"strings"
)

func StructsDefinitionsForLocation(locations map[string]*properties.Location) (string, error) {
	keys := make([]string, 0, len(locations))
	for key := range locations {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var builder strings.Builder
	builder.WriteString("type Location struct {\n")
	for _, name := range keys {
		switch name {
		case "shared":
			builder.WriteString("\tShared       bool                 `json:\"shared\"`\n")
		case "from_panorama":
			builder.WriteString("\tFromPanorama bool                 `json:\"from_panorama\"`\n")
		case "vsys":
			builder.WriteString("\tVsys         *VsysLocation        `json:\"vsys,omitempty\"`\n")
		case "device_group":
			builder.WriteString("\tDeviceGroup  *DeviceGroupLocation `json:\"device_group,omitempty\"`\n")
		}
	}
	builder.WriteString("}\n\n")

	nestedStructsDefinitionsForLocation(locations, "vsys", "VsysLocation", &builder)
	nestedStructsDefinitionsForLocation(locations, "device_group", "DeviceGroupLocation", &builder)

	return builder.String(), nil
}

func nestedStructsDefinitionsForLocation(locations map[string]*properties.Location, locationName string, structName string, builder *strings.Builder) {
	if _, ok := locations[locationName]; ok {
		keys := make([]string, 0, len(locations[locationName].Vars))
		for key := range locations[locationName].Vars {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		var namesOriginal, namesCamelCaseWithSpaces []string
		for _, name := range keys {
			namesOriginal = append(namesOriginal, name)
			namesCamelCaseWithSpaces = append(namesCamelCaseWithSpaces, naming.CamelCase("", name, "", true))
		}
		namesCamelCaseWithSpaces = MakeIndentationEqual(namesCamelCaseWithSpaces)

		builder.WriteString("type " + structName + " struct {\n")
		for idx, name := range namesCamelCaseWithSpaces {
			builder.WriteString("\t" + name + " string  `json:\"" + namesOriginal[idx] + "\"`\n")
		}
		builder.WriteString("}\n\n")
	}
}
