# Task Breakdown: Migrate Standard Project Type to SDK

## Overview

This work unit migrates the existing standard project type from `internal/project/standard/` to a new SDK-based implementation at `internal/projects/standard/`. The migration involves copying templates and prompt logic, adapting types to work with the SDK's `*state.Project`, defining states/events/transitions, embedding metadata schemas, and testing the complete lifecycle.

**Key Constraint**: The old `internal/project/standard/` package must remain completely untouched. Both implementations will coexist temporarily until Unit 5 migrates CLI commands.

---

## Tasks

### Task 1: Create Package Structure and Copy Templates
**Agent**: implementer
**Description**: Set up the new package directory structure and physically copy all template files from the old package to the new location.

**Acceptance Criteria**:
- [ ] Create directory `internal/projects/standard/`
- [ ] Create subdirectory `internal/projects/standard/cue/` for metadata schemas
- [ ] Create subdirectory `internal/projects/standard/templates/` for prompt templates
- [ ] Copy all 8 template files from `internal/project/standard/templates/*.md` to `internal/projects/standard/templates/`
  - `planning_active.md`
  - `implementation_planning.md`
  - `implementation_executing.md`
  - `review_active.md`
  - `finalize_documentation.md`
  - `finalize_checks.md`
  - `finalize_delete.md`
  - `no_project.md`
- [ ] Verify all copied templates are identical to originals (no modifications)
- [ ] Old `internal/project/standard/` package remains untouched

**Dependencies**: None

---

### Task 2: Define States and Events
**Agent**: implementer
**Description**: Create state and event constant definitions for the SDK-based implementation.

**Acceptance Criteria**:
- [ ] Create `internal/projects/standard/states.go` with all state constants:
  - `NoProject` (from SDK state package)
  - `PlanningActive`
  - `ImplementationPlanning`
  - `ImplementationExecuting`
  - `ReviewActive`
  - `FinalizeDocumentation`
  - `FinalizeChecks`
  - `FinalizeDelete`
- [ ] States use `state.State` type (SDK type) not `statechart.State` (old type)
- [ ] Create `internal/projects/standard/events.go` with all event constants:
  - `EventProjectInit`
  - `EventCompletePlanning`
  - `EventTasksApproved`
  - `EventAllTasksComplete`
  - `EventReviewPass`
  - `EventReviewFail`
  - `EventDocumentationDone`
  - `EventChecksDone`
  - `EventProjectDelete`
- [ ] Events use `state.Event` type (SDK type) not `statechart.Event` (old type)
- [ ] Constants include clear documentation comments

**Dependencies**: Task 1 (package structure exists)

**Note**: Use SDK types (`internal/sdks/state`) not old statechart types.

---

### Task 3: Create Metadata Schemas
**Agent**: implementer
**Description**: Define minimal CUE schemas for phase-specific metadata validation and create Go embeddings.

**Acceptance Criteria**:
- [ ] Create `internal/projects/standard/cue/implementation_metadata.cue`:
  ```cue
  package standard

  // Metadata for implementation phase
  {
      tasks_approved?: bool
  }
  ```
- [ ] Create `internal/projects/standard/cue/review_metadata.cue`:
  ```cue
  package standard

  // Metadata for review phase
  {
      iteration?: int & >=1
  }
  ```
- [ ] Create `internal/projects/standard/cue/finalize_metadata.cue`:
  ```cue
  package standard

  // Metadata for finalize phase
  {
      project_deleted?: bool
      pr_url?: string
  }
  ```
- [ ] Create `internal/projects/standard/metadata.go` with embedded schemas:
  ```go
  package standard

  import _ "embed"

  //go:embed cue/implementation_metadata.cue
  var implementationMetadataSchema string

  //go:embed cue/review_metadata.cue
  var reviewMetadataSchema string

  //go:embed cue/finalize_metadata.cue
  var finalizeMetadataSchema string
  ```
- [ ] Schemas compile and embed correctly

**Dependencies**: Task 1 (cue/ directory exists)

**Note**: Keep schemas minimal - only the fields currently used by guards and prompts.

---

### Task 4: Create Prompt Functions
**Agent**: implementer
**Description**: Extract and refactor the prompt generation logic into simple functions matching the SDK's `PromptGenerator` signature (`func(*state.Project) string`).

**Acceptance Criteria**:
- [ ] Create `internal/projects/standard/prompts.go` with prompt functions
- [ ] Each function signature: `func(p *state.Project) string` (matching SDK `PromptGenerator`)
- [ ] Create 7 prompt functions (no NoProject prompt needed):
  - `generatePlanningPrompt(p *state.Project) string`
  - `generateImplementationPlanningPrompt(p *state.Project) string`
  - `generateImplementationExecutingPrompt(p *state.Project) string`
  - `generateReviewPrompt(p *state.Project) string`
  - `generateFinalizeDocumentationPrompt(p *state.Project) string`
  - `generateFinalizeChecksPrompt(p *state.Project) string`
  - `generateFinalizeDeletePrompt(p *state.Project) string`
