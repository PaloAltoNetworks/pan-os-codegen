package properties

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	// given
	const content = `output:
  go_sdk: "../generated/pango"
  terraform_provider: "../generated/terraform-provider-panos"
assets:
  util_package:
    source: "assets/util"
    target:
      go_sdk: true
      terraform_provider: false
    destination: "util"
`

	// when
	config, _ := ParseConfig([]byte(content))

	// then
	assert.NotNilf(t, config, "Unmarshalled data cannot be nil")
	assert.NotEmptyf(t, config.Output, "Config output cannot be empty")
	assert.NotEmptyf(t, config.Output.GoSdk, "Config Go SDK path cannot be empty")
	assert.NotEmptyf(t, config.Output.TerraformProvider, "Config Terraform provider path cannot be empty")
	assert.NotEmpty(t, config.Assets)
	assert.Len(t, config.Assets, 1)
	assert.Len(t, config.Assets, 1)
	assert.Equal(t, "assets/util", config.Assets["util_package"].Source)
	assert.True(t, config.Assets["util_package"].Target.GoSdk)
	assert.False(t, config.Assets["util_package"].Target.TerraformProvider)
	assert.Equal(t, "util", config.Assets["util_package"].Destination)
}
