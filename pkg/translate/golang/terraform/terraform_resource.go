package terraform

import (
	"fmt"
	_ "sort"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

func GenerateTerraformResource(spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile) error {

	tfName := spec.Name
	metaName := fmt.Sprintf("_%s", tfName)
	structName := naming.CamelCase("", tfName, "GenerateTerraformResource", false)

	terraformResourceTemplateFunction := template.FuncMap{
		"metaName":   func() string { return metaName },
		"structName": func() string { return structName },
	}
	resourceTemplate := template.Must(
		template.New(
			fmt.Sprintf("terraform-resource-%s", tfName),
		).Funcs(
			terraformResourceTemplateFunction,
		).Parse(resourceTemplateStr),
	)
	var renderedTemplate strings.Builder
	if err := resourceTemplate.Execute(&renderedTemplate, spec); err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}
	terraformProvider.Code = renderedTemplate
	terraformProvider.Resources = append(terraformProvider.Resources, structName)
	return nil
}
