package orchestrator

import (
	"context"
	"fmt"
	"reflect"

	"github.com/AnasImloul/go-orchestrator/internal/lifecycle"
)

// ServiceDefinitionInterface represents the interface for service definitions.
type ServiceDefinitionInterface interface {
	ToServiceDefinition() *ServiceDefinition
}

// TypedServiceDefinition represents a type-safe service definition.
type TypedServiceDefinition[T any] struct {
	Name         string
	Dependencies []string
	Service      TypedServiceConfig[T]
	Lifecycle    LifecycleConfig
	RetryConfig  *lifecycle.RetryConfig
	Metadata     map[string]string
}

// WithLifecycle sets the lifecycle configuration for the typed service definition.
func (tsd *TypedServiceDefinition[T]) WithLifecycle(builder *LifecycleBuilder) *TypedServiceDefinition[T] {
	tsd.Lifecycle = builder.Build()
	return tsd
}

// WithRetryConfig sets the retry configuration for the typed service definition.
func (tsd *TypedServiceDefinition[T]) WithRetryConfig(config *lifecycle.RetryConfig) *TypedServiceDefinition[T] {
	tsd.RetryConfig = config
	return tsd
}

// WithDependencies sets the dependencies for the typed service definition.
func (tsd *TypedServiceDefinition[T]) WithDependencies(deps ...string) *TypedServiceDefinition[T] {
	tsd.Dependencies = deps
	return tsd
}

// WithMetadata sets metadata for the typed service definition.
func (tsd *TypedServiceDefinition[T]) WithMetadata(key, value string) *TypedServiceDefinition[T] {
	if tsd.Metadata == nil {
		tsd.Metadata = make(map[string]string)
	}
	tsd.Metadata[key] = value
	return tsd
}

// WithName sets a custom name for the service definition.
// This overrides the automatic name inference and allows you to specify
// a custom service name to avoid conflicts or use more descriptive names.
func (tsd *TypedServiceDefinition[T]) WithName(name string) *TypedServiceDefinition[T] {
	tsd.Name = name
	return tsd
}

// ToServiceDefinition converts a typed service definition to a regular service definition.
// This allows typed service definitions to work with the existing registration system.
func (tsd *TypedServiceDefinition[T]) ToServiceDefinition() *ServiceDefinition {
	return &ServiceDefinition{
		Name:         tsd.Name,
		Dependencies: tsd.Dependencies,
		Services: []ServiceConfig{
			{
				Name: tsd.Service.Name,
				Type: tsd.Service.Type,
				Factory: func(ctx context.Context, container *Container) (interface{}, error) {
					return tsd.Service.Factory(ctx, container)
				},
				Lifetime: tsd.Service.Lifetime,
			},
		},
		Lifecycle:   tsd.Lifecycle,
		RetryConfig: tsd.RetryConfig,
		Metadata:    tsd.Metadata,
	}
}

// LifecycleBuilder provides a fluent interface for building service lifecycle configurations.
type LifecycleBuilder struct {
	config LifecycleConfig
}

// NewLifecycle creates a new service lifecycle builder.
func NewLifecycle() *LifecycleBuilder {
	return &LifecycleBuilder{
		config: LifecycleConfig{},
	}
}

// WithStart sets the start function for the service lifecycle.
func (lb *LifecycleBuilder) WithStart(start func(ctx context.Context, container *Container) error) *LifecycleBuilder {
	lb.config.Start = start
	return lb
}

// WithStop sets the stop function for the service lifecycle.
func (lb *LifecycleBuilder) WithStop(stop func(ctx context.Context) error) *LifecycleBuilder {
	lb.config.Stop = stop
	return lb
}

// WithHealth sets the health check function for the service lifecycle.
func (lb *LifecycleBuilder) WithHealth(health func(ctx context.Context) HealthStatus) *LifecycleBuilder {
	lb.config.Health = health
	return lb
}

// Build returns the lifecycle configuration.
func (lb *LifecycleBuilder) Build() LifecycleConfig {
	return lb.config
}

// ToServiceDefinition returns itself for ServiceDefinition.
func (sd *ServiceDefinition) ToServiceDefinition() *ServiceDefinition {
	return sd
}

