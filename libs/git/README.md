# sow Git

Git and GitHub operations library for sow.

## Quick Start

```go
import "github.com/jmgilman/sow/libs/git"

// Git operations
g, err := git.NewGit("/path/to/repo")
if err != nil {
    log.Fatal(err)
}

branch, _ := g.CurrentBranch()
fmt.Printf("Current branch: %s\n", branch)

// GitHub operations
client, err := git.NewGitHubClient()
if err != nil {
    log.Fatal(err)
}

issues, _ := client.ListIssues("bug", "open")
for _, issue := range issues {
    fmt.Printf("#%d: %s\n", issue.Number, issue.Title)
}
```

## Usage

### Create a Git Instance

```go
g, err := git.NewGit("/path/to/repo")
if err != nil {
    var notRepo git.ErrNotGitRepository
    if errors.As(err, &notRepo) {
        fmt.Printf("%s is not a git repository\n", notRepo.Path)
    }
    return err
}

// Check current branch
branch, err := g.CurrentBranch()

// Check for protected branches
if g.IsProtectedBranch(branch) {
    fmt.Println("Warning: on protected branch")
}

// Check for uncommitted changes
hasChanges, err := g.HasUncommittedChanges()
```

### GitHub Operations

```go
// Create client (auto-detects environment)
client, err := git.NewGitHubClient()
if err != nil {
    return err
}

// Check availability
if err := client.CheckAvailability(); err != nil {
    return fmt.Errorf("GitHub not available: %w", err)
}

// List issues
issues, err := client.ListIssues("sow", "open")

// Get single issue
issue, err := client.GetIssue(123)

// Create issue
issue, err := client.CreateIssue("Bug title", "Description", []string{"bug"})

// Create PR
prNum, prURL, err := client.CreatePullRequest("My PR", "Description", true)
```

### Worktree Operations

```go
// Get worktree path
wtPath := git.WorktreePath(repoRoot, "feat/auth")

// Check for uncommitted changes first
if err := git.CheckUncommittedChanges(g); err != nil {
    return err
}

// Create worktree
if err := git.EnsureWorktree(g, repoRoot, wtPath, "feat/auth"); err != nil {
    return err
}
```

### Testing with Mocks

```go
import "github.com/jmgilman/sow/libs/git/mocks"

func TestMyService(t *testing.T) {
    mock := &mocks.GitHubClientMock{
        CheckAvailabilityFunc: func() error { return nil },
        ListIssuesFunc: func(label, state string) ([]git.Issue, error) {
            return []git.Issue{
                {Number: 1, Title: "Test Issue"},
            }, nil
        },
    }

    svc := NewMyService(mock)
    // ... test logic
}
```

## Troubleshooting

### "GitHub CLI (gh) not found"

Install the GitHub CLI from https://cli.github.com/

### "GitHub CLI not authenticated"

Run `gh auth login` to authenticate.

### "not a git repository"

Ensure the path points to a directory containing `.git/`.

### Worktree creation fails

- Ensure there are no uncommitted changes in the main repo
- Check that the branch name is valid
- Verify you have write permissions to the worktree path

## Links

- [Go Package Documentation](https://pkg.go.dev/github.com/jmgilman/sow/libs/git)
