package manager

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/PaloAltoNetworks/pango/version"
	"github.com/PaloAltoNetworks/pango/xmlapi"
)

type Error struct {
	message string
	err     error
}

func (o *Error) Error() string {
	if o.err != nil {
		return fmt.Sprintf("%s: %s", o.message, o.err)
	}

	return o.message
}

func (o *Error) Unwrap() error {
	return o.err
}

var (
	ErrPlanConflict                = errors.New("multiple plan entries with shared name")
	ErrConflict                    = errors.New("entry from the plan already exists on the server")
	ErrMissingUuid                 = errors.New("entry is missing required uuid")
	ErrMarshaling                  = errors.New("failed to marshal entry to XML document")
	ErrInvalidPosition             = errors.New("position is not valid")
	ErrMissingPivotPoint           = errors.New("provided pivot entry does not exist")
	ErrInternal                    = errors.New("internal provider error")
	ErrObjectNotFound              = errors.New("Object not found")
	ErrLocationCacheNotInitialized = errors.New("cache location not initialized, call SetInitialized first")
)

// Entry represents the minimal interface for cache entries - read-only name access.
type Entry interface {
	EntryName() string
}

// EntryWithAttributes extends Entry with XML attribute access for location filtering.
type EntryWithAttributes interface {
	Entry
	GetMiscAttributes() []xml.Attr
}

// EntryObject extends EntryWithAttributes with mutable name for CRUD operations.
type EntryObject interface {
	EntryWithAttributes
	SetEntryName(name string)
}

// UuidObject extends EntryObject with UUID fields for UUID-based resources.
type UuidObject interface {
	EntryObject
	EntryUuid() *string
	SetEntryUuid(value *string)
}

type entryState string

const (
	entryUnknown  entryState = "unknown"
	entryMissing  entryState = "missing"
	entryOutdated entryState = "outdated"
	entryRenamed  entryState = "renamed"
	entryDeleted  entryState = "deleted"
	entryOk       entryState = "ok"
)

// entryWithState tracks an entry's CRUD lifecycle state and ordering for UpdateMany operations.
type entryWithState[E Entry] struct {
	Entry    E
	State    entryState
	StateIdx int
	NewName  string
}

type SDKClient interface {
	Versioning() version.Number
	GetTarget() string
	ChunkedMultiConfig(context.Context, *xmlapi.MultiConfig, bool, url.Values) ([]xmlapi.ChunkedMultiConfigResponse, error)
	MultiConfig(context.Context, *xmlapi.MultiConfig, bool, url.Values) ([]byte, *http.Response, *xmlapi.MultiConfigResponse, error)
}

// BatchingConfig configures batching behavior for SDK operations.
type BatchingConfig struct {
	MultiConfigBatchSize int              // Batch size for write operations
	ReadBatchSize        int              // Batch size for lazy read operations
	ListStrategy         ListStrategy     // Eager or Lazy listing
	ShardingStrategy     ShardingStrategy // Disabled or Enabled
	CacheStrategy        CacheStrategy    // Caching strategy
}

// ListStrategy controls how resources are listed.
type ListStrategy string

const (
	StrategyEager ListStrategy = "eager" // Single query, fetch all entries with full details
	StrategyLazy  ListStrategy = "lazy"  // List names first, then batch read entries
)

// ShardingStrategy controls whether name listing is sharded.
type ShardingStrategy string

const (
	ShardingDisabled ShardingStrategy = "disabled" // Single query for all names
	ShardingEnabled  ShardingStrategy = "enabled"  // Multiple queries sharded by name prefix
)

// CacheStrategy controls resource caching.
type CacheStrategy string

const (
	CacheStrategyDisabled CacheStrategy = "disabled" // Caching disabled
	CacheStrategyEnabled  CacheStrategy = "enabled"  // Caching enabled
)
