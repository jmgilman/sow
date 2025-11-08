# Task 040: Prompt Entry Enhancement and Project Finalization with Issue Context

## Context

This task completes the GitHub issue integration workflow by enhancing the prompt entry screen to display issue context and updating the finalization logic to store issue metadata in the project state. This is the final implementation task that ties together all previous work (GitHub validation, issue listing, validation, and branch creation).

The prompt entry screen already exists from Work Unit 002 but needs enhancement to display issue information when present. The finalization logic in `shared.go` already handles project creation and Claude launch but needs to be updated to accept and store issue metadata.

## Requirements

### 1. Enhance Prompt Entry to Show Issue Context

Modify `handlePromptEntry()` in `cli/cmd/project/wizard_state.go` to display different context based on whether the project was created from an issue or branch name:

```go
func (w *Wizard) handlePromptEntry() error {
    var prompt string

    // Build context display based on project source
    var contextLines []string

    // Check for issue context (GitHub issue path)
    if issue, ok := w.choices["issue"].(*sow.Issue); ok {
        contextLines = append(contextLines,
            fmt.Sprintf("Issue: #%d - %s", issue.Number, issue.Title))
    }

    // Show branch name
    if branchName, ok := w.choices["branch"].(string); ok {
        contextLines = append(contextLines, fmt.Sprintf("Branch: %s", branchName))
    } else if name, ok := w.choices["name"].(string); ok {
        // Branch name path - compute branch name for display
        projectType := w.choices["type"].(string)
        prefix := getTypePrefix(projectType)
        normalized := normalizeName(name)
        contextLines = append(contextLines,
            fmt.Sprintf("Branch: %s%s", prefix, normalized))
    }

    // Add project type for clarity
    if projectType, ok := w.choices["type"].(string); ok {
        typeConfig := projectTypes[projectType]
        contextLines = append(contextLines,
            fmt.Sprintf("Type: %s", typeConfig.Description))
    }

    contextDisplay := strings.Join(contextLines, "\n")
    instructionText := contextDisplay + "\n\nPress Ctrl+E to open $EDITOR for multi-line input"

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewText().
                Title("Enter your task or question for Claude (optional):").
                Description(instructionText).
                CharLimit(10000).
                Value(&prompt).
                EditorExtension(".md"),
        ),
    )

    if err := form.Run(); err != nil {
        if errors.Is(err, huh.ErrUserAborted) {
            w.state = StateCancelled
            return nil
        }
        return fmt.Errorf("prompt entry error: %w", err)
    }

    w.choices["prompt"] = prompt
    w.state = StateComplete

    return nil
}
```

### 2. Update Finalization to Store Issue Metadata

The `finalize()` method in `wizard_state.go` already calls `initializeProject()`. Update it to pass issue information:

```go
func (w *Wizard) finalize() error {
    // Extract wizard choices
    name, ok := w.choices["name"].(string)
    if !ok {
        return fmt.Errorf("name choice not set or invalid")
    }

    branch, ok := w.choices["branch"].(string)
    if !ok {
        return fmt.Errorf("branch choice not set or invalid")
    }

    initialPrompt := ""
    if prompt, ok := w.choices["prompt"].(string); ok {
        initialPrompt = prompt
    }

    // Extract issue if present (GitHub issue path)
    var issue *sow.Issue
    if issueData, ok := w.choices["issue"].(*sow.Issue); ok {
        issue = issueData
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

    // Step 3: Initialize project in worktree WITH issue metadata
    worktreeCtx, err := sow.NewContext(worktreePath)
    if err != nil {
        return fmt.Errorf("failed to create worktree context: %w", err)
    }

    // Pass issue to initializeProject (will be nil for branch name path)
    project, err := initializeProject(worktreeCtx, branch, name, issue)
    if err != nil {
        return fmt.Errorf("failed to initialize project: %w", err)
    }

    // Step 4: Generate 3-layer prompt
    prompt, err := generateNewProjectPrompt(project, initialPrompt)
    if err != nil {
        return fmt.Errorf("failed to generate prompt: %w", err)
    }

    // Step 5: Display success message
    _, _ = fmt.Fprintf(os.Stdout, "✓ Initialized project '%s' on branch %s\n", name, branch)
    if issue != nil {
        _, _ = fmt.Fprintf(os.Stdout, "✓ Linked to issue #%d: %s\n", issue.Number, issue.Title)
    }
    _, _ = fmt.Fprintf(os.Stdout, "✓ Launching Claude in worktree...\n")

    // Step 6: Launch Claude Code
    if w.cmd != nil {
        if err := launchClaudeCode(w.cmd, worktreeCtx, prompt, w.claudeFlags); err != nil {
            return fmt.Errorf("failed to launch Claude: %w", err)
        }
    }

    return nil
}
```

