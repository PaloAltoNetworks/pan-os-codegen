package terraform

import (
	"fmt"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/translate"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

// NameProvider encapsulates the naming conventions for Terraform Provider resource.
type NameProvider struct {
	TfName     string
	MetaName   string
	StructName string
}

// NewNameProvider creates a new NameProvider instance based on given specifications.
func NewNameProvider(spec *properties.Normalization, resourceName string) *NameProvider {
	tfName := spec.Name
	metaName := fmt.Sprintf("_%s", tfName)
	structName := naming.CamelCase("", tfName, resourceName, true)
	return &NameProvider{tfName, metaName, structName}
}

// GenerateTerraformProvider handles generation of Terraform resources and data sources.
type GenerateTerraformProvider struct{}

// generateTerraformTemplate handles the common logic for generating both resources and data sources.
func (tfp *GenerateTerraformProvider) generateTerraformTemplate(resourceType string, structName string, spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile, templateStr string, funcMap template.FuncMap) error {
	err := tfp.generateTemplate(fmt.Sprintf("terraform-%s-%%s", resourceType), templateStr, spec, terraformProvider, funcMap)
	if err != nil {
		return fmt.Errorf("error generating %s template: %v", resourceType, err)
	}

	if resourceType == "Resource" {
		terraformProvider.Resources = append(terraformProvider.Resources, structName)
	} else if resourceType == "DataSource" {
		terraformProvider.DataSources = append(terraformProvider.DataSources, structName)
	}
	return nil
}

// GenerateTerraformResource generates a Terraform resource.
func (tfp *GenerateTerraformProvider) GenerateTerraformResource(spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile) error {
	resourceType := "Resource"
	names := NewNameProvider(spec, resourceType)
	funcMap := template.FuncMap{
		"metaName":   func() string { return names.MetaName },
		"structName": func() string { return names.StructName },
	}

	return tfp.generateTerraformTemplate(resourceType, names.StructName, spec, terraformProvider, resourceTemplateStr, funcMap)
}

// GenerateTerraformDataSource generates a Terraform data source.
func (tfp *GenerateTerraformProvider) GenerateTerraformDataSource(spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile) error {
	resourceType := "DataSource"
	names := NewNameProvider(spec, resourceType)
	funcMap := template.FuncMap{
		"metaName":   func() string { return names.MetaName },
		"structName": func() string { return names.StructName },
	}

	return tfp.generateTerraformTemplate(resourceType, names.StructName, spec, terraformProvider, dataSourceTemplateStr, funcMap)
}

func (tfp *GenerateTerraformProvider) GenerateTerraformProviderFile(spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile) error {
	funcMap := template.FuncMap{
		"renderImports":     translate.RenderImports,
		"structSDKLocation": func() string { return strings.Join(spec.GoSdkPath, "/") }}

	return tfp.generateTerraformTemplate("ProviderFile", "ProviderFile", spec, terraformProvider, providerFileTemplateStr, funcMap)

}

// generateTemplate creates and executes a template based on the given parameters.
func (tfp *GenerateTerraformProvider) generateTemplate(templateNamePattern, templateStr string, spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile, function template.FuncMap) error {
	resourceTemplate := template.Must(template.New(fmt.Sprintf(templateNamePattern, spec.Name)).Funcs(function).Parse(templateStr))
	if terraformProvider.Code == nil {
		terraformProvider.Code = new(strings.Builder)
	}

	var renderedTemplate strings.Builder
	if err := resourceTemplate.Execute(&renderedTemplate, spec); err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}
	if _, err := terraformProvider.Code.WriteString(renderedTemplate.String()); err != nil {
		return fmt.Errorf("error writing template: %v", err)
	}
	return nil
}
