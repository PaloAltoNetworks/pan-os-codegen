package terraform_provider

import (
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

type Entry struct {
	Name string
	Type string
}

type EntryData struct {
	EntryName string
	Entries   []Entry
}

type spec struct {
	Name            string
	HasEntryName    bool
	PangoType       string
	PangoReturnType string
	TerraformType   string
	ModelOrObject   string
	Params          map[string]*properties.SpecParam
	OneOf           map[string]*properties.SpecParam
}

func returnPangoTypeForProperty(pkgName string, parent string, prop *properties.SpecParam) string {
	if prop.Type == "" {
		return fmt.Sprintf("%s.%s", pkgName, parent)
	} else if prop.Type == "list" {
		if prop.Items.Type == "entry" {
			return fmt.Sprintf("%s.%s", pkgName, parent)
		} else {
			return fmt.Sprintf("%s.%s", pkgName, parent)
		}
	} else {
		if prop.Required {
			return fmt.Sprintf("%s.%s", pkgName, parent)
		} else {
			return fmt.Sprintf("%s.%s", pkgName, parent)
		}
	}
}

func generateFromTerraformToPangoSpec(pangoTypePrefix string, terraformPrefix string, paramSpec *properties.SpecParam) []spec {
	if paramSpec.Spec == nil {
		return nil
	}

	var specs []spec

	pangoType := fmt.Sprintf("%s%s", pangoTypePrefix, paramSpec.Name.CamelCase)

	pangoReturnType := fmt.Sprintf("%s%s", pangoTypePrefix, paramSpec.Name.CamelCase)
	terraformType := fmt.Sprintf("%s%s", terraformPrefix, paramSpec.Name.CamelCase)
	element := spec{
		PangoType:       pangoType,
		PangoReturnType: pangoReturnType,
		TerraformType:   terraformType,
		ModelOrObject:   "Object",
		Params:          paramSpec.Spec.Params,
		OneOf:           paramSpec.Spec.OneOf,
	}
	specs = append(specs, element)
	log.Printf("generateFromTerraformToPangoSpec() spec: %v", element)

	renderSpecsForParams := func(params map[string]*properties.SpecParam) {
		for _, elt := range params {
			if elt.Spec == nil {
				continue
			}
			terraformPrefix := fmt.Sprintf("%s%s", terraformPrefix, paramSpec.Name.CamelCase)
			log.Printf("Element: %s, pangoType: %s, terraformPrefix: %s", elt.Name.CamelCase, pangoType, terraformPrefix)
			specs = append(specs, generateFromTerraformToPangoSpec(pangoType, terraformPrefix, elt)...)
		}
	}

	renderSpecsForParams(paramSpec.Spec.Params)
	renderSpecsForParams(paramSpec.Spec.OneOf)

	return specs
}

func generateFromTerraformToPangoParameter(pkgName string, terraformPrefix string, pangoPrefix string, prop *properties.Normalization, parentName string) []spec {
	var specs []spec

	var pangoReturnType string
	if parentName == "" {
		pangoReturnType = fmt.Sprintf("%s.%s", pkgName, prop.EntryOrConfig())
		pangoPrefix = fmt.Sprintf("%s.", pkgName)
	} else {
		pangoReturnType = fmt.Sprintf("%s.%s", pkgName, parentName)
	}

	specs = append(specs, spec{
		HasEntryName:    prop.Entry != nil,
		PangoType:       pangoPrefix,
		PangoReturnType: pangoReturnType,
		ModelOrObject:   "Model",
		TerraformType:   terraformPrefix,
		Params:          prop.Spec.Params,
		OneOf:           prop.Spec.OneOf,
	})

	for _, elt := range prop.Spec.Params {
		specs = append(specs, generateFromTerraformToPangoSpec(pangoPrefix, terraformPrefix, elt)...)
	}

	for _, elt := range prop.Spec.OneOf {
		specs = append(specs, generateFromTerraformToPangoSpec(pangoPrefix, terraformPrefix, elt)...)
	}

	return specs
}

const copyToPangoTmpl = `
{{- define "terraformNestedElementsAssign" }}
  {{- with .Parameter }}

  {{- $result := .Name.LowerCamelCase }}
  {{- $diag := .Name.LowerCamelCase | printf "%s_diags" }}
	var {{ $result }}_entry *{{ $.Spec.PangoType }}{{ .Name.CamelCase }}
	if o.{{ .Name.CamelCase }} != nil {
		var {{ $diag }} diag.Diagnostics
		{{ $result }}_entry, {{ $diag }} = o.{{ .Name.CamelCase }}.CopyToPango(ctx)
		diags.Append({{ $diag }}...)
	}

  {{- end }}
{{- end }}

{{- define "terraformListElementsAs" }}
  {{- with .Parameter }}
    {{- $pangoType := printf "%s%s" $.Spec.PangoType .Name.CamelCase }}
    {{- $terraformType := printf "%s%s%s" $.Spec.TerraformType .Name.CamelCase $.Spec.ModelOrObject }}
    {{- $pangoEntries := printf "%s_pango_entries" .Name.LowerCamelCase }}
    {{- $tfEntries := printf "%s_tf_entries" .Name.LowerCamelCase }}
    {{- if eq .Items.Type "entry" }}
		var {{ $tfEntries }} []{{ $terraformType }}
		var {{ $pangoEntries }} []{{ $pangoType }}
	{
		d := o.{{ .Name.CamelCase }}.ElementsAs(ctx, &{{ $tfEntries }}, false)
		diags.Append(d...)
		for _, elt := range {{ $tfEntries }} {
			entry, d := elt.CopyToPango(ctx)
			diags.Append(d...)
			{{ $pangoEntries }} = append({{ $pangoEntries }}, *entry)
		}
	}
    {{- else }}
		{{ $pangoEntries }} := make([]{{ .Items.Type }}, 0)
	{
		d := o.{{ .Name.CamelCase }}.ElementsAs(ctx, &{{ $pangoEntries }}, false)
		diags.Append(d...)
	}
    {{- end }}
  {{- end }}
{{- end }}

{{- range .Specs }}
{{- $spec := . }}
func (o *{{ .TerraformType }}{{ .ModelOrObject }}) CopyToPango(ctx context.Context) (*{{ .PangoReturnType }}, diag.Diagnostics) {
	var diags diag.Diagnostics
  {{- range .Params }}
    {{- $terraformType := printf "%s%s" $spec.TerraformType .Name.CamelCase }}
    {{- if eq .Type "" }}
      {{- $pangoType := printf "%sObject" $spec.PangoType }}
	{{- template "terraformNestedElementsAssign" Map "Parameter" . "Spec" $spec }}
    {{- else if eq .Type "list" }}
      {{- $pangoType := printf "%s%s" $spec.PangoType .Name.CamelCase }}
	{{- template "terraformListElementsAs" Map "Parameter" . "Spec" $spec }}
    {{- end }}
  {{- end }}

  {{- range .OneOf }}
    {{- if eq .Type "" }}
      {{- $pangoType := printf "%sObject" $spec.PangoType }}
	{{- template "terraformNestedElementsAssign" Map "Parameter" . "Spec" $spec }}
    {{- else if eq .Type "list" }}
	{{- template "terraformListElementsAs" Map "Parameter" . "Spec" $spec }}
    {{- end }}
  {{- end }}

	result := &{{ .PangoReturnType }}{
{{- if .HasEntryName }}
	Name: *o.Name.ValueStringPointer(),
{{- end }}
  {{- range .Params }}
    {{- if eq .Type "" }}
	{{ .Name.CamelCase }}: {{ .Name.LowerCamelCase }}_entry,
    {{- else if eq .Type "list" }}
	{{ .Name.CamelCase }}: {{ .Name.LowerCamelCase }}_pango_entries,
    {{- else }}
	{{ .Name.CamelCase }}: o.{{ .Name.CamelCase }}.Value{{ CamelCaseType .Type }}Pointer(),
    {{- end }}
  {{- end }}

  {{- range .OneOf }}
    {{- if eq .Type "" }}
	{{ .Name.CamelCase }}: {{ .Name.LowerCamelCase }}_entry,
    {{- else if eq .Type "list" }}
	{{ .Name.CamelCase }}: {{ .Name.LowerCamelCase }}_pango_entries,
    {{- else }}
	{{ .Name.CamelCase }}: o.{{ .Name.CamelCase }}.Value{{ CamelCaseType .Type }}Pointer(),
    {{- end }}
  {{- end }}
	}

	return result, diags
}
{{- end }}
`

const copyFromPangoTmpl = `
{{- define "renderFromPangoToTfParameter" }}
  {{- if eq .Type "" }}
	// TODO: Missing implementation
  {{- else if eq .Type "list" }}
	{{ .Name.CamelCase }}: {{ .Name.LowerCamelCase }}_list,
  {{- end }}
{{- end }}

{{- define "renderListValueSimple" }}
var {{ .Name.LowerCamelCase }}_list types.List
{
	schema := rsschema.{{ .Type | PascalCase }}Attribute{}
	{{ .Name.LowerCamelCase }}_list, {{ .Name.LowerCamelCase }}_diags := types.ListValueFrom(ctx, obj.{{ .Name.CamelCase }}, schema.GetType())
	diags.Append({{ .Name.LowerCamelCase }}_diags...)
}
{{- end }}

{{- define "renderNestedValues" }}
  {{- range .Spec.Params }}
    {{- $terraformType := printf "%s%s" $.TerraformType (.Name.CamelCase) }}
    {{- if eq .Type "" }}
	// TODO {{ .Name.CamelCase }} {{ .Type }}
    {{- else if (and (eq .Type "list") (eq .Items.Type "entry")) }}
	{{- template "renderListValueEntry" Map "Name" .Name "Type" $terraformType }}
    {{- else if (and (eq .Type "list") (eq .Items.Type "member")) }}
	// TODO: {{ .Name.CamelCase }} {{ .Items.Type }}
    {{- else if (eq .Type "list") }}
	{{- template "renderListValueSimple" Map "Name" .Name "Type" .Items.Type }}
    {{- else }}
	// TODO: {{ .Name.CamelCase }} {{ .Type }}
    {{- end }}
  {{- end }}

  {{- range .Spec.OneOf }}
	// TODO: .Spec.OneOf {{ .Name.CamelCase }}
  {{- end }}
{{- end }}

{{- define "renderObjectListElement" }}
	entry := &{{ .TerraformType }} {
  {{- range .Element.Spec.Params }}
	{{- template "renderFromPangoToTfParameter" . }}
  {{- end }}
  {{- range .Element.Spec.OneOf }}
	{{- template "renderFromPangoToTfParameter" . }}
  {{- end }}
	}
	{{ .TfEntries }} = append({{ .TfEntries }}, *entry)
{{- end }}

{{- define "terraformListElementsAsParam" }}
  {{- with .Parameter }}
    {{- $pangoType := printf "%s%s" $.Spec.PangoType .Name.CamelCase }}
    {{- $terraformType := printf "%s%s%s" $.Spec.TerraformType .Name.CamelCase $.Spec.ModelOrObject }}
    {{- $terraformList := printf "%s_list" .Name.LowerCamelCase }}
    {{- $pangoEntries := printf "%s_pango_entries" .Name.LowerCamelCase }}
    {{- $tfEntries := printf "%s_tf_entries" .Name.LowerCamelCase }}
    {{- if eq .Items.Type "entry" }}
	var {{ $terraformList }} types.List
	{
		var {{ $tfEntries }} []{{ $terraformType }}
		for _, elt := range obj.{{ .Name.CamelCase }} {
			var entry {{ $terraformType }}
			entry_diags := entry.CopyFromPango(ctx, &elt)
			diags.Append(entry_diags...)
			{{ $tfEntries }} = append({{ $tfEntries }}, entry)
		}
		// var list_diags diags.Diagnostics
		// schemaType := ???
		// {{ $terraformList }} list_diags = types.ListValueFrom(cts, schemaType.GetType(), obj.{{ .Name.CamelCase }})
	}
    {{- else }}
		var {{ .Name.LowerCamelCase }}_list types.List
		{
			var list_diags diag.Diagnostics
			{{ .Name.LowerCamelCase }}_list, list_diags = types.ListValueFrom(ctx, types.{{ .Items.Type | PascalCase }}Type, obj.{{ .Name.CamelCase }})
			diags.Append(list_diags...)
		}
    {{- end }}
  {{- end }}
{{- end }}

{{- define "terraformListElementsAs" }}
  {{- range .Params }}
    {{- if eq .Type "list" }}
      {{- template "terraformListElementsAsParam" Map "Spec" $ "Parameter" . }}
    {{- end }}
  {{- end }}

  {{- range .OneOf }}
    {{- if eq .Type "list" }}
      {{- template "terraformListElementsAsParam" Map "Spec" $ "Parameter" . }}
    {{- end }}
  {{- end }}
{{- end }}

{{- define "terraformCreateEntryAssignmentForParam" }}
  {{- with .Parameter }}
  {{- $result := .Name.LowerCamelCase }}
  {{- $diag := .Name.LowerCamelCase | printf "%s_diags" }}
  var {{ $result }}_object *{{ $.Spec.TerraformType }}{{ .Name.CamelCase }}Object
  if obj.{{ .Name.CamelCase }} != nil {
	{{ $result }}_object = new({{ $.Spec.TerraformType }}{{ .Name.CamelCase }}Object)
  
	var {{ $diag }} diag.Diagnostics
	{{ $diag }} = {{ $result }}_object.CopyFromPango(ctx, obj.{{ .Name.CamelCase }})
	diags.Append({{ $diag }}...)
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

{{- define "terraformCreateSimpleValues" }}
  {{- range .Params }}
    {{- $terraformType := printf "types.%s" (.Type | PascalCase) }}
    {{- if (not (or (eq .Type "") (eq .Type "list"))) }}
	var {{ .Name.LowerCamelCase }}_value {{ $terraformType }}
	if obj.{{ .Name.CamelCase }} != nil {
		{{ .Name.LowerCamelCase }}_value = types.{{ .Type | PascalCase }}Value(*obj.{{ .Name.CamelCase }})
	}
    {{- end }}
  {{- end }}

  {{- range .OneOf }}
    {{- $terraformType := printf "types.%s" (.Type | PascalCase) }}
    {{- if (not (or (eq .Type "") (eq .Type "list"))) }}
	var {{ .Name.LowerCamelCase }}_value {{ $terraformType }}
	if obj.{{ .Name.CamelCase }} != nil {
		{{ .Name.LowerCamelCase }}_value = types.{{ .Type | PascalCase }}Value(*obj.{{ .Name.CamelCase }})
	}
    {{- end }}
  {{- end }}
{{- end }}

{{- define "assignFromPangoToTerraform" }}
  {{- with .Parameter }}
  {{- if eq .Type "" }}
	o.{{ .Name.CamelCase }} = {{ .Name.LowerCamelCase }}_object
  {{- else if eq .Type "list" }}
	o.{{ .Name.CamelCase }} = {{ .Name.LowerCamelCase }}_list
  {{- else }}
	o.{{ .Name.CamelCase }} = {{ .Name.LowerCamelCase }}_value
  {{- end }}
  {{- end }}
{{- end }}

{{- range .Specs }}
{{- $spec := . }}
{{ $terraformType := printf "%s%s" .TerraformType .ModelOrObject }}
func (o *{{ $terraformType }}) CopyFromPango(ctx context.Context, obj *{{ .PangoReturnType }}) diag.Diagnostics {
	var diags diag.Diagnostics
  {{- template "terraformListElementsAs" $spec }}
  {{- template "terraformCreateEntryAssignment" $spec }}
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

	var result []string
	for _, elt := range parts {
		result = append(result, strings.Title(elt))
	}

	return strings.Join(result, "")
}

func RenderCopyToPangoFunctions(pkgName string, terraformTypePrefix string, property *properties.Normalization) (string, error) {
	specs := generateFromTerraformToPangoParameter(pkgName, terraformTypePrefix, "", property, "")

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

func RenderCopyFromPangoFunctions(pkgName string, terraformTypePrefix string, property *properties.Normalization) (string, error) {
	specs := generateFromTerraformToPangoParameter(pkgName, terraformTypePrefix, "", property, "")

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

const renderLocationTmpl = `
{{- range .Locations }}
type {{ .StructName }} struct {
  {{- range .Fields }}
	{{ .Name }} {{ .Type }} {{ range .Tags }}{{ . }} {{ end }}
  {{- end }}
}
{{- end }}
`

func RenderLocationStructs(names *NameProvider, spec *properties.Normalization) (string, error) {
	type fieldCtx struct {
		Name string
		Type string
		Tags []string
	}

	type locationCtx struct {
		StructName string
		Fields     []fieldCtx
	}

	type context struct {
		Locations []locationCtx
	}

	var locations []locationCtx

	// Create the top location structure that references other locations
	topLocation := locationCtx{
		StructName: fmt.Sprintf("%sLocation", names.StructName),
	}

	for _, data := range spec.Locations {
		structName := fmt.Sprintf("%s%sLocation", names.StructName, data.Name.CamelCase)
		tfTag := fmt.Sprintf("`tfsdk:\"%s\"`", data.Name.Underscore)
		var structType string
		if len(data.Vars) > 0 {
			structType = fmt.Sprintf("*%s", structName)
		} else {
			structType = "types.Bool"
		}

		topLocation.Fields = append(topLocation.Fields, fieldCtx{
			Name: data.Name.CamelCase,
			Type: structType,
			Tags: []string{tfTag},
		})

		if len(data.Vars) == 0 {
			continue
		}

		var fields []fieldCtx
		for _, param := range data.Vars {
			paramTag := fmt.Sprintf("`tfsdk:\"%s\"`", param.Name.Underscore)
			fields = append(fields, fieldCtx{
				Name: param.Name.CamelCase,
				Type: "types.String",
				Tags: []string{paramTag},
			})
		}

		location := locationCtx{
			StructName: structName,
			Fields:     fields,
		}
		locations = append(locations, location)
	}

	locations = append(locations, topLocation)

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
	Required: true
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

