# Task 050: Implement Project Finalization and Claude Launch

## Context

This task implements the finalization flow that creates the project, initializes the worktree, generates the 3-layer prompt, and launches Claude Code. This is the culmination of the branch name creation path - where all the user's selections are turned into a real, working sow project.

This is part of Work Unit 002 (Project Creation Workflow - Branch Name Path). The wizard foundation from Work Unit 001 provides the state machine and `finalize()` method stub. The shared utilities (`initializeProject`, `generateNewProjectPrompt`, `launchClaudeCode`) already exist in `shared.go` and just need to be integrated.

**Project Goal**: Build an interactive wizard for creating new sow projects via branch name selection, including type selection, name entry with real-time preview, prompt entry with external editor support, and project initialization in git worktrees.

**Why This Task**: This is where the wizard's work pays off - creating a properly initialized project in a git worktree and launching Claude with a comprehensive prompt. The conditional uncommitted changes check is critical for avoiding git worktree errors.

## Requirements

### Handler Implementation

Create the `finalize()` function in `cli/cmd/project/wizard_state.go` to replace the current stub implementation.

**Function Location**: Replace the stub at lines 173-178 in `wizard_state.go`

**Called When**: `w.state == StateComplete` in the wizard's `Run()` loop

**Steps** (in order):

1. **Extract wizard choices**
   - Project type from `w.choices["type"]`
   - Project name from `w.choices["name"]`
   - Branch name from `w.choices["branch"]`
   - Initial prompt from `w.choices["prompt"]` (may be empty)

2. **Conditional uncommitted changes check**
   - Get current branch using `w.ctx.Git().CurrentBranch()`
   - If current branch == target branch, check for uncommitted changes
   - If uncommitted changes exist, return detailed error with fix instructions
   - If current branch != target branch, skip check (worktree creation will switch branches)

3. **Ensure worktree exists**
   - Compute worktree path using `sow.WorktreePath(w.ctx.RepoRoot(), branch)`
   - Create/verify worktree using `sow.EnsureWorktree(w.ctx, worktreePath, branch)`
   - Handle errors (likely git issues)

4. **Initialize project in worktree**
   - Create worktree context using `sow.NewContext(worktreePath)`
   - Call `initializeProject(worktreeCtx, branch, name, nil)` (no issue for branch path)
   - Handle errors (likely filesystem issues)

5. **Generate 3-layer prompt**
   - Call `generateNewProjectPrompt(project, initialPrompt)`
   - Prompt includes: Base Orchestrator + Project Type Orchestrator + Initial State + User Prompt
   - Handle errors (template rendering issues)

6. **Display success message**
   - Show project name and branch
   - Indicate Claude is launching

7. **Launch Claude Code**
   - Call `launchClaudeCode(cmd, worktreeCtx, prompt, w.claudeFlags)`
   - This transfers control to Claude (blocking call)
   - Return any errors

### Conditional Uncommitted Changes Check

**Logic**:
```go
currentBranch, err := w.ctx.Git().CurrentBranch()
if err != nil {
    return fmt.Errorf("failed to get current branch: %w", err)
}

// Only check if we're on the branch we're trying to create a worktree for
if currentBranch == selectedBranch {
    if err := sow.CheckUncommittedChanges(w.ctx); err != nil {
        return fmt.Errorf("repository has uncommitted changes\n\n"+
            "You are currently on branch '%s'.\n"+
            "Creating a worktree requires switching to a different branch first.\n\n"+
            "To fix:\n"+
            "  Commit: git add . && git commit -m \"message\"\n"+
            "  Or stash: git stash", currentBranch)
    }
}
```

**Why Conditional**: If you're on `feat/auth` and try to create a worktree for `explore/research`, git can create it without issues. But if you're on `feat/auth` and try to create a worktree for `feat/auth`, git needs to switch branches first, which requires a clean working tree.

### Success Message

Display before launching Claude:

```go
fmt.Fprintf(os.Stdout, "✓ Initialized project '%s' on branch %s\n", name, branch)
fmt.Fprintf(os.Stdout, "✓ Launching Claude in worktree...\n")
```

### Error Handling

**Git Errors** (worktree creation):
- Provide clear error message
- Include the git error details
- Return error (wizard will display it)

**Filesystem Errors** (project initialization):
- Provide clear error message
- Include the filesystem error details
- Return error

**Template Errors** (prompt generation):
- Provide clear error message
- Include template error details
- Return error

**Claude Launch Errors**:
- If Claude not found, `launchClaudeCode` will display error and return
- Propagate error up to wizard

### Integration Points

**Upstream**: Called from `Run()` when `w.state == StateComplete`, triggered by `handlePromptEntry()` after successful prompt entry

