// Package orchestrator provides a simple, unified API for application orchestration.
//
// This is the main entry point for the Go Orchestrator library. Most users
// should import this package directly:
//
//	import "github.com/AnasImloul/go-orchestrator"
//
// The library provides a declarative, less verbose API for building applications
// with dependency injection, lifecycle management, and orchestration.
package orchestrator

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/AnasImloul/go-orchestrator/internal/di"
	"github.com/AnasImloul/go-orchestrator/internal/lifecycle"
	"github.com/AnasImloul/go-orchestrator/internal/logger"
)

// ServiceRegistry represents the main service registry for dependency injection and lifecycle management.
// This is the single entry point for the library.
type ServiceRegistry struct {
	container        di.Container
	lifecycleManager lifecycle.LifecycleManager
	services         map[string]*ServiceDefinition
	config           Config
	logger           logger.Logger
	mu               sync.RWMutex
}

// Config holds configuration for the application.
type Config struct {
	StartupTimeout      time.Duration
	ShutdownTimeout     time.Duration
	HealthCheckInterval time.Duration
	EnableMetrics       bool
	EnableTracing       bool
	LogLevel            slog.Level
}

// DefaultConfig returns the default application configuration.
func DefaultConfig() Config {
	return Config{
		StartupTimeout:      30 * time.Second,
		ShutdownTimeout:     15 * time.Second,
		HealthCheckInterval: 30 * time.Second,
		EnableMetrics:       true,
		EnableTracing:       false,
		LogLevel:            slog.LevelInfo,
	}
}

// New creates a new application with the default configuration.
func New() *ServiceRegistry {
	return NewWithConfig(DefaultConfig())
}

// NewWithConfig creates a new service registry with the specified configuration.
func NewWithConfig(config Config) *ServiceRegistry {
	// Create logger
	logger := logger.NewSlogAdapter(slog.Default())

	// Create DI container
	diConfig := di.ContainerConfig{
		EnableValidation:    true,
		EnableCircularCheck: true,
		EnableInterception:  config.EnableTracing,
		DefaultLifetime:     di.Singleton,
		MaxResolutionDepth:  50,
		EnableMetrics:       config.EnableMetrics,
	}

	container := di.NewContainer(diConfig, logger)

	// Create lifecycle manager
	lifecycleManager := lifecycle.NewLifecycleManager(logger)

	return &ServiceRegistry{
		container:        container,
		lifecycleManager: lifecycleManager,
		services:         make(map[string]*ServiceDefinition),
		config:           config,
		logger:           logger,
	}
}

// Feature represents a declarative feature configuration.
type ServiceDefinition struct {
	Name         string
	Dependencies []string
	Services     []ServiceConfig
	Lifecycle    LifecycleConfig
	RetryConfig  *lifecycle.RetryConfig
	Metadata     map[string]string
}

// ServiceConfig represents a service registration configuration.
type ServiceConfig struct {
	Name     string
	Type     reflect.Type
	Factory  func(ctx context.Context, container *Container) (interface{}, error)
	Lifetime Lifetime
}

// Lifetime represents the service lifetime.
type Lifetime int

const (
	// Transient creates a new instance for each resolution
	Transient Lifetime = iota
	// Scoped creates one instance per scope
	Scoped
	// Singleton creates one instance for the entire container
	Singleton
)

// LifecycleConfig represents a service lifecycle configuration.
type LifecycleConfig struct {
	Start  func(ctx context.Context, container *Container) error
	Stop   func(ctx context.Context) error
	Health func(ctx context.Context) HealthStatus
}

// LifecycleBuilder provides a fluent interface for building service lifecycle configurations.
type LifecycleBuilder struct {
	config LifecycleConfig
}

