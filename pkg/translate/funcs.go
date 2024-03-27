package translate

import (
	"fmt"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"strings"
)

// GenerateEntryXpathForLocation functions used in location.tmpl to generate XPath for location.
func GenerateEntryXpathForLocation(location, xpath string) (string, error) {
	if !strings.Contains(xpath, "$") || !strings.Contains(xpath, "}") {
		return "", fmt.Errorf("xpath '%s' is missing '$' followed by '}'", xpath)
	}
	asEntryXpath := generateEntryXpathForLocation(location, xpath)
	return asEntryXpath, nil
}

func generateEntryXpathForLocation(location string, xpath string) string {
	xpathPartWithoutDollar := strings.SplitAfter(xpath, "$")
	xpathPartWithoutBrackets := strings.TrimSpace(strings.Trim(xpathPartWithoutDollar[1], "${}"))
	xpathPartCamelCase := naming.CamelCase("", xpathPartWithoutBrackets, "", true)
	asEntryXpath := fmt.Sprintf("util.AsEntryXpath([]string{o.%s.%s}),", location, xpathPartCamelCase)
	return asEntryXpath
}

// NormalizeAssignment generates a string, which contains entry/config assignment in Normalize() function
// in entry.tmpl/config.tmpl template. If param contains nested specs, then recursively are executed
// internal functions, which are creating entry assignment.
func NormalizeAssignment(objectType string, param *properties.SpecParam, version string) string {
	return prepareAssignment(objectType, param, "util.MemToStr", "Spec", "", version)
}

// SpecifyEntryAssignment generates a string, which contains entry/config assignment in SpecifyEntry() function
// in entry.tmpl/config.tmpl template. If param contains nested specs, then recursively are executed
// internal functions, which are creating entry assignment.
func SpecifyEntryAssignment(objectType string, param *properties.SpecParam, version string) string {
	return prepareAssignment(objectType, param, "util.StrToMem", "spec", "Xml", version)
}

func prepareAssignment(objectType string, param *properties.SpecParam, listFunction, specPrefix, specSuffix string, version string) string {
	var builder strings.Builder

	if ParamSupportedInVersion(param, version) {
		if param.Spec != nil {
			if specSuffix == "Xml" {
				appendSpecObjectAssignment(param, objectType, version, specPrefix, specSuffix, &builder)
			} else {
				appendSpecObjectAssignment(param, objectType, "", specPrefix, specSuffix, &builder)
			}
		} else if isParamListAndProfileTypeIsMember(param) {
			appendListFunctionAssignment(param, objectType, listFunction, &builder)
		} else {
			appendSimpleAssignment(param, objectType, &builder)
		}
	}

	return builder.String()
}

func isParamListAndProfileTypeIsMember(param *properties.SpecParam) bool {
	return param.Type == "list" && param.Profiles != nil && len(param.Profiles) > 0 && param.Profiles[0].Type == "member"
}

func appendSimpleAssignment(param *properties.SpecParam, objectType string, builder *strings.Builder) {
	builder.WriteString(fmt.Sprintf("%s.%s = o.%s", objectType, param.Name.CamelCase, param.Name.CamelCase))
}

func appendListFunctionAssignment(param *properties.SpecParam, objectType string, listFunction string, builder *strings.Builder) {
	builder.WriteString(fmt.Sprintf("%s.%s = %s(o.%s)", objectType, param.Name.CamelCase, listFunction, param.Name.CamelCase))
}

func appendSpecObjectAssignment(param *properties.SpecParam, objectType string, version, prefix, suffix string, builder *strings.Builder) {
	defineNestedObject([]string{param.Name.CamelCase}, param, objectType, version, prefix, suffix, builder)
	builder.WriteString(fmt.Sprintf("%s.%s = nested%s\n", objectType, param.Name.CamelCase, param.Name.CamelCase))
}

