package normalized

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/imports"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
)

var (
	_ Item = &String{}
)

type String struct {
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

	Default      *string      `json:"default" yaml:"default"`
	IsPassword   *bool        `json:"is_password" yaml:"is_password"`
	Example      *string      `json:"example" yaml:"example"`
	MinLength    *int64       `json:"min_length" yaml:"min_length"`
	MaxLength    *int64       `json:"max_length" yaml:"max_length"`
	Values       []string     `json:"values" yaml:"values"`
	AddValues    []string     `json:"add_values" yaml:"add_values"`
	RemoveValues []string     `json:"remove_values" yaml:"remove_values"`
	Regex        *string      `json:"regex" yaml:"regex"`
	HashProfile  *HashProfile `json:"hash_profile" yaml:"hash_profile"`

	Namer *naming.Namer
}

func (o *String) Path() []string {
	if o.Parent != nil {
		return append(o.Parent.Path(), o.Name)
	}

	return []string{o.Name}
}

func (o *String) Copy() Item {
	if o == nil {
		return nil
	}

	ans := String{
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
		Values:          append([]string(nil), o.Values...),
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

	if o.IsPassword != nil {
		x := *o.IsPassword
		ans.IsPassword = &x
	}

	if o.Example != nil {
		x := *o.Example
		ans.Example = &x
	}

	if o.MinLength != nil {
		x := *o.MinLength
		ans.MinLength = &x
	}

	if o.MaxLength != nil {
		x := *o.MaxLength
		ans.MaxLength = &x
	}

	if o.Regex != nil {
		x := *o.Regex
		ans.Regex = &x
	}

	return &ans
}

func (o *String) ApplyUserConfig(vi Item) {
	if vi == nil {
		return
	}

	v := vi.(*String)
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

	if v.IsPassword != nil {
		x := *v.IsPassword
		o.IsPassword = &x
	}

	if v.Example != nil {
		x := *v.Example
		o.Example = &x
	}

	if v.MinLength != nil {
		x := *v.MinLength
		o.MinLength = &x
	}

	if v.MaxLength != nil {
		x := *v.MaxLength
		o.MaxLength = &x
	}

	if len(v.Values) != 0 {
		o.Values = append([]string(nil), v.Values...)
	} else if len(v.AddValues) != 0 || len(v.RemoveValues) != 0 {
		var newv []string
		for _, val := range o.Values {
			keep := true
			for _, rm := range o.RemoveValues {
				if val == rm {
					keep = false
					break
				}
			}

			if keep {
				newv = append(newv, val)
			}
		}

		newv = append(newv, o.AddValues...)

		o.Values = newv
	}

	if v.Regex != nil {
		x := *v.Regex
		o.Regex = &x
	}

	if v.HashProfile != nil {
		o.HashProfile = v.HashProfile
	}
}

func (o *String) String() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("String:%q", o.Name))
	s.WriteString(fmt.Sprintf(" un:%q ccn:%q", o.UnderscoreName, o.CamelCaseName))

	if o.Reference != "" {
		s.WriteString(fmt.Sprintf(" ref:%q", o.Reference))
	}

	if o.ReadOnly != nil && *o.ReadOnly {
		s.WriteString(" read-only:true")
	}

	if o.Required != nil && *o.Required {
		s.WriteString(" required:true")
	}

	if o.Default != nil {
		s.WriteString(fmt.Sprintf(" default:%q", *o.Default))
	}

	if o.IsPassword != nil && *o.IsPassword {
		s.WriteString(" ispasswd:true")
	}

	if o.MinLength != nil {
		s.WriteString(fmt.Sprintf(" minlen:%d", *o.MinLength))
	}

	if o.MaxLength != nil {
		s.WriteString(fmt.Sprintf(" maxlen:%d", *o.MaxLength))
	}

	if len(o.Values) != 0 {
		s.WriteString(fmt.Sprintf(" values:%#v", o.Values))
	}

	if o.Regex != nil {
		s.WriteString(fmt.Sprintf(" regex:%q", *o.Regex))
	}

	return s.String()
}

func (o *String) NameAs(style int) string {
	switch style {
	case 0:
		return o.UnderscoreName
	case 1:
		return o.CamelCaseName
	default:
		return o.Name
	}
}

func (o *String) GolangType(includeShortName bool, schemas map[string]Item) (string, error) {
	if o == nil {
		return "", fmt.Errorf("string is nil")
	}

	return "string", nil
}

