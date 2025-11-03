# Project SDK Architecture Exploration

**Date:** November 2025
**Branch:** explore/artifact-task-lifecycle
**Status:** Research complete
**Participants:** Josh Gilman, Claude

---

## Context

This exploration investigated how to design a unified Project SDK that would enable:
1. Defining project types through a single fluent builder API
2. Eliminating CUE schema dependency while maintaining validation
3. Supporting universal command structure (set/get/add/remove operations)
4. Enabling future task/artifact state machines without architectural changes

The exploration complements the [Unified Command Hierarchy Design](../designs/command-hierarchy-design.md), which defines the CLI interface. This document defines the internal SDK architecture that powers that CLI.

---

## What We Researched

### 1. State Management Architecture

**Problem identified:** Current architecture has shared mutable state with multiple writers, leading to save conflicts and potential data loss. Multiple components modify the same state tree and save independently, causing last-write-wins scenarios.

**Solutions explored:**
- Separate fields (type safety + partial reuse) - Rejected: field duplication
- Shared pointer + transactions (reuse + type safety) - Rejected: no boundary enforcement
- CUE containers (reuse + flexibility) - Rejected: loss of compile-time safety, high complexity
- **Metadata map (SELECTED)** - Common fields type-safe, custom fields in metadata, CUE validates structure

**Key finding:** Cannot have all three of flexibility, code reuse, and compile-time type safety. Must pick two. We chose flexibility + code reuse, accepting runtime type safety for custom fields (~10% of access).

### 2. CUE Dependency Analysis

**Decision:** Remove CUE dependency (except potentially for metadata validation, later replaced with go-playground/validator).

**Rationale:**
- Standardized project/phase/artifact/task structures eliminate need for polymorphic schema validation
- Go developers more common than CUE experts (lower barrier to entry)
- Single source of truth (Go code) vs dual (Go + CUE schemas)
- Better IDE support (autocomplete, refactoring)
- Simpler mental model for contributors

**Trade-off accepted:** Lose external validation capability (`cue vet state.yaml`). Mitigation: Could generate JSON Schema from SDK config for external tools.

### 3. Collection Pattern

**Three collection types for structural operations:**
- `PhaseCollection` - map[string]*Phase with Get() method
- `ArtifactCollection` - []Artifact with Get(index), Add() methods
- `TaskCollection` - []Task with Get(id) method

**Design principle:** Collections provide structural operations (Get, Add, Remove, bounds checking). Direct field access for mutations (idiomatic Go, type-safe).

**Serialization:** Collections are type wrappers over basic types (map/slice) and serialize/deserialize transparently with YAML. No special handling needed.

### 4. Critical Architectural Constraint

**Data-only child types:** Everything below Project in hierarchy MUST be pure data with zero business logic:
- Phase - Pure data (status, metadata, artifacts, tasks)
- Artifact - Pure data (type, path, approved, metadata)
- Task - Pure data (status, inputs, outputs, metadata)

**All orchestration at Project level:**
- `project.Advance()` - Fires project state machine events
- `project.AdvanceTask(taskID)` - (future) Fires task state machine events
- `project.AdvanceArtifact(artifactID)` - (future) Fires artifact state machine events

**Why this matters:**
1. No circular dependencies - child types never reference parent
2. Clean serialization - data-only types serialize to/from YAML trivially
3. Centralized logic - all state machines managed by Project
4. Simple testing - child types are just structs with fields
5. No forward references - architecture remains clean and maintainable

---

## Key Findings

### Finding 1: Unified Project SDK with Builder Pattern

**Complete project definition through single builder API:**

```go
func NewStandardProjectConfig() *sdk.ProjectTypeConfig {
    return sdk.NewProjectTypeConfigBuilder("standard").

        // Phase configuration with options pattern
        WithPhase("planning",
            WithStartState(PlanningActive),
            WithEndState(PlanningActive),
            WithInputs("context"),
            WithOutputs("task_list"),
        ).

        // State machine configuration
        SetInitialState(NoProject).
        AddTransition(
            PlanningActive,
            ImplementationPlanning,
            EventCompletePlanning,
            sdk.WithGuard(func(p *Project) bool {
                return p.PhaseOutputApproved("planning", "task_list")
            }),
        ).

        // Prompt configuration
        WithPrompt(PlanningActive, func(p *Project) string {
            return "Planning phase active. Create and approve task list."
        }).

        Build()
}
```

**Benefits:**
- Single builder interface for complete project configuration
- Options pattern (not nested builders) for simplicity
- Phase status updates automatic based on StartState/EndState
- Guards bind to project instances via closures at runtime
- No manual wiring of phase transitions

### Finding 2: Universal Project/Phase/Artifact/Task Types

**Single set of types works for all project types:**

```go
type Project struct {
    Name        string
    Type        string
    Phases      PhaseCollection
    Statechart  StatechartState

    config  *ProjectTypeConfig  // Runtime: defines rules
    machine *statechart.Machine // Runtime: built once on Load
}

type Phase struct {
    Status      string
    Metadata    map[string]interface{}
    Inputs      ArtifactCollection
    Outputs     ArtifactCollection
    Tasks       TaskCollection
}

type Artifact struct {
    Type      string
    Path      string
    Approved  bool
    Metadata  map[string]interface{}
}
```

