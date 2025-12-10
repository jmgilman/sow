# Task 090: Update Consumers and Remove Old SDKs

## Context

This task is part of the `libs/project` module consolidation effort. With the new `libs/project` module complete (Tasks 010-080), this task updates all consumers to use the new module and removes the old SDK packages.

This is a large mechanical task involving:
1. Updating import paths throughout the codebase
2. Updating type references from old to new packages
3. Updating function calls to use new APIs
4. Removing deprecated SDK packages
5. Updating go.mod to include new dependency

## Requirements

### 1. Update cli/go.mod

Add the new libs/project module as a dependency:

```go
require (
    // existing deps...
    github.com/jmgilman/sow/libs/project v0.0.0
)

replace (
    // existing replaces...
    github.com/jmgilman/sow/libs/project => ../libs/project
)
```

### 2. Update Project Type Packages

Update each project type to use the new module:

**cli/internal/projects/standard/**
- `standard.go` - Update imports and type references
- `guards.go` - Update guard function signatures
- `prompts.go` - Update prompt function signatures
- `states.go` - Update state constants to use `project.State`
- `events.go` - Update event constants to use `project.Event`

**cli/internal/projects/exploration/**
- Same updates as standard

**cli/internal/projects/design/**
- Same updates as standard

**cli/internal/projects/breakdown/**
- Same updates as standard

### 3. Update Import Mappings

Replace old imports with new ones:

| Old Import | New Import |
|------------|------------|
| `cli/internal/sdks/state` | `libs/project` |
| `cli/internal/sdks/project` | `libs/project` |
| `cli/internal/sdks/project/state` | `libs/project/state` |
| `sdkstate "cli/internal/sdks/state"` | (remove alias, use `project` directly) |

### 4. Update Type References

Replace old types with new ones:

| Old Type | New Type |
|----------|----------|
| `sdkstate.State` | `project.State` |
| `sdkstate.Event` | `project.Event` |
| `project.ProjectTypeConfig` | `project.ProjectTypeConfig` |
| `project.ProjectTypeConfigBuilder` | `project.ProjectTypeConfigBuilder` |
| `state.Project` | `state.Project` |
| `state.Backend` | `state.Backend` |

### 5. Update Function Calls

Update loader function calls to use Backend:

**Old pattern:**
```go
proj, err := state.Load(ctx)  // ctx is sow.Context
```

**New pattern:**
```go
backend := state.NewYAMLBackend(fs)
proj, err := state.Load(context.Background(), backend)
// Or use convenience function:
proj, err := state.LoadFromFS(context.Background(), fs)
```

### 6. Update CLI Commands

Update all CLI commands that use project state:

- `cli/cmd/advance.go`
- `cli/cmd/phase.go`
- `cli/cmd/input.go`
- `cli/cmd/output.go`
- `cli/cmd/task.go`
- `cli/cmd/task_input.go`
- `cli/cmd/task_output.go`
- `cli/cmd/helpers.go`
- `cli/cmd/agent/spawn.go`
- `cli/cmd/agent/resume.go`
- `cli/cmd/project/shared.go`
- `cli/cmd/project/status.go`
- `cli/cmd/project/set.go`
- `cli/cmd/project/wizard_state.go`
- `cli/cmd/project/wizard_helpers.go`

### 7. Update Internal Packages

Update internal utility packages:

- `cli/internal/cmdutil/artifacts.go`
- `cli/internal/cmdutil/fieldpath.go`
- `cli/internal/templates/renderer.go`

### 8. Handle sow.Context Changes

The key API change is that `state.Load` no longer takes `sow.Context`. Update call sites:

**Old:**
```go
func doSomething(ctx *sow.Context) error {
    proj, err := state.Load(ctx)
    // ...
}
```

**New:**
```go
func doSomething(sowCtx *sow.Context) error {
    backend := state.NewYAMLBackend(sowCtx.FS())
    proj, err := state.Load(context.Background(), backend)
    // ...
}
```

### 9. Remove Old SDK Packages

After all consumers are updated and tests pass, delete:

- `cli/internal/sdks/state/` (entire directory)
- `cli/internal/sdks/project/` (entire directory)

### 10. Update Tests

Update all test files to use new imports and APIs:

- All `*_test.go` files in updated packages
- Test fixtures that reference old types
- Mock implementations

## Acceptance Criteria

1. [ ] `cli/go.mod` includes libs/project dependency
2. [ ] All project types updated to use new module
3. [ ] All CLI commands updated to use new APIs
4. [ ] All internal packages updated
5. [ ] sow.Context usage replaced with Backend interface
6. [ ] Old SDK packages deleted
7. [ ] All unit tests pass
8. [ ] All integration tests pass
9. [ ] `go build ./...` succeeds from cli directory
10. [ ] `golangci-lint run ./...` passes with no issues in both libs/project and cli
11. [ ] `go test -race ./...` passes with no failures in both libs/project and cli
12. [ ] No references to old SDK packages remain (verified via grep)
13. [ ] Code adheres to `.standards/STYLE.md` (file organization, naming, error handling)
14. [ ] Tests adhere to `.standards/TESTING.md` (table-driven, testify assertions, behavioral coverage)

### Test Requirements

This task primarily updates existing tests rather than writing new ones:

- Update test imports
- Update type references in tests
- Update mock implementations if needed
- Ensure all existing tests pass after migration
- Run full test suite: `go test ./...`

## Technical Details

### Import Update Script

Consider using `goimports` or a script to bulk-update imports:

```bash
# Find files with old imports
grep -r "sdks/state" cli/ --include="*.go"
grep -r "sdks/project" cli/ --include="*.go"
```

### Incremental Migration Strategy

To reduce risk, migrate in stages:

1. **Stage 1**: Update project types (standard, exploration, design, breakdown)
2. **Stage 2**: Update CLI commands
3. **Stage 3**: Update internal packages
4. **Stage 4**: Update tests
5. **Stage 5**: Remove old packages

Run tests after each stage to catch issues early.

### API Differences

Key API differences to watch for:

1. **Load/Create signatures changed**:
   - Old: `state.Load(ctx *sow.Context)`
   - New: `state.Load(ctx context.Context, backend Backend)`

2. **Save is now standalone**:
   - Old: `project.Save()` (method)
   - New: `state.Save(ctx, project)` (function)

3. **Registry location changed**:
   - Old: `state.Register(name, config)`
   - New: `project.Register(name, config)`

4. **Type locations changed**:
   - State/Event: `project` package
   - Project wrapper: `project/state` package

### Handling Breaking Changes

Some changes may require interface updates in `sow.Context`:

```go
// Potentially add helper method to sow.Context
func (c *Context) ProjectBackend() state.Backend {
    return state.NewYAMLBackend(c.fs)
}
```

## Relevant Inputs

- All files in `cli/internal/sdks/state/` - Files to delete
- All files in `cli/internal/sdks/project/` - Files to delete
- `cli/internal/projects/*/` - Project type packages to update
- `cli/cmd/` - CLI commands to update
- `cli/internal/cmdutil/` - Utility packages to update
- Files listed in grep results above - All consumer files
- `.standards/STYLE.md` - Code style requirements

## Examples

### Updating Project Type Registration

**Old (cli/internal/projects/standard/standard.go):**
```go
import (
    "github.com/jmgilman/sow/cli/internal/sdks/project"
    "github.com/jmgilman/sow/cli/internal/sdks/project/state"
    sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
)

