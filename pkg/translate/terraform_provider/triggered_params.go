package terraform_provider

import (
	"fmt"
	"log"
	"runtime/debug"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

// triggeredParamsContext holds the context for rendering triggered parameter extraction code.
type triggeredParamsContext struct {
	PluralType                      string
	StructName                      string
	ResourceTFStructName            string
	ListAttribute                   *properties.NameVariant
	ParametersWithTriggerOnChangeOf []struct {
		Param   *properties.SpecParam
		Trigger *properties.SpecParam
	}
}

// RenderTriggeredParamsExtractCreate generates code to extract triggered parameters during create.
func RenderTriggeredParamsExtractCreate(resourceType properties.ResourceType, spec *properties.Normalization) (string, error) {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("** PANIC: %v", e)
			debug.PrintStack()
			panic(e)
		}
	}()

	// Select template based on resource type
	var templateName string
	switch resourceType {
	case properties.ResourceUuid, properties.ResourceUuidPlural:
		templateName = "triggered/extract_create_uuid.tmpl"
	case properties.ResourceEntryPlural:
		templateName = "triggered/extract_create_entry_list.tmpl"
	case properties.ResourceEntry, properties.ResourceConfig:
		templateName = "triggered/extract_create_singular.tmpl"
	default:
		return "", fmt.Errorf("unsupported resource type for triggered params: %s", resourceType)
	}

	data := buildTriggeredParamsContext(resourceType, spec)

	return processTemplate(templateName, "triggered-params-extract-create", data, nil)
}

// RenderTriggeredParamsExtractUpdate generates code to extract triggered parameters during update.
func RenderTriggeredParamsExtractUpdate(resourceType properties.ResourceType, spec *properties.Normalization) (string, error) {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("** PANIC: %v", e)
			debug.PrintStack()
			panic(e)
		}
	}()

	// Select template based on resource type
	var templateName string
	switch resourceType {
	case properties.ResourceUuid, properties.ResourceUuidPlural:
		templateName = "triggered/extract_update_uuid.tmpl"
	case properties.ResourceEntryPlural:
		templateName = "triggered/extract_update_entry_list.tmpl"
	case properties.ResourceEntry, properties.ResourceConfig:
		templateName = "triggered/extract_update_singular.tmpl"
	default:
		return "", fmt.Errorf("unsupported resource type for triggered params: %s", resourceType)
	}

	data := buildTriggeredParamsContext(resourceType, spec)

	return processTemplate(templateName, "triggered-params-extract-update", data, nil)
}

// buildTriggeredParamsContext creates the template context for triggered params templates.
func buildTriggeredParamsContext(resourceType properties.ResourceType, spec *properties.Normalization) triggeredParamsContext {
	names := NewNameProvider(spec, resourceType)

	var resourceTFStructName string
	var listAttribute *properties.NameVariant
	var pluralType string

	if resourceType == properties.ResourceUuid || resourceType == properties.ResourceUuidPlural ||
		resourceType == properties.ResourceEntryPlural {
		listAttribute = properties.NewNameVariant(spec.TerraformProviderConfig.PluralName)
		resourceTFStructName = names.ResourceStructName + listAttribute.CamelCase + "Object"
		if spec.TerraformProviderConfig.PluralType != "" {
			pluralType = string(spec.TerraformProviderConfig.PluralType)
		}
	}

	return triggeredParamsContext{
		PluralType:                      pluralType,
		StructName:                      names.ResourceStructName,
		ResourceTFStructName:            resourceTFStructName,
		ListAttribute:                   listAttribute,
		ParametersWithTriggerOnChangeOf: spec.GetParametersWithTriggerOnChangeOf(),
	}
}
