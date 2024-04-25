package translate

import (
	"fmt"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"strings"
)

// GenerateEntryXpath functions used in location.tmpl to generate XPath for location.
func GenerateEntryXpath(prefix, suffix, location, xpath string) (string, error) {
	if !strings.Contains(xpath, "$") || !strings.Contains(xpath, "}") {
		return "", fmt.Errorf("xpath '%s' is missing '$' followed by '}'", xpath)
	}
	asEntryXpath := generateEntryXpathForLocation(prefix, suffix, location, xpath)
	return asEntryXpath, nil
}

func generateEntryXpathForLocation(prefix, suffix, location, xpath string) string {
	xpathPartWithoutDollar := strings.SplitAfter(xpath, "$")
	xpathPartWithoutBrackets := strings.TrimSpace(strings.Trim(xpathPartWithoutDollar[1], "${}"))
	xpathPartCamelCase := naming.CamelCase("", xpathPartWithoutBrackets, "", true)
	asEntryXpath := fmt.Sprintf("%so.%s.%s%s,", prefix, location, xpathPartCamelCase, suffix)
	return asEntryXpath
}

// NormalizeAssignment generates a string, which contains entry/config assignment in Normalize() function
// in entry.tmpl/config.tmpl template. If param contains nested specs, then recursively are executed
// internal functions, which are creating entry assignment.
func NormalizeAssignment(objectType string, param *properties.SpecParam, version string) string {
	return prepareAssignment(objectType, param, "util.MemToStr", "util.EntToStr", "util.AsBool", "Spec", "", version)
}

// SpecifyEntryAssignment generates a string, which contains entry/config assignment in SpecifyEntry() function
// in entry.tmpl/config.tmpl template. If param contains nested specs, then recursively are executed
// internal functions, which are creating entry assignment.
func SpecifyEntryAssignment(objectType string, param *properties.SpecParam, version string) string {
	return prepareAssignment(objectType, param, "util.StrToMem", "util.StrToEnt", "util.YesNo", "spec", "Xml", version)
}

func prepareAssignment(objectType string, param *properties.SpecParam, listFunction, entryFunction, boolFunction, prefix, suffix string, version string) string {
	var builder strings.Builder

	if ParamSupportedInVersion(param, version) {
		switch {
		case param.Spec != nil:
			appendSpecObjectAssignment(param, nil, objectType, paramVersionInAssignment(suffix, version),
				listFunction, entryFunction, boolFunction, prefix, suffix, &builder)
		case isParamListAndProfileTypeIsMember(param):
			appendFunctionAssignment(param, objectType, listFunction, "", &builder)
		case isParamListAndProfileTypeIsSingleEntry(param):
			appendFunctionAssignment(param, objectType, entryFunction, "", &builder)
		case param.Type == "bool":
			appendFunctionAssignment(param, objectType, boolFunction,
				useBoolFunctionForAdditionalArguments(suffix, param), &builder)
		default:
			appendSimpleAssignment(param, objectType, &builder)
		}
	}

	return builder.String()
}

func paramVersionInAssignment(suffix, version string) string {
	if suffix == "Xml" {
		return version
	}
	return ""
}

func useBoolFunctionForAdditionalArguments(suffix string, param *properties.SpecParam) string {
	if suffix == "Xml" && param.Default != "" {
		return fmt.Sprintf("util.Bool(%s)", param.Default)
	}
	return "nil"
}

func isParamListAndProfileTypeIsMember(param *properties.SpecParam) bool {
	return param.Type == "list" && param.Profiles != nil && len(param.Profiles) > 0 && param.Profiles[0].Type == "member"
}

func isParamListAndProfileTypeIsSingleEntry(param *properties.SpecParam) bool {
	return param.Type == "list" && param.Profiles != nil && len(param.Profiles) > 0 && param.Profiles[0].Type == "entry" && param.Items != nil && param.Items.Type == "string"
}

func isParamListAndProfileTypeIsExtendedEntry(param *properties.SpecParam) bool {
	return param != nil && param.Type == "list" && param.Profiles != nil && len(param.Profiles) > 0 && param.Profiles[0].Type == "entry" && param.Items != nil && param.Items.Type != "string"
}

func appendSimpleAssignment(param *properties.SpecParam, objectType string, builder *strings.Builder) {
	builder.WriteString(fmt.Sprintf("%s.%s = o.%s", objectType, param.Name.CamelCase, param.Name.CamelCase))
}

