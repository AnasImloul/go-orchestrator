package di

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/AnasImloul/go-orchestrator/internal/di"
	"github.com/AnasImloul/go-orchestrator/internal/logger"
)

// ContainerConfig holds configuration for the DI container
type ContainerConfig struct {
	EnableValidation    bool
	EnableCircularCheck bool
	EnableInterception  bool
	DefaultLifetime     Lifetime
	MaxResolutionDepth  int
	EnableMetrics       bool
}

// PublicContainer wraps the internal container to provide a public API
type PublicContainer struct {
	internal *di.DefaultContainer
}

// NewContainer creates a new DI container
func NewContainer(config ContainerConfig, logger logger.Logger) Container {
	internalConfig := di.ContainerConfig{
		EnableValidation:    config.EnableValidation,
		EnableCircularCheck: config.EnableCircularCheck,
		EnableInterception:  config.EnableInterception,
		DefaultLifetime:     di.Lifetime(config.DefaultLifetime),
		MaxResolutionDepth:  config.MaxResolutionDepth,
		EnableMetrics:       config.EnableMetrics,
	}

	internalContainer := di.NewContainer(internalConfig, logger)
	return &PublicContainer{internal: internalContainer}
}

// Register registers a service with the container
func (c *PublicContainer) Register(serviceType reflect.Type, factory Factory, options ...Option) error {
	internalFactory := func(ctx context.Context, container di.Container) (interface{}, error) {
		return factory(ctx, c)
	}

	var internalOptions []di.Option
	for _, opt := range options {
		internalOptions = append(internalOptions, &internalOptionWrapper{opt})
	}

	return c.internal.Register(serviceType, internalFactory, internalOptions...)
}

// RegisterInstance registers a service instance
func (c *PublicContainer) RegisterInstance(serviceType reflect.Type, instance interface{}) error {
	return c.internal.RegisterInstance(serviceType, instance)
}

// RegisterSingleton registers a singleton service
func (c *PublicContainer) RegisterSingleton(serviceType reflect.Type, factory Factory, options ...Option) error {
	internalFactory := func(ctx context.Context, container di.Container) (interface{}, error) {
		return factory(ctx, c)
	}

	var internalOptions []di.Option
	for _, opt := range options {
		internalOptions = append(internalOptions, &internalOptionWrapper{opt})
	}

	return c.internal.RegisterSingleton(serviceType, internalFactory, internalOptions...)
}

// Resolve resolves a service from the container
func (c *PublicContainer) Resolve(serviceType reflect.Type) (interface{}, error) {
	return c.internal.Resolve(serviceType)
}

// ResolveWithScope resolves a service with a specific scope
func (c *PublicContainer) ResolveWithScope(serviceType reflect.Type, scope Scope) (interface{}, error) {
	internalScope, ok := scope.(*PublicScope)
	if !ok {
		return nil, fmt.Errorf("invalid scope type")
	}
	return c.internal.ResolveWithScope(serviceType, internalScope.internal)
}

// CreateScope creates a new scope
func (c *PublicContainer) CreateScope() Scope {
	internalScope := c.internal.CreateScope()
	return &PublicScope{internal: internalScope}
}

// Validate validates the container configuration
func (c *PublicContainer) Validate() error {
	return c.internal.Validate()
}

// PublicScope wraps the internal scope to provide a public API
type PublicScope struct {
	internal di.Scope
}

// Resolve resolves a service within this scope
func (s *PublicScope) Resolve(serviceType reflect.Type) (interface{}, error) {
	return s.internal.Resolve(serviceType)
}

// Dispose disposes the scope and all its resources
func (s *PublicScope) Dispose() error {
	return s.internal.Dispose()
}

// internalOptionWrapper wraps public options for internal use
type internalOptionWrapper struct {
	option Option
}

// Apply applies the option to the internal registration
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