type attributeCtx struct {
	Package      string
	Name         *properties.NameVariant
	SchemaType   string
	ElementType  string
	Description  string
	Required     bool
	Computed     bool
	Optional     bool
	Default      *defaultCtx
	ModifierType string
	Attributes   []attributeCtx
}

type schemaCtx struct {
	StructName  string
	ReturnType  string
	Package     string
	Description string
	Required    bool
	Computed    bool
	Optional    bool
	Attributes  []attributeCtx
}

func RenderLocationSchemaGetter(names *NameProvider, spec *properties.Normalization) (string, error) {
	var attributes []attributeCtx
	for _, data := range spec.Locations {
		var schemaType string
		if len(data.Vars) == 0 {
			schemaType = "rsschema.BoolAttribute"
		} else {
			schemaType = "rsschema.SingleNestedAttribute"
		}

		var variableAttrs []attributeCtx
		for _, variable := range data.Vars {
			attribute := attributeCtx{
				Name:        variable.Name,
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

		var modifierType string
		if len(variableAttrs) > 0 {
			modifierType = "Object"
		} else {
			modifierType = "Bool"
		}

		attribute := attributeCtx{
			Name:         data.Name,
			SchemaType:   schemaType,
			Description:  data.Description,
			Required:     false,
			Attributes:   variableAttrs,
			ModifierType: modifierType,
		}
		attributes = append(attributes, attribute)
	}

	locationName := &properties.NameVariant{
		Underscore:     naming.Underscore("", "location", ""),
		CamelCase:      naming.CamelCase("", "location", "", true),
		LowerCamelCase: naming.CamelCase("", "location", "", false),
	}

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

func createSchemaSpecForParameter(structPrefix string, packageName string, param *properties.SpecParam) []schemaCtx {
	var schemas []schemaCtx

	if param.Spec == nil {
		return nil
	}

	var returnType string
	switch param.Type {
	case "":
		returnType = "SingleNestedAttribute"
	case "list":
		switch param.Items.Type {
		case "entry":
			returnType = "NestedAttributeObject"
		}
	}

	structName := fmt.Sprintf("%s%s", structPrefix, param.Name.CamelCase)

	var attributes []attributeCtx
	for _, elt := range param.Spec.Params {
		attributes = append(attributes, createSchemaAttributeForParameter(packageName, elt))
	}

	for _, elt := range param.Spec.OneOf {
		attributes = append(attributes, createSchemaAttributeForParameter(packageName, elt))
	}

	schemas = append(schemas, schemaCtx{
		Package:     packageName,
		StructName:  structName,
		ReturnType:  returnType,
		Description: "",
		Required:    param.Required,
		Optional:    !param.Required,
		Attributes:  attributes,
	})

	for _, elt := range param.Spec.Params {
		if elt.Type == "" || (elt.Type == "list" && elt.Items.Type == "entry") {
			schemas = append(schemas, createSchemaSpecForParameter(structName, packageName, elt)...)
		}
	}

	for _, elt := range param.Spec.OneOf {
		if elt.Type == "" || (elt.Type == "list" && elt.Items.Type == "entry") {
			schemas = append(schemas, createSchemaSpecForParameter(structName, packageName, elt)...)
		}
	}

	return schemas
}

func createSchemaAttributeForParameter(packageName string, param *properties.SpecParam) attributeCtx {
	var schemaType, elementType string
	switch param.Type {
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
	default:
		schemaType = fmt.Sprintf("%sAttribute", pascalCase(param.Type))
	}

	var defaultValue *defaultCtx
	if param.Default != "" {
		defaultValue = &defaultCtx{
			Type:  fmt.Sprintf("%sdefault.%sDefault", param.Type, pascalCase(param.Type)),
			Value: param.Default,
		}
	}

	return attributeCtx{
		Package:     packageName,
		Name:        param.Name,
		SchemaType:  schemaType,
		ElementType: elementType,
		Description: param.Description,
		Required:    param.Required,
		Optional:    !param.Required,
		Default:     defaultValue,
	}
}

type schemaType int

const (
	schemaDataSource schemaType = iota
	schemaResource
)

func createSchemaSpecForNormalization(typ schemaType, spec *properties.Normalization) []schemaCtx {
	var schemas []schemaCtx

	var packageName string
	switch typ {
	case schemaDataSource:
		packageName = "dsschema"
	case schemaResource:
		packageName = "rsschema"
	}

	if spec.Spec == nil {
		return nil
	}

	names := NewNameProvider(spec)

	var structName string
	switch typ {
	case schemaDataSource:
		structName = names.DataSourceStructName
	case schemaResource:
		structName = names.ResourceStructName
	}

	var attributes []attributeCtx

	location := &properties.NameVariant{
		Underscore:     naming.Underscore("", "location", ""),
		CamelCase:      naming.CamelCase("", "location", "", true),
		LowerCamelCase: naming.CamelCase("", "location", "", false),
	}

	attributes = append(attributes, attributeCtx{
		Package:    packageName,
		Name:       location,
		Required:   true,
		SchemaType: "SingleNestedAttribute",
	})

	tfid := &properties.NameVariant{
		Underscore:     naming.Underscore("", "tfid", ""),
		CamelCase:      naming.CamelCase("", "tfid", "", true),
		LowerCamelCase: naming.CamelCase("", "tfid", "", false),
	}

	attributes = append(attributes, attributeCtx{
		Package:     packageName,
		Name:        tfid,
		SchemaType:  "StringAttribute",
		Description: "The Terraform ID.",
		Computed:    true,
	})

	if spec.HasEntryName() {
		name := &properties.NameVariant{
			Underscore:     naming.Underscore("", "name", ""),
			CamelCase:      naming.CamelCase("", "name", "", true),
			LowerCamelCase: naming.CamelCase("", "name", "", false),
		}

		attributes = append(attributes, attributeCtx{
			Package:    packageName,
			Name:       name,
			SchemaType: "StringAttribute",
			Required:   true,
		})
	}

	for _, elt := range spec.Spec.Params {
		attributes = append(attributes, createSchemaAttributeForParameter(packageName, elt))
		schemas = append(schemas, createSchemaSpecForParameter(structName, packageName, elt)...)
	}

	for _, elt := range spec.Spec.OneOf {
		attributes = append(attributes, createSchemaAttributeForParameter(packageName, elt))
		schemas = append(schemas, createSchemaSpecForParameter(structName, packageName, elt)...)
	}

	schemas = append(schemas, schemaCtx{
		Package:    packageName,
		StructName: structName,
		ReturnType: "Schema",
		Attributes: attributes,
	})

	return schemas
}

const renderSchemaTemplate = `
{{- define "renderSchemaListAttribute" }}
	"{{ .Name.Underscore }}": {{ .Package }}.{{ .SchemaType }} {
		Required: {{ .Required }},
		Optional: {{ .Optional }},
		Computed: {{ .Computed }},
		ElementType: {{ .ElementType }},
	},
{{- end }}

{{- define "renderSchemaListNestedAttribute" }}
  {{- with .Attribute }}
	"{{ .Name.Underscore }}": {{ .Package }}.{{ .SchemaType }} {
		Required: {{ .Required }},
		Optional: {{ .Optional }},
		Computed: {{ .Computed }},
		NestedObject: {{ $.StructName }}{{ .Name.CamelCase }}Schema(),
	},
  {{- end }}
{{- end }}

{{- define "renderSchemaSingleNestedAttribute" }}
  {{- with .Attribute }}
	"{{ .Name.Underscore }}": {{ $.StructName }}{{ .Name.CamelCase }}Schema(),
  {{- end }}
{{- end }}

{{- define "renderSchemaSimpleAttribute" }}
	"{{ .Name.Underscore }}": {{ .Package }}.{{ .SchemaType }} {
		Description: "{{ .Description }}",
		Computed: {{ .Computed }},
		Required: {{ .Required }},
		Optional: {{ .Optional }},
	},
{{- end }}

{{- define "renderSchemaAttribute" }}
  {{- with .Attribute }}
    {{ if eq .SchemaType "ListAttribute" }}
      {{- template "renderSchemaListAttribute" . }}
    {{- else if eq .SchemaType "ListNestedAttribute" }}
      {{- template "renderSchemaListNestedAttribute" Map "StructName" $.StructName "Attribute" . }}
    {{- else if eq .SchemaType "SingleNestedAttribute" }}
      {{- template "renderSchemaSingleNestedAttribute" Map "StructName" $.StructName "Attribute" . }}
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
		Required: {{ .Required }},
		Computed: {{ .Computed }},
		Optional: {{ .Optional }},
{{- end }}
		Attributes: map[string]{{ .Package }}.Attribute{
  {{- range .Attributes -}}
	{{- template "renderSchemaAttribute" Map "StructName" $schema.StructName "Attribute" . }}
  {{- end }}
		},
	}
}
{{- end }}
`

func RenderResourceSchema(names *NameProvider, spec *properties.Normalization) (string, error) {
	type context struct {
		Schemas []schemaCtx
	}

	data := context{
		Schemas: createSchemaSpecForNormalization(schemaResource, spec),
	}

	return processTemplate(renderSchemaTemplate, "render-resource-schema", data, commonFuncMap)
}

type locationFieldCtx struct {
	Name string
	Type string
}

type locationCtx struct {
	Name                string
	TerraformStructName string
	SdkStructName       string
	IsBool              bool
	Fields              []locationFieldCtx
}

func renderLocationsGetContext(names *NameProvider, spec *properties.Normalization) []locationCtx {
	var locations []locationCtx

	for _, location := range spec.Locations {
		var fields []locationFieldCtx
		for _, variable := range location.Vars {
			fields = append(fields, locationFieldCtx{
				Name: variable.Name.CamelCase,
				Type: "String",
			})
		}
		locations = append(locations, locationCtx{
			Name:                location.Name.CamelCase,
			TerraformStructName: fmt.Sprintf("%s%sLocation", names.StructName, location.Name.CamelCase),
			SdkStructName:       fmt.Sprintf("%s.%sLocation", names.PackageName, location.Name.CamelCase),
			IsBool:              len(location.Vars) == 0,
			Fields:              fields,
		})
	}

	type context struct {
		Locations []locationCtx
	}

	return locations
}

const locationsPangoToState = `
{{- range .Locations }}
  {{- if .IsBool }}
if loc.Location.{{ .Name }} {
	state.Location.{{ .Name }} = types.BoolValue(true)
}
  {{- else }}
if loc.Location.{{ .Name }} != nil {
	location := &{{ .TerraformStructName }}{
    {{ $locationName := .Name }}
    {{- range .Fields }}
		{{ .Name }}: types.{{ .Type }}Value(loc.Location.{{ $locationName }}.{{ .Name }}),
    {{- end }}
	}
	state.Location.{{ .Name }} = location
}
  {{- end }}
{{- end }}
`

func RenderLocationsPangoToState(names *NameProvider, spec *properties.Normalization) (string, error) {
	type context struct {
		Locations []locationCtx
	}
	data := context{Locations: renderLocationsGetContext(names, spec)}
	return processTemplate(locationsPangoToState, "render-locations-pango-to-state", data, commonFuncMap)
}

const locationsStateToPango = `
{{- range .Locations }}
  {{- if .IsBool }}
if !state.Location.{{ .Name }}.IsNull() && state.Location.{{ .Name }}.ValueBool() {
	loc.Location.{{ .Name }} = true
}
  {{- else }}
if state.Location.{{ .Name }} != nil {
	location := &{{ .SdkStructName }}{
    {{ $locationName := .Name }}
    {{- range .Fields }}
		{{ .Name }}: state.Location.{{ $locationName }}.{{ .Name }}.ValueString(),
    {{- end }}
	}
	loc.Location.{{ .Name }} = location
}
  {{- end }}
{{- end }}
`

func RenderLocationsStateToPango(names *NameProvider, spec *properties.Normalization) (string, error) {
	type context struct {
		Locations []locationCtx
	}
	data := context{Locations: renderLocationsGetContext(names, spec)}
	return processTemplate(locationsStateToPango, "render-locations-state-to-pango", data, commonFuncMap)
}

const dataSourceStructs = `
{{- range .Structs }}
type {{ .StructName }}{{ .ModelOrObject }} struct {
  {{- range .Fields }}
	{{ .Name }} {{ .Type }} {{ range .Tags }}{{ . }} {{ end }}
  {{- end }}
}
{{- end }}
`

type datasourceStructFieldSpec struct {
	Name string
	Type string
	Tags []string
}

type datasourceStructSpec struct {
	StructName    string
	ModelOrObject string
	Fields        []datasourceStructFieldSpec
}

func terraformTypeForProperty(structPrefix string, prop *properties.SpecParam) string {
	if prop.Type == "" {
		return fmt.Sprintf("*%s%sObject", structPrefix, prop.Name.CamelCase)
	}

	if prop.Type == "list" && prop.Items.Type == "entry" {
		return fmt.Sprintf("[]%s%sObject", structPrefix, prop.Name.CamelCase)
	}

	if prop.Type == "list" {
		return fmt.Sprintf("[]types.%s", pascalCase(prop.Items.Type))
	}

	return fmt.Sprintf("types.%s", pascalCase(prop.Type))
}

func structFieldSpec(param *properties.SpecParam, structPrefix string) datasourceStructFieldSpec {
	tfTag := fmt.Sprintf("`tfsdk:\"%s\"`", param.Name.Underscore)
	return datasourceStructFieldSpec{
		Name: param.Name.CamelCase,
		Type: terraformTypeForProperty(structPrefix, param),
		Tags: []string{tfTag},
	}
}

func dataSourceStructContextForParam(structPrefix string, param *properties.SpecParam) []datasourceStructSpec {
	var structs []datasourceStructSpec

	structName := fmt.Sprintf("%s%s", structPrefix, param.Name.CamelCase)

	var params []datasourceStructFieldSpec
	if param.Spec != nil {
		for _, elt := range param.Spec.Params {
			params = append(params, structFieldSpec(elt, structName))
		}

		for _, elt := range param.Spec.OneOf {
			params = append(params, structFieldSpec(elt, structName))
		}
	}

	structs = append(structs, datasourceStructSpec{
		StructName:    structName,
		ModelOrObject: "Object",
		Fields:        params,
	})

	if param.Spec == nil {
		return structs
	}

	for _, elt := range param.Spec.Params {
		if elt.Type == "" || (elt.Type == "list" && elt.Items.Type == "entry") {
			structs = append(structs, dataSourceStructContextForParam(structName, elt)...)
		}
	}

	for _, elt := range param.Spec.OneOf {
		if elt.Type == "" || (elt.Type == "list" && elt.Items.Type == "entry") {
			structs = append(structs, dataSourceStructContextForParam(structName, elt)...)
		}
	}

	return structs
}

func dataSourceStructContext(spec *properties.Normalization) []datasourceStructSpec {
	var structs []datasourceStructSpec

	if spec.Spec == nil {
		return nil
	}

	names := NewNameProvider(spec)

	var fields []datasourceStructFieldSpec

	for _, elt := range spec.Spec.Params {
		fields = append(fields, structFieldSpec(elt, names.DataSourceStructName))
		if elt.Type == "" || (elt.Type == "list" && elt.Items.Type == "entry") {
			structs = append(structs, dataSourceStructContextForParam(names.DataSourceStructName, elt)...)
		}
	}

	for _, elt := range spec.Spec.OneOf {
		fields = append(fields, structFieldSpec(elt, names.DataSourceStructName))
		if elt.Type == "" || (elt.Type == "list" && elt.Items.Type == "entry") {
			structs = append(structs, dataSourceStructContextForParam(names.DataSourceStructName, elt)...)
		}
	}

	structs = append(structs, datasourceStructSpec{
		StructName:    names.DataSourceStructName,
		ModelOrObject: "Model",
		Fields:        fields,
	})
	return structs
}

func RenderDataSourceStructs(names *NameProvider, spec *properties.Normalization) (string, error) {
	type context struct {
		Structs []datasourceStructSpec
	}

	data := context{
		Structs: dataSourceStructContext(spec),
	}

	return processTemplate(dataSourceStructs, "render-locations-state-to-pango", data, commonFuncMap)
}

func ResourceCreateFunction(names *NameProvider, serviceName string, paramSpec *properties.Normalization, terraformProvider *properties.TerraformProviderFile, resourceSDKName string) (string, error) {
	funcMap := template.FuncMap{
		"ConfigToEntry":               ConfigEntry,
		"RenderLocationsStateToPango": func() (string, error) { return RenderLocationsStateToPango(names, paramSpec) },
		"ResourceParamToSchema": func(paramName string, paramParameters properties.SpecParam) (string, error) {
			return ParamToSchemaResource(paramName, paramParameters, terraformProvider)
		},
	}

	if strings.Contains(serviceName, "group") && serviceName != "Device group" {
		serviceName = "group"
	}

	data := map[string]interface{}{
		"HasEncryptedResources": paramSpec.HasEncryptedResources(),
		"EntryOrConfig":         paramSpec.EntryOrConfig(),
		"HasEntryName":          paramSpec.HasEntryName(),
		"structName":            names.ResourceStructName,
		"serviceName":           naming.CamelCase("", serviceName, "", false),
		"paramSpec":             paramSpec.Spec,
		"resourceSDKName":       resourceSDKName,
		"locations":             paramSpec.Locations,
	}

	return processTemplate(resourceCreateFunction, "resource-create-function", data, funcMap)
}

func ResourceReadFunction(names *NameProvider, serviceName string, paramSpec *properties.Normalization, resourceSDKName string) (string, error) {
	if strings.Contains(serviceName, "group") {
		serviceName = "group"
	}

	data := map[string]interface{}{
		"HasEncryptedResources": paramSpec.HasEncryptedResources(),
		"EntryOrConfig":         paramSpec.EntryOrConfig(),
		"HasEntryName":          paramSpec.HasEntryName(),
		"structName":            names.StructName,
		"resourceStructName":    names.ResourceStructName,
		"serviceName":           naming.CamelCase("", serviceName, "", false),
		"resourceSDKName":       resourceSDKName,
		"locations":             paramSpec.Locations,
	}

	funcMap := template.FuncMap{
		"RenderLocationsPangoToState": func() (string, error) { return RenderLocationsPangoToState(names, paramSpec) },
	}

	return processTemplate(resourceReadFunction, "resource-read-function", data, funcMap)
}

func ResourceUpdateFunction(names *NameProvider, serviceName string, paramSpec *properties.Normalization, resourceSDKName string) (string, error) {
	if strings.Contains(serviceName, "group") {
		serviceName = "group"
	}

	data := map[string]interface{}{
		"HasEncryptedResources": paramSpec.HasEncryptedResources(),
		"EntryOrConfig":         paramSpec.EntryOrConfig(),
		"HasEntryName":          paramSpec.HasEntryName(),
		"structName":            names.ResourceStructName,
		"serviceName":           naming.CamelCase("", serviceName, "", false),
		"resourceSDKName":       resourceSDKName,
	}

	funcMap := template.FuncMap{
		"RenderLocationsStateToPango": func() (string, error) { return RenderLocationsStateToPango(names, paramSpec) },
	}

	return processTemplate(resourceUpdateFunction, "resource-update-function", data, funcMap)
}

func ResourceDeleteFunction(structName string, serviceName string, paramSpec *properties.Normalization, resourceSDKName string) (string, error) {
	if strings.Contains(serviceName, "group") {
		serviceName = "group"
	}

	data := map[string]interface{}{
		"HasEncryptedResources": paramSpec.HasEncryptedResources(),
		"EntryOrConfig":         paramSpec.EntryOrConfig(),
		"HasEntryName":          paramSpec.HasEntryName(),
		"structName":            structName,
		"serviceName":           naming.CamelCase("", serviceName, "", false),
		"resourceSDKName":       resourceSDKName,
	}

	return processTemplate(resourceDeleteFunction, "resource-delete-function", data, nil)
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
