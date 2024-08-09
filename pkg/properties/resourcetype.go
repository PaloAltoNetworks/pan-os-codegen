package properties

type ResourceType int

const (
	ResourceEntry       ResourceType = iota
	ResourceEntryPlural ResourceType = iota
	ResourceUuid        ResourceType = iota
	ResourceUuidPlural  ResourceType = iota
)
