package provider

import (
	"sync"
	"testing"
	"time"
)

// Integration tests for DeferredWriter
// These tests verify the entire system working together, including:
// - SDK Manager runtime mode detection
// - Configuration flow (main → provider → DeferredWriter)
// - Write deduplication
// - Safe mode behavior
// - Deferred/periodic mode conditional calls

// T034: Test configuration flow (main → provider → DeferredWriter)
func TestConfigurationFlow(t *testing.T) {
	defer resetSingleton()

	// Test each configuration mode
	testCases := []struct {
		name   string
		config WriterConfig
	}{
		{
			name: "Safe mode configuration",
			config: WriterConfig{
				Mode:             WriteModeSafe,
				CheckIntervalMs:  10,
				FlushIntervalSec: 30,
			},
		},
		{
			name: "Deferred mode configuration",
			config: WriterConfig{
				Mode:             WriteModeDeferred,
				CheckIntervalMs:  10,
				FlushIntervalSec: 30,
			},
		},
		{
			name: "Periodic mode configuration",
			config: WriterConfig{
				Mode:             WriteModePeriodic,
				CheckIntervalMs:  10,
				FlushIntervalSec: 1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset singleton between test cases
			resetSingleton()
			wg := &sync.WaitGroup{}
			shutdownChan := make(chan struct{})

			client := &mockClient{}

			// Initialize DeferredWriter through configuration flow
			initDeferredWriter(tc.config, client, wg, shutdownChan)

			// Verify singleton was initialized
			if writerInstance == nil {
				t.Fatal("DeferredWriter singleton not initialized")
			}

			// Verify configuration was applied
			if writerInstance.mode != tc.config.Mode {
				t.Errorf("Expected mode %v, got %v", tc.config.Mode, writerInstance.mode)
			}

			if writerInstance.checkIntervalMs != tc.config.CheckIntervalMs {
				t.Errorf("Expected checkIntervalMs %d, got %d", tc.config.CheckIntervalMs, writerInstance.checkIntervalMs)
			}

			if writerInstance.flushIntervalSec != tc.config.FlushIntervalSec {
				t.Errorf("Expected flushIntervalSec %d, got %d", tc.config.FlushIntervalSec, writerInstance.flushIntervalSec)
			}

			// Cleanup
			close(shutdownChan)
			if tc.config.Mode != WriteModeSafe {
				// Wait for background goroutine to stop
				done := make(chan struct{})
				go func() {
					wg.Wait()
					close(done)
				}()
				select {
				case <-done:
				case <-time.After(500 * time.Millisecond):
					t.Fatal("Timeout waiting for shutdown")
				}
			}
		})
	}
}

// T035: Test write deduplication (multiple SetDirty calls → single write)
func TestWriteDeduplication(t *testing.T) {
	defer resetSingleton()

	wg := &sync.WaitGroup{}
	shutdownChan := make(chan struct{})

	config := WriterConfig{
		Mode:             WriteModeDeferred,
		CheckIntervalMs:  50, // Longer interval to allow batching
		FlushIntervalSec: 30,
	}

	client := &mockClient{}
	initDeferredWriter(config, client, wg, shutdownChan)

	// Call SetDirty multiple times rapidly (100 times)
	for i := 0; i < 100; i++ {
		SetDirty()
		time.Sleep(1 * time.Millisecond) // Very short delay
	}

	// Wait for writes to occur (50ms interval + some buffer)
	time.Sleep(150 * time.Millisecond)

	// Get write count
	writeCount := client.GetSaveCallCount()

	// Cleanup
	close(shutdownChan)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Timeout waiting for shutdown")
	}

	// Verify deduplication occurred
	// With 100 SetDirty calls over 100ms and 50ms check interval,
	// we expect significantly fewer than 100 writes (should be ~3-5 writes)
	t.Logf("SetDirty calls: 100, Actual writes: %d", writeCount)

	if writeCount >= 10 {
		t.Errorf("Expected write deduplication to reduce writes to <10, got %d writes", writeCount)
	}

	if writeCount == 0 {
		t.Error("Expected at least one write to occur")
	}

	// Calculate deduplication ratio
	deduplicationRatio := float64(100) / float64(writeCount)
	t.Logf("Deduplication ratio: %.1fx (100 calls → %d writes)", deduplicationRatio, writeCount)

	// Verify at least 10x deduplication (100 calls → ≤10 writes)
	if deduplicationRatio < 10.0 {
		t.Errorf("Expected at least 10x deduplication, got %.1fx", deduplicationRatio)
	}
}

