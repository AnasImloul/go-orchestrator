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
├── pkg/             # Public API (importable by external projects)
│   ├── di/          # Public DI interfaces and types
│   ├── lifecycle/   # Public lifecycle interfaces and types
│   ├── logger/      # Public logger interfaces and adapters
│   └── orchestrator/ # Main orchestrator package
├── examples/        # Usage examples
│   ├── basic/       # Simple usage example
│   ├── advanced/    # Complex orchestration example
│   └── external-usage/ # External usage example
├── docs/           # Documentation
│   ├── api.md      # API documentation
│   └── external-usage.md # External usage guide
├── .gitignore
├── LICENSE
├── CHANGELOG.md
├── CONTRIBUTING.md
├── Makefile
├── go.mod
└── README.md
```

## Public API

The library provides a clean public API through the following packages:

- **`github.com/AnasImloul/go-orchestrator/pkg/orchestrator`** - Main orchestrator functionality
- **`github.com/AnasImloul/go-orchestrator/pkg/di`** - Dependency injection interfaces and types
- **`github.com/AnasImloul/go-orchestrator/pkg/lifecycle`** - Component lifecycle management
- **`github.com/AnasImloul/go-orchestrator/pkg/logger`** - Logging interfaces and adapters

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
    "log/slog"
    
    "github.com/AnasImloul/go-orchestrator/pkg/orchestrator"
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
```

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
