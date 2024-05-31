package terraform_provider

import (
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"strings"
)

func ResourceCreateFunction(structName string, serviceName string) (string, error) {
	if strings.Contains(serviceName, "group") && serviceName != "Device group" {
		serviceName = "group"
	}
	data := map[string]interface{}{
		"structName":  structName,
		"serviceName": naming.CamelCase("", serviceName, "", false),
	}
	return centralTemplateExec(resourceCreateTemplateStr, "resource-create-function", data, nil)
}

func ResourceReadFunction(structName string, serviceName string) (string, error) {
	if strings.Contains(serviceName, "group") {
		serviceName = "group"
	}
	data := map[string]interface{}{
		"structName":  structName,
		"serviceName": naming.CamelCase("", serviceName, "", false),
	}
	return centralTemplateExec(resourceReadTemplateStr, "resource-read-function", data, nil)
}

func ResourceUpdateFunction(structName string, serviceName string) (string, error) {
	if strings.Contains(serviceName, "group") {
		serviceName = "group"
	}
	data := map[string]interface{}{
		"structName":  structName,
		"serviceName": naming.CamelCase("", serviceName, "", false),
	}
	return centralTemplateExec(resourceUpdateTemplateStr, "resource-update-function", data, nil)
}

func ResourceDeleteFunction(structName string, serviceName string) (string, error) {
	if strings.Contains(serviceName, "group") {
		serviceName = "group"
	}
	data := map[string]interface{}{
		"structName":  structName,
		"serviceName": naming.CamelCase("", serviceName, "", false),
	}
	return centralTemplateExec(resourceDeleteTemplateStr, "resource-delete-function", data, nil)
}
