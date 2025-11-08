# Issue #69: Project Creation Workflow (Branch Name Path)

**URL**: https://github.com/jmgilman/sow/issues/69
**State**: OPEN

## Description

# Work Unit 002: Project Creation Workflow (Branch Name Path)

## Behavioral Goal (User Story)

**As a** sow user creating a new project,
**I need** an interactive wizard that guides me through branch name creation with type selection, real-time preview, and validation,
**so that** I can quickly create well-named projects without understanding git branch conventions, and immediately start working with Claude in the correct worktree.

### Success Criteria for Reviewers

1. User can complete the entire creation flow from type selection to Claude launch without errors
2. Branch name preview updates in real-time as user types, showing the exact branch that will be created
3. All validation rules are enforced (protected branches, empty names, existing projects)
4. External editor integration (Ctrl+E) works for multi-line prompt entry
5. Project is created in correct worktree with proper initialization
6. Claude launches with correctly structured 3-layer prompt

## Existing Code Context

This work unit builds on **Work Unit 001's foundation** which provides:
- State machine framework (WizardState type, state transitions)
- Name normalization utilities (`normalizeName` function)
- Branch validation logic (`isProtectedBranch`, `isValidBranchName`)
- Shared wizard structures and patterns

### Code to Leverage and Extend

**Existing Functions to Keep As-Is:**
- `cli/internal/sow/worktree.go:14-16` - `WorktreePath()`: Maps branch to worktree path
- `cli/internal/sow/worktree.go:21-87` - `EnsureWorktree()`: Idempotent worktree creation
- `cli/internal/sow/worktree.go:92-122` - `CheckUncommittedChanges()`: Validates no uncommitted changes
- `cli/internal/sow/context.go:30-111` - `NewContext()`: Creates context for repo or worktree

**Existing Logic to Extract and Share:**

From `cli/cmd/project/new.go`:
- Lines 362-398: `generateNewProjectPrompt()` - Creates 3-layer prompt structure
  - Extract to shared utility: `cli/cmd/project/shared.go`
  - Will be reused by GitHub issue path (work unit 003)

- Lines 402-421: `launchClaudeCode()` - Launches Claude with prompt and flags
  - Extract to shared utility: `cli/cmd/project/shared.go`
  - Will be reused by continuation path (work unit 004)

- Lines 148-196: Project initialization logic (create directories, write artifacts, call SDK)
  - Extract core logic to shared function: `initializeProject()`
  - Will be reused by GitHub issue path (work unit 003)

**After Extraction:** `cli/cmd/project/new.go` can be deleted entirely (replaced by wizard)

### Integration with State Machine

This work unit implements handlers for these states (from work unit 001):
- `StateTypeSelect` → Show type selection screen
- `StateNameEntry` → Show name entry with real-time preview
- `StatePromptEntry` → Show prompt entry with Ctrl+E support
- `StateComplete` → Trigger finalization flow

## Existing Documentation Context

### Interactive Wizard Design

**UX Flow Design** (`.sow/knowledge/designs/interactive-wizard-ux-flow.md`):
- Lines 175-279: Path 1B (Branch Name Path) - Complete user flow specification
- Lines 285-287: Real-time preview requirement - Critical UX feature
- Lines 216-231: Name normalization rules - Must follow exactly
- Lines 384-419: Validation rules - All must be enforced

**Technical Implementation** (`.sow/knowledge/designs/interactive-wizard-technical-implementation.md`):
- Lines 183-234: Example 2 - Name entry with validation and preview implementation
- Lines 236-276: Example 3 - Text area with external editor pattern
- Lines 523-566: Helper functions - Type configuration and normalization

**Library Verification** (`.sow/knowledge/designs/huh-library-verification.md`):
- Lines 113-151: Real-time preview capability using `huh.NewNote()` with `DescriptionFunc`
- Lines 183-248: External editor support - **Uses Ctrl+E not Ctrl+O**
- Lines 88-103: Validation patterns with inline error display

### Project Type System

