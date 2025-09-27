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
	Type      reflect.Type
	Factory   func(ctx context.Context, container *Container) (interface{}, error)
	Instance  interface{}
	Lifetime  Lifetime
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
	Start func(ctx context.Context, container *Container) error
	Stop  func(ctx context.Context) error
	Health func(ctx context.Context) HealthStatus
}

// ComponentBuilder provides a fluent interface for building component configurations.
type ComponentBuilder struct {
	config ComponentConfig
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

// HealthStatus represents the health status of a component.
type HealthStatus struct {
	Status  string
	Message string
	Details map[string]interface{}
}

// Container provides a simplified interface to the DI container.
type Container struct {
	container di.Container
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

// RegisterInstance registers a service instance.
func (c *Container) RegisterInstance(serviceType reflect.Type, instance interface{}) error {
	return c.container.RegisterInstance(serviceType, instance)
}

// Resolve resolves a service from the container.
func (c *Container) Resolve(serviceType reflect.Type) (interface{}, error) {
	return c.container.Resolve(serviceType)
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
	
	instance, err := c.container.Resolve(serviceType)
	if err != nil {
		return zero, err
	}
	return instance.(T), nil
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
			if service.Factory != nil {
				if err := container.Register(service.Type, service.Factory, service.Lifetime); err != nil {
					return fmt.Errorf("failed to register service %s: %w", service.Type.String(), err)
				}
			} else if service.Instance != nil {
				// For non-singleton lifetimes, create a factory that returns the instance
				// For singleton, we can use RegisterInstance directly
				if service.Lifetime == Singleton {
					if err := container.RegisterInstance(service.Type, service.Instance); err != nil {
						return fmt.Errorf("failed to register service instance %s: %w", service.Type.String(), err)
					}
				} else {
					// For Transient and Scoped, create a factory that returns new instances
					originalInstance := service.Instance
					factory := func(ctx context.Context, container *Container) (interface{}, error) {
						if service.Lifetime == Transient {
							// For Transient, create a new instance by cloning the original
							return cloneInstance(originalInstance), nil
						} else {
							// For Scoped, return the same instance (for now, until proper scoping is implemented)
							return originalInstance, nil
						}
					}
					if err := container.Register(service.Type, factory, service.Lifetime); err != nil {
						return fmt.Errorf("failed to register service factory %s: %w", service.Type.String(), err)
					}
				}
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


// WithService adds a service to the feature.
func (f *Feature) WithService(serviceType reflect.Type, factory func(ctx context.Context, container *Container) (interface{}, error), lifetime Lifetime) *Feature {
	f.Services = append(f.Services, ServiceConfig{
		Type:     serviceType,
		Factory:  factory,
		Lifetime: lifetime,
	})
	return f
}

// WithServiceInstance adds a service instance to the feature.
func (f *Feature) WithServiceInstance(serviceType reflect.Type, instance interface{}) *Feature {
	f.Services = append(f.Services, ServiceConfig{
		Type:     serviceType,
		Instance: instance,
		Lifetime: Singleton,
	})
	return f
}


// WithServiceInstanceT is a helper function that creates a feature with a service instance using generics.
func WithServiceInstanceT[T any](name string, instance interface{}, lifetime Lifetime) *Feature {
	return WithServiceInstanceGeneric[T](instance, lifetime)(NewFeature(name))
}

// WithServiceInstanceGeneric adds a service instance to the feature using generics.
// T must be an interface type that the instance implements.
func WithServiceInstanceGeneric[T any](instance interface{}, lifetime Lifetime) func(*Feature) *Feature {
	return func(f *Feature) *Feature {
		serviceType := reflect.TypeOf((*T)(nil)).Elem()
		
		// Verify that the instance implements the interface T
		if !reflect.TypeOf(instance).Implements(serviceType) {
			panic(fmt.Sprintf("instance of type %T does not implement interface %s", instance, serviceType.String()))
		}
		
		f.Services = append(f.Services, ServiceConfig{
			Type:     serviceType,
			Instance: instance,
			Lifetime: lifetime,
		})
		return f
	}
}

// WithComponent sets the component configuration for the feature using a builder.
func (f *Feature) WithComponent(builder *ComponentBuilder) *Feature {
	f.Component = builder.Build()
	return f
}

// WithRetryConfig sets the retry configuration for the feature.
func (f *Feature) WithRetryConfig(config *lifecycle.RetryConfig) *Feature {
	f.RetryConfig = config
	return f
}

// WithMetadata adds metadata to the feature.
func (f *Feature) WithMetadata(key, value string) *Feature {
	f.Metadata[key] = value
	return f
}

// cloneInstance creates a shallow copy of an instance using reflection
func cloneInstance(original interface{}) interface{} {
	if original == nil {
		return nil
	}
	
	originalValue := reflect.ValueOf(original)
	originalType := originalValue.Type()
	
	// Handle pointers
	if originalType.Kind() == reflect.Ptr {
		// Create a new pointer to the same type
		newPtr := reflect.New(originalType.Elem())
		// Copy the value that the original pointer points to
		newPtr.Elem().Set(originalValue.Elem())
		return newPtr.Interface()
	}
	
	// Handle structs
	if originalType.Kind() == reflect.Struct {
		// Create a new struct of the same type
		newStruct := reflect.New(originalType)
		// Copy all fields
		for i := 0; i < originalValue.NumField(); i++ {
			if newStruct.Elem().Field(i).CanSet() {
				newStruct.Elem().Field(i).Set(originalValue.Field(i))
			}
		}
		return newStruct.Elem().Interface()
	}
	
	// For other types, return a copy if possible
	return original
}
