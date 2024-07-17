package translate

import (
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateEntryXpath(t *testing.T) {
	// given

	// when
	asEntryXpath, _ := GenerateEntryXpath("util.AsEntryXpath([]string{", "})", "DeviceGroup", "{{ Entry $panorama_device }}")

	// then
	assert.Equal(t, "util.AsEntryXpath([]string{o.DeviceGroup.PanoramaDevice}),", asEntryXpath)
}

func TestSpecMatchesFunction(t *testing.T) {
	// given
	paramTypeString := properties.SpecParam{
		Name: &properties.NameVariant{
			Underscore: "test",
			CamelCase:  "Test",
		},
		Type: "string",
	}
	paramTypeListString := properties.SpecParam{
		Name: &properties.NameVariant{
			Underscore: "test",
			CamelCase:  "Test",
		},
		Type: "list",
		Items: &properties.SpecParamItems{
			Type: "string",
		},
	}

	// when
	calculatedSpecMatchFunctionString := SpecMatchesFunction(&paramTypeString)
	calculatedSpecMatchFunctionListString := SpecMatchesFunction(&paramTypeListString)

	// then
	assert.Equal(t, "util.StringsMatch", calculatedSpecMatchFunctionString)
	assert.Equal(t, "util.OrderedListsMatch", calculatedSpecMatchFunctionListString)
}

func TestNestedSpecMatchesFunction(t *testing.T) {
	// given
	spec := properties.Spec{
		Params: map[string]*properties.SpecParam{
			"a": {
				Name: &properties.NameVariant{
					Underscore: "a",
					CamelCase:  "A",
				},
				Spec: &properties.Spec{
					Params: map[string]*properties.SpecParam{
						"b": {
							Name: &properties.NameVariant{
								Underscore: "b",
								CamelCase:  "B",
							},
							Spec: &properties.Spec{
								Params: map[string]*properties.SpecParam{
									"c": {
										Name: &properties.NameVariant{
											Underscore: "c",
											CamelCase:  "C",
										},
										Type: "string",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	expectedNestedSpec := `func matchAB(a *AB, b *AB) bool {if a == nil && b != nil || a != nil && b == nil {
	return false
} else if a == nil && b == nil {
	return true
}
if !util.StringsMatch(a.C, b.C) {
	return false
}
return true
}
func matchA(a *A, b *A) bool {if a == nil && b != nil || a != nil && b == nil {
	return false
} else if a == nil && b == nil {
	return true
}
if !matchAB(a.B, b.B) {
	return false
}
return true
}
`

	// when
	renderedNestedSpecMatches := NestedSpecMatchesFunction(&spec)

	// then
	assert.Equal(t, expectedNestedSpec, renderedNestedSpecMatches)
}

func TestXmlPathSuffixes(t *testing.T) {
	// given
	spec := properties.Spec{
		Params: map[string]*properties.SpecParam{
			"a": {
				Profiles: []*properties.SpecParamProfile{{
					Xpath: []string{"test", "a"},
				}},
				Name: &properties.NameVariant{
					Underscore: "a",
					CamelCase:  "A",
				},
			},
		},
	}
	expectedXpathSuffixes := []string{"test/a"}

	// when
	actualXpathSuffixes := XmlPathSuffixes(spec.Params["a"])

	// then
	assert.Equal(t, expectedXpathSuffixes, actualXpathSuffixes)
}
