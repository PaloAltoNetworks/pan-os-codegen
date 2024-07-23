package terraform_provider

import (
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

// Package-level function map to avoid repetition in each function
var centralFuncMap = template.FuncMap{
	"CamelCaseName":  func(paramName string) string { return naming.CamelCase("", paramName, "", true) },
	"UnderscoreName": func(paramName string) string { return naming.Underscore("", paramName, "") },
	"CamelCaseType":  func(paramType string) string { return naming.CamelCase("", paramType, "", true) },
}

// mergeFuncMaps merges two template.FuncMap instances.
// In case of a key conflict, the second map's value will override the first one.
func mergeFuncMaps(map1, map2 template.FuncMap) template.FuncMap {
	mergedMap := make(template.FuncMap)

	for key, value := range map1 {
		mergedMap[key] = value
	}

	for key, value := range map2 {
		mergedMap[key] = value
	}

	return mergedMap
}

// centralTemplateExec handles the creation and execution of templates
func centralTemplateExec(templateText, templateName string, data interface{}, funcMap template.FuncMap) (string, error) {
	if len(funcMap) == 0 {
		funcMap = centralFuncMap
	} else {
		funcMap = mergeFuncMaps(funcMap, centralFuncMap)
	}

	tmpl, err := template.New(templateName).Funcs(funcMap).Parse(templateText)
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", err
	}
	return builder.String(), nil
}

// ParamToModel converts the given parameter name and properties to a model representation.
func ParamToModel(paramName string, paramProp properties.TerraformProviderParams) (string, error) {
	funcMap := template.FuncMap{
		"CamelCaseName":  func() string { return naming.CamelCase("", paramName, "", true) },
		"UnderscoreName": func() string { return naming.Underscore("", paramName, "") },
		"CamelCaseType":  func() string { return naming.CamelCase("", paramProp.Type, "", true) },
	}
	modelTemplate := template.Must(
		template.New(
			"param-to-model",
		).Funcs(
			funcMap,
		).Parse(`
{{- /* Begin */ -}}
{{ "    " }}{{ CamelCaseName }} types.{{ CamelCaseType }} ` + "`" + `tfsdk:"{{ UnderscoreName }}"` + "`" + `
{{- /* Done */ -}}`,
		),
	)
	var builder strings.Builder
	err := modelTemplate.Execute(&builder, paramProp)
	return builder.String(), err
}

// ParamToSchema converts the given parameter name and properties to a schema representation.
func ParamToSchema(paramName string, paramProp properties.TerraformProviderParams) (string, error) {
	return centralTemplateExec(`
{{- /* Begin */ -}}
    "`+paramName+`": schema.{{ CamelCaseType .Type }}Attribute{
        Description: ProviderParamDescription(
            "{{ .Description }}",
            "{{ .DefaultValue }}",
            "{{ .EnvName }}",
            "`+paramName+`",
        ),
        Optional: {{ .Optional }},
{{- if .Sensitive }}
        Sensitive: true,
{{- end }}
    },
{{- /* Done */ -}}`, "describe-param", paramProp, nil)
}

// CreateTfIdStruct generates a template for a struct based on the provided structType and structName.
func CreateTfIdStruct(structType string, structName string) (string, error) {
	if structType == "entry" {
		return centralTemplateExec(`
{{- /* Begin */ -}}
    Name     string          `+"`json:\"name\"`"+`
    Location `+structName+`.Location `+"`json:\"location\"`"+`
{{- /* Done */ -}}`, "describe-param", nil, nil)
	}
	return "", nil
}

// CreateTfIdResourceModel generates a Terraform resource struct part for TFID.
func CreateTfIdResourceModel(structType string, structName string) (string, error) {
	if structType == "entry" {
		return centralTemplateExec(`
{{- /* Begin */ -}}
    Tfid     types.String          `+"`tfsdk:\"tfid\"`"+`
    Location `+structName+`Location `+"`tfsdk:\"location\"`"+`
{{- /* Done */ -}}`, "describe-param", nil, nil)
	}
	return "", nil
}

// ParamToModelResource converts the given parameter name and properties to a model representation.
func ParamToModelResource(paramName string, paramProp *properties.SpecParam, structName string) (string, error) {
	data := map[string]interface{}{
		"Name":       paramName,
		"Type":       paramProp.Type,
		"structName": structName,
	}
	templateText := `
{{- /* Begin */ -}}
    {{- if eq .Type "" }}
        {{ CamelCaseName .Name }} {{ .structName }}{{ CamelCaseName .Name }}Object ` + "`tfsdk:\"{{ UnderscoreName .Name }}\"`" + `
    {{- else }}
        {{ CamelCaseName .Name }} types.{{ CamelCaseType .Type }} ` + "`tfsdk:\"{{ UnderscoreName .Name }}\"`" + `
    {{- end -}}
{{- /* Done */ -}}`
	return centralTemplateExec(templateText, "param-to-model", data, nil)
}

// ModelNestedStruct manages nested structure definitions.
func ModelNestedStruct(paramName string, paramProp *properties.SpecParam, structName string) (string, error) {
	if paramProp.Type == "" && paramProp.Spec != nil {
		nestedStructsString := strings.Builder{}
		createdStructs := make(map[string]bool)
		nestedStruct, err := CreateNestedStruct(paramName, paramProp, structName, &nestedStructsString, createdStructs)
		if err != nil {
			return "", err
		}
		return nestedStruct, nil
	}

	return "", nil
}

// CreateNestedStruct recursively creates nested struct definitions.
func CreateNestedStruct(paramName string, paramProp *properties.SpecParam, structName string, nestedStructString *strings.Builder, createdStructs map[string]bool) (string, error) {
	nestedStructName := fmt.Sprintf("%s%s", structName, naming.CamelCase("", paramName, "", true))
	if _, exists := createdStructs[nestedStructName]; exists {
		return "", nil // Avoid recreating existing structs to prevent infinite loops
	}
	createdStructs[nestedStructName] = true

	nestedStructFuncMap := template.FuncMap{
		"structItems": func(paramName string, paramProp *properties.SpecParam) (string, error) {
			return ParamToModelResource(paramName, paramProp, nestedStructName)
		}}

	data := map[string]interface{}{
		"Spec":       paramProp.Spec,
		"structName": nestedStructName,
	}
	nestedStruct, err := centralTemplateExec(resourceModelNestedStructTemplate, "nested-struct", data, nestedStructFuncMap)
	if err != nil {
		log.Printf("[ ERROR ] Executing nested struct template failed: %v", err)
		return "", err
	}

	nestedStructString.WriteString(nestedStruct)

	for nestedIndex, nestedParam := range paramProp.Spec.Params {
		if nestedParam.Type == "" && nestedParam.Spec != nil {
			_, err := CreateNestedStruct(nestedIndex, nestedParam, nestedStructName, nestedStructString, createdStructs)
			if err != nil {
				log.Printf("[ ERROR ] Error creating further nested structures: %v", err)
				return "", err
			}
		}
	}

	for nestedIndex, nestedParam := range paramProp.Spec.OneOf {
		if nestedParam.Type == "" && nestedParam.Spec != nil {
			_, err := CreateNestedStruct(nestedIndex, nestedParam, nestedStructName, nestedStructString, createdStructs)
			if err != nil {
				log.Printf("[ ERROR ] Error creating further nested structures: %v", err)
				return "", err
			}
		}
	}

	return nestedStructString.String(), nil
}
