package git_test

import (
	"encoding/json"
	"testing"

	"github.com/jmgilman/sow/libs/exec/mocks"
	"github.com/jmgilman/sow/libs/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockError is a simple error type for mock errors.
type mockError struct {
	message string
}

func (e *mockError) Error() string {
	return e.message
}

// =============================================================================
// CheckAvailability Tests
// =============================================================================

func TestGitHubCLI_CheckAvailability_ReturnsNilWhenInstalled(t *testing.T) {
	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil // Auth check passes
		},
	}

	client := git.NewGitHubCLI(mock)
	err := client.CheckAvailability()

	require.NoError(t, err)
}

func TestGitHubCLI_CheckAvailability_ReturnsErrGHNotInstalled(t *testing.T) {
	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return false },
	}

	client := git.NewGitHubCLI(mock)
	err := client.CheckAvailability()

	require.Error(t, err)
	var notInstalled git.ErrGHNotInstalled
	assert.ErrorAs(t, err, &notInstalled)
}

func TestGitHubCLI_CheckAvailability_ReturnsErrGHNotAuthenticated(t *testing.T) {
	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return &mockError{message: "not authenticated"}
		},
	}

	client := git.NewGitHubCLI(mock)
	err := client.CheckAvailability()

	require.Error(t, err)
	var notAuthenticated git.ErrGHNotAuthenticated
	assert.ErrorAs(t, err, &notAuthenticated)
}

// =============================================================================
// ListIssues Tests
// =============================================================================

func TestGitHubCLI_ListIssues_ParsesJSONCorrectly(t *testing.T) {
	issues := []git.Issue{
		{Number: 1, Title: "First issue", State: "open", URL: "https://github.com/owner/repo/issues/1"},
		{Number: 2, Title: "Second issue", State: "open", URL: "https://github.com/owner/repo/issues/2"},
	}
	jsonData, _ := json.Marshal(issues)

	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(_ ...string) (string, string, error) {
			return string(jsonData), "", nil
		},
	}

	client := git.NewGitHubCLI(mock)
	result, err := client.ListIssues("sow", "open")

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, 1, result[0].Number)
	assert.Equal(t, "First issue", result[0].Title)
	assert.Equal(t, 2, result[1].Number)
	assert.Equal(t, "Second issue", result[1].Title)
}

func TestGitHubCLI_ListIssues_HandlesEmptyResult(t *testing.T) {
	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(_ ...string) (string, string, error) {
			return "[]", "", nil
		},
	}

	client := git.NewGitHubCLI(mock)
	result, err := client.ListIssues("nonexistent", "open")

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestGitHubCLI_ListIssues_PassesCorrectArguments(t *testing.T) {
	var capturedArgs []string

	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(args ...string) (string, string, error) {
			capturedArgs = args
			return "[]", "", nil
		},
	}

	client := git.NewGitHubCLI(mock)
	_, _ = client.ListIssues("bug", "closed")

	expected := []string{
		"issue", "list",
		"--label", "bug",
		"--state", "closed",
		"--json", "number,title,url,state,labels",
		"--limit", "1000",
	}
	assert.Equal(t, expected, capturedArgs)
}

func TestGitHubCLI_ListIssues_ReturnsErrorWhenNotInstalled(t *testing.T) {
	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return false },
	}

	client := git.NewGitHubCLI(mock)
	_, err := client.ListIssues("sow", "open")

	require.Error(t, err)
	var notInstalled git.ErrGHNotInstalled
	assert.ErrorAs(t, err, &notInstalled)
}

func TestGitHubCLI_ListIssues_ReturnsErrGHCommandOnFailure(t *testing.T) {
	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(_ ...string) (string, string, error) {
			return "", "some error", &mockError{message: "command failed"}
		},
	}

	client := git.NewGitHubCLI(mock)
	_, err := client.ListIssues("sow", "open")

	require.Error(t, err)
	var ghErr git.ErrGHCommand
	assert.ErrorAs(t, err, &ghErr)
}

// =============================================================================
// GetIssue Tests
// =============================================================================

