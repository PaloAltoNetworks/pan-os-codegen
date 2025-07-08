package terraform_provider

import (
	"fmt"
	"log"
	"runtime/debug"
	"sort"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/imports"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/object"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/parameter"
)

type Entry struct {
	Name string
	Type string
}

type EntryData struct {
	EntryName string
	Entries   []Entry
}

type parameterEncryptionSpec struct {
	HashingType parameter.HashingType
	HashingFunc string
}

type parameterSpec struct {
	PangoName     *properties.NameVariant
	TerraformName *properties.NameVariant
	ComplexType   string
	Type          string
	Required      bool
	ItemsType     string
	Encryption    *parameterEncryptionSpec
}

type spec struct {
	Name                   string
	HasEntryName           bool
	HasEncryptedParameters bool
	PangoType              string
	PangoReturnType        string
	TerraformType          string
	ModelOrObject          string
	Params                 []parameterSpec
	OneOf                  []parameterSpec
}

func renderSpecsForParams(ancestors []*properties.SpecParam, params []*properties.SpecParam) []parameterSpec {
	var specs []parameterSpec
	for _, elt := range params {
		if elt.IsTerraformOnly() {
			continue
		}

		if elt.IsPrivateParameter() {
			continue
		}

		var encryptionSpec *parameterEncryptionSpec
		if elt.Hashing != nil {
			switch spec := elt.Hashing.Spec.(type) {
			case *parameter.HashingSoloSpec:
				encryptionSpec = &parameterEncryptionSpec{
					HashingType: elt.Hashing.Type,
				}
			case *parameter.HashingClientSpec:
				encryptionSpec = &parameterEncryptionSpec{
					HashingType: elt.Hashing.Type,
					HashingFunc: spec.HashingFunc.Name,
				}
			default:
				panic(fmt.Sprintf("unsupported hashing type: %T", spec))
			}

		}

		var itemsType string
		if elt.Type == "list" {
			itemsType = elt.Items.Type
		}

		specs = append(specs, parameterSpec{
			PangoName:     elt.PangoNameVariant(),
			TerraformName: elt.TerraformNameVariant(),
			ComplexType:   elt.ComplexType(),
			Type:          elt.FinalType(),
			ItemsType:     itemsType,
			Encryption:    encryptionSpec,
		})

	}
	return specs
}

func generateFromTerraformToPangoSpec(pangoTypePrefix string, terraformPrefix string, paramSpec *properties.SpecParam, ancestors []*properties.SpecParam) []spec {
	if paramSpec.Spec == nil {
		return nil
	}

	var specs []spec

	pangoType := fmt.Sprintf("%s%s", pangoTypePrefix, paramSpec.PangoNameVariant().CamelCase)

	pangoReturnType := fmt.Sprintf("%s%s", pangoTypePrefix, paramSpec.PangoNameVariant().CamelCase)
	terraformType := fmt.Sprintf("%s%s", terraformPrefix, paramSpec.TerraformNameVariant().CamelCase)

	ancestors = append(ancestors, paramSpec)

	paramSpecs := renderSpecsForParams(ancestors, paramSpec.Spec.SortedParams())
	oneofSpecs := renderSpecsForParams(ancestors, paramSpec.Spec.SortedOneOf())

	var hasEntryName bool
	if paramSpec.Type == "list" && paramSpec.Items.Type == "entry" {
		hasEntryName = true
	}

	element := spec{
		PangoType:              pangoType,
		PangoReturnType:        pangoReturnType,
		TerraformType:          terraformType,
		ModelOrObject:          "Object",
		HasEncryptedParameters: paramSpec.HasEncryptedResources(),
		HasEntryName:           hasEntryName,
		Params:                 paramSpecs,
		OneOf:                  oneofSpecs,
	}
	specs = append(specs, element)

	renderSpecsForParams := func(params []*properties.SpecParam) {
		for _, elt := range params {
			if elt.Spec == nil || elt.IsPrivateParameter() {
				continue
			}

			terraformPrefix := fmt.Sprintf("%s%s", terraformPrefix, paramSpec.TerraformNameVariant().CamelCase)
			specs = append(specs, generateFromTerraformToPangoSpec(pangoType, terraformPrefix, elt, ancestors)...)
		}
	}

	renderSpecsForParams(paramSpec.Spec.SortedParams())
	renderSpecsForParams(paramSpec.Spec.SortedOneOf())

	return specs
}

func generateFromTerraformToPangoParameter(resourceTyp properties.ResourceType, pkgName string, terraformPrefix string, pangoPrefix string, prop *properties.Normalization, ancestors []*properties.SpecParam) []spec {
	var specs []spec

	var pangoReturnType string
	if ancestors == nil {
		pangoReturnType = fmt.Sprintf("%s.%s", pkgName, prop.EntryOrConfig())
		pangoPrefix = fmt.Sprintf("%s.", pkgName)
	} else {
		pangoReturnType = fmt.Sprintf("%s.%s", pkgName, ancestors[0].Name.CamelCase)
	}

	paramSpecs := renderSpecsForParams(ancestors, prop.Spec.SortedParams())
	oneofSpecs := renderSpecsForParams(ancestors, prop.Spec.SortedOneOf())

	switch resourceTyp {
	case properties.ResourceEntry, properties.ResourceConfig:
		specs = append(specs, spec{
			HasEntryName:    prop.Entry != nil,
			PangoType:       pangoPrefix,
			PangoReturnType: pangoReturnType,
			ModelOrObject:   "Model",
			TerraformType:   terraformPrefix,
			Params:          paramSpecs,
			OneOf:           oneofSpecs,
		})
	case properties.ResourceEntryPlural, properties.ResourceUuid, properties.ResourceUuidPlural:
		terraformPrefix = fmt.Sprintf("%s%s", terraformPrefix, pascalCase(prop.TerraformProviderConfig.PluralName))
		var hasEntryName bool
		if prop.Entry != nil && (resourceTyp != properties.ResourceEntryPlural || prop.TerraformProviderConfig.PluralType != object.TerraformPluralMapType) {
			hasEntryName = true
		}
		specs = append(specs, spec{
			HasEntryName:    hasEntryName,
			PangoType:       pangoPrefix,
			PangoReturnType: pangoReturnType,
			ModelOrObject:   "Object",
			TerraformType:   terraformPrefix,
			Params:          paramSpecs,
			OneOf:           oneofSpecs,
		})
	case properties.ResourceCustom:
		panic("custom resources don't generate anything")
	}

	for _, elt := range prop.Spec.SortedParams() {
		if elt.IsPrivateParameter() {
			continue
		}

		specs = append(specs, generateFromTerraformToPangoSpec(pangoPrefix, terraformPrefix, elt, nil)...)
	}

	for _, elt := range prop.Spec.SortedOneOf() {
		if elt.IsPrivateParameter() {
			continue
		}

		specs = append(specs, generateFromTerraformToPangoSpec(pangoPrefix, terraformPrefix, elt, nil)...)
	}

	return specs
}

const copyToPangoTmpl = `
{{- define "terraformNestedElementsAssign" }}
  {{- with .Parameter }}

  {{- $result := .TerraformName.LowerCamelCase }}
  {{- $diag := .TerraformName.LowerCamelCase | printf "%s_diags" }}
	var {{ $result }}_entry *{{ $.Spec.PangoType }}{{ .PangoName.CamelCase }}
	if o.{{ .TerraformName.CamelCase }} != nil {
		if *obj != nil && (*obj).{{ .PangoName.CamelCase }} != nil {
			{{ $result }}_entry = (*obj).{{ .PangoName.CamelCase }}
		} else {
			{{ $result }}_entry = new({{ $.Spec.PangoType }}{{ .PangoName.CamelCase }})
		}
		// ModelOrObject: {{ $.Spec.ModelOrObject }}
    {{- if eq $.Spec.ModelOrObject "Model" }}
		diags.Append(o.{{ .TerraformName.CamelCase }}.CopyToPango(ctx, ancestors, &{{ $result }}_entry, ev)...)
    {{- else }}
		diags.Append(o.{{ .TerraformName.CamelCase }}.CopyToPango(ctx, append(ancestors, o), &{{ $result }}_entry, ev)...)
    {{- end }}
		if diags.HasError() {
			return diags
		}
	}

  {{- end }}
{{- end }}

{{- define "terraformListElementsAs" }}
  {{- with .Parameter }}
    {{- $pangoType := printf "%s%s" $.Spec.PangoType .PangoName.CamelCase }}
    {{- $terraformType := printf "%s%sObject" $.Spec.TerraformType .TerraformName.CamelCase }}
    {{- $pangoEntries := printf "%s_pango_entries" .TerraformName.LowerCamelCase }}
    {{- $tfEntries := printf "%s_tf_entries" .TerraformName.LowerCamelCase }}
    {{- if eq .ItemsType "entry" }}
		var {{ $tfEntries }} []{{ $terraformType }}
		var {{ $pangoEntries }} []{{ $pangoType }}
	{
		d := o.{{ .TerraformName.CamelCase }}.ElementsAs(ctx, &{{ $tfEntries }}, false)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		for _, elt := range {{ $tfEntries }} {
			var entry *{{ $pangoType }}
			diags.Append(elt.CopyToPango(ctx, append(ancestors, elt), &entry, ev)...)
			if diags.HasError() {
				return diags
			}
			{{ $pangoEntries }} = append({{ $pangoEntries }}, *entry)
		}
	}
    {{- else }}
	{{ $pangoEntries }} := make([]{{ .ItemsType }}, 0)
	diags.Append(o.{{ .TerraformName.CamelCase }}.ElementsAs(ctx, &{{ $pangoEntries }}, false)...)
	if diags.HasError() {
		return diags
	}
    {{- end }}
  {{- end }}
{{- end }}

{{- define "renderStringAsMemberAssignment" }}
  {{- with .Parameter }}
    {{- $pangoType := printf "%s%s" $.Spec.PangoType .PangoName.CamelCase }}
    {{- $pangoEntries := printf "%s_pango_entries" .TerraformName.LowerCamelCase }}
    {{ $pangoEntries }} := []string{o.{{ .TerraformName.CamelCase }}.ValueString()}
  {{- end }}
{{- end }}

{{- define "renderSimpleAssignment" }}
  {{- if .Encryption }}
	valueKey, err := CreateXpathForAttributeWithAncestors(ancestors, "{{ .TerraformName.Original }}")
	if err != nil {
		diags.AddError("Failed to create encrypted values state key", err.Error())
		return diags
	}

	var {{ .TerraformName.LowerCamelCase }}_value *string

    {{- if eq .Encryption.HashingType "client" }}
	stateValue, found := ev.GetPlaintextValue(valueKey)
	if !found || stateValue != o.{{ .TerraformName.CamelCase }}.Value{{ CamelCaseType .Type }}() {
		hashed, err := {{ .Encryption.HashingFunc }}(o.{{ .TerraformName.CamelCase }}.Value{{ CamelCaseType .Type }}())
		if err != nil {
			diags.AddError("Failed to hash sensitive value", err.Error())
			return diags
		}

		err = ev.StoreEncryptedValue(valueKey, "{{ .Encryption.HashingType }}", hashed)
		if err != nil {
			diags.AddError("Failed to manage encrypted values state", err.Error())
			return diags
		}

		err = ev.StorePlaintextValue(valueKey, "{{ .Encryption.HashingType }}", o.{{ .TerraformName.CamelCase }}.ValueString())
		if err != nil {
			diags.AddError("Failed to manage encrypted values state", err.Error())
			return diags
		}

		{{ .TerraformName.LowerCamelCase }}_value = &hashed
	} else {
		{{ .TerraformName.LowerCamelCase }}_value = &stateValue
	}
    {{- else }}
	err = ev.StorePlaintextValue(valueKey, "{{ .Encryption.HashingType }}", o.{{ .TerraformName.CamelCase }}.ValueString())
	if err != nil {
		diags.AddError("Failed to manage encrypted values state", err.Error())
		return diags
	}
	{{ .TerraformName.LowerCamelCase }}_value = o.{{ .TerraformName.CamelCase }}.Value{{ CamelCaseType .Type }}Pointer()
    {{- end }}
  {{- else }}
	{{ .TerraformName.LowerCamelCase }}_value := o.{{ .TerraformName.CamelCase }}.Value{{ CamelCaseType .Type }}Pointer()
  {{- end }}
{{- end }}

{{- range .Specs }}
{{- $spec := . }}
func (o *{{ .TerraformType }}{{ .ModelOrObject }}) CopyToPango(ctx context.Context, ancestors []Ancestor, obj **{{ .PangoReturnType }}, ev *EncryptedValuesManager) diag.Diagnostics {
	var diags diag.Diagnostics
  {{- range .Params }}
    {{- $terraformType := printf "%s%s" $spec.TerraformType .TerraformName.CamelCase }}
    {{- if eq .ComplexType "string-as-member" }}
      {{- template "renderStringAsMemberAssignment" Map "Parameter" . "Spec" $spec }}
    {{- else if eq .Type "" }}
      {{- $pangoType := printf "%sObject" $spec.PangoType }}
	{{- template "terraformNestedElementsAssign" Map "Parameter" . "Spec" $spec }}
    {{- else if or (eq .Type "list") (eq .Type "set") }}
      {{- $pangoType := printf "%s%s" $spec.PangoType .TerraformName.CamelCase }}
	{{- template "terraformListElementsAs" Map "Parameter" . "Spec" $spec }}
    {{- else }}
        {{- template "renderSimpleAssignment" . }}
    {{- end }}
  {{- end }}

  {{- range .OneOf }}
    {{- if eq .ComplexType "string-as-member" }}
      {{- template "renderStringAsMemberAssignment" Map "Parameter" . "Spec" $spec }}
    {{- else if eq .Type "" }}
      {{- $pangoType := printf "%sObject" $spec.PangoType }}
	{{- template "terraformNestedElementsAssign" Map "Parameter" . "Spec" $spec }}
    {{- else if or (eq .Type "list") (eq .Type "set") }}
	{{- template "terraformListElementsAs" Map "Parameter" . "Spec" $spec }}
    {{- else }}
        {{- template "renderSimpleAssignment" . }}
    {{- end }}
  {{- end }}

  if (*obj) == nil {
	*obj = new({{ .PangoReturnType }})
  }
  {{- if .HasEntryName }}
	(*obj).Name = o.Name.ValueString()
  {{- end }}
  {{- range .Params }}
    {{- if eq .ComplexType "string-as-member" }}
	(*obj).{{ .PangoName.CamelCase }} = {{ .TerraformName.LowerCamelCase }}_pango_entries
    {{- else if eq .Type "" }}
	(*obj).{{ .PangoName.CamelCase }} = {{ .TerraformName.LowerCamelCase }}_entry
    {{- else if or (eq .Type "list") (eq .Type "set") }}
	(*obj).{{ .PangoName.CamelCase }} = {{ .TerraformName.LowerCamelCase }}_pango_entries
    {{- else }}
	(*obj).{{ .PangoName.CamelCase }} = {{ .TerraformName.LowerCamelCase }}_value
    {{- end }}
  {{- end }}

  {{- range .OneOf }}
    {{- if eq .ComplexType "string-as-member" }}
	(*obj).{{ .PangoName.CamelCase }} = {{ .TerraformName.LowerCamelCase }}_pango_entries
    {{- else if eq .Type "" }}
	(*obj).{{ .PangoName.CamelCase }} = {{ .TerraformName.LowerCamelCase }}_entry
    {{- else if or (eq .Type "list") (eq .Type "set") }}
	(*obj).{{ .PangoName.CamelCase }} = {{ .TerraformName.LowerCamelCase }}_pango_entries
    {{- else }}
	(*obj).{{ .PangoName.CamelCase }} = {{ .TerraformName.LowerCamelCase }}_value
    {{- end }}
  {{- end }}

	return diags
}
{{- end }}
`