func appendFunctionAssignment(param *properties.SpecParam, objectType string, functionName, additionalArguments string, builder *strings.Builder) {
	if additionalArguments != "" {
		builder.WriteString(fmt.Sprintf("%s.%s = %s(o.%s, %s)", objectType, param.Name.CamelCase, functionName, param.Name.CamelCase, additionalArguments))
	} else {
		builder.WriteString(fmt.Sprintf("%s.%s = %s(o.%s)", objectType, param.Name.CamelCase, functionName, param.Name.CamelCase))
	}
}

func appendSpecObjectAssignment(param, parentParam *properties.SpecParam, objectType string, version, listFunction, entryFunction, boolFunction, prefix, suffix string, builder *strings.Builder) {
	defineNestedObject([]*properties.SpecParam{param}, param, parentParam, objectType, version, listFunction, entryFunction, boolFunction, prefix, suffix, builder)
	builder.WriteString(fmt.Sprintf("%s.%s = nested%s\n", objectType, param.Name.CamelCase, param.Name.CamelCase))
}

func defineNestedObject(parent []*properties.SpecParam, param, parentParam *properties.SpecParam, objectType string, version, listFunction, entryFunction, boolFunction, prefix, suffix string, builder *strings.Builder) {
	declareRootOfNestedObject(parent, builder, version, prefix, suffix)

	if ParamSupportedInVersion(param, version) {
		startIfBlockForParamNotNil(parent, param, parentParam, builder)

		switch {
		case param.Spec != nil:
			assignEmptyStructForNestedObject(parent, builder, param, objectType, version, prefix, suffix)
			defineNestedObjectForChildParams(parent, param.Spec.Params, param, objectType, version, listFunction, entryFunction, boolFunction, prefix, suffix, builder)
			defineNestedObjectForChildParams(parent, param.Spec.OneOf, param, objectType, version, listFunction, entryFunction, boolFunction, prefix, suffix, builder)
		case isParamListAndProfileTypeIsMember(param):
			assignFunctionForNestedObject(parent, listFunction, "", builder, param, parentParam)
		case isParamListAndProfileTypeIsSingleEntry(param):
			assignFunctionForNestedObject(parent, entryFunction, "", builder, param, parentParam)
		case param.Type == "bool":
			assignFunctionForNestedObject(parent, boolFunction,
				useBoolFunctionForAdditionalArguments(suffix, param), builder, param, parentParam)
		default:
			assignValueForNestedObject(parent, builder, param, parentParam)
		}

		finishNestedObjectIfBlock(parent, param, builder)
	}
}

func startIfBlockForParamNotNil(parent []*properties.SpecParam, param *properties.SpecParam, parentParam *properties.SpecParam, builder *strings.Builder) {
	if isParamListAndProfileTypeIsExtendedEntry(parentParam) {
		if isParamName(param) {
			builder.WriteString(fmt.Sprintf("if o%s.%s != \"\" {\n",
				renderNestedVariableName(parent[:len(parent)-1], false, false, false), param.Name.CamelCase))
		} else {
			builder.WriteString(fmt.Sprintf("if o%s.%s != nil {\n",
				renderNestedVariableName(parent[:len(parent)-1], false, false, false), param.Name.CamelCase))
		}
	} else {
		builder.WriteString(fmt.Sprintf("if o.%s != nil {\n", renderNestedVariableName(parent, true, false, false)))
	}
}

func finishNestedObjectIfBlock(parent []*properties.SpecParam, param *properties.SpecParam, builder *strings.Builder) {
	if isParamListAndProfileTypeIsExtendedEntry(param) {
		builder.WriteString(fmt.Sprintf("nested%s = append(nested%s, nested%s)\n",
			renderNestedVariableName(parent, true, false, false),
			renderNestedVariableName(parent, true, false, false),
			renderNestedVariableName(parent, false, false, false)))
	}
	builder.WriteString("}\n")
}

func isParamName(param *properties.SpecParam) bool {
	return param.Name.CamelCase == "Name"
}

func declareRootOfNestedObject(parent []*properties.SpecParam, builder *strings.Builder, version, prefix, suffix string) {
	if len(parent) == 1 {
		builder.WriteString(fmt.Sprintf("var nested%s *%s%s%s%s\n",
			renderNestedVariableName(parent, true, true, false), prefix,
			renderNestedVariableName(parent, false, false, false), suffix,
			CreateGoSuffixFromVersion(version)))
	}
}

