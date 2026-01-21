package terraform_provider

import (
	"fmt"
	"log"
	"runtime/debug"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/imports"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/object"
)

// defaultCtx describes a default value in a schema.
type defaultCtx struct {
	Type  string
	Value string
}

// modifierCtx describes schema modifiers.
type modifierCtx struct {
	SchemaType string
	Modifiers  []string
}

// validatorFunctionCtx describes a validator function.
type validatorFunctionCtx struct {
	Type              string
	Function          string
	FunctionOverriden bool
	Expressions       []string
	Values            []string
}

// validatorCtx describes schema validators.
type validatorCtx struct {
	ListType  string
	Package   string
	Functions []validatorFunctionCtx
}

// attributeCtx describes a schema attribute.
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

// schemaCtx describes a complete schema.
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

// generateValidatorFnsMapForVariants creates a map of validator functions for variant parameters.
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

// createSchemaSpecForParameter creates schema specifications for a nested parameter.
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
	case properties.SchemaAction:
		required = param.FinalRequired()
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
		Optional:      param.FinalOptional(),
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

// createSchemaAttributeForParameter creates a schema attribute for a parameter.
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

	var computed, required, optional bool
	switch schemaTyp {
	case properties.SchemaDataSource:
		optional = true
		required = false
		computed = true
	case properties.SchemaAction:
		required = param.FinalRequired()
		optional = param.FinalOptional()
	case properties.SchemaResource, properties.SchemaEphemeralResource:
		optional = param.FinalOptional()
		computed = param.FinalComputed()
		required = param.FinalRequired()
	case properties.SchemaCommon, properties.SchemaProvider:
		panic(fmt.Sprintf("unreachable for schemaTyp '%s'", schemaTyp))
	}

	return attributeCtx{
		Package:     packageName,
		Name:        param.TerraformNameVariant(),
		SchemaType:  schemaType,
		ElementType: elementType,
		Description: param.Description,
		Required:    required,
		Optional:    optional,
		Sensitive:   param.FinalSensitive(),
		Default:     defaultValue,
		Computed:    computed,
		Validators:  validators,
	}
}

// createSchemaSpecForUuidModel creates a schema for uuid-type resources.
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

// createSchemaSpecForEntryListModel creates a schema for entry-type plural resources.
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
	case properties.SchemaAction:
		packageName = "schema"
	case properties.SchemaCommon, properties.SchemaProvider:
		fallthrough
	default:
		panic(fmt.Sprintf("unsupported schemaTyp: '%s'", schemaTyp))
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
	case properties.SchemaAction:
		structName = names.ActionStructName()
	case properties.SchemaCommon, properties.SchemaProvider:
		fallthrough
	default:
		panic(fmt.Sprintf("unsupported schemaTyp: '%s'", schemaTyp))
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

// createSchemaSpecForNormalization creates schema attributes and nested schemas for a normalization.
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

// RenderResourceSchema generates resource schema code.
func RenderResourceSchema(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization, manager *imports.Manager) (string, error) {
	type context struct {
		Schemas []schemaCtx
	}

	data := context{
		Schemas: createSchemaSpecForModel(resourceTyp, properties.SchemaResource, spec, manager),
	}

	return processTemplate("schema/schema.tmpl", "render-resource-schema", data, commonFuncMap)
}

// RenderDataSourceSchema generates data source schema code.
func RenderDataSourceSchema(resourceTyp properties.ResourceType, names *NameProvider, spec *properties.Normalization, manager *imports.Manager) (string, error) {
	type context struct {
		Schemas []schemaCtx
	}

	data := context{
		Schemas: createSchemaSpecForModel(resourceTyp, properties.SchemaDataSource, spec, manager),
	}

	return processTemplate("schema/schema.tmpl", "render-resource-schema", data, commonFuncMap)
}

// RenderSchema generates schema code for any schema type.
func RenderSchema(resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, names *NameProvider, spec *properties.Normalization, manager *imports.Manager) (string, error) {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("** PANIC: %v", e)
			debug.PrintStack()
			panic(e)
		}
	}()

	type context struct {
		Schemas []schemaCtx
	}

	data := context{
		Schemas: createSchemaSpecForModel(resourceTyp, schemaTyp, spec, manager),
	}

	return processTemplate("schema/schema.tmpl", "render-schema", data, commonFuncMap)
}
