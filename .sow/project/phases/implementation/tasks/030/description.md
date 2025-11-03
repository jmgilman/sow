# Task 030: Registry Implementation

# Task 030: Registry Implementation

## Objective

Implement the global registry for project type registration and lookup. The registry enables project types to be registered at startup (via `init()`) and retrieved during `Load()`.

## Context

**Design Reference:** See `.sow/knowledge/designs/project-sdk-implementation.md`:
- Section "Component Breakdown - Project SDK" (lines 404-408) for registry overview
- Section "Complete Usage Example" (lines 692-695) for registration pattern

**Existing Code:**
- `cli/internal/sdks/project/state/registry.go` exists with stub Registry variable
- Needs Register() and Get() functions added

**Prerequisite:** Task 020 completed (ProjectTypeConfig exists from builder)

**What This Task Builds:**
Global registry that stores project type configurations and provides lookup during project load.

## Requirements

### Registry Functions

Modify `cli/internal/sdks/project/state/registry.go`:

**Current State:**
```go
var Registry = make(map[string]*ProjectTypeConfig)
```

**Add Functions:**

```go
// Register adds a project type configuration to the global registry.
// Panics if a project type with the same name is already registered.
// This prevents accidental duplicate registrations which could cause
// non-deterministic behavior.
//
// Typical usage in project type packages:
//   func init() {
//       Register("standard", NewStandardProjectConfig())
//   }
func Register(typeName string, config *ProjectTypeConfig)
```

```go
// Get retrieves a project type configuration from the registry.
// Returns (config, true) if found, (nil, false) if not found.
//
// Used by Load() to attach project type config to loaded project:
//   config, exists := Registry[project.Type]
//   if !exists {
//       return fmt.Errorf("unknown project type: %s", project.Type)
//   }
func Get(typeName string) (*ProjectTypeConfig, bool)
```

### Implementation Details

**Register():**
- Check if typeName already exists in Registry map
- If exists, panic with message: `"project type already registered: %s"`
- Otherwise, add config to Registry map
- No return value

**Get():**
- Look up typeName in Registry map
- Return (config, true) if found
- Return (nil, false) if not found

## Files to Modify

1. `cli/internal/sdks/project/state/registry.go` - Add Register() and Get() functions
2. `cli/internal/sdks/project/state/registry_test.go` (create) - Behavioral tests

## Testing Requirements (TDD)

Create `cli/internal/sdks/project/state/registry_test.go`:

**Register Tests:**
- Register() adds config to registry
- Register() stores config under correct name
- Register() panics when registering duplicate name
- Multiple different types can be registered

**Get Tests:**
- Get() returns (config, true) for registered type
- Get() returns correct config for registered type
- Get() returns (nil, false) for unregistered type
- Get() works after multiple types registered

**Integration Test:**
- Register type → Get type → verify it's the same config

**Test Pattern for Panic:**
```go
func TestRegisterDuplicatePanics(t *testing.T) {
    // Clear registry for test isolation
    Registry = make(map[string]*ProjectTypeConfig)

    config := &ProjectTypeConfig{name: "test"}
    Register("test", config)

    defer func() {
        if r := recover(); r == nil {
            t.Error("expected panic on duplicate registration")
        }
    }()

    Register("test", config) // Should panic
}
```

**Test Isolation:**
Each test should reset Registry at start:
```go
Registry = make(map[string]*ProjectTypeConfig)
```

## Acceptance Criteria

- [ ] Register() adds config to global Registry map
- [ ] Register() panics with clear message on duplicate registration
- [ ] Get() returns (config, true) for registered types
- [ ] Get() returns (nil, false) for unregistered types
- [ ] Multiple project types can be registered
- [ ] Registry correctly stores and retrieves configs
- [ ] All tests pass (100% coverage of registry behavior)
- [ ] Code compiles without errors

## Dependencies

**Required:** Task 020 (ProjectTypeConfig type exists from builder)

## Technical Notes

- Registry is global mutable state (acceptable for this use case)
- Panic on duplicate is intentional (prevents configuration bugs)
- Registry is populated via init() functions in project type packages
- Get() uses map lookup for O(1) access
- No thread safety needed (populated during startup, read-only after)

**Usage Pattern:**
```go
// In project type package (e.g., internal/projects/standard/standard.go)
func init() {
    state.Register("standard", NewStandardProjectConfig())
}

// In Load() function
config, exists := state.Registry[project.Type]
if !exists {
    return fmt.Errorf("unknown project type: %s", project.Type)
}
```

Note: The existing `registry.go` file uses `Registry` as exported variable. Keep this pattern - Register() and Get() are convenience functions.

## Estimated Time

1 hour
