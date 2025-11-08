# Task 040: Continuation Finalization Flow

## Context

This task implements the continuation-specific path in the wizard's `finalize()` method. When a user completes the continuation workflow (StateProjectSelect → StateContinuePrompt → StateComplete), the finalize method must load the project state, generate the appropriate 3-layer continuation prompt, and launch Claude in the project's worktree.

This is part of work unit 004 (Project Continuation Workflow). The wizard foundation was implemented in work unit 001, and the continuation state handlers were implemented in tasks 020 and 030.

A critical difference from the creation path: continuation does NOT perform uncommitted changes check, since the worktree already exists and we're not switching branches in the main repo.

## Requirements

Extend the existing `finalize()` method in `cli/cmd/project/wizard_state.go` to handle the continuation path.

### Current finalize() Structure

The existing finalize() method (from work unit 001) currently handles the creation path. It needs to be extended to detect and handle continuation:

```go
func (w *Wizard) finalize() error {
    // Determine which path we're on
    action, ok := w.choices["action"].(string)
    if !ok {
        return fmt.Errorf("internal error: action choice not set")
    }

    if action == "create" {
        // Existing creation path (from work unit 001)
        // ...
    } else if action == "continue" {
        // NEW: Continuation path (this task)
        return w.finalizeContinuation()
    }

    return fmt.Errorf("internal error: unknown action: %s", action)
}
```

### Continuation Finalization Steps

Implement `finalizeContinuation()` method with the following steps:

**Step 1: Extract Wizard Choices**

```go
func (w *Wizard) finalizeContinuation() error {
    // Extract project info
    proj, ok := w.choices["project"].(ProjectInfo)
    if !ok {
        return fmt.Errorf("internal error: project choice not set or invalid")
    }

    // Extract user's continuation prompt (may be empty)
    userPrompt, ok := w.choices["prompt"].(string)
    if !ok {
        return fmt.Errorf("internal error: prompt choice not set or invalid")
    }

    // ... continue with finalization
}
```

**Step 2: Ensure Worktree Exists (Idempotent)**

Even though the worktree should exist, call EnsureWorktree for robustness:

```go
worktreePath := sow.WorktreePath(w.ctx.MainRepoRoot(), proj.Branch)
if err := sow.EnsureWorktree(w.ctx, worktreePath, proj.Branch); err != nil {
    return fmt.Errorf("failed to ensure worktree: %w", err)
}
```

Note: EnsureWorktree is idempotent - if the worktree exists, it does nothing. This handles edge cases where the worktree was deleted between selection and finalization.

**Step 3: Create Worktree Context**

Create a fresh context rooted in the worktree:

```go
worktreeCtx, err := sow.NewContext(worktreePath)
if err != nil {
    return fmt.Errorf("failed to create worktree context: %w", err)
}
```

**Step 4: Load Fresh Project State**

Load the current project state from the worktree (not the cached ProjectInfo):

```go
projectState, err := state.Load(worktreeCtx)
if err != nil {
    return fmt.Errorf("failed to load project state: %w", err)
}
```

This ensures we have the absolute latest state, including any changes made by other processes.

**Step 5: Generate 3-Layer Continuation Prompt**

Use the existing `generateContinuePrompt()` function:

```go
basePrompt, err := generateContinuePrompt(projectState)
if err != nil {
    return fmt.Errorf("failed to generate continuation prompt: %w", err)
}
```

**Step 6: Append User Prompt if Provided**

Add the user's continuation prompt if they provided one:

```go
fullPrompt := basePrompt
if userPrompt != "" {
    fullPrompt += "\n\nUser request:\n" + userPrompt
}
```

**Step 7: Display Success Message**

Show confirmation before launching Claude:

```go
fmt.Fprintf(os.Stderr, "✓ Continuing project '%s' on branch %s\n", proj.Name, proj.Branch)
```

**Step 8: Launch Claude in Worktree**

Use the existing `launchClaudeCode()` function:

```go
if w.cmd != nil {
    if err := launchClaudeCode(w.cmd, worktreeCtx, fullPrompt, w.claudeFlags); err != nil {
        return fmt.Errorf("failed to launch Claude: %w", err)
    }
}

return nil
```

