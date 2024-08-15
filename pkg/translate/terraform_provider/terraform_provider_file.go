package terraform_provider

import (
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/imports"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

// NameProvider encapsulates naming conventions for Terraform resources.
type NameProvider struct {
	TfName               string
	MetaName             string
	StructName           string
	DataSourceStructName string
	ResourceStructName   string
	PackageName          string
}

// NewNameProvider creates a new NameProvider based on given specifications.
func NewNameProvider(spec *properties.Normalization, resourceTyp properties.ResourceType) *NameProvider {
	var tfName string
	switch resourceTyp {
	case properties.ResourceEntry:
		tfName = spec.Name
	case properties.ResourceEntryPlural:
		tfName = spec.TerraformProviderConfig.PluralSuffix
	case properties.ResourceUuid:
		tfName = spec.TerraformProviderConfig.Suffix
	case properties.ResourceUuidPlural:
		suffix := spec.TerraformProviderConfig.Suffix
		pluralName := spec.TerraformProviderConfig.PluralName
		tfName = fmt.Sprintf("%s_%s", suffix, pluralName)
	}
	objectName := tfName

	metaName := fmt.Sprintf("_%s", naming.Underscore("", strings.ToLower(objectName), ""))
	structName := naming.CamelCase("", tfName, "", true)
	dataSourceStructName := naming.CamelCase("", tfName, "DataSource", true)
	resourceStructName := naming.CamelCase("", tfName, "Resource", true)
	packageName := spec.GoSdkPath[len(spec.GoSdkPath)-1]
	return &NameProvider{tfName, metaName, structName, dataSourceStructName, resourceStructName, packageName}
}

// GenerateTerraformProvider handles the generation of Terraform resources and data sources.
type GenerateTerraformProvider struct{}

// createTemplate parses the provided template string using the given FuncMap and returns a new template.
func (g *GenerateTerraformProvider) createTemplate(resourceType string, spec *properties.Normalization, templateStr string, funcMap template.FuncMap) (*template.Template, error) {
	templateName := fmt.Sprintf("terraform-%s-%s", resourceType, spec.Name)
	return template.New(templateName).Funcs(funcMap).Parse(templateStr)
}

// executeTemplate executes the provided resource template using the given spec and returns an error if it fails.
func (g *GenerateTerraformProvider) executeTemplate(template *template.Template, spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile, resourceType string, names *NameProvider) error {
	var renderedTemplate strings.Builder
	if err := template.Execute(&renderedTemplate, spec); err != nil {
		return fmt.Errorf("error executing %s template: %v", resourceType, err)
	}
	renderedTemplate.WriteString("\n")
	return g.updateProviderFile(&renderedTemplate, terraformProvider, resourceType, names)
}

// updateProviderFile updates the Terraform provider file by appending the rendered template
// to the appropriate slice in the TerraformProviderFile based on the provided resourceType.
func (g *GenerateTerraformProvider) updateProviderFile(renderedTemplate *strings.Builder, terraformProvider *properties.TerraformProviderFile, resourceType string, names *NameProvider) error {
	switch resourceType {
	case "ProviderFile":
		terraformProvider.Code = renderedTemplate
	default:
		log.Printf("updateProviderFile() renderedTemplate: %d", renderedTemplate.Len())
		if _, err := terraformProvider.Code.WriteString(renderedTemplate.String()); err != nil {
			return fmt.Errorf("error writing %s template: %v", resourceType, err)
		}
	}
	return g.appendResourceType(terraformProvider, resourceType, names)
}

// appendResourceType appends the given struct name to the appropriate slice in the TerraformProviderFile
// based on the provided resourceType.
func (g *GenerateTerraformProvider) appendResourceType(terraformProvider *properties.TerraformProviderFile, resourceType string, names *NameProvider) error {
	switch resourceType {
	case "DataSource", "DataSourceList":
		terraformProvider.DataSources = append(terraformProvider.DataSources, names.DataSourceStructName)
	case "Resource":
		terraformProvider.Resources = append(terraformProvider.Resources, names.ResourceStructName)
	}
	return nil
}

// generateTerraformEntityTemplate is the common logic for generating Terraform resources and data sources.
func (g *GenerateTerraformProvider) generateTerraformEntityTemplate(resourceType string, names *NameProvider, spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile, templateStr string, funcMap template.FuncMap) error {
	if templateStr == "" {
		return nil
	}
	template, err := g.createTemplate(resourceType, spec, templateStr, funcMap)
	if err != nil {
		log.Fatalf("Error creating template: %v", err)
		return err
	}
	return g.executeTemplate(template, spec, terraformProvider, resourceType, names)
}

// GenerateTerraformResource generates a Terraform resource template.
func (g *GenerateTerraformProvider) GenerateTerraformResource(resourceTyp properties.ResourceType, spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile) error {
	resourceType := "Resource"
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
	case properties.ResourceEntry, properties.ResourceEntryPlural, properties.ResourceUuid:
		hasPosition = false
	}

	funcMap := template.FuncMap{
		"HasPosition":             func() bool { return hasPosition },
		"metaName":                func() string { return names.MetaName },
		"structName":              func() string { return names.StructName },
		"dataSourceStructName":    func() string { return names.DataSourceStructName },
		"resourceStructName":      func() string { return names.ResourceStructName },
		"serviceName":             func() string { return names.TfName },
		"CreateTfIdStruct":        func() (string, error) { return CreateTfIdStruct(structType, spec.GoSdkPath[len(spec.GoSdkPath)-1]) },
		"CreateTfIdResourceModel": func() (string, error) { return CreateTfIdResourceModel(structType, names.StructName) },
		"RenderResourceStructs":   func() (string, error) { return RenderResourceStructs(resourceTyp, names, spec) },
		"RenderResourceSchema":    func() (string, error) { return RenderResourceSchema(resourceTyp, names, spec) },
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

		// entry or uuid style resource
		if spec.Entry != nil {
			if spec.Spec == nil || spec.Spec.Params == nil {
				return fmt.Errorf("invalid resource configuration")
			}

			switch resourceTyp {
			case properties.ResourceUuid:
				terraformProvider.ImportManager.AddSdkImport("github.com/PaloAltoNetworks/pango/xmlapi", "")
				terraformProvider.ImportManager.AddSdkImport("github.com/PaloAltoNetworks/pango/util", "")
				terraformProvider.ImportManager.AddSdkImport("github.com/PaloAltoNetworks/pango/rule", "")
			case properties.ResourceUuidPlural:
				terraformProvider.ImportManager.AddSdkImport("github.com/PaloAltoNetworks/pango/xmlapi", "")
				terraformProvider.ImportManager.AddSdkImport("github.com/PaloAltoNetworks/pango/util", "")
			case properties.ResourceEntryPlural:
				terraformProvider.ImportManager.AddSdkImport("github.com/PaloAltoNetworks/pango/xmlapi", "")
				terraformProvider.ImportManager.AddSdkImport("github.com/PaloAltoNetworks/pango/util", "")
			case properties.ResourceEntry:
			}

			// Generate Resource with entry style
			terraformProvider.ImportManager.AddStandardImport("fmt", "")

			terraformProvider.ImportManager.AddSdkImport(sdkPkgPath(spec), "")

			conditionallyAddValidators(terraformProvider.ImportManager, spec)
			conditionallyAddModifiers(terraformProvider.ImportManager, spec)
			conditionallyAddDefaults(terraformProvider.ImportManager, spec.Spec)

			terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/path", "")
			terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/diag", "")
			terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/attr", "")
			terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource", "")
			terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema", "rsschema")
			terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault", "")
			terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/types", "")
			terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-log/tflog", "")

			err := g.generateTerraformEntityTemplate(resourceType, names, spec, terraformProvider, resourceObj, funcMap)
			if err != nil {
				return err
			}
		} else {
			// Generate Resource with config style

			terraformProvider.ImportManager.AddStandardImport("fmt", "")

			terraformProvider.ImportManager.AddSdkImport(sdkPkgPath(spec), "")

			conditionallyAddValidators(terraformProvider.ImportManager, spec)
			conditionallyAddModifiers(terraformProvider.ImportManager, spec)
			conditionallyAddDefaults(terraformProvider.ImportManager, spec.Spec)

			terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/path", "")
			terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/diag", "")
			terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/attr", "")
			terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource", "")
			terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema", "rsschema")
			terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault", "")
			terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/types", "")
			terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-log/tflog", "")

			err := g.generateTerraformEntityTemplate(resourceType, names, spec, terraformProvider, resourceObj, funcMap)
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

		resourceType := "DataSource"
		names := NewNameProvider(spec, resourceTyp)
		funcMap := template.FuncMap{
			"metaName":             func() string { return names.MetaName },
			"structName":           func() string { return names.StructName },
			"serviceName":          func() string { return names.TfName },
			"CreateTfIdStruct":     func() (string, error) { return CreateTfIdStruct(structType, spec.GoSdkPath[len(spec.GoSdkPath)-1]) },
			"dataSourceStructName": func() string { return names.DataSourceStructName },
			"DataSourceReadFunction": func(structName string, serviceName string) (string, error) {
				return DataSourceReadFunction(resourceTyp, names, serviceName, spec, names.PackageName)
			},
			"RenderCopyFromPangoFunctions": func() (string, error) {
				return RenderCopyFromPangoFunctions(resourceTyp, names.PackageName, names.DataSourceStructName, spec)
			},
			"RenderCopyToPangoFunctions": func() (string, error) {
				return RenderCopyToPangoFunctions(resourceTyp, names.PackageName, names.DataSourceStructName, spec)
			},
			"RenderDataSourceStructs": func() (string, error) { return RenderDataSourceStructs(resourceTyp, names, spec) },
			"RenderDataSourceSchema":  func() (string, error) { return RenderDataSourceSchema(resourceTyp, names, spec) },
		}
		err := g.generateTerraformEntityTemplate(resourceType, names, spec, terraformProvider, dataSourceSingletonObj, funcMap)
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
	names := NewNameProvider(spec, resourceTyp)
	funcMap := template.FuncMap{
		"RenderLocationStructs":      func() (string, error) { return RenderLocationStructs(resourceTyp, names, spec) },
		"RenderLocationSchemaGetter": func() (string, error) { return RenderLocationSchemaGetter(names, spec) },
	}
	return g.generateTerraformEntityTemplate("Common", names, spec, terraformProvider, commonTemplate, funcMap)
}

// GenerateTerraformProviderFile generates the entire provider file.
func (g *GenerateTerraformProvider) GenerateTerraformProviderFile(spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile) error {
	// TODO: Uncomment this once support for config and uuid style resouces is added
	// terraformProvider.ImportManager.AddSdkImport(sdkPkgPath(spec), "")

	funcMap := template.FuncMap{
		"renderImports": func() (string, error) { return terraformProvider.ImportManager.RenderImports() },
		"renderCode":    func() string { return terraformProvider.Code.String() },
	}
	return g.generateTerraformEntityTemplate("ProviderFile", &NameProvider{}, spec, terraformProvider, providerFile, funcMap)
}

func (g *GenerateTerraformProvider) GenerateTerraformProvider(terraformProvider *properties.TerraformProviderFile, spec *properties.Normalization, providerConfig properties.TerraformProvider) error {
	terraformProvider.ImportManager.AddStandardImport("context", "")
	terraformProvider.ImportManager.AddSdkImport("github.com/PaloAltoNetworks/pango", "sdk")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/datasource", "")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/provider", "")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/provider/schema", "")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource", "")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/types", "")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-log/tflog", "")

	funcMap := template.FuncMap{
		"RenderImports":  func() (string, error) { return terraformProvider.ImportManager.RenderImports() },
		"DataSources":    func() []string { return terraformProvider.DataSources },
		"Resources":      func() []string { return terraformProvider.Resources },
		"ProviderParams": func() map[string]properties.TerraformProviderParams { return providerConfig.TerraformProviderParams },
		"ParamToModelBasic": func(paramName string, paramProp properties.TerraformProviderParams) (string, error) {
			return ParamToModelBasic(paramName, paramProp)
		},
		"ParamToSchemaProvider": func(paramName string, paramProp properties.TerraformProviderParams) (string, error) {
			return ParamToSchemaProvider(paramName, paramProp)
		},
	}
	return g.generateTerraformEntityTemplate("ProviderFile", &NameProvider{}, spec, terraformProvider, provider, funcMap)
}

func sdkPkgPath(spec *properties.Normalization) string {
	path := fmt.Sprintf("github.com/PaloAltoNetworks/pango/%s", strings.Join(spec.GoSdkPath, "/"))

	return path
}

func conditionallyAddValidators(manager *imports.Manager, spec *properties.Normalization) {

}

func conditionallyAddModifiers(manager *imports.Manager, spec *properties.Normalization) {
	manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier", "")
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