**Discovery Analysis** (`.sow/project/discovery/analysis.md`):
- Lines 98-107: Project type configuration table with prefixes
- Lines 44-69: Core infrastructure that must be reused

The project type prefix mapping is:
- Standard → `feat/`
- Exploration → `explore/`
- Design → `design/`
- Breakdown → `breakdown/`

This is NOT configurable per-project; it's part of the type definition registered in `cli/internal/projects/{type}/{type}.go`.

## Implementation Scope

### 1. Type Selection Screen

**Create:** `handleTypeSelect()` function in wizard

**Requirements:**
- Display all four project types with descriptions
- Use `huh.NewSelect[string]()` with options
- Store selection in `w.choices["type"]`
- Map type names to type keys: "standard", "exploration", "design", "breakdown"
- Handle cancel (transition to `StateCancelled`)

**Display format:**
```
Standard - Feature work and bug fixes
Exploration - Research and investigation
Design - Architecture and design documents
Breakdown - Decompose work into tasks
Cancel
```

### 2. Name Entry Screen with Real-Time Preview

**Create:** `handleNameEntry()` function in wizard

**Requirements:**
- Single text input field for project name
- Placeholder: "e.g., Web Based Agents"
- Real-time branch preview using `huh.NewNote()` with `DescriptionFunc`
- Preview format: `<prefix>/<normalized-name>`
- Inline validation on submit

**Real-Time Preview Implementation:**
```go
huh.NewNote().
    Title("Branch Preview").
    DescriptionFunc(func() string {
        if name == "" {
            return fmt.Sprintf("%s/<project-name>", prefix)
        }
        normalized := normalizeName(name)  // From work unit 001
        return fmt.Sprintf("%s/%s", prefix, normalized)
    }, &name)  // Bind to name variable for auto-updates
```

**Validation Rules (on submit):**
1. Not empty or whitespace-only
2. After normalization and prefix addition, not a protected branch (main/master)
3. Valid git branch name (uses `isValidBranchName` from work unit 001)

**Error Messages (inline):**
- Empty: "project name cannot be empty"
- Protected: "cannot use protected branch name"
- Invalid: "invalid characters in branch name"

### 3. Branch State Validation

**Create:** `checkBranchState()` helper function

**Check:**
- Does branch exist? (check git branches)
- Does worktree exist for branch? (check `.sow/worktrees/<branch>`)
- Does worktree have project? (check `.sow/worktrees/<branch>/.sow/project/state.yaml`)

**Error Handling:**
If project already exists on branch:
```
Error: Branch '<branch>' already has a project

To continue this project:
  Select "Continue existing project" from the main menu

To create a different project:
  Choose a different project name
```

Use `showError()` to display, then stay in `StateNameEntry` to retry.

### 4. Prompt Entry Screen

**Create:** `handlePromptEntry()` function in wizard

**Requirements:**
- Multi-line text area using `huh.NewText()`
- Title: "Enter your task or question for Claude (optional):"
- Context display: "Type: <type>\nBranch: <prefix>/<normalized-name>"
- Description: "Press Ctrl+E to open $EDITOR for multi-line input"
- Character limit: 5000-10000 (choose 10000 for flexibility)
- Enable external editor: `.WithEditor(true)` or `.EditorExtension(".md")`
- Optional - user can skip (leave empty)

**Implementation Pattern:**
```go
huh.NewText().
    Title("Enter your task or question for Claude (optional):").
    Description(fmt.Sprintf("Type: %s\nBranch: %s\n\nPress Ctrl+E to open $EDITOR",
        projectType, branchName)).
    CharLimit(10000).
    Value(&prompt).
    EditorExtension(".md")
```

### 5. Finalization Flow

**Create:** `finalize()` function in wizard (called when state == `StateComplete`)

**Steps:**

1. **Conditional Uncommitted Changes Check**
   ```go
   // Only check if current branch == target branch
   currentBranch, _ := ctx.Git().CurrentBranch()
   if currentBranch == selectedBranch {
       if err := sow.CheckUncommittedChanges(ctx); err != nil {
           return fmt.Errorf("repository has uncommitted changes\n\n"+
               "You are currently on branch '%s'.\n"+
               "Creating a worktree requires switching to a different branch first.\n\n"+
               "To fix:\n"+
               "  Commit: git add . && git commit -m \"message\"\n"+
               "  Or stash: git stash", currentBranch)
       }
   }
   ```