const copyFromPangoTmpl = `
{{- define "renderFromPangoToTfParameter" }}
  {{- if eq .Type "" }}
	// TODO: Missing implementation
  {{- else if or (eq .Type "list") (eq .Type "set") }}
	{{ .TerraformName.CamelCase }}: {{ .TerraformName.LowerCamelCase }}_list,
  {{- end }}
{{- end }}

{{- define "renderListValueSimple" }}
var {{ .TerraformName.LowerCamelCase }}_list types.List
{
	schema := rsschema.{{ .Type | PascalCase }}Attribute{}
	{{ .TerraformName.LowerCamelCase }}_list, {{ .TerraformName.LowerCamelCase }}_diags := types.ListValueFrom(ctx, obj.{{ .PangoName.CamelCase }}, schema.GetType())
	diags.Append({{ .TerraformName.LowerCamelCase }}_diags...)
}
{{- end }}

{{- define "renderSetValueSimple" }}
var {{ .TerraformName.LowerCamelCase }}_list types.Set
{
	schema := rsschema.{{ .Type | PascalCase }}Attribute{}
	{{ .TerraformName.LowerCamelCase }}_list, {{ .TerraformName.LowerCamelCase }}_diags := types.ValueFrom(ctx, obj.{{ .PangoName.CamelCase }}, schema.GetType())
	diags.Append({{ .TerraformName.LowerCamelCase }}_diags...)
}
{{- end }}

{{- define "renderNestedValues" }}
  {{- range .Spec.SortedParams }}
    {{- $terraformType := printf "%s%s" $.TerraformType (.TerraformName.CamelCase) }}
    {{- if eq .Type "" }}
	// TODO {{ .TerraformName.CamelCase }} {{ .Type }}
    {{- else if (and (or (eq .Type "list") (eq .Type "set")) (eq .ItemsType "entry")) }}
	{{- template "renderListValueEntry" Map "Name" .TerraformName "Type" $terraformType }}
    {{- else if (and (or (eq .Type "list") (eq .Type "set")) (eq .ItemsType "member")) }}
	// TODO: {{ .TerraformName.CamelCase }} {{ .ItemsType }}
    {{- else if (eq .Type "list") }}
	{{- template "renderListValueSimple" Map "Name" .TerraformName "Type" .ItemsType }}
    {{- else if (eq .Type "set") }}
	{{- template "renderSetValueSimple" Map "Name" .TerraformName "Type" .ItemsType }}
    {{- else }}
	// TODO: {{ .TerraformName.CamelCase }} {{ .Type }}
    {{- end }}
  {{- end }}

  {{- range .Spec.SortedOneOf }}
	// TODO: .Spec.SortedOneOf {{ .TerraformName.CamelCase }}
  {{- end }}
{{- end }}

{{- define "renderObjectListElement" }}
	entry := &{{ .TerraformType }} {
  {{- range .Element.Spec.SortedParams }}
	{{- template "renderFromPangoToTfParameter" . }}
  {{- end }}
  {{- range .Element.Spec.SortedOneOf }}
	{{- template "renderFromPangoToTfParameter" . }}
  {{- end }}
	}
	{{ .TfEntries }} = append({{ .TfEntries }}, *entry)
{{- end }}

{{- define "terraformListElementsAsParam" }}
  {{- with .Parameter }}
    {{- $pangoType := printf "%s%s" $.Spec.PangoType .TerraformName.CamelCase }}
    {{- $terraformType := printf "%s%sObject" $.Spec.TerraformType .TerraformName.CamelCase }}
    {{- $terraformList := printf "%s_list" .TerraformName.LowerCamelCase }}
    {{- $pangoEntries := printf "%s_pango_entries" .TerraformName.LowerCamelCase }}
    {{- $tfEntries := printf "%s_tf_entries" .TerraformName.LowerCamelCase }}
    {{- if eq .ItemsType "entry" }}
	var {{ $terraformList }} types.{{ $.ListOrSet }}
	{
		var {{ $tfEntries }} []{{ $terraformType }}
		for _, elt := range obj.{{ .PangoName.CamelCase }} {
			entry := {{ $terraformType }}{
				Name: types.StringValue(elt.Name),
			}
			diags.Append(entry.CopyFromPango(ctx, append(ancestors, entry), &elt, ev)...)
			if diags.HasError() {
				return diags
			}
			{{ $tfEntries }} = append({{ $tfEntries }}, entry)
		}
		var list_diags diag.Diagnostics
		schemaType := o.getTypeFor("{{ .TerraformName.Underscore }}")
		{{ $terraformList }}, list_diags = types.{{ $.ListOrSet }}ValueFrom(ctx, schemaType, {{ $tfEntries }})
		diags.Append(list_diags...)
	}
    {{- else }}
		var {{ .TerraformName.LowerCamelCase }}_list types.{{ $.ListOrSet }}
		{
			var list_diags diag.Diagnostics
			{{ .TerraformName.LowerCamelCase }}_list, list_diags = types.{{ $.ListOrSet }}ValueFrom(ctx, types.{{ .ItemsType | PascalCase }}Type, obj.{{ .PangoName.CamelCase }})
			diags.Append(list_diags...)
			if diags.HasError() {
				return diags
			}
		}
    {{- end }}
  {{- end }}
{{- end }}

{{- define "terraformSetElementsAs" }}
  {{- range .Params }}
    {{- if eq .Type "set" }}
      {{- template "terraformListElementsAsParam" Map "Spec" $ "Parameter" . "ListOrSet" "Set" }}
    {{- end }}
  {{- end }}

  {{- range .OneOf }}
    {{- if eq .Type "set" }}
      {{- template "terraformListElementsAsParam" Map "Spec" $ "Parameter" . "ListOrSet" "Set" }}
    {{- end }}
  {{- end }}
{{- end }}

{{- define "terraformListElementsAs" }}
  {{- range .Params }}
    {{- if eq .Type "list" }}
      {{- template "terraformListElementsAsParam" Map "Spec" $ "Parameter" . "ListOrSet" "List" }}
    {{- end }}
  {{- end }}

  {{- range .OneOf }}
    {{- if eq .Type "list" }}
      {{- template "terraformListElementsAsParam" Map "Spec" $ "Parameter" . "ListOrSet" "List" }}
    {{- end }}
  {{- end }}
{{- end }}

{{- define "terraformCreateEntryAssignmentForParam" }}
  {{- with .Parameter }}
  {{- $result := .TerraformName.LowerCamelCase }}
  {{- $diag := .TerraformName.LowerCamelCase | printf "%s_diags" }}
  var {{ $result }}_object *{{ $.Spec.TerraformType }}{{ .TerraformName.CamelCase }}Object
  if obj.{{ .PangoName.CamelCase }} != nil {
	{{ $result }}_object = new({{ $.Spec.TerraformType }}{{ .TerraformName.CamelCase }}Object)

    {{- if eq $.Spec.ModelOrObject "Model" }}
	diags.Append({{ $result }}_object.CopyFromPango(ctx, ancestors, obj.{{ .PangoName.CamelCase }}, ev)...)
    {{- else }}
	diags.Append({{ $result }}_object.CopyFromPango(ctx, append(ancestors, o), obj.{{ .PangoName.CamelCase }}, ev)...)
    {{- end }}
	if diags.HasError() {
		return diags
	}
  }
  {{- end }}
{{- end }}

{{- define "terraformCreateEntryAssignment" }}
  {{- range .Params }}
    {{- if eq .Type "" }}
      {{- template "terraformCreateEntryAssignmentForParam" Map "Spec" $ "Parameter" . }}
    {{- end }}
  {{- end }}

  {{- range .OneOf }}
    {{- if eq .Type "" }}
      {{- template "terraformCreateEntryAssignmentForParam" Map "Spec" $ "Parameter" . }}
    {{- end }}
  {{- end }}
{{- end }}

{{- define "terraformCreateStringAsMemberValues" }}
  {{- range .Params }}
    {{ if not (eq .ComplexType "string-as-member") }}
      {{- continue }}
    {{- end }}
    {{ .TerraformName.LowerCamelCase }}_value := types.StringValue(obj.{{ .PangoName.CamelCase }}[0])
  {{- end }}

  {{- range .OneOf }}
    {{ if not (eq .ComplexType "string-as-member") }}
      {{- continue }}
    {{- end }}
    {{ .TerraformName.LowerCamelCase }}_value := types.StringValue(*obj.{{ .PangoName.CamelCase }}[0])
  {{- end }}
{{- end }}

{{- define "terraformCreateSimpleValues" }}
  {{- range .Params }}
    {{- $terraformType := printf "types.%s" (.Type | PascalCase) }}
    {{- if (not (or (eq .Type "") (eq .Type "list") (eq .Type "set") (eq .ComplexType "string-as-member"))) }}
	var {{ .TerraformName.LowerCamelCase }}_value {{ $terraformType }}
	if obj.{{ .PangoName.CamelCase }} != nil {
{{- if .Encryption }}
		valueKey, err := CreateXpathForAttributeWithAncestors(ancestors, "{{ .TerraformName.Original }}")
		if err != nil {
			diags.AddError("Failed to create encrypted values state key", err.Error())
			return diags
		}

		if evFromState, found := ev.GetEncryptedValue(valueKey); found && ev.PreferServerState() && *obj.{{  .PangoName.CamelCase }} != evFromState {
			{{ .TerraformName.LowerCamelCase }}_value = types.StringPointerValue(obj.{{ .PangoName.CamelCase }})
		} else if value, found := ev.GetPlaintextValue(valueKey); found {
			{{ .TerraformName.LowerCamelCase }}_value = types.StringValue(value)
		} else {
			diags.AddError("Failed to read encrypted values state", fmt.Sprintf("Missing plaintext value for %s", valueKey))
			return diags
		}

		if !ev.PreferServerState() {
			err = ev.StoreEncryptedValue(valueKey, "{{ .Encryption.HashingType }}", *obj.{{ .PangoName.CamelCase }})
			if err != nil {
				diags.AddError("Failed to store encrypted values state", err.Error())
				return diags
			}
		}


{{- else }}
		{{ .TerraformName.LowerCamelCase }}_value = types.{{ .Type | PascalCase }}Value(*obj.{{ .PangoName.CamelCase }})
{{- end }}
	}
    {{- end }}
  {{- end }}

  {{- range .OneOf }}
    {{- $terraformType := printf "types.%s" (.Type | PascalCase) }}
    {{- if (not (or (eq .Type "") (eq .Type "list") (eq .Type "set") (eq .ComplexType "string-as-member"))) }}
	var {{ .TerraformName.LowerCamelCase }}_value {{ $terraformType }}
	if obj.{{ .PangoName.CamelCase }} != nil {
		{{ .TerraformName.LowerCamelCase }}_value = types.{{ .Type | PascalCase }}Value(*obj.{{ .PangoName.CamelCase }})
	}
    {{- end }}
  {{- end }}
{{- end }}

{{- define "assignFromPangoToTerraform" }}
  {{- with .Parameter }}
  {{- if eq .ComplexType "string-as-member" }}
	o.{{ .TerraformName.CamelCase }} = {{ .TerraformName.LowerCamelCase }}_value
  {{- else if eq .Type "" }}
	o.{{ .TerraformName.CamelCase }} = {{ .TerraformName.LowerCamelCase }}_object
  {{- else if or (eq .Type "list") (eq .Type "set") }}
	o.{{ .TerraformName.CamelCase }} = {{ .TerraformName.LowerCamelCase }}_list
  {{- else }}
	o.{{ .TerraformName.CamelCase }} = {{ .TerraformName.LowerCamelCase }}_value
  {{- end }}
  {{- end }}
{{- end }}

{{- range .Specs }}
{{- $spec := . }}
{{ $terraformType := printf "%s%s" .TerraformType .ModelOrObject }}
func (o *{{ $terraformType }}) CopyFromPango(ctx context.Context, ancestors []Ancestor, obj *{{ .PangoReturnType }}, ev *EncryptedValuesManager) diag.Diagnostics {
	var diags diag.Diagnostics

  {{- template "terraformSetElementsAs" $spec }}
  {{- template "terraformListElementsAs" $spec }}
  {{- template "terraformCreateEntryAssignment" $spec }}
  {{- template "terraformCreateStringAsMemberValues" $spec }}
  {{- template "terraformCreateSimpleValues" $spec }}

  {{- if .HasEntryName }}
	o.Name = types.StringValue(obj.Name)
  {{- end }}
  {{- range .Params }}
    {{- template "assignFromPangoToTerraform" Map "Spec" $spec "Parameter" . }}
  {{- end }}
  {{- range .OneOf }}
    {{- template "assignFromPangoToTerraform" Map "Spec" $spec "Parameter" . }}
  {{- end }}

	return diags
}
{{- end }}
`

func pascalCase(value string) string {
	var parts []string
	if strings.Contains(value, "-") {
		parts = strings.Split(value, "-")
	} else if strings.Contains(value, "_") {
		parts = strings.Split(value, "_")
	} else {
		parts = []string{value}
	}

	caser := cases.Title(language.English)

	var result []string
	for _, elt := range parts {
		result = append(result, caser.String(elt))
	}

	return strings.Join(result, "")
}

const encryptedValuesManagerInitializationTmpl = `
{{- if or (eq .SchemaType "datasource") (eq .Method "create") (eq .Method "import") }}
var encryptedValues []byte
{{- else }}
encryptedValues, diags := req.Private.GetKey(ctx, "encrypted_values")
resp.Diagnostics.Append(diags...)
if resp.Diagnostics.HasError() {
	return
}
{{- end }}
{{- if eq .Method "read" }}
ev, err := NewEncryptedValuesManager(encryptedValues, true)
{{- else }}
ev, err := NewEncryptedValuesManager(encryptedValues, false)
{{- end }}
if err != nil {
	resp.Diagnostics.AddError("Failed to read encrypted values from private state", err.Error())
	return
}
`

type encryptedValuesContext struct {
	SchemaType properties.SchemaType
	Method     string
}

func RenderEncryptedValuesInitialization(schemaTyp properties.SchemaType, spec *properties.Normalization, method string) (string, error) {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("** PANIC: %v", e)
			debug.PrintStack()
			panic(e)
		}
	}()

	data := encryptedValuesContext{
		SchemaType: schemaTyp,
		Method:     method,
	}

	return processTemplate(encryptedValuesManagerInitializationTmpl, "encrypted-values-manager-initialization", data, nil)
}

const encryptedValuesManagerFinalizerTmpl = `
{{- if eq .SchemaType "resource" }}
	payload, err := json.Marshal(ev)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal encrypted values state", err.Error())
		return
	}
	resp.Private.SetKey(ctx, "encrypted_values", payload)
{{- end }}
`

func RenderEncryptedValuesFinalizer(schemaTyp properties.SchemaType, spec *properties.Normalization) (string, error) {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("** PANIC: %v", e)
			debug.PrintStack()
			panic(e)
		}
	}()

	data := encryptedValuesContext{
		SchemaType: schemaTyp,
	}

	return processTemplate(encryptedValuesManagerFinalizerTmpl, "encrypted-values-manager-finalizer", data, nil)
}

func RenderCopyToPangoFunctions(resourceTyp properties.ResourceType, pkgName string, terraformTypePrefix string, property *properties.Normalization) (string, error) {
	if resourceTyp == properties.ResourceCustom {
		return "", nil
	}

	specs := generateFromTerraformToPangoParameter(resourceTyp, pkgName, terraformTypePrefix, "", property, nil)

	type context struct {
		Specs []spec
	}

	data := context{
		Specs: specs,
	}
	funcMap := mergeFuncMaps(commonFuncMap, template.FuncMap{
		"PascalCase": pascalCase,
	})
	return processTemplate(copyToPangoTmpl, "copy-to-pango", data, funcMap)
}

func RenderCopyFromPangoFunctions(resourceTyp properties.ResourceType, pkgName string, terraformTypePrefix string, property *properties.Normalization) (string, error) {
	if resourceTyp == properties.ResourceCustom {
		return "", nil
	}

	specs := generateFromTerraformToPangoParameter(resourceTyp, pkgName, terraformTypePrefix, "", property, nil)

	type context struct {
		Specs []spec
	}

	data := context{
		Specs: specs,
	}

	funcMap := mergeFuncMaps(commonFuncMap, template.FuncMap{
		"PascalCase": pascalCase,
	})
	return processTemplate(copyFromPangoTmpl, "copy-from-pango", data, funcMap)
}

const xpathComponentsGetterTmpl = `
func (o *{{ .StructName }}Model) resourceXpathParentComponents() ([]string, error) {
	var components []string
{{- range .Components }}
  {{- if eq .Type "value" }}
	components = append(components, (o.{{ .Name.CamelCase }}.ValueString()))
  {{- else if eq .Type "entry" }}
	components = append(components, pangoutil.AsEntryXpath(o.{{ .Name.CamelCase }}.ValueString()))
  {{- end }}
{{- end }}
	return components, nil
}
`

