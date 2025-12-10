package state

import (
	"fmt"
	"sort"
	"sync"
)

// registry is the global project type configuration registry.
// It maps project type names to their configurations.
// Access is thread-safe via the registry mutex.
var (
	registry   = make(map[string]ProjectTypeConfig)
	registryMu sync.RWMutex
)

// Register adds a project type configuration to the global registry.
// Panics if a project type with the same name is already registered.
// This prevents accidental duplicate registrations which could cause
// non-deterministic behavior.
//
// Typical usage in project type packages:
//
//	func init() {
//	    project.Register("standard", BuildStandardConfig())
//	}
func Register(typeName string, config ProjectTypeConfig) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := registry[typeName]; exists {
		panic(fmt.Sprintf("project type already registered: %s", typeName))
	}
	registry[typeName] = config
}

// GetConfig retrieves a project type configuration from the registry.
// Returns (config, true) if found, (nil, false) if not found.
func GetConfig(typeName string) (ProjectTypeConfig, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	config, exists := registry[typeName]
	return config, exists
}

// RegisteredTypes returns a list of all registered type names.
// Useful for documentation and CLI help text.
// Returns names in sorted order for deterministic output.
func RegisteredTypes() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	types := make([]string, 0, len(registry))
	for name := range registry {
		types = append(types, name)
	}
	sort.Strings(types)
	return types
}

// RegisterConfig registers a project type configuration in the global registry.
// This is a convenience wrapper around Register that uses config.Name() as the type name.
// Panics if a config with the same name already exists.
func RegisterConfig(config ProjectTypeConfig) {
	Register(config.Name(), config)
}

// ClearRegistry removes all registered configurations.
// This is primarily intended for testing.
func ClearRegistry() {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry = make(map[string]ProjectTypeConfig)
}
