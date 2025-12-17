package terraform_provider

import (
	"strings"
	"testing"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/stretchr/testify/assert"
)

func TestCreateTemplate(t *testing.T) {
	// Given
	g := &GenerateTerraformProvider{}
	resourceType := "Resource"
	spec := &properties.Normalization{Name: "testResource"}
	templateStr := "{{.Name}}"
	funcMap := template.FuncMap{
		"testFunc": func() string { return "test" },
	}

	// When
	resultTemplate, err := g.createTemplate(resourceType, spec, templateStr, funcMap)

	// Then
	assert.NoError(t, err, "createTemplate should not return an error")
	assert.NotNil(t, resultTemplate, "resultTemplate should not be nil")
	assert.IsType(t, &template.Template{}, resultTemplate, "resultTemplate should be of type *template.Template")
}

func TestExecuteTemplate(t *testing.T) {
	// Given
	g := &GenerateTerraformProvider{}
	tmpl, _ := template.New("test").Parse("Name: {{.Name}}")
	spec := &properties.Normalization{Name: "testResource"}
	terraformProvider := &properties.TerraformProviderFile{
		SpecMetadata: make(map[string]properties.TerraformProviderSpecMetadata),
		Code:         new(strings.Builder),
	}
	names := &NameProvider{TfName: "testResource", MetaName: "_testResource", ResourceStructName: "TestResource"}

	// When
	err := g.executeTemplate(tmpl, spec, terraformProvider, properties.ResourceEntry, properties.SchemaResource, names)

	// Then
	assert.NoError(t, err, "executeTemplate should not return an error")
	assert.Contains(t, terraformProvider.Code.String(), "Name: testResource", "The template should be executed with correct data")
}

func TestGenerateTerraformEntityTemplate(t *testing.T) {
	// Given
	g := &GenerateTerraformProvider{}
	names := &NameProvider{TfName: "testResource", MetaName: "_testResource", ResourceStructName: "TestResource"}
	spec := &properties.Normalization{Name: "testResource"}
	terraformProvider := &properties.TerraformProviderFile{
		SpecMetadata: make(map[string]properties.TerraformProviderSpecMetadata),
		Code:         new(strings.Builder),
	}
	templateStr := "Name: {{.Name}}"
	funcMap := template.FuncMap{"testFunc": func() string { return "test" }}

	// When
	err := g.generateTerraformEntityTemplate(properties.ResourceEntry, properties.SchemaResource, names, spec, terraformProvider, templateStr, funcMap)

	// Then
	assert.NoError(t, err, "generateTerraformEntityTemplate should not return an error")
	assert.Contains(t, terraformProvider.Code.String(), "Name: testResource", "The template should be processed and appended to the provider code")
}
