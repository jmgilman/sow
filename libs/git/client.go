package git

//go:generate go run github.com/matryer/moq@latest -out mocks/client.go -pkg mocks . GitHubClient

// GitHubClient defines operations for GitHub issue, PR, and branch management.
//
// This interface enables multiple client implementations:
//   - GitHubCLI: Wraps the gh CLI tool for local development
//   - Future: GitHubAPI for web VMs or CI/CD environments
//
// Use NewGitHubClient() factory for automatic environment detection.
//
//nolint:revive // Name matches established API pattern from cli/internal/sow
type GitHubClient interface {
	// CheckAvailability verifies that GitHub access is available and ready.
	// Returns ErrGHNotInstalled or ErrGHNotAuthenticated on failure.
	CheckAvailability() error

	// ListIssues returns issues matching the specified label and state.
	// state can be "open", "closed", or "all". Returns up to 1000 issues.
	ListIssues(label, state string) ([]Issue, error)

	// GetIssue retrieves a single issue by number.
	GetIssue(number int) (*Issue, error)

	// CreateIssue creates a new GitHub issue.
	CreateIssue(title, body string, labels []string) (*Issue, error)

	// GetLinkedBranches returns branches linked to an issue.
	GetLinkedBranches(number int) ([]LinkedBranch, error)

	// CreateLinkedBranch creates a branch linked to an issue.
	// branchName can be empty for auto-generated name.
	CreateLinkedBranch(issueNumber int, branchName string, checkout bool) (string, error)

	// CreatePullRequest creates a PR, optionally as draft.
	// Returns PR number, URL, and error.
	CreatePullRequest(title, body string, draft bool) (number int, url string, err error)

	// UpdatePullRequest updates an existing PR's title and body.
	UpdatePullRequest(number int, title, body string) error

	// MarkPullRequestReady converts a draft PR to ready for review.
	MarkPullRequestReady(number int) error
}
