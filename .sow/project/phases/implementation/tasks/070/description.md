# Task 070: Skip Type Selection for GitHub Issues

## Context

During manual testing, a UX issue was discovered: when a user selects a GitHub issue, they are shown the type selection screen with all 4 project type options. However, GitHub issues should **only** create "standard" project types.

**Current Flow** (problematic):
1. User selects issue #71
2. User sees type selection with 4 options
3. User must manually select "Feature work and bug fixes"
4. Proceeds to prompt entry

**Expected Flow** (streamlined):
1. User selects issue #71
2. System defaults to "standard" type automatically
3. Proceeds directly to prompt entry (no type selection screen)

This reduces friction and prevents users from selecting inappropriate project types for GitHub issues.

## Requirements

### Skip Type Selection for GitHub Issue Path

Modify the wizard flow to bypass type selection when coming from the GitHub issue path:

**Changes to `showIssueSelectScreen()`** (wizard_state.go):

After storing the issue in choices, instead of transitioning to `StateTypeSelect`, set the type to "standard" and transition directly to prompt entry:

```go
// Store issue in choices for next steps
w.choices["issue"] = issue
w.choices["type"] = "standard" // NEW: Default to standard for GitHub issues

// Proceed directly to prompt entry (skip type selection)
w.state = StatePromptEntry // CHANGED from StateTypeSelect

return nil
```

**Changes to `handleTypeSelect()`** (wizard_state.go):

The `hasIssue` routing logic will become unreachable for the GitHub issue path, but should be **left in place** for potential future use cases or manual invocation. No changes needed here.

**Alternative Consideration**:

If we want to preserve the ability to manually override the type (for power users), we could add a flag or environment variable. However, for the initial implementation, **hardcoding to "standard" is recommended** for simplicity.

## Acceptance Criteria

### Functional Requirements
- [ ] GitHub issue path defaults type to "standard" automatically
- [ ] GitHub issue path skips type selection screen entirely
- [ ] Transitions directly from issue selection to prompt entry
- [ ] Branch name path still shows type selection (no regression)
- [ ] Branch creation in `createLinkedBranch()` uses "standard" prefix for GitHub issues

### Test Requirements
- [ ] Update `TestShowIssueSelectScreen_NoLinkedBranch` to verify type set and correct transition
- [ ] Add test `TestGitHubIssuePath_SkipsTypeSelection` to validate end-to-end flow
- [ ] Verify `TestHandleTypeSelect_RoutingWithIssue` still passes (or becomes obsolete)
- [ ] All existing tests still pass
- [ ] Manual test: Select issue → verify NO type selection screen → prompt entry appears

### Non-Functional Requirements
- [ ] Branch names generated correctly with "feat/" prefix (standard type)
- [ ] Prompt entry screen still displays correct type context
- [ ] No breaking changes to branch name path

## Relevant Inputs

### Implementation Files
- `cli/cmd/project/wizard_state.go` - Contains showIssueSelectScreen() to modify

### Test Files
- `cli/cmd/project/wizard_state_test.go` - Tests to update/add
- `cli/cmd/project/wizard_integration_test.go` - May need updates to complete workflow test

### Reference
- `.sow/project/phases/review/reports/001.md` - Review report describing the issue
- `.sow/project/context/issue-70.md` - Original specification (doesn't explicitly require type selection)

## Technical Notes

### Type Selection Override (Future Enhancement)

If we later want to support custom types for GitHub issues, we could:
- Add `SOW_ISSUE_TYPE` environment variable
- Add `--type` flag to CLI
- Show type selection if a modifier key is held (Shift+Enter)

For now, **keep it simple with hardcoded "standard"**.

### Branch Prefix Verification

Verify that `createLinkedBranch()` correctly uses `getTypePrefix("standard")` to generate "feat/" prefix:

```go
prefix := getTypePrefix(projectType)  // projectType = "standard"
issueSlug := normalizeName(issue.Title)
branchName := fmt.Sprintf("%s%s-%d", prefix, issueSlug, issue.Number)
// Expected: "feat/issue-title-71"
```

### Prompt Entry Context

The prompt entry screen already shows type context:
```
Issue: #71 - Project Continuation Workflow
Branch: feat/project-continuation-workflow-71
Type: Feature work and bug fixes
```

This will continue to work correctly with the hardcoded "standard" type.

## Examples

### Before (Current Behavior - 3 screens)
1. Issue selection → User selects #71
2. **Type selection** → User selects "Feature work and bug fixes"
3. Prompt entry → User enters optional prompt

### After (Expected Behavior - 2 screens)
1. Issue selection → User selects #71
2. Prompt entry → User enters optional prompt (type defaulted to "standard")

This is 33% fewer screens and a more streamlined experience.