// T036: Test safe mode (no SetDirty/CheckError calls)
func TestSafeModeNoBackgroundCalls(t *testing.T) {
	defer resetSingleton()

	wg := &sync.WaitGroup{}
	shutdownChan := make(chan struct{})
	defer close(shutdownChan)

	config := WriterConfig{
		Mode:             WriteModeSafe,
		CheckIntervalMs:  10,
		FlushIntervalSec: 30,
	}

	client := &mockClient{}
	initDeferredWriter(config, client, wg, shutdownChan)

	// In safe mode, SetDirty() should be a no-op
	SetDirty()
	SetDirty()
	SetDirty()

	// Wait a moment
	time.Sleep(50 * time.Millisecond)

	// Verify no writes occurred (safe mode doesn't use background writer)
	writeCount := client.GetSaveCallCount()
	if writeCount != 0 {
		t.Errorf("Safe mode should not trigger writes, got %d writes", writeCount)
	}

	// Verify CheckError returns nil (no background errors in safe mode)
	err := CheckError()
	if err != nil {
		t.Errorf("CheckError should return nil in safe mode, got: %v", err)
	}
}

// T037: Test deferred/periodic modes (conditional calls)
func TestDeferredPeriodicModeConditionalCalls(t *testing.T) {
	testCases := []struct {
		name           string
		mode           XmlWriteMode
		checkInterval  int
		flushInterval  int
		expectWrites   bool
		expectSetDirty bool
	}{
		{
			name:           "Deferred mode uses SetDirty",
			mode:           WriteModeDeferred,
			checkInterval:  10,
			flushInterval:  30,
			expectWrites:   true,
			expectSetDirty: true,
		},
		{
			name:           "Periodic mode uses timer",
			mode:           WriteModePeriodic,
			checkInterval:  10,
			flushInterval:  1, // 1 second for faster test
			expectWrites:   true,
			expectSetDirty: false, // Periodic mode doesn't use dirty flag
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer resetSingleton()

			wg := &sync.WaitGroup{}
			shutdownChan := make(chan struct{})

			config := WriterConfig{
				Mode:             tc.mode,
				CheckIntervalMs:  tc.checkInterval,
				FlushIntervalSec: tc.flushInterval,
			}

			client := &mockClient{}
			initDeferredWriter(config, client, wg, shutdownChan)

			if tc.mode == WriteModeDeferred {
				// Test SetDirty functionality
				SetDirty()
				time.Sleep(time.Duration(tc.checkInterval*2) * time.Millisecond)
			} else if tc.mode == WriteModePeriodic {
				// Wait for periodic timer
				time.Sleep(time.Duration(tc.flushInterval+1) * time.Second)
			}

			writeCount := client.GetSaveCallCount()

			// Cleanup
			close(shutdownChan)
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()
			select {
			case <-done:
			case <-time.After(500 * time.Millisecond):
				t.Fatal("Timeout waiting for shutdown")
			}

			// Verify writes occurred if expected
			if tc.expectWrites {
				if writeCount == 0 {
					t.Errorf("%s: Expected writes to occur, got 0", tc.name)
				}
			}

			// For deferred mode, verify dirty flag was used
			if tc.mode == WriteModeDeferred && tc.expectSetDirty {
				dirtyFlips := writerInstance.dirtyFlagFlips.Load()
				if dirtyFlips == 0 {
					t.Error("Deferred mode: Expected dirty flag flips to be tracked")
				}
			}

			// For periodic mode, verify no dirty flag usage
			if tc.mode == WriteModePeriodic && !tc.expectSetDirty {
				dirtyFlips := writerInstance.dirtyFlagFlips.Load()
				if dirtyFlips != 0 {
					t.Errorf("Periodic mode: Expected no dirty flag usage, got %d flips", dirtyFlips)
				}
			}
		})
	}
}

