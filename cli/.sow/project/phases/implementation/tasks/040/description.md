# Fix MultiSelect Title Display Issue

## Context

During manual testing of the file selector feature (tasks 010-030), a critical UX issue was discovered: the title set on the `huh.MultiSelect` component is not rendering in the UI.

**Current Behavior:**
- Only the description line "Type to filter • Space to select • Enter to confirm" displays
- The title "Select knowledge files to provide context (optional):" is completely missing

**Impact:**
- Users don't know what they're selecting
- Users don't know that selection is optional
- Poor user experience and discoverability

**Root Cause (Suspected):**
- Related to huh library issue: "If no title is added to the menu, the filter text will not be displayed"
- Our code DOES set `.Title()` but it's not rendering
- Using huh v0.8.0

## Requirements

Fix the `handleFileSelect` method to ensure the title displays properly in the UI.

### Functional Requirements

1. **Title must be visible** in the file selection UI
2. **Title text** should read: "Select knowledge files to provide context (optional):"
3. **Description** should remain: "Type to filter • Space to select • Enter to confirm"
4. **All existing functionality** must continue to work (filtering, selection, cancellation)
5. **All existing tests** must continue to pass

### Investigation Steps

1. Research huh library documentation for MultiSelect title configuration
2. Check if title needs to be set on the Group instead of/in addition to the MultiSelect
3. Test different title configuration approaches
4. Consider whether huh library needs to be upgraded

### Possible Solutions

**Option 1: Add title to Group**
```go
huh.NewForm(
    huh.NewGroup(
        huh.NewMultiSelect[string]().
            Title("Select knowledge files to provide context (optional):").
            ...
    ).Title("File Selection"),  // Add title here
)
```

**Option 2: Use different huh configuration**
- Check for other MultiSelect configuration options
- Review huh examples and documentation

**Option 3: Upgrade huh library**
- Check if newer version has fix
- Test with upgraded version
- Ensure no breaking changes

**Option 4: Use huh.NewNote or similar for title**
- Add separate Note element above MultiSelect to display title
- Keep MultiSelect without title

## Acceptance Criteria

### Functional Requirements

1. **Title displays in UI**:
   - Title text is visible above the file list
   - Title reads: "Select knowledge files to provide context (optional):"
   - Title is clearly readable and well-formatted

2. **Description displays correctly**:
   - Description remains visible
   - Description reads: "Type to filter • Space to select • Enter to confirm"

3. **All existing functionality works**:
   - File filtering works
   - File selection works (space to toggle)
   - Multiple files can be selected
   - Enter confirms selection
   - Ctrl+C cancels (transitions to StateCancelled)
   - Empty selection is valid
   - State transitions correctly to StatePromptEntry

4. **Edge cases still handled**:
   - Empty knowledge directory: skips selection
   - Non-existent directory: skips selection
   - Test mode (SOW_TEST=1): bypasses interactive form

### Test Requirements

1. **Manual testing required**:
   - Run wizard with actual knowledge files present
   - Verify title displays correctly
   - Verify description displays correctly
   - Test filtering, selection, and all interactions

2. **Automated tests**:
   - All existing tests must continue to pass
   - No new automated tests required (this is a UI rendering fix)

3. **Verification checklist**:
   - [ ] Title visible in UI
   - [ ] Title text correct
   - [ ] Description visible
   - [ ] Filtering works
   - [ ] Selection works
   - [ ] All existing tests pass

### Code Quality

- Minimal changes to existing code
- Follow existing huh patterns in codebase
- Add code comments if non-obvious solution
- No breaking changes to API or behavior

## Technical Details

### Current Implementation

File: `cli/cmd/project/wizard_state.go:417-447`

```go
func (w *Wizard) handleFileSelect() error {
    // ... discovery logic ...

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewMultiSelect[string]().
                Title("Select knowledge files to provide context (optional):").
                Description("Type to filter • Space to select • Enter to confirm").
                Options(options...).
                Value(&selectedFiles).
                Filterable(true).
                Limit(10),
        ),
    )

    // ... error handling ...
}
```

### Investigation Notes

**Check huh library patterns:**
- Review how other huh.MultiSelect instances are configured in the wild
- Check huh examples and documentation for title configuration
- Search for similar issues in huh GitHub issues

**Check other handlers:**
- `handleNameEntry`, `handlePromptEntry` use huh.NewInput which may handle titles differently
- `handleTypeSelect` uses huh.NewSelect which may have similar or different behavior

**Test systematically:**
1. Add title to Group (in addition to MultiSelect)
2. Remove title from MultiSelect, add only to Group
3. Try different formatting or configuration
4. Check if huh version is factor

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/feat/wizard-select-files/cli/cmd/project/wizard_state.go` - Contains handleFileSelect implementation
- `/Users/josh/code/sow/.sow/worktrees/feat/wizard-select-files/.sow/project/phases/review/reports/001.md` - Review report documenting the issue
- `/Users/josh/code/sow/.sow/worktrees/feat/wizard-select-files/go.mod` - Check huh library version

## Examples

### Expected UI (with title visible)

```
┌─────────────────────────────────────────────────────────────┐
│ Select knowledge files to provide context (optional):      │
│                                                             │
│ Type to filter • Space to select • Enter to confirm        │
│                                                             │
│ [ ] designs/interactive-wizard-ux-flow.md                  │
│ [x] designs/sdk-addbranch-api.md                           │
│ [ ] designs/cli-enhanced-advance.md                        │
│ [x] adrs/001-project-state-machine.md                      │
│ [ ] guides/testing-conventions.md                          │
└─────────────────────────────────────────────────────────────┘
```

### Current UI (title missing)

```
Type to filter • Space to select • Enter to confirm

• adrs/001-consolidate-modes-to-projects.md
• adrs/002-explicit-event-selection.md
...
```

## Dependencies

None - this is a standalone UI fix for task 020.

## Constraints

### Time Constraint

This is a critical UX issue blocking finalization. Should be fixed quickly.

### Compatibility Constraint

- Must work with current huh library version (or justify upgrade)
- Must not break any existing functionality
- Must not change behavior of other wizard screens

### Testing Constraint

- Manual testing is REQUIRED (screenshot/verification needed)
- Automated tests should continue to pass
- Test mode (SOW_TEST=1) must continue to work

## Notes

- This is a bug fix for work completed in task 020
- Review found the issue during manual testing
- All code functionality works correctly; this is purely a display issue
- Fix should be minimal and targeted
