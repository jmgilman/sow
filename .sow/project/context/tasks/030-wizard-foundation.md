# Task 030: Wizard Foundation and State Machine

## Context

This task creates the foundational structure for the interactive project wizard that will replace the flag-based `sow project new` and `sow project continue` commands. The wizard uses a state machine to guide users through creating or continuing projects with an interactive terminal UI.

This is foundational work that all subsequent wizard screens will build upon. The state machine defines 10 possible states and handles transitions between them. For this task, we implement the full state machine infrastructure with the entry screen as the only fully functional handler - other handlers can be stubs.

**Architecture Decision**: The wizard operates independently of the current git branch. Users can run `sow project` from any branch, and the wizard will work on target branches through worktrees. This eliminates entire categories of edge cases and state detection complexity.

## Requirements

### File 1: wizard.go - Command Entry Point

Create `cli/cmd/project/wizard.go` with:

**Command Structure**:
```go
func newWizardCmd() *cobra.Command
```
- Command: `sow project`
- No flags required (fully interactive)
- Support for Claude Code flags after `--` separator
- Short description: "Create or continue a project (interactive)"

**Run Function**:
```go
func runWizard(cmd *cobra.Command, args []string) error
```
- Validate sow is initialized (check `mainCtx.IsInitialized()`)
- Extract Claude Code flags after `--` separator
- Create and initialize wizard struct
- Call `wizard.Run()`
- Handle cancellation gracefully (exit without error)

**Wizard Struct**:
```go
type Wizard struct {
    state        WizardState
    ctx          *sow.Context
    choices      map[string]interface{}
    claudeFlags  []string
}

func NewWizard(ctx *sow.Context, claudeFlags []string) *Wizard
```
- Initialize with `StateEntry`
- Create empty choices map
- Store context and flags

### File 2: wizard_state.go - State Machine

Create `cli/cmd/project/wizard_state.go` with:

**State Type Definition**:
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
```

**Main Loop**:
```go
func (w *Wizard) Run() error {
    for w.state != StateComplete && w.state != StateCancelled {
        if err := w.handleState(); err != nil {
            return err
        }
    }

    if w.state == StateCancelled {
        return nil  // User cancelled, not an error
    }

    return w.finalize()
}
```

**State Dispatcher**:
```go
func (w *Wizard) handleState() error {
    switch w.state {
    case StateEntry:
        return w.handleEntry()
    case StateCreateSource:
        return w.handleCreateSource()
    case StateIssueSelect:
        return w.handleIssueSelect()
    case StateTypeSelect:
        return w.handleTypeSelect()
    case StateNameEntry:
        return w.handleNameEntry()
    case StatePromptEntry:
        return w.handlePromptEntry()
    case StateProjectSelect:
        return w.handleProjectSelect()
    case StateContinuePrompt:
        return w.handleContinuePrompt()
    default:
        return fmt.Errorf("unknown state: %s", w.state)
    }
}
```

**Entry Screen Handler (Full Implementation)**:
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
        return fmt.Errorf("entry screen error: %w", err)
    }

    w.choices["action"] = action

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

**Stub Handlers (For Other States)**:
```go
func (w *Wizard) handleCreateSource() error {
    fmt.Println("Create source screen (stub)")
    w.state = StateComplete
    return nil
}

