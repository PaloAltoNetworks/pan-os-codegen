package properties

import (
	"strings"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
)

type NameVariant struct {
	Original       string
	Underscore     string
	Dashed         string
	CamelCase      string
	LowerCamelCase string
}

func NewNameVariant(name string) *NameVariant {
	return &NameVariant{
		Original:       name,
		Dashed:         naming.Dashed("", name, ""),
		Underscore:     naming.Underscore("", name, ""),
		CamelCase:      naming.CamelCase("", name, "", true),
		LowerCamelCase: naming.CamelCase("", name, "", false),
	}
}

func (o NameVariant) Components() []string {
	return strings.Split(o.Original, "-")
}

func (o NameVariant) IsEmpty() bool {
	return o.Original == ""
}

func (o NameVariant) WithSuffix(suffix *NameVariant) *NameVariant {
	if o.Original == "" {
		return NewNameVariant(suffix.Original)
	} else {
		return NewNameVariant(o.Original + "-" + suffix.Original)
	}
}

func (o NameVariant) WithLiteralSuffix(suffix string) *NameVariant {
	return &NameVariant{
		Original:       o.Original + suffix,
		CamelCase:      o.CamelCase + suffix,
		LowerCamelCase: o.LowerCamelCase + suffix,
		Underscore:     o.Underscore + suffix,
	}
}
