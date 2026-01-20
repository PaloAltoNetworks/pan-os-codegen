package codegen

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
	"github.com/stretchr/testify/assert"
)

func TestNewCommand(t *testing.T) {
	// given
	tests := []struct {
		name        string
		commandType properties.CommandType
		wantPath    string
	}{
		{"SDK Command", properties.CommandTypeSDK, "templates/sdk"},
		{"Terraform Command", properties.CommandTypeTerraform, "templates/terraform"},
	}

	for _, tt := range tests {
		// when
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cmd, err := NewCommand(ctx, tt.commandType, "dummyArg")

			// then
			assert.NoError(t, err)
			assert.Equal(t, tt.wantPath, cmd.templatePath)
		})
	}
}

func TestCommandFunctionality(t *testing.T) {
	// given
	ctx := context.Background()
	cmdType := properties.CommandTypeSDK

	// Create a temporary file to simulate the config file
	tmpDir := t.TempDir()
	tmpFile, err := os.Create(filepath.Join(tmpDir, "config-*.yaml"))
	assert.NoError(t, err)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			assert.NoError(t, err)
		}
	}(tmpFile.Name())

	// Write the necessary configuration data to the temporary file
	configData := `
output:
  go_sdk: "../generated/pango"
  terraform_provider: "../generated/terraform-provider-panos"
`
	_, err = tmpFile.WriteString(configData)
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)

	// when
	cmd, err := NewCommand(ctx, cmdType, tmpFile.Name())
	assert.NoError(t, err)

	// then
	assert.NoError(t, cmd.Setup(), "Setup should not return an error")
}
func TestCommandSetup(t *testing.T) {
	//given
	ctx := context.Background()

	cmd, err := NewCommand(ctx, properties.CommandTypeSDK, "config.yml")
	assert.NoError(t, err)

	// when
	err = cmd.Setup()

	//then
	assert.NoError(t, err)

}
