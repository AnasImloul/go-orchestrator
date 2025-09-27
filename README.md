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

## Project Structure

This project follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout):

```
go-orchestrator/
├── cmd/              # Example applications
│   └── example/      # Basic example application
├── internal/         # Private implementation (not importable)
│   ├── di/          # Dependency injection container
│   ├── lifecycle/   # Lifecycle management
│   └── logger/      # Logging interface
├── examples/        # Usage examples
│   ├── basic/       # Simple usage example
│   ├── advanced/    # Complex orchestration example
│   ├── simple/      # New declarative API example
│   └── external-usage/ # External usage example
├── docs/           # Documentation
│   ├── api.md      # API documentation
│   └── external-usage.md # External usage guide
├── orchestrator.go # Single entry point for the library
├── .gitignore
├── LICENSE
├── CHANGELOG.md
├── CONTRIBUTING.md
├── Makefile
├── go.mod
└── README.md
```

## Public API

The library provides a **single entry point** with a clean, declarative API:

- **`github.com/AnasImloul/go-orchestrator`** - Single entry point with all functionality

**Note**: The `internal/` packages are not accessible to external projects and should not be imported.

## Installation

```bash
go get github.com/AnasImloul/go-orchestrator
```

### Using Specific Version

```bash
# Get latest version
go get github.com/AnasImloul/go-orchestrator@latest

# Get specific version
go get github.com/AnasImloul/go-orchestrator@v1.0.0
```

## Quick Start

```go
package main

import (
    "context"
    "reflect"
    
    "github.com/AnasImloul/go-orchestrator"
)

func main() {
    // Create application
    app := orchestrator.New()
    
    // Add features declaratively
    app.AddFeature(
        orchestrator.WithServiceInstanceT[DatabaseService]("database",
            &databaseService{host: "localhost", port: 5432},
        ).
            WithComponent(
                orchestrator.NewComponent().
                    WithStart(func(ctx context.Context, container *orchestrator.Container) error {
                        db, _ := orchestrator.ResolveType[DatabaseService](container)
                        return db.Connect()
                    }).
                    WithStop(func(ctx context.Context) error {
                        db, _ := orchestrator.ResolveType[DatabaseService](app.Container())
                        return db.Disconnect()
                    }),
            ),
    )
    
    // Start application
    ctx := context.Background()
    if err := app.Start(ctx); err != nil {
        panic(err)
    }
    
    // Application is now running...
    
    // Graceful shutdown
    app.Stop(ctx)
}
```

## Key Features

### Declarative API
- **Single entry point**: Import only `github.com/AnasImloul/go-orchestrator`
- **Fluent interface**: Chain method calls for clean, readable code
- **Type-safe**: Generic helpers for service resolution
- **Interface-based DI**: Enforces dependency on interfaces, not concrete types
- **Parallel execution**: Independent components start simultaneously for better performance
- **Less verbose**: Minimal boilerplate code

### Dependency Injection
```go
// Define interfaces (best practice)
type DatabaseService interface {
    Connect() error
    Disconnect() error
}

// Register services declaratively (two approaches)

// Approach 1: Using reflection (traditional)
app.AddFeature(
    orchestrator.NewFeature("database").
        WithServiceInstance(
            reflect.TypeOf((*DatabaseService)(nil)).Elem(),
            &databaseService{host: "localhost", port: 5432},
        ),
)

// Approach 2: Using generics (recommended - type-safe)
app.AddFeature(
    orchestrator.WithServiceInstanceT[DatabaseService]("database",
        &databaseService{host: "localhost", port: 5432},
    ),
)

// Resolve services by interface (enforced by library)
service, err := orchestrator.ResolveType[DatabaseService](container)

// ❌ This will fail at runtime - concrete types not allowed
// service, err := orchestrator.ResolveType[*databaseService](container)
```

### Lifecycle Management
```go
// Automatic startup/shutdown ordering based on dependencies
// Independent components start in parallel for better performance
app.AddFeature(
    orchestrator.NewFeature("database").
        WithDependencies("config"). // Depends on config
        WithComponent(
            orchestrator.NewComponent().
                WithStart(startFunc).
                WithStop(stopFunc).
                WithHealth(healthFunc),
        ),
)
```

