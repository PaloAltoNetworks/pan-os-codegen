package codegen

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCommand(t *testing.T) {
	// given
	tests := []struct {
		name        string
		commandType CommandType
		wantPath    string
	}{
		{"SDK Command", CommandTypeSDK, "templates/sdk"},
		{"Terraform Command", CommandTypeTerraform, "templates/terraform"},
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

func TestExecute_NoArgs(t *testing.T) {
	// given
	ctx := context.Background()
	cmd, _ := NewCommand(ctx, CommandTypeSDK)

	// when
	err := cmd.Execute()

	// then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "path to configuration file is required")
}
