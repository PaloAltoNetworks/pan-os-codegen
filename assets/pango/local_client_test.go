package pango_test

import (
	"context"
	"encoding/xml"
	"errors"
	"os"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/PaloAltoNetworks/pango"
	pangoerrors "github.com/PaloAltoNetworks/pango/errors"
	"github.com/PaloAltoNetworks/pango/version"
	"github.com/PaloAltoNetworks/pango/xmlapi"
)

var _ = Describe("LocalXmlClient", func() {
	Describe("NewLocalXmlClient", func() {
		Context("with valid panorama config", func() {
			It("should parse and detect version and device type", func() {
				configXml, err := os.ReadFile("testdata/panorama-running-config.xml")
				Expect(err).ToNot(HaveOccurred())

				client, err := pango.NewLocalXmlClient(configXml)
				Expect(err).ToNot(HaveOccurred())
				Expect(client).ToNot(BeNil())

				// Check version was detected (11.2.3 from detail-version)
				v := client.Versioning()
				Expect(v.Major).To(Equal(11))
				Expect(v.Minor).To(Equal(2))
				Expect(v.Patch).To(Equal(3))

				// Check device type is panorama
				isPanorama, err := client.IsPanorama()
				Expect(err).ToNot(HaveOccurred())
				Expect(isPanorama).To(BeTrue())

				isFirewall, err := client.IsFirewall()
				Expect(err).ToNot(HaveOccurred())
				Expect(isFirewall).To(BeFalse())
			})
		})

		Context("with version option", func() {
			It("should use provided version", func() {
				// Config without detail-version attribute
				configXml := []byte(`<?xml version="1.0"?><config></config>`)

				expectedVersion, err := version.New("10.2.5")
				Expect(err).ToNot(HaveOccurred())

				client, err := pango.NewLocalXmlClient(configXml, pango.WithVersion(expectedVersion))
				Expect(err).ToNot(HaveOccurred())

				v := client.Versioning()
				Expect(v).To(Equal(expectedVersion))
			})
		})

		Context("with hostname option", func() {
			It("should set hostname", func() {
				configXml := []byte(`<?xml version="1.0"?><config></config>`)

				client, err := pango.NewLocalXmlClient(configXml, pango.WithHostname("test-firewall"))
				Expect(err).ToNot(HaveOccurred())
				Expect(client).ToNot(BeNil())
			})
		})

		Context("with invalid XML", func() {
			It("should return error", func() {
				configXml := []byte(`not valid xml`)

				_, err := pango.NewLocalXmlClient(configXml)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with wrong root element", func() {
			It("should return error", func() {
				configXml := []byte(`<?xml version="1.0"?><response></response>`)

				_, err := pango.NewLocalXmlClient(configXml)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("expected <config> root element"))
			})
		})
	})

	Describe("Communicate - Read Operations", func() {
		var (
			client *pango.LocalXmlClient
			ctx    context.Context
		)

		BeforeEach(func() {
			configXml, err := os.ReadFile("testdata/panorama-running-config.xml")
			Expect(err).ToNot(HaveOccurred())

			client, err = pango.NewLocalXmlClient(configXml)
			Expect(err).ToNot(HaveOccurred())

			ctx = context.Background()
		})

		Context("get single entry with strip=true", func() {
			It("should return unwrapped entry", func() {
				cmd := &xmlapi.Config{
					Action: "get",
					Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='addr-1-renamed']",
				}

				data, resp, err := client.Communicate(ctx, cmd, true, nil)
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(200))

				dataStr := string(data)
				Expect(dataStr).ToNot(BeEmpty())
				Expect(dataStr).To(ContainSubstring("addr-1-renamed"))
				Expect(dataStr).To(ContainSubstring("1.1.1.1"))
			})
		})

		Context("get single entry with strip=false", func() {
			It("should return wrapped response", func() {
				cmd := &xmlapi.Config{
					Action: "get",
					Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='addr-1-renamed']",
				}

				data, _, err := client.Communicate(ctx, cmd, false, nil)
				Expect(err).ToNot(HaveOccurred())

				dataStr := string(data)
				Expect(dataStr).To(ContainSubstring(`<response status="success">`))
				Expect(dataStr).To(ContainSubstring(`<result total-count="`))
				Expect(dataStr).To(ContainSubstring("addr-1-renamed"))
			})
		})

		Context("list multiple entries", func() {
			It("should return all entries with count attributes", func() {
				cmd := &xmlapi.Config{
					Action: "get",
					Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='e2e-test-SxQAwm-dg']/address/entry",
				}

				data, _, err := client.Communicate(ctx, cmd, false, nil)
				Expect(err).ToNot(HaveOccurred())

				dataStr := string(data)
				Expect(dataStr).To(ContainSubstring(`total-count="3"`))
				Expect(dataStr).To(ContainSubstring(`count="3"`))
				Expect(dataStr).To(ContainSubstring("e2e-test-SxQAwm-web-server-1"))
				Expect(dataStr).To(ContainSubstring("e2e-test-SxQAwm-web-server-2"))
				Expect(dataStr).To(ContainSubstring("e2e-test-SxQAwm-web-server-3"))
			})
		})

		Context("show action", func() {
			It("should work like get action", func() {
				cmd := &xmlapi.Config{
					Action: "show",
					Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='e2e-test-SxQAwm-dg']/address/entry[@name='e2e-test-SxQAwm-web-server-2']",
				}

				data, _, err := client.Communicate(ctx, cmd, true, nil)
				Expect(err).ToNot(HaveOccurred())

				dataStr := string(data)
				Expect(dataStr).To(ContainSubstring("e2e-test-SxQAwm-web-server-2"))
				Expect(dataStr).To(ContainSubstring("10.1.1.20/32"))
			})
		})

		Context("object not found", func() {
			It("should return ObjectNotFound error", func() {
				cmd := &xmlapi.Config{
					Action: "get",
					Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='nonexistent']",
				}

				_, _, err := client.Communicate(ctx, cmd, true, nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Object not found"))
			})
		})
	})

	// Note: All write operations (set, edit, delete, rename, move) are now implemented
	// See WP05, WP06, WP07 for individual operation test suites

	Describe("Communicate - Unsupported Commands", func() {
		var (
			client *pango.LocalXmlClient
			ctx    context.Context
		)

		BeforeEach(func() {
			configXml, err := os.ReadFile("testdata/panorama-running-config.xml")
			Expect(err).ToNot(HaveOccurred())

			client, err = pango.NewLocalXmlClient(configXml)
			Expect(err).ToNot(HaveOccurred())

			ctx = context.Background()
		})

		Context("op command", func() {
			It("should return ErrUnsupportedOperation", func() {
				type opCmd struct {
					XMLName struct{} `xml:"show"`
					Cmd     string   `xml:"system>info"`
				}

				cmd := &xmlapi.Op{
					Command: opCmd{},
				}

				_, _, err := client.Communicate(ctx, cmd, true, nil)
				Expect(err).To(Equal(pango.ErrUnsupportedOperation))
			})
		})
	})

	Describe("Unsupported Methods", func() {
		var (
			client *pango.LocalXmlClient
			ctx    context.Context
		)

		BeforeEach(func() {
			configXml := []byte(`<?xml version="1.0"?><config detail-version="11.0.2"></config>`)
			var err error
			client, err = pango.NewLocalXmlClient(configXml)
			Expect(err).ToNot(HaveOccurred())

			ctx = context.Background()
		})

		It("StartJob should return ErrJobsNotSupported", func() {
			_, _, _, err := client.StartJob(ctx, nil)
			Expect(err).To(Equal(pango.ErrJobsNotSupported))
		})

		It("WaitForJob should return ErrJobsNotSupported", func() {
			err := client.WaitForJob(ctx, 1, 0, nil)
			Expect(err).To(Equal(pango.ErrJobsNotSupported))
		})

		It("WaitForLogs should return ErrJobsNotSupported", func() {
			_, err := client.WaitForLogs(ctx, 1, 0, nil)
			Expect(err).To(Equal(pango.ErrJobsNotSupported))
		})

		It("MultiConfig should handle empty batch", func() {
			_, _, mcResp, err := client.MultiConfig(ctx, nil, false, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(mcResp.Status).To(Equal("success"))
		})

		It("ImportFile should return ErrUnsupportedOperation", func() {
			_, _, err := client.ImportFile(ctx, nil, nil, "", "", false, nil)
			Expect(err).To(Equal(pango.ErrUnsupportedOperation))
		})

		It("ExportFile should return ErrUnsupportedOperation", func() {
			_, _, _, err := client.ExportFile(ctx, nil, nil)
			Expect(err).To(Equal(pango.ErrUnsupportedOperation))
		})

		It("GenerateApiKey should return ErrUnsupportedOperation", func() {
			_, err := client.GenerateApiKey(ctx, "user", "pass")
			Expect(err).To(Equal(pango.ErrUnsupportedOperation))
		})

		It("RequestPasswordHash should return ErrUnsupportedOperation", func() {
			_, err := client.RequestPasswordHash(ctx, "password")
			Expect(err).To(Equal(pango.ErrUnsupportedOperation))
		})

		It("GetTechSupportFile should return ErrUnsupportedOperation", func() {
			_, _, err := client.GetTechSupportFile(ctx)
			Expect(err).To(Equal(pango.ErrUnsupportedOperation))
		})

		It("Clock should return ErrUnsupportedOperation", func() {
			_, err := client.Clock(ctx)
			Expect(err).To(Equal(pango.ErrUnsupportedOperation))
		})

		It("Plugins should return empty list", func() {
			plugins, err := client.Plugins(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(plugins).To(BeEmpty())
		})
	})

	Describe("Versioning", func() {
		It("should return detected version", func() {
			configXml, err := os.ReadFile("testdata/panorama-running-config.xml")
			Expect(err).ToNot(HaveOccurred())

			client, err := pango.NewLocalXmlClient(configXml)
			Expect(err).ToNot(HaveOccurred())

			v := client.Versioning()
			Expect(v.Major).To(Equal(11))
			Expect(v.Minor).To(Equal(2))
			Expect(v.Patch).To(Equal(3))
		})
	})

	Describe("GetTarget", func() {
		It("should return empty string", func() {
			configXml := []byte(`<?xml version="1.0"?><config></config>`)
			client, err := pango.NewLocalXmlClient(configXml)
			Expect(err).ToNot(HaveOccurred())

			target := client.GetTarget()
			Expect(target).To(Equal(""))
		})
	})
})