**Downstream**: Launches Claude Code CLI with project context, transferring control to the orchestrator agent

## Acceptance Criteria

### Functional Requirements

1. **Uncommitted Changes Check Works**
   - Only runs when current branch == target branch
   - Clear error message if check fails
   - Error includes fix instructions (commit or stash)
   - No false positives from untracked files (handled by CheckUncommittedChanges)

2. **Worktree Created Correctly**
   - Worktree created at `.sow/worktrees/<branch>/`
   - Branch created if doesn't exist
   - Idempotent (can run twice without error)

3. **Project Initialized**
   - `.sow/project/` directory exists in worktree
   - `.sow/project/context/` directory exists
   - `state.yaml` file exists and is valid
   - Project type matches user's selection
   - Project name matches user's input

4. **Prompt Generated Correctly**
   - Includes base orchestrator layer
   - Includes project type orchestrator layer
   - Includes initial state layer
   - Includes user's initial prompt (if provided)
   - Layers separated by "---"

5. **Claude Launches**
   - Launches in worktree directory (not main repo)
   - Receives generated prompt
   - Receives Claude flags from wizard (if any)
   - User can start working immediately

6. **Success Messages Displayed**
   - Shows project name and branch
   - Indicates Claude is launching
   - Messages appear before Claude takes over terminal

### Test Requirements (TDD Approach)

Write tests BEFORE implementing the handler:

**Unit Tests** (add to `wizard_state_test.go`):

```go
func TestFinalize_CreatesWorktree(t *testing.T) {
    // Set up wizard with choices populated
    // Call finalize()
    // Verify worktree directory exists
}

func TestFinalize_InitializesProject(t *testing.T) {
    // Set up wizard with choices populated
    // Call finalize()
    // Verify project state.yaml exists
    // Verify project has correct type and name
}

func TestFinalize_GeneratesPrompt(t *testing.T) {
    // Set up wizard with choices populated (including prompt)
    // Call finalize()
    // Mock launchClaudeCode to capture prompt
    // Verify prompt has 3 layers
    // Verify prompt includes user's initial prompt
}

func TestFinalize_WithEmptyPrompt(t *testing.T) {
    // Set up wizard with empty prompt choice
    // Call finalize()
    // Verify prompt generation still works
    // Verify no "User's Initial Request" section
}

func TestFinalize_UncommittedChangesError(t *testing.T) {
    // Create uncommitted changes
    // Set current branch == target branch
    // Call finalize()
    // Verify error returned with instructions
}

func TestFinalize_SkipsUncommittedCheckWhenDifferentBranch(t *testing.T) {
    // Create uncommitted changes
    // Set current branch != target branch
    // Call finalize()
    // Verify no error (check skipped)
}
```

**Integration Tests**:

```go
func TestFullWizardFlow_BranchNamePath(t *testing.T) {
    // Mock all wizard screens
    // Simulate: Entry → CreateSource → TypeSelect → NameEntry → PromptEntry → Finalize
    // Verify: Project created, prompt generated, ready to launch Claude
}
```

**Manual Testing** (critical for Claude launch):

1. **Happy Path**:
   - Run `sow project`
   - Select: Create → From branch name → Exploration → "Web Based Agents" → "Research agents"
   - Verify: Worktree created, project initialized, Claude launches with prompt

2. **Uncommitted Changes**:
   - Make uncommitted changes on current branch
   - Try to create project on current branch
   - Verify: Error shown with fix instructions

3. **Different Branch**:
   - Make uncommitted changes on current branch
   - Create project on different branch
   - Verify: Works without error (check skipped)

4. **Empty Prompt**:
   - Run wizard, leave prompt empty
   - Verify: Project created, Claude launches (no initial prompt section)

5. **Claude Flags**:
   - Run `sow project -- --model opus`
   - Complete wizard
   - Verify: Claude launches with --model opus flag

## Technical Details

### Implementation Pattern

