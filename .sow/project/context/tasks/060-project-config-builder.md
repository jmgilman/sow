# Task 060: Migrate Project Configuration and Builder

## Context

This task is part of the `libs/project` module consolidation effort. It migrates the project type configuration system from `cli/internal/sdks/project/` to `libs/project/`.

The ProjectTypeConfig and ProjectTypeConfigBuilder provide:
- Configuration for project types (standard, exploration, design, breakdown)
- Phase definitions with start/end states, inputs/outputs, tasks support
- Transition definitions with guards, actions, and phase status management
- Branch point configuration for decision states
- Machine building with bound guards and actions

This is the core SDK API that project types use to define their lifecycle.

## Requirements

### 1. Migrate ProjectTypeConfig (config.go)

Migrate the configuration type from `cli/internal/sdks/project/config.go`:

```go
// ProjectTypeConfig defines the complete configuration for a project type.
// It contains all phases, transitions, branches, prompts, and initialization logic.
type ProjectTypeConfig struct {
    name         string
    phases       map[string]*PhaseConfig
    transitions  []*TransitionConfig
    branches     map[State]*BranchConfig
    prompts      map[State]PromptGenerator
    initialState State
    initializer  func(*state.Project, map[string][]project.ArtifactState) error
    validator    func(*state.Project) error
}

// Name returns the project type name (e.g., "standard", "exploration").
func (ptc *ProjectTypeConfig) Name() string

// InitialState returns the initial state for new projects of this type.
func (ptc *ProjectTypeConfig) InitialState() State

// Phases returns the phase configurations.
func (ptc *ProjectTypeConfig) Phases() map[string]*PhaseConfig

// GetPhaseForState returns the phase name that owns the given state.
func (ptc *ProjectTypeConfig) GetPhaseForState(s State) string

// IsPhaseStartState returns true if the state is the phase's start state.
func (ptc *ProjectTypeConfig) IsPhaseStartState(phaseName string, s State) bool

// IsPhaseEndState returns true if the state is the phase's end state.
func (ptc *ProjectTypeConfig) IsPhaseEndState(phaseName string, s State) bool

// GetTransition returns the transition config for the given state change.
func (ptc *ProjectTypeConfig) GetTransition(from, to State, event Event) *TransitionConfig

// Initialize sets up a new project with phases and initial state.
func (ptc *ProjectTypeConfig) Initialize(p *state.Project, initialInputs map[string][]project.ArtifactState) error

// Validate validates project state against type-specific rules.
func (ptc *ProjectTypeConfig) Validate(p *state.Project) error

// BuildMachine creates a state machine bound to a project instance.
func (ptc *ProjectTypeConfig) BuildMachine(p *state.Project, initialState State) *Machine

// FireWithPhaseUpdates fires an event and updates phase statuses automatically.
func (ptc *ProjectTypeConfig) FireWithPhaseUpdates(machine *Machine, event Event, p *state.Project) error

// GetEventDeterminer returns the event determiner for a branch state.
func (ptc *ProjectTypeConfig) GetEventDeterminer(s State) EventDeterminer

// AvailableTransitions returns human-readable descriptions of transitions from a state.
func (ptc *ProjectTypeConfig) AvailableTransitions(s State) []string
```

### 2. Migrate PhaseConfig (phase_config.go)

Migrate phase configuration:

```go
// PhaseConfig defines the configuration for a single phase.
type PhaseConfig struct {
    name              string
    startState        State
    endState          State
    allowedInputTypes []string
    allowedOutputTypes []string
    supportsTasks     bool
    metadataSchema    string
}

// Name returns the phase name.
func (pc *PhaseConfig) Name() string

// StartState returns the phase's start state.
func (pc *PhaseConfig) StartState() State

// EndState returns the phase's end state.
func (pc *PhaseConfig) EndState() State

// AllowedInputTypes returns the allowed input artifact types.
func (pc *PhaseConfig) AllowedInputTypes() []string

// AllowedOutputTypes returns the allowed output artifact types.
func (pc *PhaseConfig) AllowedOutputTypes() []string

// SupportsTasks returns whether the phase supports task management.
func (pc *PhaseConfig) SupportsTasks() bool

// MetadataSchema returns the CUE schema for phase metadata validation.
func (pc *PhaseConfig) MetadataSchema() string
```

