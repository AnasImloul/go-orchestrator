package di

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/AnasImloul/go-orchestrator/internal/logger"
)

// DefaultContainer implements the Container interface
type DefaultContainer struct {
	registrations map[reflect.Type]*ServiceRegistration
	namedServices map[string]*ServiceRegistration
	singletons    map[reflect.Type]interface{}
	config        ContainerConfig
	logger        logger.Logger
	mu            sync.RWMutex
	disposed      bool
}


// NewContainer creates a new DI container
func NewContainer(config ContainerConfig, logger logger.Logger) *DefaultContainer {
	return &DefaultContainer{
		registrations: make(map[reflect.Type]*ServiceRegistration),
		namedServices: make(map[string]*ServiceRegistration),
		singletons:    make(map[reflect.Type]interface{}),
		config:        config,
		logger:        logger,
	}
}

// Register registers a service with the container
func (c *DefaultContainer) Register(serviceType reflect.Type, factory Factory, options ...Option) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.disposed {
		return fmt.Errorf("container is disposed")
	}

	// Apply options
	opts := ServiceOptions{}
	for _, option := range options {
		option(&opts)
	}

	registration := &ServiceRegistration{
		ServiceType: serviceType,
		Name:        opts.Name,
		Factory:     factory,
		Lifetime:    c.config.DefaultLifetime,
		Options:     opts,
	}

	// Validate registration if enabled
	if c.config.EnableValidation {
		if err := c.validateRegistration(*registration); err != nil {
			return fmt.Errorf("registration validation failed: %w", err)
		}
	}

	// Register by type
	c.registrations[serviceType] = registration

	// Register by name if provided
	if opts.Name != "" {
		c.namedServices[opts.Name] = registration
	}

	// Record metrics if enabled
	if c.config.EnableMetrics && c.config.MetricsProvider != nil {
		c.config.MetricsProvider.RecordRegistration(serviceType, registration.Lifetime)
	}

	if c.logger != nil {
		c.logger.Debug("Service registered",
			"type", serviceType.String(),
			"name", opts.Name,
			"lifetime", registration.Lifetime,
		)
	}

	return nil
}

// RegisterInstance registers a service instance
func (c *DefaultContainer) RegisterInstance(serviceType reflect.Type, instance interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.disposed {
		return fmt.Errorf("container is disposed")
	}

	registration := &ServiceRegistration{
		ServiceType: serviceType,
		Instance:    instance,
		Lifetime:    Singleton,
	}

	c.registrations[serviceType] = registration
	c.singletons[serviceType] = instance

	if c.logger != nil {
		c.logger.Debug("Service instance registered",
			"type", serviceType.String(),
		)
	}

	return nil
}

// RegisterSingleton registers a singleton service
func (c *DefaultContainer) RegisterSingleton(serviceType reflect.Type, factory Factory, options ...Option) error {
	// Apply options and set lifetime to Singleton
	opts := ServiceOptions{}
	for _, option := range options {
		option(&opts)
	}

	return c.register(serviceType, factory, Singleton, opts)
}

// Resolve resolves a service from the container
func (c *DefaultContainer) Resolve(serviceType reflect.Type) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.disposed {
		return nil, fmt.Errorf("container is disposed")
	}

	// Check resolution depth to prevent infinite recursion
	if c.config.MaxResolutionDepth > 0 {
		// For public API, we start with depth 0
		return c.resolve(context.Background(), serviceType, 0)
	}

	return c.resolve(context.Background(), serviceType, 0)
}

// ResolveByName resolves a service by name
func (c *DefaultContainer) ResolveByName(name string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.disposed {
		return nil, fmt.Errorf("container is disposed")
	}

	registration, exists := c.namedServices[name]
	if !exists {
		return nil, fmt.Errorf("service with name '%s' not found", name)
	}

	return c.resolve(context.Background(), registration.ServiceType, 0)
}

// TryResolve attempts to resolve a service, returns false if not found
func (c *DefaultContainer) TryResolve(serviceType reflect.Type) (interface{}, bool) {
	instance, err := c.Resolve(serviceType)
	return instance, err == nil
}

// Contains checks if a service is registered
func (c *DefaultContainer) Contains(serviceType reflect.Type) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, exists := c.registrations[serviceType]
	return exists
}

// ContainsByName checks if a named service is registered
func (c *DefaultContainer) ContainsByName(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, exists := c.namedServices[name]
	return exists
}

// GetRegistrations returns all service registrations
func (c *DefaultContainer) GetRegistrations() []ServiceRegistration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	registrations := make([]ServiceRegistration, 0, len(c.registrations))
	for _, reg := range c.registrations {
		registrations = append(registrations, *reg)
	}

	return registrations
}

