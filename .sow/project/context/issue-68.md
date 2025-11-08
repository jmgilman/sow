# Issue #68: Wizard Foundation and State Machine

**URL**: https://github.com/jmgilman/sow/issues/68
**State**: OPEN

## Description

# Work Unit 001: Wizard Foundation and State Machine

**Size**: Large (2-3 days)
**Dependencies**: None - foundational work
**Status**: Ready for implementation

---

## 1. Behavioral Goal (User Story)

**As a** sow user
**I need** a single `sow project` command that launches an interactive wizard
**So that** I can create or continue projects without memorizing complex flags and command variations

### Success Criteria

- Running `sow project` displays an interactive entry screen with three options: create, continue, or cancel
- User can navigate with arrow keys, select with Enter, and cancel with Esc
- Wizard maintains state correctly as user moves through screens
- Helper functions normalize project names and provide type configuration correctly
- Error messages display in a formatted, user-friendly manner
- Loading spinners appear for long-running operations
- All state transitions work correctly (can be tested with stub handlers)

---

## 2. Existing Code Context

### Explanatory Context

This work unit establishes the foundation for an interactive wizard that **replaces** the existing flag-based `sow project new` and `sow project continue` commands. The current implementation (in `new.go` and `continue.go`) contains critical logic for:

1. **Project initialization** - Creating project state files, managing worktrees
2. **Prompt generation** - Building 3-layer prompts (orchestrator + type + state + user)
3. **Claude launch** - Executing Claude Code CLI with proper context

**Before** deleting `new.go` and `continue.go`, we must **extract** their reusable logic into shared utility functions. This work unit creates the infrastructure (state machine, wizard command, helpers) that all subsequent wizard screens will build upon.

### Key Files Reference

**Will be DELETED** (after extracting logic):
- `cli/cmd/project/new.go` (440 lines)
  - Lines 86-210: `runNew()` - orchestrates creation flow
  - Lines 213-259: `determineBranchAndDescription()` - routing logic
  - Lines 261-296: `handleIssueScenarioNew()` - GitHub issue integration
  - Lines 298-342: `handleBranchScenarioNew()` - branch creation
  - Lines 362-398: `generateNewProjectPrompt()` - 3-layer prompt builder
  - Lines 400-421: `launchClaudeCode()` - Claude CLI launcher

- `cli/cmd/project/continue.go` (196 lines)
  - Lines 48-121: `runContinue()` - orchestrates continuation flow
  - Lines 167-196: `generateContinuePrompt()` - 3-layer prompt builder

**Will be REUSED** (no changes):
- `cli/internal/sow/worktree.go`
  - Lines 14-16: `WorktreePath()` - maps branch to worktree path
  - Lines 21-87: `EnsureWorktree()` - idempotent worktree creation
  - Lines 92-122: `CheckUncommittedChanges()` - validates clean repo state

- `cli/internal/sow/github.go`
  - GitHub operations: `GetIssue()`, `GetLinkedBranches()`, `CreateLinkedBranch()`, `ListIssues()`

- `cli/internal/sow/context.go`
  - Context management for main repo and worktrees

- `cli/internal/sdks/project/state/registry.go`
  - Project type system with configs and phases

**Will be CREATED**:
- `cli/cmd/project/wizard.go` - Wizard command and state machine
- `cli/cmd/project/wizard_state.go` - State handlers
- `cli/cmd/project/wizard_helpers.go` - Helper functions (normalization, type config)
- `cli/cmd/project/shared.go` - Shared utilities extracted from new.go/continue.go

---

## 3. Existing Documentation Context

### Design Documents (Registered as Inputs)

**UX Flow Design** (`.sow/knowledge/designs/interactive-wizard-ux-flow.md` - 728 lines):
- Complete user journeys for all three workflows
- All screen layouts and navigation flows
- Comprehensive validation rules and error messages
- **Key insight** (line 32): "Current branch is irrelevant" - wizard operates on target branches only
- **Entry screen** (lines 39-55): Single screen with create/continue/cancel options
- **Type configuration** (lines 129-135): Maps project types to branch prefixes

