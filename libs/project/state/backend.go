package state

import (
	"context"

	"github.com/jmgilman/sow/libs/schemas/project"
)

// Backend defines the interface for project state persistence.
// Implementations handle reading and writing project state to various
// storage systems (files, databases, remote APIs, etc.)
type Backend interface {
	// Load reads project state from storage.
	// Returns the raw ProjectState (CUE-generated type).
	// The caller is responsible for wrapping this in a Project with runtime fields.
	Load(ctx context.Context) (*project.ProjectState, error)

	// Save writes project state to storage.
	// Takes the raw ProjectState (CUE-generated type).
	// Implementation should handle atomic writes where possible.
	Save(ctx context.Context, state *project.ProjectState) error

	// Exists checks if a project exists in storage.
	Exists(ctx context.Context) (bool, error)

	// Delete removes project state from storage.
	Delete(ctx context.Context) error
}
