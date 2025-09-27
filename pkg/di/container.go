package di

import (
	"context"
	"fmt"
	"reflect"

	"github.com/AnasImloul/go-orchestrator/internal/di"
)

// containerWrapper wraps the internal container to provide a public API.
type containerWrapper struct {
	internal di.Container
}

// NewContainer creates a new DI container with default configuration.
func NewContainer() Container {
	internalConfig := di.ContainerConfig{
		EnableValidation:    true,
		EnableCircularCheck: true,
		EnableInterception:  false,
		DefaultLifetime:     di.Singleton,
		MaxResolutionDepth:  50,
		EnableMetrics:       false,
	}

	internalContainer := di.NewContainer(internalConfig, nil)
	return &containerWrapper{internal: internalContainer}
}

// Register registers a service with the container.
func (c *containerWrapper) Register(serviceType reflect.Type, factory Factory, options ...Option) error {
	internalFactory := func(ctx context.Context, container di.Container) (interface{}, error) {
		return factory(ctx, c)
	}

	var internalOptions []di.Option
	for _, opt := range options {
		internalOptions = append(internalOptions, &internalOptionWrapper{opt})
	}

	return c.internal.Register(serviceType, internalFactory, internalOptions...)
}

// RegisterInstance registers a service instance.
func (c *containerWrapper) RegisterInstance(serviceType reflect.Type, instance interface{}) error {
	return c.internal.RegisterInstance(serviceType, instance)
}

// RegisterSingleton registers a singleton service.
func (c *containerWrapper) RegisterSingleton(serviceType reflect.Type, factory Factory, options ...Option) error {
	internalFactory := func(ctx context.Context, container di.Container) (interface{}, error) {
		return factory(ctx, c)
	}

	var internalOptions []di.Option
	for _, opt := range options {
		internalOptions = append(internalOptions, &internalOptionWrapper{opt})
	}

	return c.internal.RegisterSingleton(serviceType, internalFactory, internalOptions...)
}

// Resolve resolves a service from the container.
func (c *containerWrapper) Resolve(serviceType reflect.Type) (interface{}, error) {
	return c.internal.Resolve(serviceType)
}

// CreateScope creates a new scope.
func (c *containerWrapper) CreateScope() Scope {
	internalScope := c.internal.CreateScope()
	return &scopeWrapper{internal: internalScope}
}

// Validate validates the container configuration.
func (c *containerWrapper) Validate() error {
	// The internal container doesn't have a Validate method, so we'll just return nil
	return nil
}

// scopeWrapper wraps the internal scope to provide a public API.
type scopeWrapper struct {
	internal di.Scope
}

// Resolve resolves a service within this scope.
func (s *scopeWrapper) Resolve(serviceType reflect.Type) (interface{}, error) {
	return s.internal.Resolve(serviceType)
}

// Dispose disposes the scope and all its resources.
func (s *scopeWrapper) Dispose() error {
	return s.internal.Dispose()
}

// internalOptionWrapper wraps public options for internal use.
type internalOptionWrapper struct {
	option Option
}

// Apply applies the option to the internal registration.
func (w *internalOptionWrapper) Apply(registration *di.ServiceRegistration) {
	// Convert internal registration to public registration
	publicReg := &ServiceRegistration{
		ServiceType:  registration.ServiceType,
		Factory:      nil, // Will be set by the container
		Lifetime:     Lifetime(registration.Lifetime),
		Options:      nil, // Not needed for internal use
		Interceptors: nil, // Not needed for internal use
	}

	// Apply the public option
	w.option.Apply(publicReg)

	// Convert back to internal registration
	registration.Lifetime = di.Lifetime(publicReg.Lifetime)
}
