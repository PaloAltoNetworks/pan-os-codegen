package generate

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateFullFilePath(t *testing.T) {
	// given
	spec := &properties.Normalization{
		GoSdkPath: []string{"path", "to", "go"},
	}
	expected := filepath.Join("test_output", "path", "to", "go", "template.go")
	// when

	generator := NewCreator("test_output", "templates", spec)
	fullFilePath := generator.createFullFilePath("template.tmpl")

	// then
	assert.Equal(t, expected, fullFilePath)
}

// NOTE - unit tests should only touch code inside package, do not reference external resources.
// Nevertheless, below tests ARE REFERENCING to text templates, because template ARE NOT EMBEDDED into Go files.
// Technically we could embed them, but then:
// 1 - we are losing clarity of the code,
// 2 - we are mixing Golang with templates expressions.
// Testing generator is crucial, so below tests we can br treated more as integration tests, not unit one.

func TestListOfTemplates(t *testing.T) {
	// given

	tempDir := t.TempDir()
	templateNames := []string{"test1.tmpl", "test2.tmpl"}
	for _, name := range templateNames {
		file, err := os.Create(filepath.Join(tempDir, name))
		assert.NoError(t, err)
		err = file.Close()
		if err != nil {
			assert.NoError(t, err)
		}
	}

	// when
	spec := &properties.Normalization{}
	generator := NewCreator("", tempDir, spec)
	templates, err := generator.listOfTemplates()

	require.NoError(t, err)
	assert.Len(t, templates, len(templateNames))

	// then
	for _, template := range templates {
		assert.True(t, strings.Contains(strings.Join(templateNames, " "), template))
	}
}

func TestParseTemplateForInterfaces(t *testing.T) {
	// given
	tempDir := t.TempDir()
	templateContent := `package {{.GoSdkPath}}

type Entry struct {}`
	templatePath := filepath.Join(tempDir, "test.tmpl")
	err := os.WriteFile(templatePath, []byte(templateContent), 0644)
	assert.NoError(t, err)

	spec := &properties.Normalization{
		GoSdkPath: []string{"object", "address"},
	}
	generator := NewCreator("", tempDir, spec)
	template, err := generator.parseTemplate("test.tmpl")
	assert.NoError(t, err)

	// when
	var output bytes.Buffer
	err = template.Execute(&output, spec)
	assert.NoError(t, err)

	expectedFileContent := "package [object address]\n\ntype Entry struct {}"

	// then
	assert.Equal(t, expectedFileContent, output.String())
}
