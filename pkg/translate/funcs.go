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
	calculatedAssignment := ""
	if param.Type == "list" && param.Profiles != nil && len(param.Profiles) > 0 && param.Profiles[0].Type == "member" {
		calculatedAssignment = "util.StrToMem(o." + param.Name.CamelCase + ")"
	} else {
		calculatedAssignment = "o." + param.Name.CamelCase
	}

	return calculatedAssignment
}

func SpecMatchesFunction(param *properties.SpecParam) string {
	calculatedFunction := "OptionalStringsMatch"
	if param.Type == "list" {
		calculatedFunction = "OrderedListsMatch"
	}
	return calculatedFunction
}