**Note**: The `initializeProject()` function in `shared.go` already accepts an `issue` parameter and handles writing the issue context file. No changes needed there - it was designed for this in Work Unit 002.

### 3. Verify Issue Context in Generated Prompt

The `generateNewProjectPrompt()` function in `shared.go` already includes the 3-layer prompt structure. When an issue is present, the issue context file will be automatically included as an implementation phase input artifact, making it available to Claude through the project's input system.

No code changes needed here - the existing implementation already works correctly:

```go
// From shared.go - already handles issue context
func initializeProject(ctx *sow.Context, branch string, description string, issue *sow.Issue) (*state.Project, error) {
    // ... setup ...

    var initialInputs map[string][]projschema.ArtifactState
    if issue != nil {
        // Write issue body to file
        issueFileName := fmt.Sprintf("issue-%d.md", issue.Number)
        issuePath := filepath.Join(contextDir, issueFileName)
        issueContent := fmt.Sprintf("# Issue #%d: %s\n\n**URL**: %s\n**State**: %s\n\n## Description\n\n%s\n",
            issue.Number, issue.Title, issue.URL, issue.State, issue.Body)

        if err := os.WriteFile(issuePath, []byte(issueContent), 0644); err != nil {
            return nil, fmt.Errorf("failed to write issue file: %w", err)
        }

        // Create github_issue artifact for implementation phase
        issueArtifact := projschema.ArtifactState{
            Type:       "github_issue",
            Path:       fmt.Sprintf("context/%s", issueFileName),
            Approved:   true,
            Created_at: time.Now(),
            Metadata: map[string]interface{}{
                "issue_number": issue.Number,
                "issue_url":    issue.URL,
                "issue_title":  issue.Title,
            },
        }

        initialInputs = map[string][]projschema.ArtifactState{
            "implementation": {issueArtifact},
        }
    }

    proj, err := state.Create(ctx, branch, description, initialInputs)
    // ...
}
```

## Acceptance Criteria

### Functional Requirements

1. **Issue context displayed**: When project created from issue, prompt entry shows "Issue: #123 - Title"
2. **Branch context displayed**: All paths show "Branch: <branch-name>"
3. **Type context displayed**: All paths show "Type: <description>"
4. **Ctrl+E supported**: External editor opens when user presses Ctrl+E
5. **Optional prompt**: User can skip prompt entry (leave blank)
6. **Issue metadata stored**: Project state contains issue number, title, URL when created from issue
7. **Issue context file created**: File `.sow/project/context/issue-<number>.md` created with issue content
8. **Success message enhanced**: Shows linked issue information when applicable
9. **Claude receives context**: Issue context available to Claude through project input system

### Test Requirements (TDD Approach)

Write tests before implementing. Add to `cli/cmd/project/wizard_state_test.go`:

#### Test 1: Prompt Entry with Issue Context
```go
func TestHandlePromptEntry_WithIssueContext(t *testing.T) {
    wizard := &Wizard{
        state: StatePromptEntry,
        ctx:   testContext(t),
        choices: map[string]interface{}{
            "issue": &sow.Issue{
                Number: 123,
                Title:  "Add JWT authentication",
            },
            "branch": "feat/add-jwt-authentication-123",
            "type":   "standard",
        },
    }

    // Note: Full form testing difficult without UI interaction
    // This test verifies method doesn't crash with issue context
    // Real verification would be integration test or manual test

    // For now, verify context building logic separately
    var contextLines []string
    if issue, ok := wizard.choices["issue"].(*sow.Issue); ok {
        contextLines = append(contextLines,
            fmt.Sprintf("Issue: #%d - %s", issue.Number, issue.Title))
    }

    expected := "Issue: #123 - Add JWT authentication"
    if contextLines[0] != expected {
        t.Errorf("expected %q, got %q", expected, contextLines[0])
    }
}
```

