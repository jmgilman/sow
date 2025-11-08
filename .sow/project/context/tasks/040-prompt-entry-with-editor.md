# Task 040: Implement Prompt Entry with External Editor Support

## Context

This task implements the optional initial prompt entry screen where users can provide their task or question for Claude. The key feature is external editor support via Ctrl+E, allowing users to compose multi-line prompts in their preferred editor (vim, VS Code, nano, etc.).

This is part of Work Unit 002 (Project Creation Workflow - Branch Name Path). The wizard foundation from Work Unit 001 provides the state machine. The huh library provides built-in external editor support through `EditorExtension()`, which was verified in the library verification document.

**Project Goal**: Build an interactive wizard for creating new sow projects via branch name selection, including type selection, name entry with real-time preview, prompt entry with external editor support, and project initialization in git worktrees.

**Why This Task**: The initial prompt allows users to give Claude context about what they want to accomplish. External editor support is critical for longer, more detailed prompts that would be awkward to type in a terminal input field.

## Requirements

### Handler Implementation

Create the `handlePromptEntry()` function in `cli/cmd/project/wizard_state.go` to replace the current stub implementation.

**Function Location**: Replace the stub at lines 152-157 in `wizard_state.go`

**Display Requirements**:
- Show context information:
  - Type: `w.choices["type"]` (e.g., "exploration")
  - Branch: `w.choices["branch"]` (e.g., "explore/web-based-agents")
- Multi-line text area using `huh.NewText()`
  - Title: "Enter your task or question for Claude (optional):"
  - Description: Show type, branch, and Ctrl+E instruction
  - Character limit: 10,000 characters (generous for flexibility)
  - Optional field (user can leave empty)
- External editor configuration:
  - Use `.EditorExtension(".md")` for markdown syntax highlighting
  - Editor triggered by Ctrl+E keybinding
  - Falls back to $EDITOR or nano if not set

**Context Display Format**:
```
Type: <type>
Branch: <branch>

Press Ctrl+E to open $EDITOR for multi-line input
```

**State Transitions**:
- User submits (with or without text) → store prompt and transition to `StateComplete`
- User presses Esc → transition back to `StateNameEntry` (go back)
- User presses Ctrl+C → catch `huh.ErrUserAborted`, transition to `StateCancelled`

**Data Storage**:
- Store prompt text in `w.choices["prompt"]` as a string
- Empty string if user leaves it blank (optional field)

### Editor Integration

**How External Editor Works**:
1. User presses Ctrl+E in the text area
2. huh creates temporary file with `.md` extension
3. huh opens $EDITOR (or nano if unset) with temp file
4. User edits in their preferred editor
5. User saves and exits editor
6. huh reads temp file contents and populates text area
7. User can edit more in terminal or submit

**Configuration**:
- Use `.EditorExtension(".md")` - gives syntax highlighting in VS Code, vim with markdown plugins
- No need to specify editor - huh automatically uses $EDITOR environment variable
- Falls back to nano if $EDITOR not set

### Error Handling

- Handle `huh.ErrUserAborted` gracefully (Ctrl+C) → transition to `StateCancelled`
- Esc key → transition back to `StateNameEntry` (user can change their project name)
- No validation needed - prompt is optional and any text is acceptable
- If editor fails to launch, huh will show error - no special handling needed

### Integration Points

**Upstream**: Called from `handleState()` when `w.state == StatePromptEntry`, triggered by `handleNameEntry()` after successful name entry

**Downstream**: Transitions to `StateComplete` which triggers `finalize()` method to create project and launch Claude

## Acceptance Criteria

### Functional Requirements

1. **Context Display Works**
   - Shows selected project type
   - Shows full branch name (prefix + normalized name)
   - Shows Ctrl+E instruction

2. **Text Entry Works**
   - Multi-line text can be entered directly in terminal
   - Character limit doesn't trigger warnings under 10,000 chars
   - Optional field - user can submit empty

3. **External Editor Integration**
   - Ctrl+E opens $EDITOR (or nano if unset)
   - Temp file has .md extension for syntax highlighting
   - Content from editor populates text area
   - User can continue editing after closing editor

4. **Navigation Works**
   - Enter submits and proceeds to completion
   - Esc goes back to name entry
   - Ctrl+C cancels wizard entirely

5. **Data Stored Correctly**
   - Prompt text stored in `w.choices["prompt"]`
   - Empty string if user skips prompt

### Test Requirements (TDD Approach)

Write tests BEFORE implementing the handler:

**Unit Tests** (add to `wizard_state_test.go`):

```go
func TestHandlePromptEntry_WithText(t *testing.T) {
    // Mock form to submit some text
    // Verify: choices["prompt"] contains the text
    // Verify: state transitions to StateComplete
}

func TestHandlePromptEntry_EmptyText(t *testing.T) {
    // Mock form to submit empty string
    // Verify: choices["prompt"] = ""
    // Verify: state transitions to StateComplete (empty is OK)
}

func TestHandlePromptEntry_UserAbort(t *testing.T) {
    // Mock form to return ErrUserAborted
    // Verify: state transitions to StateCancelled
}

func TestHandlePromptEntry_UserGoesBack(t *testing.T) {
    // Mock form to simulate Esc key
    // Verify: state transitions to StateNameEntry
}

func TestHandlePromptEntry_ContextDisplay(t *testing.T) {
    // Set up wizard with type and branch in choices
    // Verify: description includes both type and branch
}
```

