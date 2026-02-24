package terraform_provider

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

// Entry describes a configuration entry.
type Entry struct {
	Name string
	Type string
}

// EntryData holds data for entry template rendering.
type EntryData struct {
	EntryName string
	Entries   []Entry
}

// pascalCase converts a string to PascalCase.
func pascalCase(value string) string {
	var parts []string
	if strings.Contains(value, "-") {
		parts = strings.Split(value, "-")
	} else if strings.Contains(value, "_") {
		parts = strings.Split(value, "_")
	} else {
		parts = []string{value}
	}

	caser := cases.Title(language.English)

	var result []string
	for _, elt := range parts {
		result = append(result, caser.String(elt))
	}

	return strings.Join(result, "")
}

// ConfigEntry generates configuration entry code.
func ConfigEntry(entryName string, param *properties.SpecParam) (string, error) {
	var entries []Entry

	paramType := param.Type
	if paramType == "" {
		paramType = "object"
	}
	entries = append(entries, Entry{
		Name: naming.CamelCase("", entryName, "", true),
		Type: paramType,
	})

	log.Printf("entries: %v", entries)

	entryData := EntryData{
		EntryName: entryName,
		Entries:   entries,
	}

	return processTemplate("resource/config_entry.tmpl", "config-entry", entryData, nil)
}

// RenderResourceFuncMap generates the resource function map for the provider.
func RenderResourceFuncMap(names map[string]properties.TerraformProviderSpecMetadata) (string, error) {
	type entry struct {
		Key        string
		StructName string
	}

	type context struct {
		Entries []entry
	}

	var entries []entry

	keys := make([]string, 0, len(names))
	for elt := range names {
		keys = append(keys, elt)
	}

	sort.Strings(keys)

	for _, key := range keys {
		if key == "" {
			continue
		}

		metadata := names[key]

		if metadata.Flags&properties.TerraformSpecImportable == 0 {
			continue
		}

		entries = append(entries, entry{
			Key:        fmt.Sprintf("panos%s", key),
			StructName: metadata.StructName,
		})
	}
	data := context{
		Entries: entries,
	}

	return processTemplate("provider/resource_func_map.tmpl", "resource-func-map", data, nil)
}
