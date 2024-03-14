package translate

import (
	"fmt"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"strings"
)

// AsEntryXpath functions used in location.tmpl to generate XPath for location.
func AsEntryXpath(location, xpath string) (string, error) {
	if !strings.Contains(xpath, "$") || !strings.Contains(xpath, "}") {
		return "", fmt.Errorf("xpath '%s' is missing '$' followed by '}'", xpath)
	}
	asEntryXpath := generateXpathForLocation(location, xpath)
	return asEntryXpath, nil
}

func generateXpathForLocation(location string, xpath string) string {
	xpathPartWithoutDollar := strings.SplitAfter(xpath, "$")
	xpathPartWithoutBrackets := strings.TrimSpace(strings.Trim(xpathPartWithoutDollar[1], "${}"))
	xpathPartCamelCase := naming.CamelCase("", xpathPartWithoutBrackets, "", true)
	asEntryXpath := fmt.Sprintf("util.AsEntryXpath([]string{o.%s.%s}),", location, xpathPartCamelCase)
	return asEntryXpath
}

// NormalizeAssignment generates a string, which contains entry assignment in Normalize() function
// in entry.tmpl template. If param contains nested specs, then recursively are executed internal functions,
// which are declaring additional variables (function nestedObjectDeclaration()) and use them in
// entry assignment (function nestedObjectAssignment()).
func NormalizeAssignment(param *properties.SpecParam) string {
	return prepareAssignment(param, "util.MemToStr", "")
}

// SpecifyEntryAssignment generates a string, which contains entry assignment in SpecifyEntry() function
// in entry.tmpl template. If param contains nested specs, then recursively are executed internal functions,
// which are declaring additional variables (function nestedObjectDeclaration()) and use them in
// entry assignment (function nestedObjectAssignment()).
func SpecifyEntryAssignment(param *properties.SpecParam) string {
	return prepareAssignment(param, "util.StrToMem", "Xml")
}

func prepareAssignment(param *properties.SpecParam, listFunction, specSuffix string) string {
	var builder strings.Builder

	if param.Type == "list" && param.Profiles != nil && len(param.Profiles) > 0 && param.Profiles[0].Type == "member" {
		builder.WriteString(fmt.Sprintf("entry.%s = %s(o.%s)", param.Name.CamelCase, listFunction, param.Name.CamelCase))
	} else if param.Spec != nil {
		for _, subParam := range param.Spec.Params {
			builder.WriteString(nestedObjectDeclaration([]string{param.Name.CamelCase}, subParam))
		}
		for _, subParam := range param.Spec.OneOf {
			builder.WriteString(nestedObjectDeclaration([]string{param.Name.CamelCase}, subParam))
		}
		builder.WriteString(fmt.Sprintf("entry.%s = &Spec%s%s{\n", param.Name.CamelCase, param.Name.CamelCase, specSuffix))
		for _, subParam := range param.Spec.Params {
			builder.WriteString(nestedObjectAssignment([]string{param.Name.CamelCase}, specSuffix, subParam))
		}
		for _, subParam := range param.Spec.OneOf {
			builder.WriteString(nestedObjectAssignment([]string{param.Name.CamelCase}, specSuffix, subParam))
		}
		builder.WriteString("}\n")
	} else {
		builder.WriteString(fmt.Sprintf("entry.%s = o.%s", param.Name.CamelCase, param.Name.CamelCase))
	}

	return builder.String()
}

func nestedObjectDeclaration(parent []string, param *properties.SpecParam) string {
	var builder strings.Builder

	if param.Spec != nil {
		for _, subParam := range param.Spec.Params {
			builder.WriteString(declareVariableForNestedObject(parent, param, subParam))
			builder.WriteString(nestedObjectDeclaration(append(parent, param.Name.CamelCase), subParam))
		}
		for _, subParam := range param.Spec.OneOf {
			builder.WriteString(declareVariableForNestedObject(parent, param, subParam))
			builder.WriteString(nestedObjectDeclaration(append(parent, param.Name.CamelCase), subParam))
		}
	}

	return builder.String()
}

func declareVariableForNestedObject(parent []string, param *properties.SpecParam, subParam *properties.SpecParam) string {
	if subParam.Spec == nil && parent != nil {
		return fmt.Sprintf("nested%s%s%s := o.%s.%s.%s\n",
			strings.Join(parent, ""),
			param.Name.CamelCase,
			subParam.Name.CamelCase,
			strings.Join(parent, "."),
			param.Name.CamelCase,
			subParam.Name.CamelCase)
	} else {
		return ""
	}
}

func nestedObjectAssignment(parent []string, suffix string, param *properties.SpecParam) string {
	var builder strings.Builder

	if param.Type == "list" && param.Profiles != nil && len(param.Profiles) > 0 && param.Profiles[0].Type == "member" {
		builder.WriteString(fmt.Sprintf("%s : util.StrToMem(o.%s),\n",
			param.Name.CamelCase, param.Name.CamelCase))
	} else if param.Spec != nil {
		builder.WriteString(fmt.Sprintf("%s : &Spec%s%s%s{\n",
			param.Name.CamelCase, strings.Join(parent, ""), param.Name.CamelCase, suffix))
		for _, subParam := range param.Spec.Params {
			builder.WriteString(nestedObjectAssignment(append(parent, param.Name.CamelCase), suffix, subParam))
		}
		for _, subParam := range param.Spec.OneOf {
			builder.WriteString(nestedObjectAssignment(append(parent, param.Name.CamelCase), suffix, subParam))
		}
		builder.WriteString("},\n")
	} else {
		builder.WriteString(fmt.Sprintf("%s : nested%s%s,\n",
			param.Name.CamelCase, strings.Join(parent, ""), param.Name.CamelCase))
	}

	return builder.String()
}

// SpecMatchesFunction return a string used in function SpecMatches() in entry.tmpl
// to compare all items of generated entry.
func SpecMatchesFunction(param *properties.SpecParam) string {
	calculatedFunction := "OptionalStringsMatch"
	if param.Type == "list" {
		calculatedFunction = "OrderedListsMatch"
	}
	return calculatedFunction
}
