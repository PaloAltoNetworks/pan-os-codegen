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
func NormalizeAssignment(objectType string, param *properties.SpecParam) string {
	return prepareAssignment(objectType, param, "util.MemToStr", "")
}

// SpecifyEntryAssignment generates a string, which contains entry/config assignment in SpecifyEntry() function
// in entry.tmpl/config.tmpl template. If param contains nested specs, then recursively are executed
// internal functions, which are creating entry assignment.
func SpecifyEntryAssignment(objectType string, param *properties.SpecParam) string {
	return prepareAssignment(objectType, param, "util.StrToMem", "Xml")
}

func prepareAssignment(objectType string, param *properties.SpecParam, listFunction, specSuffix string) string {
	var builder strings.Builder

	if param.Spec != nil {
		appendSpecObjectAssignment(param, objectType, specSuffix, &builder)
	} else if isParamListAndProfileTypeIsMember(param) {
		appendListFunctionAssignment(param, objectType, listFunction, &builder)
	} else {
		appendSimpleAssignment(param, objectType, &builder)
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

func appendSpecObjectAssignment(param *properties.SpecParam, objectType string, suffix string, builder *strings.Builder) {
	builder.WriteString(fmt.Sprintf("%s.%s = &Spec%s%s{\n", objectType, param.Name.CamelCase, param.Name.CamelCase, suffix))

	appendNestedObjectAssignment([]string{param.Name.CamelCase}, param.Spec.Params, suffix, builder)
	appendNestedObjectAssignment([]string{param.Name.CamelCase}, param.Spec.OneOf, suffix, builder)

	builder.WriteString("}\n")
}

func appendNestedObjectAssignment(parent []string, params map[string]*properties.SpecParam, suffix string, builder *strings.Builder) {
	for _, subParam := range params {
		appendAssignmentForNestedObject(parent, subParam, suffix, builder)
	}
}

func appendAssignmentForNestedObject(parent []string, param *properties.SpecParam, suffix string, builder *strings.Builder) {
	if param.Spec != nil {
		builder.WriteString(fmt.Sprintf("%s : &Spec%s%s%s{\n", param.Name.CamelCase,
			strings.Join(parent, ""), param.Name.CamelCase, suffix))
		appendNestedObjectAssignment(append(parent, param.Name.CamelCase), param.Spec.Params, suffix, builder)
		appendNestedObjectAssignment(append(parent, param.Name.CamelCase), param.Spec.OneOf, suffix, builder)
		builder.WriteString("},\n")
	} else if isParamListAndProfileTypeIsMember(param) {
		builder.WriteString(fmt.Sprintf("%s : util.StrToMem(o.%s),\n",
			param.Name.CamelCase, param.Name.CamelCase))
	} else {
		builder.WriteString(fmt.Sprintf("%s : o.%s.%s,\n",
			param.Name.CamelCase, strings.Join(parent, "."), param.Name.CamelCase))
	}
}

// SpecMatchesFunction return a string used in function SpecMatches() in entry.tmpl/config.tmpl
// to compare all items of generated entry.
func SpecMatchesFunction(param *properties.SpecParam) string {
	if param.Type == "list" {
		return "OrderedListsMatch"
	}
	return "OptionalStringsMatch"
}
