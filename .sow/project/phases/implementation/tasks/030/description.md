# Task 030: Implement Phase Configuration

## Context

This task implements the phase configuration for the exploration project type. The exploration workflow has a 2-phase structure:
1. **Exploration phase** - Contains two states (Active, Summarizing) within a single phase
2. **Finalization phase** - Contains one state (Finalizing)

Each phase has specific configuration including allowed artifact types, task support, metadata schemas, and state boundaries. The configuration uses the Project SDK builder pattern with functional options.

## Requirements

### Update exploration.go

In `cli/internal/projects/exploration/exploration.go`, implement the `configurePhases` function:

1. **Function signature**:
   ```go
   func configurePhases(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder
   ```

2. **Exploration phase configuration**:
   - Phase name: "exploration"
   - Start state: `sdkstate.State(Active)`
   - End state: `sdkstate.State(Summarizing)`
   - Allowed output types: "summary", "findings"
   - Supports tasks: Yes (use `project.WithTasks()`)
   - Metadata schema: `explorationMetadataSchema` (will be defined in Task 060)

3. **Finalization phase configuration**:
   - Phase name: "finalization"
   - Start state: `sdkstate.State(Finalizing)`
   - End state: `sdkstate.State(Finalizing)` (single-state phase)
   - Allowed output types: "pr"
   - Does not support tasks
   - Metadata schema: `finalizationMetadataSchema` (will be defined in Task 060)

4. **Implementation pattern**:
   ```go
   return builder.
       WithPhase("exploration",
           project.WithStartState(sdkstate.State(Active)),
           project.WithEndState(sdkstate.State(Summarizing)),
           project.WithOutputs("summary", "findings"),
           project.WithTasks(),
           project.WithMetadataSchema(explorationMetadataSchema),
       ).
       WithPhase("finalization",
           project.WithStartState(sdkstate.State(Finalizing)),
           project.WithEndState(sdkstate.State(Finalizing)),
           project.WithOutputs("pr"),
           project.WithMetadataSchema(finalizationMetadataSchema),
       )
   ```

5. **Temporary placeholder for metadata schemas**:
   Add at top of file (will be replaced in Task 060):
   ```go
   var explorationMetadataSchema = ""
   var finalizationMetadataSchema = ""
   ```

## Test-Driven Development

This task follows TDD methodology:

1. **Write tests first** for:
   - `configurePhases()` returns non-nil builder
   - Exploration phase configuration:
     - Has correct start/end states
     - Allows "summary" and "findings" output types
     - Supports tasks
   - Finalization phase configuration:
     - Has correct start/end states
     - Allows "pr" output type
     - Does not support tasks

2. **Run tests** - they should fail initially (red phase)

3. **Implement functionality** - write phase configuration (green phase)

4. **Refactor** - improve readability and organization

Add tests to `cli/internal/projects/exploration/exploration_test.go`.

## Acceptance Criteria

- [ ] `configurePhases` function implemented in `exploration.go`
- [ ] **Unit tests written before implementation**
- [ ] Tests verify phase configuration options
- [ ] All tests pass
- [ ] Exploration phase configured with correct states and options
- [ ] Finalization phase configured with correct states and options
- [ ] Exploration phase allows "summary" and "findings" output types
- [ ] Finalization phase allows "pr" output type
- [ ] Exploration phase supports tasks
- [ ] Finalization phase does not support tasks
- [ ] Both phases reference metadata schema variables
- [ ] Placeholder metadata schema variables defined
- [ ] Function returns builder for chaining
- [ ] Code follows Go formatting standards (gofmt)
- [ ] No compilation errors

## Technical Details

### Phase Configuration Options

The SDK provides these configuration options via `project.WithXXX()` functions:

- `WithStartState(state)` - Sets phase start state
- `WithEndState(state)` - Sets phase end state
- `WithOutputs(types...)` - Constrains allowed output artifact types
- `WithInputs(types...)` - Constrains allowed input artifact types (not used in exploration)
- `WithTasks()` - Enables task support for the phase
- `WithMetadataSchema(schema)` - Sets CUE schema for metadata validation

### Artifact Type Constraints

Artifact types are validated when outputs are added to a phase. Only types listed in `WithOutputs()` are allowed. This prevents:
- Adding "review" artifacts to exploration phase (not a review workflow)
- Adding task-related artifacts to finalization phase
- Type mismatches that could break workflows

### Phase State Boundaries

Start and end states define the state range for a phase:
- Exploration: Active → Summarizing (2 states in same phase)
- Finalization: Finalizing → Finalizing (1 state in phase)

This is unusual - typically phases have start != end, but finalization is a simple single-state phase.

### Task Support

Only the exploration phase supports tasks (research topics). Finalization uses simple checklist-style operations without formal task tracking.

The SDK uses this information for:
- Validating task operations (can only add tasks to task-supporting phases)
- Determining default phase for task commands
- Schema validation

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/internal/projects/standard/standard.go` - Reference phase configuration (lines 61-82)
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/internal/sdks/project/options.go` - Phase option functions
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/internal/sdks/project/config.go` - PhaseConfig structure
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/.sow/knowledge/designs/project-modes/exploration-design.md` - Design specification (lines 88-117)
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/.sow/project/context/issue-36.md` - Requirements

## Examples

### Standard Project Phase Configuration (Reference)

From `cli/internal/projects/standard/standard.go:61-82`:

```go
func configurePhases(builder *project.ProjectTypeConfigBuilder) *project.ProjectTypeConfigBuilder {
    return builder.
        WithPhase("implementation",
            project.WithStartState(sdkstate.State(ImplementationPlanning)),
            project.WithEndState(sdkstate.State(ImplementationExecuting)),
            project.WithOutputs("task_list"),
            project.WithTasks(),
            project.WithMetadataSchema(implementationMetadataSchema),
        ).
        WithPhase("review",
            project.WithStartState(sdkstate.State(ReviewActive)),
            project.WithEndState(sdkstate.State(ReviewActive)),
            project.WithOutputs("review"),
            project.WithMetadataSchema(reviewMetadataSchema),
        ).
        WithPhase("finalize",
            project.WithStartState(sdkstate.State(FinalizeChecks)),
            project.WithEndState(sdkstate.State(FinalizeCleanup)),
            project.WithOutputs("pr_body"),
            project.WithMetadataSchema(finalizeMetadataSchema),
        )
}
```

### Phase Option Pattern

Options are functions that modify a `PhaseConfig`:

```go
func WithOutputs(types ...string) PhaseOpt {
    return func(pc *PhaseConfig) {
        pc.allowedOutputTypes = types
    }
}
```

Applied during builder construction via variadic arguments.

## Dependencies

- Task 010 (Package structure) - Provides `exploration.go` file
- Task 020 (States and events) - Provides state constants
- Will be used by Task 040 (Transitions) to map states to phases
- Metadata schema variables will be populated in Task 060

## Constraints

- Must use exact artifact type names: "summary", "findings", "pr"
- Phase names must be lowercase: "exploration", "finalization"
- Must follow builder pattern (return builder for chaining)
- Cannot change phase structure from design (2 phases, not 3)
- Empty metadata schemas are acceptable temporarily (will be populated later)
- Start and end states must match state machine design