// Similar stubs for:
// - handleIssueSelect
// - handleTypeSelect
// - handleNameEntry
// - handlePromptEntry
// - handleProjectSelect
// - handleContinuePrompt
```

**Finalize Method (Stub for now)**:
```go
func (w *Wizard) finalize() error {
    fmt.Println("Finalize: create/continue project and launch Claude")
    fmt.Printf("Choices: %+v\n", w.choices)
    return nil
}
```

## Acceptance Criteria

### Command Integration
- [ ] `newWizardCmd()` returns valid cobra command
- [ ] Command can be invoked with `sow project`
- [ ] Command accepts no required arguments
- [ ] Command supports Claude Code flags after `--` separator
- [ ] Command validates sow is initialized before running

### Wizard Initialization
- [ ] `NewWizard()` creates wizard with correct initial state (`StateEntry`)
- [ ] Wizard initializes with empty choices map
- [ ] Wizard stores context and claudeFlags correctly

### State Machine Operations
- [ ] `Run()` loops until `StateComplete` or `StateCancelled`
- [ ] `Run()` returns nil for `StateCancelled` (user chose to cancel)
- [ ] `Run()` calls `finalize()` for `StateComplete`
- [ ] `handleState()` dispatches to correct handler based on state
- [ ] `handleState()` returns error for unknown states

### Entry Screen Functionality
- [ ] Entry screen displays three options: Create, Continue, Cancel
- [ ] Arrow keys navigate between options
- [ ] Enter key selects option
- [ ] Esc key triggers cancellation (sets `StateCancelled`)
- [ ] Selecting "Create" transitions to `StateCreateSource`
- [ ] Selecting "Continue" transitions to `StateProjectSelect`
- [ ] Selecting "Cancel" transitions to `StateCancelled`
- [ ] Selection is stored in `choices["action"]`

### Stub Handlers
- [ ] All other state handlers print stub messages
- [ ] Stub handlers transition to `StateComplete` for testing
- [ ] No compilation errors from stub handlers

### Error Handling
- [ ] `huh.ErrUserAborted` is caught and handled as cancellation
- [ ] Form errors are wrapped with context
- [ ] Unknown states return descriptive error

## Relevant Inputs

- `cli/cmd/project/project.go` - Current command structure to understand how to integrate wizard
- `cli/cmd/project/new.go` - Reference for command flags pattern and Claude Code flags extraction
- `cli/cmd/project/continue.go` - Reference for command structure
- `cli/internal/sow/context.go` - Context type that wizard uses
- `.sow/knowledge/designs/interactive-wizard-ux-flow.md` - Entry screen design (lines 39-55)
- `.sow/knowledge/designs/interactive-wizard-technical-implementation.md` - State machine architecture (lines 20-102)
- `.sow/knowledge/designs/huh-library-verification.md` - How to use huh select prompts (lines 26-58)
- `.sow/project/context/issue-68.md` - Reference implementation (lines 629-667)

## Examples

### Example 1: Running the Wizard

```bash
$ sow project
╔══════════════════════════════════════════════════════════╗
║                     Sow Project Manager                  ║
╚══════════════════════════════════════════════════════════╝

What would you like to do?

  ○ Create new project
  ○ Continue existing project
  ○ Cancel

[Use arrow keys to navigate, Enter to select]
```

### Example 2: Wizard Initialization

```go
func runWizard(cmd *cobra.Command, args []string) error {
    mainCtx := cmdutil.GetContext(cmd.Context())

    if !mainCtx.IsInitialized() {
        fmt.Fprintln(os.Stderr, "Error: sow not initialized in this repository")
        fmt.Fprintln(os.Stderr, "Run: sow init")
        return fmt.Errorf("not initialized")
    }

    // Extract Claude Code flags
    var claudeFlags []string
    if dashIndex := cmd.ArgsLenAtDash(); dashIndex >= 0 {
        allArgs := cmd.Flags().Args()
        if dashIndex < len(allArgs) {
            claudeFlags = allArgs[dashIndex:]
        }
    }

    wizard := NewWizard(mainCtx, claudeFlags)
    return wizard.Run()
}
```

### Example 3: State Transition Pattern

```go
func (w *Wizard) handleSomeState() error {
    var userInput string

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("Choose an option:").
                Options(
                    huh.NewOption("Option 1", "opt1"),
                    huh.NewOption("Option 2", "opt2"),
                ).
                Value(&userInput),
        ),
    )

    if err := form.Run(); err != nil {
        if errors.Is(err, huh.ErrUserAborted) {
            w.state = StateCancelled
            return nil
        }
        return fmt.Errorf("some state error: %w", err)
    }

    w.choices["key"] = userInput

    // Transition based on input
    switch userInput {
    case "opt1":
        w.state = StateNextState1
    case "opt2":
        w.state = StateNextState2
    }

    return nil
}
```

## Dependencies

- **Task 010**: Huh library must be installed
- **Task 020**: Shared utilities extracted (wizard doesn't use them yet, but good to have)

## Constraints

- **Entry screen only**: Only implement entry screen fully; other screens are stubs
- **No worktree operations**: Don't create worktrees or projects yet (that's in finalize)
- **Simple error messages**: Keep error messages clear and user-friendly
- **No backward compatibility**: This replaces the old commands; no need to support old flags

## Testing Requirements

### Unit Tests

Create `cli/cmd/project/wizard_test.go`:

**Test: Wizard Initialization**
```go
func TestNewWizard_InitializesCorrectly(t *testing.T) {
    // Verify wizard starts in StateEntry
    // Verify choices map is empty
    // Verify context and flags are stored
}
```

**Test: State Machine Loop**
```go
func TestWizardRun_LoopsUntilTerminalState(t *testing.T) {
    // Test that Run() loops through states
    // Test that Run() exits on StateComplete
    // Test that Run() exits on StateCancelled
}
```

**Test: State Transitions**
```go
func TestHandleEntry_CreateTransition(t *testing.T) {
    // Simulate selecting "create"
    // Verify state transitions to StateCreateSource
    // Verify choices["action"] = "create"
}

