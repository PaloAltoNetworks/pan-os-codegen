package translate

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/version"
)

const addressSpecPath = "../../specs/objects/address.yaml"

func TestLocationType(t *testing.T) {
	sampleSpec, err := os.ReadFile(addressSpecPath)
	assert.Nil(t, err, "failed to read address spec")
	// given
	yamlParsedData, _ := properties.ParseSpec([]byte(sampleSpec))

	locationKeys := []string{"device-group", "shared"}
	locations := yamlParsedData.Locations
	var locationTypes []string

	// when
	for _, locationKey := range locationKeys {
		locationTypes = append(locationTypes, LocationType(locations[locationKey], true))
	}

	// then
	assert.NotEmpty(t, locationTypes)
	assert.Contains(t, locationTypes, "*DeviceGroupLocation")
	assert.Contains(t, locationTypes, "bool")
}

func TestSpecParamType(t *testing.T) {
	// given
	paramTypeRequiredString := properties.SpecParam{
		Type:     "string",
		Required: true,
	}
	itemsForParam := properties.SpecParamItems{
		Type: "string",
	}
	paramTypeListString := properties.SpecParam{
		Type:  "list",
		Items: &itemsForParam,
	}
	paramTypeOptionalString := properties.SpecParam{
		Type: "string",
	}

	// when
	calculatedTypeRequiredString := SpecParamType("", &paramTypeRequiredString)
	calculatedTypeListString := SpecParamType("", &paramTypeListString)
	calculatedTypeOptionalString := SpecParamType("", &paramTypeOptionalString)

	// then
	assert.Equal(t, "string", calculatedTypeRequiredString)
	assert.Equal(t, "[]string", calculatedTypeListString)
	assert.Equal(t, "*string", calculatedTypeOptionalString)
}

func TestOmitEmpty(t *testing.T) {
	sampleSpec, err := os.ReadFile(addressSpecPath)
	assert.Nil(t, err, "failed to read address spec")

	// given
	yamlParsedData, _ := properties.ParseSpec([]byte(sampleSpec))
	locationKeys := []string{"device-group", "shared"}
	locations := yamlParsedData.Locations
	var omitEmptyLocations []string

	// when
	for _, locationKey := range locationKeys {
		omitEmptyLocations = append(omitEmptyLocations, OmitEmpty(locations[locationKey]))
	}

	// then
	assert.NotEmpty(t, omitEmptyLocations)
	assert.Contains(t, omitEmptyLocations, ",omitempty")
	assert.Contains(t, omitEmptyLocations, "")
}

func TestXmlParamType(t *testing.T) {
	// given
	paramTypeRequiredString := properties.SpecParam{
		Type:     "string",
		Required: true,
		Profiles: []*properties.SpecParamProfile{
			{
				Type:  "string",
				Xpath: []string{"description"},
			},
		},
	}
	paramTypeListString := properties.SpecParam{
		Type: "list",
		Items: &properties.SpecParamItems{
			Type: "string",
		},
		Profiles: []*properties.SpecParamProfile{
			{
				Type:  "member",
				Xpath: []string{"tag"},
			},
		},
	}

	// when
	calculatedTypeRequiredString := XmlParamType("", &paramTypeRequiredString)
	calculatedTypeListString := XmlParamType("", &paramTypeListString)

	// then
	assert.Equal(t, "string", calculatedTypeRequiredString)
	assert.Equal(t, "*util.MemberType", calculatedTypeListString)
}

func TestXmlTag(t *testing.T) {
	// given
	paramTypeRequiredString := properties.SpecParam{
		Type:     "string",
		Required: true,
		Profiles: []*properties.SpecParamProfile{
			{
				Type:  "string",
				Xpath: []string{"description"},
			},
		},
	}
	paramTypeUuid := properties.SpecParam{
		Type:     "string",
		Required: false,
		Name:     properties.NewNameVariant("uuid"),
		Profiles: []*properties.SpecParamProfile{
			{
				Type:  "string",
				Xpath: []string{"uuid"},
			},
		},
	}
	paramTypeListString := properties.SpecParam{
		Type: "list",
		Items: &properties.SpecParamItems{
			Type: "string",
		},
		Profiles: []*properties.SpecParamProfile{
			{
				Type:  "member",
				Xpath: []string{"tag"},
			},
		},
	}

	// when
	calculatedXmlTagRequiredString := XmlTag(&paramTypeRequiredString)
	calculatedXmlTagListString := XmlTag(&paramTypeListString)
	calculatedXmlTagUuid := XmlTag(&paramTypeUuid)

	// then
	assert.Equal(t, "`xml:\"description\"`", calculatedXmlTagRequiredString)
	assert.Equal(t, "`xml:\"tag,omitempty\"`", calculatedXmlTagListString)
	assert.Equal(t, "`xml:\"uuid,attr,omitempty\"`", calculatedXmlTagUuid)
}

