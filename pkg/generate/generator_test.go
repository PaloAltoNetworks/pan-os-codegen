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
	spec := &properties.Normalization{
		GoSdkPath: []string{"path", "to", "go"},
	}
	generator := NewCreator("test_output", "templates", spec)

	expected := filepath.Join("test_output", "path", "to", "go", "template.go")
	fullFilePath := generator.createFullFilePath("template.tmpl")

	assert.Equal(t, expected, fullFilePath)
}

func TestListOfTemplates(t *testing.T) {
	// Setup: Create a temporary directory for templates
	tempDir := t.TempDir()
	// Create dummy template files
	templateNames := []string{"test1.tmpl", "test2.tmpl"}
	for _, name := range templateNames {
		file, err := os.Create(filepath.Join(tempDir, name))
		require.NoError(t, err)
		file.Close()
	}

	spec := &properties.Normalization{}
	generator := NewCreator("", tempDir, spec)
	templates, err := generator.listOfTemplates()

	require.NoError(t, err)
	assert.Len(t, templates, len(templateNames))

	// Verify template names without directory paths
	for _, template := range templates {
		assert.True(t, strings.Contains(strings.Join(templateNames, " "), template))
	}
}

func TestParseTemplate(t *testing.T) {
	tempDir := t.TempDir()
	templateContent := `package {{.GoSdkPath}}

type Entry struct {}`
	templatePath := filepath.Join(tempDir, "test.tmpl")
	err := os.WriteFile(templatePath, []byte(templateContent), 0644)
	require.NoError(t, err)

	spec := &properties.Normalization{
		GoSdkPath: []string{"object", "address"},
	}
	generator := NewCreator("", tempDir, spec)
	template, err := generator.parseTemplate("test.tmpl")
	require.NoError(t, err)

	var output bytes.Buffer
	err = template.Execute(&output, spec)
	require.NoError(t, err)

	expectedFileContent := "package [object address]\n\ntype Entry struct {}"
	assert.Equal(t, expectedFileContent, output.String())
}
