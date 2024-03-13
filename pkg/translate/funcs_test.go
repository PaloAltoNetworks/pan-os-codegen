package translate

import (
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAsEntryXpath(t *testing.T) {
	// given

	// when
	asEntryXpath, _ := AsEntryXpath("DeviceGroup", "{{ Entry $panorama_device }}")

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
	calculatedAssignmentString := SpecifyEntryAssignment(&paramTypeString)
	calculatedAssignmentListString := SpecifyEntryAssignment(&paramTypeListString)

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
	expectedAssignmentStreing := `nestedABC := o.A.B.C
entry.A = &SpecAXml{
B : &SpecABXml{
C : nestedABC,
},
}
`
	// when
	calculatedAssignmentString := SpecifyEntryAssignment(spec.Params["a"])

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
	expectedAssignmentStreing := `nestedABC := o.A.B.C
entry.A = &SpecA{
B : &SpecAB{
C : nestedABC,
},
}
`
	// when
	calculatedAssignmentString := NormalizeAssignment(spec.Params["a"])

	// then
	assert.Equal(t, expectedAssignmentStreing, calculatedAssignmentString)
}

func TestSpecMatchesFunction(t *testing.T) {
	// given
	paramTypeString := properties.SpecParam{
		Type: "string",
	}
	paramTypeListString := properties.SpecParam{
		Type: "list",
	}

	// when
	calculatedSpecMatchFunctionString := SpecMatchesFunction(&paramTypeString)
	calculatedSpecMatchFunctionListString := SpecMatchesFunction(&paramTypeListString)

	// then
	assert.Equal(t, "OptionalStringsMatch", calculatedSpecMatchFunctionString)
	assert.Equal(t, "OrderedListsMatch", calculatedSpecMatchFunctionListString)
}