2. **Ensure Worktree Exists**
   ```go
   worktreePath := sow.WorktreePath(ctx.RepoRoot(), selectedBranch)
   if err := sow.EnsureWorktree(ctx, worktreePath, selectedBranch); err != nil {
       return fmt.Errorf("failed to create worktree: %w", err)
   }
   ```

3. **Initialize Project**
   - Create worktree context
   - Validate worktree is initialized (has `.sow/` directory)
   - Create `.sow/project/` and `.sow/project/context/` directories
   - Call `state.Create(worktreeCtx, selectedBranch, projectName, nil)` (no initial inputs for branch path)
   - Project name comes from `w.choices["name"]`

4. **Generate 3-Layer Prompt**
   - Use extracted `generateNewProjectPrompt()` function
   - Pass project and user's initial prompt (from `w.choices["prompt"]`)
   - Structure: Base Orchestrator + Project Type Orchestrator + Initial State + User Prompt

5. **Launch Claude**
   - Use extracted `launchClaudeCode()` function
   - Launch in worktree directory
   - Pass generated prompt

**Success Message:**
```
✓ Initialized project '<name>' on branch <branch>
✓ Launching Claude in worktree...
```

### 6. Shared Utilities

**Create:** `cli/cmd/project/shared.go`

**Functions to Extract:**

```go
// generateNewProjectPrompt creates the 3-layer prompt for new projects
func generateNewProjectPrompt(proj *state.Project, initialPrompt string) (string, error)

// launchClaudeCode launches Claude Code CLI with prompt
func launchClaudeCode(cmd *cobra.Command, ctx *sow.Context, prompt string, claudeFlags []string) error

// initializeProject creates project structure in worktree
func initializeProject(worktreeCtx *sow.Context, branch, name string, initialInputs map[string][]projschema.ArtifactState) (*state.Project, error)
```

## Acceptance Criteria

### Functional Requirements

1. **Type Selection Works**
   - All four project types displayed with descriptions
   - Selection stored in wizard choices
   - Cancel returns to entry screen

2. **Name Entry with Preview**
   - As user types, preview updates showing `<prefix>/<normalized-name>`
   - Preview updates happen in real-time without lag
   - Empty input shows placeholder: `<prefix>/<project-name>`

3. **Validation Catches Errors**
   - Empty names rejected with inline error
   - Protected branch names (main, master) rejected
   - Invalid git characters rejected
   - Validation errors display inline below input field

4. **Branch State Checked**
   - Error shown if branch already has project
   - Error message guides user to continuation path
   - User can retry with different name

5. **Prompt Entry Functional**
   - Multi-line text area accepts input
   - User can leave empty and continue
   - No character limit warnings if under 10000

6. **External Editor Works**
   - Pressing Ctrl+E opens $EDITOR (or falls back to nano)
   - Content from editor populates prompt field
   - Editor uses .md file extension for syntax highlighting

7. **Uncommitted Changes Check**
   - Only runs when current branch == target branch
   - Clear error message if check fails
   - No false positives from untracked files

8. **Project Created**
   - Worktree created at `.sow/worktrees/<branch>/`
   - Project initialized with correct type
   - `state.yaml` file exists and valid
   - Success message shows project name and branch

9. **Claude Launched**
   - Claude launches in worktree directory (not main repo)
   - Prompt includes all 3 layers plus user's initial prompt
   - User can start working immediately

### Non-Functional Requirements

1. **Performance**: Real-time preview updates within 50ms of keystroke
2. **Usability**: All error messages include actionable remediation steps
3. **Reliability**: Wizard handles interruption (Ctrl+C) gracefully without corrupting state

## Testing Requirements

### Unit Tests

