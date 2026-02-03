package manager

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ErrCacheDeepCopyFailed is returned when cache deep copy produces unexpected normalization result.
var ErrCacheDeepCopyFailed = errors.New("cache deep copy failed: unexpected normalization result")

// Entry interface that cache entries must satisfy.
type Entry interface {
	EntryName() string
}

// extractKeys returns the keys from a map as a slice.
func extractKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// cachedEntryWithState holds an entry along with its state index for ordering.
type cachedEntryWithState[E Entry] struct {
	Entry    E
	StateIdx int
}

// LocationCacheEntry stores all entries for a specific location.
type LocationCacheEntry[E Entry] struct {
	Entries     map[string]cachedEntryWithState[E]
	Initialized bool
}

// Normalizer extracts the normalized result from XML unmarshalling.
type Normalizer[E Entry] interface {
	Normalize() ([]E, error)
}

// CacheManager provides pure storage operations for resource entries.
// IMPORTANT: NO methods use internal locking - caller MUST handle all synchronization.
// Two implementations exist: EnabledCacheManager (actual caching) and NoOpCacheManager (passthrough).
type CacheManager[E Entry] interface {
	// IsCachingEnabled returns true if this cache manager actually performs caching
	IsCachingEnabled() bool

	// IsInitialized checks if a location has been cached
	// NOT thread-safe - caller must hold lock
	IsInitialized(locationXpath string) bool

	// SetInitialized stores entries for a location and marks it initialized
	// NOT thread-safe - caller must hold lock
	SetInitialized(ctx context.Context, locationXpath string, entries []E) error

	// Get retrieves a single entry by name
	// NOT thread-safe - caller must hold lock
	// Returns (entry, found, error)
	Get(ctx context.Context, locationXpath, name string) (E, bool, error)

	// GetAll retrieves all entries in device order
	// NOT thread-safe - caller must hold lock
	GetAll(ctx context.Context, locationXpath string) ([]E, error)

	// Put adds or updates a single entry
	// NOT thread-safe - caller must hold lock
	Put(ctx context.Context, locationXpath, name string, entry E) error

	// Delete removes an entry
	// NOT thread-safe - caller must hold lock
	Delete(ctx context.Context, locationXpath, name string) error

	// Invalidate clears the cache for a specific location
	// NOT thread-safe - caller must hold lock
	Invalidate(ctx context.Context, locationXpath string) error

	// Clear flushes the entire cache (called on provider shutdown, no locking needed)
	Clear()
}

// ResourceCache provides two-level caching: location XPath → entry name → entry.
// All methods are NOT thread-safe - caller must handle locking.
type ResourceCache[E Entry] struct {
	locations      map[string]*LocationCacheEntry[E]
	normalizerType reflect.Type
	specifier      func(E) (any, error)
}

// NewResourceCache creates a new resource cache instance.
// normalizer: template instance used to determine the normalizer type
// specifier: used for deep copy (XML marshal)
func NewResourceCache[E Entry](normalizer Normalizer[E], specifier func(E) (any, error)) *ResourceCache[E] {
	// Store the type of the normalizer so we can create fresh instances in deepCopy
	normalizerType := reflect.TypeOf(normalizer)
	if normalizerType.Kind() == reflect.Ptr {
		normalizerType = normalizerType.Elem()
	}

	return &ResourceCache[E]{
		locations:      make(map[string]*LocationCacheEntry[E]),
		normalizerType: normalizerType,
		specifier:      specifier,
	}
}

// IsInitialized checks if a location has been cached.
// NOT thread-safe - caller must hold lock.
func (c *ResourceCache[E]) IsInitialized(locationXpath string) bool {
	locEntry, found := c.locations[locationXpath]
	return found && locEntry.Initialized
}

// SetInitialized stores entries for a location and marks it initialized.
// NOT thread-safe - caller must hold lock.
func (c *ResourceCache[E]) SetInitialized(ctx context.Context, locationXpath string, entries []E) error {
	entriesMap := make(map[string]cachedEntryWithState[E], len(entries))
	for idx, entry := range entries {
		name := entry.EntryName() // Use Entry interface method
		entriesMap[name] = cachedEntryWithState[E]{
			Entry:    entry,
			StateIdx: idx,
		}
	}

	c.locations[locationXpath] = &LocationCacheEntry[E]{
		Entries:     entriesMap,
		Initialized: true,
	}

	tflog.Debug(ctx, "cache initialized", map[string]any{
		"location_xpath": locationXpath,
		"entry_count":    len(entries),
	})

	return nil
}

