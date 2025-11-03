# Task Breakdown: Project SDK Builder & Configuration

## Overview

This task breakdown implements the fluent builder API, options pattern, registry, OnAdvance configuration, BuildMachine with guard closure binding, and two-tier validation for the Project SDK. All work builds on top of the completed unit-002 (state types & persistence layer).

**Existing Foundation:**
- CUE schemas and generated types (`cli/schemas/project/`)
- State machine SDK (`cli/internal/sdks/state/`)
- Wrapper types and collections (`cli/internal/sdks/project/state/`)
- Load/Save with CUE validation
- Helper methods for guards (PhaseOutputApproved, PhaseMetadataBool, AllTasksComplete)

**What We're Building:**
- ProjectTypeConfig structure with all necessary fields
- ProjectTypeConfigBuilder with fluent API
- Options pattern (PhaseOpt, TransitionOption)
- BuildMachine() that binds guards to project instances via closures
- OnAdvance configuration for state-specific event determination
- Two-tier validation (structure via CUE, metadata via embedded schemas)
- Complete registry implementation

**Testing Approach:** TDD with behavior-only tests. No testing of internal implementation details.

---

## Tasks

### Task 1: Core Configuration Types and Options Pattern

**Agent:** implementer
**Estimated Time:** 1.5 hours

**Description:**

Implement the foundational types and options pattern for project type configuration. This includes:
- `PhaseConfig` structure with start/end states, allowed artifact types, task support, and metadata schema
- `TransitionConfig` structure with from/to/event, guard template, and action functions
- `ProjectTypeConfig` structure with phases, transitions, onAdvance handlers, and prompts
- `PhaseOpt` functions for phase configuration options
- `TransitionOption` functions for transition configuration options
- Function types: `GuardTemplate`, `Action`, `EventDeterminer`, `PromptGenerator`

**Files to Create/Modify:**
- `cli/internal/sdks/project/config.go` - Configuration structures
- `cli/internal/sdks/project/options.go` - Option functions
- `cli/internal/sdks/project/types.go` - Function type definitions
- `cli/internal/sdks/project/options_test.go` - Behavioral tests

**Acceptance Criteria:**
- [ ] PhaseConfig holds start/end states, allowed input/output types, task support flag, metadata schema string
- [ ] TransitionConfig holds from/to/event, guard template (func(*Project) bool), onEntry/onExit actions
- [ ] ProjectTypeConfig holds name, phase configs map, initial state, transitions slice, onAdvance map, prompts map
- [ ] PhaseOpt functions (WithStartState, WithEndState, WithInputs, WithOutputs, WithTasks, WithMetadataSchema) correctly modify PhaseConfig
- [ ] TransitionOption functions (WithGuard, WithOnEntry, WithOnExit) correctly modify TransitionConfig
- [ ] Function types defined: GuardTemplate, Action, EventDeterminer, PromptGenerator
- [ ] All tests pass, covering option application behavior

**Dependencies:** None

---

### Task 2: Builder API Implementation

**Agent:** implementer
**Estimated Time:** 2 hours

**Description:**

Implement the fluent builder API for defining project types. The builder provides a declarative interface for configuring all aspects of a project type (phases, state machine, event determination, prompts).

**Files to Create/Modify:**
- `cli/internal/sdks/project/builder.go` - ProjectTypeConfigBuilder implementation
- `cli/internal/sdks/project/builder_test.go` - Behavioral tests

**Acceptance Criteria:**
- [ ] NewProjectTypeConfigBuilder(name) creates builder with empty collections initialized
- [ ] WithPhase(name, opts...) adds phase config, applies all options, returns builder (chainable)
- [ ] SetInitialState(state) sets initial state, returns builder
- [ ] AddTransition(from, to, event, opts...) adds transition config, applies all options, returns builder
- [ ] OnAdvance(state, determiner) configures event determiner for state, returns builder
- [ ] WithPrompt(state, generator) configures prompt generator for state, returns builder
- [ ] Build() returns ProjectTypeConfig with all configured data
- [ ] Builder is reusable (can call Build() multiple times)
- [ ] Multiple phases and transitions can be added to single builder
- [ ] All tests pass, covering fluent API chaining behavior