**Manual Testing** (critical for editor integration):

1. **Terminal Entry**:
   - Run wizard, reach prompt entry
   - Type multi-line text directly in terminal
   - Submit → verify proceeds to completion

2. **External Editor - Default**:
   - Unset $EDITOR
   - Run wizard, reach prompt entry
   - Press Ctrl+E → verify nano opens
   - Type text, save, exit → verify text appears in prompt field

3. **External Editor - Vim**:
   - Set EDITOR=vim
   - Run wizard, reach prompt entry
   - Press Ctrl+E → verify vim opens with .md file
   - Type text with markdown, save, exit → verify text appears

4. **External Editor - VS Code**:
   - Set EDITOR="code --wait"
   - Run wizard, reach prompt entry
   - Press Ctrl+E → verify VS Code opens
   - Type text, close file → verify text appears

5. **Empty Prompt**:
   - Run wizard, reach prompt entry
   - Leave field empty and submit → verify proceeds to completion

6. **Character Limit**:
   - Enter 5000 characters → no warning
   - Enter 10000 characters → no warning
   - Enter 10001 characters → huh shows limit warning

## Technical Details

### Implementation Pattern

```go
func (w *Wizard) handlePromptEntry() error {
    var prompt string

    projectType := w.choices["type"].(string)
    branchName := w.choices["branch"].(string)

    contextInfo := fmt.Sprintf(
        "Type: %s\nBranch: %s\n\nPress Ctrl+E to open $EDITOR for multi-line input",
        projectType, branchName)

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewText().
                Title("Enter your task or question for Claude (optional):").
                Description(contextInfo).
                CharLimit(10000).
                Value(&prompt).
                EditorExtension(".md"),  // Enable external editor with .md extension
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

### External Editor Configuration

**Editor Selection Priority** (handled by huh automatically):
1. $EDITOR environment variable (user's preferred editor)
2. nano (default fallback)

**Common Editor Configurations**:
- Vim: `EDITOR=vim`
- Nano: `EDITOR=nano` (or unset, nano is default)
- VS Code: `EDITOR="code --wait"` (--wait flag critical!)
- Emacs: `EDITOR=emacs`

**Why .md Extension**:
- VS Code provides markdown preview and syntax highlighting
- Vim with markdown plugins provides syntax highlighting
- Helps users see if they're writing markdown-formatted text
- No downside - plain text editors ignore the extension

### Handling Esc vs Ctrl+C

**Note**: The current implementation treats Esc as going back to `StateNameEntry`. However, huh's default behavior for Esc in text fields is to abort the form (same as Ctrl+C).

We should handle this by catching `ErrUserAborted` and checking if we want to go back or cancel. For simplicity, we'll treat both Esc and Ctrl+C as cancel (transition to `StateCancelled`).

If we want Esc to go back, we need to:
1. Catch the error
2. Somehow distinguish Esc from Ctrl+C (may not be possible with huh)
3. Transition appropriately

For this task, keep it simple: both Esc and Ctrl+C → `StateCancelled`.

### Package and Imports

All required imports are already in `wizard_state.go`:
- `errors` - for `errors.Is()`
- `fmt` - for string formatting
- `github.com/charmbracelet/huh` - for form building

No new imports needed.

### File Structure

```
cli/cmd/project/
├── wizard_state.go           # MODIFY: Replace handlePromptEntry stub
├── wizard_state_test.go      # MODIFY: Add tests for this handler
```

## Relevant Inputs

### Existing Code to Understand

- `cli/cmd/project/wizard_state.go:12-26` - WizardState constants including `StatePromptEntry`, `StateComplete`, `StateCancelled`
- `cli/cmd/project/wizard_state.go:28-44` - Wizard struct showing `choices` field for data storage

### Design Documents

- `.sow/knowledge/designs/interactive-wizard-ux-flow.md:253-275` - Prompt entry screen specification with external editor
- `.sow/knowledge/designs/interactive-wizard-technical-implementation.md:236-276` - Text area implementation example with editor
- `.sow/knowledge/designs/huh-library-verification.md:183-248` - External editor support verification showing Ctrl+E keybinding and EditorExtension usage
- `.sow/project/context/issue-69.md:179-202` - Prompt entry requirements including editor integration

### Library Capabilities

From huh library verification:
- Ctrl+E is the keybinding (NOT Ctrl+O as originally designed)
- `EditorExtension(".md")` sets file extension for syntax highlighting
- `CharLimit()` sets maximum characters
- `Title()` and `Description()` for labeling
- External editor is enabled by default for Text fields

### Testing Patterns

- `cli/cmd/project/shared_test.go:23-70` - Test setup utilities for context creation

## Examples

### Example: Terminal Entry

```
Type: Exploration
Branch: explore/web-based-agents

