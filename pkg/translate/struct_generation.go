package translate

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/version"
)

var xmlNameVariant = &properties.NameVariant{
	Original:       "xml-name",
	LowerCamelCase: "xmlName",
	CamelCase:      "XMLName",
	Underscore:     "xml_name",
}

type entryStructFieldContext struct {
	Name             *properties.NameVariant
	IsInternal       bool
	Required         bool
	FieldType        string
	Type             string
	ItemsType        string
	XmlType          string
	XmlContainerType string
	ItemsXmlType     string
	Tags             string
	version          *version.Version
}

func (o entryStructFieldContext) FinalType() string {
	switch o.FieldType {
	case "list-entry", "list-member":
		return o.ItemsType
	case "object", "simple":
		if o.Required {
			return o.Type
		} else {
			return "*" + o.Type
		}
	case "internal":
		return o.Type
	default:
		panic(fmt.Sprintf("unreachable FieldType '%s' for '%s'", o.FieldType, o.Name.CamelCase))
	}
}

func (o entryStructFieldContext) FinalXmlType() string {
	switch o.FieldType {
	case "list-entry":
		if o.XmlContainerType != "" {
			return "*" + o.XmlContainerType
		} else {
			return o.ItemsXmlType
		}
	case "list-member":
		return "*" + o.ItemsXmlType
	case "object", "simple":
		if o.Required {
			return o.XmlType
		} else {
			return "*" + o.XmlType
		}
	case "internal":
		return o.XmlType
	default:
		panic(fmt.Sprintf("unreachable FieldType '%s' for '%s'", o.FieldType, o.Name.CamelCase))
	}
}

type entryStructContext struct {
	TopLevel       bool
	IsXmlContainer bool
	Fields         []entryStructFieldContext

	version *version.Version
	name    *properties.NameVariant
}

func (o entryStructContext) versionSuffix() string {
	if o.version == nil {
		return ""
	}

	return fmt.Sprintf("_%s", strings.ReplaceAll(o.version.String(), ".", "_"))
}

func (o entryStructContext) StructName() string {
	return o.name.CamelCase
}

func (o entryStructContext) XmlStructName() string {
	return o.name.LowerCamelCase + "Xml" + o.versionSuffix()
}

func (o entryStructContext) XmlContainerStructName() string {
	return o.name.LowerCamelCase + "XmlContainer" + o.versionSuffix()
}

func (o entryStructContext) SpecifierFuncName(suffix string) string {
	return "specify" + suffix + o.versionSuffix()
}

func getTypesForParam(structTyp structType, parent *properties.NameVariant, param *properties.SpecParam, version *version.Version, overrideForXmlContainer bool) (string, string, string) {
	var versionSuffix string
	if version != nil {
		versionSuffix = fmt.Sprintf("_%s", strings.ReplaceAll(version.String(), ".", "_"))
	}

	if structTyp == structXmlType {
		typ := ParamType(structXmlType, parent, param, versionSuffix)
		var itemsType string
		if param.Type == "list" && param.Items.Type == "entry" {
			itemsType = "[]" + typ
		} else if param.Type == "list" {
			itemsType = "util.MemberType"
		}
		var xmlContainerType string
		if overrideForXmlContainer {
			xmlContainerType = parent.WithSuffix(param.Name).WithSuffix(properties.NewNameVariant("container")).WithSuffix(properties.NewNameVariant("xml")).WithLiteralSuffix(versionSuffix).LowerCamelCase
		}
		return typ, itemsType, xmlContainerType
	} else {
		typ := ParamType(structApiType, parent, param, "")
		var itemsType string
		if param.Type == "list" && param.Items.Type == "string" {
			itemsType = "[]string"
		} else if param.Type == "list" && param.Items.Type == "int64" {
			itemsType = "[]int64"
		} else if param.Type == "list" && param.Items.Type == "entry" {
			itemsType = "[]" + typ
		}
		return typ, itemsType, ""
	}
}

func getFieldTypeForParam(param *properties.SpecParam) string {
	if param.Type == "" {
		return "object"
	}

	if param.Type == "list" && param.Items.Type == "entry" {
		return "list-entry"
	}

	if param.Type == "list" {
		return "list-member"
	}

	return "simple"
}

func createStructSpecForXmlListContainer(prefix *properties.NameVariant, param *properties.SpecParam, version *version.Version) []entryStructContext {
	typ, itemsType, _ := getTypesForParam(structApiType, prefix, param, version, false)
	xmlType, itemsXmlType, _ := getTypesForParam(structXmlType, prefix, param, version, false)
	fieldType := "list-entry"

	fields := []entryStructFieldContext{
		{
			Name:         properties.NewNameVariant("entries"),
			Required:     false,
			FieldType:    fieldType,
			Type:         typ,
			ItemsType:    itemsType,
			XmlType:      xmlType,
			ItemsXmlType: itemsXmlType,
			Tags:         "`xml:\"entry\"`",
			version:      version,
		},
	}

	return []entryStructContext{{
		IsXmlContainer: true,
		Fields:         fields,
		name:           prefix.WithSuffix(param.PangoNameVariant()).WithSuffix(properties.NewNameVariant("container")),
		version:        version,
	}}
}

