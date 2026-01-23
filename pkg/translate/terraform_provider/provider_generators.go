package terraform_provider

import (
	"fmt"
	"log/slog"
	"strings"
	"text/template"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/imports"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

var _ = slog.Debug

// GenerateTerraformProviderFile generates the entire provider file.
func (g *GenerateTerraformProvider) GenerateTerraformProviderFile(resourceTyp properties.ResourceType, spec *properties.Normalization, terraformProvider *properties.TerraformProviderFile) error {
	// TODO: Uncomment this once support for config and uuid style resouces is added
	// terraformProvider.ImportManager.AddSdkImport(sdkPkgPath(spec), "")

	funcMap := template.FuncMap{
		"renderImports": func() (string, error) { return terraformProvider.ImportManager.RenderImports() },
		"renderCode":    func() string { return terraformProvider.Code.String() },
	}
	return g.generateTerraformEntityTemplate(properties.ResourceCustom, properties.SchemaProvider, &NameProvider{}, spec, terraformProvider, "provider/provider_file.tmpl", funcMap)
}

// GenerateTerraformProvider generates the main Terraform provider configuration.
func (g *GenerateTerraformProvider) GenerateTerraformProvider(terraformProvider *properties.TerraformProviderFile, spec *properties.Normalization, providerConfig properties.TerraformProvider) error {
	terraformProvider.ImportManager.AddStandardImport("context", "")
	terraformProvider.ImportManager.AddStandardImport("strings", "")
	terraformProvider.ImportManager.AddStandardImport("fmt", "")
	terraformProvider.ImportManager.AddStandardImport("log/slog", "")
	terraformProvider.ImportManager.AddSdkImport("github.com/PaloAltoNetworks/pango", "sdk")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/datasource", "")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/function", "")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/provider", "")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/action", "")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/provider/schema", "")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource", "")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/ephemeral", "")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/types", "")
	terraformProvider.ImportManager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-log/tflog", "")
	terraformProvider.ImportManager.AddOtherImport("github.com/PaloAltoNetworks/terraform-provider-panos/internal/manager", "sdkmanager")

	funcMap := template.FuncMap{
		"RenderImports":         func() (string, error) { return terraformProvider.ImportManager.RenderImports() },
		"DataSources":           func() []string { return terraformProvider.DataSources },
		"EphemeralResources":    func() []string { return terraformProvider.EphemeralResources },
		"Resources":             func() []string { return terraformProvider.Resources },
		"Actions":               func() []string { return terraformProvider.Actions },
		"RenderResourceFuncMap": func() (string, error) { return RenderResourceFuncMap(terraformProvider.SpecMetadata) },
		"ProviderParams":        func() map[string]properties.TerraformProviderParams { return providerConfig.TerraformProviderParams },
		"ParamToModelBasic": func(paramName string, paramProp properties.TerraformProviderParams) (string, error) {
			return ParamToModelBasic(paramName, paramProp)
		},
		"ParamToSchemaProvider": func(paramName string, paramProp properties.TerraformProviderParams) (string, error) {
			return ParamToSchemaProvider(paramName, paramProp)
		},
	}

	return g.generateTerraformEntityTemplate(properties.ResourceCustom, properties.SchemaProvider, &NameProvider{}, spec, terraformProvider, "provider/provider.tmpl", funcMap)
}

// sdkPkgPath returns the SDK package path for the given specification.
func sdkPkgPath(spec *properties.Normalization) string {
	path := fmt.Sprintf("github.com/PaloAltoNetworks/pango/%s", strings.Join(spec.GoSdkPath, "/"))

	return path
}

// hasVariantsImpl recursively checks if any parameters have variants (enum values or oneOf).
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
		params = append(params, elt.Spec.SortedParams()...)
		params = append(params, elt.Spec.SortedOneOf()...)

		if hasVariantsImpl(params) {
			return true
		}
	}

	return false
}

// conditionallyAddValidators adds validator imports if the spec requires them.
func conditionallyAddValidators(manager *imports.Manager, spec *properties.Normalization) {
	if spec.Spec == nil {
		return
	}

	validatorRequired := func() bool {
		if len(spec.Spec.OneOf) > 0 {
			return true
		}

		var params []*properties.SpecParam
		params = append(params, spec.Spec.SortedParams()...)
		params = append(params, spec.Spec.SortedOneOf()...)
		return hasVariantsImpl(params)
	}

	if validatorRequired() || len(spec.Locations) > 1 {
		manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/path", "")
		manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/schema/validator", "")
	}

}

// conditionallyAddModifiers adds plan modifier imports if the spec has locations.
func conditionallyAddModifiers(manager *imports.Manager, spec *properties.Normalization) {
	if len(spec.Locations) > 0 {
		manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier", "")
	}

	for _, loc := range spec.Locations {
		manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier", "")
		if len(loc.Vars) > 0 {
			manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier", "")
		}
	}
}

// conditionallyAddDefaults recursively adds default value imports for parameters that have defaults.
func conditionallyAddDefaults(manager *imports.Manager, spec *properties.Spec) {
	for _, elt := range spec.SortedParams() {
		if elt.Type == "" {
			conditionallyAddDefaults(manager, elt.Spec)
		} else if elt.Default != "" {
			packageName := fmt.Sprintf("%sdefault", elt.Type)
			fullPackage := fmt.Sprintf("github.com/hashicorp/terraform-plugin-framework/resource/schema/%s", packageName)
			manager.AddHashicorpImport(fullPackage, "")
		}
	}

	for _, elt := range spec.SortedOneOf() {
		if elt.Type == "" {
			conditionallyAddDefaults(manager, elt.Spec)
		} else if elt.Default != "" {
			packageName := fmt.Sprintf("%sdefault", elt.Type)
			fullPackage := fmt.Sprintf("github.com/hashicorp/terraform-plugin-framework/resource/schema/%s", packageName)
			manager.AddHashicorpImport(fullPackage, "")
		}
	}
}
