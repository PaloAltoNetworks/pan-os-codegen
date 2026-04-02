package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
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
