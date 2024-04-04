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

    "github.com/PaloAltoNetworks/pango/errors"
    "github.com/PaloAltoNetworks/pango/util"
    "github.com/PaloAltoNetworks/pango/version"
)`

	// when
	actualImports, _ := RenderImports("location")

	// then
	assert.NotNil(t, actualImports)
	assert.Equal(t, expectedImports, actualImports)
}
