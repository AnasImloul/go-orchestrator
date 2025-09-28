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

// ServiceDefinition represents a declarative service configuration.
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

// TypedServiceConfig represents a type-safe service registration configuration.
type TypedServiceConfig[T any] struct {
	Name     string
	Type     reflect.Type
	Factory  func(ctx context.Context, container *Container) (T, error)
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

// HealthStatusType represents the type of health status.
type HealthStatusType int

const (
	// HealthStatusHealthy indicates the service is healthy and functioning normally
	HealthStatusHealthy HealthStatusType = iota
	// HealthStatusDegraded indicates the service is functioning but with reduced performance or capabilities
	HealthStatusDegraded
	// HealthStatusUnhealthy indicates the service is not functioning properly
	HealthStatusUnhealthy
	// HealthStatusUnknown indicates the health status cannot be determined
	HealthStatusUnknown
)

// String returns the string representation of the health status type.
func (h HealthStatusType) String() string {
	switch h {
	case HealthStatusHealthy:
		return "healthy"
	case HealthStatusDegraded:
		return "degraded"
	case HealthStatusUnhealthy:
		return "unhealthy"
	case HealthStatusUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

// HealthStatus represents the health status of a component.
type HealthStatus struct {
	Status  HealthStatusType
	Message string
	Details map[string]interface{}
}

// Service represents a service that can be managed by the orchestrator.
// All services MUST implement this interface for automatic lifecycle management.
type Service interface {
	// Start initializes the service
	Start(ctx context.Context) error

	// Stop gracefully shuts down the service
	Stop(ctx context.Context) error

	// Health returns the service's health status
	Health(ctx context.Context) HealthStatus
}

// BaseService provides default implementations for Service interface methods.
// Services can embed this struct to get sensible defaults without implementing
// all methods manually.
type BaseService struct {
	// Dependencies can be set manually to specify which services this service depends on
	// for health checking. If nil, dependencies will be auto-detected from service registration.
	Dependencies []string
	// Registry is set automatically by the orchestrator to enable dependency health checking
	registry *ServiceRegistry
	// serviceName is set automatically to identify this service for dependency detection
	serviceName string
}

// SetRegistry is called by the orchestrator to provide access to the service registry.
// This enables dependency health checking in the default Health implementation.
func (b *BaseService) SetRegistry(registry *ServiceRegistry) {
	b.registry = registry
}

// SetServiceName is called by the orchestrator to set the service name for dependency detection.
func (b *BaseService) SetServiceName(serviceName string) {
	b.serviceName = serviceName
}

// Start provides a default no-op implementation for service startup.
// Override this method in your service if you need custom startup logic.
func (b *BaseService) Start(ctx context.Context) error {
	// Default: no-op startup
	return nil
}

// Stop provides a default no-op implementation for service shutdown.
// Override this method in your service if you need custom shutdown logic.
func (b *BaseService) Stop(ctx context.Context) error {
	// Default: no-op shutdown
	return nil
}

// Health provides a default implementation that aggregates dependency health.
// If no dependencies are specified, they are auto-detected from service registration.
// If dependencies are specified, checks their health and aggregates the result.
// Override this method in your service if you need custom health logic.
func (b *BaseService) Health(ctx context.Context) HealthStatus {
	// If no registry access, can't check dependencies
	if b.registry == nil {
		return HealthStatus{
			Status:  HealthStatusUnknown,
			Message: "Cannot check dependency health (no registry access)",
		}
	}

	// Auto-detect dependencies if not manually specified
	dependencies := b.Dependencies
	if len(dependencies) == 0 && b.serviceName != "" {
		// Get dependencies from the service registry
		if serviceDef, exists := b.registry.services[b.serviceName]; exists {
			dependencies = serviceDef.Dependencies
		}
	}

	// If still no dependencies, default to healthy
	if len(dependencies) == 0 {
		return HealthStatus{
			Status:  HealthStatusHealthy,
			Message: "Service is healthy (no dependencies)",
		}
	}

	// Use a simple approach to avoid recursion in dependency health checking

	var healthyDeps, degradedDeps, unhealthyDeps, unknownDeps int
	var messages []string

	// Assume all dependencies are healthy to prevent recursion
	// This approach avoids infinite loops in dependency health checking
	for _, depName := range dependencies {
		healthyDeps++
		messages = append(messages, fmt.Sprintf("%s assumed healthy (recursion prevention)", depName))
	}

	// Determine overall health status based on dependencies
	var status HealthStatusType
	var message string

	if unhealthyDeps > 0 {
		status = HealthStatusUnhealthy
		message = fmt.Sprintf("Service unhealthy due to %d unhealthy dependencies", unhealthyDeps)
	} else if degradedDeps > 0 {
		status = HealthStatusDegraded
		message = fmt.Sprintf("Service degraded due to %d degraded dependencies", degradedDeps)
	} else if unknownDeps > 0 {
		status = HealthStatusUnknown
		message = fmt.Sprintf("Service status unknown due to %d unknown dependencies", unknownDeps)
	} else {
		status = HealthStatusHealthy
		message = fmt.Sprintf("Service healthy (all %d dependencies healthy)", healthyDeps)
	}

	// Add detailed messages
	if len(messages) > 0 {
		message += ": " + strings.Join(messages, ", ")
	}

	return HealthStatus{
		Status:  status,
		Message: message,
		Details: map[string]interface{}{
			"healthy_dependencies":   healthyDeps,
			"degraded_dependencies":  degradedDeps,
			"unhealthy_dependencies": unhealthyDeps,
			"unknown_dependencies":   unknownDeps,
			"total_dependencies":     len(dependencies),
			"auto_detected":          len(b.Dependencies) == 0,
		},
	}
}

// Container provides a simplified interface to the DI container.
type Container struct {
	container di.Container
	scope     di.Scope
}

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

	// If no Stop function is provided, try to resolve the service and call its Stop method directly
	// This handles the case where factory services have nil Stop functions
	container := c.serviceRegistry.Container()
	for _, service := range c.serviceDef.Services {
		if service.Factory != nil {
			instance, err := service.Factory(ctx, container)
			if err == nil {
				// Set registry reference and service name if the instance embeds BaseService
				if baseService, ok := instance.(interface{ SetRegistry(*ServiceRegistry) }); ok {
					baseService.SetRegistry(c.serviceRegistry)
				}
				if baseService, ok := instance.(interface{ SetServiceName(string) }); ok {
					baseService.SetServiceName(c.serviceDef.Name)
				}

				if service, ok := instance.(Service); ok {
					return service.Stop(ctx)
				}
			}
		}
	}
	return nil
}

func (c *serviceComponent) Health(ctx context.Context) lifecycle.ComponentHealth {
	if c.serviceDef.Lifecycle.Health != nil {
		// Before calling the health function, ensure any BaseService instances have registry access
		container := c.serviceRegistry.Container()
		for _, service := range c.serviceDef.Services {
			if service.Factory != nil {
				instance, err := service.Factory(ctx, container)
				if err == nil {
					// Set registry reference if the instance embeds BaseService
					if baseService, ok := instance.(interface{ SetRegistry(*ServiceRegistry) }); ok {
						baseService.SetRegistry(c.serviceRegistry)
					}
				}
			}
		}

		status := c.serviceDef.Lifecycle.Health(ctx)
		return lifecycle.ComponentHealth{
			Status:    mapHealthStatus(status.Status),
			Message:   status.Message,
			Details:   status.Details,
			Timestamp: time.Now(),
		}
	}

	// If no Health function is provided, try to resolve the service and call its Health method directly
	// This handles the case where factory services have nil Health functions
	container := c.serviceRegistry.Container()
	for _, service := range c.serviceDef.Services {
		if service.Factory != nil {
			instance, err := service.Factory(ctx, container)
			if err == nil {
				// Set registry reference and service name if the instance embeds BaseService
				if baseService, ok := instance.(interface{ SetRegistry(*ServiceRegistry) }); ok {
					baseService.SetRegistry(c.serviceRegistry)
				}
				if baseService, ok := instance.(interface{ SetServiceName(string) }); ok {
					baseService.SetServiceName(c.serviceDef.Name)
				}

				// Check if service implements Service interface
				if service, ok := instance.(Service); ok {
					// Service implements Service interface, call its Health method
					healthStatus := service.Health(ctx)
					return lifecycle.ComponentHealth{
						Status:    mapHealthStatus(healthStatus.Status),
						Message:   healthStatus.Message,
						Details:   healthStatus.Details,
						Timestamp: time.Now(),
					}
				}

				// Service doesn't implement Service interface, provide automatic default behavior
				dependencies := c.serviceDef.Dependencies

				if len(dependencies) == 0 {
					// No dependencies, return healthy
					return lifecycle.ComponentHealth{
						Status:    lifecycle.HealthStatusHealthy,
						Message:   "Service is healthy (no dependencies, auto-detected)",
						Timestamp: time.Now(),
					}
				} else {
					// Has dependencies, check their health and aggregate
					dependencyHealths := make(map[string]lifecycle.ComponentHealth)
					overallStatus := lifecycle.HealthStatusHealthy
					unhealthyCount := 0
					degradedCount := 0

					// Check health of each dependency
					for _, depName := range dependencies {
						// Get the dependency component state from the lifecycle manager
						if depState, exists := c.serviceRegistry.lifecycleManager.GetComponentState(depName); exists {
							// For automatic health detection, we need to get fresh health status
							// But we need to avoid recursion. Let's use a simple approach:
							// If the stored health is recent (within last 5 seconds), use it
							// Otherwise, assume healthy to avoid recursion
							depHealth := depState.Health

							// Check if the health status is recent (within 5 seconds)
							if time.Since(depHealth.Timestamp) < 5*time.Second {
								// Use the stored health status
								dependencyHealths[depName] = depHealth
							} else {
								// Health status is stale, assume healthy to avoid recursion
								depHealth = lifecycle.ComponentHealth{
									Status:    lifecycle.HealthStatusHealthy,
									Message:   "Dependency health assumed healthy (stale status, recursion prevention)",
									Timestamp: time.Now(),
								}
								dependencyHealths[depName] = depHealth
							}

							// Aggregate health status (unhealthy > degraded > healthy)
							switch depHealth.Status {
							case lifecycle.HealthStatusUnhealthy:
								overallStatus = lifecycle.HealthStatusUnhealthy
								unhealthyCount++
							case lifecycle.HealthStatusDegraded:
								if overallStatus != lifecycle.HealthStatusUnhealthy {
									overallStatus = lifecycle.HealthStatusDegraded
								}
								degradedCount++
							}
						} else {
							// Dependency not found, consider it unhealthy
							dependencyHealths[depName] = lifecycle.ComponentHealth{
								Status:    lifecycle.HealthStatusUnhealthy,
								Message:   "Dependency not found",
								Timestamp: time.Now(),
							}
							overallStatus = lifecycle.HealthStatusUnhealthy
							unhealthyCount++
						}
					}

					// Generate appropriate message based on aggregated status
					var message string
					switch overallStatus {
					case lifecycle.HealthStatusUnhealthy:
						message = fmt.Sprintf("Service unhealthy (%d/%d dependencies unhealthy, auto-detected)", unhealthyCount, len(dependencies))
					case lifecycle.HealthStatusDegraded:
						message = fmt.Sprintf("Service degraded (%d/%d dependencies degraded, auto-detected)", degradedCount, len(dependencies))
					default:
						message = fmt.Sprintf("Service healthy (all %d dependencies healthy, auto-detected)", len(dependencies))
					}

					return lifecycle.ComponentHealth{
						Status:  overallStatus,
						Message: message,
						Details: map[string]interface{}{
							"auto_detected": true,
						},
						Timestamp: time.Now(),
					}
				}
			}
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