func TestGitHubCLI_GetIssue_ParsesSingleIssueJSON(t *testing.T) {
	issue := git.Issue{
		Number: 123,
		Title:  "Test issue",
		Body:   "Test body content",
		State:  "open",
		URL:    "https://github.com/owner/repo/issues/123",
		Labels: []git.Label{{Name: "bug"}, {Name: "sow"}},
	}
	jsonData, _ := json.Marshal(issue)

	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(_ ...string) (string, string, error) {
			return string(jsonData), "", nil
		},
	}

	client := git.NewGitHubCLI(mock)
	result, err := client.GetIssue(123)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 123, result.Number)
	assert.Equal(t, "Test issue", result.Title)
	assert.Equal(t, "Test body content", result.Body)
	assert.Equal(t, "open", result.State)
	require.Len(t, result.Labels, 2)
	assert.Equal(t, "bug", result.Labels[0].Name)
	assert.Equal(t, "sow", result.Labels[1].Name)
}

func TestGitHubCLI_GetIssue_ReturnsErrorForNonExistent(t *testing.T) {
	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(_ ...string) (string, string, error) {
			return "", "issue not found", &mockError{message: "not found"}
		},
	}

	client := git.NewGitHubCLI(mock)
	result, err := client.GetIssue(99999)

	require.Error(t, err)
	assert.Nil(t, result)
	var ghErr git.ErrGHCommand
	assert.ErrorAs(t, err, &ghErr)
}

func TestGitHubCLI_GetIssue_PassesCorrectArguments(t *testing.T) {
	var capturedArgs []string

	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(args ...string) (string, string, error) {
			capturedArgs = args
			return `{"number":42,"title":"Test","body":"","state":"open","url":"","labels":[]}`, "", nil
		},
	}

	client := git.NewGitHubCLI(mock)
	_, _ = client.GetIssue(42)

	expected := []string{
		"issue", "view", "42",
		"--json", "number,title,body,url,state,labels",
	}
	assert.Equal(t, expected, capturedArgs)
}

// =============================================================================
// CreateIssue Tests
// =============================================================================

func TestGitHubCLI_CreateIssue_ParsesURLToExtractIssueNumber(t *testing.T) {
	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(_ ...string) (string, string, error) {
			return "https://github.com/owner/repo/issues/456\n", "", nil
		},
	}

	client := git.NewGitHubCLI(mock)
	result, err := client.CreateIssue("New Issue", "Issue body", []string{"bug"})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 456, result.Number)
	assert.Equal(t, "New Issue", result.Title)
	assert.Equal(t, "Issue body", result.Body)
	assert.Equal(t, "open", result.State)
	assert.Equal(t, "https://github.com/owner/repo/issues/456", result.URL)
}

func TestGitHubCLI_CreateIssue_HandlesLabelsCorrectly(t *testing.T) {
	var capturedArgs []string

	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(args ...string) (string, string, error) {
			capturedArgs = args
			return "https://github.com/owner/repo/issues/1\n", "", nil
		},
	}

	client := git.NewGitHubCLI(mock)
	_, _ = client.CreateIssue("Title", "Body", []string{"bug", "urgent", "sow"})

	expected := []string{
		"issue", "create",
		"--title", "Title",
		"--body", "Body",
		"--label", "bug",
		"--label", "urgent",
		"--label", "sow",
	}
	assert.Equal(t, expected, capturedArgs)
}

func TestGitHubCLI_CreateIssue_HandlesNoLabels(t *testing.T) {
	var capturedArgs []string

	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(args ...string) (string, string, error) {
			capturedArgs = args
			return "https://github.com/owner/repo/issues/1\n", "", nil
		},
	}

	client := git.NewGitHubCLI(mock)
	_, _ = client.CreateIssue("Title", "Body", nil)

	expected := []string{
		"issue", "create",
		"--title", "Title",
		"--body", "Body",
	}
	assert.Equal(t, expected, capturedArgs)
}

func TestGitHubCLI_CreateIssue_ReturnsErrorOnFailure(t *testing.T) {
	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(_ ...string) (string, string, error) {
			return "", "failed to create issue", &mockError{message: "create failed"}
		},
	}

	client := git.NewGitHubCLI(mock)
	result, err := client.CreateIssue("Title", "Body", nil)

	require.Error(t, err)
	assert.Nil(t, result)
	var ghErr git.ErrGHCommand
	assert.ErrorAs(t, err, &ghErr)
}

