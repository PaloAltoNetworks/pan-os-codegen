package main

import (
	"context"
	"flag"
	"log"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/commands/codegen"
)

// Config holds the configuration values for the application
type Config struct {
	ConfigFile string
	OpType     string
}

// parseFlags parses the command line flags
func parseFlags() Config {
	var cfg Config
	flag.StringVar(&cfg.ConfigFile, "config", "./config.yaml", "Path to the configuration file")
	flag.StringVar(&cfg.OpType, "t", "", "Operation type: 'mktp', 'mksdk' or leave empty for both")
	flag.StringVar(&cfg.OpType, "type", "", "Operation type: 'mktp', 'mksdk' or leave empty for both")
	flag.Parse()
	return cfg
}

func executeCommand(ctx context.Context, cmdArgs []string, setupAndExecute func(context.Context, []string) error) {
	err := setupAndExecute(ctx, cmdArgs)
	if err != nil {
		log.Fatalf("There was an error during the execution: %s", err)
	}
}

func main() {
	cfg := parseFlags()

	ctx := context.Background()
	log.SetFlags(log.Ldate | log.Lshortfile)

	// Log the operation type and configuration file being used
	opTypeMessage := "Operation type: "
	if cfg.OpType == "" {
		opTypeMessage += "default option, create both PAN-OS SDK and Terraform Provider"
	} else {
		opTypeMessage += cfg.OpType
	}
	log.Printf("Using configuration file: %s\n", cfg.ConfigFile)
	log.Println(opTypeMessage)

	cmdType := codegen.CommandTypeSDK // Default command type
	if cfg.OpType == "mktp" {
		cmdType = codegen.CommandTypeTerraform
	} else if cfg.OpType == "mksdk" {
		cmdType = codegen.CommandTypeSDK
	}

	if cfg.OpType == "mktp" || cfg.OpType == "mksdk" {
		cmd, err := codegen.NewCommand(ctx, cmdType, cfg.ConfigFile)
		if err != nil {
			log.Fatalf("Failed to create command: %s", err)
		}
		if err := cmd.Setup(); err != nil {
			log.Fatalf("Setup failed: %s", err)
		}
		if err := cmd.Execute(); err != nil {
			log.Fatalf("Execution failed: %s", err)
		}
	} else { // Default behavior to execute both if no specific OpType is provided
		// Execute SDK
		cmdSDK, err := codegen.NewCommand(ctx, codegen.CommandTypeSDK, cfg.ConfigFile)
		if err != nil {
			log.Fatalf("Failed to create command: %s", err)
		}
		if err := cmdSDK.Setup(); err != nil {
			log.Fatalf("Setup SDK failed: %s", err)
		}
		if err := cmdSDK.Execute(); err != nil {
			log.Fatalf("Execution SDK failed: %s", err)
		}

		// Execute Terraform
		cmdTP, err := codegen.NewCommand(ctx, codegen.CommandTypeTerraform, cfg.ConfigFile)
		if err != nil {
			log.Fatalf("Failed to create command: %s", err)
		}
		if err := cmdTP.Setup(); err != nil {
			log.Fatalf("Setup Terraform failed: %s", err)
		}
		if err := cmdTP.Execute(); err != nil {
			log.Fatalf("Execution Terraform failed: %s", err)
		}
	}
}