func RenderXpathComponentsGetter(structName string, property *properties.Normalization) (string, error) {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("** PANIC: %v", e)
			debug.PrintStack()
			panic(e)
		}
	}()

	type componentSpec struct {
		Type     string
		Name     *properties.NameVariant
		Variants []*properties.NameVariant
	}

	var components []componentSpec
	for _, elt := range property.PanosXpath.Variables {
		if elt.Name == "name" {
			continue
		}

		xpathProperty, err := property.ParameterForPanosXpathVariable(elt)
		if err != nil {
			return "", err
		}

		switch elt.Spec.Type {
		case object.PanosXpathVariableValue:
			components = append(components, componentSpec{
				Type: "value",
				Name: xpathProperty.Name,
			})
		case object.PanosXpathVariableEntry:
			components = append(components, componentSpec{
				Type: "entry",
				Name: xpathProperty.Name,
			})
		case object.PanosXpathVariableStatic:
		default:
			panic(fmt.Sprintf("invalid panos xpath variable type: '%s'", elt.Spec.Type))
		}
	}

	data := struct {
		StructName string
		Components []componentSpec
	}{
		StructName: structName,
		Components: components,
	}

	return processTemplate(xpathComponentsGetterTmpl, "xpath-components", data, commonFuncMap)
}

const renderLocationTmpl = `
{{- range .Locations }}
type {{ .StructName }} struct {
  {{- range .Fields }}
	{{ .Name.CamelCase }} {{ .Type }} {{ range .Tags }}{{ . }} {{ end }}
  {{- end }}
}
{{- end }}
`

type locationStructFieldCtx struct {
	Name          *properties.NameVariant
	TerraformType string
	Type          string
	Tags          []string
}

type locationStructCtx struct {
	StructName string
	Fields     []locationStructFieldCtx
}

func getLocationStructsContext(names *NameProvider, spec *properties.Normalization) []locationStructCtx {
	var locations []locationStructCtx

	if len(spec.Locations) == 0 {
		return nil
	}

	// Create the top location structure that references other locations
	topLocation := locationStructCtx{
		StructName: fmt.Sprintf("%sLocation", names.StructName),
	}

	for _, data := range spec.OrderedLocations() {
		structName := fmt.Sprintf("%s%sLocation", names.StructName, data.Name.CamelCase)
		tfTag := fmt.Sprintf("`tfsdk:\"%s\"`", data.Name.Underscore)
		structType := "types.Object"

		topLocation.Fields = append(topLocation.Fields, locationStructFieldCtx{
			Name:          data.Name,
			TerraformType: structName,
			Type:          structType,
			Tags:          []string{tfTag},
		})

		var fields []locationStructFieldCtx

		for _, i := range spec.Imports {
			if i.Type.CamelCase != data.Name.CamelCase {
				continue
			}

			for _, elt := range i.OrderedLocations() {
				if elt.Required {
					fields = append(fields, locationStructFieldCtx{
						Name: elt.Name,
						Type: "types.String",
						Tags: []string{fmt.Sprintf("`tfsdk:\"%s\"`", elt.Name.Underscore)},
					})
				}
			}
		}

		for _, param := range data.OrderedVars() {
			paramTag := fmt.Sprintf("`tfsdk:\"%s\"`", param.Name.Underscore)
			name := param.Name
			if name.CamelCase == data.Name.CamelCase {
				name = properties.NewNameVariant("name")
				paramTag = "`tfsdk:\"name\"`"
			}
			fields = append(fields, locationStructFieldCtx{
				Name: name,
				Type: "types.String",
				Tags: []string{paramTag},
			})
		}

		location := locationStructCtx{
			StructName: structName,
			Fields:     fields,
		}
		locations = append(locations, location)
	}

	locations = append(locations, topLocation)

	return locations
}

func RenderLocationStructs(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization) (string, error) {
	type context struct {
		Locations []locationStructCtx
	}

	locations := getLocationStructsContext(names, spec)

	data := context{
		Locations: locations,
	}
	return processTemplate(renderLocationTmpl, "render-location-structs", data, commonFuncMap)
}

const locationSchemaGetterTmpl = `
{{- define "renderLocationAttribute" }}
"{{ .Name.Underscore }}": {{ .SchemaType }}{
	Description: "{{ .Description }}",
  {{- if .Required }}
	Required: true,
  {{- else }}
	Optional: true,
  {{- end }}
  {{- if .Computed }}
	Computed: true,
  {{- end }}
  {{- if .Default }}
	Default: {{ .Default.Type }}({{ .Default.Value }}),
  {{- end }}
  {{- if .Attributes }}
	Attributes: map[string]rsschema.Attribute{
    {{- range .Attributes }}
		{{- template "renderLocationAttribute" . }}
    {{- end }}
	},
  {{- end }}
	PlanModifiers: []planmodifier.{{ .ModifierType }}{
		{{ .ModifierType | LowerCase }}planmodifier.RequiresReplace(),
	},
  {{- with .Validators }}
    {{ $package := .Package }}
		Validators: []validator.{{ .ListType }}{
    {{- range .Functions }}
			{{ $package }}.{{ .Function }}(path.Expressions{
      {{- range .Expressions }}
				{{ . }},
      {{- end }}
			}...),
    {{- end }}
		},
  {{- end }}
},
{{- end }}

func {{ .StructName }}LocationSchema() rsschema.Attribute {
  {{- with .Schema }}
	return rsschema.SingleNestedAttribute{
		Description: "{{ .Description }}",
		Required: true,
		Attributes: map[string]rsschema.Attribute{
{{- range .Attributes }}
{{- template "renderLocationAttribute" . }}
{{- end }}
		},
	}
}
  {{- end }}
`

type defaultCtx struct {
	Type  string
	Value string
}

type modifierCtx struct {
	SchemaType string
	Modifiers  []string
}

type validatorFunctionCtx struct {
	Type              string
	Function          string
	FunctionOverriden bool
	Expressions       []string
	Values            []string
}

type validatorCtx struct {
	ListType  string
	Package   string
	Functions []validatorFunctionCtx
}

type attributeCtx struct {
	Package       string
	Name          *properties.NameVariant
	Private       bool
	SchemaType    string
	ExternalType  string
	ElementType   string
	Description   string
	Required      bool
	Computed      bool
	Optional      bool
	Sensitive     bool
	Default       *defaultCtx
	ModifierType  string
	Attributes    []attributeCtx
	PlanModifiers *modifierCtx
	Validators    *validatorCtx
}

type schemaCtx struct {
	IsResource    bool
	ObjectOrModel string
	StructName    string
	ReturnType    string
	Package       string
	Description   string
	Required      bool
	Computed      bool
	Optional      bool
	Sensitive     bool
	Attributes    []attributeCtx
	Validators    *validatorCtx
}

func RenderLocationSchemaGetter(names *NameProvider, spec *properties.Normalization, manager *imports.Manager) (string, error) {
	var attributes []attributeCtx

	if len(spec.Locations) == 0 {
		return "", nil
	}

	var locations []string
	for _, loc := range spec.OrderedLocations() {
		locations = append(locations, loc.Name.Underscore)
	}

	var idx int
	for _, data := range spec.OrderedLocations() {
		var variableAttrs []attributeCtx

		for _, i := range spec.Imports {
			if i.Type.CamelCase != data.Name.CamelCase {
				continue
			}

			for _, elt := range i.OrderedLocations() {
				if elt.Required {
					var defaultValue *defaultCtx
					for varName, variable := range elt.XpathVariables {
						if varName == elt.Name.Original && variable.Default != "" {
							defaultValue = &defaultCtx{
								Type:  "stringdefault.StaticString",
								Value: fmt.Sprintf(`"%s"`, variable.Default),
							}
						}
					}
					variableAttrs = append(variableAttrs, attributeCtx{
						Name:         elt.Name,
						SchemaType:   "rsschema.StringAttribute",
						Required:     defaultValue == nil,
						Optional:     defaultValue != nil,
						Computed:     defaultValue != nil,
						ModifierType: "String",
						Default:      defaultValue,
					})
				}
			}
		}

		for _, variable := range data.OrderedVars() {
			name := variable.Name
			if name.CamelCase == data.Name.CamelCase {
				name = properties.NewNameVariant("name")
			}
			attribute := attributeCtx{
				Name:        name,
				Description: variable.Description,
				SchemaType:  "rsschema.StringAttribute",
				Required:    false,
				Computed:    true,
				Default: &defaultCtx{
					Type:  "stringdefault.StaticString",
					Value: fmt.Sprintf(`"%s"`, variable.Default),
				},
				ModifierType: "String",
			}
			variableAttrs = append(variableAttrs, attribute)
		}

		modifierType := "Object"

		var validators *validatorCtx
		if len(locations) > 1 && idx == 0 {
			var expressions []string
			for _, location := range locations {
				expressions = append(expressions, fmt.Sprintf(`path.MatchRelative().AtParent().AtName("%s")`, location))
			}

			functions := []validatorFunctionCtx{{
				Function:    "ExactlyOneOf",
				Expressions: expressions,
			}}

			typ := data.ValidatorType()
			validatorImport := fmt.Sprintf("github.com/hashicorp/terraform-plugin-framework-validators/%svalidator", typ)
			manager.AddHashicorpImport(validatorImport, "")

			validators = &validatorCtx{
				ListType:  pascalCase(typ),
				Package:   fmt.Sprintf("%svalidator", typ),
				Functions: functions,
			}
		}

		attribute := attributeCtx{
			Name:         data.Name,
			SchemaType:   "rsschema.SingleNestedAttribute",
			Description:  data.Description,
			Required:     false,
			Attributes:   variableAttrs,
			ModifierType: modifierType,
			Validators:   validators,
		}
		attributes = append(attributes, attribute)

		idx += 1
	}

	locationName := properties.NewNameVariant("location")

	topAttribute := attributeCtx{
		Name:         locationName,
		SchemaType:   "rsschema.SingleNestedAttribute",
		Description:  "The location of this object.",
		Required:     true,
		Attributes:   attributes,
		ModifierType: "Object",
	}

	type context struct {
		StructName string
		Schema     attributeCtx
	}

	data := context{
		StructName: names.StructName,
		Schema:     topAttribute,
	}

	return processTemplate(locationSchemaGetterTmpl, "render-location-schema-getter", data, commonFuncMap)
}

type marshallerFieldSpec struct {
	Name       *properties.NameVariant
	Type       string
	StructName string
	Tags       string
}

type marshallerSpec struct {
	StructName string
	Fields     []marshallerFieldSpec
}

func createLocationMarshallerSpecs(names *NameProvider, spec *properties.Normalization) []marshallerSpec {
	var specs []marshallerSpec

	var topFields []marshallerFieldSpec
	for _, loc := range spec.OrderedLocations() {
		topFields = append(topFields, marshallerFieldSpec{
			Name:       loc.Name,
			Type:       "types.Object",
			StructName: fmt.Sprintf("%s%sLocation", names.StructName, loc.Name.CamelCase),
			Tags:       fmt.Sprintf("`json:\"%s,omitempty\"`", loc.Name.Underscore),
		})

		var fields []marshallerFieldSpec
		for _, field := range loc.OrderedVars() {
			name := field.Name
			tag := field.Name.Underscore
			if name.CamelCase == loc.Name.CamelCase {
				name = properties.NewNameVariant("name")
				tag = "name"
			}

			fields = append(fields, marshallerFieldSpec{
				Name: name,
				Type: "string",
				Tags: fmt.Sprintf("`json:\"%s,omitempty\"`", tag),
			})
		}

		// Add import location (e.g. vsys) name to location
		for _, i := range spec.Imports {
			if i.Type.CamelCase != loc.Name.CamelCase {
				continue
			}

			for _, elt := range i.OrderedLocations() {
				if elt.Required {
					fields = append(fields, marshallerFieldSpec{
						Name: elt.Name,
						Type: "string",
						Tags: fmt.Sprintf("`tfsdk:\"%s\"`", elt.Name.Underscore),
					})
				}
			}
		}

		specs = append(specs, marshallerSpec{
			StructName: fmt.Sprintf("%s%sLocation", names.StructName, loc.Name.CamelCase),
			Fields:     fields,
		})
	}

	specs = append(specs, marshallerSpec{
		StructName: fmt.Sprintf("%sLocation", names.StructName),
		Fields:     topFields,
	})

	return specs
}

func RenderLocationMarshallers(names *NameProvider, spec *properties.Normalization) (string, error) {
	var context struct {
		Specs []marshallerSpec
	}
	context.Specs = createLocationMarshallerSpecs(names, spec)

	return processTemplate(locationMarshallersTmpl, "render-location-marshallers", context, commonFuncMap)
}

func RenderImportStateMarshallers(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization) (string, error) {
	// Only singular entries can be imported at the time
	if resourceTyp == properties.ResourceCustom || resourceTyp == properties.ResourceConfig {
		return "", nil
	}

	var context struct {
		Specs []marshallerSpec
	}
	context.Specs = createImportStateMarshallerSpecs(resourceTyp, names, spec)

	return processTemplate(locationMarshallersTmpl, "render-import-state-marshallers", context, commonFuncMap)
}

func RenderCustomImports(spec *properties.Normalization) string {
	template, _ := getCustomTemplateForFunction(spec, "Imports")
	return template
}

func RenderCustomCommonCode(names *NameProvider, spec *properties.Normalization) string {
	template, _ := getCustomTemplateForFunction(spec, "Common")
	return template

}

func generateValidatorFnsMapForVariants(variants []*properties.SpecParam) map[int]*validatorFunctionCtx {
	validatorFns := make(map[int]*validatorFunctionCtx)

	for _, elt := range variants {
		if elt.IsPrivateParameter() {
			continue
		}

		validatorFn := "ExactlyOneOf"
		var validatorFnOverride *string
		if elt.TerraformProviderConfig != nil && elt.TerraformProviderConfig.VariantCheck != nil {
			validatorFnOverride = elt.TerraformProviderConfig.VariantCheck
		}

		validator, found := validatorFns[elt.VariantGroupId]
		if !found {
			validator = &validatorFunctionCtx{
				Type:     "Expressions",
				Function: validatorFn,
			}

			if validatorFnOverride != nil {
				validator.FunctionOverriden = true
				validator.Function = *validatorFnOverride
			}
		} else {
			if validator.FunctionOverriden {
				if validatorFnOverride != nil && validator.Function != *validatorFnOverride {
					panic("invalid yaml spec: parameter codegen override variant_check must be equal within variant group")
				}
			} else if validatorFnOverride != nil {
				validator.Function = *validatorFnOverride
			}
		}

		pathExpr := fmt.Sprintf(`path.MatchRelative().AtParent().AtName("%s")`, elt.TerraformNameVariant().Underscore)
		validator.Expressions = append(validator.Expressions, pathExpr)
		validatorFns[elt.VariantGroupId] = validator
	}

	return validatorFns
}