### 3. Migrate TransitionConfig (transition_config.go)

Migrate transition configuration:

```go
// TransitionConfig defines a single state transition.
type TransitionConfig struct {
    From          State
    To            State
    Event         Event
    guardTemplate GuardTemplate
    onEntry       Action
    onExit        Action
    failedPhase   string
    description   string
}

// GuardDescription returns the guard description.
func (tc *TransitionConfig) GuardDescription() string

// Description returns the transition description.
func (tc *TransitionConfig) Description() string
```

### 4. Migrate BranchConfig (branch.go)

Migrate branch configuration from `cli/internal/sdks/project/branch.go`:

```go
// BranchConfig represents a state-determined branch point.
type BranchConfig struct {
    from          State
    discriminator func(*state.Project) string
    branches      map[string]*BranchPath
}

// BranchPath represents one possible branch destination.
type BranchPath struct {
    value         string
    event         Event
    to            State
    description   string
    guardTemplate GuardTemplate
    onEntry       Action
    onExit        Action
    failedPhase   string
}
```

### 5. Migrate ProjectTypeConfigBuilder (builder.go in project package)

Migrate the builder from `cli/internal/sdks/project/builder.go`:

```go
// ProjectTypeConfigBuilder provides a fluent API for building project type configs.
type ProjectTypeConfigBuilder struct {
    config *ProjectTypeConfig
}

// NewProjectTypeConfigBuilder creates a new builder for a project type.
func NewProjectTypeConfigBuilder(name string) *ProjectTypeConfigBuilder

// SetInitialState sets the initial state for new projects.
func (b *ProjectTypeConfigBuilder) SetInitialState(s State) *ProjectTypeConfigBuilder

// AddPhase adds a phase configuration.
func (b *ProjectTypeConfigBuilder) AddPhase(name string, opts ...PhaseOpt) *ProjectTypeConfigBuilder

// AddTransition adds a state transition.
func (b *ProjectTypeConfigBuilder) AddTransition(from, to State, event Event, opts ...TransitionOption) *ProjectTypeConfigBuilder

// AddBranch adds a branch point configuration.
func (b *ProjectTypeConfigBuilder) AddBranch(from State, opts ...BranchOption) *ProjectTypeConfigBuilder

// SetPrompt sets the prompt generator for a state.
func (b *ProjectTypeConfigBuilder) SetPrompt(s State, gen PromptGenerator) *ProjectTypeConfigBuilder

// SetInitializer sets the project initialization function.
func (b *ProjectTypeConfigBuilder) SetInitializer(fn func(*state.Project, map[string][]project.ArtifactState) error) *ProjectTypeConfigBuilder

// SetValidator sets the project validation function.
func (b *ProjectTypeConfigBuilder) SetValidator(fn func(*state.Project) error) *ProjectTypeConfigBuilder

// Build creates the final ProjectTypeConfig.
func (b *ProjectTypeConfigBuilder) Build() *ProjectTypeConfig
```

### 6. Migrate Options (options.go in project package)

Migrate option functions from `cli/internal/sdks/project/options.go`:

