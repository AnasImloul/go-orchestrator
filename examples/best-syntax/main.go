package main

import (
	"context"
	"fmt"
	"time"

	"github.com/AnasImloul/go-orchestrator"
)

// DatabaseService interface
type DatabaseService interface {
	Connect() error
	Disconnect() error
	GetConnectionID() string
}

// databaseService implementation
type databaseService struct {
	host         string
	port         int
	connectionID string
}

func (d *databaseService) Connect() error {
	d.connectionID = fmt.Sprintf("conn_%d", time.Now().UnixNano())
	fmt.Printf("Connecting to database at %s:%d (ID: %s)\n", d.host, d.port, d.connectionID)
	time.Sleep(100 * time.Millisecond)
	fmt.Println("Database connected successfully")
	return nil
}

func (d *databaseService) Disconnect() error {
	fmt.Println("Disconnecting from database")
	time.Sleep(50 * time.Millisecond)
	fmt.Println("Database disconnected")
	return nil
}

func (d *databaseService) GetConnectionID() string {
	return d.connectionID
}

// CacheService interface
type CacheService interface {
	Connect() error
	Disconnect() error
	GetInstanceID() string
}

// cacheService implementation
type cacheService struct {
	host       string
	port       int
	instanceID string
}

func (c *cacheService) Connect() error {
	c.instanceID = fmt.Sprintf("cache_%d", time.Now().UnixNano())
	fmt.Printf("Connecting to cache at %s:%d (ID: %s)\n", c.host, c.port, c.instanceID)
	time.Sleep(100 * time.Millisecond)
	fmt.Println("Cache connected successfully")
	return nil
}

func (c *cacheService) Disconnect() error {
	fmt.Println("Disconnecting from cache")
	time.Sleep(50 * time.Millisecond)
	fmt.Println("Cache disconnected")
	return nil
}

func (c *cacheService) GetInstanceID() string {
	if c.instanceID == "" {
		c.instanceID = fmt.Sprintf("cache_%d", time.Now().UnixNano())
	}
	return c.instanceID
}

// APIService interface
type APIService interface {
	Start() error
	Stop() error
	Health() string
}

// apiService implementation
type apiService struct {
	port  int
	db    DatabaseService
	cache CacheService
}

func (a *apiService) Start() error {
	fmt.Printf("Starting API server on port %d\n", a.port)
	time.Sleep(150 * time.Millisecond)
	fmt.Println("API server started successfully")
	return nil
}

func (a *apiService) Stop() error {
	fmt.Println("Stopping API server")
	time.Sleep(75 * time.Millisecond)
	fmt.Println("API server stopped")
	return nil
}

func (a *apiService) Health() string {
	return "healthy"
}

func main() {
	fmt.Println("Go Orchestrator - Best Syntax Example")
	fmt.Println("====================================")

	// Create application with default configuration
	app := orchestrator.New()

	// Method 1: Ultra-clean syntax using NewFeatureWithInstance (factory-based)
	app.AddFeature(
		orchestrator.WithComponentFor[DatabaseService](
			orchestrator.NewFeatureWithInstance("database", DatabaseService(&databaseService{host: "localhost", port: 5432}), orchestrator.Singleton),
			app,
		).
			WithStartFor(func(db DatabaseService) error { return db.Connect() }).
			WithStopFor(func(db DatabaseService) error { return db.Disconnect() }).
			Build(),
	)

	// Method 2: Clean syntax using NewFeatureWithInstance with Transient lifetime (factory-based)
	app.AddFeature(
		orchestrator.WithComponentFor[CacheService](
			orchestrator.NewFeatureWithInstance("cache", CacheService(&cacheService{host: "localhost", port: 6379}), orchestrator.Transient),
			app,
		).
			WithStartFor(func(cache CacheService) error { return cache.Connect() }).
			WithStopFor(func(cache CacheService) error { return cache.Disconnect() }).
			Build(),
	)

	// Method 3: Factory-based service with dependencies using NewFeatureWithFactory
	app.AddFeature(
		orchestrator.WithComponentFor[APIService](
			orchestrator.NewFeatureWithFactory("api", 
				func(ctx context.Context, container *orchestrator.Container) (APIService, error) {
					db, err := orchestrator.ResolveType[DatabaseService](container)
					if err != nil {
						return nil, err
					}
					cache, err := orchestrator.ResolveType[CacheService](container)
					if err != nil {
						return nil, err
					}
					return &apiService{port: 8080, db: db, cache: cache}, nil
				}, 
				orchestrator.Singleton,
			).WithDependencies("database", "cache"),
			app,
		).
			WithStartFor(func(api APIService) error { return api.Start() }).
			WithStopFor(func(api APIService) error { return api.Stop() }).
			WithHealthFor(func(api APIService) orchestrator.HealthStatus {
				return orchestrator.HealthStatus{
					Status:  api.Health(),
					Message: "API server is running",
				}
			}).
			Build(),
	)

	// Start the application
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("Starting application...")
	if err := app.Start(ctx); err != nil {
		panic(fmt.Errorf("Failed to start application: %w", err))
	}
	fmt.Println("Application started successfully!")

	// Demonstrate different lifetimes by resolving services multiple times
	fmt.Println("\nDemonstrating service lifetimes:")
	fmt.Println("================================")

	container := app.Container()

	// Resolve services multiple times to show lifetime behavior
	for i := 0; i < 3; i++ {
		fmt.Printf("\n--- Resolution #%d ---\n", i+1)

		// Singleton: Same instance every time
		db1, _ := orchestrator.ResolveType[DatabaseService](container)
		db2, _ := orchestrator.ResolveType[DatabaseService](container)
		fmt.Printf("Database (Singleton): %s == %s? %t\n", 
			db1.GetConnectionID(), db2.GetConnectionID(), 
			db1.GetConnectionID() == db2.GetConnectionID())

		// Transient: New instance every time
		cache1, _ := orchestrator.ResolveType[CacheService](container)
		cache2, _ := orchestrator.ResolveType[CacheService](container)
		fmt.Printf("Cache (Transient): %s == %s? %t\n", 
			cache1.GetInstanceID(), cache2.GetInstanceID(), 
			cache1.GetInstanceID() == cache2.GetInstanceID())

		time.Sleep(100 * time.Millisecond)
	}

	// Check health
	fmt.Println("\nChecking application health...")
	health := app.Health(ctx)
	for name, status := range health {
		fmt.Printf("  %s: %s - %s\n", name, status.Status, status.Message)
	}

	// Run for a bit
	fmt.Println("\nRunning for 2 seconds...")
	time.Sleep(2 * time.Second)

	// Stop the application
	fmt.Println("Stopping application...")
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer stopCancel()

	if err := app.Stop(stopCtx); err != nil {
		panic(fmt.Errorf("Failed to stop application: %w", err))
	}
	fmt.Println("Application stopped successfully!")
}