**Test:** Type configuration mapping
```go
func TestGetTypePrefix(t *testing.T) {
    tests := []struct{type, expectedPrefix string}{
        {"standard", "feat/"},
        {"exploration", "explore/"},
        {"design", "design/"},
        {"breakdown", "breakdown/"},
    }
    // Verify correct prefix returned for each type
}
```

**Test:** Name normalization (inherited from work unit 001, verify integration)
```go
func TestNameEntryNormalization(t *testing.T) {
    // Test that handleNameEntry correctly applies normalizeName()
    // Test preview generation with various inputs
}
```

**Test:** Branch state checking
```go
func TestCheckBranchState(t *testing.T) {
    // Test detection of existing projects on branches
    // Test handling of branches without projects
    // Test handling of non-existent branches
}
```

**Test:** Conditional uncommitted changes check
```go
func TestShouldCheckUncommittedChanges(t *testing.T) {
    // Test returns true when current == target
    // Test returns false when current != target
}
```

### Integration Tests

**Test:** Full creation flow
```go
func TestBranchNameCreationFlow(t *testing.T) {
    // Simulate: TypeSelect → NameEntry → PromptEntry → Complete
    // Verify state transitions at each step
    // Verify choices populated correctly
    // Mock worktree creation and project initialization
    // Verify Claude launch called with correct arguments
}
```

**Test:** Validation error recovery
```go
func TestValidationErrorRecovery(t *testing.T) {
    // Enter protected branch name
    // Verify error shown
    // Verify stays in NameEntry state
    // Verify can retry with valid name
}
```

**Test:** Branch conflict handling
```go
func TestBranchConflictHandling(t *testing.T) {
    // Create project on branch
    // Try to create another project on same branch
    // Verify error shown with guidance
    // Verify stays in NameEntry state for retry
}
```

### Manual Testing Scenarios

1. **Happy Path:**
   - Run `sow project`
   - Select "Create new project"
   - Select "From branch name"
   - Select "Exploration" type
   - Enter "Web Based Agents"
   - Observe preview: `explore/web-based-agents`
   - Press Ctrl+E, write prompt in editor, save and exit
   - Verify worktree created
   - Verify Claude launches with prompt

2. **Validation Errors:**
   - Try empty name → see inline error
   - Try "main" → see protected branch error
   - Try "existing-project-branch" → see conflict error

3. **External Editor:**
   - Test with $EDITOR set to vim
   - Test with $EDITOR set to nano
   - Test with $EDITOR unset (should use nano)
   - Test with $EDITOR set to code (VS Code)

4. **Uncommitted Changes:**
   - Have uncommitted changes on current branch
   - Try to create project on current branch → see error
   - Create project on different branch → should work

5. **Empty Prompt:**
   - Skip prompt entry (leave empty)
   - Verify project still created
   - Verify Claude still launches

## Technical Details

### File Structure

```
cli/cmd/project/
├── root.go                 # Entry point (from work unit 001)
├── wizard.go              # Wizard struct and Run() (from work unit 001)
├── handlers_creation.go   # NEW: This work unit's handlers
├── shared.go              # NEW: Extracted shared utilities
└── new.go                 # DELETE after extraction
```

### Implementation in handlers_creation.go

