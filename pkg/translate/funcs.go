package translate

import (
	"fmt"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"strings"
)

// GenerateEntryXpathForLocation functions used in location.tmpl to generate XPath for location.
func GenerateEntryXpathForLocation(prefix, suffix, location, xpath string) (string, error) {
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

func prepareAssignment(objectType string, param *properties.SpecParam, listFunction, entryFunction, boolFunction, specPrefix, specSuffix string, version string) string {
	var builder strings.Builder

	if ParamSupportedInVersion(param, version) {
		if param.Spec != nil {
			if specSuffix == "Xml" {
				appendSpecObjectAssignment(param, nil, objectType, version, listFunction, entryFunction, boolFunction, specPrefix, specSuffix, &builder)
			} else {
				appendSpecObjectAssignment(param, nil, objectType, "", listFunction, entryFunction, boolFunction, specPrefix, specSuffix, &builder)
			}
		} else if isParamListAndProfileTypeIsMember(param) {
			appendFunctionAssignment(param, objectType, listFunction, "", &builder)
		} else if isParamListAndProfileTypeIsSingleEntry(param) {
			appendFunctionAssignment(param, objectType, entryFunction, "", &builder)
		} else if param.Type == "bool" {
			if specSuffix == "Xml" {
				appendFunctionAssignment(param, objectType, boolFunction, useBoolFunctionToConvertAdditionalArguments(param), &builder)
			} else {
				appendFunctionAssignment(param, objectType, boolFunction, "nil", &builder)
			}
		} else {
			appendSimpleAssignment(param, objectType, &builder)
		}
	}

	return builder.String()
}

func useBoolFunctionToConvertAdditionalArguments(param *properties.SpecParam) string {
	if param.Default != "" {
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
	defineNestedObject([]string{param.Name.CamelCase}, param, parentParam, objectType, version, listFunction, entryFunction, boolFunction, prefix, suffix, builder)
	builder.WriteString(fmt.Sprintf("%s.%s = nested%s\n", objectType, param.Name.CamelCase, param.Name.CamelCase))
}

func defineNestedObject(parent []string, param, parentParam *properties.SpecParam, objectType string, version, listFunction, entryFunction, boolFunction, prefix, suffix string, builder *strings.Builder) {
	declareRootOfNestedObject(parent, builder, version, prefix, suffix)

	if ParamSupportedInVersion(param, version) {
		if isParamListAndProfileTypeIsExtendedEntry(parentParam) {
			if param.Name.CamelCase == "Name" {
				builder.WriteString(fmt.Sprintf("if entryItem.%s != \"\" {\n", param.Name.CamelCase))
			} else {
				builder.WriteString(fmt.Sprintf("if entryItem.%s != nil {\n", param.Name.CamelCase))
			}
		} else {
			builder.WriteString(fmt.Sprintf("if o.%s != nil {\n", strings.Join(parent, ".")))
		}

		if param.Spec != nil {
			assignEmptyStructForNestedObject(parent, builder, param, objectType, version, prefix, suffix)
			defineNestedObjectForChildParams(parent, param.Spec.Params, param, objectType, version, listFunction, entryFunction, boolFunction, prefix, suffix, builder)
			defineNestedObjectForChildParams(parent, param.Spec.OneOf, param, objectType, version, listFunction, entryFunction, boolFunction, prefix, suffix, builder)
		} else if isParamListAndProfileTypeIsMember(param) {
			assignFunctionForNestedObject(parent, listFunction, "", builder, param, parentParam)
		} else if isParamListAndProfileTypeIsSingleEntry(param) {
			assignFunctionForNestedObject(parent, entryFunction, "", builder, param, parentParam)
		} else if param.Type == "bool" {
			if suffix == "Xml" {
				assignFunctionForNestedObject(parent, boolFunction, useBoolFunctionToConvertAdditionalArguments(param), builder, param, parentParam)
			} else {
				assignFunctionForNestedObject(parent, boolFunction, "nil", builder, param, parentParam)
			}
		} else {
			assignValueForNestedObject(parent, builder, param, parentParam)
		}
		if isParamListAndProfileTypeIsExtendedEntry(param) {
			builder.WriteString(fmt.Sprintf("nested%s = append(nested%s, nestedItem)\n",
				strings.Join(parent, "."), strings.Join(parent, ".")))
		}
		builder.WriteString("}\n")
	}
}

func declareRootOfNestedObject(parent []string, builder *strings.Builder, version, prefix, suffix string) {
	if len(parent) == 1 {
		builder.WriteString(fmt.Sprintf("var nested%s *%s%s%s%s\n",
			strings.Join(parent, "."), prefix,
			strings.Join(parent, ""), suffix,
			CreateGoSuffixFromVersion(version)))
	}
}

func assignEmptyStructForNestedObject(parent []string, builder *strings.Builder, param *properties.SpecParam, objectType, version, prefix, suffix string) {
	if isParamListAndProfileTypeIsExtendedEntry(param) {
		builder.WriteString(fmt.Sprintf("nested%s = []%s%s%s%s{}\n",
			strings.Join(parent, "."), prefix, strings.Join(parent, ""), suffix,
			CreateGoSuffixFromVersion(version)))

		builder.WriteString(fmt.Sprintf("for _, entryItem := range o.%s {\n",
			strings.Join(parent, ".")))
		builder.WriteString(fmt.Sprintf("nestedItem := %s%s%s%s{}\n",
			prefix, strings.Join(parent, ""), suffix,
			CreateGoSuffixFromVersion(version)))

		if suffix == "Xml" {
			builder.WriteString(fmt.Sprintf("if _, ok := o.Misc[\"%s\"]; ok {\n",
				strings.Join(parent, "")))
			builder.WriteString(fmt.Sprintf("nestedItem.Misc = o.Misc[\"%s\"]\n",
				strings.Join(parent, ""),
			))
		} else {
			builder.WriteString("if entryItem.Misc != nil {\n")
			builder.WriteString(fmt.Sprintf("entry.Misc[\"%s\"] = entryItem.Misc\n",
				strings.Join(parent, ""),
			))
		}
		builder.WriteString("}\n")
	} else {
		builder.WriteString(fmt.Sprintf("nested%s = &%s%s%s%s{}\n",
			strings.Join(parent, "."), prefix, strings.Join(parent, ""), suffix,
			CreateGoSuffixFromVersion(version)))

		if suffix == "Xml" {
			builder.WriteString(fmt.Sprintf("if _, ok := o.Misc[\"%s\"]; ok {\n",
				strings.Join(parent, "")))
			builder.WriteString(fmt.Sprintf("nested%s.Misc = o.Misc[\"%s\"]\n",
				strings.Join(parent, "."), strings.Join(parent, ""),
			))
		} else {
			builder.WriteString(fmt.Sprintf("if o.%s.Misc != nil {\n",
				strings.Join(parent, ".")))
			builder.WriteString(fmt.Sprintf("%s.Misc[\"%s\"] = o.%s.Misc\n",
				objectType, strings.Join(parent, ""), strings.Join(parent, "."),
			))
		}
		builder.WriteString("}\n")
	}
}

func assignValueForNestedObject(parent []string, builder *strings.Builder, param, parentParam *properties.SpecParam) {
	if isParamListAndProfileTypeIsExtendedEntry(parentParam) {
		builder.WriteString(fmt.Sprintf("nestedItem.%s = entryItem.%s\n",
			param.Name.CamelCase, param.Name.CamelCase))
	} else {
		builder.WriteString(fmt.Sprintf("nested%s = o.%s\n",
			strings.Join(parent, "."), strings.Join(parent, ".")))
	}
}

func assignFunctionForNestedObject(parent []string, functionName, additionalArguments string, builder *strings.Builder, param, parentParam *properties.SpecParam) {
	if isParamListAndProfileTypeIsExtendedEntry(parentParam) {
		if additionalArguments != "" {
			builder.WriteString(fmt.Sprintf("nestedItem.%s = %s(entryItem.%s, %s)\n",
				param.Name.CamelCase, functionName, param.Name.CamelCase, additionalArguments))
		} else {
			builder.WriteString(fmt.Sprintf("nestedItem.%s = %s(entryItem.%s)\n",
				param.Name.CamelCase, functionName, param.Name.CamelCase))
		}
	} else {
		if additionalArguments != "" {
			builder.WriteString(fmt.Sprintf("nested%s = %s(o.%s, %s)\n",
				strings.Join(parent, "."), functionName, strings.Join(parent, "."), additionalArguments))
		} else {
			builder.WriteString(fmt.Sprintf("nested%s = %s(o.%s)\n",
				strings.Join(parent, "."), functionName, strings.Join(parent, ".")))
		}
	}
}

func defineNestedObjectForChildParams(parent []string, params map[string]*properties.SpecParam, parentParam *properties.SpecParam, objectType string, version, listFunction, entryFunction, boolFunction, prefix, suffix string, builder *strings.Builder) {
	for _, param := range params {
		defineNestedObject(append(parent, param.Name.CamelCase), param, parentParam, objectType, version, listFunction, entryFunction, boolFunction, prefix, suffix, builder)
		if isParamListAndProfileTypeIsExtendedEntry(param) {
			builder.WriteString("}\n")
		}
	}
}

// SpecMatchesFunction return a string used in function SpecMatches() in entry.tmpl/config.tmpl
// to compare all items of generated entry.
func SpecMatchesFunction(param *properties.SpecParam) string {
	return specMatchFunctionName([]string{}, param)
}

func specMatchFunctionName(parent []string, param *properties.SpecParam) string {
	if param.Type == "list" && param.Items != nil && param.Items.Type == "string" {
		return "util.OrderedListsMatch"
	} else if param.Type == "string" && param.Name != nil && param.Name.CamelCase != "Name" {
		return "util.StringsMatch"
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

			renderSpecMatchesFunctionNameWithArguments(parent, builder, param)
			checkInSpecMatchesFunctionIfVariablesAreNil(builder)

			if isParamListAndProfileTypeIsExtendedEntry(param) {
				builder.WriteString(fmt.Sprintf("for _, a := range a {\n"))
				builder.WriteString(fmt.Sprintf("for _, b := range b {\n"))
				for _, subParam := range param.Spec.Params {
					renderInSpecMatchesFunctionIfToCheckIfVariablesMatches(parent, builder, param, subParam)
				}
				for _, subParam := range param.Spec.OneOf {
					renderInSpecMatchesFunctionIfToCheckIfVariablesMatches(parent, builder, param, subParam)
				}
				builder.WriteString("}\n")
				builder.WriteString("}\n")
			} else {
				for _, subParam := range param.Spec.Params {
					renderInSpecMatchesFunctionIfToCheckIfVariablesMatches(parent, builder, param, subParam)
				}
				for _, subParam := range param.Spec.OneOf {
					renderInSpecMatchesFunctionIfToCheckIfVariablesMatches(parent, builder, param, subParam)
				}
			}

			builder.WriteString("return true\n")
			builder.WriteString("}\n")
		} else if param.Type != "list" && (param.Type != "string" || param.Name.CamelCase == "Name") {
			// whole section should be removed, when there will be dedicated function to compare integers
			// in file https://github.com/PaloAltoNetworks/pango/blob/develop/util/comparison.go
			renderSpecMatchesFunctionNameWithArguments(parent, builder, param)

			if param.Name.CamelCase == "Name" {
				builder.WriteString("return a == b\n")
			} else {
				checkInSpecMatchesFunctionIfVariablesAreNil(builder)
				builder.WriteString("return *a == *b\n")
			}
			builder.WriteString("}\n")
		}
	}
}

func renderSpecMatchesFunctionNameWithArguments(parent []string, builder *strings.Builder, param *properties.SpecParam) {
	if param.Name.CamelCase == "Name" {
		builder.WriteString(fmt.Sprintf("func specMatch%s%s(a, b string) bool {",
			strings.Join(parent, ""), param.Name.CamelCase))
	} else {
		prefix := determinePrefix(param, false)
		builder.WriteString(fmt.Sprintf("func specMatch%s%s(a %s%s, b %s%s) bool {",
			strings.Join(parent, ""), param.Name.CamelCase,
			prefix, argumentTypeForSpecMatchesFunction(parent, param),
			prefix, argumentTypeForSpecMatchesFunction(parent, param)))
	}
}

func checkInSpecMatchesFunctionIfVariablesAreNil(builder *strings.Builder) {
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
