package location

import (
	validatorschema "github.com/paloaltonetworks/pan-os-codegen/pkg/schema/validator"
	xpathschema "github.com/paloaltonetworks/pan-os-codegen/pkg/schema/xpath"
)

type Device string

const (
	DevicePanorama Device = "panorama"
	DeviceNgfw     Device = "ngfw"
)

type Location struct {
	Name        string                       `yaml:"name"`
	Required    bool                         `yaml:"required"`
	ReadOnly    bool                         `yaml:"read_only"`
	Description string                       `yaml:"description"`
	Devices     []Device                     `yaml:"devices"`
	Xpath       xpathschema.Xpath            `yaml:"xpath"`
	Validators  []*validatorschema.Validator `yaml:"validators"`
}