Note: The `w.cmd != nil` check allows tests to skip launching Claude.

### Complete Implementation

```go
func (w *Wizard) finalizeContinuation() error {
    // 1. Extract choices
    proj, ok := w.choices["project"].(ProjectInfo)
    if !ok {
        return fmt.Errorf("internal error: project choice not set or invalid")
    }

    userPrompt, ok := w.choices["prompt"].(string)
    if !ok {
        return fmt.Errorf("internal error: prompt choice not set or invalid")
    }

    // 2. Ensure worktree exists (idempotent)
    worktreePath := sow.WorktreePath(w.ctx.MainRepoRoot(), proj.Branch)
    if err := sow.EnsureWorktree(w.ctx, worktreePath, proj.Branch); err != nil {
        return fmt.Errorf("failed to ensure worktree: %w", err)
    }

    // 3. Create worktree context
    worktreeCtx, err := sow.NewContext(worktreePath)
    if err != nil {
        return fmt.Errorf("failed to create worktree context: %w", err)
    }

    // 4. Load fresh project state
    projectState, err := state.Load(worktreeCtx)
    if err != nil {
        return fmt.Errorf("failed to load project state: %w", err)
    }

    // 5. Generate 3-layer continuation prompt
    basePrompt, err := generateContinuePrompt(projectState)
    if err != nil {
        return fmt.Errorf("failed to generate continuation prompt: %w", err)
    }

    // 6. Append user prompt if provided
    fullPrompt := basePrompt
    if userPrompt != "" {
        fullPrompt += "\n\nUser request:\n" + userPrompt
    }

    // 7. Success message
    fmt.Fprintf(os.Stderr, "✓ Continuing project '%s' on branch %s\n", proj.Name, proj.Branch)

    // 8. Launch Claude
    if w.cmd != nil {
        if err := launchClaudeCode(w.cmd, worktreeCtx, fullPrompt, w.claudeFlags); err != nil {
            return fmt.Errorf("failed to launch Claude: %w", err)
        }
    }

    return nil
}
```

## Acceptance Criteria

### Functional Requirements

1. **Detects continuation path correctly**
   - Checks w.choices["action"] == "continue"
   - Routes to finalizeContinuation() method
   - Returns error for unknown actions

2. **Extracts choices correctly**
   - Retrieves project as ProjectInfo
   - Retrieves prompt as string (may be empty)
   - Returns descriptive errors for missing/invalid choices

3. **Ensures worktree exists**
   - Calls EnsureWorktree (idempotent operation)
   - Handles case where worktree was deleted
   - Uses MainRepoRoot() for correct path resolution

4. **Creates worktree context**
   - Context rooted at worktree directory
   - Independent of main wizard context
   - Used for loading state and launching Claude

5. **Loads fresh project state**
   - Uses state.Load() with worktree context
   - Gets latest state (not cached ProjectInfo)
   - Returns error if state loading fails

6. **Generates correct prompt**
   - Uses generateContinuePrompt() for 3-layer base
   - Appends user prompt if provided
   - Skips user prompt section if empty

7. **Prompt structure is correct**
   - Layer 1: Base orchestrator prompt
   - Layer 2: Project type orchestrator prompt
   - Layer 3: Current state prompt
   - Optional: User request appended with clear delimiter

8. **Launches Claude successfully**
   - Launches in worktree directory
   - Passes full prompt (base + user)
   - Passes through Claude flags if provided
   - Skips launch in tests (when w.cmd is nil)

9. **Does NOT check uncommitted changes**
   - No call to CheckUncommittedChanges()
   - This is intentional and correct for continuation
   - Documented in comments

10. **Success message displayed**
    - Shows project name and branch
    - Displayed to stderr (not captured in tests)
    - Uses checkmark character for visual feedback

### Test Requirements (TDD Approach)

Write tests FIRST in `cli/cmd/project/wizard_test.go`, then implement to pass them:

**Unit tests for finalizeContinuation:**
- Valid choices → all steps execute successfully
- Missing project choice → returns error
- Missing prompt choice → returns error
- Worktree doesn't exist → EnsureWorktree creates it
- State loading fails → returns error
- Prompt generation fails → returns error
- Empty user prompt → base prompt only (no "User request:" section)
- Non-empty user prompt → appended to base prompt

