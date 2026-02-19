package terraform_provider_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/translate/terraform_provider"
)

func TestParamToModel(t *testing.T) {
	// Given
	paramName := "testParam"
	paramProp := properties.TerraformProviderParams{
		Type: "string",
	}

	// When
	result, err := terraform_provider.ParamToModelBasic(paramName, paramProp)

	// Then
	assert.NoError(t, err)
	assert.Contains(t, result, naming.CamelCase("", paramName, "", true))
	assert.Contains(t, result, "types.String")
	assert.Contains(t, result, "`tfsdk:\"test_param\"`")
}
