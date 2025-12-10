# Task 080: Migrate Registry and Validation

## Context

This task is part of the `libs/project` module consolidation effort. It migrates the project type registry and CUE validation from `cli/internal/sdks/project/state/` to `libs/project/state/`.

The registry provides:
- Global registration of project type configurations
- Lookup by type name for Load/Create operations
- Panic on duplicate registration (early error detection)

The validation provides:
- CUE-based structural validation of ProjectState
- Metadata validation against type-specific schemas
- Artifact type validation for phases

## Requirements

### 1. Migrate Registry (state/registry.go)

Migrate the registry from `cli/internal/sdks/project/state/registry.go`:

```go
// registry holds registered project type configurations.
var registry = make(map[string]*ProjectTypeConfig)

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
func Register(typeName string, config *ProjectTypeConfig)

// GetConfig retrieves a project type configuration from the registry.
// Returns (config, true) if found, (nil, false) if not found.
func GetConfig(typeName string) (*ProjectTypeConfig, bool)

// RegisteredTypes returns a list of all registered type names.
// Useful for documentation and CLI help text.
func RegisteredTypes() []string
```

**Note**: The registry needs to reference `ProjectTypeConfig` which is in the parent package. Use the interface approach established in Task 040/060 or accept the parent package's type directly if import cycles can be avoided.

### 2. Migrate Structure Validation (state/validate.go)

Migrate CUE validation from `cli/internal/sdks/project/state/validate.go`:

```go
// projectSchemas holds the loaded CUE value for the project subpackage.
var projectSchemas cue.Value

// init loads the project subpackage schemas from the embedded filesystem.
func init()

// validateStructure performs CUE-based structural validation on a ProjectState.
// It validates universal fields (name, type, status, etc.), regex patterns,
// required fields, and collection structures.
//
// This validation is run on both Load() and Save() operations to ensure
// structural integrity of the project state.
func validateStructure(projectState *project.ProjectState) error
```

### 3. Migrate Metadata Validation

Migrate metadata validation functions:

```go
// ValidateMetadata validates a metadata map against a CUE schema string.
// This is used for project-type-specific metadata validation on phases and tasks.
//
// If cueSchema is empty, validation is skipped (no schema = no validation).
func ValidateMetadata(metadata map[string]interface{}, cueSchema string) error

// ValidateArtifactTypes checks if artifact types are in the allowed list.
// Empty allowed list means "allow all types" (no validation).
func ValidateArtifactTypes(
    artifacts []project.ArtifactState,
    allowedTypes []string,
    phaseName string,
    category string, // "input" or "output"
) error
```

### 4. CUE Schema Loading

Load CUE schemas from the embedded filesystem in `libs/schemas`:

```go
func init() {
    // Create in-memory filesystem
    memFS := billy.NewMemory()

    // Copy embedded schemas to in-memory filesystem
    if err := core.CopyFromEmbedFS(schemas.CUESchemas, memFS, "."); err != nil {
        panic(fmt.Errorf("failed to copy embedded schemas: %w", err))
    }

    // Create loader and load the project subpackage
    loader := cuepkg.NewLoader(memFS)
    var err error
    projectSchemas, err = loader.LoadPackage(context.Background(), "project")
    if err != nil {
        panic(fmt.Errorf("failed to load project schemas: %w", err))
    }
}
```

### 5. Error Types

Define validation-specific errors:

```go
var (
    // ErrValidationFailed indicates project state validation failed.
    ErrValidationFailed = errors.New("validation failed")

    // ErrInvalidArtifactType indicates an artifact type is not allowed.
    ErrInvalidArtifactType = errors.New("invalid artifact type")
)
```

## Acceptance Criteria

1. [ ] `state/registry.go` provides Register and GetConfig functions
2. [ ] `state/validate.go` provides validateStructure function
3. [ ] ValidateMetadata validates against CUE schemas
4. [ ] ValidateArtifactTypes validates artifact types
5. [ ] CUE schemas load correctly from embedded filesystem
6. [ ] Registry panics on duplicate registration
7. [ ] Validation errors are descriptive and actionable
8. [ ] All functions are properly documented with doc comments
9. [ ] Code compiles without errors
10. [ ] `golangci-lint run ./...` passes with no issues
11. [ ] `go test -race ./...` passes with no failures
12. [ ] Code adheres to `.standards/STYLE.md` (file organization, naming, error handling)
13. [ ] Tests adhere to `.standards/TESTING.md` (table-driven, testify assertions, behavioral coverage)

### Test Requirements (TDD)

**state/registry_test.go:**
- Register adds config to registry
- Register panics on duplicate
- GetConfig returns registered config
- GetConfig returns false for unknown type
- RegisteredTypes returns all registered names

**state/validate_test.go:**
- validateStructure passes for valid state
- validateStructure fails for missing required fields
- validateStructure fails for invalid field values
- ValidateMetadata passes for valid metadata
- ValidateMetadata skips if schema is empty
- ValidateMetadata fails for invalid metadata
- ValidateArtifactTypes passes for allowed types
- ValidateArtifactTypes skips if allowed list empty
- ValidateArtifactTypes fails for disallowed types

**Integration tests:**
- Create project, validate, save, reload, validate again
- Round-trip validation preserves all fields

## Technical Details

### Import Dependencies

