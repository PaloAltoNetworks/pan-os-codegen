package translate

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

// LocationType function used in template location.tmpl to generate location type name.
func LocationType(location *properties.Location, pointer bool) string {
	prefix := ""
	if pointer {
		prefix = "*"
	}
	return fmt.Sprintf("%s%sLocation", prefix, location.Name.CamelCase)
}

type importXpathVariableSpec struct {
	Name        *properties.NameVariant
	Description string
	Default     *string
}

type importLocationSpec struct {
	Name              *properties.NameVariant
	XpathElements     []string
	XpathVariables    []importXpathVariableSpec
	XpathFinalElement string
	ResponseXpath     string
}

type importSpec struct {
	Variant   *properties.NameVariant
	Type      *properties.NameVariant
	Locations []importLocationSpec
}

func createImportLocationSpecsForLocation(location properties.ImportLocation) importLocationSpec {
	var variables []importXpathVariableSpec
	variablesByName := make(map[string]importXpathVariableSpec, len(location.XpathVariables))

	for _, elt := range location.OrderedXpathVariables() {
		var defaultValue *string
		if elt.Default != "" {
			defaultValue = &elt.Default
		}
		variableSpec := importXpathVariableSpec{
			Name:        elt.Name,
			Description: elt.Description,
			Default:     defaultValue,
		}

		variables = append(variables, variableSpec)
		variablesByName[elt.Name.Underscore] = variableSpec
	}

	var elements []string
	for _, elt := range location.XpathElements {
		if strings.HasPrefix(elt, "$") {
			variableName := elt[1:]
			asEntryXpath := fmt.Sprintf("util.AsEntryXpath(o.%s)", variablesByName[variableName].Name.LowerCamelCase)
			elements = append(elements, asEntryXpath)
		} else {
			elements = append(elements, fmt.Sprintf("\"%s\"", elt))
		}
	}

	xpathFinalElement := location.XpathElements[len(location.XpathElements)-1]
	responseXpath := fmt.Sprintf("result>%s>member", xpathFinalElement)

	return importLocationSpec{
		Name:              location.Name,
		XpathElements:     elements,
		XpathVariables:    variables,
		XpathFinalElement: xpathFinalElement,
		ResponseXpath:     responseXpath,
	}
}

func createImportSpecsForNormalization(spec *properties.Normalization) []importSpec {
	var specs []importSpec

	for _, imp := range spec.Imports {
		var locations []importLocationSpec
		for _, location := range imp.OrderedLocations() {
			locations = append(locations, createImportLocationSpecsForLocation(*location))
		}

		specs = append(specs, importSpec{
			Variant:   imp.Variant,
			Type:      imp.Type,
			Locations: locations,
		})
	}

	return specs
}

// RenderEntryImportStructs generates import location structs for a normalization spec.
func RenderEntryImportStructs(spec *properties.Normalization) (string, error) {
	type renderContext struct {
		Specs []importSpec
	}

	tmplContent, err := loadTemplate("partials/import_location_struct.tmpl")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("render-entry-import-structs").Parse(tmplContent)
	if err != nil {
		return "", err
	}

	data := renderContext{
		Specs: createImportSpecsForNormalization(spec),
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

type locationVariableSpec struct {
	Name           *properties.NameVariant
	LocationFilter bool
}

type locationSpec struct {
	Name      *properties.NameVariant
	Variables []locationVariableSpec
	HasFilter bool
}

func createLocationVariableSpecForLocation(loc *properties.Location) []locationVariableSpec {
	var variables []locationVariableSpec

	for _, elt := range loc.OrderedVars() {
		variables = append(variables, locationVariableSpec{
			Name:           elt.Name,
			LocationFilter: elt.LocationFilter,
		})
	}

	return variables
}

func createLocationSpecForNormalization(spec *properties.Normalization) []locationSpec {
	var locations []locationSpec
	for _, elt := range spec.OrderedLocations() {
		locations = append(locations, locationSpec{
			Name:      elt.Name,
			HasFilter: elt.HasFilter(),
			Variables: createLocationVariableSpecForLocation(elt),
		})
	}

	return locations
}

// RenderLocationFilter generates location filter structs for a normalization spec.
func RenderLocationFilter(spec *properties.Normalization) (string, error) {
	type renderContext struct {
		Specs []locationSpec
	}

	tmplContent, err := loadTemplate("partials/location_filter.tmpl")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("render-entry-import-structs").Parse(tmplContent)
	if err != nil {
		return "", err
	}

	data := renderContext{
		Specs: createLocationSpecForNormalization(spec),
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
