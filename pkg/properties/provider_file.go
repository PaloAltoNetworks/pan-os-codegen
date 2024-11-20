package properties

import (
	"fmt"
	"strings"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/imports"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/naming"
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
		SpecMetadata:  make(map[string]TerraformProviderSpecMetadata),
		Code:          &code,
	}
}

type TerraformNameProvider struct {
	TfName               string
	MetaName             string
	StructName           string
	DataSourceStructName string
	ResourceStructName   string
	PackageName          string
}

func NewTerraformNameProvider(spec *Normalization, resourceTyp ResourceType) *TerraformNameProvider {
	var tfName string
	switch resourceTyp {
	case ResourceEntry, ResourceCustom, ResourceConfig:
		tfName = spec.TerraformProviderConfig.Suffix
	case ResourceEntryPlural:
		tfName = spec.TerraformProviderConfig.PluralSuffix
	case ResourceUuid:
		tfName = spec.TerraformProviderConfig.Suffix
	case ResourceUuidPlural:
		suffix := spec.TerraformProviderConfig.Suffix
		pluralName := spec.TerraformProviderConfig.PluralName
		tfName = fmt.Sprintf("%s_%s", suffix, pluralName)
	}
	objectName := tfName

	metaName := fmt.Sprintf("_%s", naming.Underscore("", strings.ToLower(objectName), ""))
	structName := naming.CamelCase("", tfName, "", true)
	dataSourceStructName := naming.CamelCase("", tfName, "DataSource", true)
	resourceStructName := naming.CamelCase("", tfName, "Resource", true)
	packageName := spec.GoSdkPath[len(spec.GoSdkPath)-1]
	return &TerraformNameProvider{tfName, metaName, structName, dataSourceStructName, resourceStructName, packageName}
}

type TerraformSpecFlags uint

const (
	TerraformSpecDatasource = 0x01
	TerraformSpecResource   = 0x02
	TerraformSpecImportable = 0x04
)

type TerraformProviderSpecMetadata struct {
	ResourceSuffix string
	StructName     string
	Flags          TerraformSpecFlags
}

// TerraformProviderFile is a Terraform provider file handler.
type TerraformProviderFile struct {
	Filename      string
	Directory     []string
	ImportManager *imports.Manager
	DataSources   []string
	Resources     []string
	SpecMetadata  map[string]TerraformProviderSpecMetadata
	Code          *strings.Builder
}