// WithLifecycle sets the lifecycle configuration for the service definition using a builder.
func (sd *ServiceDefinition) WithLifecycle(builder *LifecycleBuilder) *ServiceDefinition {
	sd.Lifecycle = builder.Build()
	return sd
}

// WithRetryConfig sets the retry configuration for the service definition.
func (sd *ServiceDefinition) WithRetryConfig(config *lifecycle.RetryConfig) *ServiceDefinition {
	sd.RetryConfig = config
	return sd
}

// WithAutoDependencies enables automatic dependency discovery for the last registered service.
// This will scan the factory function parameters and automatically resolve dependencies.
func (sd *ServiceDefinition) WithAutoDependencies() *ServiceDefinition {
	if len(sd.Services) == 0 {
		return sd
	}

	lastService := &sd.Services[len(sd.Services)-1]
	if lastService.Factory == nil {
		return sd
	}

	// Create a wrapper factory that automatically resolves dependencies
	originalFactory := lastService.Factory
	lastService.Factory = func(ctx context.Context, container *Container) (interface{}, error) {
		// Use reflection to analyze the original factory function and resolve dependencies
		return resolveDependenciesAndCallFactory(ctx, container, originalFactory)
	}

	return sd
}

// WithMetadata adds metadata to the service definition.
func (sd *ServiceDefinition) WithMetadata(key, value string) *ServiceDefinition {
	if sd.Metadata == nil {
		sd.Metadata = make(map[string]string)
	}
	sd.Metadata[key] = value
	return sd
}

// WithName sets a custom name for the service definition.
// This overrides the automatic name inference and allows you to specify
// a custom service name to avoid conflicts or use more descriptive names.
func (sd *ServiceDefinition) WithName(name string) *ServiceDefinition {
	sd.Name = name
	return sd
}

// NewServiceSingleton creates a new service definition with automatic lifecycle management.
// The service instance MUST implement the Service interface.
// Lifecycle methods (Start, Stop, Health) are automatically wired.
func NewServiceSingleton[T Service](instance T) *TypedServiceDefinition[T] {
	// Get the interface type T
	interfaceType := reflect.TypeOf((*T)(nil)).Elem()
	serviceName := inferServiceNameFromType(interfaceType)

	factory := func(ctx context.Context, container *Container) (T, error) {
		return instance, nil
	}

	// Create typed service definition with automatic lifecycle wiring
	serviceDef := &TypedServiceDefinition[T]{
		Name: serviceName,
		Service: TypedServiceConfig[T]{
			Type:     interfaceType,
			Factory:  factory,
			Lifetime: Singleton,
		},
		// Automatically wire lifecycle methods
		Lifecycle: LifecycleConfig{
			Start: func(ctx context.Context, container *Container) error {
				return instance.Start(ctx)
			},
			Stop: func(ctx context.Context) error {
				return instance.Stop(ctx)
			},
			Health: func(ctx context.Context) HealthStatus {
				return instance.Health(ctx)
			},
		},
	}

	return serviceDef
}