func createSchemaSpecForParameter(schemaTyp properties.SchemaType, manager *imports.Manager, structPrefix string, packageName string, param *properties.SpecParam, validators *validatorCtx) []schemaCtx {
	var schemas []schemaCtx

	if param.Spec == nil {
		return nil
	}

	var returnType string
	switch param.FinalType() {
	case "":
		returnType = "SingleNestedAttribute"
	case "list", "set":
		switch param.Items.Type {
		case "entry":
			returnType = "NestedAttributeObject"
		}
	}

	structName := fmt.Sprintf("%s%s", structPrefix, param.TerraformNameVariant().CamelCase)

	var attributes []attributeCtx
	if param.HasEntryName() {
		name := properties.NewNameVariant("name")

		attributes = append(attributes, attributeCtx{
			Package:    packageName,
			Name:       name,
			SchemaType: "StringAttribute",
			Required:   true,
		})
	}

	for _, elt := range param.Spec.SortedParams() {
		if elt.IsPrivateParameter() {
			continue
		}

		var functions []validatorFunctionCtx
		if len(elt.EnumValues) > 0 && schemaTyp == properties.SchemaResource {
			var values []string
			for _, elt := range elt.EnumValues {
				values = append(values, elt.Name)
			}

			functions = append(functions, validatorFunctionCtx{
				Type:     "Values",
				Function: "OneOf",
				Values:   values,
			})
		}

		var validators *validatorCtx
		if len(functions) > 0 {
			typ := elt.ValidatorType()
			validatorImport := fmt.Sprintf("github.com/hashicorp/terraform-plugin-framework-validators/%svalidator", typ)
			manager.AddHashicorpImport(validatorImport, "")
			validators = &validatorCtx{
				ListType:  pascalCase(typ),
				Package:   fmt.Sprintf("%svalidator", typ),
				Functions: functions,
			}
		}

		attributes = append(attributes, createSchemaAttributeForParameter(schemaTyp, manager, packageName, elt, validators))
	}

	// Generating schema validation for variants. By default, ExactlyOneOf validation
	// is performed, unless XML API allows for no variant to be provided, in which case
	// validation is performed by ConflictsWith.
	validatorFns := generateValidatorFnsMapForVariants(param.Spec.SortedOneOf())

	var idx int
	for _, elt := range param.Spec.SortedOneOf() {
		if elt.IsPrivateParameter() {
			continue
		}

		var validators *validatorCtx
		if schemaTyp == properties.SchemaResource {
			validatorFn, found := validatorFns[elt.VariantGroupId]
			if found && validatorFn.Function != "Disabled" {
				typ := elt.ValidatorType()
				validatorImport := fmt.Sprintf("github.com/hashicorp/terraform-plugin-framework-validators/%svalidator", typ)
				manager.AddHashicorpImport(validatorImport, "")

				validators = &validatorCtx{
					ListType:  pascalCase(typ),
					Package:   fmt.Sprintf("%svalidator", typ),
					Functions: []validatorFunctionCtx{*validatorFn},
				}

				delete(validatorFns, elt.VariantGroupId)
			}
		}
		attributes = append(attributes, createSchemaAttributeForParameter(schemaTyp, manager, packageName, elt, validators))
		idx += 1
	}

	var isResource bool
	if schemaTyp == properties.SchemaResource {
		isResource = true
	}

	var computed, required bool
	switch schemaTyp {
	case properties.SchemaDataSource:
		computed = true
		required = false
	case properties.SchemaResource, properties.SchemaEphemeralResource:
		computed = param.FinalComputed()
		required = param.FinalRequired()
	case properties.SchemaCommon, properties.SchemaProvider:
		panic("unreachable")
	}

	schemas = append(schemas, schemaCtx{
		IsResource:    isResource,
		ObjectOrModel: "Object",
		Package:       packageName,
		StructName:    structName,
		ReturnType:    returnType,
		Description:   "",
		Required:      required,
		Optional:      !param.FinalRequired(),
		Computed:      computed,
		Sensitive:     param.FinalSensitive(),
		Attributes:    attributes,
		Validators:    validators,
	})

	for _, elt := range param.Spec.SortedParams() {
		if elt.IsPrivateParameter() {
			continue
		}

		var functions []validatorFunctionCtx
		if len(elt.EnumValues) > 0 && schemaTyp == properties.SchemaResource {
			var values []string
			for _, elt := range elt.EnumValues {
				values = append(values, elt.Name)
			}

			functions = append(functions, validatorFunctionCtx{
				Type:     "Values",
				Function: "OneOf",
				Values:   values,
			})
		}

		var validators *validatorCtx
		if len(functions) > 0 {
			typ := elt.ValidatorType()
			validatorImport := fmt.Sprintf("github.com/hashicorp/terraform-plugin-framework-validators/%svalidator", typ)
			manager.AddHashicorpImport(validatorImport, "")
			validators = &validatorCtx{
				ListType:  pascalCase(typ),
				Package:   fmt.Sprintf("%svalidator", typ),
				Functions: functions,
			}
		}

		if elt.Type == "" || ((elt.FinalType() == "list" || elt.FinalType() == "set") && elt.Items.Type == "entry") {
			schemas = append(schemas, createSchemaSpecForParameter(schemaTyp, manager, structName, packageName, elt, validators)...)
		}
	}

	validatorFns = generateValidatorFnsMapForVariants(param.Spec.SortedOneOf())

	for _, elt := range param.Spec.SortedOneOf() {
		if elt.IsPrivateParameter() {
			continue
		}

		if elt.Type == "" || ((elt.FinalType() == "list" || elt.FinalType() == "set") && elt.Items.Type == "entry") {
			var validators *validatorCtx

			validatorFn, found := validatorFns[elt.VariantGroupId]
			if found && validatorFn.Function != "Disabled" {
				validatorImport := fmt.Sprintf("github.com/hashicorp/terraform-plugin-framework-validators/%svalidator", "object")
				manager.AddHashicorpImport(validatorImport, "")
				validators = &validatorCtx{
					ListType:  "Object",
					Package:   "objectvalidator",
					Functions: []validatorFunctionCtx{*validatorFn},
				}
			}
			schemas = append(schemas, createSchemaSpecForParameter(schemaTyp, manager, structName, packageName, elt, validators)...)
		}
	}

	return schemas
}

func createSchemaAttributeForParameter(schemaTyp properties.SchemaType, manager *imports.Manager, packageName string, param *properties.SpecParam, validators *validatorCtx) attributeCtx {
	var schemaType, elementType string

	switch param.ComplexType() {
	case "string-as-member":
		schemaType = "StringAttribute"
	default:
		switch param.FinalType() {
		case "":
			schemaType = "SingleNestedAttribute"
		case "list":
			switch param.Items.Type {
			case "entry":
				schemaType = "ListNestedAttribute"
			case "member":
				schemaType = "ListAttribute"
				elementType = "types.StringType"
			default:
				schemaType = "ListAttribute"
				elementType = fmt.Sprintf("types.%sType", pascalCase(param.Items.Type))
			}
		case "set":
			switch param.Items.Type {
			case "entry":
				schemaType = "SetNestedAttribute"
			case "member":
				schemaType = "SetAttribute"
				elementType = "types.StringType"
			default:
				schemaType = "SetAttribute"
				elementType = fmt.Sprintf("types.%sType", pascalCase(param.Items.Type))
			}
		default:
			schemaType = fmt.Sprintf("%sAttribute", pascalCase(param.Type))
		}
	}

	var defaultValue *defaultCtx
	if schemaTyp == properties.SchemaResource && param.Default != "" {
		defaultImport := fmt.Sprintf("github.com/hashicorp/terraform-plugin-framework/resource/schema/%sdefault", param.DefaultType())
		manager.AddHashicorpImport(defaultImport, "")

		var value string
		switch param.Type {
		case "string":
			value = fmt.Sprintf("\"%s\"", param.Default)
		default:
			value = param.Default
		}
		defaultValue = &defaultCtx{
			Type:  fmt.Sprintf("%sdefault.Static%s", param.Type, pascalCase(param.Type)),
			Value: value,
		}
	}

	var computed, required bool
	switch schemaTyp {
	case properties.SchemaDataSource:
		required = false
		computed = true
	case properties.SchemaResource, properties.SchemaEphemeralResource:
		computed = param.FinalComputed()
		required = param.FinalRequired()
	case properties.SchemaCommon, properties.SchemaProvider:
		panic("unreachable")
	}

	return attributeCtx{
		Package:     packageName,
		Name:        param.TerraformNameVariant(),
		SchemaType:  schemaType,
		ElementType: elementType,
		Description: param.Description,
		Required:    required,
		Optional:    !required,
		Sensitive:   param.FinalSensitive(),
		Default:     defaultValue,
		Computed:    computed,
		Validators:  validators,
	}
}

// createSchemaSpecForUuidModel creates a schema for uuid-type resources, where top-level model describes a list of objects.
func createSchemaSpecForUuidModel(resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, spec *properties.Normalization, packageName string, structName string, manager *imports.Manager) []schemaCtx {
	var schemas []schemaCtx
	var attributes []attributeCtx

	if len(spec.Locations) > 0 {
		location := properties.NewNameVariant("location")

		attributes = append(attributes, attributeCtx{
			Package:    packageName,
			Name:       location,
			Required:   true,
			SchemaType: "SingleNestedAttribute",
		})
	}

	if resourceTyp == properties.ResourceUuidPlural {
		position := properties.NewNameVariant("position")

		attributes = append(attributes, attributeCtx{
			Package:      packageName,
			Name:         position,
			Required:     true,
			SchemaType:   "ExternalAttribute",
			ExternalType: "TerraformPositionObject",
		})
	}

	listNameStr := spec.TerraformProviderConfig.PluralName
	listName := properties.NewNameVariant(listNameStr)

	attributes = append(attributes, attributeCtx{
		Package:     packageName,
		Name:        listName,
		Required:    true,
		Description: spec.TerraformProviderConfig.PluralDescription,
		SchemaType:  "ListNestedAttribute",
	})

	var isResource bool
	if schemaTyp == properties.SchemaResource {
		isResource = true
	}
	schemas = append(schemas, schemaCtx{
		Package:       packageName,
		ObjectOrModel: "Model",
		IsResource:    isResource,
		StructName:    structName,
		ReturnType:    "Schema",
		Attributes:    attributes,
	})

	structName = fmt.Sprintf("%s%s", structName, listName.CamelCase)
	normalizationAttrs, normalizationSchemas := createSchemaSpecForNormalization(resourceTyp, schemaTyp, spec, packageName, structName, manager)

	schemas = append(schemas, schemaCtx{
		Package:       packageName,
		ObjectOrModel: "Object",
		IsResource:    isResource,
		StructName:    structName,
		ReturnType:    "NestedAttributeObject",
		Attributes:    normalizationAttrs,
	})

	schemas = append(schemas, normalizationSchemas...)

	return schemas
}

// createSchemaSpecForEntrySingularModel creates a schema for entry-type singular resources.
//
// Entry-type singular resources are resources that manage a single object in PAN-OS, e.g. `resource_ethernet_interface`.
func createSchemaSpecForEntrySingularModel(resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, spec *properties.Normalization, packageName string, structName string, manager *imports.Manager) []schemaCtx {
	var schemas []schemaCtx
	var attributes []attributeCtx

	if len(spec.Locations) > 0 {
		location := properties.NewNameVariant("location")

		attributes = append(attributes, attributeCtx{
			Package:    packageName,
			Name:       location,
			Required:   true,
			SchemaType: "SingleNestedAttribute",
		})
	}

	normalizationAttrs, normalizationSchemas := createSchemaSpecForNormalization(resourceTyp, schemaTyp, spec, packageName, structName, manager)
	attributes = append(attributes, normalizationAttrs...)

	var isResource bool
	if schemaTyp == properties.SchemaResource {
		isResource = true
	}
	schemas = append(schemas, schemaCtx{
		Package:       packageName,
		ObjectOrModel: "Model",
		IsResource:    isResource,
		StructName:    structName,
		ReturnType:    "Schema",
		Attributes:    attributes,
	})

	schemas = append(schemas, normalizationSchemas...)

	return schemas
}

// createSchemaSpecForEntrySingularModel creates a schema for entry-type plural resources.
//
// Entry-type plural resources are resources that manage multiple PAN-OS objects within
// single terraform resource, e.g. `panos_address_objects`. For such objects, we want to
// provide users with a simple way of indexing into specific objects based on their name,
// so the terraform object represents lists as sets, where key is object name, and the value
// is an terraform nested attribute describing the rest of object parameters.
func createSchemaSpecForEntryListModel(resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, spec *properties.Normalization, packageName string, structName string, manager *imports.Manager) []schemaCtx {
	var schemas []schemaCtx
	var attributes []attributeCtx

	if len(spec.Locations) > 0 {
		location := properties.NewNameVariant("location")

		attributes = append(attributes, attributeCtx{
			Package:    packageName,
			Name:       location,
			Required:   true,
			SchemaType: "SingleNestedAttribute",
		})
	}

	listNameStr := spec.TerraformProviderConfig.PluralName
	listName := properties.NewNameVariant(listNameStr)

	var listAttributeSchemaType string
	switch spec.TerraformProviderConfig.PluralType {
	case object.TerraformPluralListType:
		listAttributeSchemaType = "ListNestedAttribute"
	case object.TerraformPluralMapType:
		listAttributeSchemaType = "MapNestedAttribute"
	case object.TerraformPluralSetType:
		listAttributeSchemaType = "SetNestedAttribute"
	}

	attributes = append(attributes, attributeCtx{
		Package:     packageName,
		Name:        listName,
		Description: spec.TerraformProviderConfig.PluralDescription,
		Required:    true,
		SchemaType:  listAttributeSchemaType,
	})

	for _, elt := range spec.PanosXpath.Variables {
		if elt.Name == "name" {
			continue
		}

		param, err := spec.ParameterForPanosXpathVariable(elt)
		if err != nil {
			panic(fmt.Sprintf("couldn't find matching param for xpath variable: %s", err.Error()))
		}

		attributes = append(attributes, attributeCtx{
			Package:    packageName,
			Name:       param.Name,
			Required:   true,
			SchemaType: "StringAttribute",
		})
	}

	var isResource bool
	if schemaTyp == properties.SchemaResource {
		isResource = true
	}
	schemas = append(schemas, schemaCtx{
		Package:       packageName,
		ObjectOrModel: "Model",
		IsResource:    isResource,
		StructName:    structName,
		ReturnType:    "Schema",
		Attributes:    attributes,
	})

	structName = fmt.Sprintf("%s%s", structName, listName.CamelCase)
	normalizationAttrs, normalizationSchemas := createSchemaSpecForNormalization(resourceTyp, schemaTyp, spec, packageName, structName, manager)

	schemas = append(schemas, schemaCtx{
		Package:       packageName,
		ObjectOrModel: "Object",
		IsResource:    isResource,
		StructName:    structName,
		ReturnType:    "NestedAttributeObject",
		Attributes:    normalizationAttrs,
	})

	schemas = append(schemas, normalizationSchemas...)

	return schemas
}

// createSchemaSpecForModel generates schema spec for the top-level object based on the ResourceType.
func createSchemaSpecForModel(resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, spec *properties.Normalization, manager *imports.Manager) []schemaCtx {
	var packageName string
	switch schemaTyp {
	case properties.SchemaDataSource:
		packageName = "dsschema"
	case properties.SchemaResource:
		if spec.TerraformProviderConfig.Ephemeral {
			packageName = "ephschema"
		} else {
			packageName = "rsschema"
		}
	case properties.SchemaEphemeralResource:
		packageName = "ephschema"
	case properties.SchemaCommon, properties.SchemaProvider:
		panic("unreachable")
	}

	if spec.Spec == nil {
		return nil
	}

	names := NewNameProvider(spec, resourceTyp)

	var structName string
	switch schemaTyp {
	case properties.SchemaDataSource:
		structName = names.DataSourceStructName
	case properties.SchemaResource, properties.SchemaEphemeralResource:
		structName = names.ResourceStructName
	case properties.SchemaCommon, properties.SchemaProvider:
		panic("unreachable")
	}

	switch resourceTyp {
	case properties.ResourceEntry, properties.ResourceCustom, properties.ResourceConfig:
		return createSchemaSpecForEntrySingularModel(resourceTyp, schemaTyp, spec, packageName, structName, manager)
	case properties.ResourceEntryPlural:
		return createSchemaSpecForEntryListModel(resourceTyp, schemaTyp, spec, packageName, structName, manager)
	case properties.ResourceUuid, properties.ResourceUuidPlural:
		return createSchemaSpecForUuidModel(resourceTyp, schemaTyp, spec, packageName, structName, manager)
	default:
		panic("unreachable")
	}
}

func createSchemaSpecForNormalization(resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, spec *properties.Normalization, packageName string, structName string, manager *imports.Manager) ([]attributeCtx, []schemaCtx) {
	var schemas []schemaCtx
	var attributes []attributeCtx

	// We don't add name for resources that have plurar type set to map, as those resources
	// handle names as map keys in the top-level model.
	if spec.HasEntryName() && (resourceTyp != properties.ResourceEntryPlural || spec.TerraformProviderConfig.PluralType != object.TerraformPluralMapType) {
		name := properties.NewNameVariant("name")

		var description string
		if spec.Entry != nil && spec.Entry.Name != nil {
			description = spec.Entry.Name.Description
		}

		attributes = append(attributes, attributeCtx{
			Description: description,
			Package:     packageName,
			Name:        name,
			SchemaType:  "StringAttribute",
			Required:    true,
		})
	}

	for _, elt := range spec.Spec.SortedParams() {
		if elt.IsPrivateParameter() {
			continue
		}

		if resourceTyp == properties.ResourceEntryPlural && elt.TerraformProviderConfig != nil && elt.TerraformProviderConfig.XpathVariable != nil {
			continue
		}

		var functions []validatorFunctionCtx
		if len(elt.EnumValues) > 0 && schemaTyp == properties.SchemaResource {
			var values []string
			for _, elt := range elt.EnumValues {
				values = append(values, elt.Name)
			}

			functions = append(functions, validatorFunctionCtx{
				Type:     "Values",
				Function: "OneOf",
				Values:   values,
			})
		}

		var validators *validatorCtx
		if len(functions) > 0 {
			typ := elt.ValidatorType()
			validatorImport := fmt.Sprintf("github.com/hashicorp/terraform-plugin-framework-validators/%svalidator", typ)
			manager.AddHashicorpImport(validatorImport, "")
			validators = &validatorCtx{
				ListType:  pascalCase(typ),
				Package:   fmt.Sprintf("%svalidator", typ),
				Functions: functions,
			}
		}

		attributes = append(attributes, createSchemaAttributeForParameter(schemaTyp, manager, packageName, elt, validators))
		schemas = append(schemas, createSchemaSpecForParameter(schemaTyp, manager, structName, packageName, elt, nil)...)
	}

	validatorFns := generateValidatorFnsMapForVariants(spec.Spec.SortedOneOf())

	for _, elt := range spec.Spec.SortedOneOf() {
		if elt.IsPrivateParameter() {
			continue
		}

		if resourceTyp == properties.ResourceEntryPlural && elt.TerraformProviderConfig != nil && elt.TerraformProviderConfig.XpathVariable != nil {
			continue
		}

		var validators *validatorCtx
		if schemaTyp == properties.SchemaResource {
			validatorFn, found := validatorFns[elt.VariantGroupId]
			if found && validatorFn.Function != "Disabled" {
				typ := elt.ValidatorType()
				validatorImport := fmt.Sprintf("github.com/hashicorp/terraform-plugin-framework-validators/%svalidator", typ)
				manager.AddHashicorpImport(validatorImport, "")

				validators = &validatorCtx{
					ListType:  pascalCase(typ),
					Package:   fmt.Sprintf("%svalidator", typ),
					Functions: []validatorFunctionCtx{*validatorFn},
				}

				delete(validatorFns, elt.VariantGroupId)
			}
		}

		attributes = append(attributes, createSchemaAttributeForParameter(schemaTyp, manager, packageName, elt, validators))
		schemas = append(schemas, createSchemaSpecForParameter(schemaTyp, manager, structName, packageName, elt, validators)...)
	}

	return attributes, schemas
}