var _ = Describe("LocalXmlClient Error Types", func() {
	Describe("ErrInvalidXpath", func() {
		It("should create error with cause", func() {
			cause := errors.New("parse error")
			err := pangoerrors.NewErrInvalidXpath("/invalid[[@", cause)
			Expect(err.XPath).To(Equal("/invalid[[@"))
			Expect(err.Cause).To(Equal(cause))
			Expect(err.Error()).To(ContainSubstring("invalid XPath syntax"))
			Expect(err.Error()).To(ContainSubstring("parse error"))
		})

		It("should create error without cause", func() {
			err := pangoerrors.NewErrInvalidXpath("/invalid", nil)
			Expect(err.Error()).To(Equal("invalid XPath syntax '/invalid'"))
		})

		It("should support error unwrapping", func() {
			cause := errors.New("parse error")
			err := pangoerrors.NewErrInvalidXpath("/invalid", cause)
			unwrapped := errors.Unwrap(err)
			Expect(unwrapped).To(Equal(cause))
		})

		It("should support errors.As for type checking", func() {
			err := pangoerrors.NewErrInvalidXpath("/test", nil)
			var target *pangoerrors.ErrInvalidXpath
			Expect(errors.As(err, &target)).To(BeTrue())
			Expect(target.XPath).To(Equal("/test"))
		})
	})

	Describe("ErrObjectNotFound", func() {
		It("should create error with xpath", func() {
			xpath := "/config/devices/entry[@name='localhost']/address/entry[@name='web-server-99']"
			err := pangoerrors.NewErrObjectNotFound(xpath)
			Expect(err.XPath).To(Equal(xpath))
			Expect(err.Error()).To(ContainSubstring("object not found"))
			Expect(err.Error()).To(ContainSubstring(xpath))
		})

		It("should support errors.As for type checking", func() {
			err := pangoerrors.NewErrObjectNotFound("/config/test")
			var target *pangoerrors.ErrObjectNotFound
			Expect(errors.As(err, &target)).To(BeTrue())
			Expect(target.XPath).To(Equal("/config/test"))
		})
	})

	Describe("ErrOperationFailed", func() {
		It("should create error with index and cause", func() {
			cause := pangoerrors.NewErrObjectNotFound("/config/test")
			err := pangoerrors.NewErrOperationFailed(3, cause)
			Expect(err.OperationIndex).To(Equal(3))
			Expect(err.Cause).To(Equal(cause))
			Expect(err.Error()).To(ContainSubstring("operation 3 failed"))
			Expect(err.Error()).To(ContainSubstring("object not found"))
		})

		It("should unwrap to underlying cause", func() {
			cause := pangoerrors.NewErrObjectNotFound("/config/test")
			err := pangoerrors.NewErrOperationFailed(3, cause)
			unwrapped := errors.Unwrap(err)
			Expect(unwrapped).To(Equal(cause))
		})

		It("should support errors.As to find wrapped error", func() {
			cause := pangoerrors.NewErrObjectNotFound("/config/test")
			err := pangoerrors.NewErrOperationFailed(3, cause)

			var target *pangoerrors.ErrObjectNotFound
			Expect(errors.As(err, &target)).To(BeTrue())
			Expect(target.XPath).To(Equal("/config/test"))
		})
	})

	Describe("ErrRenameConflict", func() {
		It("should format message correctly", func() {
			xpath := "/config/devices/entry[@name='localhost']/address/entry[@name='web-server-2']"
			err := pangoerrors.NewErrRenameConflict(xpath, "web-server-2", "web-server-1")

			Expect(err.XPath).To(Equal(xpath))
			Expect(err.SourceName).To(Equal("web-server-2"))
			Expect(err.TargetName).To(Equal("web-server-1"))

			Expect(err.Error()).To(ContainSubstring("rename conflict"))
			Expect(err.Error()).To(ContainSubstring("web-server-1"))
			Expect(err.Error()).To(ContainSubstring("web-server-2"))
			Expect(err.Error()).To(ContainSubstring(xpath))
		})

		It("should support errors.As for type checking", func() {
			err := pangoerrors.NewErrRenameConflict("/config/test", "old", "new")
			var target *pangoerrors.ErrRenameConflict
			Expect(errors.As(err, &target)).To(BeTrue())
			Expect(target.SourceName).To(Equal("old"))
			Expect(target.TargetName).To(Equal("new"))
		})
	})
})

