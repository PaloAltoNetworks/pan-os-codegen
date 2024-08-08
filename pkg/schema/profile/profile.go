package profile

import "github.com/paloaltonetworks/pan-os-codegen/pkg/version"

// Profile describes parameter versioning information
//
// MinimumVersion and MaximumVersion can be used to limit
// parameter visibility across PAN-OS versions. This is used
// in the marshalling/unmarshalling code to generate per-version
// structures and properly marshal objects into XML documents.
type Profile struct {
	Type           string           `yaml:"type"`
	MinimumVersion *version.Version `yaml:"min_version"`
	MaximumVersion *version.Version `yaml:"max_version"`
	Xpath          []string         `yaml:"xpath"`
}
