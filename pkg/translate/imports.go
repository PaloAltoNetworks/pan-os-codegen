package translate

import "github.com/paloaltonetworks/pan-os-codegen/pkg/imports"

func RenderImports(templateType string) (string, error) {
	manager := imports.NewManager()

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
		manager.AddSdkImport("github.com/PaloAltoNetworks/pango/filtering", "")
		manager.AddSdkImport("github.com/PaloAltoNetworks/pango/util", "")
		manager.AddSdkImport("github.com/PaloAltoNetworks/pango/xmlapi", "")
	}

	return manager.RenderImports()
}
