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
	Name            string
	PangoType       string
	PangoReturnType string
	TerraformType   string
	ModelOrObject   string
	Params          map[string]*properties.SpecParam
	OneOf           map[string]*properties.SpecParam
}

func getReturnPangoTypeForProperty(pkgName string, parent string, prop *properties.SpecParam) string {
	if prop.Type == "" {
		return fmt.Sprintf("%s.%s", pkgName, parent)
	} else if prop.Type == "list" {
		if prop.Items.Type == "entry" {
			return fmt.Sprintf("%s.%s", pkgName, parent)
		} else {
			return fmt.Sprintf("%s.%s", pkgName, parent)
		}
	} else {
		if prop.Required {
			return fmt.Sprintf("%s.%s", pkgName, parent)
		} else {
			return fmt.Sprintf("%s.%s", pkgName, parent)
		}
	}
}

func generateFromTerraformToPangoSpec(pangoTypePrefix string, terraformPrefix string, paramSpec *properties.SpecParam) []spec {
	if paramSpec.Spec == nil {
		return nil
	}

	var specs []spec

	pangoType := fmt.Sprintf("%s%s", pangoTypePrefix, paramSpec.Name.CamelCase)

	pangoReturnType := fmt.Sprintf("%s%s", pangoTypePrefix, paramSpec.Name.CamelCase)
	terraformType := fmt.Sprintf("%s%s", terraformPrefix, paramSpec.Name.CamelCase)
	element := spec{
		PangoType:       pangoType,
		PangoReturnType: pangoReturnType,
		TerraformType:   terraformType,
		ModelOrObject:   "Object",
		Params:          paramSpec.Spec.Params,
		OneOf:           paramSpec.Spec.OneOf,
	}
	specs = append(specs, element)
	log.Printf("generateFromTerraformToPangoSpec() spec: %v", element)

	renderSpecsForParams := func(params map[string]*properties.SpecParam) {
		for _, elt := range params {
			if elt.Spec == nil {
				continue
			}
			terraformPrefix := fmt.Sprintf("%s%s", terraformPrefix, paramSpec.Name.CamelCase)
			log.Printf("Element: %s, pangoType: %s, terraformPrefix: %s", elt.Name.CamelCase, pangoType, terraformPrefix)
			specs = append(specs, generateFromTerraformToPangoSpec(pangoType, terraformPrefix, elt)...)
		}
	}

	renderSpecsForParams(paramSpec.Spec.Params)
	renderSpecsForParams(paramSpec.Spec.OneOf)

	return specs
}

func generateFromTerraformToPangoParameter(pkgName string, terraformPrefix string, pangoPrefix string, prop *properties.Normalization, parentName string) []spec {
	var specs []spec

	var pangoReturnType string
	if parentName == "" {
		pangoReturnType = fmt.Sprintf("%s.Entry", pkgName)
		pangoPrefix = fmt.Sprintf("%s.", pkgName)
	} else {
		pangoReturnType = fmt.Sprintf("%s.%s", pkgName, parentName)
	}

	specs = append(specs, spec{
		PangoType:       pangoPrefix,
		PangoReturnType: pangoReturnType,
		ModelOrObject:   "Model",
		TerraformType:   terraformPrefix,
		Params:          prop.Spec.Params,
		OneOf:           prop.Spec.OneOf,
	})

	for _, elt := range prop.Spec.Params {
		specs = append(specs, generateFromTerraformToPangoSpec(pangoPrefix, terraformPrefix, elt)...)
	}

	for _, elt := range prop.Spec.OneOf {
		specs = append(specs, generateFromTerraformToPangoSpec(pangoPrefix, terraformPrefix, elt)...)
	}

	return specs
}

