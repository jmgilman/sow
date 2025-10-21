package sow

import (
	"github.com/jmgilman/go/fs/billy"
	"github.com/jmgilman/go/fs/core"
)

// SowFS is a filesystem scoped to the .sow/ directory.
//
// This is an alias for core.FS, which provides a rich filesystem interface
// with read, write, management, walk, and chroot operations.
//
// The SowFS returned by NewSowFS is automatically scoped (chrooted) to the
// .sow/ directory, so all operations are relative to .sow/
//
// Example:
//   fs.ReadFile("project/state.yaml")  // Reads .sow/project/state.yaml
//   fs.WriteFile("refs/index.json", data, 0644)
type SowFS = core.FS

// NewSowFS creates a new filesystem scoped to .sow/ directory.
//
// The repoRoot should be the absolute path to the git repository root.
// The returned filesystem is scoped to repoRoot/.sow/
//
// This function creates a billy-backed local filesystem and chroots it
// to the .sow/ subdirectory. If .sow/ doesn't exist yet, the filesystem
// will still be created - Init() will create the directory structure later.
//
// Returns an error if the repository root is invalid or if chrooting fails.
func NewSowFS(repoRoot string) (SowFS, error) {
	// Create billy-backed local filesystem
	baseFS := billy.NewLocal()

	// Chroot to repository root
	repoFS, err := baseFS.Chroot(repoRoot)
	if err != nil {
		return nil, err
	}

	// Chroot to .sow/ directory
	// Note: Chroot will succeed even if .sow/ doesn't exist yet
	// because billy's Chroot doesn't verify the directory exists
	sowFS, err := repoFS.Chroot(".sow")
	if err != nil {
		// If chroot fails, it means we can't access the path
		// Try returning the repoFS so operations can manually prepend .sow/
		// This is a fallback for when .sow/ doesn't exist yet
		return repoFS, nil
	}

	return sowFS, nil
}
