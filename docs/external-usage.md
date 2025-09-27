# Using Go Orchestrator as External Dependency

This guide explains how to use the Go Orchestrator library in your own projects.

## Current Status

The Go Orchestrator library is production-ready with a stable public API. All features are fully functional and can be used in production applications.

## Installation

### Using Go Modules

```bash
go get github.com/AnasImloul/go-orchestrator
```

### Using Specific Version

```bash
# Get latest version
go get github.com/AnasImloul/go-orchestrator@latest

# Get specific version (when available)
go get github.com/AnasImloul/go-orchestrator@v1.0.0
```

## Project Structure

When using the library in your project, your structure might look like:

```
your-project/
├── cmd/
│   └── your-app/
│       └── main.go
├── internal/
│   ├── features/
│   │   ├── database/
│   │   ├── cache/
│   │   └── api/
│   └── services/
├── go.mod
└── go.sum
```

## Basic Usage Pattern

### 1. Create Your Main Application

```go
package main

import (
    "context"
    "log/slog"
    "os"
    "time"

    "github.com/AnasImloul/go-orchestrator/pkg/orchestrator"
    "github.com/AnasImloul/go-orchestrator/pkg/logger"
)

func main() {
    // Create logger
    slogLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
    logger := logger.NewSlogAdapter(slogLogger)

    // Create orchestrator configuration
    config := orchestrator.DefaultOrchestratorConfig()
    config.StartupTimeout = 30 * time.Second
    config.ShutdownTimeout = 15 * time.Second

    // Create orchestrator
    orch, err := orchestrator.NewOrchestrator(config, logger)
    if err != nil {
        panic(err)
    }

    // Register your features
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
    ctx := context.Background()
    if err := orch.Start(ctx); err != nil {
        panic(err)
    }

    // Graceful shutdown
    defer orch.Stop(ctx)
}
```

### 2. Implement Your Features

```go
package features

import (
    "context"
    "time"

    "github.com/AnasImloul/go-orchestrator/pkg/di"
    "github.com/AnasImloul/go-orchestrator/pkg/lifecycle"
    "github.com/AnasImloul/go-orchestrator/pkg/orchestrator"
)

type DatabaseFeature struct {
    name string
}

func (f *DatabaseFeature) GetName() string {
    return f.name
}

func (f *DatabaseFeature) GetDependencies() []string {
    return []string{} // No dependencies
}

func (f *DatabaseFeature) GetPriority() int {
    return 10 // Lower numbers start first
}

func (f *DatabaseFeature) RegisterServices(container di.Container) error {
    return container.RegisterSingleton(
        di.TypeOf[DatabaseService](),
        func(ctx context.Context, c di.Container) (interface{}, error) {
            return &DatabaseServiceImpl{
                host:     "localhost",
                port:     5432,
                database: "myapp",
            }, nil
        },
    )
}

func (f *DatabaseFeature) CreateComponent(container di.Container) (lifecycle.Component, error) {
    return &DatabaseComponent{name: f.name}, nil
}

func (f *DatabaseFeature) GetRetryConfig() *lifecycle.RetryConfig {
    return nil // Use default retry config
}

func (f *DatabaseFeature) GetMetadata() orchestrator.FeatureMetadata {
    return orchestrator.FeatureMetadata{
        Name:        f.name,
        Version:     "1.0.0",
        Description: "PostgreSQL database service",
        Tags:        []string{"database", "postgresql", "persistence"},
    }
}
```

### 3. Implement Your Components

```go
type DatabaseComponent struct {
    name string
}

func (c *DatabaseComponent) Name() string {
    return c.name
}

func (c *DatabaseComponent) Dependencies() []string {
    return []string{}
}

func (c *DatabaseComponent) Priority() int {
    return 10
}

func (c *DatabaseComponent) Start(ctx context.Context) error {
    // Initialize your database connection
    return nil
}

func (c *DatabaseComponent) Stop(ctx context.Context) error {
    // Cleanup database connections
    return nil
}

func (c *DatabaseComponent) Health(ctx context.Context) lifecycle.ComponentHealth {
    return lifecycle.ComponentHealth{
        Status:    lifecycle.HealthStatusHealthy,
        Message:   "Database is healthy",
        Timestamp: time.Now(),
    }
}

func (c *DatabaseComponent) GetRetryConfig() *lifecycle.RetryConfig {
    return nil
}
```

