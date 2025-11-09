package sow

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/exec"
)

// GitHubCLI implements GitHubClient using the gh CLI tool.
//
// All operations require the GitHub CLI (gh) to be installed and authenticated.
// The client accepts an Executor interface, making it easy to mock in tests.
//
// For auto-detection between CLI and API clients, use NewGitHubClient() factory.
type GitHubCLI struct {
	gh exec.Executor
}

// NewGitHubCLI creates a new GitHub CLI client with the given executor.
//
// The executor should be configured for the "gh" command. For production use:
//
//	ghExec := exec.NewLocal("gh")
//	github := sow.NewGitHubCLI(ghExec)
//
// For testing with a mock:
//
//	mockExec := &MockExecutor{...}
//	github := sow.NewGitHubCLI(mockExec)
//
// Note: This does NOT check if gh is installed or authenticated.
// Those checks happen on first operation via Ensure().
func NewGitHubCLI(executor exec.Executor) *GitHubCLI {
	return &GitHubCLI{
		gh: executor,
	}
}

// NewGitHub creates a GitHub CLI client.
// Deprecated: Use NewGitHubCLI() for explicit CLI client, or NewGitHubClient() for auto-detection.
func NewGitHub(executor exec.Executor) *GitHubCLI {
	return NewGitHubCLI(executor)
}

// Error types for GitHub operations

// ErrGHNotInstalled is returned when the gh CLI is not found.
type ErrGHNotInstalled struct{}

func (e ErrGHNotInstalled) Error() string {
	return "GitHub CLI (gh) not found. Install from: https://cli.github.com/"
}

// ErrGHNotAuthenticated is returned when gh is not authenticated.
type ErrGHNotAuthenticated struct{}

func (e ErrGHNotAuthenticated) Error() string {
	return "GitHub CLI not authenticated. Run: gh auth login"
}

// ErrGHCommand is returned when a gh command fails.
type ErrGHCommand struct {
	Command string
	Stderr  string
	Err     error
}

func (e ErrGHCommand) Error() string {
	if e.Stderr != "" {
		return fmt.Sprintf("gh %s failed: %s", e.Command, e.Stderr)
	}
	return fmt.Sprintf("gh %s failed: %v", e.Command, e.Err)
}

func (e ErrGHCommand) Unwrap() error {
	return e.Err
}

// Data types

