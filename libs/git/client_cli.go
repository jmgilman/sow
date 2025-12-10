package git

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jmgilman/sow/libs/exec"
)

// GitHubCLI implements GitHubClient using the gh CLI tool.
//
// All operations require the gh CLI to be installed and authenticated.
// The client accepts an Executor interface for testability.
//
//nolint:revive // Name matches established API pattern from cli/internal/sow
type GitHubCLI struct {
	exec exec.Executor
}

// NewGitHubCLI creates a new GitHub CLI client with the given executor.
//
// Example:
//
//	ghExec := exec.NewLocalExecutor("gh")
//	github := git.NewGitHubCLI(ghExec)
func NewGitHubCLI(executor exec.Executor) *GitHubCLI {
	return &GitHubCLI{
		exec: executor,
	}
}

// checkInstalled verifies that the gh CLI is installed and available.
func (g *GitHubCLI) checkInstalled() error {
	if !g.exec.Exists() {
		return ErrGHNotInstalled{}
	}
	return nil
}

// checkAuthenticated verifies that the gh CLI is authenticated.
func (g *GitHubCLI) checkAuthenticated() error {
	// gh auth status exits with code 1 if not authenticated
	// but writes to stderr in both success and failure cases
	if err := g.exec.RunSilent("auth", "status"); err != nil {
		return ErrGHNotAuthenticated{}
	}

	return nil
}

// ensure checks that gh is installed and authenticated.
func (g *GitHubCLI) ensure() error {
	if err := g.checkInstalled(); err != nil {
		return err
	}
	if err := g.checkAuthenticated(); err != nil {
		return err
	}
	return nil
}

// CheckAvailability implements GitHubClient.
// For CLI client, this checks that gh is installed and authenticated.
func (g *GitHubCLI) CheckAvailability() error {
	return g.ensure()
}

// ListIssues lists issues with the given label and state.
func (g *GitHubCLI) ListIssues(label, state string) ([]Issue, error) {
	if err := g.ensure(); err != nil {
		return nil, err
	}

	stdout, stderr, err := g.exec.Run(
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
	if err := g.ensure(); err != nil {
		return nil, err
	}

	stdout, stderr, err := g.exec.Run(
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

// CreateIssue creates a GitHub issue using gh CLI.
func (g *GitHubCLI) CreateIssue(title, body string, labels []string) (*Issue, error) {
	if err := g.ensure(); err != nil {
		return nil, err
	}

	// Build command arguments
	args := []string{"issue", "create", "--title", title, "--body", body}

	// Add labels if provided
	for _, label := range labels {
		args = append(args, "--label", label)
	}

	stdout, stderr, err := g.exec.Run(args...)
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

// GetLinkedBranches returns branches linked to an issue.
func (g *GitHubCLI) GetLinkedBranches(number int) ([]LinkedBranch, error) {
	if err := g.ensure(); err != nil {
		return nil, err
	}

	stdout, stderr, err := g.exec.Run(
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
func (g *GitHubCLI) CreateLinkedBranch(issueNumber int, branchName string, checkout bool) (string, error) {
	if err := g.ensure(); err != nil {
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

	stdout, stderr, err := g.exec.Run(args...)
	if err != nil {
		return "", ErrGHCommand{
			Command: fmt.Sprintf("issue develop %d", issueNumber),
			Stderr:  stderr,
			Err:     err,
		}
	}

	// If we specified a name, use that
	if branchName != "" {
		return branchName, nil
	}

	// Parse output to get branch name
	// gh issue develop outputs: "Created branch <name> based on <base>"
	// or: "Checked out branch <name> for issue #<number>"
	lines := strings.Split(stdout, "\n")
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

// CreatePullRequest creates a pull request using gh CLI, optionally as a draft.
func (g *GitHubCLI) CreatePullRequest(title, body string, draft bool) (int, string, error) {
	if err := g.ensure(); err != nil {
		return 0, "", err
	}

	// Build command arguments
	args := []string{"pr", "create", "--title", title, "--body", body}
	if draft {
		args = append(args, "--draft")
	}

	stdout, stderr, err := g.exec.Run(args...)
	if err != nil {
		return 0, "", ErrGHCommand{
			Command: "pr create",
			Stderr:  stderr,
			Err:     err,
		}
	}

	// gh pr create outputs the PR URL on the last line
	output := strings.TrimSpace(stdout)
	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		return 0, "", fmt.Errorf("unexpected empty output from gh pr create")
	}

	// The URL is typically the last line
	prURL := strings.TrimSpace(lines[len(lines)-1])

	// Validate it looks like a URL
	if !strings.HasPrefix(prURL, "http") {
		return 0, "", fmt.Errorf("unexpected pr create output (no URL found): %s", output)
	}

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
}

// UpdatePullRequest updates an existing pull request's title and body.
func (g *GitHubCLI) UpdatePullRequest(number int, title, body string) error {
	if err := g.ensure(); err != nil {
		return err
	}

	_, stderr, err := g.exec.Run(
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

// MarkPullRequestReady converts a draft pull request to "ready for review" state.
func (g *GitHubCLI) MarkPullRequestReady(number int) error {
	if err := g.ensure(); err != nil {
		return err
	}

	_, stderr, err := g.exec.Run(
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

// Compile-time interface check.
var _ GitHubClient = (*GitHubCLI)(nil)
