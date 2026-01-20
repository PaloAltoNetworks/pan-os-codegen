package main

import (
	"context"
	"flag"
	"fmt"
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
		slog.Error("Failed to create command", "error", err)
		os.Exit(1)
	}
	if err := cmd.Setup(); err != nil {
		slog.Error("Setup failed", "error", err)
		os.Exit(1)
	}
	if err := cmd.Execute(); err != nil {
		slog.Error("Execution failed", "error", err)
		os.Exit(1)
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

	// Log the operation type and configuration file being used
	opTypeMessage := "Operation type: "
	if cfg.OpType == "" {
		opTypeMessage += "default option, create both PAN-OS SDK and Terraform Provider"
	} else {
		opTypeMessage += cfg.OpType
	}
	slog.Debug("Parsed configuration file", "path", cfg.ConfigFile, "operation", opTypeMessage)

	cmdType := properties.CommandTypeSDK // Default command type
	switch cfg.OpType {
	case "mktp":
		cmdType = properties.CommandTypeTerraform
	case "mksdk":
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

	slog.Info("Generation complete")
}
