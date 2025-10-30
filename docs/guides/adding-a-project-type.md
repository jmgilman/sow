# Adding a New Project Type

This guide shows you how to implement a new project type using the State Machine SDK. It assumes you're familiar with Go and the sow architecture.

## Prerequisites

Before implementing a new project type:

- Read [`docs/architecture/05-building-blocks.md`](../architecture/05-building-blocks.md) - Explains the SDK layer
- Review [`cli/internal/project/standard/`](../../cli/internal/project/standard/) - Reference implementation
- Understand finite state machines and the builder pattern

**Estimated time**: 12-16 hours for a typical project type (4-5 states, 3-4 phases)

## Quick Reference

### Implementation Order

1. Define CUE schema (`cli/schemas/projects/{type}.cue`)
2. Define states (`cli/internal/project/{type}/states.go`)
3. Define events (`cli/internal/project/{type}/events.go`)
4. Implement guards (`cli/internal/project/{type}/guards.go`)
5. Implement phases (`cli/internal/project/{type}/{phase}.go`)
6. Create prompt generator (`cli/internal/project/{type}/prompts.go`)
7. Build main project struct (`cli/internal/project/{type}/project.go`)
8. Update loader (`cli/internal/project/loader/loader.go`)

### File Structure

```
cli/internal/project/{type}/
├── project.go          # Main struct, buildStateMachine(), Project interface
├── states.go           # State constants
├── events.go           # Event constants
├── guards.go           # Guard functions
├── prompts.go          # PromptGenerator implementation
├── {phase1}.go         # Phase implementations
├── {phase2}.go
└── phases_test.go      # Integration tests
```

## 1. Define CUE Schema

Create `cli/schemas/projects/{type}.cue`:

```cue
package projects

import "time"
import "github.com/jmgilman/sow/cli/schemas/phases"

#ExplorationProjectState: {
	statechart: {
		current_state:  string
		previous_state: string | *null
		history: [...{
			from:      string
			to:        string
			event:     string
			timestamp: time.Time
		}] | *null
	}

	project: {
		type:         "exploration"  // Must match type name
		name:         string
		branch:       string
		description:  string
		github_issue: int64 | *null
		created_at:   time.Time
		updated_at:   time.Time
	}

	phases: {
		discovery: phases.#Phase
		analysis:  phases.#Phase
		synthesis: phases.#Phase
	}
}
```

**Required fields**: `statechart`, `project.type` (literal), `project.name/branch/description`, `project.created_at/updated_at`, `phases.<your-phases>`

Generate Go types: `cd cli && make generate`

## 2. Define States

Create `cli/internal/project/{type}/states.go`:

```go
package exploration

import "github.com/jmgilman/sow/cli/internal/project/statechart"

const (
	// DiscoveryActive - orchestrator gathers context
	DiscoveryActive = statechart.State("DiscoveryActive")

	// AnalysisActive - orchestrator analyzes options
	AnalysisActive = statechart.State("AnalysisActive")

	// SynthesisActive - orchestrator synthesizes findings
	SynthesisActive = statechart.State("SynthesisActive")

	// ExplorationComplete - all findings documented
	ExplorationComplete = statechart.State("ExplorationComplete")
)
```

**Naming**: Use PascalCase descriptive nouns/adjectives. See `cli/internal/project/standard/states.go` for examples.

## 3. Define Events

Create `cli/internal/project/{type}/events.go`:

```go
package exploration

import "github.com/jmgilman/sow/cli/internal/project/statechart"

const (
	// EventProjectInit - project created
	// Transitions: NoProject → DiscoveryActive
	EventProjectInit = statechart.Event("project_init")

	// EventBeginAnalysis - discovery artifacts approved
	// Transitions: DiscoveryActive → AnalysisActive
	EventBeginAnalysis = statechart.Event("begin_analysis")

	// EventBeginSynthesis - analysis complete
	// Transitions: AnalysisActive → SynthesisActive
	EventBeginSynthesis = statechart.Event("begin_synthesis")

	// EventCompleteExploration - synthesis complete
	// Transitions: SynthesisActive → ExplorationComplete
	EventCompleteExploration = statechart.Event("complete_exploration")

	// EventProjectDelete - project cleanup
	// Transitions: any state → NoProject
	EventProjectDelete = statechart.Event("project_delete")
)
```

