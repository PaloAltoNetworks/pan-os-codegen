package terraform

import (
	"fmt"
	_ "sort"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

func Resource(spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile) error {

	tfName := spec.Name
	metaName := fmt.Sprintf("_%s", tfName)
	structName := naming.CamelCase("", tfName, "Resource", false)
	modelName := naming.CamelCase("", tfName, "RsModel", false)
	newFuncName := naming.CamelCase("New", structName, "", true)

	resourcetemplateFunctions := template.FuncMap{
		"MetaName":         func() string { return metaName },
		"StructName":       func() string { return structName },
		"ModelName":        func() string { return modelName },
		"NewFuncName":      func() string { return newFuncName },
		"RepoShortName":    func() string { return spec.RepositoryShortName },
		"ServiceShortName": func() string { return "fleflyf" },
		"ProviderName":     func() string { return spec.Name },
	}
	resourceTemplate := template.Must(
		template.New(
			fmt.Sprintf("terraform-resource-%s", tfName),
		).Funcs(
			resourcetemplateFunctions,
		).Parse(`
{{- /* Begin */ -}}
// Resource object

var (
	_ resource.Resource                = &nestedAddressObjectResource{}
	_ resource.ResourceWithConfigure   = &nestedAddressObjectResource{}
	_ resource.ResourceWithImportState = &nestedAddressObjectResource{}
)


{{- /* Done */ -}}`,
		),
	)
	var renderedTemplate strings.Builder
	if err := resourceTemplate.Execute(&renderedTemplate, spec); err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}
	terraformProvider.Code = renderedTemplate
	terraformProvider.Resources = append(terraformProvider.Resources, structName)
	return nil
}
