# Task 020: Replace Project Command with New/Continue Subcommands (TDD)

# Task 020: Replace Project Command with New/Continue Subcommands (TDD)

## Overview

Delete the existing unified `sow project` command and replace it with explicit `sow project new` and `sow project continue` subcommands that use the SDK state layer.

## Context

**Current state**: `cli/cmd/project.go` contains a unified command that automatically creates or continues projects based on whether state exists.

**Goal**: Split into explicit subcommands while preserving all functionality (worktree management, issue linking, Claude Code launching).

**Key change**: Use SDK `state.Load(ctx)` and `project.Save()` instead of the old `internal/project` package.

## Design References

- **Command specs**: `.sow/knowledge/designs/command-hierarchy-design.md` lines 364-377, 453-511
- **SDK patterns**: `.sow/knowledge/designs/project-sdk-implementation.md` lines 643-672 (Load/Save API)
- **Existing implementation**: `cli/cmd/project.go` (review before deleting)

## Requirements

### Commands to Implement

#### 1. `sow project new`

**Syntax**: `sow project new --branch <branch> [--issue <number>] "<description>"`

**Behavior**:
- Creates worktree for branch (if doesn't exist)
- Initializes new project using SDK
- Links GitHub issue if provided
- Generates initial prompt
- Launches Claude Code

**Must preserve** from existing code:
- Worktree management logic
- GitHub issue linking
- Branch validation (no protected branches)
- Uncommitted changes check
- Initial prompt generation

#### 2. `sow project continue`

**Syntax**: `sow project continue [--branch <branch>]`

**Behavior**:
- Loads existing project using SDK
- Generates continue prompt with project status
- Launches Claude Code

**Must preserve** from existing code:
- Worktree navigation
- Project state reading
- Continue prompt generation
- Task status summarization

#### 3. `sow project set`

**Syntax**: `sow project set <field-path> <value>`

**Behavior**:
- Loads project state
- Uses field path parser (from Task 010)
- Sets field (direct or metadata)
- Saves state

**Examples**:
```bash
sow project set description "Updated description"
sow project set metadata.custom_field value
```

#### 4. `sow project delete`

**Syntax**: `sow project delete`

**Behavior**:
- Deletes `.sow/project/` directory
- Same as existing implementation

## TDD Approach

### Step 1: Write Integration Test First

Create `cli/testdata/script/unified_commands/project/project_lifecycle.txtar`:

```txtar
# Test: Project Lifecycle Commands
# Coverage: new, continue, set, delete

# Setup git repo
exec git init
exec git config user.email 'test@example.com'
exec git config user.name 'Test User'
exec git commit --allow-empty -m 'Initial commit'
exec git checkout -b feat/test-project
exec sow init

# Test: sow project new
exec sow project new --branch feat/test-project --no-launch "Test project description"
stderr '✓ Initialized project'
exists .sow/project/state.yaml
exec cat .sow/project/state.yaml
stdout 'name: test-project'
stdout 'description: Test project description'
stdout 'branch: feat/test-project'

# Test: sow project set (direct field)
exec sow project set description "Updated description"
exec cat .sow/project/state.yaml
stdout 'description: Updated description'

# Test: sow project set (metadata field)
exec sow project set metadata.custom_field custom_value
exec cat .sow/project/state.yaml
stdout 'custom_field: custom_value'

# Test: sow project delete
exec sow project delete
! exists .sow/project
```

### Step 2: Delete Old Command

Remove `cli/cmd/project.go` entirely.

### Step 3: Implement New Commands

Create command structure and implement using SDK.

### Step 4: Run Integration Test

Verify test passes.

## Implementation Details

### File Structure

```
cli/cmd/project/
├── project.go       # Root command with subcommands
├── new.go           # sow project new
├── continue.go      # sow project continue
├── set.go           # sow project set
└── delete.go        # sow project delete
```

### SDK Usage Pattern

All commands follow this pattern:

```go
// Load
ctx := cmdutil.GetContext(cmd.Context())
project, err := state.Load(ctx)
if err != nil {
    return err
}

// Mutate
project.Description = newDescription
// OR for metadata:
if project.Metadata == nil {
    project.Metadata = make(map[string]interface{})
}
project.Metadata["custom_field"] = value

// Save
return project.Save()
```

### Migrating Logic from Old project.go

**From `runProject()` function**:
- Worktree management → keep in `new.go` and `continue.go`
- Issue handling → keep in `new.go`
- Branch validation → keep in `new.go`
- State detection → split between `new` (create) and `continue` (load)

**From helper functions**:
- `initializeProject()` → use in `new.go`, but call SDK `state.Create()`
- `generateNewProjectPrompt()` → keep in `new.go`
- `generateContinuePrompt()` → keep in `continue.go`
- `launchClaudeCode()` → keep as shared helper

### Integration with Existing Code

**Keep using**:
- `sow.Context` - File system operations
- `sow.WorktreePath()` - Worktree path generation
- `sow.EnsureWorktree()` - Worktree creation
- `sow.GitHub` - Issue linking
- `sowexec.NewLocal("claude")` - Claude Code launching

**Replace**:
- `loader.Load()` → `state.Load()`
- `loader.Create()` → `state.Create()`
- `loader.Exists()` → Check if `.sow/project/state.yaml` exists
- `project.Machine().ProjectState()` → Direct access to project fields

## Files to Create

### `cli/cmd/project/project.go`

Root command that registers subcommands:

```go
func NewProjectCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "project",
        Short: "Manage projects",
    }

    cmd.AddCommand(newNewCmd())
    cmd.AddCommand(newContinueCmd())
    cmd.AddCommand(newSetCmd())
    cmd.AddCommand(newDeleteCmd())

    return cmd
}
```

### `cli/cmd/project/new.go`

Extract logic from current `runProject()` for the "create new" path.

### `cli/cmd/project/continue.go`

Extract logic from current `runProject()` for the "continue existing" path.

### `cli/cmd/project/set.go`

New command using field path parser:

```go
func newSetCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "set <field-path> <value>",
        Short: "Set project field",
        Args:  cobra.ExactArgs(2),
        RunE:  runSet,
    }
}

func runSet(cmd *cobra.Command, args []string) error {
    ctx := cmdutil.GetContext(cmd.Context())
    project, err := state.Load(ctx)
    if err != nil {
        return err
    }

    fieldPath := args[0]
    value := args[1]

    // Use field path parser from Task 010
    if err := cmdutil.SetField(project, fieldPath, value); err != nil {
        return err
    }

    return project.Save()
}
```

### `cli/cmd/project/delete.go`

Simple deletion command.

### `cli/testdata/script/unified_commands/project/project_lifecycle.txtar`

Integration test covering all subcommands.

## Files to Delete

- `cli/cmd/project.go` - Entire file

## Acceptance Criteria

- [ ] Integration test written first
- [ ] Old `project.go` deleted
- [ ] `sow project new` creates project using SDK
- [ ] `sow project continue` loads project using SDK
- [ ] `sow project set` modifies both direct and metadata fields
- [ ] `sow project delete` removes project
- [ ] All worktree/issue/branch logic preserved
- [ ] Integration test passes
- [ ] `--no-launch` flag works for testing

## Testing Strategy

**Integration test only** - No unit tests for command logic.

Test scenarios:
1. Create new project
2. Set direct field (description)
3. Set metadata field (metadata.custom_field)
4. Delete project
5. Error: set on non-existent project
6. Error: invalid field path

## Dependencies

- Task 010 (Field Path Parsing) - Required for `set` command

## References

- **Existing implementation**: `cli/cmd/project.go`
- **SDK state types**: `cli/internal/sdks/project/state/project.go`
- **Context utilities**: `cli/internal/sow/context.go`
- **Worktree management**: `cli/internal/sow/worktree.go`

## Migration Notes

**Key differences from old implementation**:
1. Explicit `new` vs `continue` - No automatic detection
2. SDK state types - Not domain types
3. Direct field mutation - Not methods on domain objects
4. Field path parser - For metadata routing

**Backward compatibility**:
- State file format unchanged (SDK uses same YAML structure)
- Worktree paths unchanged
- Issue linking unchanged
