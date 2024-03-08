package translate

import (
	"errors"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
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