**Technical Implementation** (`.sow/knowledge/designs/interactive-wizard-technical-implementation.md` - 899 lines):
- State machine architecture (lines 20-82)
- huh library integration patterns (lines 105-140)
- Helper function implementations (lines 523-646)
- Testing strategy (lines 700-803)
- **Name normalization** (lines 571-599): Algorithm for sanitizing project names
- **Project type config** (lines 525-565): Data structure for type-to-prefix mapping
- **Loading indicators** (lines 865-881): Spinner integration for async operations

**Library Verification** (`.sow/knowledge/designs/huh-library-verification.md` - 762 lines):
- Confirms huh library supports all design requirements
- **Critical finding** (lines 195-232): External editor uses **Ctrl+E** not Ctrl+O
- Form structure patterns (lines 612-657)
- Error handling patterns (lines 660-695)

### Relevant ADRs and Architecture

None directly applicable - this is new functionality. However, the wizard follows established patterns:
- **Registry pattern** (from project type system): Extensible type configuration
- **3-layer prompt system**: Orchestrator + project type + state prompts
- **Worktree isolation**: Operations don't affect current branch

---

## 4. Implementation Scope

### What Needs to Be Built

#### 4.1 Dependency Addition
- Add `github.com/charmbracelet/huh` to `go.mod` via `go get`
- Add `github.com/charmbracelet/huh/spinner` for loading indicators
- Version: Use latest stable version

#### 4.2 Command Structure
Create new `wizard.go` with:
- `newWizardCmd()` - Cobra command for `sow project`
- `runWizard()` - Entry point that initializes wizard and calls `Run()`
- Replace current `newProjectCmd()` in `cli/cmd/project/project.go` to use wizard

#### 4.3 State Machine (wizard_state.go)
```go
type WizardState string

const (
    StateEntry          WizardState = "entry"
    StateCreateSource   WizardState = "create_source"
    StateIssueSelect    WizardState = "issue_select"
    StateTypeSelect     WizardState = "type_select"
    StateNameEntry      WizardState = "name_entry"
    StatePromptEntry    WizardState = "prompt_entry"
    StateProjectSelect  WizardState = "project_select"
    StateContinuePrompt WizardState = "continue_prompt"
    StateComplete       WizardState = "complete"
    StateCancelled      WizardState = "cancelled"
)

type Wizard struct {
    state   WizardState
    ctx     *sow.Context
    choices map[string]interface{}
}

func (w *Wizard) Run() error
func (w *Wizard) handleState() error
func (w *Wizard) finalize() error
```

Implement state handler dispatcher pattern per technical design (lines 61-81).

#### 4.4 Entry Screen Handler (wizard_state.go)
Implement `handleEntry()` using huh.NewSelect:
- Options: "Create new project", "Continue existing project", "Cancel"
- Store selection in `choices["action"]`
- Transition to appropriate next state
- Reference: Technical design lines 145-181

#### 4.5 Helper Functions (wizard_helpers.go)

**Name Normalization Function**:
```go
func normalizeName(name string) string
```
Algorithm (per technical design lines 571-599):
1. Trim whitespace
2. Convert to lowercase
3. Replace spaces with hyphens
4. Remove invalid characters (keep only a-z, 0-9, -, _)
5. Collapse multiple consecutive hyphens
6. Remove leading/trailing hyphens

**Project Type Configuration**:
```go
type ProjectTypeConfig struct {
    Prefix      string
    Description string
}

var projectTypes map[string]ProjectTypeConfig

func getTypePrefix(projectType string) string
func getTypeOptions() []huh.Option[string]
```

Data (from UX design lines 129-135):
- standard → "feat/" - Feature work and bug fixes
- exploration → "explore/" - Research and investigation
- design → "design/" - Architecture and design documents
- breakdown → "breakdown/" - Decompose work into tasks