```go
// PhaseOpt is a function that modifies a PhaseConfig.
type PhaseOpt func(*PhaseConfig)

func WithStartState(s State) PhaseOpt
func WithEndState(s State) PhaseOpt
func WithInputs(types ...string) PhaseOpt
func WithOutputs(types ...string) PhaseOpt
func WithTasks() PhaseOpt
func WithMetadataSchema(schema string) PhaseOpt

// TransitionOption is a function that modifies a TransitionConfig.
// (Note: This may conflict with state machine TransitionOption - consider namespacing)
type TransitionOption func(*TransitionConfig)

func WithGuard(description string, guardFunc func(*state.Project) bool) TransitionOption
func WithOnEntry(action Action) TransitionOption
func WithOnExit(action Action) TransitionOption
func WithFailedPhase(phaseName string) TransitionOption
func WithDescription(description string) TransitionOption

// BranchOption configures a BranchConfig.
type BranchOption func(*BranchConfig)

func BranchOn(discriminator func(*state.Project) string) BranchOption
func When(value string, event Event, to State, opts ...TransitionOption) BranchOption
```

### 7. Resolve Naming Conflicts

The project package has `TransitionOption` for project-level transitions, but the state machine also has `TransitionOption`. Handle this by:

**Option A**: Rename project-level options
```go
// In project package
type ProjectTransitionOption func(*TransitionConfig)
func WithProjectGuard(...) ProjectTransitionOption
```

**Option B**: Keep both with clear documentation
```go
// project.TransitionOption - for project type configuration
// project.WithGuard() - takes *state.Project

// State machine uses MachineBuilder with its own options internally
```

**Recommended: Option B** - Keep the names since they're in different contexts. The builder internally converts project-level options to state machine options.

## Acceptance Criteria

1. [ ] `config.go` defines ProjectTypeConfig with all methods
2. [ ] `phase_config.go` defines PhaseConfig with all methods
3. [ ] `transition_config.go` defines TransitionConfig
4. [ ] `branch.go` defines BranchConfig and BranchPath
5. [ ] `builder.go` defines ProjectTypeConfigBuilder with fluent API
6. [ ] `options.go` defines all option functions
7. [ ] BuildMachine correctly creates bound state machine
8. [ ] FireWithPhaseUpdates manages phase status correctly
9. [ ] All types properly documented with doc comments
10. [ ] Code compiles without import cycles
11. [ ] `golangci-lint run ./...` passes with no issues
12. [ ] `go test -race ./...` passes with no failures
13. [ ] Code adheres to `.standards/STYLE.md` (file organization, naming, error handling)
14. [ ] Tests adhere to `.standards/TESTING.md` (table-driven, testify assertions, behavioral coverage)

### Test Requirements (TDD)

Write comprehensive tests:

**config_test.go:**
- Name returns correct name
- InitialState returns configured state
- GetPhaseForState returns correct phase
- IsPhaseStartState/IsPhaseEndState return correct values
- Initialize creates phases correctly
- Validate runs validator function

**builder_test.go:**
- NewProjectTypeConfigBuilder creates builder
- SetInitialState sets state
- AddPhase adds phase with options
- AddTransition adds transition
- AddBranch adds branch config
- SetPrompt sets prompt generator
- Build creates complete config

**machine_integration_test.go:**
- BuildMachine creates working machine
- Guards are bound to project instance
- Actions execute with project
- FireWithPhaseUpdates manages phase status

**branch_test.go:**
- BranchOn sets discriminator
- When adds branch paths
- GetEventDeterminer returns working determiner

## Technical Details

### Import Structure

```go
// In project package
import (
    "github.com/jmgilman/sow/libs/project/state"
    "github.com/jmgilman/sow/libs/schemas/project"
)

// In state package
import (
    // NO import of parent project package - would create cycle
)
```

### Avoiding Import Cycles

The state package needs to know about ProjectTypeConfig for the Project type, but project package imports state. Solutions:

1. **Define interface in state package**:
```go
// In state/config.go
type ProjectTypeConfig interface {
    Name() string
    InitialState() string  // Return string, not project.State
    Initialize(p *Project, inputs map[string][]schemas.ArtifactState) error
    Validate(p *Project) error
    BuildMachine(p *Project, initialState string) interface{}
}
```

2. **Project type holds interface, cast at runtime**:
```go
// Project.Config() returns state.ProjectTypeConfig
// project.ProjectTypeConfig implements state.ProjectTypeConfig
```

### BuildMachine Implementation

