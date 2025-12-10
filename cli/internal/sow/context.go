package sow

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jmgilman/sow/libs/git"
)

// Context provides unified access to sow subsystems: filesystem, git, and GitHub.
// It is created once per CLI command invocation and passed to all subsystems.
type Context struct {
	fs       FS
	repo     *git.Git
	github   git.GitHubClient
	repoRoot string

	// Worktree support
	isWorktree   bool   // true if this context is in a worktree
	worktreePath string // path to worktree (only set if isWorktree=true)
	mainRepoRoot string // path to main repo root (always set)
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
	gitRepo, err := git.NewGit(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	// Detect if we're in a worktree by checking if .git is a file or directory
	gitPath := filepath.Join(repoRoot, ".git")
	info, err := os.Stat(gitPath)

	isWorktree := false
	var mainRepoRoot string

	if err == nil && !info.IsDir() {
		// .git is a file, not a directory → this is a worktree
		isWorktree = true

		// Read the .git file to find main repo location
		content, err := os.ReadFile(gitPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read .git file: %w", err)
		}

		// Parse "gitdir: /path/to/.git/worktrees/branch"
		gitdirLine := strings.TrimSpace(string(content))
		if !strings.HasPrefix(gitdirLine, "gitdir: ") {
			return nil, fmt.Errorf("invalid .git file format: expected 'gitdir:' prefix, got: %s", gitdirLine)
		}

		gitdirPath := strings.TrimPrefix(gitdirLine, "gitdir: ")

		// Main repo root is the parent of the .git directory
		// Example: /repo/.git/worktrees/feat/auth → /repo
		// Find the .git directory by looking for "worktrees" in the path
		// and going up 2 levels from there (.git/worktrees -> .git -> repo)

		// Split the path and find where "worktrees" is
		parts := strings.Split(gitdirPath, string(filepath.Separator))
		gitDirIndex := -1
		for i := len(parts) - 1; i >= 0; i-- {
			if parts[i] == "worktrees" && i > 0 && parts[i-1] == ".git" {
				gitDirIndex = i - 1
				break
			}
		}

		if gitDirIndex == -1 {
			return nil, fmt.Errorf("invalid worktree gitdir path: cannot find .git/worktrees in path: %s", gitdirPath)
		}

		// Reconstruct path up to parent of .git
		// Handle absolute paths correctly (first element might be empty for Unix paths starting with /)
		if len(parts[:gitDirIndex]) == 0 {
			mainRepoRoot = string(filepath.Separator)
		} else if parts[0] == "" {
			// Unix absolute path - prepend separator
			mainRepoRoot = string(filepath.Separator) + filepath.Join(parts[1:gitDirIndex]...)
		} else {
			mainRepoRoot = filepath.Join(parts[:gitDirIndex]...)
		}
	} else {
		// Normal repository - mainRepoRoot is same as repoRoot
		mainRepoRoot = repoRoot
	}

	return &Context{
		fs:           sowFS, // Will be nil if .sow doesn't exist
		repo:         gitRepo,
		repoRoot:     repoRoot,
		isWorktree:   isWorktree,
		worktreePath: repoRoot, // if worktree, this is the worktree path
		mainRepoRoot: mainRepoRoot,
	}, nil
}

// FS returns the filesystem for accessing .sow/ directory.
func (c *Context) FS() FS {
	return c.fs
}

// Git returns the Git repository wrapper.
func (c *Context) Git() *git.Git {
	return c.repo
}

// GitHub returns the GitHub client (lazy-loaded).
func (c *Context) GitHub() git.GitHubClient {
	if c.github == nil {
		c.github, _ = git.NewGitHubClient()
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

// IsWorktree returns true if this context is running in a git worktree.
func (c *Context) IsWorktree() bool {
	return c.isWorktree
}

// MainRepoRoot returns the root directory of the main repository.
// If in a worktree, returns the main repo's root. Otherwise, returns the current repo root.
func (c *Context) MainRepoRoot() string {
	return c.mainRepoRoot
}

// SinksPath returns the path to the .sow/sinks/ directory.
// If in a worktree, returns the path in the main repository.
// If in the main repo, returns the local path.
func (c *Context) SinksPath() string {
	if c.isWorktree {
		return filepath.Join(c.mainRepoRoot, ".sow", "sinks")
	}
	return filepath.Join(c.repoRoot, ".sow", "sinks")
}

// ReposPath returns the path to the .sow/repos/ directory.
// If in a worktree, returns the path in the main repository.
// If in the main repo, returns the local path.
func (c *Context) ReposPath() string {
	if c.isWorktree {
		return filepath.Join(c.mainRepoRoot, ".sow", "repos")
	}
	return filepath.Join(c.repoRoot, ".sow", "repos")
}