const renderSchemaTemplate = `
{{- define "renderSchemaListAttribute" }}
	"{{ .Name.Underscore }}": {{ .Package }}.{{ .SchemaType }} {
		Description: "{{ .Description }}",
		Required: {{ .Required }},
		Optional: {{ .Optional }},
		Computed: {{ .Computed }},
		Sensitive: {{ .Sensitive }},
		ElementType: {{ .ElementType }},
  {{- with .Validators }}
    {{ $package := .Package }}
		Validators: []validator.{{ .ListType }}{
    {{- range .Functions }}
      {{- if eq .Type "Expressions" }}
			{{ $package }}.{{ .Function }}(path.Expressions{
        {{- range .Expressions }}
				{{ . }},
        {{- end }}
			}...),

      {{- else if eq .Type "Values" }}
			{{ $package }}.{{ .Function }}([]string{
          {{- range .Values }}
				{{ . }},
          {{- end }}
			}...),
      {{- end }}
    {{- end }}
		},
  {{- end }}
	},
{{- end }}

{{- define "renderSchemaMapAttribute" }}
	"{{ .Name.Underscore }}": {{ .Package }}.{{ .SchemaType }} {
		Description: "{{ .Description }}",
		Required: {{ .Required }},
		Optional: {{ .Optional }},
		Computed: {{ .Computed }},
		Sensitive: {{ .Sensitive }},
		ElementType: {{ .ElementType }},
	},
{{- end }}

{{- define "renderSchemaListNestedAttribute" }}
  {{- with .Attribute }}
	"{{ .Name.Underscore }}": {{ .Package }}.{{ .SchemaType }} {
		Description: "{{ .Description }}",
		Required: {{ .Required }},
		Optional: {{ .Optional }},
		Computed: {{ .Computed }},
		Sensitive: {{ .Sensitive }},
		NestedObject: {{ $.StructName }}{{ .Name.CamelCase }}Schema(),
	},
  {{- end }}
{{- end }}

{{- define "renderSchemaMapNestedAttribute" }}
  {{- template "renderSchemaListNestedAttribute" . }}
{{- end }}


{{- define "renderSchemaSingleNestedAttribute" }}
  {{- with .Attribute }}
	"{{ .Name.Underscore }}": {{ $.StructName }}{{ .Name.CamelCase }}Schema(),
  {{- end }}
{{- end }}

{{- define "renderSchemaExternalAttribute" }}
  {{- with .Attribute }}
	"{{ .Name.Underscore }}": {{ .ExternalType }}Schema(),
  {{- end }}
{{- end }}

{{- define "renderSchemaSimpleAttribute" }}
	"{{ .Name.Underscore }}": {{ .Package }}.{{ .SchemaType }} {
		Description: "{{ .Description }}",
		Computed: {{ .Computed }},
		Required: {{ .Required }},
		Optional: {{ .Optional }},
		Sensitive: {{ .Sensitive }},
  {{- if .Default }}
		Default: {{ .Default.Type }}({{ .Default.Value }}),
  {{- end }}
  {{- if .PlanModifiers }}
		PlanModifiers: []{{ .PlanModifiers.SchemaType }}{
    {{- range .PlanModifiers.Modifiers }}
			{{ . }},
    {{- end }}
		},
  {{- end }}

  {{- with .Validators }}
    {{ $package := .Package }}
		Validators: []validator.{{ .ListType }}{
    {{- range .Functions }}
      {{- if eq .Type "Expressions" }}
			{{ $package }}.{{ .Function }}(path.Expressions{
        {{- range .Expressions }}
				{{ . }},
        {{- end }}
			}...),
      {{- else if eq .Type "Values" }}
			{{ $package }}.{{ .Function }}([]string{
        {{- range .Values }}
				"{{ . }}",
        {{- end }}
			}...),
      {{- end }}
    {{- end }}
		},
  {{- end }}
	},
{{- end }}

{{- define "renderSchemaAttribute" }}
  {{- with .Attribute }}
    {{ if or (eq .SchemaType "ListAttribute") (eq .SchemaType "SetAttribute") }}
      {{- template "renderSchemaListAttribute" . }}
    {{- else if eq .SchemaType "MapAttribute" }}
      {{- template "renderSchemaMapAttribute" . }}
    {{- else if eq .SchemaType "ListNestedAttribute" }}
      {{- template "renderSchemaListNestedAttribute" Map "StructName" $.StructName "Attribute" . }}
    {{ else if eq .SchemaType "MapNestedAttribute" }}
      {{- template "renderSchemaMapNestedAttribute" Map "StructName" $.StructName "Attribute" . }}
    {{- else if eq .SchemaType "SingleNestedAttribute" }}
      {{- template "renderSchemaSingleNestedAttribute" Map "StructName" $.StructName "Attribute" . }}
    {{- else if eq .SchemaType "ExternalAttribute" }}
      {{- template "renderSchemaExternalAttribute" Map "Attribute" . }}
    {{- else }}
      {{- template "renderSchemaSimpleAttribute" . }}
    {{- end }}
  {{- end }}
{{- end }}

{{- range .Schemas }}
{{ $schema := . }}

func {{ .StructName }}Schema() {{ .Package }}.{{ .ReturnType }} {
	return {{ .Package }}.{{ .ReturnType }}{
{{- if not (or (eq .ReturnType "Schema") (eq .ReturnType "NestedAttributeObject")) }}
		Description: "{{ .Description }}",
		Required: {{ .Required }},
		Computed: {{ .Computed }},
		Optional: {{ .Optional }},
		Sensitive: {{ .Sensitive }},
{{- end }}
  {{- with .Validators }}
    {{ $package := .Package }}
		Validators: []validator.{{ .ListType }}{
    {{- range .Functions }}
			{{ $package }}.{{ .Function }}(path.Expressions{
      {{- range .Expressions }}
				{{ . }},
      {{- end }}
			}...),
    {{- end }}
		},
  {{- end }}
		Attributes: map[string]{{ .Package }}.Attribute{
  {{- range .Attributes -}}
	{{- template "renderSchemaAttribute" Map "StructName" $schema.StructName "Attribute" . }}
  {{- end }}
		},
	}
}

func (o *{{ .StructName }}{{ .ObjectOrModel }}) getTypeFor(name string) attr.Type {
	schema := {{ .StructName }}Schema()
	if attr, ok := schema.Attributes[name]; !ok {
		panic(fmt.Sprintf("could not resolve schema for attribute %s", name))
	} else {
		switch attr := attr.(type) {
		case {{ .Package }}.ListNestedAttribute:
			return attr.NestedObject.Type()
		case {{ .Package }}.MapNestedAttribute:
			return attr.NestedObject.Type()
		default:
			return attr.GetType()
		}
	}

	panic("unreachable")
}

{{- end }}
`

func RenderResourceSchema(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization, manager *imports.Manager) (string, error) {
	type context struct {
		Schemas []schemaCtx
	}

	data := context{
		Schemas: createSchemaSpecForModel(resourceTyp, properties.SchemaResource, spec, manager),
	}

	return processTemplate(renderSchemaTemplate, "render-resource-schema", data, commonFuncMap)
}

func RenderDataSourceSchema(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization, manager *imports.Manager) (string, error) {
	type context struct {
		Schemas []schemaCtx
	}

	data := context{
		Schemas: createSchemaSpecForModel(resourceTyp, properties.SchemaDataSource, spec, manager),
	}

	return processTemplate(renderSchemaTemplate, "render-resource-schema", data, commonFuncMap)
}

const importLocationAssignmentTmpl = `
{
var terraformLocation {{ .TerraformStructName }}
resp.Diagnostics.Append({{ $.Source }}.As(ctx, &terraformLocation, basetypes.ObjectAsOptions{})...)
if resp.Diagnostics.HasError() {
	return
}
{{- range .Specs }}
{{ $type := . }}
{{ $locationName := .Name }}
if location.{{ .Name.CamelCase }} != nil {
  {{- range .Locations }}
	{
	var terraformInnerLocation {{ .TerraformStructName }}
	resp.Diagnostics.Append(terraformLocation.{{ $locationName.CamelCase }}.As(ctx, &terraformInnerLocation, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}
    {{- $pangoStruct := GetPangoStructForLocation $.Variants $type.Name .Name }}
	{{ $.Dest }} = {{ $.PackageName }}.New{{ $pangoStruct }}({{ $.PackageName }}.{{ $pangoStruct }}Spec{
    {{- range .Fields }}
		{{ . }}: terraformInnerLocation.{{ . }}.ValueString(),
    {{- end }}
	})
	}
  {{- end }}
}
}
{{- end }}
`

func RenderImportLocationAssignment(names *NameProvider, spec *properties.Normalization, source string, dest string) (string, error) {
	if len(spec.Imports) == 0 {
		return "", nil
	}

	type importVariantSpec struct {
		PangoStructNames *map[string]string
	}

	type importLocationSpec struct {
		TerraformStructName string
		Name                *properties.NameVariant
		Fields              []string
	}

	type importSpec struct {
		TerraformStructName string
		Name                *properties.NameVariant
		Locations           []importLocationSpec
	}

	var importSpecs []importSpec
	variantsByName := make(map[string]importVariantSpec)
	for _, elt := range spec.Imports {
		existing, found := variantsByName[elt.Type.CamelCase]
		if !found {
			pangoStructNames := make(map[string]string)
			existing = importVariantSpec{
				PangoStructNames: &pangoStructNames,
			}
		}

		var locations []importLocationSpec
		for _, loc := range elt.Locations {
			if !loc.Required {
				continue
			}

			var fields []string
			for _, elt := range loc.XpathVariables {
				fields = append(fields, elt.Name.CamelCase)
			}

			tfStructName := fmt.Sprintf("%s%sLocation", names.StructName, elt.Type.CamelCase)
			pangoStructName := fmt.Sprintf("%s%s%sImportLocation", elt.Variant.CamelCase, elt.Type.CamelCase, loc.Name.CamelCase)
			(*existing.PangoStructNames)[loc.Name.CamelCase] = pangoStructName
			locations = append(locations, importLocationSpec{
				TerraformStructName: tfStructName,
				Name:                loc.Name,
				Fields:              fields,
			})
		}
		variantsByName[elt.Type.CamelCase] = existing

		importSpecs = append(importSpecs, importSpec{
			Name:      elt.Type,
			Locations: locations,
		})
	}

	type context struct {
		TerraformStructName string
		PackageName         string
		Source              string
		Dest                string
		Variants            map[string]importVariantSpec
		Specs               []importSpec
	}

	data := context{
		TerraformStructName: fmt.Sprintf("%sLocation", names.StructName),
		PackageName:         names.PackageName,
		Source:              source,
		Dest:                dest,
		Variants:            variantsByName,
		Specs:               importSpecs,
	}

	funcMap := template.FuncMap{
		"GetPangoStructForLocation": func(variants map[string]importVariantSpec, typ *properties.NameVariant, location *properties.NameVariant) (string, error) {
			variantSpec, found := variants[typ.CamelCase]
			if !found {
				return "", fmt.Errorf("failed to find variant for type '%s'", typ.CamelCase)
			}

			structName, found := (*variantSpec.PangoStructNames)[location.CamelCase]
			if !found {
				return "", fmt.Errorf("failed to find variant for type '%s', location '%s'", typ.CamelCase, location.CamelCase)
			}

			return structName, nil
		},
	}

	return processTemplate(importLocationAssignmentTmpl, "render-locations-pango-to-state", data, funcMap)
}

type locationFieldCtx struct {
	PangoName     string
	TerraformName string
	Type          string
}

type locationCtx struct {
	Name                string
	PangoStructName     string
	TerraformStructName string
	SdkStructName       string
	Fields              []locationFieldCtx
}

func renderLocationsGetContext(names *NameProvider, spec *properties.Normalization) []locationCtx {
	var locations []locationCtx

	for _, location := range spec.OrderedLocations() {
		var fields []locationFieldCtx
		for _, variable := range location.OrderedVars() {
			name := variable.Name.CamelCase
			if variable.Name.CamelCase == location.Name.CamelCase {
				name = "Name"
			}

			fields = append(fields, locationFieldCtx{
				PangoName:     variable.Name.CamelCase,
				TerraformName: name,
				Type:          "String",
			})
		}
		locations = append(locations, locationCtx{
			Name:                location.Name.CamelCase,
			PangoStructName:     fmt.Sprintf("%s.%sLocation", names.PackageName, location.Name.CamelCase),
			TerraformStructName: fmt.Sprintf("%s%sLocation", names.StructName, location.Name.CamelCase),
			SdkStructName:       fmt.Sprintf("%s.%sLocation", names.PackageName, location.Name.CamelCase),
			Fields:              fields,
		})
	}

	return locations
}

const locationsPangoToState = `
{{- range .Locations }}
if {{ $.Source }}.{{ .Name }} != nil {
	{{ $.Dest }}.{{ .Name }} = &{{ .TerraformStructName }}{
    {{ $locationName := .Name }}
  {{- range .Fields }}
		{{ .TerraformName }}: types.{{ .Type }}Value({{ $.Source }}.{{ $locationName }}.{{ .PangoName }}),
  {{- end }}
	}
}
{{- end }}
`

func RenderLocationsPangoToState(names *NameProvider, spec *properties.Normalization, source string, dest string) (string, error) {
	type context struct {
		Source    string
		Dest      string
		Locations []locationCtx
	}
	data := context{Source: source, Dest: dest, Locations: renderLocationsGetContext(names, spec)}
	return processTemplate(locationsPangoToState, "render-locations-pango-to-state", data, commonFuncMap)
}

const locationsStateToPango = `
{
var terraformLocation {{ .TerraformStructName }}
resp.Diagnostics.Append({{ $.Source }}.As(ctx, &terraformLocation, basetypes.ObjectAsOptions{})...)
if resp.Diagnostics.HasError() {
	return
}

{{- range .Locations }}
{{ $locationType := .Name }}
if !terraformLocation.{{ $locationType }}.IsNull() {
	{{ $.Dest }}.{{ $locationType }} = &{{ .PangoStructName }}{}
	var innerLocation {{ .TerraformStructName }}
	resp.Diagnostics.Append(terraformLocation.{{ .Name }}.As(ctx, &innerLocation, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}
  {{- range .Fields }}
	{{ $.Dest }}.{{ $locationType }}.{{ .PangoName }} = innerLocation.{{ .TerraformName }}.ValueString()
  {{- end }}
}
{{- end }}
}
`

func RenderLocationsStateToPango(names *NameProvider, spec *properties.Normalization, source string, dest string) (string, error) {
	type context struct {
		TerraformStructName string
		Source              string
		Dest                string
		Locations           []locationCtx
	}
	data := context{
		TerraformStructName: fmt.Sprintf("%sLocation", names.StructName),
		Locations:           renderLocationsGetContext(names, spec),
		Source:              source,
		Dest:                dest,
	}
	return processTemplate(locationsStateToPango, "render-locations-state-to-pango", data, commonFuncMap)
}

func RendeCreateUpdateMovementRequired(state string, entries string) (string, error) {
	type context struct {
		State   string
		Entries string
	}
	data := context{State: state, Entries: entries}
	return processTemplate(resourceCreateUpdateMovementRequiredTmpl, "render-create-update-movement-required", data, nil)
}