```go
package project

// handleTypeSelect shows project type selection screen
func (w *Wizard) handleTypeSelect() error {
    var selectedType string

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("What type of project?").
                Options(
                    huh.NewOption("Standard - Feature work and bug fixes", "standard"),
                    huh.NewOption("Exploration - Research and investigation", "exploration"),
                    huh.NewOption("Design - Architecture and design documents", "design"),
                    huh.NewOption("Breakdown - Decompose work into tasks", "breakdown"),
                    huh.NewOption("Cancel", "cancel"),
                ).
                Value(&selectedType),
        ),
    )

    if err := form.Run(); err != nil {
        if errors.Is(err, huh.ErrUserAborted) {
            w.state = StateCancelled
            return nil
        }
        return err
    }

    if selectedType == "cancel" {
        w.state = StateCancelled
        return nil
    }

    w.choices["type"] = selectedType
    w.state = StateNameEntry
    return nil
}

// handleNameEntry shows name entry with real-time preview
func (w *Wizard) handleNameEntry() error {
    var name string
    projectType := w.choices["type"].(string)
    prefix := getTypePrefix(projectType)

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().
                Title("Enter project name:").
                Placeholder("e.g., Web Based Agents").
                Value(&name).
                Validate(func(s string) error {
                    if strings.TrimSpace(s) == "" {
                        return fmt.Errorf("project name cannot be empty")
                    }

                    normalized := normalizeName(s)
                    branchName := fmt.Sprintf("%s/%s", prefix, normalized)

                    if isProtectedBranch(branchName) {
                        return fmt.Errorf("cannot use protected branch name")
                    }

                    if err := isValidBranchName(branchName); err != nil {
                        return err
                    }

                    return nil
                }),

            // Real-time preview
            huh.NewNote().
                Title("Branch Preview").
                DescriptionFunc(func() string {
                    if name == "" {
                        return fmt.Sprintf("%s/<project-name>", prefix)
                    }
                    normalized := normalizeName(name)
                    return fmt.Sprintf("%s/%s", prefix, normalized)
                }, &name),
        ),
    )

    if err := form.Run(); err != nil {
        if errors.Is(err, huh.ErrUserAborted) {
            w.state = StateTypeSelect  // Go back to type selection
            return nil
        }
        return err
    }

    // Check if branch already has project
    normalized := normalizeName(name)
    branchName := fmt.Sprintf("%s/%s", prefix, normalized)

    state, err := checkBranchState(w.ctx, branchName)
    if err != nil {
        return err
    }

    if state.ProjectExists {
        showError(fmt.Sprintf(
            "Branch '%s' already has a project\n\n"+
            "To continue this project:\n"+
            "  Select \"Continue existing project\" from the main menu\n\n"+
            "To create a different project:\n"+
            "  Choose a different project name",
            branchName))
        return nil  // Stay in current state to retry
    }

    w.choices["name"] = name
    w.choices["branch"] = branchName
    w.state = StatePromptEntry
    return nil
}

// handlePromptEntry shows prompt entry with external editor support
func (w *Wizard) handlePromptEntry() error {
    var prompt string

    projectType := w.choices["type"].(string)
    branchName := w.choices["branch"].(string)

    contextInfo := fmt.Sprintf("Type: %s\nBranch: %s\n\nPress Ctrl+E to open $EDITOR for multi-line input",
        projectType, branchName)

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewText().
                Title("Enter your task or question for Claude (optional):").
                Description(contextInfo).
                CharLimit(10000).
                Value(&prompt).
                EditorExtension(".md"),
        ),
    )

    if err := form.Run(); err != nil {
        if errors.Is(err, huh.ErrUserAborted) {
            w.state = StateNameEntry  // Go back to name entry
            return nil
        }
        return err
    }

    w.choices["prompt"] = prompt
    w.state = StateComplete
    return nil
}

// Helper: getTypePrefix returns branch prefix for project type
func getTypePrefix(projectType string) string {
    prefixes := map[string]string{
        "standard":    "feat/",
        "exploration": "explore/",
        "design":      "design/",
        "breakdown":   "breakdown/",
    }
    return prefixes[projectType]
}

// Helper: checkBranchState checks if branch/worktree/project exists
type BranchState struct {
    BranchExists   bool
    WorktreeExists bool
    ProjectExists  bool
}

func checkBranchState(ctx *sow.Context, branchName string) (*BranchState, error) {
    state := &BranchState{}

    // Check if branch exists
    branches, err := ctx.Git().Branches()
    if err != nil {
        return nil, err
    }
    for _, b := range branches {
        if b == branchName {
            state.BranchExists = true
            break
        }
    }

    // Check if worktree exists
    worktreePath := sow.WorktreePath(ctx.RepoRoot(), branchName)
    if _, err := os.Stat(worktreePath); err == nil {
        state.WorktreeExists = true

        // Check if project exists in worktree
        projectPath := filepath.Join(worktreePath, ".sow", "project", "state.yaml")
        if _, err := os.Stat(projectPath); err == nil {
            state.ProjectExists = true
        }
    }

    return state, nil
}
```