// LifecycleBuilderFor provides a type-safe fluent API for building service lifecycle configurations.
type LifecycleBuilderFor[T any] struct {
	serviceRegistry *ServiceRegistry
	builder         *LifecycleBuilder
	serviceDef      *ServiceDefinition
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

// WithStartFor sets the start function for the service lifecycle with type-safe service resolution.
func (lbf *LifecycleBuilderFor[T]) WithStartFor(startFunc func(T) error) *LifecycleBuilderFor[T] {
	lbf.builder.WithStart(func(ctx context.Context, container *Container) error {
		service, err := ResolveType[T](container)
		if err != nil {
			return err
		}
		return startFunc(service)
	})
	return lbf
}

// WithStopFor sets the stop function for the service lifecycle with type-safe service resolution.
func (lbf *LifecycleBuilderFor[T]) WithStopFor(stopFunc func(T) error) *LifecycleBuilderFor[T] {
	lbf.builder.WithStop(func(ctx context.Context) error {
		service, err := ResolveType[T](lbf.serviceRegistry.Container())
		if err != nil {
			return err
		}
		return stopFunc(service)
	})
	return lbf
}

// WithHealthFor sets the health function for the service lifecycle with type-safe service resolution.
func (lbf *LifecycleBuilderFor[T]) WithHealthFor(healthFunc func(T) HealthStatus) *LifecycleBuilderFor[T] {
	lbf.builder.WithHealth(func(ctx context.Context) HealthStatus {
		service, err := ResolveType[T](lbf.serviceRegistry.Container())
		if err != nil {
			return HealthStatus{
				Status:  "unhealthy",
				Message: "Failed to resolve service",
			}
		}
		return healthFunc(service)
	})
	return lbf
}

// Build completes the lifecycle configuration and returns the service definition.
func (lbf *LifecycleBuilderFor[T]) Build() *ServiceDefinition {
	lbf.serviceDef.Lifecycle = lbf.builder.Build()
	return lbf.serviceDef
}

// HealthStatus represents the health status of a component.
type HealthStatus struct {
	Status  string
	Message string
	Details map[string]interface{}
}

// Container provides a simplified interface to the DI container.
type Container struct {
	container di.Container
	scope     di.Scope
}

// Register registers a service with the container.
func (c *Container) Register(serviceType reflect.Type, factory func(ctx context.Context, container *Container) (interface{}, error), lifetime Lifetime) error {
	// Convert public Lifetime to internal ServiceLifetime
	var internalLifetime di.ServiceLifetime
	switch lifetime {
	case Transient:
		internalLifetime = di.Transient
	case Scoped:
		internalLifetime = di.Scoped
	case Singleton:
		internalLifetime = di.Singleton
	default:
		internalLifetime = di.Singleton
	}

	// Register with the internal container using the proper lifetime
	return c.container.Register(serviceType, func(ctx context.Context, cont di.Container) (interface{}, error) {
		return factory(ctx, c)
	}, di.WithLifetime(internalLifetime))
}

// RegisterNamed registers a named service with the container.
func (c *Container) RegisterNamed(name string, serviceType reflect.Type, factory func(ctx context.Context, container *Container) (interface{}, error), lifetime Lifetime) error {
	// Convert public Lifetime to internal ServiceLifetime
	var internalLifetime di.ServiceLifetime
	switch lifetime {
	case Transient:
		internalLifetime = di.Transient
	case Scoped:
		internalLifetime = di.Scoped
	case Singleton:
		internalLifetime = di.Singleton
	default:
		internalLifetime = di.Singleton
	}

	// Register with the internal container using the proper lifetime and name
	return c.container.Register(serviceType, func(ctx context.Context, cont di.Container) (interface{}, error) {
		return factory(ctx, c)
	}, di.WithLifetime(internalLifetime), di.WithName(name))
}

// RegisterInstance registers a service instance.
func (c *Container) RegisterInstance(serviceType reflect.Type, instance interface{}) error {
	return c.container.RegisterInstance(serviceType, instance)
}

// RegisterNamedInstance registers a named service instance.
func (c *Container) RegisterNamedInstance(name string, serviceType reflect.Type, instance interface{}) error {
	return c.container.Register(serviceType, func(ctx context.Context, cont di.Container) (interface{}, error) {
		return instance, nil
	}, di.WithName(name), di.WithLifetime(di.Singleton))
}

// Resolve resolves a service from the container.
func (c *Container) Resolve(serviceType reflect.Type) (interface{}, error) {
	if c.scope != nil {
		return c.scope.Resolve(serviceType)
	}
	return c.container.Resolve(serviceType)
}

// ResolveByName resolves a service by name from the container.
func (c *Container) ResolveByName(name string) (interface{}, error) {
	if c.scope != nil {
		return c.scope.ResolveByName(name)
	}
	return c.container.ResolveByName(name)
}

// ResolveType resolves a service by interface type.
// T must be an interface type, not a concrete struct.
func ResolveType[T any](c *Container) (T, error) {
	var zero T
	serviceType := reflect.TypeOf((*T)(nil)).Elem()

	// Enforce that T is an interface type
	if serviceType.Kind() != reflect.Interface {
		return zero, fmt.Errorf("ResolveType[T] requires T to be an interface type, got %s", serviceType.Kind())
	}

	instance, err := c.Resolve(serviceType)
	if err != nil {
		return zero, err
	}
	return instance.(T), nil
}

// ResolveNamedType resolves a named service by interface type.
// T must be an interface type, not a concrete struct.
func ResolveNamedType[T any](c *Container, name string) (T, error) {
	var zero T
	serviceType := reflect.TypeOf((*T)(nil)).Elem()

	// Enforce that T is an interface type
	if serviceType.Kind() != reflect.Interface {
		return zero, fmt.Errorf("ResolveNamedType[T] requires T to be an interface type, got %s", serviceType.Kind())
	}

	instance, err := c.ResolveByName(name)
	if err != nil {
		return zero, err
	}

	return instance.(T), nil
}

// Component helpers for common patterns
func WithStartFunc[T any](fn func(T) error) func(context.Context, *Container) error {
	return func(ctx context.Context, container *Container) error {
		service, err := ResolveType[T](container)
		if err != nil {
			return err
		}
		return fn(service)
	}
}

func WithStopFuncWithService[T any](serviceRegistry *ServiceRegistry, fn func(T) error) func(context.Context) error {
	return func(ctx context.Context) error {
		svc, err := ResolveType[T](serviceRegistry.Container())
		if err != nil {
			return err
		}
		return fn(svc)
	}
}

func WithHealthFunc[T any](serviceRegistry *ServiceRegistry, fn func(T) HealthStatus) func(context.Context) HealthStatus {
	return func(ctx context.Context) HealthStatus {
		svc, err := ResolveType[T](serviceRegistry.Container())
		if err != nil {
			return HealthStatus{
				Status:  "unhealthy",
				Message: "Failed to resolve service",
			}
		}
		return fn(svc)
	}
}

// MustResolveType resolves a service by interface type, panicking on error.
// T must be an interface type, not a concrete struct.
func MustResolveType[T any](c *Container) T {
	instance, err := ResolveType[T](c)
	if err != nil {
		panic(err)
	}
	return instance
}

// Register registers a service definition in the service registry.
func (sr *ServiceRegistry) Register(serviceDef *ServiceDefinition) *ServiceRegistry {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if _, exists := sr.services[serviceDef.Name]; exists {
		panic(fmt.Sprintf("service %s is already registered", serviceDef.Name))
	}

	sr.services[serviceDef.Name] = serviceDef
	return sr
}

// Start starts the service registry.
func (sr *ServiceRegistry) Start(ctx context.Context) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	sr.logger.Info("Starting service registry")

	// Register all services first
	container := sr.Container()
	for _, serviceDef := range sr.services {
		for _, service := range serviceDef.Services {
			// All services now use factories for consistent behavior
			if service.Factory != nil {
				if service.Name != "" {
					// Register named service
					if err := container.RegisterNamed(service.Name, service.Type, service.Factory, service.Lifetime); err != nil {
						return fmt.Errorf("failed to register named service %s (%s): %w", service.Name, service.Type.String(), err)
					}
				} else {
					// Register unnamed service
					if err := container.Register(service.Type, service.Factory, service.Lifetime); err != nil {
						return fmt.Errorf("failed to register service %s: %w", service.Type.String(), err)
					}
				}
			} else {
				return fmt.Errorf("service %s has no factory - all services must use factory-based registration", service.Type.String())
			}
		}
	}

	// Register all service definitions as lifecycle components
	for name, serviceDef := range sr.services {
		component := &serviceComponent{
			serviceDef: serviceDef,
			serviceRegistry: sr,
		}

		if err := sr.lifecycleManager.RegisterComponent(component); err != nil {
			return fmt.Errorf("failed to register lifecycle component %s: %w", name, err)
		}
	}

	// Start the lifecycle manager
	return sr.lifecycleManager.Start(ctx)
}

