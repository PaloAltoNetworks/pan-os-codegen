package properties

type TerraformProvider struct {
	TerraformProviderParams map[string]TerraformProviderParams `json:"params" yaml:"params"`
}

type TerraformProviderParams struct {
	Description  string              `json:"description" yaml:"description"`
	DefaultValue string              `json:"default_value" yaml:"default_value"`
	EnvName      string              `json:"env_name" yaml:"env_name"`
	Optional     string              `json:"optional" yaml:"optional"`
	Type         string              `json:"type" yaml:"type"`
	Sensitive    string              `json:"sensitive" yaml:"sensitive"`
	Items        *ProviderParamItems `json:"items" yaml:"items,omitempty"`
}

type ProviderParamItems struct {
	Type   string                    `json:"type" yaml:"type"`
	Length *ProviderParamItemsLength `json:"length" yaml:"length"`
	Ref    []*string                 `json:"ref" yaml:"ref"`
}

type ProviderParamItemsLength struct {
	Min *int64 `json:"min" yaml:"min"`
	Max *int64 `json:"max" yaml:"max"`
}
