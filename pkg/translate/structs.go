package translate

import (
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
		var vars []string
		for name := range locations["vsys"].Vars {
			vars = append(vars, name)
		}
		vars = MakeIndentationEqual(vars)

		builder.WriteString("type VsysLocation struct {\n")
		for _, name := range vars {
			builder.WriteString("\t" + name + "\tstring  `json:\"" + name + "\"`\n")
		}
		builder.WriteString("}\n\n")
	}

	if _, ok := locations["device_group"]; ok {
		var vars []string
		for name := range locations["device_group"].Vars {
			vars = append(vars, name)
		}
		vars = MakeIndentationEqual(vars)

		builder.WriteString("type DeviceGroupLocation struct {\n")
		for _, name := range vars {
			builder.WriteString("\t" + name + "\tstring  `json:\"" + strings.TrimSpace(name) + "\"`\n")
		}
		builder.WriteString("}\n\n")
	}

	return builder.String(), nil
}