const dataSourceStructs = `
{{- range .Structs }}
type {{ .StructName }}{{ .ModelOrObject }} struct {
  {{- range .Fields }}
    {{- if .Private }}
	{{ .Name.LowerCamelCase }} {{ .Type }} {{ range .Tags }}{{ . }}{{ end }}
    {{- else }}
	{{ .Name.CamelCase }} {{ .Type }} {{ range .Tags }}{{ . }} {{ end }}
    {{- end }}
  {{- end }}
}
{{- end }}
`

type datasourceStructFieldSpec struct {
	Name          *properties.NameVariant
	Private       bool
	TerraformType string
	Type          string
	Tags          []string
}

type datasourceStructSpec struct {
	StructName          string
	AncestorName        string
	TerraformPluralType object.TerraformPluralType
	HasEntryName        bool
	ModelOrObject       string
	Fields              []datasourceStructFieldSpec
}

func terraformTypeForProperty(structPrefix string, prop *properties.SpecParam, hackStructsAsTypeObjects bool) string {
	if prop.Type == "" {
		if hackStructsAsTypeObjects {
			return "types.Object"
		} else {
			return fmt.Sprintf("*%s%sObject", structPrefix, prop.TerraformNameVariant().CamelCase)
		}
	}

	switch prop.ComplexType() {
	case "string-as-member":
		return "types.String"
	}

	if prop.FinalType() == "list" && prop.Items.Type == "entry" {
		return "types.List"
	}

	if prop.FinalType() == "set" && prop.Items.Type == "entry" {
		return "types.Set"
	}

	if prop.FinalType() == "list" {
		return "types.List"
	}

	if prop.FinalType() == "set" {
		return "types.Set"
	}

	return fmt.Sprintf("types.%s", pascalCase(prop.Type))
}

func structFieldSpec(param *properties.SpecParam, structPrefix string, hackStructsAsTypeObjects bool) datasourceStructFieldSpec {
	tfTag := fmt.Sprintf("`tfsdk:\"%s\"`", param.TerraformNameVariant().Underscore)

	return datasourceStructFieldSpec{
		Name:          param.TerraformNameVariant(),
		TerraformType: terraformTypeForProperty(structPrefix, param, false),
		Type:          terraformTypeForProperty(structPrefix, param, hackStructsAsTypeObjects),
		Tags:          []string{tfTag},
	}
}

func dataSourceStructContextForParam(structPrefix string, param *properties.SpecParam, hackStructsAsTypeObjects bool) []datasourceStructSpec {
	var structs []datasourceStructSpec

	structName := fmt.Sprintf("%s%s", structPrefix, param.TerraformNameVariant().CamelCase)

	var fields []datasourceStructFieldSpec

	if param.HasEntryName() {
		fields = append(fields, datasourceStructFieldSpec{
			Name: properties.NewNameVariant("name"),
			Type: "types.String",
			Tags: []string{"`tfsdk:\"name\"`"},
		})
	}

	if param.Spec != nil {
		for _, elt := range param.Spec.SortedParams() {
			if elt.IsPrivateParameter() {
				continue
			}
			fields = append(fields, structFieldSpec(elt, structName, hackStructsAsTypeObjects))
		}

		for _, elt := range param.Spec.SortedOneOf() {
			if elt.IsPrivateParameter() {
				continue
			}
			fields = append(fields, structFieldSpec(elt, structName, hackStructsAsTypeObjects))
		}
	}

	structs = append(structs, datasourceStructSpec{
		AncestorName:  param.TerraformNameVariant().Original,
		HasEntryName:  param.HasEntryName(),
		StructName:    structName,
		ModelOrObject: "Object",
		Fields:        fields,
	})

	if param.Spec == nil {
		return structs
	}

	for _, elt := range param.Spec.SortedParams() {
		if elt.IsPrivateParameter() {
			continue
		}
		if elt.Type == "" || (elt.Type == "list" && elt.Items.Type == "entry") {
			structs = append(structs, dataSourceStructContextForParam(structName, elt, hackStructsAsTypeObjects)...)
		}
	}

	for _, elt := range param.Spec.SortedOneOf() {
		if elt.IsPrivateParameter() {
			continue
		}

		if elt.Type == "" || (elt.Type == "list" && elt.Items.Type == "entry") {
			structs = append(structs, dataSourceStructContextForParam(structName, elt, hackStructsAsTypeObjects)...)
		}
	}

	return structs
}

func createStructSpecForUuidModel(resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, spec *properties.Normalization, names *NameProvider, hackStructsAsTypeObjects bool) []datasourceStructSpec {
	var structs []datasourceStructSpec

	var fields []datasourceStructFieldSpec

	if len(spec.Locations) > 0 {
		fields = append(fields, datasourceStructFieldSpec{
			Name:          properties.NewNameVariant("location"),
			TerraformType: fmt.Sprintf("%sLocation", names.StructName),
			Type:          "types.Object",
			Tags:          []string{"`tfsdk:\"location\"`"},
		})
	}

	if resourceTyp == properties.ResourceUuidPlural {

		position := properties.NewNameVariant("position")

		fields = append(fields, datasourceStructFieldSpec{
			Name:          position,
			TerraformType: "TerraformPositionObject",
			Type:          "types.Object",
			Tags:          []string{"`tfsdk:\"position\"`"},
		})
	}

	var structName string
	switch schemaTyp {
	case properties.SchemaResource, properties.SchemaEphemeralResource:
		structName = names.ResourceStructName
	case properties.SchemaDataSource:
		structName = names.DataSourceStructName
	case properties.SchemaCommon, properties.SchemaProvider:
		panic("unreachable")
	}

	listNameStr := spec.TerraformProviderConfig.PluralName
	listName := properties.NewNameVariant(listNameStr)

	tag := fmt.Sprintf("`tfsdk:\"%s\"`", listName.Underscore)
	fields = append(fields, datasourceStructFieldSpec{
		Name: listName,
		Type: "types.List",
		Tags: []string{tag},
	})

	structs = append(structs, datasourceStructSpec{
		StructName:    structName,
		ModelOrObject: "Model",
		Fields:        fields,
	})

	structName = fmt.Sprintf("%s%s", structName, listName.CamelCase)
	fields, normalizationStructs := createStructSpecForNormalization(resourceTyp, structName, spec, hackStructsAsTypeObjects)

	structs = append(structs, datasourceStructSpec{
		AncestorName:  listName.Original,
		HasEntryName:  true,
		StructName:    structName,
		ModelOrObject: "Object",
		Fields:        fields,
	})

	structs = append(structs, normalizationStructs...)

	return structs
}

func createStructSpecForEntryListModel(resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, spec *properties.Normalization, names *NameProvider, hackStructsAsTypeObjects bool) []datasourceStructSpec {
	var structs []datasourceStructSpec

	var fields []datasourceStructFieldSpec
	if len(spec.Locations) > 0 {
		fields = append(fields, datasourceStructFieldSpec{
			Name:          properties.NewNameVariant("location"),
			TerraformType: fmt.Sprintf("%sLocation", names.StructName),
			Type:          "types.Object",
			Tags:          []string{"`tfsdk:\"location\"`"},
		})
	}

	var structName string
	switch schemaTyp {
	case properties.SchemaResource, properties.SchemaEphemeralResource:
		structName = names.ResourceStructName
	case properties.SchemaDataSource:
		structName = names.DataSourceStructName
	case properties.SchemaCommon, properties.SchemaProvider:
		panic("unreachable")
	}

	for _, elt := range spec.PanosXpath.Variables {
		if elt.Name == "name" {
			continue
		}

		param, err := spec.ParameterForPanosXpathVariable(elt)
		if err != nil {
			panic(fmt.Sprintf("couldn't find matching param for xpath variable: %s", err.Error()))
		}

		xmlTags := []string{fmt.Sprintf("`tfsdk:\"%s\"`", param.Name.Underscore)}
		fields = append(fields, datasourceStructFieldSpec{
			Name: param.Name,
			Type: "types.String",
			Tags: xmlTags,
		})
	}

	listNameStr := spec.TerraformProviderConfig.PluralName
	listName := properties.NewNameVariant(listNameStr)

	var listEltType string
	switch spec.TerraformProviderConfig.PluralType {
	case object.TerraformPluralMapType:
		listEltType = "types.Map"
	case object.TerraformPluralListType:
		listEltType = "types.List"
	case object.TerraformPluralSetType:
		listEltType = "types.Set"
	}

	tag := fmt.Sprintf("`tfsdk:\"%s\"`", listName.Underscore)
	fields = append(fields, datasourceStructFieldSpec{
		Name: listName,
		Type: listEltType,
		Tags: []string{tag},
	})

	structs = append(structs, datasourceStructSpec{
		StructName:    structName,
		ModelOrObject: "Model",
		Fields:        fields,
	})

	structName = fmt.Sprintf("%s%s", structName, listName.CamelCase)
	fields, normalizationStructs := createStructSpecForNormalization(resourceTyp, structName, spec, hackStructsAsTypeObjects)

	structs = append(structs, datasourceStructSpec{
		AncestorName:        listName.Original,
		TerraformPluralType: spec.TerraformProviderConfig.PluralType,
		HasEntryName:        true,
		StructName:          structName,
		ModelOrObject:       "Object",
		Fields:              fields,
	})

	structs = append(structs, normalizationStructs...)

	return structs
}

func createStructSpecForEntryModel(resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, spec *properties.Normalization, names *NameProvider, hackStructAsTypeObjects bool) []datasourceStructSpec {
	var structs []datasourceStructSpec

	var fields []datasourceStructFieldSpec

	if len(spec.Locations) > 0 {
		fields = append(fields, datasourceStructFieldSpec{
			Name:          properties.NewNameVariant("location"),
			TerraformType: fmt.Sprintf("%sLocation", names.StructName),
			Type:          "types.Object",
			Tags:          []string{"`tfsdk:\"location\"`"},
		})
	}

	var structName string
	switch schemaTyp {
	case properties.SchemaDataSource:
		structName = names.DataSourceStructName
	case properties.SchemaResource, properties.SchemaEphemeralResource:
		structName = names.ResourceStructName
	case properties.SchemaCommon, properties.SchemaProvider:
		panic("unreachable")
	}

	normalizationFields, normalizationStructs := createStructSpecForNormalization(resourceTyp, structName, spec, hackStructAsTypeObjects)
	fields = append(fields, normalizationFields...)

	structs = append(structs, datasourceStructSpec{
		StructName:    structName,
		ModelOrObject: "Model",
		Fields:        fields,
	})

	structs = append(structs, normalizationStructs...)

	return structs
}

func createStructSpecForModel(resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, spec *properties.Normalization, names *NameProvider, hackStructsAsTypeObjects bool) []datasourceStructSpec {
	if spec.Spec == nil {
		return nil
	}

	switch resourceTyp {
	case properties.ResourceEntry, properties.ResourceCustom, properties.ResourceConfig:
		return createStructSpecForEntryModel(resourceTyp, schemaTyp, spec, names, hackStructsAsTypeObjects)
	case properties.ResourceEntryPlural:
		return createStructSpecForEntryListModel(resourceTyp, schemaTyp, spec, names, hackStructsAsTypeObjects)
	case properties.ResourceUuid, properties.ResourceUuidPlural:
		return createStructSpecForUuidModel(resourceTyp, schemaTyp, spec, names, hackStructsAsTypeObjects)
	default:
		panic("unreachable")
	}
}

func createStructSpecForNormalization(resourceTyp properties.ResourceType, structName string, spec *properties.Normalization, hackStructAsTypeObjects bool) ([]datasourceStructFieldSpec, []datasourceStructSpec) {
	var fields []datasourceStructFieldSpec
	var structs []datasourceStructSpec

	// We don't add name field for entry-style list resources, as they
	// represent lists as maps with name being a key.
	if spec.HasEntryName() {
		var private bool
		typ := "types.String"
		tag := "`tfsdk:\"name\"`"

		if resourceTyp == properties.ResourceEntryPlural && spec.TerraformProviderConfig.PluralType == object.TerraformPluralMapType {
			private = true
			typ = "string"
			tag = "`tfsdk:\"-\"`"
		}

		fields = append(fields, datasourceStructFieldSpec{
			Name:    properties.NewNameVariant("name"),
			Private: private,
			Type:    typ,
			Tags:    []string{tag},
		})
	}

	for _, elt := range spec.Spec.SortedParams() {
		if elt.IsPrivateParameter() {
			continue
		}

		if resourceTyp == properties.ResourceEntryPlural && elt.TerraformProviderConfig != nil && elt.TerraformProviderConfig.XpathVariable != nil {
			continue
		}

		fields = append(fields, structFieldSpec(elt, structName, hackStructAsTypeObjects))
		if elt.Type == "" || (elt.Type == "list" && elt.Items.Type == "entry") {
			structs = append(structs, dataSourceStructContextForParam(structName, elt, hackStructAsTypeObjects)...)
		}
	}

	for _, elt := range spec.Spec.SortedOneOf() {
		if elt.IsPrivateParameter() {
			continue
		}

		if resourceTyp == properties.ResourceEntryPlural && elt.TerraformProviderConfig != nil && elt.TerraformProviderConfig.XpathVariable != nil {
			continue
		}

		fields = append(fields, structFieldSpec(elt, structName, hackStructAsTypeObjects))
		if elt.Type == "" || (elt.Type == "list" && elt.Items.Type == "entry") {
			structs = append(structs, dataSourceStructContextForParam(structName, elt, hackStructAsTypeObjects)...)
		}
	}

	return fields, structs
}

func RenderResourceStructs(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization) (string, error) {
	type context struct {
		Structs []datasourceStructSpec
	}

	data := context{
		Structs: createStructSpecForModel(resourceTyp, properties.SchemaResource, spec, names, false),
	}

	return processTemplate(dataSourceStructs, "render-structs", data, commonFuncMap)
}

func RenderDataSourceStructs(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization) (string, error) {
	type context struct {
		Structs []datasourceStructSpec
	}

	data := context{
		Structs: createStructSpecForModel(resourceTyp, properties.SchemaDataSource, spec, names, false),
	}

	return processTemplate(dataSourceStructs, "render-structs", data, commonFuncMap)
}

const attributeTypesTmpl = `
{{- range .Structs }}
func (o *{{ .StructName }}{{ .ModelOrObject }}) AttributeTypes() map[string]attr.Type {
  {{- range .Fields }}
    {{- if .Private }}{{ continue }}{{- end }}
    {{ if (eq .Type "types.Object") }}
	var {{ .Name.LowerCamelCase }}Obj {{ .TerraformType }}
    {{- end }}
  {{- end }}
	return map[string]attr.Type{
  {{- range .Fields }}
    {{- if .Private }}{{ continue }}{{- end }}
    {{- if eq .Type "types.Object" }}
	"{{ .Name.Underscore }}": {{ .Type }}Type{
		AttrTypes: {{ .Name.LowerCamelCase }}Obj.AttributeTypes(),
	},
    {{- else if or (eq .Type "types.List") (eq .Type "types.Set") (eq .Type "types.Map") }}
		"{{ .Name.Underscore }}": {{ .Type }}Type{},
    {{- else }}
		"{{ .Name.Underscore }}": {{ .Type }}Type,
    {{- end }}
  {{- end }}
	}
}

func (o {{ .StructName }}{{ .ModelOrObject }}) AncestorName() string {
	return "{{ .AncestorName }}"
}

func (o {{ .StructName }}{{ .ModelOrObject }}) EntryName() *string {
    {{- if and .HasEntryName (eq .TerraformPluralType "map") }}
	return &o.name
    {{- else if .HasEntryName }}
	return o.Name.ValueStringPointer()
    {{- else }}
	return nil
    {{- end }}
}
{{- end }}
`

func RenderModelAttributeTypesFunction(resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, names *NameProvider, spec *properties.Normalization) (string, error) {
	if resourceTyp == properties.ResourceCustom {
		return "", nil
	}

	type context struct {
		Structs []datasourceStructSpec
	}

	data := context{
		Structs: createStructSpecForModel(resourceTyp, schemaTyp, spec, names, true),
	}

	return processTemplate(attributeTypesTmpl, "attribute-types", data, nil)
}

const locationAttributeTypesTmpl = `
{{- range .Specs }}
func (o *{{ .StructName }}) AttributeTypes() map[string]attr.Type{
  {{- range .Fields }}
    {{- if eq .Type "types.Object" }}
	var {{ .Name.LowerCamelCase }}Obj {{ .TerraformType }}
    {{- end }}
  {{- end }}
	return map[string]attr.Type{
  {{- range .Fields }}
    {{- if eq .Type "types.Object" }}
		"{{ .Name.Underscore }}": {{ .Type }}Type{
			AttrTypes: {{ .Name.LowerCamelCase }}Obj.AttributeTypes(),
		},
    {{- else }}
		"{{ .Name.Underscore }}": {{ .Type }}Type,
    {{- end }}
  {{- end }}
	}
}
{{- end }}
`

