package state

import (
	"fmt"

	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
)

// Registry maps project type names to their configurations.
// This is a global registry that is populated during initialization.
// The configs are stored as ProjectTypeConfig (interface{}) to avoid import cycles.
var Registry = make(map[string]ProjectTypeConfig)

// Register adds a project type configuration to the global registry.
// Panics if a project type with the same name is already registered.
// This prevents accidental duplicate registrations which could cause
// non-deterministic behavior.
//
// Typical usage in project type packages:
//
//	func init() {
//	    Register("standard", NewStandardProjectConfig())
//	}
func Register(typeName string, config ProjectTypeConfig) {
	if _, exists := Registry[typeName]; exists {
		panic(fmt.Sprintf("project type already registered: %s", typeName))
	}
	Registry[typeName] = config
}

// Get retrieves a project type configuration from the registry.
// Returns (config, true) if found, (nil, false) if not found.
//
// Used by Load() to attach project type config to loaded project:
//
//	config, exists := Registry[project.Type]
//	if !exists {
//	    return fmt.Errorf("unknown project type: %s", project.Type)
//	}
func Get(typeName string) (ProjectTypeConfig, bool) {
	config, exists := Registry[typeName]
	return config, exists
}

// State is a type alias for SDK state type for backward compatibility.
type State = sdkstate.State
