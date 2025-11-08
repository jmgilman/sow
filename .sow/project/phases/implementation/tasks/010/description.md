# Task 010: Implement Create Source Selection Handler

## Context

This task implements the source selection screen that appears after the user chooses "Create new project" in the wizard entry screen. The source selection allows users to choose between creating a project from a GitHub issue or from a custom branch name.

This is part of Work Unit 002 (Project Creation Workflow - Branch Name Path) which builds on the wizard foundation from Work Unit 001. The wizard state machine, normalization utilities, and shared project functions already exist.

**Project Goal**: Build an interactive wizard for creating new sow projects via branch name selection, including type selection, name entry with real-time preview, prompt entry with external editor support, and project initialization in git worktrees.

**Why This Task**: This screen is the entry point that directs users into either the GitHub issue path (work unit 003) or the branch name path (this work unit). It's a simple branching screen but critical for proper state machine flow.

## Requirements

### Handler Implementation

Create the `handleCreateSource()` function in `cli/cmd/project/wizard_state.go` to replace the current stub implementation.

**Function Location**: Replace the stub at lines 124-129 in `wizard_state.go`

**Display Requirements**:
- Show two options for project creation:
  - "From GitHub issue" (value: "issue")
  - "From branch name" (value: "branch")
  - "Cancel" (value: "cancel")
- Use `huh.NewSelect[string]()` for the selection prompt
- Title: "How would you like to create the project?"

**State Transitions**:
- If user selects "issue" → transition to `StateIssueSelect`
- If user selects "branch" → transition to `StateTypeSelect`
- If user selects "cancel" → transition to `StateCancelled`
- If user presses Ctrl+C/Esc → catch `huh.ErrUserAborted` and transition to `StateCancelled`

**Data Storage**:
- Store the selection in `w.choices["source"]` as a string
- This allows other handlers to know which creation path was chosen

### Error Handling

- Handle `huh.ErrUserAborted` gracefully by transitioning to `StateCancelled`
- Return other errors to allow the wizard to display them
- Ensure the wizard doesn't crash on unexpected form errors

### Integration Points

**Upstream**: Called from `handleState()` when `w.state == StateCreateSource`

