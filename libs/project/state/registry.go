package state

import (
	"sync"
)

// registry is the global project type configuration registry.
// It maps project type names to their configurations.
// Access is thread-safe via the registry mutex.
var (
	registry   = make(map[string]ProjectTypeConfig)
	registryMu sync.RWMutex
)

// RegisterConfig registers a project type configuration in the global registry.
// If a config with the same name already exists, it will be overwritten.
// This function is typically called during package initialization.
func RegisterConfig(config ProjectTypeConfig) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[config.Name()] = config
}

// GetConfig returns the configuration for a project type.
// Returns the config and true if found, nil and false otherwise.
func GetConfig(typeName string) (ProjectTypeConfig, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	config, exists := registry[typeName]
	return config, exists
}

// ClearRegistry removes all registered configurations.
// This is primarily intended for testing.
func ClearRegistry() {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry = make(map[string]ProjectTypeConfig)
}