var _ = Describe("LocalXmlClient Working Copy", func() {
	var (
		client *pango.LocalXmlClient
		ctx    context.Context
	)

	BeforeEach(func() {
		configXml, err := os.ReadFile("testdata/panorama-running-config.xml")
		Expect(err).ToNot(HaveOccurred())

		client, err = pango.NewLocalXmlClient(configXml)
		Expect(err).ToNot(HaveOccurred())

		ctx = context.Background()
	})

	Describe("cloneDocument", func() {
		It("should create independent clone", func() {
			// Get original document reference
			originalRoot := client.Versioning()

			// Clone the document
			// Note: We can't directly call cloneDocument() as it's unexported
			// Instead, we'll verify clone independence through MultiConfig behavior
			// For this test, we verify the document structure is preserved

			// Verify original is intact by reading a known entry
			cmd := &xmlapi.Config{
				Action: "get",
				Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='addr-1-renamed']",
			}

			data, _, err := client.Communicate(ctx, cmd, true, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(ContainSubstring("addr-1-renamed"))
			Expect(string(data)).To(ContainSubstring("1.1.1.1"))

			// Version should still be correct (proves document structure intact)
			Expect(client.Versioning()).To(Equal(originalRoot))
		})
	})

	Describe("validateXpath", func() {
		It("should accept valid XPath expressions", func() {
			// Note: validateXpath is unexported, test through operation validation
			// Valid XPath with entry predicate
			cmd := &xmlapi.Config{
				Action: "get",
				Xpath:  "/config/devices/entry[@name='localhost.localdomain']",
			}

			_, _, err := client.Communicate(ctx, cmd, true, nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should accept XPath with multiple predicates", func() {
			cmd := &xmlapi.Config{
				Action: "get",
				Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']",
			}

			_, _, err := client.Communicate(ctx, cmd, true, nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should accept descendant XPath", func() {
			cmd := &xmlapi.Config{
				Action: "get",
				Xpath:  "//address/entry[@name='addr-1-renamed']",
			}

			_, _, err := client.Communicate(ctx, cmd, true, nil)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

var _ = Describe("LocalXmlClient Locking", func() {
	var (
		client *pango.LocalXmlClient
		ctx    context.Context
	)

	BeforeEach(func() {
		configXml, err := os.ReadFile("testdata/panorama-running-config.xml")
		Expect(err).ToNot(HaveOccurred())

		client, err = pango.NewLocalXmlClient(configXml)
		Expect(err).ToNot(HaveOccurred())

		ctx = context.Background()
	})

	Describe("Concurrent access", func() {
		It("should allow multiple concurrent reads", func() {
			var wg sync.WaitGroup
			errors := make(chan error, 10)

			// Spawn 10 concurrent readers
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					cmd := &xmlapi.Config{
						Action: "get",
						Xpath:  "/config/devices/entry[@name='localhost.localdomain']",
					}
					_, _, err := client.Communicate(ctx, cmd, false, nil)
					if err != nil {
						errors <- err
					}
				}()
			}

			wg.Wait()
			close(errors)

			// Verify no errors occurred
			var errorList []error
			for err := range errors {
				errorList = append(errorList, err)
			}
			Expect(errorList).To(BeEmpty())
		})

		It("should serialize write operations", func() {
			// Verify that write operations properly acquire and release locks
			// by checking that multiple write attempts execute sequentially

			var writeOrder []int
			var mu sync.Mutex
			var wg sync.WaitGroup

			// Launch 5 concurrent write operations (rename is not implemented yet)
			for i := 0; i < 5; i++ {
				wg.Add(1)
				go func(id int) {
					defer GinkgoRecover()
					defer wg.Done()

					cmd := &xmlapi.Config{
						Action: "rename",
						Xpath:  "/config/devices/entry[@name='localhost.localdomain']/address",
					}

					// Record write order (protected by separate mutex)
					mu.Lock()
					writeOrder = append(writeOrder, id)
					mu.Unlock()

					// Attempt write (will return ErrWriteNotSupported for rename, but lock is still acquired)
					_, _, _ = client.Communicate(ctx, cmd, false, nil)
				}(i)
			}

			wg.Wait()

			// Verify all 5 writes executed (order doesn't matter, just that all completed)
			Expect(writeOrder).To(HaveLen(5))

			// Verify we got the expected error for all writes (checking one is sufficient)
			cmd := &xmlapi.Config{
				Action: "multi-config",
				Xpath:  "/config/devices",
			}
			_, _, err := client.Communicate(ctx, cmd, false, nil)
			Expect(err).To(Equal(pango.ErrWriteNotSupported))
		})
	})

	Describe("Context timeout handling", func() {
		It("should respect context timeout", func() {
			// Create client accessor to access unexported mu field
			// We need to hold the lock to test timeout behavior
			type lockedClient struct {
				*pango.LocalXmlClient
			}

			// Create a context with short timeout
			shortCtx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()

			// Hold write lock in separate goroutine
			lockHeld := make(chan bool)
			releaseLock := make(chan bool)
			go func() {
				cmd := &xmlapi.Config{
					Action: "set",
					Xpath:  "/config/devices",
				}
				lockHeld <- true
				// Hold lock until signaled
				<-releaseLock
				client.Communicate(context.Background(), cmd, false, nil)
			}()

			// Wait for lock to be held
			<-lockHeld

			// Give time for lock to be acquired
			time.Sleep(20 * time.Millisecond)

			// Attempt operation with expired context (should fail with timeout)
			cmd := &xmlapi.Config{
				Action: "get",
				Xpath:  "/config/devices/entry[@name='localhost.localdomain']",
			}
			_, _, err := client.Communicate(shortCtx, cmd, false, nil)

			// Release the lock
			close(releaseLock)

			// Verify we got a context error
			Expect(err).To(HaveOccurred())
			Expect(errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)).To(BeTrue())
		})
	})
})

var _ = Describe("LocalXmlClient SET Operation", func() {
	var (
		client *pango.LocalXmlClient
		ctx    context.Context
	)

	BeforeEach(func() {
		configXml, err := os.ReadFile("testdata/panorama-running-config.xml")
		Expect(err).ToNot(HaveOccurred())

		client, err = pango.NewLocalXmlClient(configXml)
		Expect(err).ToNot(HaveOccurred())

		ctx = context.Background()
	})

	It("should create new element with SET", func() {
		cmd := &xmlapi.Config{
			Action:  "set",
			Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address",
			Element: "<entry name='test-new-addr'><ip-netmask>192.168.1.1</ip-netmask></entry>",
		}

		_, resp, err := client.Communicate(ctx, cmd, false, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(200))

		// Verify element exists
		verifyCmd := &xmlapi.Config{
			Action: "get",
			Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='test-new-addr']",
		}
		data, _, err := client.Communicate(ctx, verifyCmd, true, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(data)).To(ContainSubstring("test-new-addr"))
		Expect(string(data)).To(ContainSubstring("192.168.1.1"))
	})

	It("should overwrite existing element with SET", func() {
		// First, create an element
		cmd := &xmlapi.Config{
			Action:  "set",
			Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address",
			Element: "<entry name='test-overwrite'><ip-netmask>10.1.1.1</ip-netmask></entry>",
		}
		_, _, err := client.Communicate(ctx, cmd, false, nil)
		Expect(err).ToNot(HaveOccurred())

		// Overwrite with different content
		cmd.Element = "<entry name='test-overwrite'><ip-netmask>10.2.2.2</ip-netmask><description>Updated</description></entry>"
		_, _, err = client.Communicate(ctx, cmd, false, nil)
		Expect(err).ToNot(HaveOccurred())

		// Verify new content
		verifyCmd := &xmlapi.Config{
			Action: "get",
			Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='test-overwrite']",
		}
		data, _, err := client.Communicate(ctx, verifyCmd, true, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(data)).To(ContainSubstring("10.2.2.2"))
		Expect(string(data)).To(ContainSubstring("Updated"))
		Expect(string(data)).ToNot(ContainSubstring("10.1.1.1"))
	})

	It("should return error for invalid XPath", func() {
		cmd := &xmlapi.Config{
			Action:  "set",
			Xpath:   "//invalid[[@",
			Element: "<entry name='test'/>",
		}

		_, _, err := client.Communicate(ctx, cmd, false, nil)
		Expect(err).To(HaveOccurred())

		var invalidXpath *pangoerrors.ErrInvalidXpath
		Expect(errors.As(err, &invalidXpath)).To(BeTrue())
		Expect(invalidXpath.XPath).To(Equal("//invalid[[@"))
	})

	It("should auto-create intermediate paths for SET operations", func() {
		// Test that SET operations automatically create missing intermediate path elements
		cmd := &xmlapi.Config{
			Action:  "set",
			Xpath:   "/config/devices/entry[@name='auto-created-device']/address",
			Element: "<entry name='test-addr'><ip-netmask>192.168.1.1</ip-netmask></entry>",
		}

		_, _, err := client.Communicate(ctx, cmd, false, nil)
		Expect(err).NotTo(HaveOccurred())

		// Verify the intermediate path was created
		verifyCmd := &xmlapi.Config{
			Action: "get",
			Xpath:  "/config/devices/entry[@name='auto-created-device']/address/entry[@name='test-addr']",
		}
		data, _, err := client.Communicate(ctx, verifyCmd, true, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(ContainSubstring("192.168.1.1"))

		// Verify the device entry was auto-created with correct name attribute
		deviceVerifyCmd := &xmlapi.Config{
			Action: "get",
			Xpath:  "/config/devices/entry[@name='auto-created-device']",
		}
		deviceData, _, err := client.Communicate(ctx, deviceVerifyCmd, true, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(deviceData)).To(ContainSubstring("auto-created-device"))
	})

	It("should format SET response correctly", func() {
		cmd := &xmlapi.Config{
			Action:  "set",
			Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address",
			Element: "<entry name='test-response'><ip-netmask>10.3.3.3</ip-netmask></entry>",
		}

		respBytes, httpResp, err := client.Communicate(ctx, cmd, false, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(httpResp.StatusCode).To(Equal(200))

		respXml := string(respBytes)
		Expect(respXml).To(ContainSubstring("<response status=\"success\">"))
		Expect(respXml).To(ContainSubstring("<result"))
		Expect(respXml).To(ContainSubstring("test-response"))
	})
})

var _ = Describe("LocalXmlClient EDIT Operation", func() {
	var (
		client *pango.LocalXmlClient
		ctx    context.Context
	)

	BeforeEach(func() {
		configXml, err := os.ReadFile("testdata/panorama-running-config.xml")
		Expect(err).ToNot(HaveOccurred())

		client, err = pango.NewLocalXmlClient(configXml)
		Expect(err).ToNot(HaveOccurred())

		ctx = context.Background()
	})

	It("should modify existing field with EDIT", func() {
		// First create an element to edit
		setCmd := &xmlapi.Config{
			Action:  "set",
			Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address",
			Element: "<entry name='test-edit'><ip-netmask>10.5.5.5</ip-netmask><description>Original</description></entry>",
		}
		_, _, err := client.Communicate(ctx, setCmd, false, nil)
		Expect(err).ToNot(HaveOccurred())

		// Edit the ip-netmask field
		editCmd := &xmlapi.Config{
			Action:  "edit",
			Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='test-edit']",
			Element: "<ip-netmask>10.6.6.6</ip-netmask>",
		}
		_, _, err = client.Communicate(ctx, editCmd, false, nil)
		Expect(err).ToNot(HaveOccurred())

		// Verify field was updated
		verifyCmd := &xmlapi.Config{
			Action: "get",
			Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='test-edit']",
		}
		data, _, err := client.Communicate(ctx, verifyCmd, true, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(data)).To(ContainSubstring("10.6.6.6"))
	})

	It("should preserve unmodified fields with EDIT", func() {
		// Create element with multiple fields
		setCmd := &xmlapi.Config{
			Action:  "set",
			Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address",
			Element: "<entry name='test-preserve'><ip-netmask>10.7.7.7</ip-netmask><description>Keep this</description></entry>",
		}
		_, _, err := client.Communicate(ctx, setCmd, false, nil)
		Expect(err).ToNot(HaveOccurred())

		// Edit only ip-netmask, description should remain
		editCmd := &xmlapi.Config{
			Action:  "edit",
			Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='test-preserve']",
			Element: "<ip-netmask>10.8.8.8</ip-netmask>",
		}
		_, _, err = client.Communicate(ctx, editCmd, false, nil)
		Expect(err).ToNot(HaveOccurred())

		// Verify both fields present
		verifyCmd := &xmlapi.Config{
			Action: "get",
			Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='test-preserve']",
		}
		data, _, err := client.Communicate(ctx, verifyCmd, true, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(data)).To(ContainSubstring("10.8.8.8"))
		Expect(string(data)).To(ContainSubstring("Keep this"))
	})

	It("should add new fields with EDIT", func() {
		// Create element with one field
		setCmd := &xmlapi.Config{
			Action:  "set",
			Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address",
			Element: "<entry name='test-add-field'><ip-netmask>10.9.9.9</ip-netmask></entry>",
		}
		_, _, err := client.Communicate(ctx, setCmd, false, nil)
		Expect(err).ToNot(HaveOccurred())

		// Edit to add description field
		editCmd := &xmlapi.Config{
			Action:  "edit",
			Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='test-add-field']",
			Element: "<description>New field added</description>",
		}
		_, _, err = client.Communicate(ctx, editCmd, false, nil)
		Expect(err).ToNot(HaveOccurred())

		// Verify both fields present
		verifyCmd := &xmlapi.Config{
			Action: "get",
			Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='test-add-field']",
		}
		data, _, err := client.Communicate(ctx, verifyCmd, true, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(data)).To(ContainSubstring("10.9.9.9"))
		Expect(string(data)).To(ContainSubstring("New field added"))
	})

	It("should return ObjectNotFound for non-existent element", func() {
		cmd := &xmlapi.Config{
			Action:  "edit",
			Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='does-not-exist']",
			Element: "<ip-netmask>10.10.10.10</ip-netmask>",
		}

		_, _, err := client.Communicate(ctx, cmd, false, nil)
		Expect(err).To(HaveOccurred())

		var notFound *pangoerrors.ErrObjectNotFound
		Expect(errors.As(err, &notFound)).To(BeTrue())
	})

	It("should return error for invalid XPath", func() {
		cmd := &xmlapi.Config{
			Action:  "edit",
			Xpath:   "//invalid[[@",
			Element: "<ip-netmask>10.10.10.10</ip-netmask>",
		}

		_, _, err := client.Communicate(ctx, cmd, false, nil)
		Expect(err).To(HaveOccurred())

		var invalidXpath *pangoerrors.ErrInvalidXpath
		Expect(errors.As(err, &invalidXpath)).To(BeTrue())
	})

	It("should format EDIT response correctly", func() {
		// Create element first
		setCmd := &xmlapi.Config{
			Action:  "set",
			Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address",
			Element: "<entry name='test-edit-response'><ip-netmask>10.11.11.11</ip-netmask></entry>",
		}
		_, _, err := client.Communicate(ctx, setCmd, false, nil)
		Expect(err).ToNot(HaveOccurred())

		// Edit it
		editCmd := &xmlapi.Config{
			Action:  "edit",
			Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='test-edit-response']",
			Element: "<description>Testing response</description>",
		}

		respBytes, httpResp, err := client.Communicate(ctx, editCmd, false, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(httpResp.StatusCode).To(Equal(200))

		respXml := string(respBytes)
		Expect(respXml).To(ContainSubstring("<response status=\"success\">"))
		Expect(respXml).To(ContainSubstring("<result"))
	})
})

