package main

import (
	"context"
	"fmt"
	"log"

	"github.com/AnasImloul/go-orchestrator"
)

// DatabaseService interface
type DatabaseService interface {
	Connect() error
	Disconnect() error
	GetInstanceID() string
}

// CacheService interface
type CacheService interface {
	Connect() error
	Disconnect() error
	GetInstanceID() string
}

// APIService interface
type APIService interface {
	Start() error
	Stop() error
	GetInstanceID() string
}

// databaseService implementation
type databaseService struct {
	host string
	port int
	id   string
}

func (d *databaseService) Connect() error {
	fmt.Printf("🔌 Database connecting to %s:%d (ID: %s)\n", d.host, d.port, d.id)
	return nil
}

func (d *databaseService) Disconnect() error {
	fmt.Printf("🔌 Database disconnecting from %s:%d (ID: %s)\n", d.host, d.port, d.id)
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

func (c *cacheService) Connect() error {
	fmt.Printf("💾 Cache connecting to %s:%d (ID: %s)\n", c.host, c.port, c.id)
	return nil
}

func (c *cacheService) Disconnect() error {
	fmt.Printf("💾 Cache disconnecting from %s:%d (ID: %s)\n", c.host, c.port, c.id)
	return nil
}

func (c *cacheService) GetInstanceID() string {
	return c.id
}

// apiService implementation
type apiService struct {
	port  int
	db    DatabaseService
	cache CacheService
	id    string
}

func (a *apiService) Start() error {
	fmt.Printf("🚀 API starting on port %d (ID: %s)\n", a.port, a.id)
	fmt.Printf("   - Database ID: %s\n", a.db.GetInstanceID())
	fmt.Printf("   - Cache ID: %s\n", a.cache.GetInstanceID())
	return nil
}

func (a *apiService) Stop() error {
	fmt.Printf("🛑 API stopping on port %d (ID: %s)\n", a.port, a.id)
	return nil
}

func (a *apiService) GetInstanceID() string {
	return a.id
}

func main() {
	fmt.Println("🎯 Automatic Dependency Discovery Example")
	fmt.Println("==========================================")

	// Create application
	app := orchestrator.New()

	// Add database feature with instance
	app.AddFeature(
		orchestrator.NewFeatureWithInstance("database", 
			DatabaseService(&databaseService{host: "localhost", port: 5432, id: "db-001"}), 
			orchestrator.Singleton,
		),
	)

	// Add cache feature with instance
	app.AddFeature(
		orchestrator.NewFeatureWithInstance("cache", 
			CacheService(&cacheService{host: "localhost", port: 6379, id: "cache-001"}), 
			orchestrator.Singleton,
		),
	)

	// Add API feature with automatic dependency discovery
	// The factory function only takes the dependencies as parameters
	// Dependencies are automatically discovered from the function parameters!
	app.AddFeature(
		orchestrator.NewFeatureWithAutoFactory[APIService]("api",
			func(db DatabaseService, cache CacheService) APIService {
				return &apiService{
					port:  8080,
					db:    db,
					cache: cache,
					id:    "api-001",
				}
			},
			orchestrator.Singleton,
		).
			WithComponent(
				orchestrator.NewComponent().
					WithStart(func(ctx context.Context, container *orchestrator.Container) error {
						api, err := orchestrator.ResolveType[APIService](container)
						if err != nil {
							return err
						}
						return api.Start()
					}).
					WithStop(func(ctx context.Context) error {
						api, err := orchestrator.ResolveType[APIService](app.Container())
						if err != nil {
							return err
						}
						return api.Stop()
					}).
					WithHealth(func(ctx context.Context) orchestrator.HealthStatus {
						return orchestrator.HealthStatus{
							Status:  "healthy",
							Message: "API is running",
						}
					}),
			),
	)

	// Start the application
	ctx := context.Background()
	fmt.Println("\n🚀 Starting application...")
	if err := app.Start(ctx); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	// Show that dependencies were automatically resolved
	fmt.Println("\n✅ Application started successfully!")
	fmt.Println("   - Dependencies were automatically discovered and injected")
	fmt.Println("   - No manual dependency resolution needed in the factory")

	// Stop the application
	fmt.Println("\n🛑 Stopping application...")
	if err := app.Stop(ctx); err != nil {
		log.Fatalf("Failed to stop application: %v", err)
	}

	fmt.Println("✅ Application stopped successfully!")
}
