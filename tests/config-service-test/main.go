package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/AnasImloul/go-orchestrator"
)

// Config represents the main application configuration
type Config struct {
	DatabaseURL string
	RedisURL    string
}

// DatabaseConfig represents database-specific configuration
type DatabaseConfig struct {
	URL string
}

// Database represents a database connection
type Database struct {
	config *Config
}

func NewDatabase(config *Config) *Database {
	return &Database{
		config: config,
	}
}

func (d *Database) GetURL() string {
	return d.config.DatabaseURL
}

func main() {
	// Set up logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	// Create orchestrator
	registry := orchestrator.New()

	// Register main config service
	registry.Register(
		orchestrator.NewAutoServiceFactory[*Config](
			func() *Config {
				return &Config{
					DatabaseURL: "postgres://localhost:5432/mydb",
					RedisURL:    "redis://localhost:6379",
				}
			},
			orchestrator.Singleton,
		),
	)

	// Register database service that depends on main config
	registry.Register(
		orchestrator.NewAutoServiceFactory[*Database](
			func(config *Config) *Database {
				return NewDatabase(config)
			},
			orchestrator.Singleton,
		),
	)

	// Start the orchestrator
	ctx := context.Background()
	if err := registry.Start(ctx); err != nil {
		slog.Error("Failed to start orchestrator", "error", err)
		os.Exit(1)
	}

	// Resolve and use the database
	database, err := orchestrator.ResolveStruct[*Database](registry.Container())
	if err != nil {
		slog.Error("Failed to resolve database", "error", err)
		os.Exit(1)
	}

	// Use the database
	url := database.GetURL()
	slog.Info("Database test successful", "url", url)

	// Stop the orchestrator
	if err := registry.Stop(ctx); err != nil {
		slog.Error("Failed to stop orchestrator", "error", err)
		os.Exit(1)
	}

	slog.Info("Test completed successfully")
}
