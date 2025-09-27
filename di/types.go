package di

import (
	"context"
	"reflect"
)

// Container represents a dependency injection container
type Container interface {
	// Register registers a service with the container
	Register(serviceType reflect.Type, factory Factory, options ...Option) error

	// RegisterInstance registers a service instance
	RegisterInstance(serviceType reflect.Type, instance interface{}) error

	// RegisterSingleton registers a singleton service
	RegisterSingleton(serviceType reflect.Type, factory Factory, options ...Option) error

	// Resolve resolves a service from the container
	Resolve(serviceType reflect.Type) (interface{}, error)

	// ResolveByName resolves a service by name
	ResolveByName(name string) (interface{}, error)

	// TryResolve attempts to resolve a service, returns false if not found
	TryResolve(serviceType reflect.Type) (interface{}, bool)

	// Contains checks if a service is registered
	Contains(serviceType reflect.Type) bool

	// ContainsByName checks if a named service is registered
	ContainsByName(name string) bool

	// GetRegistrations returns all service registrations
	GetRegistrations() []ServiceRegistration

	// CreateScope creates a new scope
	CreateScope() Scope

	// Dispose disposes the container and all its resources
	Dispose() error
}

// Scope represents a dependency injection scope
type Scope interface {
	// Resolve resolves a service within this scope
	Resolve(serviceType reflect.Type) (interface{}, error)

	// ResolveByName resolves a service by name within this scope
	ResolveByName(name string) (interface{}, error)

	// Dispose disposes the scope and all scoped instances
	Dispose() error
}

// Factory represents a factory function for creating service instances
type Factory func(ctx context.Context, container Container) (interface{}, error)

// ServiceLifetime represents the lifetime of a service
type ServiceLifetime int

const (
	// Transient creates a new instance every time
	Transient ServiceLifetime = iota

	// Singleton creates a single instance for the entire application
	Singleton

	// Scoped creates a single instance per scope
	Scoped
)

// ServiceRegistration represents a service registration
type ServiceRegistration struct {
	ServiceType reflect.Type
	Name        string
	Factory     Factory
	Instance    interface{}
	Lifetime    ServiceLifetime
	Options     ServiceOptions
}

// ServiceOptions holds options for service registration
type ServiceOptions struct {
	Name         string
	Tags         []string
	Dependencies []reflect.Type
	Interceptors []Interceptor
	Metadata     map[string]interface{}
}

// Option represents a service registration option
type Option func(*ServiceOptions)

// WithName sets the service name
func WithName(name string) Option {
	return func(o *ServiceOptions) {
		o.Name = name
	}
}

// WithTags sets service tags
func WithTags(tags ...string) Option {
	return func(o *ServiceOptions) {
		o.Tags = tags
	}
}

// WithDependencies specifies explicit dependencies
func WithDependencies(deps ...reflect.Type) Option {
	return func(o *ServiceOptions) {
		o.Dependencies = deps
	}
}

// WithInterceptors adds interceptors to the service
func WithInterceptors(interceptors ...Interceptor) Option {
	return func(o *ServiceOptions) {
		o.Interceptors = interceptors
	}
}

// WithMetadata adds metadata to the service registration
func WithMetadata(key string, value interface{}) Option {
	return func(o *ServiceOptions) {
		if o.Metadata == nil {
			o.Metadata = make(map[string]interface{})
		}
		o.Metadata[key] = value
	}
}

// Interceptor represents a service interceptor
type Interceptor interface {
	// Intercept intercepts service creation/resolution
	Intercept(ctx context.Context, serviceType reflect.Type, next func() (interface{}, error)) (interface{}, error)
}

// InterceptorFunc is a function adapter for Interceptor
type InterceptorFunc func(ctx context.Context, serviceType reflect.Type, next func() (interface{}, error)) (interface{}, error)

// Intercept implements the Interceptor interface
func (f InterceptorFunc) Intercept(ctx context.Context, serviceType reflect.Type, next func() (interface{}, error)) (interface{}, error) {
	return f(ctx, serviceType, next)
}

// ServiceProvider represents a service provider that can register services
type ServiceProvider interface {
	// RegisterServices registers services with the container
	RegisterServices(container Container) error

	// GetName returns the provider name
	GetName() string

	// GetDependencies returns provider dependencies
	GetDependencies() []string
}

// Module represents a DI module that groups related services
type Module interface {
	ServiceProvider

	// Configure configures the module
	Configure(config ModuleConfig) error

	// GetConfig returns the module configuration
	GetConfig() ModuleConfig
}

// ModuleConfig represents module configuration
type ModuleConfig struct {
	Name         string
	Enabled      bool
	Dependencies []string
	Settings     map[string]interface{}
}

// ContainerBuilder helps build and configure containers
type ContainerBuilder interface {
	// AddServiceProvider adds a service provider
	AddServiceProvider(provider ServiceProvider) ContainerBuilder

	// AddModule adds a module
	AddModule(module Module) ContainerBuilder

	// Configure configures the container
	Configure(config ContainerConfig) ContainerBuilder

	// Build builds the container
	Build() (Container, error)
}

// ContainerConfig represents container configuration
type ContainerConfig struct {
	EnableValidation    bool
	EnableCircularCheck bool
	EnableInterception  bool
	DefaultLifetime     ServiceLifetime
	MaxResolutionDepth  int
	EnableMetrics       bool
	MetricsProvider     MetricsProvider
}

// MetricsProvider provides metrics for DI operations
type MetricsProvider interface {
	// RecordResolution records a service resolution
	RecordResolution(serviceType reflect.Type, duration int64, success bool)

	// RecordRegistration records a service registration
	RecordRegistration(serviceType reflect.Type, lifetime ServiceLifetime)

	// GetMetrics returns current metrics
	GetMetrics() map[string]interface{}
}

// ServiceRegistry provides service discovery capabilities
type ServiceRegistry interface {
	// FindServices finds services by criteria
	FindServices(criteria ServiceCriteria) []ServiceRegistration

	// GetServicesByTag gets services by tag
	GetServicesByTag(tag string) []ServiceRegistration

	// GetServicesByType gets services by type
	GetServicesByType(serviceType reflect.Type) []ServiceRegistration
}

// ServiceCriteria represents criteria for service discovery
type ServiceCriteria struct {
	Type     reflect.Type
	Name     string
	Tags     []string
	Metadata map[string]interface{}
}

// HealthChecker provides health checking for services
type HealthChecker interface {
	// CheckHealth checks the health of a service
	CheckHealth(ctx context.Context, service interface{}) error
}

// Validator validates service registrations and resolutions
type Validator interface {
	// ValidateRegistration validates a service registration
	ValidateRegistration(registration ServiceRegistration) error

	// ValidateResolution validates a service resolution
	ValidateResolution(serviceType reflect.Type, instance interface{}) error
}
