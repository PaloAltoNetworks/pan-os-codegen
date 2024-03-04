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

	if _, ok := locations["vsys"]; ok {
		var namesOriginal, namesCamelCaseWithSpaces []string
		for name := range locations["vsys"].Vars {
			namesOriginal = append(namesOriginal, name)
			namesCamelCaseWithSpaces = append(namesCamelCaseWithSpaces, naming.CamelCase("", name, "", true))
		}
		namesCamelCaseWithSpaces = MakeIndentationEqual(namesCamelCaseWithSpaces)

		builder.WriteString("type VsysLocation struct {\n")
		for idx, name := range namesCamelCaseWithSpaces {
			builder.WriteString("\t" + name + " string  `json:\"" + namesOriginal[idx] + "\"`\n")
		}
		builder.WriteString("}\n\n")
	}

	if _, ok := locations["device_group"]; ok {
		var namesOriginal, namesCamelCaseWithSpaces []string
		for name := range locations["device_group"].Vars {
			namesOriginal = append(namesOriginal, name)
			namesCamelCaseWithSpaces = append(namesCamelCaseWithSpaces, naming.CamelCase("", name, "", true))
		}
		namesCamelCaseWithSpaces = MakeIndentationEqual(namesCamelCaseWithSpaces)

		builder.WriteString("type DeviceGroupLocation struct {\n")
		for idx, name := range namesCamelCaseWithSpaces {
			builder.WriteString("\t" + name + " string  `json:\"" + namesOriginal[idx] + "\"`\n")
		}
		builder.WriteString("}\n\n")
	}

	return builder.String(), nil
}
