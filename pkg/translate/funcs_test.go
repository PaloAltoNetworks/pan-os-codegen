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

func TestSpecifyEntryAssignmentForFlatStructure(t *testing.T) {
	// given
	paramTypeString := properties.SpecParam{
		Name: &properties.NameVariant{
			CamelCase:  "Description",
			Underscore: "description",
		},
		Profiles: []*properties.SpecParamProfile{
			{
				Type:  "string",
				Xpath: []string{"description"},
			},
		},
	}
	paramTypeListString := properties.SpecParam{
		Type: "list",
		Name: &properties.NameVariant{
			CamelCase:  "Tags",
			Underscore: "tags",
		},
		Profiles: []*properties.SpecParamProfile{
			{
				Type:  "member",
				Xpath: []string{"tags"},
			},
		},
	}

	// when
	calculatedAssignmentString := SpecifyEntryAssignment("entry", &paramTypeString, "")
	calculatedAssignmentListString := SpecifyEntryAssignment("entry", &paramTypeListString, "")

	// then
	assert.Equal(t, "entry.Description = o.Description", calculatedAssignmentString)
	assert.Equal(t, "entry.Tags = util.StrToMem(o.Tags)", calculatedAssignmentListString)
}

func TestSpecifyEntryAssignmentForNestedObject(t *testing.T) {
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
									},
								},
							},
						},
					},
				},
			},
		},
	}
	expectedAssignmentStreing := `var nestedA *specAXml
if o.A != nil {
nestedA = &specAXml{}
if _, ok := o.Misc["A"]; ok {
nestedA.Misc = o.Misc["A"]
}
if o.A.B != nil {
nestedA.B = &specABXml{}
if _, ok := o.Misc["AB"]; ok {
nestedA.B.Misc = o.Misc["AB"]
}
if o.A.B.C != nil {
nestedA.B.C = o.A.B.C
}
}
}
entry.A = nestedA
`
	// when
	calculatedAssignmentString := SpecifyEntryAssignment("entry", spec.Params["a"], "")

	// then
	assert.Equal(t, expectedAssignmentStreing, calculatedAssignmentString)
}

func TestNormalizeAssignmentForNestedObject(t *testing.T) {
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
									},
								},
							},
						},
					},
				},
			},
		},
	}
	expectedAssignmentStreing := `var nestedA *SpecA
if o.A != nil {
nestedA = &SpecA{}
if o.A.Misc != nil {
entry.Misc["A"] = o.A.Misc
}
if o.A.B != nil {
nestedA.B = &SpecAB{}
if o.A.B.Misc != nil {
entry.Misc["AB"] = o.A.B.Misc
}
if o.A.B.C != nil {
nestedA.B.C = o.A.B.C
}
}
}
entry.A = nestedA
`
	// when
	calculatedAssignmentString := NormalizeAssignment("entry", spec.Params["a"], "")

	// then
	assert.Equal(t, expectedAssignmentStreing, calculatedAssignmentString)
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
	expectedNestedSpec := `func specMatchAB(a *SpecAB, b *SpecAB) bool {if a == nil && b != nil || a != nil && b == nil {
	return false
} else if a == nil && b == nil {
	return true
}
if !util.StringsMatch(a.C, b.C) {
	return false
}
return true
}
func specMatchA(a *SpecA, b *SpecA) bool {if a == nil && b != nil || a != nil && b == nil {
	return false
} else if a == nil && b == nil {
	return true
}
if !specMatchAB(a.B, b.B) {
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
