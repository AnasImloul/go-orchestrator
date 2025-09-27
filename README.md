# Go Orchestrator

A powerful Go library for application orchestration, dependency injection, and lifecycle management.

## Features

- **Dependency Injection Container**: Type-safe DI with support for singletons, scoped, and transient lifetimes
- **Lifecycle Management**: Automatic startup/shutdown ordering based on dependencies and priorities
- **Application Orchestration**: Coordinate multiple features with proper dependency resolution
- **Health Checking**: Built-in health monitoring for all components
- **Generic Helpers**: Type-safe resolution with `di.Resolve[T]()` syntax
- **DAG-based Dependencies**: Directed Acyclic Graph for proper dependency ordering
- **Extensible**: Plugin architecture for custom features and components

## Installation

```bash
go get github.com/imloulanas/go-orchestrator
```

## Quick Start

```go
package main

import (
    "context"
    "log/slog"
    
    "github.com/imloulanas/go-orchestrator/di"
    "github.com/imloulanas/go-orchestrator/lifecycle"
    "github.com/imloulanas/go-orchestrator/orchestrator"
)

func main() {
    // Create logger
    logger := slog.Default()
    
    // Create orchestrator
    config := orchestrator.DefaultOrchestratorConfig()
    orch, err := orchestrator.NewOrchestrator(config, logger)
    if err != nil {
        panic(err)
    }
    
    // Register features
    orch.RegisterFeature(&MyFeature{})
    
    // Start application
    ctx := context.Background()
    if err := orch.Start(ctx); err != nil {
        panic(err)
    }
    
    // Application is now running...
    
    // Graceful shutdown
    orch.Stop(ctx)
}
```

## Dependency Injection

### Basic Usage

```go
// Register a service
container.RegisterSingleton(di.TypeOf[MyService](), func(ctx context.Context, c di.Container) (interface{}, error) {
    return &MyService{}, nil
})

// Resolve a service
service, err := di.Resolve[MyService](container)
if err != nil {
    panic(err)
}
```

### Generic Helpers

```go
// Type-safe resolution
service := di.MustResolve[MyService](container)

// Try resolution (returns false if not found)
if service, ok := di.TryResolve[MyService](container); ok {
    // Use service
}

// Get type information
serviceType := di.TypeOf[MyService]()
```

## Lifecycle Management

### Creating a Component

```go
type MyComponent struct {
    name string
}

func (c *MyComponent) Name() string {
    return c.name
}

func (c *MyComponent) Dependencies() []string {
    return []string{"database", "cache"}
}

func (c *MyComponent) Priority() int {
    return 100
}

func (c *MyComponent) Start(ctx context.Context) error {
    // Initialize component
    return nil
}

func (c *MyComponent) Stop(ctx context.Context) error {
    // Cleanup component
    return nil
}

func (c *MyComponent) Health(ctx context.Context) lifecycle.ComponentHealth {
    return lifecycle.ComponentHealth{
        Status:    lifecycle.HealthStatusHealthy,
        Message:   "Component is healthy",
        Timestamp: time.Now(),
    }
}
```

### Registering Components

```go
manager := lifecycle.NewLifecycleManager(logger)
manager.RegisterComponent(&MyComponent{name: "my-service"})

// Start all components in dependency order
ctx := context.Background()
if err := manager.Start(ctx); err != nil {
    panic(err)
}

// Stop all components in reverse order
manager.Stop(ctx)
```

## Application Orchestration

### Creating a Feature

```go
type MyFeature struct{}

func (f *MyFeature) GetName() string {
    return "my-feature"
}

func (f *MyFeature) GetDependencies() []string {
    return []string{"database"}
}

func (f *MyFeature) GetPriority() int {
    return 200
}

func (f *MyFeature) RegisterServices(container di.Container) error {
    // Register services for this feature
    return container.RegisterSingleton(di.TypeOf[MyService](), func(ctx context.Context, c di.Container) (interface{}, error) {
        return &MyService{}, nil
    })
}

func (f *MyFeature) CreateComponent(container di.Container) (lifecycle.Component, error) {
    return &MyComponent{}, nil
}

func (f *MyFeature) GetMetadata() orchestrator.FeatureMetadata {
    return orchestrator.FeatureMetadata{
        Name:        "my-feature",
        Description: "My awesome feature",
        Version:     "1.0.0",
    }
}
```

## Configuration

### Container Configuration

```go
config := di.ContainerConfig{
    EnableValidation:    true,
    EnableCircularCheck: true,
    EnableInterception:  false,
    DefaultLifetime:     di.Singleton,
    MaxResolutionDepth:  50,
    EnableMetrics:       true,
}

container := di.NewContainer(config, logger)
```

### Orchestrator Configuration

```go
config := orchestrator.OrchestratorConfig{
    StartupTimeout:       30 * time.Second,
    ShutdownTimeout:      15 * time.Second,
    HealthCheckInterval:  30 * time.Second,
    EnableMetrics:        true,
    EnableTracing:        false,
}

orch, err := orchestrator.NewOrchestrator(config, logger)
```

## Health Checking

```go
// Perform health check
report := orch.HealthCheck(ctx)

fmt.Printf("Overall Status: %s\n", report.Status)
fmt.Printf("Healthy Features: %d\n", report.Summary.HealthyFeatures)
fmt.Printf("Unhealthy Features: %d\n", report.Summary.UnhealthyFeatures)

// Check specific feature health
if featureHealth, exists := report.Features["my-feature"]; exists {
    fmt.Printf("Feature Status: %s\n", featureHealth.Status)
}
```

## Advanced Features

### Service Interceptors

```go
interceptor := di.InterceptorFunc(func(ctx context.Context, serviceType reflect.Type, next func() (interface{}, error)) (interface{}, error) {
    // Pre-processing
    fmt.Printf("Creating service: %s\n", serviceType.String())
    
    // Call next in chain
    instance, err := next()
    if err != nil {
        return nil, err
    }
    
    // Post-processing
    fmt.Printf("Service created: %T\n", instance)
    return instance, nil
})

container.Register(di.TypeOf[MyService](), factory, di.WithInterceptors(interceptor))
```

### Lifecycle Hooks

```go
hook := func(ctx context.Context, event lifecycle.Event) error {
    fmt.Printf("Lifecycle event: %s for component %s\n", event.Phase, event.Component)
    return nil
}

manager.AddHook(lifecycle.PhaseStartup, hook)
manager.AddHook(lifecycle.PhaseShutdown, hook)
```

## License

MIT License - see LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