// CreateScope creates a new scope
func (c *DefaultContainer) CreateScope() Scope {
	return NewScope(c, c.logger)
}

// Dispose disposes the container and all its resources
func (c *DefaultContainer) Dispose() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.disposed {
		return nil
	}

	c.disposed = true

	// Dispose all singletons that implement disposable interface
	// Skip disposing the container itself to prevent recursive disposal
	for serviceType, instance := range c.singletons {
		// Skip if this is the container itself to prevent recursive disposal
		if instance == c {
			continue
		}
		
		if disposable, ok := instance.(Disposable); ok {
			if err := disposable.Dispose(); err != nil {
				if c.logger != nil {
					c.logger.Error("Failed to dispose singleton",
						"type", serviceType.String(),
						"error", err.Error(),
					)
				}
			}
		}
	}

	// Clear all collections
	c.registrations = nil
	c.namedServices = nil
	c.singletons = nil

	if c.logger != nil {
		c.logger.Info("Container disposed")
	}

	return nil
}

// Private helper methods

// register is the internal registration method
func (c *DefaultContainer) register(serviceType reflect.Type, factory Factory, lifetime ServiceLifetime, opts ServiceOptions) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.disposed {
		return fmt.Errorf("container is disposed")
	}

	registration := &ServiceRegistration{
		ServiceType: serviceType,
		Name:        opts.Name,
		Factory:     factory,
		Lifetime:    lifetime,
		Options:     opts,
	}

	// Validate registration if enabled
	if c.config.EnableValidation {
		if err := c.validateRegistration(*registration); err != nil {
			return fmt.Errorf("registration validation failed: %w", err)
		}
	}

	// Register by type
	c.registrations[serviceType] = registration

	// Register by name if provided
	if opts.Name != "" {
		c.namedServices[opts.Name] = registration
	}

	// Record metrics if enabled
	if c.config.EnableMetrics && c.config.MetricsProvider != nil {
		c.config.MetricsProvider.RecordRegistration(serviceType, lifetime)
	}

	if c.logger != nil {
		c.logger.Debug("Service registered",
			"type", serviceType.String(),
			"name", opts.Name,
			"lifetime", lifetime,
		)
	}

	return nil
}

// resolve is the internal resolution method
func (c *DefaultContainer) resolve(ctx context.Context, serviceType reflect.Type, depth int) (interface{}, error) {
	start := time.Now()
	var success bool
	defer func() {
		duration := time.Since(start).Nanoseconds()
		if c.config.EnableMetrics && c.config.MetricsProvider != nil {
			c.config.MetricsProvider.RecordResolution(serviceType, duration, success)
		}
	}()

	// Check resolution depth to prevent infinite recursion
	if depth > c.config.MaxResolutionDepth {
		return nil, fmt.Errorf("maximum resolution depth exceeded for type %s", serviceType.String())
	}

	// Find registration
	registration, exists := c.registrations[serviceType]
	if !exists {
		return nil, fmt.Errorf("service of type %s is not registered", serviceType.String())
	}

	// Handle instance registration
	if registration.Instance != nil {
		success = true
		return registration.Instance, nil
	}

	// Handle singleton lifetime
	if registration.Lifetime == Singleton {
		if instance, exists := c.singletons[serviceType]; exists {
			success = true
			return instance, nil
		}
	}

	// Create instance using factory
	instance, err := c.createInstance(ctx, registration, depth)
	if err != nil {
		return nil, err
	}

	// Store singleton
	if registration.Lifetime == Singleton {
		c.singletons[serviceType] = instance
	}

	// Validate instance if enabled
	if c.config.EnableValidation {
		if err := c.validateInstance(serviceType, instance); err != nil {
			return nil, fmt.Errorf("instance validation failed: %w", err)
		}
	}

	success = true
	return instance, nil
}

// createInstance creates a service instance using the factory
func (c *DefaultContainer) createInstance(ctx context.Context, registration *ServiceRegistration, depth int) (instance interface{}, err error) {
	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic during instance creation for type %s: %v", registration.ServiceType.String(), r)
		}
	}()

	if registration.Factory == nil {
		return nil, fmt.Errorf("no factory provided for service %s", registration.ServiceType.String())
	}

	// Apply interceptors if enabled
	if c.config.EnableInterception && len(registration.Options.Interceptors) > 0 {
		return c.applyInterceptors(ctx, registration, depth)
	}

	// Use retry logic if configured
	if registration.Options.RetryConfig != nil {
		var result interface{}
		retryErr := RetryWithBackoff(ctx, *registration.Options.RetryConfig, func() error {
			var err error
			result, err = registration.Factory(ctx, c)
			return err
		})
		
		if retryErr != nil {
			return nil, retryErr
		}
		return result, nil
	}

	// Create instance directly
	return registration.Factory(ctx, c)
}

