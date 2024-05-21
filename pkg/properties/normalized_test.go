package properties

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

const sampleSpec = `name: 'Address'
terraform_provider_config:
  suffix: 'address'
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
imports:
  'template':
    xpath:
      - 'config'
      - 'devices'
      - '{{ Entry $ngfw_device }}'
      - 'vsys'
      - '{{ Entry $vsys }}'
      - 'import'
      - 'network'
      - 'interface'
    vars:
      'ngfw_device':
        description: 'The NGFW device.'
        default: 'localhost.localdomain'
      'vsys':
        description: 'The vsys.'
        default: 'vsys1'
    only_for_params:
      - layer3
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
          from_version: "10.1.1"
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
          from_version: "11.1.2"
const:
  color:
    values:
      red:
        value: color1
      light green:
        value: color9
      blue:
        value: color3
`

func TestUnmarshallAddressSpecFile(t *testing.T) {
	// given

	// when
	yamlParsedData, _ := ParseSpec([]byte(sampleSpec))

	// then
	assert.NotNilf(t, yamlParsedData, "Unmarshalled data cannot be nil")
	assert.Equal(t, "Address", yamlParsedData.Name, "Unmarshalled data should contain `name`")
	assert.NotNilf(t, yamlParsedData.TerraformProviderConfig.Suffix, "Unmarshalled data should contain `suffix`")
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
terraform_provider_config:
    skip_resource: false
    skip_datasource: false
    skip_datasource_listing: false
    suffix: address
go_sdk_path:
    - objects
    - address
xpath_suffix:
    - address
locations:
    device_group:
        name:
            underscore: device_group
            camelcase: DeviceGroup
            lowercamelcase: deviceGroup
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
                name:
                    underscore: device_group
                    camelcase: DeviceGroup
                    lowercamelcase: deviceGroup
                description: The device group.
                default: ""
                required: true
                validation:
                    not_values:
                        shared: The device group cannot be "shared". Use the "shared" path instead.
            panorama_device:
                name:
                    underscore: panorama_device
                    camelcase: PanoramaDevice
                    lowercamelcase: panoramaDevice
                description: The panorama device.
                default: localhost.localdomain
                required: false
                validation: null
    from_panorama:
        name:
            underscore: from_panorama
            camelcase: FromPanorama
            lowercamelcase: fromPanorama
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
        name:
            underscore: shared
            camelcase: Shared
            lowercamelcase: shared
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
        name:
            underscore: vsys
            camelcase: Vsys
            lowercamelcase: vsys
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
                name:
                    underscore: ngfw_device
                    camelcase: NgfwDevice
                    lowercamelcase: ngfwDevice
                description: The NGFW device.
                default: localhost.localdomain
                required: false
                validation: null
            vsys:
                name:
                    underscore: vsys
                    camelcase: Vsys
                    lowercamelcase: vsys
                description: The vsys.
                default: vsys1
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
imports:
    template:
        name:
            underscore: template
            camelcase: Template
            lowercamelcase: template
        xpath:
            - config
            - devices
            - '{{ Entry $ngfw_device }}'
            - vsys
            - '{{ Entry $vsys }}'
            - import
            - network
            - interface
        vars:
            ngfw_device:
                name:
                    underscore: ngfw_device
                    camelcase: NgfwDevice
                    lowercamelcase: ngfwDevice
                description: The NGFW device.
                default: localhost.localdomain
            vsys:
                name:
                    underscore: vsys
                    camelcase: Vsys
                    lowercamelcase: vsys
                description: The vsys.
                default: vsys1
        only_for_params:
            - layer3
version: 10.1.0
spec:
    params:
        description:
            name:
                underscore: description
                camelcase: Description
                lowercamelcase: description
            description: The description.
            type: string
            default: ""
            required: false
            length:
                min: 0
                max: 1023
            profiles:
                - xpath:
                    - description
                  not_present: false
                  from_version: 10.1.1
            spec: null
        tags:
            name:
                underscore: tags
                camelcase: Tags
                lowercamelcase: tags
            description: The administrative tags.
            type: list
            default: ""
            required: false
            count:
                min: null
                max: 64
            items:
                type: string
                length:
                    min: null
                    max: 127
                ref: []
            profiles:
                - xpath:
                    - tag
                  type: member
                  not_present: false
                  from_version: ""
            spec: null
    one_of:
        fqdn:
            name:
                underscore: fqdn
                camelcase: Fqdn
                lowercamelcase: fqdn
            description: The FQDN value.
            type: string
            default: ""
            required: false
            length:
                min: 1
                max: 255
            regex: ^[a-zA-Z0-9_]([a-zA-Z0-9:_-])+[a-zA-Z0-9]$
            profiles:
                - xpath:
                    - fqdn
                  not_present: false
                  from_version: ""
            spec: null
        ip_netmask:
            name:
                underscore: ip_netmask
                camelcase: IpNetmask
                lowercamelcase: ipNetmask
            description: The IP netmask value.
            type: string
            default: ""
            required: false
            profiles:
                - xpath:
                    - ip-netmask
                  not_present: false
                  from_version: ""
            spec: null
        ip_range:
            name:
                underscore: ip_range
                camelcase: IpRange
                lowercamelcase: ipRange
            description: The IP range value.
            type: string
            default: ""
            required: false
            profiles:
                - xpath:
                    - ip-range
                  not_present: false
                  from_version: ""
            spec: null
        ip_wildcard:
            name:
                underscore: ip_wildcard
                camelcase: IpWildcard
                lowercamelcase: ipWildcard
            description: The IP wildcard value.
            type: string
            default: ""
            required: false
            profiles:
                - xpath:
                    - ip-wildcard
                  not_present: false
                  from_version: 11.1.2
            spec: null
const:
    color:
        name:
            underscore: color
            camelcase: Color
            lowercamelcase: color
        values:
            blue:
                name:
                    underscore: blue
                    camelcase: Blue
                    lowercamelcase: blue
                value: color3
            light green:
                name:
                    underscore: light_green
                    camelcase: LightGreen
                    lowercamelcase: lightGreen
                value: color9
            red:
                name:
                    underscore: red
                    camelcase: Red
                    lowercamelcase: red
                value: color1
`

	// when
	yamlParsedData, _ := ParseSpec([]byte(sampleSpec))
	yamlDump, _ := yaml.Marshal(&yamlParsedData)

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
	assert.LessOrEqual(t, 15, len(config), "Expected to have 15 spec YAML files")
}

func TestSanity(t *testing.T) {
	// given
	var sampleInvalidSpec = `
name: 'Address'
terraform_provider_config:
    skip_resource: false
    skip_datasource: false
    skip_datasource_listing: false
    suffix: address
go_sdk_path:
  - 'objects'
  - 'address'
xpath_suffix:
  - 'address'
`
	// when
	yamlParsedData := Normalization{}
	err := yaml.Unmarshal([]byte(sampleInvalidSpec), &yamlParsedData)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	err = yamlParsedData.Sanity()

	// then
	assert.ErrorContainsf(t, err, "at least 1 location is required", "error message %s", err)
}

func TestValidation(t *testing.T) {
	// given
	var sampleInvalidSpec = `
name: 'Address'
terraform_provider_config:
 suffix: 'address'
xpath_suffix:
  - 'address'
`
	// when
	yamlParsedData := Normalization{}
	err := yaml.Unmarshal([]byte(sampleInvalidSpec), &yamlParsedData)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	problems := yamlParsedData.Validate()

	// then
	assert.Len(t, problems, 2, "Not all expected validation checks failed")
}

func TestGettingListOfSupportedVersions(t *testing.T) {
	// given
	yamlParsedData, _ := ParseSpec([]byte(sampleSpec))

	// when
	versions := yamlParsedData.SupportedVersions()

	// then
	assert.NotNilf(t, yamlParsedData, "Unmarshalled data cannot be nil")
	assert.Contains(t, versions, "10.1.1")
}

func TestCustomType(t *testing.T) {
	// given

	// when
	yamlParsedData, _ := ParseSpec([]byte(sampleSpec))

	// then
	assert.NotNil(t, yamlParsedData.Const)
	assert.Equal(t, "Red", yamlParsedData.Const["color"].Values["red"].Name.CamelCase)
	assert.Equal(t, "color1", yamlParsedData.Const["color"].Values["red"].Value)
}