func TestHandleEntry_ContinueTransition(t *testing.T) {
    // Simulate selecting "continue"
    // Verify state transitions to StateProjectSelect
}

func TestHandleEntry_CancelTransition(t *testing.T) {
    // Simulate selecting "cancel"
    // Verify state transitions to StateCancelled
}

func TestHandleEntry_EscapeKeyCancels(t *testing.T) {
    // Simulate Esc key (huh.ErrUserAborted)
    // Verify state transitions to StateCancelled
}
```

**Test: Error Handling**
```go
func TestHandleState_UnknownStateReturnsError(t *testing.T) {
    // Set wizard to unknown state
    // Verify handleState() returns error
}
```

**Note**: Testing huh forms may require mocking or integration tests. At minimum, test the state transition logic.

### Integration Tests

**Manual Testing**:
1. Run `sow project` in an initialized repository
2. Verify entry screen displays
3. Use arrow keys to navigate options
4. Press Enter on "Create" - should see stub message
5. Run again, press Enter on "Continue" - should see stub message
6. Run again, press Enter on "Cancel" - should exit cleanly
7. Run again, press Esc - should exit cleanly

**Test in Different Scenarios**:
- From main branch
- From feature branch
- With uncommitted changes (wizard shouldn't care at this stage)
- In uninitialized repo (should show error)

## Implementation Notes

### Why Stubs for Other States?

Stubs allow us to:
1. **Test state machine**: Verify transitions work correctly
2. **Validate architecture**: Ensure the structure is sound
3. **Enable parallel work**: Other tasks can implement real handlers
4. **Incremental development**: Build one screen at a time

### State Machine Pattern

Each handler follows the same pattern:
1. Create huh form with appropriate fields
2. Run form and capture input
3. Handle cancellation (Esc key)
4. Store choices in map
5. Transition to next state based on input

This pattern makes the code predictable and easy to extend.

### Error Handling Philosophy

**User errors** (Esc, Cancel):
- Not fatal errors
- Set state to `StateCancelled`
- Return nil (wizard handles gracefully)

**System errors** (file I/O, git failures):
- Fatal errors
- Return error with context
- Wizard exits with error message

### Cancel vs. Abort

- **Cancel** (user selects "Cancel" option): Explicit choice to exit
- **Abort** (user presses Esc): Implicit choice to exit
- Both are treated the same: transition to `StateCancelled`, exit cleanly

## Success Indicators

After completing this task:
1. User can run `sow project` and see interactive entry screen
2. Navigation works correctly (arrow keys, Enter, Esc)
3. State transitions work for all entry screen options
4. Stub handlers demonstrate the pattern for future screens
5. State machine architecture is proven and testable
6. Foundation is solid for implementing real handlers in subsequent tasks
7. No compilation errors, all tests pass