**Downstream**:
- Transitions to `StateIssueSelect` (to be implemented in work unit 003)
- Transitions to `StateTypeSelect` (task 020 in this work unit)
- Transitions to `StateCancelled` (handled by wizard's Run() method)

## Acceptance Criteria

### Functional Requirements

1. **Display Works**
   - All three options are displayed in the selection menu
   - Option labels are clear and descriptive
   - Cancel option is available for users to exit

2. **Selection Stored**
   - User's selection is stored in `w.choices["source"]`
   - Value is one of: "issue", "branch", or "cancel"

3. **State Transitions Correct**
   - "From GitHub issue" → `StateIssueSelect`
   - "From branch name" → `StateTypeSelect`
   - "Cancel" → `StateCancelled`
   - Ctrl+C/Esc → `StateCancelled`

4. **Error Handling Works**
   - User abort (Ctrl+C/Esc) doesn't crash the wizard
   - Unexpected errors are propagated correctly

### Test Requirements (TDD Approach)

Write tests BEFORE implementing the handler:

**Unit Tests** (add to `wizard_state_test.go`):

```go
func TestHandleCreateSource_SelectsIssue(t *testing.T) {
    // Test that selecting "issue" transitions to StateIssueSelect
    // and stores "issue" in choices
}

func TestHandleCreateSource_SelectsBranch(t *testing.T) {
    // Test that selecting "branch" transitions to StateTypeSelect
    // and stores "branch" in choices
}

func TestHandleCreateSource_SelectsCancel(t *testing.T) {
    // Test that selecting "cancel" transitions to StateCancelled
}

func TestHandleCreateSource_UserAbort(t *testing.T) {
    // Test that Ctrl+C/Esc transitions to StateCancelled
    // Mock huh to return ErrUserAborted
}
```

**Manual Testing**:
1. Run `sow project` and select "Create new project"
2. Verify source selection screen appears
3. Select "From branch name" and verify it proceeds to type selection
4. Go back and select "Cancel" - verify wizard exits cleanly
5. Press Ctrl+C during selection - verify wizard exits gracefully

## Technical Details

### Implementation Pattern

Follow the existing pattern from `handleEntry()` at lines 86-122 in `wizard_state.go`:

```go
func (w *Wizard) handleCreateSource() error {
    var source string

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("How would you like to create the project?").
                Options(
                    huh.NewOption("From GitHub issue", "issue"),
                    huh.NewOption("From branch name", "branch"),
                    huh.NewOption("Cancel", "cancel"),
                ).
                Value(&source),
        ),
    )

    if err := form.Run(); err != nil {
        if errors.Is(err, huh.ErrUserAborted) {
            w.state = StateCancelled
            return nil
        }
        return fmt.Errorf("create source screen error: %w", err)
    }

    w.choices["source"] = source

    switch source {
    case "issue":
        w.state = StateIssueSelect
    case "branch":
        w.state = StateTypeSelect
    case "cancel":
        w.state = StateCancelled
    }

    return nil
}
```

### Package and Imports

All required imports are already in `wizard_state.go`:
- `errors` - for `errors.Is()`
- `fmt` - for error formatting
- `github.com/charmbracelet/huh` - for form building

No new imports needed.

### File Structure

```
cli/cmd/project/
├── wizard_state.go           # MODIFY: Replace handleCreateSource stub
├── wizard_state_test.go      # CREATE: Add tests for this handler
```

## Relevant Inputs

### Existing Code to Understand

- `cli/cmd/project/wizard_state.go:86-122` - Example pattern from `handleEntry()` showing select prompt with state transitions
- `cli/cmd/project/wizard_state.go:12-26` - WizardState constants including `StateCreateSource`, `StateIssueSelect`, `StateTypeSelect`, `StateCancelled`
- `cli/cmd/project/wizard_state.go:28-44` - Wizard struct showing `state` and `choices` fields
- `cli/cmd/project/wizard_state.go:62-83` - `handleState()` dispatcher showing how handlers are called

### Design Documents

- `.sow/knowledge/designs/interactive-wizard-ux-flow.md:123-174` - Complete UX flow specification showing source selection screen
- `.sow/knowledge/designs/interactive-wizard-technical-implementation.md:121-182` - Select prompt implementation pattern with state transitions
- `.sow/knowledge/designs/huh-library-verification.md:28-87` - Select field capabilities and validation patterns

### Testing Patterns

- `cli/cmd/project/wizard_helpers_test.go:1-253` - Example test patterns for wizard helper functions
- `cli/cmd/project/shared_test.go:23-70` - Test setup utilities (`setupTestRepo`, `setupTestContext`)

## Examples

### Example User Flow

```
$ sow project

[Entry Screen]
What would you like to do?
  ● Create new project
  ○ Continue existing project
  ○ Cancel

[User selects "Create new project"]

[Create Source Screen - THIS TASK]
How would you like to create the project?
  ○ From GitHub issue
  ● From branch name
  ○ Cancel

[User selects "From branch name"]

[Transitions to Type Selection Screen - Task 020]
```

### Example Error Recovery

```
[Create Source Screen]
How would you like to create the project?
  ○ From GitHub issue
  ○ From branch name
  ○ Cancel

[User presses Ctrl+C]

[Wizard exits cleanly, no error message]
$ _
```

## Dependencies

### Upstream Dependencies (Must Complete First)

- **Work Unit 001**: Wizard Foundation and State Machine ✅ COMPLETE
  - Provides: `WizardState` enum with `StateCreateSource`, `StateIssueSelect`, `StateTypeSelect`, `StateCancelled`
  - Provides: Wizard struct with `state` and `choices` fields
  - Provides: State machine dispatch via `handleState()`

### Downstream Dependencies (Will Use This Task)

- **Task 020**: Type selection handler (triggered when user selects "From branch name")
- **Work Unit 003**: Issue selection handler (triggered when user selects "From GitHub issue")

## Constraints

### UX Requirements

- Options must be in logical order: issue first, branch second, cancel last
- Title must clearly indicate this is about choosing HOW to create, not WHAT to create
- Cancel option must always be available for user escape

### State Machine Requirements

- Never transition to an invalid state
- Always store selection before transitioning
- Handle user abort gracefully without error messages

### Testing Requirements

- Tests must be written BEFORE implementation (TDD)
- Tests must not require actual UI interaction (mock huh forms)
- Tests must verify both state transitions and data storage

### What NOT to Do

- ❌ Don't implement issue selection logic here (that's work unit 003)
- ❌ Don't implement type selection logic here (that's task 020)
- ❌ Don't show error messages for user cancellation (it's intentional)
- ❌ Don't add new state transitions beyond the three specified
- ❌ Don't modify the Wizard struct or WizardState enum (those are from work unit 001)

## Notes

### Critical Implementation Details

1. **Choice Storage**: Always store the selection in `w.choices["source"]` before transitioning. This allows the finalization logic to know which creation path was used.

2. **Error Handling**: The only "expected" error is `huh.ErrUserAborted`. All other errors should be propagated up to the wizard's Run() method.

3. **State Transitions**: The state transition logic is a simple switch statement. No validation needed since the form only allows the three specified values.

### Testing Strategy

Since we can't easily test interactive forms, the tests should focus on:
- State transitions (mock the form to return specific values)
- Data storage (verify choices map is populated)
- Error handling (verify ErrUserAborted is handled correctly)

The manual testing scenarios verify the actual UX works as expected.
