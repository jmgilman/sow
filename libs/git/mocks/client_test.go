package mocks_test

import (
	"testing"

	"github.com/jmgilman/sow/libs/git"
	"github.com/jmgilman/sow/libs/git/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// TestGitHubClientMock tests that all mock methods work correctly.
func TestGitHubClientMock(t *testing.T) {
	t.Run("CheckAvailability returns configured value", func(t *testing.T) {
		mock := &mocks.GitHubClientMock{
			CheckAvailabilityFunc: func() error { return nil },
		}

		err := mock.CheckAvailability()

		require.NoError(t, err)
	})

	t.Run("ListIssues returns configured issues", func(t *testing.T) {
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

		require.NoError(t, err)
		require.Len(t, issues, 2)
		assert.Equal(t, "Test Issue", issues[0].Title)
	})

	t.Run("GetIssue returns configured issue", func(t *testing.T) {
		expectedIssue := &git.Issue{Number: 42, Title: "Found Issue"}
		mock := &mocks.GitHubClientMock{
			GetIssueFunc: func(_ int) (*git.Issue, error) {
				return expectedIssue, nil
			},
		}

		issue, err := mock.GetIssue(42)

		require.NoError(t, err)
		assert.Equal(t, 42, issue.Number)
	})

	t.Run("CreateIssue returns configured issue", func(t *testing.T) {
		mock := &mocks.GitHubClientMock{
			CreateIssueFunc: func(title, _ string, _ []string) (*git.Issue, error) {
				return &git.Issue{Number: 100, Title: title}, nil
			},
		}

		issue, err := mock.CreateIssue("New Issue", "Body text", []string{"enhancement"})

		require.NoError(t, err)
		assert.Equal(t, 100, issue.Number)
	})

	t.Run("GetLinkedBranches returns configured branches", func(t *testing.T) {
		mock := &mocks.GitHubClientMock{
			GetLinkedBranchesFunc: func(_ int) ([]git.LinkedBranch, error) {
				return []git.LinkedBranch{{Name: "feat/123"}}, nil
			},
		}

		branches, err := mock.GetLinkedBranches(123)

		require.NoError(t, err)
		require.Len(t, branches, 1)
	})

	t.Run("CreateLinkedBranch returns configured branch name", func(t *testing.T) {
		mock := &mocks.GitHubClientMock{
			CreateLinkedBranchFunc: func(_ int, _ string, _ bool) (string, error) {
				return "feat/issue-123", nil
			},
		}

		name, err := mock.CreateLinkedBranch(123, "", false)

		require.NoError(t, err)
		assert.Equal(t, "feat/issue-123", name)
	})

	t.Run("CreatePullRequest returns configured PR details", func(t *testing.T) {
		mock := &mocks.GitHubClientMock{
			CreatePullRequestFunc: func(_, _ string, _ bool) (int, string, error) {
				return 456, "https://github.com/owner/repo/pull/456", nil
			},
		}

		num, url, err := mock.CreatePullRequest("My PR", "Description", true)

		require.NoError(t, err)
		assert.Equal(t, 456, num)
		assert.Equal(t, "https://github.com/owner/repo/pull/456", url)
	})

	t.Run("UpdatePullRequest returns configured result", func(t *testing.T) {
		mock := &mocks.GitHubClientMock{
			UpdatePullRequestFunc: func(_ int, _, _ string) error {
				return nil
			},
		}

		err := mock.UpdatePullRequest(456, "Updated Title", "Updated Body")

		require.NoError(t, err)
	})

	t.Run("MarkPullRequestReady returns configured result", func(t *testing.T) {
		mock := &mocks.GitHubClientMock{
			MarkPullRequestReadyFunc: func(_ int) error {
				return nil
			},
		}

		err := mock.MarkPullRequestReady(456)

		require.NoError(t, err)
	})
}
