package terraform_provider

import (
	"fmt"
	"os"
	"path/filepath"
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

// processInlineTemplate processes an inline template string and executes it with the given data.
// This is the core template execution function that handles parsing and rendering.
func processInlineTemplate(tmplContent, templateName string, data interface{}, funcMap template.FuncMap) (string, error) {
	if len(funcMap) == 0 {
		funcMap = commonFuncMap
	} else {
		funcMap = mergeFuncMaps(funcMap, commonFuncMap)
	}

	tmpl, err := template.New(templateName).Funcs(funcMap).Parse(tmplContent)
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", err
	}
	return builder.String(), nil
}

// processTemplate handles the creation and execution of templates.
// It loads template content from .tmpl files in the templates/terraform-provider/ directory.
// The templateText parameter can be:
// - A relative file path ending in .tmpl (e.g., "schema/schema.tmpl") - loads from file and calls processInlineTemplate
// - An inline template string - passes directly to processInlineTemplate (for spec-provided custom templates)
func processTemplate(templateText, templateName string, data interface{}, funcMap template.FuncMap) (string, error) {
	var tmplContent string

	// If it looks like a file path, try to load from file
	if strings.HasSuffix(templateText, ".tmpl") {
		// Try current directory first, then parent directories (for tests)
		templatePath := filepath.Join("templates", "terraform-provider", templateText)
		content, err := os.ReadFile(templatePath)
		if err != nil {
			// Try from parent directories (for when running tests)
			for i := 1; i <= 3; i++ {
				prefix := strings.Repeat("../", i)
				altPath := filepath.Join(prefix, "templates", "terraform-provider", templateText)
				content, err = os.ReadFile(altPath)
				if err == nil {
					templatePath = altPath
					break
				}
			}
			if err != nil {
				return "", fmt.Errorf("failed to read template %s: %w", templatePath, err)
			}
		}
		tmplContent = string(content)
	} else {
		// Use as inline template (for custom templates from specs)
		tmplContent = templateText
	}

	return processInlineTemplate(tmplContent, templateName, data, funcMap)
}
