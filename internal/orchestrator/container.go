package orchestrator

import (
	"context"
	"fmt"
	"reflect"

	"github.com/AnasImloul/go-orchestrator/internal/di"
)

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

// ResolveStruct resolves a service by struct type.
// T can be any type (interface or struct).
func ResolveStruct[T any](c *Container) (T, error) {
	var zero T
	serviceType := reflect.TypeOf((*T)(nil)).Elem()

	instance, err := c.Resolve(serviceType)
	if err != nil {
		return zero, err
	}
	return instance.(T), nil
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