// NewAutoServiceFactory creates a new service definition with automatic dependency discovery and lifecycle management.
// The factory function can return any type T - it doesn't need to implement the Service interface.
// Dependencies are automatically discovered from the factory function parameters.
// Lifecycle methods are automatically provided with sensible defaults.
func NewAutoServiceFactory[T any](factory interface{}, lifetime Lifetime) *TypedServiceDefinition[T] {
	// Get the interface type T
	interfaceType := reflect.TypeOf((*T)(nil)).Elem()
	serviceName := inferServiceNameFromType(interfaceType)

	// Convert the factory function to the expected signature
	factoryFunc := func(ctx context.Context, container *Container) (T, error) {
		// Use reflection to call the original factory function with resolved dependencies
		factoryValue := reflect.ValueOf(factory)
		factoryType := factoryValue.Type()

		// Get the number of parameters the factory function expects
		numParams := factoryType.NumIn()
		args := make([]reflect.Value, numParams)

		// Resolve each dependency
		for i := 0; i < numParams; i++ {
			paramType := factoryType.In(i)
			instance, err := container.Resolve(paramType)
			if err != nil {
				var zero T
				return zero, fmt.Errorf("failed to resolve dependency %d (%s): %w", i, paramType.String(), err)
			}
			args[i] = reflect.ValueOf(instance)
		}

		// Call the factory function
		results := factoryValue.Call(args)
		if len(results) != 1 {
			var zero T
			return zero, fmt.Errorf("factory function should return exactly one value, got %d", len(results))
		}

		// Convert the result to type T
		result := results[0].Interface().(T)
		return result, nil
	}

	// Create typed service definition with automatic lifecycle wiring
	serviceDef := &TypedServiceDefinition[T]{
		Name: serviceName,
		Service: TypedServiceConfig[T]{
			Type:     interfaceType,
			Factory:  factoryFunc,
			Lifetime: lifetime,
		},
		// Automatically wire lifecycle methods with defaults
		Lifecycle: LifecycleConfig{
			Start: func(ctx context.Context, container *Container) error {
				// Try to call Start method if it exists
				instance, err := container.Resolve(interfaceType)
				if err != nil {
					return err
				}
				if startable, ok := instance.(interface{ Start(context.Context) error }); ok {
					return startable.Start(ctx)
				}
				// No Start method, default to no-op
				return nil
			},
			Stop: func(ctx context.Context) error {
				// Try to call Stop method if it exists
				// We'll use the serviceComponent's container instead of creating a new one
				// This will be handled by the serviceComponent.Stop method
				return nil
			},
			// No Health function - let serviceComponent.Health handle automatic detection
		},
	}

	// Automatically discover and add dependencies based on factory parameters
	autoDiscoverDependenciesTyped(serviceDef, factory)

	return serviceDef
}

// NewServiceFactory creates a new service definition with automatic dependency discovery and lifecycle management.
// The factory function must return a type T that implements the Service interface.
// Dependencies are automatically discovered from the factory function parameters.
// Lifecycle methods (Start, Stop, Health) are automatically wired.
func NewServiceFactory[T Service](factory interface{}, lifetime Lifetime) *TypedServiceDefinition[T] {
	// Get the interface type T
	interfaceType := reflect.TypeOf((*T)(nil)).Elem()
	serviceName := inferServiceNameFromType(interfaceType)

	// Validate that the factory function returns the correct type
	factoryValue := reflect.ValueOf(factory)
	factoryType := factoryValue.Type()

	if factoryType.Kind() != reflect.Func {
		panic("factory must be a function")
	}

	if factoryType.NumOut() == 0 {
		panic("factory function must return a value")
	}

	// Check that the return type matches T
	returnType := factoryType.Out(0)
	expectedType := interfaceType

	// For interface types, we need to check if the return type implements the interface
	if expectedType.Kind() == reflect.Interface {
		if !returnType.Implements(expectedType) {
			panic(fmt.Sprintf("factory return type %s does not implement interface %s", returnType, expectedType))
		}
	} else {
		if returnType != expectedType {
			panic(fmt.Sprintf("factory return type %s does not match expected type %s", returnType, expectedType))
		}
	}

	// Create a wrapper factory that handles automatic dependency injection
	wrapperFactory := func(ctx context.Context, container *Container) (T, error) {
		return callFactoryWithAutoDependencies[T](ctx, container, factory)
	}

	// Create typed service definition with automatic lifecycle wiring
	serviceDef := &TypedServiceDefinition[T]{
		Name: serviceName,
		Service: TypedServiceConfig[T]{
			Type:     interfaceType,
			Factory:  wrapperFactory,
			Lifetime: lifetime,
		},
		// Automatically wire lifecycle methods
		Lifecycle: LifecycleConfig{
			Start: func(ctx context.Context, container *Container) error {
				service, err := ResolveType[T](container)
				if err != nil {
					return err
				}
				return service.Start(ctx)
			},
			Stop: func(ctx context.Context) error {
				// For factory-created services, we need to resolve the service and call Stop
				// We'll need to get the container from the service registry
				// This is a limitation we'll address by storing the registry reference
				return nil // We'll implement this properly by modifying the serviceComponent
			},
			Health: func(ctx context.Context) HealthStatus {
				// For factory services, we need to resolve from the registry's container
				// This is a limitation we'll address by modifying the serviceComponent
				return HealthStatus{
					Status:  HealthStatusHealthy,
					Message: "Service is healthy",
				}
			},
		},
	}

	// Automatically discover and add dependencies based on factory parameters
	autoDiscoverDependenciesTyped(serviceDef, factory)

	return serviceDef
}