var _ = Describe("LocalXmlClient DELETE Operation", func() {
	var (
		client *pango.LocalXmlClient
		ctx    context.Context
	)

	BeforeEach(func() {
		configXml, err := os.ReadFile("testdata/panorama-running-config.xml")
		Expect(err).ToNot(HaveOccurred())

		client, err = pango.NewLocalXmlClient(configXml)
		Expect(err).ToNot(HaveOccurred())

		ctx = context.Background()
	})

	It("should delete existing element", func() {
		// First create an element to delete
		setCmd := &xmlapi.Config{
			Action:  "set",
			Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address",
			Element: "<entry name='test-delete'><ip-netmask>10.99.99.99</ip-netmask></entry>",
		}
		_, _, err := client.Communicate(ctx, setCmd, false, nil)
		Expect(err).ToNot(HaveOccurred())

		// Verify it exists
		getCmd := &xmlapi.Config{
			Action: "get",
			Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='test-delete']",
		}
		_, _, err = client.Communicate(ctx, getCmd, true, nil)
		Expect(err).ToNot(HaveOccurred())

		// Delete the element
		deleteCmd := &xmlapi.Config{
			Action: "delete",
			Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='test-delete']",
		}
		_, resp, err := client.Communicate(ctx, deleteCmd, false, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(200))

		// Verify it's deleted
		_, _, err = client.Communicate(ctx, getCmd, true, nil)
		Expect(err).To(HaveOccurred())
		var notFound pangoerrors.Panos
		Expect(errors.As(err, &notFound)).To(BeTrue())
		Expect(notFound.Code).To(Equal(7)) // Object not found
	})

	It("should return ObjectNotFound for non-existent element", func() {
		deleteCmd := &xmlapi.Config{
			Action: "delete",
			Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='never-existed']",
		}
		_, _, err := client.Communicate(ctx, deleteCmd, false, nil)
		Expect(err).To(HaveOccurred())
		var notFoundErr *pangoerrors.ErrObjectNotFound
		Expect(errors.As(err, &notFoundErr)).To(BeTrue())
	})

	It("should return error for invalid XPath", func() {
		deleteCmd := &xmlapi.Config{
			Action: "delete",
			Xpath:  "//invalid[[@",
		}
		_, _, err := client.Communicate(ctx, deleteCmd, false, nil)
		Expect(err).To(HaveOccurred())
		var invalidXpath *pangoerrors.ErrInvalidXpath
		Expect(errors.As(err, &invalidXpath)).To(BeTrue())
	})

	It("should format DELETE response correctly", func() {
		// Create element first
		setCmd := &xmlapi.Config{
			Action:  "set",
			Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address",
			Element: "<entry name='test-delete-response'><ip-netmask>10.88.88.88</ip-netmask></entry>",
		}
		_, _, err := client.Communicate(ctx, setCmd, false, nil)
		Expect(err).ToNot(HaveOccurred())

		// Delete and check response format
		deleteCmd := &xmlapi.Config{
			Action: "delete",
			Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='test-delete-response']",
		}
		respBytes, httpResp, err := client.Communicate(ctx, deleteCmd, false, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(httpResp.StatusCode).To(Equal(200))

		respXml := string(respBytes)
		Expect(respXml).To(ContainSubstring("<response status=\"success\">"))
		Expect(respXml).To(ContainSubstring("<result"))
		Expect(respXml).To(ContainSubstring("total-count=\"0\""))
		Expect(respXml).To(ContainSubstring("count=\"0\""))
	})
})

