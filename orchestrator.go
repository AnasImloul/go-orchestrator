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

// App represents the main application orchestrator.
// This is the single entry point for the library.
type App struct {
	container        di.Container
	lifecycleManager lifecycle.LifecycleManager
	features         map[string]*Feature
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
func New() *App {
	return NewWithConfig(DefaultConfig())
}

// NewWithConfig creates a new application with the specified configuration.
func NewWithConfig(config Config) *App {
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

	return &App{
		container:        container,
		lifecycleManager: lifecycleManager,
		features:         make(map[string]*Feature),
		config:           config,
		logger:           logger,
	}
}

// Feature represents a declarative feature configuration.
type Feature struct {
	Name         string
	Dependencies []string
	Services     []ServiceConfig
	Component    ComponentConfig
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

// ComponentConfig represents a component configuration.
type ComponentConfig struct {
	Start  func(ctx context.Context, container *Container) error
	Stop   func(ctx context.Context) error
	Health func(ctx context.Context) HealthStatus
}

// ComponentBuilder provides a fluent interface for building component configurations.
type ComponentBuilder struct {
	config ComponentConfig
}

// ComponentBuilderFor provides a type-safe fluent API for building component configurations.
type ComponentBuilderFor[T any] struct {
	app     *App
	builder *ComponentBuilder
	feature *Feature
}

// NewComponent creates a new component builder.
func NewComponent() *ComponentBuilder {
	return &ComponentBuilder{
		config: ComponentConfig{},
	}
}

// WithStart sets the start function for the component.
func (cb *ComponentBuilder) WithStart(start func(ctx context.Context, container *Container) error) *ComponentBuilder {
	cb.config.Start = start
	return cb
}

// WithStop sets the stop function for the component.
func (cb *ComponentBuilder) WithStop(stop func(ctx context.Context) error) *ComponentBuilder {
	cb.config.Stop = stop
	return cb
}

// WithHealth sets the health check function for the component.
func (cb *ComponentBuilder) WithHealth(health func(ctx context.Context) HealthStatus) *ComponentBuilder {
	cb.config.Health = health
	return cb
}

// Build returns the component configuration.
func (cb *ComponentBuilder) Build() ComponentConfig {
	return cb.config
}

// WithStartFor sets the start function for the component with type-safe service resolution.
func (cbf *ComponentBuilderFor[T]) WithStartFor(startFunc func(T) error) *ComponentBuilderFor[T] {
	cbf.builder.WithStart(func(ctx context.Context, container *Container) error {
		service, err := ResolveType[T](container)
		if err != nil {
			return err
		}
		return startFunc(service)
	})
	return cbf
}

// WithStopFor sets the stop function for the component with type-safe service resolution.
func (cbf *ComponentBuilderFor[T]) WithStopFor(stopFunc func(T) error) *ComponentBuilderFor[T] {
	cbf.builder.WithStop(func(ctx context.Context) error {
		service, err := ResolveType[T](cbf.app.Container())
		if err != nil {
			return err
		}
		return stopFunc(service)
	})
	return cbf
}

// WithHealthFor sets the health function for the component with type-safe service resolution.
func (cbf *ComponentBuilderFor[T]) WithHealthFor(healthFunc func(T) HealthStatus) *ComponentBuilderFor[T] {
	cbf.builder.WithHealth(func(ctx context.Context) HealthStatus {
		service, err := ResolveType[T](cbf.app.Container())
		if err != nil {
			return HealthStatus{
				Status:  "unhealthy",
				Message: "Failed to resolve service",
			}
		}
		return healthFunc(service)
	})
	return cbf
}

// Build completes the component configuration and returns the feature.
func (cbf *ComponentBuilderFor[T]) Build() *Feature {
	cbf.feature.Component = cbf.builder.Build()
	return cbf.feature
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

func WithStopFuncWithApp[T any](app *App, fn func(T) error) func(context.Context) error {
	return func(ctx context.Context) error {
		service, err := ResolveType[T](app.Container())
		if err != nil {
			return err
		}
		return fn(service)
	}
}

func WithHealthFunc[T any](app *App, fn func(T) HealthStatus) func(context.Context) HealthStatus {
	return func(ctx context.Context) HealthStatus {
		service, err := ResolveType[T](app.Container())
		if err != nil {
			return HealthStatus{
				Status:  "unhealthy",
				Message: "Failed to resolve service",
			}
		}
		return fn(service)
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

// AddFeature adds a feature to the application.
func (a *App) AddFeature(feature *Feature) *App {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.features[feature.Name]; exists {
		panic(fmt.Sprintf("feature %s is already registered", feature.Name))
	}

	a.features[feature.Name] = feature
	return a
}

// Start starts the application.
func (a *App) Start(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.logger.Info("Starting application")

	// Register all services first
	container := a.Container()
	for _, feature := range a.features {
		for _, service := range feature.Services {
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

	// Register all features as components
	for name, feature := range a.features {
		component := &featureComponent{
			feature: feature,
			app:     a,
		}

		if err := a.lifecycleManager.RegisterComponent(component); err != nil {
			return fmt.Errorf("failed to register component %s: %w", name, err)
		}
	}

	// Start the lifecycle manager
	return a.lifecycleManager.Start(ctx)
}

// Stop stops the application.
func (a *App) Stop(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.logger.Info("Stopping application")
	return a.lifecycleManager.Stop(ctx)
}

// Health returns the health status of the application.
func (a *App) Health(ctx context.Context) map[string]HealthStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()

	states := a.lifecycleManager.GetAllComponentStates()
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
func (a *App) Container() *Container {
	return &Container{container: a.container}
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

// featureComponent wraps a feature as a lifecycle component.
type featureComponent struct {
	feature *Feature
	app     *App
}

func (c *featureComponent) Name() string {
	return c.feature.Name
}

func (c *featureComponent) Dependencies() []string {
	return c.feature.Dependencies
}

func (c *featureComponent) Start(ctx context.Context) error {
	// Services are already registered in App.Start()
	// Just start the component
	if c.feature.Component.Start != nil {
		container := c.app.Container()
		return c.feature.Component.Start(ctx, container)
	}

	return nil
}

func (c *featureComponent) Stop(ctx context.Context) error {
	if c.feature.Component.Stop != nil {
		return c.feature.Component.Stop(ctx)
	}
	return nil
}

func (c *featureComponent) Health(ctx context.Context) lifecycle.ComponentHealth {
	if c.feature.Component.Health != nil {
		status := c.feature.Component.Health(ctx)
		return lifecycle.ComponentHealth{
			Status:    lifecycle.HealthStatusHealthy, // Default to healthy
			Message:   status.Message,
			Details:   status.Details,
			Timestamp: time.Now(),
		}
	}

	return lifecycle.ComponentHealth{
		Status:    lifecycle.HealthStatusHealthy,
		Message:   "Component is healthy",
		Timestamp: time.Now(),
	}
}

func (c *featureComponent) GetRetryConfig() *lifecycle.RetryConfig {
	return c.feature.RetryConfig
}

// Helper functions for creating features

// NewFeature creates a new feature with the given name.
func NewFeature(name string) *Feature {
	return &Feature{
		Name:     name,
		Services: make([]ServiceConfig, 0),
		Metadata: make(map[string]string),
	}
}

// WithDependencies sets the dependencies for the feature.
func (f *Feature) WithDependencies(deps ...string) *Feature {
	f.Dependencies = deps
	return f
}

// WithLifetime sets the lifetime for the last registered service.
func (f *Feature) WithLifetime(lifetime Lifetime) *Feature {
	if len(f.Services) == 0 {
		panic("WithLifetime must be called after WithService")
	}
	// Update the lifetime of the last registered service
	f.Services[len(f.Services)-1].Lifetime = lifetime
	return f
}


// WithServiceFactory adds a service factory to the feature using generics.
// T must be an interface type that the factory returns.
func WithServiceFactory[T any](factory func(ctx context.Context, container *Container) (T, error)) func(*Feature) *Feature {
	return func(f *Feature) *Feature {
		serviceType := reflect.TypeOf((*T)(nil)).Elem()

		// Create a wrapper factory that returns interface{}
		wrapperFactory := func(ctx context.Context, container *Container) (interface{}, error) {
			result, err := factory(ctx, container)
			if err != nil {
				return nil, err
			}
			return result, nil
		}

		f.Services = append(f.Services, ServiceConfig{
			Type:     serviceType,
			Factory:  wrapperFactory,
			Lifetime: Singleton, // Default to Singleton, can be overridden with WithLifetime
		})
		return f
	}
}

// WithNamedService adds a named service to the feature.
func (f *Feature) WithNamedService(name string, serviceType reflect.Type, factory func(ctx context.Context, container *Container) (interface{}, error), lifetime Lifetime) *Feature {
	f.Services = append(f.Services, ServiceConfig{
		Name:     name,
		Type:     serviceType,
		Factory:  factory,
		Lifetime: lifetime,
	})
	return f
}

// WithComponent sets the component configuration for the feature using a builder.
func (f *Feature) WithComponent(builder *ComponentBuilder) *Feature {
	f.Component = builder.Build()
	return f
}

// WithComponentFor creates a component builder with type-safe lifecycle methods.
// This is a shorthand for common component patterns.
func WithComponentFor[T any](f *Feature, app *App) *ComponentBuilderFor[T] {
	return &ComponentBuilderFor[T]{
		app:      app,
		builder:  NewComponent(),
		feature:  f,
	}
}


// NewFeatureWithFactory creates a new feature with a service factory using the cleanest syntax.
// This is the most concise way to create a feature with a factory.
func NewFeatureWithFactory[T any](name string, factory func(ctx context.Context, container *Container) (T, error), lifetime Lifetime) *Feature {
	return WithServiceFactory[T](factory)(NewFeature(name)).
		WithLifetime(lifetime)
}

// NewFeatureWithAutoFactory creates a new feature with automatic dependency injection.
// The factory function should only take the dependencies as parameters, and they will be automatically resolved.
// Dependencies are also automatically discovered and added to the dependency graph.
func NewFeatureWithAutoFactory[T any](name string, factory interface{}, lifetime Lifetime) *Feature {
	// Create a wrapper factory that handles automatic dependency injection
	wrapperFactory := func(ctx context.Context, container *Container) (T, error) {
		return callFactoryWithAutoDependencies[T](ctx, container, factory)
	}
	
	// Create the feature with the wrapper factory
	feature := NewFeatureWithFactory(name, wrapperFactory, lifetime)
	
	// Automatically discover and add dependencies based on factory parameters
	autoDiscoverDependencies(feature, factory)
	
	return feature
}

// NewFeatureWithInstance creates a new feature with a service instance by wrapping it in a factory.
// This is the recommended way to register simple services without dependencies.
func NewFeatureWithInstance[T any](name string, instance T, lifetime Lifetime) *Feature {
	// Create a factory that returns the instance
	factory := func(ctx context.Context, container *Container) (T, error) {
		// For Transient services, create a new instance by cloning
		if lifetime == Transient {
			return cloneInstance(instance).(T), nil
		}
		// For Singleton and Scoped, return the same instance
		return instance, nil
	}
	return NewFeatureWithFactory(name, factory, lifetime)
}

// WithRetryConfig sets the retry configuration for the feature.
func (f *Feature) WithRetryConfig(config *lifecycle.RetryConfig) *Feature {
	f.RetryConfig = config
	return f
}

// WithAutoDependencies enables automatic dependency discovery for the last registered service.
// This will scan the factory function parameters and automatically resolve dependencies.
func (f *Feature) WithAutoDependencies() *Feature {
	if len(f.Services) == 0 {
		return f
	}
	
	lastService := &f.Services[len(f.Services)-1]
	if lastService.Factory == nil {
		return f
	}
	
	// Create a wrapper factory that automatically resolves dependencies
	originalFactory := lastService.Factory
	lastService.Factory = func(ctx context.Context, container *Container) (interface{}, error) {
		// Use reflection to analyze the original factory function and resolve dependencies
		return resolveDependenciesAndCallFactory(ctx, container, originalFactory)
	}
	
	return f
}

// WithMetadata adds metadata to the feature.
func (f *Feature) WithMetadata(key, value string) *Feature {
	f.Metadata[key] = value
	return f
}

// autoDiscoverDependencies analyzes the factory function parameters and automatically
// adds them as dependencies to the feature.
func autoDiscoverDependencies(feature *Feature, factory interface{}) {
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
	
	// Add the discovered dependencies to the feature
	if len(dependencies) > 0 {
		feature.Dependencies = append(feature.Dependencies, dependencies...)
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
