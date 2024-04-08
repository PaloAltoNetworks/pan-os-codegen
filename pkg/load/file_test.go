package load

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestFile(t *testing.T) {
	// Given
	testFilename := "testfile.txt"
	testContent := []byte("This is a test file content")
	err := os.WriteFile(testFilename, testContent, 0644)
	if err != nil {
		t.Fatalf("Unable to write test file: %v", err)
	}
	defer os.Remove(testFilename)

	// When
	resultContent, err := File(testFilename)

	// Then
	assert.NoError(t, err, "File function should complete without error")
	assert.Equal(t, testContent, resultContent, "Content of the result should be identical to the test content")
}

func TestFile_FileDoesNotExist(t *testing.T) {
	// Given
	testFilename := "nonexistentfile.txt"

	// When
	_, err := File(testFilename)

	// Then
	assert.Error(t, err, "File function should return error on nonexistent file")
}
