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
	port int
	db   DatabaseService
	cache CacheService
}

func (a *apiService) Start() error {
	fmt.Printf("Starting API server on port %d\n", a.port)
	time.Sleep(100 * time.Millisecond)
	fmt.Println("API server started successfully")
	return nil
}

func (a *apiService) Stop() error {
	fmt.Println("Stopping API server")
	time.Sleep(50 * time.Millisecond)
	fmt.Println("API server stopped")
	return nil
}

func (a *apiService) Health() string {
	return "healthy"
}


func main() {
	fmt.Println("Go Orchestrator - Clean API Example")
	fmt.Println("===================================")

	// Create application with default configuration
	app := orchestrator.New()

	// Add database feature
	app.AddFeature(
		orchestrator.NewFeatureWithInstance("database", DatabaseService(&databaseService{host: "localhost", port: 5432}), orchestrator.Singleton),
			orchestrator.NewFeature("database"),
		).
			WithLifetime(orchestrator.Singleton).
			WithComponent(
				orchestrator.NewComponent().
					WithStart(orchestrator.WithStartFunc[DatabaseService](func(db DatabaseService) error {
						return db.Connect()
					})).
					WithStop(orchestrator.WithStopFuncWithApp[DatabaseService](app, func(db DatabaseService) error {
						return db.Disconnect()
					})).
					WithHealth(func(ctx context.Context) orchestrator.HealthStatus {
						return orchestrator.HealthStatus{
							Status:  "healthy",
							Message: "Database is connected",
						}
					}),
			),
	)

	// Add cache feature with Transient lifetime
	app.AddFeature(
		orchestrator.NewFeatureWithInstance("cache", CacheService(&cacheService{
			host: "localhost", 
			port: 6379,
		})(
			orchestrator.NewFeature("cache"),
		).
			WithLifetime(orchestrator.Transient).
			WithComponent(
				orchestrator.NewComponent().
					WithStart(orchestrator.WithStartFunc[CacheService](func(cache CacheService) error {
						return cache.Connect()
					})).
					WithStop(orchestrator.WithStopFuncWithApp[CacheService](app, func(cache CacheService) error {
						return cache.Disconnect()
					})).
					WithHealth(func(ctx context.Context) orchestrator.HealthStatus {
						return orchestrator.HealthStatus{
							Status:  "healthy",
							Message: "Cache is connected",
						}
					}),
			),
	)

	// Add API feature that depends on database and cache
	app.AddFeature(
		orchestrator.WithServiceFactory[APIService](
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
		)(
			orchestrator.NewFeature("api").
				WithDependencies("database", "cache"),
		).
			WithLifetime(orchestrator.Singleton).
			WithComponent(
				orchestrator.NewComponent().
					WithStart(orchestrator.WithStartFunc[APIService](func(api APIService) error {
						return api.Start()
					})).
					WithStop(orchestrator.WithStopFuncWithApp[APIService](app, func(api APIService) error {
						return api.Stop()
					})).
					WithHealth(orchestrator.WithHealthFunc[APIService](app, func(api APIService) orchestrator.HealthStatus {
						return orchestrator.HealthStatus{
							Status:  api.Health(),
							Message: "API server is running",
						}
					})),
			),
	)

	// Start the application
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("Starting application...")
	if err := app.Start(ctx); err != nil {
		panic(fmt.Errorf("Failed to start application: %w", err))
	}
	fmt.Println("Application started successfully!")

	// Check health
	fmt.Println("Checking application health...")
	healthReport := app.Health(ctx)
	for name, status := range healthReport {
		fmt.Printf("  %s: %s - %s\n", name, status.Status, status.Message)
	}

	// Demonstrate service lifetimes
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
