package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/paloaltonetworks/pan-os-codegen/pkg/commands/codegen"
	"github.com/paloaltonetworks/pan-os-codegen/pkg/properties"
)

// Config holds the configuration values for the application.
type Config struct {
	ConfigFile string
	OpType     string
}

// parseFlags parses the command line flags.
func parseFlags() Config {
	var cfg Config
	flag.StringVar(&cfg.ConfigFile, "config", "./cmd/codegen/config.yaml", "Path to the configuration file")
	flag.StringVar(&cfg.OpType, "t", "", "Operation type: 'mktp', 'mksdk' or leave empty for both")
	flag.StringVar(&cfg.OpType, "type", "", "Operation type: 'mktp', 'mksdk' or leave empty for both")
	flag.Parse()
	return cfg
}

// runCommand executed command to generate code for SDK or Terraform.
func runCommand(ctx context.Context, cmdType properties.CommandType, cfg string) {
	cmd, err := codegen.NewCommand(ctx, cmdType, cfg)
	if err != nil {
		log.Fatalf("Failed to create command: %s", err)
	}
	if err := cmd.Setup(); err != nil {
		log.Fatalf("Setup failed: %s", err)
	}
	if err := cmd.Execute(); err != nil {
		log.Fatalf("Execution failed: %s", err)
	}
}

func main() {
	logLevel := os.Getenv("CODEGEN_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "ERROR"
	}
	var level slog.Level
	var err = level.UnmarshalText([]byte(logLevel))
	if err != nil {
		fmt.Print(err.Error())
		return
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)

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
	slog.Debug("Parsed configuration file", "path", cfg.ConfigFile)

	cmdType := properties.CommandTypeSDK // Default command type
	if cfg.OpType == "mktp" {
		cmdType = properties.CommandTypeTerraform
	} else if cfg.OpType == "mksdk" {
		cmdType = properties.CommandTypeSDK
	}

	if cfg.OpType == "mktp" || cfg.OpType == "mksdk" {
		runCommand(ctx, cmdType, cfg.ConfigFile)
	} else { // Default behavior to execute both if no specific OpType is provided
		// Execute SDK
		runCommand(ctx, properties.CommandTypeSDK, cfg.ConfigFile)

		// Execute Terraform
		runCommand(ctx, properties.CommandTypeTerraform, cfg.ConfigFile)
	}

	log.Println("Generation complete.")
}
