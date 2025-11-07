# Task 020: Extract Shared Utilities from new.go and continue.go

## Context

This task is part of replacing the flag-based `sow project new` and `sow project continue` commands with an interactive wizard. The existing commands contain critical logic for project initialization, prompt generation, and Claude Code launching that must be preserved.

**Critical Migration Requirement**: Before we can delete `new.go` and `continue.go`, we must extract their reusable functions into a shared module. The wizard will use these same functions, just invoked through an interactive flow instead of command flags.

This task extracts four key functions that implement core functionality:
1. **Project initialization** - Creates project directories and state files
2. **New project prompt generation** - Builds 3-layer prompts for new projects
3. **Continue project prompt generation** - Builds 3-layer prompts for existing projects
4. **Claude Code launcher** - Executes the Claude CLI with proper context

## Requirements

Create a new file `cli/cmd/project/shared.go` containing extracted utility functions. These functions must work identically to their original implementations.

### Function 1: initializeProject

**Extract from**: `new.go` lines 148-196 (logic embedded in `runNew`)

**Signature**:
```go
func initializeProject(
    ctx *sow.Context,
    branch string,
    description string,
    issue *sow.Issue,
) (*state.Project, error)
```

**Responsibilities**:
1. Create `.sow/project` directory in worktree
2. Create `.sow/project/context` directory
3. If issue is provided:
   - Write issue body to `context/issue-{number}.md`
   - Create github_issue artifact with metadata
   - Attach artifact to implementation phase
4. Call `state.Create()` with initial inputs
5. Return created project

**Key details**:
- Issue file format: `# Issue #{number}: {title}\n\n**URL**: {url}\n**State**: {state}\n\n## Description\n\n{body}\n`
- Artifact type: `"github_issue"`
- Artifact path: `"context/issue-{number}.md"` (relative to .sow/project)
- Artifact auto-approved: `true`
- Initial inputs only for implementation phase

### Function 2: generateNewProjectPrompt

**Extract from**: `new.go` lines 359-395

**Signature**:
```go
func generateNewProjectPrompt(proj *state.Project, initialPrompt string) (string, error)
```

**Responsibilities**:
1. Render base orchestrator prompt from template
2. Get project type orchestrator prompt
3. Get initial state prompt
4. Append user's initial prompt if provided
5. Return combined 3-layer prompt

**3-Layer Structure**:
- Layer 1: Base orchestrator introduction (from template)
- Layer 2: Project type orchestrator prompt (from config)
- Layer 3: Initial state prompt (from state machine)
- Optional: User's initial request

**Separators**: Use `\n\n---\n\n` between layers

### Function 3: generateContinuePrompt

**Extract from**: `continue.go` lines 167-196

**Signature**:
```go
func generateContinuePrompt(proj *state.Project) (string, error)
```

**Responsibilities**:
1. Render base orchestrator prompt from template
2. Get project type orchestrator prompt
3. Get **current** state prompt (not initial)
4. Return combined 3-layer prompt

**Difference from new prompt**: Uses current state instead of initial state, no user prompt appended.

### Function 4: launchClaudeCode

**Extract from**: `new.go` lines 397-418

**Signature**:
```go
func launchClaudeCode(
    cmd *cobra.Command,
    ctx *sow.Context,
    prompt string,
    claudeFlags []string,
) error
```

**Responsibilities**:
1. Check if Claude CLI exists using `sowexec.NewLocal("claude")`
2. If not found, print error message and installation instructions
3. Build command args: prompt first, then claudeFlags
4. Execute Claude with:
   - Working directory = `ctx.RepoRoot()`
   - Stdin/Stdout/Stderr inherited from parent
   - Command context from cobra command
5. Return any execution error

**Error messages**:
- "Error: Claude Code CLI not found"
- "Install from: https://claude.com/download"

## Acceptance Criteria

### Code Quality
- [ ] File `cli/cmd/project/shared.go` created
- [ ] All four functions implemented with correct signatures
- [ ] All necessary imports included
- [ ] Functions have godoc comments explaining purpose and parameters
- [ ] No compilation errors

### Functional Correctness
- [ ] `initializeProject()` creates directories in correct locations
- [ ] `initializeProject()` writes issue files with correct format
- [ ] `initializeProject()` creates artifacts with correct metadata
- [ ] `generateNewProjectPrompt()` produces 3-layer prompt structure
- [ ] `generateNewProjectPrompt()` includes user prompt when provided
- [ ] `generateContinuePrompt()` produces 3-layer prompt structure
- [ ] `generateContinuePrompt()` uses current state, not initial state
- [ ] `launchClaudeCode()` checks for Claude CLI existence
- [ ] `launchClaudeCode()` builds args in correct order (prompt first)
- [ ] `launchClaudeCode()` inherits stdin/stdout/stderr

### No Breaking Changes
- [ ] Original `new.go` and `continue.go` remain unchanged in this task
- [ ] Original commands still work (they still use embedded logic)
- [ ] No existing functionality is broken

## Relevant Inputs

- `cli/cmd/project/new.go` - Source of initializeProject, generateNewProjectPrompt, launchClaudeCode logic
- `cli/cmd/project/continue.go` - Source of generateContinuePrompt logic
- `cli/internal/sow/context.go` - Context type definition
- `cli/internal/sow/github.go` - Issue type definition
- `cli/internal/sdks/project/state/project.go` - Project and state.Create() function
- `cli/internal/sdks/project/templates/templates.go` - Template rendering
- `cli/internal/prompts/fs.go` - Prompts filesystem
- `cli/internal/exec/executor.go` - sowexec.NewLocal() function
- `.sow/project/context/issue-68.md` - Reference for issue file format (section 11)

