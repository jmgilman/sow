# GitHub Issue Integration - Manual Testing Guide

## Prerequisites
- Repository with GitHub remote
- `gh` CLI installed and authenticated
- Issues in repository with 'sow' label

## Test Scenarios

### Scenario 1: Happy Path - Create from Issue
1. Create test issue in GitHub with 'sow' label
2. Run `sow project`
3. Select "Create new project"
4. Select "From GitHub issue"
5. Verify: Spinner shows "Fetching issues from GitHub..."
6. Verify: Issue appears in list with format "#123: Title"
7. Select the test issue
8. Select project type (e.g., "Standard")
9. Verify: Type selection shows "Issue: #123 - Title"
10. Verify: Spinner shows "Creating linked branch..."
11. Enter optional prompt (or skip)
12. Verify: Success message shows issue link
13. Verify: Claude launches in worktree
14. Verify in GitHub: Issue shows linked branch in UI

**Expected**: Project created successfully with issue context

### Scenario 2: Issue Already Linked
1. Create project from issue (Scenario 1)
2. Exit Claude, run `sow project` again
3. Select "Create new project" → "From GitHub issue"
4. Select the SAME issue
5. Verify: Error shows branch name and suggests "Continue existing project"
6. Press Enter
7. Verify: Returns to issue list
8. Select "Cancel" to exit

**Expected**: Clear error, can select different issue or cancel

### Scenario 3: GitHub CLI Not Installed
1. Temporarily rename `gh` CLI (or unset PATH)
2. Run `sow project`
3. Select "Create new project" → "From GitHub issue"
4. Verify: Error shows installation instructions
5. Verify: Error suggests "From branch name" alternative
6. Press Enter
7. Verify: Returns to source selection
8. Select "From branch name" path
9. Verify: Can create project without GitHub

**Expected**: Helpful error, fallback path works

### Scenario 4: GitHub CLI Not Authenticated
1. Run `gh auth logout`
2. Run `sow project`
3. Select "Create new project" → "From GitHub issue"
4. Verify: Error shows "gh auth login" instructions
5. Run `gh auth login` in another terminal
6. Run `sow project` again
7. Verify: Can now access issues

**Expected**: Clear auth instructions, can retry after fixing

### Scenario 5: No Issues with 'sow' Label
1. Ensure no issues in repo have 'sow' label
2. Run `sow project`
3. Select "Create new project" → "From GitHub issue"
4. Verify: Error explains no issues found
5. Verify: Error explains how to create and label issue
6. Press Enter
7. Verify: Returns to source selection

**Expected**: Helpful guidance on creating labeled issues

### Scenario 6: External Editor Integration
1. Set `EDITOR` environment variable (e.g., `export EDITOR=vim`)
2. Start creating project from issue
3. At prompt entry, press Ctrl+E
4. Verify: Editor opens with temp file
5. Write multi-line prompt:
   ```
   Focus on:
   - Middleware implementation
   - Integration tests
   - Error handling
   ```
6. Save and exit editor
7. Verify: Content appears in wizard
8. Press Enter
9. Verify: Project created with full prompt

**Expected**: External editor works, multi-line content preserved

### Scenario 7: Network Error During Fetch
1. Disconnect network (or use network simulator)
2. Run `sow project`
3. Select "Create new project" → "From GitHub issue"
4. Verify: Error shows network issue explanation
5. Reconnect network
6. Return to source selection, try again
7. Verify: Works after network restored

**Expected**: Graceful degradation, can retry

### Scenario 8: Unicode in Issue Title
1. Create issue with Unicode title: "Add 日本語 support"
2. Create project from this issue
3. Verify: Title displays correctly in wizard
4. Verify: Branch name is ASCII-safe: "feat/add-support-123"
5. Verify: Issue context file contains Unicode title

**Expected**: Unicode handled correctly, branch name sanitized

### Scenario 9: Very Long Issue Title
1. Create issue with 200+ character title
2. Create project from this issue
3. Verify: Title displays (may wrap in terminal)
4. Verify: Branch name is reasonable length
5. Verify: Project created successfully

**Expected**: Long titles handled gracefully

### Scenario 10: Branch Name Path Still Works
1. Run `sow project`
2. Select "Create new project" → "From branch name"
3. Select type, enter name, enter prompt
4. Verify: Works exactly as before (no regression)
5. Verify: No issue metadata in project state

**Expected**: Existing functionality unaffected

## Debug Mode Testing

### Scenario 11: Debug Mode Provides Useful Info
1. Set `SOW_DEBUG=1` environment variable
2. Run `sow project`
3. Select "Create new project" → "From GitHub issue"
4. Verify: Debug output shows:
   - State transitions: `[DEBUG] Wizard: State=issue_select`
   - GitHub calls: `[DEBUG] GitHub: Fetched 2 issues`
   - Issue details: `[DEBUG] GitHub: Issue #123: Add JWT auth`
5. Complete project creation
6. Verify: Debug output helps understand wizard flow
7. Unset `SOW_DEBUG` and verify no debug output

**Expected**: Debug mode provides actionable troubleshooting info

### Scenario 12: State Transition Validation in Debug Mode
1. Set `SOW_DEBUG=1`
2. Run `sow project`
3. Go through normal workflow
4. Verify: No warnings about invalid state transitions
5. If warnings appear, report as bug (invalid transition detected)

**Expected**: All transitions are valid, no warnings

## Error Message Quality Testing

### Scenario 13: Error Messages Are Helpful
For each error scenario (2-5, 7), verify error messages include:
1. **What went wrong** - Clear description of the problem
2. **Why/How to fix** - Explanation or troubleshooting steps
3. **Next steps** - Clear action items or alternatives

**Example good error**:
```
Failed to fetch issues from GitHub

This may be due to:
  • Network connectivity issues
  • GitHub API being temporarily unavailable
  • GitHub API rate limits

Please try again in a moment, or select 'From branch name' to continue without GitHub integration.

[Press Enter to return to source selection]
```

**Expected**: Users understand what went wrong and what to do next

## Performance Testing

### Scenario 14: Integration Tests Run Quickly
1. Run integration tests: `go test ./cli/cmd/project -run Integration`
2. Verify: All tests complete in <5 seconds
3. Verify: No flaky failures

**Expected**: Fast, reliable test execution

### Scenario 15: Debug Mode Has Minimal Overhead
1. Time project creation without debug: `time sow project`
2. Time with debug: `SOW_DEBUG=1 time sow project`
3. Compare execution times
4. Verify: <1% performance difference

**Expected**: Debug mode adds negligible overhead

## Regression Testing

### Scenario 16: All Existing Tests Still Pass
1. Run all tests: `go test ./cli/cmd/project -v`
2. Verify: No failures in existing tests
3. Verify: Test coverage >80% for new code

**Expected**: No regressions, good coverage

## Post-Release Testing

### Scenario 17: Real-World Usage
1. Use GitHub issue integration for actual work
2. Create 3-5 projects from real issues
3. Verify: Workflow feels natural
4. Verify: Error handling helps when things go wrong
5. Document any unexpected behaviors or edge cases

**Expected**: Feature works well in practice

## Feedback Collection

After manual testing, document:
- Which scenarios worked perfectly
- Which scenarios had rough edges
- Suggestions for improvement
- New edge cases discovered
- Error messages that could be clearer

This feedback informs future iterations and improvements.