- [ ] Use embedded templates from Task 1 via local registry:
  ```go
  //go:embed templates/*.md
  var templatesFS embed.FS

  var standardRegistry *prompts.Registry[StatePromptID]
  // Initialize in init() like old implementation
  ```
- [ ] Preserve dynamic logic from old implementation:
  - Project header (name, branch, description)
  - Git status (via `p.ctx.Git()` if available)
  - Task summaries (iterate `phase.Tasks`)
  - Iteration tracking (from review metadata)
  - Recent commits (conditional on task completion)
  - Artifact status displays
- [ ] Adapt field access to SDK types:
  - `phase, _ := p.Phases.Get("planning")` (collection pattern)
  - `phase.Outputs` (direct slice access)
  - `phase.Metadata["iteration"]` (metadata map access)
  - `phase.Tasks` (direct slice access)
- [ ] Preserve helper functions:
  - `findPreviousReviewArtifact(p *state.Project, iteration int) *Artifact`
  - `extractReviewAssessment(artifact *Artifact) string`
- [ ] Remove all references to old `StandardPromptGenerator` struct/interface
- [ ] Remove all references to old `statechart.PromptComponents`
- [ ] Access git/GitHub via `p.ctx` (from `*state.Project`) if needed

**Dependencies**: Task 1 (templates copied), Task 2 (states defined)

**Note**: This simplifies the prompt system significantly - no generator struct, no interfaces, just plain functions that match the SDK signature. The functions build prompts using string concatenation, template rendering, and direct access to project state.

---

### Task 5: Implement Guard Functions (TDD)
**Agent**: implementer
**Description**: Create guard functions that work with SDK types and check transition conditions using Test-Driven Development.

**TDD Requirements**:
- [ ] Write tests FIRST in `internal/projects/standard/guards_test.go`
- [ ] Implement guards to make tests pass
- [ ] Refactor while keeping tests green

**Acceptance Criteria**:
- [ ] Create `internal/projects/standard/guards_test.go` with tests for:
  - `phaseOutputApproved()` returns true when output exists and approved
  - `phaseOutputApproved()` returns false when output not approved
  - `phaseMetadataBool()` returns correct boolean value
  - `phaseMetadataBool()` returns false when key missing
  - `allTasksComplete()` returns true when all tasks completed
  - `allTasksComplete()` returns false when pending tasks exist
  - Guards handle missing phases gracefully (return false, not panic)
- [ ] Create `internal/projects/standard/guards.go` with guard helper functions:
  - `phaseOutputApproved(p *state.Project, phaseName, outputType string) bool` - Check if specific output type is approved
  - `phaseMetadataBool(p *state.Project, phaseName, key string) bool` - Get boolean from phase metadata
  - `allTasksComplete(p *state.Project) bool` - Check if all implementation tasks completed
  - `latestReviewApproved(p *state.Project) bool` - Check if latest review is approved
  - `projectDeleted(p *state.Project) bool` - Check project_deleted flag in finalize metadata
- [ ] All guards use SDK collection pattern:
  ```go
  phase, err := p.Phases.Get(phaseName)
  if err != nil {
      return false
  }
  ```
- [ ] Guards handle missing phases/artifacts gracefully (return false, not panic)
- [ ] Guards check metadata using `phase.Metadata[key]` with type assertions
- [ ] Guards match logic from old implementation but use new types
- [ ] All tests pass

**Dependencies**: Task 2 (states/events defined)

**Note**: These are helper functions, not the guard closures themselves. The closures are defined inline in `standard.go` (Task 6).

---

### Task 6: Implement SDK Configuration (TDD)
**Agent**: implementer
**Description**: Create the main project type configuration using the SDK builder API with Test-Driven Development.

**TDD Requirements**:
- [ ] Write integration tests FIRST in `internal/projects/standard/lifecycle_test.go`
- [ ] Implement configuration to make tests pass
- [ ] Refactor while keeping tests green

**Acceptance Criteria**:
- [ ] Create `internal/projects/standard/lifecycle_test.go` with tests for:
  - Full successful lifecycle (NoProject → PlanningActive → ... → NoProject)
  - Review fail loop (ReviewActive → ImplementationPlanning rework)
  - Guard blocking (cannot advance without meeting conditions)
  - Event determination (OnAdvance returns correct events)
  - Prompt generation (each state generates non-empty, correct prompts)
  - Metadata validation (schemas validate/reject correctly)
- [ ] Create `internal/projects/standard/standard.go` with:
  - `init()` function that calls `project.Register("standard", NewStandardProjectConfig())`
  - `NewStandardProjectConfig()` that returns configured `*project.ProjectTypeConfig`
