package translate

import "github.com/paloaltonetworks/pan-os-codegen/pkg/imports"

// RenderImports render string, which contains import required in entry, location or service template.
func RenderImports(templateType string) (string, error) {
	manager := imports.NewImportManager()

	addTemplateImports(templateType, manager)

	return manager.RenderImports()
}

func addTemplateImports(templateType string, manager *imports.ImportManager) {
	switch templateType {
	case "entry":
		manager.AddImport(imports.Standard, "encoding/xml", "")
		manager.AddImport(imports.Standard, "fmt", "")
		manager.AddImport(imports.Sdk, "github.com/PaloAltoNetworks/pango/filtering", "")
		manager.AddImport(imports.Sdk, "github.com/PaloAltoNetworks/pango/generic", "")
		manager.AddImport(imports.Sdk, "github.com/PaloAltoNetworks/pango/util", "")
		manager.AddImport(imports.Sdk, "github.com/PaloAltoNetworks/pango/version", "")
	case "location":
		manager.AddImport(imports.Standard, "fmt", "")
		manager.AddImport(imports.Sdk, "github.com/PaloAltoNetworks/pango/errors", "")
		manager.AddImport(imports.Sdk, "github.com/PaloAltoNetworks/pango/util", "")
		manager.AddImport(imports.Sdk, "github.com/PaloAltoNetworks/pango/version", "")
	case "service":
		manager.AddImport(imports.Standard, "context", "")
		manager.AddImport(imports.Standard, "fmt", "")
		manager.AddImport(imports.Sdk, "github.com/PaloAltoNetworks/pango/errors", "")
		manager.AddImport(imports.Sdk, "github.com/PaloAltoNetworks/pango/filtering", "")
		manager.AddImport(imports.Sdk, "github.com/PaloAltoNetworks/pango/util", "")
		manager.AddImport(imports.Sdk, "github.com/PaloAltoNetworks/pango/xmlapi", "")
	case "filtering":
		manager.AddImport(imports.Sdk,"github.com/PaloAltoNetworks/pango/filtering","")
	case "provider":
		manager.AddImport(imports.Standard, "context", "")
		manager.AddImport(imports.Sdk, "github.com/PaloAltoNetworks/pango", "sdk")
		manager.AddImport(imports.Hashicorp, "github.com/hashicorp/terraform-plugin-framework/datasource", "")
		manager.AddImport(imports.Hashicorp, "github.com/hashicorp/terraform-plugin-framework/provider", "")
		manager.AddImport(imports.Hashicorp, "github.com/hashicorp/terraform-plugin-framework/provider/schema", "")
		manager.AddImport(imports.Hashicorp, "github.com/hashicorp/terraform-plugin-framework/resource", "")
		manager.AddImport(imports.Hashicorp, "github.com/hashicorp/terraform-plugin-framework/types", "")
		manager.AddImport(imports.Hashicorp, "github.com/hashicorp/terraform-plugin-log/tflog", "")
	}
}
