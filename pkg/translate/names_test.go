package translate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackageName(t *testing.T) {
	// given
	sampleGoSdkPath := []string{"objects", "address"}

	// when
	packageName := PackageName(sampleGoSdkPath)

	// then
	assert.Equal(t, "address", packageName)
}
