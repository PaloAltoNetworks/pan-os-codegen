package generate

import (
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCopyAssetsWithError(t *testing.T) {
	// Given
	dummyConfig := &properties.Config{
		Assets: map[string]*properties.Asset{
			"tst": {
				Source: "inexistent/source/path",
			}},
	}
	dummyRunType := "sdk"

	// When
	err := CopyAssets(dummyConfig, dummyRunType)

	// Then
	assert.Error(t, err) // we expect an error because the source asset does not exist
}
