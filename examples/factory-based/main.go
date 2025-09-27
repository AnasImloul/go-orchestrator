package main

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/AnasImloul/go-orchestrator"
)

// DatabaseService interface
type DatabaseService interface {
	Connect() error
	GetConnectionID() string
	Disconnect() error
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

func (d *databaseService) GetConnectionID() string {
	return d.connectionID
}

func (d *databaseService) Disconnect() error {
	fmt.Println("Disconnecting from database")
	time.Sleep(50 * time.Millisecond)
	fmt.Println("Database disconnected")
	return nil
}

// CacheService interface
type CacheService interface {
	Connect() error
	GetInstanceID() string
	Disconnect() error
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

func (c *cacheService) GetInstanceID() string {
	return c.instanceID
}

func (c *cacheService) Disconnect() error {
	fmt.Println("Disconnecting from cache")
	time.Sleep(50 * time.Millisecond)
	fmt.Println("Cache disconnected")
	return nil
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
	level      string
}

func (l *loggerService) Log(message string) {
	fmt.Printf("[%s] %s: %s\n", l.level, l.instanceID, message)
}

func (l *loggerService) GetInstanceID() string {
	return l.instanceID
}

func (l *loggerService) Connect() error {
	l.instanceID = fmt.Sprintf("logger_%d", time.Now().UnixNano())
	fmt.Printf("Initializing logger (ID: %s, Level: %s)\n", l.instanceID, l.level)
	return nil
}

func main() {
	fmt.Println("Go Orchestrator - Factory-Based Registration Example")
	fmt.Println("====================================================")

	// Create application with default configuration
	app := orchestrator.New()

	// Add database feature using factory registration
	app.AddFeature(
		orchestrator.NewFeature("database").
			WithService(
				reflect.TypeOf((*DatabaseService)(nil)).Elem(),
				func(ctx context.Context, container *orchestrator.Container) (interface{}, error) {
					// Factory creates a new instance each time (for Transient)
					return &databaseService{host: "localhost", port: 5432}, nil
				},
				orchestrator.Singleton, // Database should be singleton
			).
			WithComponent(
				orchestrator.NewComponent().
					WithStart(func(ctx context.Context, container *orchestrator.Container) error {
						db, err := orchestrator.ResolveType[DatabaseService](container)
						if err != nil {
							return err
						}
						return db.Connect()
					}).
					WithStop(func(ctx context.Context) error {
						db, err := orchestrator.ResolveType[DatabaseService](app.Container())
						if err != nil {
							return err
						}
						return db.Disconnect()
					}),
			),
	)

	// Add cache feature using factory registration
	app.AddFeature(
		orchestrator.NewFeature("cache").
			WithService(
				reflect.TypeOf((*CacheService)(nil)).Elem(),
				func(ctx context.Context, container *orchestrator.Container) (interface{}, error) {
					// Factory creates a new instance each time (for Transient)
					return &cacheService{host: "localhost", port: 6379}, nil
				},
				orchestrator.Transient, // Cache should be transient
			).
			WithComponent(
				orchestrator.NewComponent().
					WithStart(func(ctx context.Context, container *orchestrator.Container) error {
						cache, err := orchestrator.ResolveType[CacheService](container)
						if err != nil {
							return err
						}
						return cache.Connect()
					}).
					WithStop(func(ctx context.Context) error {
						cache, err := orchestrator.ResolveType[CacheService](app.Container())
						if err != nil {
							return err
						}
						return cache.Disconnect()
					}),
			),
	)

	// Add multiple logger services using named registration
	app.AddFeature(
		orchestrator.NewFeature("loggers").
			WithNamedService(
				"info-logger",
				reflect.TypeOf((*LoggerService)(nil)).Elem(),
				func(ctx context.Context, container *orchestrator.Container) (interface{}, error) {
					return &loggerService{level: "INFO"}, nil
				},
				orchestrator.Transient,
			).
			WithNamedService(
				"error-logger",
				reflect.TypeOf((*LoggerService)(nil)).Elem(),
				func(ctx context.Context, container *orchestrator.Container) (interface{}, error) {
					return &loggerService{level: "ERROR"}, nil
				},
				orchestrator.Transient,
			).
			WithComponent(
				orchestrator.NewComponent().
					WithStart(func(ctx context.Context, container *orchestrator.Container) error {
						// Initialize both loggers
						infoLogger, err := orchestrator.ResolveNamedType[LoggerService](container, "info-logger")
						if err != nil {
							return err
						}
						infoLogger.Connect()

						errorLogger, err := orchestrator.ResolveNamedType[LoggerService](container, "error-logger")
						if err != nil {
							return err
						}
						errorLogger.Connect()

						// Test logging
						infoLogger.Log("Application started")
						errorLogger.Log("No errors during startup")

						return nil
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

		// Transient: New instance every time
		cache1, _ := orchestrator.ResolveType[CacheService](container)
		cache2, _ := orchestrator.ResolveType[CacheService](container)
		fmt.Printf("Cache (Transient): %s == %s? %t\n", 
			cache1.GetInstanceID(), cache2.GetInstanceID(), 
			cache1.GetInstanceID() == cache2.GetInstanceID())

		// Named services: Different instances
		infoLogger1, _ := orchestrator.ResolveNamedType[LoggerService](container, "info-logger")
		infoLogger2, _ := orchestrator.ResolveNamedType[LoggerService](container, "info-logger")
		fmt.Printf("Info Logger (Transient): %s == %s? %t\n", 
			infoLogger1.GetInstanceID(), infoLogger2.GetInstanceID(), 
			infoLogger1.GetInstanceID() == infoLogger2.GetInstanceID())

		errorLogger1, _ := orchestrator.ResolveNamedType[LoggerService](container, "error-logger")
		errorLogger2, _ := orchestrator.ResolveNamedType[LoggerService](container, "error-logger")
		fmt.Printf("Error Logger (Transient): %s == %s? %t\n", 
			errorLogger1.GetInstanceID(), errorLogger2.GetInstanceID(), 
			errorLogger1.GetInstanceID() == errorLogger2.GetInstanceID())

		time.Sleep(100 * time.Millisecond)
	}

	// Test scoped services
	fmt.Println("\nTesting scoped services:")
	fmt.Println("=======================")
	
	// Create a scope
	scopedContainer := container.CreateScope()
	defer scopedContainer.Dispose()

	// Resolve services within the scope
	scopedCache1, _ := orchestrator.ResolveType[CacheService](scopedContainer)
	scopedCache2, _ := orchestrator.ResolveType[CacheService](scopedContainer)
	fmt.Printf("Scoped Cache: %s == %s? %t\n", 
		scopedCache1.GetInstanceID(), scopedCache2.GetInstanceID(), 
		scopedCache1.GetInstanceID() == scopedCache2.GetInstanceID())

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
