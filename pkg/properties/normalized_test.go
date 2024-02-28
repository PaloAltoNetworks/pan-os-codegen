package properties

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestUnmarshallAddressSpecFile(t *testing.T) {
	// given

	// when
	yamlParser, _ := ParseSpec("../../specs/objects/address.yaml")

	// then
	assert.NotNilf(t, yamlParser, "Unmarshalled data cannot be nil")
	assert.Equal(t, "Address", yamlParser.Name, "Unmarshalled data should contain `name`")
	assert.Equal(t, "address", yamlParser.TerraformProviderSuffix, "Unmarshalled data should contain `terraform_provider_suffix`")
	assert.NotNilf(t, yamlParser.GoSdkPath, "Unmarshalled data should contain `go_sdk_path`")
	assert.NotNilf(t, yamlParser.XpathSuffix, "Unmarshalled data should contain `xpath_suffix`")
	assert.NotNilf(t, yamlParser.Locations, "Unmarshalled data should contain `locations`")
	assert.NotNilf(t, yamlParser.Entry, "Unmarshalled data should contain `entry`")
	assert.NotNilf(t, yamlParser.Version, "Unmarshalled data should contain `version`")
	assert.NotNilf(t, yamlParser.Spec, "Unmarshalled data should contain `spec`")
}

func TestGetNormalizations(t *testing.T) {
	// given

	// when
	config, _ := GetNormalizations()

	// then
	assert.NotNil(t, config)
	assert.GreaterOrEqual(t, 15, len(config), "Expected to have 15 spec YAML files")
}

func TestSanityCheck(t *testing.T) {
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
