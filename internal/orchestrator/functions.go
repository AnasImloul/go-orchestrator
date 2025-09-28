package orchestrator

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/AnasImloul/go-orchestrator/internal/lifecycle"
)

// autoDiscoverDependenciesTyped automatically discovers dependencies from factory function parameters for typed service definitions.
// Only adds dependencies that are likely to be lifecycle components (interfaces that implement Service).
func autoDiscoverDependenciesTyped[T any](serviceDef *TypedServiceDefinition[T], factory interface{}) {
	factoryValue := reflect.ValueOf(factory)
	factoryType := factoryValue.Type()

	// Get parameter types from the factory function
	for i := 0; i < factoryType.NumIn(); i++ {
		paramType := factoryType.In(i)
		
		// Only add as lifecycle dependency if it's likely to be a service
		// Check if it's an interface type or if it implements the Service interface
		if isLikelyService(paramType) {
			dependencyName := typeToDependencyName(paramType)
			serviceDef.Dependencies = append(serviceDef.Dependencies, dependencyName)
		}
		// Struct types are handled by DI container resolution, not lifecycle dependencies
	}
}

// isLikelyService determines if a type is likely to be a service that needs lifecycle management
func isLikelyService(paramType reflect.Type) bool {
	// Remove pointer if present
	if paramType.Kind() == reflect.Ptr {
		paramType = paramType.Elem()
	}
	
	// If it's an interface, it's likely a service
	if paramType.Kind() == reflect.Interface {
		return true
	}
	
	// If it's a struct, check if it implements the Service interface
	if paramType.Kind() == reflect.Struct {
		serviceInterface := reflect.TypeOf((*Service)(nil)).Elem()
		return paramType.Implements(serviceInterface)
	}
	
	// For other types (like logger.Logger), don't treat as lifecycle dependency
	return false
}

// inferServiceNameFromType automatically infers a robust service name from a reflect.Type.
// It creates unique names by including the full package path to avoid conflicts.
// Uses standard naming convention: "package::ServiceName"
// Examples:
//   - "github.com/user/pkg1.DatabaseService" -> "github.com/user/pkg1::DatabaseService"
//   - "github.com/user/pkg2.DatabaseService" -> "github.com/user/pkg2::DatabaseService"
func inferServiceNameFromType(serviceType reflect.Type) string {
	if serviceType == nil {
		return "service"
	}

	typeName := serviceType.String()

	// Split package path and type name
	lastDot := strings.LastIndex(typeName, ".")
	if lastDot == -1 {
		// No package path, just use the type name
		return sanitizeServiceName(typeName)
	}

	packagePath := typeName[:lastDot]
	typeNameOnly := typeName[lastDot+1:]

	// Clean up the type name
	typeNameClean := sanitizeServiceName(typeNameOnly)

	// Use standard format: package::ServiceName
	if packagePath != "" {
		return packagePath + "::" + typeNameClean
	}
	return typeNameClean
}

// sanitizeServiceName cleans up a type name to create a valid service name.
// Preserves the original Go naming convention (PascalCase).
func sanitizeServiceName(typeName string) string {
	// Keep original case for Go naming convention
	serviceName := typeName

	// Remove "Service" suffix if present (case-insensitive)
	if strings.HasSuffix(strings.ToLower(serviceName), "service") {
		serviceName = serviceName[:len(serviceName)-7] // Remove "Service"
	}

	// Remove "Interface" suffix if present (case-insensitive)
	if strings.HasSuffix(strings.ToLower(serviceName), "interface") {
		serviceName = serviceName[:len(serviceName)-9] // Remove "Interface"
	}

	// Remove any non-alphanumeric characters except underscores
	var result strings.Builder
	for _, r := range serviceName {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			result.WriteRune(r)
		}
	}

	serviceName = result.String()

	// Ensure we have a valid name
	if serviceName == "" {
		return "Service"
	}

	return serviceName
}