**Dependencies:** Task 1 (config types and options)

---

### Task 3: Registry Implementation

**Agent:** implementer
**Estimated Time:** 1 hour

**Description:**

Implement the global registry for project type registration and lookup. The registry enables project types to be registered at startup and retrieved during Load().

**Files to Modify:**
- `cli/internal/sdks/project/state/registry.go` - Add Register() and Get() functions
- `cli/internal/sdks/project/state/registry_test.go` (create) - Behavioral tests

**Acceptance Criteria:**
- [ ] Register(typeName, config) adds config to global Registry map
- [ ] Register() panics if typeName already registered (prevents accidental duplicates)
- [ ] Get(typeName) returns (config, true) for registered types
- [ ] Get(typeName) returns (nil, false) for unregistered types
- [ ] Registry correctly stores and retrieves multiple project types
- [ ] All tests pass, covering registration and retrieval behavior

**Dependencies:** Task 2 (builder creates ProjectTypeConfig)

---

### Task 4: BuildMachine with Closure Binding

**Agent:** implementer
**Estimated Time:** 2 hours

**Description:**

Implement BuildMachine() that creates a state machine with guards bound to project instances via closures. This enables guards to access live project state while being defined declaratively in project type configs.

**Files to Modify:**
- `cli/internal/sdks/project/state/registry.go` - Implement BuildMachine() method
- `cli/internal/sdks/project/state/machine.go` (create) - Machine-related helpers if needed
- `cli/internal/sdks/project/state/machine_test.go` (create) - Behavioral tests

**Acceptance Criteria:**
- [ ] BuildMachine(project, initialState) creates state machine initialized with given state
- [ ] All transitions from ProjectTypeConfig are added to machine
- [ ] Guards are bound to project instance via closures (guard templates become GuardFunc)
- [ ] Bound guards can access project state and return bool
- [ ] OnEntry actions are bound and can mutate project state
- [ ] OnExit actions are bound and can mutate project state
- [ ] Machine correctly enforces guards (CanFire returns false when guard fails)
- [ ] All tests pass, covering guard binding and state machine behavior

**Dependencies:** Task 3 (ProjectTypeConfig available)

**Technical Notes:**
- Use `internal/sdks/state.MachineBuilder` to construct the machine
- Bind guard templates by creating closures: `func() bool { return guardTemplate(project) }`
- Import from `internal/sdks/state` for Event, State, and builder types

---

### Task 5: OnAdvance Configuration and Project.Advance()

**Agent:** implementer
**Estimated Time:** 2 hours

**Description:**

Implement OnAdvance configuration and the generic Project.Advance() method. This enables the generic `sow advance` command to work across all project types by delegating event determination to project-type-specific logic.

**Files to Modify:**
- `cli/internal/sdks/project/state/registry.go` - Add GetEventDeterminer() method to ProjectTypeConfig
- `cli/internal/sdks/project/state/project.go` - Implement Advance() method
- `cli/internal/sdks/project/state/advance_test.go` (create) - Behavioral tests

**Acceptance Criteria:**
- [ ] GetEventDeterminer(state) returns configured EventDeterminer for given state
- [ ] GetEventDeterminer(state) returns nil for states without configured determiner
- [ ] Project.Advance() gets current state from machine
- [ ] Advance() calls GetEventDeterminer() to get determiner for current state
- [ ] Advance() calls determiner function with project to get next event
- [ ] Advance() checks if transition is allowed via machine.CanFire()
- [ ] Advance() fires event if allowed (executes transition with guards/actions)
- [ ] Advance() returns error if no determiner configured
- [ ] Advance() returns error if guard prevents transition
- [ ] Advance() returns error if determiner returns error
- [ ] All tests pass, covering event determination and transition behavior