func RenderLocationAttributeTypes(names *NameProvider, spec *properties.Normalization) (string, error) {
	type context struct {
		Specs []locationStructCtx
	}

	locations := getLocationStructsContext(names, spec)

	data := context{
		Specs: locations,
	}
	return processTemplate(locationAttributeTypesTmpl, "render-location-structs", data, commonFuncMap)
}

func getCustomTemplateForFunction(spec *properties.Normalization, function string) (string, error) {
	if resource, found := customResourceFuncsMap[spec.TerraformProviderConfig.Suffix]; !found {
		return "", fmt.Errorf("cannot find a list of custom functions for %s", spec.TerraformProviderConfig.Suffix)
	} else {
		if template, found := resource[function]; !found {
			return "", fmt.Errorf("No template for function '%s'", function)
		} else {
			return template, nil
		}
	}
}

func ResourceCreateFunction(resourceTyp properties.ResourceType, names *NameProvider, serviceName string, paramSpec *properties.Normalization, terraformProvider *properties.TerraformProviderFile, resourceSDKName string) (string, error) {
	funcMap := template.FuncMap{
		"ConfigToEntry": ConfigEntry,
		"RenderEncryptedValuesInitialization": func() (string, error) {
			return RenderEncryptedValuesInitialization(properties.SchemaResource, paramSpec, "create")
		},
		"RenderEncryptedValuesFinalizer": func() (string, error) {
			return RenderEncryptedValuesFinalizer(properties.SchemaResource, paramSpec)
		},
		"RenderImportLocationAssignment": func(source string, dest string) (string, error) {
			return RenderImportLocationAssignment(names, paramSpec, source, dest)
		},
		"RenderCreateUpdateMovementRequired": func(state string, entries string) (string, error) {
			return RendeCreateUpdateMovementRequired(state, entries)
		},
		"RenderLocationsStateToPango": func(source string, dest string) (string, error) {
			return RenderLocationsStateToPango(names, paramSpec, source, dest)
		},
		"ResourceParamToSchema": func(paramName string, paramParameters properties.SpecParam) (string, error) {
			return ParamToSchemaResource(paramName, paramParameters, terraformProvider)
		},
	}

	if strings.Contains(serviceName, "group") && serviceName != "Device group" {
		serviceName = "group"
	}

	var tmpl string
	var listAttribute string
	var exhaustive bool
	switch resourceTyp {
	case properties.ResourceEntry, properties.ResourceConfig:
		exhaustive = true
		tmpl = resourceCreateFunction
	case properties.ResourceEntryPlural:
		exhaustive = false
		tmpl = resourceCreateEntryListFunction
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
	case properties.ResourceUuid:
		exhaustive = true
		tmpl = resourceCreateManyFunction
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
	case properties.ResourceUuidPlural:
		exhaustive = false
		tmpl = resourceCreateManyFunction
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
	case properties.ResourceCustom:
		var err error
		tmpl, err = getCustomTemplateForFunction(paramSpec, "Create")
		if err != nil {
			return "", err
		}
	}

	listAttributeVariant := properties.NewNameVariant(listAttribute)

	data := map[string]interface{}{
		"PluralType":            paramSpec.TerraformProviderConfig.PluralType,
		"HasEncryptedResources": paramSpec.HasEncryptedResources(),
		"HasImports":            len(paramSpec.Imports) > 0,
		"Exhaustive":            exhaustive,
		"ListAttribute":         listAttributeVariant,
		"EntryOrConfig":         paramSpec.EntryOrConfig(),
		"HasEntryName":          paramSpec.HasEntryName(),
		"structName":            names.ResourceStructName,
		"serviceName":           naming.CamelCase("", serviceName, "", false),
		"paramSpec":             paramSpec.Spec,
		"resourceSDKName":       resourceSDKName,
		"locations":             paramSpec.OrderedLocations(),
	}

	return processTemplate(tmpl, "resource-create-function", data, funcMap)
}

func DataSourceReadFunction(resourceTyp properties.ResourceType, names *NameProvider, serviceName string, paramSpec *properties.Normalization, resourceSDKName string) (string, error) {
	if strings.Contains(serviceName, "group") {
		serviceName = "group"
	}

	var tmpl string
	var listAttribute string
	var exhaustive bool
	switch resourceTyp {
	case properties.ResourceEntry, properties.ResourceConfig:
		tmpl = resourceReadFunction
	case properties.ResourceEntryPlural:
		tmpl = resourceReadEntryListFunction
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
	case properties.ResourceUuid:
		tmpl = resourceReadManyFunction
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
		exhaustive = true
	case properties.ResourceUuidPlural:
		tmpl = resourceReadManyFunction
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
	case properties.ResourceCustom:
		var err error
		tmpl, err = getCustomTemplateForFunction(paramSpec, "DataSourceRead")
		if err != nil {
			return "", err
		}
	}

	listAttributeVariant := properties.NewNameVariant(listAttribute)

	data := map[string]interface{}{
		"PluralType":                       paramSpec.TerraformProviderConfig.PluralType,
		"ResourceXpathVariablesWithChecks": paramSpec.ResourceXpathVariablesWithChecks(false),
		"ResourceOrDS":                     "DataSource",
		"HasEncryptedResources":            paramSpec.HasEncryptedResources(),
		"ListAttribute":                    listAttributeVariant,
		"Exhaustive":                       exhaustive,
		"EntryOrConfig":                    paramSpec.EntryOrConfig(),
		"HasEntryName":                     paramSpec.HasEntryName(),
		"structName":                       names.StructName,
		"resourceStructName":               names.ResourceStructName,
		"dataSourceStructName":             names.DataSourceStructName,
		"serviceName":                      naming.CamelCase("", serviceName, "", false),
		"resourceSDKName":                  resourceSDKName,
		"locations":                        paramSpec.OrderedLocations(),
	}

	funcMap := template.FuncMap{
		"RenderEncryptedValuesInitialization": func() (string, error) {
			return RenderEncryptedValuesInitialization(properties.SchemaDataSource, paramSpec, "read")
		},
		"RenderEncryptedValuesFinalizer": func() (string, error) {
			return RenderEncryptedValuesFinalizer(properties.SchemaDataSource, paramSpec)
		},
		"AttributesFromXpathComponents": func(target string) (string, error) { return paramSpec.AttributesFromXpathComponents(target) },
		"RenderLocationsPangoToState": func(source string, dest string) (string, error) {
			return RenderLocationsPangoToState(names, paramSpec, source, dest)
		},
		"RenderLocationsStateToPango": func(source string, dest string) (string, error) {
			return RenderLocationsStateToPango(names, paramSpec, source, dest)
		},
	}

	return processTemplate(tmpl, "datasource-read-function", data, funcMap)
}

func ResourceReadFunction(resourceTyp properties.ResourceType, names *NameProvider, serviceName string, paramSpec *properties.Normalization, resourceSDKName string) (string, error) {
	if strings.Contains(serviceName, "group") {
		serviceName = "group"
	}

	var tmpl string
	var listAttribute string
	var exhaustive bool
	switch resourceTyp {
	case properties.ResourceEntry, properties.ResourceConfig:
		tmpl = resourceReadFunction
	case properties.ResourceEntryPlural:
		tmpl = resourceReadEntryListFunction
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
	case properties.ResourceUuid:
		tmpl = resourceReadManyFunction
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
		exhaustive = true
	case properties.ResourceUuidPlural:
		tmpl = resourceReadManyFunction
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
	case properties.ResourceCustom:
		var err error
		tmpl, err = getCustomTemplateForFunction(paramSpec, "ResourceRead")
		if err != nil {
			return "", err
		}
	}

	listAttributeVariant := properties.NewNameVariant(listAttribute)

	data := map[string]interface{}{
		"PluralType":                       paramSpec.TerraformProviderConfig.PluralType,
		"ResourceXpathVariablesWithChecks": paramSpec.ResourceXpathVariablesWithChecks(false),
		"ResourceOrDS":                     "Resource",
		"HasEncryptedResources":            paramSpec.HasEncryptedResources(),
		"ListAttribute":                    listAttributeVariant,
		"Exhaustive":                       exhaustive,
		"EntryOrConfig":                    paramSpec.EntryOrConfig(),
		"HasEntryName":                     paramSpec.HasEntryName(),
		"structName":                       names.StructName,
		"datasourceStructName":             names.DataSourceStructName,
		"resourceStructName":               names.ResourceStructName,
		"serviceName":                      naming.CamelCase("", serviceName, "", false),
		"resourceSDKName":                  resourceSDKName,
		"locations":                        paramSpec.OrderedLocations(),
	}

	funcMap := template.FuncMap{
		"RenderEncryptedValuesInitialization": func() (string, error) {
			return RenderEncryptedValuesInitialization(properties.SchemaResource, paramSpec, "read")
		},
		"RenderEncryptedValuesFinalizer": func() (string, error) {
			return RenderEncryptedValuesFinalizer(properties.SchemaResource, paramSpec)
		},
		"AttributesFromXpathComponents": func(target string) (string, error) { return paramSpec.AttributesFromXpathComponents(target) },
		"RenderLocationsPangoToState": func(source string, dest string) (string, error) {
			return RenderLocationsPangoToState(names, paramSpec, source, dest)
		},
		"RenderLocationsStateToPango": func(source string, dest string) (string, error) {
			return RenderLocationsStateToPango(names, paramSpec, source, dest)
		},
	}

	return processTemplate(tmpl, "resource-read-function", data, funcMap)
}

func ResourceUpdateFunction(resourceTyp properties.ResourceType, names *NameProvider, serviceName string, paramSpec *properties.Normalization, resourceSDKName string) (string, error) {
	if strings.Contains(serviceName, "group") {
		serviceName = "group"
	}

	var tmpl string
	var listAttribute string
	var exhaustive bool
	switch resourceTyp {
	case properties.ResourceEntry, properties.ResourceConfig:
		tmpl = resourceUpdateFunction
	case properties.ResourceEntryPlural:
		tmpl = resourceUpdateEntryListFunction
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
	case properties.ResourceUuid:
		tmpl = resourceUpdateManyFunction
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
		exhaustive = true
	case properties.ResourceUuidPlural:
		tmpl = resourceUpdateManyFunction
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
	case properties.ResourceCustom:
		var err error
		tmpl, err = getCustomTemplateForFunction(paramSpec, "Update")
		if err != nil {
			return "", err
		}
	}

	listAttributeVariant := properties.NewNameVariant(listAttribute)

	data := map[string]interface{}{
		"PluralType":            paramSpec.TerraformProviderConfig.PluralType,
		"HasEncryptedResources": paramSpec.HasEncryptedResources(),
		"ListAttribute":         listAttributeVariant,
		"Exhaustive":            exhaustive,
		"EntryOrConfig":         paramSpec.EntryOrConfig(),
		"HasEntryName":          paramSpec.HasEntryName(),
		"structName":            names.ResourceStructName,
		"serviceName":           naming.CamelCase("", serviceName, "", false),
		"resourceSDKName":       resourceSDKName,
	}

	funcMap := template.FuncMap{
		"RenderEncryptedValuesInitialization": func() (string, error) {
			return RenderEncryptedValuesInitialization(properties.SchemaResource, paramSpec, "update")
		},
		"RenderEncryptedValuesFinalizer": func() (string, error) {
			return RenderEncryptedValuesFinalizer(properties.SchemaResource, paramSpec)
		},
		"RenderCreateUpdateMovementRequired": func(state string, entries string) (string, error) {
			return RendeCreateUpdateMovementRequired(state, entries)
		},
		"RenderLocationsStateToPango": func(source string, dest string) (string, error) {
			return RenderLocationsStateToPango(names, paramSpec, source, dest)
		},
		"RenderLocationsPangoToState": func(source string, dest string) (string, error) {
			return RenderLocationsPangoToState(names, paramSpec, source, dest)
		},
	}

	return processTemplate(tmpl, "resource-update-function", data, funcMap)
}

func ResourceDeleteFunction(resourceTyp properties.ResourceType, names *NameProvider, serviceName string, paramSpec *properties.Normalization, resourceSDKName string) (string, error) {
	if strings.Contains(serviceName, "group") {
		serviceName = "group"
	}

	var tmpl string
	var listAttribute string
	var exhaustive string
	switch resourceTyp {
	case properties.ResourceEntry, properties.ResourceConfig:
		tmpl = resourceDeleteFunction
	case properties.ResourceEntryPlural:
		tmpl = resourceDeleteManyFunction
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
	case properties.ResourceUuid:
		tmpl = resourceDeleteManyFunction
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
		exhaustive = "exhaustive"
	case properties.ResourceUuidPlural:
		tmpl = resourceDeleteManyFunction
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
		exhaustive = "non-exhaustive"
	case properties.ResourceCustom:
		var err error
		tmpl, err = getCustomTemplateForFunction(paramSpec, "Delete")
		if err != nil {
			return "", err
		}
	}

	listAttributeVariant := properties.NewNameVariant(listAttribute)

	data := map[string]interface{}{
		"PluralType":            paramSpec.TerraformProviderConfig.PluralType,
		"HasEncryptedResources": paramSpec.HasEncryptedResources(),
		"HasImports":            len(paramSpec.Imports) > 0,
		"EntryOrConfig":         paramSpec.EntryOrConfig(),
		"ListAttribute":         listAttributeVariant,
		"Exhaustive":            exhaustive,
		"HasEntryName":          paramSpec.HasEntryName(),
		"structName":            names.ResourceStructName,
		"serviceName":           naming.CamelCase("", serviceName, "", false),
		"resourceSDKName":       resourceSDKName,
	}

	funcMap := template.FuncMap{
		"RenderEncryptedValuesInitialization": func() (string, error) {
			return RenderEncryptedValuesInitialization(properties.SchemaResource, paramSpec, "delete")
		},
		"RenderImportLocationAssignment": func(source string, dest string) (string, error) {
			return RenderImportLocationAssignment(names, paramSpec, source, dest)
		},
		"RenderLocationsStateToPango": func(source string, dest string) (string, error) {
			return RenderLocationsStateToPango(names, paramSpec, source, dest)
		},
	}

	return processTemplate(tmpl, "resource-delete-function", data, funcMap)
}

func ResourceOpenFunction(resourceTyp properties.ResourceType, names *NameProvider, serviceName string, paramSpec *properties.Normalization, resourceSDKName string) (string, error) {
	var tmpl string
	switch resourceTyp {
	case properties.ResourceEntry, properties.ResourceConfig:
	case properties.ResourceEntryPlural:
	case properties.ResourceUuid:
	case properties.ResourceUuidPlural:
		return "", fmt.Errorf("Ephemeral resources are only implemented for custom specs")
	case properties.ResourceCustom:
		var err error
		tmpl, err = getCustomTemplateForFunction(paramSpec, "Open")
		if err != nil {
			return "", err
		}
	}

	return processTemplate(tmpl, "resource-open-function", nil, nil)
}

func ResourceRenewFunction(resourceTyp properties.ResourceType, names *NameProvider, serviceName string, paramSpec *properties.Normalization, resourceSDKName string) (string, error) {
	var tmpl string
	switch resourceTyp {
	case properties.ResourceEntry, properties.ResourceConfig:
	case properties.ResourceEntryPlural:
	case properties.ResourceUuid:
	case properties.ResourceUuidPlural:
		return "", fmt.Errorf("Ephemeral resources are only implemented for custom specs")
	case properties.ResourceCustom:
		var err error
		tmpl, err = getCustomTemplateForFunction(paramSpec, "Renew")
		if err != nil {
			return "", err
		}
	}

	return processTemplate(tmpl, "resource-renew-function", nil, nil)
}

func ResourceCloseFunction(resourceTyp properties.ResourceType, names *NameProvider, serviceName string, paramSpec *properties.Normalization, resourceSDKName string) (string, error) {
	var tmpl string
	switch resourceTyp {
	case properties.ResourceEntry, properties.ResourceConfig:
	case properties.ResourceEntryPlural:
	case properties.ResourceUuid:
	case properties.ResourceUuidPlural:
		return "", fmt.Errorf("Ephemeral resources are only implemented for custom specs")
	case properties.ResourceCustom:
		var err error
		tmpl, err = getCustomTemplateForFunction(paramSpec, "Close")
		if err != nil {
			return "", err
		}
	}

	return processTemplate(tmpl, "resource-close-function", nil, nil)
}

