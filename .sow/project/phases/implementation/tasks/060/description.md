# Task 060: Fix Separate Context Note Screen

## Context

During manual testing, a UX issue was discovered: when a user selects a GitHub issue and proceeds to type selection, the issue context displays as a **separate screen** requiring "enter next" before showing the type selection options. This creates an unnecessary intermediate screen that interrupts the flow.

**Root Cause**: In `handleTypeSelect()` (lines 269-286 in wizard_state.go), the context Note is added as a separate form group:

```go
if contextNote != nil {
    groups = append(groups, huh.NewGroup(contextNote))
}
```

This causes huh to display each group as its own screen.

**Expected Behavior**: The issue context should either:
- Option A: Be removed entirely (simplest fix)
- Option B: Be displayed WITH the type selection on the same screen (combined in one group)

Given that the context will be shown again on the prompt entry screen anyway, **Option A (removal) is recommended** for the cleanest UX.

## Requirements

### Remove Separate Context Note Screen

Update `handleTypeSelect()` in `wizard_state.go` to remove the separate context note functionality:

1. Remove the contextNote variable and conditional logic (lines 269-275)
2. Remove the conditional group appending (lines 279-281)
3. Simplify to single group with just the type selection
4. Update the test `TestHandleTypeSelect_RoutingWithIssue` to remove context note expectations

**Before** (problematic):
```go
// Check if we have issue context
var contextNote *huh.Note
issue, hasIssue := w.choices["issue"].(*sow.Issue)
if hasIssue {
    contextNote = huh.NewNote().
        Description(fmt.Sprintf("Issue: #%d - %s", issue.Number, issue.Title))
}

// Build form groups
groups := []*huh.Group{}

// Add context note if present
if contextNote != nil {
    groups = append(groups, huh.NewGroup(contextNote))
}

// Add type selection
groups = append(groups, huh.NewGroup(
    huh.NewSelect[string]().
        Title("What type of project?").
        Options(getTypeOptions()...).
        Value(&selectedType),
))

form := huh.NewForm(groups...)
```

**After** (clean):
```go
// Check if we have issue context
issue, hasIssue := w.choices["issue"].(*sow.Issue)

// Build form with just type selection
form := huh.NewForm(
    huh.NewGroup(
        huh.NewSelect[string]().
            Title("What type of project?").
            Options(getTypeOptions()...).
            Value(&selectedType),
    ),
)
```

## Acceptance Criteria

### Functional Requirements
- [ ] Context note code removed from `handleTypeSelect()`
- [ ] Type selection shows as single screen (no intermediate screen)
- [ ] Issue path routing still works correctly (hasIssue check preserved)
- [ ] Issue context still displays on prompt entry screen (unchanged)

### Test Requirements
- [ ] Update `TestHandleTypeSelect_RoutingWithIssue` to remove context expectations
- [ ] All existing tests still pass
- [ ] Manual test: Select GitHub issue → type selection appears immediately with no intermediate screen

## Relevant Inputs

### Implementation Files
- `cli/cmd/project/wizard_state.go` - Contains handleTypeSelect() to modify

### Test Files
- `cli/cmd/project/wizard_state_test.go` - Contains TestHandleTypeSelect_RoutingWithIssue to update

### Reference
- `.sow/project/phases/review/reports/001.md` - Review report describing the issue

## Notes

- The issue context is already displayed on the prompt entry screen, so removing it from type selection doesn't lose information
- This fix improves UX by reducing the number of screens in the flow
- The `hasIssue` variable must be preserved as it's used for routing logic
