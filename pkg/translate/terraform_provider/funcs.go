package terraform_provider

import (
	"fmt"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"reflect"
	"strings"
	"text/template"
)

type Entry struct {
	Name string
	Type string
}

type EntryData struct {
	EntryName string
	Entries   []Entry
}

func ResourceCreateFunction(structName string, serviceName string, paramSpec interface{}, terraformProvider *properties.TerraformProviderFile, resourceSDKName string) (string, error) {
	funcMap := template.FuncMap{
		"LoadConfigToEntry": CreateEntryConfig,
		"ResourceParamToSchema": func(paramName string, paramParameters properties.SpecParam) (string, error) {
			return ParamToSchemaResource(paramName, paramParameters, terraformProvider)
		},
	}
	if strings.Contains(serviceName, "group") && serviceName != "Device group" {
		serviceName = "group"
	}
	data := map[string]interface{}{
		"structName":      structName,
		"serviceName":     naming.CamelCase("", serviceName, "", false),
		"paramSpec":       paramSpec,
		"resourceSDKName": resourceSDKName,
	}
	return processTemplate(resourceCreateTemplateStr, "resource-create-function", data, funcMap)
}

func ResourceReadFunction(structName string, serviceName string, paramSpec interface{}, resourceSDKName string) (string, error) {
	if strings.Contains(serviceName, "group") {
		serviceName = "group"
	}
	data := map[string]interface{}{
		"structName":      structName,
		"serviceName":     naming.CamelCase("", serviceName, "", false),
		"resourceSDKName": resourceSDKName,
	}
	return processTemplate(resourceReadTemplateStr, "resource-read-function", data, nil)
}

func ResourceUpdateFunction(structName string, serviceName string, paramSpec interface{}, resourceSDKName string) (string, error) {
	if strings.Contains(serviceName, "group") {
		serviceName = "group"
	}
	data := map[string]interface{}{
		"structName":      structName,
		"serviceName":     naming.CamelCase("", serviceName, "", false),
		"resourceSDKName": resourceSDKName,
	}
	return processTemplate(resourceUpdateTemplateStr, "resource-update-function", data, nil)
}

func ResourceDeleteFunction(structName string, serviceName string, paramSpec interface{}, resourceSDKName string) (string, error) {
	if strings.Contains(serviceName, "group") {
		serviceName = "group"
	}
	data := map[string]interface{}{
		"structName":      structName,
		"serviceName":     naming.CamelCase("", serviceName, "", false),
		"resourceSDKName": resourceSDKName,
	}
	return processTemplate(resourceDeleteTemplateStr, "resource-delete-function", data, nil)
}

func CreateEntryConfig(spec interface{}, entryName string) (string, error) {
	specValReflect := reflect.ValueOf(spec)
	specValIndirect := reflect.Indirect(specValReflect)
	t := specValIndirect.Type()

	var entries []Entry
	for i := 0; i < t.NumField(); i++ {
		entry := t.Field(i)
		if entry.Name == "Type" {
			valueRaw := reflect.Indirect(specValReflect).Field(i).Interface()
			valueFormatted := fmt.Sprintf("%v", valueRaw)
			entries = append(entries, Entry{
				Name: naming.CamelCase("", entryName, "", true),
				Type: fmt.Sprintf("%v", valueFormatted),
			})
		}
	}

	entryData := EntryData{
		EntryName: entryName,
		Entries:   entries,
	}

	return processTemplate(resourceEntryConfigTemplate, "config-to-entry", entryData, nil)
}