**Branch Preview Function**:
```go
func previewBranchName(projectType, name string) string
```
Combines prefix and normalized name: `<prefix>/<normalized-name>`

**Error Display Utilities**:
```go
func showError(message string) error
func showErrorWithOptions(message string, options map[string]string) (string, error)
```
Use huh forms to display errors with acknowledgment. Reference: Technical design lines 654-695.

#### 4.6 Loading Indicator Utilities (wizard_helpers.go)
```go
func withSpinner(title string, action func() error) error
```
Wraps spinner.New() for consistent loading indicators across wizard.
Reference: Technical design lines 865-881.

#### 4.7 Cancel and Abort Handling
- Detect `huh.ErrUserAborted` in all form handlers
- Set state to `StateCancelled` and return cleanly
- Exit wizard without error (user chose to cancel)

### What Needs to Be Migrated/Extracted

Before deleting `new.go` and `continue.go`, extract to new `shared.go`:

#### From new.go:
**Project Initialization Logic** (lines 149-196):
```go
func initializeProject(
    ctx *sow.Context,
    branch, description string,
    issue *sow.Issue,
) (*state.Project, error)
```
Creates:
- `.sow/project` and `.sow/project/context` directories
- Issue context file if issue provided
- Initial project state with inputs

**Prompt Generation** (lines 362-398):
```go
func generateNewProjectPrompt(
    proj *state.Project,
    userPrompt string,
) (string, error)
```
Builds 3-layer prompt:
1. Base orchestrator introduction
2. Project type orchestrator prompt
3. Initial state prompt
4. User's initial request (if provided)

**Claude Launch** (lines 400-421):
```go
func launchClaudeCode(
    ctx context.Context,
    sowCtx *sow.Context,
    prompt string,
    claudeFlags []string,
) error
```
Executes Claude CLI with:
- Prompt as first argument
- Additional flags appended
- Working directory = worktree path
- Stdin/Stdout/Stderr inherited

#### From continue.go:
**Continue Prompt Generation** (lines 167-196):
```go
func generateContinuePrompt(proj *state.Project) (string, error)
```
Similar to new prompt but uses current state instead of initial state.

**Note**: The branch scenario handlers (`handleBranchScenarioNew`, etc.) are NOT extracted - wizard replaces this routing logic entirely.

### What Can Be Reused As-Is

**No modifications needed** for:
- `cli/internal/sow/worktree.go` - All functions used directly
- `cli/internal/sow/github.go` - All functions used directly
- `cli/internal/sow/context.go` - Context creation unchanged
- `cli/internal/sdks/project/state/` - Project SDK unchanged
- `cli/internal/prompts/` - Template rendering unchanged

---

## 5. Acceptance Criteria

### AC1: Dependency Management
- [ ] `go.mod` contains `github.com/charmbracelet/huh` with version
- [ ] `go.mod` contains `github.com/charmbracelet/huh/spinner`
- [ ] `go mod tidy` runs without errors
- [ ] No dependency conflicts

### AC2: Command Integration
- [ ] `sow project` command launches wizard (no flags required)
- [ ] Old `sow project new` and `sow project continue` commands removed
- [ ] Command appears in `sow project --help`
- [ ] No compilation errors

### AC3: Entry Screen Functionality
- [ ] Running `sow project` displays entry screen
- [ ] Screen shows three options: Create, Continue, Cancel
- [ ] Arrow keys navigate options
- [ ] Enter key selects option
- [ ] Esc or Cancel selection exits cleanly without error
- [ ] Title displays correctly

### AC4: State Machine Operations
- [ ] Wizard initializes in `StateEntry`
- [ ] State transitions work correctly based on user selections
- [ ] Selecting "Cancel" transitions to `StateCancelled`
- [ ] Selecting "Create" transitions to `StateCreateSource`
- [ ] Selecting "Continue" transitions to `StateProjectSelect`
- [ ] `StateCancelled` exits wizard without error
- [ ] State handler dispatcher routes to correct handler