**Key points:**
- Universal structure across project types
- Project.Phases is map[string]*Phase (standardized, not project-type-specific structs)
- Metadata maps provide flexibility for custom fields
- Collections serialize/deserialize transparently (just type wrappers)
- Validation via go-playground/validator for metadata constraints

### Finding 3: Metadata Separation

**Explicit metadata command to avoid "magic" routing:**

```bash
# Common fields - direct commands
sow output set --index 0 approved true

# Metadata fields - explicit subcommand
sow output metadata set --index 0 assessment pass
```

**Alternative rejected:** Automatic routing to metadata for unknown fields. Too surprising - typos in field names silently succeed, CUE errors confusing.

**Implementation:**
```go
// On Artifact (collection provides Get)
artifact, err := phase.Outputs.Get(index)

// Direct field mutation
artifact.Approved = true

// Metadata mutation (explicit command)
if artifact.Metadata == nil {
    artifact.Metadata = make(map[string]interface{})
}
artifact.Metadata[field] = value
```

### Finding 4: Template Functions + Closure Binding

**How state machines bind to project instances:**

**Step 1 - Registration (app startup):**
```go
// Guard defined as template function
sdk.WithGuard(func(p *Project) bool {
    return p.PhaseOutputApproved("planning", "task_list")
})

// Stored in ProjectTypeConfig.transitions[]
tc.guardFunc = func(p *Project) bool { ... }
```

**Step 2 - Machine building (per command):**
```go
func (ptc *ProjectTypeConfig) BuildMachine(project *Project, initialState State) *Machine {
    for _, tc := range ptc.transitions {
        // Bind template to THIS project instance
        boundGuard := func() bool {
            return tc.guardFunc(project)  // project captured in closure
        }

        builder.AddTransition(..., statechart.WithGuard(boundGuard))
    }
}
```

**Result:** Guards always operate on the specific project instance, seeing live state during transitions.

### Finding 5: Load/Save Pattern

**Load - deserialize and initialize:**
```go
func Load(ctx) (*Project, error) {
    // 1. Deserialize YAML into universal Project struct
    var project Project
    yaml.Unmarshal(data, &project)

    // 2. Attach type config from registry
    project.config = Registry[project.Type]

    // 3. Build state machine ONCE, initialized with saved state
    project.machine = project.config.BuildMachine(&project, project.Statechart.CurrentState)

    // 4. Validate instance against type rules
    project.config.Validate(&project)

    return &project, nil
}
```

**Save - validate and serialize:**
```go
func (p *Project) Save() error {
    // 1. Sync machine state to project
    p.Statechart.CurrentState = p.machine.State().String()

    // 2. Validate before saving
    p.config.Validate(p)

    // 3. Marshal to YAML (collections serialize as underlying map/slice)
    data := yaml.Marshal(p)

    // 4. Atomic write (temp file + rename)
    os.WriteFile(tmpPath, data, 0644)
    os.Rename(tmpPath, path)
}
```

### Finding 6: CLI Integration Pattern

**Most commands: Direct mutation + save**
```go
func runOutputSet(cmd *cobra.Command, args []string) error {
    project := loader.Load(ctx)

    // Use collections for navigation
    phase, err := project.Phases.Get(phaseName)
    output, err := phase.Outputs.Get(index)

    // Direct field mutation
    output.Approved = true

    // Save at end
    return project.Save()
}
```

**Advance command: State machine orchestration**
```go
func (p *Project) Advance() error {
    event := p.determineNextEvent()  // Based on current state

    can, _ := p.machine.CanFire(event)  // Check guards
    if !can {
        return ErrCannotAdvance
    }

    p.machine.Fire(event)  // Execute transition (guards, entry/exit actions, prompts)

    return p.Save()  // Save updated state
}
```

---

## Architectural Decisions

### AD1: Remove CUE Dependency

**Decision:** Use Go + go-playground/validator instead of CUE schemas.

**Context:** Standardized universal types (Project/Phase/Artifact/Task) eliminate need for polymorphic schema validation. CUE added complexity and required dual source of truth (Go structs + CUE schemas).

**Consequences:**
- ✅ Single language (Go only)
- ✅ Better IDE support
- ✅ Simpler mental model
- ❌ Lose external validation (`cue vet`)
- ❌ Runtime validation only for metadata

### AD2: Metadata Map for Flexibility

**Decision:** Common fields typed, custom fields in `metadata map[string]interface{}`.

**Context:** Cannot have compile-time type safety, code reuse, AND flexibility. Must pick two.

**Consequences:**
- ✅ No field duplication across project types
- ✅ Common fields (90% of access) type-safe
- ✅ Flexible for custom project types
- ❌ Custom field access requires type assertions
- ❌ Typos in metadata keys not caught by compiler
- Mitigation: go-playground/validator validates structure, tests catch bugs

### AD3: Data-Only Child Types

