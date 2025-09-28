package orchestrator

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/AnasImloul/go-orchestrator/internal/di"
	"github.com/AnasImloul/go-orchestrator/internal/lifecycle"
	"github.com/AnasImloul/go-orchestrator/internal/logger"
)

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

// Register registers a service definition in the service registry.
// Accepts both ServiceDefinition and TypedServiceDefinition[T] through the interface.
func (sr *ServiceRegistry) Register(serviceDefInterface ServiceDefinitionInterface) *ServiceRegistry {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	serviceDef := serviceDefInterface.ToServiceDefinition()

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
				// Create a wrapper factory that sets the registry reference and service name for BaseService instances
				wrappedFactory := func(ctx context.Context, container *Container) (interface{}, error) {
					instance, err := service.Factory(ctx, container)
					if err != nil {
						return instance, err
					}
					
					// Set registry reference and service name if the instance embeds BaseService
					if baseService, ok := instance.(interface{ SetRegistry(*ServiceRegistry) }); ok {
						baseService.SetRegistry(sr)
					}
					if baseService, ok := instance.(interface{ SetServiceName(string) }); ok {
						serviceName := service.Name
						if serviceName == "" {
							serviceName = service.Type.String()
						}
						baseService.SetServiceName(serviceName)
					}
					
					return instance, nil
				}
				
				if service.Name != "" {
					// Register named service
					if err := container.RegisterNamed(service.Name, service.Type, wrappedFactory, service.Lifetime); err != nil {
						return fmt.Errorf("failed to register named service %s (%s): %w", service.Name, service.Type.String(), err)
					}
				} else {
					// Register unnamed service
					if err := container.Register(service.Type, wrappedFactory, service.Lifetime); err != nil {
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
			serviceDef:      serviceDef,
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

	// Use HealthCheck to actually call the component's Health method
	componentHealth := sr.lifecycleManager.HealthCheck(ctx)
	health := make(map[string]HealthStatus)

	for name, componentHealth := range componentHealth {
		// Convert lifecycle HealthStatus to our HealthStatusType
		var statusType HealthStatusType
		switch componentHealth.Status {
		case lifecycle.HealthStatusHealthy:
			statusType = HealthStatusHealthy
		case lifecycle.HealthStatusDegraded:
			statusType = HealthStatusDegraded
		case lifecycle.HealthStatusUnhealthy:
			statusType = HealthStatusUnhealthy
		default:
			statusType = HealthStatusUnknown
		}

		health[name] = HealthStatus{
			Status:  statusType,
			Message: componentHealth.Message,
			Details: componentHealth.Details,
		}
	}

	return health
}

// Container returns the DI container.
func (sr *ServiceRegistry) Container() *Container {
	return &Container{container: sr.container}
}
