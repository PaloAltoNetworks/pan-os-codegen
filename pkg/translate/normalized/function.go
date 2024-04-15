package normalized

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
)

type Function struct {
	Parent      *Namespace
	Name        string `json:"name" yaml:"name"`
	Uri         string
	Method      string
	Description string `json:"description" yaml:"description"`

	Schema     string `json:"schema" yaml:"schema"`
	Properties map[string]bool

	PathParams  []Item
	QueryParams []Item
	Request     Item
	Output      Item
	Namer       *naming.Namer
}

func (o *Function) Rename(name, description string) {
	o.Name = name
	o.Description = description
	if o.Output != nil {
		o.Output.Rename(name)
	}
}

func (o *Function) FunctionHasInput() bool {
	if o == nil {
		return false
	}

	return o.Request != nil || len(o.PathParams) != 0 || len(o.QueryParams) != 0
}

func (o *Function) ItemizeInput() (*Object, error) {
	if o == nil {
		return nil, nil
	}

	if len(o.PathParams) == 0 && len(o.QueryParams) == 0 && o.Request == nil {
		return nil, nil
	}

	params := make(map[string]Item)

	for i := range o.PathParams {
		x := o.PathParams[i]
		name := x.NameAs(2)
		if _, ok := params[name]; ok {
			return nil, fmt.Errorf("Name collision for %q.", name)
		}
		params[name] = x
	}

	for i := range o.QueryParams {
		x := o.QueryParams[i]
		name := x.NameAs(2)
		if _, ok := params[name]; ok {
			return nil, fmt.Errorf("Name collision for %q.", name)
		}
		params[name] = x
	}

	if o.Request != nil {
		name := "request"
		if _, ok := params[name]; ok {
			return nil, fmt.Errorf("Name collision for %q.", name)
		}
		params[name] = o.Request
	}

	var shortName string
	if o.Parent != nil {
		shortName = o.Parent.ShortName
	}
	name := fmt.Sprintf("%sInput", o.Name)
	uname := naming.Underscore("", o.Name, "input")
	t := true
	return &Object{
		Name:           fmt.Sprintf("%sInput", o.Name),
		Description:    fmt.Sprintf("handles input for the %s function.", o.Name),
		Required:       &t,
		UnderscoreName: uname,
		CamelCaseName:  naming.CamelCase("", uname, "", true),
		ShortName:      shortName,
		ClassName:      name,
		Params:         params,
	}, nil
}

func (o *Function) PathNames(style int) []string {
	if o == nil || len(o.PathParams) == 0 {
		return nil
	}

	ans := make([]string, 0, len(o.PathParams))
	for i := range o.PathParams {
		ans = append(ans, o.PathParams[i].NameAs(style))
	}

	return ans
}

func (o *Function) QueryNames(style int) []string {
	if o == nil || len(o.QueryParams) == 0 {
		return nil
	}

	ans := make([]string, 0, len(o.QueryParams))
	for i := range o.QueryParams {
		ans = append(ans, o.QueryParams[i].NameAs(style))
	}

	return ans
}

func (o *Function) String() string {
	var b strings.Builder

	b.WriteString(" * ")
	b.WriteString(o.Name)
	b.WriteString(" (")
	b.WriteString(o.Method)
	b.WriteString(")")

	b.WriteString(" ")
	b.WriteString(o.Uri)

	if o.Request != nil {
		b.WriteString(fmt.Sprintf(" in:%v", o.Request.SchemaReferences()))
	}

	if o.Output != nil {
		b.WriteString(fmt.Sprintf(" out:%v", o.Output.SchemaReferences()))
	}

	return b.String()
}

func (o *Function) AssociatedSchema() string {
	if o == nil {
		return ""
	}

	if o.Schema != "" {
		return o.Schema
	}

	if o.Request != nil {
		if ans := o.Request.SchemaReferences(); len(ans) > 0 {
			return ans[0]
		}
	}

	if o.Output != nil {
		if ans := o.Output.SchemaReferences(); len(ans) > 0 {
			return ans[0]
		}
	}

	return ""
}

