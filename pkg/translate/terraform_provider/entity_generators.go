package terraform_provider

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

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

	// Try to load from file if templateStr looks like a file path
	var tmplContent string
	if strings.HasSuffix(templateStr, ".tmpl") || strings.Contains(templateStr, "/") {
		templatePath := filepath.Join("templates", "terraform-provider", templateStr)
		if content, err := os.ReadFile(templatePath); err == nil {
			tmplContent = string(content)
		} else {
			// Fallback to embedded string if file doesn't exist
			tmplContent = templateStr
		}
	} else {
		tmplContent = templateStr
	}

	return template.New(templateName).Funcs(funcMap).Parse(tmplContent)
}

// executeTemplate executes the provided resource template using the given spec and returns an error if it fails.
func (g *GenerateTerraformProvider) executeTemplate(template *template.Template, spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile, resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, names *NameProvider) error {
	var renderedTemplate strings.Builder
	if err := template.Execute(&renderedTemplate, spec); err != nil {
		return fmt.Errorf("error executing %v template: %v", resourceTyp, err)
	}
	renderedTemplate.WriteString("\n")
	return g.updateProviderFile(spec, &renderedTemplate, terraformProvider, resourceTyp, schemaTyp, names)
}

// updateProviderFile updates the Terraform provider file by appending the rendered template
// to the appropriate slice in the TerraformProviderFile based on the provided resourceType.
func (g *GenerateTerraformProvider) updateProviderFile(spec *properties.Normalization, renderedTemplate *strings.Builder, terraformProvider *properties.TerraformProviderFile, resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, names *NameProvider) error {
	if schemaTyp == properties.SchemaProvider {
		terraformProvider.Code = renderedTemplate
	} else {
		slog.Debug("updateProviderFile() renderedTemplate", "renderedTemplate.Len()", renderedTemplate.Len())
		if _, err := terraformProvider.Code.WriteString(renderedTemplate.String()); err != nil {
			return fmt.Errorf("error writing %v template: %v", resourceTyp, err)
		}
	}
	return g.appendResourceType(spec, terraformProvider, resourceTyp, schemaTyp, names)
}

