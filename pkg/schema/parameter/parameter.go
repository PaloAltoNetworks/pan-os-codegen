package parameter

import (
	"fmt"
	"log/slog"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/errors"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/profile"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/validator"
)

// Parameter describes a single parameter for the given object spec.
//
// Parameter can describe both a top-level parameter of the object
// or nested parameter of another object parameter.
//
// Spec type is any, as its unmarshalling is done in a custom
// UnmarshalYAML function.
type Parameter struct {
	Name             string                `yaml:"name"`
	Description      string                `yaml:"description"`
	Type             string                `yaml:"type"`
	CodegenOverrides *CodegenOverrides     `yaml:"codegen_overrides"`
	Hashing          *HashingSpec          `yaml:"hashing"`
	Required         bool                  `yaml:"required"`
	Profiles         []profile.Profile     `yaml:"profiles"`
	Validators       []validator.Validator `yaml:"validators"`
	Spec             any                   `yaml:"-"`
}

// SimpleSpec describes a parameter of a simple type.
type SimpleSpec struct {
	Default  any  `yaml:"default"`
	Required bool `yaml:"required"`
}

// EnumSpecValue describes a single enum value.
type EnumSpecValue struct {
	Value string `yaml:"value"`
	Const string `yaml:"const"`
}

type VariantCheckType string

const (
	VariantCheckConflictsWith VariantCheckType = "ConflictsWith"
	VariantCheckExactlyOneOf  VariantCheckType = "ExactlyOneOf"
)

type CodegenOverridesGoSdk struct {
	Skip *bool `yaml:"skip"`
}

type CodegenOverridesTerraform struct {
	Name         *string           `yaml:"name"`
	Type         *string           `yaml:"type"`
	Private      *bool             `yaml:"private"`
	Sensitive    *bool             `yaml:"sensitive"`
	Computed     *bool             `yaml:"computed"`
	Required     *bool             `yaml:"required"`
	VariantCheck *VariantCheckType `yaml:"variant_check"`
}

type CodegenOverrides struct {
	GoSdk     CodegenOverridesGoSdk     `yaml:"gosdk"`
	Terraform CodegenOverridesTerraform `yaml:"terraform"`
}

type HashingSpec struct {
	Type string `yaml:"type"`
}

// EnumSpec describes a parameter of type enum
//
// Values is a list of EnumSpecValue, where each one consisting of the PAN-OS Value
// and its optional Const representation. This allows to generate a more meaningful
// types, for example when spec value is "color1", and its const is "red", the following
// type will be marshalled for pan-os-go SDK:
//
//	ParameterRed ParameterType = "color1"
//
// when spec value is "up" and spec const is empty, the following type will be marshalled
// instead:
//
//	ParameterUp ParameterType = "up"
type EnumSpec struct {
	Required bool            `yaml:"required"`
	Default  string          `yaml:"default"`
	Values   []EnumSpecValue `yaml:"values"`
}

type StructSpec struct {
	Required   bool         `yaml:"required"`
	Parameters []*Parameter `yaml:"params"`
	Variants   []*Parameter `yaml:"variants"`
}

type ListSpecElement struct {
	Type     string     `yaml:"type"`
	Required bool       `yaml:"required"`
	Spec     StructSpec `yaml:"spec"`
}

type ListSpec struct {
	Required bool            `yaml:"required"`
	Items    ListSpecElement `yaml:"items"`
}

type NilSpec struct{}

// UnmarshalYAML implements custom unmarshalling logic for parameters
//
// When unmarshalling yaml objects into parameters, their spec is unmarshalled
// into different structures based on the spec type. This custom UnmarshalYAML
// functions handles this logic.
func (p *Parameter) UnmarshalYAML(n *yaml.Node) error {
	// Create an empty Parameter value and a new temporary structure
	// that is based on the parameter, but overrides Spec field to be
	// of a generic yaml.Node type.
	// This new type, and the casting below is needed so that yaml
	// unmarshaller doesn't recursively call UnmarshalYAML.
	type P Parameter
	type S struct {
		*P   `yaml:",inline"`
		Spec yaml.Node `yaml:"spec"`
	}

	// Cast "this" parameter pointer to a new S type and then decode
	// entire parameter yaml object into this new object. This will
	// unmarshal spec field from yaml into yaml.Node Spec field in the
	// temporary structure.
	obj := &S{P: (*P)(p)}
	if err := n.Decode(obj); err != nil {
		return err
	}

	// Now that we have unmarshalled entire parameter object into a temporary
	// structure, we can assign proper structure to the parameter Spec field.
	switch p.Type {
	case "object":
		p.Spec = new(StructSpec)
	case "list":
		p.Spec = new(ListSpec)
	case "enum":
		p.Spec = new(EnumSpec)
	case "nil":
		p.Spec = new(NilSpec)
	case "string", "bool", "int64", "float64":
		p.Spec = new(SimpleSpec)
	default:
		return errors.NewSchemaError(fmt.Sprintf("unsupported parameter type: '%s'", p.Type))
	}

	// Escape value of a description, making sure all backslashes are handled properly
	obj.Description = strings.ReplaceAll(obj.Description, "\\", "\\\\")

	// Finally, decode obj.Spec (which is yaml.Node type) into the parameter
	// spec structure
	return obj.Spec.Decode(p.Spec)
}

// SingularName returns a singular name for parameter.
//
// When Parameter type is list, and list profile type is either
// "entry" or "member", we us first path of an profile xpath array
// to determine a singular name for the given parameter.
//
// When called for non-list parameters, parameter name is returned
// instead.
func (p *Parameter) SingularName() string {
	switch p.Spec.(type) {
	case *ListSpec:
		var singularName string
		for _, profile := range p.Profiles {
			switch profile.Type {
			case "entry", "member":
				if len(profile.Xpath) >= 1 {
					singularName = profile.Xpath[0]
				}
			}
		}

		if singularName == "" {
			slog.Warn("Couldn't generate singular name for list parameter", "parameter", p.Name)
		}

		return singularName
	}

	return p.Name
}

// SpecItemsType is a shorthand accessor to list items type
//
// When checking parameter list item type, parameter's spec has to
// be first cast to the ListSpec type. This function gives
// a quick access to the type without having to do casting manually.
//
// When called on any other parameter type, it returns an empty
// string, so the caller (mostly templates) must ensure that it's
// being called on list parameters.
func (p *Parameter) SpecItemsType() string {
	switch spec := p.Spec.(type) {
	case *ListSpec:
		return spec.Items.Type
	default:
		return ""
	}
}
