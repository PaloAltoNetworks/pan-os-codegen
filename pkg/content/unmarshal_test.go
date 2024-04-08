package content

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalJSONSuccess(t *testing.T) {
	// Given
	var check map[string]interface{}
	jsonInBytes := []byte(`{"key": "value"}`)
	// When
	err := Unmarshal(jsonInBytes, &check)
	// Then
	assert.Nil(t, err)
	assert.Equal(t, check["key"], "value")
}

func TestUnmarshalYAMLSuccess(t *testing.T) {
	// Given
	var check map[string]interface{}
	yamlInBytes := []byte(`key: value`)
	// When
	err := Unmarshal(yamlInBytes, &check)
	// Then
	assert.Nil(t, err)
	assert.Equal(t, check["key"], "value")
}

func TestUnmarshalNoData(t *testing.T) {
	// Given no data
	var check map[string]interface{}
	empty := []byte(``)
	// When
	err := Unmarshal(empty, &check)
	// Then
	assert.NotNil(t, err)
	assert.Equal(t, "no data in file", err.Error())
}