// appendResourceType appends the given struct name to the appropriate slice in the TerraformProviderFile
// based on the provided resourceType.
func (g *GenerateTerraformProvider) appendResourceType(spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile, resourceTyp properties.ResourceType, schemaTyp properties.SchemaType, names *NameProvider) error {
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
	case properties.SchemaAction:
		terraformProvider.Actions = append(terraformProvider.Actions, names.ActionStructName())
	case properties.SchemaProvider, properties.SchemaCommon:
	default:
		panic(fmt.Sprintf("unsupported schemaTyp: '%s'", schemaTyp))
	}

	switch resourceTyp {
	case properties.ResourceEntry, properties.ResourceEntryPlural, properties.ResourceUuid, properties.ResourceUuidPlural:
		if !spec.TerraformProviderConfig.SkipResource {
			flags |= properties.TerraformSpecImportable
		}
	case properties.ResourceCustom, properties.ResourceConfig:
	}

	// Accumulate flags if metadata already exists for this resource
	if existing, ok := terraformProvider.SpecMetadata[names.MetaName]; ok {
		flags |= existing.Flags
	}

	terraformProvider.SpecMetadata[names.MetaName] = properties.TerraformProviderSpecMetadata{
		ResourceSuffix: names.MetaName,
		StructName:     names.StructName,
		Subcategory:    spec.TerraformProviderConfig.Subcategory,
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
	case properties.SchemaAction:
		resourceType = "Action"
	default:
		panic(fmt.Sprintf("unsupported schemaTyp: '%+v'", schemaTyp))
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

		"IsCustom":    func() bool { return spec.TerraformProviderConfig.ResourceType == properties.TerraformResourceCustom },
		"IsUuid":      func() bool { return spec.HasEntryUuid() },
		"IsConfig":    func() bool { return !spec.HasEntryName() && !spec.HasEntryUuid() },
		"IsEphemeral": func() bool { return spec.TerraformProviderConfig.Ephemeral },
		"ListAttribute": func() *properties.NameVariant {
			return properties.NewNameVariant(spec.TerraformProviderConfig.PluralName)
		},
		"HasLocations": func() bool { return len(spec.Locations) > 0 },
		"IsImportable": func() bool {
			switch resourceTyp {
			case properties.ResourceEntry, properties.ResourceEntryPlural, properties.ResourceUuid, properties.ResourceUuidPlural:
				return true
			case properties.ResourceConfig, properties.ResourceCustom:
				return false
			}

			panic("unreachable")
		},
		"IsResourcePlural": func() bool {
			switch resourceTyp {
			case properties.ResourceEntryPlural, properties.ResourceUuid, properties.ResourceUuidPlural:
				return true
			case properties.ResourceEntry, properties.ResourceConfig, properties.ResourceCustom:
				return false
			}

			panic("unreachable")
		},
		"tfresourcepkg": func() string {
			if spec.TerraformProviderConfig.Ephemeral {
				return "ephemeral"
			} else {
				return "resource"
			}
		},
		"resourceSDKName":       func() string { return names.PackageName },
		"HasPosition":           func() bool { return hasPosition },
		"HasCustomValidation":   func() bool { return spec.TerraformProviderConfig.CustomValidation },
		"metaName":              func() string { return names.MetaName },
		"structName":            func() string { return names.StructName },
		"dataSourceStructName":  func() string { return names.DataSourceStructName },
		"resourceStructName":    func() string { return names.ResourceStructName },
		"serviceName":           func() string { return names.TfName },
		"RenderResourceStructs": func() (string, error) { return RenderResourceStructs(resourceTyp, names, spec) },
		"RenderResourceValidators": func() (string, error) {
			return RenderResourceValidators(resourceTyp, names, spec, terraformProvider.ImportManager)
		},
		"RenderResourceSchema": func() (string, error) {
			return RenderResourceSchema(resourceTyp, names, spec, terraformProvider.ImportManager)
		},
		"RenderModelAttributeTypesFunction": func() (string, error) {
			return RenderModelAttributeTypesFunction(resourceTyp, properties.SchemaResource, names, spec)
		},
		"RenderCopyToPangoFunctions": func() (string, error) {
			return RenderCopyToPangoFunctions(names, resourceTyp, properties.SchemaResource, names.PackageName, names.ResourceStructName, spec)
		},
		"RenderCopyFromPangoFunctions": func() (string, error) {
			return RenderCopyFromPangoFunctions(names, resourceTyp, properties.SchemaResource, names.PackageName, names.ResourceStructName, spec)
		},
		"RenderXpathComponentsGetter": func() (string, error) {
			return RenderXpathComponentsGetter(names.ResourceStructName, spec)
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
		"RenderImportStateMarshallers": func() (string, error) {
			return RenderImportStateMarshallers(resourceTyp, names, spec)
		},
		"RenderImportStateCreator": func() (string, error) {
			return RenderImportStateCreator(resourceTyp, names, spec)
		},
		"ResourceImportStateFunction": func() (string, error) {
			return ResourceImportStateFunction(resourceTyp, names, spec)
		},
		"ResourceSchemaLocationAttribute": CreateResourceSchemaLocationAttribute,
	}

	if !spec.TerraformProviderConfig.SkipResource {
		terraformProvider.ImportManager.AddStandardImport("context", "")

		terraformProvider.ImportManager.AddSdkImport("github.com/PaloAltoNetworks/pango", "")

		switch spec.TerraformProviderConfig.ResourceType {
		case properties.TerraformResourceEntry, properties.TerraformResourceUuid:
			terraformProvider.ImportManager.AddStandardImport("errors", "")
			terraformProvider.ImportManager.AddStandardImport("encoding/base64", "")
			terraformProvider.ImportManager.AddOtherImport("github.com/PaloAltoNetworks/terraform-provider-panos/internal/manager", "sdkmanager")
			terraformProvider.ImportManager.AddSdkImport("github.com/PaloAltoNetworks/pango/util", "pangoutil")
		case properties.TerraformResourceConfig:
			terraformProvider.ImportManager.AddOtherImport("github.com/PaloAltoNetworks/terraform-provider-panos/internal/manager", "sdkmanager")
		case properties.TerraformResourceCustom:
		}

		// entry or uuid style resource
		if spec.Entry != nil {
			if spec.Spec == nil || spec.Spec.Params == nil {
				return fmt.Errorf("invalid resource configuration")
			}

			switch resourceTyp {
			case properties.ResourceUuid, properties.ResourceUuidPlural:
				terraformProvider.ImportManager.AddSdkImport("github.com/PaloAltoNetworks/pango/movement", "")
				terraformProvider.ImportManager.AddStandardImport("strings", "")
			case properties.ResourceEntry:
			case properties.ResourceEntryPlural:
			case properties.ResourceCustom, properties.ResourceConfig:
			}

			switch resourceTyp {
			case properties.ResourceEntry, properties.ResourceConfig:
			case properties.ResourceEntryPlural, properties.ResourceUuid, properties.ResourceUuidPlural, properties.ResourceCustom:
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
				terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/ephemeral", "")
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
				if resourceTyp != properties.ResourceCustom {
					terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-log/tflog", "")
				}
			}

			if spec.ResourceXpathVariablesWithChecks(false) {
				terraformProvider.ImportManager.AddStandardImport("strings", "")
				terraformProvider.ImportManager.AddSdkImport("github.com/PaloAltoNetworks/pango/util", "pangoutil")
			}

			err := g.generateTerraformEntityTemplate(resourceTyp, schemaTyp, names, spec, terraformProvider, "resource/resource.tmpl", funcMap)
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
				terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/ephemeral", "")
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
				if resourceTyp != properties.ResourceCustom {
					terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-log/tflog", "")
				}
			}

			if spec.ResourceXpathVariablesWithChecks(false) {
				terraformProvider.ImportManager.AddStandardImport("strings", "")
				terraformProvider.ImportManager.AddSdkImport("github.com/PaloAltoNetworks/pango/util", "pangoutil")
			}

			err := g.generateTerraformEntityTemplate(resourceTyp, schemaTyp, names, spec, terraformProvider, "resource/resource.tmpl", funcMap)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GenerateTerraformAction generates a Terraform action template.
func (o *GenerateTerraformProvider) GenerateTerraformAction(spec *properties.Normalization, provider *properties.TerraformProviderFile) error {
	provider.ImportManager.AddStandardImport("context", "")
	provider.ImportManager.AddStandardImport("fmt", "")

	provider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/action", "")
	provider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/action/schema", "")
	provider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/attr", "")

	if len(spec.Spec.OneOf) > 0 || len(spec.Spec.Params) > 0 {
		provider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/types", "")
	}

	provider.ImportManager.AddSdkImport("github.com/PaloAltoNetworks/pango", "")

	resourceTyp := properties.ResourceEntry
	names := NewNameProvider(spec, resourceTyp)

	funcMap := template.FuncMap{
		"structName":          func() string { return names.ActionStructName() },
		"metaName":            func() string { return names.MetaName },
		"HasCustomValidation": func() bool { return spec.TerraformProviderConfig.CustomValidation },

		"RenderStructs": func() (string, error) { return RenderStructs(resourceTyp, properties.SchemaAction, names, spec) },
		"RenderSchema": func() (string, error) {
			return RenderSchema(resourceTyp, properties.SchemaAction, names, spec, provider.ImportManager)
		},
	}

	return o.generateTerraformEntityTemplate(resourceTyp, properties.SchemaAction, names, spec, provider, "action/action.tmpl", funcMap)
}

// GenerateTerraformDataSource generates a Terraform data source and data source template.
func (g *GenerateTerraformProvider) GenerateTerraformDataSource(resourceTyp properties.ResourceType, spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile) error {

	if !spec.TerraformProviderConfig.SkipDatasource {
		terraformProvider.ImportManager.AddStandardImport("context", "")
		terraformProvider.ImportManager.AddStandardImport("fmt", "")

		if spec.TerraformProviderConfig.ResourceType != properties.TerraformResourceCustom {
			terraformProvider.ImportManager.AddStandardImport("errors", "")
			terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-log/tflog", "")
			terraformProvider.ImportManager.AddOtherImport("github.com/PaloAltoNetworks/terraform-provider-panos/internal/manager", "sdkmanager")
			terraformProvider.ImportManager.AddSdkImport("github.com/PaloAltoNetworks/pango/util", "pangoutil")
		}

		terraformProvider.ImportManager.AddSdkImport("github.com/PaloAltoNetworks/pango", "")
		if !spec.GoSdkSkip {
			terraformProvider.ImportManager.AddSdkImport(sdkPkgPath(spec), "")
		}

		terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/attr", "")
		terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/diag", "")
		terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/datasource/schema", "dsschema")
		terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/types", "")
		terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/datasource", "")

		terraformProvider.ImportManager.AddStandardImport("context", "")
		terraformProvider.ImportManager.AddStandardImport("fmt", "")
		terraformProvider.ImportManager.AddSdkImport("github.com/PaloAltoNetworks/pango", "")
		if spec.TerraformProviderConfig.ResourceType != properties.TerraformResourceCustom {
			terraformProvider.ImportManager.AddOtherImport("github.com/PaloAltoNetworks/terraform-provider-panos/internal/manager", "sdkmanager")
		}

		conditionallyAddValidators(terraformProvider.ImportManager, spec)
		conditionallyAddModifiers(terraformProvider.ImportManager, spec)

		if !spec.GoSdkSkip {
			terraformProvider.ImportManager.AddSdkImport(sdkPkgPath(spec), "")
		}

		terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/attr", "")
		terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/diag", "")
		terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema", "rsschema")
		terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault", "")

		names := NewNameProvider(spec, resourceTyp)
		funcMap := template.FuncMap{
			"GoSDKSkipped": func() bool { return spec.GoSdkSkip },
			"IsEntry":      func() bool { return spec.HasEntryName() && !spec.HasEntryUuid() },
			"IsResourcePlural": func() bool {
				switch resourceTyp {
				case properties.ResourceEntryPlural, properties.ResourceUuid, properties.ResourceUuidPlural:
					return true
				case properties.ResourceEntry, properties.ResourceConfig, properties.ResourceCustom:
					return false
				}

				panic("unreachable")
			},
			"HasImports":           func() bool { return len(spec.Imports) > 0 },
			"HasLocations":         func() bool { return len(spec.Locations) > 0 },
			"IsCustom":             func() bool { return spec.TerraformProviderConfig.ResourceType == properties.TerraformResourceCustom },
			"IsUuid":               func() bool { return spec.HasEntryUuid() },
			"resourceSDKName":      func() string { return names.PackageName },
			"IsConfig":             func() bool { return !spec.HasEntryName() && !spec.HasEntryUuid() },
			"metaName":             func() string { return names.MetaName },
			"structName":           func() string { return names.StructName },
			"serviceName":          func() string { return names.TfName },
			"dataSourceStructName": func() string { return names.DataSourceStructName },
			"DataSourceReadFunction": func(structName string, serviceName string) (string, error) {
				return DataSourceReadFunction(resourceTyp, names, serviceName, spec, names.PackageName)
			},
			"FunctionSupported": func(function string) (bool, error) {
				return FunctionSupported(spec, function)
			},
			"RenderModelAttributeTypesFunction": func() (string, error) {
				return RenderModelAttributeTypesFunction(resourceTyp, properties.SchemaDataSource, names, spec)
			},
			"RenderCopyToPangoFunctions": func() (string, error) {
				return RenderCopyToPangoFunctions(names, resourceTyp, properties.SchemaDataSource, names.PackageName, names.DataSourceStructName, spec)
			},
			"RenderCopyFromPangoFunctions": func() (string, error) {
				return RenderCopyFromPangoFunctions(names, resourceTyp, properties.SchemaDataSource, names.PackageName, names.DataSourceStructName, spec)
			},
			"RenderXpathComponentsGetter": func() (string, error) {
				return RenderXpathComponentsGetter(names.DataSourceStructName, spec)
			},
			"RenderDataSourceStructs": func() (string, error) { return RenderDataSourceStructs(resourceTyp, names, spec) },
			"RenderDataSourceSchema": func() (string, error) {
				return RenderDataSourceSchema(resourceTyp, names, spec, terraformProvider.ImportManager)
			},
		}
		err := g.generateTerraformEntityTemplate(resourceTyp, properties.SchemaDataSource, names, spec, terraformProvider, "datasource/datasource.tmpl", funcMap)
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

// GenerateCommonCode generates common Terraform code for resources and data sources.
func (g *GenerateTerraformProvider) GenerateCommonCode(resourceTyp properties.ResourceType, spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile) error {

	// For specs with locations, we need to implement MarshalJSON and UnmarshalJSON methods
	if len(spec.Locations) > 0 {
		terraformProvider.ImportManager.AddStandardImport("encoding/json", "")
	}
	// Imports required by resources that can be imported into state
	switch resourceTyp {
	case properties.ResourceEntry, properties.ResourceEntryPlural, properties.ResourceUuid, properties.ResourceUuidPlural:
		if !spec.TerraformProviderConfig.SkipResource {
			terraformProvider.ImportManager.AddStandardImport("encoding/base64", "")
		}
		terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/types/basetypes", "")
		terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/path", "")
	case properties.ResourceConfig:
		terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/types/basetypes", "")
	case properties.ResourceCustom:
		if len(spec.Locations) > 0 {
			terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/types/basetypes", "")
		}
	}

	names := NewNameProvider(spec, resourceTyp)
	funcMap := template.FuncMap{
		"HasLocations":          func() bool { return len(spec.Locations) > 0 },
		"RenderLocationStructs": func() (string, error) { return RenderLocationStructs(resourceTyp, names, spec) },
		"RenderLocationSchemaGetter": func() (string, error) {
			return RenderLocationSchemaGetter(names, spec, terraformProvider.ImportManager)
		},
		"RenderLocationMarshallers": func() (string, error) { return RenderLocationMarshallers(names, spec) },
		"RenderLocationAttributeTypes": func() (string, error) {
			return RenderLocationAttributeTypes(names, spec)
		},
	}
	return g.generateTerraformEntityTemplate(resourceTyp, properties.SchemaCommon, names, spec, terraformProvider, "common/common.tmpl", funcMap)
}