### AC5: Name Normalization
Test cases (all must pass):
- [ ] "Web Based Agents" → "web-based-agents"
- [ ] "API V2" → "api-v2"
- [ ] "feature--name" → "feature-name"
- [ ] "-leading-trailing-" → "leading-trailing"
- [ ] "With!Invalid@Chars#" → "withinvalidchars"
- [ ] "UPPERCASE" → "uppercase"
- [ ] "  spaces  " → "spaces"

### AC6: Type Configuration
- [ ] `projectTypes` map contains all four types
- [ ] Each type has correct prefix (feat/, explore/, design/, breakdown/)
- [ ] Each type has correct description
- [ ] `getTypePrefix()` returns correct prefix for each type
- [ ] `getTypePrefix()` defaults to "feat/" for unknown types
- [ ] `getTypeOptions()` returns huh-compatible options

### AC7: Error Display
- [ ] `showError()` displays formatted error message
- [ ] User can acknowledge error with Enter
- [ ] Error screen uses huh form structure
- [ ] Error message is readable and properly formatted

### AC8: Loading Indicator
- [ ] `withSpinner()` displays spinner during action
- [ ] Spinner shows provided title
- [ ] Spinner disappears when action completes
- [ ] Errors from action are propagated correctly

### AC9: Shared Functions Extracted
- [ ] `initializeProject()` exists in shared.go
- [ ] `generateNewProjectPrompt()` exists in shared.go
- [ ] `generateContinuePrompt()` exists in shared.go
- [ ] `launchClaudeCode()` exists in shared.go
- [ ] All functions have correct signatures
- [ ] Functions work identically to original implementations

### AC10: Code Cleanup
- [ ] `new.go` deleted (after extraction complete)
- [ ] `continue.go` deleted (after extraction complete)
- [ ] No references to deleted files remain
- [ ] No dead code or unused imports

---

## 6. Testing Requirements

### Unit Tests Required

**File: `wizard_helpers_test.go`**

Test name normalization:
```go
func TestNormalizeName(t *testing.T)
```
All test cases from AC5 plus edge cases:
- Empty string
- Only spaces
- Only special characters
- Very long names
- Unicode characters

Test type configuration:
```go
func TestGetTypePrefix(t *testing.T)
func TestGetTypeOptions(t *testing.T)
```
Verify all types return correct prefixes and descriptions.

Test branch preview:
```go
func TestPreviewBranchName(t *testing.T)
```
Verify correct combination of prefix and normalized name.

**File: `shared_test.go`**

Test prompt generation:
```go
func TestGenerateNewProjectPrompt(t *testing.T)
func TestGenerateContinuePrompt(t *testing.T)
```
Verify 3-layer structure and correct formatting.

### Integration Tests Required

**File: `wizard_test.go`**

Test wizard initialization:
```go
func TestWizardInitialization(t *testing.T)
```
Verify wizard starts in correct state with empty choices.

Test state transitions:
```go
func TestStateTransitions(t *testing.T)
```
Simulate user choices and verify state changes correctly.

Test cancellation:
```go
func TestWizardCancellation(t *testing.T)
```
Verify Esc and Cancel selection both work correctly.

### Manual Testing Required

**Entry Screen**:
- [ ] Run `sow project` from main repo
- [ ] Verify entry screen displays
- [ ] Test arrow key navigation
- [ ] Test Enter to select
- [ ] Test Esc to cancel
- [ ] Verify terminal renders correctly

**Error Display**:
- [ ] Trigger an error condition
- [ ] Verify error displays in formatted box
- [ ] Verify can acknowledge and continue

**Loading Indicator**:
- [ ] Add test spinner somewhere (e.g., simulated delay)
- [ ] Verify spinner appears and animates
- [ ] Verify spinner disappears on completion

**Terminal Compatibility**:
- [ ] Test in macOS Terminal
- [ ] Test in iTerm2
- [ ] Test in VS Code integrated terminal
- [ ] Test with different terminal sizes

---

## 7. Implementation Notes

### Migration Strategy

