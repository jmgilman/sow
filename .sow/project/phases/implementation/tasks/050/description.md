# Task 050: Implement Guard Functions (TDD)

# Task 050: Implement Guard Functions (TDD)

## Overview

Create guard helper functions that check transition conditions using Test-Driven Development. Guards determine whether state transitions should be allowed based on project state.

**TDD Requirement**: Write tests FIRST, then implement guards to make tests pass, then refactor.

## Context

**Design Reference**: `.sow/knowledge/designs/project-sdk-implementation.md` (lines 738-760) shows guard usage in transitions

**SDK Guard Pattern**: Guards are closures that capture project instance:
```go
WithGuard(func(p *state.Project) bool {
    return phaseOutputApproved(p, "planning", "task_list")
})
```

**Old Implementation Reference**: `cli/internal/project/standard/guards.go` contains logic to adapt (uses different types)

## TDD Workflow

### Red Phase: Write Failing Tests

Create `cli/internal/projects/standard/guards_test.go` with comprehensive test cases covering all guard functions and edge cases.

### Green Phase: Implement Guards

Create `cli/internal/projects/standard/guards.go` with minimal implementation to make tests pass.

### Refactor Phase: Clean Up

Improve implementation while keeping tests green.

## Requirements

### Test File Structure

Create `cli/internal/projects/standard/guards_test.go`:

```go
package standard

import (
	"testing"
	"time"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	"github.com/jmgilman/sow/cli/schemas/project"
)

func TestPhaseOutputApproved(t *testing.T) {
	tests := []struct {
		name       string
		project    *state.Project
		phaseName  string
		outputType string
		want       bool
	}{
		{
			name: "returns true when output exists and approved",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"planning": {
							Outputs: []project.ArtifactState{
								{
									Type:     "task_list",
									Path:     "tasks.md",
									Approved: true,
								},
							},
						},
					},
				},
			},
			phaseName:  "planning",
			outputType: "task_list",
			want:       true,
		},
		{
			name: "returns false when output not approved",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"planning": {
							Outputs: []project.ArtifactState{
								{
									Type:     "task_list",
									Path:     "tasks.md",
									Approved: false,
								},
							},
						},
					},
				},
			},
			phaseName:  "planning",
			outputType: "task_list",
			want:       false,
		},
		{
			name: "returns false when phase missing",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{},
				},
			},
			phaseName:  "planning",
			outputType: "task_list",
			want:       false,
		},
		{
			name: "returns false when output type not found",
			project: &state.Project{
				ProjectState: project.ProjectState{
					Phases: map[string]project.PhaseState{
						"planning": {
							Outputs: []project.ArtifactState{
								{
									Type:     "other",
									Path:     "other.md",
									Approved: true,
								},
							},
						},
					},
				},
			},
			phaseName:  "planning",
			outputType: "task_list",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := phaseOutputApproved(tt.project, tt.phaseName, tt.outputType); got != tt.want {
				t.Errorf("phaseOutputApproved() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Add similar comprehensive tests for:
// - TestPhaseMetadataBool
// - TestAllTasksComplete
// - TestLatestReviewApproved
// - TestProjectDeleted
```

**Test Coverage Requirements**:
1. **phaseOutputApproved**: approved output, not approved, missing phase, wrong type
2. **phaseMetadataBool**: true value, false value, missing key, missing phase, wrong type
3. **allTasksComplete**: all completed, mix of completed/pending, abandoned tasks, no tasks
4. **latestReviewApproved**: latest approved, latest not approved, no reviews, missing phase
5. **projectDeleted**: true value, false value, missing metadata, missing phase

### Implementation File Structure

Create `cli/internal/projects/standard/guards.go`:

```go
package standard

import (
	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
)

// Guard helper functions check transition conditions.
// These are called by guard closures defined in standard.go.

// phaseOutputApproved checks if a specific output artifact type is approved.
func phaseOutputApproved(p *state.Project, phaseName, outputType string) bool {
	phase, exists := p.Phases[phaseName]
	if !exists {
		return false
	}

	for _, output := range phase.Outputs {
		if output.Type == outputType && output.Approved {
			return true
		}
	}
	return false
}

// phaseMetadataBool gets a boolean value from phase metadata.
// Returns false if key missing, wrong type, or phase missing.
func phaseMetadataBool(p *state.Project, phaseName, key string) bool {
	phase, exists := p.Phases[phaseName]
	if !exists {
		return false
	}

	if phase.Metadata == nil {
		return false
	}

	val, ok := phase.Metadata[key]
	if !ok {
		return false
	}

	boolVal, ok := val.(bool)
	if !ok {
		return false
	}

	return boolVal
}

// allTasksComplete checks if all implementation tasks are completed or abandoned.
func allTasksComplete(p *state.Project) bool {
	phase, exists := p.Phases["implementation"]
	if !exists {
		return false
	}

	if len(phase.Tasks) == 0 {
		return false
	}

	for _, task := range phase.Tasks {
		if task.Status != "completed" && task.Status != "abandoned" {
			return false
		}
	}
	return true
}

// latestReviewApproved checks if the most recent review artifact is approved.
func latestReviewApproved(p *state.Project) bool {
	phase, exists := p.Phases["review"]
	if !exists {
		return false
	}

	// Find latest review output
	var latestReview *state.Artifact
	for i := len(phase.Outputs) - 1; i >= 0; i-- {
		if phase.Outputs[i].Type == "review" {
			latestReview = &phase.Outputs[i]
			break
		}
	}

	if latestReview == nil {
		return false
	}

	return latestReview.Approved
}

// projectDeleted checks if the project_deleted flag is set in finalize metadata.
func projectDeleted(p *state.Project) bool {
	return phaseMetadataBool(p, "finalize", "project_deleted")
}
```

## Acceptance Criteria

### TDD Process
- [ ] Tests written FIRST in `guards_test.go`
- [ ] Tests fail initially (red phase)
- [ ] Implementation written to make tests pass (green phase)
- [ ] Code refactored while keeping tests green

### Test Coverage
- [ ] `TestPhaseOutputApproved` with 4+ test cases
- [ ] `TestPhaseMetadataBool` with 5+ test cases
- [ ] `TestAllTasksComplete` with 4+ test cases
- [ ] `TestLatestReviewApproved` with 4+ test cases
- [ ] `TestProjectDeleted` with 3+ test cases
- [ ] All edge cases covered (missing phases, nil values, wrong types)
- [ ] Tests use table-driven approach
- [ ] All tests pass: `go test ./cli/internal/projects/standard/...`

### Implementation Quality
- [ ] Five guard functions implemented:
  - `phaseOutputApproved(p *state.Project, phaseName, outputType string) bool`
  - `phaseMetadataBool(p *state.Project, phaseName, key string) bool`
  - `allTasksComplete(p *state.Project) bool`
  - `latestReviewApproved(p *state.Project) bool`
  - `projectDeleted(p *state.Project) bool`
- [ ] Guards handle missing phases gracefully (return false, not panic)
- [ ] Guards handle nil metadata gracefully
- [ ] Guards use direct map access (not SDK collections - phase is already retrieved)
- [ ] Clear documentation comments for each function
- [ ] Package compiles: `go build ./cli/internal/projects/standard/...`
- [ ] Old package untouched

## Validation Commands

```bash
# Run tests (should fail initially, then pass after implementation)
go test ./cli/internal/projects/standard/ -v

# Check test coverage
go test ./cli/internal/projects/standard/ -cover

# Verify compilation
go build ./cli/internal/projects/standard/...

# Run linter
golangci-lint run ./cli/internal/projects/standard/...

# Verify old package untouched
git diff cli/internal/project/standard/
```

## Dependencies

- Task 020 (states/events defined)

## Standards

### TDD Standards
- **Red-Green-Refactor**: Follow cycle strictly
- **Test First**: Write tests before implementation
- **One Test at a Time**: Make one test pass before moving to next
- **Refactor Continuously**: Improve code while tests stay green

### Code Standards
- Use table-driven tests with descriptive names
- Test both happy path and edge cases
- Guards never panic (return false on errors)
- Clear error handling in tests
- Comprehensive documentation comments

### Quality Gates
- [ ] All tests pass
- [ ] Test coverage > 90%
- [ ] Linter passes with no warnings
- [ ] No code duplication
- [ ] Edge cases covered

## Notes

**Guard Helper Pattern**: These are helper functions, not the final guard closures. Task 6 will create closures that call these helpers:
```go
WithGuard(func(p *state.Project) bool {
    return phaseOutputApproved(p, "planning", "task_list")
})
```

**Type Access**: Guards use direct map/slice access since they work with already-retrieved phase objects. The SDK collections (`Phases.Get()`) are used in the guard closures, not in these helpers.

**Error Handling**: Guards return `false` for any error condition (missing phase, wrong type, etc.). This is safer than panicking and allows state machine to block invalid transitions.

**Old Implementation**: Reference `cli/internal/project/standard/guards.go` for logic but adapt types completely. Old uses `*schemas.ProjectState`, new uses `*state.Project`.