```go
import (
    "context"
    "errors"
    "fmt"

    "cuelang.org/go/cue"
    "cuelang.org/go/cue/cuecontext"
    cuepkg "github.com/jmgilman/go/cue"
    "github.com/jmgilman/go/fs/billy"
    "github.com/jmgilman/go/fs/core"
    "github.com/jmgilman/sow/libs/schemas"
    "github.com/jmgilman/sow/libs/schemas/project"
)
```

### Registry Type Reference

The registry needs to store `ProjectTypeConfig`. Options:

**Option 1: Use interface**
```go
// In state/registry.go
type ProjectTypeConfigInterface interface {
    Name() string
    InitialState() State
    // etc.
}

var registry = make(map[string]ProjectTypeConfigInterface)
```

**Option 2: Store as interface{} and type assert**
```go
var registry = make(map[string]interface{})

func GetConfig(typeName string) (interface{}, bool) {
    config, exists := registry[typeName]
    return config, exists
}
```

**Option 3: Move registry to parent package**
```go
// In project/registry.go (parent package)
var registry = make(map[string]*ProjectTypeConfig)
```

**Recommended: Option 3** - The registry belongs in the parent `project` package alongside `ProjectTypeConfig`. This avoids import cycles since state package imports parent package (not vice versa).

### Validation Flow

```
Load()
  ├── backend.Load() -> ProjectState
  ├── validateStructure(ProjectState) -> CUE validation
  ├── project.GetConfig(type) -> ProjectTypeConfig
  ├── config.Initialize() if creating
  ├── config.BuildMachine()
  └── config.Validate() -> Metadata validation

Save()
  ├── sync machine state
  ├── update timestamps
  ├── validateStructure(ProjectState) -> CUE validation
  ├── config.Validate() -> Metadata validation
  └── backend.Save(ProjectState)
```

### CUE Validation Example

```go
func validateStructure(projectState *project.ProjectState) error {
    ctx := cuecontext.New()

    // Lookup the ProjectState schema
    schema := projectSchemas.LookupPath(cue.ParsePath("#ProjectState"))
    if schema.Err() != nil {
        return fmt.Errorf("schema lookup failed: %w", schema.Err())
    }

    // Encode the project state
    value := ctx.Encode(projectState)
    if value.Err() != nil {
        return fmt.Errorf("encode failed: %w", value.Err())
    }

    // Unify and validate
    result := schema.Unify(value)
    if err := result.Validate(cue.Concrete(true)); err != nil {
        return fmt.Errorf("%w: %s", ErrValidationFailed, err)
    }

    return nil
}
```

## Relevant Inputs

- `cli/internal/sdks/project/state/registry.go` - Current registry implementation
- `cli/internal/sdks/project/state/validate.go` - Current validation implementation
- `libs/schemas/project/cue_types_gen.go` - ProjectState type
- `libs/schemas/embed.go` - Embedded CUE schemas
- `.standards/STYLE.md` - Code style requirements
- `.standards/TESTING.md` - Testing requirements

## Examples

### Registering a Project Type

```go
// In cli/internal/projects/standard/standard.go
func init() {
    config := BuildStandardConfig()
    project.Register("standard", config)
}
```

### Looking Up a Config

```go
// In loader.go
func Load(ctx context.Context, backend Backend) (*Project, error) {
    projectState, err := backend.Load(ctx)
    if err != nil {
        return nil, err
    }

    if err := validateStructure(projectState); err != nil {
        return nil, fmt.Errorf("invalid structure: %w", err)
    }

    config, exists := GetConfig(projectState.Type)
    if !exists {
        return nil, fmt.Errorf("unknown project type: %s", projectState.Type)
    }

    // ...
}
```

### Validating Metadata

```go
// In config.Validate()
func (ptc *ProjectTypeConfig) Validate(p *state.Project) error {
    for phaseName, phase := range p.Phases {
        phaseConfig := ptc.Phases()[phaseName]
        if phaseConfig == nil {
            continue
        }

        // Validate phase metadata
        if err := state.ValidateMetadata(phase.Metadata, phaseConfig.MetadataSchema()); err != nil {
            return fmt.Errorf("phase %s metadata: %w", phaseName, err)
        }

        // Validate artifact types
        if err := state.ValidateArtifactTypes(
            phase.Inputs,
            phaseConfig.AllowedInputTypes(),
            phaseName,
            "input",
        ); err != nil {
            return err
        }
    }
    return nil
}
```

### Testing Validation

```go
func TestValidateStructure_ValidState(t *testing.T) {
    state := &project.ProjectState{
        Name:   "valid-project",
        Type:   "standard",
        Branch: "feat/test",
        Phases: map[string]project.PhaseState{},
        Statechart: project.StatechartState{
            Current_state: "PlanningActive",
        },
    }

    err := validateStructure(state)
    assert.NoError(t, err)
}

func TestValidateStructure_InvalidName(t *testing.T) {
    state := &project.ProjectState{
        Name:   "INVALID_NAME", // uppercase not allowed
        Type:   "standard",
        Branch: "feat/test",
    }

    err := validateStructure(state)
    assert.ErrorIs(t, err, ErrValidationFailed)
}
```

## Dependencies

- Task 010: Core types (State, Event)
- Task 040: State wrapper types (for validation targets)
- Task 060: Project config (for registry type)
- libs/schemas module (for embedded CUE and generated types)

## Constraints

- CUE schemas are loaded at package init time - init must not fail
- Registry uses panic for duplicates (fail-fast on programmer error)
- Validation errors should be informative about what failed
- Do NOT change CUE schema definitions - only use existing schemas
- Keep validation functions pure (no side effects)
- Registry is thread-safe for concurrent reads (writes happen at init)
