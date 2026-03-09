package terraform_provider

import (
	"fmt"
	"strings"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

type Field struct {
	Name    string
	Type    string
	TagName string
}

type StructData struct {
	StructName string
	Fields     []Field
}

// ParamToModelBasic converts the given parameter name and properties to a model representation.
func ParamToModelBasic(paramName string, paramProp interface{}) (string, error) {
	data := map[string]interface{}{
		"paramName": paramName,
	}
	paramPropMap := structToMap(paramProp)
	for k, v := range paramPropMap {
		data[k] = v
	}

	return processTemplate(`
{{- /* Begin */ -}}
{{ "    " }}{{ CamelCaseName .paramName }} types.{{ CamelCaseType .Type }} `+"`"+`tfsdk:"{{ UnderscoreName .paramName }}"`+"`"+`
{{- /* Done */ -}}`, "param-to-model", data, nil)
}

// ParamToSchemaProvider converts the given parameter name and properties to a schema representation.
func ParamToSchemaProvider(paramName string, paramProp interface{}) (string, error) {
	return processTemplate(`
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
{{- if .Items }}
		ElementType: types.{{CamelCaseType .Items.Type}}Type,
{{- end }}
    },
{{- /* Done */ -}}`, "describe-param", paramProp, nil)
}

func ParamToSchemaResource(paramName string, paramProp interface{}, terraformProvider *properties.TerraformProviderFile) (string, error) {
	switch v := paramProp.(type) {
	case *properties.SpecParam:
		if v.Type == "bool" && v.Default != "" {
			terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault", "")
		}

		return processTemplate(`
{{- /* Begin */ -}}
{{- if .Type }}
    "`+strings.ReplaceAll(paramName, "-", "_")+`": rsschema.{{ CamelCaseType .Type }}Attribute{
{{- else }}
    "`+strings.ReplaceAll(paramName, "-", "_")+`": rsschema.SingleNestedAttribute{
{{- end }}
        Description: "{{ .Description }}",
		{{- if .Required }}
        Required: true,
		{{- else }}
		Optional:    true,
		{{- end }}
		{{- if .Items }}
		ElementType: types.{{CamelCaseType .Items.Type}}Type,
		{{- end }}
		{{- if .Default }}
		Default: {{.Type}}default.Static{{ CamelCaseType .Type }}({{- if eq .Type "string" }}{{ printf "%q" .Default }}{{ else if eq .Type "bool" }}{{ .Default }}{{ else }}{{ .Default }}{{ end }}),
		Computed: true ,
		{{- end }}
    },
{{- /* Done */ -}}`, "describe-param", v, nil)

	case *properties.Location:
		return processTemplate(`
{{- /* Begin */ -}}
{{- if .Vars }}
     "`+strings.ReplaceAll(paramName, "-", "_")+`": rsschema.SingleNestedAttribute{
{{- else }}
    "`+strings.ReplaceAll(paramName, "-", "_")+`": rsschema.StringAttribute{
{{- end }}
        Description: "{{ .Description }}",
		Required: true,
    },
{{- /* Done */ -}}`, "describe-location", v, nil)

	default:
		return "", fmt.Errorf("unsupported type: %T", paramProp)
	}
}

func CreateResourceSchemaLocationAttribute() (string, error) {
	// Stub implementation - resource schema location attribute not yet implemented
	return "", nil
}

// CreateTfIdStruct generates a template for a struct based on the provided structType and structName.
func CreateTfIdStruct(structType string, structName string) (string, error) {
	if structType == "entry" {
		return processTemplate(`
{{- /* Begin */ -}}
    Name     string          `+"`json:\"name\"`"+`
    Location `+structName+`.Location `+"`json:\"location\"`"+`
{{- /* Done */ -}}`, "describe-param", nil, nil)
	} else {
		return processTemplate(`
{{- /* Begin */ -}}
    Location `+structName+`.Location `+"`json:\"location\"`"+`
{{- /* Done */ -}}`, "describe-param", nil, nil)
	}
}

// CreateTfIdResourceModel generates a Terraform resource struct part for TFID.
func CreateTfIdResourceModel(structType string, structName string) (string, error) {
	if structType == "entry" {
		return processTemplate(`
{{- /* Begin */ -}}
    Tfid     types.String          `+"`tfsdk:\"tfid\"`"+`
    Location `+structName+`Location `+"`tfsdk:\"location\"`"+`
{{- /* Done */ -}}`, "describe-param", nil, nil)
	} else {
		return processTemplate(`
{{- /* Begin */ -}}
    Tfid     types.String          `+"`tfsdk:\"tfid\"`"+`
    Location `+structName+`Location `+"`tfsdk:\"location\"`"+`
{{- /* Done */ -}}`, "describe-param", nil, nil)
	}
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
        {{ CamelCaseName .Name }} *{{ .structName }}{{ CamelCaseName .Name }}Object ` + "`tfsdk:\"{{ UnderscoreName .Name }}\"`" + `
    {{- else }}
        {{ CamelCaseName .Name }} types.{{ CamelCaseType .Type }} ` + "`tfsdk:\"{{ UnderscoreName .Name }}\"`" + `
    {{- end -}}
{{- /* Done */ -}}`
	return processTemplate(templateText, "param-to-model", data, nil)
}