func assignEmptyStructForNestedObject(parent []*properties.SpecParam, builder *strings.Builder, param *properties.SpecParam, objectType, version, prefix, suffix string) {
	if isParamListAndProfileTypeIsExtendedEntry(param) {
		createListAndLoopForNestedEntry(parent, builder, prefix, suffix, version)
		miscForUnknownXmlWithExtendedEntry(parent, builder, suffix)
	} else {
		createStructForParamWithSpec(parent, builder, prefix, suffix, version)
		miscForUnknownXmlWithSpec(parent, builder, suffix, objectType)
	}
	builder.WriteString("}\n")
}

func createStructForParamWithSpec(parent []*properties.SpecParam, builder *strings.Builder, prefix string, suffix string, version string) {
	builder.WriteString(fmt.Sprintf("nested%s = &%s%s%s%s{}\n",
		renderNestedVariableName(parent, true, true, false), prefix,
		renderNestedVariableName(parent, false, false, false), suffix,
		CreateGoSuffixFromVersion(version)))
}

func createListAndLoopForNestedEntry(parent []*properties.SpecParam, builder *strings.Builder, prefix string, suffix string, version string) {
	builder.WriteString(fmt.Sprintf("nested%s = []%s%s%s%s{}\n",
		renderNestedVariableName(parent, true, false, false), prefix,
		renderNestedVariableName(parent, false, false, false), suffix,
		CreateGoSuffixFromVersion(version)))

	builder.WriteString(fmt.Sprintf("for _, o%s := range o.%s {\n",
		renderNestedVariableName(parent, false, false, false),
		renderNestedVariableName(parent, true, false, false)))
	builder.WriteString(fmt.Sprintf("nested%s := %s%s%s%s{}\n",
		renderNestedVariableName(parent, false, false, false),
		prefix, renderNestedVariableName(parent, false, false, false), suffix,
		CreateGoSuffixFromVersion(version)))
}

func miscForUnknownXmlWithSpec(parent []*properties.SpecParam, builder *strings.Builder, suffix string, objectType string) {
	if suffix == "Xml" {
		builder.WriteString(fmt.Sprintf("if _, ok := o.Misc[\"%s\"]; ok {\n",
			renderNestedVariableName(parent, false, false, false)))
		builder.WriteString(fmt.Sprintf("nested%s.Misc = o.Misc[\"%s\"]\n",
			renderNestedVariableName(parent, true, true, false),
			renderNestedVariableName(parent, false, false, false),
		))
	} else {
		builder.WriteString(fmt.Sprintf("if o.%s.Misc != nil {\n",
			renderNestedVariableName(parent, true, true, false)))
		builder.WriteString(fmt.Sprintf("%s.Misc[\"%s\"] = o.%s.Misc\n",
			objectType, renderNestedVariableName(parent, false, false, false),
			renderNestedVariableName(parent, true, true, false),
		))
	}
}

func miscForUnknownXmlWithExtendedEntry(parent []*properties.SpecParam, builder *strings.Builder, suffix string) {
	if suffix == "Xml" {
		builder.WriteString(fmt.Sprintf("if _, ok := o.Misc[\"%s\"]; ok {\n",
			renderNestedVariableName(parent, false, false, false)))
		builder.WriteString(fmt.Sprintf("nested%s.Misc = o.Misc[\"%s\"]\n",
			renderNestedVariableName(parent, false, false, false),
			renderNestedVariableName(parent, false, false, false),
		))
	} else {
		builder.WriteString(fmt.Sprintf("if o%s.Misc != nil {\n",
			renderNestedVariableName(parent, false, false, false)))
		builder.WriteString(fmt.Sprintf("entry.Misc[\"%s\"] = o%s.Misc\n",
			renderNestedVariableName(parent, false, false, false),
			renderNestedVariableName(parent, false, false, false),
		))
	}
}

func assignValueForNestedObject(parent []*properties.SpecParam, builder *strings.Builder, param, parentParam *properties.SpecParam) {
	if isParamListAndProfileTypeIsExtendedEntry(parentParam) {
		builder.WriteString(fmt.Sprintf("nested%s.%s = o%s.%s\n",
			renderNestedVariableName(parent[:len(parent)-1], false, false, false), param.Name.CamelCase,
			renderNestedVariableName(parent[:len(parent)-1], false, false, false), param.Name.CamelCase))
	} else {
		builder.WriteString(fmt.Sprintf("nested%s = o.%s\n",
			renderNestedVariableName(parent, true, true, false),
			renderNestedVariableName(parent, true, true, false)))
	}
}

