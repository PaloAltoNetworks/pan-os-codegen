package object

import (
	"fmt"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/imports"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/location"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/parameter"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/schema/validator"
	"gopkg.in/yaml.v3"
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

type TerraformPluralType string

const (
	TerraformPluralListType TerraformPluralType = "list"
	TerraformPluralMapType  TerraformPluralType = "map"
	TerraformPluralSetType  TerraformPluralType = "set"
)

type TerraformConfig struct {
	Description           string                     `yaml:"description"`
	Action                bool                       `yaml:"action"`
	Epheneral             bool                       `yaml:"ephemeral"`
	CustomValidation      bool                       `yaml:"custom_validation"`
	SkipResource          bool                       `yaml:"skip_resource"`
	SkipDatasource        bool                       `yaml:"skip_datasource"`
	SkipdatasourceListing bool                       `yaml:"skip_datasource_listing"`
	ResourceType          TerraformResourceType      `yaml:"resource_type"`
	XmlNode               *string                    `yaml:"xml_node"`
	CustomFunctions       map[string]bool            `yaml:"custom_functions"`
	ResourceVariants      []TerraformResourceVariant `yaml:"resource_variants"`
	Suffix                string                     `yaml:"suffix"`
	PluralSuffix          string                     `yaml:"plural_suffix"`
	PluralName            string                     `yaml:"plural_name"`
	PluralType            TerraformPluralType        `yaml:"plural_type"`
	PluralDescription     string                     `yaml:"plural_description"`
}

type GoSdkMethod string

const (
	GoSdkMethodCreate = "create"
	GoSdkMethodUpdate = "update"
	GoSdkMethodDelete = "delete"
	GoSdkMethodRead   = "read"
	GoSdkMethodList   = "list"
)

type GoSdkConfig struct {
	Skip             bool          `yaml:"skip"`
	SupportedMethods []GoSdkMethod `yaml:"supported_methods"`
	Package          []string      `yaml:"package"`
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

type PanosXpathVariableType string

const (
	PanosXpathVariableStatic PanosXpathVariableType = "static"
	PanosXpathVariableValue  PanosXpathVariableType = "value"
	PanosXpathVariableEntry  PanosXpathVariableType = "entry"
)

type PanosXpathVariableSpec struct {
	Type  PanosXpathVariableType `yaml:"type"`
	Xpath string                 `yaml:"xpath"`
}

type PanosXpathVariable struct {
	Name string                 `yaml:"name"`
	Spec PanosXpathVariableSpec `yaml:"spec"`
}

type PanosXpath struct {
	Path      []string             `yaml:"path"`
	Variables []PanosXpathVariable `yaml:"vars"`
}

type Object struct {
	Name            string              `yaml:"-"`
	DisplayName     string              `yaml:"name"`
	PanosXpath      PanosXpath          `yaml:"panos_xpath"`
	TerraformConfig *TerraformConfig    `yaml:"terraform_provider_config"`
	Version         string              `yaml:"version"`
	GoSdkConfig     *GoSdkConfig        `yaml:"go_sdk_config"`
	Locations       []location.Location `yaml:"locations"`
	Entries         []Entry             `yaml:"entries"`
	Imports         []imports.Import    `yaml:"imports"`
	Spec            *Spec               `yaml:"spec"`
}

func (o *Object) UnmarshalYAML(n *yaml.Node) error {
	type O Object
	type S struct {
		*O `yaml:",inline"`
	}

	obj := &S{O: (*O)(o)}

	if err := n.Decode(obj); err != nil {
		return err
	}

	if o.GoSdkConfig == nil || o.GoSdkConfig.SupportedMethods == nil {
		o.GoSdkConfig.SupportedMethods = []GoSdkMethod{
			GoSdkMethodCreate,
			GoSdkMethodDelete,
			GoSdkMethodRead,
			GoSdkMethodList,
			GoSdkMethodUpdate,
		}
	}

	switch obj.TerraformConfig.ResourceType {
	case TerraformResourceEntry, TerraformResourceUuid:
		if obj.PanosXpath.Path[len(obj.PanosXpath.Path)-1] != "$name" {
			obj.PanosXpath.Path = append(obj.PanosXpath.Path, "$name")
			obj.PanosXpath.Variables = append(obj.PanosXpath.Variables, PanosXpathVariable{
				Name: "name",
				Spec: PanosXpathVariableSpec{
					Type:  PanosXpathVariableEntry,
					Xpath: "/params[@name=\"name\"]",
				},
			})
		}
	case TerraformResourceCustom, TerraformResourceConfig:
	}

	if obj.TerraformConfig.PluralType == "" {
		switch obj.TerraformConfig.ResourceType {
		case TerraformResourceUuid:
			obj.TerraformConfig.PluralType = TerraformPluralListType
		case TerraformResourceEntry:
			obj.TerraformConfig.PluralType = TerraformPluralMapType
		case TerraformResourceConfig, TerraformResourceCustom:
		}
	} else if obj.TerraformConfig.ResourceType == TerraformResourceUuid && obj.TerraformConfig.PluralType != "list" {
		return fmt.Errorf("failed to unmarshal yaml spec: plural_type must be list for uuid resource types")
	}

	return nil
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
