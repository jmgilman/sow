package sow

import (
	"errors"
	"fmt"

	"github.com/jmgilman/go/git"
)

// Context provides unified access to sow subsystems: filesystem, git, and GitHub.
// It is created once per CLI command invocation and passed to all subsystems.
type Context struct {
	fs       FS
	repo     *Git
	github   *GitHub
	repoRoot string
}

// NewContext creates a new sow context rooted at the given repository directory.
// If .sow/ doesn't exist, the context is still created but with fs set to nil.
// Use IsInitialized() to check if .sow/ exists.
func NewContext(repoRoot string) (*Context, error) {
	// Try to create FS (chrooted to .sow/)
	// If .sow doesn't exist, set fs to nil but don't fail
	sowFS, err := NewFS(repoRoot)
	if err != nil && !errors.Is(err, ErrNotInitialized) {
		return nil, fmt.Errorf("failed to create FS: %w", err)
	}

	// Open git repository
	gitRepo, err := git.Open(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	return &Context{
		fs:       sowFS, // Will be nil if .sow doesn't exist
		repo:     &Git{repo: gitRepo},
		repoRoot: repoRoot,
	}, nil
}

// FS returns the filesystem for accessing .sow/ directory.
func (c *Context) FS() FS {
	return c.fs
}

// Git returns the Git repository wrapper.
func (c *Context) Git() *Git {
	return c.repo
}

// GitHub returns the GitHub client (lazy-loaded).
func (c *Context) GitHub() *GitHub {
	if c.github == nil {
		c.github = &GitHub{}
	}
	return c.github
}

// RepoRoot returns the repository root directory path.
func (c *Context) RepoRoot() string {
	return c.repoRoot
}

// IsInitialized returns true if .sow/ directory exists (FS is not nil).
func (c *Context) IsInitialized() bool {
	return c.fs != nil
}
