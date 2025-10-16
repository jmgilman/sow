package sowfs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmgilman/go/fs/billy"
	"github.com/jmgilman/go/fs/core"
	"github.com/jmgilman/sow/cli/schemas"
)

// Common errors for SowFS operations.
var (
	// ErrNotInGitRepo indicates the current directory is not in a git repository.
	ErrNotInGitRepo = errors.New("not in git repository")

	// ErrSowNotInitialized indicates .sow directory was not found.
	ErrSowNotInitialized = errors.New(".sow directory not found - run 'sow init' first")

	// ErrProjectNotFound indicates no active project exists.
	ErrProjectNotFound = errors.New("no active project found")

	// ErrTaskNotFound indicates the requested task does not exist.
	ErrTaskNotFound = errors.New("task not found")

	// ErrInvalidTaskID indicates the task ID format is invalid.
	ErrInvalidTaskID = errors.New("invalid task ID - must be gap-numbered (010, 020, etc)")
)


// SowFS provides access to the .sow directory structure with domain-specific abstractions.
//
// This is the main entry point for all .sow directory operations. It enforces the
// directory structure and provides type-safe, validated access to state files.
//
// The filesystem is automatically chrooted to the .sow directory, so all paths
// in domain-specific interfaces are relative to .sow/.
type SowFS interface {
	// Knowledge returns the knowledge domain for accessing repository documentation.
	// This includes architecture docs, ADRs, and other committed knowledge.
	Knowledge() KnowledgeFS

	// Refs returns the refs domain for managing external knowledge and code references.
	// This handles both committed refs (team-shared) and local refs (developer-specific).
	Refs() RefsFS

	// Project returns the project domain for accessing active project state.
	// Returns ErrProjectNotFound if no project is currently active.
	Project() (ProjectFS, error)

	// RepoRoot returns the absolute path to the git repository root.
	RepoRoot() string

	// Close cleans up any resources held by the filesystem.
	// Should be called when done using SowFS.
	Close() error
}

// NewSowFS creates a SowFS from the current working directory.
//
// This function:
//  1. Walks up the directory tree to find the git repository root
//  2. Verifies a .sow directory exists at the repository root
//  3. Creates a billy filesystem chrooted to .sow/
//  4. Returns a SowFS implementation
//
// Returns:
//   - ErrNotInGitRepo if not in a git repository
//   - ErrSowNotInitialized if .sow directory not found
func NewSowFS() (*SowFSImpl, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	return NewSowFSFromPath(cwd)
}

// NewSowFSFromPath creates a SowFS from a specific repository path.
//
// Similar to NewSowFS but starts from the given path instead of the
// current working directory.
//
// Parameters:
//   - repoPath: absolute path to a directory within the git repository
//
// Returns:
//   - ErrNotInGitRepo if path is not in a git repository
//   - ErrSowNotInitialized if .sow directory not found
func NewSowFSFromPath(repoPath string) (*SowFSImpl, error) {
	// Find git repository root
	repoRoot, err := findGitRepoRoot(repoPath)
	if err != nil {
		return nil, err
	}

	// Create local filesystem
	localFS := billy.NewLocal()

	// Verify .sow exists and is a directory at repo root
	sowPath := filepath.Join(repoRoot, ".sow")
	info, err := localFS.Stat(sowPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrSowNotInitialized
		}
		return nil, fmt.Errorf("failed to check .sow directory: %w", err)
	}
	if !info.IsDir() {
		return nil, ErrSowNotInitialized
	}

	// Chroot filesystem to .sow/
	chrootFS, err := localFS.Chroot(sowPath)
	if err != nil {
		return nil, fmt.Errorf("failed to chroot to .sow: %w", err)
	}

	// Load CUE validator
	validator, err := schemas.NewCUEValidator()
	if err != nil {
		return nil, fmt.Errorf("failed to load CUE schemas: %w", err)
	}

	// Create SowFS implementation
	sowFS := &SowFSImpl{
		fs:        chrootFS,
		repoRoot:  repoRoot,
		validator: validator,
	}

	// Eagerly initialize all domain implementations and validate
	if err := sowFS.initialize(); err != nil {
		return nil, err
	}

	return sowFS, nil
}

