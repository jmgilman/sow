// Package sowfs provides filesystem abstractions for the .sow directory structure.
//
// # Overview
//
// The sowfs package enforces the .sow directory structure and provides type-safe,
// validated access to all state files. It serves as the single source of truth for
// the directory layout, ensuring consistent access patterns across the CLI.
//
// # Architecture
//
// The package uses domain-specific interfaces to organize functionality:
//
//   - SowFS: Main entry point, provides access to all domains
//   - KnowledgeFS: Repository documentation (.sow/knowledge/)
//   - RefsFS: External references (.sow/refs/)
//   - ProjectFS: Active project state (.sow/project/)
//   - TaskFS: Individual task state (.sow/project/phases/implementation/tasks/{id}/)
//
// Each domain encapsulates all operations for its area, including:
//   - File I/O (reading/writing)
//   - Schema validation (via CUE)
//   - Type marshaling (YAML/JSON ↔ Go structs)
//
// # Directory Structure
//
// The .sow directory structure enforced by this package:
//
//	.sow/
//	├── knowledge/              # Repository-specific documentation (KnowledgeFS)
//	│   ├── overview.md
//	│   ├── architecture/
//	│   └── adrs/
//	├── refs/                   # External references (RefsFS)
//	│   ├── index.json          # Committed refs index
//	│   ├── index.local.json    # Local refs index (git-ignored)
//	│   └── {symlinks}          # Symlinks to cached repos
//	└── project/                # Active project (ProjectFS)
//	    ├── state.yaml          # Project state
//	    ├── log.md              # Orchestrator log
//	    ├── context/            # Project context files
//	    └── phases/
//	        └── implementation/
//	            └── tasks/
//	                └── {id}/   # Individual task (TaskFS)
//	                    ├── state.yaml
//	                    ├── description.md
//	                    ├── log.md
//	                    └── feedback/
//
// # Usage Example
//
//	// Create SowFS from current directory
//	sowFS, err := sowfs.NewSowFS()
//	if err != nil {
//	    return err
//	}
//	defer sowFS.Close()
//
//	// Access refs domain
//	refsFS := sowFS.Refs()
//	index, err := refsFS.CommittedIndex()
//
//	// Access project domain
//	projectFS, err := sowFS.Project()
//	if err != nil {
//	    return err // No active project
//	}
//	state, err := projectFS.State()
//
//	// Access task domain
//	taskFS, err := projectFS.Task("010")
//	taskState, err := taskFS.State()
//
// # Testing
//
// For testing, use NewSowFSWithFS to provide a memory filesystem:
//
//	import "github.com/jmgilman/go/fs/billy"
//
//	func TestMyFunction(t *testing.T) {
//	    // Create in-memory filesystem
//	    fs := billy.NewMemory()
//	    fs.MkdirAll(".sow/refs", 0755)
//	    fs.WriteFile(".sow/refs/index.json", testData, 0644)
//
//	    // Create SowFS with test filesystem
//	    sowFS, err := sowfs.NewSowFSWithFS(fs, "/test/repo")
//
//	    // Test operations...
//	}
//
// # Error Handling
//
// The package defines standard errors for common failure modes:
//
//   - ErrNotInGitRepo: Not in a git repository
//   - ErrSowNotInitialized: .sow directory not found
//   - ErrProjectNotFound: No active project
//   - ErrTaskNotFound: Task doesn't exist
//   - ErrInvalidTaskID: Invalid task ID format
//
// Use errors.Is() for error checking:
//
//	if errors.Is(err, sowfs.ErrProjectNotFound) {
//	    // Handle no project case
//	}
//
// # Design Principles
//
// 1. Single Source of Truth: All .sow structure knowledge in one package
// 2. Type Safety: CUE validation + typed structs
// 3. Encapsulation: Implementation details hidden from callers
// 4. Testability: Interface-based, works with memory filesystem
// 5. Discoverability: Domain-specific methods clearly organized
package sowfs
