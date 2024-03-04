package generate

import (
	"bytes"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateFullFilePath(t *testing.T) {
	// given
	spec := properties.Normalization{
		GoSdkPath: []string{"go"},
	}
	generator := NewCreator("test", "../../templates/sdk", &spec)

	// when
	fullFilePath := generator.createFullFilePath("output", &spec, "template.tmpl")

	// then
	assert.NotNil(t, fullFilePath)
	assert.Equal(t, "output/go/template.go", fullFilePath)
}

// NOTE - unit tests should only touch code inside package, do not reference external resources.
// Nevertheless, below tests ARE REFERENCING to text templates, because template ARE NOT EMBEDDED into Go files.
// Technically we could embed them, but then:
// 1 - we are losing clarity of the code,
// 2 - we are mixing Golang with templates expressions.
// Testing generator is crucial, so below tests we can br treated more as integration tests, not unit one.

func TestListOfTemplates(t *testing.T) {
	// given
	spec := properties.Normalization{
		GoSdkPath: []string{"go"},
	}
	generator := NewCreator("test", "../../templates/sdk", &spec)

	// when
	var templates []string
	templates, _ = generator.listOfTemplates(templates)

	// then
	assert.Equal(t, 4, len(templates))
}

func TestParseTemplate(t *testing.T) {
	// given
	spec := properties.Normalization{
		GoSdkPath: []string{"object", "address"},
	}
	generator := NewCreator("test", "../../templates/sdk", &spec)
	expectedFileContent := `package address

type Specifier func(Entry) (any, error)

type Normalizer interface {
    Normalize() ([]Entry, error)
}`

	// when
	template, _ := generator.parseTemplate("interfaces.tmpl")
	var output bytes.Buffer
	_ = generator.generateOutputFileFromTemplate(template, &output, generator.Spec)

	// then
	assert.Equal(t, expectedFileContent, output.String())
}
