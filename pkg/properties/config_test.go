package properties

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig(t *testing.T) {
	// given
	const content = `output:
  go_sdk: "../generated/pango"
  terraform_provider: "../generated/terraform-provider-panos"
`

	// when
	config, _ := ParseConfig([]byte(content))

	// then
	assert.NotNilf(t, config, "Unmarshalled data cannot be nil")
	assert.NotEmptyf(t, config.Output, "Config output cannot be empty")
	assert.NotEmptyf(t, config.Output.GoSdk, "Config Go SDK path cannot be empty")
	assert.NotEmptyf(t, config.Output.TerraformProvider, "Config Terraform provider path cannot be empty")
}
