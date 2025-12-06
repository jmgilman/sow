# Task 030: Status Command Tests

## Context

Write tests for the `sow project status` command to verify output format and error handling.

## Goals

- Verify command output matches expected format
- Test error cases (no project, no .sow)
- Ensure task counting is accurate

## Requirements

### Test File

Create `cli/cmd/project/status_test.go`

### Test Cases

1. **TestStatusNoProject**
   - Setup: No `.sow/project/state.yaml`
   - Expected: Error "no active project"

2. **TestStatusBasicOutput**
   - Setup: Create minimal project state
   - Verify: Output contains project name, branch, type, state
   - Verify: Output contains phase section
   - Verify: Output contains task section

3. **TestStatusTaskCounts**
   - Setup: Project with mixed task statuses
   - Verify: Correct X/Y calculation per phase

4. **TestStatusNoTasks**
   - Setup: Project with phases but no tasks
   - Verify: Shows "0/0 tasks completed" or appropriate message

### Test Helpers

Use existing test patterns from `cli/cmd/project/`:
- `project_test.go` - Command testing patterns
- `shared_test.go` - Shared test utilities

### Mock Project State

Create test fixtures with known state:

```go
func createTestProject(t *testing.T, dir string, phases map[string][]project.TaskState) {
    // Create .sow/project/state.yaml with given phases and tasks
}
```

## Acceptance Criteria

- [ ] All test cases pass
- [ ] Tests cover happy path and error cases
- [ ] Tests verify output format
- [ ] Tests verify task counting logic
- [ ] Tests follow existing patterns in the codebase

## Technical Notes

### Capturing Output

To test command output:
```go
cmd := newStatusCmd()
var stdout bytes.Buffer
cmd.SetOut(&stdout)
cmd.SetErr(&stderr)
err := cmd.Execute()
// Assert on stdout.String()
```

### Test State Creation

Create a helper that writes valid YAML:
```go
stateYAML := `
name: test-project
type: standard
branch: feat/test
phases:
  implementation:
    status: in_progress
    tasks:
      - id: "010"
        name: Test task
        phase: implementation
        status: completed
        ...
statechart:
  current_state: ImplementationExecuting
  updated_at: 2024-01-01T00:00:00Z
`
```

## Relevant Inputs

- `cli/cmd/project/project_test.go` - Existing test patterns
- `cli/cmd/project/shared_test.go` - Test utilities
- `cli/cmd/project/wizard_test.go` - More test examples
- `cli/internal/sdks/project/state/loader_test.go` - State loading tests
- `cli/schemas/project/test_helper.cue` - Test helper schemas