// Get retrieves a single entry by name with deep copy.
// NOT thread-safe - caller must hold lock.
// Returns (entry, found, error).
func (c *ResourceCache[E]) Get(ctx context.Context, locationXpath, name string) (E, bool, error) {
	locEntry, found := c.locations[locationXpath]
	if !found || !locEntry.Initialized {
		return *new(E), false, nil
	}

	ews, found := locEntry.Entries[name]
	if !found {
		return *new(E), false, nil
	}

	copied, err := c.deepCopy(ctx, ews.Entry)
	if err != nil {
		return *new(E), false, err
	}

	return copied, true, nil
}

// GetAll retrieves all entries in device order with deep copies.
// NOT thread-safe - caller must hold lock.
func (c *ResourceCache[E]) GetAll(ctx context.Context, locationXpath string) ([]E, error) {
	locEntry, found := c.locations[locationXpath]
	if !found || !locEntry.Initialized {
		return nil, nil
	}

	// Collect entries with indices
	entriesWithIdx := make([]cachedEntryWithState[E], 0, len(locEntry.Entries))
	for _, ews := range locEntry.Entries {
		entriesWithIdx = append(entriesWithIdx, ews)
	}

	// Sort by StateIdx to preserve device order
	// Using a simple sort since Go doesn't have sort.Slice for generic types easily
	for i := 0; i < len(entriesWithIdx)-1; i++ {
		for j := i + 1; j < len(entriesWithIdx); j++ {
			if entriesWithIdx[i].StateIdx > entriesWithIdx[j].StateIdx {
				entriesWithIdx[i], entriesWithIdx[j] = entriesWithIdx[j], entriesWithIdx[i]
			}
		}
	}

	// Extract entries in order and deep copy
	result := make([]E, 0, len(entriesWithIdx))
	for _, ews := range entriesWithIdx {
		copied, err := c.deepCopy(ctx, ews.Entry)
		if err != nil {
			return nil, err
		}
		result = append(result, copied)
	}

	return result, nil
}

// Put adds or updates a single entry in the cache.
// NOT thread-safe - caller must hold lock.
func (c *ResourceCache[E]) Put(ctx context.Context, locationXpath, name string, entry E) error {
	locEntry, found := c.locations[locationXpath]
	if !found {
		locEntry = &LocationCacheEntry[E]{
			Entries:     make(map[string]cachedEntryWithState[E]),
			Initialized: false,
		}
		c.locations[locationXpath] = locEntry
	}

	if existing, exists := locEntry.Entries[name]; exists {
		// Update existing, preserve StateIdx
		locEntry.Entries[name] = cachedEntryWithState[E]{
			Entry:    entry,
			StateIdx: existing.StateIdx,
		}
	} else {
		// New entry, assign next StateIdx
		maxIdx := -1
		for _, ews := range locEntry.Entries {
			if ews.StateIdx > maxIdx {
				maxIdx = ews.StateIdx
			}
		}
		locEntry.Entries[name] = cachedEntryWithState[E]{
			Entry:    entry,
			StateIdx: maxIdx + 1,
		}
	}

	return nil
}

// Delete removes an entry from the cache.
// NOT thread-safe - caller must hold lock.
func (c *ResourceCache[E]) Delete(ctx context.Context, locationXpath, name string) error {
	locEntry, found := c.locations[locationXpath]
	if !found {
		return nil
	}

	delete(locEntry.Entries, name)
	return nil
}

// Invalidate clears the cache for a specific location.
// NOT thread-safe - caller must hold lock.
func (c *ResourceCache[E]) Invalidate(ctx context.Context, locationXpath string) error {
	delete(c.locations, locationXpath)
	return nil
}

// Clear flushes the entire cache.
func (c *ResourceCache[E]) Clear() {
	// Note: This doesn't acquire locks, as it's meant for provider shutdown/reset.
	// For production use, you might want to iterate through locations and lock each.
	c.locations = make(map[string]*LocationCacheEntry[E])
}