// Stop stops the service registry.
func (sr *ServiceRegistry) Stop(ctx context.Context) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	sr.logger.Info("Stopping service registry")
	return sr.lifecycleManager.Stop(ctx)
}

// Health returns the health status of the service registry.
func (sr *ServiceRegistry) Health(ctx context.Context) map[string]HealthStatus {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	states := sr.lifecycleManager.GetAllComponentStates()
	health := make(map[string]HealthStatus)

	for name, state := range states {
		health[name] = HealthStatus{
			Status:  string(state.Health.Status),
			Message: state.Health.Message,
			Details: state.Health.Details,
		}
	}

	return health
}

// Container returns the DI container.
func (sr *ServiceRegistry) Container() *Container {
	return &Container{container: sr.container}
}

// CreateScope creates a new scope for the container.
func (c *Container) CreateScope() *Container {
	scope := c.container.CreateScope()
	return &Container{
		container: c.container,
		scope:     scope,
	}
}

// Dispose disposes the scope if one exists.
func (c *Container) Dispose() error {
	if c.scope != nil {
		return c.scope.Dispose()
	}
	return nil
}

// serviceComponent wraps a service definition as a lifecycle component.
type serviceComponent struct {
	serviceDef      *ServiceDefinition
	serviceRegistry *ServiceRegistry
}