**Naming**: Use `Event` prefix + past tense verb. Document transitions and guards.

## 4. Implement Guards

Create `cli/internal/project/{type}/guards.go`:

```go
package exploration

import (
	"github.com/jmgilman/sow/cli/internal/project/statechart"
	"github.com/jmgilman/sow/cli/schemas"
)

// DiscoveryArtifactsApproved checks if discovery has approved context artifact
func DiscoveryArtifactsApproved(state *schemas.ProjectState) bool {
	hasContext := statechart.HasArtifactWithType(
		state.Phases.Discovery.Artifacts,
		"context",
	)
	if !hasContext {
		return false
	}
	return statechart.ArtifactsApproved(state.Phases.Discovery.Artifacts)
}

// AnalysisComplete checks if analysis has approved options artifact
func AnalysisComplete(state *schemas.ProjectState) bool {
	hasOptions := statechart.HasArtifactWithType(
		state.Phases.Analysis.Artifacts,
		"options",
	)
	return hasOptions && statechart.ArtifactsApproved(state.Phases.Analysis.Artifacts)
}
```

**Critical**: Guards must be **pure functions** - no side effects, only read state. Reuse SDK guards (`TasksComplete`, `ArtifactsApproved`, `HasArtifactWithType`, `MinTaskCount`).

## 5. Implement Phases

Create one file per phase (e.g., `cli/internal/project/{type}/discovery.go`):

```go
package exploration

import (
	"fmt"
	"time"
	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/domain"
	"github.com/jmgilman/sow/cli/internal/sow"
	phasesSchema "github.com/jmgilman/sow/cli/schemas/phases"
)

type DiscoveryPhase struct {
	state     *phasesSchema.Phase
	artifacts *project.ArtifactCollection
	project   *ExplorationProject
	ctx       *sow.Context
}

func NewDiscoveryPhase(state *phasesSchema.Phase, proj *ExplorationProject, ctx *sow.Context) *DiscoveryPhase {
	return &DiscoveryPhase{
		state:     state,
		artifacts: project.NewArtifactCollection(state, proj),
		project:   proj,
		ctx:       ctx,
	}
}

// Implement Phase interface methods

func (p *DiscoveryPhase) Name() string { return "discovery" }
func (p *DiscoveryPhase) Status() string { return p.state.Status }
func (p *DiscoveryPhase) Enabled() bool { return p.state.Enabled }

// Artifact operations (if supported)
func (p *DiscoveryPhase) AddArtifact(path string, opts ...domain.ArtifactOption) error {
	return p.artifacts.Add(path, opts...)
}

func (p *DiscoveryPhase) ApproveArtifact(path string) (*domain.PhaseOperationResult, error) {
	if err := p.artifacts.Approve(path); err != nil {
		return nil, err
	}

	// Save BEFORE returning event
	if err := p.project.Save(); err != nil {
		return nil, err
	}

	// Conditionally return event when all artifacts approved
	if allApproved(p.state.Artifacts) {
		return domain.WithEvent(EventBeginAnalysis), nil
	}
	return domain.NoEvent(), nil
}

func (p *DiscoveryPhase) ListArtifacts() []*phasesSchema.Artifact {
	return p.artifacts.List()
}

// Task operations (not supported - return ErrNotSupported)
func (p *DiscoveryPhase) AddTask(_ string, _ ...domain.TaskOption) (*domain.Task, error) {
	return nil, project.ErrNotSupported
}

func (p *DiscoveryPhase) Complete() (*domain.PhaseOperationResult, error) {
	p.state.Status = "completed"
	now := time.Now()
	p.state.Completed_at = &now

	// Save BEFORE returning event
	if err := p.project.Save(); err != nil {
		return nil, err
	}

	return domain.WithEvent(EventBeginAnalysis), nil
}

// ... implement remaining Phase interface methods
```

