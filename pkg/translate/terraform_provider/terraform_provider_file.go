package terraform_provider

import (
	"fmt"
	"log"
	"log/slog"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/imports"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

var _ = slog.Debug

type NameProvider = properties.TerraformNameProvider

// NewNameProvider creates a new NameProvider based on given specifications.
func NewNameProvider(spec *properties.Normalization, resourceTyp properties.ResourceType) *NameProvider {
	return properties.NewTerraformNameProvider(spec, resourceTyp)
}

// GenerateTerraformProvider handles the generation of Terraform resources and data sources.
type GenerateTerraformProvider struct{}

// createTemplate parses the provided template string using the given FuncMap and returns a new template.
func (g *GenerateTerraformProvider) createTemplate(resourceType string, spec *properties.Normalization, templateStr string, funcMap template.FuncMap) (*template.Template, error) {
	templateName := fmt.Sprintf("terraform-%s-%s", resourceType, spec.Name)
	return template.New(templateName).Funcs(funcMap).Parse(templateStr)
}

// executeTemplate executes the provided resource template using the given spec and returns an error if it fails.
func (g *GenerateTerraformProvider) executeTemplate(template *template.Template, spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile, resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, names *NameProvider) error {
	var renderedTemplate strings.Builder
	if err := template.Execute(&renderedTemplate, spec); err != nil {
		return fmt.Errorf("error executing %v template: %v", resourceTyp, err)
	}
	renderedTemplate.WriteString("\n")
	return g.updateProviderFile(&renderedTemplate, terraformProvider, resourceTyp, schemaTyp, names)
}

// updateProviderFile updates the Terraform provider file by appending the rendered template
// to the appropriate slice in the TerraformProviderFile based on the provided resourceType.
func (g *GenerateTerraformProvider) updateProviderFile(renderedTemplate *strings.Builder, terraformProvider *properties.TerraformProviderFile, resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, names *NameProvider) error {
	if schemaTyp == properties.SchemaProvider {
		terraformProvider.Code = renderedTemplate
	} else {
		log.Printf("updateProviderFile() renderedTemplate: %d", renderedTemplate.Len())
		if _, err := terraformProvider.Code.WriteString(renderedTemplate.String()); err != nil {
			return fmt.Errorf("error writing %v template: %v", resourceTyp, err)
		}
	}
	return g.appendResourceType(terraformProvider, resourceTyp, schemaTyp, names)
}

// appendResourceType appends the given struct name to the appropriate slice in the TerraformProviderFile
// based on the provided resourceType.
func (g *GenerateTerraformProvider) appendResourceType(terraformProvider *properties.TerraformProviderFile, resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, names *NameProvider) error {
	var flags properties.TerraformSpecFlags
	switch schemaTyp {
	case properties.SchemaDataSource:
		flags |= properties.TerraformSpecDatasource
		terraformProvider.DataSources = append(terraformProvider.DataSources, names.DataSourceStructName)
	case properties.SchemaResource:
		flags |= properties.TerraformSpecResource
		terraformProvider.Resources = append(terraformProvider.Resources, names.ResourceStructName)
	case properties.SchemaEphemeralResource:
		flags |= properties.TerraformSpecEphemeralResource
		terraformProvider.EphemeralResources = append(terraformProvider.EphemeralResources, names.ResourceStructName)
	case properties.SchemaProvider, properties.SchemaCommon:
	}

	switch resourceTyp {
	case properties.ResourceEntry:
		flags |= properties.TerraformSpecImportable
	case properties.ResourceCustom, properties.ResourceEntryPlural, properties.ResourceUuid, properties.ResourceUuidPlural, properties.ResourceConfig:
	}

	terraformProvider.SpecMetadata[names.MetaName] = properties.TerraformProviderSpecMetadata{
		ResourceSuffix: names.MetaName,
		StructName:     names.StructName,
		Flags:          flags,
	}
	return nil
}

// generateTerraformEntityTemplate is the common logic for generating Terraform resources and data sources.
func (g *GenerateTerraformProvider) generateTerraformEntityTemplate(resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, names *NameProvider, spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile, templateStr string, funcMap template.FuncMap) error {
	if templateStr == "" {
		return nil
	}

	var resourceType string
	switch schemaTyp {
	case properties.SchemaDataSource:
		resourceType = "DataSource"
	case properties.SchemaResource:
		resourceType = "Resource"
	case properties.SchemaEphemeralResource:
		resourceType = "EphemeralResource"
	case properties.SchemaCommon:
		resourceType = "Common"
	case properties.SchemaProvider:
		resourceType = "ProviderFile"
	}

	template, err := g.createTemplate(resourceType, spec, templateStr, funcMap)
	if err != nil {
		log.Fatalf("Error creating template: %v", err)
		return err
	}
	return g.executeTemplate(template, spec, terraformProvider, resourceTyp, schemaTyp, names)
}

// GenerateTerraformResource generates a Terraform resource template.
func (g *GenerateTerraformProvider) GenerateTerraformResource(resourceTyp properties.ResourceType, spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile) error {
	names := NewNameProvider(spec, resourceTyp)

	var structType string
	if spec.Entry != nil {
		structType = "entry"
	} else {
		structType = "config"
	}

	var hasPosition bool
	switch resourceTyp {
	case properties.ResourceUuidPlural:
		hasPosition = true
	case properties.ResourceEntry, properties.ResourceEntryPlural, properties.ResourceUuid, properties.ResourceCustom, properties.ResourceConfig:
		hasPosition = false
	}

	schemaTyp := properties.SchemaResource
	if spec.TerraformProviderConfig.Ephemeral {
		schemaTyp = properties.SchemaEphemeralResource
	}

	funcMap := template.FuncMap{
		"GoSDKSkipped": func() bool { return spec.GoSdkSkip },
		"IsEntry":      func() bool { return spec.HasEntryName() && !spec.HasEntryUuid() },
		"HasImports":   func() bool { return len(spec.Imports) > 0 },

		"IsCustom":     func() bool { return spec.TerraformProviderConfig.ResourceType == properties.TerraformResourceCustom },
		"IsUuid":       func() bool { return spec.HasEntryUuid() },
		"IsConfig":     func() bool { return !spec.HasEntryName() && !spec.HasEntryUuid() },
		"IsImportable": func() bool { return resourceTyp == properties.ResourceEntry },
		"ListAttribute": func() *properties.NameVariant {
			return properties.NewNameVariant(spec.TerraformProviderConfig.PluralName)
		},
		"HasLocations": func() bool { return len(spec.Locations) > 0 },
		"IsEphemeral":  func() bool { return spec.TerraformProviderConfig.Ephemeral },
		"tfresourcepkg": func() string {
			if spec.TerraformProviderConfig.Ephemeral {
				return "ephemeral"
			} else {
				return "resource"
			}
		},
		"resourceSDKName":         func() string { return names.PackageName },
		"HasPosition":             func() bool { return hasPosition },
		"metaName":                func() string { return names.MetaName },
		"structName":              func() string { return names.StructName },
		"dataSourceStructName":    func() string { return names.DataSourceStructName },
		"resourceStructName":      func() string { return names.ResourceStructName },
		"serviceName":             func() string { return names.TfName },
		"CreateTfIdStruct":        func() (string, error) { return CreateTfIdStruct(structType, spec.GoSdkPath[len(spec.GoSdkPath)-1]) },
		"CreateTfIdResourceModel": func() (string, error) { return CreateTfIdResourceModel(structType, names.StructName) },
		"RenderResourceStructs":   func() (string, error) { return RenderResourceStructs(resourceTyp, names, spec) },
		"RenderResourceSchema": func() (string, error) {
			return RenderResourceSchema(resourceTyp, names, spec, terraformProvider.ImportManager)
		},
		"RenderCopyToPangoFunctions": func() (string, error) {
			return RenderCopyToPangoFunctions(resourceTyp, names.PackageName, names.ResourceStructName, spec)
		},
		"RenderCopyFromPangoFunctions": func() (string, error) {
			return RenderCopyFromPangoFunctions(resourceTyp, names.PackageName, names.ResourceStructName, spec)
		},
		"ResourceCreateFunction": func(structName string, serviceName string) (string, error) {
			return ResourceCreateFunction(resourceTyp, names, serviceName, spec, terraformProvider, names.PackageName)
		},
		"ResourceReadFunction": func(structName string, serviceName string) (string, error) {
			return ResourceReadFunction(resourceTyp, names, serviceName, spec, names.PackageName)
		},
		"ResourceUpdateFunction": func(structName string, serviceName string) (string, error) {
			return ResourceUpdateFunction(resourceTyp, names, serviceName, spec, names.PackageName)
		},
		"ResourceDeleteFunction": func(structName string, serviceName string) (string, error) {
			return ResourceDeleteFunction(resourceTyp, names, serviceName, spec, names.PackageName)
		},
		"ResourceOpenFunction": func(structName string, serviceName string) (string, error) {
			return ResourceOpenFunction(resourceTyp, names, serviceName, spec, names.PackageName)
		},
		"ResourceRenewFunction": func(structName string, serviceName string) (string, error) {
			return ResourceOpenFunction(resourceTyp, names, serviceName, spec, names.PackageName)
		},
		"ResourceCloseFunction": func(structName string, serviceName string) (string, error) {
			return ResourceOpenFunction(resourceTyp, names, serviceName, spec, names.PackageName)
		},
		"FunctionSupported": func(function string) (bool, error) {
			return FunctionSupported(spec, function)
		},
		"RenderImportStateStructs": func() (string, error) {
			return RenderImportStateStructs(resourceTyp, names, spec)
		},
		"RenderImportStateCreator": func() (string, error) {
			return RenderImportStateCreator(resourceTyp, names, spec)
		},
		"ResourceImportStateFunction": func() (string, error) {
			return ResourceImportStateFunction(resourceTyp, names, spec)
		},
		"ParamToModelResource": ParamToModelResource,
		"ModelNestedStruct":    ModelNestedStruct,
		"ResourceParamToSchema": func(paramName string, paramParameters interface{}) (string, error) {
			return ParamToSchemaResource(paramName, paramParameters, terraformProvider)
		},
		"ResourceSchemaLocationAttribute": CreateResourceSchemaLocationAttribute,
	}

	if !spec.TerraformProviderConfig.SkipResource {
		terraformProvider.ImportManager.AddStandardImport("context", "")
		terraformProvider.ImportManager.AddSdkImport("github.com/PaloAltoNetworks/pango", "")
		if spec.TerraformProviderConfig.ResourceType != properties.TerraformResourceCustom {
			terraformProvider.ImportManager.AddOtherImport("github.com/PaloAltoNetworks/terraform-provider-panos/internal/manager", "sdkmanager")
		}

		// entry or uuid style resource
		if spec.Entry != nil {
			if spec.Spec == nil || spec.Spec.Params == nil {
				return fmt.Errorf("invalid resource configuration")
			}

			terraformProvider.ImportManager.AddStandardImport("errors", "")
			switch resourceTyp {
			case properties.ResourceUuid:
				terraformProvider.ImportManager.AddSdkImport("github.com/PaloAltoNetworks/pango/movement", "")
			case properties.ResourceEntry:
			case properties.ResourceUuidPlural:
			case properties.ResourceEntryPlural:
			case properties.ResourceCustom, properties.ResourceConfig:
			}

			// Generate Resource with entry style
			terraformProvider.ImportManager.AddStandardImport("fmt", "")

			if !spec.GoSdkSkip {
				terraformProvider.ImportManager.AddSdkImport(sdkPkgPath(spec), "")
			}

			conditionallyAddValidators(terraformProvider.ImportManager, spec)
			conditionallyAddModifiers(terraformProvider.ImportManager, spec)
			conditionallyAddDefaults(terraformProvider.ImportManager, spec.Spec)

			if spec.TerraformProviderConfig.Ephemeral {
				terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/ephemeral/schema", "ephschema")
			} else {
				terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema", "rsschema")
			}

			terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/attr", "")
			terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource", "")
			terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/types", "")

			if len(spec.Locations) > 0 {
				terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/diag", "")
				terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault", "")
				terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-log/tflog", "")
			}

			err := g.generateTerraformEntityTemplate(resourceTyp, schemaTyp, names, spec, terraformProvider, resourceObj, funcMap)
			if err != nil {
				return err
			}
		} else {
			// Generate Resource with config style

			terraformProvider.ImportManager.AddStandardImport("fmt", "")

			if !spec.GoSdkSkip {
				terraformProvider.ImportManager.AddSdkImport(sdkPkgPath(spec), "")
			}

			switch resourceTyp {
			case properties.ResourceEntry, properties.ResourceConfig:
				terraformProvider.ImportManager.AddStandardImport("errors", "")
			case properties.ResourceEntryPlural, properties.ResourceUuid:
			case properties.ResourceUuidPlural, properties.ResourceCustom:
			}
			conditionallyAddValidators(terraformProvider.ImportManager, spec)
			conditionallyAddModifiers(terraformProvider.ImportManager, spec)
			conditionallyAddDefaults(terraformProvider.ImportManager, spec.Spec)

			if spec.TerraformProviderConfig.Ephemeral {
				terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/ephemeral/schema", "ephschema")
			} else {
				terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema", "rsschema")
			}

			terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/attr", "")
			terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource", "")
			terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/types", "")

			if len(spec.Locations) > 0 {
				terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/diag", "")
				terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault", "")
				terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-log/tflog", "")
			}

			err := g.generateTerraformEntityTemplate(resourceTyp, schemaTyp, names, spec, terraformProvider, resourceObj, funcMap)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GenerateTerraformDataSource generates a Terraform data source and data source template.
func (g *GenerateTerraformProvider) GenerateTerraformDataSource(resourceTyp properties.ResourceType, spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile) error {

	var structType string
	if spec.Entry != nil {
		structType = "entry"
	} else {
		structType = "config"
	}

	if !spec.TerraformProviderConfig.SkipDatasource {
		terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/datasource/schema", "dsschema")
		terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/types", "")
		terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/datasource", "")

		conditionallyAddValidators(terraformProvider.ImportManager, spec)
		conditionallyAddModifiers(terraformProvider.ImportManager, spec)

		terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema", "rsschema")
		terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault", "")

		names := NewNameProvider(spec, resourceTyp)
		funcMap := template.FuncMap{
			"GoSDKSkipped":         func() bool { return spec.GoSdkSkip },
			"IsEntry":              func() bool { return spec.HasEntryName() && !spec.HasEntryUuid() },
			"HasImports":           func() bool { return len(spec.Imports) > 0 },
			"HasLocations":         func() bool { return len(spec.Locations) > 0 },
			"IsCustom":             func() bool { return spec.TerraformProviderConfig.ResourceType == properties.TerraformResourceCustom },
			"IsUuid":               func() bool { return spec.HasEntryUuid() },
			"resourceSDKName":      func() string { return names.PackageName },
			"IsConfig":             func() bool { return !spec.HasEntryName() && !spec.HasEntryUuid() },
			"metaName":             func() string { return names.MetaName },
			"structName":           func() string { return names.StructName },
			"serviceName":          func() string { return names.TfName },
			"CreateTfIdStruct":     func() (string, error) { return CreateTfIdStruct(structType, spec.GoSdkPath[len(spec.GoSdkPath)-1]) },
			"dataSourceStructName": func() string { return names.DataSourceStructName },
			"DataSourceReadFunction": func(structName string, serviceName string) (string, error) {
				return DataSourceReadFunction(resourceTyp, names, serviceName, spec, names.PackageName)
			},
			"FunctionSupported": func(function string) (bool, error) {
				return FunctionSupported(spec, function)
			},
			"RenderCopyToPangoFunctions": func() (string, error) {
				return RenderCopyToPangoFunctions(resourceTyp, names.PackageName, names.DataSourceStructName, spec)
			},
			"RenderCopyFromPangoFunctions": func() (string, error) {
				return RenderCopyFromPangoFunctions(resourceTyp, names.PackageName, names.DataSourceStructName, spec)
			},
			"RenderDataSourceStructs": func() (string, error) { return RenderDataSourceStructs(resourceTyp, names, spec) },
			"RenderDataSourceSchema": func() (string, error) {
				return RenderDataSourceSchema(resourceTyp, names, spec, terraformProvider.ImportManager)
			},
		}
		err := g.generateTerraformEntityTemplate(resourceTyp, properties.SchemaDataSource, names, spec, terraformProvider, dataSourceSingletonObj, funcMap)
		if err != nil {
			return err
		}
	}

	//TODO: I've disable creating DS List to omit creation the non existing object, please remove this comment once the DS List is implemented.
	//if !spec.TerraformProviderConfig.SkipDatasourceListing {
	//	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/datasource/schema", "dsschema")
	//
	//	resourceType := "DataSourceList"
	//	names := NewNameProvider(spec, resourceType)
	//	funcMap := template.FuncMap{
	//		"metaName":   func() string { return names.MetaName },
	//		"structName": func() string { return names.StructName },
	//	}
	//	err := g.generateTerraformEntityTemplate(resourceType, names, spec, terraformProvider, dataSourceListObj, funcMap)
	//	if err != nil {
	//		return err
	//	}
	//}

	return nil
}

func (g *GenerateTerraformProvider) GenerateCommonCode(resourceTyp properties.ResourceType, spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile) error {

	// For specs with locations, we need to implement MarshalJSON and UnmarshalJSON methods
	if len(spec.Locations) > 0 {
		terraformProvider.ImportManager.AddStandardImport("encoding/json", "")
	}
	// Imports required by resources that can be imported into state
	if resourceTyp == properties.ResourceEntry {
		terraformProvider.ImportManager.AddStandardImport("encoding/base64", "")
		terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/types/basetypes", "")
		terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/path", "")
	}
	names := NewNameProvider(spec, resourceTyp)
	funcMap := template.FuncMap{
		"HasLocations":          func() bool { return len(spec.Locations) > 0 },
		"RenderLocationStructs": func() (string, error) { return RenderLocationStructs(resourceTyp, names, spec) },
		"RenderLocationSchemaGetter": func() (string, error) {
			return RenderLocationSchemaGetter(names, spec, terraformProvider.ImportManager)
		},
		"RenderLocationMarshallers": func() (string, error) { return RenderLocationMarshallers(names, spec) },
		"RenderCustomCommonCode":    func() string { return RenderCustomCommonCode(names, spec) },
	}
	return g.generateTerraformEntityTemplate(resourceTyp, properties.SchemaCommon, names, spec, terraformProvider, commonTemplate, funcMap)
}

// GenerateTerraformProviderFile generates the entire provider file.
func (g *GenerateTerraformProvider) GenerateTerraformProviderFile(spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile) error {
	// TODO: Uncomment this once support for config and uuid style resouces is added
	// terraformProvider.ImportManager.AddSdkImport(sdkPkgPath(spec), "")

	funcMap := template.FuncMap{
		"renderImports":       func() (string, error) { return terraformProvider.ImportManager.RenderImports() },
		"renderCustomImports": func() string { return RenderCustomImports(spec) },
		"renderCode":          func() string { return terraformProvider.Code.String() },
	}
	return g.generateTerraformEntityTemplate(properties.ResourceCustom, properties.SchemaProvider, &NameProvider{}, spec, terraformProvider, providerFile, funcMap)
}

func (g *GenerateTerraformProvider) GenerateTerraformProvider(terraformProvider *properties.TerraformProviderFile, spec *properties.Normalization, providerConfig properties.TerraformProvider) error {
	terraformProvider.ImportManager.AddStandardImport("context", "")
	terraformProvider.ImportManager.AddStandardImport("strings", "")
	terraformProvider.ImportManager.AddStandardImport("fmt", "")
	terraformProvider.ImportManager.AddStandardImport("log/slog", "")
	terraformProvider.ImportManager.AddSdkImport("github.com/PaloAltoNetworks/pango", "sdk")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/datasource", "")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/function", "")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/provider", "")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/provider/schema", "")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource", "")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/ephemeral", "")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/types", "")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-log/tflog", "")

	funcMap := template.FuncMap{
		"RenderImports":         func() (string, error) { return terraformProvider.ImportManager.RenderImports() },
		"DataSources":           func() []string { return terraformProvider.DataSources },
		"EphemeralResources":    func() []string { return terraformProvider.EphemeralResources },
		"Resources":             func() []string { return terraformProvider.Resources },
		"RenderResourceFuncMap": func() (string, error) { return RenderResourceFuncMap(terraformProvider.SpecMetadata) },
		"ProviderParams":        func() map[string]properties.TerraformProviderParams { return providerConfig.TerraformProviderParams },
		"ParamToModelBasic": func(paramName string, paramProp properties.TerraformProviderParams) (string, error) {
			return ParamToModelBasic(paramName, paramProp)
		},
		"ParamToSchemaProvider": func(paramName string, paramProp properties.TerraformProviderParams) (string, error) {
			return ParamToSchemaProvider(paramName, paramProp)
		},
	}
	return g.generateTerraformEntityTemplate(properties.ResourceCustom, properties.SchemaProvider, &NameProvider{}, spec, terraformProvider, provider, funcMap)
}

