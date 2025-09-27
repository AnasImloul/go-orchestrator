package orchestrator

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/imloulanas/go-orchestrator/di"
	"github.com/imloulanas/go-orchestrator/lifecycle"
	"github.com/imloulanas/go-orchestrator/logger"
)


// DefaultOrchestrator implements the Orchestrator interface
type DefaultOrchestrator struct {
	container         di.Container
	lifecycleManager  lifecycle.LifecycleManager
	features          map[string]Feature
	config            OrchestratorConfig
	logger            logger.Logger
	mu                sync.RWMutex
	healthCheckTicker *time.Ticker
	stopHealthCheck   chan struct{}
}

// NewOrchestrator creates a new application orchestrator
func NewOrchestrator(config OrchestratorConfig, logger logger.Logger) (*DefaultOrchestrator, error) {
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

	orchestrator := &DefaultOrchestrator{
		container:        container,
		lifecycleManager: lifecycleManager,
		features:         make(map[string]Feature),
		config:           config,
		logger:           logger,
		stopHealthCheck:  make(chan struct{}),
	}

	// Register core services in DI container
	if err := orchestrator.registerCoreServices(); err != nil {
		return nil, fmt.Errorf("failed to register core services: %w", err)
	}

	return orchestrator, nil
}

// RegisterFeature registers a feature with the orchestrator
func (o *DefaultOrchestrator) RegisterFeature(feature Feature) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	name := feature.GetName()

	// Check if feature is already registered
	if _, exists := o.features[name]; exists {
		return fmt.Errorf("feature %s is already registered", name)
	}

	// Store feature
	o.features[name] = feature

	// Create component wrapper
	wrapper := &ComponentWrapper{
		feature:   feature,
		container: o.container,
	}

	// Register component with lifecycle manager
	if err := o.lifecycleManager.RegisterComponent(wrapper); err != nil {
		delete(o.features, name)
		return fmt.Errorf("failed to register feature component: %w", err)
	}

	if o.logger != nil {
		o.logger.Info("Feature registered",
			"feature", name,
			"dependencies", feature.GetDependencies(),
			"priority", feature.GetPriority(),
		)
	}

	return nil
}

// Start starts the application in the correct order
func (o *DefaultOrchestrator) Start(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.logger != nil {
		o.logger.Info("Starting application orchestrator",
			"features", len(o.features),
			"timeout", o.config.StartupTimeout,
		)
	}

	// Create timeout context
	startupCtx, cancel := context.WithTimeout(ctx, o.config.StartupTimeout)
	defer cancel()

	// Add startup hooks
	if err := o.addStartupHooks(); err != nil {
		return fmt.Errorf("failed to add startup hooks: %w", err)
	}

	// Start lifecycle manager (which will start all features in order)
	if err := o.lifecycleManager.Start(startupCtx); err != nil {
		return fmt.Errorf("failed to start lifecycle manager: %w", err)
	}

	// Start periodic health checks
	if o.config.HealthCheckInterval > 0 {
		o.startHealthCheckRoutine()
	}

	if o.logger != nil {
		o.logger.Info("Application orchestrator started successfully")
	}
	return nil
}

// Stop stops the application gracefully
func (o *DefaultOrchestrator) Stop(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.logger != nil {
		o.logger.Info("Stopping application orchestrator")
	}

	// Stop health check routine
	if o.healthCheckTicker != nil {
		close(o.stopHealthCheck)
		o.healthCheckTicker.Stop()
		o.healthCheckTicker = nil
	}

	// Create timeout context
	shutdownCtx, cancel := context.WithTimeout(ctx, o.config.ShutdownTimeout)
	defer cancel()

	// Add shutdown hooks
	if err := o.addShutdownHooks(); err != nil {
		if o.logger != nil {
			o.logger.Error("Failed to add shutdown hooks",
				"error", err.Error(),
			)
		}
	}

	// Stop lifecycle manager (which will stop all features in reverse order)
	if err := o.lifecycleManager.Stop(shutdownCtx); err != nil {
		if o.logger != nil {
			o.logger.Error("Failed to stop lifecycle manager gracefully",
				"error", err.Error(),
			)
		}
	}

	// Dispose DI container
	if err := o.container.Dispose(); err != nil {
		if o.logger != nil {
			o.logger.Error("Failed to dispose DI container",
				"error", err.Error(),
			)
		}
	}

	if o.logger != nil {
		o.logger.Info("Application orchestrator stopped")
	}
	return nil
}

// GetContainer returns the DI container
func (o *DefaultOrchestrator) GetContainer() di.Container {
	return o.container
}

// GetLifecycleManager returns the lifecycle manager
func (o *DefaultOrchestrator) GetLifecycleManager() lifecycle.LifecycleManager {
	return o.lifecycleManager
}

// GetPhase returns the current application phase
func (o *DefaultOrchestrator) GetPhase() lifecycle.Phase {
	return o.lifecycleManager.GetPhase()
}

