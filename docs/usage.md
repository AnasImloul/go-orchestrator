# Using Go Orchestrator as External Dependency

This guide shows how to use the Go Orchestrator library in your own projects.

## Installation

### Using Go Modules (Recommended)

```bash
go get github.com/AnasImloul/go-orchestrator
```

### Using Specific Version

```bash
# Get latest version
go get github.com/AnasImloul/go-orchestrator@latest

# Get specific version
go get github.com/AnasImloul/go-orchestrator@v1.0.0

# Get specific commit
go get github.com/AnasImloul/go-orchestrator@abc1234
```

> **Note**: The public API is production-ready and fully functional. All examples in this documentation are working and can be used in production applications.

## Basic Usage

### 1. Import the Library

```go
package main

import (
    "context"
    "log/slog"
    "os"
    
    "github.com/AnasImloul/go-orchestrator/pkg/orchestrator"
)
```

### 2. Create a Simple Application

```go
func main() {
    // Create logger
    logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
    
    // Create orchestrator with default configuration
    config := orchestrator.DefaultOrchestratorConfig()
    orch, err := orchestrator.NewOrchestrator(config, logger)
    if err != nil {
        panic(err)
    }
    
    // Register your features
    orch.RegisterFeature(&MyFeature{})
    
    // Start the application
    ctx := context.Background()
    if err := orch.Start(ctx); err != nil {
        panic(err)
    }
    
    // Your application is now running...
    
    // Graceful shutdown
    defer orch.Stop(ctx)
}
```

## Creating Custom Features

### 1. Implement the Feature Interface

```go
type MyFeature struct {
    name string
}

func (f *MyFeature) GetName() string {
    return f.name
}

func (f *MyFeature) GetDependencies() []string {
    return []string{} // No dependencies
}

func (f *MyFeature) GetPriority() int {
    return 100 // Lower numbers start first
}

func (f *MyFeature) RegisterServices(container di.Container) error {
    // Register your services here
    return container.RegisterSingleton(
        di.TypeOf[MyService](),
        func(ctx context.Context, c di.Container) (interface{}, error) {
            return &MyService{}, nil
        },
    )
}

func (f *MyFeature) CreateComponent(container di.Container) (lifecycle.Component, error) {
    return &MyComponent{name: f.name}, nil
}

func (f *MyFeature) GetRetryConfig() *lifecycle.RetryConfig {
    return nil // Use default retry config
}

func (f *MyFeature) GetMetadata() orchestrator.FeatureMetadata {
    return orchestrator.FeatureMetadata{
        Name:        f.name,
        Version:     "1.0.0",
        Description: "My custom feature",
    }
}
```

### 2. Implement the Component Interface

```go
type MyComponent struct {
    name string
}

func (c *MyComponent) Name() string {
    return c.name
}

func (c *MyComponent) Dependencies() []string {
    return []string{}
}

func (c *MyComponent) Priority() int {
    return 100
}

func (c *MyComponent) Start(ctx context.Context) error {
    // Initialize your component
    return nil
}

func (c *MyComponent) Stop(ctx context.Context) error {
    // Cleanup your component
    return nil
}

func (c *MyComponent) Health(ctx context.Context) lifecycle.ComponentHealth {
    return lifecycle.ComponentHealth{
        Status:    lifecycle.HealthStatusHealthy,
        Message:   "Component is healthy",
        Timestamp: time.Now(),
    }
}

func (c *MyComponent) GetRetryConfig() *lifecycle.RetryConfig {
    return nil
}
```

## Advanced Usage

### 1. Using Dependency Injection

```go
// Define your service interface
type DatabaseService interface {
    Connect() error
    Query(sql string) ([]map[string]interface{}, error)
    Close() error
}

// Implement your service
type PostgreSQLService struct {
    connectionString string
}

func (s *PostgreSQLService) Connect() error {
    // Connect to PostgreSQL
    return nil
}

func (s *PostgreSQLService) Query(sql string) ([]map[string]interface{}, error) {
    // Execute query
    return nil, nil
}

func (s *PostgreSQLService) Close() error {
    // Close connection
    return nil
}

// Register in your feature
func (f *DatabaseFeature) RegisterServices(container di.Container) error {
    return container.RegisterSingleton(
        di.TypeOf[DatabaseService](),
        func(ctx context.Context, c di.Container) (interface{}, error) {
            return &PostgreSQLService{
                connectionString: "postgres://user:pass@localhost/db",
            }, nil
        },
    )
}

// Use in your component
func (c *DatabaseComponent) Start(ctx context.Context) error {
    // Resolve the service
    db, err := di.Resolve[DatabaseService](c.container)
    if err != nil {
        return err
    }
    
    return db.Connect()
}
```

