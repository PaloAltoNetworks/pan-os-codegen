package terraform_provider

import (
	"reflect"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	codegentmpl "github.com/paloaltonetworks/pan-os-codegen/pkg/template"
)

// Package-level function map to avoid repetition in each function
var commonFuncMap = template.FuncMap{
	"Map":            codegentmpl.TemplateMap,
	"LowerCase":      func(value string) string { return strings.ToLower(value) },
	"CamelCaseName":  func(paramName string) string { return naming.CamelCase("", paramName, "", true) },
	"UnderscoreName": func(paramName string) string { return naming.Underscore("", paramName, "") },
	"CamelCaseType":  func(paramType string) string { return naming.CamelCase("", paramType, "", true) },
}

// mergeFuncMaps merges two template.FuncMap instances.
// In case of a key conflict, the second map's value will override the first one.
func mergeFuncMaps(map1, map2 template.FuncMap) template.FuncMap {
	mergedMap := make(template.FuncMap)

	for key, value := range map1 {
		mergedMap[key] = value
	}

	for key, value := range map2 {
		mergedMap[key] = value
	}

	return mergedMap
}

// structToMap converts a struct to a map[string]interface{}.
// It takes an item of any type and returns a map where the keys are the exported field names and the values are the field values.
func structToMap(item interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		f := t.Field(i)
		// Exported field
		if f.PkgPath == "" {
			out[f.Name] = v.Field(i).Interface()
		}
	}
	return out
}

// processTemplate handles the creation and execution of templates
func processTemplate(templateText, templateName string, data interface{}, funcMap template.FuncMap) (string, error) {
	if len(funcMap) == 0 {
		funcMap = commonFuncMap
	} else {
		funcMap = mergeFuncMaps(funcMap, commonFuncMap)
	}

	tmpl, err := template.New(templateName).Funcs(funcMap).Parse(templateText)
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", err
	}
	return builder.String(), nil
}
