package validator

import (
	"fmt"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Validator describes any validator that can be defined across
// xpath variables, locations, entries and parameters.
//
// Spec type is any, as its unmarshalling is done in a custom
// UnmarshalYAML function.
type Validator struct {
	Type string `yaml:"type"`
	Spec any    `yaml:"-"`
}

// UnmarshalYAML implements custom unmarshalling logic for parameters
//
// When unmarshalling yaml objects into parameters, their spec is unmarshalled
// into different structures based on the spec type. This custom UnmarshalYAML
//
// For a detailed description of how this is achieved check `parameter.Parameter`
// documentation.
func (v *Validator) UnmarshalYAML(n *yaml.Node) error {
	type V Validator
	type S struct {
		*V   `yaml:",inline"`
		Spec yaml.Node `yaml:"spec"`
	}

	obj := &S{V: (*V)(v)}
	if err := n.Decode(obj); err != nil {
		return err
	}

	switch v.Type {
	case "length":
		v.Spec = new(StringLengthSpec)
	case "values":
		v.Spec = new(ValuesSpec)
	case "not-values":
		v.Spec = new(NotValuesSpec)
	case "range":
		v.Spec = new(RangeSpec)
	case "count":
		v.Spec = new(CountSpec)
	case "regexp":
		v.Spec = new(RegexpSpec)
	case "required":
		v.Spec = new(RequiredSpec)
	default:
		return errors.NewSchemaError(fmt.Sprintf("unsupported validator: '%s'", obj.Type))
	}

	return obj.Spec.Decode(v.Spec)
}

// RequiredSpec validates that a given value is set.
type RequiredSpec struct{}

// StringLengthSpec validates given string's length.
type StringLengthSpec struct {
	Min int
	Max int
}

// ValuesSpec validates that a given value is within validator values.
type ValuesSpec struct {
	Values []any
}

type NotValuesSpecItem struct {
	Value string
	Error string
}

// NotValuesSpec validates that a given value is not withing validator values.
type NotValuesSpec struct {
	Values []NotValuesSpecItem
}

// RangeSpec validates that a given numeric value is within a given range.
type RangeSpec struct {
	Min int
	Max int
}

// CountSpec validates that a given slice length is within a given range.
type CountSpec struct {
	Min int
	Max int
}

// RegexpSpec validates that a given value matches regex expression.
type RegexpSpec struct {
	Expr string
}