// =============================================================================
// GetLinkedBranches Tests
// =============================================================================

func TestGitHubCLI_GetLinkedBranches_ParsesTabSeparatedOutput(t *testing.T) {
	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(_ ...string) (string, string, error) {
			// gh issue develop --list returns tab-separated format with header
			output := "BRANCH\tURL\n" +
				"feat/add-feature\thttps://github.com/owner/repo/tree/feat/add-feature\n" +
				"fix/bug-123\thttps://github.com/owner/repo/tree/fix/bug-123\n"
			return output, "", nil
		},
	}

	client := git.NewGitHubCLI(mock)
	result, err := client.GetLinkedBranches(123)

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "feat/add-feature", result[0].Name)
	assert.Equal(t, "https://github.com/owner/repo/tree/feat/add-feature", result[0].URL)
	assert.Equal(t, "fix/bug-123", result[1].Name)
}

func TestGitHubCLI_GetLinkedBranches_ReturnsEmptySliceWhenNoBranches(t *testing.T) {
	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(_ ...string) (string, string, error) {
			// gh issue develop --list returns error when no branches linked
			return "", "no linked branches found for this issue", &mockError{message: "exit 1"}
		},
	}

	client := git.NewGitHubCLI(mock)
	result, err := client.GetLinkedBranches(123)

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestGitHubCLI_GetLinkedBranches_HandlesNoLinkedBranchesErrorGracefully(t *testing.T) {
	testCases := []struct {
		name   string
		stderr string
	}{
		{"lowercase", "no linked branches found for this issue"},
		{"capitalized", "No linked branches found for this issue"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mocks.ExecutorMock{
				ExistsFunc: func() bool { return true },
				RunSilentFunc: func(_ ...string) error {
					return nil
				},
				RunFunc: func(_ ...string) (string, string, error) {
					return "", tc.stderr, &mockError{message: "exit 1"}
				},
			}

			client := git.NewGitHubCLI(mock)
			result, err := client.GetLinkedBranches(456)

			require.NoError(t, err)
			assert.Empty(t, result)
		})
	}
}

func TestGitHubCLI_GetLinkedBranches_PassesCorrectArguments(t *testing.T) {
	var capturedArgs []string

	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(args ...string) (string, string, error) {
			capturedArgs = args
			return "BRANCH\tURL\n", "", nil
		},
	}

	client := git.NewGitHubCLI(mock)
	_, _ = client.GetLinkedBranches(789)

	expected := []string{"issue", "develop", "--list", "789"}
	assert.Equal(t, expected, capturedArgs)
}

// =============================================================================
// CreateLinkedBranch Tests
// =============================================================================

func TestGitHubCLI_CreateLinkedBranch_PassesCorrectArgsForCustomName(t *testing.T) {
	var capturedArgs []string

	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(args ...string) (string, string, error) {
			capturedArgs = args
			return "Created branch feat/my-branch based on main\n", "", nil
		},
	}

	client := git.NewGitHubCLI(mock)
	branchName, err := client.CreateLinkedBranch(123, "feat/my-branch", false)

	require.NoError(t, err)
	assert.Equal(t, "feat/my-branch", branchName)

	expected := []string{"issue", "develop", "123", "--name", "feat/my-branch"}
	assert.Equal(t, expected, capturedArgs)
}

func TestGitHubCLI_CreateLinkedBranch_PassesCorrectArgsForAutoGeneratedName(t *testing.T) {
	var capturedArgs []string
	callCount := 0

	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(args ...string) (string, string, error) {
			callCount++
			if callCount == 1 {
				capturedArgs = args
				return "Created branch 123-my-test-issue based on main\n", "", nil
			}
			// This shouldn't be called if branch name is in output
			return "", "", &mockError{message: "unexpected call"}
		},
	}

	client := git.NewGitHubCLI(mock)
	branchName, err := client.CreateLinkedBranch(123, "", false)

	require.NoError(t, err)
	assert.Equal(t, "123-my-test-issue", branchName)

	expected := []string{"issue", "develop", "123"}
	assert.Equal(t, expected, capturedArgs)
}

