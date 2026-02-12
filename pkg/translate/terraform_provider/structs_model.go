package terraform_provider

import (
	"fmt"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/imports"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/object"
)

// datasourceStructFieldSpec describes a field in a data source or resource struct.
type datasourceStructFieldSpec struct {
	Name                *properties.NameVariant
	Private             bool
	TerraformType       string
	TerraformStructType string
	Type                string
	ItemsType           string
	Tags                []string
}

// datasourceStructSpec describes a data source or resource struct.
type datasourceStructSpec struct {
	StructName          string
	AncestorName        string
	TerraformPluralType object.TerraformPluralType
	HasEntryName        bool
	ModelOrObject       string
	Fields              []datasourceStructFieldSpec
}

// modelFieldValidationType describes the type of validation for a model field.
type modelFieldValidationType string

const (
	modelFieldValidationPlaintextPlaceholder = "plaintext-placeholder"
)

// nestedObjectField describes a nested object field that requires validation.
type nestedObjectField struct {
	FieldName    *properties.NameVariant
	FieldType    string // "object", "list", "set"
	NestedStruct string // Name of nested struct (e.g., "ResourceNameSettings")
}

// modelFieldValidatorSpec describes validation for a single model field.
type modelFieldValidatorSpec struct {
	FieldName      *properties.NameVariant
	ValidationType modelFieldValidationType
}

// modelValidatorSpec describes validation for an entire model.
type modelValidatorSpec struct {
	StructName    string
	ModelOrObject string
	Fields        []modelFieldValidatorSpec
	NestedObjects []nestedObjectField
}

// terraformTypeForProperty returns the Terraform type for a property.
func terraformTypeForProperty(structPrefix string, prop *properties.SpecParam, hackStructsAsTypeObjects bool) string {
	if prop.Type == "" || prop.Type == "list" && prop.Items.Type == "entry" {
		if hackStructsAsTypeObjects {
			if prop.Type == "" {
				return "types.Object"
			} else if prop.FinalType() == "set" {
				return "types.Set"
			} else if prop.FinalType() == "map" {
				return "types.Map"
			} else {
				return "types.List"
			}
		} else {
			return fmt.Sprintf("*%s%sObject", structPrefix, prop.TerraformNameVariant().CamelCase)
		}
	}

	switch prop.ComplexType() {
	case "string-as-member":
		return "types.String"
	}

	if prop.FinalType() == "list" {
		return "types.List"
	}

	if prop.FinalType() == "set" {
		return "types.Set"
	}

	return fmt.Sprintf("types.%s", pascalCase(prop.Type))
}

// structFieldSpec creates a field specification for a parameter.
func structFieldSpec(param *properties.SpecParam, structPrefix string, hackStructsAsTypeObjects bool) datasourceStructFieldSpec {
	tfTag := fmt.Sprintf("`tfsdk:\"%s\"`", param.TerraformNameVariant().Underscore)

	var itemsType string
	if param.Type == "list" {
		if param.Items.Type == "entry" {
			itemsType = "types.Object"
		} else {
			itemsType = fmt.Sprintf("types.%sType", pascalCase(param.Items.Type))
		}
	}

	return datasourceStructFieldSpec{
		Name:                param.TerraformNameVariant(),
		TerraformType:       terraformTypeForProperty(structPrefix, param, hackStructsAsTypeObjects),
		TerraformStructType: terraformTypeForProperty(structPrefix, param, false),
		Type:                terraformTypeForProperty(structPrefix, param, hackStructsAsTypeObjects),
		ItemsType:           itemsType,
		Tags:                []string{tfTag},
	}
}

// dataSourceStructContextForParam creates struct specifications for a nested parameter.
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

