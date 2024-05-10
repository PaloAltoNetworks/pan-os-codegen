package terraform_provider_test

import (
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/translate/terraform_provider"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestParamToModel(t *testing.T) {
	// Given
	paramName := "testParam"
	paramProp := properties.TerraformProviderParams{
		Type: "string",
	}

	// When
	result, err := terraform_provider.ParamToModel(paramName, paramProp)

	// Then
	assert.NoError(t, err)
	assert.Contains(t, result, naming.CamelCase("", paramName, "", true))
	assert.Contains(t, result, "types.String")
	assert.Contains(t, result, "`tfsdk:\"test_param\"`")
}

func TestTFIDStruct(t *testing.T) {
	// Given
	structType := "entry"
	structName := "TestStruct"

	// When
	result, err := terraform_provider.CreateTfIdStruct(structType, structName)

	// Then
	assert.NoError(t, err)
	assert.Contains(t, result, "Name     string          `json:\"name\"`")
	assert.Contains(t, result, "Location TestStruct.Location `json:\"location\"`")
}

func TestCreateNestedStruct(t *testing.T) {
	// Given
	paramName := "nested"
	paramProp := &properties.SpecParam{
		Spec: &properties.Spec{
			Params: map[string]*properties.SpecParam{
				"inner": {Type: "string"},
			},
		},
	}
	structName := "Base"
	nestedStructString := new(strings.Builder)
	createdStructs := make(map[string]bool)

	// When
	result, err := terraform_provider.CreateNestedStruct(paramName, paramProp, structName, nestedStructString, createdStructs)

	// Then
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "BaseNestedObject")
}
