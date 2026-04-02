package manager

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// ErrCacheDeepCopyFailed is returned when cache deep copy produces unexpected normalization result.
var ErrCacheDeepCopyFailed = errors.New("cache deep copy failed: unexpected normalization result")

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

// CacheManager provides caching for PAN-OS resources with external locking.
//
// CRITICAL LOCKING REQUIREMENT:
//
// ALL methods in this interface require external locking. Callers MUST acquire
// appropriate locks before calling any cache method:
//   - RLock() for READ operations: Get, GetAll, IsInitialized
//   - Lock() for WRITE operations: Put, Delete, SetInitialized, Invalidate, Clear
//
// Calling without locks WILL CAUSE DATA RACES.
//
// External Locking Model:
// This cache uses external locking where callers control lock acquisition and
// release. This design allows:
//   - Lock scope optimization (minimize critical sections)
//   - Coordination of multiple operations under single lock
//   - Flexibility in lock granularity
//
// The trade-off is that callers bear responsibility for thread-safety.
//
// Lock Type Requirements:
//
//   READ Operations (require RLock):
//     - Get(...)          // Read single entry
//     - GetAll(...)       // Read all entries
//     - IsInitialized(...) // Check initialization status
//
//   WRITE Operations (require Lock):
//     - Put(...)          // Add/update entry
//     - Delete(...)       // Remove entry
//     - SetInitialized(...) // Initialize location
//     - Invalidate(...)   // Clear specific location
//     - Clear()           // Remove all entries
//
// Multiple readers can hold RLock concurrently. Writes require exclusive Lock.
//
// Standard Locking Patterns:
//
// Read Pattern (concurrent-safe):
//
//	mu := locking.GetRWMutex(locking.XpathLockCategory, locationXpath)
//	mu.RLock()
//	entry, found, err := cache.Get(ctx, locationXpath, name)
//	mu.RUnlock()
//
// Write Pattern (exclusive):
//
//	mu := locking.GetRWMutex(locking.XpathLockCategory, locationXpath)
//	mu.Lock()
//	defer mu.Unlock()  // Always use defer for writes
//	err := cache.Put(ctx, locationXpath, name, entry)
//
// Double-Check Pattern (initialization):
//
//	mu := locking.GetRWMutex(locking.XpathLockCategory, locationXpath)
//
//	// Fast path: check with RLock
//	mu.RLock()
//	if cache.IsInitialized(locationXpath) {
//	    entries, _ := cache.GetAll(ctx, locationXpath)
//	    mu.RUnlock()
//	    return entries
//	}
//	mu.RUnlock()
//
//	// Slow path: initialize with Lock
//	mu.Lock()
//	defer mu.Unlock()
//	if cache.IsInitialized(locationXpath) {  // Double-check
//	    return cache.GetAll(ctx, locationXpath)
//	}
//	entries := fetchFromAPI()
//	cache.SetInitialized(ctx, locationXpath, entries)
//	return entries
//
// Two implementations exist: EnabledCacheManager (actual caching) and NoOpCacheManager (passthrough).
type CacheManager[E Entry] interface {
	// IsCachingEnabled returns true if this cache manager actually performs caching
	IsCachingEnabled() bool

	// IsInitialized checks if a location has been cached.
	//
	// IMPORTANT: Caller MUST hold at least RLock before calling.
	//
	// Returns true if the location has been initialized, false otherwise.
	IsInitialized(locationXpath string) bool

	// SetInitialized stores entries for a location and marks it initialized.
	//
	// IMPORTANT: Caller MUST hold Lock (exclusive write lock) before calling.
	//
	// This is a write operation that replaces all entries for the location.
	SetInitialized(ctx context.Context, locationXpath string, entries []E) error

	// Get retrieves a single entry by name with deep copy for safe mutation.
	//
	// IMPORTANT: Caller MUST hold at least RLock before calling.
	//
	// Return values disambiguate cache miss from errors:
	//   - (entry, true, nil)  = found in cache (success)
	//   - (zero, false, nil)  = not found in cache (normal cache miss, not an error)
	//   - (zero, false, error) = internal error (deep copy failed)
	//
	// Cache misses return false (not an error) because missing entries are expected
	// during normal operation. Callers should check the bool flag:
	//   if !found { /* handle miss */ }
	//
	// The zero value is returned when not found to maintain type safety, but
	// callers should only use the entry when found==true.
	Get(ctx context.Context, locationXpath, name string) (E, bool, error)

	// GetAll retrieves all entries in device order.
	//
	// IMPORTANT: Caller MUST hold at least RLock before calling.
	//
	// Returns entries in the order they were received from the device.
	GetAll(ctx context.Context, locationXpath string) ([]E, error)

	// Put adds or updates a single entry in the cache.
	//
	// IMPORTANT: Caller MUST hold Lock (exclusive write lock) before calling.
	//
	// IMPORTANT: Location must be initialized first via SetInitialized().
	// Calling Put() on an uninitialized location returns ErrLocationCacheNotInitialized.
	//
	// This design ensures cache coherence - all entries for a location must be loaded
	// together via SetInitialized() before individual updates. This prevents partial
	// cache states where GetAll() would return incomplete results.
	//
	// Typical usage:
	//   1. Call SetInitialized(location, allEntries) to populate cache
	//   2. Call Put(location, name, entry) to update individual entries after mutations
	//
	// This is a write operation that modifies cache state.
	Put(ctx context.Context, locationXpath, name string, entry E) error

	// Delete removes an entry.
	//
	// IMPORTANT: Caller MUST hold Lock (exclusive write lock) before calling.
	//
	// This is a write operation that modifies cache state.
	Delete(ctx context.Context, locationXpath, name string) error

	// Invalidate clears the cache for a specific location.
	//
	// IMPORTANT: Caller MUST hold Lock (exclusive write lock) before calling.
	//
	// This is a write operation that removes all entries for the location.
	Invalidate(ctx context.Context, locationXpath string) error

	// Clear removes all cached data from the cache.
	// CRITICAL: Caller MUST hold write locks for ALL locations before calling.
	// See ResourceCache.Clear() for detailed documentation and usage examples.
	Clear()
}

