package properties

type ResourceType int

const (
	ResourceEntry ResourceType = iota
	ResourceCustom
	ResourceConfig
	ResourceEntryPlural
	ResourceUuid
	ResourceUuidPlural
)

type SchemaType int

const (
	SchemaResource   SchemaType = iota
	SchemaDataSource SchemaType = iota
	SchemaCommon     SchemaType = iota
	SchemaProvider   SchemaType = iota
)
