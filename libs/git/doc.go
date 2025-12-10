// Package git provides Git and GitHub operations for sow using the
// ports and adapters pattern.
//
// This package defines interfaces (ports) for Git repository operations and
// GitHub API interactions. The interfaces enable dependency injection and easy
// mocking in tests, decoupling business logic from the actual Git/GitHub
// implementation mechanism.
//
// # Design Pattern
//
// The package follows the ports and adapters (hexagonal) architecture:
//
//   - Ports: Interfaces defining contracts for Git and GitHub operations
//   - Adapters: Concrete implementations using git CLI, gh CLI, or go-git library
//
// This design allows consumers to:
//   - Easily mock dependencies for unit testing
//   - Swap implementations without changing business logic
//   - Test code in isolation from external systems
//
// # Data Types
//
// The package provides data types for GitHub entities:
//
//   - [Issue]: Represents a GitHub issue with labels and metadata
//   - [Label]: Represents a GitHub label
//   - [LinkedBranch]: Represents a branch linked to an issue
//
// # Error Types
//
// The package defines error types for common failure scenarios:
//
//   - [ErrGHNotInstalled]: GitHub CLI (gh) not found in PATH
//   - [ErrGHNotAuthenticated]: gh CLI not authenticated
//   - [ErrGHCommand]: gh CLI command failed
//   - [ErrNotGitRepository]: Path is not a git repository
//   - [ErrBranchExists]: Branch already exists
//
// # Usage
//
// Working with issues:
//
//	issue := &git.Issue{
//	    Number: 123,
//	    Title:  "My Issue",
//	    Labels: []git.Label{{Name: "bug"}, {Name: "sow"}},
//	}
//
//	if issue.HasLabel("sow") {
//	    fmt.Println("This is a sow issue")
//	}
//
// Handling errors:
//
//	err := someGitOperation()
//	var notInstalled git.ErrGHNotInstalled
//	if errors.As(err, &notInstalled) {
//	    fmt.Println("Please install gh: https://cli.github.com/")
//	}
//
// # Testing
//
// Use the generated mocks from the mocks subpackage for testing:
//
//	import "github.com/jmgilman/sow/libs/git/mocks"
//
// Mocks are generated using moq for all public interfaces.
//
// # Implementations
//
// This package provides:
//   - Data types for GitHub entities (Issue, Label, LinkedBranch)
//   - Error types for Git/GitHub operations
//   - Generated mocks via moq in the mocks subpackage for testing
//
// Additional implementations (GitClient, GitHubClient) are defined in separate
// files within this package.
package git
