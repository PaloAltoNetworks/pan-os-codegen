package terraform

import (
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestGenerateTerraformResourceValidInput tests the successful generation of a Terraform resource.
func TestGenerateTerraformResourceValidInput(t *testing.T) {
	// Given
	spec := &properties.Normalization{
		Name: "example",
	}
	terraformProvider := &properties.TerraformProviderFile{
		Resources: []string{},
	}

	gtp := GenerateTerraformProvider{}

	// When
	err := gtp.GenerateTerraformResource(spec, terraformProvider)

	// Then
	assert.NoError(t, err, "GenerateTerraformResource should not return an error")
	assert.Contains(t, terraformProvider.Resources, "ExampleResource", "The resource name should be added to the Terraform provider resources")
}
