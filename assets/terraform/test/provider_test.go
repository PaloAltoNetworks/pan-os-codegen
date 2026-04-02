package provider_test

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	sdk "github.com/PaloAltoNetworks/pango"
	pangoutil "github.com/PaloAltoNetworks/pango/util"
	"github.com/PaloAltoNetworks/terraform-provider-panos/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
)

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary.
	version string = "dev"

	// sdkClient can be either *sdk.Client (API mode) or *sdk.LocalXmlClient (XML mode)
	sdkClient pangoutil.PangoClient

	// Deferred writer lifecycle coordination
	writerWg       sync.WaitGroup
	writerShutdown chan struct{}

	testAccProviders = map[string]func() (tfprotov6.ProviderServer, error){
		"panos": providerserver.NewProtocol6WithError(provider.New(version, &writerWg, writerShutdown)()),
		"echo":  echoprovider.NewProviderServer(),
	}
)

func init() {
	// Initialize shutdown channel for deferred writer coordination
	writerShutdown = make(chan struct{})

	ctx := context.Background()

	// Check if running in local XML mode or API mode
	xmlFilePath := os.Getenv("PANOS_XML_FILE_PATH")

	if xmlFilePath != "" {
		// Local XML mode - create LocalXmlClient
		slog.Info("Initializing test client for local XML mode", "file_path", xmlFilePath)

		localClient, err := sdk.NewLocalXmlClient(
			xmlFilePath,
			sdk.WithAutoSave(false), // Tests manage saves explicitly
			sdk.WithCheckEnvironment(true),
		)
		if err != nil {
			slog.Error("creating local XML client", slog.String("error", err.Error()))
			return
		}

		if err := localClient.Setup(); err != nil {
			slog.Error("setting up local XML client", slog.String("error", err.Error()))
			return
		}

		sdkClient = localClient

		// Initialize deferred writer for XML mode tests
		initTestDeferredWriter()
	} else {
		// API mode - create regular pango.Client
		slog.Info("Initializing test client for API mode")

		apiClient := &sdk.Client{
			CheckEnvironment: true,
		}

		if err := apiClient.Setup(); err != nil {
			slog.Error("setting up pango client", slog.String("error", err.Error()))
			return
		}

		if err := apiClient.Initialize(ctx); err != nil {
			slog.Error("initialization pango client", slog.String("error", err.Error()))
			return
		}

		sdkClient = apiClient
	}
}

// initTestDeferredWriter initializes the deferred background writer for acceptance tests
// running in local XML mode.
//
// Configuration is read from environment variables:
//
//	PANOS_XML_WRITE_MODE - Write mode (safe|deferred|periodic), default: safe
//	PANOS_XML_WRITE_CHECK_INTERVAL_MS - Check interval for deferred mode (5-20), default: 10
//	PANOS_XML_WRITE_FLUSH_INTERVAL_SEC - Flush interval for periodic mode (1-3600), default: 30
func initTestDeferredWriter() {
	// Parse write mode from environment (default: safe)
	writeMode := provider.WriteModeSafe
	if mode := os.Getenv("PANOS_XML_WRITE_MODE"); mode != "" {
		switch mode {
		case "safe":
			writeMode = provider.WriteModeSafe
		case "deferred":
			writeMode = provider.WriteModeDeferred
		case "periodic":
			writeMode = provider.WriteModePeriodic
		default:
			slog.Warn("Invalid PANOS_XML_WRITE_MODE, using safe mode", "mode", mode)
		}
	}

	// Only initialize if not using safe mode (safe mode doesn't need background writer)
	if writeMode == provider.WriteModeSafe {
		return
	}

	// Parse check interval (deferred mode only, default: 10ms)
	checkIntervalMs := 10
	if interval := os.Getenv("PANOS_XML_WRITE_CHECK_INTERVAL_MS"); interval != "" {
		if val, err := strconv.Atoi(interval); err == nil && val >= 5 && val <= 20 {
			checkIntervalMs = val
		} else {
			slog.Warn("Invalid PANOS_XML_WRITE_CHECK_INTERVAL_MS, using default 10", "interval", interval)
		}
	}

	// Parse flush interval (periodic mode only, default: 30s)
	flushIntervalSec := 30
	if interval := os.Getenv("PANOS_XML_WRITE_FLUSH_INTERVAL_SEC"); interval != "" {
		if val, err := strconv.Atoi(interval); err == nil && val >= 1 && val <= 3600 {
			flushIntervalSec = val
		} else {
			slog.Warn("Invalid PANOS_XML_WRITE_FLUSH_INTERVAL_SEC, using default 30", "interval", interval)
		}
	}

	// Get LocalXmlClient from SDK client
	localClient, ok := sdkClient.(*sdk.LocalXmlClient)
	if !ok {
		slog.Info("SDK client is not LocalXmlClient, deferred writer will not initialize")
		return
	}

	// Create writer configuration
	config := provider.WriterConfig{
		Mode:             writeMode,
		CheckIntervalMs:  checkIntervalMs,
		FlushIntervalSec: flushIntervalSec,
	}

	// Initialize the deferred writer
	provider.InitDeferredWriter(config, localClient, &writerWg, writerShutdown)

	slog.Info("Deferred writer initialized for tests",
		"mode", writeMode,
		"check_interval_ms", checkIntervalMs,
		"flush_interval_sec", flushIntervalSec)
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("PANOS_HOSTNAME") == "" && os.Getenv("PANOS_XML_FILE_PATH") == "" {
		t.Fatal("PANOS_HOSTNAME or PANOS_XML_FILE_PATH must be set for acceptance tests")
	}

	if os.Getenv("PANOS_API") != "" {
		return
	}

	if os.Getenv("PANOS_XML_FILE_PATH") == "" && os.Getenv("PANOS_USERNAME") == "" {
		t.Fatal("PANOS_USERNAME must be set for acceptance tests")
	}

	if os.Getenv("PANOS_XML_FILE_PATH") == "" && os.Getenv("PANOS_PASSWORD") == "" {
		t.Fatal("PANOS_PASSWORD must be set for acceptance tests")
	}
}

