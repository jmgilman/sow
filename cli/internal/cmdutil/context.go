// Package cmdutil provides shared utilities for CLI commands.
package cmdutil

import (
	"context"

	"github.com/jmgilman/go/fs/core"
	"github.com/jmgilman/sow/cli/internal/sow"
)

// Context keys for storing values in command context.
type contextKey string

const (
	sowKey        contextKey = "sow"
	filesystemKey contextKey = "filesystem"
)

// SowFromContext retrieves the Sow instance from the command context.
// Panics if not found (should always be available via root command setup).
func SowFromContext(ctx context.Context) *sow.Sow {
	s, ok := ctx.Value(sowKey).(*sow.Sow)
	if !ok {
		panic("sow instance not found in context")
	}
	return s
}

// WithSow adds a Sow instance to the context.
func WithSow(ctx context.Context, s *sow.Sow) context.Context {
	return context.WithValue(ctx, sowKey, s)
}

// WithFilesystem adds a filesystem to the context (for backwards compatibility).
func WithFilesystem(ctx context.Context, fs core.FS) context.Context {
	return context.WithValue(ctx, filesystemKey, fs)
}
