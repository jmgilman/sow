// Package sow provides a unified domain abstraction layer for the sow CLI.
//
// This package encapsulates filesystem operations and state machine logic,
// allowing CLI commands to be a thin presentation layer. All business logic
// for managing projects, tasks, and phases lives here.
package sow

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/jmgilman/sow/cli/internal/statechart"
	"gopkg.in/yaml.v3"
)

const (
	// StructureVersion is the current .sow structure version.
	// This version is written to .sow/.version during initialization.
	StructureVersion = "1.0.0"
)

// Sow is the main entrypoint for interacting with the sow system.
// It provides high-level operations for managing projects, initialization,
// and repository state.
type Sow struct {
	fs billy.Filesystem
}

// New creates a new Sow instance with the given filesystem.
// The filesystem should be rooted at the repository directory.
func New(fs billy.Filesystem) *Sow {
	return &Sow{fs: fs}
}

// FS returns the underlying filesystem for advanced operations.
// Most callers should not need direct filesystem access.
func (s *Sow) FS() billy.Filesystem {
	return s.fs
}

// RepoRoot returns the repository root directory path.
func (s *Sow) RepoRoot() string {
	return s.fs.Root()
}

// Branch returns the current git branch name.
func (s *Sow) Branch() string {
	branch, err := s.getCurrentBranch()
	if err != nil {
		return ""
	}
	return branch
}

// DetectContext detects the current workspace context.
// Returns the context type ("none", "project", or "task") and task ID if applicable.
func (s *Sow) DetectContext() (string, string) {
	// Check if we're in a task directory
	// Task directories are at .sow/project/phases/{phase}/tasks/{task-id}/
	cwd, err := os.Getwd()
	if err != nil {
		return "none", ""
	}

	repoRoot := s.RepoRoot()

	// Check if current dir is under .sow/project/phases/*/tasks/*
	relPath, err := filepath.Rel(repoRoot, cwd)
	if err != nil {
		return "none", ""
	}

	// Normalize to forward slashes for parsing
	relPath = filepath.ToSlash(relPath)

	// Split path into components
	parts := strings.Split(relPath, "/")

	if len(parts) >= 6 {
		if parts[0] == ".sow" && parts[1] == "project" && parts[2] == "phases" && parts[4] == "tasks" {
			return "task", parts[5]
		}
	}

	// Check if we're anywhere under .sow/project
	if len(parts) >= 2 && parts[0] == ".sow" && parts[1] == "project" {
		return "project", ""
	}

	return "none", ""
}

// Init initializes the sow structure in the repository.
// Creates .sow/ directory with knowledge/ and refs/ subdirectories.
func (s *Sow) Init() error {
	// Check if in a git repository
	if _, err := s.fs.Stat(".git"); err != nil {
		return fmt.Errorf("not in git repository")
	}

	// Check if already initialized
	if s.IsInitialized() {
		return fmt.Errorf(".sow directory already exists - repository already initialized")
	}

	// Create base .sow directory
	if err := s.fs.MkdirAll(".sow", 0755); err != nil {
		return fmt.Errorf("failed to create .sow directory: %w", err)
	}

	// Create knowledge directory
	if err := s.fs.MkdirAll(".sow/knowledge", 0755); err != nil {
		return fmt.Errorf("failed to create knowledge directory: %w", err)
	}

	// Create knowledge/adrs directory
	if err := s.fs.MkdirAll(".sow/knowledge/adrs", 0755); err != nil {
		return fmt.Errorf("failed to create adrs directory: %w", err)
	}

	// Create refs directory
	if err := s.fs.MkdirAll(".sow/refs", 0755); err != nil {
		return fmt.Errorf("failed to create refs directory: %w", err)
	}

	// Create .gitignore for refs
	gitignorePath := filepath.Join(".sow", "refs", ".gitignore")
	gitignoreContent := []byte("# Ignore all symlinks and local refs\n*\n!.gitignore\n!index.json\n!index.local.json\n")
	if err := s.writeFile(gitignorePath, gitignoreContent); err != nil {
		return fmt.Errorf("failed to create refs .gitignore: %w", err)
	}

	// Create version file
	versionPath := filepath.Join(".sow", ".version")
	versionContent := []byte(StructureVersion + "\n")
	if err := s.writeFile(versionPath, versionContent); err != nil {
		return fmt.Errorf("failed to create version file: %w", err)
	}

	// Create refs index with version
	indexPath := filepath.Join(".sow", "refs", "index.json")
	indexContent := []byte(`{
  "version": "1.0.0",
  "refs": []
}
`)
	if err := s.writeFile(indexPath, indexContent); err != nil {
		return fmt.Errorf("failed to create refs index: %w", err)
	}

	return nil
}

