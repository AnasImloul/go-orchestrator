package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/AnasImloul/go-orchestrator/internal/di"
	"github.com/AnasImloul/go-orchestrator/internal/lifecycle"
	"github.com/AnasImloul/go-orchestrator/internal/logger"
	"github.com/AnasImloul/go-orchestrator/pkg/orchestrator"
)

// BasicExample demonstrates the simplest usage of the orchestrator
func main() {
	// Create a simple logger
	slogLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	
	// Create logger adapter
	logger := logger.NewSlogAdapter(slogLogger)

	// Create orchestrator with default configuration
	config := orchestrator.DefaultOrchestratorConfig()
	orch, err := orchestrator.NewOrchestrator(config, logger)
	if err != nil {
		fmt.Printf("Failed to create orchestrator: %v\n", err)
		os.Exit(1)
	}

	// Register a simple feature
	feature := &SimpleFeature{name: "basic-service"}
	if err := orch.RegisterFeature(feature); err != nil {
		fmt.Printf("Failed to register feature: %v\n", err)
		os.Exit(1)
	}

	// Start the application
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := orch.Start(ctx); err != nil {
		fmt.Printf("Failed to start: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Application started successfully!")

	// Run for a short time
	time.Sleep(1 * time.Second)

	// Stop gracefully
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()

	if err := orch.Stop(stopCtx); err != nil {
		fmt.Printf("Failed to stop: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Application stopped successfully!")
}

// SimpleFeature is a basic feature implementation
type SimpleFeature struct {
	name string
}

func (f *SimpleFeature) GetName() string {
	return f.name
}

func (f *SimpleFeature) GetDependencies() []string {
	return []string{}
}

func (f *SimpleFeature) GetPriority() int {
	return 100
}

func (f *SimpleFeature) RegisterServices(container di.Container) error {
	fmt.Printf("Registering services for %s\n", f.name)
	return nil
}

func (f *SimpleFeature) CreateComponent(container di.Container) (lifecycle.Component, error) {
	return &SimpleComponent{name: f.name}, nil
}

func (f *SimpleFeature) GetRetryConfig() *lifecycle.RetryConfig {
	return nil
}

func (f *SimpleFeature) GetMetadata() orchestrator.FeatureMetadata {
	return orchestrator.FeatureMetadata{
		Name:        f.name,
		Version:     "1.0.0",
		Description: "A simple feature for basic example",
	}
}

// SimpleComponent is a basic component implementation
type SimpleComponent struct {
	name string
}

func (c *SimpleComponent) Name() string {
	return c.name
}

func (c *SimpleComponent) Dependencies() []string {
	return []string{}
}

func (c *SimpleComponent) Priority() int {
	return 100
}

func (c *SimpleComponent) Start(ctx context.Context) error {
	fmt.Printf("Starting %s\n", c.name)
	return nil
}

func (c *SimpleComponent) Stop(ctx context.Context) error {
	fmt.Printf("Stopping %s\n", c.name)
	return nil
}

func (c *SimpleComponent) Health(ctx context.Context) lifecycle.ComponentHealth {
	return lifecycle.ComponentHealth{
		Status:    lifecycle.HealthStatusHealthy,
		Message:   "Component is healthy",
		Timestamp: time.Now(),
	}
}

func (c *SimpleComponent) GetRetryConfig() *lifecycle.RetryConfig {
	return nil
}
