// ============================================================================
// WARNING: This file has a corresponding CUE schema file (common.cue).
// When modifying these Go types, you MUST manually update common.cue
// to keep the schemas in sync. Do not rely on code generation.
// ============================================================================

// Package phases provides hand-written Go types for phase schemas.
// These types must be kept in sync with the CUE schemas in common.cue.
package phases

import "time"

// GenericStatus defines the common phase status values that most project types use.
// Project types can use this for their phases or define custom status enums per phase.
type GenericStatus string

// Common phase status constants.
const (
	StatusPending    GenericStatus = "pending"
	StatusInProgress GenericStatus = "in_progress"
	StatusCompleted  GenericStatus = "completed"
	StatusSkipped    GenericStatus = "skipped"
)

// Phase is the universal schema for all phases in all project types.
// What makes a phase unique is its guards, prompts, and which operations it supports.
// NOTE: Project types will typically define their own phase types with additional fields.
type Phase struct {
	// Common metadata
	Status  string `json:"status"`
	Enabled bool   `json:"enabled"`

	// Timestamps
	Created_at   time.Time  `json:"created_at"`
	Started_at   *time.Time `json:"started_at,omitempty"`
	Completed_at *time.Time `json:"completed_at,omitempty"`

	// Generic collections (used by phases that need them)
	Artifacts []Artifact `json:"artifacts"` // Used by discovery, design, review
	Tasks     []Task     `json:"tasks"`     // Used by implementation

	// Escape hatch for unanticipated fields
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Artifact represents a phase artifact requiring human approval.
type Artifact struct {
	// Path relative to .sow/project/
	Path string `json:"path"`

	// Human approval status
	Approved bool `json:"approved"`

	// When artifact was created
	Created_at time.Time `json:"created_at"`

	// Artifact type (e.g., "task_list", "review", "documentation")
	Type *string `json:"type,omitempty"`

	// Review assessment result ("pass" or "fail")
	Assessment *string `json:"assessment,omitempty"`

	// Escape hatch for unanticipated fields
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Task represents an implementation task.
type Task struct {
	// Gap-numbered ID (010, 020, 030...)
	// nolint:revive // Id is intentional to match JSON field name
	Id string `json:"id"`

	// Task name
	Name string `json:"name"`

	// Task status
	Status string `json:"status"` // "pending" | "in_progress" | "completed" | "abandoned"

	// Can run in parallel with other tasks
	Parallel bool `json:"parallel"`

	// Task IDs this task depends on
	Dependencies []string `json:"dependencies,omitempty"`
}