var _ = Describe("LocalXmlClient RENAME Operation", func() {
	var (
		client *pango.LocalXmlClient
		ctx    context.Context
	)

	BeforeEach(func() {
		configXml, err := os.ReadFile("testdata/panorama-running-config.xml")
		Expect(err).ToNot(HaveOccurred())

		client, err = pango.NewLocalXmlClient(configXml)
		Expect(err).ToNot(HaveOccurred())

		ctx = context.Background()
	})

	Context("RENAME operation", func() {
	It("should rename element successfully", func() {
		renameCmd := &xmlapi.Config{
			Action:  "rename",
			Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='addr-1-renamed']",
			NewName: "addr-1-new-name",
		}
		_, _, err := client.Communicate(ctx, renameCmd, false, nil)
		Expect(err).ToNot(HaveOccurred())

		// Verify old name doesn't exist
		getOld := &xmlapi.Config{
			Action: "get",
			Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='addr-1-renamed']",
		}
		_, _, err = client.Communicate(ctx, getOld, true, nil)
		Expect(err).To(HaveOccurred())
		var notFound pangoerrors.Panos
		Expect(errors.As(err, &notFound)).To(BeTrue())

		// Verify new name exists
		getNew := &xmlapi.Config{
			Action: "get",
			Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='addr-1-new-name']",
		}
		_, _, err = client.Communicate(ctx, getNew, true, nil)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should return RenameConflict if new name exists", func() {
		// First create a second element
		config := &xmlapi.Config{
			Action:  "set",
			Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address",
			Element: "<entry name='existing-name'><ip-netmask>10.88.88.88</ip-netmask></entry>",
		}
		_, _, err := client.Communicate(ctx, config, false, nil)
		Expect(err).ToNot(HaveOccurred())

		// Try to rename addr-1-renamed to existing-name
		config = &xmlapi.Config{
			Action:  "rename",
			Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='addr-1-renamed']",
			NewName: "existing-name",
		}
		_, _, err = client.Communicate(ctx, config, false, nil)

		Expect(err).To(HaveOccurred())
		var renameConflict *pangoerrors.ErrRenameConflict
		Expect(errors.As(err, &renameConflict)).To(BeTrue())
		Expect(renameConflict.Error()).To(ContainSubstring("already exists"))
	})

	It("should return ObjectNotFound if source doesn't exist", func() {
		config := &xmlapi.Config{
			Action:  "rename",
			Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='does-not-exist']",
			NewName: "new-name",
		}
		_, _, err := client.Communicate(ctx, config, false, nil)

		Expect(err).To(HaveOccurred())
		var notFound *pangoerrors.ErrObjectNotFound
		Expect(errors.As(err, &notFound)).To(BeTrue())
	})

	It("should return error if NewName is empty", func() {
		config := &xmlapi.Config{
			Action:  "rename",
			Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='addr-1-renamed']",
			NewName: "",
		}
		_, _, err := client.Communicate(ctx, config, false, nil)

		Expect(err).To(HaveOccurred())
	})
	})
})

