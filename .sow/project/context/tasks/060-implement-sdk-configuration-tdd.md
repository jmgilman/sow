# Task 060: Implement SDK Configuration (TDD)

## Overview

Create the main project type configuration using the SDK builder API with Test-Driven Development. This wires together all components (phases, states, events, guards, prompts) into a complete, functional project type.

**TDD Requirement**: Write integration tests FIRST, then implement configuration to make tests pass.

## Context

**Design Reference**:
- `.sow/knowledge/designs/project-sdk-implementation.md` (lines 673-831) - Complete configuration example
- `cli/internal/sdks/project/builder.go` - SDK builder API
- `cli/internal/sdks/project/machine.go` - Machine building with closures

**What This Does**: Configures the standard project type by:
1. Defining 4 phases (planning, implementation, review, finalize)
2. Setting up state machine with 10 transitions and guards
3. Registering OnAdvance event determiners for each state
4. Attaching prompt generators to each state
5. Registering with global project type registry

## TDD Workflow

### Red Phase: Write Failing Integration Tests

Create `cli/internal/projects/standard/lifecycle_test.go` with comprehensive end-to-end tests.

### Green Phase: Implement Configuration

Create `cli/internal/projects/standard/standard.go` with SDK builder configuration.

### Refactor Phase: Clean Up

Improve implementation while keeping tests green.

## Requirements

### Test File Structure

Create `cli/internal/projects/standard/lifecycle_test.go`:

```go
package standard

import (
	"context"
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/internal/sdks/project"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
)

// TestFullLifecycle tests complete project lifecycle from start to finish
func TestFullLifecycle(t *testing.T) {
	// Setup: Create minimal project in NoProject state
	proj := createTestProject(t, NoProject)

	// NoProject → PlanningActive
	t.Run("init transitions to PlanningActive", func(t *testing.T) {
		err := proj.Machine().Fire(EventProjectInit)
		if err != nil {
			t.Fatalf("Fire(EventProjectInit) failed: %v", err)
		}
		if got := proj.Machine().State(); got != sdkstate.State(PlanningActive) {
			t.Errorf("state = %v, want %v", got, PlanningActive)
		}
	})

	// PlanningActive → ImplementationPlanning
	t.Run("planning completion transitions to implementation planning", func(t *testing.T) {
		// Add and approve task_list output
		addApprovedOutput(t, proj, "planning", "task_list", "tasks.md")

		err := proj.Machine().Fire(EventCompletePlanning)
		if err != nil {
			t.Fatalf("Fire(EventCompletePlanning) failed: %v", err)
		}
		if got := proj.Machine().State(); got != sdkstate.State(ImplementationPlanning) {
			t.Errorf("state = %v, want %v", got, ImplementationPlanning)
		}
	})

	// ImplementationPlanning → ImplementationExecuting
	t.Run("task approval transitions to execution", func(t *testing.T) {
		// Set tasks_approved metadata
		setPhaseMetadata(t, proj, "implementation", "tasks_approved", true)

		err := proj.Machine().Fire(EventTasksApproved)
		if err != nil {
			t.Fatalf("Fire(EventTasksApproved) failed: %v", err)
		}
		if got := proj.Machine().State(); got != sdkstate.State(ImplementationExecuting) {
			t.Errorf("state = %v, want %v", got, ImplementationExecuting)
		}
	})

	// ImplementationExecuting → ReviewActive
	t.Run("all tasks complete transitions to review", func(t *testing.T) {
		// Add completed tasks
		addCompletedTask(t, proj, "implementation", "001", "Task 1")
		addCompletedTask(t, proj, "implementation", "002", "Task 2")

		err := proj.Machine().Fire(EventAllTasksComplete)
		if err != nil {
			t.Fatalf("Fire(EventAllTasksComplete) failed: %v", err)
		}
		if got := proj.Machine().State(); got != sdkstate.State(ReviewActive) {
			t.Errorf("state = %v, want %v", got, ReviewActive)
		}
	})

	// ReviewActive → FinalizeDocumentation
	t.Run("review pass transitions to finalize", func(t *testing.T) {
		// Add approved review with pass assessment
		addApprovedReview(t, proj, "review", "pass", "review.md")

		err := proj.Machine().Fire(EventReviewPass)
		if err != nil {
			t.Fatalf("Fire(EventReviewPass) failed: %v", err)
		}
		if got := proj.Machine().State(); got != sdkstate.State(FinalizeDocumentation) {
			t.Errorf("state = %v, want %v", got, FinalizeDocumentation)
		}
	})

	// Finalize substates
	t.Run("finalize substates progress correctly", func(t *testing.T) {
		// FinalizeDocumentation → FinalizeChecks
		err := proj.Machine().Fire(EventDocumentationDone)
		if err != nil {
			t.Fatalf("Fire(EventDocumentationDone) failed: %v", err)
		}
		if got := proj.Machine().State(); got != sdkstate.State(FinalizeChecks) {
			t.Errorf("state = %v, want %v", got, FinalizeChecks)
		}

		// FinalizeChecks → FinalizeDelete
		err = proj.Machine().Fire(EventChecksDone)
		if err != nil {
			t.Fatalf("Fire(EventChecksDone) failed: %v", err)
		}
		if got := proj.Machine().State(); got != sdkstate.State(FinalizeDelete) {
			t.Errorf("state = %v, want %v", got, FinalizeDelete)
		}

		// FinalizeDelete → NoProject
		setPhaseMetadata(t, proj, "finalize", "project_deleted", true)
		err = proj.Machine().Fire(EventProjectDelete)
		if err != nil {
			t.Fatalf("Fire(EventProjectDelete) failed: %v", err)
		}
		if got := proj.Machine().State(); got != sdkstate.State(NoProject) {
			t.Errorf("state = %v, want %v", got, NoProject)
		}
	})
}

// TestReviewFailLoop tests review failure rework loop
func TestReviewFailLoop(t *testing.T) {
	proj := createTestProject(t, ReviewActive)

	// Add approved review with fail assessment
	addApprovedReview(t, proj, "review", "fail", "review-fail.md")

	// ReviewActive → ImplementationPlanning (rework)
	err := proj.Machine().Fire(EventReviewFail)
	if err != nil {
		t.Fatalf("Fire(EventReviewFail) failed: %v", err)
	}
	if got := proj.Machine().State(); got != sdkstate.State(ImplementationPlanning) {
		t.Errorf("state = %v, want %v", got, ImplementationPlanning)
	}
}

// TestGuardsBlockInvalidTransitions tests guards prevent invalid transitions
func TestGuardsBlockInvalidTransitions(t *testing.T) {
	tests := []struct {
		name         string
		initialState sdkstate.State
		setupFunc    func(*state.Project)
		event        sdkstate.Event
		shouldBlock  bool
	}{
		{
			name:         "planning without approved task_list blocks",
			initialState: sdkstate.State(PlanningActive),
			setupFunc:    func(p *state.Project) {}, // No task list
			event:        sdkstate.Event(EventCompletePlanning),
			shouldBlock:  true,
		},
		{
			name:         "implementation planning without tasks_approved blocks",
			initialState: sdkstate.State(ImplementationPlanning),
			setupFunc:    func(p *state.Project) {}, // No metadata set
			event:        sdkstate.Event(EventTasksApproved),
			shouldBlock:  true,
		},
		{
			name:         "implementation executing without completed tasks blocks",
			initialState: sdkstate.State(ImplementationExecuting),
			setupFunc: func(p *state.Project) {
				// Add pending task
				addPendingTask(p, "implementation", "001", "Task 1")
			},
			event:       sdkstate.Event(EventAllTasksComplete),
			shouldBlock: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proj := createTestProject(t, tt.initialState)
			tt.setupFunc(proj)

			can, err := proj.Machine().CanFire(tt.event)
			if err != nil {
				t.Fatalf("CanFire() error: %v", err)
			}

			if tt.shouldBlock && can {
				t.Errorf("guard should block transition but allowed it")
			}
			if !tt.shouldBlock && !can {
				t.Errorf("guard should allow transition but blocked it")
			}
		})
	}
}

// TestPromptGeneration tests prompts generate correctly for each state
func TestPromptGeneration(t *testing.T) {
	states := []sdkstate.State{
		sdkstate.State(PlanningActive),
		sdkstate.State(ImplementationPlanning),
		sdkstate.State(ImplementationExecuting),
		sdkstate.State(ReviewActive),
		sdkstate.State(FinalizeDocumentation),
		sdkstate.State(FinalizeChecks),
		sdkstate.State(FinalizeDelete),
	}

	for _, st := range states {
		t.Run(string(st), func(t *testing.T) {
			proj := createTestProject(t, st)

			prompt := proj.Machine().GeneratePrompt()
			if prompt == "" {
				t.Error("prompt is empty")
			}
			if !contains(prompt, "Project:") {
				t.Error("prompt missing project header")
			}
		})
	}
}

// TestOnAdvanceEventDetermination tests event determiners work correctly
func TestOnAdvanceEventDetermination(t *testing.T) {
	t.Run("ReviewActive determines pass event", func(t *testing.T) {
		proj := createTestProject(t, sdkstate.State(ReviewActive))
		addApprovedReview(t, proj, "review", "pass", "review.md")

		config := project.GetConfig("standard")
		determiner := config.GetEventDeterminer(sdkstate.State(ReviewActive))
		if determiner == nil {
			t.Fatal("no event determiner for ReviewActive")
		}

		event, err := determiner(proj)
		if err != nil {
			t.Fatalf("determiner failed: %v", err)
		}
		if event != sdkstate.Event(EventReviewPass) {
			t.Errorf("event = %v, want %v", event, EventReviewPass)
		}
	})

	t.Run("ReviewActive determines fail event", func(t *testing.T) {
		proj := createTestProject(t, sdkstate.State(ReviewActive))
		addApprovedReview(t, proj, "review", "fail", "review.md")

		config := project.GetConfig("standard")
		determiner := config.GetEventDeterminer(sdkstate.State(ReviewActive))

		event, err := determiner(proj)
		if err != nil {
			t.Fatalf("determiner failed: %v", err)
		}
		if event != sdkstate.Event(EventReviewFail) {
			t.Errorf("event = %v, want %v", event, EventReviewFail)
		}
	})
}

// Helper functions
func createTestProject(t *testing.T, initialState sdkstate.State) *state.Project {
	// Create minimal project state
	// Setup phases, machine, etc.
	// ... implementation details ...
}

func addApprovedOutput(t *testing.T, p *state.Project, phase, typ, path string) {
	// Helper to add approved output artifact
}

func addApprovedReview(t *testing.T, p *state.Project, phase, assessment, path string) {
	// Helper to add approved review artifact with assessment metadata
}

func setPhaseMetadata(t *testing.T, p *state.Project, phase, key string, value interface{}) {
	// Helper to set phase metadata
}

func addCompletedTask(t *testing.T, p *state.Project, phase, id, name string) {
	// Helper to add completed task
}

func addPendingTask(p *state.Project, phase, id, name string) {
	// Helper to add pending task
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
```

