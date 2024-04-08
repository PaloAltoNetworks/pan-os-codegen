package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestParseFlags(t *testing.T) {
	// Given
	expectedConfig := Config{
		ConfigFile: "./cmd/codegen/config.yaml",
		OpType:     "mksdk",
	}
	os.Args = []string{"cmd", "--config=./cmd/codegen/config.yaml", "--type=mksdk"}

	// When
	actualConfig := parseFlags()

	// Then
	assert.Equal(t, expectedConfig, actualConfig)
}