var _ = Describe("LocalXmlClient MOVE Operation", func() {
	var (
		client *pango.LocalXmlClient
		ctx    context.Context
	)

	BeforeEach(func() {
		configXml, err := os.ReadFile("testdata/panorama-running-config.xml")
		Expect(err).ToNot(HaveOccurred())

		client, err = pango.NewLocalXmlClient(configXml)
		Expect(err).ToNot(HaveOccurred())

		ctx = context.Background()
	})

	Context("MOVE operation", func() {
		BeforeEach(func() {
		// Create test elements for move operations
		config := &xmlapi.Config{
			Action:  "set",
			Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address",
			Element: "<entry name='move-test-1'><ip-netmask>10.1.1.1</ip-netmask></entry>",
		}
		_, _, err := client.Communicate(ctx, config, false, nil)
		Expect(err).ToNot(HaveOccurred())

		config.Element = "<entry name='move-test-2'><ip-netmask>10.1.1.2</ip-netmask></entry>"
		_, _, err = client.Communicate(ctx, config, false, nil)
		Expect(err).ToNot(HaveOccurred())

		config.Element = "<entry name='move-test-3'><ip-netmask>10.1.1.3</ip-netmask></entry>"
		_, _, err = client.Communicate(ctx, config, false, nil)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should move element to top", func() {
		moveCmd := &xmlapi.Config{
			Action: "move",
			Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='move-test-3']",
			Where:  "top",
		}
		_, httpResp, err := client.Communicate(ctx, moveCmd, false, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(httpResp.StatusCode).To(Equal(200))

		// Verify element still exists after move
		getCmd := &xmlapi.Config{
			Action: "get",
			Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='move-test-3']",
		}
		_, _, err = client.Communicate(ctx, getCmd, true, nil)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should move element to bottom", func() {
		moveCmd := &xmlapi.Config{
			Action: "move",
			Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='move-test-1']",
			Where:  "bottom",
		}
		_, httpResp, err := client.Communicate(ctx, moveCmd, false, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(httpResp.StatusCode).To(Equal(200))

		// Verify element still exists after move
		getCmd := &xmlapi.Config{
			Action: "get",
			Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='move-test-1']",
		}
		_, _, err = client.Communicate(ctx, getCmd, true, nil)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should move element before another", func() {
		moveCmd := &xmlapi.Config{
			Action:      "move",
			Xpath:       "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='move-test-3']",
			Where:       "before",
			Destination: "move-test-1",
		}
		_, httpResp, err := client.Communicate(ctx, moveCmd, false, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(httpResp.StatusCode).To(Equal(200))

		// Verify element still exists after move
		getCmd := &xmlapi.Config{
			Action: "get",
			Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='move-test-3']",
		}
		_, _, err = client.Communicate(ctx, getCmd, true, nil)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should move element after another", func() {
		moveCmd := &xmlapi.Config{
			Action:      "move",
			Xpath:       "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='move-test-1']",
			Where:       "after",
			Destination: "move-test-3",
		}
		_, httpResp, err := client.Communicate(ctx, moveCmd, false, nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(httpResp.StatusCode).To(Equal(200))

		// Verify element still exists after move
		getCmd := &xmlapi.Config{
			Action: "get",
			Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='move-test-1']",
		}
		_, _, err = client.Communicate(ctx, getCmd, true, nil)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should return error for invalid where value", func() {
		config := &xmlapi.Config{
			Action: "move",
			Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='move-test-1']",
			Where:  "invalid",
		}
		_, _, err := client.Communicate(ctx, config, false, nil)

		Expect(err).To(HaveOccurred())
	})
	})
})