The BuildMachine method creates a state machine with guards/actions bound to the project:

```go
func (ptc *ProjectTypeConfig) BuildMachine(p *state.Project, initialState State) *Machine {
    builder := NewBuilder(initialState, ptc.makePromptFunc(p))

    for _, tc := range ptc.transitions {
        var opts []TransitionOption

        // Bind guard to project
        if tc.guardTemplate.Func != nil {
            opts = append(opts, WithGuardDescription(
                tc.guardTemplate.Description,
                func() bool { return tc.guardTemplate.Func(p) },
            ))
        }

        // Bind actions to project
        if tc.onEntry != nil {
            opts = append(opts, WithOnEntry(func(ctx context.Context, args ...any) error {
                return tc.onEntry(p)
            }))
        }

        // ... similar for onExit

        builder.AddTransition(tc.From, tc.To, tc.Event, opts...)
    }

    return builder.Build()
}
```

## Relevant Inputs

- `cli/internal/sdks/project/config.go` - Current ProjectTypeConfig
- `cli/internal/sdks/project/builder.go` - Current ProjectTypeConfigBuilder
- `cli/internal/sdks/project/options.go` - Current option functions
- `cli/internal/sdks/project/branch.go` - Current branch configuration
- `cli/internal/sdks/project/machine.go` - BuildMachine and FireWithPhaseUpdates
- `cli/internal/sdks/project/types.go` - Type definitions
- `.standards/STYLE.md` - Code style requirements
- `.standards/TESTING.md` - Testing requirements

## Examples

### Building a Project Type Config

```go
config := project.NewProjectTypeConfigBuilder("standard").
    SetInitialState(project.State("PlanningActive")).
    AddPhase("planning",
        project.WithStartState(project.State("PlanningActive")),
        project.WithEndState(project.State("PlanningActive")),
        project.WithOutputs("task_list"),
        project.WithTasks(),
    ).
    AddPhase("implementation",
        project.WithStartState(project.State("ImplementationPlanning")),
        project.WithEndState(project.State("ImplementationExecuting")),
        project.WithInputs("task_list"),
        project.WithTasks(),
    ).
    AddTransition(
        project.State("PlanningActive"),
        project.State("ImplementationPlanning"),
        project.Event("AdvancePlanning"),
        project.WithGuard("planning artifacts approved", func(p *state.Project) bool {
            return allPlanningArtifactsApproved(p)
        }),
        project.WithDescription("Complete planning and begin implementation"),
    ).
    SetPrompt(project.State("PlanningActive"), func(p *state.Project) string {
        return "Create task breakdown for: " + p.Description
    }).
    SetInitializer(func(p *state.Project, inputs map[string][]schemas.ArtifactState) error {
        // Initialize phases
        return nil
    }).
    Build()

// Register with registry
project.Register("standard", config)
```

### Using Branch Configuration

```go
config := project.NewProjectTypeConfigBuilder("standard").
    // ... phases ...
    AddBranch(project.State("ReviewActive"),
        project.BranchOn(func(p *state.Project) string {
            // Examine latest review artifact
            return getReviewAssessment(p) // "pass" or "fail"
        }),
        project.When("pass",
            project.Event("ReviewPass"),
            project.State("FinalizeChecks"),
            project.WithDescription("Review approved - proceed to finalization"),
        ),
        project.When("fail",
            project.Event("ReviewFail"),
            project.State("ImplementationPlanning"),
            project.WithFailedPhase("review"),
            project.WithDescription("Review rejected - return to implementation"),
        ),
    ).
    Build()
```

## Dependencies

- Task 010: Core types (State, Event, Guard, Action, etc.)
- Task 040: State wrapper types (state.Project)
- Task 050: State machine and builder

## Constraints

- Avoid import cycles between `project` and `project/state` packages
- Keep builder API fluent and chainable
- Preserve all existing functionality from the CLI packages
- Document the relationship between project-level and state machine-level options
- FireWithPhaseUpdates must correctly handle failed phase marking
