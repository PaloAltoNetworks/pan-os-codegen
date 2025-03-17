package properties

type ResourceType string

const (
	ResourceEntry       ResourceType = "entry"
	ResourceCustom      ResourceType = "custom"
	ResourceConfig      ResourceType = "config"
	ResourceEntryPlural ResourceType = "entry-plural"
	ResourceUuid        ResourceType = "uuid"
	ResourceUuidPlural  ResourceType = "uuid-plural"
)

type SchemaType int

const (
	SchemaResource          SchemaType = iota
	SchemaEphemeralResource SchemaType = iota
	SchemaDataSource        SchemaType = iota
	SchemaCommon            SchemaType = iota
	SchemaProvider          SchemaType = iota
)
