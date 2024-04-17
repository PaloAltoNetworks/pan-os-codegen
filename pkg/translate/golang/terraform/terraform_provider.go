package terraform

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

type GenerateTerraformProvider struct{}

func (tfp *GenerateTerraformProvider) GenerateTerraformResource(spec *properties.Normalization,
	terraformProvider *properties.TerraformProviderFile) error {

	tfName := spec.Name
	metaName := fmt.Sprintf("_%s", tfName)
	structName := naming.CamelCase("", tfName, "Resource", true)

	terraformResourceTemplateFunction := template.FuncMap{
		"metaName":   func() string { return metaName },
		"structName": func() string { return structName },
	}

	err := tfp.generateTemplate("terraform-resource-%s", resourceTemplateStr, spec, terraformProvider, terraformResourceTemplateFunction)
	if err != nil {
		return fmt.Errorf("error generating template: %v", err)
	}

	terraformProvider.Resources = append(terraformProvider.Resources, structName)
	return nil
}

func (tfp *GenerateTerraformProvider) GenerateTerraformDataSource(spec *properties.Normalization,
	terraformProvider *properties.TerraformProviderFile) error {

	tfName := spec.Name
	metaName := fmt.Sprintf("_%s", tfName)
	structName := naming.CamelCase("", tfName, "DataSource", true)

	terraformResourceTemplateFunction := template.FuncMap{
		"metaName":   func() string { return metaName },
		"structName": func() string { return structName },
	}

	err := tfp.generateTemplate("terraform-data-source-%s", dataSourceTemplateStr, spec, terraformProvider, terraformResourceTemplateFunction)
	if err != nil {
		return fmt.Errorf("error generating template: %v", err)
	}
	terraformProvider.DataSources = append(terraformProvider.DataSources, structName)
	return nil
}

func (tfp *GenerateTerraformProvider) generateTemplate(templateNamePattern, templateStr string, spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile, function template.FuncMap) error {

	resourceTemplate := template.Must(
		template.New(
			fmt.Sprintf(templateNamePattern, spec.Name),
		).Funcs(
			function,
		).Parse(templateStr),
	)
	var renderedTemplate strings.Builder
	if err := resourceTemplate.Execute(&renderedTemplate, spec); err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}
	if _, err := terraformProvider.Code.WriteString(renderedTemplate.String()); err != nil {
		return fmt.Errorf("error writing template: %v", err)
	}
	return nil
}