## Examples

### Example 1: initializeProject Usage

```go
// In wizard finalization:
proj, err := initializeProject(
    worktreeCtx,
    "feat/add-auth",
    "Add JWT authentication",
    issue, // or nil if no issue
)
if err != nil {
    return fmt.Errorf("failed to initialize project: %w", err)
}
```

### Example 2: generateNewProjectPrompt Usage

```go
// After project creation:
prompt, err := generateNewProjectPrompt(proj, "Start by reviewing existing auth code")
if err != nil {
    return fmt.Errorf("failed to generate prompt: %w", err)
}
// prompt contains 3-layer structure + user request
```

### Example 3: generateContinuePrompt Usage

```go
// When continuing existing project:
proj, err := state.Load(worktreeCtx)
if err != nil {
    return err
}

prompt, err := generateContinuePrompt(proj)
if err != nil {
    return fmt.Errorf("failed to generate prompt: %w", err)
}
// prompt contains 3-layer structure with current state
```

### Example 4: launchClaudeCode Usage

```go
// Final step of wizard:
err := launchClaudeCode(cmd, worktreeCtx, prompt, []string{"--model", "opus"})
if err != nil {
    return fmt.Errorf("failed to launch Claude: %w", err)
}
```

## Dependencies

- **Task 010**: Huh library must be added (though this task doesn't use it yet)

## Constraints

- **No behavior changes**: Functions must work identically to original implementations
- **No test changes**: Don't modify existing tests in this task
- **Preserve error messages**: Use exact same error messages as originals
- **Preserve formatting**: Issue files, prompts must have identical formatting
- **No refactoring**: Extract as-is, don't improve or optimize yet

## Testing Requirements

### Unit Tests

Create `cli/cmd/project/shared_test.go` with tests for each function:

**Test: initializeProject**
```go
func TestInitializeProject_CreatesDirectories(t *testing.T)
func TestInitializeProject_WithIssue_WritesIssueFile(t *testing.T)
func TestInitializeProject_WithIssue_CreatesArtifact(t *testing.T)
func TestInitializeProject_WithoutIssue_NoArtifacts(t *testing.T)
```

Test that:
- Directories are created
- Issue file has correct format
- Artifact has correct metadata fields
- state.Create() is called with correct inputs

**Test: generateNewProjectPrompt**
```go
func TestGenerateNewProjectPrompt_Has3Layers(t *testing.T)
func TestGenerateNewProjectPrompt_WithUserPrompt(t *testing.T)
func TestGenerateNewProjectPrompt_WithoutUserPrompt(t *testing.T)
```

Test that:
- Output contains base orchestrator prompt
- Output contains project type prompt
- Output contains initial state prompt
- Layers are separated by `\n\n---\n\n`
- User prompt is appended when provided

**Test: generateContinuePrompt**
```go
func TestGenerateContinuePrompt_Has3Layers(t *testing.T)
func TestGenerateContinuePrompt_UsesCurrentState(t *testing.T)
```

Test that:
- Output contains 3 layers
- Uses current state, not initial state

**Test: launchClaudeCode**
```go
func TestLaunchClaudeCode_ChecksClaudeExists(t *testing.T)
func TestLaunchClaudeCode_BuildsArgsCorrectly(t *testing.T)
```

Test that:
- Function checks for Claude CLI
- Args are built in correct order (prompt first)
- Working directory is set to ctx.RepoRoot()

**Note**: Use test helpers, mocks, or temporary directories as needed. Reference existing test patterns in the codebase.

### Integration Tests

Not required for this task - the existing `new.go` and `continue.go` commands serve as integration tests until they're replaced.

## Implementation Notes

### Extraction Strategy

1. **Copy, don't move**: Copy logic from original files, leave originals unchanged
2. **Keep context**: Include any helper variables or error messages
3. **Preserve imports**: Make sure all necessary imports are included
4. **Test extraction**: Write tests to verify extracted functions work correctly

### Original Code Locations

**initializeProject logic** (new.go:148-196):
- Creates projectDir and contextDir
- Writes issue file if issue != nil
- Creates issueArtifact with metadata
- Calls state.Create() with initialInputs

**generateNewProjectPrompt** (new.go:359-395):
- Renders base orchestrator from template
- Gets project type orchestrator prompt
- Gets initial state prompt
- Appends user prompt if provided

**generateContinuePrompt** (continue.go:167-196):
- Same structure as generateNewProjectPrompt
- Uses currentState instead of initialState
- No user prompt appending

**launchClaudeCode** (new.go:397-418):
- Uses sowexec.NewLocal("claude")
- Checks claude.Exists()
- Builds args, creates exec.Command
- Sets Dir, Stdin, Stdout, Stderr
- Runs command

### Why Extract First?

This extraction happens BEFORE building the wizard because:
1. **Safety**: Existing commands continue to work during development
2. **Testing**: We can test extracted functions independently
3. **Verification**: We can verify extracted logic is identical
4. **No rushing**: We take time to extract correctly before deleting originals

### Post-Extraction Plan

After this task is complete:
- Task 030+ will build the wizard using these shared functions
- Later tasks will update `new.go` and `continue.go` to use shared functions
- Final task will delete `new.go` and `continue.go` entirely

## Success Indicators

After completing this task:
1. All four functions exist in `shared.go` with correct signatures
2. All functions have test coverage
3. Tests pass and verify functions work correctly
4. Original `new.go` and `continue.go` remain unchanged and working
5. Foundation is ready for wizard to use these functions
6. Code is well-documented with godoc comments
