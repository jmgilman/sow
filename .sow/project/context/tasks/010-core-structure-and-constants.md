# Task 010: Core Structure and Constants

## Context

This task establishes the foundational structure for the design project type implementation. The design project type is a new workflow for creating architecture and design documentation, following the 3-state pattern: Active → Finalizing → Completed.

The design project type is built using the Project SDK builder pattern, similar to the exploration and standard project types. It requires:
- A package structure at `cli/internal/projects/design/`
- State and event constants defining the workflow
- Package initialization with registry integration

This is part of implementing GitHub issue #37, which specifies a complete design workflow for tracking design artifacts through planning, drafting, review, approval, and finalization phases.

## Requirements

### Package Structure

Create the package at `cli/internal/projects/design/` with the following files:

1. **states.go** - State constants for the state machine:
   - `Active` - Active design phase (design.active status)
   - `Finalizing` - Finalization in progress (finalization.in_progress status)
   - `Completed` - Terminal state (design finished)

2. **events.go** - Event constants that trigger transitions:
   - `EventCompleteDesign` - Transitions from Active to Finalizing
   - `EventCompleteFinalization` - Transitions from Finalizing to Completed

3. **metadata.go** - Embedded CUE metadata schemas:
   - Embed `cue/design_metadata.cue` for design phase validation
   - Embed `cue/finalization_metadata.cue` for finalization phase validation

4. **cue/design_metadata.cue** - Design phase metadata schema:
   - Empty schema (no required metadata for design phase)
   - Allow optional metadata as needed

5. **cue/finalization_metadata.cue** - Finalization phase metadata schema:
   - `pr_url` (optional string) - URL of created pull request
   - `project_deleted` (optional bool) - Flag indicating .sow/project/ has been deleted

### File Contents

**states.go**:
```go
package design

import "github.com/jmgilman/sow/cli/internal/sdks/state"

// Design project states

const (
	// Active indicates active design phase
	Active = state.State("Active")

	// Finalizing indicates finalization in progress
	Finalizing = state.State("Finalizing")

	// Completed indicates design finished
	Completed = state.State("Completed")
)
```

**events.go**:
```go
package design

import "github.com/jmgilman/sow/cli/internal/sdks/state"

// Design project events trigger state transitions

const (
	// EventCompleteDesign transitions from Active to Finalizing
	// Fired when all documents are approved
	EventCompleteDesign = state.Event("complete_design")

	// EventCompleteFinalization transitions from Finalizing to Completed
	// Fired when all finalization tasks are completed
	EventCompleteFinalization = state.Event("complete_finalization")
)
```

**metadata.go**:
```go
package design

import _ "embed"

// Embedded CUE metadata schemas for phase validation

//go:embed cue/design_metadata.cue
var designMetadataSchema string

//go:embed cue/finalization_metadata.cue
var finalizationMetadataSchema string
```

**cue/design_metadata.cue**:
```cue
package design

// Metadata schema for design phase
{
	// No required metadata for design phase
	// Optional metadata can be added as needed
}
```

**cue/finalization_metadata.cue**:
```cue
package design

// Metadata schema for finalization phase
{
	// pr_url: Optional URL of created pull request
	pr_url?: string

	// project_deleted: Flag indicating .sow/project/ has been deleted
	project_deleted?: bool
}
```

## Acceptance Criteria

### Functional Requirements

- [ ] Package `cli/internal/projects/design` is created
- [ ] `states.go` defines all three state constants (Active, Finalizing, Completed)
- [ ] `events.go` defines both event constants (EventCompleteDesign, EventCompleteFinalization)
- [ ] `metadata.go` embeds both CUE schema files
- [ ] `cue/design_metadata.cue` exists with empty schema structure
- [ ] `cue/finalization_metadata.cue` exists with pr_url and project_deleted fields
- [ ] All constants use correct type aliases from `sdks/state` package
- [ ] State and event names match the design specification exactly

### Test Requirements (TDD)

Write unit tests in parallel with implementation:

**states_test.go**:
- [ ] Test that state constants have correct string values
- [ ] Test that states are of correct type (state.State)

**events_test.go**:
- [ ] Test that event constants have correct string values
- [ ] Test that events are of correct type (state.Event)

**metadata_test.go**:
- [ ] Test that designMetadataSchema is not empty
- [ ] Test that finalizationMetadataSchema is not empty
- [ ] Test that embedded CUE schemas are valid CUE syntax

### Code Quality

- [ ] All files have package documentation comments
- [ ] Event constants include clear documentation explaining when they fire
- [ ] State constants include documentation explaining their purpose
- [ ] Follow Go naming conventions
- [ ] Code passes `go vet` and `golint`

## Technical Details

### Import Structure

Use the following imports:
- `github.com/jmgilman/sow/cli/internal/sdks/state` - For State and Event type aliases
- `_ "embed"` - For embedding CUE schema files

### Type Aliases

States and events use type aliases from the SDK:
- `state.State` - Type for state machine states
- `state.Event` - Type for state machine events

These are string-based types that provide type safety while remaining serializable.

### Embedded Files

The CUE schemas are embedded using Go's `//go:embed` directive:
- Files are embedded at compile time
- Accessible as string variables
- Used by the Project SDK for runtime validation

### CUE Schema Structure

CUE schemas define the structure and validation rules for phase metadata:
- Design phase: No required fields (open schema for flexibility)
- Finalization phase: Optional fields for tracking PR and cleanup status

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/.sow/project/context/issue-37.md` - GitHub issue with full requirements
- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/.sow/knowledge/designs/project-modes/design-design.md` - Complete design specification
- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/cli/internal/projects/exploration/states.go` - Example state constants pattern
- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/cli/internal/projects/exploration/events.go` - Example event constants pattern
- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/cli/internal/projects/exploration/metadata.go` - Example embedded schema pattern
- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/cli/internal/projects/exploration/cue/exploration_metadata.cue` - Example CUE schema
- `/Users/josh/code/sow/.sow/worktrees/37-design-project-type-implementation/cli/internal/projects/exploration/cue/finalization_metadata.cue` - Example finalization schema

## Examples

### State Usage in State Machine

States are used in transition configuration:
```go
AddTransition(
    sdkstate.State(Active),
    sdkstate.State(Finalizing),
    sdkstate.Event(EventCompleteDesign),
    // ... options
)
```

### Event Usage in Determiners

Events are returned by event determiners:
```go
OnAdvance(sdkstate.State(Active), func(_ *state.Project) (sdkstate.Event, error) {
    return sdkstate.Event(EventCompleteDesign), nil
})
```

### CUE Schema Validation

Schemas are passed to phase configuration:
```go
WithPhase("design",
    project.WithMetadataSchema(designMetadataSchema),
)
```

## Dependencies

None - this is the foundational task that other tasks will depend on.

## Constraints

### Naming Conventions

- State names use PascalCase: `Active`, `Finalizing`, `Completed`
- Event names use camelCase with Event prefix: `EventCompleteDesign`
- State constants represent high-level workflow states
- Event constants describe the completion conditions

### Design Decisions

- 3-state workflow (simpler than standard's 7-state workflow)
- No intra-phase transitions in design phase (single Active state)
- Terminal state is `Completed` (matches exploration pattern)
- Event names indicate completion rather than initiation

### Compatibility

- Must work with existing Project SDK
- State and event types must be compatible with state machine SDK
- CUE schemas must be valid CUE syntax for runtime validation
