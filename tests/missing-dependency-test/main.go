package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/AnasImloul/go-orchestrator"
)

// DatabaseConfig represents database-specific configuration
type DatabaseConfig struct {
	URL string
}

// Database represents a database connection
type Database struct {
	config *DatabaseConfig
}

func NewDatabase(config *DatabaseConfig) *Database {
	return &Database{
		config: config,
	}
}

func main() {
	// Set up logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	// Create orchestrator
	registry := orchestrator.New()

	// Register database service that depends on DatabaseConfig
	// But DON'T register DatabaseConfig - this should fail
	registry.Register(
		orchestrator.NewAutoServiceFactory[*Database](
			func(config *DatabaseConfig) *Database {
				return NewDatabase(config)
			},
			orchestrator.Singleton,
		),
	)

	// Start the orchestrator - this should fail because DatabaseConfig is not registered
	ctx := context.Background()
	if err := registry.Start(ctx); err != nil {
		slog.Error("Failed to start orchestrator (expected)", "error", err)
		slog.Info("Test completed - DAG validation correctly caught missing dependency")
		return
	}

	// This should not be reached
	slog.Error("Unexpected success - DAG validation should have failed")
}