```go
func (w *Wizard) finalize() error {
    // Extract choices
    projectType := w.choices["type"].(string)
    name := w.choices["name"].(string)
    branch := w.choices["branch"].(string)
    initialPrompt := ""
    if prompt, ok := w.choices["prompt"].(string); ok {
        initialPrompt = prompt
    }

    // Step 1: Conditional uncommitted changes check
    currentBranch, err := w.ctx.Git().CurrentBranch()
    if err != nil {
        return fmt.Errorf("failed to get current branch: %w", err)
    }

    if currentBranch == branch {
        if err := sow.CheckUncommittedChanges(w.ctx); err != nil {
            return fmt.Errorf("repository has uncommitted changes\n\n"+
                "You are currently on branch '%s'.\n"+
                "Creating a worktree requires switching to a different branch first.\n\n"+
                "To fix:\n"+
                "  Commit: git add . && git commit -m \"message\"\n"+
                "  Or stash: git stash", currentBranch)
        }
    }

    // Step 2: Ensure worktree exists
    worktreePath := sow.WorktreePath(w.ctx.RepoRoot(), branch)
    if err := sow.EnsureWorktree(w.ctx, worktreePath, branch); err != nil {
        return fmt.Errorf("failed to create worktree: %w", err)
    }

    // Step 3: Initialize project in worktree
    worktreeCtx, err := sow.NewContext(worktreePath)
    if err != nil {
        return fmt.Errorf("failed to create worktree context: %w", err)
    }

    project, err := initializeProject(worktreeCtx, branch, name, nil)
    if err != nil {
        return fmt.Errorf("failed to initialize project: %w", err)
    }

    // Step 4: Generate 3-layer prompt
    prompt, err := generateNewProjectPrompt(project, initialPrompt)
    if err != nil {
        return fmt.Errorf("failed to generate prompt: %w", err)
    }

    // Step 5: Display success message
    fmt.Fprintf(os.Stdout, "✓ Initialized project '%s' on branch %s\n", name, branch)
    fmt.Fprintf(os.Stdout, "✓ Launching Claude in worktree...\n")

    // Step 6: Launch Claude Code
    // Note: Need access to cobra.Command for launchClaudeCode
    // This will require passing cmd to finalize() or storing it in Wizard
    if err := launchClaudeCode(w.cmd, worktreeCtx, prompt, w.claudeFlags); err != nil {
        return fmt.Errorf("failed to launch Claude: %w", err)
    }

    return nil
}
```

### Required Wizard Struct Change

The `finalize()` method needs access to the cobra Command for `launchClaudeCode()`. Two options:

**Option 1: Add cmd field to Wizard struct**:
```go
type Wizard struct {
    state       WizardState
    ctx         *sow.Context
    choices     map[string]interface{}
    claudeFlags []string
    cmd         *cobra.Command  // Add this field
}
```

Modify `NewWizard()` to accept cmd:
```go
func NewWizard(cmd *cobra.Command, ctx *sow.Context, claudeFlags []string) *Wizard {
    return &Wizard{
        state:       StateEntry,
        ctx:         ctx,
        choices:     make(map[string]interface{}),
        claudeFlags: claudeFlags,
        cmd:         cmd,
    }
}
```

Update `runWizard()` to pass cmd:
```go
wizard := NewWizard(cmd, mainCtx, claudeFlags)
```

**Option 2: Pass cmd to finalize()** (cleaner):
Change `finalize()` signature:
```go
func (w *Wizard) finalize(cmd *cobra.Command) error
```

Call from `Run()`:
```go
if w.state == StateComplete {
    return w.finalize(cmd)  // But cmd isn't available here
}
```

This doesn't work without changing Run() signature too.

**Recommendation**: Use Option 1 (add cmd field to Wizard struct). It's cleaner and follows the pattern of storing claudeFlags.

### Package and Imports

Add to `wizard_state.go`:
```go
import (
    "os"  // For os.Stdout
)
```

### File Structure

```
cli/cmd/project/
├── wizard_state.go           # MODIFY: Add cmd field, update NewWizard, implement finalize
├── wizard.go                 # MODIFY: Update NewWizard call
├── shared.go                 # READ: Use initializeProject, generateNewProjectPrompt, launchClaudeCode
├── wizard_state_test.go      # MODIFY: Add tests for finalize
```

## Relevant Inputs

### Existing Code to Understand

- `cli/cmd/project/shared.go:20-90` - `initializeProject()` function showing project creation logic
- `cli/cmd/project/shared.go:92-134` - `generateNewProjectPrompt()` function showing 3-layer prompt structure
- `cli/cmd/project/shared.go:172-207` - `launchClaudeCode()` function showing Claude launch
- `cli/internal/sow/worktree.go:11-87` - `WorktreePath()` and `EnsureWorktree()` functions
- `cli/internal/sow/worktree.go:89-122` - `CheckUncommittedChanges()` function
- `cli/internal/sow/context.go:30-111` - `NewContext()` for creating worktree context

### Design Documents

- `.sow/knowledge/designs/interactive-wizard-ux-flow.md:281-342` - Finalization flow specification
- `.sow/knowledge/designs/interactive-wizard-technical-implementation.md:450-521` - Finalization implementation example
- `.sow/project/context/issue-69.md:204-255` - Finalization requirements with conditional check

### Testing Patterns

- `cli/cmd/project/shared_test.go:74-238` - Tests for `initializeProject()` showing project creation testing
- `cli/cmd/project/shared_test.go:240-325` - Tests for `generateNewProjectPrompt()` showing prompt testing