func (c *serviceComponent) Name() string {
	return c.serviceDef.Name
}

func (c *serviceComponent) Dependencies() []string {
	return c.serviceDef.Dependencies
}

func (c *serviceComponent) Start(ctx context.Context) error {
	// Services are already registered in ServiceRegistry.Start()
	// Just start the lifecycle
	if c.serviceDef.Lifecycle.Start != nil {
		container := c.serviceRegistry.Container()
		return c.serviceDef.Lifecycle.Start(ctx, container)
	}

	return nil
}

func (c *serviceComponent) Stop(ctx context.Context) error {
	if c.serviceDef.Lifecycle.Stop != nil {
		return c.serviceDef.Lifecycle.Stop(ctx)
	}
	return nil
}

func (c *serviceComponent) Health(ctx context.Context) lifecycle.ComponentHealth {
	if c.serviceDef.Lifecycle.Health != nil {
		status := c.serviceDef.Lifecycle.Health(ctx)
		return lifecycle.ComponentHealth{
			Status:    lifecycle.HealthStatusHealthy, // Default to healthy
			Message:   status.Message,
			Details:   status.Details,
			Timestamp: time.Now(),
		}
	}

	return lifecycle.ComponentHealth{
		Status:    lifecycle.HealthStatusHealthy,
		Message:   "Service is healthy",
		Timestamp: time.Now(),
	}
}

func (c *serviceComponent) GetRetryConfig() *lifecycle.RetryConfig {
	return c.serviceDef.RetryConfig
}

// Helper functions for creating service definitions

// NewService creates a new service definition with the given name.
func NewService(name string) *ServiceDefinition {
	return &ServiceDefinition{
		Name:     name,
		Services: make([]ServiceConfig, 0),
		Metadata: make(map[string]string),
	}
}

// WithDependencies sets the dependencies for the service definition.
func (sd *ServiceDefinition) WithDependencies(deps ...string) *ServiceDefinition {
	sd.Dependencies = deps
	return sd
}

// WithLifetime sets the lifetime for the last registered service.
func (sd *ServiceDefinition) WithLifetime(lifetime Lifetime) *ServiceDefinition {
	if len(sd.Services) == 0 {
		panic("WithLifetime must be called after WithService")
	}
	// Update the lifetime of the last registered service
	sd.Services[len(sd.Services)-1].Lifetime = lifetime
	return sd
}


