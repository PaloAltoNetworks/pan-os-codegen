package terraform_provider

import (
	"fmt"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"log"
	"strings"
	"text/template"
)

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
func ParamToSchema(paramName string, paramProp properties.TerraformProviderParams) (string, error) {
	funcMap := template.FuncMap{
		"AttName":       func() string { return paramName },
		"CamelCaseType": func() string { return naming.CamelCase("", paramProp.Type, "", true) },
	}
	schemaTemplate := template.Must(
		template.New(
			"describe-param",
		).Funcs(
			funcMap,
		).Parse(`
{{- /* Begin */ -}}
            "{{ AttName }}": schema.{{ CamelCaseType }}Attribute{
                Description: ProviderParamDescription(
                    "{{ .Description }}",
                    "{{ .DefaultValue }}",
                    "{{ .EnvName }}",
					"{{ AttName }}",
                ),
                Optional: {{ .Optional }},
{{- if .Sensitive }}
                Sensitive: true,
{{- end }}
            },
{{- /* Done */ -}}`,
		),
	)
	var builder strings.Builder
	err := schemaTemplate.Execute(&builder, paramProp)
	return builder.String(), err
}
func TfidStruct(structType string, structName string) (string, error) {
	var tfidStructTemplate *template.Template
	if structType == "entry" {
		funcMap := template.FuncMap{}
		tfidStructTemplate = template.Must(
			template.New(
				"describe-param",
			).Funcs(
				funcMap,
			).Parse(`
{{- /* Begin */ -}}
	Name     string          ` + "`" + `json:"name"` + "`" + `
	Location ` + fmt.Sprintf(`%s.Location `, structName) + "`" + `json:"location"` + "`" + `
{{- /* Done */ -}}`,
			),
		)
	}
	var builder strings.Builder
	err := tfidStructTemplate.Execute(&builder, "")
	return builder.String(), err
}

func ParamToModelResource(paramName string, paramProp *properties.SpecParam, structName string) (string, error) {
	funcMap := template.FuncMap{
		"CamelCaseName":  func() string { return naming.CamelCase("", paramName, "", true) },
		"UnderscoreName": func() string { return naming.Underscore("", paramName, "") },
		"CamelCaseType":  func() string { return naming.CamelCase("", paramProp.Type, "", true) },
		"structName":     func() string { return structName },
	}

	var builder strings.Builder
	var modelTemplate *template.Template

	// If the parameter has nested parameters (type is empty)
	if paramProp.Type == "" {
		modelTemplate = template.Must(
			template.New(
				"param-to-model",
			).Funcs(
				funcMap,
			).Parse(`
           {{- /* Begin */ -}}
	   {{ "    " }}{{ CamelCaseName }} ` + structName + `{{ CamelCaseName }}Object ` + "`" + `tfsdk:"{{ UnderscoreName }}"` + "`" + `
           {{- /* Done */ -}}`,
			),
		)
	} else {
		modelTemplate = template.Must(
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
	}
	err := modelTemplate.Execute(&builder, paramProp)
	if err != nil {
		return "", err
	}
	return builder.String(), nil
}

func TfidResourceModel(structType string, structName string) (string, error) {
	var tfidStructTemplate *template.Template
	if structType == "entry" {
		funcMap := template.FuncMap{}
		tfidStructTemplate = template.Must(
			template.New(
				"describe-param",
			).Funcs(
				funcMap,
			).Parse(`
{{- /* Begin */ -}}
	Tfid     types.String          ` + "`" + `tfsdk:"tfid"` + "`" + `
	Location ` + fmt.Sprintf(`%sLocation `, structName) + "`" + `tfsdk:"location"` + "`" + `
{{- /* Done */ -}}`,
			),
		)
	}
	var builder strings.Builder
	err := tfidStructTemplate.Execute(&builder, "")
	return builder.String(), err
}

func ModelNestedStruct(paramName string, paramProp *properties.SpecParam, structName string) (string, error) {
	if paramProp.Type == "" && paramProp.Spec != nil {
		nestedStructsString := strings.Builder{}
		createdStructs := make(map[string]bool)
		nestedStruct, err := createNestedStruct(paramName, paramProp, structName, &nestedStructsString, createdStructs)
		if err != nil {
			return "", err
		}
		return nestedStruct, err
	}

	return "", nil
}

func createNestedStruct(paramName string, paramProp *properties.SpecParam, structName string, nestedStructString *strings.Builder, createdStructs map[string]bool) (string, error) {
	if paramProp.Spec.Params != nil || paramProp.Spec.OneOf != nil {
		var iterator map[string]*properties.SpecParam
		if paramProp.Spec.Params != nil {
			iterator = paramProp.Spec.Params
		} else if paramProp.Spec.OneOf != nil {
			log.Printf("[ DEBUG ] Found OneOf: %s, %s, %s \n", paramName, paramProp.Name, structName)
			iterator = paramProp.Spec.OneOf
		}

		for nestedIndex, nestedParam := range iterator {
			nestedStructName := fmt.Sprintf("%s%s", structName, naming.CamelCase("", paramName, "", true))
			if _, exists := createdStructs[nestedStructName]; !exists {
				createdStructs[nestedStructName] = true
				log.Printf("paramName: %s, nestedIndex: %s, nestedParams: %v, structName: %s \n", paramName, nestedIndex, nestedParam, structName)
				funcMap := template.FuncMap{
					"structName": func() string { return nestedStructName },
					"structItems": func(paramName string, paramProp *properties.SpecParam) (string, error) {
						return ParamToModelResource(paramName, paramProp, nestedStructName)
					},
				}
				nestedStructTemplate := template.Must(
					template.New("nested-struct").Funcs(funcMap).Parse(resourceModelNestedStructTemplate))
				err := nestedStructTemplate.Execute(nestedStructString, paramProp)
				if err != nil {
					log.Printf("Error executing nestedStructTemplate: %v", err)
					return "", err
				}
				if nestedParam.Type == "" && nestedParam.Spec != nil {
					nestedStruct, err := createNestedStruct(nestedIndex, nestedParam, nestedStructName, nestedStructString, createdStructs)
					if err != nil {
						log.Printf("Error executing nestedStructTemplate: %v", err)
						return "", err
					}
					return nestedStruct, err
				}
			}
		}
		return nestedStructString.String(), nil
	}

	return "", nil
}
