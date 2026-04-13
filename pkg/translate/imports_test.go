package translate

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

func TestRenderImports(t *testing.T) {
	// given - no location entry xpath vars, util should not be imported
	expectedImports := `
import (
    "fmt"

    "github.com/PaloAltoNetworks/pango/errors"
    "github.com/PaloAltoNetworks/pango/version"
)`

	// when
	spec := &properties.Normalization{
		PanosXpath: properties.PanosXpath{
			Path: []string{"test"},
		},
	}

	actualImports, _ := RenderImports(spec, "location")

	// then
	assert.NotNil(t, actualImports)
	assert.Equal(t, expectedImports, actualImports)
}

func TestRenderImportsWithEntryXpathVars(t *testing.T) {
	// given - location has entry xpath vars, util should be imported
	expectedImports := `
import (
    "fmt"

    "github.com/PaloAltoNetworks/pango/errors"
    "github.com/PaloAltoNetworks/pango/util"
    "github.com/PaloAltoNetworks/pango/version"
)`

	// when
	spec := &properties.Normalization{
		PanosXpath: properties.PanosXpath{
			Path: []string{"test"},
		},
		Locations: map[string]*properties.Location{
			"vsys": {
				Xpath: []string{"config", "devices", "Entry", "vsys", "Entry"},
			},
		},
	}

	actualImports, _ := RenderImports(spec, "location")

	// then
	assert.NotNil(t, actualImports)
	assert.Equal(t, expectedImports, actualImports)
}
