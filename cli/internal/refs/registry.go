package refs

import (
	"context"
	"fmt"
	"sync"
)

// registry holds all registered reference types.
var (
	registry   = make(map[string]RefType)
	registryMu sync.RWMutex
)

// Register registers a new reference type.
// Panics if a type with the same name is already registered.
//
// This is typically called in init() functions of type implementations.
func Register(t RefType) {
	registryMu.Lock()
	defer registryMu.Unlock()

	name := t.Name()
	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("ref type already registered: %s", name))
	}

	registry[name] = t
}

// GetType returns a registered type by name.
// Returns an error if the type is not registered.
func GetType(name string) (RefType, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	t, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown ref type: %s", name)
	}

	return t, nil
}

// AllTypes returns all registered types.
func AllTypes() []RefType {
	registryMu.RLock()
	defer registryMu.RUnlock()

	types := make([]RefType, 0, len(registry))
	for _, t := range registry {
		types = append(types, t)
	}

	return types
}

// EnabledTypes returns all types that are enabled on this system.
// A type is enabled if its IsEnabled() method returns true.
func EnabledTypes(ctx context.Context) ([]RefType, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	var enabled []RefType
	for _, t := range registry {
		ok, err := t.IsEnabled(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to check if %s enabled: %w", t.Name(), err)
		}
		if ok {
			enabled = append(enabled, t)
		}
	}

	return enabled, nil
}

// DisabledTypes returns all types that are disabled on this system.
// A type is disabled if its IsEnabled() method returns false.
func DisabledTypes(ctx context.Context) ([]RefType, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	var disabled []RefType
	for _, t := range registry {
		ok, err := t.IsEnabled(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to check if %s enabled: %w", t.Name(), err)
		}
		if !ok {
			disabled = append(disabled, t)
		}
	}

	return disabled, nil
}

// TypeForScheme returns the type that handles the given URL scheme.
// Returns an error if no type handles the scheme.
//
// Example: "git+https" returns the git type.
func TypeForScheme(_ context.Context, scheme string) (RefType, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()

	// Parse scheme to determine type
	typeName := InferTypeFromScheme(scheme)

	t, ok := registry[typeName]
	if !ok {
		return nil, fmt.Errorf("no type registered for scheme: %s", scheme)
	}

	return t, nil
}
