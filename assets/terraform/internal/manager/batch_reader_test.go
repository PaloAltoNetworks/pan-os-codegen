package manager_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/PaloAltoNetworks/terraform-provider-panos/internal/manager"
)

var _ = Describe("BatchReader Helper Functions", func() {
	Describe("ExtractNames", func() {
		It("should extract names from empty slice", func() {
			entries := []*MockEntryObject{}
			names := manager.ExtractNames(entries)
			Expect(names).To(BeEmpty())
		})

		It("should extract names from populated slice", func() {
			entries := []*MockEntryObject{
				{Name: "entry1"},
				{Name: "entry2"},
				{Name: "entry3"},
			}
			names := manager.ExtractNames(entries)
			Expect(names).To(Equal([]string{"entry1", "entry2", "entry3"}))
		})
	})

	Describe("BuildShardedXpath", func() {
		It("should build xpath with single prefix", func() {
			xpath := manager.BuildShardedXpath("/test/path", []string{"a"})
			Expect(xpath).To(Equal("/test/path[starts-with(@name, 'a')]/@name"))
		})

		It("should build xpath with multiple prefixes", func() {
			xpath := manager.BuildShardedXpath("/test/path", []string{"a", "b", "c"})
			Expect(xpath).To(Equal("/test/path[starts-with(@name, 'a') or starts-with(@name, 'b') or starts-with(@name, 'c')]/@name"))
		})

		It("should handle empty prefix list", func() {
			xpath := manager.BuildShardedXpath("/test/path", []string{})
			Expect(xpath).To(Equal("/test/path[]/@name"))
		})
	})

	Describe("BuildBatchXpath", func() {
		It("should build xpath with single name", func() {
			xpath := manager.BuildBatchXpath("/test/path/entry", []string{"name1"})
			Expect(xpath).To(Equal("/test/path/entry[@name='name1']"))
		})

		It("should build xpath with multiple names", func() {
			xpath := manager.BuildBatchXpath("/test/path/entry", []string{"name1", "name2", "name3"})
			Expect(xpath).To(Equal("/test/path/entry[@name='name1' or @name='name2' or @name='name3']"))
		})

		It("should handle empty names list", func() {
			xpath := manager.BuildBatchXpath("/test/path/entry", []string{})
			Expect(xpath).To(Equal("/test/path/entry[]"))
		})
	})

	Describe("FilterEntriesByLocation", func() {
		Context("when location has no filter", func() {
			It("should return all entries", func() {
				location := MockLocation{Filter: ""}
				entries := []*MockEntryObject{
					{Name: "entry1"},
					{Name: "entry2"},
				}
				filtered := manager.FilterEntriesByLocation(location, entries)
				Expect(filtered).To(Equal(entries))
			})
		})

		Context("when location has a filter", func() {
			It("should return entries matching the filter", func() {
				location := MockLocation{Filter: "vsys1"}
				entries := []*MockEntryObject{
					{
						Name:     "entry1",
						Location: "vsys1",
					},
					{
						Name:     "entry2",
						Location: "vsys2",
					},
					{
						Name:     "entry3",
						Location: "vsys1",
					},
				}
				filtered := manager.FilterEntriesByLocation(location, entries)
				Expect(filtered).To(HaveLen(2))
				Expect(filtered[0].Name).To(Equal("entry1"))
				Expect(filtered[1].Name).To(Equal("entry3"))
			})

			It("should return entries with no loc attribute", func() {
				location := MockLocation{Filter: "vsys1"}
				entries := []*MockEntryObject{
					{
						Name:     "entry1",
						Location: "",
					},
					{
						Name:     "entry2",
						Location: "vsys2",
					},
				}
				filtered := manager.FilterEntriesByLocation(location, entries)
				Expect(filtered).To(HaveLen(1))
				Expect(filtered[0].Name).To(Equal("entry1"))
			})
		})
	})
})

