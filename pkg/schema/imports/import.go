package imports

type Import struct {
	Variants     []string `yaml:"variants"`
	Target       string   `yaml:"target"`
	DefaultValue *string  `yaml:"default_value,omitempty"`
}
