package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/AnasImloul/go-orchestrator/pkg/di"
	"github.com/AnasImloul/go-orchestrator/pkg/lifecycle"
	"github.com/AnasImloul/go-orchestrator/pkg/logger"
	"github.com/AnasImloul/go-orchestrator/pkg/orchestrator"
)

// This example shows how to use the Go Orchestrator library
// as an external dependency in your own project

func main() {
	fmt.Println("🚀 Go Orchestrator External Usage Example")
	fmt.Println("==========================================")

	// Create logger
	slogLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	logger := logger.NewSlogAdapter(slogLogger)

	// Create orchestrator with custom configuration
	config := orchestrator.OrchestratorConfig{
		StartupTimeout:      15 * time.Second,
		ShutdownTimeout:     10 * time.Second,
		HealthCheckInterval: 30 * time.Second,
		EnableMetrics:       true,
		EnableTracing:       false,
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
			"api": map[string]interface{}{
				"port": 8080,
			},
		},
	}

	orch, err := orchestrator.NewOrchestrator(config, logger)
	if err != nil {
		panic(fmt.Sprintf("Failed to create orchestrator: %v", err))
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
			panic(fmt.Sprintf("Failed to register feature %s: %v", feature.GetName(), err))
		}
	}

	// Start the application
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("\n📡 Starting application...")
	if err := orch.Start(ctx); err != nil {
		panic(fmt.Sprintf("Failed to start application: %v", err))
	}

	fmt.Println("✅ Application started successfully!")

	// Simulate running the application
	fmt.Println("\n⏳ Application running for 3 seconds...")
	time.Sleep(3 * time.Second)

	// Perform health check
	fmt.Println("\n🏥 Performing health check...")
	healthReport := orch.HealthCheck(ctx)
	fmt.Printf("📊 Health Status: %s\n", healthReport.Status)
	fmt.Printf("📈 Total Features: %d\n", healthReport.Summary.TotalFeatures)
	fmt.Printf("✅ Healthy Features: %d\n", healthReport.Summary.HealthyFeatures)

	// Graceful shutdown
	fmt.Println("\n🛑 Initiating graceful shutdown...")
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer stopCancel()

	if err := orch.Stop(stopCtx); err != nil {
		panic(fmt.Sprintf("Failed to stop application: %v", err))
	}

	fmt.Println("✅ Application stopped successfully!")
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
	fmt.Printf("🔧 Registering database services for %s\n", f.name)
	
	// Register database service
	return container.RegisterSingleton(
		di.TypeOf[DatabaseService](),
		func(ctx context.Context, c di.Container) (interface{}, error) {
			return &DatabaseServiceImpl{
				host:     "localhost",
				port:     5432,
				database: "myapp",
			}, nil
		},
	)
}

func (f *DatabaseFeature) CreateComponent(container di.Container) (lifecycle.Component, error) {
	return &DatabaseComponent{name: f.name}, nil
}

func (f *DatabaseFeature) GetRetryConfig() *lifecycle.RetryConfig {
	return nil // Use default retry config
}

func (f *DatabaseFeature) GetMetadata() orchestrator.FeatureMetadata {
	return orchestrator.FeatureMetadata{
		Name:        f.name,
		Version:     "1.0.0",
		Description: "PostgreSQL database service",
		Tags:        []string{"database", "postgresql", "persistence"},
	}
}

// DatabaseService interface
type DatabaseService interface {
	Connect() error
	Query(sql string) ([]map[string]interface{}, error)
	Close() error
}

// DatabaseServiceImpl implements DatabaseService
type DatabaseServiceImpl struct {
	host     string
	port     int
	database string
}

func (s *DatabaseServiceImpl) Connect() error {
	fmt.Printf("🔌 Connecting to database %s:%d/%s\n", s.host, s.port, s.database)
	time.Sleep(200 * time.Millisecond) // Simulate connection time
	fmt.Println("✅ Database connected successfully")
	return nil
}

func (s *DatabaseServiceImpl) Query(sql string) ([]map[string]interface{}, error) {
	fmt.Printf("📝 Executing query: %s\n", sql)
	return []map[string]interface{}{}, nil
}

func (s *DatabaseServiceImpl) Close() error {
	fmt.Println("🔌 Closing database connection")
	return nil
}

