package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/AnasImloul/go-orchestrator/pkg/di"
	"github.com/AnasImloul/go-orchestrator/pkg/lifecycle"
	"github.com/AnasImloul/go-orchestrator/pkg/logger"
	"github.com/AnasImloul/go-orchestrator/pkg/orchestrator"
)

// ExampleService represents a simple service
type ExampleService struct {
	name string
}

func (s *ExampleService) GetName() string {
	return s.name
}

// ExampleComponent represents a simple component
type ExampleComponent struct {
	name    string
	service *ExampleService
	logger  logger.Logger
}

func (c *ExampleComponent) Name() string {
	return c.name
}

func (c *ExampleComponent) Dependencies() []string {
	return []string{}
}

func (c *ExampleComponent) Priority() int {
	return 100
}

func (c *ExampleComponent) Start(ctx context.Context) error {
	c.logger.Info("Starting example component", "name", c.name)
	time.Sleep(100 * time.Millisecond) // Simulate startup time
	c.logger.Info("Example component started", "name", c.name)
	return nil
}

func (c *ExampleComponent) Stop(ctx context.Context) error {
	c.logger.Info("Stopping example component", "name", c.name)
	time.Sleep(50 * time.Millisecond) // Simulate shutdown time
	c.logger.Info("Example component stopped", "name", c.name)
	return nil
}

func (c *ExampleComponent) Health(ctx context.Context) lifecycle.ComponentHealth {
	return lifecycle.ComponentHealth{
		Status:    lifecycle.HealthStatusHealthy,
		Message:   "Component is healthy",
		Timestamp: time.Now(),
	}
}

func (c *ExampleComponent) GetRetryConfig() *lifecycle.RetryConfig {
	return &lifecycle.RetryConfig{
		MaxAttempts:       3,
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          1 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

// ExampleFeature represents a simple feature
type ExampleFeature struct {
	name string
}

func (f *ExampleFeature) GetName() string {
	return f.name
}

func (f *ExampleFeature) GetDependencies() []string {
	return []string{}
}

func (f *ExampleFeature) GetPriority() int {
	return 100
}

func (f *ExampleFeature) RegisterServices(container di.Container) error {
	// Register the example service
	serviceType := di.TypeOf[*ExampleService]()
	factory := func(ctx context.Context, c di.Container) (interface{}, error) {
		return &ExampleService{name: "example-service"}, nil
	}
	
	return container.RegisterSingleton(serviceType, factory)
}

func (f *ExampleFeature) CreateComponent(container di.Container) (lifecycle.Component, error) {
	// Resolve the service
	serviceType := di.TypeOf[*ExampleService]()
	service, err := container.Resolve(serviceType)
	if err != nil {
		return nil, err
	}
	
	// Create a logger adapter
	slogLogger := slog.Default()
	logger := logger.NewSlogAdapter(slogLogger)
	
	// Create the component
	component := &ExampleComponent{
		name:    f.name,
		service: service.(*ExampleService),
		logger:  logger,
	}
	
	return component, nil
}

func (f *ExampleFeature) GetRetryConfig() *lifecycle.RetryConfig {
	return &lifecycle.RetryConfig{
		MaxAttempts:       3,
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          1 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

func (f *ExampleFeature) GetMetadata() orchestrator.FeatureMetadata {
	return orchestrator.FeatureMetadata{
		Name:        f.name,
		Description: "An example feature demonstrating the orchestrator",
		Version:     "1.0.0",
		Author:      "Example Author",
		Tags:        []string{"example", "demo"},
	}
}

func main() {
	fmt.Println("Go Orchestrator External Usage Example")
	fmt.Println("=====================================")

	// Create a new orchestrator
	orch, err := orchestrator.NewOrchestrator()
	if err != nil {
		panic(fmt.Sprintf("Failed to create orchestrator: %v", err))
	}

	// Create and register a feature
	feature := &ExampleFeature{name: "example-feature"}
	if err := orch.RegisterFeature(feature); err != nil {
		panic(fmt.Sprintf("Failed to register feature: %v", err))
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start the orchestrator
	fmt.Println("Starting orchestrator...")
	if err := orch.Start(ctx); err != nil {
		panic(fmt.Sprintf("Failed to start orchestrator: %v", err))
	}

	// Perform a health check
	fmt.Println("Performing health check...")
	health := orch.HealthCheck(ctx)
	fmt.Printf("Health Status: %s\n", health.Status)
	fmt.Printf("Total Features: %d\n", health.Summary.TotalFeatures)
	fmt.Printf("Healthy Features: %d\n", health.Summary.HealthyFeatures)

	// Let it run for a bit
	fmt.Println("Running for 2 seconds...")
	time.Sleep(2 * time.Second)

	// Stop the orchestrator
	fmt.Println("Stopping orchestrator...")
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopCancel()

	if err := orch.Stop(stopCtx); err != nil {
		panic(fmt.Sprintf("Failed to stop orchestrator: %v", err))
	}

	fmt.Println("Orchestrator stopped successfully!")
}