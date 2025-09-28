package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/AnasImloul/go-orchestrator"
)

// DatabaseService interface - now implements the Service interface
type DatabaseService interface {
	orchestrator.Service
	Connect() error
	Disconnect() error
	GetInstanceID() string
}

// CacheService interface - now implements the Service interface
type CacheService interface {
	orchestrator.Service
	Connect() error
	Disconnect() error
	GetInstanceID() string
}

// APIService interface - no Service interface needed! Health is automatic
type APIService interface {
	GetInstanceID() string
}

// databaseService implementation
type databaseService struct {
	host string
	port int
	id   string
}

func (d *databaseService) Start(ctx context.Context) error {
	fmt.Printf("Database starting connection to %s:%d (ID: %s)\n", d.host, d.port, d.id)
	time.Sleep(100 * time.Millisecond) // Simulate connection time
	return d.Connect()
}

func (d *databaseService) Stop(ctx context.Context) error {
	fmt.Printf("Database stopping connection to %s:%d (ID: %s)\n", d.host, d.port, d.id)
	time.Sleep(50 * time.Millisecond) // Simulate disconnection time
	return d.Disconnect()
}

func (d *databaseService) Health(ctx context.Context) orchestrator.HealthStatus {
	return orchestrator.HealthStatus{
		Status:  orchestrator.HealthStatusDegraded,
		Message: fmt.Sprintf("Database %s is connected but experiencing slow queries", d.id),
	}
}

func (d *databaseService) Connect() error {
	fmt.Printf("   Database connected to %s:%d\n", d.host, d.port)
	return nil
}

func (d *databaseService) Disconnect() error {
	fmt.Printf("   Database disconnected from %s:%d\n", d.host, d.port)
	return nil
}

func (d *databaseService) GetInstanceID() string {
	return d.id
}

// cacheService implementation
type cacheService struct {
	host string
	port int
	id   string
}

func (c *cacheService) Start(ctx context.Context) error {
	fmt.Printf("Cache starting connection to %s:%d (ID: %s)\n", c.host, c.port, c.id)
	time.Sleep(75 * time.Millisecond) // Simulate connection time
	return c.Connect()
}

func (c *cacheService) Stop(ctx context.Context) error {
	fmt.Printf("Cache stopping connection to %s:%d (ID: %s)\n", c.host, c.port, c.id)
	time.Sleep(25 * time.Millisecond) // Simulate disconnection time
	return c.Disconnect()
}

func (c *cacheService) Health(ctx context.Context) orchestrator.HealthStatus {
	return orchestrator.HealthStatus{
		Status:  orchestrator.HealthStatusHealthy,
		Message: fmt.Sprintf("Cache %s is connected", c.id),
	}
}

func (c *cacheService) Connect() error {
	fmt.Printf("   Cache connected to %s:%d\n", c.host, c.port)
	return nil
}

func (c *cacheService) Disconnect() error {
	fmt.Printf("   Cache disconnected from %s:%d\n", c.host, c.port)
	return nil
}

func (c *cacheService) GetInstanceID() string {
	return c.id
}

// apiService implementation - no Health method needed! Dependencies are auto-detected
type apiService struct {
	port  int
	db    DatabaseService
	cache CacheService
	id    string
}

func (a *apiService) Start(ctx context.Context) error {
	fmt.Printf("API starting on port %d (ID: %s)\n", a.port, a.id)
	fmt.Printf("   - Database ID: %s\n", a.db.GetInstanceID())
	fmt.Printf("   - Cache ID: %s\n", a.cache.GetInstanceID())
	time.Sleep(150 * time.Millisecond) // Simulate startup time
	fmt.Printf("   API server started successfully\n")
	return nil
}

func (a *apiService) Stop(ctx context.Context) error {
	fmt.Printf("API stopping on port %d (ID: %s)\n", a.port, a.id)
	time.Sleep(75 * time.Millisecond) // Simulate shutdown time
	fmt.Printf("   API server stopped successfully\n")
	return nil
}

// No Health method needed! The library automatically detects dependencies
// and provides health aggregation based on the factory parameters

func (a *apiService) GetInstanceID() string {
	return a.id
}

func main() {
	fmt.Println("Improved Developer Experience Example")
	fmt.Println("========================================")
	fmt.Println("No more boilerplate lifecycle configuration!")
	fmt.Println("Just implement the Service interface and you're done!")
	fmt.Println("Completely automatic dependency health propagation!")
	fmt.Println("   - Database reports degraded health")
	fmt.Println("   - API service automatically reflects degraded status")
	fmt.Println("   - Dependencies are auto-detected from factory parameters")
	fmt.Println("   - No BaseService embedding needed - it's all automatic!")
	fmt.Println()

	// Create service registry
	registry := orchestrator.New()

	// Register services with ZERO boilerplate!
	// Just pass the instance - lifecycle is automatically wired!
	registry.Register(
		orchestrator.NewServiceSingleton[DatabaseService](
			&databaseService{host: "localhost", port: 5432, id: "db-001"},
		),
	)

	registry.Register(
		orchestrator.NewServiceSingleton[CacheService](
			&cacheService{host: "localhost", port: 6379, id: "cache-001"},
		),
	)

	// Factory-based service with automatic dependency discovery AND lifecycle wiring
	// Dependencies are automatically detected from the factory parameters!
	// No Service interface needed - health aggregation is completely automatic!
	registry.Register(
		orchestrator.NewAutoServiceFactory[APIService](
			func(db DatabaseService, cache CacheService) APIService {
				return &apiService{
					port:  8080,
					db:    db,
					cache: cache,
					id:    "api-001",
				}
			},
			orchestrator.Singleton,
		),
	)

	// Start the service registry
	ctx := context.Background()
	fmt.Println("Starting service registry...")
	if err := registry.Start(ctx); err != nil {
		log.Fatalf("Failed to start service registry: %v", err)
	}

	fmt.Println("\nService registry started successfully!")
	fmt.Println("   - All lifecycle methods were automatically wired")
	fmt.Println("   - Dependencies were automatically discovered and injected")
	fmt.Println("   - No manual configuration needed!")

	// Check health
	fmt.Println("\nChecking service health...")
	health := registry.Health(ctx)
	for name, status := range health {
		fmt.Printf("   %s: %s - %s\n", name, status.Status, status.Message)
		if status.Details != nil {
			for key, value := range status.Details {
				fmt.Printf("      %s: %v\n", key, value)
			}
		}
	}

	// Run for a bit
	fmt.Println("\nRunning for 2 seconds...")
	time.Sleep(2 * time.Second)

	// Stop the service registry
	fmt.Println("\nStopping service registry...")
	if err := registry.Stop(ctx); err != nil {
		log.Fatalf("Failed to stop service registry: %v", err)
	}

	fmt.Println("Service registry stopped successfully!")
	fmt.Println("\nThat's it! No boilerplate, just clean, simple service registration!")
}
