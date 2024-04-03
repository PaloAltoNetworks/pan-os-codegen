package translate

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRenderImports(t *testing.T) {
	// given
	expectedImports := `
import (
    "fmt"

    errors "github.com/PaloAltoNetworks/pango/errors"
    util "github.com/PaloAltoNetworks/pango/util"
    version "github.com/PaloAltoNetworks/pango/version"
)`

	// when
	actualImports, _ := RenderImports("location")

	// then
	assert.NotNil(t, actualImports)
	assert.Equal(t, expectedImports, actualImports)
}
