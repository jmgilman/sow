package cmd

import (
	"context"

	"github.com/jmgilman/go/fs/core"
	"github.com/jmgilman/sow/cli/internal/sow"
)

// WithFilesystem adds filesystem adapter to context.
func WithFilesystem(ctx context.Context, fs core.FS) context.Context {
	return context.WithValue(ctx, "filesystem", fs)
}

// FilesystemFromContext retrieves filesystem adapter from context.
func FilesystemFromContext(ctx context.Context) core.FS {
	fs, ok := ctx.Value("filesystem").(core.FS)
	if !ok {
		panic("filesystem not found in context")
	}
	return fs
}

// WithSow adds the unified Sow instance to context.
func WithSow(ctx context.Context, s *sow.Sow) context.Context {
	return context.WithValue(ctx, "sow", s)
}

// SowFromContext retrieves the unified Sow instance from context.
// Panics if Sow is not found in context (should always be available).
func SowFromContext(ctx context.Context) *sow.Sow {
	s, ok := ctx.Value("sow").(*sow.Sow)
	if !ok {
		panic("sow instance not found in context")
	}
	return s
}
