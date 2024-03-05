package translate

import (
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"sort"
	"strings"
)

func AsEntryXpath(location, xpath string) string {
	location = naming.CamelCase("", location, "", true)
	xpath = strings.TrimSpace(strings.Split(strings.Split(xpath, "$")[1], "}")[0])
	xpath = naming.CamelCase("", xpath, "", true)
	return "util.AsEntryXpath([]string{o." + location + "." + xpath + "}),"
}

func FuncBodyForLocation(locations map[string]*properties.Location) (string, error) {
	keys := make([]string, 0, len(locations))
	for key := range locations {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	mapIsValid := map[string]string{
		"shared": `
	case o.Shared:
		count++
`,

		"from_panorama": `
	case o.FromPanorama:
		count++
`,

		"vsys": `
	case o.Vsys != nil:
		if o.Vsys.Vsys == "" {
			return fmt.Errorf("vsys.vsys is unspecified")
		}
		if o.Vsys.NgfwDevice == "" {
			return fmt.Errorf("vsys.ngfw_device is unspecified")
		}
		count++
`,

		"device_group": `
	case o.DeviceGroup != nil:
		if o.DeviceGroup.DeviceGroup == "" {
			return fmt.Errorf("device_group.device_group is unspecified")
		}
		if o.DeviceGroup.PanoramaDevice == "" {
			return fmt.Errorf("device_group.panorama_device is unspecified")
		}
		count++
`,
	}
	mapXpath := map[string]string{
		"shared": `
	case o.Shared:
		ans = []string{
			"config",
			"shared",
		}
`,

		"from_panorama": `
	case o.FromPanorama:
		ans = []string{"config", "panorama"}
`,

		"vsys": `
	case o.Vsys != nil:
		if o.Vsys.NgfwDevice == "" {
			return nil, fmt.Errorf("NgfwDevice is unspecified")
		}
		if o.Vsys.Vsys == "" {
			return nil, fmt.Errorf("Vsys is unspecified")
		}
		ans = []string{
			"config",
			"devices",
			util.AsEntryXpath([]string{o.Vsys.NgfwDevice}),
			"vsys",
			util.AsEntryXpath([]string{o.Vsys.Vsys}),
		}
`,

		"device_group": `
	case o.DeviceGroup != nil:
		if o.DeviceGroup.PanoramaDevice == "" {
			return nil, fmt.Errorf("PanoramaDevice is unspecified")
		}
		if o.DeviceGroup.DeviceGroup == "" {
			return nil, fmt.Errorf("DeviceGroup is unspecified")
		}
		ans = []string{
			"config",
			"devices",
			util.AsEntryXpath([]string{o.DeviceGroup.PanoramaDevice}),
			"device-group",
			util.AsEntryXpath([]string{o.DeviceGroup.DeviceGroup}),
		}
`,
	}

	var builder strings.Builder
	funcBodyForLocation(&builder, keys, "IsValid", "()", "error", false,
		`	count := 0

	switch {
`,
		`	}
	
	if count == 0 {
		return fmt.Errorf("no path specified")
	}

	if count > 1 {
		return fmt.Errorf("multiple paths specified: only one should be specified")
	}

	return nil
`,
		mapIsValid)
	funcBodyForLocation(&builder, keys, "Xpath", "(vn version.Number, name string)", "([]string, error)",
		true, `
	var ans []string

	switch {
`,
		`
	default:
		return nil, errors.NoLocationSpecifiedError
	}

	ans = append(ans, Suffix...)
	ans = append(ans, util.AsEntryXpath([]string{name}))

	return ans, nil
`,
		mapXpath)

	return builder.String(), nil
}

func funcBodyForLocation(builder *strings.Builder, keys []string, funcName string, funcInput string, funcOutput string,
	startFromNewLine bool, funcBegin string, funcEnd string, funcCases map[string]string) {
	if startFromNewLine {
		builder.WriteString("\n")
	}
	builder.WriteString("func (o Location) " + funcName + funcInput + " " + funcOutput + " {\n")
	builder.WriteString(funcBegin)
	for _, name := range keys {
		builder.WriteString(funcCases[name])
	}
	builder.WriteString(funcEnd)
	builder.WriteString("}\n")
}
