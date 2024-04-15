package normalized

import (
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/imports"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
)

var (
	_ Item = &Object{}
)

type Object struct {
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
	ShortName               string   `json:"short_name" yaml:"short_name"`
	ClassName               string   `json:"class_name" yaml:"class_name"`
	Location                string   `json:"-" yaml:"-"`
	DeriveResourceNamesFrom string   `json:"derive_resource_names_from" yaml:"derive_resource_names_from"`

	Params map[string]Item `json:"-" yaml:"-"`
	OneOf  []string        `json:"one_of" yaml:"one_of"`

	Namer *naming.Namer
}

func (o *Object) Path() []string {
	if o.Parent != nil {
		return append(o.Parent.Path(), o.Name)
	}

	return []string{o.Name}
}

func (o *Object) Copy() Item {
	if o == nil {
		return nil
	}

	ans := Object{
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

		OneOf: append([]string(nil), o.OneOf...),
	}

	if o.ReadOnly != nil {
		x := *o.ReadOnly
		ans.ReadOnly = &x
	}

	if o.Required != nil {
		x := *o.Required
		ans.Required = &x
	}

	for key, value := range o.Params {
		ans.Params[key] = value.Copy()
	}

	return &ans
}

func (o *Object) ApplyUserConfig(vi Item) {
	if vi == nil {
		return
	}

	v := vi.(*Object)
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

	if len(v.OneOf) != 0 {
		o.OneOf = append([]string(nil), v.OneOf...)
	}
}

func (o *Object) NameAs(style int) string {
	switch style {
	case 0:
		return o.UnderscoreName
	case 1:
		return o.CamelCaseName
	default:
		return o.Name
	}
}

// style=0 means sort by underscore name
// style=1 means sort by golang name
// style=2 means sort by JSON tag
func (o *Object) OrderedParams(style int) []string {
	v := make([]string, 0, len(o.Params))
	namemap := make(map[string]string)

	for orig := range o.Params {
		stylized := o.Params[orig].NameAs(style)
		namemap[stylized] = orig
		v = append(v, stylized)
	}

	sort.Strings(v)

	ans := make([]string, 0, len(v))
	for _, stylized := range v {
		ans = append(ans, namemap[stylized])
	}

	return ans
}

func (o *Object) String() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("Object:%q un:%q ccn:%q", o.Name, o.UnderscoreName, o.CamelCaseName))
	s.WriteString(fmt.Sprintf(" OneOf:%#v", o.OneOf))

	s.WriteString(fmt.Sprintf(" params:%d", len(o.Params)))
	if len(o.Params) != 0 {
		s.WriteString("\n")
		for _, key := range o.OrderedParams(2) {
			value := o.Params[key]
			s.WriteString("\n -")
			s.WriteString(fmt.Sprintf(" key:%q", key))
			s.WriteString(fmt.Sprintf(" %s", value))
		}
	}

	return s.String()
}

func (o *Object) GolangType(includeShortName bool, schemas map[string]Item) (string, error) {
	if o == nil {
		return "", fmt.Errorf("object is nil")
	}

	if o.Reference != "" {
		v, ok := schemas[o.Reference]
		if !ok {
			return "", fmt.Errorf("obj:%s ref:%s is not present", o.Name, o.Reference)
		}
		return v.GolangType(true, schemas)
	}

	if includeShortName {
		return fmt.Sprintf("%s.%s", o.GetShortName(), o.ClassName), nil
	}

	return o.ClassName, nil
}

func (o *Object) ValidatorString(_ bool) string {
	// TODO: revisit this.
	return ""
}

func (o *Object) GetInternalName() string   { return o.Name }
func (o *Object) GetUnderscoreName() string { return o.UnderscoreName }
func (o *Object) GetCamelCaseName() string  { return o.CamelCaseName }

func (o *Object) SchemaInit(shortName, loc string) error {
	o.ShortName = shortName
	if o.SdkPath == nil {
		o.SdkPath = naming.SchemaNameToSdkPath(loc)
	}

	return nil
}

