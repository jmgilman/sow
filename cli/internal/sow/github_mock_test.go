package sow_test

import (
	"errors"
	"testing"

	"github.com/jmgilman/sow/cli/internal/sow"
)

func TestMockGitHub_ImplementsInterface(_ *testing.T) {
	var _ sow.GitHubClient = (*sow.MockGitHub)(nil)
}

func TestMockGitHub_CheckAvailability_WithCustomFunc(t *testing.T) {
	expectedErr := errors.New("mock error")
	mock := &sow.MockGitHub{
		CheckAvailabilityFunc: func() error {
			return expectedErr
		},
	}

	err := mock.CheckAvailability()

	if err == nil || err.Error() != expectedErr.Error() {
		t.Errorf("expected mock error, got %v", err)
	}
}

func TestMockGitHub_CheckAvailability_WithNilFunc(t *testing.T) {
	mock := &sow.MockGitHub{}

	err := mock.CheckAvailability()

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockGitHub_ListIssues_WithCustomFunc(t *testing.T) {
	expectedIssues := []sow.Issue{
		{Number: 1, Title: "Issue 1"},
		{Number: 2, Title: "Issue 2"},
	}
	mock := &sow.MockGitHub{
		ListIssuesFunc: func(_, _ string) ([]sow.Issue, error) {
			return expectedIssues, nil
		},
	}

	issues, err := mock.ListIssues("sow", "open")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 2 {
		t.Errorf("expected 2 issues, got %d", len(issues))
	}
	if issues[0].Number != 1 {
		t.Errorf("expected first issue number 1, got %d", issues[0].Number)
	}
}

func TestMockGitHub_ListIssues_WithNilFunc(t *testing.T) {
	mock := &sow.MockGitHub{}

	issues, err := mock.ListIssues("sow", "open")

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if issues == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(issues) != 0 {
		t.Errorf("expected empty slice, got %d issues", len(issues))
	}
}

func TestMockGitHub_GetIssue_WithCustomFunc(t *testing.T) {
	mock := &sow.MockGitHub{
		GetIssueFunc: func(number int) (*sow.Issue, error) {
			return &sow.Issue{
				Number: number,
				Title:  "Custom Issue",
				State:  "open",
			}, nil
		},
	}

	issue, err := mock.GetIssue(123)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issue == nil {
		t.Fatal("expected issue, got nil")
	}
	if issue.Number != 123 {
		t.Errorf("expected number 123, got %d", issue.Number)
	}
	if issue.Title != "Custom Issue" {
		t.Errorf("expected title 'Custom Issue', got %s", issue.Title)
	}
}

func TestMockGitHub_GetIssue_WithNilFunc(t *testing.T) {
	mock := &sow.MockGitHub{}

	issue, err := mock.GetIssue(123)

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if issue != nil {
		t.Errorf("expected nil issue, got %+v", issue)
	}
}

func TestMockGitHub_CreateIssue_WithCustomFunc(t *testing.T) {
	mock := &sow.MockGitHub{
		CreateIssueFunc: func(title, body string, _ []string) (*sow.Issue, error) {
			return &sow.Issue{
				Number: 42,
				Title:  title,
				Body:   body,
				State:  "open",
			}, nil
		},
	}

	issue, err := mock.CreateIssue("Test Title", "Test Body", []string{"bug"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issue == nil {
		t.Fatal("expected issue, got nil")
	}
	if issue.Number != 42 {
		t.Errorf("expected number 42, got %d", issue.Number)
	}
	if issue.Title != "Test Title" {
		t.Errorf("expected title 'Test Title', got %s", issue.Title)
	}
}

func TestMockGitHub_CreateIssue_WithNilFunc(t *testing.T) {
	mock := &sow.MockGitHub{}

	issue, err := mock.CreateIssue("Title", "Body", nil)

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if issue != nil {
		t.Errorf("expected nil issue, got %+v", issue)
	}
}

func TestMockGitHub_GetLinkedBranches_WithCustomFunc(t *testing.T) {
	expectedBranches := []sow.LinkedBranch{
		{Name: "branch-1", URL: "https://github.com/owner/repo/tree/branch-1"},
		{Name: "branch-2", URL: "https://github.com/owner/repo/tree/branch-2"},
	}
	mock := &sow.MockGitHub{
		GetLinkedBranchesFunc: func(_ int) ([]sow.LinkedBranch, error) {
			return expectedBranches, nil
		},
	}

	branches, err := mock.GetLinkedBranches(123)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(branches) != 2 {
		t.Errorf("expected 2 branches, got %d", len(branches))
	}
	if branches[0].Name != "branch-1" {
		t.Errorf("expected first branch name 'branch-1', got %s", branches[0].Name)
	}
}

func TestMockGitHub_GetLinkedBranches_WithNilFunc(t *testing.T) {
	mock := &sow.MockGitHub{}

	branches, err := mock.GetLinkedBranches(123)

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if branches == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(branches) != 0 {
		t.Errorf("expected empty slice, got %d branches", len(branches))
	}
}

func TestMockGitHub_CreateLinkedBranch_WithCustomFunc(t *testing.T) {
	mock := &sow.MockGitHub{
		CreateLinkedBranchFunc: func(_ int, branchName string, _ bool) (string, error) {
			if branchName == "" {
				return "123-auto-generated-branch", nil
			}
			return branchName, nil
		},
	}

	branchName, err := mock.CreateLinkedBranch(123, "custom-branch", true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if branchName != "custom-branch" {
		t.Errorf("expected branch name 'custom-branch', got %s", branchName)
	}

	// Test with empty branch name (auto-generated)
	branchName, err = mock.CreateLinkedBranch(123, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if branchName != "123-auto-generated-branch" {
		t.Errorf("expected auto-generated branch name, got %s", branchName)
	}
}

func TestMockGitHub_CreateLinkedBranch_WithNilFunc(t *testing.T) {
	mock := &sow.MockGitHub{}

	branchName, err := mock.CreateLinkedBranch(123, "branch", true)

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if branchName != "" {
		t.Errorf("expected empty string, got %s", branchName)
	}
}

func TestMockGitHub_CreatePullRequest_WithCustomFunc(t *testing.T) {
	mock := &sow.MockGitHub{
		CreatePullRequestFunc: func(_, _ string, _ bool) (int, string, error) {
			return 42, "https://github.com/owner/repo/pull/42", nil
		},
	}

	number, url, err := mock.CreatePullRequest("Title", "Body", true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if number != 42 {
		t.Errorf("expected number 42, got %d", number)
	}
	if url != "https://github.com/owner/repo/pull/42" {
		t.Errorf("unexpected url: %s", url)
	}
}

func TestMockGitHub_CreatePullRequest_WithNilFunc(t *testing.T) {
	mock := &sow.MockGitHub{}

	number, url, err := mock.CreatePullRequest("Title", "Body", false)

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if number != 0 {
		t.Errorf("expected number 0, got %d", number)
	}
	if url != "" {
		t.Errorf("expected empty url, got %s", url)
	}
}

func TestMockGitHub_UpdatePullRequest_WithCustomFunc(t *testing.T) {
	updateCalled := false
	mock := &sow.MockGitHub{
		UpdatePullRequestFunc: func(number int, _, _ string) error {
			updateCalled = true
			if number != 42 {
				t.Errorf("expected number 42, got %d", number)
			}
			return nil
		},
	}

	err := mock.UpdatePullRequest(42, "New Title", "New Body")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updateCalled {
		t.Error("expected UpdatePullRequestFunc to be called")
	}
}

func TestMockGitHub_UpdatePullRequest_WithNilFunc(t *testing.T) {
	mock := &sow.MockGitHub{}

	err := mock.UpdatePullRequest(42, "Title", "Body")

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestMockGitHub_MarkPullRequestReady_WithCustomFunc(t *testing.T) {
	markReadyCalled := false
	mock := &sow.MockGitHub{
		MarkPullRequestReadyFunc: func(number int) error {
			markReadyCalled = true
			if number != 42 {
				t.Errorf("expected number 42, got %d", number)
			}
			return nil
		},
	}

	err := mock.MarkPullRequestReady(42)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !markReadyCalled {
		t.Error("expected MarkPullRequestReadyFunc to be called")
	}
}

func TestMockGitHub_MarkPullRequestReady_WithNilFunc(t *testing.T) {
	mock := &sow.MockGitHub{}

	err := mock.MarkPullRequestReady(42)

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

// TestMockGitHub_ErrorPropagation tests that errors from custom funcs are properly propagated.
func TestMockGitHub_ErrorPropagation(t *testing.T) {
	expectedErr := errors.New("custom error")

	tests := []struct {
		name string
		mock *sow.MockGitHub
		test func(*sow.MockGitHub) error
	}{
		{
			name: "CheckAvailability",
			mock: &sow.MockGitHub{
				CheckAvailabilityFunc: func() error { return expectedErr },
			},
			test: func(m *sow.MockGitHub) error { return m.CheckAvailability() },
		},
		{
			name: "ListIssues",
			mock: &sow.MockGitHub{
				ListIssuesFunc: func(_, _ string) ([]sow.Issue, error) { return nil, expectedErr },
			},
			test: func(m *sow.MockGitHub) error { _, err := m.ListIssues("", ""); return err },
		},
		{
			name: "GetIssue",
			mock: &sow.MockGitHub{
				GetIssueFunc: func(_ int) (*sow.Issue, error) { return nil, expectedErr },
			},
			test: func(m *sow.MockGitHub) error { _, err := m.GetIssue(1); return err },
		},
		{
			name: "CreateIssue",
			mock: &sow.MockGitHub{
				CreateIssueFunc: func(_, _ string, _ []string) (*sow.Issue, error) {
					return nil, expectedErr
				},
			},
			test: func(m *sow.MockGitHub) error { _, err := m.CreateIssue("", "", nil); return err },
		},
		{
			name: "GetLinkedBranches",
			mock: &sow.MockGitHub{
				GetLinkedBranchesFunc: func(_ int) ([]sow.LinkedBranch, error) { return nil, expectedErr },
			},
			test: func(m *sow.MockGitHub) error { _, err := m.GetLinkedBranches(1); return err },
		},
		{
			name: "CreateLinkedBranch",
			mock: &sow.MockGitHub{
				CreateLinkedBranchFunc: func(_ int, _ string, _ bool) (string, error) {
					return "", expectedErr
				},
			},
			test: func(m *sow.MockGitHub) error { _, err := m.CreateLinkedBranch(1, "", false); return err },
		},
		{
			name: "CreatePullRequest",
			mock: &sow.MockGitHub{
				CreatePullRequestFunc: func(_, _ string, _ bool) (int, string, error) {
					return 0, "", expectedErr
				},
			},
			test: func(m *sow.MockGitHub) error { _, _, err := m.CreatePullRequest("", "", false); return err },
		},
		{
			name: "UpdatePullRequest",
			mock: &sow.MockGitHub{
				UpdatePullRequestFunc: func(_ int, _, _ string) error { return expectedErr },
			},
			test: func(m *sow.MockGitHub) error { return m.UpdatePullRequest(1, "", "") },
		},
		{
			name: "MarkPullRequestReady",
			mock: &sow.MockGitHub{
				MarkPullRequestReadyFunc: func(_ int) error { return expectedErr },
			},
			test: func(m *sow.MockGitHub) error { return m.MarkPullRequestReady(1) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.test(tt.mock)
			if err == nil || err.Error() != expectedErr.Error() {
				t.Errorf("expected error %v, got %v", expectedErr, err)
			}
		})
	}
}
