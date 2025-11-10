package sow_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/jmgilman/sow/cli/internal/exec"
	"github.com/jmgilman/sow/cli/internal/sow"
)

// Example demonstrates how to test GitHub functionality using MockExecutor.
func Example() {
	// Create a mock executor that simulates gh CLI responses
	mock := &exec.MockExecutor{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			// Simulate successful authentication
			return nil
		},
		RunFunc: func(_ ...string) (string, string, error) {
			// Simulate gh issue view response
			issue := map[string]interface{}{
				"number": 123,
				"title":  "Test Issue",
				"body":   "Test body",
				"url":    "https://github.com/org/repo/issues/123",
				"state":  "open",
				"labels": []map[string]string{},
			}
			stdout, _ := json.Marshal(issue)
			return string(stdout), "", nil
		},
	}

	// Create GitHub client with mock executor
	github := sow.NewGitHubCLI(mock)

	// Now you can test GitHub operations without calling the real gh CLI
	issue, err := github.GetIssue(123)
	if err != nil {
		panic(err)
	}

	// Verify the mock worked
	if issue.Number != 123 {
		panic("unexpected issue number")
	}
	if issue.Title != "Test Issue" {
		panic("unexpected issue title")
	}

	// Output:
}

func TestGitHub_CheckInstalled_MockNotInstalled(t *testing.T) {
	// Test that CheckInstalled correctly detects when gh is not available
	mock := &exec.MockExecutor{
		ExistsFunc: func() bool { return false },
	}

	github := sow.NewGitHubCLI(mock)

	err := github.CheckInstalled()
	if err == nil {
		t.Error("expected error when gh not installed, got nil")
	}

	// Verify error type
	var notInstalled sow.ErrGHNotInstalled
	if !errors.As(err, &notInstalled) {
		t.Errorf("expected ErrGHNotInstalled, got %T", err)
	}
}

func TestGitHub_CheckAuthenticated_MockNotAuthenticated(t *testing.T) {
	// Test that CheckAuthenticated correctly detects authentication failure
	mock := &exec.MockExecutor{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			// Simulate gh auth status failing
			return &MockError{Message: "not authenticated"}
		},
	}

	github := sow.NewGitHubCLI(mock)

	err := github.CheckAuthenticated()
	if err == nil {
		t.Error("expected error when not authenticated, got nil")
	}

	// Verify error type
	var notAuthenticated sow.ErrGHNotAuthenticated
	if !errors.As(err, &notAuthenticated) {
		t.Errorf("expected ErrGHNotAuthenticated, got %T", err)
	}
}

// MockError is a simple error type for mock errors.
type MockError struct {
	Message string
}

func (e *MockError) Error() string {
	return e.Message
}

// Tests for UpdatePullRequest

func TestGitHubCLI_UpdatePullRequest_Success(t *testing.T) {
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
	if len(capturedArgs) != len(expected) {
		t.Errorf("expected %d args, got %d", len(expected), len(capturedArgs))
	}
	for i, arg := range expected {
		if i >= len(capturedArgs) || capturedArgs[i] != arg {
			t.Errorf("arg %d: expected %q, got %q", i, arg, capturedArgs[i])
		}
	}
}

func TestGitHubCLI_UpdatePullRequest_NotInstalled(t *testing.T) {
	mock := &exec.MockExecutor{
		ExistsFunc: func() bool { return false },
	}

	gh := sow.NewGitHubCLI(mock)
	err := gh.UpdatePullRequest(123, "Title", "Body")

	if err == nil {
		t.Fatal("expected error when gh not installed, got nil")
	}

	var notInstalled sow.ErrGHNotInstalled
	if !errors.As(err, &notInstalled) {
		t.Errorf("expected ErrGHNotInstalled, got %T", err)
	}
}

func TestGitHubCLI_UpdatePullRequest_CommandFails(t *testing.T) {
	mock := &exec.MockExecutor{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil // Auth check passes
		},
		RunFunc: func(_ ...string) (string, string, error) {
			return "", "PR not found", &MockError{Message: "exit code 1"}
		},
	}

	gh := sow.NewGitHubCLI(mock)
	err := gh.UpdatePullRequest(999, "Title", "Body")

	if err == nil {
		t.Fatal("expected error when command fails, got nil")
	}

	var ghErr sow.ErrGHCommand
	if !errors.As(err, &ghErr) {
		t.Errorf("expected ErrGHCommand, got %T", err)
	}
}

// Tests for MarkPullRequestReady

func TestGitHubCLI_MarkPullRequestReady_Success(t *testing.T) {
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
	err := gh.MarkPullRequestReady(42)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify correct command was called
	expected := []string{"pr", "ready", "42"}
	if len(capturedArgs) != len(expected) {
		t.Errorf("expected %d args, got %d", len(expected), len(capturedArgs))
	}
	for i, arg := range expected {
		if i >= len(capturedArgs) || capturedArgs[i] != arg {
			t.Errorf("arg %d: expected %q, got %q", i, arg, capturedArgs[i])
		}
	}
}