### Type Configuration

The project type prefixes are **NOT** stored in a configuration file. They are implicit in the project type registration:

- Standard project registered in `cli/internal/projects/standard/standard.go`
- Branch conventions use `feat/` prefix
- No explicit prefix configuration in the code

For the wizard, we maintain the mapping directly:
```go
var projectTypePrefixes = map[string]string{
    "standard":    "feat/",
    "exploration": "explore/",
    "design":      "design/",
    "breakdown":   "breakdown/",
}
```

This mapping is **derived from convention**, not configuration. If we later need to make this configurable, it should be added to the project type config builder.

## Examples

### Example 1: Successful Creation

```
$ sow project

[Entry Screen]
→ User selects "Create new project"

[Source Selection]
→ User selects "From branch name"

[Type Selection]
What type of project?
  ○ Standard - Feature work and bug fixes
  ● Exploration - Research and investigation
  ○ Design - Architecture and design documents
  ○ Breakdown - Decompose work into tasks
  ○ Cancel

[Name Entry]
Enter project name:
┌────────────────────────────────────────┐
│ Web Based Agents                       │
└────────────────────────────────────────┘

Branch Preview: explore/web-based-agents
[User presses Enter]

[Prompt Entry]
Type: Exploration
Branch: explore/web-based-agents

Enter your task or question for Claude (optional):
Press Ctrl+E to open $EDITOR for multi-line input

┌────────────────────────────────────────┐
│ Research the landscape of web-based    │
│ agent frameworks and compare their     │
│ architectures.                         │
└────────────────────────────────────────┘
[User presses Enter]

[Finalization]
✓ Initialized project 'Web Based Agents' on branch explore/web-based-agents
✓ Launching Claude in worktree...

[Claude Code launches]
```

### Example 2: Validation Error Recovery

```
[Name Entry]
Enter project name:
┌────────────────────────────────────────┐
│ main                                   │
└────────────────────────────────────────┘

Branch Preview: feat/main
⚠ cannot use protected branch name

[User corrects]
┌────────────────────────────────────────┐
│ authentication                         │
└────────────────────────────────────────┘

Branch Preview: feat/authentication
[Validation passes, continues to prompt entry]
```

### Example 3: External Editor Usage

```
[Prompt Entry]
Type: Design
Branch: design/api-architecture

Enter your task or question for Claude (optional):
Press Ctrl+E to open $EDITOR for multi-line input

┌────────────────────────────────────────┐
│                                        │
└────────────────────────────────────────┘

[User presses Ctrl+E]

[nano opens (or $EDITOR)]
Design a REST API for the new microservice with these requirements:
- OAuth2 authentication
- Rate limiting per API key
- Pagination for list endpoints
- OpenAPI 3.0 documentation
[User saves and exits]

[Back to wizard, prompt populated]
┌────────────────────────────────────────┐
│ Design a REST API for the new micro…  │
└────────────────────────────────────────┘

[User presses Enter to continue]
```

## Dependencies

### Upstream Dependencies (Must Complete First)

- **Work Unit 001**: Wizard Foundation and State Machine
  - Provides: `WizardState` enum, state transition framework
  - Provides: `normalizeName()`, `isValidBranchName()`, `isProtectedBranch()`
  - Provides: Wizard struct with `state` and `choices` fields
  - Provides: Basic error handling patterns (`showError()`)

### Downstream Dependencies (Will Use This Work Unit)

- **Work Unit 003**: GitHub Issue Integration
  - Will reuse: `generateNewProjectPrompt()` from `shared.go`
  - Will reuse: `launchClaudeCode()` from `shared.go`
  - Will reuse: `initializeProject()` from `shared.go`
  - Will reuse: Type selection screen pattern

- **Work Unit 005**: Validation and Testing Utilities
  - May enhance: Branch name validation with additional checks
  - May enhance: Error messages with additional context

## Constraints

### Performance Requirements

