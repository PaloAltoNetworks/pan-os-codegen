package provider_test

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	pangoversion "github.com/PaloAltoNetworks/pango/version"
)

// TestMode is a bitmask of supported test execution modes.
type TestMode uint8

const (
	LocalXml TestMode = 1 << iota
	Panorama
	Firewall
)

func (m TestMode) String() string {
	if m == 0 {
		return "none"
	}

	var names []string
	if m&LocalXml != 0 {
		names = append(names, "local-xml")
	}
	if m&Panorama != 0 {
		names = append(names, "panorama")
	}
	if m&Firewall != 0 {
		names = append(names, "firewall")
	}
	return strings.Join(names, "|")
}

// VersionRange represents a supported PAN-OS version range.
// Min is inclusive, Max is exclusive: [Min, Max).
type VersionRange struct {
	Min string // e.g. "11.0.0"; empty = no lower bound
	Max string // e.g. "11.2.0"; empty = no upper bound
}

// TestConfig describes the execution requirements for a test.
type TestConfig struct {
	// Modes is a bitmask of supported execution modes.
	// Zero value means the test runs in all modes (no mode check).
	Modes TestMode

	// MinVersion is a shorthand for a single VersionRange with only a lower bound.
	// Cannot be combined with Versions.
	MinVersion string

	// Versions is a list of version ranges. The test runs if the current version
	// falls within ANY range. Empty list means no version constraint.
	// Cannot be combined with MinVersion.
	Versions []VersionRange
}

var (
	detectedMode    TestMode
	detectedVersion pangoversion.Number
	detectOnce      sync.Once
	detectErr       error
)

func detectEnvironment() {
	if sdkClient == nil {
		detectErr = fmt.Errorf("sdkClient not initialized")
		return
	}

	xmlFilePath := os.Getenv("PANOS_XML_FILE_PATH")
	if xmlFilePath != "" {
		detectedMode = LocalXml
	} else {
		isPanorama, err := sdkClient.IsPanorama()
		if err != nil {
			detectErr = fmt.Errorf("detecting panorama mode: %w", err)
			return
		}
		if isPanorama {
			detectedMode = Panorama
		} else {
			detectedMode = Firewall
		}
	}

	detectedVersion = sdkClient.Versioning()
}

func isZeroVersion(v pangoversion.Number) bool {
	return v.Major == 0 && v.Minor == 0 && v.Patch == 0
}

func versionInRange(v pangoversion.Number, r VersionRange) bool {
	if r.Min != "" {
		min, err := pangoversion.New(r.Min)
		if err != nil {
			panic(fmt.Sprintf("invalid min version %q: %v", r.Min, err))
		}
		if !v.Gte(min) {
			return false
		}
	}
	if r.Max != "" {
		max, err := pangoversion.New(r.Max)
		if err != nil {
			panic(fmt.Sprintf("invalid max version %q: %v", r.Max, err))
		}
		if !v.Lt(max) {
			return false
		}
	}
	return true
}

func versionInRanges(v pangoversion.Number, ranges []VersionRange) bool {
	for _, r := range ranges {
		if versionInRange(v, r) {
			return true
		}
	}
	return false
}

func formatVersionRanges(ranges []VersionRange) string {
	parts := make([]string, len(ranges))
	for i, r := range ranges {
		switch {
		case r.Min != "" && r.Max != "":
			parts[i] = fmt.Sprintf("[%s, %s)", r.Min, r.Max)
		case r.Min != "":
			parts[i] = fmt.Sprintf("[%s, ...)", r.Min)
		case r.Max != "":
			parts[i] = fmt.Sprintf("[..., %s)", r.Max)
		default:
			parts[i] = "[..., ...)"
		}
	}
	return strings.Join(parts, " or ")
}

// testAccSetup validates that the current execution environment matches the
// test's requirements and skips the test if it does not.
//
// Call this at the beginning of each acceptance test, after t.Parallel():
//
//	func TestAccAddress_Basic(t *testing.T) {
//	    t.Parallel()
//	    testAccSetup(t, TestConfig{
//	        Modes:      LocalXml | Panorama,
//	        MinVersion: "11.0.0",
//	    })
//	    // ... rest of test
//	}
func testAccSetup(t *testing.T, cfg TestConfig) {
	t.Helper()

	detectOnce.Do(detectEnvironment)
	if detectErr != nil {
		t.Fatalf("test setup: environment detection failed: %v", detectErr)
	}

	if cfg.MinVersion != "" && len(cfg.Versions) > 0 {
		t.Fatal("test setup: MinVersion and Versions are mutually exclusive in TestConfig")
	}

	// Mode check
	if cfg.Modes != 0 && (cfg.Modes&detectedMode) == 0 {
		t.Skipf("skipping: test supports modes [%s], current mode is %s",
			cfg.Modes, detectedMode)
	}

	// Version check
	ranges := cfg.Versions
	if cfg.MinVersion != "" {
		ranges = []VersionRange{{Min: cfg.MinVersion}}
	}

	if len(ranges) > 0 {
		if isZeroVersion(detectedVersion) {
			t.Skipf("skipping: PAN-OS version unknown (%s), test requires version constraints %s",
				detectedVersion, formatVersionRanges(ranges))
		}

		if !versionInRanges(detectedVersion, ranges) {
			t.Skipf("skipping: PAN-OS version %s not in supported ranges %s",
				detectedVersion, formatVersionRanges(ranges))
		}
	}
}