func TestGitHubCLI_MarkPullRequestReady_NotInstalled(t *testing.T) {
	mock := &exec.MockExecutor{
		ExistsFunc: func() bool { return false },
	}

	gh := sow.NewGitHubCLI(mock)
	err := gh.MarkPullRequestReady(42)

	if err == nil {
		t.Fatal("expected error when gh not installed, got nil")
	}

	var notInstalled sow.ErrGHNotInstalled
	if !errors.As(err, &notInstalled) {
		t.Errorf("expected ErrGHNotInstalled, got %T", err)
	}
}

func TestGitHubCLI_MarkPullRequestReady_CommandFails(t *testing.T) {
	mock := &exec.MockExecutor{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil // Auth check passes
		},
		RunFunc: func(_ ...string) (string, string, error) {
			return "", "PR is not a draft", &MockError{Message: "exit code 1"}
		},
	}

	gh := sow.NewGitHubCLI(mock)
	err := gh.MarkPullRequestReady(42)

	if err == nil {
		t.Fatal("expected error when command fails, got nil")
	}

	var ghErr sow.ErrGHCommand
	if !errors.As(err, &ghErr) {
		t.Errorf("expected ErrGHCommand, got %T", err)
	}
}

// Tests for CreatePullRequest (enhanced with draft parameter)

func TestGitHubCLI_CreatePullRequest_Draft(t *testing.T) {
	var capturedArgs []string

	mock := &exec.MockExecutor{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil // Auth check passes
		},
		RunFunc: func(args ...string) (string, string, error) {
			capturedArgs = args
			return "https://github.com/owner/repo/pull/42\n", "", nil
		},
	}

	gh := sow.NewGitHubCLI(mock)
	number, url, err := gh.CreatePullRequest("Draft Title", "Draft Body", true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify --draft flag is present
	found := false
	for _, arg := range capturedArgs {
		if arg == "--draft" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected --draft flag in arguments")
	}

	// Verify PR number and URL are correct
	if number != 42 {
		t.Errorf("expected PR number 42, got %d", number)
	}
	if url != "https://github.com/owner/repo/pull/42" {
		t.Errorf("expected URL 'https://github.com/owner/repo/pull/42', got %q", url)
	}
}

func TestGitHubCLI_CreatePullRequest_NotDraft(t *testing.T) {
	var capturedArgs []string

	mock := &exec.MockExecutor{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil // Auth check passes
		},
		RunFunc: func(args ...string) (string, string, error) {
			capturedArgs = args
			return "https://github.com/owner/repo/pull/100\n", "", nil
		},
	}

	gh := sow.NewGitHubCLI(mock)
	number, url, err := gh.CreatePullRequest("Ready Title", "Ready Body", false)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify --draft flag is NOT present
	for _, arg := range capturedArgs {
		if arg == "--draft" {
			t.Error("did not expect --draft flag in arguments when draft=false")
		}
	}

	// Verify PR number and URL are correct
	if number != 100 {
		t.Errorf("expected PR number 100, got %d", number)
	}
	if url != "https://github.com/owner/repo/pull/100" {
		t.Errorf("expected URL 'https://github.com/owner/repo/pull/100', got %q", url)
	}
}

func TestGitHubCLI_CreatePullRequest_ParseError(t *testing.T) {
	mock := &exec.MockExecutor{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil // Auth check passes
		},
		RunFunc: func(_ ...string) (string, string, error) {
			// Return invalid URL that can't be parsed
			return "invalid-url\n", "", nil
		},
	}

	gh := sow.NewGitHubCLI(mock)
	_, _, err := gh.CreatePullRequest("Title", "Body", false)

	if err == nil {
		t.Fatal("expected error when URL cannot be parsed, got nil")
	}
}

func TestGitHubCLI_CreatePullRequest_NotInstalled(t *testing.T) {
	mock := &exec.MockExecutor{
		ExistsFunc: func() bool { return false },
	}

	gh := sow.NewGitHubCLI(mock)
	_, _, err := gh.CreatePullRequest("Title", "Body", false)

	if err == nil {
		t.Fatal("expected error when gh not installed, got nil")
	}

	var notInstalled sow.ErrGHNotInstalled
	if !errors.As(err, &notInstalled) {
		t.Errorf("expected ErrGHNotInstalled, got %T", err)
	}
}

func TestGitHubCLI_CreatePullRequest_CommandFails(t *testing.T) {
	mock := &exec.MockExecutor{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil // Auth check passes
		},
		RunFunc: func(_ ...string) (string, string, error) {
			return "", "no commits to create PR", &MockError{Message: "exit code 1"}
		},
	}

	gh := sow.NewGitHubCLI(mock)
	_, _, err := gh.CreatePullRequest("Title", "Body", false)

	if err == nil {
		t.Fatal("expected error when command fails, got nil")
	}

	var ghErr sow.ErrGHCommand
	if !errors.As(err, &ghErr) {
		t.Errorf("expected ErrGHCommand, got %T", err)
	}
}