// T033: Test SDK Manager runtime mode detection simulation
// Note: This test simulates SDK Manager behavior since actual SDK Manager
// integration requires generated code. In production, SDK Managers will
// call SetDirty() and CheckError() based on runtime mode detection.
func TestSDKManagerModeDetection(t *testing.T) {
	testCases := []struct {
		name                string
		mode                XmlWriteMode
		shouldCallSetDirty  bool
		shouldCallCheckErr bool
	}{
		{
			name:                "Safe mode - no calls",
			mode:                WriteModeSafe,
			shouldCallSetDirty:  false,
			shouldCallCheckErr: false,
		},
		{
			name:                "Deferred mode - both calls",
			mode:                WriteModeDeferred,
			shouldCallSetDirty:  true,
			shouldCallCheckErr: true,
		},
		{
			name:                "Periodic mode - only CheckError",
			mode:                WriteModePeriodic,
			shouldCallSetDirty:  false, // Periodic doesn't use dirty flag
			shouldCallCheckErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer resetSingleton()

			wg := &sync.WaitGroup{}
			shutdownChan := make(chan struct{})

			config := WriterConfig{
				Mode:             tc.mode,
				CheckIntervalMs:  10,
				FlushIntervalSec: 30,
			}

			client := &mockClient{}
			initDeferredWriter(config, client, wg, shutdownChan)

			// Simulate SDK Manager runtime mode detection
			// In actual implementation, SDK Managers will call these based on mode
			if tc.shouldCallSetDirty {
				SetDirty()
			}

			if tc.shouldCallCheckErr {
				err := CheckError()
				if err != nil {
					t.Errorf("Unexpected error from CheckError: %v", err)
				}
			}

			// Verify behavior matches expectations
			if tc.mode == WriteModeSafe {
				// Safe mode: verify no writes triggered
				time.Sleep(50 * time.Millisecond)
				if client.GetSaveCallCount() != 0 {
					t.Error("Safe mode should not trigger writes")
				}
			} else if tc.mode == WriteModeDeferred && tc.shouldCallSetDirty {
				// Deferred mode: verify dirty flag was set
				if !writerInstance.dirtyFlag.Load() {
					t.Error("Deferred mode: dirty flag should be set after SetDirty()")
				}
			}

			// Cleanup for non-safe modes
			if tc.mode != WriteModeSafe {
				close(shutdownChan)
				done := make(chan struct{})
				go func() {
					wg.Wait()
					close(done)
				}()
				select {
				case <-done:
				case <-time.After(500 * time.Millisecond):
					t.Fatal("Timeout waiting for shutdown")
				}
			} else {
				// Safe mode: just close the channel
				close(shutdownChan)
			}
		})
	}
}

// TestIntegrationEndToEnd tests the complete flow from configuration to shutdown
func TestIntegrationEndToEnd(t *testing.T) {
	defer resetSingleton()

	wg := &sync.WaitGroup{}
	shutdownChan := make(chan struct{})

	// Configure in deferred mode
	config := WriterConfig{
		Mode:             WriteModeDeferred,
		CheckIntervalMs:  10,
		FlushIntervalSec: 30,
	}

	client := &mockClient{}

	// 1. Initialize (simulates main → provider → DeferredWriter)
	initDeferredWriter(config, client, wg, shutdownChan)

	// 2. Simulate SDK Manager operations
	for i := 0; i < 10; i++ {
		SetDirty() // Simulate resource modifications

		// Check for errors (fail-fast)
		if err := CheckError(); err != nil {
			t.Fatalf("Unexpected error during operations: %v", err)
		}

		time.Sleep(2 * time.Millisecond)
	}

	// 3. Wait for writes to occur
	time.Sleep(30 * time.Millisecond)

	// 4. Verify writes happened
	initialWrites := client.GetSaveCallCount()
	if initialWrites == 0 {
		t.Error("Expected writes to occur during operation")
	}

	// 5. Shutdown (simulates main.go shutdown coordination)
	close(shutdownChan)

	// 6. Wait for final flush
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - shutdown completed
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for graceful shutdown")
	}

	// 7. Verify final flush occurred if needed
	finalWrites := client.GetSaveCallCount()
	t.Logf("Total writes during end-to-end test: %d", finalWrites)

	if finalWrites == 0 {
		t.Error("Expected at least one write during end-to-end test")
	}
}