func FunctionSupported(spec *properties.Normalization, function string) (bool, error) {
	switch function {
	case "Create", "Delete", "Read", "Update":
		return !spec.TerraformProviderConfig.Ephemeral, nil
	case "Open", "Close", "Renew":
		if !spec.TerraformProviderConfig.Ephemeral {
			return false, nil
		}

		if resource, found := customResourceFuncsMap[spec.TerraformProviderConfig.Suffix]; !found {
			return false, fmt.Errorf("cannot find a list of custom functions for %s", spec.TerraformProviderConfig.Suffix)
		} else {
			_, found := resource[function]
			return found, nil
		}
	default:
		return false, fmt.Errorf("invalid custom function name: %s", function)
	}
}

type importStateStructFieldSpec struct {
	Name          string
	TerraformType string
	Type          string
	Tags          string
}

type importStateStructSpec struct {
	StructName string
	Fields     []importStateStructFieldSpec
}

func createImportStateStructSpecs(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization) []importStateStructSpec {
	var specs []importStateStructSpec

	var fields []importStateStructFieldSpec
	fields = append(fields, importStateStructFieldSpec{
		Name:          "Location",
		TerraformType: fmt.Sprintf("%sLocation", names.StructName),
		Type:          "types.Object",
		Tags:          "`json:\"location\"`",
	})

	var resourceHasParent bool
	if spec.ResourceXpathVariablesWithChecks(false) {
		resourceHasParent = true
	}

	switch resourceTyp {
	case properties.ResourceEntry:
		if resourceHasParent {
			var xpathVariable *object.PanosXpathVariable
			for _, elt := range spec.PanosXpath.Variables {
				if elt.Name == "parent" {
					xpathVariable = &elt
				}
			}

			if xpathVariable == nil {
				panic("couldn't find parent variable for a child spec")
			}

			parentParam, err := spec.ParameterForPanosXpathVariable(*xpathVariable)
			if err != nil {
				panic(fmt.Sprintf("couldn't find matching param for xpath variable: %s", err.Error()))
			}

			fields = append(fields, importStateStructFieldSpec{
				Name: parentParam.Name.CamelCase,
				Type: "types.String",
				Tags: fmt.Sprintf("`json:\"%s\"`", parentParam.Name.Underscore),
			})
		}

		fields = append(fields, importStateStructFieldSpec{
			Name: "Name",
			Type: "types.String",
			Tags: "`json:\"name\"`",
		})
	case properties.ResourceEntryPlural:
		if resourceHasParent {
			var xpathVariable *object.PanosXpathVariable
			for _, elt := range spec.PanosXpath.Variables {
				if elt.Name == "parent" {
					xpathVariable = &elt
				}
			}

			if xpathVariable == nil {
				panic("couldn't find parent variable for a child spec")
			}

			parentParam, err := spec.ParameterForPanosXpathVariable(*xpathVariable)
			if err != nil {
				panic(fmt.Sprintf("couldn't find matching param for xpath variable: %s", err.Error()))
			}

			fields = append(fields, importStateStructFieldSpec{
				Name: parentParam.Name.CamelCase,
				Type: "types.String",
				Tags: fmt.Sprintf("`json:\"%s\"`", parentParam.Name.Underscore),
			})
		} else {
			fields = append(fields, importStateStructFieldSpec{
				Name: "Names",
				Type: "types.List",
				Tags: "`json:\"names\"`",
			})
		}
	case properties.ResourceUuid:
		fields = append(fields, importStateStructFieldSpec{
			Name: "Names",
			Type: "types.List",
			Tags: "`json:\"names\"`",
		})
	case properties.ResourceUuidPlural:
		fields = append(fields, importStateStructFieldSpec{
			Name: "Names",
			Type: "types.List",
			Tags: "`json:\"names\"`",
		})
		fields = append(fields, importStateStructFieldSpec{
			Name:          "Position",
			TerraformType: "TerraformPositionObject",
			Type:          "types.Object",
			Tags:          "`json:\"position\"`",
		})
	case properties.ResourceCustom, properties.ResourceConfig:
		panic("unreachable")
	}

	specs = append(specs, importStateStructSpec{
		StructName: fmt.Sprintf("%sImportState", names.StructName),
		Fields:     fields,
	})

	return specs
}

func createImportStateMarshallerSpecs(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization) []marshallerSpec {
	var specs []marshallerSpec

	var fields []marshallerFieldSpec

	fields = append(fields, marshallerFieldSpec{
		Name:       properties.NewNameVariant("location"),
		Type:       "types.Object",
		StructName: fmt.Sprintf("%sLocation", names.StructName),
		Tags:       "`json:\"location\"`",
	})

	var resourceHasParent bool
	if spec.ResourceXpathVariablesWithChecks(false) {
		resourceHasParent = true
	}

	switch resourceTyp {
	case properties.ResourceEntry:
		if resourceHasParent {
			var xpathVariable *object.PanosXpathVariable
			for _, elt := range spec.PanosXpath.Variables {
				if elt.Name == "parent" {
					xpathVariable = &elt
				}
			}

			if xpathVariable == nil {
				panic("couldn't find parent variable for a child spec")
			}

			parentParam, err := spec.ParameterForPanosXpathVariable(*xpathVariable)
			if err != nil {
				panic(fmt.Sprintf("couldn't find matching param for xpath variable: %s", err.Error()))
			}

			fields = append(fields, marshallerFieldSpec{
				Name: parentParam.Name,
				Type: "string",
				Tags: fmt.Sprintf("`json:\"%s\"`", parentParam.Name.Underscore),
			})
		}

		fields = append(fields, marshallerFieldSpec{
			Name: properties.NewNameVariant("name"),
			Type: "string",
			Tags: "`json:\"name\"`",
		})
	case properties.ResourceEntryPlural:
		if resourceHasParent {
			var xpathVariable *object.PanosXpathVariable
			for _, elt := range spec.PanosXpath.Variables {
				if elt.Name == "parent" {
					xpathVariable = &elt
				}
			}

			if xpathVariable == nil {
				panic("couldn't find parent variable for a child spec")
			}

			parentParam, err := spec.ParameterForPanosXpathVariable(*xpathVariable)
			if err != nil {
				panic(fmt.Sprintf("couldn't find matching param for xpath variable: %s", err.Error()))
			}

			fields = append(fields, marshallerFieldSpec{
				Name: parentParam.Name,
				Type: "string",
				Tags: fmt.Sprintf("`json:\"%s\"`", parentParam.Name.Underscore),
			})
		} else {
			fields = append(fields, marshallerFieldSpec{
				Name:       properties.NewNameVariant("names"),
				Type:       "types.List",
				StructName: "[]string",
				Tags:       "`json:\"names\"`",
			})
		}
	case properties.ResourceUuid:
		fields = append(fields, marshallerFieldSpec{
			Name:       properties.NewNameVariant("names"),
			Type:       "types.List",
			StructName: "[]string",
			Tags:       "`json:\"names\"`",
		})
	case properties.ResourceUuidPlural:
		fields = append(fields, marshallerFieldSpec{
			Name:       properties.NewNameVariant("names"),
			Type:       "types.List",
			StructName: "[]string",
			Tags:       "`json:\"names\"`",
		})
		fields = append(fields, marshallerFieldSpec{
			Name:       properties.NewNameVariant("position"),
			Type:       "types.Object",
			StructName: "TerraformPositionObject",
			Tags:       "`json:\"position\"`",
		})
	case properties.ResourceCustom, properties.ResourceConfig:
		panic(fmt.Sprintf("unreachable state: %s", resourceTyp))
	}

	specs = append(specs, marshallerSpec{
		StructName: fmt.Sprintf("%sImportState", names.StructName),
		Fields:     fields,
	})

	return specs
}

func RenderImportStateStructs(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization) (string, error) {
	// Only singular entries can be imported at the time
	if resourceTyp == properties.ResourceCustom || resourceTyp == properties.ResourceConfig {
		return "", nil
	}

	type context struct {
		Specs []importStateStructSpec
	}

	data := context{
		Specs: createImportStateStructSpecs(resourceTyp, names, spec),
	}

	return processTemplate(renderImportStateStructsTmpl, "render-import-state-structs", data, nil)
}

func ResourceImportStateFunction(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization) (string, error) {
	if resourceTyp == properties.ResourceConfig || resourceTyp == properties.ResourceCustom {
		return "", nil
	}

	type context struct {
		StructName      string
		PluralType      object.TerraformPluralType
		ResourceIsList  bool
		HasPosition     bool
		HasEntryName    bool
		ListAttribute   *properties.NameVariant
		ListStructName  string
		PangoStructName string
		HasParent       bool
		ParentAttribute *properties.NameVariant
	}

	data := context{
		StructName: names.StructName,
	}

	var resourceHasParent bool
	if spec.ResourceXpathVariablesWithChecks(false) {
		resourceHasParent = true
	}

	switch resourceTyp {
	case properties.ResourceEntry:
		if resourceHasParent {
			var xpathVariable *object.PanosXpathVariable
			for _, elt := range spec.PanosXpath.Variables {
				if elt.Name == "parent" {
					xpathVariable = &elt
				}
			}

			if xpathVariable == nil {
				panic("couldn't find parent variable for a child spec")
			}

			parentParam, err := spec.ParameterForPanosXpathVariable(*xpathVariable)
			if err != nil {
				panic(fmt.Sprintf("couldn't find matching param for xpath variable: %s", err.Error()))
			}

			data.ParentAttribute = parentParam.Name
			data.HasParent = true
		}
		data.HasEntryName = spec.HasEntryName()
	case properties.ResourceEntryPlural:
		if resourceHasParent {
			var xpathVariable *object.PanosXpathVariable
			for _, elt := range spec.PanosXpath.Variables {
				if elt.Name == "parent" {
					xpathVariable = &elt
				}
			}

			if xpathVariable == nil {
				panic("couldn't find parent variable for a child spec")
			}

			parentParam, err := spec.ParameterForPanosXpathVariable(*xpathVariable)
			if err != nil {
				panic(fmt.Sprintf("couldn't find matching param for xpath variable: %s", err.Error()))
			}

			data.PluralType = spec.TerraformProviderConfig.PluralType
			data.ParentAttribute = parentParam.Name
			data.HasParent = true

		} else {
			listAttribute := properties.NewNameVariant(spec.TerraformProviderConfig.PluralName)
			data.PluralType = spec.TerraformProviderConfig.PluralType
			data.ListAttribute = listAttribute
			data.ListStructName = fmt.Sprintf("%sResource%sObject", names.StructName, listAttribute.CamelCase)
			data.PangoStructName = fmt.Sprintf("%s.Entry", names.PackageName)
		}
	case properties.ResourceUuid, properties.ResourceUuidPlural:
		data.ResourceIsList = true
		data.PluralType = spec.TerraformProviderConfig.PluralType
		listAttribute := properties.NewNameVariant(spec.TerraformProviderConfig.PluralName)
		data.ListAttribute = properties.NewNameVariant(spec.TerraformProviderConfig.PluralName)
		data.ListStructName = fmt.Sprintf("%sResource%sObject", names.StructName, listAttribute.CamelCase)
		data.PangoStructName = fmt.Sprintf("%s.Entry", names.PackageName)
		if resourceTyp == properties.ResourceUuidPlural {
			data.HasPosition = true
		}
	case properties.ResourceCustom, properties.ResourceConfig:
		panic("unreachable")
	}

	funcMap := template.FuncMap{
		"ConfigToEntry": ConfigEntry,
		"RenderEncryptedValuesInitialization": func() (string, error) {
			return RenderEncryptedValuesInitialization(properties.SchemaResource, spec, "import")
		},
	}

	return processTemplate(resourceImportStateFunctionTmpl, "resource-import-state-function", data, funcMap)
}

func RenderImportStateCreator(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization) (string, error) {
	if resourceTyp == properties.ResourceConfig || resourceTyp == properties.ResourceCustom {
		return "", nil
	}

	type context struct {
		FuncName         string
		ModelName        string
		StructNamePrefix string
		ListAttribute    *properties.NameVariant
		ListStructName   string
		ResourceType     properties.ResourceType
		HasParent        bool
		ParentAttribute  *properties.NameVariant
	}

	data := context{
		FuncName:         fmt.Sprintf("%sImportStateCreator", names.StructName),
		ModelName:        fmt.Sprintf("%sModel", names.ResourceStructName),
		ResourceType:     resourceTyp,
		StructNamePrefix: names.StructName,
	}

	var resourceHasParent bool
	if spec.ResourceXpathVariablesWithChecks(false) {
		resourceHasParent = true
	}

	switch resourceTyp {
	case properties.ResourceEntry:
		if resourceHasParent {
			var xpathVariable *object.PanosXpathVariable
			for _, elt := range spec.PanosXpath.Variables {
				if elt.Name == "parent" {
					xpathVariable = &elt
				}
			}

			if xpathVariable == nil {
				panic("couldn't find parent variable for a child spec")
			}

			parentParam, err := spec.ParameterForPanosXpathVariable(*xpathVariable)
			if err != nil {
				panic(fmt.Sprintf("couldn't find matching param for xpath variable: %s", err.Error()))
			}

			data.HasParent = true
			data.ParentAttribute = parentParam.Name
		}
	case properties.ResourceEntryPlural:
		if resourceHasParent {
			var xpathVariable *object.PanosXpathVariable
			for _, elt := range spec.PanosXpath.Variables {
				if elt.Name == "parent" {
					xpathVariable = &elt
				}
			}

			if xpathVariable == nil {
				panic("couldn't find parent variable for a child spec")
			}

			parentParam, err := spec.ParameterForPanosXpathVariable(*xpathVariable)
			if err != nil {
				panic(fmt.Sprintf("couldn't find matching param for xpath variable: %s", err.Error()))
			}

			data.HasParent = true
			data.ParentAttribute = parentParam.Name
		} else {
			listAttribute := properties.NewNameVariant(spec.TerraformProviderConfig.PluralName)
			data.ListAttribute = listAttribute
			data.ListStructName = fmt.Sprintf("%sResource%sObject", names.StructName, listAttribute.CamelCase)
		}
	case properties.ResourceUuid, properties.ResourceUuidPlural:
		listAttribute := properties.NewNameVariant(spec.TerraformProviderConfig.PluralName)
		data.ListAttribute = listAttribute
		data.ListStructName = fmt.Sprintf("%sResource%sObject", names.StructName, listAttribute.CamelCase)
	case properties.ResourceCustom, properties.ResourceConfig:
		panic("unreachable")
	}

	return processTemplate(resourceImportStateCreatorTmpl, "render-import-state-creator", data, commonFuncMap)
}

func ConfigEntry(entryName string, param *properties.SpecParam) (string, error) {
	var entries []Entry

	paramType := param.Type
	if paramType == "" {
		paramType = "object"
	}
	entries = append(entries, Entry{
		Name: naming.CamelCase("", entryName, "", true),
		Type: paramType,
	})

	log.Printf("entries: %v", entries)

	entryData := EntryData{
		EntryName: entryName,
		Entries:   entries,
	}

	return processTemplate(resourceConfigEntry, "config-entry", entryData, nil)
}

func RenderResourceFuncMap(names map[string]properties.TerraformProviderSpecMetadata) (string, error) {
	type entry struct {
		Key        string
		StructName string
	}

	type context struct {
		Entries []entry
	}

	var entries []entry

	keys := make([]string, 0, len(names))
	for elt := range names {
		keys = append(keys, elt)
	}

	sort.Strings(keys)

	for _, key := range keys {
		if key == "" {
			continue
		}

		metadata := names[key]

		if metadata.Flags&properties.TerraformSpecImportable == 0 {
			continue
		}

		entries = append(entries, entry{
			Key:        fmt.Sprintf("panos%s", key),
			StructName: metadata.StructName,
		})
	}
	data := context{
		Entries: entries,
	}

	return processTemplate(resourceFuncMapTmpl, "resource-func-map", data, nil)
}

var customResourceFuncsMap = map[string]map[string]string{
	"device_group_parent": {
		"Imports":        deviceGroupParentImports,
		"DataSourceRead": deviceGroupParentDataSourceRead,
		"ResourceRead":   deviceGroupParentResourceRead,
		"Create":         deviceGroupParentResourceCreate,
		"Update":         deviceGroupParentResourceUpdate,
		"Delete":         deviceGroupParentResourceDelete,
		"Common":         deviceGroupParentCommon,
	},
	"api_key": {
		"Imports": apiKeyImports,
		"Open":    apiKeyOpen,
	},
	"vm_auth_key": {
		"Common":  vmAuthKeyCommon,
		"Imports": vmAuthKeyImports,
		"Open":    vmAuthKeyOpen,
	},
}
