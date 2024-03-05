package translate

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

func TestMakeIndentationEqual(t *testing.T) {
	// given
	givenItems := []string{"test", "a"}
	expectedItems := []string{"test", "a   "}

	// when
	changedItems := MakeIndentationEqual(givenItems)

	// then
	assert.Equal(t, expectedItems, changedItems)
}
