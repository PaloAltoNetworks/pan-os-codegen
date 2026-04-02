package properties

import (
	"os"
	"testing"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/parameter"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

const addressSpecPath = "../../specs/objects/address.yaml"

func TestUnmarshallAddressSpecFile(t *testing.T) {
	// given
	sampleSpec, err := os.ReadFile(addressSpecPath)
	assert.Nil(t, err, "failed to read address spec")
	// when
	yamlParsedData, err := ParseSpec([]byte(sampleSpec))
	assert.Nil(t, err)

	// then
	assert.NotNilf(t, yamlParsedData, "Unmarshalled data cannot be nil")
	assert.Equal(t, "address", yamlParsedData.Name, "Unmarshalled data should contain `name`")
	assert.NotNilf(t, yamlParsedData.TerraformProviderConfig.Suffix, "Unmarshalled data should contain `suffix`")
	assert.NotNilf(t, yamlParsedData.GoSdkPath, "Unmarshalled data should contain `go_sdk_path`")
	assert.NotNilf(t, yamlParsedData.XpathSuffix, "Unmarshalled data should contain `xpath_suffix`")
	assert.NotNilf(t, yamlParsedData.Locations, "Unmarshalled data should contain `locations`")
	assert.NotNilf(t, yamlParsedData.Entry, "Unmarshalled data should contain `entry`")
	assert.NotNilf(t, yamlParsedData.Version, "Unmarshalled data should contain `version`")
	assert.NotNilf(t, yamlParsedData.Spec, "Unmarshalled data should contain `spec`")
}

func TestGetNormalizations(t *testing.T) {
	// given

	// when
	config, _ := GetNormalizations()

	// then
	assert.NotNil(t, config)
	assert.LessOrEqual(t, 15, len(config), "Expected to have 15 spec YAML files")
}

func TestSanity(t *testing.T) {
	// given
	var sampleInvalidSpec = `
name: 'Address'
terraform_provider_config:
    skip_resource: false
    skip_datasource: false
    skip_datasource_listing: false
    suffix: address
go_sdk_path:
  - 'objects'
  - 'address'
xpath_suffix:
  - 'address'
`
	// when
	yamlParsedData := Normalization{}
	err := yaml.Unmarshal([]byte(sampleInvalidSpec), &yamlParsedData)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	err = yamlParsedData.Sanity()

	// then
	assert.ErrorContainsf(t, err, "at least 1 location is required", "error message %s", err)
}

func TestValidation(t *testing.T) {
	// given
	var sampleInvalidSpec = `
name: 'Address'
terraform_provider_config:
 suffix: 'address'
xpath_suffix:
  - 'address'
`
	// when
	yamlParsedData := Normalization{}
	err := yaml.Unmarshal([]byte(sampleInvalidSpec), &yamlParsedData)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	problems := yamlParsedData.Validate()

	// then
	assert.Len(t, problems, 2, "Not all expected validation checks failed")
}

func TestFinalLocalXmlSupported_ExplicitTrue(t *testing.T) {
	// given - explicit supports_local_xml: true
	trueVal := true
	spec := &Normalization{
		TerraformProviderConfig: TerraformProviderConfig{
			ResourceType:     TerraformResourceEntry,
			SupportsLocalXml: &trueVal,
		},
		Spec: &Spec{
			Params: map[string]*SpecParam{}, // No hashed fields
		},
	}

	// when
	result := spec.FinalLocalXmlSupported()

	// then - explicit value takes priority
	assert.True(t, result, "Explicit supports_local_xml: true should return true")
}

func TestFinalLocalXmlSupported_ExplicitFalse(t *testing.T) {
	// given - explicit supports_local_xml: false
	falseVal := false
	spec := &Normalization{
		TerraformProviderConfig: TerraformProviderConfig{
			ResourceType:     TerraformResourceEntry,
			SupportsLocalXml: &falseVal,
		},
		Spec: &Spec{
			Params: map[string]*SpecParam{}, // No hashed fields
		},
	}

	// when
	result := spec.FinalLocalXmlSupported()

	// then - explicit value takes priority even when auto-detection would return true
	assert.False(t, result, "Explicit supports_local_xml: false should return false")
}

func TestFinalLocalXmlSupported_AutoDetect_NonCustomNoHash(t *testing.T) {
	// given - nil supports_local_xml, entry type, no hashed fields
	spec := &Normalization{
		TerraformProviderConfig: TerraformProviderConfig{
			ResourceType:     TerraformResourceEntry,
			SupportsLocalXml: nil, // Auto-detection
		},
		Spec: &Spec{
			Params: map[string]*SpecParam{
				"test_param": {
					Name: &NameVariant{
						CamelCase: "TestParam",
					},
					Type:    "string",
					Hashing: nil, // No hashing
				},
			},
		},
	}

	// when
	result := spec.FinalLocalXmlSupported()

	// then - auto-detection returns true for non-Custom without hashed fields
	assert.True(t, result, "Non-Custom type without hashed fields should auto-detect as supported")
}

func TestFinalLocalXmlSupported_AutoDetect_CustomType(t *testing.T) {
	// given - nil supports_local_xml, Custom type
	spec := &Normalization{
		TerraformProviderConfig: TerraformProviderConfig{
			ResourceType:     TerraformResourceCustom,
			SupportsLocalXml: nil, // Auto-detection
		},
		Spec: &Spec{
			Params: map[string]*SpecParam{},
		},
	}

	// when
	result := spec.FinalLocalXmlSupported()

	// then - Custom type always returns false
	assert.False(t, result, "Custom resource type should auto-detect as unsupported")
}

func TestFinalLocalXmlSupported_AutoDetect_WithHashedField(t *testing.T) {
	// given - nil supports_local_xml, entry type, has hashed field
	spec := &Normalization{
		TerraformProviderConfig: TerraformProviderConfig{
			ResourceType:     TerraformResourceEntry,
			SupportsLocalXml: nil, // Auto-detection
		},
		Spec: &Spec{
			Params: map[string]*SpecParam{
				"hashed_param": {
					Name: &NameVariant{
						CamelCase: "HashedParam",
					},
					Type: "string",
					Hashing: &parameter.Hashing{
						Type: parameter.HashingSoloType,
					},
				},
			},
		},
	}

	// when
	result := spec.FinalLocalXmlSupported()

	// then - hashed fields prevent local XML support
	assert.False(t, result, "Resources with hashed fields should auto-detect as unsupported")
}

func TestFinalLocalXmlSupported_AutoDetect_NestedHashedField(t *testing.T) {
	// given - nil supports_local_xml, entry type, nested parameter with hashing
	spec := &Normalization{
		TerraformProviderConfig: TerraformProviderConfig{
			ResourceType:     TerraformResourceEntry,
			SupportsLocalXml: nil, // Auto-detection
		},
		Spec: &Spec{
			Params: map[string]*SpecParam{
				"parent_param": {
					Name: &NameVariant{
						CamelCase: "ParentParam",
					},
					Type:    "object",
					Hashing: nil,
					Spec: &Spec{
						Params: map[string]*SpecParam{
							"nested_hashed": {
								Name: &NameVariant{
									CamelCase: "NestedHashed",
								},
								Type: "string",
								Hashing: &parameter.Hashing{
									Type: parameter.HashingClientType,
								},
							},
						},
					},
				},
			},
		},
	}

	// when
	result := spec.FinalLocalXmlSupported()

	// then - nested hashed fields also prevent local XML support
	assert.False(t, result, "Resources with nested hashed fields should auto-detect as unsupported")
}

func TestFinalLocalXmlSupported_ExplicitTrue_OverridesCustomType(t *testing.T) {
	// given - explicit true overrides Custom type auto-detection
	trueVal := true
	spec := &Normalization{
		TerraformProviderConfig: TerraformProviderConfig{
			ResourceType:     TerraformResourceCustom,
			SupportsLocalXml: &trueVal,
		},
		Spec: &Spec{
			Params: map[string]*SpecParam{},
		},
	}

	// when
	result := spec.FinalLocalXmlSupported()

	// then - explicit value has priority over auto-detection
	assert.True(t, result, "Explicit supports_local_xml: true should override Custom type auto-detection")
}

func TestFinalLocalXmlSupported_ExplicitTrue_OverridesHashedFields(t *testing.T) {
	// given - explicit true overrides hashed field auto-detection
	trueVal := true
	spec := &Normalization{
		TerraformProviderConfig: TerraformProviderConfig{
			ResourceType:     TerraformResourceEntry,
			SupportsLocalXml: &trueVal,
		},
		Spec: &Spec{
			Params: map[string]*SpecParam{
				"hashed_param": {
					Name: &NameVariant{
						CamelCase: "HashedParam",
					},
					Type: "string",
					Hashing: &parameter.Hashing{
						Type: parameter.HashingSoloType,
					},
				},
			},
		},
	}

	// when
	result := spec.FinalLocalXmlSupported()

	// then - explicit value has priority
	assert.True(t, result, "Explicit supports_local_xml: true should override hashed field auto-detection")
}

func TestFinalLocalXmlSupported_NilSpec(t *testing.T) {
	// given - nil spec (edge case)
	spec := &Normalization{
		TerraformProviderConfig: TerraformProviderConfig{
			ResourceType:     TerraformResourceEntry,
			SupportsLocalXml: nil,
		},
		Spec: nil,
	}

	// when
	result := spec.FinalLocalXmlSupported()

	// then - no spec means no hashed fields, should return true for non-Custom
	assert.True(t, result, "Nil spec with non-Custom type should auto-detect as supported")
}