- [ ] Configure all 4 phases using `WithPhase()`:
  - Planning: `WithStartState(PlanningActive)`, `WithEndState(PlanningActive)`, `WithInputs("context")`, `WithOutputs("task_list")`
  - Implementation: `WithStartState(ImplementationPlanning)`, `WithEndState(ImplementationExecuting)`, `WithTasks()`, `WithMetadataSchema(implementationMetadataSchema)`
  - Review: `WithStartState(ReviewActive)`, `WithEndState(ReviewActive)`, `WithOutputs("review")`, `WithMetadataSchema(reviewMetadataSchema)`
  - Finalize: `WithStartState(FinalizeDocumentation)`, `WithEndState(FinalizeDelete)`, `WithMetadataSchema(finalizeMetadataSchema)`
- [ ] Set initial state: `SetInitialState(NoProject)`
- [ ] Add all 10 transitions with guards:
  - NoProject → PlanningActive (EventProjectInit, no guard)
  - PlanningActive → ImplementationPlanning (EventCompletePlanning, guard: task_list approved)
  - ImplementationPlanning → ImplementationExecuting (EventTasksApproved, guard: tasks_approved metadata)
  - ImplementationExecuting → ReviewActive (EventAllTasksComplete, guard: all tasks complete)
  - ReviewActive → FinalizeDocumentation (EventReviewPass, guard: review approved)
  - ReviewActive → ImplementationPlanning (EventReviewFail, guard: review approved with fail assessment)
  - FinalizeDocumentation → FinalizeChecks (EventDocumentationDone, always true)
  - FinalizeChecks → FinalizeDelete (EventChecksDone, always true)
  - FinalizeDelete → NoProject (EventProjectDelete, guard: project_deleted metadata)
- [ ] Configure OnAdvance event determiners for each state (except NoProject):
  - Simple states return single event (PlanningActive → EventCompletePlanning)
  - ReviewActive examines review metadata to determine EventReviewPass vs EventReviewFail
- [ ] Configure prompts using `WithPrompt()` for all states:
  - Call appropriate prompt generator methods from Task 4
- [ ] All guards use helper functions from Task 5
- [ ] Configuration < 250 lines total
- [ ] All integration tests pass

**Dependencies**: Task 2 (states/events), Task 3 (metadata schemas), Task 4 (prompts), Task 5 (guards)

**Note**: This is the core integration task that wires everything together using the SDK builder. Write integration tests first to drive the configuration implementation.

---

### Task 7: Verify Old Implementation Untouched
**Agent**: implementer
**Description**: Final verification that the migration didn't affect the existing implementation.

**Acceptance Criteria**:
- [ ] Run `git diff internal/project/standard/` shows no changes
- [ ] Existing tests in `internal/project/standard/` still pass
- [ ] Old package imports unchanged
- [ ] No new dependencies added to old package
- [ ] Both implementations can coexist (no naming conflicts)

**Dependencies**: Task 6 (SDK configuration complete)

**Note**: This is a safety check to ensure we haven't broken anything during the migration.

---

## Success Metrics

**Functional**:
- [ ] All states, events, and transitions defined correctly
- [ ] All guards implement correct logic
- [ ] All prompts generate correctly with dynamic content
- [ ] Metadata schemas validate correctly
- [ ] Full lifecycle works end-to-end (PlanningActive → NoProject)
- [ ] Review fail loop works (ReviewActive → ImplementationPlanning)

**Code Quality**:
- [ ] Configuration code < 250 lines (excluding tests)
- [ ] Comprehensive test coverage written during implementation (TDD)
- [ ] All tests pass (unit tests for guards/events/metadata, integration tests for lifecycle)
- [ ] No duplication between old and new implementations
- [ ] Clear separation of concerns (states, events, guards, prompts, config)

**Safety**:
- [ ] Old `internal/project/standard/` completely untouched
- [ ] Both implementations can coexist
- [ ] No regressions in existing tests

---

## Notes

**Type Adaptation Strategy**:
The main challenge is adapting from old types (`*schemas.ProjectState`, `statechart.State/Event`) to SDK types (`*state.Project`, `state.State/Event`). The SDK provides collection methods (`Phases.Get()`) and metadata access patterns that differ from the old direct field access.

**Prompt System Simplification**:
The SDK uses a simple signature for prompts: `func(*state.Project) string`. We extract the existing prompt generation logic (dynamic components like git status, task summaries, iteration tracking) into plain functions matching this signature, eliminating the old `StandardPromptGenerator` struct and `PromptComponents` dependencies. Template rendering remains via the local registry pattern.

**Guard Patterns**:
Guards in the SDK are closures that capture the project instance. Helper functions in `guards.go` encapsulate the checking logic, then `standard.go` creates closures that call these helpers.

**Testing Strategy (TDD)**:
- **Task 5**: Write unit tests FIRST for guard functions, then implement guards
- **Task 6**: Write integration tests FIRST for lifecycle/transitions/prompts, then implement configuration
- Unit tests focus on individual guard/event functions
- Integration tests exercise state machine transitions, full lifecycle, prompts, and metadata validation
- Follow red-green-refactor cycle throughout implementation
- No need to test old implementation (already has tests)

**Migration Path**:
This unit creates the new implementation in parallel. Unit 5 will migrate CLI commands to use the new SDK types. Unit 6 will delete the old `internal/project/` package.