**Dependencies:** Task 4 (BuildMachine creates machine with bound guards)

---

### Task 6: Two-Tier Validation Implementation

**Agent:** implementer
**Estimated Time:** 2.5 hours

**Description:**

Implement the two-tier validation system: CUE validates structure (already done in unit-002), and runtime validates metadata against embedded schemas. This enables project types to define custom metadata schemas that are validated on Save().

**Files to Modify:**
- `cli/internal/sdks/project/state/registry.go` - Implement Validate() method on ProjectTypeConfig
- `cli/internal/sdks/project/state/validate.go` - Add validateMetadata() helper
- `cli/internal/sdks/project/state/validate_test.go` - Add metadata validation tests

**Acceptance Criteria:**
- [ ] Validate(project) iterates over all phase configs
- [ ] For each phase in state, validate input artifacts against allowed input types (if specified)
- [ ] For each phase in state, validate output artifacts against allowed output types (if specified)
- [ ] Empty allowed types list means "allow all types" (no validation)
- [ ] If phase has metadata schema, validate phase.Metadata against embedded CUE schema
- [ ] If phase has no metadata schema but metadata present, return error
- [ ] validateMetadata() compiles embedded CUE schema, encodes metadata, unifies, and validates
- [ ] Validation errors include phase name and clear error message
- [ ] Schema compilation errors reported clearly
- [ ] All tests pass, covering artifact type validation and metadata validation behavior

**Dependencies:** Task 2 (ProjectTypeConfig with phase configs)

**Technical Notes:**
- Use `cuelang.org/go/cue` for CUE runtime validation
- Use `cue.Concrete(true)` for strict validation
- Reference existing `validateStructure()` in `validate.go` for CUE validation patterns

---

### Task 7: Integration Test - Complete Project Type Configuration

**Agent:** implementer
**Estimated Time:** 1.5 hours

**Description:**

Create a comprehensive integration test that demonstrates a complete project type can be defined using the SDK, registered, loaded, advanced through states, and validated. This satisfies the acceptance criteria "Example project type can be fully configured using SDK."

**Files to Create:**
- `cli/internal/sdks/project/integration_test.go` - Full workflow integration test

**Acceptance Criteria:**
- [ ] Test defines a simple project type with 2-3 phases using builder API
- [ ] Test includes phase with metadata schema (embedded CUE string)
- [ ] Test includes transitions with guards that access project state
- [ ] Test includes OnAdvance configuration for each state
- [ ] Test registers project type in registry
- [ ] Test creates project state, calls BuildMachine(), verifies machine created
- [ ] Test calls Advance() multiple times, progressing through lifecycle
- [ ] Test verifies guards prevent invalid transitions
- [ ] Test verifies guards allow valid transitions
- [ ] Test verifies metadata validation works (both pass and fail cases)
- [ ] Test verifies artifact type validation works
- [ ] Test demonstrates full lifecycle: configure → register → build machine → advance → validate
- [ ] All assertions pass, proving complete SDK functionality

**Dependencies:** All previous tasks (complete SDK implementation)

**Technical Notes:**
- Use in-memory test project state (no disk I/O required)
- Create simple but realistic example (e.g., 3-state workflow: Start → Working → Done)
- Include both positive tests (valid operations succeed) and negative tests (invalid operations fail)

---

## Summary

**Total Tasks:** 7
**Estimated Time:** 12.5 hours
**Testing Strategy:** TDD with behavior-only tests throughout

**Deliverables:**
- Fluent builder API for complete project type configuration
- Options pattern for flexible phase and transition configuration
- Global registry for project type registration
- BuildMachine() with guard closure binding
- OnAdvance configuration for generic advance command
- Two-tier validation (structure + metadata)
- Complete integration test demonstrating full SDK capabilities

**Success Metrics:**
- All unit tests pass
- Integration test demonstrates complete workflow
- Builder API is self-documenting and easy to use
- No modifications to core SDK code needed when adding new project types
