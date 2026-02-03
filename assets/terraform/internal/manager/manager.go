package manager

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/PaloAltoNetworks/pango/util"
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
	ErrPlanConflict      = errors.New("multiple plan entries with shared name")
	ErrConflict          = errors.New("entry from the plan already exists on the server")
	ErrMissingUuid       = errors.New("entry is missing required uuid")
	ErrMarshaling        = errors.New("failed to marshal entry to XML document")
	ErrInvalidPosition   = errors.New("position is not valid")
	ErrMissingPivotPoint = errors.New("provided pivot entry does not exist")
	ErrInternal          = errors.New("internal provider error")
	ErrObjectNotFound    = errors.New("Object not found")
)

type entryState string

const (
	entryUnknown  entryState = "unknown"
	entryMissing  entryState = "missing"
	entryOutdated entryState = "outdated"
	entryRenamed  entryState = "renamed"
	entryDeleted  entryState = "deleted"
	entryOk       entryState = "ok"
)

type SDKClient interface {
	Versioning() version.Number
	GetTarget() string
	ChunkedMultiConfig(context.Context, *xmlapi.MultiConfig, bool, url.Values) ([]xmlapi.ChunkedMultiConfigResponse, error)
	MultiConfig(context.Context, *xmlapi.MultiConfig, bool, url.Values) ([]byte, *http.Response, *xmlapi.MultiConfigResponse, error)
}

type ImportLocation interface {
	XpathForLocation(version.Number, util.ILocation) ([]string, error)
	MarshalPangoXML([]string) (string, error)
	UnmarshalPangoXML([]byte) ([]string, error)
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