func sdkPkgPath(spec *properties.Normalization) string {
	path := fmt.Sprintf("github.com/PaloAltoNetworks/pango/%s", strings.Join(spec.GoSdkPath, "/"))

	return path
}

func hasVariantsImpl(props []*properties.SpecParam) bool {
	for _, elt := range props {
		if len(elt.EnumValues) > 0 {
			return true
		}

		if elt.Spec == nil {
			continue
		}

		if len(elt.Spec.OneOf) > 0 {
			return true
		}

		var params []*properties.SpecParam
		for _, elt := range elt.Spec.Params {
			params = append(params, elt)
		}
		for _, elt := range elt.Spec.OneOf {
			params = append(params, elt)
		}

		if hasVariantsImpl(params) {
			return true
		}
	}

	return false
}

func conditionallyAddValidators(manager *imports.Manager, spec *properties.Normalization) {
	if spec.Spec == nil {
		return
	}

	validatorRequired := func() bool {
		if len(spec.Spec.OneOf) > 0 {
			return true
		}

		var params []*properties.SpecParam
		for _, elt := range spec.Spec.Params {
			params = append(params, elt)
		}
		for _, elt := range spec.Spec.OneOf {
			params = append(params, elt)
		}
		return hasVariantsImpl(params)
	}

	if validatorRequired() || len(spec.Locations) > 1 {
		manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/path", "")
		manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/schema/validator", "")
	}

}