// WithServiceFactory adds a service factory to the service definition using generics.
// T must be an interface type that the factory returns.
func WithServiceFactory[T any](factory func(ctx context.Context, container *Container) (T, error)) func(*ServiceDefinition) *ServiceDefinition {
	return func(sd *ServiceDefinition) *ServiceDefinition {
		serviceType := reflect.TypeOf((*T)(nil)).Elem()

		// Create a wrapper factory that returns interface{}
		wrapperFactory := func(ctx context.Context, container *Container) (interface{}, error) {
			result, err := factory(ctx, container)
			if err != nil {
				return nil, err
			}
			return result, nil
		}

		sd.Services = append(sd.Services, ServiceConfig{
			Type:     serviceType,
			Factory:  wrapperFactory,
			Lifetime: Singleton, // Default to Singleton, can be overridden with WithLifetime
		})
		return sd
	}
}

// WithNamedService adds a named service to the service definition.
func (sd *ServiceDefinition) WithNamedService(name string, serviceType reflect.Type, factory func(ctx context.Context, container *Container) (interface{}, error), lifetime Lifetime) *ServiceDefinition {
	sd.Services = append(sd.Services, ServiceConfig{
		Name:     name,
		Type:     serviceType,
		Factory:  factory,
		Lifetime: lifetime,
	})
	return sd
}

// WithLifecycle sets the lifecycle configuration for the service definition using a builder.
func (sd *ServiceDefinition) WithLifecycle(builder *LifecycleBuilder) *ServiceDefinition {
	sd.Lifecycle = builder.Build()
	return sd
}

// WithLifecycleFor creates a lifecycle builder with type-safe lifecycle methods.
// This is a shorthand for common lifecycle patterns.
func WithLifecycleFor[T any](sd *ServiceDefinition, sr *ServiceRegistry) *LifecycleBuilderFor[T] {
	return &LifecycleBuilderFor[T]{
		serviceRegistry: sr,
		builder:         NewLifecycle(),
		serviceDef:      sd,
	}
}


// NewServiceWithFactory creates a new service definition with a service factory using the cleanest syntax.
// This is the most concise way to create a service definition with a factory.
func NewServiceWithFactory[T any](name string, factory func(ctx context.Context, container *Container) (T, error), lifetime Lifetime) *ServiceDefinition {
	return WithServiceFactory[T](factory)(NewService(name)).
		WithLifetime(lifetime)
}

// NewServiceWithAutoFactory creates a new service definition with automatic dependency injection.
// The factory function should only take the dependencies as parameters, and they will be automatically resolved.
// Dependencies are also automatically discovered and added to the dependency graph.
func NewServiceWithAutoFactory[T any](name string, factory interface{}, lifetime Lifetime) *ServiceDefinition {
	// Create a wrapper factory that handles automatic dependency injection
	wrapperFactory := func(ctx context.Context, container *Container) (T, error) {
		return callFactoryWithAutoDependencies[T](ctx, container, factory)
	}
	
	// Create the service definition with the wrapper factory
	serviceDef := NewServiceWithFactory(name, wrapperFactory, lifetime)
	
	// Automatically discover and add dependencies based on factory parameters
	autoDiscoverDependencies(serviceDef, factory)
	
	return serviceDef
}

// NewServiceWithInstance creates a new service definition with a service instance by wrapping it in a factory.
// This is the recommended way to register simple services without dependencies.
func NewServiceWithInstance[T any](name string, instance T, lifetime Lifetime) *ServiceDefinition {
	// Create a factory that returns the instance
	factory := func(ctx context.Context, container *Container) (T, error) {
		// For Transient services, create a new instance by cloning
		if lifetime == Transient {
			return cloneInstance(instance).(T), nil
		}
		// For Singleton and Scoped, return the same instance
		return instance, nil
	}
	return NewServiceWithFactory(name, factory, lifetime)
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
	sd.Metadata[key] = value
	return sd
}

