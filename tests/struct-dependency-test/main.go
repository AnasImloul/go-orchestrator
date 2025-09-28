package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/AnasImloul/go-orchestrator"
)

// Database represents a database connection
type Database struct {
	connectionString string
}

func NewDatabase(connectionString string) *Database {
	return &Database{
		connectionString: connectionString,
	}
}

// Repository depends on Database
type Repository struct {
	db *Database
}

func NewRepository(db *Database) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) GetData() string {
	return fmt.Sprintf("Data from database: %s", r.db.connectionString)
}

func main() {
	// Set up logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	// Create orchestrator
	registry := orchestrator.New()

	// Register database service
	registry.Register(
		orchestrator.NewAutoServiceFactory[*Database](
			func() *Database {
				return NewDatabase("postgres://localhost:5432/mydb")
			},
			orchestrator.Singleton,
		),
	)

	// Register repository service that depends on database
	registry.Register(
		orchestrator.NewAutoServiceFactory[*Repository](
			func(db *Database) *Repository {
				return NewRepository(db)
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

	// Resolve and use the repository
	repository, err := orchestrator.ResolveStruct[*Repository](registry.Container())
	if err != nil {
		slog.Error("Failed to resolve repository", "error", err)
		os.Exit(1)
	}

	// Use the repository
	data := repository.GetData()
	slog.Info("Repository test successful", "data", data)

	// Stop the orchestrator
	if err := registry.Stop(ctx); err != nil {
		slog.Error("Failed to stop orchestrator", "error", err)
		os.Exit(1)
	}

	slog.Info("Test completed successfully")
}
