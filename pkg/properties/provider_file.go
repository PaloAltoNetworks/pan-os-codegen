package properties

import (
	"strings"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/imports"
)

// NewTerraformProviderFile returns a new handler for a file that the
// generator needs to create for the Terraform provider.  This will
// help aggregate all the essential information for the provider file,
// such as imports and the code itself.
func NewTerraformProviderFile(filename string) *TerraformProviderFile {
	var code strings.Builder
	code.Grow(1e4)

	return &TerraformProviderFile{
		Filename:      filename,
		ImportManager: imports.NewManager(),
		DataSources:   make([]string, 0, 10),
		Resources:     make([]string, 0, 10),
		Code:          &code,
	}
}

// TerraformProviderFile is a Terraform provider file handler.
type TerraformProviderFile struct {
	Filename      string
	Directory     []string
	ImportManager *imports.Manager
	DataSources   []string
	Resources     []string
	Code          *strings.Builder
}
