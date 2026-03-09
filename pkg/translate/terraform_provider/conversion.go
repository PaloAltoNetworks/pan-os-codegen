package terraform_provider

import (
	"fmt"
	"log"
	"runtime/debug"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/object"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/parameter"
)

// parameterEncryptionSpec holds encryption configuration for a parameter.
type parameterEncryptionSpec struct {
	HashingType parameter.HashingType
	HashingFunc string
}

// parameterSpec describes a parameter used in conversion templates.
type parameterSpec struct {
	PangoName     *properties.NameVariant
	TerraformName *properties.NameVariant
	TerraformType string
	ComplexType   string
	Type          string
	Required      bool
	ItemsType     string
	Encryption    *parameterEncryptionSpec
}

// spec describes a struct specification used in conversion templates.
type spec struct {
	Name                   string
	HasEntryName           bool
	HasEncryptedParameters bool
	PangoType              string
	PangoReturnType        string
	TerraformType          string
	TerraformStructType    string
	ModelOrObject          string
	Params                 []parameterSpec
	OneOf                  []parameterSpec
}

// renderSpecsForParams creates parameter specifications for conversion templates.
func renderSpecsForParams(names *NameProvider, schemaTyp properties.SchemaType, ancestors []*properties.SpecParam, params []*properties.SpecParam) []parameterSpec {
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

		var structPrefix string
		switch schemaTyp {
		case properties.SchemaDataSource:
			structPrefix = names.DataSourceStructName
		case properties.SchemaResource, properties.SchemaEphemeralResource:
			structPrefix = names.ResourceStructName
		case properties.SchemaListResource, properties.SchemaCommon, properties.SchemaProvider, properties.SchemaAction, properties.SchemaCustom:
			panic(fmt.Sprintf("invalid schema type: %s", schemaTyp))
		}

		for _, elt := range ancestors {
			structPrefix += elt.TerraformNameVariant().CamelCase
		}

		specs = append(specs, parameterSpec{
			PangoName:     elt.PangoNameVariant(),
			TerraformName: elt.TerraformNameVariant(),
			TerraformType: terraformTypeForProperty(structPrefix, elt, false),
			ComplexType:   elt.ComplexType(),
			Type:          elt.FinalType(),
			ItemsType:     itemsType,
			Encryption:    encryptionSpec,
		})

	}
	return specs
}

// generateFromTerraformToPangoSpec creates nested object conversion specs.
func generateFromTerraformToPangoSpec(names *NameProvider, resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, pangoTypePrefix string, terraformPrefix string, paramSpec *properties.SpecParam, ancestors []*properties.SpecParam) []spec {
	if paramSpec.Spec == nil {
		return nil
	}

	var specs []spec

	pangoType := fmt.Sprintf("%s%s", pangoTypePrefix, paramSpec.PangoNameVariant().CamelCase)

	pangoReturnType := fmt.Sprintf("%s%s", pangoTypePrefix, paramSpec.PangoNameVariant().CamelCase)
	terraformType := fmt.Sprintf("%s%s", terraformPrefix, paramSpec.TerraformNameVariant().CamelCase)

	ancestors = append(ancestors, paramSpec)

	paramSpecs := renderSpecsForParams(names, schemaTyp, ancestors, paramSpec.Spec.SortedParams())
	oneofSpecs := renderSpecsForParams(names, schemaTyp, ancestors, paramSpec.Spec.SortedOneOf())

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
			specs = append(specs, generateFromTerraformToPangoSpec(names, resourceTyp, schemaTyp, pangoType, terraformPrefix, elt, ancestors)...)
		}
	}

	renderSpecsForParams(paramSpec.Spec.SortedParams())
	renderSpecsForParams(paramSpec.Spec.SortedOneOf())

	return specs
}

