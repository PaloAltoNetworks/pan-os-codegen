package translate

import "github.com/paloaltonetworks/pan-os-codegen/pkg/imports"

// RenderImports render string, which contains import required in entry, location or service template.
func RenderImports(templateTypes ...string) (string, error) {
	manager := imports.NewManager()

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
		}
	}

	return manager.RenderImports()
}