func createEntryXmlStructSpecsForParameter(structTyp structType, parentPrefix *properties.NameVariant, param *properties.SpecParam, version *version.Version) []entryStructContext {
	var fields []entryStructFieldContext
	var entries []entryStructContext

	if param.Type == "list" && param.Items.Type == "entry" {
		if structTyp == structXmlType {
			fields = append(fields, entryStructFieldContext{
				IsInternal: true,
				FieldType:  "internal",
				Name:       xmlNameVariant,
				XmlType:    "xml.Name",
				Tags:       "`xml:\"entry\"`",
			})
		}
		fields = append(fields, entryStructFieldContext{
			Name:      properties.NewNameVariant("name"),
			Required:  true,
			FieldType: "simple",
			Type:      "string",
			XmlType:   "string",
			Tags:      "`xml:\"name,attr\"`",
		})
	}

	processParameter := func(prefix *properties.NameVariant, param *properties.SpecParam) {
		if param.GoSdkConfig != nil && param.GoSdkConfig.Skip != nil && *param.GoSdkConfig.Skip {
			return
		}

		if !ParamSupportedInVersion(param, version) {
			return
		}

		var overrideTypeForXmlContainer bool
		if structTyp == structXmlType && param.Type == "list" && param.Items.Type == "entry" {
			overrideTypeForXmlContainer = true
			entries = append(entries, createStructSpecForXmlListContainer(prefix, param, version)...)
		}

		typ, itemsType, _ := getTypesForParam(structApiType, prefix, param, version, overrideTypeForXmlContainer)
		xmlType, itemsXmlType, xmlContainerType := getTypesForParam(structXmlType, prefix, param, version, overrideTypeForXmlContainer)
		fieldType := getFieldTypeForParam(param)

		fields = append(fields, entryStructFieldContext{
			Name:             param.PangoNameVariant(),
			Required:         param.Required,
			FieldType:        fieldType,
			Type:             typ,
			ItemsType:        itemsType,
			XmlType:          xmlType,
			XmlContainerType: xmlContainerType,
			ItemsXmlType:     itemsXmlType,
			Tags:             XmlTag(param),
			version:          version,
		})

		if param.Type == "" || (param.Type == "list" && param.Items.Type == "entry") {
			entries = append(entries, createEntryXmlStructSpecsForParameter(structTyp, prefix, param, version)...)
		}

	}

	prefixName := parentPrefix.WithSuffix(param.Name)
	for _, elt := range param.Spec.SortedParams() {
		processParameter(prefixName, elt)
	}

	for _, elt := range param.Spec.SortedOneOf() {
		processParameter(prefixName, elt)
	}

	fields = append(fields, []entryStructFieldContext{
		{
			Name:      properties.NewNameVariant("misc"),
			FieldType: "internal",
			Type:      "[]generic.Xml",
			XmlType:   "[]generic.Xml",
			Tags:      "`xml:\",any\"`",
		},
		{
			Name:      properties.NewNameVariant("misc-attributes"),
			FieldType: "internal",
			Type:      "[]xml.Attr",
			XmlType:   "[]xml.Attr",
			Tags:      "`xml:\",any,attr\"`",
		},
	}...)

	name := parentPrefix.WithSuffix(param.PangoNameVariant())
	entries = append([]entryStructContext{{
		Fields:  fields,
		name:    name,
		version: version,
	}}, entries...)

	return entries
}

