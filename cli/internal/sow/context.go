// Package sow provides the core context and primitives for the sow system.
//
// The Context type is the primary interface for all sow subsystems, providing
// access to the filesystem, git repository, and GitHub client.
package sow

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/exec"
)

// Context provides access to sow subsystems.
//
// Context is created once per CLI command invocation and provides a unified
// interface for filesystem operations, git repository access, and GitHub API
// interactions. All subsystems (project, refs, etc.) receive a Context to
// perform their operations.
//
// The Context is immutable after creation - subsystems should not modify it.
type Context struct {
	fs     SowFS
	git    *Git
	github *GitHub // Lazy-loaded
}

// NewContext creates a new Context rooted at the repository directory.
//
// This function:
//   - Creates a SowFS scoped to .sow/ directory
//   - Initializes Git repository access
//   - Prepares GitHub client (lazy-loaded on first access)
//
// The repoRoot should be the absolute path to the git repository root.
// This is typically detected by walking up from the current directory to find .git/
//
// Returns an error if:
//   - repoRoot is not a valid directory
//   - Cannot create filesystem abstraction
//   - Git repository cannot be accessed
func NewContext(repoRoot string) (*Context, error) {
	// Create SowFS scoped to .sow/
	sowFS, err := NewSowFS(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to create sow filesystem: %w", err)
	}

	// Create Git repository access
	git, err := NewGit(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize git repository: %w", err)
	}

	return &Context{
		fs:     sowFS,
		git:    git,
		github: nil, // Lazy-loaded
	}, nil
}

// FS returns the SowFS for filesystem operations scoped to .sow/
//
// All filesystem operations should use this interface rather than direct
// file system access to ensure proper scoping and portability.
func (c *Context) FS() SowFS {
	return c.fs
}

// Git returns the Git repository for version control operations.
//
// Provides access to git operations like getting current branch,
// checking repository status, etc.
func (c *Context) Git() *Git {
	return c.git
}

// GitHub returns the GitHub client for API operations.
//
// The client is lazy-loaded on first access. This avoids the overhead
// of GitHub initialization for commands that don't need it.
//
// Note: GitHub operations may fail if not authenticated or if network
// is unavailable. Callers should handle errors appropriately.
func (c *Context) GitHub() *GitHub {
	if c.github == nil {
		ghExec := exec.NewLocal("gh")
		c.github = NewGitHub(ghExec)
	}
	return c.github
}
