package properties

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

const addressSpecPath = "../../specs/objects/address.yaml"

func TestUnmarshallAddressSpecFile(t *testing.T) {
	// given
	sampleSpec, err := os.ReadFile(addressSpecPath)
	assert.Nil(t, err, "failed to read address spec")
	// when
	yamlParsedData, err := ParseSpec([]byte(sampleSpec))
	assert.Nil(t, err)

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