// autoDiscoverDependencies analyzes the factory function parameters and automatically
// adds them as dependencies to the service definition.
func autoDiscoverDependencies(serviceDef *ServiceDefinition, factory interface{}) {
	factoryValue := reflect.ValueOf(factory)
	factoryType := factoryValue.Type()
	
	// Check if it's a function
	if factoryType.Kind() != reflect.Func {
		return
	}
	
	// Get the number of parameters
	numIn := factoryType.NumIn()
	dependencies := make([]string, 0, numIn)
	
	// Analyze each parameter to determine dependency names
	for i := 0; i < numIn; i++ {
		paramType := factoryType.In(i)
		
		// Convert type to a dependency name
		// For interfaces, we use a simplified name based on the type
		dependencyName := typeToDependencyName(paramType)
		if dependencyName != "" {
			dependencies = append(dependencies, dependencyName)
		}
	}
	
	// Add the discovered dependencies to the service definition
	if len(dependencies) > 0 {
		serviceDef.Dependencies = append(serviceDef.Dependencies, dependencies...)
	}
}

// typeToDependencyName converts a Go type to a dependency name.
// This is a simple heuristic - in a real implementation, you might want
// to use type tags or a more sophisticated mapping.
func typeToDependencyName(paramType reflect.Type) string {
	typeName := paramType.String()
	
	// Remove pointer prefix if present
	if paramType.Kind() == reflect.Ptr {
		typeName = paramType.Elem().String()
	}
	
	// Convert common patterns to dependency names
	switch {
	case strings.Contains(typeName, "DatabaseService"):
		return "database"
	case strings.Contains(typeName, "CacheService"):
		return "cache"
	case strings.Contains(typeName, "LoggerService"):
		return "logger"
	case strings.Contains(typeName, "APIService"):
		return "api"
	case strings.Contains(typeName, "MetricsService"):
		return "metrics"
	case strings.Contains(typeName, "LoggingService"):
		return "logging"
	default:
		// For unknown types, try to extract a reasonable name
		// Remove package prefix and "Service" suffix
		parts := strings.Split(typeName, ".")
		if len(parts) > 0 {
			name := parts[len(parts)-1]
			if strings.HasSuffix(name, "Service") {
				return strings.ToLower(strings.TrimSuffix(name, "Service"))
			}
			return strings.ToLower(name)
		}
		return ""
	}
}

// callFactoryWithAutoDependencies uses reflection to automatically resolve dependencies
// and call the factory function with the resolved dependencies.
func callFactoryWithAutoDependencies[T any](ctx context.Context, container *Container, factory interface{}) (T, error) {
	var zero T
	
	factoryValue := reflect.ValueOf(factory)
	factoryType := factoryValue.Type()
	
	// Check if it's a function
	if factoryType.Kind() != reflect.Func {
		return zero, fmt.Errorf("factory must be a function, got %s", factoryType.Kind())
	}
	
	// Get the number of parameters
	numIn := factoryType.NumIn()
	args := make([]reflect.Value, numIn)
	
	// Resolve each parameter as a dependency
	for i := 0; i < numIn; i++ {
		paramType := factoryType.In(i)
		
		// Try to resolve the dependency from the container
		dependency, err := container.Resolve(paramType)
		if err != nil {
			return zero, fmt.Errorf("failed to resolve dependency %d of type %s: %w", i, paramType.String(), err)
		}
		
		args[i] = reflect.ValueOf(dependency)
	}
	
	// Call the factory function
	results := factoryValue.Call(args)
	
	// Handle the return values
	if len(results) == 0 {
		return zero, fmt.Errorf("factory function must return at least one value")
	}
	
	// Check if the first result is an error
	if len(results) > 1 {
		if err, ok := results[1].Interface().(error); ok && err != nil {
			return zero, err
		}
	}
	
	// Return the first result
	return results[0].Interface().(T), nil
}

