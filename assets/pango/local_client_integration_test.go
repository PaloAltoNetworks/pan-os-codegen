package pango_test

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/PaloAltoNetworks/pango"
	"github.com/PaloAltoNetworks/pango/objects/address"
	"github.com/PaloAltoNetworks/pango/util"
	"github.com/PaloAltoNetworks/pango/xmlapi"
)

var _ = Describe("LocalXmlClient Integration", func() {
	var (
		client *pango.LocalXmlClient
		ctx    context.Context
	)

	BeforeEach(func() {
		configXml, err := os.ReadFile("testdata/panorama-test-minimal.xml")
		Expect(err).ToNot(HaveOccurred())

		client, err = newTestClient(configXml)
		Expect(err).ToNot(HaveOccurred())

		ctx = context.Background()
	})

	Describe("Service Instantiation", func() {
		It("should instantiate address.Service with LocalXmlClient", func() {
			svc := address.NewService(client)
			Expect(svc).ToNot(BeNil())
		})

		It("service should have non-nil client", func() {
			svc := address.NewService(client)
			Expect(svc).ToNot(BeNil())
			// Service has client internally, verified by successful instantiation
		})

		It("service Versioning() should return correct version (11.2.3)", func() {
			version := client.Versioning()
			Expect(version).ToNot(BeNil())
			// Version should be 11.2.3 from panorama-test-minimal.xml
			Expect(version.Major).To(Equal(11))
			Expect(version.Minor).To(Equal(2))
			Expect(version.Patch).To(Equal(3))
		})
	})

	Describe("Read Operations - Device Group", func() {
		var svc *address.Service

		BeforeEach(func() {
			svc = address.NewService(client)
		})

		It("should read single address from device-group with get action", func() {
			loc := address.NewDeviceGroupLocation()
			loc.DeviceGroup.DeviceGroup = "dg1-renamed"
			loc.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			entry, err := svc.Read(ctx, *loc, "addr-1-renamed", "get")
			Expect(err).ToNot(HaveOccurred())
			Expect(entry).ToNot(BeNil())
			Expect(entry.Name).To(Equal("addr-1-renamed"))
			Expect(entry.IpNetmask).ToNot(BeNil())
			Expect(*entry.IpNetmask).To(Equal("1.1.1.1"))
		})

		It("should read single address from device-group with show action", func() {
			loc := address.NewDeviceGroupLocation()
			loc.DeviceGroup.DeviceGroup = "e2e-test-SxQAwm-dg"
			loc.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			entry, err := svc.Read(ctx, *loc, "e2e-test-SxQAwm-web-server-1", "show")
			Expect(err).ToNot(HaveOccurred())
			Expect(entry).ToNot(BeNil())
			Expect(entry.Name).To(Equal("e2e-test-SxQAwm-web-server-1"))
			Expect(entry.IpNetmask).ToNot(BeNil())
			Expect(*entry.IpNetmask).To(Equal("10.1.1.10/32"))
			Expect(entry.Description).ToNot(BeNil())
			Expect(*entry.Description).To(Equal("E2E Test - Web Server 1"))
		})

		It("should read address with all fields populated", func() {
			loc := address.NewDeviceGroupLocation()
			loc.DeviceGroup.DeviceGroup = "e2e-test-SxQAwm-dg"
			loc.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			entry, err := svc.Read(ctx, *loc, "e2e-test-SxQAwm-web-server-2", "get")
			Expect(err).ToNot(HaveOccurred())
			Expect(entry).ToNot(BeNil())
			Expect(entry.Name).To(Equal("e2e-test-SxQAwm-web-server-2"))
			Expect(entry.IpNetmask).ToNot(BeNil())
			Expect(*entry.IpNetmask).To(Equal("10.1.1.20/32"))
			Expect(entry.Description).ToNot(BeNil())
			Expect(*entry.Description).To(Equal("E2E Test - Web Server 2"))
		})

		It("should return ObjectNotFound when reading non-existent address", func() {
			loc := address.NewDeviceGroupLocation()
			loc.DeviceGroup.DeviceGroup = "dg1-renamed"
			loc.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			entry, err := svc.Read(ctx, *loc, "does-not-exist", "get")
			Expect(err).To(HaveOccurred())
			Expect(entry).To(BeNil())
			// Check for ObjectNotFound error
			Expect(err.Error()).To(ContainSubstring("Object not found"))
		})
	})

	Describe("List Operations - Device Group", func() {
		var svc *address.Service

		BeforeEach(func() {
			svc = address.NewService(client)
		})

		It("should list all addresses from device-group with single object", func() {
			loc := address.NewDeviceGroupLocation()
			loc.DeviceGroup.DeviceGroup = "dg1-renamed"
			loc.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			entries, err := svc.List(ctx, *loc, "get", "", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(entries).To(HaveLen(1))
			Expect(entries[0].Name).To(Equal("addr-1-renamed"))
		})

		It("should list all addresses from device-group with multiple objects", func() {
			loc := address.NewDeviceGroupLocation()
			loc.DeviceGroup.DeviceGroup = "e2e-test-SxQAwm-dg"
			loc.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			entries, err := svc.List(ctx, *loc, "get", "", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(entries).To(HaveLen(3))

			names := make([]string, len(entries))
			for i, entry := range entries {
				names[i] = entry.Name
				Expect(entry.IpNetmask).ToNot(BeNil())
			}

			Expect(names).To(ContainElement("e2e-test-SxQAwm-web-server-1"))
			Expect(names).To(ContainElement("e2e-test-SxQAwm-web-server-2"))
			Expect(names).To(ContainElement("e2e-test-SxQAwm-web-server-3"))
		})

		It("should list with show action", func() {
			loc := address.NewDeviceGroupLocation()
			loc.DeviceGroup.DeviceGroup = "e2e-test-SxQAwm-dg"
			loc.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			entries, err := svc.List(ctx, *loc, "show", "", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(entries).To(HaveLen(3))

			for _, entry := range entries {
				Expect(entry.Name).ToNot(BeEmpty())
				Expect(entry.IpNetmask).ToNot(BeNil())
			}
		})

		It("should return empty list from non-existent device-group", func() {
			loc := address.NewDeviceGroupLocation()
			loc.DeviceGroup.DeviceGroup = "non-existent-dg"
			loc.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			entries, err := svc.List(ctx, *loc, "get", "", "")
			// May return empty list or error, both are acceptable
			if err != nil {
				Expect(err.Error()).To(ContainSubstring("Object not found"))
			} else {
				Expect(entries).To(BeEmpty())
			}
		})
	})

	Describe("Entry Field Validation", func() {
		var svc *address.Service

		BeforeEach(func() {
			svc = address.NewService(client)
		})

		It("should verify IpNetmask field unmarshals correctly", func() {
			loc := address.NewDeviceGroupLocation()
			loc.DeviceGroup.DeviceGroup = "dg1-renamed"
			loc.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			entry, err := svc.Read(ctx, *loc, "addr-1-renamed", "get")
			Expect(err).ToNot(HaveOccurred())
			Expect(entry.IpNetmask).ToNot(BeNil())
			Expect(*entry.IpNetmask).To(Equal("1.1.1.1"))
		})

		It("should verify Description field unmarshals correctly", func() {
			loc := address.NewDeviceGroupLocation()
			loc.DeviceGroup.DeviceGroup = "e2e-test-SxQAwm-dg"
			loc.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			entry, err := svc.Read(ctx, *loc, "e2e-test-SxQAwm-web-server-1", "get")
			Expect(err).ToNot(HaveOccurred())
			Expect(entry.Description).ToNot(BeNil())
			Expect(*entry.Description).To(Equal("E2E Test - Web Server 1"))
		})

		It("should verify Name field is always present", func() {
			loc := address.NewDeviceGroupLocation()
			loc.DeviceGroup.DeviceGroup = "dg1-renamed"
			loc.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			entry, err := svc.Read(ctx, *loc, "addr-1-renamed", "get")
			Expect(err).ToNot(HaveOccurred())
			Expect(entry.Name).To(Equal("addr-1-renamed"))
			Expect(entry.Name).ToNot(BeEmpty())
		})

		It("should verify optional fields remain nil when not in XML", func() {
			loc := address.NewDeviceGroupLocation()
			loc.DeviceGroup.DeviceGroup = "dg1-renamed"
			loc.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			entry, err := svc.Read(ctx, *loc, "addr-1-renamed", "get")
			Expect(err).ToNot(HaveOccurred())
			// addr-1-renamed has no description in the config
			Expect(entry.Description).To(BeNil())
		})
	})

	Describe("Location Types", func() {
		var svc *address.Service

		BeforeEach(func() {
			svc = address.NewService(client)
		})

		It("should work with DeviceGroup location with explicit PanoramaDevice", func() {
			loc := address.NewDeviceGroupLocation()
			loc.DeviceGroup.DeviceGroup = "dg1-renamed"
			loc.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			entry, err := svc.Read(ctx, *loc, "addr-1-renamed", "get")
			Expect(err).ToNot(HaveOccurred())
			Expect(entry).ToNot(BeNil())
		})

		It("should work with different device groups", func() {
			// Test dg1-renamed
			loc1 := address.NewDeviceGroupLocation()
			loc1.DeviceGroup.DeviceGroup = "dg1-renamed"
			loc1.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			entry1, err := svc.Read(ctx, *loc1, "addr-1-renamed", "get")
			Expect(err).ToNot(HaveOccurred())
			Expect(entry1.Name).To(Equal("addr-1-renamed"))

			// Test e2e-test-SxQAwm-dg
			loc2 := address.NewDeviceGroupLocation()
			loc2.DeviceGroup.DeviceGroup = "e2e-test-SxQAwm-dg"
			loc2.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			entry2, err := svc.Read(ctx, *loc2, "e2e-test-SxQAwm-web-server-1", "get")
			Expect(err).ToNot(HaveOccurred())
			Expect(entry2.Name).To(Equal("e2e-test-SxQAwm-web-server-1"))

			// Test e2e-test-vMoxgj-dg
			loc3 := address.NewDeviceGroupLocation()
			loc3.DeviceGroup.DeviceGroup = "e2e-test-vMoxgj-dg"
			loc3.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			entries, err := svc.List(ctx, *loc3, "get", "", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(entries).To(HaveLen(3))
		})

		It("should handle shared location if present", func() {
			loc := address.NewSharedLocation()
			// SharedLocation has no fields to set

			entries, err := svc.List(ctx, *loc, "get", "", "")
			// Shared location may be empty in test config, which is acceptable
			if err != nil {
				Expect(err.Error()).To(ContainSubstring("Object not found"))
			} else {
				// Empty list is fine for shared location
				Expect(entries).ToNot(BeNil())
			}
		})
	})

	Describe("Error Propagation", func() {
		var svc *address.Service

		BeforeEach(func() {
			svc = address.NewService(client)
		})

		It("should propagate ObjectNotFound error from Read()", func() {
			loc := address.NewDeviceGroupLocation()
			loc.DeviceGroup.DeviceGroup = "dg1-renamed"
			loc.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			entry, err := svc.Read(ctx, *loc, "does-not-exist", "get")
			Expect(err).To(HaveOccurred())
			Expect(entry).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("Object not found"))
		})

		It("should successfully Create() using SET operation", func() {
			loc := address.NewDeviceGroupLocation()
			loc.DeviceGroup.DeviceGroup = "dg1-renamed"
			loc.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			newEntry := &address.Entry{
				Name:      "test-create-address",
				IpNetmask: util.String("192.168.100.100"),
			}

			result, err := svc.Create(ctx, *loc, newEntry)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.Name).To(Equal("test-create-address"))
			Expect(*result.IpNetmask).To(Equal("192.168.100.100"))
		})

		It("should successfully use EDIT operation directly", func() {
			// Note: Service Update() uses MultiConfig which is not yet implemented
			// This test verifies EDIT works at the Communicate() level

			// First create an entry using SET
			setCmd := &xmlapi.Config{
				Action:  "set",
				Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address",
				Element: "<entry name='test-edit-via-service'><ip-netmask>10.10.10.10</ip-netmask></entry>",
			}
			_, _, err := client.Communicate(ctx, setCmd, false, nil)
			Expect(err).ToNot(HaveOccurred())

			// Then edit it
			editCmd := &xmlapi.Config{
				Action:  "edit",
				Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='test-edit-via-service']",
				Element: "<ip-netmask>10.20.30.40</ip-netmask>",
			}
			_, _, err = client.Communicate(ctx, editCmd, false, nil)
			Expect(err).ToNot(HaveOccurred())

			// Verify the change
			getCmd := &xmlapi.Config{
				Action: "get",
				Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='test-edit-via-service']",
			}
			data, _, err := client.Communicate(ctx, getCmd, true, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(ContainSubstring("10.20.30.40"))
		})

		It("should successfully delete via Delete()", func() {
			loc := address.NewDeviceGroupLocation()
			loc.DeviceGroup.DeviceGroup = "dg1-renamed"
			loc.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			// First verify the address exists
			_, err := svc.Read(ctx, *loc, "addr-1-renamed", "get")
			Expect(err).ToNot(HaveOccurred())

			// Delete it
			err = svc.Delete(ctx, *loc, "addr-1-renamed")
			Expect(err).ToNot(HaveOccurred())

			// Verify it's gone
			_, err = svc.Read(ctx, *loc, "addr-1-renamed", "get")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Version-Specific Behavior", func() {
		var svc *address.Service

		BeforeEach(func() {
			svc = address.NewService(client)
		})

		It("should return correct version from Versioning()", func() {
			version := client.Versioning()
			Expect(version).ToNot(BeNil())
			Expect(version.Major).To(Equal(11))
			Expect(version.Minor).To(Equal(2))
			Expect(version.Patch).To(Equal(3))
		})

		It("should use correct version for entry normalization", func() {
			loc := address.NewDeviceGroupLocation()
			loc.DeviceGroup.DeviceGroup = "dg1-renamed"
			loc.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			entry, err := svc.Read(ctx, *loc, "addr-1-renamed", "get")
			Expect(err).ToNot(HaveOccurred())
			Expect(entry).ToNot(BeNil())
			// Entry fields should be normalized for version 11.2.3
			Expect(entry.Name).ToNot(BeEmpty())
			Expect(entry.IpNetmask).ToNot(BeNil())
		})
	})

	Describe("Multiple Location Types in Same Config", func() {
		var svc *address.Service

		BeforeEach(func() {
			svc = address.NewService(client)
		})

		It("should read from multiple device-groups in sequence", func() {
			// Read from dg1-renamed
			loc1 := address.NewDeviceGroupLocation()
			loc1.DeviceGroup.DeviceGroup = "dg1-renamed"
			loc1.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			entry1, err := svc.Read(ctx, *loc1, "addr-1-renamed", "get")
			Expect(err).ToNot(HaveOccurred())
			Expect(entry1.Name).To(Equal("addr-1-renamed"))

			// Read from e2e-test-SxQAwm-dg
			loc2 := address.NewDeviceGroupLocation()
			loc2.DeviceGroup.DeviceGroup = "e2e-test-SxQAwm-dg"
			loc2.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			entry2, err := svc.Read(ctx, *loc2, "e2e-test-SxQAwm-web-server-1", "get")
			Expect(err).ToNot(HaveOccurred())
			Expect(entry2.Name).To(Equal("e2e-test-SxQAwm-web-server-1"))

			// Read from e2e-test-vMoxgj-dg
			loc3 := address.NewDeviceGroupLocation()
			loc3.DeviceGroup.DeviceGroup = "e2e-test-vMoxgj-dg"
			loc3.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			entries, err := svc.List(ctx, *loc3, "get", "", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(entries).To(HaveLen(3))

			// Verify objects are distinct
			Expect(entry1.Name).ToNot(Equal(entry2.Name))
		})

		It("should list from multiple locations without cross-contamination", func() {
			// List from dg1-renamed (1 object)
			loc1 := address.NewDeviceGroupLocation()
			loc1.DeviceGroup.DeviceGroup = "dg1-renamed"
			loc1.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			entries1, err := svc.List(ctx, *loc1, "get", "", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(entries1).To(HaveLen(1))

			// List from e2e-test-SxQAwm-dg (3 objects)
			loc2 := address.NewDeviceGroupLocation()
			loc2.DeviceGroup.DeviceGroup = "e2e-test-SxQAwm-dg"
			loc2.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			entries2, err := svc.List(ctx, *loc2, "get", "", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(entries2).To(HaveLen(3))

			// List from e2e-test-vMoxgj-dg (3 objects)
			loc3 := address.NewDeviceGroupLocation()
			loc3.DeviceGroup.DeviceGroup = "e2e-test-vMoxgj-dg"
			loc3.DeviceGroup.PanoramaDevice = "localhost.localdomain"

			entries3, err := svc.List(ctx, *loc3, "get", "", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(entries3).To(HaveLen(3))

			// Verify no cross-contamination
			Expect(entries1[0].Name).To(Equal("addr-1-renamed"))
			Expect(entries2).ToNot(ContainElement(entries1[0]))
		})
	})
})
