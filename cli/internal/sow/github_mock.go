package sow

// MockGitHub is a mock implementation of GitHubClient for testing.
//
// MockGitHub allows you to control the behavior of GitHub operations in tests
// by providing custom functions for each interface method. This is useful for
// testing higher-level code that depends on GitHubClient (like commands and
// wizards) without making actual GitHub API calls or CLI invocations.
//
// Usage in tests:
//
//	mock := &sow.MockGitHub{
//	    CheckAvailabilityFunc: func() error { return nil },
//	    GetIssueFunc: func(n int) (*sow.Issue, error) {
//	        return &sow.Issue{Number: n, Title: "Test Issue"}, nil
//	    },
//	}
//	// Use mock in your test code
//	issue, err := mock.GetIssue(123)
//
// Default Behavior (when function fields are nil):
//
//   - CheckAvailability: returns nil (success)
//   - ListIssues: returns empty slice []Issue{} with no error
//   - GetIssue: returns nil issue with no error
//   - CreateIssue: returns nil issue with no error
//   - GetLinkedBranches: returns empty slice []LinkedBranch{} with no error
//   - CreateLinkedBranch: returns empty string with no error
//   - CreatePullRequest: returns 0, "" with no error
//   - UpdatePullRequest: returns nil (success)
//   - MarkPullRequestReady: returns nil (success)
//
// This default behavior allows you to mock only the methods you care about in
// your test, while other methods return safe, non-error values.
//
// When to Use MockGitHub vs MockExecutor:
//
// Use MockGitHub when:
//   - Testing code that depends on the GitHubClient interface
//   - Testing commands (issue list, issue check, etc.)
//   - Testing wizard flows that use GitHub operations
//   - Testing any high-level code that calls GitHub operations
//
// Use MockExecutor (from exec package) when:
//   - Testing the GitHubCLI implementation itself (github_cli.go)
//   - Verifying that correct gh CLI commands are being called
//   - Testing error handling in the CLI layer
//   - Testing command argument formatting and parsing
//
// Examples:
//
// Testing with custom functions:
//
//	mock := &sow.MockGitHub{
//	    ListIssuesFunc: func(label, state string) ([]sow.Issue, error) {
//	        return []sow.Issue{
//	            {Number: 1, Title: "Issue 1"},
//	            {Number: 2, Title: "Issue 2"},
//	        }, nil
//	    },
//	}
//	issues, err := mock.ListIssues("sow", "open")
//	// Returns 2 test issues
//
// Testing error conditions:
//
//	mock := &sow.MockGitHub{
//	    CheckAvailabilityFunc: func() error {
//	        return ErrGHNotInstalled{}
//	    },
//	}
//	err := mock.CheckAvailability()
//	// Returns ErrGHNotInstalled
//
// Minimal mock (using defaults for unused methods):
//
//	mock := &sow.MockGitHub{
//	    CheckAvailabilityFunc: func() error { return nil },
//	}
//	// All other methods return sensible defaults
//	issues, _ := mock.ListIssues("sow", "open")  // Returns []
//	issue, _ := mock.GetIssue(123)               // Returns nil
type MockGitHub struct {
	CheckAvailabilityFunc    func() error
	ListIssuesFunc           func(label, state string) ([]Issue, error)
	GetIssueFunc             func(number int) (*Issue, error)
	CreateIssueFunc          func(title, body string, labels []string) (*Issue, error)
	GetLinkedBranchesFunc    func(number int) ([]LinkedBranch, error)
	CreateLinkedBranchFunc   func(issueNumber int, branchName string, checkout bool) (string, error)
	CreatePullRequestFunc    func(title, body string, draft bool) (number int, url string, err error)
	UpdatePullRequestFunc    func(number int, title, body string) error
	MarkPullRequestReadyFunc func(number int) error
}

// CheckAvailability calls the mock function if set, otherwise returns nil (success).
func (m *MockGitHub) CheckAvailability() error {
	if m.CheckAvailabilityFunc != nil {
		return m.CheckAvailabilityFunc()
	}
	return nil
}

// ListIssues calls the mock function if set, otherwise returns an empty slice.
func (m *MockGitHub) ListIssues(label, state string) ([]Issue, error) {
	if m.ListIssuesFunc != nil {
		return m.ListIssuesFunc(label, state)
	}
	return []Issue{}, nil
}

// GetIssue calls the mock function if set, otherwise returns nil.
func (m *MockGitHub) GetIssue(number int) (*Issue, error) {
	if m.GetIssueFunc != nil {
		return m.GetIssueFunc(number)
	}
	return nil, nil
}

// CreateIssue calls the mock function if set, otherwise returns nil.
func (m *MockGitHub) CreateIssue(title, body string, labels []string) (*Issue, error) {
	if m.CreateIssueFunc != nil {
		return m.CreateIssueFunc(title, body, labels)
	}
	return nil, nil
}

// GetLinkedBranches calls the mock function if set, otherwise returns an empty slice.
func (m *MockGitHub) GetLinkedBranches(number int) ([]LinkedBranch, error) {
	if m.GetLinkedBranchesFunc != nil {
		return m.GetLinkedBranchesFunc(number)
	}
	return []LinkedBranch{}, nil
}

// CreateLinkedBranch calls the mock function if set, otherwise returns an empty string.
func (m *MockGitHub) CreateLinkedBranch(issueNumber int, branchName string, checkout bool) (string, error) {
	if m.CreateLinkedBranchFunc != nil {
		return m.CreateLinkedBranchFunc(issueNumber, branchName, checkout)
	}
	return "", nil
}

// CreatePullRequest calls the mock function if set, otherwise returns zeros.
func (m *MockGitHub) CreatePullRequest(title, body string, draft bool) (number int, url string, err error) {
	if m.CreatePullRequestFunc != nil {
		return m.CreatePullRequestFunc(title, body, draft)
	}
	return 0, "", nil
}

// UpdatePullRequest calls the mock function if set, otherwise returns nil (success).
func (m *MockGitHub) UpdatePullRequest(number int, title, body string) error {
	if m.UpdatePullRequestFunc != nil {
		return m.UpdatePullRequestFunc(number, title, body)
	}
	return nil
}

// MarkPullRequestReady calls the mock function if set, otherwise returns nil (success).
func (m *MockGitHub) MarkPullRequestReady(number int) error {
	if m.MarkPullRequestReadyFunc != nil {
		return m.MarkPullRequestReadyFunc(number)
	}
	return nil
}

// Compile-time check that MockGitHub implements GitHubClient.
var _ GitHubClient = (*MockGitHub)(nil)
