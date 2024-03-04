package translate

import (
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStructsDefinitionsForLocation(t *testing.T) {
	// given
	var sampleSpec = `name: 'Address'
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
  'from_panorama':
    description: 'Located in the config pushed from Panorama.'
    read_only: true
    device:
      ngfw: true
    xpath: ['config', 'panorama']
  'vsys':
    description: 'Located in a specific vsys.'
    device:
      panorama: true
      ngfw: true
    xpath:
      - 'config'
      - 'devices'
      - '{{ Entry $ngfw_device }}'
      - 'vsys'
      - '{{ Entry $vsys }}'
    vars:
      'ngfw_device':
        description: 'The NGFW device.'
        default: 'localhost.localdomain'
      'vsys':
        description: 'The vsys.'
        default: 'vsys1'
        validation:
          not_values:
            'shared': 'The vsys cannot be "shared". Use the "shared" path instead.'
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
spec:
  params:
    description:
      description: 'The description.'
      type: 'string'
      length:
        min: 0
        max: 1023
      profiles:
        -
          xpath: ["description"]
    tags:
      description: 'The administrative tags.'
      type: 'list'
      count:
        max: 64
      items:
        type: 'string'
        length:
          max: 127
      profiles:
        -
          type: 'member'
          xpath: ["tag"]
  one_of:
    'ip_netmask':
      description: 'The IP netmask value.'
      profiles:
        -
          xpath: ["ip-netmask"]
    'ip_range':
      description: 'The IP range value.'
      profiles:
        -
          xpath: ["ip-range"]
    'fqdn':
      description: 'The FQDN value.'
      regex: '^[a-zA-Z0-9_]([a-zA-Z0-9:_-])+[a-zA-Z0-9]$'
      length:
        min: 1
        max: 255
      profiles:
        -
          xpath: ["fqdn"]
    'ip_wildcard':
      description: 'The IP wildcard value.'
      profiles:
        -
          xpath: ["ip-wildcard"]
`
	var expectedStructsForLocation = "type Location struct {\n" +
		"\tShared       bool                 `json:\"shared\"`\n" +
		"\tFromPanorama bool                 `json:\"from_panorama\"`\n" +
		"\tVsys         *VsysLocation        `json:\"vsys,omitempty\"`\n" +
		"\tDeviceGroup  *DeviceGroupLocation `json:\"device_group,omitempty\"`\n" +
		"}\n" +
		"\ntype VsysLocation struct {\n" +
		"\tNgfwDevice string  `json:\"ngfw_device\"`\n" +
		"\tVsys       string  `json:\"vsys\"`\n}\n" +
		"\ntype DeviceGroupLocation struct {\n" +
		"\tPanoramaDevice string  `json:\"panorama_device\"`\n" +
		"\tDeviceGroup    string  `json:\"device_group\"`\n" +
		"}\n\n"

	// when
	yamlParsedData, _ := properties.ParseSpec([]byte(sampleSpec))
	structsForLocation, _ := StructsDefinitionsForLocation(yamlParsedData.Locations)

	// then
	assert.Equal(t, expectedStructsForLocation, structsForLocation)
}