func conditionallyAddModifiers(manager *imports.Manager, spec *properties.Normalization) {
	if len(spec.Locations) > 0 {
		manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier", "")
	}

	for _, loc := range spec.Locations {
		if len(loc.Vars) == 0 {
			manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier", "")
		} else {
			manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier", "")
			manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier", "")
		}
	}
}

func conditionallyAddDefaults(manager *imports.Manager, spec *properties.Spec) {
	for _, elt := range spec.Params {
		if elt.Type == "" {
			conditionallyAddDefaults(manager, elt.Spec)
		} else if elt.Default != "" {
			packageName := fmt.Sprintf("%sdefault", elt.Type)
			fullPackage := fmt.Sprintf("github.com/hashicorp/terraform-plugin-framework/resource/schema/%s", packageName)
			manager.AddHashicorpImport(fullPackage, "")
		}
	}

	for _, elt := range spec.OneOf {
		if elt.Type == "" {
			conditionallyAddDefaults(manager, elt.Spec)
		} else if elt.Default != "" {
			packageName := fmt.Sprintf("%sdefault", elt.Type)
			fullPackage := fmt.Sprintf("github.com/hashicorp/terraform-plugin-framework/resource/schema/%s", packageName)
			manager.AddHashicorpImport(fullPackage, "")
		}
	}
}
