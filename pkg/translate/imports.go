package translate

import "github.com/paloaltonetworks/pan-os-codegen/pkg/imports"

func RenderImports(templateType string) (string, error) {
	manager := imports.NewManager()

	switch templateType {
	case "entry":
		manager.AddStandardImport("encoding/xml", "")
		manager.AddStandardImport("fmt", "")
		manager.AddSdkImport("github.com/PaloAltoNetworks/pango/filtering", "filtering")
		manager.AddSdkImport("github.com/PaloAltoNetworks/pango/generic", "generic")
		manager.AddSdkImport("github.com/PaloAltoNetworks/pango/util", "util")
		manager.AddSdkImport("github.com/PaloAltoNetworks/pango/version", "version")
	case "location":
		manager.AddStandardImport("fmt", "")
		manager.AddSdkImport("github.com/PaloAltoNetworks/pango/errors", "errors")
		manager.AddSdkImport("github.com/PaloAltoNetworks/pango/util", "util")
		manager.AddSdkImport("github.com/PaloAltoNetworks/pango/version", "version")
	case "service":
		manager.AddStandardImport("context", "")
		manager.AddStandardImport("fmt", "")
		manager.AddSdkImport("github.com/PaloAltoNetworks/pango/errors", "errors")
		manager.AddSdkImport("github.com/PaloAltoNetworks/pango/filtering", "filtering")
		manager.AddSdkImport("github.com/PaloAltoNetworks/pango/util", "util")
		manager.AddSdkImport("github.com/PaloAltoNetworks/pango/xmlapi", "xmlapi")
	}

	return manager.RenderImports()
}
