# Task 030: Implement New GitHub Methods

## Context

This task is part of a refactoring project to extract a GitHubClient interface from the existing GitHub CLI implementation. Tasks 010 and 020 created the interface and renamed the implementation. Now we need to implement the new methods required by the interface that don't exist in the current implementation.

**Project Goal**: Extract GitHubClient interface from existing GitHub CLI implementation to enable dual client support (CLI + API) while maintaining backward compatibility.

**Previous Tasks**:
- Task 010: Created GitHubClient interface
- Task 020: Renamed GitHub to GitHubCLI

**This Task's Role**: Implement three new methods to complete the GitHubClient interface:
1. UpdatePullRequest - Update PR title and body
2. MarkPullRequestReady - Convert draft PR to ready for review
3. Enhanced CreatePullRequest - Add draft support and return PR number

## Requirements

Add/modify the following methods in `cli/internal/sow/github_cli.go`:

**1. UpdatePullRequest (NEW)**
- Update existing PR's title and/or body
- Use gh CLI: `gh pr edit <number> --title <title> --body <body>`
- Return error if command fails
- Follow existing error handling patterns (ErrGHCommand)

**2. MarkPullRequestReady (NEW)**
- Convert a draft PR to ready for review
- Use gh CLI: `gh pr ready <number>`
- Return error if command fails
- Follow existing error handling patterns (ErrGHCommand)

**3. CreatePullRequest (ENHANCE)**
- Add draft parameter to signature
- Return PR number in addition to URL
- Use gh CLI: `gh pr create --title <title> --body <body> [--draft]`
- Parse PR number from returned URL
- Follow pattern from CreateIssue (lines 327-391) for URL parsing

**Method Signatures:**
```go
func (g *GitHubCLI) UpdatePullRequest(number int, title, body string) error
func (g *GitHubCLI) MarkPullRequestReady(number int) error
func (g *GitHubCLI) CreatePullRequest(title, body string, draft bool) (int, string, error)
```

**Error Handling:**
- Call g.Ensure() before any gh CLI operation
- Use ErrGHCommand for command failures
- Include command name and stderr in errors
- Follow patterns from existing methods

**gh CLI Commands:**

UpdatePullRequest:
```bash
gh pr edit <number> --title "New Title" --body "New Body"
```

MarkPullRequestReady:
```bash
gh pr ready <number>
```

CreatePullRequest with draft:
```bash
gh pr create --title "Title" --body "Body" --draft
```

## Acceptance Criteria

- [ ] UpdatePullRequest method implemented in github_cli.go
- [ ] MarkPullRequestReady method implemented in github_cli.go
- [ ] CreatePullRequest signature updated to include draft parameter and return number
- [ ] All three methods call Ensure() before operations
- [ ] All three methods use ErrGHCommand for failures
- [ ] CreatePullRequest parses PR number from URL correctly
- [ ] Unit tests written for UpdatePullRequest (following TDD)
- [ ] Unit tests written for MarkPullRequestReady (following TDD)
- [ ] Unit tests written for CreatePullRequest with draft=true (following TDD)
- [ ] Unit tests written for CreatePullRequest with draft=false (following TDD)
- [ ] All new tests use MockExecutor (no real gh CLI calls)
- [ ] All tests pass: `go test ./cli/internal/sow`
- [ ] Package compiles: `go build ./cli/internal/sow`
- [ ] Methods have godoc comments explaining parameters and return values

## Technical Details

**Implementation Pattern to Follow:**

Look at existing methods like `CreateIssue` (lines 327-391) and `GetIssue` (lines 176-199) for patterns:

1. **Check availability first:**
```go
if err := g.Ensure(); err != nil {
    return err // or appropriate zero values for multi-return
}
```

2. **Build command arguments:**
```go
args := []string{"pr", "edit", fmt.Sprintf("%d", number)}
args = append(args, "--title", title, "--body", body)
```

3. **Execute command:**
```go
stdout, stderr, err := g.gh.Run(args...)
if err != nil {
    return ErrGHCommand{
        Command: "pr edit",
        Stderr:  stderr,
        Err:     err,
    }
}
```

4. **Parse output if needed** (for CreatePullRequest PR number)

**UpdatePullRequest Implementation:**

```go
// UpdatePullRequest updates an existing pull request's title and body.
//
// Parameters:
//   - number: PR number to update
//   - title: New PR title
//   - body: New PR description (supports markdown)
func (g *GitHubCLI) UpdatePullRequest(number int, title, body string) error {
    if err := g.Ensure(); err != nil {
        return err
    }

    stdout, stderr, err := g.gh.Run(
        "pr", "edit", fmt.Sprintf("%d", number),
        "--title", title,
        "--body", body,
    )
    if err != nil {
        return ErrGHCommand{
            Command: fmt.Sprintf("pr edit %d", number),
            Stderr:  stderr,
            Err:     err,
        }
    }

    return nil
}
```

**MarkPullRequestReady Implementation:**

```go
// MarkPullRequestReady marks a draft pull request as ready for review.
//
// Parameters:
//   - number: Draft PR number to mark ready
func (g *GitHubCLI) MarkPullRequestReady(number int) error {
    if err := g.Ensure(); err != nil {
        return err
    }

    stdout, stderr, err := g.gh.Run(
        "pr", "ready", fmt.Sprintf("%d", number),
    )
    if err != nil {
        return ErrGHCommand{
            Command: fmt.Sprintf("pr ready %d", number),
            Stderr:  stderr,
            Err:     err,
        }
    }

    return nil
}
```

**CreatePullRequest Enhancement:**

The current implementation (lines 393-434) returns only URL. Update to:
1. Add draft parameter
2. Include --draft flag if draft=true
3. Parse PR number from URL
4. Return (number, url, error)