// createTempXMLCopy creates an isolated temporary copy of the XML configuration file.
//
// The temp file is:
//   - Created in the system temp directory (os.TempDir())
//   - Named with pattern "panos-test-*.xml" (unique random suffix)
//   - Byte-for-byte copy of the original file
//   - Automatically cleaned up after test (unless TF_TEST_NO_CLEANUP=1/true)
//
// Behavior:
//   - Reads entire original file into memory (assumes <10MB typical size)
//   - Creates temp file with unique name to prevent collisions
//   - Writes exact copy of original content
//   - Registers cleanup via t.Cleanup() to delete temp file
//   - Cleanup runs even if test panics or fails
//
// Environment Variables:
//
//	TF_TEST_NO_CLEANUP: Set to "1" or "true" to preserve temp files for debugging
//	  - Default (unset/other values): Temp files are cleaned up
//	  - Preserved files remain in temp directory for manual inspection
//
// Error Handling:
//   - Original file unreadable: t.Fatalf with file path and error
//   - Temp file creation fails: t.Fatalf with error (check permissions)
//   - Write fails: Cleans up partial file and t.Fatalf
//   - Cleanup fails: Logs warning, doesn't fail test
//
// Parameters:
//
//	t: Test context for cleanup registration and failure reporting
//	originalPath: Absolute path to original XML file to copy
//
// Returns:
//
//	Absolute path to created temporary XML file
//
// Example:
//
//	tempPath := createTempXMLCopy(t, "/tmp/panos-config.xml")
//	// tempPath might be: "/tmp/panos-test-abc123.xml"
//	// File will be deleted when test completes (unless TF_TEST_NO_CLEANUP=1)
func createTempXMLCopy(t *testing.T, originalPath string) string {
	// Read original file
	data, err := os.ReadFile(originalPath)
	if err != nil {
		t.Fatalf("Failed to read original XML file %s: %v", originalPath, err)
	}

	// Create temp file with unique name
	tempFile, err := os.CreateTemp("", "panos-test-*.xml")
	if err != nil {
		t.Fatalf("Failed to create temp XML file: %v", err)
	}
	tempPath := tempFile.Name()

	// Write copy
	if _, err := tempFile.Write(data); err != nil {
		tempFile.Close()
		os.Remove(tempPath) // Clean up on failure
		t.Fatalf("Failed to write temp XML file: %v", err)
	}
	tempFile.Close()

	// Check TF_TEST_NO_CLEANUP environment variable
	noCleanup := os.Getenv("TF_TEST_NO_CLEANUP")
	shouldCleanup := true

	if noCleanup == "1" || strings.ToLower(noCleanup) == "true" {
		shouldCleanup = false
	}

	// Conditionally register cleanup
	if shouldCleanup {
		t.Cleanup(func() {
			if err := os.Remove(tempPath); err != nil {
				t.Logf("Warning: failed to cleanup temp XML file %s: %v", tempPath, err)
			}
		})
	} else {
		t.Logf("Preserving temp XML file for debugging: %s", tempPath)
	}

	return tempPath
}

// testAccProviderFactories returns Terraform provider factories with test isolation.
//
// In local XML mode (PANOS_XML_FILE_PATH set):
//   - Creates an isolated temporary XML file copy for this test
//   - Registers cleanup to delete temp file (unless TF_TEST_NO_CLEANUP=1/true)
//   - Temporarily overrides PANOS_XML_FILE_PATH env var for test duration
//   - Restores original env var on cleanup
//
// In API mode (PANOS_XML_FILE_PATH not set):
//   - Returns original global testAccProviders unchanged
//   - No file copying or environment modification
//
// Usage:
//
//	func TestAccResource_Basic(t *testing.T) {
//	    t.Parallel()
//	    resource.Test(t, resource.TestCase{
//	        ProtoV6ProviderFactories: testAccProviderFactories(t),
//	        // ... rest of test
//	    })
//	}
//
// Environment Variables:
//   - PANOS_XML_FILE_PATH: Original XML config file path (mode detection)
//   - TF_TEST_NO_CLEANUP: Set to "1" or "true" to preserve temp files for debugging
//
// Thread Safety:
//
//	Environment variable override is process-global but safe because:
//	- Each test worker runs sequentially (t.Parallel() only parallelizes across tests)
//	- resource.Test() is synchronous within each worker
//	- provider.Configure() reads env var before worker continues
//	- t.Cleanup() restores env var after test completes
//
// Returns:
//
//	Provider factory map compatible with ProtoV6ProviderFactories field
func testAccProviderFactories(t *testing.T) map[string]func() (tfprotov6.ProviderServer, error) {
	// Detect test mode
	xmlPath := os.Getenv("PANOS_XML_FILE_PATH")
	if xmlPath == "" {
		// API mode - no isolation needed
		return testAccProviders
	}

	// Local XML mode - create temp copy
	tempXMLPath := createTempXMLCopy(t, xmlPath)

	// Override environment variable
	originalEnv := os.Getenv("PANOS_XML_FILE_PATH")
	os.Setenv("PANOS_XML_FILE_PATH", tempXMLPath)

	t.Cleanup(func() {
		os.Setenv("PANOS_XML_FILE_PATH", originalEnv)
	})

	// Return original factories (which now read temp path from environment)
	return testAccProviders
}
