package main

import (
	"context"
	"fmt"
	"time"

	"github.com/AnasImloul/go-orchestrator"
)

// SimpleService demonstrates a service that uses BaseService for default handlers
type SimpleService interface {
	orchestrator.Service
	GetName() string
}

type simpleService struct {
	orchestrator.BaseService
	name string
}

func NewSimpleService(name string) SimpleService {
	return &simpleService{
		BaseService: orchestrator.BaseService{
			Dependencies: []string{}, // No dependencies
		},
		name: name,
	}
}

func (s *simpleService) GetName() string {
	return s.name
}

// ServiceWithDependencies demonstrates a service that depends on other services
type ServiceWithDependencies interface {
	orchestrator.Service
	GetDependencyCount() int
}

type serviceWithDependencies struct {
	orchestrator.BaseService
	dependencyCount int
}

func NewServiceWithDependencies(deps []string) ServiceWithDependencies {
	return &serviceWithDependencies{
		BaseService: orchestrator.BaseService{
			Dependencies: deps, // Specify dependencies for health checking
		},
		dependencyCount: len(deps),
	}
}

func (s *serviceWithDependencies) GetDependencyCount() int {
	return s.dependencyCount
}

// CustomService demonstrates a service that overrides some default handlers
type CustomService interface {
	orchestrator.Service
	GetStartupMessage() string
}

type customService struct {
	orchestrator.BaseService
	startupMessage string
}

func NewCustomService() CustomService {
	return &customService{
		BaseService: orchestrator.BaseService{
			Dependencies: []string{}, // No dependencies
		},
		startupMessage: "Custom service started with custom logic!",
	}
}

// Override the default Start method
func (c *customService) Start(ctx context.Context) error {
	fmt.Printf("🚀 Custom startup logic: %s\n", c.startupMessage)
	time.Sleep(100 * time.Millisecond) // Simulate startup work
	return nil
}

// Override the default Health method
func (c *customService) Health(ctx context.Context) orchestrator.HealthStatus {
	return orchestrator.HealthStatus{
		Status:  orchestrator.HealthStatusHealthy,
		Message: "Custom health check: Service is running perfectly!",
		Details: map[string]interface{}{
			"custom_field": "custom_value",
			"startup_time": time.Now().Format(time.RFC3339),
		},
	}
}

func (c *customService) GetStartupMessage() string {
	return c.startupMessage
}

// UnhealthyService demonstrates a service that reports unhealthy status
type UnhealthyService interface {
	orchestrator.Service
	GetError() string
}

type unhealthyService struct {
	orchestrator.BaseService
	error string
}

func NewUnhealthyService() UnhealthyService {
	return &unhealthyService{
		BaseService: orchestrator.BaseService{
			Dependencies: []string{}, // No dependencies
		},
		error: "Simulated service failure",
	}
}

// Override the default Health method to report unhealthy status
func (u *unhealthyService) Health(ctx context.Context) orchestrator.HealthStatus {
	return orchestrator.HealthStatus{
		Status:  orchestrator.HealthStatusUnhealthy,
		Message: fmt.Sprintf("Service is unhealthy: %s", u.error),
		Details: map[string]interface{}{
			"error_code": "SIMULATED_FAILURE",
			"timestamp":  time.Now().Format(time.RFC3339),
		},
	}
}

func (u *unhealthyService) GetError() string {
	return u.error
}

func main() {
	fmt.Println("🎯 Default Handlers Example")
	fmt.Println("===========================")
	fmt.Println("✨ Demonstrates BaseService with default Start, Stop, and Health handlers")
	fmt.Println("✨ Shows dependency health aggregation")
	fmt.Println("✨ Shows how to override default handlers when needed")
	fmt.Println()

	// Create service registry
	registry := orchestrator.New()

	// Register services with different configurations
	registry.Register(
		orchestrator.NewServiceSingleton[SimpleService](
			NewSimpleService("simple-1"),
		).WithName("simple-1"),
	)

	registry.Register(
		orchestrator.NewServiceSingleton[SimpleService](
			NewSimpleService("simple-2"),
		).WithName("simple-2"),
	)

	registry.Register(
		orchestrator.NewServiceSingleton[CustomService](
			NewCustomService(),
		),
	)

	registry.Register(
		orchestrator.NewServiceSingleton[UnhealthyService](
			NewUnhealthyService(),
		),
	)

	// Register a service that depends on other services
	registry.Register(
		orchestrator.NewServiceSingleton[ServiceWithDependencies](
			NewServiceWithDependencies([]string{
				"simple-1", // Reference to the first simple service
				"simple-2", // Reference to the second simple service
				"main::Custom", // This will be the service name for CustomService
			}),
		),
	)

	// Start the service registry
	ctx := context.Background()
	fmt.Println("🚀 Starting service registry...")
	if err := registry.Start(ctx); err != nil {
		fmt.Printf("❌ Failed to start service registry: %v\n", err)
		return
	}

	fmt.Println("\n✅ Service registry started successfully!")
	fmt.Println("   - Services with default handlers started automatically")
	fmt.Println("   - Custom startup logic executed where overridden")
	fmt.Println("   - Registry references set for dependency health checking")

	// Check health
	fmt.Println("\n📊 Checking service health...")
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
	fmt.Println("\n⏱️  Running for 3 seconds...")
	time.Sleep(3 * time.Second)

	// Stop the service registry
	fmt.Println("\n🛑 Stopping service registry...")
	if err := registry.Stop(ctx); err != nil {
		fmt.Printf("❌ Failed to stop service registry: %v\n", err)
		return
	}

	fmt.Println("✅ Service registry stopped successfully!")
	fmt.Println("\n🎉 Default handlers working perfectly!")
	fmt.Println("   - No-op Start/Stop handlers for services that don't need them")
	fmt.Println("   - Automatic dependency health aggregation")
	fmt.Println("   - Easy to override handlers when custom logic is needed")
}
