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
	assert.Equal(t, 4, len(unmarshallData), "Unmarshalled data should contain 4 keys")
	assert.Equal(t, "Address", unmarshallData["name"], "Unmarshalled data should contain `name`")
}

func TestMarshall(t *testing.T) {
	// given
	parsedData := make(map[interface{}]interface{})
	parsedData["name"] = "Address"

	// when
	marshallData, _ := MarshallYaml(parsedData)

	// then
	fmt.Printf("%s", marshallData)
	assert.NotNilf(t, marshallData, "Marshalled data cannot be nil")
	assert.Equal(t, "name: Address\n", marshallData, "Marshalled data should contain 1 key `name`")
}
