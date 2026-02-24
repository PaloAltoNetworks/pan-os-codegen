package translate

import (
	"bytes"
	"fmt"
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
