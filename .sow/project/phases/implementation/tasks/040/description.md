# Task 040: Project Configuration and SDK Integration

## Context

This task implements the core SDK configuration for the design project type, bringing together all previous components (states, events, guards, prompts) into a declarative configuration using the Project SDK builder pattern.

The design project type follows a 2-phase, 3-state workflow:
- **Phase 1: design** - Single Active state for planning, drafting, and reviewing documents
- **Phase 2: finalization** - Single Finalizing state for moving docs and creating PR
- **Terminal state: Completed** - Project finished

The SDK builder pattern provides a fluent API for configuring phases, transitions, guards, actions, event determiners, and prompts. The configuration is registered globally on package initialization, making it available throughout the application.

This task is critical as it integrates all previous tasks into a working project type that can be instantiated and used by the orchestrator.

## Requirements

### Main Configuration File

Implement `design.go` with the following components:

1. **Package initialization**
   ```go
   func init() {
       state.Register("design", NewDesignProjectConfig())
   }
   ```
   - Registers design project type on package load
   - Makes it available via `state.Registry["design"]`

2. **NewDesignProjectConfig() *project.ProjectTypeConfig**
   - Entry point for creating configuration
   - Uses builder pattern to assemble all components
   - Calls helper functions for each aspect of configuration
   - Returns fully built ProjectTypeConfig

3. **initializeDesignProject(p *state.Project, initialInputs map[string][]projschema.ArtifactState) error**
   - Called during project creation to set up phases
   - Creates both "design" and "finalization" phases
   - Design phase starts with status="active", enabled=true
   - Finalization phase starts with status="pending", enabled=false
   - Handles optional initial inputs for each phase
   - Initializes all phase collections (Inputs, Outputs, Tasks, Metadata)