// mapHealthStatus converts a HealthStatusType to the lifecycle HealthStatus enum
func mapHealthStatus(status HealthStatusType) lifecycle.HealthStatus {
	switch status {
	case HealthStatusHealthy:
		return lifecycle.HealthStatusHealthy
	case HealthStatusDegraded:
		return lifecycle.HealthStatusDegraded
	case HealthStatusUnhealthy:
		return lifecycle.HealthStatusUnhealthy
	default:
		return lifecycle.HealthStatusUnknown
	}
}

// typeToDependencyName converts a Go type to a dependency name.
// Uses the same robust naming strategy as service registration to ensure consistency.
func typeToDependencyName(paramType reflect.Type) string {
	// Remove pointer prefix if present
	if paramType.Kind() == reflect.Ptr {
		paramType = paramType.Elem()
	}

	// Use the same naming strategy as service registration
	return inferServiceNameFromType(paramType)
}

// callFactoryWithAutoDependencies uses reflection to automatically resolve dependencies
// and call the factory function with the resolved dependencies.
func callFactoryWithAutoDependencies[T any](ctx context.Context, container *Container, factory interface{}) (T, error) {
	var zero T

	factoryValue := reflect.ValueOf(factory)
	factoryType := factoryValue.Type()

	// Check if it's a function
	if factoryType.Kind() != reflect.Func {
		return zero, fmt.Errorf("factory must be a function, got %s", factoryType.Kind())
	}

	// Get the number of parameters
	numIn := factoryType.NumIn()
	args := make([]reflect.Value, numIn)

	// Resolve each parameter as a dependency
	for i := 0; i < numIn; i++ {
		paramType := factoryType.In(i)

		// Try to resolve the dependency from the container
		dependency, err := container.Resolve(paramType)
		if err != nil {
			return zero, fmt.Errorf("failed to resolve dependency %d of type %s: %w", i, paramType.String(), err)
		}

		args[i] = reflect.ValueOf(dependency)
	}

	// Call the factory function
	results := factoryValue.Call(args)

	// Handle the return values
	if len(results) == 0 {
		return zero, fmt.Errorf("factory function must return at least one value")
	}

	// Check if the first result is an error
	if len(results) > 1 {
		if err, ok := results[1].Interface().(error); ok && err != nil {
			return zero, err
		}
	}

	// Return the first result
	return results[0].Interface().(T), nil
}

// resolveDependenciesAndCallFactory uses reflection to automatically resolve dependencies
// and call the original factory function with the resolved dependencies.
func resolveDependenciesAndCallFactory(ctx context.Context, container *Container, factory interface{}) (interface{}, error) {
	factoryValue := reflect.ValueOf(factory)
	factoryType := factoryValue.Type()

	// Check if it's a function
	if factoryType.Kind() != reflect.Func {
		return nil, fmt.Errorf("factory must be a function, got %s", factoryType.Kind())
	}

	// Get the number of parameters (should be 2: context.Context and *Container)
	numIn := factoryType.NumIn()
	if numIn != 2 {
		return nil, fmt.Errorf("factory function must have exactly 2 parameters (context.Context, *Container), got %d", numIn)
	}

	// Verify the first parameter is context.Context
	if !factoryType.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		return nil, fmt.Errorf("first parameter must be context.Context")
	}

	// Verify the second parameter is *Container
	if factoryType.In(1) != reflect.TypeOf((*Container)(nil)) {
		return nil, fmt.Errorf("second parameter must be *Container")
	}

	// Call the original factory with the provided parameters
	args := []reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(container),
	}

	results := factoryValue.Call(args)

	// Handle the return values (should be (interface{}, error))
	if len(results) != 2 {
		return nil, fmt.Errorf("factory function must return exactly 2 values (interface{}, error), got %d", len(results))
	}

	// Check for error
	if !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	return results[0].Interface(), nil
}
