package properties

import (
	"github.com/stretchr/testify/assert"
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
