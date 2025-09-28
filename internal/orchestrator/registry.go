package orchestrator

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
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
	appLogger := logger.NewSlogAdapter(slog.Default())

	// Create DI container
	diConfig := di.ContainerConfig{
		EnableValidation:    true,
		EnableCircularCheck: true,
		EnableInterception:  config.EnableTracing,
		DefaultLifetime:     di.Singleton,
		MaxResolutionDepth:  50,
		EnableMetrics:       config.EnableMetrics,
	}

	container := di.NewContainer(diConfig, appLogger)

	// Register the logger in the DI container immediately for automatic injection
	loggerType := reflect.TypeOf((*logger.Logger)(nil)).Elem()
	if err := container.RegisterInstance(loggerType, appLogger); err != nil {
		// If registration fails, log the error but don't fail the entire initialization
		// This ensures the orchestrator can still work even if logger registration fails
		appLogger.Error("Failed to register logger in DI container", "error", err)
	}

	// Also register the logger as a named service for dependency discovery
	// This allows the logger to be resolved by both type and name
	loggerName := "logger::Logger" // Use the same naming convention as typeToDependencyName
	containerWrapper := &Container{container: container}
	if err := containerWrapper.RegisterNamedInstance(loggerName, loggerType, appLogger); err != nil {
		appLogger.Error("Failed to register logger as named service", "error", err)
	}

	// Create lifecycle manager
	lifecycleManager := lifecycle.NewLifecycleManager(appLogger)

	// Register the logger as a virtual component in the lifecycle manager
	// This allows dependency validation to pass for services that depend on the logger
	loggerComponent := &loggerComponent{name: loggerName}
	if err := lifecycleManager.RegisterComponent(loggerComponent); err != nil {
		appLogger.Error("Failed to register logger as lifecycle component", "error", err)
	}

	return &ServiceRegistry{
		container:        container,
		lifecycleManager: lifecycleManager,
		services:         make(map[string]*ServiceDefinition),
		config:           config,
		logger:           appLogger,
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

// Logger returns the orchestrator's logger instance.
// This allows services and other components to use the same logger
// that the orchestrator uses for consistent logging.
func (sr *ServiceRegistry) Logger() logger.Logger {
	return sr.logger
}

// loggerComponent is a virtual component that represents the logger in the lifecycle manager.
// It doesn't have any lifecycle methods since the logger doesn't need to be started/stopped.
type loggerComponent struct {
	name string
}

func (c *loggerComponent) Name() string {
	return c.name
}

func (c *loggerComponent) Dependencies() []string {
	return []string{} // Logger has no dependencies
}

func (c *loggerComponent) Start(ctx context.Context) error {
	// Logger doesn't need to be started
	return nil
}

func (c *loggerComponent) Stop(ctx context.Context) error {
	// Logger doesn't need to be stopped
	return nil
}

func (c *loggerComponent) Health(ctx context.Context) lifecycle.ComponentHealth {
	// Logger is always healthy
	return lifecycle.ComponentHealth{
		Status:    lifecycle.HealthStatusHealthy,
		Message:   "Logger is healthy",
		Timestamp: time.Now(),
	}
}

func (c *loggerComponent) GetRetryConfig() *lifecycle.RetryConfig {
	return nil // Logger doesn't need retry configuration
}
