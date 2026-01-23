package manager

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"

	sdkerrors "github.com/PaloAltoNetworks/pango/errors"
)

// Type constraints for batch reading operations

// EntryLike represents an entry that has a name field
type EntryLike interface {
	EntryName() string
}

// EntryWithAttributes represents an entry that has XML attributes
// Used for filtering by location
type EntryWithAttributes interface {
	EntryLike
	GetMiscAttributes() []xml.Attr
}

// LocationLike represents a location that can filter entries
type LocationLike interface {
	LocationFilter() *string
}

// ServiceWithListXpath represents a service that can list entries using XPath
type ServiceWithListXpath[E any] interface {
	ListWithXpath(ctx context.Context, xpath string, action string, vsysName string, deviceName string) ([]E, error)
}

// Shard definitions for name-based sharding strategy
// 7 shards based on first character: 0-4, 5-9, a-f, g-l, m-r, s-z, special
var NameShards = []struct {
	Name     string
	Prefixes []string
}{
	{"0-4", []string{"0", "1", "2", "3", "4"}},
	{"5-9", []string{"5", "6", "7", "8", "9"}},
	{"a-f", []string{"a", "A", "b", "B", "c", "C", "d", "D", "e", "E", "f", "F"}},
	{"g-l", []string{"g", "G", "h", "H", "i", "I", "j", "J", "k", "K", "l", "L"}},
	{"m-r", []string{"m", "M", "n", "N", "o", "O", "p", "P", "q", "Q", "r", "R"}},
	{"s-z", []string{"s", "S", "t", "T", "u", "U", "v", "V", "w", "W", "x", "X", "y", "Y", "z", "Z"}},
	{"special", []string{"-", "_", ".", "@", "#", "!", "$", "%", "&", "*", "+", "=", "~"}},
}

// Helper functions

// ExtractNames extracts the name field from a slice of entries
func ExtractNames[E EntryLike](entries []E) []string {
	var names []string
	for _, entry := range entries {
		names = append(names, entry.EntryName())
	}
	return names
}

// BuildShardedXpath builds an XPath with starts-with predicates for sharding
func BuildShardedXpath(baseXpath string, prefixes []string) string {
	var predicates []string
	for _, prefix := range prefixes {
		predicates = append(predicates, fmt.Sprintf("starts-with(@name, '%s')", prefix))
	}

	predicate := strings.Join(predicates, " or ")
	return fmt.Sprintf("%s[%s]/@name", baseXpath, predicate)
}

// BuildBatchXpath builds an XPath with OR predicates for batch reading
func BuildBatchXpath(baseXpath string, names []string) string {
	var predicates []string
	for _, name := range names {
		// TODO: Escape name for XPath safety
		predicates = append(predicates, fmt.Sprintf("@name='%s'", name))
	}

	predicate := strings.Join(predicates, " or ")
	return fmt.Sprintf("%s/entry[%s]", baseXpath, predicate)
}

// FilterEntriesByLocation filters entries to only those matching the location filter
func FilterEntriesByLocation[E EntryWithAttributes, L LocationLike](location L, entries []E) []E {
	filter := location.LocationFilter()
	if filter == nil {
		return entries
	}

	getLocAttribute := func(entry E) *string {
		for _, elt := range entry.GetMiscAttributes() {
			if elt.Name.Local == "loc" {
				return &elt.Value
			}
		}
		return nil
	}

	var filtered []E
	for _, elt := range entries {
		entryLoc := getLocAttribute(elt)
		if entryLoc == nil || *entryLoc == *filter {
			filtered = append(filtered, elt)
		}
	}

	return filtered
}

// BatchReader provides shared batching operations for reading entries
type BatchReader[E EntryLike, S ServiceWithListXpath[E]] struct {
	service        S
	batchingConfig BatchingConfig
}

