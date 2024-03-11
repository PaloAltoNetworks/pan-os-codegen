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

func TestSpecifyEntryAssignment(t *testing.T) {
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
	assert.Equal(t, "o.Description", calculatedAssignmentString)
	assert.Equal(t, "util.StrToMem(o.Tags)", calculatedAssignmentListString)
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