func creasteStructSpecsForNormalization(structTyp structType, parentPrefix *properties.NameVariant, spec *properties.Normalization, version *version.Version) []entryStructContext {
	var entries []entryStructContext
	var fields []entryStructFieldContext

	if structTyp == structXmlType {
		var xmlTags string
		if spec.TerraformProviderConfig.XmlNode != nil {
			xmlTags = fmt.Sprintf("`xml:\"%s\"`", *spec.TerraformProviderConfig.XmlNode)
		} else {
			switch spec.TerraformProviderConfig.ResourceType {
			case properties.TerraformResourceEntry, properties.TerraformResourceUuid:
				xmlTags = "`xml:\"entry\"`"
			case properties.TerraformResourceConfig:
				xmlTags = "`xml:\"system\"`"
			case properties.TerraformResourceCustom:
				fallthrough
			default:
				panic(fmt.Sprintf("unreachable resource type: '%s'", spec.TerraformProviderConfig.ResourceType))
			}
		}

		fields = append(fields, entryStructFieldContext{
			IsInternal: true,
			FieldType:  "internal",
			Name:       xmlNameVariant,
			XmlType:    "xml.Name",
			Tags:       xmlTags,
		})
	}

	switch spec.TerraformProviderConfig.ResourceType {
	case properties.TerraformResourceEntry, properties.TerraformResourceUuid:
		fields = append(fields, entryStructFieldContext{
			Name:      properties.NewNameVariant("name"),
			Required:  true,
			FieldType: "simple",
			Type:      "string",
			XmlType:   "string",
			Tags:      "`xml:\"name,attr\"`",
		})
	case properties.TerraformResourceConfig:
	case properties.TerraformResourceCustom:
		fallthrough
	default:
		panic(fmt.Sprintf("unreachable resource type: '%s'", spec.TerraformProviderConfig.ResourceType))
	}

	processParameter := func(prefix *properties.NameVariant, param *properties.SpecParam) {
		if param.GoSdkConfig != nil && param.GoSdkConfig.Skip != nil && *param.GoSdkConfig.Skip {
			return
		}

		if !ParamSupportedInVersion(param, version) {
			return
		}

		var overrideTypeForXmlContainer bool
		if structTyp == structXmlType && param.Type == "list" && param.Items.Type == "entry" {
			overrideTypeForXmlContainer = true
			entries = append(entries, createStructSpecForXmlListContainer(prefix, param, version)...)
		}

		typ, itemsType, _ := getTypesForParam(structApiType, prefix, param, version, overrideTypeForXmlContainer)
		xmlType, itemsXmlType, xmlContainerType := getTypesForParam(structXmlType, prefix, param, version, overrideTypeForXmlContainer)
		fieldType := getFieldTypeForParam(param)

		fields = append(fields, entryStructFieldContext{
			Name:             param.PangoNameVariant(),
			Required:         param.Required,
			FieldType:        fieldType,
			Type:             typ,
			ItemsType:        itemsType,
			XmlType:          xmlType,
			XmlContainerType: xmlContainerType,
			ItemsXmlType:     itemsXmlType,
			Tags:             XmlTag(param),
			version:          version,
		})

		if param.Type == "" || (param.Type == "list" && param.Items.Type == "entry") {
			entries = append(entries, createEntryXmlStructSpecsForParameter(structTyp, properties.NewNameVariant(""), param, version)...)
		}
	}

	for _, elt := range spec.Spec.SortedParams() {
		processParameter(parentPrefix, elt)
	}

	for _, elt := range spec.Spec.SortedOneOf() {
		processParameter(parentPrefix, elt)
	}

	fields = append(fields, []entryStructFieldContext{
		{
			Name:      properties.NewNameVariant("misc"),
			FieldType: "internal",
			Type:      "[]generic.Xml",
			XmlType:   "[]generic.Xml",
			Tags:      "`xml:\",any\"`",
		},
		{
			Name:      properties.NewNameVariant("misc-attributes"),
			FieldType: "internal",
			Type:      "[]xml.Attr",
			XmlType:   "[]xml.Attr",
			Tags:      "`xml:\",any,attr\"`",
		},
	}...)

	var name *properties.NameVariant
	switch spec.TerraformProviderConfig.ResourceType {
	case properties.TerraformResourceEntry, properties.TerraformResourceUuid:
		name = properties.NewNameVariant("entry")
	case properties.TerraformResourceConfig:
		name = properties.NewNameVariant("config")
	case properties.TerraformResourceCustom:
		fallthrough
	default:
		panic(fmt.Sprintf("unreachable resource type: %v", spec.TerraformProviderConfig.ResourceType))
	}

	entries = append([]entryStructContext{{
		TopLevel: true,
		Fields:   fields,
		name:     name,
		version:  version,
	}}, entries...)

	return entries
}

func createStructSpecs(structTyp structType, spec *properties.Normalization, version *version.Version) []entryStructContext {
	return creasteStructSpecsForNormalization(structTyp, properties.NewNameVariant(""), spec, version)
}

// RenderEntryApiStructs generates API struct definitions for a normalization spec.
func RenderEntryApiStructs(spec *properties.Normalization) (string, error) {
	tmplContent, err := loadTemplate("partials/api_structs.tmpl")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("render-entry-api-structs").Parse(tmplContent)
	if err != nil {
		return "", err
	}

	specs := createStructSpecs(structApiType, spec, nil)
	type context struct {
		Specs []entryStructContext
	}

	data := context{Specs: specs}

	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", err
	}

	return builder.String(), nil
}

// RenderEntryXmlStructs generates XML struct definitions for a normalization spec.
func RenderEntryXmlStructs(spec *properties.Normalization) (string, error) {
	tmplContent, err := loadTemplate("partials/xml_structs.tmpl")
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("render-entry-xml-structs").Parse(tmplContent)
	if err != nil {
		return "", err
	}

	specs := createStructSpecs(structXmlType, spec, nil)
	for _, elt := range spec.SupportedVersionRanges() {
		specs = append(specs, createStructSpecs(structXmlType, spec, &elt.Minimum)...)
	}

	type context struct {
		Specs []entryStructContext
	}

	data := context{Specs: specs}

	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", err
	}

	return builder.String(), nil
}
