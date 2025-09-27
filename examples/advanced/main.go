package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/AnasImloul/go-orchestrator/internal/di"
	"github.com/AnasImloul/go-orchestrator/internal/lifecycle"
	"github.com/AnasImloul/go-orchestrator/internal/logger"
	"github.com/AnasImloul/go-orchestrator/pkg/orchestrator"
)

// AdvancedExample demonstrates complex orchestration with dependencies
func main() {
	// Create a structured logger
	slogLogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	
	// Create logger adapter
	logger := logger.NewSlogAdapter(slogLogger)

	// Create orchestrator with custom configuration
	config := orchestrator.OrchestratorConfig{
		StartupTimeout:      15 * time.Second,
		ShutdownTimeout:     10 * time.Second,
		HealthCheckInterval: 30 * time.Second,
		EnableMetrics:       true,
		EnableTracing:       true,
		FeatureConfig: map[string]interface{}{
			"database": map[string]interface{}{
				"host":     "localhost",
				"port":     5432,
				"database": "myapp",
			},
			"cache": map[string]interface{}{
				"host": "localhost",
				"port": 6379,
			},
		},
	}

	orch, err := orchestrator.NewOrchestrator(config, logger)
	if err != nil {
		fmt.Printf("Failed to create orchestrator: %v\n", err)
		os.Exit(1)
	}

	// Register features with dependencies
	features := []orchestrator.Feature{
		&DatabaseFeature{name: "database", priority: 10},
		&CacheFeature{name: "cache", priority: 20},
		&APIFeature{name: "api-server", priority: 30},
		&WorkerFeature{name: "background-worker", priority: 40},
	}

	for _, feature := range features {
		if err := orch.RegisterFeature(feature); err != nil {
			fmt.Printf("Failed to register feature %s: %v\n", feature.GetName(), err)
			os.Exit(1)
		}
	}

	// Start the application
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.Info("Starting application with advanced orchestration")
	if err := orch.Start(ctx); err != nil {
		fmt.Printf("Failed to start: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Application started successfully")

	// Simulate running the application
	time.Sleep(3 * time.Second)

	// Perform health checks
	logger.Info("Performing health check")
	healthReport := orch.HealthCheck(ctx)
	logger.Info("Health check completed", "report", healthReport)

	// Stop gracefully
	logger.Info("Initiating graceful shutdown")
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer stopCancel()

	if err := orch.Stop(stopCtx); err != nil {
		fmt.Printf("Failed to stop: %v\n", err)
		os.Exit(1)
	}

	logger.Info("Application stopped successfully")
}

// DatabaseFeature represents a database service
type DatabaseFeature struct {
	name     string
	priority int
}

func (f *DatabaseFeature) GetName() string {
	return f.name
}

func (f *DatabaseFeature) GetDependencies() []string {
	return []string{} // No dependencies
}

func (f *DatabaseFeature) GetPriority() int {
	return f.priority
}

func (f *DatabaseFeature) RegisterServices(container di.Container) error {
	fmt.Printf("Registering database services for %s\n", f.name)
	return nil
}

func (f *DatabaseFeature) CreateComponent(container di.Container) (lifecycle.Component, error) {
	return &DatabaseComponent{name: f.name}, nil
}

func (f *DatabaseFeature) GetRetryConfig() *lifecycle.RetryConfig {
	return nil
}

func (f *DatabaseFeature) GetMetadata() orchestrator.FeatureMetadata {
	return orchestrator.FeatureMetadata{
		Name:        f.name,
		Version:     "2.1.0",
		Description: "PostgreSQL database service",
		Tags:        []string{"database", "postgresql", "persistence"},
	}
}

// CacheFeature represents a cache service
type CacheFeature struct {
	name     string
	priority int
}

func (f *CacheFeature) GetName() string {
	return f.name
}

func (f *CacheFeature) GetDependencies() []string {
	return []string{"database"} // Depends on database
}

func (f *CacheFeature) GetPriority() int {
	return f.priority
}

func (f *CacheFeature) RegisterServices(container di.Container) error {
	fmt.Printf("Registering cache services for %s\n", f.name)
	return nil
}

func (f *CacheFeature) CreateComponent(container di.Container) (lifecycle.Component, error) {
	return &CacheComponent{name: f.name}, nil
}

func (f *CacheFeature) GetRetryConfig() *lifecycle.RetryConfig {
	return nil
}

func (f *CacheFeature) GetMetadata() orchestrator.FeatureMetadata {
	return orchestrator.FeatureMetadata{
		Name:        f.name,
		Version:     "1.5.0",
		Description: "Redis cache service",
		Tags:        []string{"cache", "redis", "performance"},
	}
}

// APIFeature represents an API server
type APIFeature struct {
	name     string
	priority int
}

func (f *APIFeature) GetName() string {
	return f.name
}

func (f *APIFeature) GetDependencies() []string {
	return []string{"database", "cache"} // Depends on both database and cache
}

func (f *APIFeature) GetPriority() int {
	return f.priority
}

func (f *APIFeature) RegisterServices(container di.Container) error {
	fmt.Printf("Registering API services for %s\n", f.name)
	return nil
}

func (f *APIFeature) CreateComponent(container di.Container) (lifecycle.Component, error) {
	return &APIComponent{name: f.name}, nil
}

func (f *APIFeature) GetRetryConfig() *lifecycle.RetryConfig {
	return nil
}

func (f *APIFeature) GetMetadata() orchestrator.FeatureMetadata {
	return orchestrator.FeatureMetadata{
		Name:        f.name,
		Version:     "3.0.0",
		Description: "REST API server",
		Tags:        []string{"api", "http", "rest"},
	}
}

// WorkerFeature represents a background worker
type WorkerFeature struct {
	name     string
	priority int
}

func (f *WorkerFeature) GetName() string {
	return f.name
}

func (f *WorkerFeature) GetDependencies() []string {
	return []string{"database", "cache", "api-server"} // Depends on all other services
}

func (f *WorkerFeature) GetPriority() int {
	return f.priority
}

func (f *WorkerFeature) RegisterServices(container di.Container) error {
	fmt.Printf("Registering worker services for %s\n", f.name)
	return nil
}

func (f *WorkerFeature) CreateComponent(container di.Container) (lifecycle.Component, error) {
	return &WorkerComponent{name: f.name}, nil
}

func (f *WorkerFeature) GetRetryConfig() *lifecycle.RetryConfig {
	return nil
}

func (f *WorkerFeature) GetMetadata() orchestrator.FeatureMetadata {
	return orchestrator.FeatureMetadata{
		Name:        f.name,
		Version:     "1.2.0",
		Description: "Background job processor",
		Tags:        []string{"worker", "background", "jobs"},
	}
}

// Component implementations
type DatabaseComponent struct {
	name string
}

func (c *DatabaseComponent) Name() string {
	return c.name
}

func (c *DatabaseComponent) Dependencies() []string {
	return []string{}
}

func (c *DatabaseComponent) Priority() int {
	return 10
}

func (c *DatabaseComponent) Start(ctx context.Context) error {
	fmt.Printf("Starting %s (connecting to PostgreSQL...)\n", c.name)
	time.Sleep(500 * time.Millisecond) // Simulate connection time
	fmt.Printf("%s started successfully\n", c.name)
	return nil
}

func (c *DatabaseComponent) Stop(ctx context.Context) error {
	fmt.Printf("Stopping %s (closing database connections...)\n", c.name)
	time.Sleep(200 * time.Millisecond) // Simulate cleanup time
	fmt.Printf("%s stopped successfully\n", c.name)
	return nil
}

func (c *DatabaseComponent) Health(ctx context.Context) lifecycle.ComponentHealth {
	return lifecycle.ComponentHealth{
		Status:    lifecycle.HealthStatusHealthy,
		Message:   "Database is healthy",
		Timestamp: time.Now(),
	}
}

func (c *DatabaseComponent) GetRetryConfig() *lifecycle.RetryConfig {
	return nil
}

type CacheComponent struct {
	name string
}

func (c *CacheComponent) Name() string {
	return c.name
}

func (c *CacheComponent) Dependencies() []string {
	return []string{"database"}
}

func (c *CacheComponent) Priority() int {
	return 20
}

func (c *CacheComponent) Start(ctx context.Context) error {
	fmt.Printf("Starting %s (connecting to Redis...)\n", c.name)
	time.Sleep(300 * time.Millisecond) // Simulate connection time
	fmt.Printf("%s started successfully\n", c.name)
	return nil
}

func (c *CacheComponent) Stop(ctx context.Context) error {
	fmt.Printf("Stopping %s (closing Redis connections...)\n", c.name)
	time.Sleep(150 * time.Millisecond) // Simulate cleanup time
	fmt.Printf("%s stopped successfully\n", c.name)
	return nil
}

func (c *CacheComponent) Health(ctx context.Context) lifecycle.ComponentHealth {
	return lifecycle.ComponentHealth{
		Status:    lifecycle.HealthStatusHealthy,
		Message:   "Cache is healthy",
		Timestamp: time.Now(),
	}
}

func (c *CacheComponent) GetRetryConfig() *lifecycle.RetryConfig {
	return nil
}

type APIComponent struct {
	name string
}

func (c *APIComponent) Name() string {
	return c.name
}

func (c *APIComponent) Dependencies() []string {
	return []string{"database", "cache"}
}

func (c *APIComponent) Priority() int {
	return 30
}

func (c *APIComponent) Start(ctx context.Context) error {
	fmt.Printf("Starting %s (binding to port 8080...)\n", c.name)
	time.Sleep(400 * time.Millisecond) // Simulate startup time
	fmt.Printf("%s started successfully\n", c.name)
	return nil
}

func (c *APIComponent) Stop(ctx context.Context) error {
	fmt.Printf("Stopping %s (shutting down HTTP server...)\n", c.name)
	time.Sleep(300 * time.Millisecond) // Simulate shutdown time
	fmt.Printf("%s stopped successfully\n", c.name)
	return nil
}

func (c *APIComponent) Health(ctx context.Context) lifecycle.ComponentHealth {
	return lifecycle.ComponentHealth{
		Status:    lifecycle.HealthStatusHealthy,
		Message:   "API server is healthy",
		Timestamp: time.Now(),
	}
}

func (c *APIComponent) GetRetryConfig() *lifecycle.RetryConfig {
	return nil
}

type WorkerComponent struct {
	name string
}

func (c *WorkerComponent) Name() string {
	return c.name
}

func (c *WorkerComponent) Dependencies() []string {
	return []string{"database", "cache", "api-server"}
}

func (c *WorkerComponent) Priority() int {
	return 40
}

func (c *WorkerComponent) Start(ctx context.Context) error {
	fmt.Printf("Starting %s (initializing job queue...)\n", c.name)
	time.Sleep(200 * time.Millisecond) // Simulate startup time
	fmt.Printf("%s started successfully\n", c.name)
	return nil
}

func (c *WorkerComponent) Stop(ctx context.Context) error {
	fmt.Printf("Stopping %s (draining job queue...)\n", c.name)
	time.Sleep(250 * time.Millisecond) // Simulate shutdown time
	fmt.Printf("%s stopped successfully\n", c.name)
	return nil
}

func (c *WorkerComponent) Health(ctx context.Context) lifecycle.ComponentHealth {
	return lifecycle.ComponentHealth{
		Status:    lifecycle.HealthStatusHealthy,
		Message:   "Worker is healthy",
		Timestamp: time.Now(),
	}
}

func (c *WorkerComponent) GetRetryConfig() *lifecycle.RetryConfig {
	return nil
}
