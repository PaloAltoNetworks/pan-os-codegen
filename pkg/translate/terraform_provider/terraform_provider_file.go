package terraform_provider

import (
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

// NameProvider encapsulates naming conventions for Terraform resources.
type NameProvider struct {
	TfName     string
	MetaName   string
	StructName string
}

// NewNameProvider creates a new NameProvider based on given specifications.
func NewNameProvider(spec *properties.Normalization, resourceName string) *NameProvider {
	tfName := spec.Name
	metaName := fmt.Sprintf("_%s", naming.Underscore("", strings.ToLower(tfName), ""))
	structName := naming.CamelCase("", tfName, resourceName, true)
	return &NameProvider{tfName, metaName, structName}
}

// GenerateTerraformProvider handles the generation of Terraform resources and data sources.
type GenerateTerraformProvider struct{}

// createTemplate parses the provided template string using the given FuncMap and returns a new template.
func (g *GenerateTerraformProvider) createTemplate(resourceType string, spec *properties.Normalization, templateStr string, funcMap template.FuncMap) (*template.Template, error) {
	templateName := fmt.Sprintf("terraform-%s-%s", resourceType, spec.Name)
	return template.New(templateName).Funcs(funcMap).Parse(templateStr)
}

// executeTemplate executes the provided resource template using the given spec and returns an error if it fails.
func (g *GenerateTerraformProvider) executeTemplate(resourceTemplate *template.Template, spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile, resourceType string, names *NameProvider) error {
	var renderedTemplate strings.Builder
	if err := resourceTemplate.Execute(&renderedTemplate, spec); err != nil {
		log.Fatalf("Error executing template: %v", err)
		return fmt.Errorf("error executing %s template: %v", resourceType, err)
	}
	renderedTemplate.WriteString("\n")
	return g.updateProviderFile(&renderedTemplate, terraformProvider, resourceType, names.StructName)
}

// updateProviderFile updates the Terraform provider file by appending the rendered template
// to the appropriate slice in the TerraformProviderFile based on the provided resourceType.
func (g *GenerateTerraformProvider) updateProviderFile(renderedTemplate *strings.Builder, terraformProvider *properties.TerraformProviderFile, resourceType string, structName string) error {
	switch resourceType {
	case "ProviderFile":
		terraformProvider.Code = renderedTemplate
	default:
		if _, err := terraformProvider.Code.WriteString(renderedTemplate.String()); err != nil {
			return fmt.Errorf("error writing %s template: %v", resourceType, err)
		}
	}
	return g.appendResourceType(terraformProvider, resourceType, structName)
}

// appendResourceType appends the given struct name to the appropriate slice in the TerraformProviderFile
// based on the provided resourceType.
func (g *GenerateTerraformProvider) appendResourceType(terraformProvider *properties.TerraformProviderFile, resourceType string, structName string) error {
	switch resourceType {
	case "DataSource", "DataSourceList":
		terraformProvider.DataSources = append(terraformProvider.DataSources, structName)
	case "Resource":
		terraformProvider.Resources = append(terraformProvider.Resources, structName)
	}
	return nil
}

// generateTerraformEntityTemplate is the common logic for generating Terraform resources and data sources.
func (g *GenerateTerraformProvider) generateTerraformEntityTemplate(resourceType string, names *NameProvider, spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile, templateStr string, funcMap template.FuncMap) error {
	resourceTemplate, err := g.createTemplate(resourceType, spec, templateStr, funcMap)
	if err != nil {
		log.Fatalf("Error creating template: %v", err)
		return err
	}
	return g.executeTemplate(resourceTemplate, spec, terraformProvider, resourceType, names)
}

// GenerateTerraformResource generates a Terraform resource template.
func (g *GenerateTerraformProvider) GenerateTerraformResource(spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile) error {
	resourceType := "Resource"
	names := NewNameProvider(spec, resourceType)
	funcMap := template.FuncMap{
		"metaName":                func() string { return names.MetaName },
		"structName":              func() string { return names.StructName },
		"CreateTfIdStruct":        func() (string, error) { return CreateTfIdStruct("entry", spec.GoSdkPath[len(spec.GoSdkPath)-1]) },
		"CreateTfIdResourceModel": func() (string, error) { return CreateTfIdResourceModel("entry", names.StructName) },
		"ParamToModelResource": func(paramName string, paramProp *properties.SpecParam, structName string) (string, error) {
			return ParamToModelResource(paramName, paramProp, structName)
		},
		"ModelNestedStruct": func(paramName string, paramProp *properties.SpecParam, structName string) (string, error) {
			return ModelNestedStruct(paramName, paramProp, structName)
		},
	}
	terraformProvider.ImportManager.AddStandardImport("context", "")
	terraformProvider.ImportManager.AddStandardImport("fmt", "")
	terraformProvider.ImportManager.AddSdkImport("github.com/PaloAltoNetworks/pango", "")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema", "rsschema")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource", "")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/types", "")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/path", "")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-log/tflog", "")
	if !spec.TerraformProviderConfig.SkipResource {
		if spec.Entry != nil {
			if spec.Spec == nil || spec.Spec.Params == nil {
				return fmt.Errorf("invalid resource configuration")
			}
			_, uuid := spec.Spec.Params["uuid"]
			if uuid {
				// Generate Resource with uuid style
				err := g.generateTerraformEntityTemplate(resourceType, names, spec, terraformProvider, resourceTemplateStr, funcMap)
				if err != nil {
					return err
				}
			} else {
				// Generate Resource with entry style
				err := g.generateTerraformEntityTemplate(resourceType, names, spec, terraformProvider, resourceTemplateStr, funcMap)
				if err != nil {
					return err
				}
			}
		} else {
			// Generate Resource with config style
			err := g.generateTerraformEntityTemplate(resourceType, names, spec, terraformProvider, resourceTemplateStr, funcMap)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GenerateTerraformDataSource generates a Terraform data source and data source template.
func (g *GenerateTerraformProvider) GenerateTerraformDataSource(spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile) error {

	if !spec.TerraformProviderConfig.SkipDatasource {
		terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/datasource/schema", "dsschema")
		terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/datasource", "")

		resourceType := "DataSource"
		names := NewNameProvider(spec, resourceType)
		funcMap := template.FuncMap{
			"metaName":   func() string { return names.MetaName },
			"structName": func() string { return names.StructName },
		}
		err := g.generateTerraformEntityTemplate(resourceType, names, spec, terraformProvider, dataSourceSingletonTemplateStr, funcMap)
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
	//	err := g.generateTerraformEntityTemplate(resourceType, names, spec, terraformProvider, dataSourceListTemplatetStr, funcMap)
	//	if err != nil {
	//		return err
	//	}
	//}

	return nil
}

// GenerateTerraformProviderFile generates the entire provider file.
func (g *GenerateTerraformProvider) GenerateTerraformProviderFile(spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile) error {
	terraformProvider.ImportManager.AddSdkImport(fmt.Sprintf("github.com/PaloAltoNetworks/pango/%s", strings.Join(spec.GoSdkPath, "/")), "")

	funcMap := template.FuncMap{
		"renderImports": func() (string, error) { return terraformProvider.ImportManager.RenderImports() },
		"renderCode":    func() string { return terraformProvider.Code.String() },
	}
	return g.generateTerraformEntityTemplate("ProviderFile", &NameProvider{}, spec, terraformProvider, providerFileTemplateStr, funcMap)
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
		"ParamToModel": func(paramName string, paramProp properties.TerraformProviderParams) (string, error) {
			return ParamToModel(paramName, paramProp)
		},
		"ParamToSchema": func(paramName string, paramProp properties.TerraformProviderParams) (string, error) {
			return ParamToSchema(paramName, paramProp)
		},
	}
	return g.generateTerraformEntityTemplate("ProviderFile", &NameProvider{}, spec, terraformProvider, providerTemplateStr, funcMap)
}