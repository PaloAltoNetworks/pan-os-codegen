package translate

import (
	"fmt"
	"strings"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
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
		return fmt.Sprintf("match%s%s", strings.Join(parent, ""), param.Name.CamelCase)
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
	builder.WriteString(fmt.Sprintf("func match%s%s(a %s%s, b %s%s) bool {",
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
		return fmt.Sprintf("%s%s",
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
