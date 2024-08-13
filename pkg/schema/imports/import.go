package imports

import (
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/location"
)

type Import struct {
	Variant   string              `yaml:"variant"`
	Type      string              `yaml:"type"`
	Locations []location.Location `yaml:"locations"`
}
