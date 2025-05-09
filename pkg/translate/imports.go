package translate

import (
	"github.com/paloaltonetworks/pan-os-codegen/pkg/imports"
)

// RenderImports render string, which contains import required in entry, location or service template.
func RenderImports(templateTypes ...string) (string, error) {
	manager := imports.NewManager()

	for _, templateType := range templateTypes {
		switch templateType {
		case "sync":
			manager.AddStandardImport("sync", "")
		case "config":
			//manager.AddStandardImport("fmt", "")
			manager.AddStandardImport("encoding/xml", "")
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/generic", "")
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/util", "")
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/version", "")
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
			manager.AddStandardImport("encoding/xml", "")
			manager.AddStandardImport("fmt", "")
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/errors", "")
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/util", "")
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/xmlapi", "")
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/xml", "pangoxml")
		case "filtering":
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/filtering", "")
		case "audit":
			manager.AddStandardImport("net/url", "")
			manager.AddStandardImport("strings", "")
			manager.AddStandardImport("time", "")
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/audit", "")
		case "rule":
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/rule", "")
		case "movement":
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/movement", "")
		case "version":
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/version", "")
		case "template":
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/panorama/template", "")
		case "imports":
			manager.AddStandardImport("encoding/xml", "")
		}
	}

	return manager.RenderImports()
}