// Issue represents a GitHub issue.
type Issue struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Body   string `json:"body"`
	State  string `json:"state"`
	URL    string `json:"url"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

// LinkedBranch represents a branch linked to an issue.
type LinkedBranch struct {
	Name string
	URL  string
}

// HasLabel checks if an issue has a specific label.
func (i *Issue) HasLabel(label string) bool {
	for _, l := range i.Labels {
		if l.Name == label {
			return true
		}
	}
	return false
}

// Installation and authentication checks

// CheckInstalled verifies that the gh CLI is installed and available.
func (g *GitHubCLI) CheckInstalled() error {
	if !g.gh.Exists() {
		return ErrGHNotInstalled{}
	}
	return nil
}

// CheckAuthenticated verifies that the gh CLI is authenticated.
func (g *GitHubCLI) CheckAuthenticated() error {
	// gh auth status exits with code 1 if not authenticated
	// but writes to stderr in both success and failure cases
	if err := g.gh.RunSilent("auth", "status"); err != nil {
		return ErrGHNotAuthenticated{}
	}

	return nil
}

// Ensure checks that gh is installed and authenticated.
//
// This should be called before any GitHub operation to provide
// clear error messages to the user.
func (g *GitHubCLI) Ensure() error {
	if err := g.CheckInstalled(); err != nil {
		return err
	}
	if err := g.CheckAuthenticated(); err != nil {
		return err
	}
	return nil
}

// CheckAvailability implements GitHubClient.
// For CLI client, this checks that gh is installed and authenticated.
func (g *GitHubCLI) CheckAvailability() error {
	return g.Ensure()
}

// Issue operations

// ListIssues lists issues with the given label and state.
//
// Parameters:
//   - label: Filter issues by label (e.g., "sow")
//   - state: Filter by state ("open", "closed", or "all")
//
// Returns up to 1000 issues matching the criteria.
func (g *GitHubCLI) ListIssues(label, state string) ([]Issue, error) {
	if err := g.Ensure(); err != nil {
		return nil, err
	}

	stdout, stderr, err := g.gh.Run(
		"issue", "list",
		"--label", label,
		"--state", state,
		"--json", "number,title,url,state,labels",
		"--limit", "1000",
	)
	if err != nil {
		return nil, ErrGHCommand{
			Command: fmt.Sprintf("issue list --label %s --state %s", label, state),
			Stderr:  stderr,
			Err:     err,
		}
	}

	var issues []Issue
	if err := json.Unmarshal([]byte(stdout), &issues); err != nil {
		return nil, fmt.Errorf("failed to parse issue list: %w", err)
	}

	return issues, nil
}

// GetIssue retrieves a single issue by number.
func (g *GitHubCLI) GetIssue(number int) (*Issue, error) {
	if err := g.Ensure(); err != nil {
		return nil, err
	}

	stdout, stderr, err := g.gh.Run(
		"issue", "view", fmt.Sprintf("%d", number),
		"--json", "number,title,body,url,state,labels",
	)
	if err != nil {
		return nil, ErrGHCommand{
			Command: fmt.Sprintf("issue view %d", number),
			Stderr:  stderr,
			Err:     err,
		}
	}

	var issue Issue
	if err := json.Unmarshal([]byte(stdout), &issue); err != nil {
		return nil, fmt.Errorf("failed to parse issue: %w", err)
	}

	return &issue, nil
}

// GetLinkedBranches returns branches linked to an issue.
//
// Returns an empty slice if no branches are linked (not an error).
func (g *GitHubCLI) GetLinkedBranches(number int) ([]LinkedBranch, error) {
	if err := g.Ensure(); err != nil {
		return nil, err
	}

	stdout, stderr, err := g.gh.Run(
		"issue", "develop", "--list", fmt.Sprintf("%d", number),
	)
	if err != nil {
		// Check if error is because no branches are linked
		if strings.Contains(stderr, "no linked branches") ||
			strings.Contains(stderr, "No linked branches") {
			return []LinkedBranch{}, nil
		}

		return nil, ErrGHCommand{
			Command: fmt.Sprintf("issue develop --list %d", number),
			Stderr:  stderr,
			Err:     err,
		}
	}

	// Parse output - gh issue develop --list returns tab-separated values
	// Format: BRANCH    URL
	lines := strings.Split(strings.TrimSpace(stdout), "\n")

	var branches []LinkedBranch
	for i, line := range lines {
		// Skip header line
		if i == 0 {
			continue
		}

		// Skip empty lines
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split by whitespace (tabs or spaces)
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			branches = append(branches, LinkedBranch{
				Name: parts[0],
				URL:  parts[1],
			})
		}
	}

	return branches, nil
}

// CreateLinkedBranch creates a branch linked to an issue using gh issue develop.
//
// Parameters:
//   - issueNumber: The issue number to link to
//   - branchName: Custom branch name (empty string for auto-generated)
//   - checkout: Whether to checkout the branch after creation
//
// Returns the branch name that was created.
func (g *GitHubCLI) CreateLinkedBranch(issueNumber int, branchName string, checkout bool) (string, error) {
	if err := g.Ensure(); err != nil {
		return "", err
	}

	args := []string{
		"issue", "develop", fmt.Sprintf("%d", issueNumber),
	}

	if branchName != "" {
		args = append(args, "--name", branchName)
	}

	if checkout {
		args = append(args, "--checkout")
	}

	stdout, stderr, err := g.gh.Run(args...)
	if err != nil {
		return "", ErrGHCommand{
			Command: fmt.Sprintf("issue develop %d", issueNumber),
			Stderr:  stderr,
			Err:     err,
		}
	}

	// Parse output to get branch name
	// gh issue develop outputs: "Created branch <name> based on <base>"
	// or: "Checked out branch <name> for issue #<number>"
	output := stdout

	// If we specified a name, use that
	if branchName != "" {
		return branchName, nil
	}

	// Otherwise, parse from output
	// Look for "Created branch" or "Checked out branch"
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Created branch") || strings.Contains(line, "Checked out branch") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				return parts[2], nil
			}
		}
	}

	// If we can't parse the output, construct the expected branch name
	// gh creates branches in the format: <issue-number>-<title-kebab-case>
	// We'll need to query the issue to get the title
	issue, err := g.GetIssue(issueNumber)
	if err != nil {
		return "", fmt.Errorf("failed to get issue for branch name: %w", err)
	}

	// Convert title to kebab-case
	branchName = fmt.Sprintf("%d-%s", issueNumber, toKebabCase(issue.Title))
	return branchName, nil
}

// Pull request operations

// CreateIssue creates a GitHub issue using gh CLI.
//
// Parameters:
//   - title: Issue title
//   - body: Issue body/description (supports markdown)
//   - labels: List of labels to add to the issue
//
// Returns the created Issue on success.
func (g *GitHubCLI) CreateIssue(title, body string, labels []string) (*Issue, error) {
	if err := g.Ensure(); err != nil {
		return nil, err
	}

	// Build command arguments
	args := []string{"issue", "create", "--title", title, "--body", body}

	// Add labels if provided
	for _, label := range labels {
		args = append(args, "--label", label)
	}

	stdout, stderr, err := g.gh.Run(args...)
	if err != nil {
		return nil, ErrGHCommand{
			Command: "issue create",
			Stderr:  stderr,
			Err:     err,
		}
	}

	// gh issue create outputs the issue URL on the last line
	output := strings.TrimSpace(stdout)
	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("unexpected empty output from gh issue create")
	}

	// The URL is typically the last line
	issueURL := strings.TrimSpace(lines[len(lines)-1])

	// Validate it looks like a URL
	if !strings.HasPrefix(issueURL, "http") {
		return nil, fmt.Errorf("unexpected issue create output (no URL found): %s", output)
	}

	// Parse issue number from URL (format: https://github.com/owner/repo/issues/NUMBER)
	parts := strings.Split(issueURL, "/")
	if len(parts) < 1 {
		return nil, fmt.Errorf("could not parse issue number from URL: %s", issueURL)
	}
	issueNumberStr := parts[len(parts)-1]
	issueNumber := 0
	_, err = fmt.Sscanf(issueNumberStr, "%d", &issueNumber)
	if err != nil {
		return nil, fmt.Errorf("could not parse issue number from URL: %s", issueURL)
	}

	return &Issue{
		Number: issueNumber,
		Title:  title,
		Body:   body,
		URL:    issueURL,
		State:  "open",
	}, nil
}

// CreatePullRequest creates a pull request using gh CLI.
//
// Parameters:
//   - title: PR title
//   - body: PR description (supports markdown)
//
// Returns the PR URL on success.
func (g *GitHubCLI) CreatePullRequest(title, body string) (string, error) {
	if err := g.Ensure(); err != nil {
		return "", err
	}

	stdout, stderr, err := g.gh.Run(
		"pr", "create",
		"--title", title,
		"--body", body,
	)
	if err != nil {
		return "", ErrGHCommand{
			Command: "pr create",
			Stderr:  stderr,
			Err:     err,
		}
	}

	// gh pr create outputs the PR URL on the last line
	output := strings.TrimSpace(stdout)
	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("unexpected empty output from gh pr create")
	}

	// The URL is typically the last line
	prURL := strings.TrimSpace(lines[len(lines)-1])

	// Validate it looks like a URL
	if !strings.HasPrefix(prURL, "http") {
		return "", fmt.Errorf("unexpected pr create output (no URL found): %s", output)
	}

	return prURL, nil
}

// Helper functions

// toKebabCase converts a string to kebab-case.
func toKebabCase(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces and underscores with hyphens
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")

	// Remove any non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	// Remove consecutive hyphens
	str := result.String()
	for strings.Contains(str, "--") {
		str = strings.ReplaceAll(str, "--", "-")
	}

	// Trim hyphens from start and end
	str = strings.Trim(str, "-")

	return str
}
