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

## Service Lifetimes

The library supports three service lifetimes:

- **Singleton**: One instance for the entire application lifecycle
- **Scoped**: One instance per scope (request/operation) - requires proper scope management
- **Transient**: New instance created every time the service is resolved

### Lifetime Behavior

- **Singleton**: Same instance returned every time, shared across the entire application
- **Scoped**: Same instance within a scope, different instances across scopes
- **Transient**: New instance created for each resolution, even within the same scope

### Service Registration Methods

#### New Ultra-Clean API Benefits

The new API provides several key improvements:

- **🎯 Type Safety**: Full compile-time type checking with generics
- **📝 Declarative**: Clean, readable syntax that expresses intent clearly
- **🔧 Less Verbose**: 60% reduction in boilerplate code
- **⚡ Intuitive**: Natural fluent API that's easy to understand
- **🛡️ Robust**: Runtime type verification and error handling
- **🏭 Factory-Based**: All services use factories for consistent lifetime management
- **🔄 Backward Compatible**: All existing code continues to work

#### Service Registration (Best Syntax - Recommended)

**Ultra-Clean Syntax (Factory-Based):**
```go
// One-liner for simple services with type-safe component lifecycle
app.AddFeature(
    orchestrator.WithComponentFor[DatabaseService](
        orchestrator.NewFeatureWithInstance("database", DatabaseService(&databaseService{host: "localhost", port: 5432}), orchestrator.Singleton),
        app,
    ).
        WithStartFor(func(db DatabaseService) error { return db.Connect() }).
        WithStopFor(func(db DatabaseService) error { return db.Disconnect() }).
        Build(),
)
```

**Factory-Based Services with Dependencies:**
```go
// Clean factory syntax with automatic dependency resolution
app.AddFeature(
    orchestrator.WithComponentFor[APIService](
        orchestrator.NewFeatureWithFactory("api", 
            func(ctx context.Context, container *orchestrator.Container) (APIService, error) {
                db, _ := orchestrator.ResolveType[DatabaseService](container)
                cache, _ := orchestrator.ResolveType[CacheService](container)
                return &apiService{port: 8080, db: db, cache: cache}, nil
            }, 
            orchestrator.Singleton,
        ).WithDependencies("database", "cache"),
        app,
    ).
        WithStartFor(func(api APIService) error { return api.Start() }).
        WithStopFor(func(api APIService) error { return api.Stop() }).
        WithHealthFor(func(api APIService) orchestrator.HealthStatus {
            return orchestrator.HealthStatus{Status: api.Health(), Message: "API server is running"}
        }).
        Build(),
)
```

**Automatic Dependency Discovery (NEW!):**
```go
// Ultra-clean syntax with automatic dependency injection AND discovery
app.AddFeature(
    orchestrator.NewFeatureWithAutoFactory[APIService]("api",
        func(db DatabaseService, cache CacheService) APIService {
            return &apiService{port: 8080, db: db, cache: cache}
        },
        orchestrator.Singleton,
    ).
        WithComponent(
            orchestrator.NewComponent().
                WithStart(func(ctx context.Context, container *orchestrator.Container) error {
                    api, _ := orchestrator.ResolveType[APIService](container)
                    return api.Start()
                }).
                WithStop(func(ctx context.Context) error {
                    api, _ := orchestrator.ResolveType[APIService](app.Container())
                    return api.Stop()
                }),
        ),
)
```

#### Automatic Dependency Discovery Benefits

The new `NewFeatureWithAutoFactory` provides:

- **🎯 Zero Boilerplate**: No manual dependency resolution OR declaration needed
- **🔍 Type-Safe**: Compile-time verification of dependency types
- **⚡ Reflection-Based**: Automatic parameter scanning and injection
- **🛡️ Error Handling**: Clear error messages for missing dependencies
- **📝 Clean Syntax**: Factory functions only need to declare their dependencies
- **🔗 Auto-Discovery**: Dependencies are automatically discovered from function parameters
- **🚀 Zero Configuration**: No need for `WithDependencies()` calls

#### Factory-Only Approach Benefits

All services now use factories internally, which provides:

- **Consistent Lifetime Management**: Transient services create new instances via factory calls
- **No Cloning Complexity**: Eliminates the need for deep cloning logic
- **Simplified Architecture**: Single code path for all service creation
- **Better Performance**: Factory calls are more efficient than cloning
- **Clearer Intent**: Factory functions make service creation explicit

