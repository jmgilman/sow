// Package cmdutil provides shared utilities for CLI commands.
package cmdutil

import (
	"context"

	"github.com/jmgilman/sow/cli/internal/sow"
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
