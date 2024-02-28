package parsing

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnmarshall(t *testing.T) {
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
	yamlParser, _ := NewYamlSpecParser([]byte(fileContent))

	// then
	assert.NotNilf(t, yamlParser, "Unmarshalled data cannot be nil")
	assert.Equal(t, "Address", yamlParser.Name, "Unmarshalled data should contain `name`")
}

func TestMarshall(t *testing.T) {
	// given
	yamlParser := YamlSpecParser{}
	yamlParser.Name = "Address"

	// when
	dumpData, _ := yamlParser.Dump()

	// then
	fmt.Printf("%s", dumpData)
	assert.NotNilf(t, dumpData, "Marshalled data cannot be nil")
	assert.Containsf(t, dumpData, "name: Address", "Marshalled data should contain key `name`")
}