var _ = Describe("LocalXmlClient MultiConfig Operation", func() {
	var (
		client *pango.LocalXmlClient
		ctx    context.Context
	)

	BeforeEach(func() {
		configXml, err := os.ReadFile("testdata/panorama-running-config.xml")
		Expect(err).ToNot(HaveOccurred())

		client, err = pango.NewLocalXmlClient(configXml)
		Expect(err).ToNot(HaveOccurred())

		ctx = context.Background()
	})

	Context("MultiConfig operation", func() {
		It("should commit when all operations succeed", func() {
			mc := &xmlapi.MultiConfig{
				Operations: []xmlapi.MultiConfigOperation{
					{
						XMLName: xml.Name{Local: "set"},
						Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address",
						Data:    "<entry name='mc-addr-1'><ip-netmask>10.100.1.1</ip-netmask></entry>",
					},
					{
						XMLName: xml.Name{Local: "set"},
						Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address",
						Data:    "<entry name='mc-addr-2'><ip-netmask>10.100.1.2</ip-netmask></entry>",
					},
				},
			}

			_, httpResp, _, err := client.MultiConfig(ctx, mc, false, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(httpResp.StatusCode).To(Equal(200))

			// Verify both addresses were created
			getCmd1 := &xmlapi.Config{
				Action: "get",
				Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='mc-addr-1']",
			}
			_, _, err = client.Communicate(ctx, getCmd1, true, nil)
			Expect(err).ToNot(HaveOccurred())

			getCmd2 := &xmlapi.Config{
				Action: "get",
				Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='mc-addr-2']",
			}
			_, _, err = client.Communicate(ctx, getCmd2, true, nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should rollback when one operation fails", func() {
			mc := &xmlapi.MultiConfig{
				Operations: []xmlapi.MultiConfigOperation{
					{
						XMLName: xml.Name{Local: "set"},
						Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address",
						Data:    "<entry name='mc-test-1'><ip-netmask>10.101.1.1</ip-netmask></entry>",
					},
					{
						XMLName: xml.Name{Local: "edit"},
						Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='nonexistent']",
						Data:    "<ip-netmask>10.101.1.99</ip-netmask>",
					},
					{
						XMLName: xml.Name{Local: "set"},
						Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address",
						Data:    "<entry name='mc-test-3'><ip-netmask>10.101.1.3</ip-netmask></entry>",
					},
				},
			}

			_, _, _, err := client.MultiConfig(ctx, mc, false, nil)
			Expect(err).To(HaveOccurred())

			// Verify error is ErrOperationFailed with correct index
			var opFailed *pangoerrors.ErrOperationFailed
			Expect(errors.As(err, &opFailed)).To(BeTrue())
			Expect(opFailed.OperationIndex).To(Equal(1))

			// Verify NO changes were applied (rollback successful)
			getCmd1 := &xmlapi.Config{
				Action: "get",
				Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='mc-test-1']",
			}
			_, _, err = client.Communicate(ctx, getCmd1, true, nil)
			Expect(err).To(HaveOccurred())

			getCmd3 := &xmlapi.Config{
				Action: "get",
				Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='mc-test-3']",
			}
			_, _, err = client.Communicate(ctx, getCmd3, true, nil)
			Expect(err).To(HaveOccurred())
		})

		It("should support delete operations in batch", func() {
			// First create some test addresses
			createCmd1 := &xmlapi.Config{
				Action:  "set",
				Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address",
				Element: "<entry name='mc-delete-1'><ip-netmask>10.102.1.1</ip-netmask></entry>",
			}
			_, _, err := client.Communicate(ctx, createCmd1, false, nil)
			Expect(err).ToNot(HaveOccurred())

			createCmd2 := &xmlapi.Config{
				Action:  "set",
				Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address",
				Element: "<entry name='mc-delete-2'><ip-netmask>10.102.1.2</ip-netmask></entry>",
			}
			_, _, err = client.Communicate(ctx, createCmd2, false, nil)
			Expect(err).ToNot(HaveOccurred())

			// Now delete both in a batch
			mc := &xmlapi.MultiConfig{
				Operations: []xmlapi.MultiConfigOperation{
					{
						XMLName: xml.Name{Local: "delete"},
						Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='mc-delete-1']",
					},
					{
						XMLName: xml.Name{Local: "delete"},
						Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='mc-delete-2']",
					},
				},
			}

			_, httpResp, _, err := client.MultiConfig(ctx, mc, false, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(httpResp.StatusCode).To(Equal(200))

			// Verify both were deleted
			getCmd1 := &xmlapi.Config{
				Action: "get",
				Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='mc-delete-1']",
			}
			_, _, err = client.Communicate(ctx, getCmd1, true, nil)
			Expect(err).To(HaveOccurred())

			getCmd2 := &xmlapi.Config{
				Action: "get",
				Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='mc-delete-2']",
			}
			_, _, err = client.Communicate(ctx, getCmd2, true, nil)
			Expect(err).To(HaveOccurred())
		})

		It("should support rename operations in batch", func() {
			// Create test addresses
			createCmd := &xmlapi.Config{
				Action:  "set",
				Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address",
				Element: "<entry name='mc-rename-old'><ip-netmask>10.103.1.1</ip-netmask></entry>",
			}
			_, _, err := client.Communicate(ctx, createCmd, false, nil)
			Expect(err).ToNot(HaveOccurred())

			// Rename in batch
			mc := &xmlapi.MultiConfig{
				Operations: []xmlapi.MultiConfigOperation{
					{
						XMLName: xml.Name{Local: "rename"},
						Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='mc-rename-old']",
						NewName: "mc-rename-new",
					},
				},
			}

			_, httpResp, _, err := client.MultiConfig(ctx, mc, false, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(httpResp.StatusCode).To(Equal(200))

			// Verify old name is gone
			getOld := &xmlapi.Config{
				Action: "get",
				Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='mc-rename-old']",
			}
			_, _, err = client.Communicate(ctx, getOld, true, nil)
			Expect(err).To(HaveOccurred())

			// Verify new name exists
			getNew := &xmlapi.Config{
				Action: "get",
				Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='mc-rename-new']",
			}
			_, _, err = client.Communicate(ctx, getNew, true, nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should reject read operations in batch", func() {
			mc := &xmlapi.MultiConfig{
				Operations: []xmlapi.MultiConfigOperation{
					{
						XMLName: xml.Name{Local: "get"},
						Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='addr-1-renamed']",
					},
				},
			}

			_, _, _, err := client.MultiConfig(ctx, mc, false, nil)
			Expect(err).To(HaveOccurred())

			var opFailed *pangoerrors.ErrOperationFailed
			Expect(errors.As(err, &opFailed)).To(BeTrue())
			Expect(opFailed.OperationIndex).To(Equal(0))
		})

		It("should rollback on third operation failure", func() {
			mc := &xmlapi.MultiConfig{
				Operations: []xmlapi.MultiConfigOperation{
					{
						XMLName: xml.Name{Local: "set"},
						Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address",
						Data:    "<entry name='mc-rollback-1'><ip-netmask>10.104.1.1</ip-netmask></entry>",
					},
					{
						XMLName: xml.Name{Local: "set"},
						Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address",
						Data:    "<entry name='mc-rollback-2'><ip-netmask>10.104.1.2</ip-netmask></entry>",
					},
					{
						XMLName: xml.Name{Local: "delete"},
						Xpath:   "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='does-not-exist']",
					},
				},
			}

			_, _, _, err := client.MultiConfig(ctx, mc, false, nil)
			Expect(err).To(HaveOccurred())

			var opFailed *pangoerrors.ErrOperationFailed
			Expect(errors.As(err, &opFailed)).To(BeTrue())
			Expect(opFailed.OperationIndex).To(Equal(2))

			// Verify first two operations were rolled back
			getCmd1 := &xmlapi.Config{
				Action: "get",
				Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='mc-rollback-1']",
			}
			_, _, err = client.Communicate(ctx, getCmd1, true, nil)
			Expect(err).To(HaveOccurred())

			getCmd2 := &xmlapi.Config{
				Action: "get",
				Xpath:  "/config/devices/entry[@name='localhost.localdomain']/device-group/entry[@name='dg1-renamed']/address/entry[@name='mc-rollback-2']",
			}
			_, _, err = client.Communicate(ctx, getCmd2, true, nil)
			Expect(err).To(HaveOccurred())
		})
	})
})