// IsInitialized checks if sow has been initialized in the repository.
func (s *Sow) IsInitialized() bool {
	_, err := s.fs.Stat(".sow")
	return err == nil
}

// HasProject checks if an active project exists.
func (s *Sow) HasProject() bool {
	_, err := s.fs.Stat(".sow/project/state.yaml")
	return err == nil
}

// GetProject loads the active project from disk.
// Returns ErrNoProject if no project exists.
func (s *Sow) GetProject() (*Project, error) {
	if !s.HasProject() {
		return nil, ErrNoProject
	}

	machine, err := statechart.LoadFS(s.fs)
	if err != nil {
		return nil, fmt.Errorf("failed to load project state: %w", err)
	}

	return &Project{
		sow:     s,
		machine: machine,
	}, nil
}

// CreateProject creates a new project with the given name and description.
// Returns ErrProjectExists if a project already exists.
func (s *Sow) CreateProject(name, description string) (*Project, error) {
	if !s.IsInitialized() {
		return nil, ErrNotInitialized
	}

	// Validate project name is kebab-case
	if !isKebabCase(name) {
		return nil, fmt.Errorf("project name must be kebab-case (lowercase letters, digits, and single hyphens only)")
	}

	// Validate description is not empty
	if description == "" {
		return nil, fmt.Errorf("description cannot be empty")
	}

	// Get current git branch
	branch, err := s.getCurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Validate not on main/master
	if branch == "main" || branch == "master" {
		return nil, fmt.Errorf("cannot create project on protected branch '%s'", branch)
	}

	// Check if project already exists
	if s.HasProject() {
		// Load existing project to get its name
		existing, err := s.GetProject()
		if err == nil {
			existingName := existing.State().Project.Name
			return nil, fmt.Errorf("project '%s' already exists on branch '%s'", existingName, branch)
		}
		return nil, ErrProjectExists
	}

	// Create project directory structure
	if err := s.fs.MkdirAll(".sow/project", 0755); err != nil {
		return nil, fmt.Errorf("failed to create project directory: %w", err)
	}

	if err := s.fs.MkdirAll(".sow/project/context", 0755); err != nil {
		return nil, fmt.Errorf("failed to create context directory: %w", err)
	}

	// Create phase directories for required phases
	requiredPhases := []string{"implementation", "review", "finalize"}
	for _, phase := range requiredPhases {
		phaseDir := filepath.Join(".sow/project/phases", phase)
		if err := s.fs.MkdirAll(phaseDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create %s phase directory: %w", phase, err)
		}

		// Create log.md for each phase
		logPath := filepath.Join(phaseDir, "log.md")
		// Capitalize first letter of phase name
		phaseName := strings.ToUpper(phase[:1]) + phase[1:]
		logContent := []byte(fmt.Sprintf("# %s Phase Log\n\n", phaseName))
		if err := s.writeFile(logPath, logContent); err != nil {
			return nil, fmt.Errorf("failed to create %s log: %w", phase, err)
		}
	}

	// Create tasks directory for implementation
	tasksDir := filepath.Join(".sow/project/phases/implementation/tasks")
	if err := s.fs.MkdirAll(tasksDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create tasks directory: %w", err)
	}

	// Create review reports directory
	reportsDir := filepath.Join(".sow/project/phases/review/reports")
	if err := s.fs.MkdirAll(reportsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create reports directory: %w", err)
	}

	// Create project log
	logPath := filepath.Join(".sow/project", "log.md")
	logContent := []byte("# Project Log\n\n")
	if err := s.writeFile(logPath, logContent); err != nil {
		return nil, fmt.Errorf("failed to create project log: %w", err)
	}

	// Create state machine with initial project state
	machine, err := statechart.NewWithProject(name, description, branch, s.fs)
	if err != nil {
		return nil, fmt.Errorf("failed to create state machine: %w", err)
	}

	// Fire project init event to transition to DiscoveryDecision
	if err := machine.Fire(statechart.EventProjectInit); err != nil {
		return nil, fmt.Errorf("failed to initialize project state: %w", err)
	}

	// Save the initial state
	if err := machine.Save(); err != nil {
		return nil, fmt.Errorf("failed to save project state: %w", err)
	}

	return &Project{
		sow:     s,
		machine: machine,
	}, nil
}