**Key patterns**:
- **Save-before-event**: Always `Save()` before returning `WithEvent()`
- **Conditional events**: Return `WithEvent()` or `NoEvent()` based on state
- **Not supported**: Return `project.ErrNotSupported` for unsupported operations
- **Use ArtifactCollection**: Reuse the helper for artifact management

See `cli/internal/project/standard/planning.go` for complete phase example.

## 6. Create Prompt Generator

Create `cli/internal/project/{type}/prompts.go`:

```go
package exploration

import (
	"fmt"
	"strings"
	"github.com/jmgilman/sow/cli/internal/prompts"
	"github.com/jmgilman/sow/cli/internal/project/statechart"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
)

type ExplorationPromptGenerator struct {
	components *statechart.PromptComponents
}

func NewExplorationPromptGenerator(ctx *sow.Context) *ExplorationPromptGenerator {
	return &ExplorationPromptGenerator{
		components: statechart.NewPromptComponents(ctx),
	}
}

func (g *ExplorationPromptGenerator) GeneratePrompt(
	state statechart.State,
	projectState *schemas.ProjectState,
) (string, error) {
	switch state {
	case statechart.NoProject:
		return "", nil
	case DiscoveryActive:
		return g.generateDiscoveryPrompt(projectState)
	case AnalysisActive:
		return g.generateAnalysisPrompt(projectState)
	default:
		return "", fmt.Errorf("unknown state: %s", state)
	}
}

func (g *ExplorationPromptGenerator) generateDiscoveryPrompt(
	projectState *schemas.ProjectState,
) (string, error) {
	var buf strings.Builder

	// Reuse SDK component for header
	buf.WriteString(g.components.ProjectHeader(projectState))
	buf.WriteString("\n")

	// Gracefully degrade if git unavailable
	gitStatus, err := g.components.GitStatus()
	if err != nil {
		buf.WriteString("## Git Status\n\nUnavailable\n\n")
	} else {
		buf.WriteString(gitStatus)
		buf.WriteString("\n")
	}

	// Render template for static guidance
	templateCtx := &prompts.StatechartContext{
		State:        string(DiscoveryActive),
		ProjectState: projectState,
	}
	guidance, err := g.components.RenderTemplate(
		prompts.PromptExplorationDiscovery,
		templateCtx,
	)
	if err != nil {
		return "", err
	}
	buf.WriteString(guidance)

	return buf.String(), nil
}

// ... implement other state-specific prompt methods
```

**Reusable components**:
- `g.components.ProjectHeader(projectState)` - Project info
- `g.components.GitStatus()` - Current git status
- `g.components.RecentCommits(n)` - Recent commits
- `g.components.TaskSummary(tasks)` - Task list
- `g.components.RenderTemplate(id, ctx)` - Static templates

See `cli/internal/project/standard/prompts.go` for complete example.

## 7. Build Main Project Struct

Create `cli/internal/project/{type}/project.go`:

