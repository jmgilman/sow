package sow

import (
	"fmt"

	"github.com/jmgilman/go/fs/billy"
	"github.com/jmgilman/go/fs/core"
)

// FS is a filesystem scoped to the .sow/ directory.
//
// This is an alias for core.FS, which provides a rich filesystem interface
// with read, write, management, walk, and chroot operations.
//
// The FS returned by NewFS is automatically scoped (chrooted) to the
// .sow/ directory, so all operations are relative to .sow/
//
// Example:
//   fs.ReadFile("project/state.yaml")  // Reads .sow/project/state.yaml
//   fs.WriteFile("refs/index.json", data, 0644)
type FS = core.FS

// NewFS creates a new filesystem scoped to .sow/ directory.
//
// The repoRoot should be the absolute path to the git repository root.
// The returned filesystem is scoped to repoRoot/.sow/
//
// Returns ErrNotInitialized if .sow/ doesn't exist (use Init() to create it first).
func NewFS(repoRoot string) (FS, error) {
	// Check if .sow exists
	if !isInitialized(repoRoot) {
		return nil, ErrNotInitialized
	}

	// Create billy-backed local filesystem
	baseFS := billy.NewLocal()

	// Chroot to repository root
	repoFS, err := baseFS.Chroot(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to chroot to repo root: %w", err)
	}

	// Chroot to .sow/ directory
	sowFS, err := repoFS.Chroot(".sow")
	if err != nil {
		return nil, fmt.Errorf("failed to chroot to .sow: %w", err)
	}

	return sowFS, nil
}
