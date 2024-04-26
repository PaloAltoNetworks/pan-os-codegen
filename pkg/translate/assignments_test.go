package translate

import (
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/stretchr/testify/assert"
	"testing"
)

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

func TestNestedVariableNameWithoutEntry(t *testing.T) {
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
	expectedNestedVariableName := `A.B.C`
	params := []*properties.SpecParam{}
	params = append(params,
		spec.Params["a"],
		spec.Params["a"].Spec.Params["b"],
		spec.Params["a"].Spec.Params["b"].Spec.Params["c"],
	)

	// when
	calculatedNestedVariableName := renderNestedVariableName(params, true, true, false)

	// then
	assert.Equal(t, expectedNestedVariableName, calculatedNestedVariableName)
}

func TestNestedVariableNameWithEntry(t *testing.T) {
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
							Type: "list",
							Items: &properties.SpecParamItems{
								Type: "entry",
							},
							Profiles: []*properties.SpecParamProfile{
								{
									Xpath: []string{"test", "entry"},
									Type:  "entry",
								},
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
	expectedNestedVariableName := `AB.C`
	params := []*properties.SpecParam{}
	params = append(params,
		spec.Params["a"],
		spec.Params["a"].Spec.Params["b"],
		spec.Params["a"].Spec.Params["b"].Spec.Params["c"],
	)

	// when
	calculatedNestedVariableName := renderNestedVariableName(params, true, true, false)

	// then
	assert.Equal(t, expectedNestedVariableName, calculatedNestedVariableName)
}
