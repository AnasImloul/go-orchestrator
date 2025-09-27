package lifecycle

import (
	"context"
	"time"
)

// RetryConfig configures retry behavior for component operations
type RetryConfig struct {
	MaxAttempts       int           // Maximum number of retry attempts (default: 3)
	InitialDelay      time.Duration // Initial delay between retries (default: 100ms)
	MaxDelay          time.Duration // Maximum delay between retries (default: 5s)
	BackoffMultiplier float64       // Multiplier for exponential backoff (default: 2.0)
	RetryableErrors   []error       // Specific errors that should trigger retry (nil means all errors)
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
	if rc.RetryableErrors == nil || len(rc.RetryableErrors) == 0 {
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

// RetryWithBackoff executes a function with retry logic and exponential backoff
func RetryWithBackoff(ctx context.Context, config RetryConfig, operation func() error) error {
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if this error should trigger a retry
		if !config.IsRetryableError(err) {
			return err
		}

		// Don't sleep after the last attempt
		if attempt == config.MaxAttempts-1 {
			break
		}

		// Calculate delay with exponential backoff
		delay := time.Duration(float64(config.InitialDelay) *
			pow(config.BackoffMultiplier, float64(attempt)))

		// Cap the delay at MaxDelay
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}

		// Sleep with context cancellation support
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return lastErr
}

// pow calculates x^y for float64 values
func pow(x, y float64) float64 {
	result := 1.0
	for i := 0; i < int(y); i++ {
		result *= x
	}
	return result
}

// Phase represents different lifecycle phases
type Phase string

const (
	PhaseStartup  Phase = "startup"
	PhaseRunning  Phase = "running"
	PhaseShutdown Phase = "shutdown"
	PhaseStopped  Phase = "stopped"
)

// Event represents a lifecycle event
type Event struct {
	Phase     Phase
	Component string
	Timestamp time.Time
	Data      map[string]interface{}
}

// Hook represents a lifecycle hook function
type Hook func(ctx context.Context, event Event) error

// Component represents a component that participates in the lifecycle
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

// ComponentHealth represents the health status of a component
type ComponentHealth struct {
	Status    HealthStatus
	Message   string
	Details   map[string]interface{}
	Timestamp time.Time
}

// HealthStatus represents different health states
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// ComponentState represents the current state of a component
type ComponentState struct {
	Name         string
	Phase        Phase
	Health       ComponentHealth
	StartedAt    *time.Time
	StoppedAt    *time.Time
	Dependencies []string
	Priority     int
	Error        error
}

// LifecycleManager manages the lifecycle of components
type LifecycleManager interface {
	// RegisterComponent registers a component for lifecycle management
	RegisterComponent(component Component) error

	// UnregisterComponent removes a component from lifecycle management
	UnregisterComponent(name string) error

	// Start starts all components in dependency order
	Start(ctx context.Context) error

	// Stop stops all components in reverse dependency order
	Stop(ctx context.Context) error

	// AddHook adds a lifecycle hook for a specific phase
	AddHook(phase Phase, hook Hook) error

	// RemoveHook removes a lifecycle hook
	RemoveHook(phase Phase, hook Hook) error

	// GetComponentState returns the state of a specific component
	GetComponentState(name string) (ComponentState, bool)

	// GetAllComponentStates returns the state of all components
	GetAllComponentStates() map[string]ComponentState

	// GetPhase returns the current lifecycle phase
	GetPhase() Phase

	// HealthCheck performs a health check on all components
	HealthCheck(ctx context.Context) map[string]ComponentHealth
}

// ComponentOption provides options for component configuration
type ComponentOption func(*ComponentConfig)

// ComponentConfig holds configuration for a component
type ComponentConfig struct {
	Name         string
	Dependencies []string
	Priority     int
	Timeout      time.Duration
	Retries      int
	HealthCheck  func(ctx context.Context) ComponentHealth
}

// WithDependencies sets component dependencies
func WithDependencies(deps ...string) ComponentOption {
	return func(c *ComponentConfig) {
		c.Dependencies = deps
	}
}

// WithPriority sets component priority
func WithPriority(priority int) ComponentOption {
	return func(c *ComponentConfig) {
		c.Priority = priority
	}
}

// WithTimeout sets component timeout
func WithTimeout(timeout time.Duration) ComponentOption {
	return func(c *ComponentConfig) {
		c.Timeout = timeout
	}
}

// WithRetries sets component retry count
func WithRetries(retries int) ComponentOption {
	return func(c *ComponentConfig) {
		c.Retries = retries
	}
}

// WithHealthCheck sets custom health check function
func WithHealthCheck(healthCheck func(ctx context.Context) ComponentHealth) ComponentOption {
	return func(c *ComponentConfig) {
		c.HealthCheck = healthCheck
	}
}
