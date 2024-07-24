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

		var params map[string]*properties.SpecParam
		var variants map[string]*properties.SpecParam
		if prop.Type == "" {
			params = prop.Spec.Params
			variants = prop.Spec.OneOf
		}

		specs = append(specs, spec{
			Name:                 name,
			TerraformType:        terraformType,
			PangoType:            returnType,
			ParentFunctionSuffix: fmt.Sprintf("%s%s", structName, parent),
			FunctionSuffix:       fmt.Sprintf("%s%s", structName, name),
			Params:               params,
			OneOf:                variants,
		})

		if prop.Type == "" {
			parent += name
			for _, p := range params {
				specs = append(specs, generateFromTerraformToPangoSpec(pkgName, structName, parent, p)...)
			}

			for _, p := range variants {
				specs = append(specs, generateFromTerraformToPangoSpec(pkgName, structName, parent, p)...)
			}
		}

	}
	return specs
}

const copyNestedFromTerraformToPangoStr = `
{{- define "terraformNestedElementsAssign" }}
  {{- with .Parameter }}

  {{- $result := .Name.LowerCamelCase }}
  {{- $diag := .Name.LowerCamelCase | printf "%s_diags" }}

	var {{ $result }} *{{ $.Parent }}{{ .Name.CamelCase }}
	var {{ $diag }} diag.Diagnostics
	{{ $result }}, {{ $diag }} = {{ $.CopyFunction }}
	diags.Append({{ $diag }}...)

  {{- end }}
{{- end }}

{{- define "terraformListElementsAs" }}
  {{- with .Parameter }}
    {{- if eq .Items.Type "entry" }}
	var {{ .Name.LowerCamelCase }}_elements []{{ .Name.CamelCase }}
    {{- else }}
	var {{ .Name.LowerCamelCase }}_elements []{{ .Items.Type }}
    {{- end }}
	{
		d := obj.{{ .Name.CamelCase }}.ElementsAs(ctx, &{{ .Name.LowerCamelCase }}_elements, false)
		diags.Append(d...)
	}
  {{- end }}
{{- end }}

{{- range .Specs }}
{{- $spec := . }}
func CopyFromTerraformToPango{{ .ParentFunctionSuffix }}{{ .Name }}(ctx context.Context, obj {{ .TerraformType }}) (*{{ .PangoType }}, diag.Diagnostics) {
	var diags diag.Diagnostics
  {{- range .Params }}
    {{- if eq .Type "" }}
        {{- $copyFn := printf "CopyFromTerraformToPango%s%s%s(ctx, obj.%s)" $spec.ParentFunctionSuffix $spec.Name .Name.CamelCase .Name.CamelCase }}
	{{- template "terraformNestedElementsAssign" Map "Parameter" . "CopyFunction" $copyFn "Parent" $spec.PangoType }}
    {{- else if eq .Type "list" }}
        {{- $copyFn := printf "COPY LIST LOL(obj.%s)" .Name.CamelCase }}
	{{- template "terraformListElementsAs" Map "Parameter" . "CopyFunction" $copyFn }}
    {{- end }}
  {{- end }}

  {{- range .OneOf }}
    {{- if eq .Type "" }}
        {{- $copyFn := printf "CopyFromTerraformToPango%s%s%s(ctx, obj.%s)" $spec.ParentFunctionSuffix $spec.Name .Name.CamelCase .Name.CamelCase }}
        {{- template "terraformNestedElementsAssign" Map "Parameter" . "CopyFunction" $copyFn "Parent" $spec.PangoType }}
    {{- else if eq .Type "list" }}
        {{- $copyFn := printf "COPY LIST LOL(obj.%s)" .Name.CamelCase }}
	{{- template "terraformListElementsAs" Map "Parameter" . "CopyFunction" $copyFn }}
    {{- end }}
  {{- end }}

	result := &{{ .PangoType }}{
  {{- range .Params }}
    {{- if eq .Type "" }}
	{{ .Name.CamelCase }}: {{ .Name.LowerCamelCase }},
    {{- else if eq .Type "list" }}
	{{- if eq .Items.Type "object" }}
		// TODO: List objects {{ .Name.CamelCase }},
        {{- else }}
		{{ .Name.CamelCase }}: {{ .Name.LowerCamelCase }}_elements,
        {{- end }}
    {{- else }}
	{{ .Name.CamelCase }}: obj.{{ .Name.CamelCase }}.Value{{ CamelCaseType .Type }}Pointer(),
    {{- end }}
  {{- end }}

  {{- range .OneOf }}
    {{- if eq .Type "" }}
	{{ .Name.CamelCase }}: {{ .Name.LowerCamelCase }},
    {{- else if eq .Type "list" }}
	{{- if eq .Items.Type "object" }}
		// TODO: List objects {{ .Name.CamelCase }},
        {{- else }}
		{{ .Name.CamelCase }}: {{ .Name.LowerCamelCase }}_elements,
        {{- end }}
    {{- else }}
	{{ .Name.CamelCase }}: obj.{{ .Name.CamelCase }}.Value{{ CamelCaseType .Type }}Pointer(),
    {{- end }}
  {{- end }}
	}

	return result, diags
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
	return processTemplate(copyNestedFromTerraformToPangoStr, "create-nested-copy-from-tf-to-pango", data, commonFuncMap)
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

	return processTemplate(resourceCreateFunction, "resource-create-function", data, funcMap)
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

	return processTemplate(resourceReadFunction, "resource-read-function", data, nil)
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

	return processTemplate(resourceUpdateFunction, "resource-update-function", data, nil)
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

	return processTemplate(resourceDeleteFunction, "resource-delete-function", data, nil)
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

	return processTemplate(resourceConfigEntry, "config-entry", entryData, nil)
}