// generateFromTerraformToPangoParameter creates top-level conversion specs.
func generateFromTerraformToPangoParameter(names *NameProvider, resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, pkgName string, terraformPrefix string, pangoPrefix string, prop *properties.Normalization, ancestors []*properties.SpecParam) []spec {
	var specs []spec

	var pangoReturnType string
	if ancestors == nil {
		pangoReturnType = fmt.Sprintf("%s.%s", pkgName, prop.EntryOrConfig())
		pangoPrefix = fmt.Sprintf("%s.", pkgName)
	} else {
		pangoReturnType = fmt.Sprintf("%s.%s", pkgName, ancestors[0].Name.CamelCase)
	}

	switch resourceTyp {
	case properties.ResourceEntry, properties.ResourceConfig:
	case properties.ResourceEntryPlural, properties.ResourceUuid, properties.ResourceUuidPlural:
		terraformPrefix = fmt.Sprintf("%s%s", terraformPrefix, pascalCase(prop.TerraformProviderConfig.PluralName))
	case properties.ResourceCustom:
		panic("custom resources don't generate anything")
	}

	switch resourceTyp {
	case properties.ResourceEntry, properties.ResourceConfig:
		paramSpecs := renderSpecsForParams(names, schemaTyp, ancestors, prop.Spec.SortedParams())
		oneofSpecs := renderSpecsForParams(names, schemaTyp, ancestors, prop.Spec.SortedOneOf())
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
		ancestors = append(ancestors, &properties.SpecParam{
			Name: properties.NewNameVariant(prop.TerraformProviderConfig.PluralName),
			Type: "list",
		})

		paramSpecs := renderSpecsForParams(names, schemaTyp, ancestors, prop.Spec.SortedParams())
		oneofSpecs := renderSpecsForParams(names, schemaTyp, ancestors, prop.Spec.SortedOneOf())

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

		specs = append(specs, generateFromTerraformToPangoSpec(names, resourceTyp, schemaTyp, pangoPrefix, terraformPrefix, elt, ancestors)...)
	}

	for _, elt := range prop.Spec.SortedOneOf() {
		if elt.IsPrivateParameter() {
			continue
		}

		specs = append(specs, generateFromTerraformToPangoSpec(names, resourceTyp, schemaTyp, pangoPrefix, terraformPrefix, elt, ancestors)...)
	}

	return specs
}

// RenderCopyToPangoFunctions generates functions to convert Terraform state to Pango types.
func RenderCopyToPangoFunctions(names *NameProvider, resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, pkgName string, terraformTypePrefix string, property *properties.Normalization) (string, error) {
	if resourceTyp == properties.ResourceCustom {
		return "", nil
	}

	specs := generateFromTerraformToPangoParameter(names, resourceTyp, schemaTyp, pkgName, terraformTypePrefix, "", property, nil)

	type context struct {
		Specs []spec
	}

	data := context{
		Specs: specs,
	}
	funcMap := mergeFuncMaps(commonFuncMap, template.FuncMap{
		"PascalCase": pascalCase,
	})
	return processTemplate("conversion/copy_to_pango.tmpl", "copy-to-pango", data, funcMap)
}

// RenderCopyFromPangoFunctions generates functions to convert Pango types to Terraform state.
func RenderCopyFromPangoFunctions(names *NameProvider, resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, pkgName string, terraformTypePrefix string, property *properties.Normalization) (string, error) {
	if resourceTyp == properties.ResourceCustom {
		return "", nil
	}

	specs := generateFromTerraformToPangoParameter(names, resourceTyp, schemaTyp, pkgName, terraformTypePrefix, "", property, nil)

	type context struct {
		Specs []spec
	}

	data := context{
		Specs: specs,
	}

	funcMap := mergeFuncMaps(commonFuncMap, template.FuncMap{
		"PascalCase": pascalCase,
	})
	return processTemplate("conversion/copy_from_pango.tmpl", "copy-from-pango", data, funcMap)
}

// RenderXpathComponentsGetter generates a function to extract XPath components from a struct.
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

	return processTemplate("conversion/xpath_components.tmpl", "xpath-components", data, commonFuncMap)
}
