package translate

import (
	"errors"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"strings"
)

func AsEntryXpath(location, xpath string) (string, error) {
	if !strings.Contains(xpath, "$") || !strings.Contains(xpath, "}") {
		return "", errors.New("$ followed by } should exists in xpath'")
	}
	xpath = strings.TrimSpace(strings.Split(strings.Split(xpath, "$")[1], "}")[0])
	xpath = naming.CamelCase("", xpath, "", true)
	return "util.AsEntryXpath([]string{o." + location + "." + xpath + "}),", nil
}

func SpecifyEntryAssignment(param *properties.SpecParam) string {
	var builder strings.Builder

	if param.Type == "list" && param.Profiles != nil && len(param.Profiles) > 0 && param.Profiles[0].Type == "member" {
		builder.WriteString("entry." + param.Name.CamelCase + " = util.StrToMem(o." + param.Name.CamelCase + ")")
	} else if param.Spec != nil {
		for _, subParam := range param.Spec.Params {
			builder.WriteString(nestedObjectDeclaration([]string{param.Name.CamelCase}, subParam))
		}
		builder.WriteString("entry." + param.Name.CamelCase + " = &Spec" + param.Name.CamelCase + "Xml{\n")
		for _, subParam := range param.Spec.Params {
			builder.WriteString(nestedObjectAssignment([]string{param.Name.CamelCase}, "Xml", subParam))
		}
		for _, subParam := range param.Spec.OneOf {
			builder.WriteString(nestedObjectAssignment([]string{param.Name.CamelCase}, "Xml", subParam))
		}
		builder.WriteString("}\n")
	} else {
		builder.WriteString("entry." + param.Name.CamelCase + " = o." + param.Name.CamelCase)
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
		return "nested" +
			strings.Join(parent, "") +
			param.Name.CamelCase +
			subParam.Name.CamelCase +
			" := o." +
			strings.Join(parent, ".") +
			"." + param.Name.CamelCase +
			"." + subParam.Name.CamelCase +
			"\n"
	} else {
		return ""
	}
}

func nestedObjectAssignment(parent []string, suffix string, param *properties.SpecParam) string {
	var builder strings.Builder

	if param.Type == "list" && param.Profiles != nil && len(param.Profiles) > 0 && param.Profiles[0].Type == "member" {
		builder.WriteString(param.Name.CamelCase +
			" : util.StrToMem(o." +
			param.Name.CamelCase +
			"),\n")
	} else if param.Spec != nil {
		builder.WriteString(param.Name.CamelCase +
			" : &Spec" +
			param.Name.CamelCase +
			suffix + "{\n")
		for _, subParam := range param.Spec.Params {
			builder.WriteString(nestedObjectAssignment(append(parent, param.Name.CamelCase), suffix, subParam))
		}
		for _, subParam := range param.Spec.OneOf {
			builder.WriteString(nestedObjectAssignment(append(parent, param.Name.CamelCase), suffix, subParam))
		}
		builder.WriteString("},\n")
	} else {
		builder.WriteString(param.Name.CamelCase +
			" : nested" +
			strings.Join(parent, "") +
			param.Name.CamelCase + ",\n")
	}

	return builder.String()
}

func NormalizeAssignment(param *properties.SpecParam) string {
	var builder strings.Builder

	if param.Type == "list" && param.Profiles != nil && len(param.Profiles) > 0 && param.Profiles[0].Type == "member" {
		builder.WriteString("entry." + param.Name.CamelCase + " = entryXml." + param.Name.CamelCase)
	} else if param.Spec != nil {
		for _, subParam := range param.Spec.Params {
			builder.WriteString(nestedObjectDeclaration([]string{param.Name.CamelCase}, subParam))
		}
		builder.WriteString("entry." + param.Name.CamelCase + " = &Spec" + param.Name.CamelCase + "{\n")
		for _, subParam := range param.Spec.Params {
			builder.WriteString(nestedObjectAssignment([]string{param.Name.CamelCase}, "", subParam))
		}
		for _, subParam := range param.Spec.OneOf {
			builder.WriteString(nestedObjectAssignment([]string{param.Name.CamelCase}, "", subParam))
		}
		builder.WriteString("}\n")
	} else {
		builder.WriteString("entry." + param.Name.CamelCase + " = entryXml." + param.Name.CamelCase)
	}

	return builder.String()
}

func SpecMatchesFunction(param *properties.SpecParam) string {
	calculatedFunction := "OptionalStringsMatch"
	if param.Type == "list" {
		calculatedFunction = "OrderedListsMatch"
	}
	return calculatedFunction
}
