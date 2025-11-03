// Package state provides the persistence layer for the Project SDK.
//
// # Architecture
//
// The state package implements a wrapper pattern where CUE-generated types
// (from cli/schemas/project/*.cue) are embedded into runtime types that add
// methods and behavior.
//
// Data flow:
//
//	YAML file → Unmarshal → CUE types → Validate → Convert → Wrapper types (runtime)
//	                                               ↓
//	YAML file ← Marshal ← CUE types ← Validate ← Convert ← Wrapper types
//
// # Core Types
//
// Project - Main orchestrator type with helper methods and embedded ProjectState.
// Provides PhaseOutputApproved(), PhaseMetadataBool(), AllTasksComplete() for guards.
//
// Phase - Pure data wrapper for phase state, embeds PhaseState from CUE.
//
// Artifact - Pure data wrapper for artifact state, embeds ArtifactState from CUE.
//
// Task - Pure data wrapper for task state, embeds TaskState from CUE.
//
// # Collections
//
// PhaseCollection - Map-based collection keyed by phase name (e.g., "planning", "implementation").
// Provides Get() for phase lookup.
//
// ArtifactCollection - Slice-based collection with Add/Remove operations.
// Supports indexed access via Get(index).
//
// TaskCollection - Slice-based collection with ID-based lookup.
// Provides Get(id), Add(task), Remove(id) operations.
//
// # Loading and Saving
//
// Load() reads project state from .sow/project/state.yaml:
//
//	ctx := context.Background()
//	project, err := state.Load(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// The Load pipeline:
//  1. Read YAML file from .sow/project/state.yaml
//  2. Unmarshal into CUE-generated ProjectState
//  3. Validate structure with CUE schema
//  4. Convert to wrapper types (Project with collections)
//  5. Lookup and attach ProjectTypeConfig from registry
//  6. Build state machine initialized with current state
//  7. Validate metadata against embedded schemas (stub for now)
//
// Save() writes project state atomically:
//
//	err := project.Save()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// The Save pipeline:
//  1. Sync statechart.current_state from machine (if present)
//  2. Update project.updated_at timestamp
//  3. Validate structure with CUE schema
//  4. Validate metadata against embedded schemas (stub for now)
//  5. Marshal to YAML
//  6. Atomic write (temp file + rename)
//
// Atomic writes ensure the state file is never left in a partially-written state.
// The pattern: write to state.yaml.tmp, then rename to state.yaml (atomic operation).
//
// # Validation
//
// Two-tier validation ensures correctness:
//
// 1. CUE structural validation - Universal fields required by all project types
// (name, type, branch, phases, statechart, etc.)
//
// 2. Metadata validation - Project-type-specific fields in phase.metadata
// (complexity, priority, approval flags, etc.)
//
// Validation timing:
// - Load(): CUE structural validation only
// - Save(): Both CUE structural + metadata validation
//
// This ensures that:
// - Projects can always be loaded from disk (even if metadata schema changed)
// - Invalid state cannot be saved to disk
// - State file always contains structurally valid data
//
// # Usage Example
//
//	// Load project
//	ctx := context.Background()
//	project, err := state.Load(ctx)
//	if err != nil {
//	    return err
//	}
//
//	// Navigate to phase
//	phase, err := PhaseCollection(project.Phases).Get("implementation")
//	if err != nil {
//	    return err
//	}
//
//	// Mutate phase state
//	phase.Status = "in_progress"
//
//	// Add artifact using collection
//	artifact := state.Artifact{
//	    ArtifactState: project.ArtifactState{
//	        Type:       "review",
//	        Path:       "review.md",
//	        Approved:   true,
//	        Created_at: time.Now(),
//	    },
//	}
//	collection := state.ArtifactCollection(phase.Outputs)
//	collection.Add(artifact)
//	phase.Outputs = collection
//
//	// Save atomically
//	if err := project.Save(); err != nil {
//	    return err
//	}
//
// # Helper Methods
//
// Project provides helper methods for common guard patterns used in
// state machine transitions (implemented in Unit 3):
//
// # PhaseOutputApproved(phaseName, outputType string) bool
//
// Checks if a phase has an approved artifact of the given type.
// Returns false if phase not found, artifact not found, or not approved.
// Example: project.PhaseOutputApproved("planning", "task_list")
//
// # PhaseMetadataBool(phaseName, key string) bool
//
// Reads a boolean value from phase metadata.
// Returns false if phase not found, key not found, or value is not a boolean.
// Example: project.PhaseMetadataBool("implementation", "tests_passing")
//
// # AllTasksComplete() bool
//
// Checks if all tasks across all phases have status "completed".
// Returns true if all tasks completed, or if no tasks exist (vacuous truth).
// Example: if project.AllTasksComplete() { ... }
//
// These methods are read-only and safe to call during guard evaluation.
//
// # Context Requirements
//
// Load() requires a context with "workingDir" key set to the repository root:
//
//	type contextKey string
//	const workingDirKey contextKey = "workingDir"
//
//	ctx := context.WithValue(context.Background(), workingDirKey, "/path/to/repo")
//	project, err := state.Load(ctx)
//
// The state file is expected at: {workingDir}/.sow/project/state.yaml
//
// # Error Handling
//
// All errors are actionable and indicate which step failed:
//
//   - "failed to read state": File not found or permission denied
//   - "failed to unmarshal": Invalid YAML syntax
//   - "CUE validation failed": Structure doesn't match schema
//   - "unknown project type": Type not registered in Registry
//   - "metadata validation failed": Metadata doesn't match embedded schema
//
// Validation errors prevent Save() from writing to disk, ensuring the state
// file always contains valid data.
//
// # Registry
//
// Project types are registered in the global Registry:
//
//	var Registry = make(map[string]*ProjectTypeConfig)
//
// ProjectTypeConfig holds project-type-specific configuration including
// state machine definitions, guards, and metadata schemas. This is populated
// during initialization (full implementation in Unit 3).
//
// # Design Document
//
// Full architecture details, implementation decisions, and rationale:
// .sow/knowledge/designs/project-sdk-implementation.md
//
// Key sections:
//   - "Architecture Overview" - High-level design
//   - "Data Flow: CLI Command Lifecycle" - Real-world usage patterns
//   - "Wrapper Types" - Why we wrap CUE types
//   - "Two-tier Validation" - Validation strategy
//   - "Atomic Writes" - How we prevent corruption
//
// # Testing
//
// The package includes comprehensive unit and integration tests:
//
// Unit tests (*_test.go):
//   - collections_test.go - Collection operations
//   - convert_test.go - Type conversion
//   - validate_test.go - Validation logic
//   - loader_test.go - Load/Save operations
//   - project_test.go - Helper methods
//
// Integration tests (integration_test.go):
//   - TestIntegration_PersistenceWorkflow - Full Load→Mutate→Save→Load cycle
//   - TestIntegration_ComplexNestedStructure - Nested phases/tasks/artifacts
//   - TestIntegration_ValidationPreventsInvalidState - Validation enforced
//   - TestIntegration_AtomicWriteProtection - Atomic write behavior
//   - TestIntegration_MetadataValidation - Metadata schemas enforced
//
// Run tests with:
//
//	go test ./cli/internal/sdks/project/state/...
//
// # Future Work
//
// This package provides the foundation for the Project SDK. Future units will add:
//
//   - Unit 3: Full state machine implementation with guards and actions
//   - Unit 3: Complete metadata validation with embedded CUE schemas
//   - Unit 3: Project type configurations (standard, exploration, etc.)
//   - Unit 4: CLI commands that use this persistence layer
package state