### Configuration File Structure

Create `cli/internal/projects/standard/standard.go`:

```go
package standard

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/sdks/project"
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
)

func init() {
	// Register standard project type on package load
	project.Register("standard", NewStandardProjectConfig())
}

// NewStandardProjectConfig creates the complete configuration for standard project type
func NewStandardProjectConfig() *project.ProjectTypeConfig {
	return project.NewProjectTypeConfigBuilder("standard").

		// ===== PHASES =====

		WithPhase("planning",
			project.WithStartState(sdkstate.State(PlanningActive)),
			project.WithEndState(sdkstate.State(PlanningActive)),
			project.WithInputs("context"),
			project.WithOutputs("task_list"),
		).

		WithPhase("implementation",
			project.WithStartState(sdkstate.State(ImplementationPlanning)),
			project.WithEndState(sdkstate.State(ImplementationExecuting)),
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
			project.WithStartState(sdkstate.State(FinalizeDocumentation)),
			project.WithEndState(sdkstate.State(FinalizeDelete)),
			project.WithMetadataSchema(finalizeMetadataSchema),
		).

		// ===== STATE MACHINE =====

		SetInitialState(sdkstate.State(NoProject)).

		// Project initialization
		AddTransition(
			sdkstate.State(NoProject),
			sdkstate.State(PlanningActive),
			sdkstate.Event(EventProjectInit),
		).

		// Planning → Implementation
		AddTransition(
			sdkstate.State(PlanningActive),
			sdkstate.State(ImplementationPlanning),
			sdkstate.Event(EventCompletePlanning),
			project.WithGuard(func(p *state.Project) bool {
				return phaseOutputApproved(p, "planning", "task_list")
			}),
		).

		// Implementation planning → execution
		AddTransition(
			sdkstate.State(ImplementationPlanning),
			sdkstate.State(ImplementationExecuting),
			sdkstate.Event(EventTasksApproved),
			project.WithGuard(func(p *state.Project) bool {
				return phaseMetadataBool(p, "implementation", "tasks_approved")
			}),
		).

		// Implementation → Review
		AddTransition(
			sdkstate.State(ImplementationExecuting),
			sdkstate.State(ReviewActive),
			sdkstate.Event(EventAllTasksComplete),
			project.WithGuard(func(p *state.Project) bool {
				return allTasksComplete(p)
			}),
		).

		// Review → Finalize (pass)
		AddTransition(
			sdkstate.State(ReviewActive),
			sdkstate.State(FinalizeDocumentation),
			sdkstate.Event(EventReviewPass),
			project.WithGuard(func(p *state.Project) bool {
				return latestReviewApproved(p)
			}),
		).

		// Review → Implementation (fail - rework)
		AddTransition(
			sdkstate.State(ReviewActive),
			sdkstate.State(ImplementationPlanning),
			sdkstate.Event(EventReviewFail),
			project.WithGuard(func(p *state.Project) bool {
				return latestReviewApproved(p)
			}),
		).

		// Finalize substates
		AddTransition(
			sdkstate.State(FinalizeDocumentation),
			sdkstate.State(FinalizeChecks),
			sdkstate.Event(EventDocumentationDone),
		).

		AddTransition(
			sdkstate.State(FinalizeChecks),
			sdkstate.State(FinalizeDelete),
			sdkstate.Event(EventChecksDone),
		).

		AddTransition(
			sdkstate.State(FinalizeDelete),
			sdkstate.State(NoProject),
			sdkstate.Event(EventProjectDelete),
			project.WithGuard(func(p *state.Project) bool {
				return projectDeleted(p)
			}),
		).

		// ===== EVENT DETERMINATION =====

		OnAdvance(sdkstate.State(PlanningActive), func(p *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventCompletePlanning), nil
		}).

		OnAdvance(sdkstate.State(ImplementationPlanning), func(p *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventTasksApproved), nil
		}).

		OnAdvance(sdkstate.State(ImplementationExecuting), func(p *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventAllTasksComplete), nil
		}).

		OnAdvance(sdkstate.State(ReviewActive), func(p *state.Project) (sdkstate.Event, error) {
			// Complex: examine review assessment
			phase, exists := p.Phases["review"]
			if !exists {
				return "", fmt.Errorf("review phase not found")
			}

			// Find latest approved review
			var latestReview *state.Artifact
			for i := len(phase.Outputs) - 1; i >= 0; i-- {
				if phase.Outputs[i].Type == "review" && phase.Outputs[i].Approved {
					latestReview = &phase.Outputs[i]
					break
				}
			}

			if latestReview == nil {
				return "", fmt.Errorf("no approved review found")
			}

			// Check assessment metadata
			assessment, ok := latestReview.Metadata["assessment"].(string)
			if !ok {
				return "", fmt.Errorf("review missing assessment")
			}

			switch assessment {
			case "pass":
				return sdkstate.Event(EventReviewPass), nil
			case "fail":
				return sdkstate.Event(EventReviewFail), nil
			default:
				return "", fmt.Errorf("invalid assessment: %s", assessment)
			}
		}).

		OnAdvance(sdkstate.State(FinalizeDocumentation), func(p *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventDocumentationDone), nil
		}).

		OnAdvance(sdkstate.State(FinalizeChecks), func(p *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventChecksDone), nil
		}).

		OnAdvance(sdkstate.State(FinalizeDelete), func(p *state.Project) (sdkstate.Event, error) {
			return sdkstate.Event(EventProjectDelete), nil
		}).

		// ===== PROMPTS =====

		WithPrompt(sdkstate.State(PlanningActive), generatePlanningPrompt).
		WithPrompt(sdkstate.State(ImplementationPlanning), generateImplementationPlanningPrompt).
		WithPrompt(sdkstate.State(ImplementationExecuting), generateImplementationExecutingPrompt).
		WithPrompt(sdkstate.State(ReviewActive), generateReviewPrompt).
		WithPrompt(sdkstate.State(FinalizeDocumentation), generateFinalizeDocumentationPrompt).
		WithPrompt(sdkstate.State(FinalizeChecks), generateFinalizeChecksPrompt).
		WithPrompt(sdkstate.State(FinalizeDelete), generateFinalizeDeletePrompt).

		Build()
}
```