### 2. Feature Dependencies

```go
type APIFeature struct {
    name string
}

func (f *APIFeature) GetDependencies() []string {
    return []string{"database", "cache"} // Depends on database and cache features
}

func (f *APIFeature) GetPriority() int {
    return 200 // Higher priority (starts after dependencies)
}
```

### 3. Custom Configuration

```go
config := orchestrator.OrchestratorConfig{
    StartupTimeout:      30 * time.Second,
    ShutdownTimeout:     15 * time.Second,
    HealthCheckInterval: 30 * time.Second,
    EnableMetrics:       true,
    EnableTracing:       true,
    FeatureConfig: map[string]interface{}{
        "database": map[string]interface{}{
            "host":     "localhost",
            "port":     5432,
            "database": "myapp",
        },
        "cache": map[string]interface{}{
            "host": "localhost",
            "port": 6379,
        },
    },
}
```

## Complete Example

Here's a complete example of using the library:

```go
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

func main() {
    // Create logger
    slogLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
    logger := logger.NewSlogAdapter(slogLogger)
    
    // Create orchestrator
    config := orchestrator.DefaultOrchestratorConfig()
    orch, err := orchestrator.NewOrchestrator(config, logger)
    if err != nil {
        panic(err)
    }
    
    // Register features
    features := []orchestrator.Feature{
        &DatabaseFeature{name: "database"},
        &CacheFeature{name: "cache"},
        &APIFeature{name: "api-server"},
    }
    
    for _, feature := range features {
        if err := orch.RegisterFeature(feature); err != nil {
            panic(err)
        }
    }
    
    // Start application
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := orch.Start(ctx); err != nil {
        panic(err)
    }
    
    fmt.Println("Application started successfully!")
    
    // Simulate running
    time.Sleep(5 * time.Second)
    
    // Health check
    health := orch.HealthCheck(ctx)
    fmt.Printf("Health: %+v\n", health)
    
    // Graceful shutdown
    stopCtx, stopCancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer stopCancel()
    
    if err := orch.Stop(stopCtx); err != nil {
        panic(err)
    }
    
    fmt.Println("Application stopped successfully!")
}

// DatabaseFeature implementation
type DatabaseFeature struct {
    name string
}

func (f *DatabaseFeature) GetName() string { return f.name }
func (f *DatabaseFeature) GetDependencies() []string { return []string{} }
func (f *DatabaseFeature) GetPriority() int { return 10 }

func (f *DatabaseFeature) RegisterServices(container di.Container) error {
    return container.RegisterSingleton(
        di.TypeOf[DatabaseService](),
        func(ctx context.Context, c di.Container) (interface{}, error) {
            return &DatabaseService{}, nil
        },
    )
}

func (f *DatabaseFeature) CreateComponent(container di.Container) (lifecycle.Component, error) {
    return &DatabaseComponent{name: f.name}, nil
}

func (f *DatabaseFeature) GetRetryConfig() *lifecycle.RetryConfig { return nil }

func (f *DatabaseFeature) GetMetadata() orchestrator.FeatureMetadata {
    return orchestrator.FeatureMetadata{
        Name:        f.name,
        Description: "Database service",
    }
}

// DatabaseService interface
type DatabaseService interface {
    Connect() error
}

type DatabaseService struct{}

func (s *DatabaseService) Connect() error {
    fmt.Println("Connecting to database...")
    time.Sleep(100 * time.Millisecond)
    fmt.Println("Database connected!")
    return nil
}

// DatabaseComponent implementation
type DatabaseComponent struct {
    name string
}

func (c *DatabaseComponent) Name() string { return c.name }
func (c *DatabaseComponent) Dependencies() []string { return []string{} }
func (c *DatabaseComponent) Priority() int { return 10 }

func (c *DatabaseComponent) Start(ctx context.Context) error {
    fmt.Printf("Starting %s...\n", c.name)
    return nil
}

func (c *DatabaseComponent) Stop(ctx context.Context) error {
    fmt.Printf("Stopping %s...\n", c.name)
    return nil
}

func (c *DatabaseComponent) Health(ctx context.Context) lifecycle.ComponentHealth {
    return lifecycle.ComponentHealth{
        Status:    lifecycle.HealthStatusHealthy,
        Message:   "Database is healthy",
        Timestamp: time.Now(),
    }
}

func (c *DatabaseComponent) GetRetryConfig() *lifecycle.RetryConfig { return nil }

// Similar implementations for CacheFeature and APIFeature...
type CacheFeature struct {
    name string
}

func (f *CacheFeature) GetName() string { return f.name }
func (f *CacheFeature) GetDependencies() []string { return []string{"database"} }
func (f *CacheFeature) GetPriority() int { return 20 }

func (f *CacheFeature) RegisterServices(container di.Container) error { return nil }
func (f *CacheFeature) CreateComponent(container di.Container) (lifecycle.Component, error) {
    return &CacheComponent{name: f.name}, nil
}
func (f *CacheFeature) GetRetryConfig() *lifecycle.RetryConfig { return nil }
func (f *CacheFeature) GetMetadata() orchestrator.FeatureMetadata {
    return orchestrator.FeatureMetadata{Name: f.name, Description: "Cache service"}
}

type CacheComponent struct {
    name string
}

func (c *CacheComponent) Name() string { return c.name }
func (c *CacheComponent) Dependencies() []string { return []string{"database"} }
func (c *CacheComponent) Priority() int { return 20 }
func (c *CacheComponent) Start(ctx context.Context) error {
    fmt.Printf("Starting %s...\n", c.name)
    return nil
}
func (c *CacheComponent) Stop(ctx context.Context) error {
    fmt.Printf("Stopping %s...\n", c.name)
    return nil
}
func (c *CacheComponent) Health(ctx context.Context) lifecycle.ComponentHealth {
    return lifecycle.ComponentHealth{
        Status:    lifecycle.HealthStatusHealthy,
        Message:   "Cache is healthy",
        Timestamp: time.Now(),
    }
}
func (c *CacheComponent) GetRetryConfig() *lifecycle.RetryConfig { return nil }

type APIFeature struct {
    name string
}

func (f *APIFeature) GetName() string { return f.name }
func (f *APIFeature) GetDependencies() []string { return []string{"database", "cache"} }
func (f *APIFeature) GetPriority() int { return 30 }

func (f *APIFeature) RegisterServices(container di.Container) error { return nil }
func (f *APIFeature) CreateComponent(container di.Container) (lifecycle.Component, error) {
    return &APIComponent{name: f.name}, nil
}
func (f *APIFeature) GetRetryConfig() *lifecycle.RetryConfig { return nil }
func (f *APIFeature) GetMetadata() orchestrator.FeatureMetadata {
    return orchestrator.FeatureMetadata{Name: f.name, Description: "API server"}
}

type APIComponent struct {
    name string
}

func (c *APIComponent) Name() string { return c.name }
func (c *APIComponent) Dependencies() []string { return []string{"database", "cache"} }
func (c *APIComponent) Priority() int { return 30 }
func (c *APIComponent) Start(ctx context.Context) error {
    fmt.Printf("Starting %s...\n", c.name)
    return nil
}
func (c *APIComponent) Stop(ctx context.Context) error {
    fmt.Printf("Stopping %s...\n", c.name)
    return nil
}
func (c *APIComponent) Health(ctx context.Context) lifecycle.ComponentHealth {
    return lifecycle.ComponentHealth{
        Status:    lifecycle.HealthStatusHealthy,
        Message:   "API is healthy",
        Timestamp: time.Now(),
    }
}
func (c *APIComponent) GetRetryConfig() *lifecycle.RetryConfig { return nil }
```

## Best Practices

### 1. Error Handling
Always handle errors properly and provide meaningful error messages.

### 2. Graceful Shutdown
Always implement proper cleanup in your components' `Stop` methods.

### 3. Health Checks
Implement meaningful health checks that reflect the actual state of your components.

### 4. Dependencies
Design your features with clear, minimal dependencies to avoid circular dependencies.

### 5. Configuration
Use the `FeatureConfig` to make your features configurable without code changes.

## Troubleshooting

### Common Issues

1. **Import Path Issues**: Make sure you're importing from `pkg/orchestrator`
2. **Interface Implementation**: Ensure all required methods are implemented
3. **Dependency Cycles**: Check for circular dependencies between features
4. **Context Timeouts**: Use appropriate timeouts for startup and shutdown

### Getting Help

- Check the [API Documentation](api.md)
- Look at the [examples](../examples/) directory
- Review the [contributing guide](../CONTRIBUTING.md)