URL parsing pattern (from CreateIssue lines 372-382):
```go
// Parse PR number from URL (format: https://github.com/owner/repo/pull/NUMBER)
parts := strings.Split(prURL, "/")
if len(parts) < 1 {
    return 0, "", fmt.Errorf("could not parse PR number from URL: %s", prURL)
}
prNumberStr := parts[len(parts)-1]
prNumber := 0
_, err = fmt.Sscanf(prNumberStr, "%d", &prNumber)
if err != nil {
    return 0, "", fmt.Errorf("could not parse PR number from URL: %s", prURL)
}

return prNumber, prURL, nil
```

**Testing Approach (TDD):**

1. **Write tests first** for each method in `github_cli_test.go`:
   - Test UpdatePullRequest with mock executor
   - Test MarkPullRequestReady with mock executor
   - Test CreatePullRequest with draft=true
   - Test CreatePullRequest with draft=false
   - Test error cases for each method

2. **Test pattern** (from existing tests):
```go
func TestGitHubCLI_UpdatePullRequest(t *testing.T) {
    var capturedArgs []string

    mock := &exec.MockExecutor{
        ExistsFunc: func() bool { return true },
        RunSilentFunc: func(_ ...string) error {
            return nil // Auth check passes
        },
        RunFunc: func(args ...string) (string, string, error) {
            capturedArgs = args
            return "", "", nil // Success
        },
    }

    gh := sow.NewGitHubCLI(mock)
    err := gh.UpdatePullRequest(123, "New Title", "New Body")

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    // Verify correct command was called
    expected := []string{"pr", "edit", "123", "--title", "New Title", "--body", "New Body"}
    if !reflect.DeepEqual(capturedArgs, expected) {
        t.Errorf("expected args %v, got %v", expected, capturedArgs)
    }
}
```

3. **Run tests** to verify implementation

**Location in File:**

Add new methods near existing PR operations (after CreatePullRequest, around line 434).

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/.sow/project/context/issue-90.md` - Complete project requirements (Step 2, lines 212-227; gh commands: lines 562-584)
- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/cli/internal/sow/github_cli.go` - Implementation file (will exist after Task 020)
- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/cli/internal/sow/github.go` - Current implementation patterns (CreateIssue: lines 327-391, CreatePullRequest: lines 393-434)
- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/cli/internal/sow/github_cli_test.go` - Test file (will exist after Task 020)
- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/cli/internal/sow/github_test.go` - Current test patterns (lines 12-54)
- `/Users/josh/code/sow/.sow/worktrees/feat/github-client-interface-extraction-refactoring-90/cli/internal/exec/mock.go` - MockExecutor usage pattern (lines 16-23)

## Examples

**UpdatePullRequest Usage:**
```go
gh := sow.NewGitHubCLI(exec.NewLocal("gh"))
err := gh.UpdatePullRequest(42, "Updated Title", "Updated description")
if err != nil {
    // Handle error
}
```

**MarkPullRequestReady Usage:**
```go
gh := sow.NewGitHubCLI(exec.NewLocal("gh"))
err := gh.MarkPullRequestReady(42)
if err != nil {
    // Handle error
}
```

**CreatePullRequest Usage:**
```go
gh := sow.NewGitHubCLI(exec.NewLocal("gh"))

// Create draft PR
number, url, err := gh.CreatePullRequest("WIP: Feature", "Still in progress", true)
if err != nil {
    // Handle error
}
fmt.Printf("Created draft PR #%d: %s\n", number, url)

// Later, mark it ready
err = gh.MarkPullRequestReady(number)
```

**Test Example for CreatePullRequest with draft:**
```go
func TestGitHubCLI_CreatePullRequest_Draft(t *testing.T) {
    mock := &exec.MockExecutor{
        ExistsFunc: func() bool { return true },
        RunSilentFunc: func(_ ...string) error { return nil },
        RunFunc: func(args ...string) (string, string, error) {
            // Verify --draft flag is present
            found := false
            for _, arg := range args {
                if arg == "--draft" {
                    found = true
                    break
                }
            }
            if !found {
                t.Error("expected --draft flag in arguments")
            }

            return "https://github.com/owner/repo/pull/42\n", "", nil
        },
    }

    gh := sow.NewGitHubCLI(mock)
    number, url, err := gh.CreatePullRequest("Title", "Body", true)

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if number != 42 {
        t.Errorf("expected PR number 42, got %d", number)
    }
    if url != "https://github.com/owner/repo/pull/42" {
        t.Errorf("unexpected URL: %s", url)
    }
}
```

## Dependencies

**Depends On:**
- Task 010 (Define GitHubClient Interface) - Interface must exist
- Task 020 (Rename Implementation to GitHubCLI) - File must be renamed and struct updated

**Reason**: These methods complete the GitHubClient interface implementation.

## Constraints

**DO NOT:**
- Break existing CreatePullRequest callers (signature change is OK if they're updated)
- Change existing method behavior beyond what's required
- Skip error handling (always call Ensure() first)
- Skip tests (TDD approach required)
- Make real gh CLI calls in tests (use MockExecutor)

**DO:**
- Follow existing code patterns from github.go
- Use ErrGHCommand for command failures
- Write comprehensive unit tests first (TDD)
- Document all parameters and return values
- Handle edge cases (empty title/body, invalid PR numbers)

**Testing Requirements:**
- Write tests FIRST (TDD methodology)
- Use MockExecutor for all tests
- Test both success and failure cases
- Verify correct gh CLI commands are called
- Test PR number parsing from URL

**gh CLI Behavior Notes:**
- `gh pr edit` can update title and/or body independently
- `gh pr ready` only works on draft PRs
- `gh pr create --draft` creates PR in draft state
- All commands output to stdout
- PR URLs are on the last line of output