**Integration tests:**
- Full continuation flow end-to-end
- Verify prompt structure (3 layers + optional user section)
- Verify Claude launched in correct directory
- Verify Claude receives correct prompt

**Negative tests:**
- Invalid action type → routed correctly or error
- Corrupted project state → error with context
- Worktree creation failure → error with context

## Technical Details

### Go Packages and Imports

```go
import (
    "fmt"
    "os"

    "github.com/jmgilman/sow/cli/internal/sdks/project/state"
    "github.com/jmgilman/sow/cli/internal/sow"
)
```

### File Location

Modify existing file: `cli/cmd/project/wizard_state.go`

- Extend `finalize()` method to detect continuation action
- Add `finalizeContinuation()` helper method

### Critical Design Decisions

**Why no uncommitted changes check?**

The creation path calls `CheckUncommittedChanges()` before creating a worktree. Continuation does NOT do this because:

1. The worktree already exists (no branch switching in main repo needed)
2. We might be running FROM a worktree (wizard supports this)
3. Main repo state is irrelevant to the worktree's work

This should be documented in a comment:

```go
// NOTE: Unlike creation, continuation does NOT check uncommitted changes.
// The worktree already exists, so there's no risk of needing to switch
// branches in the main repo. This is intentional.
```

**Why call EnsureWorktree if it exists?**

EnsureWorktree is idempotent - if the worktree exists, it immediately returns nil. Calling it handles the edge case where:
- User selected a project
- Worktree was deleted between selection and finalization
- EnsureWorktree recreates it

This makes the workflow more robust without adding complexity.

**Why load fresh state instead of using ProjectInfo?**

The ProjectInfo from discovery is a snapshot at discovery time. Between discovery and finalization:
- Another process might have modified the project
- Tasks might have been added/completed
- State machine might have transitioned

Loading fresh ensures the continuation prompt reflects the absolute current state.

### generateContinuePrompt Function

This function already exists in `cli/cmd/project/shared.go` (from the original continue.go extraction). It:

1. Renders base orchestrator template
2. Gets project type orchestrator prompt
3. Gets current state prompt
4. Combines all three with "---" separators

See `cli/cmd/project/shared.go` lines 143-177 for implementation.

### User Prompt Formatting

When appending the user's prompt, use a clear delimiter:

```
[3-layer base prompt]

User request:
[user's prompt text]
```

This makes it obvious to Claude that this is a new request from the user, separate from the state-based context.

## Relevant Inputs

**Design documents:**
- `.sow/knowledge/designs/interactive-wizard-ux-flow.md` - Lines 362-376 contain finalization flow specification
- `.sow/knowledge/designs/interactive-wizard-technical-implementation.md` - Finalization implementation details

**Existing finalize() implementation:**
- `cli/cmd/project/wizard_state.go` - Lines 349-419 show existing creation finalization pattern

**Shared utilities:**
- `cli/cmd/project/shared.go` - generateContinuePrompt() function (lines 143-177)
- `cli/cmd/project/shared.go` - launchClaudeCode() function (lines 179-213)

**Task dependencies:**
- Task 010: ProjectInfo struct definition
- Task 020: handleProjectSelect sets w.choices["project"]
- Task 030: handleContinuePrompt sets w.choices["prompt"]

**Context and worktree utilities:**
- `cli/internal/sow/context.go` - NewContext() for worktree context creation
- `cli/internal/sow/context.go` - MainRepoRoot() method
- `cli/internal/sow/worktree.go` - EnsureWorktree() function
- `cli/internal/sow/worktree.go` - WorktreePath() function

**State loading:**
- `cli/internal/sdks/project/state/loader.go` - Load() function for loading project state

## Examples

### Example 1: Complete finalizeContinuation Implementation

See "Complete Implementation" section in Requirements above.

### Example 2: finalize() Method with Routing