func assignFunctionForNestedObject(parent []*properties.SpecParam, functionName, additionalArguments string, builder *strings.Builder, param, parentParam *properties.SpecParam) {
	if isParamListAndProfileTypeIsExtendedEntry(parentParam) {
		if additionalArguments != "" {
			builder.WriteString(fmt.Sprintf("nested%s.%s = %s(o%s.%s, %s)\n",
				renderNestedVariableName(parent[:len(parent)-1], false, false, false), param.Name.CamelCase, functionName,
				renderNestedVariableName(parent[:len(parent)-1], false, false, false), param.Name.CamelCase, additionalArguments))
		} else {
			builder.WriteString(fmt.Sprintf("nested%s.%s = %s(o%s.%s)\n",
				renderNestedVariableName(parent[:len(parent)-1], false, false, false), param.Name.CamelCase, functionName,
				renderNestedVariableName(parent[:len(parent)-1], false, false, false), param.Name.CamelCase))
		}
	} else {
		if additionalArguments != "" {
			builder.WriteString(fmt.Sprintf("nested%s = %s(o.%s, %s)\n",
				renderNestedVariableName(parent, true, true, false), functionName,
				renderNestedVariableName(parent, true, true, false), additionalArguments))
		} else {
			builder.WriteString(fmt.Sprintf("nested%s = %s(o.%s)\n",
				renderNestedVariableName(parent, true, true, false), functionName,
				renderNestedVariableName(parent, true, true, false)))
		}
	}
}

func defineNestedObjectForChildParams(parent []*properties.SpecParam, params map[string]*properties.SpecParam, parentParam *properties.SpecParam, objectType string, version, listFunction, entryFunction, boolFunction, prefix, suffix string, builder *strings.Builder) {
	for _, param := range params {
		defineNestedObject(append(parent, param), param, parentParam, objectType, version, listFunction, entryFunction, boolFunction, prefix, suffix, builder)
		if isParamListAndProfileTypeIsExtendedEntry(param) {
			builder.WriteString("}\n")
		}
	}
}

func renderNestedVariableName(params []*properties.SpecParam, useDot, searchForParamWithEntry, startFromDot bool) string {
	var builder strings.Builder

	indexOfLastParamWithExtendedEntry := 0

	if searchForParamWithEntry {
		for i := len(params) - 1; i >= 0; i-- {
			if isParamListAndProfileTypeIsExtendedEntry(params[i]) {
				indexOfLastParamWithExtendedEntry = i
				break
			}
		}
	}

	for i, param := range params {
		if useDot && startFromDot && i == 0 && (!searchForParamWithEntry ||
			searchForParamWithEntry && i >= indexOfLastParamWithExtendedEntry) {
			builder.WriteString(".")
		}
		builder.WriteString(param.Name.CamelCase)
		if useDot && i < len(params)-1 && (!searchForParamWithEntry ||
			searchForParamWithEntry && i >= indexOfLastParamWithExtendedEntry) {
			builder.WriteString(".")
		}
	}

	return builder.String()
}

// SpecMatchesFunction return a string used in function SpecMatches() in entry.tmpl/config.tmpl
// to compare all items of generated entry.
func SpecMatchesFunction(param *properties.SpecParam) string {
	return specMatchFunctionName([]string{}, param)
}

func specMatchFunctionName(parent []string, param *properties.SpecParam) string {
	if param.Type == "list" && param.Items != nil && param.Items.Type == "string" {
		return "util.OrderedListsMatch"
	} else if param.Type == "string" {
		if param.Name != nil && param.Name.CamelCase == "Name" {
			return "util.StringsEqual"
		} else {
			return "util.StringsMatch"
		}
	} else if param.Type == "bool" {
		return "util.BoolsMatch"
	} else if param.Type == "int" {
		return "util.IntsMatch"
	} else if param.Type == "int64" {
		return "util.Ints64Match"
	} else {
		return fmt.Sprintf("specMatch%s%s", strings.Join(parent, ""), param.Name.CamelCase)
	}
}

