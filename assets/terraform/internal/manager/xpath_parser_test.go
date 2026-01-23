package manager_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("XPath Parser", func() {
	Describe("extractNamesFromPredicate", func() {
		It("should extract single name", func() {
			names := extractNamesFromPredicate("@name='admin'")
			Expect(names).To(Equal([]string{"admin"}))
		})

		It("should extract multiple names with OR", func() {
			names := extractNamesFromPredicate("@name='admin' or @name='user1'")
			Expect(names).To(Equal([]string{"admin", "user1"}))
		})

		It("should extract three names with OR", func() {
			names := extractNamesFromPredicate("@name='admin' or @name='user1' or @name='backup'")
			Expect(names).To(Equal([]string{"admin", "user1", "backup"}))
		})

		It("should handle extra whitespace", func() {
			names := extractNamesFromPredicate("@name='admin'  or  @name='user1'")
			Expect(names).To(Equal([]string{"admin", "user1"}))
		})

		It("should return empty for invalid predicate", func() {
			names := extractNamesFromPredicate("invalid")
			Expect(names).To(BeEmpty())
		})

		It("should return empty for empty predicate", func() {
			names := extractNamesFromPredicate("")
			Expect(names).To(BeEmpty())
		})
	})

	Describe("extractPrefixesFromPredicate", func() {
		It("should extract single prefix", func() {
			prefixes := extractPrefixesFromPredicate("starts-with(@name, 'a')")
			Expect(prefixes).To(Equal([]string{"a"}))
		})

		It("should extract single prefix without space after comma", func() {
			prefixes := extractPrefixesFromPredicate("starts-with(@name,'a')")
			Expect(prefixes).To(Equal([]string{"a"}))
		})

		It("should extract multiple prefixes with OR", func() {
			prefixes := extractPrefixesFromPredicate("starts-with(@name, 'a') or starts-with(@name, 'A')")
			Expect(prefixes).To(Equal([]string{"a", "A"}))
		})

		It("should extract three prefixes with OR", func() {
			prefixes := extractPrefixesFromPredicate("starts-with(@name, 'a') or starts-with(@name, 'b') or starts-with(@name, 'c')")
			Expect(prefixes).To(Equal([]string{"a", "b", "c"}))
		})

		It("should handle mixed spacing", func() {
			prefixes := extractPrefixesFromPredicate("starts-with(@name,'a') or starts-with(@name, 'b')")
			Expect(prefixes).To(Equal([]string{"a", "b"}))
		})

		It("should return empty for invalid predicate", func() {
			prefixes := extractPrefixesFromPredicate("invalid")
			Expect(prefixes).To(BeEmpty())
		})

		It("should return empty for empty predicate", func() {
			prefixes := extractPrefixesFromPredicate("")
			Expect(prefixes).To(BeEmpty())
		})
	})

	Describe("filterByNames", func() {
		var entries []*MockEntryObject

		BeforeEach(func() {
			entries = []*MockEntryObject{
				{Name: "admin", Value: "value1"},
				{Name: "user1", Value: "value2"},
				{Name: "backup", Value: "value3"},
			}
		})

		It("should filter entries by single name", func() {
			filtered := filterByNames(entries, []string{"admin"})
			Expect(filtered).To(HaveLen(1))
			Expect(filtered[0].EntryName()).To(Equal("admin"))
		})

		It("should filter entries by multiple names", func() {
			filtered := filterByNames(entries, []string{"admin", "backup"})
			Expect(filtered).To(HaveLen(2))
			Expect(filtered[0].EntryName()).To(Equal("admin"))
			Expect(filtered[1].EntryName()).To(Equal("backup"))
		})

		It("should return empty for non-matching names", func() {
			filtered := filterByNames(entries, []string{"nonexistent"})
			Expect(filtered).To(BeEmpty())
		})

		It("should return empty for empty name list", func() {
			filtered := filterByNames(entries, []string{})
			Expect(filtered).To(BeEmpty())
		})

		It("should handle duplicate names in filter", func() {
			filtered := filterByNames(entries, []string{"admin", "admin"})
			Expect(filtered).To(HaveLen(1))
			Expect(filtered[0].EntryName()).To(Equal("admin"))
		})
	})

	Describe("filterByPrefixes", func() {
		var entries []*MockEntryObject

		BeforeEach(func() {
			entries = []*MockEntryObject{
				{Name: "admin", Value: "value1"},
				{Name: "backup", Value: "value2"},
				{Name: "user1", Value: "value3"},
				{Name: "1-config", Value: "value4"},
				{Name: "_special", Value: "value5"},
			}
		})

		It("should filter entries by single prefix", func() {
			filtered := filterByPrefixes(entries, []string{"a"})
			Expect(filtered).To(HaveLen(1))
			Expect(filtered[0].EntryName()).To(Equal("admin"))
		})

		It("should filter entries by multiple prefixes", func() {
			filtered := filterByPrefixes(entries, []string{"a", "b"})
			Expect(filtered).To(HaveLen(2))
			Expect(filtered[0].EntryName()).To(Equal("admin"))
			Expect(filtered[1].EntryName()).To(Equal("backup"))
		})

		It("should filter entries with numeric prefix", func() {
			filtered := filterByPrefixes(entries, []string{"1"})
			Expect(filtered).To(HaveLen(1))
			Expect(filtered[0].EntryName()).To(Equal("1-config"))
		})

		It("should filter entries with special character prefix", func() {
			filtered := filterByPrefixes(entries, []string{"_"})
			Expect(filtered).To(HaveLen(1))
			Expect(filtered[0].EntryName()).To(Equal("_special"))
		})

		It("should be case-sensitive", func() {
			filtered := filterByPrefixes(entries, []string{"A"})
			Expect(filtered).To(BeEmpty())
		})

		It("should return empty for non-matching prefixes", func() {
			filtered := filterByPrefixes(entries, []string{"z"})
			Expect(filtered).To(BeEmpty())
		})

		It("should return empty for empty prefix list", func() {
			filtered := filterByPrefixes(entries, []string{})
			Expect(filtered).To(BeEmpty())
		})
	})

	Describe("parseXpathPredicate", func() {
		var entries []*MockEntryObject

		BeforeEach(func() {
			entries = []*MockEntryObject{
				{Name: "admin", Value: "value1"},
				{Name: "user1", Value: "value2"},
				{Name: "backup", Value: "value3"},
			}
		})

		Context("with no predicate", func() {
			It("should return all entries for xpath without brackets", func() {
				filtered := parseXpathPredicate("/some/location/entry", entries)
				Expect(filtered).To(HaveLen(3))
			})

			It("should return all entries for xpath ending with @name", func() {
				filtered := parseXpathPredicate("/some/location/entry/@name", entries)
				Expect(filtered).To(HaveLen(3))
			})
		})

		Context("with name-based OR predicates", func() {
			It("should filter by single name", func() {
				filtered := parseXpathPredicate("/some/location/entry[@name='admin']", entries)
				Expect(filtered).To(HaveLen(1))
				Expect(filtered[0].EntryName()).To(Equal("admin"))
			})

			It("should filter by multiple names with OR", func() {
				filtered := parseXpathPredicate("/some/location/entry[@name='admin' or @name='backup']", entries)
				Expect(filtered).To(HaveLen(2))
				Expect(filtered[0].EntryName()).To(Equal("admin"))
				Expect(filtered[1].EntryName()).To(Equal("backup"))
			})

			It("should handle xpath with trailing @name", func() {
				filtered := parseXpathPredicate("/some/location/entry[@name='admin' or @name='user1']/@name", entries)
				Expect(filtered).To(HaveLen(2))
			})
		})

		Context("with starts-with predicates", func() {
			It("should filter by single prefix", func() {
				filtered := parseXpathPredicate("/some/location/entry[starts-with(@name, 'a')]", entries)
				Expect(filtered).To(HaveLen(1))
				Expect(filtered[0].EntryName()).To(Equal("admin"))
			})

			It("should filter by multiple prefixes with OR", func() {
				filtered := parseXpathPredicate("/some/location/entry[starts-with(@name, 'a') or starts-with(@name, 'b')]", entries)
				Expect(filtered).To(HaveLen(2))
				Expect(filtered[0].EntryName()).To(Equal("admin"))
				Expect(filtered[1].EntryName()).To(Equal("backup"))
			})

			It("should handle xpath with trailing @name", func() {
				filtered := parseXpathPredicate("/some/location/entry[starts-with(@name, 'a') or starts-with(@name, 'u')]/@name", entries)
				Expect(filtered).To(HaveLen(2))
			})
		})

		Context("with malformed predicates", func() {
			It("should return all entries for unbalanced brackets", func() {
				filtered := parseXpathPredicate("/some/location/entry[@name='admin'", entries)
				Expect(filtered).To(HaveLen(3))
			})

			It("should return all entries for unknown predicate format", func() {
				filtered := parseXpathPredicate("/some/location/entry[unknown-predicate]", entries)
				Expect(filtered).To(HaveLen(3))
			})

			It("should handle empty predicate", func() {
				filtered := parseXpathPredicate("/some/location/entry[]", entries)
				Expect(filtered).To(HaveLen(3))
			})
		})

		Context("with edge cases", func() {
			It("should handle empty entries list", func() {
				emptyEntries := []*MockEntryObject{}
				filtered := parseXpathPredicate("/some/location/entry[@name='admin']", emptyEntries)
				Expect(filtered).To(BeEmpty())
			})

			It("should handle whitespace in xpath", func() {
				filtered := parseXpathPredicate("  /some/location/entry[@name='admin']  ", entries)
				Expect(filtered).To(HaveLen(1))
			})
		})
	})
})