func TestGitHubCLI_CreateLinkedBranch_RespectsCheckoutFlag(t *testing.T) {
	var capturedArgs []string

	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(args ...string) (string, string, error) {
			capturedArgs = args
			return "Checked out branch feat/checkout-test for issue #123\n", "", nil
		},
	}

	client := git.NewGitHubCLI(mock)
	branchName, err := client.CreateLinkedBranch(123, "feat/checkout-test", true)

	require.NoError(t, err)
	assert.Equal(t, "feat/checkout-test", branchName)

	expected := []string{"issue", "develop", "123", "--name", "feat/checkout-test", "--checkout"}
	assert.Equal(t, expected, capturedArgs)
}

func TestGitHubCLI_CreateLinkedBranch_FallsBackToQueryingIssueForBranchName(t *testing.T) {
	callCount := 0

	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(_ ...string) (string, string, error) {
			callCount++
			if callCount == 1 {
				// First call: create branch - return output without parseable branch name
				return "branch created successfully\n", "", nil
			}
			// Second call: get issue to construct branch name
			issue := git.Issue{
				Number: 123,
				Title:  "Add New Feature",
				State:  "open",
			}
			jsonData, _ := json.Marshal(issue)
			return string(jsonData), "", nil
		},
	}

	client := git.NewGitHubCLI(mock)
	branchName, err := client.CreateLinkedBranch(123, "", false)

	require.NoError(t, err)
	assert.Equal(t, "123-add-new-feature", branchName)
}

func TestGitHubCLI_CreateLinkedBranch_ReturnsErrorOnFailure(t *testing.T) {
	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(_ ...string) (string, string, error) {
			return "", "branch already exists", &mockError{message: "exit 1"}
		},
	}

	client := git.NewGitHubCLI(mock)
	_, err := client.CreateLinkedBranch(123, "existing-branch", false)

	require.Error(t, err)
	var ghErr git.ErrGHCommand
	assert.ErrorAs(t, err, &ghErr)
}

// =============================================================================
// CreatePullRequest Tests
// =============================================================================

func TestGitHubCLI_CreatePullRequest_IncludesDraftFlagWhenDraftTrue(t *testing.T) {
	var capturedArgs []string

	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(args ...string) (string, string, error) {
			capturedArgs = args
			return "https://github.com/owner/repo/pull/42\n", "", nil
		},
	}

	client := git.NewGitHubCLI(mock)
	number, url, err := client.CreatePullRequest("Draft PR", "Draft body", true)

	require.NoError(t, err)
	assert.Equal(t, 42, number)
	assert.Equal(t, "https://github.com/owner/repo/pull/42", url)

	// Check --draft flag is present
	found := false
	for _, arg := range capturedArgs {
		if arg == "--draft" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected --draft flag in arguments")
}

func TestGitHubCLI_CreatePullRequest_OmitsDraftFlagWhenDraftFalse(t *testing.T) {
	var capturedArgs []string

	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(args ...string) (string, string, error) {
			capturedArgs = args
			return "https://github.com/owner/repo/pull/100\n", "", nil
		},
	}

	client := git.NewGitHubCLI(mock)
	number, url, err := client.CreatePullRequest("Ready PR", "Ready body", false)

	require.NoError(t, err)
	assert.Equal(t, 100, number)
	assert.Equal(t, "https://github.com/owner/repo/pull/100", url)

	// Check --draft flag is NOT present
	for _, arg := range capturedArgs {
		assert.NotEqual(t, "--draft", arg, "did not expect --draft flag when draft=false")
	}
}

func TestGitHubCLI_CreatePullRequest_ParsesPRNumberFromURL(t *testing.T) {
	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(_ ...string) (string, string, error) {
			return "https://github.com/myorg/myrepo/pull/789\n", "", nil
		},
	}

	client := git.NewGitHubCLI(mock)
	number, url, err := client.CreatePullRequest("Title", "Body", false)

	require.NoError(t, err)
	assert.Equal(t, 789, number)
	assert.Equal(t, "https://github.com/myorg/myrepo/pull/789", url)
}

func TestGitHubCLI_CreatePullRequest_ReturnsErrorOnInvalidURL(t *testing.T) {
	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(_ ...string) (string, string, error) {
			return "not-a-url\n", "", nil
		},
	}

	client := git.NewGitHubCLI(mock)
	_, _, err := client.CreatePullRequest("Title", "Body", false)

	require.Error(t, err)
}