## Examples

### Example: Successful Finalization

```
[User completes prompt entry]

[Finalization starts]

✓ Initialized project 'Web Based Agents' on branch explore/web-based-agents
✓ Launching Claude in worktree...

[Claude Code launches in new terminal instance]

Hi! I see you're working on a new project "Web Based Agents" on branch explore/web-based-agents.

[Base Orchestrator prompt]
...
[Project Type Orchestrator prompt]
...
[Initial State prompt]
...

## User's Initial Request

Research the landscape of web-based agent frameworks and compare their architectures.
```

### Example: Uncommitted Changes Error

```
[User on feat/auth branch with uncommitted changes, tries to create project on feat/auth]

Error: repository has uncommitted changes

You are currently on branch 'feat/auth'.
Creating a worktree requires switching to a different branch first.

To fix:
  Commit: git add . && git commit -m "message"
  Or stash: git stash

[Wizard exits]
```

### Example: Different Branch (Check Skipped)

```
[User on main branch with uncommitted changes, creates project on feat/auth]

[Check skipped because main != feat/auth]

✓ Initialized project 'Add Authentication' on branch feat/auth
✓ Launching Claude in worktree...

[Claude launches successfully]
```

## Dependencies

### Upstream Dependencies (Must Complete First)

- **Work Unit 001**: Wizard Foundation and State Machine ✅ COMPLETE
  - Provides: `WizardState` enum with `StateComplete`
  - Provides: Wizard struct and Run() method
  - Provides: Shared utilities in `shared.go`

- **All previous tasks** (010-040) in this work unit
  - Provide: All wizard choices (type, name, branch, prompt)
  - Provide: State transition to `StateComplete`

### Downstream Dependencies (Will Use This Task)

- None - this is the terminal handler that completes the wizard

## Constraints

### Git Worktree Requirements

- Must check for uncommitted changes only when current == target branch
- Must use `sow.EnsureWorktree()` (idempotent, handles branch creation)
- Must create worktree context after worktree exists
- Must initialize project in worktree context (not main repo context)

### Prompt Generation Requirements

- Must use `generateNewProjectPrompt()` for consistency
- Must pass empty string if no initial prompt
- Must handle template rendering errors

### Claude Launch Requirements

- Must launch in worktree directory
- Must pass all Claude flags from wizard
- Must display success message BEFORE launching (Claude takes over terminal)

### Testing Requirements

- Tests written BEFORE implementation (TDD)
- Unit tests for each step of finalization
- Integration test for full wizard flow
- Manual tests for Claude launch and uncommitted changes scenarios

### What NOT to Do

- ❌ Don't skip uncommitted changes check - causes git errors
- ❌ Don't always run uncommitted changes check - wrong when current != target
- ❌ Don't initialize project in main repo context - use worktree context
- ❌ Don't modify `initializeProject()` or other shared utilities - they're shared
- ❌ Don't skip success messages - user needs feedback before Claude launches
- ❌ Don't launch Claude in main repo - must use worktree path

## Notes

### Critical Implementation Details

1. **Conditional Check Logic**: The uncommitted changes check ONLY runs when `currentBranch == targetBranch`. This is a git worktree requirement, not a sow design choice. Getting this wrong causes either false errors or git failures.

2. **Context Switch**: After creating the worktree, you MUST create a new context for it. The wizard's context points to the main repo. The project needs to be initialized in the worktree.

3. **Cobra Command Access**: The wizard needs access to the cobra Command for `launchClaudeCode()`. Adding it as a field in the Wizard struct is the cleanest solution.

4. **Success Messages**: Display BEFORE launching Claude. Once Claude launches, it takes over the terminal and the wizard's output is lost.

### Testing Strategy

**Unit Tests**: Test each step independently:
- Worktree creation
- Project initialization
- Prompt generation
- Uncommitted changes check (both paths)

**Integration Tests**: Test full wizard flow end-to-end (except Claude launch)

**Manual Tests**: Critical for:
- Claude launch (can't easily mock)
- Uncommitted changes check (need real git state)
- Success messages display correctly

### Error Message Philosophy

All error messages should:
- Explain what went wrong
- Explain why it went wrong
- Provide concrete fix instructions
- Use the same phrasing as git (when relevant)

The uncommitted changes error follows this pattern - it tells you why (need to switch branches), what to do (commit or stash), and gives exact commands.

### Future Work

If we add support for continuing from the wizard later, we'll need to:
1. Add a continuation path that skips worktree creation (already exists)
2. Generate continuation prompt instead of new project prompt
3. Handle project already initialized case

This task focuses on the new project creation path only.