// DatabaseComponent implementation
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
	fmt.Printf("🚀 Starting %s component\n", c.name)
	time.Sleep(100 * time.Millisecond) // Simulate startup time
	fmt.Printf("✅ %s component started\n", c.name)
	return nil
}

func (c *DatabaseComponent) Stop(ctx context.Context) error {
	fmt.Printf("🛑 Stopping %s component\n", c.name)
	time.Sleep(50 * time.Millisecond) // Simulate shutdown time
	fmt.Printf("✅ %s component stopped\n", c.name)
	return nil
}

func (c *DatabaseComponent) Health(ctx context.Context) lifecycle.ComponentHealth {
	return lifecycle.ComponentHealth{
		Status:    lifecycle.HealthStatusHealthy,
		Message:   "Database is healthy and responsive",
		Timestamp: time.Now(),
	}
}

func (c *DatabaseComponent) GetRetryConfig() *lifecycle.RetryConfig {
	return nil
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
	fmt.Printf("🔧 Registering cache services for %s\n", f.name)
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
		Version:     "1.0.0",
		Description: "Redis cache service",
		Tags:        []string{"cache", "redis", "performance"},
	}
}

// CacheComponent implementation
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
	fmt.Printf("🚀 Starting %s component\n", c.name)
	time.Sleep(150 * time.Millisecond) // Simulate startup time
	fmt.Printf("✅ %s component started\n", c.name)
	return nil
}

func (c *CacheComponent) Stop(ctx context.Context) error {
	fmt.Printf("🛑 Stopping %s component\n", c.name)
	time.Sleep(75 * time.Millisecond) // Simulate shutdown time
	fmt.Printf("✅ %s component stopped\n", c.name)
	return nil
}

func (c *CacheComponent) Health(ctx context.Context) lifecycle.ComponentHealth {
	return lifecycle.ComponentHealth{
		Status:    lifecycle.HealthStatusHealthy,
		Message:   "Cache is healthy and responsive",
		Timestamp: time.Now(),
	}
}

func (c *CacheComponent) GetRetryConfig() *lifecycle.RetryConfig {
	return nil
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
	fmt.Printf("🔧 Registering API services for %s\n", f.name)
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
		Version:     "1.0.0",
		Description: "REST API server",
		Tags:        []string{"api", "http", "rest"},
	}
}

// APIComponent implementation
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
	fmt.Printf("🚀 Starting %s component\n", c.name)
	time.Sleep(200 * time.Millisecond) // Simulate startup time
	fmt.Printf("✅ %s component started on port 8080\n", c.name)
	return nil
}

func (c *APIComponent) Stop(ctx context.Context) error {
	fmt.Printf("🛑 Stopping %s component\n", c.name)
	time.Sleep(100 * time.Millisecond) // Simulate shutdown time
	fmt.Printf("✅ %s component stopped\n", c.name)
	return nil
}

func (c *APIComponent) Health(ctx context.Context) lifecycle.ComponentHealth {
	return lifecycle.ComponentHealth{
		Status:    lifecycle.HealthStatusHealthy,
		Message:   "API server is healthy and accepting requests",
		Timestamp: time.Now(),
	}
}

func (c *APIComponent) GetRetryConfig() *lifecycle.RetryConfig {
	return nil
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
	fmt.Printf("🔧 Registering worker services for %s\n", f.name)
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
		Version:     "1.0.0",
		Description: "Background job processor",
		Tags:        []string{"worker", "background", "jobs"},
	}
}

// WorkerComponent implementation
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
	fmt.Printf("🚀 Starting %s component\n", c.name)
	time.Sleep(100 * time.Millisecond) // Simulate startup time
	fmt.Printf("✅ %s component started\n", c.name)
	return nil
}

func (c *WorkerComponent) Stop(ctx context.Context) error {
	fmt.Printf("🛑 Stopping %s component\n", c.name)
	time.Sleep(50 * time.Millisecond) // Simulate shutdown time
	fmt.Printf("✅ %s component stopped\n", c.name)
	return nil
}

func (c *WorkerComponent) Health(ctx context.Context) lifecycle.ComponentHealth {
	return lifecycle.ComponentHealth{
		Status:    lifecycle.HealthStatusHealthy,
		Message:   "Background worker is healthy and processing jobs",
		Timestamp: time.Now(),
	}
}

func (c *WorkerComponent) GetRetryConfig() *lifecycle.RetryConfig {
	return nil
}