**Decision:** Phase/Artifact/Task are pure data. All orchestration at Project level.

**Context:** Prevents circular dependencies, enables clean serialization, centralizes logic.

**Consequences:**
- ✅ No forward references to Project
- ✅ Simple serialization (pure data)
- ✅ Testable (just structs)
- ✅ Enables future task/artifact state machines (all managed by Project)
- ⚠️ Must maintain discipline (no business logic in child types)

### AD4: Collection Pattern for API

**Decision:** Collections for structural operations, direct field access for mutations.

**Context:** Provide clean API without losing idiomatic Go patterns.

**Consequences:**
- ✅ Bounds checking centralized in collections
- ✅ Direct field access (type-safe, idiomatic)
- ✅ Collections serialize transparently
- ✅ Reusable operations (Get, Add, Remove)

### AD5: Explicit Metadata Subcommand

**Decision:** Separate `metadata set` command instead of automatic routing.

**Context:** Automatic routing creates surprising behavior (typos succeed, confusing errors).

**Consequences:**
- ✅ Clear separation: common fields vs metadata
- ✅ Predictable behavior (no magic)
- ✅ Better error messages
- ❌ Slightly more verbose for metadata operations

### AD6: Options Pattern for Builder

**Decision:** Use options pattern (`WithPhase(name, opts...)`) instead of nested builders.

**Context:** Nested builders require `Done()` methods and parent references (complex).

**Consequences:**
- ✅ Trivial implementation
- ✅ Standard Go idiom
- ✅ No `Done()` needed
- ✅ Easy to extend with new options

---

## Future: Task/Artifact State Machines

**This architecture enables future task/artifact state machines without breaking changes:**

```go
// Project orchestrates task state machine (no state machine on Task itself)
func (p *Project) AdvanceTask(taskID string) error {
    // Build task machine from project config
    taskMachine := p.config.BuildTaskMachine(taskID, p)

    event := determineNextTaskEvent(taskMachine, p)
    taskMachine.Fire(event)

    return p.Save()
}

// Guards can close over project state (no task reference to project needed)
func (config *ProjectTypeConfig) BuildTaskMachine(taskID string, p *Project) *Machine {
    builder.AddTransition(...,
        sdk.WithGuard(func() bool {
            // Closure captures p - task can check parent phase state
            return p.Phases["implementation"].Status == "in_progress"
        }),
    )
}
```

**Key points:**
- Tasks remain pure data (no machine field, no forward reference)
- Project builds task machines on-demand
- Guards close over project state for cross-entity checks
- Same pattern for artifact state machines
- All state machines managed centrally at Project level

This will be explored further in a dedicated task/artifact state machine exploration.

---

## Open Questions

- [ ] Should metadata validation use go-playground/validator tags or custom validation logic?
- [ ] How do we handle task metadata schemas (similar to phase metadata)?
- [ ] Should artifact metadata validation be per-type or per-phase-per-type?
- [ ] Do we need a way to define conditional phase features (e.g., tasks only if metadata flag set)?
- [ ] How do we version ProjectTypeConfig schemas for backward compatibility?

---

## References

- **Complementary design:** [Unified Command Hierarchy Design](../designs/command-hierarchy-design.md) - Defines the CLI interface this SDK powers
- **State machine current implementation:** cli/internal/project/statechart/
- **Existing collections:** cli/internal/project/artifacts.go, tasks_collection.go
- **Exploration artifacts:**
  - 01-current-state-machine-architecture.md - Analysis of replication effort
  - 02-coupling-complexity-assessment.md - Parent-child state machine patterns
  - 03-shell-architecture-patterns.md - Optional state machines via interfaces
  - 04-state-management-architecture.md - Metadata map decision
  - 05-project-sdk-api.md - Complete API specification (builder, Load/Save, CLI integration)

---

## Summary

This exploration defined a comprehensive Project SDK architecture that:

1. **Eliminates CUE dependency** - Go + go-playground/validator provides validation with better DX
2. **Unified builder API** - Single fluent interface defines phases, state machines, and prompts
3. **Universal types** - Project/Phase/Artifact/Task work for all project types
4. **Clean separation** - Data-only child types, orchestration at Project level
5. **Collection pattern** - Structural operations via collections, mutations via direct field access
6. **Explicit metadata** - Dedicated commands avoid surprising behavior
7. **Template binding** - Guards defined as templates, bound to instances via closures
8. **Future-ready** - Architecture supports task/artifact state machines without changes

**Next steps:**
1. Implement Project SDK based on 05-project-sdk-api.md specification
2. Explore task/artifact state machines (dedicated exploration session)
3. Migrate standard project to use new SDK
4. Update CLI commands to use collection pattern + explicit metadata commands

**Accepted trade-offs:**
- Runtime type safety for metadata (~10% of field access) vs compile-time safety everywhere
- Metadata accessed via type assertions vs strongly-typed accessors
- Validation at Save() time vs continuous validation
- Single save point at command end vs transaction-based intermediate saves

The architecture prioritizes simplicity, maintainability, and extensibility while accepting minimal runtime safety trade-offs in infrequently-accessed code paths.