```go
func (w *Wizard) finalize() error {
    // Determine path
    action, ok := w.choices["action"].(string)
    if !ok {
        return fmt.Errorf("internal error: action choice not set")
    }

    switch action {
    case "create":
        return w.finalizeCreation() // Existing from work unit 001
    case "continue":
        return w.finalizeContinuation() // New in this task
    default:
        return fmt.Errorf("internal error: unknown action: %s", action)
    }
}

func (w *Wizard) finalizeCreation() error {
    // Existing creation logic from work unit 001
    // (move existing finalize() body here)
}

func (w *Wizard) finalizeContinuation() error {
    // New continuation logic (this task)
}
```

### Example 3: Generated Prompt Structure

```markdown
[LAYER 1: Base Orchestrator]
You are the orchestrator agent for the sow system...
(full base orchestrator prompt from templates/greet/orchestrator.md)

---

[LAYER 2: Project Type Orchestrator]
You are coordinating a Standard project...
Current phase: implementation
(project-type-specific orchestrator guidance)

---

[LAYER 3: Current State Prompt]
You are in the implementation phase, task 4 of 5 is in progress.

Current task: Implement token refresh logic

Review the task description and continue work.
(state-specific guidance)

User request:
Let's focus on the token refresh logic and write
comprehensive integration tests for the auth flow
```

### Example 4: Test Structure

```go
func TestFinalizeContinuation(t *testing.T) {
    t.Run("valid continuation flow completes successfully", func(t *testing.T) {
        // Setup: Create wizard with valid choices
        // Execute: Call finalizeContinuation
        // Assert: No error, prompt generated correctly
    })

    t.Run("empty user prompt excluded from final prompt", func(t *testing.T) {
        // Setup: wizard.choices["prompt"] = ""
        // Execute: Call finalizeContinuation
        // Assert: Final prompt has no "User request:" section
    })

    t.Run("non-empty user prompt appended", func(t *testing.T) {
        // Setup: wizard.choices["prompt"] = "focus on tests"
        // Execute: Call finalizeContinuation
        // Assert: Final prompt includes "\n\nUser request:\nfocus on tests"
    })

    t.Run("missing project choice returns error", func(t *testing.T) {
        // Setup: wizard.choices without "project"
        // Execute: Call finalizeContinuation
        // Assert: Returns error about missing project
    })

    t.Run("state loading failure returns error", func(t *testing.T) {
        // Setup: Create wizard, corrupt state file
        // Execute: Call finalizeContinuation
        // Assert: Returns error about failed state load
    })
}
```

## Dependencies

**Required from work unit 001:**
- Wizard struct with finalize() method exists
- launchClaudeCode() function exists in shared.go
- generateContinuePrompt() function exists in shared.go

**Required from task 010:**
- ProjectInfo struct defined
- Used by w.choices["project"]

**Required from task 020:**
- handleProjectSelect sets w.choices["project"] as ProjectInfo

**Required from task 030:**
- handleContinuePrompt sets w.choices["prompt"] as string

**Existing infrastructure:**
- `cli/internal/sow/worktree.go` - EnsureWorktree, WorktreePath
- `cli/internal/sow/context.go` - NewContext, MainRepoRoot
- `cli/internal/sdks/project/state/` - Load function

## Constraints

### Critical Behavior Differences from Creation

**NO uncommitted changes check:**
- Creation path: Must check (might need to switch branches)
- Continuation path: Must NOT check (worktree exists, no switching needed)

This is documented in issue #71 and the design docs.

### State Freshness

- Must load fresh state via state.Load()
- Do NOT rely on cached ProjectInfo.Phase/TasksCompleted
- ProjectInfo is for display only, not for prompt generation

### Prompt Structure

- Must use generateContinuePrompt() for base 3-layer prompt
- Must append user prompt with clear delimiter if provided
- Must NOT modify or filter user's prompt text
- Empty user prompt is valid (don't append anything)

### Error Handling

- Missing choices → internal error (shouldn't happen if state machine correct)
- Worktree issues → descriptive error (actionable for user)
- State loading failure → descriptive error
- Claude launch failure → descriptive error

### Testing Considerations

- w.cmd may be nil in tests → skip Claude launch
- Mock or stub file system for test isolation
- Verify prompt structure without actually launching Claude
- Test both empty and non-empty user prompts
