package sow_test

import (
	"encoding/json"
	"testing"

	"github.com/jmgilman/sow/cli/internal/exec"
	"github.com/jmgilman/sow/cli/internal/sow"
)

// Example demonstrates how to test GitHub functionality using MockExecutor.
func Example() {
	// Create a mock executor that simulates gh CLI responses
	mock := &exec.MockExecutor{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(args ...string) error {
			// Simulate successful authentication
			return nil
		},
		RunFunc: func(args ...string) (string, string, error) {
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
	github := sow.NewGitHub(mock)

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

	github := sow.NewGitHub(mock)

	err := github.CheckInstalled()
	if err == nil {
		t.Error("expected error when gh not installed, got nil")
	}

	// Verify error type
	if _, ok := err.(sow.ErrGHNotInstalled); !ok {
		t.Errorf("expected ErrGHNotInstalled, got %T", err)
	}
}

func TestGitHub_CheckAuthenticated_MockNotAuthenticated(t *testing.T) {
	// Test that CheckAuthenticated correctly detects authentication failure
	mock := &exec.MockExecutor{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(args ...string) error {
			// Simulate gh auth status failing
			return &MockError{Message: "not authenticated"}
		},
	}

	github := sow.NewGitHub(mock)

	err := github.CheckAuthenticated()
	if err == nil {
		t.Error("expected error when not authenticated, got nil")
	}

	// Verify error type
	if _, ok := err.(sow.ErrGHNotAuthenticated); !ok {
		t.Errorf("expected ErrGHNotAuthenticated, got %T", err)
	}
}

// MockError is a simple error type for mock errors
type MockError struct {
	Message string
}

func (e *MockError) Error() string {
	return e.Message
}
