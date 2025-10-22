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

type SchemaType string

const (
	SchemaResource          SchemaType = "resource"
	SchemaEphemeralResource SchemaType = "ephemeral-resource"
	SchemaListResource      SchemaType = "list-resource"
	SchemaAction            SchemaType = "action"
	SchemaDataSource        SchemaType = "datasource"
	SchemaCommon            SchemaType = "common"
	SchemaProvider          SchemaType = "provider"
	SchemaCustom            SchemaType = "custom"
)
