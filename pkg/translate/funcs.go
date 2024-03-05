package translate

import (
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"strings"
)

func AsEntryXpath(location, xpath string) string {
	location = naming.CamelCase("", location, "", true)
	xpath = strings.TrimSpace(strings.Split(strings.Split(xpath, "$")[1], "}")[0])
	xpath = naming.CamelCase("", xpath, "", true)
	return "util.AsEntryXpath([]string{o." + location + "." + xpath + "}),"
}
