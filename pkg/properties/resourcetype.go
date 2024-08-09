package properties

type ResourceType int

const (
	ResourceEntry      ResourceType = iota
	ResourceUuid       ResourceType = iota
	ResourceUuidPlural ResourceType = iota
)
