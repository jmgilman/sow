package git

import "fmt"

// ErrGHNotInstalled is returned when the gh CLI is not found in PATH.
type ErrGHNotInstalled struct{}

func (e ErrGHNotInstalled) Error() string {
	return "GitHub CLI (gh) not found. Install from: https://cli.github.com/"
}

// ErrGHNotAuthenticated is returned when gh CLI is installed but not authenticated.
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

// ErrNotGitRepository is returned when the path is not a git repository.
type ErrNotGitRepository struct {
	Path string
}

func (e ErrNotGitRepository) Error() string {
	return fmt.Sprintf("%s is not a git repository", e.Path)
}

// ErrBranchExists is returned when attempting to create a branch that already exists.
type ErrBranchExists struct {
	Branch string
}

func (e ErrBranchExists) Error() string {
	return fmt.Sprintf("branch %s already exists", e.Branch)
}