### 4. Define Your Services

```go
type DatabaseService interface {
    Connect() error
    Query(sql string) ([]map[string]interface{}, error)
    Close() error
}

type DatabaseServiceImpl struct {
    host     string
    port     int
    database string
}

func (s *DatabaseServiceImpl) Connect() error {
    // Connect to database
    return nil
}

func (s *DatabaseServiceImpl) Query(sql string) ([]map[string]interface{}, error) {
    // Execute query
    return []map[string]interface{}{}, nil
}

func (s *DatabaseServiceImpl) Close() error {
    // Close connection
    return nil
}
```

## Advanced Usage

### Feature Dependencies

```go
type APIFeature struct {
    name string
}

func (f *APIFeature) GetDependencies() []string {
    return []string{"database", "cache"} // Depends on database and cache
}

func (f *APIFeature) GetPriority() int {
    return 30 // Higher priority (starts after dependencies)
}
```

### Custom Configuration

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

### Health Checks

```go
// Perform health check
health := orch.HealthCheck(ctx)
fmt.Printf("Overall Health: %s\n", health.Status)
fmt.Printf("Healthy Features: %d/%d\n", 
    health.Summary.HealthyFeatures, 
    health.Summary.TotalFeatures)

// Check specific feature health
if featureHealth, exists := health.Features["database"]; exists {
    fmt.Printf("Database Health: %s\n", featureHealth.Status)
}
```

## Best Practices

### 1. Error Handling

Always handle errors properly and provide meaningful error messages:

```go
func (c *DatabaseComponent) Start(ctx context.Context) error {
    if err := c.connect(); err != nil {
        return fmt.Errorf("failed to connect to database: %w", err)
    }
    return nil
}
```

### 2. Graceful Shutdown

Implement proper cleanup in your components:

```go
func (c *DatabaseComponent) Stop(ctx context.Context) error {
    // Wait for ongoing operations to complete
    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-c.waitForOperations():
        // Operations completed
    }
    
    // Close connections
    return c.close()
}
```

### 3. Health Checks

Implement meaningful health checks:

```go
func (c *DatabaseComponent) Health(ctx context.Context) lifecycle.ComponentHealth {
    if err := c.ping(); err != nil {
        return lifecycle.ComponentHealth{
            Status:    lifecycle.HealthStatusUnhealthy,
            Message:   fmt.Sprintf("Database ping failed: %v", err),
            Timestamp: time.Now(),
        }
    }
    
    return lifecycle.ComponentHealth{
        Status:    lifecycle.HealthStatusHealthy,
        Message:   "Database is responsive",
        Timestamp: time.Now(),
    }
}
```

### 4. Dependency Design

Design your features with clear, minimal dependencies:

```go
// Good: Clear, minimal dependencies
func (f *APIFeature) GetDependencies() []string {
    return []string{"database", "cache"}
}

// Avoid: Circular dependencies
func (f *BadFeature) GetDependencies() []string {
    return []string{"feature-that-depends-on-me"}
}
```

## Working Examples

For complete, working examples, see:

- `examples/basic/main.go` - Simple usage example
- `examples/advanced/main.go` - Complex orchestration example
- `cmd/example/main.go` - Command-line application example

## Troubleshooting

### Common Issues

1. **Import Path Issues**: Make sure you're importing from the correct packages
2. **Interface Implementation**: Ensure all required methods are implemented
3. **Dependency Cycles**: Check for circular dependencies between features
4. **Context Timeouts**: Use appropriate timeouts for startup and shutdown

### Getting Help

- Check the [API Documentation](api.md)
- Look at the [examples](../examples/) directory
- Review the [contributing guide](../CONTRIBUTING.md)
- Open an issue on GitHub for bugs or feature requests

## Migration Guide

If you're migrating from an older version or different orchestration library:

1. **Update Import Paths**: Change to the new package structure
2. **Implement New Interfaces**: Ensure your features implement the required interfaces
3. **Update Configuration**: Use the new configuration structure
4. **Test Thoroughly**: Verify all features work as expected

## Roadmap

The public API is being finalized with these planned features:

- [ ] Stable public API interfaces
- [ ] Complete documentation
- [ ] More examples and tutorials
- [ ] Performance optimizations
- [ ] Additional middleware and plugins