**Critical**: Follow this exact order to avoid breaking existing functionality:

1. **Phase 1: Extract without breaking**
   - Create `shared.go` with extracted functions
   - Keep `new.go` and `continue.go` unchanged
   - Update `new.go` and `continue.go` to use shared functions
   - Test that existing commands still work

2. **Phase 2: Build wizard foundation**
   - Add huh dependency
   - Create wizard command structure
   - Implement state machine skeleton
   - Implement entry screen only
   - Test entry screen in isolation

3. **Phase 3: Replace commands**
   - Update `project.go` to use wizard command
   - Remove `new` and `continue` subcommands
   - Delete `new.go` and `continue.go`
   - Test that `sow project` launches wizard

### State Machine Pattern

Each state handler follows this pattern:
```go
func (w *Wizard) handleStateName() error {
    // 1. Create huh form
    form := huh.NewForm(...)

    // 2. Run form
    if err := form.Run(); err != nil {
        if errors.Is(err, huh.ErrUserAborted) {
            w.state = StateCancelled
            return nil
        }
        return err
    }

    // 3. Store choices
    w.choices["key"] = value

    // 4. Transition to next state
    w.state = NextState

    return nil
}
```

### Testing with Stub Handlers

For this work unit, subsequent screens can be stubs:
```go
func (w *Wizard) handleCreateSource() error {
    fmt.Println("Create source screen (stub)")
    w.state = StateComplete
    return nil
}
```

This allows testing state transitions without implementing all screens.

### Error Handling Philosophy

**User errors** (validation, conflicts):
- Display with `showError()`
- Allow retry or navigation back
- Never exit wizard unless user chooses

**System errors** (file I/O, git failures):
- Return error immediately
- Exit wizard
- Print error to stderr

### Spinner Usage Guidelines

Use spinners for operations > 500ms:
- GitHub API calls (fetching issues)
- Git operations (creating worktrees)
- File system scans (listing projects)

Don't use spinners for:
- User input (they control timing)
- Quick operations (< 500ms)

---

## 8. Files to Create

### Primary Implementation Files

```
cli/cmd/project/
├── wizard.go              # NEW: Wizard command entry point
├── wizard_state.go        # NEW: State machine and handlers
├── wizard_helpers.go      # NEW: Helper functions
├── wizard_helpers_test.go # NEW: Helper tests
├── wizard_test.go         # NEW: Wizard integration tests
└── shared.go              # NEW: Extracted utilities
    └── shared_test.go     # NEW: Shared function tests
```

### Files to Modify

```
cli/cmd/project/
└── project.go             # MODIFY: Replace new/continue with wizard command
```

### Files to Delete (after extraction)

```
cli/cmd/project/
├── new.go                 # DELETE: Replaced by wizard
└── continue.go            # DELETE: Replaced by wizard
```

### Dependency Files

```
go.mod                     # MODIFY: Add huh dependencies
go.sum                     # MODIFY: Auto-updated by go get
```

---

## 9. Dependencies on This Work Unit

**All subsequent work units depend on this foundation**:

- **Work Unit 002**: Branch name creation workflow - builds on state machine and helpers
- **Work Unit 003**: GitHub issue workflow - uses state machine and loading indicators
- **Work Unit 004**: Continue workflow - uses state machine and shared functions
- **Work Unit 005**: Validation and error handling - extends error display utilities

**This work unit must be complete and tested before starting any other work unit.**

---

## 10. Definition of Done

- [ ] All acceptance criteria met
- [ ] All unit tests passing
- [ ] All integration tests passing
- [ ] Manual testing checklist complete
- [ ] `new.go` and `continue.go` deleted
- [ ] No compilation errors or warnings
- [ ] `go mod tidy` completes successfully
- [ ] Code follows existing project conventions
- [ ] No TODO or FIXME comments in production code
- [ ] Running `sow project` launches wizard with entry screen
- [ ] User can navigate entry screen and cancel cleanly
- [ ] State machine handles all defined states correctly
- [ ] All helper functions tested and working
- [ ] Documentation strings (godoc) added to all public functions