// NewSowFSWithFS creates a SowFS with an existing filesystem (for testing).
//
// This constructor is primarily for testing. It doesn't validate git repository
// structure, only that the .sow directory exists in the provided filesystem.
//
// Parameters:
//   - fs: filesystem containing .sow directory
//   - repoRoot: absolute path to use as repository root
//
// Returns:
//   - ErrSowNotInitialized if .sow directory not found in fs
func NewSowFSWithFS(fs core.FS, repoRoot string) (*SowFSImpl, error) {
	// Verify .sow exists and is a directory
	info, err := fs.Stat(".sow")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrSowNotInitialized
		}
		return nil, fmt.Errorf("failed to check .sow directory: %w", err)
	}
	if !info.IsDir() {
		return nil, ErrSowNotInitialized
	}

	// Chroot filesystem to .sow/
	chrootFS, err := fs.Chroot(".sow")
	if err != nil {
		return nil, fmt.Errorf("failed to chroot to .sow: %w", err)
	}

	// Load CUE validator
	validator, err := schemas.NewCUEValidator()
	if err != nil {
		return nil, fmt.Errorf("failed to load CUE schemas: %w", err)
	}

	// Create SowFS implementation
	sowFS := &SowFSImpl{
		fs:        chrootFS,
		repoRoot:  repoRoot,
		validator: validator,
	}

	// Eagerly initialize all domain implementations and validate
	if err := sowFS.initialize(); err != nil {
		return nil, err
	}

	return sowFS, nil
}

// SowFSImpl is the concrete implementation of SowFS.
// Exported for type assertions but use via SowFS interface.
//
//nolint:revive // Name is intentional: SowFS is the interface, SowFSImpl is the implementation
type SowFSImpl struct {
	// fs is chrooted to .sow/ directory
	fs core.FS

	// repoRoot is the absolute path to the git repository root
	repoRoot string

	// validator provides CUE schema validation
	validator *schemas.CUEValidator

	// Domain-specific implementations (eagerly initialized during NewSowFS*)
	knowledge *KnowledgeFSImpl
	refs      *RefsFSImpl
	project   *ProjectFSImpl
}

// Ensure SowFSImpl implements SowFS interface.
var _ SowFS = (*SowFSImpl)(nil)

// Knowledge returns the knowledge domain (pre-initialized during construction).
func (s *SowFSImpl) Knowledge() KnowledgeFS {
	return s.knowledge
}

// Refs returns the refs domain (pre-initialized during construction).
func (s *SowFSImpl) Refs() RefsFS {
	return s.refs
}

// Project returns the project domain (pre-initialized during construction)
// Returns ErrProjectNotFound if no active project exists in the repository.
func (s *SowFSImpl) Project() (ProjectFS, error) {
	// Check if project exists
	exists, err := s.project.Exists()
	if err != nil {
		return nil, fmt.Errorf("failed to check project: %w", err)
	}
	if !exists {
		return nil, ErrProjectNotFound
	}

	return s.project, nil
}

// RepoRoot returns the repository root path.
func (s *SowFSImpl) RepoRoot() string {
	return s.repoRoot
}

// Close cleans up resources.
func (s *SowFSImpl) Close() error {
	// No resources to cleanup currently
	return nil
}

// initialize eagerly initializes all domain implementations and validates the .sow structure.
//
// This method is called by all NewSowFS* constructors to ensure that:
//  1. All domain implementations are initialized upfront
//  2. The .sow directory structure is validated before any operations
//  3. Commands only operate on valid .sow structures (fail-fast)
//
// Returns an error if validation fails, preventing the SowFS from being used.
func (s *SowFSImpl) initialize() error {
	// Eagerly initialize all domain implementations
	s.knowledge = NewKnowledgeFS(s, s.validator)
	s.refs = NewRefsFS(s, s.validator)
	s.project = NewProjectFS(s, s.validator)

	// Validate entire .sow structure
	result := s.ValidateAll()
	if result.HasErrors() {
		return fmt.Errorf("validation failed: %w", result)
	}

	return nil
}

// ValidateAll performs comprehensive validation of the entire .sow directory structure.
//
// This validates all files that exist without fail-fast behavior, collecting
// all validation errors into a single result. Optional components (project, tasks)
// are only validated if they exist.
//
// Validated components:
//   - .sow/refs/index.json (if exists)
//   - .sow/refs/index.local.json (if exists)
//   - .sow/project/state.yaml (if exists)
//   - .sow/project/phases/implementation/tasks/*/state.yaml (if exist)
//
// Returns a ValidationResult containing all validation errors found.
// If no errors are found, ValidationResult.HasErrors() returns false.
func (s *SowFSImpl) ValidateAll() *ValidationResult {
	result := &ValidationResult{}

	// Validate refs committed index (if exists)
	s.validateRefsCommittedIndex(result)

	// Validate refs local index (if exists)
	s.validateRefsLocalIndex(result)

	// Validate project (if exists)
	s.validateProject(result)

	return result
}

// validateRefsCommittedIndex validates the committed refs index.
func (s *SowFSImpl) validateRefsCommittedIndex(result *ValidationResult) {
	// Check if file exists
	exists, err := s.fs.Exists(FileRefsCommittedIndex)
	if err != nil {
		result.Add(FileRefsCommittedIndex, "refs-committed", fmt.Errorf("failed to check file: %w", err))
		return
	}
	if !exists {
		// File is optional - not an error
		return
	}

	// Read file
	data, err := s.fs.ReadFile(FileRefsCommittedIndex)
	if err != nil {
		result.Add(FileRefsCommittedIndex, "refs-committed", fmt.Errorf("failed to read file: %w", err))
		return
	}

	// Validate against schema
	if err := s.validator.ValidateRefsCommittedIndex(data); err != nil {
		result.Add(FileRefsCommittedIndex, "refs-committed", err)
	}
}

