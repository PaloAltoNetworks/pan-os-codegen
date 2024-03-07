package codegen

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCommand(t *testing.T) {
	tests := []struct {
		name        string
		commandType CommandType
		wantPath    string
	}{
		{"SDK Command", CommandTypeSDK, "../templates/sdk"},
		{"Terraform Command", CommandTypeTerraform, "../templates/terraform"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cmd, err := NewCommand(ctx, tt.commandType, "dummyArg")
			require.NoError(t, err)
			assert.Equal(t, tt.wantPath, cmd.templatePath)
		})
	}
}

func TestCommandFunctionality(t *testing.T) {
	ctx := context.Background()
	cmdType := CommandTypeSDK

	// Create a temporary file to simulate the config file
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name()) // Clean up the file after the test

	// Write the necessary configuration data to the temporary file
	configData := `
output:
  go_sdk: "../generated/pango"
  terraform_provider: "../generated/terraform-provider-panos"
`
	_, err = tmpFile.WriteString(configData)
	require.NoError(t, err)
	err = tmpFile.Close() // Close the file after writing
	require.NoError(t, err)

	cmd, err := NewCommand(ctx, cmdType, tmpFile.Name())
	require.NoError(t, err)

	require.NoError(t, cmd.Setup(), "Setup should not return an error")
}
func TestCommand_Setup(t *testing.T) {
	ctx := context.Background()

	cmd, err := NewCommand(ctx, CommandTypeSDK, "config.yml")
	require.NoError(t, err)

	err = cmd.Setup()
	require.NoError(t, err)

}