---

## 11. Reference Implementation Examples

### Entry Screen Implementation

From technical design (lines 145-181):

```go
func (w *Wizard) handleEntry() error {
    var action string

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("What would you like to do?").
                Options(
                    huh.NewOption("Create new project", "create"),
                    huh.NewOption("Continue existing project", "continue"),
                    huh.NewOption("Cancel", "cancel"),
                ).
                Value(&action),
        ),
    )

    if err := form.Run(); err != nil {
        if errors.Is(err, huh.ErrUserAborted) {
            w.state = StateCancelled
            return nil
        }
        return err
    }

    w.choices["action"] = action

    // Transition
    switch action {
    case "create":
        w.state = StateCreateSource
    case "continue":
        w.state = StateProjectSelect
    case "cancel":
        w.state = StateCancelled
    }

    return nil
}
```

### Name Normalization Implementation

From technical design (lines 571-599):

```go
func normalizeName(name string) string {
    // Trim whitespace
    name = strings.TrimSpace(name)

    // Convert to lowercase
    name = strings.ToLower(name)

    // Replace spaces with hyphens
    name = strings.ReplaceAll(name, " ", "-")

    // Remove invalid characters (keep only a-z, 0-9, -, _)
    var result strings.Builder
    for _, r := range name {
        if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
            result.WriteRune(r)
        }
    }
    name = result.String()

    // Collapse multiple consecutive hyphens
    for strings.Contains(name, "--") {
        name = strings.ReplaceAll(name, "--", "-")
    }

    // Remove leading/trailing hyphens
    name = strings.Trim(name, "-")

    return name
}
```

### Spinner Usage Example

From technical design (lines 865-881):

```go
func withSpinner(title string, action func() error) error {
    var err error

    _ = spinner.New().
        Title(title).
        Action(func() {
            err = action()
        }).
        Run()

    return err
}

// Usage:
err := withSpinner("Creating worktree...", func() error {
    return sow.EnsureWorktree(ctx, path, branch)
})
```

---

## 12. Open Questions and Decisions

### Resolved
- **Q**: Should we maintain backward compatibility with `sow project new` flags?
  - **A**: No - design specifies complete replacement, not wrapper

- **Q**: Which keybinding for external editor?
  - **A**: Ctrl+E (verified in library verification doc, lines 195-232)

- **Q**: Should we check uncommitted changes for all operations?
  - **A**: Only when creating and current branch == target branch (UX design lines 395-420)

### For Implementer to Decide
- Exact error message formatting (follow library defaults vs. custom styling)
- Spinner type choice (Dot, Line, etc.) - recommend Dot for consistency
- Whether to add `--accessible` flag now or in polish phase

---

## 13. Related Work Units

- **Work Unit 002**: Branch Name Creation Workflow
- **Work Unit 003**: GitHub Issue Integration Workflow
- **Work Unit 004**: Continue Existing Project Workflow
- **Work Unit 005**: Validation and Error Handling
- **Work Unit 006**: Finalization and Claude Launch

---

## 14. Estimated Effort Breakdown

- **Dependency setup**: 15 minutes
- **Command structure and state machine**: 3 hours
- **Entry screen implementation**: 2 hours
- **Helper functions**: 4 hours
- **Extract shared utilities**: 4 hours
- **Unit tests**: 4 hours
- **Integration tests**: 2 hours
- **Manual testing and fixes**: 3 hours
- **Code cleanup and documentation**: 2 hours

**Total**: ~24 hours (3 working days)

---

## 15. Success Indicators

After this work unit is complete:

1. Developer can run `sow project` and see interactive entry screen
2. All helper functions have test coverage > 90%
3. State machine correctly routes between defined states
4. No existing functionality is broken
5. Code is ready for other work units to build upon
6. Foundation is solid enough that subsequent screens are "just more handlers"

The measure of success is that implementing the next work unit (branch name workflow) requires only adding new state handlers - no changes to the foundation code.