func TestGitHubCLI_CreatePullRequest_ReturnsErrGHCommandOnFailure(t *testing.T) {
	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(_ ...string) (string, string, error) {
			return "", "no commits to create PR", &mockError{message: "exit 1"}
		},
	}

	client := git.NewGitHubCLI(mock)
	_, _, err := client.CreatePullRequest("Title", "Body", false)

	require.Error(t, err)
	var ghErr git.ErrGHCommand
	assert.ErrorAs(t, err, &ghErr)
}

func TestGitHubCLI_CreatePullRequest_ReturnsErrorWhenNotInstalled(t *testing.T) {
	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return false },
	}

	client := git.NewGitHubCLI(mock)
	_, _, err := client.CreatePullRequest("Title", "Body", false)

	require.Error(t, err)
	var notInstalled git.ErrGHNotInstalled
	assert.ErrorAs(t, err, &notInstalled)
}

// =============================================================================
// UpdatePullRequest Tests
// =============================================================================

func TestGitHubCLI_UpdatePullRequest_PassesCorrectArguments(t *testing.T) {
	var capturedArgs []string

	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(args ...string) (string, string, error) {
			capturedArgs = args
			return "", "", nil
		},
	}

	client := git.NewGitHubCLI(mock)
	err := client.UpdatePullRequest(123, "New Title", "New Body")

	require.NoError(t, err)

	expected := []string{"pr", "edit", "123", "--title", "New Title", "--body", "New Body"}
	assert.Equal(t, expected, capturedArgs)
}

func TestGitHubCLI_UpdatePullRequest_ReturnsErrorOnFailure(t *testing.T) {
	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(_ ...string) (string, string, error) {
			return "", "PR not found", &mockError{message: "exit 1"}
		},
	}

	client := git.NewGitHubCLI(mock)
	err := client.UpdatePullRequest(999, "Title", "Body")

	require.Error(t, err)
	var ghErr git.ErrGHCommand
	assert.ErrorAs(t, err, &ghErr)
}

func TestGitHubCLI_UpdatePullRequest_ReturnsErrorWhenNotInstalled(t *testing.T) {
	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return false },
	}

	client := git.NewGitHubCLI(mock)
	err := client.UpdatePullRequest(123, "Title", "Body")

	require.Error(t, err)
	var notInstalled git.ErrGHNotInstalled
	assert.ErrorAs(t, err, &notInstalled)
}

// =============================================================================
// MarkPullRequestReady Tests
// =============================================================================

func TestGitHubCLI_MarkPullRequestReady_PassesCorrectArguments(t *testing.T) {
	var capturedArgs []string

	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(args ...string) (string, string, error) {
			capturedArgs = args
			return "", "", nil
		},
	}

	client := git.NewGitHubCLI(mock)
	err := client.MarkPullRequestReady(42)

	require.NoError(t, err)

	expected := []string{"pr", "ready", "42"}
	assert.Equal(t, expected, capturedArgs)
}

func TestGitHubCLI_MarkPullRequestReady_ReturnsErrorOnFailure(t *testing.T) {
	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return true },
		RunSilentFunc: func(_ ...string) error {
			return nil
		},
		RunFunc: func(_ ...string) (string, string, error) {
			return "", "PR is not a draft", &mockError{message: "exit 1"}
		},
	}

	client := git.NewGitHubCLI(mock)
	err := client.MarkPullRequestReady(42)

	require.Error(t, err)
	var ghErr git.ErrGHCommand
	assert.ErrorAs(t, err, &ghErr)
}

func TestGitHubCLI_MarkPullRequestReady_ReturnsErrorWhenNotInstalled(t *testing.T) {
	mock := &mocks.ExecutorMock{
		ExistsFunc: func() bool { return false },
	}

	client := git.NewGitHubCLI(mock)
	err := client.MarkPullRequestReady(42)

	require.Error(t, err)
	var notInstalled git.ErrGHNotInstalled
	assert.ErrorAs(t, err, &notInstalled)
}

// =============================================================================
// Example Test
// =============================================================================

func Example() {
	// Create a mock executor that simulates gh CLI responses
	mock := &mocks.ExecutorMock{
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
	github := git.NewGitHubCLI(mock)

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
