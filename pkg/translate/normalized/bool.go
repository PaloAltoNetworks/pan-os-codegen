package normalized

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/imports"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
)

var (
	_ Item = &Bool{}
)

type Bool struct {
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

	Default      *bool `json:"default" yaml:"default"`
	IsObjectBool *bool `json:"is_object_bool" yaml:"is_object_bool"`

	Namer *naming.Namer
}

func (o *Bool) Path() []string {
	if o.Parent != nil {
		return append(o.Parent.Path(), o.Name)
	}

	return []string{o.Name}
}

func (o *Bool) Copy() Item {
	if o == nil {
		return nil
	}

	ans := Bool{
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

	if o.IsObjectBool != nil {
		x := *o.IsObjectBool
		ans.IsObjectBool = &x
	}

	return &ans
}

func (o *Bool) ApplyUserConfig(vi Item) {
	if vi == nil {
		return
	}

	v := vi.(*Bool)
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

	if v.Default != nil {
		x := *v.Default
		o.Default = &x
	}

	if v.IsObjectBool != nil {
		x := *v.IsObjectBool
		o.IsObjectBool = &x
	}
}

func (o *Bool) String() string {
	return fmt.Sprintf("Bool:%q un:%q ccn:%q isobj:%v", o.Name, o.UnderscoreName, o.CamelCaseName, o.IsObjectBool)
}

func (o *Bool) NameAs(style int) string {
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

func (o *Bool) GolangType(includeShortName bool, schemas map[string]Item) (string, error) {
	if o == nil {
		return "", fmt.Errorf("bool is nil")
	}

	if o.IsObjectBool != nil && *o.IsObjectBool {
		return "any", nil
	}

	return "bool", nil
}

func (o *Bool) ValidatorString(includeDefault bool) string {
	if includeDefault && o.Default != nil {
		if *o.Default {
			return " Default: `true`."
		} else {
			return " Default: `false`."
		}
	}

	return ""
}

func (o *Bool) GetInternalName() string   { return o.Name }
func (o *Bool) GetUnderscoreName() string { return o.UnderscoreName }
func (o *Bool) GetCamelCaseName() string  { return o.CamelCaseName }
func (o *Bool) SchemaInit(_, _ string) error {
	return fmt.Errorf("bool cannot currently be a schema endpoint")
}

func (o *Bool) GetShortName() string {
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

func (o *Bool) Items() []Item {
	return nil
}

func (o *Bool) GetItems(isTop, all bool, schemas map[string]Item) ([]Item, error) {
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

func (o *Bool) GetSdkImports(all bool, schemas map[string]Item) (map[string]bool, error) {
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

func (o *Bool) ToGolangSdkString(prefix, suffix string, schemas map[string]Item) (string, error) {
	return "", fmt.Errorf("unsupported bool to sdk conversion")
}

func (o *Bool) SchemaReferences() []string { return nil }

func (o *Bool) ApplyParameterConfig(loc string, req bool) error {
	o.Location = loc
	o.Required = &req

	return nil
}

func (o *Bool) GetLocation() string  { return o.Location }
func (o *Bool) GetReference() string { return o.Reference }
func (o *Bool) GetSdkPath() []string { return nil }
func (o *Bool) PackageName() string  { return "" }
func (o *Bool) ToGolangSdkQueryParam() (string, bool, error) {
	if o == nil {
		return "", false, fmt.Errorf("bool is nil")
	} else if o.IsObjectBool != nil && *o.IsObjectBool {
		return "", false, fmt.Errorf("not sure how to turn object bool to query param")
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
			"bool-to-golang-sdk-param",
		).Funcs(
			fm,
		).Parse(`
{{- /* Begin */ -}}
{{ "    " }}
{{- if IsTrue .Required -}}
    uv.Set("{{ .Name }}", strconv.FormatBool(input.{{ .CamelCaseName }}))
{{- else -}}
    if input.{{ .CamelCaseName }} != nil {
        uv.Set("{{ .Name }}", strconv.FormatBool(*input.{{ .CamelCaseName }}))
    }
{{- end }}
{{- /* Done */ -}}`,
		),
	)

	err := t.Execute(&b, o)

	return b.String(), true, err
}

func (o *Bool) ToGolangSdkPathParam() (string, bool, error) {
	if o == nil {
		return "", false, fmt.Errorf("bool is nil")
	}

	var b strings.Builder

	t := template.Must(
		template.New(
			"bool-to-golang-sdk-path-param",
		).Parse(`
{{- /* Begin */ -}}
{{ "    " -}}
    path = strings.ReplaceAll(path, "{{ "{" }}{{ .Name }}{{ "}" }}", strconv.FormatBool(input.{{ .CamelCaseName }}))
{{- /* End */ -}}`,
		),
	)

	err := t.Execute(&b, o)

	return b.String(), true, err
}

func (o *Bool) Rename(v string) {
	o.UnderscoreName = naming.Underscore("", v, "output")
	o.CamelCaseName = naming.CamelCase("", v, "Output", true)
	if o.Description == "" {
		o.Description = fmt.Sprintf("handles output for the %s function.", v)
	}
}

func (o *Bool) TerraformModelType(_, _ string, _ map[string]Item) (string, error) {
	return "types.Bool", nil
}

func (o *Bool) TflogString() (string, error) {
	if o == nil {
		return "", fmt.Errorf("bool is nil")
	}

	var b strings.Builder

	t := template.Must(
		template.New(
			"bool-to-tflog-string",
		).Parse(`
{{- /* Begin */ -}}
{{ "        " }}"{{ .UnderscoreName }}": state.{{ .CamelCaseName }}.ValueBool(),
{{- if not .Required }}
        "has_{{ .UnderscoreName }}": !state.{{ .CamelCaseName }}.IsNull(),
{{- end }}
{{- /* Done */ -}}`,
		),
	)

	err := t.Execute(&b, o)

	return b.String(), err
}

func (o *Bool) IsRequired() bool                                      { return o.Required != nil && *o.Required }
func (o *Bool) IsReadOnly() bool                                      { return o.ReadOnly != nil && *o.ReadOnly }
func (o *Bool) HasDefault() bool                                      { return o.Default != nil }
func (o *Bool) ClearDefault()                                         { o.Default = nil }
func (o *Bool) GetObjects(schemas map[string]Item) ([]*Object, error) { return nil, nil }

func (o *Bool) RenderTerraformDefault() (string, map[string]string, error) {
	if o.Default == nil {
		return "", nil, fmt.Errorf("bool doesn't have a default")
	}

	hclibs := map[string]string{
		"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault": "",
	}

	return fmt.Sprintf("booldefault.StaticBool(%t)", *o.Default), hclibs, nil
}

func (o *Bool) GetPlanModifierClass() (string, error) { return "Bool", nil }
func (o *Bool) GetPlanModifierLibrary() string        { return "bool" }
func (o *Bool) EncryptedParams() []*String            { return nil }

func (o *Bool) RootParent() Item {
	if o.Parent != nil {
		return o.Parent.RootParent()
	}
	return o
}

func (o *Bool) EncHasName() (bool, error)                         { return false, fmt.Errorf("this is not a string") }
func (o *Bool) GetParent() Item                                   { return o.Parent }
func (o *Bool) SetParent(i Item)                                  { o.Parent = i }
func (o *Bool) IsEncrypted() bool                                 { return false }
func (o *Bool) HasEncryptedItems(_ map[string]Item) (bool, error) { return false, nil }
func (o *Bool) GetEncryptionKey(_, _ string, _ bool, _ byte) (string, error) {
	return "", fmt.Errorf("bool not encrypted")
}

func (o *Bool) RenderTerraformValidation() ([]string, *imports.Manager, error) {
	manager := imports.NewManager()
	manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator", "")

	return nil, manager, nil
}