4. **configurePhases(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder**
   - Defines phase structure using WithPhase
   - Design phase: start state = Active, end state = Active, supports tasks
   - Finalization phase: start state = Finalizing, end state = Finalizing, no tasks
   - Sets output types: design phase allows "design", "adr", "architecture", "diagram", "spec"
   - Finalization phase allows "pr"
   - Attaches metadata schemas to each phase

5. **configureTransitions(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder**
   - Sets initial state to Active
   - Defines Active → Finalizing transition with:
     - Event: EventCompleteDesign
     - Guard: allDocumentsApproved
     - OnExit: mark design phase as completed
     - OnEntry: enable and activate finalization phase
   - Defines Finalizing → Completed transition with:
     - Event: EventCompleteFinalization
     - Guard: allFinalizationTasksComplete
     - OnEntry: mark finalization phase as completed

6. **configureEventDeterminers(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder**
   - Maps states to events for advancement
   - Active → EventCompleteDesign
   - Finalizing → EventCompleteFinalization

### Phase Configuration Details

**Design Phase**:
- Name: "design"
- Start state: Active
- End state: Active (no intra-phase transitions)
- Tasks: Enabled (each task = one document)
- Outputs: "design", "adr", "architecture", "diagram", "spec"
- Metadata schema: designMetadataSchema (from Task 010)
- Initial status: "active"
- Initial enabled: true

**Finalization Phase**:
- Name: "finalization"
- Start state: Finalizing
- End state: Finalizing (no intra-phase transitions)
- Tasks: Disabled (finalization is direct orchestrator work)
- Outputs: "pr"
- Metadata schema: finalizationMetadataSchema (from Task 010)
- Initial status: "pending"
- Initial enabled: false

### Transition Actions

**Active → Finalizing OnExit**:
```go
phase := p.Phases["design"]
phase.Status = "completed"
phase.Completed_at = time.Now()
p.Phases["design"] = phase
```

**Active → Finalizing OnEntry**:
```go
phase := p.Phases["finalization"]
phase.Enabled = true
phase.Status = "in_progress"
phase.Started_at = time.Now()
p.Phases["finalization"] = phase
```

**Finalizing → Completed OnEntry**:
```go
phase := p.Phases["finalization"]
phase.Status = "completed"
phase.Completed_at = time.Now()
p.Phases["finalization"] = phase
```

## Acceptance Criteria

### Functional Requirements

- [ ] `design.go` implements all required functions
- [ ] Package init() registers "design" project type
- [ ] NewDesignProjectConfig builds complete configuration
- [ ] initializeDesignProject creates both phases correctly
- [ ] configurePhases defines phase structure properly
- [ ] configureTransitions sets up all transitions with guards and actions
- [ ] configureEventDeterminers maps all advanceable states
- [ ] Phase initial states match specification (design active, finalization pending)
- [ ] Transition actions update phase status and timestamps correctly

### Test Requirements (TDD)

Write comprehensive tests in `design_test.go`:

**Package registration tests**:
- [ ] "design" type is registered in global registry
- [ ] Registry entry is non-nil

**NewDesignProjectConfig tests**:
- [ ] Returns non-nil config
- [ ] Config has correct initial state (Active)
- [ ] Config has all transitions configured
- [ ] Config has prompts registered

**initializeDesignProject tests**:
- [ ] Creates both phases
- [ ] Design phase starts active with enabled=true
- [ ] Finalization phase starts pending with enabled=false
- [ ] Both phases have initialized collections (Inputs, Outputs, Tasks, Metadata)
- [ ] Handles nil initialInputs gracefully
- [ ] Handles provided initialInputs correctly
- [ ] Sets correct timestamps (Created_at)

**configurePhases tests**:
- [ ] Design phase supports tasks
- [ ] Finalization phase does not support tasks
- [ ] GetTaskSupportingPhases returns only "design"
- [ ] GetDefaultTaskPhase returns "design" for Active state
- [ ] Design phase allows document output types
- [ ] Finalization phase allows "pr" output
- [ ] Design phase rejects invalid output types (e.g., "pr")
- [ ] Finalization phase rejects invalid output types (e.g., "design")

**configureTransitions tests**:
- [ ] Initial state is Active
- [ ] All transitions configured correctly
- [ ] Guards are bound to transitions
- [ ] OnEntry/OnExit actions are bound

**Integration with guards**:
- [ ] Active → Finalizing blocked when documents not approved
- [ ] Active → Finalizing allowed when documents approved
- [ ] Finalizing → Completed blocked when tasks incomplete
- [ ] Finalizing → Completed allowed when tasks complete

### Code Quality

- [ ] All functions documented with clear descriptions
- [ ] Builder pattern used consistently
- [ ] Functions return builder for chaining
- [ ] Error handling in initializer function
- [ ] Timestamps use project creation time
- [ ] Tests use testify/assert and testify/require

## Technical Details

### Import Structure

```go
package design

import (
	"time"

	"github.com/jmgilman/sow/cli/internal/sdks/project"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
)
```

### Builder Pattern Usage

The SDK builder provides a fluent API:
```go
builder := project.NewProjectTypeConfigBuilder("design")
builder = configurePhases(builder)
builder = configureTransitions(builder)
builder = configureEventDeterminers(builder)
builder = configurePrompts(builder)
builder = builder.WithInitializer(initializeDesignProject)
return builder.Build()
```

Each configure function takes a builder and returns it for chaining.

### Phase Options

Use these option functions from the SDK:
- `project.WithStartState(state)` - Sets phase start state
- `project.WithEndState(state)` - Sets phase end state
- `project.WithOutputs(types...)` - Sets allowed output artifact types
- `project.WithTasks()` - Enables task support
- `project.WithMetadataSchema(schema)` - Sets CUE validation schema

### Transition Options

Use these option functions:
- `project.WithGuard(description, guardFunc)` - Adds guard with description
- `project.WithOnExit(action)` - Adds action executed on exit from source state
- `project.WithOnEntry(action)` - Adds action executed on entry to target state

### State Machine Integration

The configuration is used to build a state machine:
```go
config := NewDesignProjectConfig()
machine := config.BuildMachine(project)
```

The state machine handles:
- State transitions via Fire(event)
- Guard evaluation via CanFire(event)
- Action execution on transitions
- Event determination via Advance()

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/.sow/knowledge/designs/project-modes/design-design.md` - Complete configuration specification (lines 37-273)
- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/cli/internal/projects/exploration/exploration.go` - Reference implementation pattern
- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/cli/internal/projects/exploration/exploration_test.go` - Reference test patterns
- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/cli/internal/projects/standard/standard.go` - Complex transition example
- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/cli/internal/sdks/project/builder.go` - SDK builder API
- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/cli/internal/sdks/project/options.go` - Available option functions

## Examples

### Complete Configuration Function

```go
func NewDesignProjectConfig() *project.ProjectTypeConfig {
	builder := project.NewProjectTypeConfigBuilder("design")
	builder = configurePhases(builder)
	builder = configureTransitions(builder)
	builder = configureEventDeterminers(builder)
	builder = configurePrompts(builder)
	builder = builder.WithInitializer(initializeDesignProject)
	return builder.Build()
}
```

### Phase Configuration with Options

```go
func configurePhases(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
	return builder.
		WithPhase("design",
			project.WithStartState(sdkstate.State(Active)),
			project.WithEndState(sdkstate.State(Active)),
			project.WithOutputs("design", "adr", "architecture", "diagram", "spec"),
			project.WithTasks(),
			project.WithMetadataSchema(designMetadataSchema),
		).
		WithPhase("finalization",
			project.WithStartState(sdkstate.State(Finalizing)),
			project.WithEndState(sdkstate.State(Finalizing)),
			project.WithOutputs("pr"),
			project.WithMetadataSchema(finalizationMetadataSchema),
		)
}
```

### Transition with Guard and Actions

```go
AddTransition(
	sdkstate.State(Active),
	sdkstate.State(Finalizing),
	sdkstate.Event(EventCompleteDesign),
	project.WithGuard("all documents approved", func(p *state.Project) bool {
		return allDocumentsApproved(p)
	}),
	project.WithOnExit(func(p *state.Project) error {
		phase := p.Phases["design"]
		phase.Status = "completed"
		phase.Completed_at = time.Now()
		p.Phases["design"] = phase
		return nil
	}),
	project.WithOnEntry(func(p *state.Project) error {
		phase := p.Phases["finalization"]
		phase.Enabled = true
		phase.Status = "in_progress"
		phase.Started_at = time.Now()
		p.Phases["finalization"] = phase
		return nil
	}),
)
```

## Dependencies

- Task 010: Core Structure and Constants - States, events, metadata schemas
- Task 020: Guard Functions and Helpers - Guard functions used in transitions
- Task 030: Prompt Templates and Generators - Prompt configuration function

## Constraints

### State Machine Semantics

The design project uses a simple 3-state machine:
- Only inter-phase transitions (no intra-phase state changes)
- Design phase has one state (Active) for entire design process
- Finalization phase has one state (Finalizing) for all finalization work
- Guards enforce workflow rules (can't skip states)

### Phase Lifecycle

Phases follow strict lifecycle:
1. Created with status="pending", enabled=false
2. Enabled and activated on entry (status changes to "active" or "in_progress")
3. Marked completed on exit (status="completed", completed_at set)

### Task Phase Mapping

Only design phase supports tasks:
- Document planning tasks live in design phase
- Finalization work is done directly by orchestrator (no tasks)
- GetDefaultTaskPhase should always return "design" since it's the only task-supporting phase

### Artifact Type Validation

The SDK validates artifact types against phase configuration:
- Must explicitly list allowed types in WithOutputs
- Unknown types are rejected at validation time
- Ensures consistent artifact taxonomy across project
