# ADR: GitHub Client Dual Implementation Strategy

## Status

Proposed

## Context

sow uses GitHub operations extensively for project workflows:
- Listing and creating issues
- Linking branches to issues (GitHub's linked branch feature)
- Creating pull requests (draft and ready states)
- Updating pull request content as work progresses
- Marking draft PRs as ready for review
- Retrieving issue metadata for project initialization

The **standard project workflow** follows this pattern:
1. Create draft pull request (early in implementation phase)
2. Perform implementation work across multiple tasks
3. Update PR body as tasks complete
4. Mark PR as ready for review (when all tasks complete)

Currently (per ADR-010), sow exclusively uses the `gh` CLI for all GitHub operations. This design decision was made to:
- Leverage user's existing GitHub authentication (via `gh auth login`)
- Avoid implementing a custom API client
- Simplify the codebase by delegating to a well-maintained tool

However, this creates a fundamental incompatibility with Claude Code web environments:

### The Web VM Challenge

Claude Code web runs in ephemeral Linux VMs with these characteristics:
- **Sandboxed**: Limited pre-installed tools (git, curl, Python, Node.js, Go)
- **GitHub proxy authentication**: Secure proxy handles git operations (clone, push, pull)
- **No gh CLI**: Not pre-installed, would require manual installation
- **No gh authentication**: Even if installed, `gh auth login` requires interactive flow

The GitHub proxy in web VMs:
- ✅ Handles git-level operations (clone, push, PR creation via git)
- ❌ Does not provide gh CLI access
- ❌ Does not expose credentials to gh CLI

### Authentication Options in Web VMs

1. **Install gh + authenticate**: Complex, requires user to generate PAT and configure
2. **Use GitHub API directly**: Requires GITHUB_TOKEN but is straightforward
3. **Hybrid approach**: Use gh locally, API in web VMs

### Current State

Without a solution, sow is **non-functional in web VMs** for any workflow involving issues:
- `sow project new --issue N` fails (can't fetch issue)
- Breakdown projects fail (can't list/create issues)
- Branch linking fails (gh CLI only feature)

Basic workflows work (create branch, code, push) but lack GitHub integration.

## Decision

**Implement a dual GitHub client with interface abstraction and automatic environment detection.**

### Architecture

Create a `GitHubClient` interface with two implementations:

```go
// Interface (abstraction)
type GitHubClient interface {
    CheckAvailability() error
    ListIssues(label, state string) ([]Issue, error)
    GetIssue(number int) (*Issue, error)
    CreateIssue(title, body string, labels []string) (*Issue, error)
    GetLinkedBranches(number int) ([]LinkedBranch, error)
    CreateLinkedBranch(issueNumber int, branchName string, checkout bool) (string, error)
    CreatePullRequest(title, body string, draft bool) (number int, url string, error)
    UpdatePullRequest(number int, title, body string) error
    MarkPullRequestReady(number int) error
}

// Implementation 1: gh CLI wrapper (existing code, renamed)
type GitHubCLI struct {
    executor exec.Executor
}

// Implementation 2: API client (new code)
type GitHubAPI struct {
    client   *github.Client  // REST API (go-github)
    clientv4 *githubv4.Client // GraphQL (githubv4)
    owner    string
    repo     string
}
```

### Auto-Detection Factory

```go
func NewGitHubClient() (GitHubClient, error) {
    // Check for GITHUB_TOKEN environment variable
    if token := os.Getenv("GITHUB_TOKEN"); token != "" {
        // API mode: Extract owner/repo from git remote
        owner, repo, err := getRepoFromGit()
        if err != nil {
            return nil, fmt.Errorf("failed to determine repo: %w", err)
        }
        return NewGitHubAPI(token, owner, repo), nil
    }

    // CLI mode: Use gh binary (existing behavior)
    return NewGitHubCLI(exec.NewLocal("gh")), nil
}
```

### User Experience

**Web VM Users**:
1. Clone repository in Claude Code web
2. Set `GITHUB_TOKEN` environment variable (via web UI, with PAT)
3. sow automatically uses API mode
4. All GitHub operations work transparently

**Local CLI Users** (unchanged):
1. Authenticate gh CLI (existing: `gh auth login`)
2. sow automatically uses CLI mode
3. Existing workflow preserved

**Explicit API Mode** (local testing):
1. Set `GITHUB_TOKEN` environment variable
2. sow uses API mode even locally
3. Useful for testing API implementation

### API Mapping

All gh CLI operations have GitHub API equivalents:

| Operation | gh CLI Command | GitHub API |
|-----------|----------------|------------|
| List issues | `gh issue list --label X --state Y` | `GET /repos/{owner}/{repo}/issues` |
| Get issue | `gh issue view N` | `GET /repos/{owner}/{repo}/issues/{number}` |
| Create issue | `gh issue create --title X --body Y` | `POST /repos/{owner}/{repo}/issues` |
| Create PR | `gh pr create --title X --body Y` | `POST /repos/{owner}/{repo}/pulls` |
| Create draft PR | `gh pr create --draft --title X --body Y` | `POST /repos/{owner}/{repo}/pulls` with `draft: true` |
| Update PR | `gh pr edit N --title X --body Y` | `PATCH /repos/{owner}/{repo}/pulls/{number}` |
| Mark PR ready | `gh pr ready N` | `PATCH /repos/{owner}/{repo}/pulls/{number}` with `draft: false` |
| Link branch | `gh issue develop N --name branch` | GraphQL `createLinkedBranch` mutation |

**Note**: Branch linking requires GraphQL. The REST API doesn't support linked branches.

### Dependencies

New dependencies for API implementation:
- `github.com/google/go-github/v66` - Official GitHub REST API client (~2MB)
- `github.com/shurcooL/githubv4` - GraphQL client (~1MB)
- `golang.org/x/oauth2` - OAuth2 token authentication (~500KB)

Total binary size increase: ~3-4MB (acceptable for full web VM support)

### Migration Path

This is a **refactoring**, not a rewrite:

1. **Extract interface** (no behavior change):
   - Define `GitHubClient` interface with current method signatures
   - Add compile-time check: `var _ GitHubClient = (*GitHub)(nil)`
   - Verify all tests pass

2. **Rename implementation** (no behavior change):
   - Rename `GitHub` struct to `GitHubCLI`
   - Update references in codebase
   - Verify all tests pass

3. **Add API implementation** (new functionality):
   - Create `GitHubAPI` struct
   - Implement all interface methods using go-github and githubv4
   - Add unit tests with mocked HTTP responses

4. **Add factory** (enable auto-detection):
   - Create `NewGitHubClient()` factory function
   - Update `Context` initialization to use factory
   - Add integration tests for both modes

**Zero breaking changes**: Consumers use the interface, don't know which implementation.

## Consequences

### Positive

1. **Web VM compatibility**: sow fully functional in Claude Code web with GITHUB_TOKEN
2. **Backward compatible**: Local CLI users see zero changes
3. **Improved testability**: Mock interface instead of exec calls
4. **Flexibility**: Can switch implementations or add new ones
5. **Single codebase**: No web-specific forks or conditional compilation
6. **User choice**: Local users can opt into API mode by setting token

### Negative

1. **Binary size**: +3-4MB for API client dependencies
2. **Maintenance**: Two implementations to maintain (but interface ensures parity)
3. **API rate limits**: GitHub API has rate limits (though higher with user token)
4. **Implementation effort**: ~15-20 hours to implement and test API client

### Neutral

1. **Performance**: API may be slightly slower than gh CLI (network latency vs local exec)
2. **Authentication management**: Users must manage GITHUB_TOKEN in web VMs (vs gh auth once locally)
3. **Error handling**: API errors differ from gh CLI errors (but interface abstracts)

### Trade-offs Considered

#### Alternative 1: gh CLI Only, Make Optional

**Approach**: Keep gh CLI, gracefully degrade when unavailable

**Pros**:
- No code changes
- Simple

**Cons**:
- Broken experience in web VMs (missing core features)
- User frustration when issue workflows fail
- Undermines sow's value proposition

**Rejected**: Unacceptable user experience

#### Alternative 2: API Only

**Approach**: Replace gh CLI with API client everywhere

**Pros**:
- Single implementation
- Works in all environments

**Cons**:
- Breaking change for local users (must set GITHUB_TOKEN)
- Loses gh CLI's auth UX (seamless `gh auth login`)
- Forces migration on existing users

**Rejected**: Breaking changes violate compatibility requirement

#### Alternative 3: Dual Implementation (Chosen)

**Approach**: Interface abstraction with auto-detection

**Pros**:
- Works in all environments
- Zero breaking changes
- Best UX for each environment
- Testable and maintainable

**Cons**:
- More code to maintain
- Binary size increase

**Chosen**: Best balance of functionality, compatibility, and UX

## Implementation Notes

### API Client Structure

```go
type GitHubAPI struct {
    client   *github.Client    // REST API
    clientv4 *githubv4.Client  // GraphQL
    owner    string            // Repository owner
    repo     string            // Repository name
}

func NewGitHubAPI(token, owner, repo string) *GitHubAPI {
    ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
    tc := oauth2.NewClient(context.Background(), ts)

    return &GitHubAPI{
        client:   github.NewClient(tc),
        clientv4: githubv4.NewClient(tc),
        owner:    owner,
        repo:     repo,
    }
}
```

### Repository Detection

Extract owner/repo from git remote:

```go
func getRepoFromGit() (owner, repo string, err error) {
    cmd := exec.Command("git", "remote", "get-url", "origin")
    output, err := cmd.Output()
    if err != nil {
        return "", "", err
    }

    // Parse: git@github.com:owner/repo.git
    // or: https://github.com/owner/repo.git
    remote := strings.TrimSpace(string(output))

    // Extract owner/repo using regex or string parsing
    // ...

    return owner, repo, nil
}
```

### Error Handling

Map API errors to user-friendly messages:

```go
func (g *GitHubAPI) GetIssue(number int) (*Issue, error) {
    issue, resp, err := g.client.Issues.Get(context.Background(), g.owner, g.repo, number)
    if err != nil {
        if resp != nil && resp.StatusCode == 404 {
            return nil, fmt.Errorf("issue #%d not found", number)
        }
        if resp != nil && resp.StatusCode == 401 {
            return nil, fmt.Errorf("authentication failed: check GITHUB_TOKEN")
        }
        return nil, fmt.Errorf("failed to get issue: %w", err)
    }
    return convertIssue(issue), nil
}
```

### Pull Request Workflow Methods

Example implementations for the draft PR workflow:

```go
func (g *GitHubAPI) CreatePullRequest(title, body string, draft bool) (number int, url string, error) {
    pr := &github.NewPullRequest{
        Title: github.String(title),
        Body:  github.String(body),
        Head:  github.String(getCurrentBranch()), // helper function
        Base:  github.String("main"),
        Draft: github.Bool(draft),
    }

    created, resp, err := g.client.PullRequests.Create(context.Background(), g.owner, g.repo, pr)
    if err != nil {
        return 0, "", fmt.Errorf("failed to create PR: %w", err)
    }

    return created.GetNumber(), created.GetHTMLURL(), nil
}

func (g *GitHubAPI) UpdatePullRequest(number int, title, body string) error {
    pr := &github.PullRequest{
        Title: github.String(title),
        Body:  github.String(body),
    }

    _, _, err := g.client.PullRequests.Edit(context.Background(), g.owner, g.repo, number, pr)
    if err != nil {
        return fmt.Errorf("failed to update PR #%d: %w", number, err)
    }

    return nil
}

func (g *GitHubAPI) MarkPullRequestReady(number int) error {
    // Mark as ready by setting draft to false
    pr := &github.PullRequest{
        Draft: github.Bool(false),
    }

    _, _, err := g.client.PullRequests.Edit(context.Background(), g.owner, g.repo, number, pr)
    if err != nil {
        return fmt.Errorf("failed to mark PR #%d ready: %w", number, err)
    }

    return nil
}
```

**Note**: The `CreatePullRequest` method now returns both the PR number (for subsequent updates) and the URL (for display to user).

### Testing Strategy

**Unit Tests**:
- Mock GitHub API responses using `httptest`
- Test error handling for common scenarios (404, 401, rate limits)
- Verify interface compliance for both implementations

**Integration Tests**:
- Test with real GitHub API (using test repository)
- Test gh CLI mode (existing tests)
- Test factory detection logic with different environments

**Manual Validation**:
- Test in Claude Code web VM with GITHUB_TOKEN
- Verify all operations work (issues, PRs, branch linking)
- Verify local CLI unchanged

## Security Considerations

### Token Handling

- Token read from environment variable (not stored on disk)
- Token not logged or exposed in error messages
- Token lifecycle managed by user (web UI or local env)

### Rate Limiting

- User tokens have higher rate limits (5000 req/hour vs 60 unauthenticated)
- API client should handle rate limit errors gracefully
- Consider caching responses where appropriate (future optimization)

### Scope Requirements

GitHub token requires these scopes:
- `repo` - Full repository access (read/write issues, PRs, code)
- `workflow` - Workflow access (if needed for PR creation with workflows)

Document in setup guide and `.env.example`.

## Documentation Requirements

1. **Web VM Setup Guide**:
   - How to generate GitHub personal access token
   - Setting GITHUB_TOKEN in Claude Code web UI
   - Troubleshooting API errors

2. **Developer Guide**:
   - GitHubClient interface contract
   - Adding new GitHub operations
   - Testing both implementations

3. **Migration Guide** (for contributors):
   - How interface extraction works
   - Running tests with API mode locally
   - Debugging API client issues

4. **Example Configuration**:
   ```bash
   # .env.example
   # GitHub personal access token for API access
   # Required in Claude Code web, optional locally
   # Scopes: repo, workflow
   GITHUB_TOKEN=ghp_your_token_here
   ```

## Success Metrics

This decision is successful when:

1. ✅ Web VM users can run all sow commands with GITHUB_TOKEN
2. ✅ Local CLI users see no changes (gh CLI still works)
3. ✅ Zero breaking changes in API (interface compatibility)
4. ✅ Test coverage >80% for API client
5. ✅ Binary size increase <5MB
6. ✅ API operations complete in <3 seconds (reasonable performance)
7. ✅ Draft PR workflow functions correctly (create draft, update, mark ready)

## References

### Related Documents

- Architecture Design: Claude Code Web Integration
- Exploration Summary: `.sow/knowledge/explorations/integrating_claude_code_web.md`
- Task 013 Findings: gh CLI dependency analysis
- Task 014 Findings: GitHub client interface design

### External Documentation

- [GitHub REST API](https://docs.github.com/en/rest)
- [GitHub GraphQL API](https://docs.github.com/en/graphql)
- [go-github](https://pkg.go.dev/github.com/google/go-github/v66/github)
- [githubv4](https://pkg.go.dev/github.com/shurcooL/githubv4)
- [oauth2](https://pkg.go.dev/golang.org/x/oauth2)
- [Claude Code on the Web](https://code.claude.com/docs/en/claude-code-on-the-web)

### Code Locations

**Current**:
- `cli/internal/sow/github.go` - GitHub client (to be refactored)

**New**:
- `cli/internal/sow/github_client.go` - GitHubClient interface
- `cli/internal/sow/github_cli.go` - gh CLI implementation
- `cli/internal/sow/github_api.go` - API implementation
- `cli/internal/sow/github_factory.go` - Factory with auto-detection
- `cli/internal/sow/github_test.go` - Interface compliance tests

## Revision History

- 2025-11-07: Initial proposal based on exploration findings