#### Test 2: Prompt Entry with Branch Name Context
```go
func TestHandlePromptEntry_WithBranchNameContext(t *testing.T) {
    wizard := &Wizard{
        state: StatePromptEntry,
        ctx:   testContext(t),
        choices: map[string]interface{}{
            "name":   "Web Based Agents",
            "type":   "exploration",
            "branch": "explore/web-based-agents",
        },
    }

    // Verify branch context computed correctly
    branchName := wizard.choices["branch"].(string)
    expected := "explore/web-based-agents"

    if branchName != expected {
        t.Errorf("expected branch %q, got %q", expected, branchName)
    }
}
```

#### Test 3: Finalization with Issue Metadata
```go
func TestFinalize_WithIssue(t *testing.T) {
    // Setup test context with temporary directory
    tmpDir := t.TempDir()

    // Create git repo in temp dir
    initGitRepo(t, tmpDir)

    ctx, err := sow.NewContext(tmpDir)
    if err != nil {
        t.Fatalf("failed to create context: %v", err)
    }

    issue := &sow.Issue{
        Number: 123,
        Title:  "Test Issue",
        Body:   "Issue description here",
        State:  "open",
        URL:    "https://github.com/test/repo/issues/123",
    }

    wizard := &Wizard{
        state: StateComplete,
        ctx:   ctx,
        choices: map[string]interface{}{
            "name":   "Test Issue",
            "branch": "feat/test-issue-123",
            "type":   "standard",
            "issue":  issue,
            "prompt": "",
        },
        cmd: nil, // Skip Claude launch in test
    }

    err = wizard.finalize()
    if err != nil {
        t.Fatalf("finalize failed: %v", err)
    }

    // Verify issue context file created
    worktreePath := sow.WorktreePath(tmpDir, "feat/test-issue-123")
    issueFilePath := filepath.Join(worktreePath, ".sow", "project", "context", "issue-123.md")

    if _, err := os.Stat(issueFilePath); os.IsNotExist(err) {
        t.Error("issue context file not created")
    }

    // Verify file contains issue information
    content, err := os.ReadFile(issueFilePath)
    if err != nil {
        t.Fatalf("failed to read issue file: %v", err)
    }

    if !strings.Contains(string(content), "Test Issue") {
        t.Error("issue file doesn't contain issue title")
    }

    if !strings.Contains(string(content), "Issue description here") {
        t.Error("issue file doesn't contain issue body")
    }
}
```

#### Test 4: Finalization without Issue (Branch Name Path)
```go
func TestFinalize_WithoutIssue(t *testing.T) {
    tmpDir := t.TempDir()
    initGitRepo(t, tmpDir)

    ctx, err := sow.NewContext(tmpDir)
    if err != nil {
        t.Fatalf("failed to create context: %v", err)
    }

    wizard := &Wizard{
        state: StateComplete,
        ctx:   ctx,
        choices: map[string]interface{}{
            "name":   "Test Project",
            "branch": "feat/test-project",
            "type":   "standard",
            "prompt": "",
            // No issue in choices
        },
        cmd: nil,
    }

    err = wizard.finalize()
    if err != nil {
        t.Fatalf("finalize failed: %v", err)
    }

    // Verify no issue context file created
    worktreePath := sow.WorktreePath(tmpDir, "feat/test-project")
    contextDir := filepath.Join(worktreePath, ".sow", "project", "context")

    entries, err := os.ReadDir(contextDir)
    if err != nil {
        t.Fatalf("failed to read context dir: %v", err)
    }

    // Should be no files (or at least no issue-*.md files)
    for _, entry := range entries {
        if strings.HasPrefix(entry.Name(), "issue-") {
            t.Errorf("unexpected issue file: %s", entry.Name())
        }
    }
}
```

### Non-Functional Requirements

- **Backward compatibility**: Branch name path continues to work without issues
- **Clear context**: All relevant information displayed before prompt entry
- **Helpful success messages**: Confirmation messages include issue link when applicable
- **Persistent metadata**: Issue information stored in project state for future reference

## Technical Details

### Context Display Format

The prompt entry screen displays context in this order:
1. Issue information (if present): "Issue: #123 - Title"
2. Branch name: "Branch: feat/..."
3. Type: "Type: Feature work and bug fixes"

This provides complete context about what's being created.

### External Editor Integration

The `.EditorExtension(".md")` tells huh to use `.md` extension for the temporary file when opening `$EDITOR`. This enables syntax highlighting in editors that support it.