// validateRefsLocalIndex validates the local refs index.
func (s *SowFSImpl) validateRefsLocalIndex(result *ValidationResult) {
	// Check if file exists
	exists, err := s.fs.Exists(FileRefsLocalIndex)
	if err != nil {
		result.Add(FileRefsLocalIndex, "refs-local", fmt.Errorf("failed to check file: %w", err))
		return
	}
	if !exists {
		// File is optional - not an error
		return
	}

	// Read file
	data, err := s.fs.ReadFile(FileRefsLocalIndex)
	if err != nil {
		result.Add(FileRefsLocalIndex, "refs-local", fmt.Errorf("failed to read file: %w", err))
		return
	}

	// Validate against schema
	if err := s.validator.ValidateRefsLocalIndex(data); err != nil {
		result.Add(FileRefsLocalIndex, "refs-local", err)
	}
}

// validateProject validates the project directory and all its contents.
func (s *SowFSImpl) validateProject(result *ValidationResult) {
	// Check if project exists
	exists, err := s.fs.Exists(FileProjectState)
	if err != nil {
		result.Add(FileProjectState, "project-state", fmt.Errorf("failed to check file: %w", err))
		return
	}
	if !exists {
		// No active project - not an error
		return
	}

	// Validate project state
	data, err := s.fs.ReadFile(FileProjectState)
	if err != nil {
		result.Add(FileProjectState, "project-state", fmt.Errorf("failed to read file: %w", err))
		return
	}

	if err := s.validator.ValidateProjectState(data); err != nil {
		result.Add(FileProjectState, "project-state", err)
	}

	// Validate all tasks
	s.validateTasks(result)
}

// validateTasks validates all task state files.
func (s *SowFSImpl) validateTasks(result *ValidationResult) {
	// Check if tasks directory exists
	exists, err := s.fs.Exists(PathProjectTasksDir)
	if err != nil {
		result.Add(PathProjectTasksDir, "tasks", fmt.Errorf("failed to check directory: %w", err))
		return
	}
	if !exists {
		// No tasks yet - not an error
		return
	}

	// Read directory entries
	entries, err := s.fs.ReadDir(PathProjectTasksDir)
	if err != nil {
		result.Add(PathProjectTasksDir, "tasks", fmt.Errorf("failed to read directory: %w", err))
		return
	}

	// Validate each task directory
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		taskID := entry.Name()
		taskStatePath := filepath.Join(PathProjectTasksDir, taskID, FileTaskState)

		// Check if state file exists
		exists, err := s.fs.Exists(taskStatePath)
		if err != nil {
			result.Add(taskStatePath, "task-state", fmt.Errorf("failed to check file: %w", err))
			continue
		}
		if !exists {
			result.Add(taskStatePath, "task-state", fmt.Errorf("task directory exists but %s is missing", FileTaskState))
			continue
		}

		// Read state file
		data, err := s.fs.ReadFile(taskStatePath)
		if err != nil {
			result.Add(taskStatePath, "task-state", fmt.Errorf("failed to read file: %w", err))
			continue
		}

		// Validate against schema
		if err := s.validator.ValidateTaskState(data); err != nil {
			result.Add(taskStatePath, "task-state", err)
		}
	}
}

// findGitRepoRoot walks up the directory tree to find the git repository root.
//
// Starting from the given path, it walks up parent directories looking for
// a .git directory. Returns the absolute path to the repository root.
//
// Parameters:
//   - startPath: absolute path to start searching from
//
// Returns:
//   - Repository root path (absolute)
//   - ErrNotInGitRepo if no .git directory found
func findGitRepoRoot(startPath string) (string, error) {
	// Make path absolute
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Walk up directory tree
	currentPath := absPath
	for {
		// Check if .git exists in current directory
		gitPath := filepath.Join(currentPath, ".git")
		info, err := os.Stat(gitPath)
		if err == nil {
			// .git exists - check if it's a directory or file (worktree case)
			if info.IsDir() {
				// Normal repository
				return currentPath, nil
			}
			// .git is a file (git worktree) - read it to find actual repo
			// For now, we'll treat the directory containing .git file as the root
			// since we're operating in the worktree context
			return currentPath, nil
		}

		// Move to parent directory
		parentPath := filepath.Dir(currentPath)

		// Check if we've reached filesystem root
		if parentPath == currentPath {
			// Reached root without finding .git
			return "", ErrNotInGitRepo
		}

		currentPath = parentPath
	}
}