func (o *Function) ApplyUserConfig(v *Function) {
	if o == nil || v == nil {
		return
	}

	if v.Name != "" {
		o.Name = v.Name
	}

	if v.Description != "" {
		o.Description = v.Description
	}

	if v.Schema != "" {
		o.Schema = v.Schema
	}
}

func (o *Function) ToGolangSdk(schemas map[string]Item) (string, error) {
	if o == nil {
		return "", fmt.Errorf("nil function")
	}

	fm := templateFuncMap(1, false, schemas)
	var b strings.Builder

	t := template.Must(
		template.New(
			"to-golang-sdk-string",
		).Funcs(
			fm,
		).Parse(`
{{- /* Begin */ -}}
{{- $dospace := false -}}


{{- /* Handle the input class definition. */ -}}
{{- if .FunctionHasInput -}}
{{- $dospace = true -}}
// {{ .Name }}Input handles input for {{ .Name }}.
{{- if ne (len .PathParams) 0 }}
// path: {{ .PathNames 2 }}
{{- end }}
{{- if ne (len .QueryParams) 0 }}
// query: {{ .QueryNames 2}}
{{- end }}
{{- if IsNotNil .Request }}
// hasinput yo
{{- end }}
// {{ .Method }} {{ .Uri }}
type {{ .Name }}Input struct {
{{- range $theItem := .PathParams }}
{{ "    " }}{{ Name $theItem }}{{ " " }}
{{- if IsArrayType $theItem }}[]
{{- else if and (IsBoolType $theItem) (BoolIsTrue $theItem.IsObjectBool) }}
{{- else if not (BoolIsTrue $theItem.Required) }}*
{{- end }}{{ GetGolangType $theItem }}
{{- end }}
{{- range $theItem := .QueryParams }}
{{ "    " }}{{ Name $theItem }}{{ " " }}
{{- if IsArrayType $theItem }}[]
{{- else if and (IsBoolType $theItem) (BoolIsTrue $theItem.IsObjectBool) }}
{{- else if not (BoolIsTrue $theItem.Required) }}*
{{- end }}{{ GetGolangType $theItem }}
{{- end }}
{{- if IsNotNil .Request }}
// ShortName: {{ .Request.GetShortName }}
// Ref: {{ .Request.Reference }}
    Config {{ .Request.ShortName}}.{{ GetGolangType .Request }}
{{- end }}
}
{{- end }}
{{- /* End input handling */ -}}


{{- /* Handle the output class definition. */ -}}
{{- if and (IsNotNil .Output) (eq .Output.Reference "") -}}
{{- $dospace = true }}
// Output here.
{{- end }}
{{- /* End output handling. */ -}}

{{- /* Output the function definition itself. */ -}}
{{- if $dospace }}

{{ end }}
// {{ .Name }} {{ .Description }}
func (c *Client) {{ .Name }}(ctx context.Context
{{- if .FunctionHasInput }}, input {{ .Name }}Input
{{- end -}}
){{ " " }}
{{- if IsNotNil .Output }}({{ .Name }}Output, {{ end }}error
{{- if IsNotNil .Output }}){{ end }}{
    // Variables.
    var err error
    var ans jaEeLI.Config
    path := "{{ .Uri }}"

    // Query parameter handling
    // uv := url.Values{}
    // uv.Set("folder", input.Folder)

    // Path param handling.
    path = strings.ReplaceAll(path, "{id}", input.ObjectId)

    // Execute the command.
    _, err = c.client.Do(ctx, "GET", path, uv, nil, &ans)

    // Done.
    return ans, err
}
{{- /* Done. */ -}}`,
		),
	)

	err := t.Execute(&b, o)

	return b.String(), err
}
