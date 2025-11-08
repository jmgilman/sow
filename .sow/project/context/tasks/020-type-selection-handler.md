# Task 020: Implement Project Type Selection Handler

## Context

This task implements the project type selection screen for the branch name creation path. After the user selects "From branch name" in the source selection screen, they need to choose which type of project they're creating (standard, exploration, design, or breakdown).

This is part of Work Unit 002 (Project Creation Workflow - Branch Name Path). The wizard foundation from Work Unit 001 provides the state machine, and the project type configuration infrastructure already exists. This task leverages the existing `projectTypes` map and `getTypeOptions()` helper from `wizard_helpers.go`.

**Project Goal**: Build an interactive wizard for creating new sow projects via branch name selection, including type selection, name entry with real-time preview, prompt entry with external editor support, and project initialization in git worktrees.

**Why This Task**: The project type determines the branch prefix (feat/, explore/, design/, breakdown/) and the project's phase structure. This selection must happen before name entry so the preview can show the correct branch name.

## Requirements

### Handler Implementation

Create the `handleTypeSelect()` function in `cli/cmd/project/wizard_state.go` to replace the current stub implementation.

**Function Location**: Replace the stub at lines 138-143 in `wizard_state.go`

**Display Requirements**:
- Show all four project types with their descriptions:
  - "Feature work and bug fixes" (value: "standard")
  - "Research and investigation" (value: "exploration")
  - "Architecture and design documents" (value: "design")
  - "Decompose work into tasks" (value: "breakdown")
  - "Cancel" (value: "cancel")
- Use `huh.NewSelect[string]()` for the selection prompt
- Title: "What type of project?"
- Use the existing `getTypeOptions()` helper function from `wizard_helpers.go` to generate options

**State Transitions**:
- If user selects a project type → store selection and transition to `StateNameEntry`
- If user selects "cancel" → transition to `StateCancelled`
- If user presses Ctrl+C/Esc → catch `huh.ErrUserAborted` and transition to `StateCancelled`

**Data Storage**:
- Store the selection in `w.choices["type"]` as a string
- Value should be one of: "standard", "exploration", "design", "breakdown", or "cancel"

### Error Handling

- Handle `huh.ErrUserAborted` gracefully by transitioning to `StateCancelled`
- Return other errors to allow the wizard to display them
- Ensure the wizard doesn't crash on unexpected form errors

### Integration Points

**Upstream**: Called from `handleState()` when `w.state == StateTypeSelect`, triggered by `handleCreateSource()` when user selects "From branch name"

**Downstream**: Transitions to `StateNameEntry` (task 030) which uses the selected type to determine branch prefix for preview

## Acceptance Criteria

### Functional Requirements

1. **All Types Displayed**
   - All four project types shown with correct descriptions
   - Descriptions match the `projectTypes` map in `wizard_helpers.go`
   - Cancel option is available

2. **Selection Stored Correctly**
   - User's selection stored in `w.choices["type"]`
   - Value is one of the four valid types or "cancel"

3. **State Transitions Work**
   - Any project type selection → `StateNameEntry`
   - "Cancel" → `StateCancelled`
   - Ctrl+C/Esc → `StateCancelled`

4. **Options Use Correct Configuration**
   - Options generated using `getTypeOptions()` helper
   - Descriptions match the projectTypes map
   - Order is: standard, exploration, design, breakdown, cancel

### Test Requirements (TDD Approach)

Write tests BEFORE implementing the handler:

**Unit Tests** (add to `wizard_state_test.go`):