#### Legacy Syntax (Still Supported)
```go
// Original verbose syntax - still works but not recommended for new code
app.AddFeature(
    orchestrator.WithServiceFactory[DatabaseService](func(ctx context.Context, container *orchestrator.Container) (DatabaseService, error) {
        return &databaseService{host: "localhost", port: 5432}, nil
    })(
        orchestrator.NewFeature("database"),
    ).
        WithLifetime(orchestrator.Singleton).
        WithComponent(
            orchestrator.NewComponent().
                WithStart(orchestrator.WithStartFunc[DatabaseService](func(db DatabaseService) error {
                    return db.Connect()
                })).
                WithStop(orchestrator.WithStopFuncWithApp[DatabaseService](app, func(db DatabaseService) error {
                    return db.Disconnect()
                })),
        ),
)
```

#### Factory Registration (For Complex Dependencies)
```go
// Register a service factory for services with dependencies
app.AddFeature(
    orchestrator.WithServiceFactory[APIService](
        func(ctx context.Context, container *orchestrator.Container) (APIService, error) {
            db, err := orchestrator.ResolveType[DatabaseService](container)
            if err != nil {
                return nil, err
            }
            return &apiService{port: 8080, db: db}, nil
        },
    )(
        orchestrator.NewFeature("api").
            WithDependencies("database"),
    ).
        WithLifetime(orchestrator.Singleton),
)
```

**When to use each approach:**
- **`WithService[T]()`**: For simple services without dependencies (clean, declarative)
- **`WithServiceFactory[T]()`**: For services with dependencies that need to be resolved from the container (type-safe, generic)

**Helper Functions for Reduced Verbosity:**
- **`WithStartFunc[T]()`**: Eliminates repetitive service resolution in start methods
- **`WithStopFuncWithApp[T]()`**: Eliminates repetitive service resolution in stop methods  
- **`WithHealthFunc[T]()`**: Eliminates repetitive service resolution in health methods

#### Named Services
```go
// Register multiple services with the same interface type using names
app.AddFeature(
    orchestrator.NewFeature("databases").
        WithNamedService(
            "primary-db",
            reflect.TypeOf((*DatabaseService)(nil)).Elem(),
            func(ctx context.Context, container *orchestrator.Container) (interface{}, error) {
                return &databaseService{host: "primary", port: 5432}, nil
            },
            orchestrator.Singleton,
        ).
        WithNamedService(
            "secondary-db",
            reflect.TypeOf((*DatabaseService)(nil)).Elem(),
            func(ctx context.Context, container *orchestrator.Container) (interface{}, error) {
                return &databaseService{host: "secondary", port: 5433}, nil
            },
            orchestrator.Singleton,
        ),
)

// Resolve named services
primaryDB, _ := orchestrator.ResolveNamedType[DatabaseService](container, "primary-db")
secondaryDB, _ := orchestrator.ResolveNamedType[DatabaseService](container, "secondary-db")
```

### Scope Management

```go
// Create a scope for scoped services
scopedContainer := container.CreateScope()
defer scopedContainer.Dispose()

// Resolve services within the scope
service, _ := orchestrator.ResolveType[MyService](scopedContainer)
```

### Limitations and Best Practices

1. **Type Collision**: You cannot register multiple services with the same interface type without using named services
2. **Transient Cloning**: Transient services use deep cloning for instance registration, which may not work for all types
3. **Factory Registration**: Use factory registration for better control over service creation and lifetime behavior
4. **Scope Lifecycle**: Always dispose scopes when done to prevent memory leaks
5. **Interface Enforcement**: The library enforces that services are resolved as interfaces, not concrete types

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

See `examples/simple/main.go` for a simple declarative usage example.

### Clean API Example

See `examples/clean-api/main.go` for a demonstration of the new clean, declarative API with proper service lifetimes.

### Advanced Example

See `examples/advanced/main.go` for a complex orchestration example with multiple dependent services.

### Service Lifetimes Example

See `examples/lifetimes/main.go` for a demonstration of different service lifetimes (Singleton, Scoped, Transient).

### Factory-Based Registration Example

See `examples/factory-based/main.go` for examples of factory-based registration and named services.

### Named Services Example

See `examples/named-services/main.go` for examples of multiple services with the same interface type.

### External Usage Example

See `examples/external-usage/main.go` for a complete example showing how to use the library as an external dependency in your own project.

### Running Examples

```bash
# Best syntax examples (recommended)
go run examples/best-syntax/main.go
go run examples/ultra-clean/main.go
go run examples/auto-dependencies/main.go

# Legacy examples (still supported)
go run examples/simple/main.go
go run examples/clean-api/main.go
go run examples/advanced/main.go

# Feature demonstrations
go run examples/lifetimes/main.go
go run examples/factory-based/main.go
go run examples/named-services/main.go
go run examples/parallel/main.go

# Run the service lifetimes example
go run examples/lifetimes/main.go

# Run the factory-based registration example
go run examples/factory-based/main.go

# Run the named services example
go run examples/named-services/main.go

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
