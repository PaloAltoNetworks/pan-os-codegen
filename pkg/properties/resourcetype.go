package properties

type ResourceType int

const (
	ResourceEntry       ResourceType = iota
	ResourceCustom      ResourceType = iota
	ResourceEntryPlural ResourceType = iota
	ResourceUuid        ResourceType = iota
	ResourceUuidPlural  ResourceType = iota
)

type SchemaType int

const (
	SchemaResource   SchemaType = iota
	SchemaDataSource SchemaType = iota
	SchemaCommon     SchemaType = iota
	SchemaProvider   SchemaType = iota
)