var _ = Describe("BatchReader with MockEntryService", func() {
	var (
		client  *MockEntryClient[*MockEntryObject]
		service *MockEntryService[*MockEntryObject, MockLocation]
		reader  *manager.BatchReader[*MockEntryObject, *MockEntryService[*MockEntryObject, MockLocation]]
		ctx     context.Context
	)

	BeforeEach(func() {
		// Initialize with test data covering various naming patterns
		initial := []*MockEntryObject{
			{Name: "admin", Value: "value1"},
			{Name: "user1", Value: "value2"},
			{Name: "backup", Value: "value3"},
			{Name: "1-config", Value: "value4"},
			{Name: "5-test", Value: "value5"},
		}
		client = NewMockEntryClient(initial)
		service = NewMockEntryService[*MockEntryObject, MockLocation](client)
		ctx = context.Background()
	})

	Describe("ListNamesUnsharded", func() {
		It("should return all names when service succeeds", func() {
			config := manager.BatchingConfig{}
			reader = manager.NewBatchReader(service, config)

			names, err := reader.ListNamesUnsharded(ctx, "/some/location/entry")
			Expect(err).ToNot(HaveOccurred())
			Expect(names).To(ConsistOf("admin", "user1", "backup", "1-config", "5-test"))
		})

		It("should return empty slice when no entries exist", func() {
			// Create empty client
			emptyClient := NewMockEntryClient([]*MockEntryObject{})
			emptyService := NewMockEntryService[*MockEntryObject, MockLocation](emptyClient)

			config := manager.BatchingConfig{}
			reader = manager.NewBatchReader(emptyService, config)

			names, err := reader.ListNamesUnsharded(ctx, "/some/location/entry")
			Expect(err).ToNot(HaveOccurred())
			Expect(names).To(BeEmpty())
		})
	})

	Describe("ListNamesSharded", func() {
		It("should fetch from all shards and combine results", func() {
			config := manager.BatchingConfig{}
			reader = manager.NewBatchReader(service, config)

			names, err := reader.ListNamesSharded(ctx, "/some/location/entry")
			Expect(err).ToNot(HaveOccurred())
			Expect(names).To(ConsistOf("admin", "user1", "backup", "1-config", "5-test"))
		})

		It("should handle entries starting with numbers", func() {
			config := manager.BatchingConfig{}
			reader = manager.NewBatchReader(service, config)

			names, err := reader.ListNamesSharded(ctx, "/some/location/entry")
			Expect(err).ToNot(HaveOccurred())
			Expect(names).To(ContainElements("1-config", "5-test"))
		})

		It("should return empty when no entries exist", func() {
			emptyClient := NewMockEntryClient([]*MockEntryObject{})
			emptyService := NewMockEntryService[*MockEntryObject, MockLocation](emptyClient)

			config := manager.BatchingConfig{}
			reader = manager.NewBatchReader(emptyService, config)

			names, err := reader.ListNamesSharded(ctx, "/some/location/entry")
			Expect(err).ToNot(HaveOccurred())
			Expect(names).To(BeEmpty())
		})
	})

	Describe("ListNames", func() {
		It("should use unsharded when sharding disabled", func() {
			config := manager.BatchingConfig{ShardingStrategy: manager.ShardingDisabled}
			reader = manager.NewBatchReader(service, config)

			names, err := reader.ListNames(ctx, "/some/location/entry")
			Expect(err).ToNot(HaveOccurred())
			Expect(names).To(ConsistOf("admin", "user1", "backup", "1-config", "5-test"))
		})

		It("should use sharded when sharding enabled", func() {
			config := manager.BatchingConfig{ShardingStrategy: manager.ShardingEnabled}
			reader = manager.NewBatchReader(service, config)

			names, err := reader.ListNames(ctx, "/some/location/entry")
			Expect(err).ToNot(HaveOccurred())
			Expect(names).To(ConsistOf("admin", "user1", "backup", "1-config", "5-test"))
		})
	})

	Describe("ReadEntriesByNames", func() {
		It("should filter entries by names using OR predicates", func() {
			config := manager.BatchingConfig{}
			reader = manager.NewBatchReader(service, config)

			entries, err := reader.ReadEntriesByNames(ctx, "/some/location", []string{"admin", "backup"})
			Expect(err).ToNot(HaveOccurred())
			Expect(entries).To(HaveLen(2))
			Expect(entries[0].EntryName()).To(Equal("admin"))
			Expect(entries[1].EntryName()).To(Equal("backup"))
		})

		It("should return empty slice when names not found", func() {
			config := manager.BatchingConfig{}
			reader = manager.NewBatchReader(service, config)

			entries, err := reader.ReadEntriesByNames(ctx, "/some/location", []string{"nonexistent"})
			Expect(err).ToNot(HaveOccurred())
			Expect(entries).To(BeEmpty())
		})

		It("should handle single name", func() {
			config := manager.BatchingConfig{}
			reader = manager.NewBatchReader(service, config)

			entries, err := reader.ReadEntriesByNames(ctx, "/some/location", []string{"admin"})
			Expect(err).ToNot(HaveOccurred())
			Expect(entries).To(HaveLen(1))
			Expect(entries[0].EntryName()).To(Equal("admin"))
		})
	})

	Describe("BatchReadEntries", func() {
		It("should read in batches", func() {
			config := manager.BatchingConfig{ReadBatchSize: 2}
			reader = manager.NewBatchReader(service, config)

			entries, err := reader.BatchReadEntries(ctx, "/some/location", []string{"admin", "user1", "backup"})
			Expect(err).ToNot(HaveOccurred())
			Expect(entries).To(HaveLen(3))
			Expect(entries[0].EntryName()).To(Equal("admin"))
			Expect(entries[1].EntryName()).To(Equal("user1"))
			Expect(entries[2].EntryName()).To(Equal("backup"))
		})

		It("should use default batch size when not configured", func() {
			config := manager.BatchingConfig{} // ReadBatchSize = 0, uses default
			reader = manager.NewBatchReader(service, config)

			entries, err := reader.BatchReadEntries(ctx, "/some/location", []string{"admin", "user1"})
			Expect(err).ToNot(HaveOccurred())
			Expect(entries).To(HaveLen(2))
		})

		It("should handle empty names list", func() {
			config := manager.BatchingConfig{}
			reader = manager.NewBatchReader(service, config)

			entries, err := reader.BatchReadEntries(ctx, "/some/location", []string{})
			Expect(err).ToNot(HaveOccurred())
			Expect(entries).To(BeEmpty())
		})
	})

	Describe("ReadManyLazy", func() {
		It("should list names then batch read", func() {
			config := manager.BatchingConfig{ShardingStrategy: manager.ShardingDisabled}
			reader = manager.NewBatchReader(service, config)

			entries, err := reader.ReadManyLazy(ctx, "/some/location/entry")
			Expect(err).ToNot(HaveOccurred())
			Expect(entries).To(HaveLen(5))
			Expect(entries[0].EntryName()).To(Equal("admin"))
		})

		It("should return ErrObjectNotFound when no entries exist", func() {
			emptyClient := NewMockEntryClient([]*MockEntryObject{})
			emptyService := NewMockEntryService[*MockEntryObject, MockLocation](emptyClient)

			config := manager.BatchingConfig{}
			reader = manager.NewBatchReader(emptyService, config)

			entries, err := reader.ReadManyLazy(ctx, "/some/location/entry")
			Expect(err).To(Equal(manager.ErrObjectNotFound))
			Expect(entries).To(BeNil())
		})

		It("should work with sharding enabled", func() {
			config := manager.BatchingConfig{ShardingStrategy: manager.ShardingEnabled}
			reader = manager.NewBatchReader(service, config)

			entries, err := reader.ReadManyLazy(ctx, "/some/location/entry")
			Expect(err).ToNot(HaveOccurred())
			Expect(entries).To(HaveLen(5))
		})
	})

	Describe("FilterEntriesByLocation integration", func() {
		It("should filter entries by location attribute", func() {
			// Create entries with different locations
			locatedEntries := []*MockEntryObject{
				{Name: "entry1", Location: "vsys1"},
				{Name: "entry2", Location: "vsys2"},
				{Name: "entry3", Location: "vsys1"},
			}
			locatedClient := NewMockEntryClient(locatedEntries)
			locatedService := NewMockEntryService[*MockEntryObject, MockLocation](locatedClient)

			config := manager.BatchingConfig{}
			reader = manager.NewBatchReader(locatedService, config)

			// Read all entries
			allEntries, err := reader.ReadManyLazy(ctx, "/some/location/entry")
			Expect(err).ToNot(HaveOccurred())
			Expect(allEntries).To(HaveLen(3))

			// Filter by location
			location := MockLocation{Filter: "vsys1"}
			filtered := manager.FilterEntriesByLocation(location, allEntries)
			Expect(filtered).To(HaveLen(2))
			Expect(filtered[0].Name).To(Equal("entry1"))
			Expect(filtered[1].Name).To(Equal("entry3"))
		})
	})
})

