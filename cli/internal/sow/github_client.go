package sow

// GitHubClient defines operations for GitHub issue, PR, and branch management.
//
// This interface enables dual client support for the sow CLI:
//   - GitHubCLI: Wraps the gh CLI tool for local development environments
//   - GitHubAPI: Uses REST/GraphQL APIs for web VMs (future implementation)
//
// The appropriate implementation is selected automatically based on the environment.
// Use the NewGitHubClient() factory function for automatic environment detection,
// which will choose between CLI and API implementations based on the presence of
// the GITHUB_TOKEN environment variable.
//
// Example usage:
//
//	client, err := sow.NewGitHubClient()
//	if err != nil {
//	    return err
//	}
//	issues, err := client.ListIssues("bug", "open")
type GitHubClient interface {
	// CheckAvailability verifies that GitHub access is available and ready for use.
	//
	// For CLI implementation: Checks that the gh CLI tool is installed and authenticated.
	// Returns ErrGHNotInstalled if gh is not found in PATH, or ErrGHNotAuthenticated
	// if gh is installed but not authenticated.
	//
	// For API implementation: Validates that the GitHub token is present and has
	// sufficient permissions, and verifies connectivity to the GitHub API.
	//
	// This should be called before performing GitHub operations to provide clear
	// error messages to users about what needs to be configured.
	CheckAvailability() error

	// ListIssues returns issues matching the specified label and state filters.
	//
	// Parameters:
	//   - label: Filter issues by label (e.g., "bug", "enhancement", "sow")
	//   - state: Filter by state - use "open", "closed", or "all"
	//
	// Returns a slice of up to 1000 issues matching the criteria. Returns an empty
	// slice if no issues match the filters (not an error condition).
	//
	// Returns an error if the GitHub client is not available or if the API/CLI
	// operation fails.
	ListIssues(label, state string) ([]Issue, error)

	// GetIssue retrieves detailed information about a single issue by its number.
	//
	// Parameters:
	//   - number: The issue number to retrieve
	//
	// Returns a pointer to the Issue with all fields populated (number, title, body,
	// state, URL, and labels). Returns nil and an error if the issue is not found or
	// if the operation fails.
	GetIssue(number int) (*Issue, error)

	// CreateIssue creates a new GitHub issue with the specified title, body, and labels.
	//
	// Parameters:
	//   - title: Issue title (required, non-empty)
	//   - body: Issue description/body text (supports GitHub-flavored markdown)
	//   - labels: List of label names to apply to the issue (can be empty)
	//
	// Returns a pointer to the created Issue with number, title, body, URL, and state
	// fields populated. The issue will be created in the "open" state.
	//
	// Returns an error if the GitHub client is not available or if issue creation fails.
	CreateIssue(title, body string, labels []string) (*Issue, error)

	// GetLinkedBranches returns all branches that are linked to the specified issue.
	//
	// Parameters:
	//   - number: The issue number to query for linked branches
	//
	// Returns a slice of LinkedBranch structs containing branch names and URLs.
	// Returns an empty slice if no branches are linked (not an error condition).
	//
	// Returns an error if the GitHub client is not available or if the operation fails.
	GetLinkedBranches(number int) ([]LinkedBranch, error)

	// CreateLinkedBranch creates a new branch linked to the specified issue.
	//
	// The branch will be associated with the issue in GitHub's issue tracking system,
	// enabling automatic PR linking and issue references.
	//
	// Parameters:
	//   - issueNumber: The issue number to link the branch to
	//   - branchName: Custom branch name (use empty string for auto-generated name based on issue title)
	//   - checkout: If true, checks out the branch after creation; if false, creates but doesn't check out
	//
	// Returns the name of the created branch. If branchName is empty, GitHub will auto-generate
	// a name in the format "<issue-number>-<title-kebab-case>".
	//
	// Returns an error if the GitHub client is not available, if the issue doesn't exist,
	// or if branch creation fails.
	CreateLinkedBranch(issueNumber int, branchName string, checkout bool) (string, error)

	// CreatePullRequest creates a new pull request, optionally as a draft.
	//
	// The PR is created from the current branch to the repository's default base branch.
	//
	// Parameters:
	//   - title: Pull request title (required, non-empty)
	//   - body: Pull request description (supports GitHub-flavored markdown)
	//   - draft: If true, creates the PR as a draft; if false, creates as ready for review
	//
	// Returns:
	//   - number: The PR number, which can be used with UpdatePullRequest and MarkPullRequestReady
	//   - url: Full GitHub URL to the pull request (e.g., "https://github.com/owner/repo/pull/42")
	//   - error: Any error during PR creation
	//
	// Draft PRs can be converted to ready for review using MarkPullRequestReady.
	//
	// Returns an error if the GitHub client is not available, if there are no commits to create
	// a PR from, or if PR creation fails.
	CreatePullRequest(title, body string, draft bool) (number int, url string, err error)

	// UpdatePullRequest updates the title and/or body of an existing pull request.
	//
	// This is useful for updating PR descriptions as implementation progresses, or for
	// correcting the title after creation.
	//
	// Parameters:
	//   - number: The PR number to update
	//   - title: New PR title (required, non-empty)
	//   - body: New PR description (supports GitHub-flavored markdown)
	//
	// Returns an error if the GitHub client is not available, if the PR doesn't exist,
	// or if the update operation fails.
	UpdatePullRequest(number int, title, body string) error

	// MarkPullRequestReady converts a draft pull request to "ready for review" state.
	//
	// This operation is only valid for PRs that were created as drafts. Calling this
	// on a PR that is already ready for review is typically a no-op or may return an error
	// depending on the implementation.
	//
	// Parameters:
	//   - number: The PR number to mark as ready
	//
	// Returns an error if the GitHub client is not available, if the PR doesn't exist,
	// if the PR is not a draft, or if the operation fails.
	MarkPullRequestReady(number int) error
}