## Acceptance Criteria

### TDD Process
- [ ] Integration tests written FIRST in `lifecycle_test.go`
- [ ] Tests fail initially (red phase)
- [ ] Configuration implemented to make tests pass (green phase)
- [ ] Code refactored while keeping tests green

### Test Coverage
- [ ] `TestFullLifecycle` - Complete happy path (NoProject → ... → NoProject)
- [ ] `TestReviewFailLoop` - Rework loop (ReviewActive → ImplementationPlanning)
- [ ] `TestGuardsBlockInvalidTransitions` - Guards prevent invalid state changes
- [ ] `TestPromptGeneration` - All states generate correct prompts
- [ ] `TestOnAdvanceEventDetermination` - Event determiners work correctly
- [ ] All tests pass: `go test ./cli/internal/projects/standard/... -v`
- [ ] Test coverage > 80%

### Configuration Quality
- [ ] File `standard.go` created with `init()` and `NewStandardProjectConfig()`
- [ ] 4 phases configured with correct options
- [ ] Initial state set to NoProject
- [ ] 9 transitions defined with correct guards
- [ ] 7 OnAdvance event determiners implemented
- [ ] 7 prompts registered
- [ ] Guards use helper functions from Task 050
- [ ] Prompts use functions from Task 040
- [ ] Metadata schemas from Task 030 embedded
- [ ] Configuration < 250 lines
- [ ] Package compiles: `go build ./cli/internal/projects/standard/...`
- [ ] Linter passes: `golangci-lint run ./cli/internal/projects/standard/...`
- [ ] Old package untouched

