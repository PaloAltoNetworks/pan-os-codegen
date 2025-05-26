package translate

import (
	"github.com/paloaltonetworks/pan-os-codegen/pkg/imports"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

// RenderImports render string, which contains import required in entry, location or service template.
func RenderImports(spec *properties.Normalization, templateTypes ...string) (string, error) {
	manager := imports.NewManager()

	for _, templateType := range templateTypes {
		switch templateType {
		case "sync":
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/locking", "")
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

			if spec.Name == "global-protect-portal" && !spec.HasParametersWithStrconv() {
				panic("WTF?")
			}
			if spec.HasParametersWithStrconv() {
				manager.AddStandardImport("errors", "")
			}
		case "location":
			manager.AddStandardImport("fmt", "")
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/errors", "")
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/util", "")
			manager.AddSdkImport("github.com/PaloAltoNetworks/pango/version", "")
			if spec.ResourceXpathVariablesWithChecks(true) {
				manager.AddStandardImport("strings", "")
			}
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