func (o *String) ValidatorString(includeDefault bool) string {
	var b strings.Builder

	if o.MinLength != nil && o.MaxLength != nil {
		b.WriteString(fmt.Sprintf(" String length must be between %d and %d characters.", *o.MinLength, *o.MaxLength))
	} else if o.MinLength != nil {
		b.WriteString(fmt.Sprintf(" String length must exceed %d characters.", *o.MinLength))
	} else if o.MaxLength != nil {
		b.WriteString(fmt.Sprintf(" String length must not exceed %d characters.", *o.MaxLength))
	}

	if len(o.Values) != 0 && o.Regex != nil {
		x := fmt.Sprintf("`\"%s\"`", strings.Join(o.Values, "\"`, `\""))
		b.WriteString(fmt.Sprintf(" String can either be a specific string(%s) or match this regex: `%s`.", x, *o.Regex))
	} else if len(o.Values) != 0 {
		x := fmt.Sprintf("`\"%s\"`", strings.Join(o.Values, "\"`, `\""))
		b.WriteString(fmt.Sprintf(" String must be one of these: %s.", x))
	} else if o.Regex != nil {
		b.WriteString(fmt.Sprintf(" String validation regex: `%s`.", *o.Regex))
	}

	if includeDefault && o.Default != nil {
		b.WriteString(fmt.Sprintf(" Default: `%q`.", *o.Default))
	}

	return b.String()
}

func (o *String) GetInternalName() string   { return o.Name }
func (o *String) GetUnderscoreName() string { return o.UnderscoreName }
func (o *String) GetCamelCaseName() string  { return o.CamelCaseName }
func (o *String) SchemaInit(_, _ string) error {
	return fmt.Errorf("string cannot currently be a schema endpoint")
}