// DeleteProject deletes the active project.
// Returns ErrNoProject if no project exists.
func (s *Sow) DeleteProject() error {
	if !s.HasProject() {
		return ErrNoProject
	}

	// Load the machine to update state before deletion
	machine, err := statechart.LoadFS(s.fs)
	if err != nil {
		return fmt.Errorf("failed to load project state: %w", err)
	}

	// Set project_deleted flag to true (required by state machine guard)
	state := machine.ProjectState()
	state.Phases.Finalize.Project_deleted = true

	// Save state with flag set
	if err := machine.Save(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	// Fire project delete event (guard will now pass)
	if err := machine.Fire(statechart.EventProjectDelete); err != nil {
		return fmt.Errorf("failed to transition state: %w", err)
	}

	// Remove the entire project directory
	if err := s.fs.Remove(".sow/project"); err != nil {
		// Try to remove all contents recursively
		if err := s.removeAll(".sow/project"); err != nil {
			return fmt.Errorf("failed to delete project: %w", err)
		}
	}

	return nil
}

// getCurrentBranch returns the current git branch name.
func (s *Sow) getCurrentBranch() (string, error) {
	// Read .git/HEAD
	headPath := ".git/HEAD"
	f, err := s.fs.Open(headPath)
	if err != nil {
		return "", fmt.Errorf("failed to open .git/HEAD: %w", err)
	}
	defer func() { _ = f.Close() }()

	var head [256]byte
	n, err := f.Read(head[:])
	if err != nil {
		return "", fmt.Errorf("failed to read .git/HEAD: %w", err)
	}

	headContent := string(head[:n])
	// HEAD typically contains: ref: refs/heads/branch-name
	const prefix = "ref: refs/heads/"
	if len(headContent) > len(prefix) && headContent[:len(prefix)] == prefix {
		branch := headContent[len(prefix):]
		// Trim newline
		if len(branch) > 0 && branch[len(branch)-1] == '\n' {
			branch = branch[:len(branch)-1]
		}
		return branch, nil
	}

	return "", fmt.Errorf("could not parse git branch from HEAD")
}

// writeFile writes data to a file atomically with 0644 permissions.
func (s *Sow) writeFile(path string, data []byte) error {
	f, err := s.fs.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	_, err = f.Write(data)
	return err
}

// readFile reads a file's contents.
func (s *Sow) readFile(path string) ([]byte, error) {
	f, err := s.fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var data []byte
	buf := make([]byte, 4096)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			data = append(data, buf[:n]...)
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
	}
	return data, nil
}

// removeAll recursively removes a directory and its contents.
func (s *Sow) removeAll(path string) error {
	entries, err := s.fs.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			if err := s.removeAll(fullPath); err != nil {
				return err
			}
		} else {
			if err := s.fs.Remove(fullPath); err != nil {
				return err
			}
		}
	}

	return s.fs.Remove(path)
}

// readYAML reads and unmarshals a YAML file.
func (s *Sow) readYAML(path string, v interface{}) error {
	data, err := s.readFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, v)
}

// writeYAML marshals and writes a YAML file atomically.
func (s *Sow) writeYAML(path string, v interface{}) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return err
	}

	// Write to temp file first
	tmpPath := path + ".tmp"
	if err := s.writeFile(tmpPath, data); err != nil {
		return err
	}

	// Atomic rename
	if err := s.fs.Rename(tmpPath, path); err != nil {
		_ = s.fs.Remove(tmpPath) // Clean up temp file
		return err
	}

	return nil
}

// readJSON reads and unmarshals a JSON file.
func (s *Sow) readJSON(path string, v interface{}) error {
	data, err := s.readFile(path)
	if err != nil {
		return err
	}
	// Use a simple JSON unmarshal approach
	return unmarshalJSON(data, v)
}

// writeJSON marshals and writes a JSON file atomically.
func (s *Sow) writeJSON(path string, v interface{}) error {
	data, err := marshalJSON(v)
	if err != nil {
		return err
	}

	// Write to temp file first
	tmpPath := path + ".tmp"
	if err := s.writeFile(tmpPath, data); err != nil {
		return err
	}

	// Atomic rename
	if err := s.fs.Rename(tmpPath, path); err != nil {
		_ = s.fs.Remove(tmpPath) // Clean up temp file
		return err
	}

	return nil
}

// marshalJSON marshals a value to JSON with indentation.
func marshalJSON(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

// unmarshalJSON unmarshals JSON data into a value.
func unmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// isKebabCase validates that a string is in kebab-case format.
// Kebab-case: lowercase letters, digits, and single hyphens only.
// Cannot start or end with hyphen, no consecutive hyphens.
func isKebabCase(s string) bool {
	if s == "" {
		return false
	}
	// Must be lowercase letters, digits, and hyphens only
	// Must not start or end with hyphen
	// No consecutive hyphens
	matched, _ := regexp.MatchString(`^[a-z0-9]+(-[a-z0-9]+)*$`, s)
	return matched
}
