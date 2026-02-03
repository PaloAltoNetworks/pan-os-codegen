package manager

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"

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
	return fmt.Sprintf("%s[%s]", baseXpath, predicate)
}

// FilterEntriesByLocation filters entries to only those matching the location filter
func FilterEntriesByLocation[E EntryWithAttributes, L LocationLike](location L, entries []E) []E {
	filter := location.LocationFilter()
	if filter == nil {
		return entries
	}

	// Use context.Background() since this is a utility function
	ctx := context.Background()

	tflog.Debug(ctx, "FilterEntriesByLocation: applying location filter", map[string]any{
		"filter":      *filter,
		"entry_count": len(entries),
	})

	getLocAttribute := func(entry E) *string {
		for _, elt := range entry.GetMiscAttributes() {
			if elt.Name.Local == "loc" {
				return &elt.Value
			}
		}
		return nil
	}

	var filtered []E
	var filteredOut int
	for _, elt := range entries {
		entryLoc := getLocAttribute(elt)
		if entryLoc == nil || *entryLoc == *filter {
			filtered = append(filtered, elt)
		} else {
			filteredOut++
			tflog.Debug(ctx, "FilterEntriesByLocation: filtering out entry", map[string]any{
				"entry_name": elt.EntryName(),
				"entry_loc":  *entryLoc,
				"filter":     *filter,
			})
		}
	}

	tflog.Debug(ctx, "FilterEntriesByLocation: filtering complete", map[string]any{
		"filter":       *filter,
		"entries_in":   len(entries),
		"entries_out":  len(filtered),
		"filtered_out": filteredOut,
		"final_order":  extractEntryNames(filtered),
	})

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

	tflog.Debug(ctx, "BatchReader: unsharded name listing", map[string]any{
		"xpath": nameXpath,
	})

	entries, err := br.service.ListWithXpath(ctx, nameXpath, "get", "", "")
	if err != nil {
		if sdkerrors.IsObjectNotFound(err) {
			tflog.Debug(ctx, "BatchReader: unsharded listing returned ObjectNotFound", map[string]any{
				"xpath": nameXpath,
			})
			return []string{}, nil
		}
		tflog.Debug(ctx, "BatchReader: unsharded listing failed", map[string]any{
			"xpath": nameXpath,
			"error": err.Error(),
		})
		return nil, &Error{err: err, message: "Failed to list entry names"}
	}

	names := ExtractNames(entries)

	tflog.Debug(ctx, "BatchReader: unsharded listing succeeded", map[string]any{
		"xpath":      nameXpath,
		"name_count": len(names),
		"name_order": firstN(names, 10),
	})

	return names, nil
}

// ListNamesSharded fetches entry names using multiple sharded queries
func (br *BatchReader[E, S]) ListNamesSharded(
	ctx context.Context,
	baseXpath string,
) ([]string, error) {
	tflog.Debug(ctx, "BatchReader: starting sharded name listing", map[string]any{
		"base_xpath": baseXpath,
	})

	var allNames []string
	shardNum := 0

	for _, shard := range NameShards {
		shardNum++
		shardXpath := BuildShardedXpath(baseXpath, shard.Prefixes)

		tflog.Debug(ctx, "BatchReader: querying shard", map[string]any{
			"shard_num":  shardNum,
			"shard_name": shard.Name,
			"xpath":      shardXpath,
		})

		entries, err := br.service.ListWithXpath(ctx, shardXpath, "get", "", "")
		if err != nil {
			if sdkerrors.IsObjectNotFound(err) {
				tflog.Debug(ctx, "BatchReader: shard returned no results", map[string]any{
					"shard_num":  shardNum,
					"shard_name": shard.Name,
				})
				continue // Empty shard
			}
			tflog.Debug(ctx, "BatchReader: shard query failed", map[string]any{
				"shard_num":  shardNum,
				"shard_name": shard.Name,
				"error":      err.Error(),
			})
			return nil, &Error{
				err:     err,
				message: fmt.Sprintf("Failed to list shard %s", shard.Name),
			}
		}

		names := ExtractNames(entries)

		tflog.Debug(ctx, "BatchReader: shard query succeeded", map[string]any{
			"shard_num":  shardNum,
			"shard_name": shard.Name,
			"name_count": len(names),
			"names":      names,
		})

		allNames = append(allNames, names...)
	}

	tflog.Debug(ctx, "BatchReader: sharded listing completed", map[string]any{
		"total_shards": len(NameShards),
		"total_names":  len(allNames),
		"final_order":  firstN(allNames, 10),
	})

	return allNames, nil
}

// ListNames fetches entry names using the configured sharding strategy
func (br *BatchReader[E, S]) ListNames(
	ctx context.Context,
	baseXpath string,
) ([]string, error) {
	tflog.Debug(ctx, "BatchReader: listing names", map[string]any{
		"base_xpath": baseXpath,
		"sharding":   br.batchingConfig.ShardingStrategy,
	})

	var names []string
	var err error

	switch br.batchingConfig.ShardingStrategy {
	case ShardingEnabled:
		names, err = br.ListNamesSharded(ctx, baseXpath)
	case ShardingDisabled:
		names, err = br.ListNamesUnsharded(ctx, baseXpath)
	default:
		names, err = br.ListNamesUnsharded(ctx, baseXpath)
	}

	if err != nil {
		tflog.Debug(ctx, "BatchReader: ListNames failed", map[string]any{
			"base_xpath": baseXpath,
			"sharding":   br.batchingConfig.ShardingStrategy,
			"error":      err.Error(),
		})
		return nil, err
	}

	tflog.Debug(ctx, "BatchReader: ListNames succeeded", map[string]any{
		"base_xpath": baseXpath,
		"sharding":   br.batchingConfig.ShardingStrategy,
		"name_count": len(names),
		"name_order": firstN(names, 10),
	})

	return names, nil
}

