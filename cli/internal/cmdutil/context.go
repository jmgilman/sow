// Package cmdutil provides shared utilities for CLI commands.
package cmdutil

import (
	"context"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/libs/project/state"
)

// Context keys for storing values in command context.
type contextKey string

const (
	sowContextKey contextKey = "sowContext"
)

// GetContext retrieves the sow.Context from the command context.
// Panics if not found (should always be available via root command setup).
func GetContext(ctx context.Context) *sow.Context {
	c, ok := ctx.Value(sowContextKey).(*sow.Context)
	if !ok {
		panic("sow context not found in context")
	}
	return c
}

// RequireInitialized retrieves the sow.Context and returns an error if .sow doesn't exist.
// Use this in commands that require .sow to be initialized.
func RequireInitialized(ctx context.Context) (*sow.Context, error) {
	c := GetContext(ctx)
	if !c.IsInitialized() {
		return nil, sow.ErrNotInitialized
	}
	return c, nil
}

// WithContext adds a sow.Context to the command context.
func WithContext(ctx context.Context, sowCtx *sow.Context) context.Context {
	return context.WithValue(ctx, sowContextKey, sowCtx)
}

// LoadProject loads the project from the sow context's filesystem.
// This is a convenience wrapper around state.Load with the YAML backend.
func LoadProject(ctx context.Context, sowCtx *sow.Context) (*state.Project, error) {
	backend := state.NewYAMLBackend(sowCtx.FS())
	proj, err := state.Load(ctx, backend)
	if err != nil {
		return nil, fmt.Errorf("load project: %w", err)
	}
	return proj, nil
}

// SaveProject saves the project to the sow context's filesystem.
// This is a convenience wrapper around state.Save.
func SaveProject(ctx context.Context, proj *state.Project) error {
	if err := state.Save(ctx, proj); err != nil {
		return fmt.Errorf("save project: %w", err)
	}
	return nil
}

// CreateProject creates a new project and saves it.
// This is a convenience wrapper around state.Create with the YAML backend.
func CreateProject(ctx context.Context, sowCtx *sow.Context, opts state.CreateOpts) (*state.Project, error) {
	backend := state.NewYAMLBackend(sowCtx.FS())
	proj, err := state.Create(ctx, backend, opts)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	return proj, nil
}