Keybinding is **Ctrl+E** (not Ctrl+O as originally drafted - verified in huh library documentation).

### Issue Context File Format

The issue context file follows this format:

```markdown
# Issue #123: Add JWT authentication

**URL**: https://github.com/owner/repo/issues/123
**State**: open

## Description

Implement JWT-based authentication for API endpoints.
Support both access tokens and refresh tokens.
```

This file is:
- Created in `.sow/project/context/issue-123.md`
- Registered as implementation phase input artifact
- Automatically available to Claude through project input system
- Includes issue metadata for future reference

### Project State with Issue Metadata

When created from an issue, the project state includes:

```yaml
name: "Add JWT authentication"
type: standard
branch: feat/add-jwt-authentication-123
phases:
  implementation:
    inputs:
      - type: github_issue
        path: context/issue-123.md
        approved: true
        metadata:
          issue_number: 123
          issue_url: https://github.com/owner/repo/issues/123
          issue_title: "Add JWT authentication"
```

This metadata enables:
- Future features to reference the original issue
- Orchestrator to include issue link in PR descriptions
- Tracking which issues have been worked on

### Success Message Enhancement

The success message differs based on project source:

**From Issue**:
```
✓ Initialized project 'Add JWT authentication' on branch feat/add-jwt-authentication-123
✓ Linked to issue #123: Add JWT authentication
✓ Launching Claude in worktree...
```

**From Branch Name**:
```
✓ Initialized project 'Web Based Agents' on branch explore/web-based-agents
✓ Launching Claude in worktree...
```

## Relevant Inputs

### Wizard State Machine
- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/cli/cmd/project/wizard_state.go`
  - Lines 293-333: Current `handlePromptEntry()` implementation to enhance
  - Lines 349-419: Current `finalize()` method to update

### Shared Project Logic
- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/cli/cmd/project/shared.go`
  - Lines 20-97: `initializeProject()` - already accepts issue parameter
  - Lines 59-88: Issue context file creation logic (already implemented)
  - Lines 99-141: `generateNewProjectPrompt()` - no changes needed

### Helper Functions
- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/cli/cmd/project/wizard_helpers.go`
  - Lines 42-91: `normalizeName()` for branch name display
  - Lines 93-106: `getTypePrefix()` for branch prefix
  - Lines 15-40: `projectTypes` map for type descriptions

### Worktree Management
- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/cli/internal/sow/worktree.go`
  - Lines 14-16: `WorktreePath()` function
  - Lines 18-87: `EnsureWorktree()` function
  - Lines 89-122: `CheckUncommittedChanges()` function

### Design Specifications
- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/.sow/knowledge/designs/interactive-wizard-ux-flow.md`
  - Lines 143-173: Initial prompt entry with issue context
  - Lines 254-278: Initial prompt entry for branch name path

- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/.sow/knowledge/designs/huh-library-verification.md`
  - Lines 183-248: External editor integration (Ctrl+E keybinding)

- `/Users/josh/code/sow/.sow/worktrees/70-github-issue-integration-workflow/.sow/project/context/issue-70.md`
  - Lines 342-388: Prompt entry enhancement specification
  - Lines 390-439: Finalization enhancement specification
  - Lines 760-775: Issue metadata storage format

## Examples

### Example: GitHub Issue Path - Complete Flow
```
User: Selects issue #124 "Refactor database schema"
User: Selects "Feature work and bug fixes" type
Wizard: Creates branch "feat/refactor-database-schema-124"
Display: Prompt entry screen:

  "Issue: #124 - Refactor database schema
  Branch: feat/refactor-database-schema-124
  Type: Feature work and bug fixes

  Enter your task or question for Claude (optional):
  Press Ctrl+E to open $EDITOR for multi-line input"

User: Types "Focus on normalizing the users table first"
User: Presses Enter
Wizard: Creates worktree at .sow/worktrees/feat/refactor-database-schema-124/
Wizard: Initializes project with issue metadata
Wizard: Creates .sow/project/context/issue-124.md
Wizard: Generates 3-layer prompt including issue context
Display: Success messages:
  "✓ Initialized project 'Refactor database schema' on branch feat/refactor-database-schema-124
  ✓ Linked to issue #124: Refactor database schema
  ✓ Launching Claude in worktree..."
Claude: Launches in worktree with issue context available
```

