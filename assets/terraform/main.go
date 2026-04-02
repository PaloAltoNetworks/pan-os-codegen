package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strconv"
	"sync"
	"syscall"

	"github.com/PaloAltoNetworks/terraform-provider-panos/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Run "go generate" to format example terraform files and generate the docs for the registry/website

// If you do not have terraform installed, you can remove the formatting command, but its suggested to
// ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name panos

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary.
	version string = "dev"

	// goreleaser can pass other information to the main package, such as the specific commit
	// https://goreleaser.com/cookbooks/using-main.version/
)

// initProfiling initializes CPU, memory, mutex, and block profiling based on environment variables.
// Returns a cleanup function that should be called before the program exits.
func initProfiling() (cleanup func()) {
	cpuProfile := os.Getenv("TF_PROF_CPU")
	memProfile := os.Getenv("TF_PROF_MEM")

	var cleanupFuncs []func()

	// Start CPU profiling if requested
	if cpuProfile != "" {
		f, err := os.Create(cpuProfile)
		if err != nil {
			log.Printf("Could not create CPU profile: %v", err)
		} else {
			if err := pprof.StartCPUProfile(f); err != nil {
				log.Printf("Could not start CPU profile: %v", err)
				f.Close()
			} else {
				cleanupFuncs = append(cleanupFuncs, func() {
					pprof.StopCPUProfile()
					f.Close()
				})
			}
		}
	}

	// Setup memory profiling cleanup if requested
	if memProfile != "" {
		cleanupFuncs = append(cleanupFuncs, func() {
			f, err := os.Create(memProfile)
			if err != nil {
				log.Printf("Could not create memory profile: %v", err)
				return
			}
			defer f.Close()

			runtime.GC() // get up-to-date statistics
			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Printf("Could not write memory profile: %v", err)
			}
		})
	}

	// Mutex profiling
	mutexProfile := os.Getenv("TF_PROF_MUTEX")
	if mutexProfile != "" {
		// Parse sampling rate (default: 1 = every event)
		mutexRate := 1
		if rateStr := os.Getenv("TF_PROF_MUTEX_RATE"); rateStr != "" {
			if rate, err := strconv.Atoi(rateStr); err == nil && rate > 0 {
				mutexRate = rate
			} else {
				log.Printf("Invalid TF_PROF_MUTEX_RATE '%s', using default 1: %v", rateStr, err)
			}
		}

		log.Printf("Mutex profiling enabled: file=%s rate=%d", mutexProfile, mutexRate)

		// Enable mutex profiling
		runtime.SetMutexProfileFraction(mutexRate)

		// Write profile on shutdown
		cleanupFuncs = append(cleanupFuncs, func() {
			f, err := os.Create(mutexProfile)
			if err != nil {
				log.Printf("Could not create mutex profile: %v", err)
				return
			}
			defer f.Close()

			if err := pprof.Lookup("mutex").WriteTo(f, 0); err != nil {
				log.Printf("Could not write mutex profile: %v", err)
			}
		})
	}

	// Block profiling
	blockProfile := os.Getenv("TF_PROF_BLOCK")
	if blockProfile != "" {
		// Parse sampling rate (default: 1 = every event)
		blockRate := 1
		if rateStr := os.Getenv("TF_PROF_BLOCK_RATE"); rateStr != "" {
			if rate, err := strconv.Atoi(rateStr); err == nil && rate > 0 {
				blockRate = rate
			} else {
				log.Printf("Invalid TF_PROF_BLOCK_RATE '%s', using default 1: %v", rateStr, err)
			}
		}

		log.Printf("Block profiling enabled: file=%s rate=%d", blockProfile, blockRate)

		// Enable block profiling
		runtime.SetBlockProfileRate(blockRate)

		// Write profile on shutdown
		cleanupFuncs = append(cleanupFuncs, func() {
			f, err := os.Create(blockProfile)
			if err != nil {
				log.Printf("Could not create block profile: %v", err)
				return
			}
			defer f.Close()

			if err := pprof.Lookup("block").WriteTo(f, 0); err != nil {
				log.Printf("Could not write block profile: %v", err)
			}
		})
	}

	// Return combined cleanup function
	return func() {
		for i := len(cleanupFuncs) - 1; i >= 0; i-- {
			cleanupFuncs[i]()
		}
	}
}

// setupSignalHandler sets up graceful shutdown on interrupt signals.
// This ensures profiling data is properly written even when the provider is terminated.
func setupSignalHandler(cleanup func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		if cleanup != nil {
			cleanup()
		}
		os.Exit(0)
	}()
}

func main() {
	var (
		debug bool
		wg    sync.WaitGroup
	)

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	// Initialize profiling and get cleanup function
	cleanup := initProfiling()
	if cleanup != nil {
		defer cleanup()
		setupSignalHandler(cleanup)
	}

	// Create shutdown channel for coordinating background writer lifecycle
	shutdownChan := make(chan struct{})
	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/paloaltonetworks/panos",
		Debug:   debug,
	}

	// Pass lifecycle coordination to provider
	err := providerserver.Serve(context.Background(), provider.New(version, &wg, shutdownChan), opts)

	if err != nil {
		log.Fatal(err.Error())
	}

	// === SHUTDOWN COORDINATION ===
	// After Serve() returns, Terraform has disconnected
	close(shutdownChan) // Signal background writers to stop
	wg.Wait()           // Block process exit until writers finish
	// Process exits after final flush completes
}
