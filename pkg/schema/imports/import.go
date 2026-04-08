package imports

// Import defines vsys import configuration for PAN-OS resources that need to be
// imported into a vsys (e.g., network interfaces, virtual routers, subinterfaces).
type Import struct {
	// Variants lists PAN-OS interface modes or location types that require vsys import.
	// Specific values (e.g., ["layer2", "layer3"]) mean import is only required when
	// the resource's location uses one of those modes; a conditional check is generated
	// in the Terraform provider for each listed variant.
	//
	// The special wildcard value ["*"] means import into vsys regardless of interface
	// mode — used for resources like tunnel interfaces, virtual routers, and subinterfaces
	// that are always importable. When "*" is present, no mode check is generated;
	// locationRequiresImport is set unconditionally to true.
	//
	// If "*" appears alongside specific variants (e.g., ["layer2", "*"]), the wildcard
	// is filtered out and only the specific variant checks are generated.
	Variants []string `yaml:"variants"`

	// Target is the PAN-OS XML element name for the import target (e.g., "interface").
	Target string `yaml:"target"`

	// DefaultValue is the default vsys value to use for import, if any.
	DefaultValue *string `yaml:"default_value,omitempty"`
}
