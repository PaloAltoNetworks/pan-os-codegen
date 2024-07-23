package terraform_provider

import (
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

type Entry struct {
	Name string
	Type string
}

type EntryData struct {
	EntryName string
	Entries   []Entry
}

type spec struct {
	Name                 string
	ParentFunctionSuffix string
	FunctionSuffix       string
	PangoType            string
	TerraformType        string
	Params               map[string]*properties.SpecParam
	OneOf                map[string]*properties.SpecParam
}

func getReturnPangoTypeForProperty(pkgName string, parent string, prop *properties.SpecParam) string {
	if prop.Type == "" {
		return fmt.Sprintf("%s.%s%s", pkgName, parent, prop.Name.CamelCase)
	} else if prop.Type == "list" {
		if prop.Items.Type == "" {
			return fmt.Sprintf("[]%s.%s%s", pkgName, parent, prop.Name.CamelCase)
		} else {
			return fmt.Sprintf("[]%s.%s%s", pkgName, parent, prop.Type)
		}
	} else {
		if prop.Required {
			return fmt.Sprintf("%s.%s%s", pkgName, parent, prop.Type)
		} else {
			return fmt.Sprintf("%s.%s%s", pkgName, parent, prop.Type)
		}
	}
}

func generateFromTerraformToPangoSpec(pkgName string, structName string, parent string, prop *properties.SpecParam) []spec {
	var specs []spec
	if prop.Type == "" {
		name := fmt.Sprintf("%s", prop.Name.CamelCase)
		terraformType := fmt.Sprintf("%s%s%sObject", structName, parent, prop.Name.CamelCase)
		returnType := getReturnPangoTypeForProperty(pkgName, parent, prop)
		specs = append(specs, spec{
			Name:                 name,
			TerraformType:        terraformType,
			PangoType:            returnType,
			ParentFunctionSuffix: fmt.Sprintf("%s%s", structName, parent),
			FunctionSuffix:       fmt.Sprintf("%s%s", structName, name),
			Params:               prop.Spec.Params,
			OneOf:                prop.Spec.OneOf,
		})

		parent += name
		for _, p := range prop.Spec.Params {
			specs = append(specs, generateFromTerraformToPangoSpec(pkgName, structName, parent, p)...)
		}

		for _, p := range prop.Spec.OneOf {
			specs = append(specs, generateFromTerraformToPangoSpec(pkgName, structName, parent, p)...)
		}

	}
	return specs
}

const copyNestedFromTerraformToPangoStr = `
{{- range .Specs }}
{{- $spec := . }}
func CopyFromTerraformToPango{{ .ParentFunctionSuffix }}{{ .Name }}(obj *{{ .TerraformType }}) *{{ .PangoType }} {
	return &{{ .PangoType }}{
  {{- range .Params }}
    {{- if eq .Type "" }}
	{{ .Name.CamelCase }}: CopyFromTerraformToPango{{ $spec.FunctionSuffix }}{{ .Name.CamelCase }}(obj.{{ .Name.CamelCase }}),
    {{- else }}
	{{ .Name.CamelCase }}: obj.{{ .Name.CamelCase }},
    {{- end }}
  {{- end }}

  {{- range .OneOf }}
    {{- if eq .Type "" }}
	{{ .Name.CamelCase }}: CopyFromTerraformToPango{{ $spec.FunctionSuffix }}{{ .Name.CamelCase }}(obj.{{ .Name.CamelCase }}),
    {{- else }}
	{{ .Name.CamelCase }}: obj.{{ .Name.CamelCase }},
    {{- end }}
  {{- end }}
	}
}
{{- end }}
`

func CopyNestedFromTerraformToPango(pkgName string, structName string, props *properties.Normalization) (string, error) {
	var specs []spec
	for _, elt := range props.Spec.Params {
		if elt.Type == "" {
			specs = append(specs, generateFromTerraformToPangoSpec(pkgName, structName, "", elt)...)
		}
	}

	for _, elt := range props.Spec.OneOf {
		if elt.Type == "" {
			specs = append(specs, generateFromTerraformToPangoSpec(pkgName, structName, "", elt)...)
		}
	}

	type context struct {
		Specs []spec
	}

	data := context{
		Specs: specs,
	}
	return processTemplate(copyNestedFromTerraformToPangoStr, "create-nested-copy-from-tf-to-pango", data, nil)
}

func ResourceCreateFunction(structName string, serviceName string, paramSpec *properties.Normalization, terraformProvider *properties.TerraformProviderFile, resourceSDKName string) (string, error) {
	funcMap := template.FuncMap{
		"ConfigToEntry": ConfigEntry,
		"ResourceParamToSchema": func(paramName string, paramParameters properties.SpecParam) (string, error) {
			return ParamToSchemaResource(paramName, paramParameters, terraformProvider)
		},
	}

	if strings.Contains(serviceName, "group") && serviceName != "Device group" {
		serviceName = "group"
	}

	data := map[string]interface{}{
		"structName":      structName,
		"serviceName":     naming.CamelCase("", serviceName, "", false),
		"paramSpec":       paramSpec.Spec,
		"resourceSDKName": resourceSDKName,
		"locations":       paramSpec.Locations,
	}

	return processTemplate(resourceCreateTemplateStr, "resource-create-function", data, funcMap)
}

func ResourceReadFunction(structName string, serviceName string, paramSpec *properties.Normalization, resourceSDKName string) (string, error) {
	if strings.Contains(serviceName, "group") {
		serviceName = "group"
	}

	data := map[string]interface{}{
		"structName":      structName,
		"serviceName":     naming.CamelCase("", serviceName, "", false),
		"resourceSDKName": resourceSDKName,
		"locations":       paramSpec.Locations,
	}

	return processTemplate(resourceReadTemplateStr, "resource-read-function", data, nil)
}

func ResourceUpdateFunction(structName string, serviceName string, paramSpec interface{}, resourceSDKName string) (string, error) {
	if strings.Contains(serviceName, "group") {
		serviceName = "group"
	}

	data := map[string]interface{}{
		"structName":      structName,
		"serviceName":     naming.CamelCase("", serviceName, "", false),
		"resourceSDKName": resourceSDKName,
	}

	return processTemplate(resourceUpdateTemplateStr, "resource-update-function", data, nil)
}

func ResourceDeleteFunction(structName string, serviceName string, paramSpec interface{}, resourceSDKName string) (string, error) {
	if strings.Contains(serviceName, "group") {
		serviceName = "group"
	}

	data := map[string]interface{}{
		"structName":      structName,
		"serviceName":     naming.CamelCase("", serviceName, "", false),
		"resourceSDKName": resourceSDKName,
	}

	return processTemplate(resourceDeleteTemplateStr, "resource-delete-function", data, nil)
}

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

	return processTemplate(resourceEntryConfigTemplate, "config-to-entry", entryData, nil)
}
