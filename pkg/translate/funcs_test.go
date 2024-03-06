package translate

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAsEntryXpath(t *testing.T) {
	// given

	// when
	asEntryXpath := AsEntryXpath("DeviceGroup", "{{ Entry $panorama_device }}")

	// then
	assert.Equal(t, "util.AsEntryXpath([]string{o.DeviceGroup.PanoramaDevice}),", asEntryXpath)
}
