package manager_test

import (
	"context"
	"encoding/xml"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/PaloAltoNetworks/terraform-provider-panos/internal/manager"
)

// Test entry type
type TestEntry struct {
	XMLName xml.Name `xml:"entry"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value"`
}

// Implement Entry interface
func (e *TestEntry) EntryName() string {
	return e.Name
}

// Test normalizer - matches production normalizer pattern with struct tags
type TestNormalizer struct {
	Entries []*TestEntry `xml:"entry"`
}

func (n *TestNormalizer) Normalize() ([]*TestEntry, error) {
	return n.Entries, nil
}

// Test specifier
func testSpecifier(entry *TestEntry) (any, error) {
	return entry, nil
}

var _ = Describe("ResourceCache", func() {
	var (
		normalizer *TestNormalizer
		specifier  func(*TestEntry) (any, error)
		ctx        context.Context
	)

	BeforeEach(func() {
		normalizer = &TestNormalizer{}
		specifier = testSpecifier
		ctx = context.Background()
	})

	Context("when cache is enabled", func() {
		var resourceCache manager.CacheManager[*TestEntry]

		BeforeEach(func() {
			resourceCache = manager.NewEnabledCacheManager[*TestEntry](normalizer, specifier)
		})

		It("should store and retrieve entries", func() {
			location := "/config/devices/entry[@name='localhost.localdomain']/vsys/entry[@name='vsys1']/address"
			entries := []*TestEntry{
				{Name: "addr1", Value: "192.168.1.1"},
				{Name: "addr2", Value: "192.168.1.2"},
			}

			err := resourceCache.SetInitialized(ctx, location, entries)
			Expect(err).ToNot(HaveOccurred())

			// Retrieve single entry
			entry, found, err := resourceCache.Get(ctx, location, "addr1")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(entry.Name).To(Equal("addr1"))
			Expect(entry.Value).To(Equal("192.168.1.1"))

			// Retrieve all entries
			allEntries, err := resourceCache.GetAll(ctx, location)
			Expect(err).ToNot(HaveOccurred())
			Expect(allEntries).To(HaveLen(2))
		})

		It("should track location initialization", func() {
			location := "/config/devices/entry[@name='localhost.localdomain']/vsys/entry[@name='vsys1']/address"

			// Not initialized yet
			Expect(resourceCache.IsInitialized(location)).To(BeFalse())

			// Populate location
			entries := []*TestEntry{
				{Name: "addr1", Value: "192.168.1.1"},
			}
			err := resourceCache.SetInitialized(ctx, location, entries)
			Expect(err).ToNot(HaveOccurred())

			// Now initialized
			Expect(resourceCache.IsInitialized(location)).To(BeTrue())
		})

		It("should preserve device order", func() {
			location := "/config/devices/entry[@name='localhost.localdomain']/vsys/entry[@name='vsys1']/address"
			entries := []*TestEntry{
				{Name: "addr3", Value: "3"},
				{Name: "addr1", Value: "1"},
				{Name: "addr2", Value: "2"},
			}

			err := resourceCache.SetInitialized(ctx, location, entries)
			Expect(err).ToNot(HaveOccurred())

			retrieved, err := resourceCache.GetAll(ctx, location)
			Expect(err).ToNot(HaveOccurred())
			Expect(retrieved).To(HaveLen(3))

			// Should be in original order (addr3, addr1, addr2)
			Expect(retrieved[0].Name).To(Equal("addr3"))
			Expect(retrieved[1].Name).To(Equal("addr1"))
			Expect(retrieved[2].Name).To(Equal("addr2"))
		})

		It("should return deep copies", func() {
			location := "/config/devices/entry[@name='localhost.localdomain']/vsys/entry[@name='vsys1']/address"
			entries := []*TestEntry{
				{Name: "addr1", Value: "original"},
			}

			err := resourceCache.SetInitialized(ctx, location, entries)
			Expect(err).ToNot(HaveOccurred())

			// Get entry and modify it
			entry1, found, err := resourceCache.Get(ctx, location, "addr1")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())

			entry1.Value = "modified"

			// Get entry again - should still have original value
			entry2, found, err := resourceCache.Get(ctx, location, "addr1")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(entry2.Value).To(Equal("original"))
		})

		It("should update entries", func() {
			location := "/config/devices/entry[@name='localhost.localdomain']/vsys/entry[@name='vsys1']/address"
			entries := []*TestEntry{
				{Name: "addr1", Value: "original"},
			}

			err := resourceCache.SetInitialized(ctx, location, entries)
			Expect(err).ToNot(HaveOccurred())

			// Update entry
			updated := &TestEntry{Name: "addr1", Value: "updated"}
			err = resourceCache.Put(ctx, location, "addr1", updated)
			Expect(err).ToNot(HaveOccurred())

			// Retrieve should show updated value
			entry, found, err := resourceCache.Get(ctx, location, "addr1")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
			Expect(entry.Value).To(Equal("updated"))
		})

		It("should delete entries", func() {
			location := "/config/devices/entry[@name='localhost.localdomain']/vsys/entry[@name='vsys1']/address"
			entries := []*TestEntry{
				{Name: "addr1", Value: "value1"},
				{Name: "addr2", Value: "value2"},
			}

			err := resourceCache.SetInitialized(ctx, location, entries)
			Expect(err).ToNot(HaveOccurred())

			// Delete entry
			err = resourceCache.Delete(ctx, location, "addr1")
			Expect(err).ToNot(HaveOccurred())

			// Should not be found
			_, found, err := resourceCache.Get(ctx, location, "addr1")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeFalse())

			// Other entry should still exist
			_, found, err = resourceCache.Get(ctx, location, "addr2")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())
		})

		It("should invalidate locations", func() {
			location := "/config/devices/entry[@name='localhost.localdomain']/vsys/entry[@name='vsys1']/address"
			entries := []*TestEntry{
				{Name: "addr1", Value: "value1"},
			}

			err := resourceCache.SetInitialized(ctx, location, entries)
			Expect(err).ToNot(HaveOccurred())
			Expect(resourceCache.IsInitialized(location)).To(BeTrue())

			// Invalidate
			err = resourceCache.Invalidate(ctx, location)
			Expect(err).ToNot(HaveOccurred())

			// Should no longer be initialized
			Expect(resourceCache.IsInitialized(location)).To(BeFalse())
		})

		It("should clear all locations", func() {
			location1 := "/config/devices/entry[@name='localhost.localdomain']/vsys/entry[@name='vsys1']/address"
			location2 := "/config/devices/entry[@name='localhost.localdomain']/vsys/entry[@name='vsys2']/address"

			entries := []*TestEntry{{Name: "addr1", Value: "value1"}}

			err := resourceCache.SetInitialized(ctx, location1, entries)
			Expect(err).ToNot(HaveOccurred())
			err = resourceCache.SetInitialized(ctx, location2, entries)
			Expect(err).ToNot(HaveOccurred())

			resourceCache.Clear()

			Expect(resourceCache.IsInitialized(location1)).To(BeFalse())
			Expect(resourceCache.IsInitialized(location2)).To(BeFalse())
		})
	})

	Context("when cache is disabled", func() {
		var resourceCache manager.CacheManager[*TestEntry]

		BeforeEach(func() {
			resourceCache = manager.NewNoOpCacheManager[*TestEntry]()
		})

		It("should return nil/false for all operations", func() {
			location := "/config/devices/entry[@name='localhost.localdomain']/vsys/entry[@name='vsys1']/address"
			entries := []*TestEntry{{Name: "addr1", Value: "value1"}}

			// SetInitialized should not error but also not store
			err := resourceCache.SetInitialized(ctx, location, entries)
			Expect(err).ToNot(HaveOccurred())

			// IsInitialized should return false
			Expect(resourceCache.IsInitialized(location)).To(BeFalse())

			// Get should return not found
			_, found, err := resourceCache.Get(ctx, location, "addr1")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeFalse())

			// GetAll should return nil
			allEntries, err := resourceCache.GetAll(ctx, location)
			Expect(err).ToNot(HaveOccurred())
			Expect(allEntries).To(BeNil())
		})
	})

})
