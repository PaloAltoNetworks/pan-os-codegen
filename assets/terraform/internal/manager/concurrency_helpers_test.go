package manager_test

import (
	"fmt"
	"sync/atomic"
	"time"
)

// AtomicInt32 wraps atomic operations for concurrency tests
type AtomicInt32 struct {
	value atomic.Int32
}

func NewAtomicInt32() *AtomicInt32 {
	return &AtomicInt32{}
}

func (a *AtomicInt32) Add(delta int32) {
	a.value.Add(delta)
}

func (a *AtomicInt32) Load() int32 {
	return a.value.Load()
}

// Sleep pauses execution for testing (milliseconds)
func Sleep(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

// Format is a wrapper for fmt.Sprintf
func Format(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

// BuildLocationXpath builds the location xpath for cache operations
// Must match manager's buildLocationXpath logic (entry.go:95-107)
func BuildLocationXpath(location MockLocation, client *MockEntryClient[*MockEntryObject]) string {
	// Replicate manager's buildLocationXpath:
	// xpath, err := location.XpathWithComponents(client.Versioning(), append(parentComponents, util.AsEntryXpath(""))...)
	// return util.AsXpath(xpath)

	components, err := location.XpathWithComponents(client.Versioning(), "")
	if err != nil {
		panic(fmt.Sprintf("BuildLocationXpath failed: %v", err))
	}

	// util.AsXpath for []string builds: fmt.Sprintf("/%s", strings.Join(val, "/"))
	xpath := ""
	for i, comp := range components {
		if i == 0 {
			xpath = "/" + comp
		} else {
			xpath += "/" + comp
		}
	}
	return xpath
}
