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
	return nil
}

func (d *databaseService) GetConnectionID() string {
	return d.connectionID
}

// CacheService interface
type CacheService interface {
	Connect() error
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
	return nil
}

func (c *cacheService) GetInstanceID() string {
	return c.instanceID
}

// LoggerService interface
type LoggerService interface {
	Log(message string)
	GetInstanceID() string
	Connect() error
}

// loggerService implementation
type loggerService struct {
	instanceID string
}

func (l *loggerService) Log(message string) {
	fmt.Printf("[%s] %s\n", l.instanceID, message)
}

func (l *loggerService) GetInstanceID() string {
	return l.instanceID
}

func (l *loggerService) Connect() error {
	l.instanceID = fmt.Sprintf("logger_%d", time.Now().UnixNano())
	fmt.Printf("Initializing logger (ID: %s)\n", l.instanceID)
	return nil
}

func main() {
	fmt.Println("Go Orchestrator - Service Lifetimes Example")
	fmt.Println("============================================")

	// Create application with default configuration
	app := orchestrator.New()

	// Add services with different lifetimes
	app.AddFeature(
		orchestrator.WithServiceInstanceGeneric[DatabaseService](
			&databaseService{host: "localhost", port: 5432},
			orchestrator.Singleton, // Single instance for entire application
		)(orchestrator.NewFeature("database")).
			WithComponent(
				orchestrator.NewComponent().
					WithStart(func(ctx context.Context, container *orchestrator.Container) error {
						db, err := orchestrator.ResolveType[DatabaseService](container)
						if err != nil {
							return err
						}
						return db.Connect()
					}),
			),
	)

	app.AddFeature(
		orchestrator.WithServiceInstanceGeneric[CacheService](
			&cacheService{host: "localhost", port: 6379},
			orchestrator.Scoped, // One instance per scope (request/operation)
		)(orchestrator.NewFeature("cache")).
			WithComponent(
				orchestrator.NewComponent().
					WithStart(func(ctx context.Context, container *orchestrator.Container) error {
						cache, err := orchestrator.ResolveType[CacheService](container)
						if err != nil {
							return err
						}
						return cache.Connect()
					}),
			),
	)

	app.AddFeature(
		orchestrator.WithServiceInstanceGeneric[LoggerService](
			&loggerService{},
			orchestrator.Transient, // New instance every time
		)(orchestrator.NewFeature("logger")).
			WithComponent(
				orchestrator.NewComponent().
					WithStart(func(ctx context.Context, container *orchestrator.Container) error {
						logger, err := orchestrator.ResolveType[LoggerService](container)
						if err != nil {
							return err
						}
						return logger.Connect()
					}),
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

		// Scoped: Same instance within scope (for now, behaves like singleton)
		cache1, _ := orchestrator.ResolveType[CacheService](container)
		cache2, _ := orchestrator.ResolveType[CacheService](container)
		fmt.Printf("Cache (Scoped): %s == %s? %t\n", 
			cache1.GetInstanceID(), cache2.GetInstanceID(), 
			cache1.GetInstanceID() == cache2.GetInstanceID())

		// Transient: New instance every time
		logger1, _ := orchestrator.ResolveType[LoggerService](container)
		logger2, _ := orchestrator.ResolveType[LoggerService](container)
		fmt.Printf("Logger (Transient): %s == %s? %t\n", 
			logger1.GetInstanceID(), logger2.GetInstanceID(), 
			logger1.GetInstanceID() == logger2.GetInstanceID())

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
