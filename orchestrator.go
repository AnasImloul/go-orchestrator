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
	"github.com/AnasImloul/go-orchestrator/internal/orchestrator"
)

// Re-export types from internal package for public API
type (
	// ServiceRegistry represents the main service registry for dependency injection and lifecycle management.
	// This is the single entry point for the library.
	ServiceRegistry = orchestrator.ServiceRegistry

	// Config holds configuration for the application.
	Config = orchestrator.Config

	// ServiceDefinitionInterface represents the interface for service definitions.
	ServiceDefinitionInterface = orchestrator.ServiceDefinitionInterface

	// ServiceDefinition represents a declarative service configuration.
	ServiceDefinition = orchestrator.ServiceDefinition

	// ServiceConfig represents a service registration configuration.
	ServiceConfig = orchestrator.ServiceConfig

	// Lifetime represents the service lifetime.
	Lifetime = orchestrator.Lifetime

	// LifecycleConfig represents a service lifecycle configuration.
	LifecycleConfig = orchestrator.LifecycleConfig

	// LifecycleBuilder provides a fluent interface for building service lifecycle configurations.
	LifecycleBuilder = orchestrator.LifecycleBuilder

	// HealthStatusType represents the type of health status.
	HealthStatusType = orchestrator.HealthStatusType

	// HealthStatus represents the health status of a component.
	HealthStatus = orchestrator.HealthStatus

	// Service represents a service that can be managed by the orchestrator.
	// All services MUST implement this interface for automatic lifecycle management.
	Service = orchestrator.Service

	// BaseService provides default implementations for Service interface methods.
	// Services can embed this struct to get sensible defaults without implementing
	// all methods manually.
	BaseService = orchestrator.BaseService

	// Container provides a simplified interface to the DI container.
	Container = orchestrator.Container
)

// Note: TypedServiceDefinition and TypedServiceConfig are available through the internal package.
// Use the factory functions (NewServiceSingleton, NewAutoServiceFactory, NewServiceFactory)
// which return the appropriate typed definitions from the internal package.

// Re-export constants from internal package
const (
	// Transient creates a new instance for each resolution
	Transient Lifetime = orchestrator.Transient
	// Scoped creates one instance per scope
	Scoped Lifetime = orchestrator.Scoped
	// Singleton creates one instance for the entire container
	Singleton Lifetime = orchestrator.Singleton
)

const (
	// HealthStatusHealthy indicates the service is healthy and functioning normally
	HealthStatusHealthy HealthStatusType = orchestrator.HealthStatusHealthy
	// HealthStatusDegraded indicates the service is functioning but with reduced performance or capabilities
	HealthStatusDegraded HealthStatusType = orchestrator.HealthStatusDegraded
	// HealthStatusUnhealthy indicates the service is not functioning properly
	HealthStatusUnhealthy HealthStatusType = orchestrator.HealthStatusUnhealthy
	// HealthStatusUnknown indicates the health status cannot be determined
	HealthStatusUnknown HealthStatusType = orchestrator.HealthStatusUnknown
)

// Public API functions - delegate to internal implementation

// DefaultConfig returns the default application configuration.
func DefaultConfig() Config {
	return orchestrator.DefaultConfig()
}

// New creates a new application with the default configuration.
func New() *ServiceRegistry {
	return orchestrator.New()
}

// NewWithConfig creates a new service registry with the specified configuration.
func NewWithConfig(config Config) *ServiceRegistry {
	return orchestrator.NewWithConfig(config)
}

// NewLifecycle creates a new service lifecycle builder.
func NewLifecycle() *LifecycleBuilder {
	return orchestrator.NewLifecycle()
}

// NewServiceSingleton creates a new service definition with automatic lifecycle management.
// The service instance MUST implement the Service interface.
// Lifecycle methods (Start, Stop, Health) are automatically wired.
func NewServiceSingleton[T Service](instance T) *orchestrator.TypedServiceDefinition[T] {
	return orchestrator.NewServiceSingleton(instance)
}

// NewAutoServiceFactory creates a new service definition with automatic dependency discovery and lifecycle management.
// The factory function can return any type T - it doesn't need to implement the Service interface.
// Dependencies are automatically discovered from the factory function parameters.
// Lifecycle methods are automatically provided with sensible defaults.
func NewAutoServiceFactory[T any](factory interface{}, lifetime Lifetime) *orchestrator.TypedServiceDefinition[T] {
	return orchestrator.NewAutoServiceFactory[T](factory, lifetime)
}

// NewServiceFactory creates a new service definition with automatic dependency discovery and lifecycle management.
// The factory function must return a type T that implements the Service interface.
// Dependencies are automatically discovered from the factory function parameters.
// Lifecycle methods (Start, Stop, Health) are automatically wired.
func NewServiceFactory[T Service](factory interface{}, lifetime Lifetime) *orchestrator.TypedServiceDefinition[T] {
	return orchestrator.NewServiceFactory[T](factory, lifetime)
}

// ResolveType resolves a service by interface type.
// T must be an interface type, not a concrete struct.
func ResolveType[T any](c *Container) (T, error) {
	return orchestrator.ResolveType[T](c)
}
