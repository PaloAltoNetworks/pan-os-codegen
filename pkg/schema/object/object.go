package object

import (
	"gopkg.in/yaml.v3"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/imports"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/location"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/parameter"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/validator"
)

type TerraformResourceType string

const (
	TerraformResourceEntry  TerraformResourceType = "entry"
	TerraformResourceUuid   TerraformResourceType = "uuid"
	TerraformResourceConfig TerraformResourceType = "config"
	TerraformResourceCustom TerraformResourceType = "custom"
)

type TerraformResourceVariant string

const (
	TerraformResourceSingular TerraformResourceVariant = "singular"
	TerraformResourcePlural   TerraformResourceVariant = "plural"
)

type TerraformConfig struct {
	Description           string                     `yaml:"description"`
	Epheneral             bool                       `yaml:"ephemeral"`
	SkipResource          bool                       `yaml:"skip_resource"`
	SkipDatasource        bool                       `yaml:"skip_datasource"`
	SkipdatasourceListing bool                       `yaml:"skip_datasource_listing"`
	ResourceType          TerraformResourceType      `yaml:"resource_type"`
	CustomFunctions       map[string]string          `yaml:"custom_functions"`
	ResourceVariants      []TerraformResourceVariant `yaml:"resource_variants"`
	Suffix                string                     `yaml:"suffix"`
	PluralSuffix          string                     `yaml:"plural_suffix"`
	PluralName            string                     `yaml:"plural_name"`
	PluralDescription     string                     `yaml:"plural_description"`
}

type GoSdkConfig struct {
	Skip    bool     `yaml:"skip"`
	Package []string `yaml:"package"`
}

type Entry struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Validators  []*validator.Validator `yaml:"validators"`
}

type Spec struct {
	Required   bool                   `yaml:"required"`
	Parameters []*parameter.Parameter `yaml:"params"`
	Variants   []*parameter.Parameter `yaml:"variants"`
}

type Object struct {
	Name            string              `yaml:"-"`
	DisplayName     string              `yaml:"name"`
	XpathSuffix     []string            `yaml:"xpath_suffix"`
	TerraformConfig *TerraformConfig    `yaml:"terraform_provider_config"`
	Version         string              `yaml:"version"`
	GoSdkConfig     *GoSdkConfig        `yaml:"go_sdk_config"`
	Locations       []location.Location `yaml:"locations"`
	Entries         []Entry             `yaml:"entries"`
	Imports         []imports.Import    `yaml:"imports"`
	Spec            *Spec               `yaml:"spec"`
}

func NewFromBytes(name string, objectBytes []byte) (*Object, error) {
	var object Object

	err := yaml.Unmarshal(objectBytes, &object)
	if err != nil {
		return nil, err
	}

	object.Name = name

	return &object, nil
}
