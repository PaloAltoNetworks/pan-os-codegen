package translate

import (
	"fmt"
	"strings"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/version"
)

// SpecParamType returns param type (it can be a nested spec) for structs based on spec from YAML files.
func SpecParamType(parent string, param *properties.SpecParam) string {
	prefix := determinePrefix(param, false)

	calculatedType := ""
	if param.Spec != nil {
		calculatedType = calculateNestedSpecType(parent, param)
	} else if param.Type == "list" && param.Items != nil {
		calculatedType = determineListType(param)
	} else {
		calculatedType = param.Type
	}

	return fmt.Sprintf("%s%s", prefix, calculatedType)
}

// ParamType return param type (it can be nested spec) (for struct based on spec from YAML files).
func ParamType(structTyp structType, parentName *properties.NameVariant, param *properties.SpecParam, suffix string) string {
	var calculatedType string
	if param.Type == "" || isParamListAndProfileTypeIsExtendedEntry(param) {
		typ := calculateNestedXmlSpecType(structTyp, parentName, param, suffix)
		if structTyp == structXmlType {
			calculatedType = typ.LowerCamelCase
		} else {
			calculatedType = typ.CamelCase
		}
	} else if isParamListAndProfileTypeIsMember(param) {
		if structTyp == structXmlType {
			calculatedType = "util.Member"
		} else {
			calculatedType = param.Items.Type
		}
	} else if isParamListAndProfileTypeIsSingleEntry(param) {
		if structTyp == structXmlType {
			calculatedType = "util.Entry"
		} else {
			calculatedType = calculateNestedXmlSpecType(structTyp, parentName, param, suffix).CamelCase
		}
	} else if param.Type == "bool" && structTyp == structXmlType {
		calculatedType = "string"
	} else {
		calculatedType = param.Type
	}

	return calculatedType
}

// XmlParamType returns the XML parameter type for a given parent and parameter.
func XmlParamType(parent string, param *properties.SpecParam) string {
	return ParamType(structXmlType, properties.NewNameVariant(parent), param, "")
}

func determinePrefix(param *properties.SpecParam, useMemberOrEntryTypeStruct bool) string {
	if param.Type == "list" {
		if useMemberOrEntryTypeStruct && (isParamListAndProfileTypeIsMember(param) || isParamListAndProfileTypeIsSingleEntry(param)) {
			return "*"
		} else {
			return "[]"
		}
	} else if !param.Required {
		return "*"
	}
	return ""
}

func determineListType(param *properties.SpecParam) string {
	if param.Items.Type == "object" && param.Items.Ref != nil {
		return "string"
	}
	return param.Items.Type
}

func calculateNestedSpecType(parent string, param *properties.SpecParam) string {
	return fmt.Sprintf("%s%s", parent, naming.CamelCase("", param.Name.CamelCase, "", true))
}

func calculateNestedXmlSpecType(structTyp structType, parentName *properties.NameVariant, param *properties.SpecParam, suffix string) *properties.NameVariant {
	var typ *properties.NameVariant
	if parentName.IsEmpty() {
		typ = param.PangoNameVariant()
	} else {
		typ = parentName.WithSuffix(param.PangoNameVariant())
	}

	if structTyp == structXmlType {
		typ = typ.WithSuffix(properties.NewNameVariant("xml")).WithLiteralSuffix(suffix)
	}

	return typ
}

// XmlName creates a string with xml name (e.g. `description`).
func XmlName(param *properties.SpecParam) string {
	if len(param.Profiles) > 0 {
		// FIXME: lists of objects have an extra "entry" element on their xpath
		if param.Type == "list" && param.Items.Type == "entry" {
			return param.Profiles[0].Xpath[0]
		}
		return strings.Join(param.Profiles[0].Xpath, ">")
	}

	return ""
}

// XmlTag creates a string with xml tag (e.g. `xml:"description,omitempty"`).
func XmlTag(param *properties.SpecParam) string {
	if len(param.Profiles) > 0 {
		suffix := ""

		if param.Name != nil && (param.Name.Underscore == "uuid" || param.Name.Underscore == "name") {
			suffix = suffix + ",attr"
		}

		if !param.Required {
			suffix = suffix + ",omitempty"
		}

		return fmt.Sprintf("`xml:\"%s%s\"`", XmlName(param), suffix)
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

// CreateGoSuffixFromVersionTmpl converts a version to a Go suffix for use in templates.
func CreateGoSuffixFromVersionTmpl(v any) (string, error) {
	if v != nil {
		typed, ok := v.(version.Version)
		if !ok {
			return "", fmt.Errorf("Failed to cast version to *version.Version: '%T'", v)
		}
		return CreateGoSuffixFromVersion(&typed), nil
	}

	return "", nil
}

// CreateGoSuffixFromVersion convert version into Go suffix e.g. 10.1.1 into _10_1_1
func CreateGoSuffixFromVersion(v *version.Version) string {
	if v == nil {
		return ""
	}

	return fmt.Sprintf("_%s", strings.ReplaceAll(v.String(), ".", "_"))
}

// ParamNotSkippedTmpl checks if a parameter should not be skipped (template version).
func ParamNotSkippedTmpl(param *properties.SpecParam) bool {
	if param.GoSdkConfig != nil && param.GoSdkConfig.Skip != nil {
		return !*param.GoSdkConfig.Skip
	}

	return true
}

// ParamSupportedInVersionTmpl checks if a parameter is supported in the given device version (template version).
func ParamSupportedInVersionTmpl(param *properties.SpecParam, deviceVersion any) (bool, error) {
	if deviceVersion == nil {
		return true, nil
	}

	typed, ok := deviceVersion.(version.Version)
	if !ok {
		return false, fmt.Errorf("Failed to cast deviceVersion to version.Version, received '%T'", deviceVersion)
	}

	return ParamSupportedInVersion(param, &typed), nil
}

// ParamSupportedInVersion checks if param is supported in specific PAN-OS version
func ParamSupportedInVersion(param *properties.SpecParam, deviceVersion *version.Version) bool {
	if deviceVersion == nil {
		return true
	}

	result := checkIfDeviceVersionSupportedByProfile(param, *deviceVersion)
	return result
}

func checkIfDeviceVersionSupportedByProfile(param *properties.SpecParam, deviceVersion version.Version) bool {
	for _, profile := range param.Profiles {
		if profile.MinVersion == nil && profile.MaxVersion == nil {
			return true
		}

		if deviceVersion.GreaterThanOrEqualTo(*profile.MinVersion) && deviceVersion.LesserThan(*profile.MaxVersion) {
			return true
		}
	}
	return false
}
