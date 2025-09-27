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

// ExampleFeature demonstrates how to create a feature
type ExampleFeature struct {
	name string
}

func NewExampleFeature(name string) *ExampleFeature {
	return &ExampleFeature{name: name}
}

func (f *ExampleFeature) GetName() string {
	return f.name
}

func (f *ExampleFeature) GetDependencies() []string {
	return []string{} // No dependencies
}

func (f *ExampleFeature) GetPriority() int {
	return 100 // Lower numbers start first
}

func (f *ExampleFeature) RegisterServices(container di.Container) error {
	// Register any services this feature needs
	fmt.Printf("Registering services for feature: %s\n", f.name)
	return nil
}

func (f *ExampleFeature) CreateComponent(container di.Container) (lifecycle.Component, error) {
	// Create the actual component
	fmt.Printf("Creating component for feature: %s\n", f.name)
	return &ExampleComponent{name: f.name}, nil
}

func (f *ExampleFeature) GetRetryConfig() *lifecycle.RetryConfig {
	return nil // Use default retry config
}

func (f *ExampleFeature) GetMetadata() orchestrator.FeatureMetadata {
	return orchestrator.FeatureMetadata{
		Name:        f.name,
		Version:     "1.0.0",
		Description: "Example feature for demonstration",
		Author:      "Go Orchestrator",
		Tags:        []string{"example", "demo"},
	}
}

// ExampleComponent implements the lifecycle component
type ExampleComponent struct {
	name string
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
	fmt.Printf("Starting component: %s\n", c.name)
	// Simulate some startup work
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("Component %s started successfully\n", c.name)
	return nil
}

func (c *ExampleComponent) Stop(ctx context.Context) error {
	fmt.Printf("Stopping component: %s\n", c.name)
	// Simulate some shutdown work
	time.Sleep(50 * time.Millisecond)
	fmt.Printf("Component %s stopped successfully\n", c.name)
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
	return nil // Use default retry config
}

func main() {
	// Create a logger
	slogLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create logger adapter (not used in simple example)
	_ = logger.NewSlogAdapter(slogLogger)

	// Create orchestrator configuration
	config := orchestrator.DefaultOrchestratorConfig()
	config.StartupTimeout = 10 * time.Second
	config.ShutdownTimeout = 5 * time.Second

	// Create orchestrator
	orch, err := orchestrator.NewOrchestrator()
	if err != nil {
		fmt.Printf("Failed to create orchestrator: %v\n", err)
		os.Exit(1)
	}

	// Register features
	features := []*ExampleFeature{
		NewExampleFeature("database"),
		NewExampleFeature("cache"),
		NewExampleFeature("api-server"),
	}

	for _, feature := range features {
		if err := orch.RegisterFeature(feature); err != nil {
			fmt.Printf("Failed to register feature %s: %v\n", feature.GetName(), err)
			os.Exit(1)
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start the application
	fmt.Println("Starting application...")
	if err := orch.Start(ctx); err != nil {
		fmt.Printf("Failed to start application: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Application started successfully!")

	// Simulate running for a while
	time.Sleep(2 * time.Second)

	// Perform health check
	fmt.Println("Performing health check...")
	healthReport := orch.HealthCheck(ctx)
	fmt.Printf("Health report: %+v\n", healthReport)

	// Stop the application
	fmt.Println("Stopping application...")
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopCancel()

	if err := orch.Stop(stopCtx); err != nil {
		fmt.Printf("Failed to stop application: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Application stopped successfully!")
}