func init() {
    state.Register("standard", BuildConfig())
}

func BuildConfig() *project.ProjectTypeConfig {
    return project.NewProjectTypeConfigBuilder("standard").
        SetInitialState(sdkstate.State(PlanningActive)).
        // ...
        Build()
}
```

**New:**
```go
import (
    "github.com/jmgilman/sow/libs/project"
    "github.com/jmgilman/sow/libs/project/state"
)

func init() {
    project.Register("standard", BuildConfig())
}

func BuildConfig() *project.ProjectTypeConfig {
    return project.NewProjectTypeConfigBuilder("standard").
        SetInitialState(project.State(PlanningActive)).
        // ...
        Build()
}
```

### Updating CLI Command

**Old (cli/cmd/advance.go):**
```go
import (
    "github.com/jmgilman/sow/cli/internal/sdks/project/state"
)

func runAdvance(cmd *cobra.Command, args []string) error {
    ctx := getContext(cmd)
    proj, err := state.Load(ctx)
    // ...
}
```

**New:**
```go
import (
    "context"
    "github.com/jmgilman/sow/libs/project/state"
)

func runAdvance(cmd *cobra.Command, args []string) error {
    sowCtx := getContext(cmd)
    backend := state.NewYAMLBackend(sowCtx.FS())
    proj, err := state.Load(context.Background(), backend)
    // ...
}
```

### Updating Guard Functions

**Old:**
```go
func allTasksComplete(p *state.Project) bool {
    // ...
}
```

**New:**
```go
import "github.com/jmgilman/sow/libs/project/state"

func allTasksComplete(p *state.Project) bool {
    // Same implementation, different import
}
```

## Dependencies

- Tasks 010-080: libs/project module must be complete
- All tests for libs/project must pass

## Constraints

- Do NOT change behavior - this is a pure migration
- Do NOT refactor consumers - only update imports and types
- Run tests after each batch of changes
- Keep commits small and focused (one package at a time)
- Ensure backward compatibility with existing project state files
- Old project state.yaml files must still load correctly
