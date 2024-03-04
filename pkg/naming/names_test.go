package naming

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPackageName(t *testing.T) {
	// given
	sampleGoSdkPath := []string{"objects", "address"}

	// when
	packageName := PackageName(sampleGoSdkPath)

	// then
	assert.Equal(t, "address", packageName)
}
