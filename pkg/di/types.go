package di

import (
	"context"
	"reflect"
)

// Container represents a dependency injection container.
type Container interface {
	// Register registers a service with the container
	Register(serviceType reflect.Type, factory Factory, options ...Option) error

	// RegisterInstance registers a service instance
	RegisterInstance(serviceType reflect.Type, instance interface{}) error

	// RegisterSingleton registers a singleton service
	RegisterSingleton(serviceType reflect.Type, factory Factory, options ...Option) error

	// Resolve resolves a service from the container
	Resolve(serviceType reflect.Type) (interface{}, error)

	// CreateScope creates a new scope
	CreateScope() Scope

	// Validate validates the container configuration
	Validate() error
}

// Factory is a function that creates a service instance.
type Factory func(ctx context.Context, container Container) (interface{}, error)

// Option represents a registration option.
type Option interface {
	Apply(registration *ServiceRegistration)
}

// Scope represents a service scope.
type Scope interface {
	// Resolve resolves a service within this scope
	Resolve(serviceType reflect.Type) (interface{}, error)

	// Dispose disposes the scope and all its resources
	Dispose() error
}

// ServiceRegistration represents a service registration.
type ServiceRegistration struct {
	ServiceType  reflect.Type
	Factory      Factory
	Lifetime     Lifetime
	Options      []Option
	Interceptors []Interceptor
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

// Interceptor represents a service interceptor.
type Interceptor interface {
	Intercept(ctx context.Context, serviceType reflect.Type, next func() (interface{}, error)) (interface{}, error)
}

// InterceptorFunc is a function-based interceptor.
type InterceptorFunc func(ctx context.Context, serviceType reflect.Type, next func() (interface{}, error)) (interface{}, error)

// Intercept implements the Interceptor interface.
func (f InterceptorFunc) Intercept(ctx context.Context, serviceType reflect.Type, next func() (interface{}, error)) (interface{}, error) {
	return f(ctx, serviceType, next)
}

// Disposable represents a service that can be disposed.
type Disposable interface {
	Dispose() error
}

// Generic helper functions for type-safe dependency resolution.

// TypeOf returns the reflect.Type for type T.
func TypeOf[T any]() reflect.Type {
	var zero T
	return reflect.TypeOf(zero)
}

// Resolve resolves a service of type T from the container.
func Resolve[T any](container Container) (T, error) {
	var zero T
	serviceType := reflect.TypeOf(zero)
	instance, err := container.Resolve(serviceType)
	if err != nil {
		return zero, err
	}
	return instance.(T), nil
}

// MustResolve resolves a service of type T from the container, panicking on error.
func MustResolve[T any](container Container) T {
	instance, err := Resolve[T](container)
	if err != nil {
		panic(err)
	}
	return instance
}

// TryResolve attempts to resolve a service of type T from the container.
func TryResolve[T any](container Container) (T, bool) {
	var zero T
	serviceType := reflect.TypeOf(zero)
	instance, err := container.Resolve(serviceType)
	if err != nil {
		return zero, false
	}
	return instance.(T), true
}
