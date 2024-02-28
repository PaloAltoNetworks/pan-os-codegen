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
	unmarshallData, _ := UnmarshallYaml([]byte(fileContent))

	// then
	assert.NotNilf(t, unmarshallData, "Unmarshalled data cannot be nil")
	assert.Equal(t, "Address", unmarshallData.Name, "Unmarshalled data should contain `name`")
}

func TestMarshall(t *testing.T) {
	// given
	parsedData := YamlSpecParser{}
	parsedData.Name = "Address"

	// when
	marshallData, _ := MarshallYaml(&parsedData)

	// then
	fmt.Printf("%s", marshallData)
	assert.NotNilf(t, marshallData, "Marshalled data cannot be nil")
	assert.Containsf(t, marshallData, "name: Address", "Marshalled data should contain key `name`")
}
