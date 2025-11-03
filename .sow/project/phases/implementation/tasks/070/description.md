# Task 070: Integration Test - Complete Project Type Configuration

# Task 070: Integration Test - Complete Project Type Configuration

## Objective

Create a comprehensive integration test that demonstrates a complete project type can be defined using the SDK, registered, loaded, advanced through states, and validated. This satisfies the acceptance criteria "Example project type can be fully configured using SDK."

## Context

**Design Reference:** See `.sow/knowledge/designs/project-sdk-implementation.md`:
- Section "Complete Usage Example" (lines 673-831) for full project type definition
- Section "Testing Strategy - Integration Tests" (lines 1498-1516) for test requirements

**Prerequisite:** All previous tasks completed (010-060) - complete SDK implementation

**What This Task Builds:**
Proof that the SDK works end-to-end: configure → register → load → advance → validate.

## Requirements

### Integration Test Scope

Create `cli/internal/sdks/project/integration_test.go` that demonstrates:

1. **Define** - Complete project type using builder API
2. **Register** - Register type in registry
3. **Create** - Create project state
4. **Build** - Build state machine for project
5. **Advance** - Progress through multiple states
6. **Validate** - Validate state at each step
7. **Guards** - Verify guards block/allow appropriately
8. **Actions** - Verify actions mutate state correctly

### Example Project Type

Create a simple but realistic 3-state workflow:
- **States:** `Idle`, `Working`, `Done`
- **Events:** `Start`, `Complete`
- **Phases:** `work` (with tasks and metadata)
- **Guards:** Check work phase has approved output before completion
- **Metadata:** Schema requires `complexity` field

### Integration Test Structure

```go
func TestCompleteProjectTypeWorkflow(t *testing.T) {
    // 1. DEFINE: Create project type using builder
    config := NewProjectTypeConfigBuilder("simple").
        WithPhase("work",
            WithOutputs("result"),
            WithTasks(),
            WithMetadataSchema(`{
                complexity: "low" | "medium" | "high"
            }`),
        ).
        SetInitialState(State("Idle")).
        AddTransition(
            State("Idle"),
            State("Working"),
            Event("Start"),
        ).
        AddTransition(
            State("Working"),
            State("Done"),
            Event("Complete"),
            WithGuard(func(p *state.Project) bool {
                return p.PhaseOutputApproved("work", "result")
            }),
        ).
        OnAdvance(State("Idle"), func(p *state.Project) (Event, error) {
            return Event("Start"), nil
        }).
        OnAdvance(State("Working"), func(p *state.Project) (Event, error) {
            return Event("Complete"), nil
        }).
        Build()

    // 2. REGISTER: Add to registry
    state.Register("simple", config)

    // 3. CREATE: Make project state
    project := createTestProject(t, "simple")

    // 4. BUILD: Create state machine
    project.machine = config.BuildMachine(project, State("Idle"))

    // 5. ADVANCE: Progress from Idle → Working
    err := project.Advance()
    if err != nil {
        t.Fatalf("Advance from Idle failed: %v", err)
    }
    if project.machine.State() != State("Working") {
        t.Errorf("expected state=Working, got %s", project.machine.State())
    }

    // 6. VALIDATE: Guard blocks without approved output
    err = project.Advance()
    if err == nil {
        t.Error("expected Advance to fail when guard blocks")
    }

    // 7. MODIFY: Approve output
    project.Phases["work"].Outputs[0].Approved = true

    // 8. ADVANCE: Now transition should work
    err = project.Advance()
    if err != nil {
        t.Fatalf("Advance to Done failed: %v", err)
    }
    if project.machine.State() != State("Done") {
        t.Errorf("expected state=Done, got %s", project.machine.State())
    }

    // 9. VALIDATE: Full project validation passes
    err = config.Validate(project)
    if err != nil {
        t.Errorf("validation failed: %v", err)
    }
}
```

### Additional Integration Tests

**Test: Metadata Validation Integration**
- Define project type with metadata schema
- Create project with valid metadata
- Validate passes
- Change metadata to invalid value
- Validate fails with clear error