// HealthCheck performs a health check on all components
func (o *DefaultOrchestrator) HealthCheck(ctx context.Context) HealthReport {
	o.mu.RLock()
	defer o.mu.RUnlock()

	report := HealthReport{
		Timestamp:  time.Now(),
		Phase:      o.lifecycleManager.GetPhase(),
		Features:   make(map[string]FeatureHealth),
		Components: make(map[string]lifecycle.ComponentHealth),
		Summary:    HealthSummary{TotalFeatures: len(o.features)},
	}

	// Get component health from lifecycle manager
	componentHealth := o.lifecycleManager.HealthCheck(ctx)
	report.Components = componentHealth

	// Check feature health
	for name, feature := range o.features {
		health := o.checkFeatureHealth(ctx, feature, componentHealth[name])
		report.Features[name] = health

		// Update summary
		switch health.Status {
		case HealthStatusHealthy:
			report.Summary.HealthyFeatures++
		case HealthStatusDegraded:
			report.Summary.DegradedFeatures++
		case HealthStatusUnhealthy:
			report.Summary.UnhealthyFeatures++
		default:
			report.Summary.UnknownFeatures++
		}
	}

	// Determine overall status
	report.Status = o.determineOverallHealth(report.Summary)

	return report
}

// Private helper methods

// registerCoreServices registers core services in the DI container
func (o *DefaultOrchestrator) registerCoreServices() error {
	// Register the DI container itself
	if err := o.container.RegisterInstance(reflect.TypeOf((*di.Container)(nil)).Elem(), o.container); err != nil {
		return fmt.Errorf("failed to register DI container: %w", err)
	}

	// Register the lifecycle manager
	if err := o.container.RegisterInstance(reflect.TypeOf((*lifecycle.LifecycleManager)(nil)).Elem(), o.lifecycleManager); err != nil {
		return fmt.Errorf("failed to register lifecycle manager: %w", err)
	}

	// Register the orchestrator itself
	if err := o.container.RegisterInstance(reflect.TypeOf((*Orchestrator)(nil)).Elem(), o); err != nil {
		return fmt.Errorf("failed to register orchestrator: %w", err)
	}

	return nil
}

// addStartupHooks adds startup lifecycle hooks
func (o *DefaultOrchestrator) addStartupHooks() error {
	startupHook := func(ctx context.Context, event lifecycle.Event) error {
		if o.logger != nil {
			o.logger.Info("Startup event",
				"component", event.Component,
				"phase", event.Phase,
			)
		}
		return nil
	}

	return o.lifecycleManager.AddHook(lifecycle.PhaseStartup, startupHook)
}

// addShutdownHooks adds shutdown lifecycle hooks
func (o *DefaultOrchestrator) addShutdownHooks() error {
	shutdownHook := func(ctx context.Context, event lifecycle.Event) error {
		if o.logger != nil {
			o.logger.Info("Shutdown event",
				"component", event.Component,
				"phase", event.Phase,
			)
		}
		return nil
	}

	return o.lifecycleManager.AddHook(lifecycle.PhaseShutdown, shutdownHook)
}

// startHealthCheckRoutine starts the periodic health check routine
func (o *DefaultOrchestrator) startHealthCheckRoutine() {
	o.healthCheckTicker = time.NewTicker(o.config.HealthCheckInterval)

	go func() {
		for {
			select {
			case <-o.healthCheckTicker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				report := o.HealthCheck(ctx)
				cancel()

				if report.Status != HealthStatusHealthy {
					if o.logger != nil {
						o.logger.Warn("Health check failed",
							"status", report.Status,
							"unhealthy_features", report.Summary.UnhealthyFeatures,
							"degraded_features", report.Summary.DegradedFeatures,
						)
					}
				} else {
					if o.logger != nil {
						o.logger.Debug("Health check passed",
							"healthy_features", report.Summary.HealthyFeatures,
						)
					}
				}

			case <-o.stopHealthCheck:
				return
			}
		}
	}()
}

// checkFeatureHealth checks the health of a specific feature
func (o *DefaultOrchestrator) checkFeatureHealth(ctx context.Context, feature Feature, componentHealth lifecycle.ComponentHealth) FeatureHealth {
	health := FeatureHealth{
		Timestamp: time.Now(),
		Metadata:  feature.GetMetadata(),
	}

	// Base health on component health
	switch componentHealth.Status {
	case lifecycle.HealthStatusHealthy:
		health.Status = HealthStatusHealthy
		health.Message = "Feature is healthy"
	case lifecycle.HealthStatusDegraded:
		health.Status = HealthStatusDegraded
		health.Message = componentHealth.Message
	case lifecycle.HealthStatusUnhealthy:
		health.Status = HealthStatusUnhealthy
		health.Message = componentHealth.Message
	default:
		health.Status = HealthStatusUnknown
		health.Message = "Feature health unknown"
	}

	return health
}

// determineOverallHealth determines the overall application health
func (o *DefaultOrchestrator) determineOverallHealth(summary HealthSummary) HealthStatus {
	if summary.UnhealthyFeatures > 0 {
		return HealthStatusUnhealthy
	}

	if summary.DegradedFeatures > 0 {
		return HealthStatusDegraded
	}

	if summary.HealthyFeatures == summary.TotalFeatures && summary.TotalFeatures > 0 {
		return HealthStatusHealthy
	}

	return HealthStatusUnknown
}