```go
func TestHandleTypeSelect_SelectsStandard(t *testing.T) {
    // Test selecting "standard" type
    // Verify: choices["type"] = "standard"
    // Verify: state transitions to StateNameEntry
}

func TestHandleTypeSelect_SelectsExploration(t *testing.T) {
    // Test selecting "exploration" type
    // Verify: choices["type"] = "exploration"
    // Verify: state transitions to StateNameEntry
}

func TestHandleTypeSelect_SelectsDesign(t *testing.T) {
    // Test selecting "design" type
    // Verify: choices["type"] = "design"
    // Verify: state transitions to StateNameEntry
}

func TestHandleTypeSelect_SelectsBreakdown(t *testing.T) {
    // Test selecting "breakdown" type
    // Verify: choices["type"] = "breakdown"
    // Verify: state transitions to StateNameEntry
}

func TestHandleTypeSelect_SelectsCancel(t *testing.T) {
    // Test selecting "cancel"
    // Verify: state transitions to StateCancelled
}

func TestHandleTypeSelect_UserAbort(t *testing.T) {
    // Test Ctrl+C/Esc handling
    // Mock huh to return ErrUserAborted
    // Verify: state transitions to StateCancelled
}

func TestHandleTypeSelect_UsesTypeOptionsHelper(t *testing.T) {
    // Verify getTypeOptions() is called
    // This ensures consistency with projectTypes configuration
}
```

**Manual Testing**:
1. Run `sow project` → Create new project → From branch name
2. Verify all four project types are displayed with descriptions
3. Select "Exploration" → verify proceeds to name entry
4. Go back and select "Cancel" → verify wizard exits
5. Press Ctrl+C → verify wizard exits gracefully

## Technical Details

### Implementation Pattern

Use the existing `getTypeOptions()` helper from `wizard_helpers.go:118-127`:

```go
func (w *Wizard) handleTypeSelect() error {
    var selectedType string

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("What type of project?").
                Options(getTypeOptions()...).  // Use helper function
                Value(&selectedType),
        ),
    )

    if err := form.Run(); err != nil {
        if errors.Is(err, huh.ErrUserAborted) {
            w.state = StateCancelled
            return nil
        }
        return fmt.Errorf("type selection error: %w", err)
    }

    if selectedType == "cancel" {
        w.state = StateCancelled
        return nil
    }

    w.choices["type"] = selectedType
    w.state = StateNameEntry

    return nil
}
```

### Key Dependencies

**From `wizard_helpers.go`**:
- `getTypeOptions()` (lines 118-127) - Returns huh.Option slice with all types + cancel
- `projectTypes` map (lines 19-36) - Configuration mapping type names to prefixes and descriptions

**Option Format**:
Each option from `getTypeOptions()` has:
- Label: The description (e.g., "Feature work and bug fixes")
- Value: The type key (e.g., "standard")

### Package and Imports

All required imports are already in `wizard_state.go`:
- `errors` - for `errors.Is()`
- `fmt` - for error formatting
- `github.com/charmbracelet/huh` - for form building

No new imports needed.

### File Structure

```
cli/cmd/project/
├── wizard_state.go           # MODIFY: Replace handleTypeSelect stub
├── wizard_helpers.go         # READ: Use getTypeOptions() helper
├── wizard_state_test.go      # CREATE: Add tests for this handler
```

## Relevant Inputs

### Existing Code to Understand

- `cli/cmd/project/wizard_helpers.go:19-36` - `projectTypes` map with all four types and their configurations
- `cli/cmd/project/wizard_helpers.go:118-127` - `getTypeOptions()` helper that generates huh options from projectTypes
- `cli/cmd/project/wizard_state.go:12-26` - WizardState constants including `StateTypeSelect`, `StateNameEntry`, `StateCancelled`
- `cli/cmd/project/wizard_state.go:86-122` - Example pattern from `handleEntry()` showing select prompt with state transitions

### Project Type Infrastructure

- `cli/internal/projects/standard/standard.go:1-26` - Standard project type registration showing how types are configured
- `cli/internal/projects/exploration/exploration.go:1-26` - Exploration project type showing different phase structure
- `cli/internal/projects/design/design.go` - Design project type
- `cli/internal/projects/breakdown/breakdown.go` - Breakdown project type

### Design Documents

- `.sow/knowledge/designs/interactive-wizard-ux-flow.md:175-194` - Type selection screen UX specification
- `.sow/knowledge/designs/interactive-wizard-technical-implementation.md:121-182` - Select prompt implementation pattern
- `.sow/project/context/issue-69.md:102-121` - Type selection requirements and display format

### Testing Patterns