func TestNestedSpecs(t *testing.T) {
	// given
	spec := properties.Spec{
		Params: map[string]*properties.SpecParam{
			"a": {
				Name: properties.NewNameVariant("a"),
				Spec: &properties.Spec{
					Params: map[string]*properties.SpecParam{
						"b": {
							Name: properties.NewNameVariant("b"),
							Spec: &properties.Spec{
								Params: map[string]*properties.SpecParam{
									"c": {
										Name: properties.NewNameVariant("c"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// when
	nestedSpecs, _ := NestedSpecs(&spec)

	// then
	assert.NotNil(t, nestedSpecs)
	assert.Contains(t, nestedSpecs, "A")
	assert.Contains(t, nestedSpecs, "AB")
}

func TestCreateGoSuffixFromVersion(t *testing.T) {
	// given

	// when
	version, _ := version.NewVersionFromString("10.1.1")
	suffix := CreateGoSuffixFromVersion(&version)

	// then
	assert.Equal(t, "_10_1_1", suffix)
}

func TestParamSupportedInVersion(t *testing.T) {
	// given
	deviceVersion110, _ := version.NewVersionFromString("11.0.0")
	deviceVersion101, _ := version.NewVersionFromString("10.1.1")
	deviceVersion90, _ := version.NewVersionFromString("9.0.3")

	paramName := properties.NewNameVariant("test")

	profileAlwaysPresent := properties.SpecParamProfile{
		Xpath: []string{"test"},
	}

	minVersion1010, _ := version.NewVersionFromString("10.1.0")
	maxVersion1013, _ := version.NewVersionFromString("10.1.3")
	profilePresentFrom10 := properties.SpecParamProfile{
		Xpath:      []string{"test"},
		MinVersion: &minVersion1010,
		MaxVersion: &maxVersion1013,
	}

	minVersion1100, _ := version.NewVersionFromString("11.0.0")
	maxVersion1110, _ := version.NewVersionFromString("11.1.0")
	profilePresentFrom11 := properties.SpecParamProfile{
		Xpath:      []string{"test"},
		MinVersion: &minVersion1100,
		MaxVersion: &maxVersion1110,
	}

	minVersion901, _ := version.NewVersionFromString("9.0.1")
	maxVersion910, _ := version.NewVersionFromString("9.1.0")
	profileNotPresentFrom10 := properties.SpecParamProfile{
		Xpath:      []string{"test"},
		MinVersion: &minVersion901,
		MaxVersion: &maxVersion910,
	}

	paramPresentFrom10 := &properties.SpecParam{
		Type: "string",
		Name: paramName,
		Profiles: []*properties.SpecParamProfile{
			&profilePresentFrom10,
		},
	}
	paramAlwaysPresent := &properties.SpecParam{
		Type: "string",
		Name: paramName,
		Profiles: []*properties.SpecParamProfile{
			&profileAlwaysPresent,
		},
	}
	paramNotPresentFrom10 := &properties.SpecParam{
		Type: "string",
		Name: paramName,
		Profiles: []*properties.SpecParamProfile{
			&profileNotPresentFrom10,
		},
	}

	paramPresentFrom10And11 := &properties.SpecParam{
		Type: "string",
		Name: paramName,
		Profiles: []*properties.SpecParamProfile{
			&profilePresentFrom10,
			&profilePresentFrom11,
		},
	}

	// when
	noVersionAndParamAlwaysPresent := ParamSupportedInVersion(paramAlwaysPresent, nil)
	noVersionAndParamNotPresentFrom10 := ParamSupportedInVersion(paramNotPresentFrom10, nil)
	device10AndParamPresentFrom10 := ParamSupportedInVersion(paramPresentFrom10, &deviceVersion101)
	device10AndParamAlwaysPresent := ParamSupportedInVersion(paramAlwaysPresent, &deviceVersion101)
	device10AndParamNotPresentFrom10 := ParamSupportedInVersion(paramNotPresentFrom10, &deviceVersion101)
	device9AndParamPresentFrom10 := ParamSupportedInVersion(paramPresentFrom10, &deviceVersion90)
	device9AndParamAlwaysPresent := ParamSupportedInVersion(paramAlwaysPresent, &deviceVersion90)
	device9AndParamNotPresentFrom10 := ParamSupportedInVersion(paramNotPresentFrom10, &deviceVersion90)
	assert.True(t, ParamSupportedInVersion(paramPresentFrom10And11, &deviceVersion110))

	// then
	assert.True(t, noVersionAndParamAlwaysPresent)
	assert.True(t, noVersionAndParamNotPresentFrom10)
	assert.True(t, device10AndParamPresentFrom10)
	assert.True(t, device10AndParamAlwaysPresent)
	assert.False(t, device10AndParamNotPresentFrom10)
	assert.False(t, device9AndParamPresentFrom10)
	assert.True(t, device9AndParamAlwaysPresent)
	assert.True(t, device9AndParamNotPresentFrom10)
}