Enter your task or question for Claude (optional):
Press Ctrl+E to open $EDITOR for multi-line input

┌────────────────────────────────────────────────────────┐
│ Research the landscape of web-based agent frameworks   │
│ and compare their architectures. Focus on:             │
│ - Multi-agent coordination                             │
│ - Browser automation capabilities                      │
│ - Integration patterns                                 │
└────────────────────────────────────────────────────────┘

[User presses Enter to submit]
[Proceeds to finalization]
```

### Example: External Editor Flow

```
Type: Design
Branch: design/api-architecture

Enter your task or question for Claude (optional):
Press Ctrl+E to open $EDITOR for multi-line input

┌────────────────────────────────────────────────────────┐
│                                                        │
└────────────────────────────────────────────────────────┘

[User presses Ctrl+E]

[Vim opens with temporary file ending in .md]
# API Architecture Design

Design a REST API for the new microservice with these requirements:

- OAuth2 authentication
- Rate limiting per API key
- Pagination for list endpoints
- OpenAPI 3.0 documentation

Consider:
- Backward compatibility with v1 API
- Migration path for existing clients
- Performance implications of rate limiting
~
~
:wq

[User saves and exits vim]

[Back to prompt screen, text is populated]
┌────────────────────────────────────────────────────────┐
│ # API Architecture Design                              │
│                                                        │
│ Design a REST API for the new microservice with...    │
└────────────────────────────────────────────────────────┘

[User presses Enter to submit]
[Proceeds to finalization]
```

### Example: Empty Prompt

```
Type: Standard
Branch: feat/authentication

Enter your task or question for Claude (optional):
Press Ctrl+E to open $EDITOR for multi-line input

┌────────────────────────────────────────────────────────┐
│                                                        │
└────────────────────────────────────────────────────────┘

[User presses Enter without typing anything]
[Proceeds to finalization with empty prompt]
```

## Dependencies

### Upstream Dependencies (Must Complete First)

- **Work Unit 001**: Wizard Foundation and State Machine ✅ COMPLETE
  - Provides: `WizardState` enum with `StatePromptEntry`, `StateComplete`, `StateCancelled`
  - Provides: Wizard struct with `choices` field

- **Task 030**: Name entry handler (this work unit)
  - Provides: `w.choices["type"]` for context display
  - Provides: `w.choices["branch"]` for context display

### Downstream Dependencies (Will Use This Task)

- **Task 050**: Finalization handler
  - Reads: `w.choices["prompt"]` to pass to `generateNewProjectPrompt()`

## Constraints

### UX Requirements

- Prompt must be optional (user can skip)
- Instructions must mention Ctrl+E explicitly
- Context (type and branch) must be visible
- No character limit warnings under 10,000 characters

### External Editor Requirements

- Must use $EDITOR environment variable
- Must fall back to nano if $EDITOR unset
- Must use .md extension for temp file
- Must handle editor failure gracefully (huh handles this)

### State Machine Requirements

- Empty prompt is valid - store as empty string
- Both Esc and Ctrl+C should cancel (transition to `StateCancelled`)
- Success always transitions to `StateComplete`

### Testing Requirements

- Tests written BEFORE implementation (TDD)
- Unit tests for handler flow and data storage
- Manual tests for external editor integration (can't be automated easily)

### What NOT to Do

- ❌ Don't require prompt text - it's optional
- ❌ Don't use Ctrl+O for editor - huh uses Ctrl+E
- ❌ Don't set editor explicitly - use $EDITOR environment variable
- ❌ Don't add complex validation - any text is acceptable
- ❌ Don't differentiate Esc vs Ctrl+C - both cancel (simplicity)
- ❌ Don't set character limit too low - 10,000 gives flexibility

## Notes

### Critical Implementation Details

1. **EditorExtension**: The `.md` extension enables syntax highlighting in editors that support it (VS Code, vim with plugins). It has no downside for editors that don't support it.

2. **Optional Field**: Users should feel comfortable leaving this blank. The "(optional)" label in the title makes this clear.

3. **Context Display**: Showing the type and branch helps users understand what project they're creating, especially if they took a break during the wizard.

4. **Ctrl+E vs Ctrl+O**: The original design docs specified Ctrl+O, but the huh library verification found it's actually Ctrl+E. All user-facing text must use Ctrl+E.

### Testing Strategy

**Unit Tests**: Focus on state transitions and data storage. Mock the form to return different values.

**Manual Tests**: Critical for verifying external editor integration actually works. Test with multiple editors (vim, nano, VS Code) to ensure compatibility.

The manual tests catch issues that unit tests can't (e.g., VS Code needing --wait flag, file extension working correctly).

### Common Editor Issues

**VS Code**: Users must set `EDITOR="code --wait"`. Without --wait, VS Code returns immediately and huh reads an empty file.

**GUI Editors**: Similar issue to VS Code - they may need special flags to wait for file close.

**Terminal Editors**: Vim, nano, emacs work out of the box - they block until user exits.

The wizard doesn't need to handle these cases - users who set $EDITOR to a GUI editor need to configure it correctly. The --wait flag is standard practice.
