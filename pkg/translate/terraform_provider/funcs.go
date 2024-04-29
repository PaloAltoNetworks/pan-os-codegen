package terraform_provider

import (
	"fmt"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
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
	Name     string          ` + "`" + `json:"name" ` + "`" + `
	Location ` + fmt.Sprintf(`%s.Location `, structName) + "`" + `json:"name" ` + "`" + `
{{- /* Done */ -}}`,
			),
		)
	}

	var builder strings.Builder
	err := tfidStructTemplate.Execute(&builder, "")

	return builder.String(), err
}