const copyNestedFromTerraformToPangoStr = `
{{- define "terraformNestedElementsAssign" }}
  {{- with .Parameter }}

  {{- $result := .Name.LowerCamelCase }}
  {{- $diag := .Name.LowerCamelCase | printf "%s_diags" }}
	var {{ $result }}_entry *{{ $.Spec.PangoType }}{{ .Name.CamelCase }}
	var {{ $diag }} diag.Diagnostics
	{{ $result }}_entry, {{ $diag }} = o.{{ .Name.CamelCase }}.CopyToPango(ctx)
	diags.Append({{ $diag }}...)

  {{- end }}
{{- end }}

{{- define "terraformListElementsAs" }}
  {{- with .Parameter }}
    {{- $pangoType := printf "%s%s" $.Spec.PangoType .Name.CamelCase }}
    {{- $terraformType := printf "%s%s%s" $.Spec.TerraformType .Name.CamelCase $.Spec.ModelOrObject }}
    {{- $pangoEntries := printf "%s_pango_entries" .Name.LowerCamelCase }}
    {{- $tfEntries := printf "%s_tf_entries" .Name.LowerCamelCase }}
    {{- if eq .Items.Type "entry" }}
		var {{ $tfEntries }} []{{ $terraformType }}
		var {{ $pangoEntries }} []{{ $pangoType }}
	{
		d := o.{{ .Name.CamelCase }}.ElementsAs(ctx, &{{ $tfEntries }}, false)
		diags.Append(d...)
		for _, elt := range {{ $tfEntries }} {
			entry, d := elt.CopyToPango(ctx)
			diags.Append(d...)
			{{ $pangoEntries }} = append({{ $pangoEntries }}, *entry)
		}
	}
    {{- else }}
		var {{ $pangoEntries }} []{{ .Items.Type }}
	{
		d := o.{{ .Name.CamelCase }}.ElementsAs(ctx, &{{ $pangoEntries }}, false)
		diags.Append(d...)
	}
    {{- end }}
  {{- end }}
{{- end }}

{{- range .Specs }}
{{- $spec := . }}
func (o *{{ .TerraformType }}{{ .ModelOrObject }}) CopyToPango(ctx context.Context) (*{{ .PangoReturnType }}, diag.Diagnostics) {
	var diags diag.Diagnostics
  {{- range .Params }}
    {{- $terraformType := printf "%s%s" $spec.TerraformType .Name.CamelCase }}
    {{- if eq .Type "" }}
      {{- $pangoType := printf "%sObject" $spec.PangoType }}
	{{- template "terraformNestedElementsAssign" Map "Parameter" . "Spec" $spec }}
    {{- else if eq .Type "list" }}
      {{- $pangoType := printf "%s%s" $spec.PangoType .Name.CamelCase }}
	{{- template "terraformListElementsAs" Map "Parameter" . "Spec" $spec }}
    {{- end }}
  {{- end }}

  {{- range .OneOf }}
    {{- if eq .Type "" }}
      {{- $pangoType := printf "%sObject" $spec.PangoType }}
	{{- template "terraformNestedElementsAssign" Map "Parameter" . "Spec" $spec }}
    {{- else if eq .Type "list" }}
	{{- template "terraformListElementsAs" Map "Parameter" . "Spec" $spec }}
    {{- end }}
  {{- end }}

	result := &{{ .PangoReturnType }}{
  {{- range .Params }}
    {{- if eq .Type "" }}
	{{ .Name.CamelCase }}: {{ .Name.LowerCamelCase }}_entry,
    {{- else if eq .Type "list" }}
	{{- if eq .Items.Type "object" }}
		// TODO: List objects {{ .Name.CamelCase }},
        {{- else }}
		{{ .Name.CamelCase }}: {{ .Name.LowerCamelCase }}_pango_entries,
        {{- end }}
    {{- else }}
	{{ .Name.CamelCase }}: o.{{ .Name.CamelCase }}.Value{{ CamelCaseType .Type }}Pointer(),
    {{- end }}
  {{- end }}

  {{- range .OneOf }}
    {{- if eq .Type "" }}
	{{ .Name.CamelCase }}: {{ .Name.LowerCamelCase }}_entry,
    {{- else if eq .Type "list" }}
	{{- if eq .Items.Type "object" }}
		// TODO: List objects {{ .Name.CamelCase }},
        {{- else }}
		{{ .Name.CamelCase }}: {{ .Name.LowerCamelCase }}_pango_entries,
        {{- end }}
    {{- else }}
	{{ .Name.CamelCase }}: o.{{ .Name.CamelCase }}.Value{{ CamelCaseType .Type }}Pointer(),
    {{- end }}
  {{- end }}
	}

	return result, diags
}
{{- end }}
`

func CopyNestedFromTerraformToPango(pkgName string, terraformTypePrefix string, property *properties.Normalization) (string, error) {
	specs := generateFromTerraformToPangoParameter(pkgName, terraformTypePrefix, "", property, "")

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