**Test: Artifact Type Validation Integration**
- Define project type with allowed artifact types
- Create project with valid artifact types
- Validate passes
- Add artifact with invalid type
- Validate fails with clear error

**Test: OnEntry/OnExit Actions Integration**
- Define transitions with entry/exit actions
- Actions mutate project state
- Verify state changes persist after transition
- Multiple transitions execute actions correctly

**Test: Multiple Phase Workflow**
- Define project type with 3+ phases
- Each phase has different configs (some with tasks, some without)
- Progress through all phases
- Validate at each step

## Files to Create

1. `cli/internal/sdks/project/integration_test.go` - Complete workflow tests

### Helper Functions

Include helpers for test setup:

```go
// createTestProject creates a minimal project state for testing
func createTestProject(t *testing.T, typeName string) *state.Project {
    return &state.Project{
        ProjectState: schemas.ProjectState{
            Name:   "test-project",
            Type:   typeName,
            Branch: "test-branch",
            Phases: map[string]schemas.PhaseState{
                "work": {
                    Status:  "pending",
                    Enabled: true,
                    Outputs: []schemas.ArtifactState{
                        {Type: "result", Approved: false},
                    },
                    Metadata: map[string]interface{}{
                        "complexity": "medium",
                    },
                },
            },
            Statechart: schemas.StatechartState{
                Current_state: "Idle",
                Updated_at:    time.Now(),
            },
        },
    }
}
```

## Testing Requirements

**Integration Tests Must Cover:**
- [ ] Complete project type definition via builder
- [ ] Registration in registry
- [ ] State machine creation via BuildMachine()
- [ ] Guard binding (guards access project state)
- [ ] Guard blocking (transition prevented when guard fails)
- [ ] Guard allowing (transition succeeds when guard passes)
- [ ] Event determination via OnAdvance
- [ ] Advance() progressing through states
- [ ] OnEntry/OnExit actions executing and mutating state
- [ ] Metadata validation (pass and fail cases)
- [ ] Artifact type validation (pass and fail cases)
- [ ] Multi-state workflow (3+ states)
- [ ] Error cases (missing determiner, guard blocks, etc.)

**Test Organization:**
- One comprehensive test covering full workflow
- Separate focused tests for specific integration points
- Clear test names describing what's being integrated
- No mocking - test real integration between components

## Acceptance Criteria

- [ ] Test defines simple project type using builder API
- [ ] Test includes phase with metadata schema
- [ ] Test includes transitions with guards accessing project state
- [ ] Test includes OnAdvance configuration for each state
- [ ] Test registers project type in registry
- [ ] Test creates project state and builds machine
- [ ] Test advances through multiple states successfully
- [ ] Test verifies guards prevent invalid transitions
- [ ] Test verifies guards allow valid transitions
- [ ] Test verifies metadata validation works (pass and fail)
- [ ] Test verifies artifact type validation works (pass and fail)
- [ ] Test verifies actions can mutate project state
- [ ] Full workflow test completes successfully
- [ ] All integration test assertions pass
- [ ] Code compiles and runs without errors

## Dependencies

**Required:** ALL previous tasks (010-060) - complete SDK implementation

## Technical Notes

- This is an integration test, not a unit test - tests real interactions
- No mocking - use real implementations of all components
- Test should be comprehensive but remain readable
- Focus on happy path + critical error cases
- Helper functions keep test code clean
- Test isolation: Each test resets Registry at start

**Test Isolation Pattern:**
```go
func TestSomething(t *testing.T) {
    // Reset registry for isolation
    state.Registry = make(map[string]*state.ProjectTypeConfig)

    // Test code...
}
```

**Import Organization:**
```go
import (
    "testing"
    "time"

    "github.com/jmgilman/sow/cli/internal/sdks/project"
    "github.com/jmgilman/sow/cli/internal/sdks/project/state"
    "github.com/jmgilman/sow/cli/schemas"
)
```

**Success Criteria:**
This test proves that:
1. SDK API is complete and usable
2. All components integrate correctly
3. Project types can be fully configured declaratively
4. Guards, actions, determiners all work as designed
5. Validation catches configuration errors

If this test passes, the SDK is ready for use by project type implementations.

## Estimated Time

1.5 hours