```go
package exploration

import (
	"fmt"
	"time"
	"github.com/jmgilman/sow/cli/internal/logging"
	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/domain"
	"github.com/jmgilman/sow/cli/internal/project/statechart"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/jmgilman/sow/cli/schemas/projects"
)

type ExplorationProject struct {
	state   *projects.ExplorationProjectState
	ctx     *sow.Context
	machine *statechart.Machine
	phases  map[string]domain.Phase
}

func New(state *projects.ExplorationProjectState, ctx *sow.Context) *ExplorationProject {
	p := &ExplorationProject{
		state:  state,
		ctx:    ctx,
		phases: make(map[string]domain.Phase),
	}

	// Create phase instances
	p.phases["discovery"] = NewDiscoveryPhase(&state.Phases.Discovery, p, ctx)
	p.phases["analysis"] = NewAnalysisPhase(&state.Phases.Analysis, p, ctx)
	p.phases["synthesis"] = NewSynthesisPhase(&state.Phases.Synthesis, p, ctx)

	// Build state machine
	p.machine = p.buildStateMachine()

	return p
}

func (p *ExplorationProject) buildStateMachine() *statechart.Machine {
	currentState := statechart.State(p.state.Statechart.Current_state)
	projectState := (*schemas.ProjectState)(p.state)
	promptGen := NewExplorationPromptGenerator(p.ctx)

	builder := statechart.NewBuilder(currentState, projectState, promptGen)

	builder.
		// Unconditional transition (no guard)
		AddTransition(
			statechart.NoProject,
			DiscoveryActive,
			EventProjectInit,
		).
		// Conditional transition with guard
		AddTransition(
			DiscoveryActive,
			AnalysisActive,
			EventBeginAnalysis,
			statechart.WithGuard(func() bool {
				return DiscoveryArtifactsApproved(projectState)
			}),
		).
		AddTransition(
			AnalysisActive,
			SynthesisActive,
			EventBeginSynthesis,
			statechart.WithGuard(func() bool {
				return AnalysisComplete(projectState)
			}),
		).
		// Deletion transitions from any state
		AddTransition(DiscoveryActive, statechart.NoProject, EventProjectDelete).
		AddTransition(AnalysisActive, statechart.NoProject, EventProjectDelete)

	machine := builder.Build()
	machine.SetFilesystem(p.ctx.FS())
	return machine
}

// Implement Project interface

func (p *ExplorationProject) Name() string { return p.state.Project.Name }
func (p *ExplorationProject) Branch() string { return p.state.Project.Branch }
func (p *ExplorationProject) Description() string { return p.state.Project.Description }
func (p *ExplorationProject) Type() string { return "exploration" }

func (p *ExplorationProject) CurrentPhase() domain.Phase {
	currentState := p.machine.State()
	switch currentState {
	case DiscoveryActive:
		return p.phases["discovery"]
	case AnalysisActive:
		return p.phases["analysis"]
	case SynthesisActive, ExplorationComplete:
		return p.phases["synthesis"]
	default:
		return nil
	}
}

func (p *ExplorationProject) Phase(name string) (domain.Phase, error) {
	phase, ok := p.phases[name]
	if !ok {
		return nil, project.ErrPhaseNotFound
	}
	return phase, nil
}

func (p *ExplorationProject) Machine() *statechart.Machine { return p.machine }
func (p *ExplorationProject) InitialState() statechart.State { return DiscoveryActive }

func (p *ExplorationProject) Save() error {
	return p.machine.Save()
}

func (p *ExplorationProject) Log(action, result string, opts ...domain.LogOption) error {
	entry := &logging.LogEntry{
		Timestamp: time.Now(),
		AgentID:   "orchestrator",
		Action:    action,
		Result:    result,
	}
	for _, opt := range opts {
		opt(entry)
	}
	return logging.AppendLog(p.ctx.FS(), "project/log.md", entry)
}

// ... implement remaining Project interface methods
```

**Critical patterns**:
- **Guard closures**: Capture `projectState` in guard functions
- **Unconditional transitions**: Omit `WithGuard()` option
- **Type conversion**: Cast `*projects.ExplorationProjectState` to `*schemas.ProjectState`

See `cli/internal/project/standard/project.go` for complete example.

## 8. Update Loader

Edit `cli/internal/project/loader/loader.go`:

