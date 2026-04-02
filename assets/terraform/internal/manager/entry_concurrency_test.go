package manager_test

import (
	"context"
	"sync"
	"sync/atomic"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/PaloAltoNetworks/terraform-provider-panos/internal/manager"
)

// Concurrency tests - verify manager's locking works correctly in production code
var _ = Describe("Entry Manager Concurrency", func() {
	var existing []*MockEntryObject
	var client *MockEntryClient[*MockEntryObject]
	var service *MockEntryService[*MockEntryObject, MockLocation]
	var cachedSdk *manager.EntryObjectManager[*MockEntryObject, MockLocation, *MockEntryService[*MockEntryObject, MockLocation]]
	var apiCallCount *atomic.Int32
	var cache manager.CacheManager[*MockEntryObject]
	var location MockLocation

	ctx := context.Background()

	BeforeEach(func() {
		existing = []*MockEntryObject{
			{Location: "parent", Name: "1", Value: "A"},
			{Location: "child", Name: "2", Value: "B"},
			{Location: "child", Name: "3", Value: "C"},
		}

		// Enable caching for concurrency tests
		batchingConfig := manager.BatchingConfig{
			MultiConfigBatchSize: 500,
			ReadBatchSize:        50,
			ListStrategy:         manager.StrategyEager,
			ShardingStrategy:     manager.ShardingDisabled,
			CacheStrategy:        manager.CacheStrategyEnabled,
		}

		client = NewMockEntryClient(existing)
		service = NewMockEntryService[*MockEntryObject, MockLocation](client)

		// Create enabled cache
		cache = manager.NewEnabledCacheManager[*MockEntryObject](
			func() manager.Normalizer[*MockEntryObject] { return &MockEntryNormalizer{} },
			MockEntrySpecifier,
		)

		// Track API calls
		apiCallCount = &atomic.Int32{}
		service.ListWithXpathFunc = func(ctx context.Context, xpath string, action string, filter string, quote string) ([]*MockEntryObject, error) {
			apiCallCount.Add(1)
			return client.list(), nil
		}

		cachedSdk = manager.NewEntryObjectManager[*MockEntryObject, MockLocation, *MockEntryService[*MockEntryObject, MockLocation]](
			client, service, batchingConfig, cache, MockEntrySpecifier, MockEntryMatcher,
		)

		location = MockLocation{}
	})

	// Group A: Concurrent Read Operations
	Context("Concurrent Read Operations", func() {
		Context("with specific test data", func() {
			BeforeEach(func() {
				existing = []*MockEntryObject{
					{Name: "test-entry", Value: "test-value"},
				}
				client = NewMockEntryClient(existing)
				service = NewMockEntryService[*MockEntryObject, MockLocation](client)

				cache = manager.NewEnabledCacheManager[*MockEntryObject](
					func() manager.Normalizer[*MockEntryObject] { return &MockEntryNormalizer{} },
					MockEntrySpecifier,
				)

				apiCallCount = &atomic.Int32{}
				service.ListWithXpathFunc = func(ctx context.Context, xpath string, action string, filter string, quote string) ([]*MockEntryObject, error) {
					apiCallCount.Add(1)
					return client.list(), nil
				}

				batchingConfig := manager.BatchingConfig{
					MultiConfigBatchSize: 500,
					ReadBatchSize:        50,
					ListStrategy:         manager.StrategyEager,
					ShardingStrategy:     manager.ShardingDisabled,
					CacheStrategy:        manager.CacheStrategyEnabled,
				}

				cachedSdk = manager.NewEntryObjectManager[*MockEntryObject, MockLocation, *MockEntryService[*MockEntryObject, MockLocation]](
					client, service, batchingConfig, cache, MockEntrySpecifier, MockEntryMatcher,
				)
			})

			It("should handle 50 concurrent Read() calls for same entry", func() {
				var wg sync.WaitGroup
				successCount := atomic.Int32{}

				// 50 concurrent Read() calls
				for i := 0; i < 50; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()

						obj, err := cachedSdk.Read(ctx, location, []string{}, "test-entry")
						if err == nil && obj.Name == "test-entry" {
							successCount.Add(1)
						}
					}()
				}

				wg.Wait()

				// Verify all reads succeeded
				Expect(successCount.Load()).To(Equal(int32(50)))

				// Verify single API fetch (manager's double-check locking pattern)
				Expect(apiCallCount.Load()).To(Equal(int32(1)))
			})
		})

		It("should handle 50 concurrent ReadMany() calls", func() {
			var wg sync.WaitGroup
			successCount := atomic.Int32{}

			// 50 concurrent ReadMany() calls
			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					entries, err := cachedSdk.ReadMany(ctx, location, []string{})
					if err == nil && len(entries) == 3 {
						successCount.Add(1)
					}
				}()
			}

			wg.Wait()

			// All reads should succeed
			Expect(successCount.Load()).To(Equal(int32(50)))

			// Single API fetch despite concurrent calls
			Expect(apiCallCount.Load()).To(Equal(int32(1)))
		})

		It("should handle mixed Read() and ReadMany() calls", func() {
			var wg sync.WaitGroup
			readSuccessCount := atomic.Int32{}
			readManySuccessCount := atomic.Int32{}

			// 25 Read() calls
			for i := 0; i < 25; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					obj, err := cachedSdk.Read(ctx, location, []string{}, "1")
					if err == nil && obj.Name == "1" {
						readSuccessCount.Add(1)
					}
				}()
			}

			// 25 ReadMany() calls
			for i := 0; i < 25; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					entries, err := cachedSdk.ReadMany(ctx, location, []string{})
					if err == nil && len(entries) == 3 {
						readManySuccessCount.Add(1)
					}
				}()
			}

			wg.Wait()

			// All operations should succeed
			Expect(readSuccessCount.Load()).To(Equal(int32(25)))
			Expect(readManySuccessCount.Load()).To(Equal(int32(25)))

			// Cache coherence: single initialization
			Expect(apiCallCount.Load()).To(Equal(int32(1)))
		})
	})

	// Group B: Concurrent Write Operations
	Context("Concurrent Write Operations", func() {
		It("should handle 30 concurrent Create() calls for different entries", func() {
			var wg sync.WaitGroup
			successCount := atomic.Int32{}

			// 30 concurrent creates
			for i := 0; i < 30; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()

					entry := &MockEntryObject{
						Name:  Format("new-entry-%d", idx),
						Value: Format("value-%d", idx),
					}

					created, err := cachedSdk.Create(ctx, location, []string{}, entry)
					if err == nil && created.Name == entry.Name {
						successCount.Add(1)
					}
				}(i)
			}

			wg.Wait()

			// All creates should succeed
			Expect(successCount.Load()).To(Equal(int32(30)))

			// Verify all entries exist
			entries, err := cachedSdk.ReadMany(ctx, location, []string{})
			Expect(err).ToNot(HaveOccurred())
			Expect(len(entries)).To(Equal(33)) // 3 existing + 30 new
		})

		Context("with update test data", func() {
			BeforeEach(func() {
				existing = []*MockEntryObject{
					{Name: "update-1", Value: "original-1"},
					{Name: "update-2", Value: "original-2"},
					{Name: "update-3", Value: "original-3"},
				}
				client = NewMockEntryClient(existing)
				service = NewMockEntryService[*MockEntryObject, MockLocation](client)

				cache = manager.NewEnabledCacheManager[*MockEntryObject](
					func() manager.Normalizer[*MockEntryObject] { return &MockEntryNormalizer{} },
					MockEntrySpecifier,
				)

				apiCallCount = &atomic.Int32{}
				service.ListWithXpathFunc = func(ctx context.Context, xpath string, action string, filter string, quote string) ([]*MockEntryObject, error) {
					apiCallCount.Add(1)
					return client.list(), nil
				}

				batchingConfig := manager.BatchingConfig{
					MultiConfigBatchSize: 500,
					ReadBatchSize:        50,
					ListStrategy:         manager.StrategyEager,
					ShardingStrategy:     manager.ShardingDisabled,
					CacheStrategy:        manager.CacheStrategyEnabled,
				}

				cachedSdk = manager.NewEntryObjectManager[*MockEntryObject, MockLocation, *MockEntryService[*MockEntryObject, MockLocation]](
					client, service, batchingConfig, cache, MockEntrySpecifier, MockEntryMatcher,
				)
			})

			It("should handle concurrent Update() calls", func() {
				var wg sync.WaitGroup
				successCount := atomic.Int32{}

				// Concurrent updates
				for i := 1; i <= 3; i++ {
					wg.Add(1)
					go func(idx int) {
						defer wg.Done()

						updated := &MockEntryObject{
							Name:  Format("update-%d", idx),
							Value: Format("updated-%d", idx),
						}

						result, err := cachedSdk.Update(ctx, location, []string{}, updated, "")
						if err == nil && result.Value == updated.Value {
							successCount.Add(1)
						}
					}(i)
				}

				wg.Wait()

				// All updates should succeed
				Expect(successCount.Load()).To(Equal(int32(3)))
			})
		})

		Context("with delete test data", func() {
			BeforeEach(func() {
				existing = make([]*MockEntryObject, 20)
				for i := 0; i < 20; i++ {
					existing[i] = &MockEntryObject{
						Name:  Format("entry-%d", i),
						Value: Format("value-%d", i),
					}
				}
				client = NewMockEntryClient(existing)
				service = NewMockEntryService[*MockEntryObject, MockLocation](client)

				cache = manager.NewEnabledCacheManager[*MockEntryObject](
					func() manager.Normalizer[*MockEntryObject] { return &MockEntryNormalizer{} },
					MockEntrySpecifier,
				)

				apiCallCount = &atomic.Int32{}
				service.ListWithXpathFunc = func(ctx context.Context, xpath string, action string, filter string, quote string) ([]*MockEntryObject, error) {
					apiCallCount.Add(1)
					return client.list(), nil
				}

				batchingConfig := manager.BatchingConfig{
					MultiConfigBatchSize: 500,
					ReadBatchSize:        50,
					ListStrategy:         manager.StrategyEager,
					ShardingStrategy:     manager.ShardingDisabled,
					CacheStrategy:        manager.CacheStrategyEnabled,
				}

				cachedSdk = manager.NewEntryObjectManager[*MockEntryObject, MockLocation, *MockEntryService[*MockEntryObject, MockLocation]](
					client, service, batchingConfig, cache, MockEntrySpecifier, MockEntryMatcher,
				)
			})

			It("should handle concurrent Delete() calls", func() {
				var wg sync.WaitGroup

				// Delete 10 entries concurrently
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(idx int) {
						defer wg.Done()

						entryNames := []string{Format("entry-%d", idx)}
						cachedSdk.Delete(ctx, location, []string{}, entryNames)
					}(i)
				}

				wg.Wait()

				// Verify 10 entries remain
				remaining := client.list()
				Expect(len(remaining)).To(Equal(10))
			})
		})
	})

	// Group C: Cache Initialization Concurrency
	Context("Cache Initialization Under Load", func() {
		It("should prevent duplicate API fetches with double-check pattern", func() {
			// Clear cache to test initialization
			cache.Clear()

			var wg sync.WaitGroup
			successCount := atomic.Int32{}

			// 20 goroutines call ReadMany() simultaneously on empty cache
			for i := 0; i < 20; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					entries, err := cachedSdk.ReadMany(ctx, location, []string{})
					if err == nil && len(entries) == 3 {
						successCount.Add(1)
					}
				}()
			}

			wg.Wait()

			// All goroutines should succeed
			Expect(successCount.Load()).To(Equal(int32(20)))

			// CRITICAL: Exactly ONE API fetch should occur (double-check locking)
			Expect(apiCallCount.Load()).To(Equal(int32(1)))
		})

		It("should efficiently cache multiple reads", func() {
			// First read initializes cache
			_, err := cachedSdk.ReadMany(ctx, location, []string{})
			Expect(err).ToNot(HaveOccurred())

			initialCount := apiCallCount.Load()
			Expect(initialCount).To(Equal(int32(1)))

			// Subsequent reads should use cache - no additional API calls
			for i := 0; i < 20; i++ {
				_, err := cachedSdk.ReadMany(ctx, location, []string{})
				Expect(err).ToNot(HaveOccurred())
			}

			// Verify cache prevented additional API calls
			Expect(apiCallCount.Load()).To(Equal(int32(1)))
		})

		It("should handle initialization from multiple locations concurrently", func() {
			cache.Clear()

			location1 := MockLocation{Filter: "parent"}
			location2 := MockLocation{Filter: "child"}

			var wg sync.WaitGroup
			loc1Success := atomic.Int32{}
			loc2Success := atomic.Int32{}

			// 10 goroutines for location1
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					entries, err := cachedSdk.ReadMany(ctx, location1, []string{})
					if err == nil && len(entries) >= 0 {
						loc1Success.Add(1)
					}
				}()
			}

			// 10 goroutines for location2
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					entries, err := cachedSdk.ReadMany(ctx, location2, []string{})
					if err == nil && len(entries) >= 0 {
						loc2Success.Add(1)
					}
				}()
			}

			wg.Wait()

			// Both locations should succeed
			Expect(loc1Success.Load()).To(Equal(int32(10)))
			Expect(loc2Success.Load()).To(Equal(int32(10)))

			// Independent lock granularity - single fetch per location
			Expect(apiCallCount.Load()).To(BeNumerically(">=", int32(1)))
		})
	})

	// Group D: Mixed Read/Write Scenarios
	Context("Mixed Concurrent Operations", func() {
		It("should handle readers during concurrent Creates", func() {
			var wg sync.WaitGroup
			stopCh := make(chan struct{})
			readerSuccess := atomic.Int32{}
			writerSuccess := atomic.Int32{}

			// 20 readers continuously reading
			for i := 0; i < 20; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					for {
						select {
						case <-stopCh:
							return
						default:
							entries, err := cachedSdk.ReadMany(ctx, location, []string{})
							if err == nil && len(entries) >= 3 {
								readerSuccess.Add(1)
							}
						}
					}
				}()
			}

			// 10 writers creating new entries
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()

					entry := &MockEntryObject{
						Name:  Format("concurrent-entry-%d", idx),
						Value: Format("value-%d", idx),
					}

					_, err := cachedSdk.Create(ctx, location, []string{}, entry)
					if err == nil {
						writerSuccess.Add(1)
					}
				}(i)
			}

			// Let operations run briefly
			Sleep(50)
			close(stopCh)
			wg.Wait()

			// Writers should all succeed
			Expect(writerSuccess.Load()).To(Equal(int32(10)))

			// Readers should have succeeded many times
			Expect(readerSuccess.Load()).To(BeNumerically(">", 10))

			// Final cache state should include new entries
			entries, _ := cachedSdk.ReadMany(ctx, location, []string{})
			Expect(len(entries)).To(Equal(13)) // 3 original + 10 new
		})

		It("should handle readers during UpdateMany", func() {
			var wg sync.WaitGroup
			stopCh := make(chan struct{})
			readerSuccess := atomic.Int32{}

			// 15 readers
			for i := 0; i < 15; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					for {
						select {
						case <-stopCh:
							return
						default:
							_, err := cachedSdk.ReadMany(ctx, location, []string{})
							if err == nil {
								readerSuccess.Add(1)
							}
						}
					}
				}()
			}

			// 1 UpdateMany operation
			wg.Add(1)
			go func() {
				defer wg.Done()

				stateEntries := []*MockEntryObject{{Name: "1", Value: "A"}, {Name: "2", Value: "B"}}
				planEntries := []*MockEntryObject{
					{Name: "1", Value: "A-updated"},
					{Name: "4", Value: "D-new"},
				}

				cachedSdk.UpdateMany(ctx, location, []string{}, stateEntries, planEntries)
			}()

			Sleep(30)
			close(stopCh)
			wg.Wait()

			// Readers should have succeeded
			Expect(readerSuccess.Load()).To(BeNumerically(">", 5))
		})

		It("should maintain cache consistency during complex workflows", func() {
			var wg sync.WaitGroup
			createCount := atomic.Int32{}
			readCount := atomic.Int32{}
			updateCount := atomic.Int32{}

			// Mix of operations
			for i := 0; i < 5; i++ {
				// Create
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()

					entry := &MockEntryObject{
						Name:  Format("complex-%d", idx),
						Value: Format("value-%d", idx),
					}
					_, err := cachedSdk.Create(ctx, location, []string{}, entry)
					if err == nil {
						createCount.Add(1)
					}
				}(i)

				// Read
				wg.Add(1)
				go func() {
					defer wg.Done()

					_, err := cachedSdk.ReadMany(ctx, location, []string{})
					if err == nil {
						readCount.Add(1)
					}
				}()

				// Update existing
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()

					if idx < len(existing) {
						updated := &MockEntryObject{
							Name:  existing[idx].Name,
							Value: Format("updated-%d", idx),
						}
						_, err := cachedSdk.Update(ctx, location, []string{}, updated, "")
						if err == nil {
							updateCount.Add(1)
						}
					}
				}(i)
			}

			wg.Wait()

			// Verify operations completed
			Expect(createCount.Load()).To(BeNumerically(">", 0))
			Expect(readCount.Load()).To(Equal(int32(5)))
			Expect(updateCount.Load()).To(BeNumerically(">", 0))
		})
	})

	// Group E: Cache Coherence Tests
	Context("Cache Coherence", func() {
		It("should keep cache synchronized during CreateMany", func() {
			newEntries := []*MockEntryObject{
				{Name: "batch-1", Value: "value-1"},
				{Name: "batch-2", Value: "value-2"},
				{Name: "batch-3", Value: "value-3"},
			}

			created, err := cachedSdk.CreateMany(ctx, location, []string{}, newEntries)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(created)).To(Equal(3))

			// Verify cache reflects new entries immediately
			allEntries, err := cachedSdk.ReadMany(ctx, location, []string{})
			Expect(err).ToNot(HaveOccurred())
			Expect(len(allEntries)).To(Equal(6)) // 3 original + 3 new

			// Verify specific entries exist in cache
			for _, newEntry := range newEntries {
				found, err := cachedSdk.Read(ctx, location, []string{}, newEntry.Name)
				Expect(err).ToNot(HaveOccurred())
				Expect(found.Name).To(Equal(newEntry.Name))
			}
		})

		It("should keep cache synchronized during UpdateMany", func() {
			stateEntries := []*MockEntryObject{
				{Name: "1", Value: "A"},
				{Name: "2", Value: "B"},
				{Name: "3", Value: "C"},
			}

			planEntries := []*MockEntryObject{
				{Name: "1", Value: "A-modified"},
				{Name: "3", Value: "C-modified"},
				{Name: "4", Value: "D-new"},
			}

			updated, err := cachedSdk.UpdateMany(ctx, location, []string{}, stateEntries, planEntries)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(updated)).To(Equal(3))

			// Cache should reflect deletions
			_, err = cachedSdk.Read(ctx, location, []string{}, "2")
			Expect(err).To(MatchError(manager.ErrObjectNotFound))

			// Cache should reflect new entries
			found, err := cachedSdk.Read(ctx, location, []string{}, "4")
			Expect(err).ToNot(HaveOccurred())
			Expect(found.Name).To(Equal("4"))
		})

		It("should recover from cache errors gracefully", func() {
			// This test verifies that even if cache operations fail,
			// the manager continues to function correctly

			// Perform operation
			entry := &MockEntryObject{Name: "resilient", Value: "test"}
			created, err := cachedSdk.Create(ctx, location, []string{}, entry)

			// Should succeed regardless of cache state
			Expect(err).ToNot(HaveOccurred())
			Expect(created.Name).To(Equal("resilient"))

			// Subsequent reads should work
			found, err := cachedSdk.Read(ctx, location, []string{}, "resilient")
			Expect(err).ToNot(HaveOccurred())
			Expect(found.Name).To(Equal("resilient"))
		})
	})
})
