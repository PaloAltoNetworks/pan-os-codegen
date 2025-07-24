package xpathschema

import (
	"gopkg.in/yaml.v3"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/validator"
)

type VariableType string

const (
	VariableEntry  VariableType = "entry"
	VariableObject VariableType = "object"
)

// Variable describes a single xpath variable
//
// Xpath variables are used to dynamically render parts of the
// object xpath, using custom logic from pan-os-go utils.
type Variable struct {
	Name           string                `yaml:"name"`
	Description    string                `yaml:"description"`
	Required       bool                  `yaml:"required"`
	LocationFilter bool                  `yaml:"location_filter"`
	Type           VariableType          `yaml:"type"`
	Default        string                `yaml:"default"`
	Validators     []validator.Validator `yaml:"validators"`
}

// Xpath describes xpath as used by locations and imports
//
// Xpath Elements that start with '$' character are variables,
// defined in the Variables field.
type Xpath struct {
	Elements  []string   `yaml:"path"`
	Variables []Variable `yaml:"vars"`
}

// UnmarshalYAML implements unmarshalling with default Type.
//
// This is temporary logic that sets default Type for all variables
// that have no type set explicitly. It can be removed once all
// schemas are generated from the XML source.
func (o *Variable) UnmarshalYAML(n *yaml.Node) error {
	type V Variable

	obj := (*V)(o)
	err := n.Decode(obj)
	if err != nil {
		return err
	}

	if o.Type == "" {
		o.Type = "entry"
	}

	return nil
}