## Validation Commands

```bash
# Run all tests
go test ./cli/internal/projects/standard/... -v

# Check coverage
go test ./cli/internal/projects/standard/... -cover

# Run specific test
go test ./cli/internal/projects/standard/... -run TestFullLifecycle -v

# Build package
go build ./cli/internal/projects/standard/...

# Run linter
golangci-lint run ./cli/internal/projects/standard/...

# Verify registration
go test ./cli/internal/projects/standard/... -run TestRegistration -v

# Verify old package untouched
git diff cli/internal/project/standard/
```

## Dependencies

- Task 020 (states/events)
- Task 030 (metadata schemas)
- Task 040 (prompts)
- Task 050 (guards)

## Standards

### TDD Standards
- Integration tests drive configuration
- Test realistic scenarios (full lifecycle, rework loops, guard blocking)
- Tests are independent (can run in any order)
- Helper functions make tests readable

### Code Standards
- Configuration uses SDK builder fluently
- Guard closures are concise (call helpers)
- OnAdvance logic handles errors gracefully
- Comments explain non-obvious decisions
- Configuration is declarative and readable

### Quality Gates
- [ ] All tests pass
- [ ] Test coverage > 80%
- [ ] Linter passes
- [ ] Configuration code < 250 lines
- [ ] No code duplication
- [ ] Clear separation of concerns

## Notes

**This is the Integration Task**: Everything comes together here. If tests pass, the SDK-based standard project type works end-to-end.

**Guard Closure Pattern**: Guards capture project instance via closures, then call helper functions:
```go
project.WithGuard(func(p *state.Project) bool {
    return phaseOutputApproved(p, "planning", "task_list")
})
```

**OnAdvance Complexity**: Most states return single event (simple). ReviewActive examines metadata to determine pass/fail (complex conditional logic).

**Registration**: The `init()` function registers the configuration globally. Unit 5 will migrate CLI commands to use this registration.

**State Machine Building**: See `cli/internal/sdks/project/machine.go` (lines 42-108) for how BuildMachine() binds guards via closures.