// deepCopy creates a deep copy via XML marshal/unmarshal.
func (c *ResourceCache[E]) deepCopy(ctx context.Context, entry E) (E, error) {
	tflog.Debug(ctx, "cache: starting deep copy")

	// Phase 1: Marshal entry to XML via Specifier
	xmlNode, err := c.specifier(entry)
	if err != nil {
		tflog.Debug(ctx, "cache: specifier failed", map[string]any{
			"error": err.Error(),
		})
		return *new(E), err
	}

	tflog.Debug(ctx, "cache: specifier returned", map[string]any{
		"node_type": fmt.Sprintf("%T", xmlNode),
	})

	// Wrap single entry in a container structure to match Normalizer expectations.
	// The Normalizer expects XML with <entry> elements as children of a parent element.
	// We create a generic wrapper struct that will marshal to the expected format.
	type wrapper struct {
		Entries []any `xml:"entry"`
	}

	container := wrapper{
		Entries: []any{xmlNode},
	}

	xmlBytes, err := xml.Marshal(container)
	if err != nil {
		tflog.Debug(ctx, "cache: marshal failed", map[string]any{
			"error": err.Error(),
		})
		return *new(E), err
	}

	tflog.Debug(ctx, "cache: marshaled to XML", map[string]any{
		"xml_length": len(xmlBytes),
		"xml_preview": string(xmlBytes),
	})

	// Phase 2: Unmarshal XML via Normalizer
	// Create a fresh normalizer instance to avoid race conditions
	normalizerPtr := reflect.New(c.normalizerType)
	normalizer := normalizerPtr.Interface().(Normalizer[E])

	err = xml.Unmarshal(xmlBytes, normalizer)
	if err != nil {
		tflog.Debug(ctx, "cache: unmarshal failed", map[string]any{
			"error": err.Error(),
		})
		return *new(E), err
	}

	normalized, err := normalizer.Normalize()
	if err != nil {
		tflog.Debug(ctx, "cache: normalize failed", map[string]any{
			"error": err.Error(),
		})
		return *new(E), err
	}

	tflog.Debug(ctx, "cache: normalize returned", map[string]any{
		"entry_count": len(normalized),
	})

	if len(normalized) != 1 {
		tflog.Debug(ctx, "cache: deep copy failed - unexpected entry count", map[string]any{
			"expected": 1,
			"actual":   len(normalized),
		})
		return *new(E), ErrCacheDeepCopyFailed
	}

	tflog.Debug(ctx, "cache: deep copy succeeded")

	return normalized[0], nil
}

// EnabledCacheManager wraps ResourceCache to provide actual caching functionality.
type EnabledCacheManager[E Entry] struct {
	cache *ResourceCache[E]
}

// NewEnabledCacheManager creates a new enabled cache manager.
func NewEnabledCacheManager[E Entry](normalizer Normalizer[E], specifier func(E) (any, error)) *EnabledCacheManager[E] {
	cache := NewResourceCache[E](normalizer, specifier)
	return &EnabledCacheManager[E]{cache: cache}
}

func (m *EnabledCacheManager[E]) IsCachingEnabled() bool {
	return true
}

func (m *EnabledCacheManager[E]) IsInitialized(locationXpath string) bool {
	return m.cache.IsInitialized(locationXpath)
}

func (m *EnabledCacheManager[E]) SetInitialized(ctx context.Context, locationXpath string, entries []E) error {
	return m.cache.SetInitialized(ctx, locationXpath, entries)
}

func (m *EnabledCacheManager[E]) Get(ctx context.Context, locationXpath, name string) (E, bool, error) {
	return m.cache.Get(ctx, locationXpath, name)
}

func (m *EnabledCacheManager[E]) GetAll(ctx context.Context, locationXpath string) ([]E, error) {
	return m.cache.GetAll(ctx, locationXpath)
}

func (m *EnabledCacheManager[E]) Put(ctx context.Context, locationXpath, name string, entry E) error {
	return m.cache.Put(ctx, locationXpath, name, entry)
}

func (m *EnabledCacheManager[E]) Delete(ctx context.Context, locationXpath, name string) error {
	return m.cache.Delete(ctx, locationXpath, name)
}

func (m *EnabledCacheManager[E]) Invalidate(ctx context.Context, locationXpath string) error {
	return m.cache.Invalidate(ctx, locationXpath)
}

func (m *EnabledCacheManager[E]) Clear() {
	m.cache.Clear()
}

// NoOpCacheManager provides a passthrough implementation that performs no caching.
type NoOpCacheManager[E Entry] struct{}

// NewNoOpCacheManager creates a new no-op cache manager.
func NewNoOpCacheManager[E Entry]() *NoOpCacheManager[E] {
	return &NoOpCacheManager[E]{}
}

func (m *NoOpCacheManager[E]) IsCachingEnabled() bool {
	return false
}

func (m *NoOpCacheManager[E]) IsInitialized(locationXpath string) bool {
	return false
}

func (m *NoOpCacheManager[E]) SetInitialized(ctx context.Context, locationXpath string, entries []E) error {
	return nil // No-op, always succeeds
}

func (m *NoOpCacheManager[E]) Get(ctx context.Context, locationXpath, name string) (E, bool, error) {
	return *new(E), false, nil
}

func (m *NoOpCacheManager[E]) GetAll(ctx context.Context, locationXpath string) ([]E, error) {
	return nil, nil
}

func (m *NoOpCacheManager[E]) Put(ctx context.Context, locationXpath, name string, entry E) error {
	return nil
}

func (m *NoOpCacheManager[E]) Delete(ctx context.Context, locationXpath, name string) error {
	return nil
}

func (m *NoOpCacheManager[E]) Invalidate(ctx context.Context, locationXpath string) error {
	return nil
}

func (m *NoOpCacheManager[E]) Clear() {
	// Nothing to clear
}
