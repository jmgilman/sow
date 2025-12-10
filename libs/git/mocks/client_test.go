package mocks_test

import (
	"testing"

	"github.com/jmgilman/sow/libs/git"
	"github.com/jmgilman/sow/libs/git/mocks"
)

// TestGitHubClientMock_Implements_Interface verifies the generated mock
// correctly implements the GitHubClient interface.
func TestGitHubClientMock_Implements_Interface(_ *testing.T) {
	// Create a mock with minimal function implementations
	mock := &mocks.GitHubClientMock{
		CheckAvailabilityFunc: func() error { return nil },
	}

	// Compile-time check: mock must implement GitHubClient
	var _ git.GitHubClient = mock
}

func TestGitHubClientMock_CheckAvailability(t *testing.T) {
	mock := &mocks.GitHubClientMock{
		CheckAvailabilityFunc: func() error { return nil },
	}

	err := mock.CheckAvailability()
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestGitHubClientMock_ListIssues(t *testing.T) {
	expectedIssues := []git.Issue{
		{Number: 1, Title: "Test Issue"},
		{Number: 2, Title: "Another Issue"},
	}

	mock := &mocks.GitHubClientMock{
		ListIssuesFunc: func(_, _ string) ([]git.Issue, error) {
			return expectedIssues, nil
		},
	}

	issues, err := mock.ListIssues("bug", "open")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if len(issues) != 2 {
		t.Errorf("expected 2 issues, got %d", len(issues))
	}
	if issues[0].Title != "Test Issue" {
		t.Errorf("expected first issue title 'Test Issue', got '%s'", issues[0].Title)
	}
}

func TestGitHubClientMock_GetIssue(t *testing.T) {
	expectedIssue := &git.Issue{Number: 42, Title: "Found Issue"}

	mock := &mocks.GitHubClientMock{
		GetIssueFunc: func(_ int) (*git.Issue, error) {
			return expectedIssue, nil
		},
	}

	issue, err := mock.GetIssue(42)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if issue.Number != 42 {
		t.Errorf("expected issue number 42, got %d", issue.Number)
	}
}

func TestGitHubClientMock_CreateIssue(t *testing.T) {
	mock := &mocks.GitHubClientMock{
		CreateIssueFunc: func(title, _ string, _ []string) (*git.Issue, error) {
			return &git.Issue{Number: 100, Title: title}, nil
		},
	}

	issue, err := mock.CreateIssue("New Issue", "Body text", []string{"enhancement"})
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if issue.Number != 100 {
		t.Errorf("expected issue number 100, got %d", issue.Number)
	}
}

func TestGitHubClientMock_GetLinkedBranches(t *testing.T) {
	mock := &mocks.GitHubClientMock{
		GetLinkedBranchesFunc: func(_ int) ([]git.LinkedBranch, error) {
			return []git.LinkedBranch{{Name: "feat/123"}}, nil
		},
	}

	branches, err := mock.GetLinkedBranches(123)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if len(branches) != 1 {
		t.Errorf("expected 1 branch, got %d", len(branches))
	}
}

func TestGitHubClientMock_CreateLinkedBranch(t *testing.T) {
	mock := &mocks.GitHubClientMock{
		CreateLinkedBranchFunc: func(_ int, _ string, _ bool) (string, error) {
			return "feat/issue-123", nil
		},
	}

	name, err := mock.CreateLinkedBranch(123, "", false)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if name != "feat/issue-123" {
		t.Errorf("expected branch name 'feat/issue-123', got '%s'", name)
	}
}

func TestGitHubClientMock_CreatePullRequest(t *testing.T) {
	mock := &mocks.GitHubClientMock{
		CreatePullRequestFunc: func(_, _ string, _ bool) (int, string, error) {
			return 456, "https://github.com/owner/repo/pull/456", nil
		},
	}

	num, url, err := mock.CreatePullRequest("My PR", "Description", true)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if num != 456 {
		t.Errorf("expected PR number 456, got %d", num)
	}
	if url != "https://github.com/owner/repo/pull/456" {
		t.Errorf("expected PR URL, got '%s'", url)
	}
}

func TestGitHubClientMock_UpdatePullRequest(t *testing.T) {
	mock := &mocks.GitHubClientMock{
		UpdatePullRequestFunc: func(_ int, _, _ string) error {
			return nil
		},
	}

	err := mock.UpdatePullRequest(456, "Updated Title", "Updated Body")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestGitHubClientMock_MarkPullRequestReady(t *testing.T) {
	mock := &mocks.GitHubClientMock{
		MarkPullRequestReadyFunc: func(_ int) error {
			return nil
		},
	}

	err := mock.MarkPullRequestReady(456)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}