### Parallel Execution

The orchestrator automatically groups components by dependency level and starts independent components in parallel:

```go
// These three services have no dependencies - they start in parallel at level 0
app.AddFeature(orchestrator.NewFeature("cache"))
app.AddFeature(orchestrator.NewFeature("metrics")) 
app.AddFeature(orchestrator.NewFeature("logging"))

// This service depends on all three - it starts at level 1 after they're ready
app.AddFeature(
    orchestrator.NewFeature("api").
        WithDependencies("cache", "metrics", "logging"),
)
```

**Execution Flow:**
- **Level 0**: cache, metrics, logging start simultaneously
- **Level 1**: api starts after all dependencies are ready

This provides significant performance improvements over sequential startup.

## Using as External Dependency

### 1. Install the Library

```bash
go get github.com/AnasImloul/go-orchestrator
```

### 2. Import Required Packages

```go
import (
    "github.com/AnasImloul/go-orchestrator/pkg/orchestrator"
    "github.com/AnasImloul/go-orchestrator/pkg/di"
    "github.com/AnasImloul/go-orchestrator/pkg/lifecycle"
    "github.com/AnasImloul/go-orchestrator/pkg/logger"
)
```

> **Note**: The public API is production-ready and fully functional. All examples are working and can be used in production applications.

### 3. Create Your Features

```go
type MyFeature struct {
    name string
}

func (f *MyFeature) GetName() string { return f.name }
func (f *MyFeature) GetDependencies() []string { return []string{} }
func (f *MyFeature) GetPriority() int { return 100 }
func (f *MyFeature) RegisterServices(container di.Container) error { return nil }
func (f *MyFeature) CreateComponent(container di.Container) (lifecycle.Component, error) {
    return &MyComponent{name: f.name}, nil
}
func (f *MyFeature) GetRetryConfig() *lifecycle.RetryConfig { return nil }
func (f *MyFeature) GetMetadata() orchestrator.FeatureMetadata {
    return orchestrator.FeatureMetadata{Name: f.name, Description: "My feature"}
}
```

### 4. Start the Orchestrator

```go
// Create logger
logger := logger.NewSlogAdapter(slog.Default())

// Create orchestrator
config := orchestrator.DefaultOrchestratorConfig()
orch, err := orchestrator.NewOrchestrator(config, logger)
if err != nil {
    panic(err)
}

// Register features
orch.RegisterFeature(&MyFeature{name: "my-service"})

// Start application
ctx := context.Background()
if err := orch.Start(ctx); err != nil {
    panic(err)
}

// Graceful shutdown
defer orch.Stop(ctx)
```

For complete examples, see the [usage documentation](docs/usage.md), [external usage guide](docs/external-usage.md), and [external usage example](examples/external-usage/).

## Examples

### Basic Example

See `examples/basic/main.go` for a simple usage example.

### Advanced Example

See `examples/advanced/main.go` for a complex orchestration example with multiple dependent services.

### External Usage Example

See `examples/external-usage/main.go` for a complete example showing how to use the library as an external dependency in your own project.

### Running Examples

```bash
# Run the basic example
go run examples/basic/main.go

# Run the advanced example
go run examples/advanced/main.go

# Run the external usage example
cd examples/external-usage
go run main.go

# Run the command-line example
go run cmd/example/main.go
```

## Development

### Building

```bash
# Build the example application
make build

# Run tests
make test

# Run with coverage
make test-coverage

# Format code
make fmt

# Lint code
make lint

# Run all checks
make check
```

### Project Layout

This project follows Go standard project layout conventions:

- **`pkg/`**: Public API that can be imported by external projects
- **`internal/`**: Private implementation details that cannot be imported externally
- **`cmd/`**: Example applications and command-line tools
- **`examples/`**: Usage examples and tutorials
- **`docs/`**: Documentation and API references

## API Documentation

See [docs/api.md](docs/api.md) for comprehensive API documentation.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
# Test commit for avatar verification
