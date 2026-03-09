package terraform_provider

import (
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

// getCustomTemplateForFunction retrieves the custom template name for a function.
func getCustomTemplateForFunction(spec *properties.Normalization, function string) (string, error) {
	data := struct {
		Function string
	}{
		Function: function,
	}
	return processTemplate("common/custom_function.tmpl", "custom-template-for-function", data, nil)
}

// ResourceCreateFunction generates the Create function for a resource.
func ResourceCreateFunction(resourceTyp properties.ResourceType, names *NameProvider, serviceName string, paramSpec *properties.Normalization, terraformProvider *properties.TerraformProviderFile, resourceSDKName string) (string, error) {
	funcMap := template.FuncMap{
		"ConfigToEntry": ConfigEntry,
		"RenderEncryptedValuesInitialization": func() (string, error) {
			return RenderEncryptedValuesInitialization(properties.SchemaResource, paramSpec, "create")
		},
		"RenderEncryptedValuesFinalizer": func() (string, error) {
			return RenderEncryptedValuesFinalizer(properties.SchemaResource, paramSpec)
		},
		"RenderImportLocationAssignment": func(source string, dest string) (string, error) {
			return RenderImportLocationAssignment(names, paramSpec, source, dest)
		},
		"RenderCreateUpdateMovementRequired": func(state string, entries string) (string, error) {
			return RendeCreateUpdateMovementRequired(state, entries)
		},
		"RenderLocationsStateToPango": func(source string, dest string) (string, error) {
			return RenderLocationsStateToPango(names, paramSpec, source, dest, "resp.Diagnostics")
		},
	}

	if strings.Contains(serviceName, "group") && serviceName != "Device group" {
		serviceName = "group"
	}

	var tmpl string
	var listAttribute string
	var exhaustive bool
	switch resourceTyp {
	case properties.ResourceEntry, properties.ResourceConfig:
		exhaustive = true
		tmpl = "resource/create.tmpl"
	case properties.ResourceEntryPlural:
		exhaustive = false
		tmpl = "resource/create_entry_list.tmpl"
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
	case properties.ResourceUuid:
		exhaustive = true
		tmpl = "resource/create_many.tmpl"
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
	case properties.ResourceUuidPlural:
		exhaustive = false
		tmpl = "resource/create_many.tmpl"
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
	case properties.ResourceCustom:
		var err error
		tmpl, err = getCustomTemplateForFunction(paramSpec, "Create")
		if err != nil {
			return "", err
		}
	}

	listAttributeVariant := properties.NewNameVariant(listAttribute)

	data := map[string]interface{}{
		"PluralType":            paramSpec.TerraformProviderConfig.PluralType,
		"HasEncryptedResources": paramSpec.HasEncryptedResources(),
		"HasImports":            len(paramSpec.Imports.Variants) > 0,
		"Exhaustive":            exhaustive,
		"ListAttribute":         listAttributeVariant,
		"EntryOrConfig":         paramSpec.EntryOrConfig(),
		"HasEntryName":          paramSpec.HasEntryName(),
		"structName":            names.ResourceStructName,
		"serviceName":           naming.CamelCase("", serviceName, "", false),
		"paramSpec":             paramSpec.Spec,
		"resourceSDKName":       resourceSDKName,
		"locations":             paramSpec.OrderedLocations(),
	}

	return processTemplate(tmpl, "resource-create-function", data, funcMap)
}

// DataSourceReadFunction generates the Read function for a data source.
func DataSourceReadFunction(resourceTyp properties.ResourceType, names *NameProvider, serviceName string, paramSpec *properties.Normalization, resourceSDKName string) (string, error) {
	if strings.Contains(serviceName, "group") {
		serviceName = "group"
	}

	var tmpl string
	var listAttribute string
	var exhaustive bool
	switch resourceTyp {
	case properties.ResourceEntry, properties.ResourceConfig:
		tmpl = "resource/read.tmpl"
	case properties.ResourceEntryPlural:
		tmpl = "resource/read_entry_list.tmpl"
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
	case properties.ResourceUuid:
		tmpl = "resource/read_many.tmpl"
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
		exhaustive = true
	case properties.ResourceUuidPlural:
		tmpl = "resource/read_many.tmpl"
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
	case properties.ResourceCustom:
		var err error
		tmpl, err = getCustomTemplateForFunction(paramSpec, "Read")
		if err != nil {
			return "", err
		}
	}

	listAttributeVariant := properties.NewNameVariant(listAttribute)

	data := map[string]interface{}{
		"PluralType":                       paramSpec.TerraformProviderConfig.PluralType,
		"ResourceXpathVariablesWithChecks": paramSpec.ResourceXpathVariablesWithChecks(false),
		"ResourceOrDS":                     "DataSource",
		"HasEncryptedResources":            paramSpec.HasEncryptedResources(),
		"ListAttribute":                    listAttributeVariant,
		"Exhaustive":                       exhaustive,
		"EntryOrConfig":                    paramSpec.EntryOrConfig(),
		"HasEntryName":                     paramSpec.HasEntryName(),
		"structName":                       names.StructName,
		"resourceStructName":               names.ResourceStructName,
		"dataSourceStructName":             names.DataSourceStructName,
		"serviceName":                      naming.CamelCase("", serviceName, "", false),
		"resourceSDKName":                  resourceSDKName,
		"locations":                        paramSpec.OrderedLocations(),
	}

	funcMap := template.FuncMap{
		"RenderEncryptedValuesInitialization": func() (string, error) {
			return RenderEncryptedValuesInitialization(properties.SchemaDataSource, paramSpec, "read")
		},
		"RenderEncryptedValuesFinalizer": func() (string, error) {
			return RenderEncryptedValuesFinalizer(properties.SchemaDataSource, paramSpec)
		},
		"AttributesFromXpathComponents": func(target string) (string, error) { return paramSpec.AttributesFromXpathComponents(target) },
		"RenderLocationsPangoToState": func(source string, dest string) (string, error) {
			return RenderLocationsPangoToState(names, paramSpec, source, dest)
		},
		"RenderLocationsStateToPango": func(source string, dest string) (string, error) {
			return RenderLocationsStateToPango(names, paramSpec, source, dest, "resp.Diagnostics")
		},
	}

	return processTemplate(tmpl, "datasource-read-function", data, funcMap)
}

// ResourceReadFunction generates the Read function for a resource.
func ResourceReadFunction(resourceTyp properties.ResourceType, names *NameProvider, serviceName string, paramSpec *properties.Normalization, resourceSDKName string) (string, error) {
	if strings.Contains(serviceName, "group") {
		serviceName = "group"
	}

	var tmpl string
	var listAttribute string
	var exhaustive bool
	switch resourceTyp {
	case properties.ResourceEntry, properties.ResourceConfig:
		tmpl = "resource/read.tmpl"
	case properties.ResourceEntryPlural:
		tmpl = "resource/read_entry_list.tmpl"
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
	case properties.ResourceUuid:
		tmpl = "resource/read_many.tmpl"
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
		exhaustive = true
	case properties.ResourceUuidPlural:
		tmpl = "resource/read_many.tmpl"
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
	case properties.ResourceCustom:
		var err error
		tmpl, err = getCustomTemplateForFunction(paramSpec, "Read")
		if err != nil {
			return "", err
		}
	}

	listAttributeVariant := properties.NewNameVariant(listAttribute)

	data := map[string]interface{}{
		"PluralType":                       paramSpec.TerraformProviderConfig.PluralType,
		"ResourceXpathVariablesWithChecks": paramSpec.ResourceXpathVariablesWithChecks(false),
		"ResourceOrDS":                     "Resource",
		"HasEncryptedResources":            paramSpec.HasEncryptedResources(),
		"ListAttribute":                    listAttributeVariant,
		"Exhaustive":                       exhaustive,
		"EntryOrConfig":                    paramSpec.EntryOrConfig(),
		"HasEntryName":                     paramSpec.HasEntryName(),
		"structName":                       names.StructName,
		"datasourceStructName":             names.DataSourceStructName,
		"resourceStructName":               names.ResourceStructName,
		"serviceName":                      naming.CamelCase("", serviceName, "", false),
		"resourceSDKName":                  resourceSDKName,
		"locations":                        paramSpec.OrderedLocations(),
	}

	funcMap := template.FuncMap{
		"RenderEncryptedValuesInitialization": func() (string, error) {
			return RenderEncryptedValuesInitialization(properties.SchemaResource, paramSpec, "read")
		},
		"RenderEncryptedValuesFinalizer": func() (string, error) {
			return RenderEncryptedValuesFinalizer(properties.SchemaResource, paramSpec)
		},
		"AttributesFromXpathComponents": func(target string) (string, error) { return paramSpec.AttributesFromXpathComponents(target) },
		"RenderLocationsPangoToState": func(source string, dest string) (string, error) {
			return RenderLocationsPangoToState(names, paramSpec, source, dest)
		},
		"RenderLocationsStateToPango": func(source string, dest string) (string, error) {
			return RenderLocationsStateToPango(names, paramSpec, source, dest, "resp.Diagnostics")
		},
	}

	return processTemplate(tmpl, "resource-read-function", data, funcMap)
}

// ResourceUpdateFunction generates the Update function for a resource.
func ResourceUpdateFunction(resourceTyp properties.ResourceType, names *NameProvider, serviceName string, paramSpec *properties.Normalization, resourceSDKName string) (string, error) {
	if strings.Contains(serviceName, "group") {
		serviceName = "group"
	}

	var tmpl string
	var listAttribute string
	var exhaustive bool
	switch resourceTyp {
	case properties.ResourceEntry, properties.ResourceConfig:
		tmpl = "resource/update.tmpl"
	case properties.ResourceEntryPlural:
		tmpl = "resource/update_entry_list.tmpl"
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
	case properties.ResourceUuid:
		tmpl = "resource/update_many.tmpl"
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
		exhaustive = true
	case properties.ResourceUuidPlural:
		tmpl = "resource/update_many.tmpl"
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
	case properties.ResourceCustom:
		var err error
		tmpl, err = getCustomTemplateForFunction(paramSpec, "Update")
		if err != nil {
			return "", err
		}
	}

	listAttributeVariant := properties.NewNameVariant(listAttribute)

	data := map[string]interface{}{
		"PluralType":            paramSpec.TerraformProviderConfig.PluralType,
		"HasEncryptedResources": paramSpec.HasEncryptedResources(),
		"ListAttribute":         listAttributeVariant,
		"Exhaustive":            exhaustive,
		"EntryOrConfig":         paramSpec.EntryOrConfig(),
		"HasEntryName":          paramSpec.HasEntryName(),
		"structName":            names.ResourceStructName,
		"serviceName":           naming.CamelCase("", serviceName, "", false),
		"resourceSDKName":       resourceSDKName,
	}

	funcMap := template.FuncMap{
		"RenderEncryptedValuesInitialization": func() (string, error) {
			return RenderEncryptedValuesInitialization(properties.SchemaResource, paramSpec, "update")
		},
		"RenderEncryptedValuesFinalizer": func() (string, error) {
			return RenderEncryptedValuesFinalizer(properties.SchemaResource, paramSpec)
		},
		"RenderCreateUpdateMovementRequired": func(state string, entries string) (string, error) {
			return RendeCreateUpdateMovementRequired(state, entries)
		},
		"RenderLocationsStateToPango": func(source string, dest string) (string, error) {
			return RenderLocationsStateToPango(names, paramSpec, source, dest, "resp.Diagnostics")
		},
		"RenderLocationsPangoToState": func(source string, dest string) (string, error) {
			return RenderLocationsPangoToState(names, paramSpec, source, dest)
		},
	}

	return processTemplate(tmpl, "resource-update-function", data, funcMap)
}

// ResourceDeleteFunction generates the Delete function for a resource.
func ResourceDeleteFunction(resourceTyp properties.ResourceType, names *NameProvider, serviceName string, paramSpec *properties.Normalization, resourceSDKName string) (string, error) {
	if strings.Contains(serviceName, "group") {
		serviceName = "group"
	}

	var tmpl string
	var listAttribute string
	var exhaustive string
	switch resourceTyp {
	case properties.ResourceEntry, properties.ResourceConfig:
		tmpl = "resource/delete.tmpl"
	case properties.ResourceEntryPlural:
		tmpl = "resource/delete_many.tmpl"
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
	case properties.ResourceUuid:
		tmpl = "resource/delete_many.tmpl"
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
		exhaustive = "exhaustive"
	case properties.ResourceUuidPlural:
		tmpl = "resource/delete_many.tmpl"
		listAttribute = pascalCase(paramSpec.TerraformProviderConfig.PluralName)
		exhaustive = "non-exhaustive"
	case properties.ResourceCustom:
		var err error
		tmpl, err = getCustomTemplateForFunction(paramSpec, "Delete")
		if err != nil {
			return "", err
		}
	}

	listAttributeVariant := properties.NewNameVariant(listAttribute)

	data := map[string]interface{}{
		"PluralType":            paramSpec.TerraformProviderConfig.PluralType,
		"HasEncryptedResources": paramSpec.HasEncryptedResources(),
		"HasImports":            len(paramSpec.Imports.Variants) > 0,
		"EntryOrConfig":         paramSpec.EntryOrConfig(),
		"ListAttribute":         listAttributeVariant,
		"Exhaustive":            exhaustive,
		"HasEntryName":          paramSpec.HasEntryName(),
		"structName":            names.ResourceStructName,
		"serviceName":           naming.CamelCase("", serviceName, "", false),
		"resourceSDKName":       resourceSDKName,
	}

	funcMap := template.FuncMap{
		"RenderEncryptedValuesInitialization": func() (string, error) {
			return RenderEncryptedValuesInitialization(properties.SchemaResource, paramSpec, "delete")
		},
		"RenderImportLocationAssignment": func(source string, dest string) (string, error) {
			return RenderImportLocationAssignment(names, paramSpec, source, dest)
		},
		"RenderLocationsStateToPango": func(source string, dest string) (string, error) {
			return RenderLocationsStateToPango(names, paramSpec, source, dest, "resp.Diagnostics")
		},
	}

	return processTemplate(tmpl, "resource-delete-function", data, funcMap)
}