// ResourceCache provides two-level caching: location XPath → entry name → entry.
// All methods are NOT thread-safe - caller must handle locking.
type ResourceCache[E Entry] struct {
	locations  map[string]*LocationCacheEntry[E]
	marshaller Marshaller[E]
}

// NewResourceCache creates a new resource cache instance.
func NewResourceCache[E Entry, M Marshaller[E]](
	marshaller M,
) *ResourceCache[E] {
	return &ResourceCache[E]{
		locations:  make(map[string]*LocationCacheEntry[E]),
		marshaller: marshaller,
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
		name := entry.EntryName()
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
	slices.SortFunc(entriesWithIdx, func(a, b cachedEntryWithState[E]) int {
		return a.StateIdx - b.StateIdx
	})

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
		return ErrLocationCacheNotInitialized
	}

	if !locEntry.Initialized {
		return ErrLocationCacheNotInitialized
	}

	if existing, exists := locEntry.Entries[name]; exists {
		locEntry.Entries[name] = cachedEntryWithState[E]{
			Entry:    entry,
			StateIdx: existing.StateIdx,
		}
	} else {
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

// Clear removes all cached data from this resource cache.
//
// CRITICAL: Caller MUST hold write locks for ALL locations before calling.
// This method does NOT acquire locks internally. Calling Clear() without
// proper locking will cause DATA RACES and potential CRASHES.
//
// External Locking Model:
// The cache uses an external locking model where callers are responsible
// for acquiring appropriate locks before cache operations. This allows
// callers to optimize lock scope and coordinate multiple operations.
//
// For Clear(), the caller must hold write locks (Lock, not RLock) for
// ALL cached locations before calling. Since Clear() affects all locations,
// partial locking is NOT sufficient.
//
// Typical Usage Context:
//   - Provider shutdown: Clear cache when provider terminates
//   - Test cleanup: Reset cache state between test cases
//   - Emergency reset: Clear corrupted cache (rare)
//
// Usage Example (Test Cleanup):
//
//	// In test cleanup, ensure no concurrent access possible
//	cache.Clear()  // Safe - single-threaded test teardown
//
// INCORRECT Usage:
//
//	cache.Clear()  // UNSAFE - no lock held, concurrent access possible
//
// CORRECT Usage (if concurrent access possible):
//
//	// Acquire write locks for ALL locations first
//	for _, locationXpath := range allLocations {
//	    mu := locking.GetRWMutex(locking.XpathLockCategory, locationXpath)
//	    mu.Lock()
//	    defer mu.Unlock()
//	}
//	cache.Clear()  // Safe - all locations locked
//
// See CacheManager interface documentation for detailed locking requirements.
func (c *ResourceCache[E]) Clear() {
	// Note: Clear affects ALL locations, so we can't validate individual location locks here.
	// The caller is responsible for holding locks for all locations before calling.
	// In the future, we could validate that at least some locks are held.
	c.locations = make(map[string]*LocationCacheEntry[E])
}

// deepCopy creates a deep copy via XML marshal/unmarshal.
func (c *ResourceCache[E]) deepCopy(ctx context.Context, entry E) (E, error) {
	tflog.Debug(ctx, "cache: starting deep copy")

	// Phase 1: Marshal entry to XML via marshaller
	xmlNode, err := c.marshaller.Specify(entry)
	if err != nil {
		tflog.Trace(ctx, "cache: specifier failed", map[string]any{
			"error": err.Error(),
		})
		return *new(E), err
	}

	tflog.Trace(ctx, "cache: specifier returned", map[string]any{
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
		tflog.Trace(ctx, "cache: marshal failed", map[string]any{
			"error": err.Error(),
		})
		return *new(E), err
	}

	tflog.Trace(ctx, "cache: marshaled to XML", map[string]any{
		"xml_length": len(xmlBytes),
		"xml_preview": string(xmlBytes),
	})

	// Phase 2: Unmarshal XML via marshaller.NewNormalizer
	// Create a fresh normalizer instance to avoid race conditions
	// Type assert from any to Normalizer[E] - safe because SDK generates compatible types
	normalizer := c.marshaller.NewNormalizer().(Normalizer[E])

	err = xml.Unmarshal(xmlBytes, normalizer)
	if err != nil {
		tflog.Trace(ctx, "cache: unmarshal failed", map[string]any{
			"error": err.Error(),
		})
		return *new(E), err
	}

	normalized, err := normalizer.Normalize()
	if err != nil {
		tflog.Trace(ctx, "cache: normalize failed", map[string]any{
			"error": err.Error(),
		})
		return *new(E), err
	}

	tflog.Trace(ctx, "cache: normalize returned", map[string]any{
		"entry_count": len(normalized),
	})

	if len(normalized) != 1 {
		tflog.Trace(ctx, "cache: deep copy failed - unexpected entry count", map[string]any{
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
func NewEnabledCacheManager[E Entry, M Marshaller[E]](
	marshaller M,
) *EnabledCacheManager[E] {
	cache := NewResourceCache[E](marshaller)
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
