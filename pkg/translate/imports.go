package translate

import (
	"fmt"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/imports"
)

// RenderImports render string, which contains import required in entry, location or service template.
func RenderImports(templateTypes ...string) (string, error) {
	manager := imports.NewManager()

	var structSDKLocation string
	if len(templateTypes) > 1 && templateTypes[0] == "terraform_provider_file" {
		structSDKLocation = templateTypes[1]
		templateTypes = templateTypes[:1]
	}

	for _, templateType := range templateTypes {
		switch templateType {
		case "entry":
			manager.AddStandardImport("encoding/xml", "")
			manager.AddStandardImport("fmt", "")
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/filtering", "")
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/generic", "")
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/util", "")
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/version", "")
		case "location":
			manager.AddStandardImport("fmt", "")
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/errors", "")
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/util", "")
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/version", "")
		case "service":
			manager.AddStandardImport("context", "")
			manager.AddStandardImport("fmt", "")
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/errors", "")
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/util", "")
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/xmlapi", "")
		case "filtering":
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/filtering", "")
		case "audit":
			manager.AddStandardImport("net/url", "")
			manager.AddStandardImport("strings", "")
			manager.AddStandardImport("time", "")
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/audit", "")
		case "rule":
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/rule", "")
		case "version":
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/version", "")
		case "terraform_provider_file":
			manager.AddStandardImport("context", "")
			manager.AddStandardImport("fmt", "")
			manager.AddStandardImport("regexp", "")
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango", "")
			manager.AddSdkImport(fmt.Sprintf("github.com/PaloAltoNetworks/pango/%s", structSDKLocation), "")
			manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator", "")
			manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework-validators/listvalidator", "")
			manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator", "")
			manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/datasource", "")
			manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/datasource/schema", "dsschema")
			manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/path", "")
			manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource", "")
			manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema", "rsschema")
			manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier", "")
			manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier", "")
			manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault", "")
			manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier", "")
			manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/schema/validator", "")
			manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-framework/types", "")
			manager.AddHashicorpImport("github.com/hashicorp/terraform-plugin-log/tflog", "")
		}
	}

	return manager.RenderImports()
}
