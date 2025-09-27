package lifecycle

import (
	"context"
	"sync"
	"time"

	"github.com/AnasImloul/go-orchestrator/internal/lifecycle"
	"github.com/AnasImloul/go-orchestrator/internal/logger"
)

// PublicLifecycleManager wraps the internal lifecycle manager to provide a public API
type PublicLifecycleManager struct {
	internal *lifecycle.DefaultLifecycleManager
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager(logger logger.Logger) LifecycleManager {
	internalManager := lifecycle.NewLifecycleManager(logger)
	return &PublicLifecycleManager{internal: internalManager}
}

// RegisterComponent registers a component for lifecycle management
func (lm *PublicLifecycleManager) RegisterComponent(component Component) error {
	// Convert public component to internal component
	internalComponent := &componentWrapper{component: component}
	return lm.internal.RegisterComponent(internalComponent)
}

// UnregisterComponent removes a component from lifecycle management
func (lm *PublicLifecycleManager) UnregisterComponent(name string) error {
	return lm.internal.UnregisterComponent(name)
}

// Start starts all registered components in dependency order
func (lm *PublicLifecycleManager) Start(ctx context.Context) error {
	return lm.internal.Start(ctx)
}

// Stop stops all registered components in reverse order
func (lm *PublicLifecycleManager) Stop(ctx context.Context) error {
	return lm.internal.Stop(ctx)
}

// GetPhase returns the current lifecycle phase
func (lm *PublicLifecycleManager) GetPhase() Phase {
	return Phase(lm.internal.GetPhase())
}

// GetComponentState returns the state of a specific component
func (lm *PublicLifecycleManager) GetComponentState(name string) (*ComponentState, error) {
	internalState, err := lm.internal.GetComponentState(name)
	if err != nil {
		return nil, err
	}

	// Convert internal state to public state
	return &ComponentState{
		Name:         internalState.Name,
		Phase:        Phase(internalState.Phase),
		Dependencies: internalState.Dependencies,
		Priority:     internalState.Priority,
		StartedAt:    internalState.StartedAt,
		StoppedAt:    internalState.StoppedAt,
		Error:        internalState.Error,
		Health: ComponentHealth{
			Status:    HealthStatus(internalState.Health.Status),
			Message:   internalState.Health.Message,
			Details:   internalState.Health.Details,
			Timestamp: internalState.Health.Timestamp,
		},
		Metadata: internalState.Metadata,
	}, nil
}

// GetAllComponentStates returns the states of all components
func (lm *PublicLifecycleManager) GetAllComponentStates() map[string]*ComponentState {
	internalStates := lm.internal.GetAllComponentStates()
	publicStates := make(map[string]*ComponentState)

	for name, internalState := range internalStates {
		publicStates[name] = &ComponentState{
			Name:         internalState.Name,
			Phase:        Phase(internalState.Phase),
			Dependencies: internalState.Dependencies,
			Priority:     internalState.Priority,
			StartedAt:    internalState.StartedAt,
			StoppedAt:    internalState.StoppedAt,
			Error:        internalState.Error,
			Health: ComponentHealth{
				Status:    HealthStatus(internalState.Health.Status),
				Message:   internalState.Health.Message,
				Details:   internalState.Health.Details,
				Timestamp: internalState.Health.Timestamp,
			},
			Metadata: internalState.Metadata,
		}
	}

	return publicStates
}

// AddHook adds a lifecycle hook
func (lm *PublicLifecycleManager) AddHook(phase Phase, hook Hook) error {
	internalHook := &hookWrapper{hook: hook}
	return lm.internal.AddHook(lifecycle.Phase(phase), internalHook)
}

// RemoveHook removes a lifecycle hook
func (lm *PublicLifecycleManager) RemoveHook(phase Phase, hook Hook) error {
	internalHook := &hookWrapper{hook: hook}
	return lm.internal.RemoveHook(lifecycle.Phase(phase), internalHook)
}

// componentWrapper wraps a public component for internal use
type componentWrapper struct {
	component Component
}

// Name returns the component name
func (w *componentWrapper) Name() string {
	return w.component.Name()
}

// Dependencies returns the component dependencies
func (w *componentWrapper) Dependencies() []string {
	return w.component.Dependencies()
}

// Priority returns the component priority
func (w *componentWrapper) Priority() int {
	return w.component.Priority()
}

// Start starts the component
func (w *componentWrapper) Start(ctx context.Context) error {
	return w.component.Start(ctx)
}

// Stop stops the component
func (w *componentWrapper) Stop(ctx context.Context) error {
	return w.component.Stop(ctx)
}

// Health returns the component health
func (w *componentWrapper) Health(ctx context.Context) lifecycle.ComponentHealth {
	publicHealth := w.component.Health(ctx)
	return lifecycle.ComponentHealth{
		Status:    lifecycle.HealthStatus(publicHealth.Status),
		Message:   publicHealth.Message,
		Details:   publicHealth.Details,
		Timestamp: publicHealth.Timestamp,
	}
}

// GetRetryConfig returns retry configuration for this component
func (w *componentWrapper) GetRetryConfig() *lifecycle.RetryConfig {
	publicConfig := w.component.GetRetryConfig()
	if publicConfig == nil {
		return nil
	}

	return &lifecycle.RetryConfig{
		MaxAttempts:       publicConfig.MaxAttempts,
		InitialDelay:      publicConfig.InitialDelay,
		MaxDelay:          publicConfig.MaxDelay,
		BackoffMultiplier: publicConfig.BackoffMultiplier,
		RetryableErrors:   publicConfig.RetryableErrors,
	}
}

// hookWrapper wraps a public hook for internal use
type hookWrapper struct {
	hook Hook
}

// Execute executes the hook
func (w *hookWrapper) Execute(ctx context.Context, event lifecycle.Event) error {
	publicEvent := Event{
		Phase:     Phase(event.Phase),
		Component: event.Component,
		Timestamp: event.Timestamp,
		Data:      event.Data,
	}
	return w.hook.Execute(ctx, publicEvent)
}
