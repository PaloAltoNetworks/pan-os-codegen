package translate

import (
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"strings"
)

func StructsDefinitionsForLocation(locations map[string]*properties.Location) (string, error) {
	var builder strings.Builder

	builder.WriteString("type Location struct {\n")
	for name := range locations {
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
		var namesOriginal, namesCamelCaseWithSpaces []string
		for name := range locations[locationName].Vars {
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
