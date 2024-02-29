package properties

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

const content = `name: 'Address'
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

func TestUnmarshallAddressSpecFile(t *testing.T) {
	// given

	// when
	yamlParsedData, _ := ParseSpec([]byte(content))

	// then
	assert.NotNilf(t, yamlParsedData, "Unmarshalled data cannot be nil")
	assert.Equal(t, "Address", yamlParsedData.Name, "Unmarshalled data should contain `name`")
	assert.Equal(t, "address", yamlParsedData.TerraformProviderSuffix, "Unmarshalled data should contain `terraform_provider_suffix`")
	assert.NotNilf(t, yamlParsedData.GoSdkPath, "Unmarshalled data should contain `go_sdk_path`")
	assert.NotNilf(t, yamlParsedData.XpathSuffix, "Unmarshalled data should contain `xpath_suffix`")
	assert.NotNilf(t, yamlParsedData.Locations, "Unmarshalled data should contain `locations`")
	assert.NotNilf(t, yamlParsedData.Entry, "Unmarshalled data should contain `entry`")
	assert.NotNilf(t, yamlParsedData.Version, "Unmarshalled data should contain `version`")
	assert.NotNilf(t, yamlParsedData.Spec, "Unmarshalled data should contain `spec`")
}

func TestMarshallAddressSpecFile(t *testing.T) {
	// given
	var expectedMarshalledData = `name: Address
terraform_provider_suffix: address
go_sdk_path:
    - objects
    - address
xpath_suffix:
    - address
locations:
    device_group:
        description: Located in a specific device group.
        device:
            panorama: true
            ngfw: false
        xpath:
            - config
            - devices
            - '{{ Entry $panorama_device }}'
            - device-group
            - '{{ Entry $device_group }}'
        read_only: false
        vars:
            device_group:
                description: The device group.
                required: true
                validation:
                    not_values:
                        shared: The device group cannot be "shared". Use the "shared" path instead.
            panorama_device:
                description: The panorama device.
                required: false
                validation: null
    from_panorama:
        description: Located in the config pushed from Panorama.
        device:
            panorama: false
            ngfw: true
        xpath:
            - config
            - panorama
        read_only: true
        vars: {}
    shared:
        description: Located in shared.
        device:
            panorama: true
            ngfw: true
        xpath:
            - config
            - shared
        read_only: false
        vars: {}
    vsys:
        description: Located in a specific vsys.
        device:
            panorama: true
            ngfw: true
        xpath:
            - config
            - devices
            - '{{ Entry $ngfw_device }}'
            - vsys
            - '{{ Entry $vsys }}'
        read_only: false
        vars:
            ngfw_device:
                description: The NGFW device.
                required: false
                validation: null
            vsys:
                description: The vsys.
                required: false
                validation:
                    not_values:
                        shared: The vsys cannot be "shared". Use the "shared" path instead.
entry:
    name:
        description: The name of the address object.
        length:
            min: 1
            max: 63
version: 10.1.0
spec:
    params:
        description:
            description: The description.
            type: string
            length:
                min: 0
                max: 1023
            profiles:
                - xpath:
                    - description
            spec: null
        tags:
            description: The administrative tags.
            type: list
            count:
                min: null
                max: 64
            items:
                type: string
                length:
                    min: null
                    max: 127
            profiles:
                - xpath:
                    - tag
                  type: member
            spec: null
    one_of:
        fqdn:
            description: The FQDN value.
            type: ""
            length:
                min: 1
                max: 255
            regex: ^[a-zA-Z0-9_]([a-zA-Z0-9:_-])+[a-zA-Z0-9]$
            profiles:
                - xpath:
                    - fqdn
            spec: null
        ip_netmask:
            description: The IP netmask value.
            type: ""
            profiles:
                - xpath:
                    - ip-netmask
            spec: null
        ip_range:
            description: The IP range value.
            type: ""
            profiles:
                - xpath:
                    - ip-range
            spec: null
        ip_wildcard:
            description: The IP wildcard value.
            type: ""
            profiles:
                - xpath:
                    - ip-wildcard
            spec: null
`

	// when
	yamlParsedData, _ := ParseSpec([]byte(content))
	yamlDump, _ := yaml.Marshal(&yamlParsedData)
	//fmt.Printf("%s", string(yamlDump))

	// then
	assert.NotNilf(t, yamlDump, "Marshalled data cannot be nil")
	assert.Equal(t, expectedMarshalledData, string(yamlDump), "Marshalled data differs from expected")
}

func TestGetNormalizations(t *testing.T) {
	// given

	// when
	config, _ := GetNormalizations()

	// then
	assert.NotNil(t, config)
	assert.GreaterOrEqual(t, 15, len(config), "Expected to have 15 spec YAML files")
}

func TestSanity(t *testing.T) {
	// given
	var fileContent = `
name: 'Address'
terraform_provider_suffix: 'address'
go_sdk_path:
  - 'objects'
  - 'address'
xpath_suffix:
  - 'address'
`
	// when
	yamlParsedData := Normalization{}
	err := yaml.Unmarshal([]byte(fileContent), &yamlParsedData)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	err = yamlParsedData.Sanity()

	// then
	assert.ErrorContainsf(t, err, "at least 1 location is required", "error message %s", err)
}

func TestValidation(t *testing.T) {
	// given
	var fileContent = `
name: 'Address'
terraform_provider_suffix: 'address'
xpath_suffix:
  - 'address'
`
	// when
	yamlParsedData := Normalization{}
	err := yaml.Unmarshal([]byte(fileContent), &yamlParsedData)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	problems := yamlParsedData.Validate()

	// then
	assert.Len(t, problems, 2, "Not all expected validation checks failed")
}
