package translate

import (
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

// RenderToXmlMarshalers generates toXml() marshaling methods for a normalization spec.
func RenderToXmlMarshalers(spec *properties.Normalization) (string, error) {
	tmplContent, err := loadTemplate("partials/struct_to_xml_marshalers.tmpl")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("render-to-xml-marsrhallers").Parse(tmplContent)
	if err != nil {
		return "", err
	}

	specs := createStructSpecs(structXmlType, spec, nil)
	for _, elt := range spec.SupportedVersionRanges() {
		specs = append(specs, createStructSpecs(structXmlType, spec, &elt.Minimum)...)
	}
	type context struct {
		EntryOrConfig string
		Specs         []entryStructContext
	}

	entryOrConfig := "Entry"
	if spec.TerraformProviderConfig.ResourceType == properties.TerraformResourceConfig {
		entryOrConfig = "Config"
	}

	data := context{
		EntryOrConfig: entryOrConfig,
		Specs:         specs,
	}

	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", err
	}

	return builder.String(), nil
}

// RenderXmlContainerNormalizers generates XML container normalizer functions for a normalization spec.
func RenderXmlContainerNormalizers(spec *properties.Normalization) (string, error) {
	tmplContent, err := loadTemplate("partials/xml_container_normalizers.tmpl")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("render-xml-container-normalizers").Parse(tmplContent)
	if err != nil {
		return "", err
	}

	specs := createStructSpecs(structXmlType, spec, nil)
	for _, elt := range spec.SupportedVersionRanges() {
		specs = append(specs, createStructSpecs(structXmlType, spec, &elt.Minimum)...)
	}
	type context struct {
		EntryOrConfig string
		Specs         []entryStructContext
	}

	entryOrConfig := "Entry"
	if spec.TerraformProviderConfig.ResourceType == properties.TerraformResourceConfig {
		entryOrConfig = "Config"
	}

	data := context{
		EntryOrConfig: entryOrConfig,
		Specs:         specs,
	}

	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", err
	}

	return builder.String(), nil
}

// RenderXmlContainerSpecifiers generates XML container specifier functions for a normalization spec.
func RenderXmlContainerSpecifiers(spec *properties.Normalization) (string, error) {
	tmplContent, err := loadTemplate("partials/xml_container_specifiers.tmpl")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("render-xml-container-specifiers").Parse(tmplContent)
	if err != nil {
		return "", err
	}

	specs := createStructSpecs(structXmlType, spec, nil)
	for _, elt := range spec.SupportedVersionRanges() {
		specs = append(specs, createStructSpecs(structXmlType, spec, &elt.Minimum)...)
	}
	type context struct {
		EntryOrConfig string
		Specs         []entryStructContext
	}

	entryOrConfig := "Entry"
	if spec.TerraformProviderConfig.ResourceType == properties.TerraformResourceConfig {
		entryOrConfig = "Config"
	}

	data := context{
		EntryOrConfig: entryOrConfig,
		Specs:         specs,
	}

	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", err
	}

	return builder.String(), nil
}

// RenderSpecMatchers generates spec matcher functions for a normalization spec.
func RenderSpecMatchers(spec *properties.Normalization) (string, error) {
	tmplContent, err := loadTemplate("partials/spec_matchers.tmpl")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("render-spec-matchers").Parse(tmplContent)
	if err != nil {
		return "", err
	}

	specs := createStructSpecs(structApiType, spec, nil)
	type context struct {
		EntryOrConfig string
		Specs         []entryStructContext
	}

	entryOrConfig := "Entry"
	if spec.TerraformProviderConfig.ResourceType == properties.TerraformResourceConfig {
		entryOrConfig = "Config"
	}

	data := context{
		EntryOrConfig: entryOrConfig,
		Specs:         specs,
	}

	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", err
	}

	return builder.String(), nil
}