// NestedSpecMatchesFunction return a string with body of specMach* functions required for nested params
func NestedSpecMatchesFunction(spec *properties.Spec) string {
	var builder strings.Builder

	defineSpecMatchesFunction([]string{}, spec.Params, &builder)
	defineSpecMatchesFunction([]string{}, spec.OneOf, &builder)

	return builder.String()
}

func defineSpecMatchesFunction(parent []string, params map[string]*properties.SpecParam, builder *strings.Builder) {
	for _, param := range params {
		if param.Spec != nil {
			defineSpecMatchesFunction(append(parent, param.Name.CamelCase), param.Spec.Params, builder)
			defineSpecMatchesFunction(append(parent, param.Name.CamelCase), param.Spec.OneOf, builder)

			renderSpecMatchesFunctionSignature(parent, builder, param)
			checkIfVariablesAreNil(builder)

			if isParamListAndProfileTypeIsExtendedEntry(param) {
				renderSpecMatchBodyForExtendedEntry(parent, builder, param)
			} else {
				renderSpecMatchBodyForTypicalParam(parent, param, builder)
			}

			builder.WriteString("return true\n")
			builder.WriteString("}\n")
		}
	}
}

func renderSpecMatchesFunctionSignature(parent []string, builder *strings.Builder, param *properties.SpecParam) {
	prefix := determinePrefix(param, false)
	builder.WriteString(fmt.Sprintf("func specMatch%s%s(a %s%s, b %s%s) bool {",
		strings.Join(parent, ""), param.Name.CamelCase,
		prefix, argumentTypeForSpecMatchesFunction(parent, param),
		prefix, argumentTypeForSpecMatchesFunction(parent, param)))
}

func argumentTypeForSpecMatchesFunction(parent []string, param *properties.SpecParam) string {
	if param.Type == "bool" {
		return "bool"
	} else if param.Type == "int" {
		return "int"
	} else {
		return fmt.Sprintf("Spec%s%s",
			strings.Join(parent, ""), param.Name.CamelCase)
	}
}

func checkIfVariablesAreNil(builder *strings.Builder) {
	builder.WriteString("if a == nil && b != nil || a != nil && b == nil {\n")
	builder.WriteString("	return false\n")
	builder.WriteString("} else if a == nil && b == nil {\n")
	builder.WriteString("	return true\n")
	builder.WriteString("}\n")
}

func renderInSpecMatchesFunctionIfToCheckIfVariablesMatches(parent []string, builder *strings.Builder, param *properties.SpecParam, subparam *properties.SpecParam) {
	builder.WriteString(fmt.Sprintf("if !%s(a.%s, b.%s) {\n",
		specMatchFunctionName(append(parent, param.Name.CamelCase), subparam), subparam.Name.CamelCase, subparam.Name.CamelCase))
	builder.WriteString("	return false\n")
	builder.WriteString("}\n")
}

func renderSpecMatchBodyForTypicalParam(parent []string, param *properties.SpecParam, builder *strings.Builder) {
	for _, subParam := range param.Spec.Params {
		renderInSpecMatchesFunctionIfToCheckIfVariablesMatches(parent, builder, param, subParam)
	}
	for _, subParam := range param.Spec.OneOf {
		renderInSpecMatchesFunctionIfToCheckIfVariablesMatches(parent, builder, param, subParam)
	}
}

func renderSpecMatchBodyForExtendedEntry(parent []string, builder *strings.Builder, param *properties.SpecParam) {
	builder.WriteString("for _, a := range a {\n")
	builder.WriteString("for _, b := range b {\n")
	for _, subParam := range param.Spec.Params {
		renderInSpecMatchesFunctionIfToCheckIfVariablesMatches(parent, builder, param, subParam)
	}
	for _, subParam := range param.Spec.OneOf {
		renderInSpecMatchesFunctionIfToCheckIfVariablesMatches(parent, builder, param, subParam)
	}
	builder.WriteString("}\n")
	builder.WriteString("}\n")
}

// XmlPathSuffixes return XML path suffixes created from profiles.
func XmlPathSuffixes(param *properties.SpecParam) []string {
	xmlPathSuffixes := []string{}
	if param.Profiles != nil {
		for _, profile := range param.Profiles {
			xmlPathSuffixes = append(xmlPathSuffixes, strings.Join(profile.Xpath, "/"))
		}
	}
	return xmlPathSuffixes
}