var _ = Describe("BatchReader with MockUuidService", func() {
	var (
		client  *MockUuidClient[*MockUuidObject]
		service *MockUuidService[*MockUuidObject, MockLocation]
		reader  *manager.BatchReader[*MockUuidObject, *MockUuidService[*MockUuidObject, MockLocation]]
		ctx     context.Context
	)

	BeforeEach(func() {
		// Initialize with test data
		initial := []*MockUuidObject{
			{Name: "admin", Value: "value1"},
			{Name: "user1", Value: "value2"},
			{Name: "backup", Value: "value3"},
			{Name: "1-config", Value: "value4"},
		}
		client = NewMockUuidClient(initial)
		service = NewMockUuidService[*MockUuidObject, MockLocation](client)
		ctx = context.Background()
	})

	Describe("ListNamesUnsharded", func() {
		It("should return all names when service succeeds", func() {
			config := manager.BatchingConfig{}
			reader = manager.NewBatchReader(service, config)

			names, err := reader.ListNamesUnsharded(ctx, "/some/location/entry")
			Expect(err).ToNot(HaveOccurred())
			Expect(names).To(ConsistOf("admin", "user1", "backup", "1-config"))
		})
	})

	Describe("ListNamesSharded", func() {
		It("should fetch from all shards and combine results", func() {
			config := manager.BatchingConfig{}
			reader = manager.NewBatchReader(service, config)

			names, err := reader.ListNamesSharded(ctx, "/some/location/entry")
			Expect(err).ToNot(HaveOccurred())
			Expect(names).To(ConsistOf("admin", "user1", "backup", "1-config"))
		})
	})

	Describe("ReadEntriesByNames", func() {
		It("should filter entries by names using OR predicates", func() {
			config := manager.BatchingConfig{}
			reader = manager.NewBatchReader(service, config)

			entries, err := reader.ReadEntriesByNames(ctx, "/some/location", []string{"admin", "backup"})
			Expect(err).ToNot(HaveOccurred())
			Expect(entries).To(HaveLen(2))
			Expect(entries[0].EntryName()).To(Equal("admin"))
			Expect(entries[1].EntryName()).To(Equal("backup"))
		})
	})

	Describe("ReadManyLazy", func() {
		It("should list names then batch read", func() {
			config := manager.BatchingConfig{ShardingStrategy: manager.ShardingDisabled}
			reader = manager.NewBatchReader(service, config)

			entries, err := reader.ReadManyLazy(ctx, "/some/location/entry")
			Expect(err).ToNot(HaveOccurred())
			Expect(entries).To(HaveLen(4))
		})

		It("should verify UUIDs are present", func() {
			config := manager.BatchingConfig{}
			reader = manager.NewBatchReader(service, config)

			entries, err := reader.ReadManyLazy(ctx, "/some/location/entry")
			Expect(err).ToNot(HaveOccurred())

			// Verify all entries have UUIDs
			for _, entry := range entries {
				Expect(entry.EntryUuid()).ToNot(BeNil())
				Expect(*entry.EntryUuid()).ToNot(BeEmpty())
			}
		})
	})
})