// createStructSpecForUuidModel creates struct specifications for UUID-type resources.
func createStructSpecForUuidModel(resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, spec *properties.Normalization, names *NameProvider, hackStructsAsTypeObjects bool) []datasourceStructSpec {
	var structs []datasourceStructSpec

	var fields []datasourceStructFieldSpec

	if len(spec.Locations) > 0 {
		fields = append(fields, datasourceStructFieldSpec{
			Name:                properties.NewNameVariant("location"),
			TerraformType:       fmt.Sprintf("%sLocation", names.StructName),
			TerraformStructType: fmt.Sprintf("%sLocation", names.StructName),
			Type:                "types.Object",
			Tags:                []string{"`tfsdk:\"location\"`"},
		})
	}

	if resourceTyp == properties.ResourceUuidPlural {

		position := properties.NewNameVariant("position")

		fields = append(fields, datasourceStructFieldSpec{
			Name:                position,
			TerraformType:       "TerraformPositionObject",
			TerraformStructType: "TerraformPositionObject",
			Type:                "types.Object",
			Tags:                []string{"`tfsdk:\"position\"`"},
		})
	}

	var structName string
	switch schemaTyp {
	case properties.SchemaResource, properties.SchemaEphemeralResource:
		structName = names.ResourceStructName
	case properties.SchemaDataSource:
		structName = names.DataSourceStructName
	case properties.SchemaCommon, properties.SchemaProvider, properties.SchemaAction:
		panic("unreachable")
	}

	listNameStr := spec.TerraformProviderConfig.PluralName
	listName := properties.NewNameVariant(listNameStr)

	tag := fmt.Sprintf("`tfsdk:\"%s\"`", listName.Underscore)
	fields = append(fields, datasourceStructFieldSpec{
		Name:                listName,
		Type:                "types.List",
		TerraformStructType: fmt.Sprintf("%s%sObject", structName, listName.CamelCase),
		ItemsType:           "types.Object",
		Tags:                []string{tag},
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

// createStructSpecForEntryListModel creates struct specifications for entry-type plural resources.
func createStructSpecForEntryListModel(resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, spec *properties.Normalization, names *NameProvider, hackStructsAsTypeObjects bool) []datasourceStructSpec {
	var structs []datasourceStructSpec

	var fields []datasourceStructFieldSpec
	if len(spec.Locations) > 0 {
		fields = append(fields, datasourceStructFieldSpec{
			Name:                properties.NewNameVariant("location"),
			TerraformType:       fmt.Sprintf("%sLocation", names.StructName),
			TerraformStructType: fmt.Sprintf("%sLocation", names.StructName),
			Type:                "types.Object",
			Tags:                []string{"`tfsdk:\"location\"`"},
		})
	}

	var structName string
	switch schemaTyp {
	case properties.SchemaResource, properties.SchemaEphemeralResource:
		structName = names.ResourceStructName
	case properties.SchemaDataSource:
		structName = names.DataSourceStructName
	case properties.SchemaCommon, properties.SchemaProvider, properties.SchemaAction:
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
		Name:                listName,
		Type:                listEltType,
		TerraformStructType: fmt.Sprintf("%s%sObject", structName, listName.CamelCase),
		ItemsType:           "types.Object",
		Tags:                []string{tag},
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

// createStructSpecForEntryModel creates struct specifications for entry-type singular resources.
func createStructSpecForEntryModel(resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, spec *properties.Normalization, names *NameProvider, hackStructAsTypeObjects bool) []datasourceStructSpec {
	var structs []datasourceStructSpec

	var fields []datasourceStructFieldSpec

	if len(spec.Locations) > 0 {
		fields = append(fields, datasourceStructFieldSpec{
			Name:                properties.NewNameVariant("location"),
			TerraformType:       fmt.Sprintf("%sLocation", names.StructName),
			TerraformStructType: fmt.Sprintf("%sLocation", names.StructName),
			Type:                "types.Object",
			Tags:                []string{"`tfsdk:\"location\"`"},
		})
	}

	var structName string
	switch schemaTyp {
	case properties.SchemaDataSource:
		structName = names.DataSourceStructName
	case properties.SchemaResource, properties.SchemaEphemeralResource:
		structName = names.ResourceStructName
	case properties.SchemaAction:
		structName = names.ActionStructName()
	case properties.SchemaCommon, properties.SchemaProvider:
		panic("unreachable")
	default:
		panic(fmt.Sprintf("unsupported schemaTyp: '%s'", schemaTyp))
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

// createStructSpecForModel creates struct specifications for any resource type.
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

// createStructSpecForNormalization creates struct fields and nested structs for a normalization.
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

// createValidatorSpecForParameter creates validator specs for a nested parameter.
func createValidatorSpecForParameter(resourceTyp properties.ResourceType, structPrefix string, param *properties.SpecParam, manager *imports.Manager) []modelValidatorSpec {
	if param.Spec == nil {
		return nil
	}

	var specs []modelValidatorSpec
	structName := fmt.Sprintf("%s%s", structPrefix, param.TerraformNameVariant().CamelCase)

	// Get validators for this parameter's spec
	fields, nestedSpecs, nestedObjects := createValidatorSpecForParameterSpec(resourceTyp, structName, param, manager)

	// Always add the spec to ensure ALL object structs have ValidateConfig methods
	// This is important because parent objects may call ValidateConfig on nested objects
	specs = append(specs, modelValidatorSpec{
		StructName:    structName,
		ModelOrObject: "Object", // Nested structures are always "Object"
		Fields:        fields,
		NestedObjects: nestedObjects,
	})

	// Add any nested specs
	specs = append(specs, nestedSpecs...)

	return specs
}

// createValidatorSpecForParameterSpec processes the Spec of a parameter.
func createValidatorSpecForParameterSpec(resourceTyp properties.ResourceType, structName string, param *properties.SpecParam, manager *imports.Manager) ([]modelFieldValidatorSpec, []modelValidatorSpec, []nestedObjectField) {
	var fields []modelFieldValidatorSpec
	var nestedSpecs []modelValidatorSpec
	var nestedObjects []nestedObjectField

	if param.Spec == nil {
		return fields, nestedSpecs, nestedObjects
	}

	// Process regular parameters
	for _, elt := range param.Spec.SortedParams() {
		if elt.IsPrivateParameter() {
			continue
		}

		if elt.Hashing != nil {
			fields = append(fields, modelFieldValidatorSpec{
				FieldName:      elt.TerraformNameVariant(),
				ValidationType: modelFieldValidationPlaintextPlaceholder,
			})
			manager.AddStandardImport("strings", "")
		}

		if elt.Type == "" {
			nestedObjects = append(nestedObjects, nestedObjectField{
				FieldName:    elt.TerraformNameVariant(),
				FieldType:    "object",
				NestedStruct: fmt.Sprintf("%s%s", structName, elt.TerraformNameVariant().CamelCase),
			})
			nestedSpecs = append(nestedSpecs, createValidatorSpecForParameter(resourceTyp, structName, elt, manager)...)
		} else if (elt.FinalType() == "list" || elt.FinalType() == "set") && elt.Items != nil && elt.Items.Type == "entry" {
			nestedObjects = append(nestedObjects, nestedObjectField{
				FieldName:    elt.TerraformNameVariant(),
				FieldType:    elt.FinalType(),
				NestedStruct: fmt.Sprintf("%s%s", structName, elt.TerraformNameVariant().CamelCase),
			})
			nestedSpecs = append(nestedSpecs, createValidatorSpecForParameter(resourceTyp, structName, elt, manager)...)
		}
	}

	// Process variant parameters
	for _, elt := range param.Spec.SortedOneOf() {
		if elt.IsPrivateParameter() {
			continue
		}

		if elt.Hashing != nil {
			fields = append(fields, modelFieldValidatorSpec{
				FieldName:      elt.TerraformNameVariant(),
				ValidationType: modelFieldValidationPlaintextPlaceholder,
			})
			manager.AddStandardImport("strings", "")
		}

		if elt.Type == "" {
			nestedObjects = append(nestedObjects, nestedObjectField{
				FieldName:    elt.TerraformNameVariant(),
				FieldType:    "object",
				NestedStruct: fmt.Sprintf("%s%s", structName, elt.TerraformNameVariant().CamelCase),
			})
			nestedSpecs = append(nestedSpecs, createValidatorSpecForParameter(resourceTyp, structName, elt, manager)...)
		} else if (elt.FinalType() == "list" || elt.FinalType() == "set") && elt.Items != nil && elt.Items.Type == "entry" {
			nestedObjects = append(nestedObjects, nestedObjectField{
				FieldName:    elt.TerraformNameVariant(),
				FieldType:    elt.FinalType(),
				NestedStruct: fmt.Sprintf("%s%s", structName, elt.TerraformNameVariant().CamelCase),
			})
			nestedSpecs = append(nestedSpecs, createValidatorSpecForParameter(resourceTyp, structName, elt, manager)...)
		}
	}

	return fields, nestedSpecs, nestedObjects
}

// createValidatorSpecForNormalization creates validator specs for a normalization's parameters.
func createValidatorSpecForNormalization(resourceTyp properties.ResourceType, structName string, spec *properties.Normalization, manager *imports.Manager) ([]modelFieldValidatorSpec, []modelValidatorSpec, []nestedObjectField) {
	var fields []modelFieldValidatorSpec
	var nestedSpecs []modelValidatorSpec
	var nestedObjects []nestedObjectField

	// Process regular parameters
	for _, elt := range spec.Spec.SortedParams() {
		if elt.IsPrivateParameter() {
			continue
		}

		// Skip XPath variables for entry plural resources (they're handled separately)
		if resourceTyp == properties.ResourceEntryPlural && elt.TerraformProviderConfig != nil && elt.TerraformProviderConfig.XpathVariable != nil {
			continue
		}

		if elt.Hashing != nil {
			fields = append(fields, modelFieldValidatorSpec{
				FieldName:      elt.TerraformNameVariant(),
				ValidationType: modelFieldValidationPlaintextPlaceholder,
			})
			manager.AddStandardImport("strings", "")
		}

		if elt.Type == "" {
			nestedObjects = append(nestedObjects, nestedObjectField{
				FieldName:    elt.TerraformNameVariant(),
				FieldType:    "object",
				NestedStruct: fmt.Sprintf("%s%s", structName, elt.TerraformNameVariant().CamelCase),
			})
			nestedSpecs = append(nestedSpecs, createValidatorSpecForParameter(resourceTyp, structName, elt, manager)...)
		} else if (elt.FinalType() == "list" || elt.FinalType() == "set") && elt.Items != nil && elt.Items.Type == "entry" {
			nestedObjects = append(nestedObjects, nestedObjectField{
				FieldName:    elt.TerraformNameVariant(),
				FieldType:    elt.FinalType(),
				NestedStruct: fmt.Sprintf("%s%s", structName, elt.TerraformNameVariant().CamelCase),
			})
			nestedSpecs = append(nestedSpecs, createValidatorSpecForParameter(resourceTyp, structName, elt, manager)...)
		}
	}

	// Process variant parameters
	for _, elt := range spec.Spec.SortedOneOf() {
		if elt.IsPrivateParameter() {
			continue
		}

		// Skip XPath variables for entry plural resources
		if resourceTyp == properties.ResourceEntryPlural && elt.TerraformProviderConfig != nil && elt.TerraformProviderConfig.XpathVariable != nil {
			continue
		}

		if elt.Hashing != nil {
			fields = append(fields, modelFieldValidatorSpec{
				FieldName:      elt.TerraformNameVariant(),
				ValidationType: modelFieldValidationPlaintextPlaceholder,
			})
			manager.AddStandardImport("strings", "")
		}

		if elt.Type == "" {
			nestedObjects = append(nestedObjects, nestedObjectField{
				FieldName:    elt.TerraformNameVariant(),
				FieldType:    "object",
				NestedStruct: fmt.Sprintf("%s%s", structName, elt.TerraformNameVariant().CamelCase),
			})
			nestedSpecs = append(nestedSpecs, createValidatorSpecForParameter(resourceTyp, structName, elt, manager)...)
		} else if (elt.FinalType() == "list" || elt.FinalType() == "set") && elt.Items != nil && elt.Items.Type == "entry" {
			nestedObjects = append(nestedObjects, nestedObjectField{
				FieldName:    elt.TerraformNameVariant(),
				FieldType:    elt.FinalType(),
				NestedStruct: fmt.Sprintf("%s%s", structName, elt.TerraformNameVariant().CamelCase),
			})
			nestedSpecs = append(nestedSpecs, createValidatorSpecForParameter(resourceTyp, structName, elt, manager)...)
		}
	}

	return fields, nestedSpecs, nestedObjects
}

// createValidatorSpecForEntryModel creates validator specs for entry-type singular resources.
func createValidatorSpecForEntryModel(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization, manager *imports.Manager) []modelValidatorSpec {
	if spec.Spec == nil {
		return nil
	}

	var specs []modelValidatorSpec
	structName := names.ResourceStructName

	fields, nestedSpecs, nestedObjects := createValidatorSpecForNormalization(resourceTyp, structName, spec, manager)

	// Always add the spec to ensure ALL structs have ValidateConfig methods
	specs = append(specs, modelValidatorSpec{
		StructName:    structName,
		ModelOrObject: "Model", // Top-level is always Model
		Fields:        fields,
		NestedObjects: nestedObjects,
	})

	specs = append(specs, nestedSpecs...)

	return specs
}

// createValidatorSpecForEntryListModel creates validator specs for entry-type plural resources.
func createValidatorSpecForEntryListModel(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization, manager *imports.Manager) []modelValidatorSpec {
	if spec.Spec == nil {
		return nil
	}

	var specs []modelValidatorSpec

	// For entry list models, validators apply to the nested object representing each entry
	listNameStr := spec.TerraformProviderConfig.PluralName
	listName := properties.NewNameVariant(listNameStr)
	structName := fmt.Sprintf("%s%s", names.ResourceStructName, listName.CamelCase)

	// Determine the field type for the list/map/set
	var fieldType string
	switch spec.TerraformProviderConfig.PluralType {
	case object.TerraformPluralMapType:
		fieldType = "map"
	case object.TerraformPluralListType:
		fieldType = "list"
	case object.TerraformPluralSetType:
		fieldType = "set"
	}

	// Create NestedObjects entry for the Model to enable validation recursion into list/map/set
	nestedObjsForModel := []nestedObjectField{
		{
			FieldName:    listName,
			FieldType:    fieldType,
			NestedStruct: structName,
		},
	}

	// Add validator for top-level Model
	// The Model contains the list/map/set field, and we need to recurse into it
	specs = append(specs, modelValidatorSpec{
		StructName:    names.ResourceStructName,
		ModelOrObject: "Model",
		Fields:        nil,                // No direct fields to validate on the Model itself
		NestedObjects: nestedObjsForModel, // Recurse into list/map/set to validate entry objects
	})

	fields, nestedSpecs, nestedObjects := createValidatorSpecForNormalization(resourceTyp, structName, spec, manager)

	// Always add the spec to ensure ALL structs have ValidateConfig methods
	specs = append(specs, modelValidatorSpec{
		StructName:    structName,
		ModelOrObject: "Object", // List elements are Objects
		Fields:        fields,
		NestedObjects: nestedObjects,
	})

	specs = append(specs, nestedSpecs...)

	return specs
}

// createValidatorSpecForUuidModel creates validator specs for UUID-type resources.
func createValidatorSpecForUuidModel(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization, manager *imports.Manager) []modelValidatorSpec {
	if spec.Spec == nil {
		return nil
	}

	var specs []modelValidatorSpec

	// For UUID models, validators apply to the nested object representing each entry
	listNameStr := spec.TerraformProviderConfig.PluralName
	listName := properties.NewNameVariant(listNameStr)
	structName := fmt.Sprintf("%s%s", names.ResourceStructName, listName.CamelCase)

	// UUID models always use a list type
	fieldType := "list"

	// Create NestedObjects entry for the Model to enable validation recursion into list
	nestedObjsForModel := []nestedObjectField{
		{
			FieldName:    listName,
			FieldType:    fieldType,
			NestedStruct: structName,
		},
	}

	// Add validator for top-level Model
	// The Model contains the list field, and we need to recurse into it
	specs = append(specs, modelValidatorSpec{
		StructName:    names.ResourceStructName,
		ModelOrObject: "Model",
		Fields:        nil,                // No direct fields to validate on the Model itself
		NestedObjects: nestedObjsForModel, // Recurse into list to validate entry objects
	})

	fields, nestedSpecs, nestedObjects := createValidatorSpecForNormalization(resourceTyp, structName, spec, manager)

	// Always add the spec to ensure ALL structs have ValidateConfig methods
	specs = append(specs, modelValidatorSpec{
		StructName:    structName,
		ModelOrObject: "Object", // UUID list elements are Objects
		Fields:        fields,
		NestedObjects: nestedObjects,
	})

	specs = append(specs, nestedSpecs...)

	return specs
}

// createValidatorSpecForModel creates validator specs for any resource type.
func createValidatorSpecForModel(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization, manager *imports.Manager) []modelValidatorSpec {
	if spec.Spec == nil {
		return nil
	}

	switch resourceTyp {
	case properties.ResourceEntry, properties.ResourceCustom, properties.ResourceConfig:
		return createValidatorSpecForEntryModel(resourceTyp, names, spec, manager)
	case properties.ResourceEntryPlural:
		return createValidatorSpecForEntryListModel(resourceTyp, names, spec, manager)
	case properties.ResourceUuid, properties.ResourceUuidPlural:
		return createValidatorSpecForUuidModel(resourceTyp, names, spec, manager)
	default:
		panic("unreachable")
	}
}

// RenderResourceStructs generates resource struct definitions.
func RenderResourceStructs(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization) (string, error) {
	type context struct {
		Structs []datasourceStructSpec
	}

	data := context{
		Structs: createStructSpecForModel(resourceTyp, properties.SchemaResource, spec, names, true),
	}

	return processTemplate("datasource/datasource_structs.tmpl", "render-structs", data, commonFuncMap)
}

// RenderResourceValidators generates resource validator methods.
func RenderResourceValidators(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization, manager *imports.Manager) (string, error) {
	type context struct {
		Validators []modelValidatorSpec
	}

	// Pass imports manager so validators can register required Go packages (e.g. "strings" for plaintext validation)
	validators := createValidatorSpecForModel(resourceTyp, names, spec, manager)

	data := context{
		Validators: validators,
	}

	return processTemplate("conversion/model_validators.tmpl", "render-structs", data, commonFuncMap)
}

// RenderDataSourceStructs generates data source struct definitions.
func RenderDataSourceStructs(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization) (string, error) {
	type context struct {
		Structs []datasourceStructSpec
	}

	data := context{
		Structs: createStructSpecForModel(resourceTyp, properties.SchemaDataSource, spec, names, true),
	}

	return processTemplate("datasource/datasource_structs.tmpl", "render-structs", data, commonFuncMap)
}

// RenderStructs generates struct definitions for any schema type.
func RenderStructs(resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, names *NameProvider, spec *properties.Normalization) (string, error) {
	type context struct {
		Structs []datasourceStructSpec
	}

	data := context{
		Structs: createStructSpecForModel(resourceTyp, schemaTyp, spec, names, true),
	}

	return processTemplate("datasource/datasource_structs.tmpl", "render-structs", data, commonFuncMap)
}

// RenderModelAttributeTypesFunction generates model attribute type definitions.
func RenderModelAttributeTypesFunction(resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, names *NameProvider, spec *properties.Normalization) (string, error) {
	type context struct {
		Structs []datasourceStructSpec
	}

	data := context{
		Structs: createStructSpecForModel(resourceTyp, schemaTyp, spec, names, true),
	}

	return processTemplate("common/attribute_types.tmpl", "attribute-types", data, nil)
}
