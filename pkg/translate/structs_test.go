package translate

import (
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/stretchr/testify/assert"
	"testing"
)

const sampleSpec = `name: 'Address'
terraform_provider_suffix: 'address'
go_sdk_path:
  - 'objects'
  - 'address'
xpath_suffix:
  - 'address'
locations:
  'shared':
    description: 'Located in shared.'
    device:
      panorama: true
      ngfw: true
    xpath: ['config', 'shared']
  'device_group':
    description: 'Located in a specific device group.'
    device:
      panorama: true
    xpath:
      - 'config'
      - 'devices'
      - '{{ Entry $panorama_device }}'
      - 'device-group'
      - '{{ Entry $device_group }}'
    vars:
      'panorama_device':
        description: 'The panorama device.'
        default: 'localhost.localdomain'
      'device_group':
        description: 'The device group.'
        required: true
        validation:
          not_values:
            'shared': 'The device group cannot be "shared". Use the "shared" path instead.'
entry:
  name:
    description: 'The name of the address object.'
    length:
      min: 1
      max: 63
version: '10.1.0'
`

func TestLocationType(t *testing.T) {
	// given
	yamlParsedData, _ := properties.ParseSpec([]byte(sampleSpec))
	locationKeys := []string{"device_group", "shared"}
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
	calculatedTypeRequiredString := SpecParamType(&paramTypeRequiredString)
	calculatedTypeListString := SpecParamType(&paramTypeListString)
	calculatedTypeOptionalString := SpecParamType(&paramTypeOptionalString)

	// then
	assert.Equal(t, "string", calculatedTypeRequiredString)
	assert.Equal(t, "[]string", calculatedTypeListString)
	assert.Equal(t, "*string", calculatedTypeOptionalString)
}

func TestOmitEmpty(t *testing.T) {
	// given
	yamlParsedData, _ := properties.ParseSpec([]byte(sampleSpec))
	locationKeys := []string{"device_group", "shared"}
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
	calculatedTypeRequiredString := XmlParamType(&paramTypeRequiredString)
	calculatedTypeListString := XmlParamType(&paramTypeListString)

	// then
	assert.Equal(t, "string", calculatedTypeRequiredString)
	assert.Equal(t, "*util.MemberType", calculatedTypeListString)
}

func TestXmlTag(t *testing.T) {
	// given
	paramTypeRequiredString := properties.SpecParam{
		Type:     "string",
		Required: false,
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
	calculatedXmlTagRequiredString := XmlTag(&paramTypeRequiredString)
	calculatedXmlTagListString := XmlTag(&paramTypeListString)

	// then
	assert.Equal(t, "`xml:\"description,omitempty\"`", calculatedXmlTagRequiredString)
	assert.Equal(t, "`xml:\"tag,omitempty\"`", calculatedXmlTagListString)
}
