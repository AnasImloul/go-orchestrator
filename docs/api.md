# API Documentation

## Overview

Go Orchestrator provides a comprehensive API for application orchestration, dependency injection, and lifecycle management.

## Core Interfaces

### Orchestrator

The main orchestrator interface that coordinates application lifecycle and dependency injection.

```go
type Orchestrator interface {
    RegisterFeature(feature Feature) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    GetContainer() di.Container
    GetLifecycleManager() lifecycle.LifecycleManager
    GetPhase() lifecycle.Phase
    HealthCheck(ctx context.Context) HealthReport
}
```

### Feature

Represents a feature that can be managed by the orchestrator.

```go
type Feature interface {
    GetName() string
    GetDependencies() []string
    GetPriority() int
    RegisterServices(container di.Container) error
    CreateComponent(container di.Container) (lifecycle.Component, error)
    GetRetryConfig() *lifecycle.RetryConfig
    GetMetadata() FeatureMetadata
}
```

### Container (Dependency Injection)

The dependency injection container interface.

```go
type Container interface {
    Register(serviceType reflect.Type, factory Factory, options ...Option) error
    RegisterInstance(serviceType reflect.Type, instance interface{}) error
    RegisterSingleton(serviceType reflect.Type, factory Factory, options ...Option) error
    Resolve(serviceType reflect.Type) (interface{}, error)
    ResolveWithScope(serviceType reflect.Type, scope Scope) (interface{}, error)
    CreateScope() Scope
    Validate() error
}
```

### LifecycleManager

Manages component lifecycle operations.

```go
type LifecycleManager interface {
    RegisterComponent(component Component) error
    UnregisterComponent(name string) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    GetPhase() Phase
    GetComponentState(name string) (*ComponentState, error)
    GetAllComponentStates() map[string]*ComponentState
    AddHook(phase Phase, hook Hook) error
    RemoveHook(phase Phase, hook Hook) error
}
```

## Configuration

### OrchestratorConfig

Configuration for the orchestrator.

```go
type OrchestratorConfig struct {
    StartupTimeout       time.Duration
    ShutdownTimeout      time.Duration
    HealthCheckInterval  time.Duration
    EnableMetrics        bool
    EnableTracing        bool
    FeatureConfig        map[string]interface{}
}
```

### ContainerConfig

Configuration for the dependency injection container.

```go
type ContainerConfig struct {
    EnableValidation     bool
    EnableCircularCheck  bool
    EnableInterception   bool
    DefaultLifetime      Lifetime
    MaxResolutionDepth   int
    EnableMetrics        bool
}
```

## Helper Functions

### Generic Resolution

Type-safe service resolution helpers:

```go
func Resolve[T any](container Container) (T, error)
func MustResolve[T any](container Container) T
func TypeOf[T any]() reflect.Type
```

## Examples

See the `examples/` directory for comprehensive usage examples.

## Error Handling

All functions return errors that should be checked and handled appropriately. Common error types include:

- `ErrServiceNotFound`: Service not registered in container
- `ErrCircularDependency`: Circular dependency detected
- `ErrComponentNotFound`: Component not found in lifecycle manager
- `ErrInvalidConfiguration`: Invalid configuration provided
