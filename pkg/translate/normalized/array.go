package normalized

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/imports"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
)

var (
	_ Item = &Array{Spec: &Bool{}}
	_ Item = &Array{Spec: &Float{}}
	_ Item = &Array{Spec: &Int{}}
	_ Item = &Array{Spec: &Object{}}
	_ Item = &Array{Spec: &String{}}
)

type Array struct {
	Parent      Item   `json:"-" yaml:"-"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	Summary     string `json:"summary" yaml:"summary"`
	ReadOnly    *bool  `json:"read_only" yaml:"read_only"`
	Required    *bool  `json:"required" yaml:"required"`

	SdkPath                 []string `json:"sdk_path" yaml:"sdk_path"`
	Reference               string   `json:"reference" yaml:"reference"`
	UnderscoreName          string   `json:"underscore_name" yaml:"underscore_name"`
	CamelCaseName           string   `json:"camel_case_name" yaml:"camel_case_name"`
	DeriveNamesFrom         string   `json:"derive_names_from" yaml:"derive_names_from"`
	ShortName               string   `json:"-" yaml:"-"`
	ClassName               string   `json:"class_name" yaml:"class_name"`
	Location                string   `json:"-" yaml:"-"`
	DeriveResourceNamesFrom string   `json:"derive_resource_names_from" yaml:"derive_resource_names_from"`

	Spec      Item   `json:"spec" yaml:"spec"`
	Unordered *bool  `json:"unordered" yaml:"unordered"`
	MinItems  *int64 `json:"min_items" yaml:"min_items"`
	MaxItems  *int64 `json:"max_items" yaml:"max_items"`

	Namer *naming.Namer
}

func (o *Array) Path() []string {
	if o == nil {
		return nil
	}

	if o.Parent != nil {
		return append(o.Parent.Path(), o.Name)
	}

	return []string{o.Name}
}

func (o *Array) Copy() Item {
	if o == nil {
		return nil
	}

	ans := Array{
		Parent:      o.Parent,
		Name:        o.Name,
		Description: o.Description,
		Summary:     o.Summary,

		SdkPath:                 append([]string(nil), o.SdkPath...),
		Reference:               o.Reference,
		UnderscoreName:          o.UnderscoreName,
		CamelCaseName:           o.CamelCaseName,
		DeriveNamesFrom:         o.DeriveNamesFrom,
		ShortName:               o.ShortName,
		ClassName:               o.ClassName,
		Location:                o.Location,
		DeriveResourceNamesFrom: o.DeriveResourceNamesFrom,
	}

	if o.ReadOnly != nil {
		x := *o.ReadOnly
		ans.ReadOnly = &x
	}

	if o.Required != nil {
		x := *o.Required
		ans.Required = &x
	}

	if o.Spec != nil {
		ans.Spec = o.Spec.Copy()
	}

	if o.Unordered != nil {
		x := *o.Unordered
		ans.Unordered = &x
	}

	if o.MinItems != nil {
		x := *o.MinItems
		ans.MinItems = &x
	}

	if o.MaxItems != nil {
		x := *o.MaxItems
		ans.MaxItems = &x
	}

	return &ans
}

func (o *Array) ApplyUserConfig(vi Item) {
	if vi == nil {
		return
	}

	v := vi.(*Array)
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

	if len(v.SdkPath) != 0 {
		o.SdkPath = append([]string(nil), v.SdkPath...)
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

	if v.ClassName != "" {
		o.ClassName = v.ClassName
	}

	if v.Unordered != nil {
		x := *v.Unordered
		o.Unordered = &x
	}

	if v.MinItems != nil {
		x := *v.MinItems
		o.MinItems = &x
	}

	if v.MaxItems != nil {
		x := *v.MaxItems
		o.MaxItems = &x
	}
}

func (o *Array) String() string {
	return fmt.Sprintf("Array:%q un:%q ccn:%q min:%v max:%v", o.Name, o.UnderscoreName, o.CamelCaseName, o.MinItems, o.MaxItems)
}

func (o *Array) NameAs(style int) string {
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

func (o *Array) GolangType(includeShortName bool, schemas map[string]Item) (string, error) {
	if o == nil || o.Spec == nil {
		return "", fmt.Errorf("array is nil")
	}

	if o.Reference != "" {
		v, ok := schemas[o.Reference]
		if !ok {
			return "", fmt.Errorf("array:%s ref:%s is not present", o.Name, o.Reference)
		}
		switch x := v.(type) {
		case *Array:
			return fmt.Sprintf("%s.%s", x.ShortName, x.ClassName), nil
		default:
			return "", fmt.Errorf("array:%s ref:%s is array, but schema is %T", o.Name, o.Reference, x)
		}
	}

	ans, err := o.Spec.GolangType(includeShortName, schemas)
	return "[]" + ans, err
}

func (o *Array) ValidatorString(_ bool) string {
	if o == nil {
		return ""
	}

	var b strings.Builder
	if o.MinItems != nil && o.MaxItems != nil {
		b.WriteString(fmt.Sprintf(" List must contain at least %d elements and at most %d elements.", *o.MinItems, *o.MaxItems))
	} else if o.MinItems != nil {
		b.WriteString(fmt.Sprintf(" List must contain at least %d elements.", *o.MinItems))
	} else if o.MaxItems != nil {
		b.WriteString(fmt.Sprintf(" List must contain at most %d elements.", *o.MaxItems))
	}

	if o.Spec != nil {
		more := o.Spec.ValidatorString(false)
		if more != "" {
			b.WriteString(" Individual elements in this list are subject to additional validation.")
			b.WriteString(more)
		}
	}

	return b.String()
}

func (o *Array) GetInternalName() string   { return o.Name }
func (o *Array) GetUnderscoreName() string { return o.UnderscoreName }
func (o *Array) GetCamelCaseName() string  { return o.CamelCaseName }

func (o *Array) SchemaInit(shortName, loc string) error {
	o.ShortName = shortName
	if o.SdkPath == nil {
		o.SdkPath = naming.SchemaNameToSdkPath(loc)
	}

	return nil
}

func (o *Array) GetShortName() string {
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

func (o *Array) Items() []Item {
	return o.Spec.Items()
}

func (o *Array) GetItems(isTop, all bool, schemas map[string]Item) ([]Item, error) {
	if o == nil {
		return nil, fmt.Errorf("array is nil")
	} else if o.Spec == nil {
		return nil, fmt.Errorf("array spec is nil")
	}

	ans := make([]Item, 0, 2)

	if isTop {
		if o.Reference != "" {
			v2, ok := schemas[o.Reference]
			if !ok {
				return nil, fmt.Errorf("array:%s ref:%s not found", o.Name, o.Reference)
			}
			return v2.GetItems(true, all, schemas)
		}
		ans = append(ans, o)
		v2, err := o.Spec.GetItems(false, all, schemas)
		ans = append(ans, v2...)
		return ans, err
	}

	if !all && o.Reference != "" || o.Spec.GetReference() != "" {
		return nil, nil
	}

	v := o
	if o.Reference != "" {
		v2, ok := schemas[o.Reference]
		if !ok {
			return nil, fmt.Errorf("array:%s ref:%s not present", o.Name, o.Reference)
		}
		switch x := v2.(type) {
		case *Array:
			v = x
		default:
			return nil, fmt.Errorf("array:%s ref:%s is array but schema is %T", o.Name, o.Reference, x)
		}
		ans = append(ans, v)
	}
	v2, err := v.Spec.GetItems(false, all, schemas)
	ans = append(ans, v2...)
	return ans, err
}

func (o *Array) GetSdkImports(all bool, schemas map[string]Item) (map[string]bool, error) {
	if o == nil {
		return nil, fmt.Errorf("array is nil")
	} else if o.Spec == nil {
		return nil, fmt.Errorf("array spec is nil")
	}

	ans := make(map[string]bool)
	v := o
	if v.Reference != "" {
		ans[v.Reference] = true
		if !all {
			return ans, nil
		}
		v2, ok := schemas[v.Reference]
		if !ok {
			return nil, fmt.Errorf("array:%s ref:%s not found", v.Name, v.Reference)
		}
		switch x := v2.(type) {
		case *Array:
			v = x
		default:
			return nil, fmt.Errorf("array:%s ref:%s is array but schema is %T", v.Name, v.Reference, x)
		}
	}

	v2, err := v.Spec.GetSdkImports(all, schemas)
	if err != nil {
		return nil, fmt.Errorf("Error getting sdk imports for array %q: %s", v.Name, err)
	}

	for key, val := range v2 {
		ans[key] = val
	}

	return ans, nil
}

func (o *Array) ToGolangSdkString(prefix, suffix string, schemas map[string]Item) (string, error) {
	if o == nil {
		return "", fmt.Errorf("array is nil")
	}

	fm := templateFuncMap(1, false, schemas)
	fm["Prefix"] = func() string { return prefix }
	fm["Suffix"] = func() string { return suffix }

	t := template.Must(
		template.New(
			"to-golang-sdk-array",
		).Funcs(
			fm,
		).Parse(`
{{- /* Begin */ -}}
// {{ Prefix }}{{ .ClassName }}{{ Suffix }} type.
// ShortName: {{ .ShortName }}
type {{ Prefix }}{{ .ClassName }}{{ Suffix }} []
{{- GetGolangType .Spec }}
{{- /* End */ -}}`,
		),
	)

	var b strings.Builder
	err := t.Execute(&b, o)

	return b.String(), err
}

func (o *Array) SchemaReferences() []string {
	if o == nil {
		return nil
	}

	if o.Reference != "" && strings.HasPrefix(o.Reference, SchemaPrefix) {
		return []string{o.Reference}
	}

	return o.Spec.SchemaReferences()
}

func (o *Array) ApplyParameterConfig(loc string, req bool) error {
	o.Location = loc
	o.Required = &req

	return nil
}

func (o *Array) GetLocation() string  { return o.Location }
func (o *Array) GetReference() string { return o.Reference }
func (o *Array) GetSdkPath() []string { return o.SdkPath }
func (o *Array) PackageName() string  { return o.SdkPath[len(o.SdkPath)-1] }
func (o *Array) ToGolangSdkQueryParam() (string, bool, error) {
	if o == nil {
		return "", false, fmt.Errorf("array is nil")
	} else if o.Spec == nil {
		return "", false, fmt.Errorf("array spec is nil")
	}

	var b strings.Builder
	var err error
	var t *template.Template

	needStrconv := false
	fm := template.FuncMap{
		"Name":          func() string { return o.Name },
		"CamelCaseName": func() string { return o.CamelCaseName },
	}

	switch x := o.Spec.(type) {
	case *Bool:
		if x.IsObjectBool != nil && *x.IsObjectBool {
			return "", false, fmt.Errorf("is object bool?")
		}
		needStrconv = true
		t = template.Must(
			template.New(
				"bool-array-to-golang-sdk-param",
			).Funcs(
				fm,
			).Parse(`
{{- /* Begin */ -}}
{{ "    " -}}
    for _, x := range input.{{ CamelCaseName }} {
        uv.Add("{{ .Name }}", strconv.FormatBool(x))
    }
{{- /* End */ -}}`,
			),
		)
	case *Int:
		needStrconv = true
		t = template.Must(
			template.New(
				"int-array-to-golang-sdk-param",
			).Funcs(
				fm,
			).Parse(`
{{- /* Begin */ -}}
{{ "    " -}}
    for _, x := range input.{{ CamelCaseName }} {
        uv.Add("{{ .Name }}", strconv.FormatInt(x, 10))
    }
{{- /* End */ -}}`,
			),
		)
	case *Float:
		needStrconv = true
		t = template.Must(
			template.New(
				"float-array-to-golang-sdk-param",
			).Funcs(
				fm,
			).Parse(`
{{- /* Begin */ -}}
{{ "    " -}}
    for _, x := range input.{{ CamelCaseName }} {
        uv.Add("{{ .Name }}", strconv.FormatFloat(x, 'g', -1, 64))
    }
{{- /* End */ -}}`,
			),
		)
	case *String:
		t = template.Must(
			template.New(
				"string-array-to-golang-sdk-param",
			).Funcs(
				fm,
			).Parse(`
{{- /* Begin */ -}}
{{ "    " -}}
    for _, x := range input.{{ CamelCaseName }} {
        uv.Add("{{ .Name }}", x)
    }
{{- /* End */ -}}`,
			),
		)
	default:
		return "", false, fmt.Errorf("type %T unsupported", x)
	}

	err = t.Execute(&b, o.Spec)

	return b.String(), needStrconv, err
}

func (o *Array) ToGolangSdkPathParam() (string, bool, error) {
	return "", false, fmt.Errorf("array cannot be a path param")
}

func (o *Array) Rename(v string) {
	o.UnderscoreName = naming.Underscore("", v, "output")
	o.CamelCaseName = naming.CamelCase("", v, "Output", true)
	if o.ClassName == naming.DelayNaming {
		o.ClassName = o.CamelCaseName
	}
	if o.Description == "" {
		o.Description = fmt.Sprintf("handles output for the %s function.", v)
	}
}

func (o *Array) TerraformModelType(prefix, defaultShortName string, schemas map[string]Item) (string, error) {
	v, err := ItemLookup(o, schemas)
	if err != nil {
		return "", err
	}
	arr := v.(*Array)
	if arr.Spec == nil {
		return "", fmt.Errorf("array spec is nil")
	}

	aspec, err := ItemLookup(arr.Spec, schemas)
	if err != nil {
		return "", fmt.Errorf("array.Spec lookup failed: %s", err)
	}

	switch aspec.(type) {
	case *Bool:
		return "types.List", nil
	case *Int:
		return "types.List", nil
	case *Float:
		return "types.List", nil
	case *String:
		return "types.List", nil
	case *Array:
		return "", fmt.Errorf("TODO: array of arrays")
	case *Object:
		sv, err := aspec.TerraformModelType(prefix, defaultShortName, schemas)
		if err != nil {
			return "", fmt.Errorf("err getting array spec tf model type: %s", err)
		}

		return "[]" + sv, nil
	}

	return "", fmt.Errorf("not sure what the spec is: %T", aspec)
}

func (o *Array) TflogString() (string, error) { return "", nil }
func (o *Array) IsRequired() bool             { return o.Required != nil && *o.Required }
func (o *Array) IsReadOnly() bool             { return o.ReadOnly != nil && *o.ReadOnly }
func (o *Array) HasDefault() bool             { return false }
func (o *Array) ClearDefault()                {}
func (o *Array) GetObjects(schemas map[string]Item) ([]*Object, error) {
	if o == nil {
		return nil, fmt.Errorf("array is nil")
	}

	v := o
	if o.Reference != "" {
		ref, ok := schemas[o.Reference]
		if !ok {
			return nil, fmt.Errorf("Array ref %q not present", o.Reference)
		}
		ar, ok := ref.(*Array)
		if !ok {
			return nil, fmt.Errorf("Array ref %q is %T not *Array", ref)
		}
		v = ar
	}

	if v.Spec == nil {
		return nil, fmt.Errorf("Array spec is nil")
	}

	return v.Spec.GetObjects(schemas)
}

func (o *Array) RenderTerraformDefault() (string, map[string]string, error) {
	return "", nil, fmt.Errorf("array doesn't support default rendering")
}

func (o *Array) EncryptedParams() []*String { return o.Spec.EncryptedParams() }

func (o *Array) RootParent() Item {
	if o.Parent != nil {
		return o.Parent.RootParent()
	}
	return o
}

func (o *Array) EncHasName() (bool, error) { return false, fmt.Errorf("this is not a string") }
func (o *Array) GetParent() Item           { return o.Parent }
func (o *Array) SetParent(i Item)          { o.Parent = i }
func (o *Array) IsEncrypted() bool         { return o.Spec.IsEncrypted() }

func (o *Array) HasEncryptedItems(schemas map[string]Item) (bool, error) {
	lo, err := ItemLookup(o, schemas)
	if err != nil {
		return false, err
	}
	arr, ok := lo.(*Array)
	if !ok {
		return false, fmt.Errorf("lookup is no array: %T", lo)
	}
	if arr.Spec == nil {
		return false, fmt.Errorf("lookup array has empty spec")
	}
	los, err := ItemLookup(arr.Spec, schemas)
	if err != nil {
		return false, err
	}

	return los.HasEncryptedItems(schemas)
}

func (o *Array) GetEncryptionKey(_, _ string, _ bool, _ byte) (string, error) {
	return "", fmt.Errorf("array not encrypted")
}

func (o *Array) RenderTerraformValidation() ([]string, *imports.Manager, error) {
	if o == nil {
		return nil, nil, fmt.Errorf("array is nil")
	}

	ans := make([]string, 0, 5)

	manager := imports.NewManager()
	manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework-validators/listvalidator", "")

	// MinItems and MaxItems.
	if o.MinItems != nil && o.MaxItems != nil {
		ans = append(ans, fmt.Sprintf("listvalidator.SizeBetween(%d, %d),", *o.MinItems, *o.MaxItems))
	} else if o.MinItems != nil {
		ans = append(ans, fmt.Sprintf("listvalidator.SizeAtLeast(%d),", *o.MinItems))
	} else if o.MaxItems != nil {
		ans = append(ans, fmt.Sprintf("listvalidator.SizeAtMost(%d),", *o.MaxItems))
	}

	if o.Spec != nil {
		var specFunc string
		switch o.Spec.(type) {
		case *Int:
			specFunc = "ValueInt64sAre"
		case *Float:
			specFunc = "ValueFloat64sAre"
		case *String:
			specFunc = "ValueStringsAre"
		}

		if specFunc != "" {
			specAns, specManager, specErr := o.Spec.RenderTerraformValidation()
			manager.Merge(specManager)
			if specErr != nil {
				return nil, nil, specErr
			}
			if len(specAns) != 0 {
				ans = append(ans, fmt.Sprintf("listvalidator.%s(", specFunc))
				for _, sav := range specAns {
					ans = append(ans, sav)
				}
				ans = append(ans, "),")
			}
		}
	}

	return ans, manager, nil
}
