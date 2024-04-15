package normalized

import (
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/imports"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
)

var (
	_ Item = &Float{}
)

type Float struct {
	Parent      Item   `json:"-" yaml:"-"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	Summary     string `json:"summary" yaml:"summary"`
	ReadOnly    *bool  `json:"read_only" yaml:"read_only"`
	Required    *bool  `json:"required" yaml:"required"`

	Reference       string `json:"reference" yaml:"reference"`
	UnderscoreName  string `json:"underscore_name" yaml:"underscore_name"`
	CamelCaseName   string `json:"camel_case_name" yaml:"camel_case_name"`
	DeriveNamesFrom string `json:"derive_names_from" yaml:"derive_names_from"`
	ShortName       string `json:"-" yaml:"-"`
	Location        string `json:"-" yaml:"-"`

	Default *float64  `json:"default" yaml:"default"`
	Example *float64  `json:"example" yaml:"example"`
	Min     *float64  `json:"min" yaml:"min"`
	Max     *float64  `json:"max" yaml:"max"`
	Values  []float64 `json:"values" yaml:"values"`

	Namer *naming.Namer
}

func (o *Float) Path() []string {
	if o.Parent != nil {
		return append(o.Parent.Path(), o.Name)
	}

	return []string{o.Name}
}

func (o *Float) Copy() Item {
	if o == nil {
		return nil
	}

	ans := Float{
		Parent:      o.Parent,
		Name:        o.Name,
		Description: o.Description,
		Summary:     o.Summary,

		Reference:       o.Reference,
		UnderscoreName:  o.UnderscoreName,
		CamelCaseName:   o.CamelCaseName,
		DeriveNamesFrom: o.DeriveNamesFrom,
		ShortName:       o.ShortName,
		Location:        o.Location,
		Values:          append([]float64(nil), o.Values...),
	}

	if o.ReadOnly != nil {
		x := *o.ReadOnly
		ans.ReadOnly = &x
	}

	if o.Required != nil {
		x := *o.Required
		ans.Required = &x
	}

	if o.Default != nil {
		x := *o.Default
		ans.Default = &x
	}

	if o.Example != nil {
		x := *o.Example
		ans.Example = &x
	}

	if o.Min != nil {
		x := *o.Min
		ans.Min = &x
	}

	if o.Max != nil {
		x := *o.Max
		ans.Max = &x
	}

	return &ans
}

func (o *Float) ApplyUserConfig(vi Item) {
	if vi == nil {
		return
	}

	v := vi.(*Float)
	if v == nil {
		return
	}

	if v.Name != "" {
		o.Name = v.Name
	}

	if v.Description != "" {
		o.Description = v.Description
	}

	if v.Summary != "" {
		o.Summary = v.Summary
	}

	if v.ReadOnly != nil {
		x := *v.ReadOnly
		o.ReadOnly = &x
	}

	if v.Required != nil {
		x := *v.Required
		o.Required = &x
	}

	if v.Reference != "" {
		o.Reference = v.Reference
	}

	if v.DeriveNamesFrom != "" {
		o.UnderscoreName = naming.Underscore("", v.DeriveNamesFrom, "")
		o.CamelCaseName = naming.CamelCase("", v.DeriveNamesFrom, "", true)
	}

	if v.UnderscoreName != "" {
		o.UnderscoreName = v.UnderscoreName
	}

	if v.CamelCaseName != "" {
		o.CamelCaseName = v.CamelCaseName
	}

	if v.ShortName != "" {
		o.ShortName = v.ShortName
	}

	if v.Default != nil {
		x := *v.Default
		o.Default = &x
	}

	if v.Example != nil {
		x := *v.Example
		o.Example = &x
	}

	if v.Min != nil {
		x := *v.Min
		o.Min = &x
	}

	if v.Max != nil {
		x := *v.Max
		o.Max = &x
	}

	if len(v.Values) != 0 {
		o.Values = append([]float64(nil), v.Values...)
	}
}

func (o *Float) String() string {
	return fmt.Sprintf("Float:%q un:%q ccn:%q min:%v max:%v", o.Name, o.UnderscoreName, o.CamelCaseName, o.Min, o.Max)
}

func (o *Float) NameAs(style int) string {
	switch style {
	case 0:
		return o.UnderscoreName
	case 1:
		return o.CamelCaseName
	case 2:
		return o.Name
	}

	panic(fmt.Sprintf("Unknown style: %d", style))
}

func (o *Float) GolangType(includeShortName bool, schemas map[string]Item) (string, error) {
	if o == nil {
		return "", fmt.Errorf("float is nil")
	}

	return "float64", nil
}

func (o *Float) ValidatorString(includeDefault bool) string {
	if o == nil {
		return ""
	}

	var b strings.Builder
	if o.Min != nil && o.Max != nil {
		b.WriteString(fmt.Sprintf(" Value must be between %g and %g.", *o.Min, *o.Max))
	} else if o.Min != nil {
		b.WriteString(fmt.Sprintf(" Value must be greater than or equal to %g.", *o.Min))
	} else if o.Max != nil {
		b.WriteString(fmt.Sprintf(" Value must be less than or equal to %g.", *o.Max))
	}

	if len(o.Values) != 0 {
		b.WriteString(" Value must be one of the following: ")
		for i, x := range o.Values {
			if i != 0 {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("`%s`", strconv.FormatFloat(x, 'g', -1, 64)))
		}
		b.WriteString(".")
	}

	if includeDefault && o.Default != nil {
		b.WriteString(fmt.Sprintf(" Default: `%g`.", *o.Default))
	}

	return b.String()
}

func (o *Float) GetInternalName() string   { return o.Name }
func (o *Float) GetUnderscoreName() string { return o.UnderscoreName }
func (o *Float) GetCamelCaseName() string  { return o.CamelCaseName }
func (o *Float) SchemaInit(_, _ string) error {
	return fmt.Errorf("float cannot currently be a schema endpoint")
}

func (o *Float) GetShortName() string {
	if o == nil {
		return ""
	}

	if o.ShortName != "" {
		return o.ShortName
	}

	if o.Parent != nil {
		return o.Parent.GetShortName()
	}

	return ""
}

func (o *Float) Items() []Item {
	return nil
}

func (o *Float) GetItems(isTop, all bool, schemas map[string]Item) ([]Item, error) {
	if o == nil {
		return nil, fmt.Errorf("bool is nil")
	}

	if isTop {
		if o.Reference != "" {
			v2, ok := schemas[o.Reference]
			if !ok {
				return nil, fmt.Errorf("bool:%s ref:%s not found", o.Name, o.Reference)
			}
			return v2.GetItems(true, all, schemas)
		}
		return []Item{o}, nil
	}

	if o.Reference != "" && all {
		v, ok := schemas[o.Reference]
		if !ok {
			return nil, fmt.Errorf("bool:%s ref:%s not present", o.Name, o.Reference)
		}
		return []Item{v}, nil
	}

	return nil, nil
}

func (o *Float) GetSdkImports(all bool, schemas map[string]Item) (map[string]bool, error) {
	if o == nil {
		return nil, fmt.Errorf("bool is nil")
	}

	if o.Reference != "" {
		return map[string]bool{
			o.Reference: true,
		}, nil
	}

	return nil, nil
}

func (o *Float) ToGolangSdkString(prefix, suffix string, schemas map[string]Item) (string, error) {
	return "", fmt.Errorf("unsupported float to sdk conversion")
}

func (o *Float) SchemaReferences() []string { return nil }

func (o *Float) ApplyParameterConfig(loc string, req bool) error {
	o.Location = loc
	o.Required = &req

	return nil
}

func (o *Float) GetLocation() string  { return o.Location }
func (o *Float) GetReference() string { return o.Reference }
func (o *Float) GetSdkPath() []string { return nil }
func (o *Float) PackageName() string  { return "" }
func (o *Float) ToGolangSdkQueryParam() (string, bool, error) {
	if o == nil {
		return "", false, fmt.Errorf("float is nil")
	}

	var b strings.Builder

	fm := template.FuncMap{
		"IsTrue": func(v *bool) bool {
			if v == nil {
				return false
			}
			return *v == true
		},
	}

	t := template.Must(
		template.New(
			"float-to-golang-sdk-param",
		).Funcs(
			fm,
		).Parse(`
{{- /* Begin */ -}}
{{ "    " }}
{{- if IsTrue .Required -}}
    uv.Set("{{ .Name }}", strconv.FormatFloat(input.{{ .CamelCaseName }}, 'g', -1, 64))
{{- else -}}
    if input.{{ .CamelCaseName }} != nil {
        uv.Set("{{ .Name }}", strconv.FormatFloat(*input.{{ .CamelCaseName }}, 'g', -1, 64))
    }
{{- end }}
{{- /* Done */ -}}`,
		),
	)

	err := t.Execute(&b, o)

	return b.String(), true, err
}

func (o *Float) ToGolangSdkPathParam() (string, bool, error) {
	if o == nil {
		return "", false, fmt.Errorf("float is nil")
	}

	var b strings.Builder

	t := template.Must(
		template.New(
			"float-to-golang-sdk-path-param",
		).Parse(`
{{- /* Begin */ -}}
{{ "    " -}}
    path = strings.ReplaceAll(path, "{{ "{" }}{{ .Name }}{{ "}" }}", strconv.FormatFloat(input.{{ .CamelCaseName }}, 'g', -1, 64))
{{- /* End */ -}}`,
		),
	)

	err := t.Execute(&b, o)

	return b.String(), true, err
}

func (o *Float) Rename(v string) {
	o.UnderscoreName = naming.Underscore("", v, "output")
	o.CamelCaseName = naming.CamelCase("", v, "Output", true)
	if o.Description == "" {
		o.Description = fmt.Sprintf("handles output for the %s function.", v)
	}
}

func (o *Float) TerraformModelType(_, _ string, _ map[string]Item) (string, error) {
	return "types.Float64", nil
}

func (o *Float) TflogString() (string, error) {
	if o == nil {
		return "", fmt.Errorf("float is nil")
	}

	var b strings.Builder

	t := template.Must(
		template.New(
			"float-to-tflog-string",
		).Parse(`
{{- /* Begin */ -}}
{{ "        " }}"{{ .UnderscoreName }}": state.{{ .CamelCaseName }}.ValueFloat64(),
{{- if not .Required }}
        "has_{{ .UnderscoreName }}": !state.{{ .CamelCaseName }}.IsNull(),
{{- end }}
{{- /* Done */ -}}`,
		),
	)

	err := t.Execute(&b, o)

	return b.String(), err
}

func (o *Float) IsRequired() bool                                      { return o.Required != nil && *o.Required }
func (o *Float) IsReadOnly() bool                                      { return o.ReadOnly != nil && *o.ReadOnly }
func (o *Float) HasDefault() bool                                      { return o.Default != nil }
func (o *Float) ClearDefault()                                         { o.Default = nil }
func (o *Float) GetObjects(schemas map[string]Item) ([]*Object, error) { return nil, nil }

func (o *Float) RenderTerraformDefault() (string, map[string]string, error) {
	if o.Default == nil {
		return "", nil, fmt.Errorf("float doesn't have a default")
	}

	hclibs := map[string]string{
		"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default": "",
	}

	return fmt.Sprintf("float64default.StaticFloat64(%g)", *o.Default), hclibs, nil
}
func (o *Float) EncryptedParams() []*String { return nil }

func (o *Float) RootParent() Item {
	if o.Parent != nil {
		return o.Parent.RootParent()
	}
	return o
}

func (o *Float) EncHasName() (bool, error)                         { return false, fmt.Errorf("this is not a string") }
func (o *Float) GetParent() Item                                   { return o.Parent }
func (o *Float) SetParent(i Item)                                  { o.Parent = i }
func (o *Float) IsEncrypted() bool                                 { return false }
func (o *Float) HasEncryptedItems(_ map[string]Item) (bool, error) { return false, nil }
func (o *Float) GetEncryptionKey(_, _ string, _ bool, _ byte) (string, error) {
	return "", fmt.Errorf("float not encrypted")
}

func (o *Float) RenderTerraformValidation() ([]string, *imports.Manager, error) {
	if o == nil {
		return nil, nil, fmt.Errorf("float is nil")
	}

	manager := imports.NewManager()
	manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework-validators/float64validator", "")

	ans := make([]string, 0, 2)

	// Min and Max.
	if o.Min != nil && o.Max != nil {
		ans = append(ans, fmt.Sprintf("float64validator.Between(%g, %g),", *o.Min, *o.Max))
	} else if o.Min != nil {
		ans = append(ans, fmt.Sprintf("float64validator.AtLeast(%g),", *o.Min))
	} else if o.Max != nil {
		ans = append(ans, fmt.Sprintf("float64validator.AtMost(%g),", *o.Max))
	}

	// Values.
	if len(o.Values) != 0 {
		var inner strings.Builder
		for vnum, val := range o.Values {
			if vnum != 0 {
				inner.WriteString(", ")
			}
			inner.WriteString(fmt.Sprintf("%g", val))
		}
		ans = append(ans, fmt.Sprintf("float64validator.OneOf(%s),", inner.String()))
	}

	return ans, manager, nil
}
