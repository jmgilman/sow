package sowfs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmgilman/go/fs/core"
)

// ErrSowAlreadyInitialized indicates .sow directory already exists.
var ErrSowAlreadyInitialized = fmt.Errorf(".sow directory already exists - repository already initialized")

// Initialize creates the initial .sow directory structure.
//
// This function:
//  1. Verifies we're in a git repository (hard requirement)
//  2. Checks if .sow already exists (error if it does)
//  3. Creates the directory structure
//  4. Writes initial files with valid content
//
// Structure created:
//   .sow/
//   ├── .version              - Version tracking for migrations
//   ├── knowledge/            - Repository documentation (empty)
//   └── refs/                 - External references
//       ├── .gitignore        - Ignore symlinks, keep indexes
//       └── index.json        - Empty committed refs index
//
// Parameters:
//   - fs: Filesystem to operate on (rooted at current working directory)
//
// Returns:
//   - ErrNotInGitRepo if not in a git repository
//   - ErrSowAlreadyInitialized if .sow already exists
func Initialize(fs core.FS) error {
	// Get current working directory to find git root
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Verify we're in a git repository (hard requirement)
	_, err = findGitRepoRoot(cwd)
	if err != nil {
		return err
	}

	// Check if .sow already exists (filesystem is already rooted at cwd)
	exists, err := fs.Exists(DirSow)
	if err != nil {
		return fmt.Errorf("failed to check if .sow exists: %w", err)
	}
	if exists {
		return ErrSowAlreadyInitialized
	}

	// Create directory structure
	if err := createDirectories(fs); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Write initial files
	if err := writeInitialFiles(fs); err != nil {
		return fmt.Errorf("failed to write initial files: %w", err)
	}

	return nil
}

// createDirectories creates the .sow directory structure.
func createDirectories(fs core.FS) error {
	// Directories to create (relative to current working directory)
	dirs := []string{
		DirSow,           // .sow/
		DirKnowledge,     // .sow/knowledge/
		DirRefs,          // .sow/refs/
	}

	for _, dir := range dirs {
		if err := fs.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// writeInitialFiles writes the initial files with valid content.
func writeInitialFiles(fs core.FS) error {
	// Files to create with their content (relative to current working directory)
	files := map[string]string{
		filepath.Join(DirSow, FileVersion):           VersionFileContent,
		filepath.Join(DirSow, FileRefsCommittedIndex): RefsIndexContent,
		filepath.Join(DirSow, FileRefsGitignore):     RefsGitignoreContent,
	}

	for path, content := range files {
		if err := fs.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", path, err)
		}
	}

	return nil
}