### Example: Branch Name Path - No Changes
```
User: Selects "From branch name"
User: Selects "Exploration" type
User: Enters "Web Based Agents"
Display: Prompt entry screen:

  "Branch: explore/web-based-agents
  Type: Research and investigation

  Enter your task or question for Claude (optional):
  Press Ctrl+E to open $EDITOR for multi-line input"

User: Presses Ctrl+E
Editor: Opens vim with temp file
User: Writes multi-line prompt, saves and exits
Wizard: Creates worktree
Wizard: Initializes project (no issue metadata)
Display: "✓ Initialized project 'Web Based Agents' on branch explore/web-based-agents
  ✓ Launching Claude in worktree..."
Claude: Launches in worktree
```

### Example: External Editor Usage
```
User: At prompt entry screen
User: Presses Ctrl+E
System: Creates temp file /tmp/huh-12345.md
System: Opens $EDITOR (vim/nano/code/etc)
User: Types multi-line content:
  "Research the landscape of web-based agent frameworks.
  Focus on:
  - Multi-agent coordination
  - Browser automation
  - Integration patterns"
User: Saves and exits editor
Wizard: Reads content from temp file
Wizard: Populates prompt field with content
User: Sees content in wizard
User: Presses Enter to continue
```

## Dependencies

### Prerequisites
- Task 010 (GitHub CLI validation) - MUST be complete
- Task 020 (issue listing) - MUST be complete
- Task 030 (issue validation and branch creation) - MUST be complete
- Work Unit 002 (branch name path and finalization) - COMPLETE (assumed)

### Depends On
- `initializeProject()` in shared.go (already handles issue parameter)
- `generateNewProjectPrompt()` in shared.go (already works correctly)
- `normalizeName()` for branch name display
- `getTypePrefix()` for branch prefix lookup

### Enables
- Complete GitHub issue integration workflow
- Future features that reference issue metadata
- PR creation with issue context

## Constraints

### Must Not
- **Break branch name path**: Changes must be backward compatible
- **Modify existing finalization**: Reuse existing logic, just pass issue parameter
- **Change file locations**: Issue context files must go in `.sow/project/context/`
- **Override user prompts**: Issue context supplements, doesn't replace user prompts

### Must Do
- **Show all context**: Display issue, branch, and type information
- **Support Ctrl+E**: External editor must work for multi-line prompts
- **Store metadata**: Issue information must persist in project state
- **Enhance success messages**: Show issue link when applicable

### Performance
- **No delays**: Context display is instant (no network calls)
- **Fast file creation**: Issue context file write is negligible
- **Prompt generation**: 3-layer prompt generation already optimized

## Notes

### Why Store Full Issue Object?

We store the complete issue in choices (not just number/title) because:
- `initializeProject()` needs title, body, URL for context file
- Finalization needs number for metadata
- Success message displays title
- Future features may need labels, assignee, etc.

### Issue Context vs User Prompt

The issue context is stored as a project input artifact, separate from the user's initial prompt. This distinction is important:

**Issue Context** (input artifact):
- Created from GitHub issue
- Contains issue title, body, URL
- Registered in project state
- Available to all orchestrator invocations

**User Prompt** (initial request):
- Optional user guidance
- Included in first Claude invocation
- Not persisted in project state
- Can provide focus or priorities

### Ctrl+E vs Ctrl+O

Original design specified Ctrl+O, but huh library uses Ctrl+E for external editor. This was discovered during library verification and updated in design docs. Implementation must use Ctrl+E.

### Claude Receives Issue Context

When Claude launches, the orchestrator reads project state and sees the `github_issue` input artifact. The orchestrator includes this context in the prompt:

```
You have access to these project inputs:
- context/issue-123.md (GitHub Issue #123: Add JWT authentication)
```

Claude can then read the issue file to understand requirements.

### Success Message Timing

The success messages are displayed **before** launching Claude. This ensures users see confirmation even if Claude launch takes time or if terminal output is redirected.

### Future: Continue Existing Project

When users continue an existing project created from an issue (Work Unit 004), the orchestrator will read the issue metadata from project state and can include it in the continuation prompt. This is already supported by the state structure.

### Testing Strategy

Due to the interactive nature of huh forms, full end-to-end testing of the prompt entry screen requires integration tests or manual testing. Unit tests focus on:
- Context string building logic
- Issue metadata handling
- Finalization with/without issue
- File creation and content

Integration tests would use actual huh forms but require terminal emulation or specialized testing tools.