// applyInterceptors applies interceptors to service creation
func (c *DefaultContainer) applyInterceptors(ctx context.Context, registration *ServiceRegistration, depth int) (interface{}, error) {
	interceptors := registration.Options.Interceptors
	if len(interceptors) == 0 {
		return registration.Factory(ctx, c)
	}

	// Create interceptor chain
	var next func() (interface{}, error)
	next = func() (interface{}, error) {
		return registration.Factory(ctx, c)
	}

	// Apply interceptors in reverse order
	for i := len(interceptors) - 1; i >= 0; i-- {
		interceptor := interceptors[i]
		currentNext := next
		next = func() (interface{}, error) {
			return interceptor.Intercept(ctx, registration.ServiceType, currentNext)
		}
	}

	return next()
}

// validateRegistration validates a service registration
func (c *DefaultContainer) validateRegistration(registration ServiceRegistration) error {
	// Check for nil factory when needed
	if registration.Factory == nil && registration.Instance == nil {
		return fmt.Errorf("either factory or instance must be provided")
	}

	// Check for circular dependencies if enabled
	if c.config.EnableCircularCheck {
		if err := c.checkCircularDependencies(registration.ServiceType, registration.Options.Dependencies); err != nil {
			return fmt.Errorf("circular dependency detected: %w", err)
		}
	}

	return nil
}

// checkCircularDependencies performs a comprehensive circular dependency check
func (c *DefaultContainer) checkCircularDependencies(serviceType reflect.Type, dependencies []reflect.Type) error {
	visited := make(map[reflect.Type]bool)
	visiting := make(map[reflect.Type]bool)
	
	// Start DFS from the current service type
	return c.dfsCircularCheck(serviceType, visited, visiting)
}

// dfsCircularCheck performs depth-first search to detect cycles
// Note: This function assumes the caller already holds the write lock
func (c *DefaultContainer) dfsCircularCheck(serviceType reflect.Type, visited, visiting map[reflect.Type]bool) error {
	if visiting[serviceType] {
		return fmt.Errorf("circular dependency detected involving type %s", serviceType.String())
	}
	
	if visited[serviceType] {
		return nil
	}
	
	visiting[serviceType] = true
	
	// Get registration for this service type (no lock needed since caller holds write lock)
	registration, exists := c.registrations[serviceType]
	
	if exists {
		// Check dependencies of this service
		for _, dep := range registration.Options.Dependencies {
			if err := c.dfsCircularCheck(dep, visited, visiting); err != nil {
				return err
			}
		}
	}
	
	visiting[serviceType] = false
	visited[serviceType] = true
	
	return nil
}

// validateInstance validates a resolved service instance
func (c *DefaultContainer) validateInstance(serviceType reflect.Type, instance interface{}) error {
	if instance == nil {
		return fmt.Errorf("factory returned nil instance for type %s", serviceType.String())
	}

	instanceType := reflect.TypeOf(instance)

	// Check if serviceType is an interface
	if serviceType.Kind() == reflect.Interface {
		// For interfaces, check if the instance implements the interface
		if !instanceType.Implements(serviceType) {
			return fmt.Errorf("instance of type %s does not implement interface %s", instanceType.String(), serviceType.String())
		}
	} else {
		// For concrete types, check if types match or if instance is assignable to serviceType
		if !instanceType.AssignableTo(serviceType) {
			return fmt.Errorf("instance of type %s is not assignable to %s", instanceType.String(), serviceType.String())
		}
	}

	return nil
}

// Disposable interface for resources that need cleanup
type Disposable interface {
	Dispose() error
}

// Generic helper functions for type-safe dependency resolution

// TypeOf returns the reflect.Type for a given type T
func TypeOf[T any]() reflect.Type {
	var p *T
	return reflect.TypeOf(p).Elem()
}

// Resolve resolves a service of type T from the container
func Resolve[T any](c Container) (T, error) {
	var zero T
	v, err := c.Resolve(TypeOf[T]())
	if err != nil {
		return zero, fmt.Errorf("failed to resolve %v: %w", TypeOf[T](), err)
	}
	t, ok := v.(T)
	if !ok {
		return zero, fmt.Errorf("wrong type: have %T, want %v", v, TypeOf[T]())
	}
	return t, nil
}

// MustResolve resolves a service of type T from the container, panicking on error
func MustResolve[T any](c Container) T {
	v, err := Resolve[T](c)
	if err != nil {
		panic(fmt.Sprintf("failed to resolve %v: %v", TypeOf[T](), err))
	}
	return v
}

// TryResolve attempts to resolve a service of type T from the container
func TryResolve[T any](c Container) (T, bool) {
	var zero T
	v, err := Resolve[T](c)
	if err != nil {
		return zero, false
	}
	return v, true
}
