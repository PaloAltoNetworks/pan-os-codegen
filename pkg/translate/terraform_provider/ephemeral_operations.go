package terraform_provider

import (
	"fmt"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

// ResourceOpenFunction generates the Open function for ephemeral resources.
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

// ResourceRenewFunction generates the Renew function for ephemeral resources.
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

// ResourceCloseFunction generates the Close function for ephemeral resources.
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

// FunctionSupported checks if a function is supported for the given spec.
func FunctionSupported(spec *properties.Normalization, function string) (bool, error) {
	if len(spec.TerraformProviderConfig.CustomFuncs) > 0 {
		supported, found := spec.TerraformProviderConfig.CustomFuncs[function]
		return (found && supported), nil
	}

	switch function {
	case "Create", "Delete", "Read", "Update":
		return !spec.TerraformProviderConfig.Ephemeral, nil
	case "Open", "Close", "Renew":
		return spec.TerraformProviderConfig.Ephemeral, nil
	case "Invoke":
		return spec.TerraformProviderConfig.Action, nil
	default:
		return false, fmt.Errorf("invalid custom function name: %s", function)
	}
}
