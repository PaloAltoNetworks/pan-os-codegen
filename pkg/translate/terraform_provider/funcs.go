package terraform_provider

import (
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"strings"
	"text/template"
)

func ParamToModel(paramName string, paramProp properties.TerraformProviderParams) (string, error) {
	funcMap := template.FuncMap{
		"CamelCaseName": func() string { return naming.CamelCase("", paramName, "", true) },
		"CamelCaseType": func() string { return naming.CamelCase("", paramProp.Type, "", true) },
	}

	t := template.Must(
		template.New(
			"param-to-model",
		).Funcs(
			funcMap,
		).Parse(`
{{- /* Begin */ -}}
{{ "    " }}{{ CamelCaseName }} types.{{ CamelCaseType }} ` + "`" + `tfsdk:"{{ .EnvName }}"` + "`" + `
{{- /* Done */ -}}`,
		),
	)

	var builder strings.Builder
	err := t.Execute(&builder, paramProp)

	return builder.String(), err
}

func ParamToSchema(paramName string, paramProp properties.TerraformProviderParams) (string, error) {
	funcMap := template.FuncMap{
		"AttName":       func() string { return paramName },
		"CamelCaseType": func() string { return naming.CamelCase("", paramProp.Type, "", true) },
	}

	t := template.Must(
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
	err := t.Execute(&builder, paramProp)

	return builder.String(), err
}