// NewBatchReader creates a new batch reader
func NewBatchReader[E EntryLike, S ServiceWithListXpath[E]](
	service S,
	batchingConfig BatchingConfig,
) *BatchReader[E, S] {
	return &BatchReader[E, S]{
		service:        service,
		batchingConfig: batchingConfig,
	}
}

// ListNamesUnsharded fetches all entry names in a single query
func (br *BatchReader[E, S]) ListNamesUnsharded(
	ctx context.Context,
	baseXpath string,
) ([]string, error) {
	nameXpath := baseXpath + "/@name"

	entries, err := br.service.ListWithXpath(ctx, nameXpath, "get", "", "")
	if err != nil {
		if sdkerrors.IsObjectNotFound(err) {
			return []string{}, nil
		}
		return nil, &Error{err: err, message: "Failed to list entry names"}
	}

	return ExtractNames(entries), nil
}

// ListNamesSharded fetches entry names using multiple sharded queries
func (br *BatchReader[E, S]) ListNamesSharded(
	ctx context.Context,
	baseXpath string,
) ([]string, error) {
	var allNames []string

	for _, shard := range NameShards {
		shardXpath := BuildShardedXpath(baseXpath, shard.Prefixes)

		entries, err := br.service.ListWithXpath(ctx, shardXpath, "get", "", "")
		if err != nil {
			if sdkerrors.IsObjectNotFound(err) {
				continue // Empty shard
			}
			return nil, &Error{
				err:     err,
				message: fmt.Sprintf("Failed to list shard %s", shard.Name),
			}
		}

		allNames = append(allNames, ExtractNames(entries)...)
	}

	return allNames, nil
}

// ListNames fetches entry names using the configured sharding strategy
func (br *BatchReader[E, S]) ListNames(
	ctx context.Context,
	baseXpath string,
) ([]string, error) {
	switch br.batchingConfig.ShardingStrategy {
	case ShardingDisabled:
		return br.ListNamesUnsharded(ctx, baseXpath)
	case ShardingEnabled:
		return br.ListNamesSharded(ctx, baseXpath)
	default:
		return br.ListNamesUnsharded(ctx, baseXpath)
	}
}

// ReadEntriesByNames fetches a batch of entries by their names
func (br *BatchReader[E, S]) ReadEntriesByNames(
	ctx context.Context,
	baseXpath string,
	names []string,
) ([]E, error) {
	batchXpath := BuildBatchXpath(baseXpath, names)

	entries, err := br.service.ListWithXpath(ctx, batchXpath, "get", "", "")
	if err != nil {
		if sdkerrors.IsObjectNotFound(err) {
			return []E{}, nil
		}
		return nil, &Error{err: err, message: "Failed to read entry batch"}
	}

	return entries, nil
}

// BatchReadEntries reads entries in batches using the configured batch size
func (br *BatchReader[E, S]) BatchReadEntries(
	ctx context.Context,
	baseXpath string,
	names []string,
) ([]E, error) {
	batchSize := br.batchingConfig.ReadBatchSize
	if batchSize <= 0 {
		batchSize = 50 // default
	}

	var allEntries []E

	// Process names in batches
	for i := 0; i < len(names); i += batchSize {
		end := i + batchSize
		if end > len(names) {
			end = len(names)
		}

		batch := names[i:end]
		entries, err := br.ReadEntriesByNames(ctx, baseXpath, batch)
		if err != nil {
			return nil, err
		}

		allEntries = append(allEntries, entries...)
	}

	return allEntries, nil
}

// ReadManyLazy implements the lazy reading strategy: list names, then batch read
func (br *BatchReader[E, S]) ReadManyLazy(
	ctx context.Context,
	baseXpath string,
) ([]E, error) {
	// Phase 1: Get list of names
	names, err := br.ListNames(ctx, baseXpath)
	if err != nil {
		return nil, err
	}

	if len(names) == 0 {
		return nil, ErrObjectNotFound
	}

	// Phase 2: Fetch entries in batches
	return br.BatchReadEntries(ctx, baseXpath, names)
}