func (o *String) GetShortName() string {
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

func (o *String) GetItems(isTop, all bool, schemas map[string]Item) ([]Item, error) {
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

func (o *String) Items() []Item {
	return nil
}

func (o *String) GetSdkImports(all bool, schemas map[string]Item) (map[string]bool, error) {
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

func (o *String) ToGolangSdkString(prefix, suffix string, schemas map[string]Item) (string, error) {
	return "", fmt.Errorf("unsupported string to sdk conversion")
}

func (o *String) SchemaReferences() []string { return nil }

func (o *String) ApplyParameterConfig(loc string, req bool) error {
	o.Location = loc
	o.Required = &req

	return nil
}

func (o *String) GetLocation() string  { return o.Location }
func (o *String) GetReference() string { return o.Reference }
func (o *String) GetSdkPath() []string { return nil }
func (o *String) PackageName() string  { return "" }
func (o *String) ToGolangSdkQueryParam() (string, bool, error) {
	if o == nil {
		return "", false, fmt.Errorf("string is nil")
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
			"string-to-golang-sdk-param",
		).Funcs(
			fm,
		).Parse(`
{{- /* Begin */ -}}
{{ "    " }}
{{- if IsTrue .Required -}}
    uv.Set("{{ .Name }}", input.{{ .CamelCaseName }})
{{- else -}}
    if input.{{ .CamelCaseName }} != nil {
        uv.Set("{{ .Name }}", *input.{{ .CamelCaseName }})
    }
{{- end }}
{{- /* Done */ -}}`,
		),
	)

	err := t.Execute(&b, o)

	return b.String(), false, err
}

func (o *String) ToGolangSdkPathParam() (string, bool, error) {
	if o == nil {
		return "", false, fmt.Errorf("string is nil")
	}

	var b strings.Builder

	t := template.Must(
		template.New(
			"string-to-golang-sdk-path-param",
		).Parse(`
{{- /* Begin */ -}}
{{ "    " -}}
    path = strings.ReplaceAll(path, "{{ "{" }}{{ .Name }}{{ "}" }}", input.{{ .CamelCaseName }})
{{- /* End */ -}}`,
		),
	)

	err := t.Execute(&b, o)

	return b.String(), false, err
}

func (o *String) Rename(v string) {
	o.UnderscoreName = naming.Underscore("", v, "output")
	o.CamelCaseName = naming.CamelCase("", v, "Output", true)
	if o.Description == "" {
		o.Description = fmt.Sprintf("handles output for the %s function.", v)
	}
}

func (o *String) TerraformModelType(_, _ string, _ map[string]Item) (string, error) {
	return "types.String", nil
}

func (o *String) TflogString() (string, error) {
	if o == nil {
		return "", fmt.Errorf("string is nil")
	}

	var b strings.Builder

	t := template.Must(
		template.New(
			"string-to-tflog-string",
		).Parse(`
{{- /* Begin */ -}}
{{ "        " }}"{{ .UnderscoreName }}": state.{{ .CamelCaseName }}.ValueString(),
{{- if not .Required }}
        "has_{{ .UnderscoreName }}": !state.{{ .CamelCaseName }}.IsNull(),
{{- end }}
{{- /* Done */ -}}`,
		),
	)

	err := t.Execute(&b, o)

	return b.String(), err
}

func (o *String) IsRequired() bool                                      { return o.Required != nil && *o.Required }
func (o *String) IsReadOnly() bool                                      { return o.ReadOnly != nil && *o.ReadOnly }
func (o *String) HasDefault() bool                                      { return o.Default != nil }
func (o *String) ClearDefault()                                         { o.Default = nil }
func (o *String) GetObjects(schemas map[string]Item) ([]*Object, error) { return nil, nil }

func (o *String) RenderTerraformDefault() (string, map[string]string, error) {
	if o.Default == nil {
		return "", nil, fmt.Errorf("string doesn't have a default")
	}

	hclibs := map[string]string{
		"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault": "",
	}

	return fmt.Sprintf("stringdefault.StaticString(%q)", *o.Default), hclibs, nil
}

func (o *String) EncryptedParams() []*String {
	if o.IsPassword != nil && *o.IsPassword {
		return []*String{o}
	}
	return nil
}

func (o *String) RootParent() Item {
	if o.Parent != nil {
		return o.Parent.RootParent()
	}
	return o
}

func (o *String) HasEncryptedItems(schemas map[string]Item) (bool, error) {
	if o.IsPassword != nil && *o.IsPassword && o.HashProfile != nil {
		return o.HashProfile.IsEncrypted(), nil
	}

	return false, nil
}

func (o *String) EncParams() ([]string, error) {
	if o.IsPassword == nil || !*o.IsPassword {
		return nil, fmt.Errorf("this is not a password")
	} else if o.Parent == nil {
		return nil, fmt.Errorf("no parent")
	}

	p, ok := o.Parent.(*Object)
	if !ok {
		return nil, fmt.Errorf("parent is %T not *Object", o.Parent)
	}

	ans := make([]string, 0, len(p.Params))
	for name := range p.Params {
		ans = append(ans, name)
	}
	return ans, nil
}

func (o *String) EncHasName() (bool, error) {
	if o.IsPassword == nil || !*o.IsPassword {
		return false, fmt.Errorf("this is not a password")
	} else if o.Parent == nil {
		return false, fmt.Errorf("no parent")
	}

	p, ok := o.Parent.(*Object)
	if !ok {
		return false, fmt.Errorf("parent is %T not *Object", o.Parent)
	}

	_, ok = p.Params["name"]
	return ok, nil
}

func (o *String) IsEncrypted() bool {
	if o.IsPassword == nil || !*o.IsPassword || o.HashProfile == nil {
		return false
	}

	return o.HashProfile.IsEncrypted()
}

func (o *String) GetEncryptionKey(src, varName string, srcIsTfType bool, keyType byte) (string, error) {
	if !o.IsEncrypted() {
		return "", fmt.Errorf("string is not encrypted")
	}

	return o.HashProfile.GetEncryptionKey(o, src, varName, srcIsTfType, keyType)
}

func (o *String) GetParent() Item  { return o.Parent }
func (o *String) SetParent(i Item) { o.Parent = i }

func (o *String) RenderTerraformValidation() ([]string, *imports.Manager, error) {
	if o == nil {
		return nil, nil, fmt.Errorf("string is nil")
	}

	manager := imports.NewManager()
	manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator", "")

	ans := make([]string, 0, 2)

	// MinLength and MaxLength.
	if o.MinLength != nil && o.MaxLength != nil {
		ans = append(ans, fmt.Sprintf("stringvalidator.LengthBetween(%d, %d),", *o.MinLength, *o.MaxLength))
	} else if o.MinLength != nil {
		ans = append(ans, fmt.Sprintf("stringvalidator.LengthAtLeast(%d),", *o.MinLength))
	} else if o.MaxLength != nil {
		ans = append(ans, fmt.Sprintf("stringvalidator.LengthAtMost(%d),", *o.MaxLength))
	}

	// Values and regex.
	if len(o.Values) != 0 && o.Regex != nil {
		// Merge static values into a single regex and use that.
		var inner strings.Builder
		inner.WriteString(*o.Regex)
		for _, val := range o.Values {
			inner.WriteString(fmt.Sprintf("|^%s$", val))
		}
		manager.AddStandardImport("regexp", "")
		ans = append(ans, fmt.Sprintf(`stringvalidator.RegexMatches(regexp.MustCompile(%q), ""),`, inner.String()))
	} else if o.Regex != nil {
		manager.AddStandardImport("regexp", "")
		ans = append(ans, fmt.Sprintf(`stringvalidator.RegexMatches(regexp.MustCompile(%q), ""),`, *o.Regex))
	} else if len(o.Values) != 0 {
		var inner strings.Builder
		for vnum, val := range o.Values {
			if vnum != 0 {
				inner.WriteString(", ")
			}
			inner.WriteString("\"")
			inner.WriteString(val)
			inner.WriteString("\"")
		}
		ans = append(ans, fmt.Sprintf("stringvalidator.OneOf(%s),", inner.String()))
	}

	return ans, manager, nil
}