// ReadEntriesByNames fetches a batch of entries by their names
func (br *BatchReader[E, S]) ReadEntriesByNames(
	ctx context.Context,
	baseXpath string,
	names []string,
) ([]E, error) {
	batchXpath := BuildBatchXpath(baseXpath, names)

	tflog.Debug(ctx, "BatchReader: executing batch XPath query", map[string]any{
		"xpath":      batchXpath,
		"name_count": len(names),
		"names":      names,
	})

	entries, err := br.service.ListWithXpath(ctx, batchXpath, "get", "", "")
	if err != nil {
		if sdkerrors.IsObjectNotFound(err) {
			tflog.Debug(ctx, "BatchReader: batch query returned ObjectNotFound, returning empty slice", map[string]any{
				"xpath":      batchXpath,
				"name_count": len(names),
				"names":      names,
			})
			return []E{}, nil // POTENTIAL BUG: Silent failure
		}
		tflog.Debug(ctx, "BatchReader: batch query failed", map[string]any{
			"xpath": batchXpath,
			"error": err.Error(),
		})
		return nil, &Error{err: err, message: "Failed to read entry batch"}
	}

	tflog.Debug(ctx, "BatchReader: batch query succeeded", map[string]any{
		"xpath":            batchXpath,
		"names_queried":    len(names),
		"entries_returned": len(entries),
		"entry_order":      extractEntryNames(entries),
	})

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

	tflog.Debug(ctx, "BatchReader: starting batched entry reads", map[string]any{
		"base_xpath":  baseXpath,
		"total_names": len(names),
		"batch_size":  batchSize,
		"batch_count": (len(names) + batchSize - 1) / batchSize,
	})

	var allEntries []E
	batchNum := 0

	// Process names in batches
	for i := 0; i < len(names); i += batchSize {
		end := i + batchSize
		if end > len(names) {
			end = len(names)
		}
		batchNum++

		batch := names[i:end]

		tflog.Debug(ctx, "BatchReader: reading batch", map[string]any{
			"batch_num":   batchNum,
			"batch_start": i,
			"batch_end":   end,
			"batch_size":  len(batch),
			"first_names": firstN(batch, 3),
		})

		entries, err := br.ReadEntriesByNames(ctx, baseXpath, batch)
		if err != nil {
			tflog.Debug(ctx, "BatchReader: batch read failed", map[string]any{
				"batch_num":  batchNum,
				"batch_size": len(batch),
				"error":      err.Error(),
			})
			return nil, err
		}

		tflog.Debug(ctx, "BatchReader: batch read succeeded", map[string]any{
			"batch_num":      batchNum,
			"names_in_batch": len(batch),
			"entries_read":   len(entries),
			"entry_names":    extractEntryNames(entries),
		})

		allEntries = append(allEntries, entries...)
	}

	tflog.Debug(ctx, "BatchReader: all batches completed", map[string]any{
		"total_batches": batchNum,
		"total_entries": len(allEntries),
		"entry_order":   extractEntryNames(allEntries),
	})

	return allEntries, nil
}

// ReadManyLazy implements the lazy reading strategy: list names, then batch read
func (br *BatchReader[E, S]) ReadManyLazy(
	ctx context.Context,
	baseXpath string,
) ([]E, error) {
	tflog.Debug(ctx, "BatchReader: starting lazy read", map[string]any{
		"base_xpath": baseXpath,
		"batch_size": br.batchingConfig.ReadBatchSize,
		"sharding":   br.batchingConfig.ShardingStrategy,
	})

	// Phase 1: Get list of names
	names, err := br.ListNames(ctx, baseXpath)
	if err != nil {
		tflog.Debug(ctx, "BatchReader: ListNames failed", map[string]any{
			"base_xpath": baseXpath,
			"error":      err.Error(),
		})
		return nil, err
	}

	tflog.Debug(ctx, "BatchReader: ListNames succeeded", map[string]any{
		"base_xpath":  baseXpath,
		"name_count":  len(names),
		"first_names": firstN(names, 5),
	})

	if len(names) == 0 {
		tflog.Debug(ctx, "BatchReader: no names found, returning ObjectNotFound", map[string]any{
			"base_xpath": baseXpath,
		})
		return nil, ErrObjectNotFound
	}

	// Phase 2: Fetch entries in batches
	entries, err := br.BatchReadEntries(ctx, baseXpath, names)
	if err != nil {
		tflog.Debug(ctx, "BatchReader: BatchReadEntries failed", map[string]any{
			"base_xpath": baseXpath,
			"name_count": len(names),
			"error":      err.Error(),
		})
		return nil, err
	}

	tflog.Debug(ctx, "BatchReader: lazy read completed", map[string]any{
		"base_xpath":    baseXpath,
		"names_listed":  len(names),
		"entries_read":  len(entries),
		"first_entries": firstNEntryNames(entries, 5),
	})

	return entries, nil
}

// Helper functions for logging

func firstN(items []string, n int) []string {
	if len(items) <= n {
		return items
	}
	return items[:n]
}

func extractEntryNames[E interface{ EntryName() string }](entries []E) []string {
	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.EntryName()
	}
	return names
}

func firstNEntryNames[E interface{ EntryName() string }](entries []E, n int) []string {
	names := extractEntryNames(entries)
	return firstN(names, n)
}