// resolveDependenciesAndCallFactory uses reflection to automatically resolve dependencies
// and call the original factory function with the resolved dependencies.
func resolveDependenciesAndCallFactory(ctx context.Context, container *Container, factory interface{}) (interface{}, error) {
	factoryValue := reflect.ValueOf(factory)
	factoryType := factoryValue.Type()
	
	// Check if it's a function
	if factoryType.Kind() != reflect.Func {
		return nil, fmt.Errorf("factory must be a function, got %s", factoryType.Kind())
	}
	
	// Get the number of parameters (should be 2: context.Context and *Container)
	numIn := factoryType.NumIn()
	if numIn != 2 {
		return nil, fmt.Errorf("factory function must have exactly 2 parameters (context.Context, *Container), got %d", numIn)
	}
	
	// Verify the first parameter is context.Context
	if !factoryType.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		return nil, fmt.Errorf("first parameter must be context.Context")
	}
	
	// Verify the second parameter is *Container
	if factoryType.In(1) != reflect.TypeOf((*Container)(nil)) {
		return nil, fmt.Errorf("second parameter must be *Container")
	}
	
	// Call the original factory with the provided parameters
	args := []reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(container),
	}
	
	results := factoryValue.Call(args)
	
	// Handle the return values (should be (interface{}, error))
	if len(results) != 2 {
		return nil, fmt.Errorf("factory function must return exactly 2 values (interface{}, error), got %d", len(results))
	}
	
	// Check for error
	if !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}
	
	return results[0].Interface(), nil
}

// cloneInstance creates a deep copy of an instance using reflection
func cloneInstance(original interface{}) interface{} {
	if original == nil {
		return nil
	}

	originalValue := reflect.ValueOf(original)
	originalType := originalValue.Type()

	// Handle pointers
	if originalType.Kind() == reflect.Ptr {
		if originalValue.IsNil() {
			return reflect.New(originalType.Elem()).Interface()
		}
		// Create a new pointer to the same type
		newPtr := reflect.New(originalType.Elem())
		// Recursively copy the value that the original pointer points to
		newPtr.Elem().Set(reflect.ValueOf(cloneInstance(originalValue.Elem().Interface())))
		return newPtr.Interface()
	}

	// Handle structs
	if originalType.Kind() == reflect.Struct {
		// Create a new struct of the same type
		newStruct := reflect.New(originalType)
		// Copy all fields recursively
		for i := 0; i < originalValue.NumField(); i++ {
			if newStruct.Elem().Field(i).CanSet() {
				fieldValue := originalValue.Field(i)
				if fieldValue.CanInterface() {
					clonedField := cloneInstance(fieldValue.Interface())
					newStruct.Elem().Field(i).Set(reflect.ValueOf(clonedField))
				} else {
					newStruct.Elem().Field(i).Set(fieldValue)
				}
			}
		}
		return newStruct.Elem().Interface()
	}

	// Handle slices
	if originalType.Kind() == reflect.Slice {
		if originalValue.IsNil() {
			return reflect.MakeSlice(originalType, 0, 0).Interface()
		}
		newSlice := reflect.MakeSlice(originalType, originalValue.Len(), originalValue.Cap())
		for i := 0; i < originalValue.Len(); i++ {
			clonedElement := cloneInstance(originalValue.Index(i).Interface())
			newSlice.Index(i).Set(reflect.ValueOf(clonedElement))
		}
		return newSlice.Interface()
	}

	// Handle maps
	if originalType.Kind() == reflect.Map {
		if originalValue.IsNil() {
			return reflect.MakeMap(originalType).Interface()
		}
		newMap := reflect.MakeMap(originalType)
		for _, key := range originalValue.MapKeys() {
			clonedKey := cloneInstance(key.Interface())
			clonedValue := cloneInstance(originalValue.MapIndex(key).Interface())
			newMap.SetMapIndex(reflect.ValueOf(clonedKey), reflect.ValueOf(clonedValue))
		}
		return newMap.Interface()
	}

	// Handle arrays
	if originalType.Kind() == reflect.Array {
		newArray := reflect.New(originalType).Elem()
		for i := 0; i < originalValue.Len(); i++ {
			clonedElement := cloneInstance(originalValue.Index(i).Interface())
			newArray.Index(i).Set(reflect.ValueOf(clonedElement))
		}
		return newArray.Interface()
	}

	// For primitive types (int, string, bool, etc.), return the value directly
	// These are already copied by value in Go
	return original
}