func (o *Object) GetShortName() string {
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

func (o *Object) Items() []Item {
	if o == nil {
		return nil
	}

	ans := make([]Item, 0, 10)
	ans = append(ans, o)

	for _, name := range o.OrderedParams(1) {
		ans = append(ans, o.Params[name].Items()...)
	}

	return ans
}

func (o *Object) GetSdkImports(all bool, schemas map[string]Item) (map[string]bool, error) {
	if o == nil {
		return nil, fmt.Errorf("bool is nil")
	}

	v := o
	ans := make(map[string]bool)
	if v.Reference != "" {
		ans[v.Reference] = true
		if !all {
			return ans, nil
		}

		v2, ok := schemas[v.Reference]
		if !ok {
			return nil, fmt.Errorf("obj:%s ref:%s not found", v.Name, v.Reference)
		}
		switch x := v2.(type) {
		case *Object:
			v = x
		default:
			return nil, fmt.Errorf("obj:%s ref:%s is obj but schema is %T", v.Name, v.Reference, x)
		}
	}

	for _, name := range v.OrderedParams(1) {
		v2, err := v.Params[name].GetSdkImports(all, schemas)
		if err != nil {
			return nil, fmt.Errorf("obj:%s param:%s err: %s", v.Name, name, err)
		}
		for key, val := range v2 {
			ans[key] = val
		}
	}

	return ans, nil
}

func (o *Object) GetItems(isTop, all bool, schemas map[string]Item) ([]Item, error) {
	if o == nil {
		return nil, fmt.Errorf("object is nil")
	}

	if o.Reference != "" && !all && !isTop {
		return nil, nil
	}

	v := o
	if o.Reference != "" {
		v2, ok := schemas[o.Reference]
		if !ok {
			return nil, fmt.Errorf("obj:%s ref:%s not found", o.Name, o.Reference)
		}
		switch x := v2.(type) {
		case *Object:
			v = x
		default:
			return nil, fmt.Errorf("obj:%s ref:%s is obj but schema is %T", o.Name, o.Reference, x)
		}
	}

	ans := make([]Item, 0, len(v.Params)+1)
	ans = append(ans, v)
	for _, name := range v.OrderedParams(1) {
		v2, err := v.Params[name].GetItems(false, all, schemas)
		if err != nil {
			return nil, fmt.Errorf("obj:%s param:%s err: %s", o.Name, name, err)
		}
		ans = append(ans, v2...)
	}

	return ans, nil
}

func (o *Object) ToGolangSdkString(prefix, suffix string, schemas map[string]Item) (string, error) {
	if o == nil {
		return "", fmt.Errorf("Object definition is nil?")
	}

	fm := templateFuncMap(1, false, schemas)
	fm["Prefix"] = func() string { return prefix }
	fm["Suffix"] = func() string { return suffix }

	t := template.Must(
		template.New(
			"to-golang-sdk-string",
		).Funcs(
			fm,
		).Parse(
			sdkGolangClass,
		),
	)

	var b strings.Builder
	err := t.Execute(&b, o)

	return b.String(), err
}

func (o *Object) SchemaReferences() []string {
	if o == nil {
		return nil
	}

	if o.Reference != "" && strings.HasPrefix(o.Reference, SchemaPrefix) {
		return []string{o.Reference}
	}

	ans := make([]string, 0, len(o.Params))

	for name := range o.Params {
		ans = append(ans, o.Params[name].SchemaReferences()...)
	}

	if len(ans) > 0 {
		return ans
	}

	return nil
}

func (o *Object) ApplyParameterConfig(loc string, req bool) error {
	return fmt.Errorf("Object cannot be a parameter")
}

func (o *Object) GetLocation() string  { return o.Location }
func (o *Object) GetReference() string { return o.Reference }
func (o *Object) GetSdkPath() []string { return o.SdkPath }
func (o *Object) PackageName() string  { return o.SdkPath[len(o.SdkPath)-1] }
func (o *Object) ToGolangSdkQueryParam() (string, bool, error) {
	return "", false, fmt.Errorf("not supported for objects")
}
func (o *Object) ToGolangSdkPathParam() (string, bool, error) {
	return "", false, fmt.Errorf("object cannot be a path param")
}

func (o *Object) Rename(v string) {
	o.UnderscoreName = naming.Underscore("", v, "output")
	o.CamelCaseName = naming.CamelCase("", v, "Output", true)
	if o.ClassName == naming.DelayNaming {
		o.ClassName = o.CamelCaseName
	}
	if o.Description == "" {
		o.Description = fmt.Sprintf("handles output for the %s function.", v)
	}
}

func (o *Object) TerraformModelType(prefix, defaultShortName string, schemas map[string]Item) (string, error) {
	v, err := ItemLookup(o, schemas)
	if err != nil {
		return "", err
	}
	obj, ok := v.(*Object)
	if !ok {
		return "", fmt.Errorf("object after lookup was not object: %T", obj)
	}

	sn := obj.GetShortName()
	if sn == "" {
		sn = defaultShortName
	}

	return fmt.Sprintf("%s_%s_%s", prefix, sn, obj.ClassName), nil
}

func (o *Object) TflogString() (string, error) { return "", nil }
func (o *Object) IsRequired() bool             { return o.Required != nil && *o.Required }
func (o *Object) IsReadOnly() bool             { return o.ReadOnly != nil && *o.ReadOnly }
func (o *Object) HasDefault() bool             { return false }
func (o *Object) ClearDefault()                {}
func (o *Object) GetObjects(schemas map[string]Item) ([]*Object, error) {
	if o == nil {
		return nil, fmt.Errorf("object is nil")
	}

	v, err := ItemLookup(o, schemas)
	if err != nil {
		return nil, fmt.Errorf("%s.GetItems failed: %s", o.Name, err)
	}

	obj, ok := v.(*Object)
	if !ok {
		return nil, fmt.Errorf("%s is obj, but lookup was %T", o.Name, v)
	}

	ans := make([]*Object, 0, len(obj.Params)+1)
	ans = append(ans, obj)

	for _, name := range obj.OrderedParams(1) {
		p := obj.Params[name]
		list, err := p.GetObjects(schemas)
		if err != nil {
			return nil, fmt.Errorf("Err on %s.Param[%s]: %s", obj.Name, name, err)
		}
		ans = append(ans, list...)
	}

	return ans, nil
}

func (o *Object) AsTerraformSchema(schemaPrefix, evs string, inputs, outputs, forceNew map[string]bool, suffixes map[string]string, schemas map[string]Item) (string, *imports.Manager, error) {
	if o == nil {
		return "", nil, fmt.Errorf("obj is nil")
	}

	manager := imports.NewManager()
	haveOutputOneOfs := false

	fm := template.FuncMap{
		"SchemaPrefix":         func() string { return schemaPrefix },
		"EncryptedValueSchema": func() string { return evs },
		"ItemDescription": func(x Item) (string, error) {
			val, err := GolangDocstring(x, false, true, suffixes[x.GetInternalName()], schemas)
			if err != nil {
				return val, err
			}

			var parent *Object
			pi := x.GetParent()
			if pi != nil {
				p2, ok := pi.(*Object)
				if ok {
					parent = p2
				}
			}

			if schemaPrefix != "rsschema" || !inputs[x.GetInternalName()] || parent == nil {
				return val, nil
			}

			oolist := append([]string(nil), parent.OneOf...)
			sort.Strings(oolist)

			var b strings.Builder
			b.Grow(200)
			b.WriteString(" Ensure that only one of the following is specified:")

			withOneOf := false
			for num, oo := range oolist {
				if oo == x.GetInternalName() {
					withOneOf = true
				}

				if num != 0 {
					b.WriteString(",")
				}
				b.WriteString(fmt.Sprintf(" `%s`", oo))
			}

			if !withOneOf {
				return val, nil
			}
			return val + b.String(), nil
		},
		"SchemaType": func(i Item) (string, error) {
			switch x := i.(type) {
			case nil:
				return "", fmt.Errorf("item is nil")
			case *Bool:
				return "BoolAttribute", nil
			case *Int:
				return "Int64Attribute", nil
			case *Float:
				return "Float64Attribute", nil
			case *String:
				return "StringAttribute", nil
			case *Object:
				return "SingleNestedAttribute", nil
			case *Array:
				if x.Spec == nil {
					return "", fmt.Errorf("spec is nil")
				}
				switch x.Spec.(type) {
				case *Array:
					return "", fmt.Errorf("TODO: array of arrays")
				case *Object:
					return "ListNestedAttribute", nil
				default:
					return "ListAttribute", nil
				}
				return "", fmt.Errorf("array spec fallthrough?")
			}
			return "", fmt.Errorf("passthrough somehow (%T)?", i)
		},
		"ElementType": func(i Item) (string, error) {
			if i == nil {
				return "", fmt.Errorf("item is nil")
			}
			arr, ok := i.(*Array)
			if !ok {
				return "", nil
			}
			if arr.Spec == nil {
				return "", fmt.Errorf("ElementType: array spec is nil")
			}
			switch arr.Spec.(type) {
			case *Bool:
				return "types.BoolType", nil
			case *Int:
				return "types.Int64Type", nil
			case *Float:
				return "types.Float64Type", nil
			case *String:
				return "types.StringType", nil
			}
			return "", nil
		},
		"ShouldIncludeDefault": func(i Item) bool {
			// Only resources can have defaults.
			if schemaPrefix != "rsschema" {
				return false
			}
			name := i.GetInternalName()
			if !inputs[name] {
				return false
			}
			return i.HasDefault()
		},
		"RenderTerraformDefault": func(i Item) (string, error) {
			base := "github.com/hashicorp/terraform-plugin-framework/resource/schema/"

			switch x := i.(type) {
			case *Bool:
				manager.AddHashicorpImport(base+"booldefault", "")
				return fmt.Sprintf("booldefault.StaticBool(%t)", *x.Default), nil
			case *Int:
				manager.AddHashicorpImport(base+"int64default", "")
				return fmt.Sprintf("int64default.StaticInt64(%d)", *x.Default), nil
			case *Float:
				manager.AddHashicorpImport(base+"float64default", "")
				return fmt.Sprintf("float64default.StaticFloat64(%g)", *x.Default), nil
			case *String:
				manager.AddHashicorpImport(base+"stringdefault", "")
				return fmt.Sprintf("stringdefault.StaticString(%q)", *x.Default), nil
			}

			return "", fmt.Errorf("Unsupported type: %T", i)
		},
		"IsComputed": func(i Item) bool {
			name := i.GetInternalName()
			// If the param is required, it cannot be computed.
			if inputs[name] && i.IsRequired() && !i.HasDefault() {
				return false
			}
			// Inputs with a default are computed.
			if inputs[name] && i.HasDefault() {
				return true
			}
			if inputs[name] && outputs[name] && i.IsReadOnly() {
				return true
			}
			// NOTE:  Turns out that if you set Optional:true and Computed:true
			// on a param the .IsNull() of it will ALWAYS return false when reading
			// it from the plan.  This means that there would be no way to tell if the
			// user configured it or left it empty.
			//
			// So if something is appears in both places then we need to have
			// Computed:false.
			if inputs[name] && outputs[name] {
				return false
			}
			if outputs[name] {
				return true
			}
			return false
		},
		"IsRequired": func(i Item) bool {
			name := i.GetInternalName()
			// A required param with no default is required.
			if inputs[name] && i.IsRequired() && !i.HasDefault() {
				return true
			}
			return false
		},
		"IsOptional": func(i Item) bool {
			name := i.GetInternalName()
			if inputs[name] {
				if i.IsReadOnly() {
					return false
				}
				if i.IsRequired() {
					return i.HasDefault()
				}
				return true
			}
			return false
		},
		"PlanModifierClass": func(i Item) (string, error) {
			manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier", "")
			switch i.(type) {
			case *Bool:
				return "Bool", nil
			case *Int:
				return "Int64", nil
			case *Float:
				return "Float64", nil
			case *String:
				return "String", nil
			}

			return "", fmt.Errorf("Unsupported plan modifier type: %T", i)
		},
		"PlanModifierLibPrefix": func(i Item) (string, error) {
			base := "github.com/hashicorp/terraform-plugin-framework/resource/schema/"
			switch i.(type) {
			case *Bool:
				manager.AddHashicorpImport(base+"boolplanmodifier", "")
				return "bool", nil
			case *Int:
				manager.AddHashicorpImport(base+"int64planmodifier", "")
				return "int64", nil
			case *Float:
				manager.AddHashicorpImport(base+"float64planmodifier", "")
				return "float64", nil
			case *String:
				manager.AddHashicorpImport(base+"stringplanmodifier", "")
				return "string", nil
			}

			return "", fmt.Errorf("Unsupported plan modifier type: %T", i)
		},
		"ShouldIncludePlanModifiers": func(i Item) bool {
			if schemaPrefix != "rsschema" {
				return false
			}

			switch i.(type) {
			case *Object:
				return false
			case *Array:
				return false
			}

			return forceNew[i.GetInternalName()] || i.IsReadOnly()
		},
		"IsForceNew": func(i Item) bool {
			return forceNew[i.GetInternalName()]
		},
		"IsSensitive": func(i Item) bool {
			str, ok := i.(*String)
			if !ok {
				return false
			}
			return str.IsPassword != nil && *str.IsPassword
		},
		"ShouldIncludeValidators": func(i Item) (bool, error) {
			if i == nil {
				return false, fmt.Errorf("no item passed in")
			}

			if schemaPrefix != "rsschema" {
				return false, nil
			}

			if !inputs[i.GetInternalName()] {
				return false, nil
			}

			vstring := i.ValidatorString(false)
			if vstring != "" {
				return true, nil
			}

			if !haveOutputOneOfs {
				for _, oo := range o.OneOf {
					if i.GetInternalName() == oo {
						return true, nil
					}
				}
			}

			return false, nil
		},
		"ExactlyOneOfLines": func(i Item) ([]string, error) {
			if i == nil {
				return nil, fmt.Errorf("exactly one of lines item is nil")
			}

			if haveOutputOneOfs {
				return nil, nil
			}

			shouldOutput := false
			for _, oo := range o.OneOf {
				if oo == i.GetInternalName() {
					shouldOutput = true
					break
				}
			}

			if !shouldOutput {
				return nil, nil
			}

			lines := make([]string, 0, len(o.OneOf)+2)

			var eooValidator string
			switch i.(type) {
			case *Bool:
				eooValidator = "boolvalidator"
			case *Int:
				eooValidator = "int64validator"
			case *Float:
				eooValidator = "float64validator"
			case *String:
				eooValidator = "stringvalidator"
			case *Array:
				eooValidator = "listvalidator"
			case *Object:
				eooValidator = "objectvalidator"
			default:
				return nil, fmt.Errorf("Unsupported item type to ExactlyOneOfLines: %T", i)
			}

			manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/path", "")

			oolist := append([]string(nil), o.OneOf...)
			sort.Strings(oolist)

			lines = append(lines, fmt.Sprintf("%s.ExactlyOneOf(", eooValidator))
			for _, oo := range oolist {
				if oo == i.GetInternalName() {
					lines = append(lines, "path.MatchRelative(),")
				} else {
					lines = append(lines, fmt.Sprintf("path.MatchRelative().AtParent().AtName(%q),", o.Params[oo].GetUnderscoreName()))
				}
			}
			lines = append(lines, "),")

			haveOutputOneOfs = true

			return lines, nil
		},
		"ValidatorClass": func(i Item) (string, error) {
			if i == nil {
				return "", fmt.Errorf("validator item is nil")
			}

			manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/schema/validator", "")

			switch i.(type) {
			case *Bool:
				return "Bool", nil
			case *Int:
				return "Int64", nil
			case *Float:
				return "Float64", nil
			case *Array:
				return "List", nil
			case *String:
				return "String", nil
			case *Object:
				return "Object", nil
			}
			return "", fmt.Errorf("Unsupported validator class type: %T", i)
		},
		"RenderValidators": func(i Item) ([]string, error) {
			if i == nil {
				return nil, fmt.Errorf("render validator item is nil")
			}

			theAns, subManager, err := i.RenderTerraformValidation()
			manager.Merge(subManager)
			return theAns, err

			ans := make([]string, 0, 2)
			switch x := i.(type) {
			case *Array:
				manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework-validators/listvalidator", "")

				// MinItems and MaxItems.
				if x.MinItems != nil && x.MaxItems != nil {
					ans = append(ans, fmt.Sprintf("listvalidator.SizeBetween(%d, %d)", *x.MinItems, *x.MaxItems))
				} else if x.MinItems != nil {
					ans = append(ans, fmt.Sprintf("listvalidator.SizeAtLeast(%d)", *x.MinItems))
				} else if x.MaxItems != nil {
					ans = append(ans, fmt.Sprintf("listvalidator.SizeAtMost(%d)", *x.MaxItems))
				}
			case *Int:
				manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework-validators/int64validator", "")

				// Min and Max.
				if x.Min != nil && x.Max != nil {
					ans = append(ans, fmt.Sprintf("int64validator.Between(%d, %d)", *x.Min, *x.Max))
				} else if x.Min != nil {
					ans = append(ans, fmt.Sprintf("int64validator.AtLeast(%d)", *x.Min))
				} else if x.Max != nil {
					ans = append(ans, fmt.Sprintf("int64validator.AtMost(%d)", *x.Max))
				}
				// Values.
				if len(x.Values) != 0 {
					var inner strings.Builder
					for vnum, val := range x.Values {
						if vnum != 0 {
							inner.WriteString(", ")
						}
						inner.WriteString(fmt.Sprintf("%d", val))
					}
					ans = append(ans, fmt.Sprintf("int64validator.OneOf(%s)", inner.String()))
				}
			case *Float:
				manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework-validators/float64validator", "")

				// Min and Max.
				if x.Min != nil && x.Max != nil {
					ans = append(ans, fmt.Sprintf("float64validator.Between(%g, %g)", *x.Min, *x.Max))
				} else if x.Min != nil {
					ans = append(ans, fmt.Sprintf("float64validator.AtLeast(%g)", *x.Min))
				} else if x.Max != nil {
					ans = append(ans, fmt.Sprintf("float64validator.AtMost(%g)", *x.Max))
				}
				// Values.
				if len(x.Values) != 0 {
					var inner strings.Builder
					for vnum, val := range x.Values {
						if vnum != 0 {
							inner.WriteString(", ")
						}
						inner.WriteString(fmt.Sprintf("%g", val))
					}
					ans = append(ans, fmt.Sprintf("float64validator.OneOf(%s)", inner.String()))
				}
			case *String:
				manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator", "")

				// MinLength and MaxLength.
				if x.MinLength != nil && x.MaxLength != nil {
					ans = append(ans, fmt.Sprintf("stringvalidator.LengthBetween(%d, %d)", *x.MinLength, *x.MaxLength))
				} else if x.MinLength != nil {
					ans = append(ans, fmt.Sprintf("stringvalidator.LengthAtLeast(%d)", *x.MinLength))
				} else if x.MaxLength != nil {
					ans = append(ans, fmt.Sprintf("stringvalidator.LengthAtMost(%d)", *x.MaxLength))
				}
				// Values and regex.
				if len(x.Values) != 0 && x.Regex != nil {
					// Merge static values into a single regex and use that.
					var inner strings.Builder
					inner.WriteString(*x.Regex)
					for _, val := range x.Values {
						inner.WriteString(fmt.Sprintf("|^%s$", val))
					}
					manager.AddStandardImport("regexp", "")
					ans = append(ans, fmt.Sprintf(`stringvalidator.RegexMatches(regexp.MustCompile(%q), "")`, inner.String()))
				} else if x.Regex != nil {
					manager.AddStandardImport("regexp", "")
					ans = append(ans, fmt.Sprintf(`stringvalidator.RegexMatches(regexp.MustCompile(%q), "")`, *x.Regex))
				} else if len(x.Values) != 0 {
					var inner strings.Builder
					for vnum, val := range x.Values {
						if vnum != 0 {
							inner.WriteString(", ")
						}
						inner.WriteString("\"")
						inner.WriteString(val)
						inner.WriteString("\"")
					}
					ans = append(ans, fmt.Sprintf("stringvalidator.OneOf(%s)", inner.String()))
				}
			default:
				return nil, fmt.Errorf("Unsupported type: %T", i)
			}

			return ans, nil
		},
		"IsListNestedAttribute": func(i Item) (bool, error) {
			if i == nil {
				return false, fmt.Errorf("item is nil")
			}
			arr, ok := i.(*Array)
			if !ok {
				return false, nil
			}
			if arr.Spec == nil {
				return false, fmt.Errorf("arr.Spec is nil")
			}
			_, ok = arr.Spec.(*Object)
			return ok, nil
		},
		"IsSingleNestedAttribute": func(i Item) (bool, error) {
			if i == nil {
				return false, fmt.Errorf("item is nil")
			}
			_, ok := i.(*Object)
			return ok, nil
		},
		"GetObjectFrom": func(i Item) (*Object, error) {
			li, err := ItemLookup(i, schemas)
			if err != nil {
				return nil, err
			}
			switch x := li.(type) {
			case *Object:
				return x, nil
			case *Array:
				sli, err := ItemLookup(x.Spec, schemas)
				if err != nil {
					return nil, err
				}
				so, ok := sli.(*Object)
				if !ok {
					return nil, fmt.Errorf("array spec is not an object")
				}
				return so, nil
			}
			return nil, fmt.Errorf("no object found in this item")
		},
		"DescribeObject": func(subobj *Object, pname string) (string, error) {
			// NOTE:  The names in *Object.Params is the names in the API, but the
			// names in the input and output maps are the underscore names.
			//uname := o.Params[pname].GetUnderscoreName()
			oInputs := make(map[string]bool)
			if inputs[pname] {
				for key := range subobj.Params {
					oInputs[key] = true
				}
			}
			oOutputs := make(map[string]bool)
			if outputs[pname] {
				for key := range subobj.Params {
					oOutputs[key] = true
				}
			}
			ans, libs, err := subobj.AsTerraformSchema(schemaPrefix, "", oInputs, oOutputs, nil, nil, schemas)
			manager.Merge(libs)
			return ans, err
		},
		"IO": func() string {
			return fmt.Sprintf("        // inputs:%#v outputs:%#v forceNew:%#v", inputs, outputs, forceNew)
		},
	}

	t := template.Must(
		template.New(
			"obj-as-terraform-schema",
		).Funcs(
			fm,
		).Parse(`
{{- /* Begin */ -}}
{{- $obj := . }}
{{- "        " }}Attributes: map[string] {{ SchemaPrefix }}.Attribute{
{{ IO }}
{{- $evs := EncryptedValueSchema }}
{{- if ne $evs "" }}
{{ $evs }}
{{- end }}
{{- range $pname := $obj.OrderedParams 0 }}
{{- $pp := index $obj.Params $pname }}
            "{{ $pp.GetUnderscoreName }}": {{ SchemaPrefix }}.{{ SchemaType $pp }}{
                Description: {{ printf "%q" (ItemDescription $pp) }},
{{- if IsRequired $pp }}
                Required: true,
{{- end }}
{{- if IsOptional $pp }}
                Optional: true,
{{- end }}
{{- if IsComputed $pp }}
                Computed: true,
{{- end }}
{{- if ShouldIncludeDefault $pp }}
                Default: {{ RenderTerraformDefault $pp }},
{{- end }}
{{- if IsSensitive $pp }}
                Sensitive: true,
{{- end }}
{{- $et := ElementType $pp }}
{{- if ne $et "" }}
                ElementType: {{ $et }},
{{- end }}
{{- if ShouldIncludePlanModifiers $pp }}
                PlanModifiers: []planmodifier.{{ PlanModifierClass $pp }}{
{{- if IsForceNew $pp }}
                    {{ PlanModifierLibPrefix $pp }}planmodifier.RequiresReplace(),
{{- end }}
{{- if $pp.IsReadOnly }}
                    {{ PlanModifierLibPrefix $pp }}planmodifier.UseStateForUnknown(),
{{- end }}
                },
{{- end }}
{{- if ShouldIncludeValidators $pp }}
                Validators: []validator.{{ ValidatorClass $pp }}{
{{- range $vline := RenderValidators $pp }}
                    {{ $vline }}
{{- end }}
{{- range $eoo := ExactlyOneOfLines $pp }}
{{ $eoo }}
{{- end }}
                },
{{- end }}
{{- if IsListNestedAttribute $pp }}
                NestedObject: {{ SchemaPrefix }}.NestedAttributeObject{
{{- $subobj := GetObjectFrom $pp }}
{{ DescribeObject $subobj $pname }}
                },
{{- end }}
{{- if IsSingleNestedAttribute $pp }}
{{- $subobj := GetObjectFrom $pp }}
{{ DescribeObject $subobj $pname }}
{{- end }}
            },
{{- end }}
        },
{{- /* End */ -}}`,
		),
	)

	var b strings.Builder
	err := t.Execute(&b, o)

	return b.String(), manager, err
}

func (o *Object) RenderTerraformDefault() (string, map[string]string, error) {
	return "", nil, fmt.Errorf("object doesn't support default rendering")
}

func (o *Object) EncryptedParams() []*String {
	ans := make([]*String, 0, len(o.Params))

	for _, p := range o.Params {
		ans = append(ans, p.EncryptedParams()...)
	}

	return ans
}

func (o *Object) RootParent() Item {
	if o.Parent != nil {
		return o.Parent.RootParent()
	}
	return o
}

func (o *Object) EncHasName() (bool, error) { return false, fmt.Errorf("this is not a string") }
func (o *Object) GetParent() Item           { return o.Parent }
func (o *Object) SetParent(i Item)          { o.Parent = i }
func (o *Object) IsEncrypted() bool         { return false }

func (o *Object) HasEncryptedItems(schemas map[string]Item) (bool, error) {
	lo, err := ItemLookup(o, schemas)
	if err != nil {
		return false, err
	}
	obj, ok := lo.(*Object)
	if !ok {
		return false, fmt.Errorf("lookup object is not *Object: %T", lo)
	}

	for _, p := range obj.Params {
		isEncrypted, err := p.HasEncryptedItems(schemas)
		if err != nil {
			return false, err
		}
		if isEncrypted {
			return true, nil
		}
	}

	return false, nil
}

func (o *Object) GetEncryptionKey(_, _ string, _ bool, _ byte) (string, error) {
	return "", fmt.Errorf("object not encrypted")
}

func (o *Object) RenderTerraformValidation() ([]string, *imports.Manager, error) {
	manager := imports.NewManager()
	manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator", "")

	return nil, manager, nil
}