```go
package loader

import (
	"fmt"
	"github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/project/domain"
	"github.com/jmgilman/sow/cli/internal/project/exploration"  // Add import
	"github.com/jmgilman/sow/cli/internal/project/standard"
	"github.com/jmgilman/sow/cli/internal/project/statechart"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas/projects"
)

func Load(ctx *sow.Context) (domain.Project, error) {
	// ... existing checks ...

	state, _, err := statechart.LoadProjectState(ctx.FS())
	if err != nil {
		return nil, err
	}

	projectType := state.Project.Type

	switch projectType {
	case "standard":
		standardState := (*projects.StandardProjectState)(state)
		return standard.New(standardState, ctx), nil

	case "exploration":  // Add your type
		explorationState := (*projects.ExplorationProjectState)(state)
		return exploration.New(explorationState, ctx), nil

	default:
		return nil, fmt.Errorf("unknown project type: %s", projectType)
	}
}
```

## Common Patterns

### Save-Before-Event Pattern

Always save state before returning events:

```go
func (p *Phase) Complete() (*domain.PhaseOperationResult, error) {
	p.state.Status = "completed"

	// SAVE FIRST
	if err := p.project.Save(); err != nil {
		return nil, err
	}

	// THEN return event
	return domain.WithEvent(EventPhaseComplete), nil
}
```

### Conditional Event Returns

Return events based on state conditions:

```go
func (p *Phase) ApproveArtifact(path string) (*domain.PhaseOperationResult, error) {
	p.artifacts.Approve(path)
	p.project.Save()

	// Check condition
	if allArtifactsApproved() {
		return domain.WithEvent(EventAllApproved), nil
	}

	return domain.NoEvent(), nil
}
```

### Guard Purity

Guards must be pure functions - no side effects:

```go
// GOOD - pure function
func AnalysisComplete(state *schemas.ProjectState) bool {
	return statechart.ArtifactsApproved(state.Phases.Analysis.Artifacts)
}

// BAD - has side effects
func AnalysisComplete(state *schemas.ProjectState) bool {
	state.Phases.Analysis.Status = "checked"  // Side effect!
	return true
}
```

### Metadata Checks

For custom flags in metadata maps:

```go
func FlagIsSet(state *schemas.ProjectState) bool {
	if state.Phases.Discovery.Metadata != nil {
		if flag, ok := state.Phases.Discovery.Metadata["my_flag"].(bool); ok {
			return flag
		}
	}
	return false
}
```

## Common Gotchas

### 1. Type Conversion

```go
// CORRECT - cast to generic ProjectState for machine
projectState := (*schemas.ProjectState)(p.state)

// WRONG - don't use specific type with builder
projectState := p.state  // Won't compile
```

### 2. Forgetting to Save

```go
// WRONG - event fires but state not saved
return domain.WithEvent(EventComplete), nil

// CORRECT - save before event
if err := p.project.Save(); err != nil {
	return nil, err
}
return domain.WithEvent(EventComplete), nil
```

### 3. Guard Side Effects

Guards are evaluated multiple times - never mutate state in guards.

### 4. Nil Metadata

Always check metadata maps before accessing:

```go
if p.state.Metadata != nil {
	val := p.state.Metadata["key"]
}
```

### 5. State vs Event Naming

- **States**: Present tense descriptors (`AnalysisActive`, `ReviewPending`)
- **Events**: Past tense actions (`EventBeginAnalysis`, `EventTasksApproved`)

## Testing

Reference `cli/internal/project/standard/phases_test.go` for integration testing patterns. Key areas to test:

- State machine construction
- Guard evaluation (pure functions, easy to unit test)
- Phase operations return correct `PhaseOperationResult`
- Prompt generation (all states covered)
- Loader integration

## References

- **SDK components**: [`docs/architecture/05-building-blocks.md`](../architecture/05-building-blocks.md)
- **Reference implementation**: [`cli/internal/project/standard/`](../../cli/internal/project/standard/)
- **Domain interfaces**: [`cli/internal/project/domain/interfaces.go`](../../cli/internal/project/domain/interfaces.go)
- **Phase helpers**: [`cli/internal/project/tasks_collection.go`](../../cli/internal/project/tasks_collection.go)
