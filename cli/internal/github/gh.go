// Package github provides integration with the GitHub CLI (gh).
package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

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

// CheckInstalled verifies that the gh CLI is installed and available.
func CheckInstalled() error {
	_, err := exec.LookPath("gh")
	if err != nil {
		return ErrGHNotInstalled{}
	}
	return nil
}

// CheckAuthenticated verifies that the gh CLI is authenticated.
func CheckAuthenticated() error {
	cmd := exec.Command("gh", "auth", "status")

	// gh auth status exits with code 1 if not authenticated
	// but writes to stderr in both success and failure cases
	if err := cmd.Run(); err != nil {
		return ErrGHNotAuthenticated{}
	}

	return nil
}

// Ensure checks that gh is installed and authenticated.
func Ensure() error {
	if err := CheckInstalled(); err != nil {
		return err
	}
	if err := CheckAuthenticated(); err != nil {
		return err
	}
	return nil
}

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

// ListIssues lists issues with the given label and state.
func ListIssues(label, state string) ([]Issue, error) {
	if err := Ensure(); err != nil {
		return nil, err
	}

	args := []string{
		"issue", "list",
		"--label", label,
		"--state", state,
		"--json", "number,title,url,state,labels",
		"--limit", "1000",
	}

	cmd := exec.Command("gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, ErrGHCommand{
			Command: strings.Join(args, " "),
			Stderr:  stderr.String(),
			Err:     err,
		}
	}

	var issues []Issue
	if err := json.Unmarshal(stdout.Bytes(), &issues); err != nil {
		return nil, fmt.Errorf("failed to parse issue list: %w", err)
	}

	return issues, nil
}

// GetIssue retrieves a single issue by number.
func GetIssue(number int) (*Issue, error) {
	if err := Ensure(); err != nil {
		return nil, err
	}

	args := []string{
		"issue", "view", fmt.Sprintf("%d", number),
		"--json", "number,title,body,url,state,labels",
	}

	cmd := exec.Command("gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, ErrGHCommand{
			Command: strings.Join(args, " "),
			Stderr:  stderr.String(),
			Err:     err,
		}
	}

	var issue Issue
	if err := json.Unmarshal(stdout.Bytes(), &issue); err != nil {
		return nil, fmt.Errorf("failed to parse issue: %w", err)
	}

	return &issue, nil
}

// GetLinkedBranches returns branches linked to an issue.
func GetLinkedBranches(number int) ([]LinkedBranch, error) {
	if err := Ensure(); err != nil {
		return nil, err
	}

	args := []string{
		"issue", "develop", "--list", fmt.Sprintf("%d", number),
	}

	cmd := exec.Command("gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Check if error is because no branches are linked
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "no linked branches") ||
			strings.Contains(stderrStr, "No linked branches") {
			return []LinkedBranch{}, nil
		}

		return nil, ErrGHCommand{
			Command: strings.Join(args, " "),
			Stderr:  stderrStr,
			Err:     err,
		}
	}

	// Parse output - gh issue develop --list returns tab-separated values
	// Format: BRANCH    URL
	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")

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

// HasLabel checks if an issue has a specific label.
func (i *Issue) HasLabel(label string) bool {
	for _, l := range i.Labels {
		if l.Name == label {
			return true
		}
	}
	return false
}

// CreateLinkedBranch creates a branch linked to an issue using gh issue develop.
// Returns the branch name that was created.
func CreateLinkedBranch(issueNumber int, branchName string, checkout bool) (string, error) {
	if err := Ensure(); err != nil {
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

	cmd := exec.Command("gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", ErrGHCommand{
			Command: strings.Join(args, " "),
			Stderr:  stderr.String(),
			Err:     err,
		}
	}

	// Parse output to get branch name
	// gh issue develop outputs: "Created branch <name> based on <base>"
	// or: "Checked out branch <name> for issue #<number>"
	output := stdout.String()

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
	issue, err := GetIssue(issueNumber)
	if err != nil {
		return "", fmt.Errorf("failed to get issue for branch name: %w", err)
	}

	// Convert title to kebab-case
	branchName = fmt.Sprintf("%d-%s", issueNumber, toKebabCase(issue.Title))
	return branchName, nil
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

// CreatePullRequest creates a pull request using gh CLI.
// Returns the PR URL on success.
func CreatePullRequest(title, body string) (string, error) {
	if err := Ensure(); err != nil {
		return "", err
	}

	args := []string{
		"pr", "create",
		"--title", title,
		"--body", body,
	}

	cmd := exec.Command("gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", ErrGHCommand{
			Command: strings.Join(args, " "),
			Stderr:  stderr.String(),
			Err:     err,
		}
	}

	// gh pr create outputs the PR URL on the last line
	output := strings.TrimSpace(stdout.String())
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