func defineNestedObject(parent []string, param *properties.SpecParam, objectType string, version, prefix, suffix string, builder *strings.Builder) {
	declareRootOfNestedObject(parent, builder, version, prefix, suffix)

	builder.WriteString(fmt.Sprintf("if o.%s != nil {\n", strings.Join(parent, ".")))
	if param.Spec != nil {
		assignEmptyStructForNestedObject(parent, builder, version, prefix, suffix)
		defineNestedObjectForChildParams(parent, param.Spec.Params, objectType, version, prefix, suffix, builder)
		defineNestedObjectForChildParams(parent, param.Spec.OneOf, objectType, version, prefix, suffix, builder)
	} else {
		assignValueForNestedObject(parent, builder)
	}
	builder.WriteString("}\n")
}

func declareRootOfNestedObject(parent []string, builder *strings.Builder, version, prefix, suffix string) {
	if len(parent) == 1 {
		builder.WriteString(fmt.Sprintf("nested%s := &%s%s%s%s{}\n",
			strings.Join(parent, "."), prefix,
			strings.Join(parent, ""), suffix,
			CreateGoSuffixFromVersion(version)))
	}
}

func assignEmptyStructForNestedObject(parent []string, builder *strings.Builder, version, prefix, suffix string) {
	if len(parent) > 1 {
		builder.WriteString(fmt.Sprintf("nested%s = &%s%s%s%s{}\n",
			strings.Join(parent, "."), prefix, strings.Join(parent, ""), suffix,
			CreateGoSuffixFromVersion(version)))

		builder.WriteString(fmt.Sprintf("if o.%s.Misc != nil {\n",
			strings.Join(parent, ".")))
		if suffix == "Xml" {
			builder.WriteString(fmt.Sprintf("nested%s.Misc = o.%s.Misc[\"%s\"]\n",
				strings.Join(parent, "."), strings.Join(parent, "."), strings.Join(parent, ""),
			))
		} else {
			builder.WriteString(fmt.Sprintf("nested%s.Misc[\"%s\"] = o.%s.Misc\n",
				strings.Join(parent, "."), strings.Join(parent, ""), strings.Join(parent, "."),
			))
		}
		builder.WriteString("}\n")
	}
}

func assignValueForNestedObject(parent []string, builder *strings.Builder) {
	builder.WriteString(fmt.Sprintf("nested%s = o.%s\n",
		strings.Join(parent, "."),
		strings.Join(parent, ".")))
}

func defineNestedObjectForChildParams(parent []string, params map[string]*properties.SpecParam, objectType string, version, prefix, suffix string, builder *strings.Builder) {
	for _, param := range params {
		defineNestedObject(append(parent, param.Name.CamelCase), param, objectType, version, prefix, suffix, builder)
	}
}

// SpecMatchesFunction return a string used in function SpecMatches() in entry.tmpl/config.tmpl
// to compare all items of generated entry.
func SpecMatchesFunction(param *properties.SpecParam) string {
	return specMatchFunctionName([]string{}, param)
}

func specMatchFunctionName(parent []string, param *properties.SpecParam) string {
	if param.Type == "list" {
		return "util.OrderedListsMatch"
	} else if param.Type == "string" {
		return "util.OptionalStringsMatch"
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

			for _, subParam := range param.Spec.Params {
				renderInSpecMatchesFunctionIfToCheckIfVariablesMatches(parent, builder, param, subParam)
			}
			for _, subParam := range param.Spec.OneOf {
				renderInSpecMatchesFunctionIfToCheckIfVariablesMatches(parent, builder, param, subParam)
			}

			builder.WriteString("return true\n")
			builder.WriteString("}\n")
		} else if param.Type != "list" && param.Type != "string" {
			renderSpecMatchesFunctionNameWithArguments(parent, builder, param)
			checkInSpecMatchesFunctionIfVariablesAreNil(builder)

			builder.WriteString("return *a == *b\n")
			builder.WriteString("}\n")
		}
	}
}

func renderSpecMatchesFunctionNameWithArguments(parent []string, builder *strings.Builder, param *properties.SpecParam) {
	builder.WriteString(fmt.Sprintf("func specMatch%s%s(a *%s, b *%s) bool {",
		strings.Join(parent, ""), param.Name.CamelCase,
		argumentTypeForSpecMatchesFunction(parent, param),
		argumentTypeForSpecMatchesFunction(parent, param)))
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
