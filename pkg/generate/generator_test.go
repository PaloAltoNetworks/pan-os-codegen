package generate

import (
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateFullFilePath(t *testing.T) {
	// given
	spec := properties.Normalization{
		GoSdkPath: []string{"go"},
	}
	generator := NewCreator("test", "templates/sdk", &spec)

	// when
	fullFilePath := generator.createFullFilePath("output", &spec, "template.tmpl")

	// then
	assert.NotNil(t, fullFilePath)
	assert.Equal(t, "output/go/template.go", fullFilePath)
}
