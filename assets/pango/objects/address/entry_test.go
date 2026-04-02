package address

import (
	"encoding/xml"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/PaloAltoNetworks/pango/util"
)

var _ = ginkgo.Describe("SelectiveOperations", func() {
	ginkgo.Describe("Selective Marshalling", func() {
		var entry *Entry

		ginkgo.BeforeEach(func() {
			entry = &Entry{
				Name:        "test-address",
				Description: util.String("Test description"),
				IpNetmask:   util.String("10.1.1.0/24"),
				Tag:         []string{"tag1", "tag2"},
				Fqdn:        util.String("example.com"),
			}
		})

		ginkgo.Context("Full marshalling", func() {
			ginkgo.It("should include entry wrapper", func() {
				var fullXml entryXml
				fullXml.MarshalFromObject(*entry)

				xmlBytes, err := xml.Marshal(fullXml)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				xmlStr := string(xmlBytes)
				ginkgo.GinkgoWriter.Printf("Full XML: %s\n", xmlStr)

				// Should have entry wrapper
				gomega.Expect(xmlStr).To(gomega.ContainSubstring(`<entry name="test-address">`))
				gomega.Expect(xmlStr).To(gomega.ContainSubstring(`</entry>`))
				gomega.Expect(xmlStr).To(gomega.ContainSubstring(`<description>Test description</description>`))
			})
		})

		ginkgo.Context("Partial marshalling", func() {
			ginkgo.It("should exclude entry wrapper", func() {
				fieldOpt := &FieldOption{
					fields: map[string]bool{
						"description": true,
						"ip-netmask":  true,
					},
					mode: FieldModeInclude,
				}

				var partialXml entryXmlPartial
				partialXml.MarshalFromObjectWithFields(*entry, fieldOpt)

				xmlBytes, err := xml.Marshal(partialXml)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				xmlStr := string(xmlBytes)
				ginkgo.GinkgoWriter.Printf("Partial XML: %s\n", xmlStr)

				// CRITICAL: Should NOT have entry wrapper
				gomega.Expect(xmlStr).ToNot(gomega.ContainSubstring(`<entry`))
				gomega.Expect(xmlStr).ToNot(gomega.ContainSubstring(`name=`))

				// Should only contain selected fields
				gomega.Expect(xmlStr).To(gomega.ContainSubstring(`<description>Test description</description>`))
				gomega.Expect(xmlStr).To(gomega.ContainSubstring(`<ip-netmask>10.1.1.0/24</ip-netmask>`))

				// Should NOT contain non-selected fields
				gomega.Expect(xmlStr).ToNot(gomega.ContainSubstring(`<tag>`))
				gomega.Expect(xmlStr).ToNot(gomega.ContainSubstring(`<fqdn>`))
			})

			ginkgo.It("should handle single field selection", func() {
				fieldOpt := &FieldOption{
					fields: map[string]bool{"description": true},
					mode:   FieldModeInclude,
				}

				var partialXml entryXmlPartial
				partialXml.MarshalFromObjectWithFields(*entry, fieldOpt)

				xmlBytes, err := xml.Marshal(partialXml)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				xmlStr := string(xmlBytes)
				ginkgo.GinkgoWriter.Printf("Partial XML (description only): %s\n", xmlStr)

				gomega.Expect(xmlStr).ToNot(gomega.ContainSubstring(`<entry`))
				gomega.Expect(xmlStr).To(gomega.ContainSubstring(`<description>`))
				gomega.Expect(xmlStr).ToNot(gomega.ContainSubstring(`<ip-netmask>`))
			})
		})
	})

	ginkgo.Describe("Field Helpers", func() {
		var entry *Entry

		ginkgo.BeforeEach(func() {
			entry = &Entry{
				Name:        "test",
				Description: util.String("Test"),
				IpNetmask:   util.String("10.1.1.0/24"),
				Tag:         []string{"tag1"},
			}
		})

		ginkgo.Describe("GetFieldValue", func() {
			ginkgo.It("should return correct value for name field", func() {
				val := entry.GetFieldValue("name")
				gomega.Expect(val).To(gomega.Equal("test"))
			})

			ginkgo.It("should return correct value for description field", func() {
				val := entry.GetFieldValue("description")
				gomega.Expect(val).To(gomega.Equal(entry.Description))
			})

			ginkgo.It("should return nil for unknown field", func() {
				val := entry.GetFieldValue("unknown")
				gomega.Expect(val).To(gomega.BeNil())
			})
		})


		ginkgo.Describe("IsFieldNilOrEmpty", func() {
			ginkgo.It("should detect nil pointer fields", func() {
				isEmpty := entry.IsFieldNilOrEmpty("fqdn")
				gomega.Expect(isEmpty).To(gomega.BeTrue())
			})

			ginkgo.It("should detect non-nil pointer fields", func() {
				isEmpty := entry.IsFieldNilOrEmpty("description")
				gomega.Expect(isEmpty).To(gomega.BeFalse())
			})

			ginkgo.It("should detect empty slices", func() {
				emptyEntry := &Entry{
					Name: "test",
					Tag:  []string{},
				}
				isEmpty := emptyEntry.IsFieldNilOrEmpty("tag")
				gomega.Expect(isEmpty).To(gomega.BeTrue())
			})

			ginkgo.It("should detect non-empty slices", func() {
				isEmpty := entry.IsFieldNilOrEmpty("tag")
				gomega.Expect(isEmpty).To(gomega.BeFalse())
			})
		})
	})

	ginkgo.Describe("Partial Specifier", func() {
		ginkgo.It("should return correct type and marshal only selected fields", func() {
			entry := &Entry{
				Name:        "test",
				Description: util.String("Test"),
				IpNetmask:   util.String("10.1.1.0/24"),
				Tag:         []string{"tag1"},
			}

			fieldOpt := &FieldOption{
				fields: map[string]bool{"description": true},
				mode:   FieldModeInclude,
			}

			spec, err := specifyEntryPartial(entry, fieldOpt)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			partialXml, ok := spec.(entryXmlPartial)
			gomega.Expect(ok).To(gomega.BeTrue(), "should return entryXmlPartial type")

			// Verify only selected field is marshalled
			gomega.Expect(partialXml.Description).To(gomega.Equal(entry.Description))
			gomega.Expect(partialXml.IpNetmask).To(gomega.BeNil())
			gomega.Expect(partialXml.Tag).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("Selective Unmarshalling", func() {
		ginkgo.Context("UnmarshalToObjectWithFields with field selection", func() {
			ginkgo.It("should unmarshal only selected fields", func() {
				// Create XML struct with all fields populated
				xmlEntry := entryXml{
					Name: "test-address",
					entryXmlFields: entryXmlFields{
						Description: util.String("Test description"),
						IpNetmask:   util.String("10.1.1.0/24"),
						Tag:         util.StrToMem([]string{"tag1", "tag2"}),
						Fqdn:        util.String("example.com"),
					},
				}

				// Unmarshal with only description and tag selected
				fieldOpt := &FieldOption{
					fields: map[string]bool{
						"description": true,
						"tag":         true,
					},
					mode: FieldModeInclude,
				}

				entry, err := xmlEntry.UnmarshalToObjectWithFields(fieldOpt)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				// Name should always be present
				gomega.Expect(entry.Name).To(gomega.Equal("test-address"))

				// Selected fields should be populated
				gomega.Expect(entry.Description).ToNot(gomega.BeNil())
				gomega.Expect(*entry.Description).To(gomega.Equal("Test description"))
				gomega.Expect(entry.Tag).To(gomega.Equal([]string{"tag1", "tag2"}))

				// Unselected fields should be nil/empty
				gomega.Expect(entry.IpNetmask).To(gomega.BeNil())
				gomega.Expect(entry.Fqdn).To(gomega.BeNil())
			})

			ginkgo.It("should unmarshal all fields when fieldOpt is nil", func() {
				xmlEntry := entryXml{
					Name: "test-address",
					entryXmlFields: entryXmlFields{
						Description: util.String("Test description"),
						IpNetmask:   util.String("10.1.1.0/24"),
						Tag:         util.StrToMem([]string{"tag1"}),
					},
				}

				entry, err := xmlEntry.UnmarshalToObjectWithFields(nil)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				// All fields should be populated
				gomega.Expect(entry.Name).To(gomega.Equal("test-address"))
				gomega.Expect(*entry.Description).To(gomega.Equal("Test description"))
				gomega.Expect(*entry.IpNetmask).To(gomega.Equal("10.1.1.0/24"))
				gomega.Expect(entry.Tag).To(gomega.Equal([]string{"tag1"}))
			})
		})

	})

})
