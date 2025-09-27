package main

import (
	"context"
	"fmt"
	"time"

	"github.com/AnasImloul/go-orchestrator"
)

// CacheService interface
type CacheService interface {
	Connect() error
	Disconnect() error
}

// cacheService implementation
type cacheService struct {
	host string
	port int
}

func (c *cacheService) Connect() error {
	fmt.Printf("Connecting to cache at %s:%d\n", c.host, c.port)
	time.Sleep(200 * time.Millisecond) // Simulate connection time
	fmt.Println("Cache connected successfully")
	return nil
}

func (c *cacheService) Disconnect() error {
	fmt.Println("Disconnecting from cache")
	time.Sleep(100 * time.Millisecond)
	fmt.Println("Cache disconnected")
	return nil
}

// MetricsService interface
type MetricsService interface {
	Start() error
	Stop() error
}

// metricsService implementation
type metricsService struct {
	port int
}

func (m *metricsService) Start() error {
	fmt.Printf("Starting metrics service on port %d\n", m.port)
	time.Sleep(150 * time.Millisecond) // Simulate startup time
	fmt.Println("Metrics service started successfully")
	return nil
}

func (m *metricsService) Stop() error {
	fmt.Println("Stopping metrics service")
	time.Sleep(50 * time.Millisecond)
	fmt.Println("Metrics service stopped")
	return nil
}

// LoggingService interface
type LoggingService interface {
	Initialize() error
	Shutdown() error
}

// loggingService implementation
type loggingService struct {
	level string
}

func (l *loggingService) Initialize() error {
	fmt.Printf("Initializing logging service with level %s\n", l.level)
	time.Sleep(100 * time.Millisecond) // Simulate initialization time
	fmt.Println("Logging service initialized successfully")
	return nil
}

func (l *loggingService) Shutdown() error {
	fmt.Println("Shutting down logging service")
	time.Sleep(50 * time.Millisecond)
	fmt.Println("Logging service shut down")
	return nil
}

// APIService interface
type APIService interface {
	Start() error
	Stop() error
}

// apiService implementation
type apiService struct {
	port    int
	cache   CacheService
	metrics MetricsService
	logging LoggingService
}

func (a *apiService) Start() error {
	fmt.Printf("Starting API service on port %d\n", a.port)
	time.Sleep(100 * time.Millisecond)
	fmt.Println("API service started successfully")
	return nil
}

func (a *apiService) Stop() error {
	fmt.Println("Stopping API service")
	time.Sleep(50 * time.Millisecond)
	fmt.Println("API service stopped")
	return nil
}

func main() {
	fmt.Println("Go Orchestrator - Parallel Execution Example")
	fmt.Println("=============================================")

	// Create application with default configuration
	app := orchestrator.New()

	// Add independent services (these will start in parallel at level 0)
	app.AddFeature(
		orchestrator.WithService[CacheService](&cacheService{host: "localhost", port: 6379})(
			orchestrator.NewFeature("cache"),
		).
			WithLifetime(orchestrator.Singleton).
			WithComponent(
				orchestrator.NewComponent().
					WithStart(orchestrator.WithStartFunc[CacheService](func(cache CacheService) error {
						return cache.Connect()
					})).
					WithStop(orchestrator.WithStopFuncWithApp[CacheService](app, func(cache CacheService) error {
						return cache.Disconnect()
					})),
			),
	)

	app.AddFeature(
		orchestrator.WithService[MetricsService](&metricsService{port: 9090})(
			orchestrator.NewFeature("metrics"),
		).
			WithLifetime(orchestrator.Singleton).
			WithComponent(
				orchestrator.NewComponent().
					WithStart(orchestrator.WithStartFunc[MetricsService](func(metrics MetricsService) error {
						return metrics.Start()
					})).
					WithStop(orchestrator.WithStopFuncWithApp[MetricsService](app, func(metrics MetricsService) error {
						return metrics.Stop()
					})),
			),
	)

	app.AddFeature(
		orchestrator.WithService[LoggingService](&loggingService{level: "info"})(
			orchestrator.NewFeature("logging"),
		).
			WithLifetime(orchestrator.Singleton).
			WithComponent(
				orchestrator.NewComponent().
					WithStart(orchestrator.WithStartFunc[LoggingService](func(logging LoggingService) error {
						return logging.Initialize()
					})).
					WithStop(orchestrator.WithStopFuncWithApp[LoggingService](app, func(logging LoggingService) error {
						return logging.Shutdown()
					})),
			),
	)

	// Add API service that depends on all three (this will start at level 1)
	app.AddFeature(
		orchestrator.WithServiceFactory[APIService](
			func(ctx context.Context, container *orchestrator.Container) (APIService, error) {
				cache, err := orchestrator.ResolveType[CacheService](container)
				if err != nil {
					return nil, err
				}
				metrics, err := orchestrator.ResolveType[MetricsService](container)
				if err != nil {
					return nil, err
				}
				logging, err := orchestrator.ResolveType[LoggingService](container)
				if err != nil {
					return nil, err
				}
				return &apiService{port: 8080, cache: cache, metrics: metrics, logging: logging}, nil
			},
		)(
			orchestrator.NewFeature("api").
				WithDependencies("cache", "metrics", "logging"),
		).
			WithLifetime(orchestrator.Singleton).
			WithComponent(
				orchestrator.NewComponent().
					WithStart(orchestrator.WithStartFunc[APIService](func(api APIService) error {
						return api.Start()
					})).
					WithStop(orchestrator.WithStopFuncWithApp[APIService](app, func(api APIService) error {
						return api.Stop()
					})),
			),
	)

	// Start the application
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("Starting application...")
	startTime := time.Now()
	if err := app.Start(ctx); err != nil {
		panic(fmt.Errorf("Failed to start application: %w", err))
	}
	startupDuration := time.Since(startTime)
	fmt.Printf("Application started successfully in %v!\n", startupDuration)

	// Check health
	fmt.Println("Checking application health...")
	healthReport := app.Health(ctx)
	for name, status := range healthReport {
		fmt.Printf("  %s: %s - %s\n", name, status.Status, status.Message)
	}

	// Run for a bit
	fmt.Println("Running for 2 seconds...")
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