- `cli/cmd/project/wizard_helpers_test.go:106-133` - `TestGetTypePrefix` showing how to test type configuration
- `cli/cmd/project/wizard_helpers_test.go:134-147` - `TestGetTypeOptions` showing how to verify option count
- `cli/cmd/project/wizard_helpers_test.go:206-252` - `TestProjectTypesMap` showing comprehensive type verification

## Examples

### Example User Flow

```
$ sow project

[Entry Screen]
→ User selects "Create new project"

[Create Source Screen]
→ User selects "From branch name"

[Type Selection Screen - THIS TASK]
What type of project?

  ○ Feature work and bug fixes
  ● Research and investigation
  ○ Architecture and design documents
  ○ Decompose work into tasks
  ○ Cancel

[User selects "Research and investigation"]

[Transitions to Name Entry Screen]
Type: Exploration
Enter project name:
...
```

### Example Type Configurations

From `projectTypes` map in `wizard_helpers.go`:

```go
"standard": {
    Prefix:      "feat/",
    Description: "Feature work and bug fixes",
}

"exploration": {
    Prefix:      "explore/",
    Description: "Research and investigation",
}

"design": {
    Prefix:      "design/",
    Description: "Architecture and design documents",
}

"breakdown": {
    Prefix:      "breakdown/",
    Description: "Decompose work into tasks",
}
```

## Dependencies

### Upstream Dependencies (Must Complete First)

- **Work Unit 001**: Wizard Foundation and State Machine ✅ COMPLETE
  - Provides: `WizardState` enum with `StateTypeSelect`, `StateNameEntry`, `StateCancelled`
  - Provides: `projectTypes` map with type configurations
  - Provides: `getTypeOptions()` helper function

- **Task 010**: Create source handler ✅ (this work unit)
  - Provides: Transition to `StateTypeSelect` when user selects "From branch name"

### Downstream Dependencies (Will Use This Task)

- **Task 030**: Name entry handler
  - Reads: `w.choices["type"]` to determine branch prefix
  - Uses: `getTypePrefix(w.choices["type"])` to generate preview

- **Task 050**: Finalization handler
  - Reads: `w.choices["type"]` to create project with correct type

## Constraints

### Configuration Consistency

- Must use `getTypeOptions()` to ensure options match `projectTypes` map
- Must not hard-code type descriptions (they come from projectTypes)
- Must maintain option order: standard, exploration, design, breakdown, cancel

### State Machine Requirements

- Always store selection in `w.choices["type"]` before transitioning
- Only transition to `StateNameEntry` for valid project types
- Transition to `StateCancelled` for cancel or user abort

### Testing Requirements

- Tests must verify all four project types can be selected
- Tests must verify cancel option works
- Tests must verify user abort is handled gracefully
- Tests must verify `getTypeOptions()` is used (ensures consistency)

### What NOT to Do

- ❌ Don't hard-code type descriptions - use `getTypeOptions()`
- ❌ Don't modify the `projectTypes` map - it's shared infrastructure
- ❌ Don't add new project types here - that requires SDK changes
- ❌ Don't change the option order - it's specified in design docs
- ❌ Don't implement name entry logic here (that's task 030)

## Notes

### Critical Implementation Details

1. **Option Helper Usage**: The `getTypeOptions()` helper ensures consistency between the wizard and the project type configuration. Always use it instead of creating options manually.

2. **Type Storage**: The selected type is stored as a string key (e.g., "exploration"), not the full config. Downstream handlers use `getTypePrefix(type)` to get the prefix.

3. **Cancel vs Abort**: There are two ways to exit: selecting "cancel" (explicit choice) or pressing Ctrl+C/Esc (implicit abort). Both transition to `StateCancelled` but the implementation is slightly different.

### Testing Strategy

Focus on:
- **Coverage**: All four types plus cancel and abort
- **Consistency**: Verify `getTypeOptions()` is used (prevents drift)
- **Behavior**: State transitions and data storage work correctly

Manual testing verifies the UX matches the design specification.

### Future Extensibility

If new project types are added in the future, they must:
1. Be registered in the SDK (in `cli/internal/projects/<type>/`)
2. Be added to the `projectTypes` map in `wizard_helpers.go`
3. Have a prefix and description defined

The wizard will automatically pick up new types through `getTypeOptions()` without code changes here.
