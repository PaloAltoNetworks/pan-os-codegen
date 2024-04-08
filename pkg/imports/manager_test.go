package imports

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddImport(t *testing.T) {
	// Given
	manager := NewImportManager()
	expectedImport := "github.com/hashicorp"
	expectedShortName := "hc"

	// When
	manager.AddImport(Sdk, expectedImport, expectedShortName)

	// Then
	actualShortName, exists := manager.Imports[Sdk][expectedImport]
	assert.True(t, exists, "Import should exist")
	assert.Equal(t, expectedShortName, actualShortName, "Short name mismatch")
}

func TestRenderImports(t *testing.T) {
	// Given
	manager := NewImportManager()
	expectedImport := "github.com/hashicorp"
	expectedShortName := "hc"
	manager.AddImport(Sdk, expectedImport, expectedShortName)

	// When
	renderedImports, err := manager.RenderImports()

	// Then
	assert.NoError(t, err)
	assert.Contains(t, renderedImports, expectedImport)
	assert.Contains(t, renderedImports, expectedShortName)
}
