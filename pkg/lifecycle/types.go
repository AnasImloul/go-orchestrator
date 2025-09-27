package lifecycle

import (
	"context"
	"time"
)

// Component represents a lifecycle component
type Component interface {
	// Name returns the component name (must be unique)
	Name() string

	// Dependencies returns the names of components this component depends on
	Dependencies() []string

	// Priority returns the component priority (lower numbers start first)
	Priority() int

	// Start initializes the component
	Start(ctx context.Context) error

	// Stop gracefully shuts down the component
	Stop(ctx context.Context) error

	// Health returns the component's health status
	Health(ctx context.Context) ComponentHealth

	// GetRetryConfig returns retry configuration for this component (optional)
	GetRetryConfig() *RetryConfig
}

// LifecycleManager manages component lifecycle operations
type LifecycleManager interface {
	// RegisterComponent registers a component for lifecycle management
	RegisterComponent(component Component) error

	// UnregisterComponent removes a component from lifecycle management
	UnregisterComponent(name string) error

	// Start starts all registered components in dependency order
	Start(ctx context.Context) error

	// Stop stops all registered components in reverse order
	Stop(ctx context.Context) error

	// GetPhase returns the current lifecycle phase
	GetPhase() Phase

	// GetComponentState returns the state of a specific component
	GetComponentState(name string) (*ComponentState, error)

	// GetAllComponentStates returns the states of all components
	GetAllComponentStates() map[string]*ComponentState

	// AddHook adds a lifecycle hook
	AddHook(phase Phase, hook Hook) error

	// RemoveHook removes a lifecycle hook
	RemoveHook(phase Phase, hook Hook) error
}

// Phase represents the lifecycle phase
type Phase int

const (
	// PhaseStopped indicates the component is stopped
	PhaseStopped Phase = iota
	// PhaseStarting indicates the component is starting
	PhaseStarting
	// PhaseRunning indicates the component is running
	PhaseRunning
	// PhaseStopping indicates the component is stopping
	PhaseStopping
	// PhaseFailed indicates the component has failed
	PhaseFailed
)

// String returns the string representation of the phase
func (p Phase) String() string {
	switch p {
	case PhaseStopped:
		return "stopped"
	case PhaseStarting:
		return "starting"
	case PhaseRunning:
		return "running"
	case PhaseStopping:
		return "stopping"
	case PhaseFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// ComponentState represents the state of a component
type ComponentState struct {
	Name         string            `json:"name"`
	Phase        Phase             `json:"phase"`
	Dependencies []string          `json:"dependencies"`
	Priority     int               `json:"priority"`
	StartedAt    *time.Time        `json:"started_at,omitempty"`
	StoppedAt    *time.Time        `json:"stopped_at,omitempty"`
	Error        error             `json:"error,omitempty"`
	Health       ComponentHealth   `json:"health"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// ComponentHealth represents the health status of a component
type ComponentHealth struct {
	Status    HealthStatus            `json:"status"`
	Message   string                  `json:"message"`
	Details   map[string]interface{}  `json:"details,omitempty"`
	Timestamp time.Time               `json:"timestamp"`
}

// HealthStatus represents the health status
type HealthStatus int

const (
	// HealthStatusUnknown indicates the health status is unknown
	HealthStatusUnknown HealthStatus = iota
	// HealthStatusHealthy indicates the component is healthy
	HealthStatusHealthy
	// HealthStatusDegraded indicates the component is degraded
	HealthStatusDegraded
	// HealthStatusUnhealthy indicates the component is unhealthy
	HealthStatusUnhealthy
)

// String returns the string representation of the health status
func (h HealthStatus) String() string {
	switch h {
	case HealthStatusUnknown:
		return "unknown"
	case HealthStatusHealthy:
		return "healthy"
	case HealthStatusDegraded:
		return "degraded"
	case HealthStatusUnhealthy:
		return "unhealthy"
	default:
		return "unknown"
	}
}

// Hook represents a lifecycle hook
type Hook interface {
	// Execute executes the hook
	Execute(ctx context.Context, event Event) error
}

// HookFunc is a function-based hook
type HookFunc func(ctx context.Context, event Event) error

// Execute implements the Hook interface
func (f HookFunc) Execute(ctx context.Context, event Event) error {
	return f(ctx, event)
}

// Event represents a lifecycle event
type Event struct {
	Phase     Phase                 `json:"phase"`
	Component string                `json:"component"`
	Timestamp time.Time             `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// RetryConfig configures retry behavior for component operations
type RetryConfig struct {
	MaxAttempts       int           `json:"max_attempts"`        // Maximum number of retry attempts (default: 3)
	InitialDelay      time.Duration `json:"initial_delay"`       // Initial delay between retries (default: 100ms)
	MaxDelay          time.Duration `json:"max_delay"`           // Maximum delay between retries (default: 5s)
	BackoffMultiplier float64       `json:"backoff_multiplier"`  // Multiplier for exponential backoff (default: 2.0)
	RetryableErrors   []error       `json:"retryable_errors"`    // Specific errors that should trigger retry (nil means all errors)
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:       3,
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          5 * time.Second,
		BackoffMultiplier: 2.0,
		RetryableErrors:   nil, // Retry on all errors
	}
}

// IsRetryableError checks if an error should trigger a retry
func (rc *RetryConfig) IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// If no specific retryable errors are configured, retry on all errors
	if len(rc.RetryableErrors) == 0 {
		return true
	}

	// Check if the error matches any of the retryable errors
	for _, retryableErr := range rc.RetryableErrors {
		if err.Error() == retryableErr.Error() {
			return true
		}
	}

	return false
}