- Real-time preview must update within 50ms of keystroke
- Name normalization must be fast enough to not cause preview lag
- Use `DescriptionFunc` binding feature for automatic caching

### Security Considerations

- Never expose git credentials in error messages
- Validate all user input before executing git commands
- Use parameterized git commands (no string concatenation)
- External editor: ensure temp files are properly cleaned up

### Compatibility Requirements

- Must work with git 2.17+ (minimum worktree support)
- Must work with or without `gh` CLI (this path doesn't require it)
- Must work on macOS, Linux, Windows (use `filepath.Join` for paths)
- Terminal must support ANSI colors (huh requirement)

### Git Conventions

- Branch names must be valid git refs (enforced by validation)
- Worktree paths must match branch names exactly (slashes preserved)
- Protected branches (main, master) cannot have projects created on them
- Current branch same as target requires uncommitted changes check

### What NOT to Do

- ❌ Don't modify existing `worktree.go` functions - they're stable and tested
- ❌ Don't create new project type registration - use existing types
- ❌ Don't modify project type prefix conventions - they're part of the type definition
- ❌ Don't add configurable prefixes yet - keep it simple, convention-based
- ❌ Don't launch Claude from main repo - always use worktree path
- ❌ Don't skip validation - all rules must be enforced
- ❌ Don't use Ctrl+O for editor - huh uses Ctrl+E (see library verification doc)

## Relevant Inputs

### Design Documents
- `.sow/knowledge/designs/interactive-wizard-ux-flow.md` - Complete UX specification for branch name path (lines 175-279)
- `.sow/knowledge/designs/interactive-wizard-technical-implementation.md` - Implementation patterns and code examples (lines 183-276, 523-599)
- `.sow/knowledge/designs/huh-library-verification.md` - Library capabilities and limitations (lines 113-151 real-time preview, lines 183-248 external editor)

### Existing Code to Understand
- `cli/cmd/project/new.go` - Current implementation to extract logic from (lines 362-421)
- `cli/internal/sow/worktree.go` - Worktree management functions to reuse as-is (lines 14-122)
- `cli/internal/sow/context.go` - Context creation for main repo and worktrees (lines 30-111)
- `cli/internal/projects/standard/standard.go` - Example project type registration (lines 1-26)
- `cli/internal/projects/exploration/exploration.go` - Example project type with different phases (lines 1-26)

### Discovery and Analysis
- `.sow/project/discovery/analysis.md` - Codebase context and project type table (lines 44-69, 98-107)

### Dependencies from Work Unit 001
- `.sow/project/context/tasks/001-wizard-foundation.md` - State machine design, validation utilities, shared patterns

### Supporting Infrastructure
- `cli/internal/sdks/project/state/registry.go` - Project type registry pattern (lines 24-29)
- `cli/internal/sdks/project/templates/` - Prompt template rendering system

## Notes

### Critical Implementation Details

1. **Real-Time Preview**: The `DescriptionFunc` binding is what makes preview work. Without the `&name` binding parameter, preview won't update.

2. **External Editor Keybinding**: Design docs originally said Ctrl+O, but library verification found it's actually Ctrl+E. Use Ctrl+E in all user-facing text.

3. **Conditional Uncommitted Changes**: Only check when `current == target` branch. This is a git worktree limitation, not a sow design choice.

4. **Name vs Branch**: User enters a "name", wizard normalizes it to a "branch name". Store both in choices for clarity.

5. **Type Prefixes**: These are convention-based, not configurable. Hard-code the mapping for now. If we need configurability later, add to project type config builder.

### Testing Strategy

- Unit tests focus on pure functions (normalization, validation, state checking)
- Integration tests focus on state transitions and data flow
- Manual tests focus on UX and external integrations (editor, git, Claude)

### Performance Optimization

- `normalizeName()` is called on every keystroke for preview
- Keep it simple (string operations only, no I/O)
- `DescriptionFunc` has automatic caching - trust it

### Error Handling Philosophy

- All errors should guide user to resolution
- Never leave user stuck - always provide path forward
- Validation errors should be immediate and inline
- Catastrophic errors (git failures) should be clear about what went wrong
