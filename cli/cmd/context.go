package cmd

import (
	"context"

	"github.com/jmgilman/go/fs/core"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/internal/sowfs"
)

// Context keys for adapters.
type contextKey string

const (
	filesystemKey contextKey = "filesystem"
	sowfsKey      contextKey = "sowfs"
	sowKey        contextKey = "sow"
)

// WithFilesystem adds filesystem adapter to context.
func WithFilesystem(ctx context.Context, fs core.FS) context.Context {
	return context.WithValue(ctx, filesystemKey, fs)
}

// FilesystemFromContext retrieves filesystem adapter from context.
func FilesystemFromContext(ctx context.Context) core.FS {
	fs, ok := ctx.Value(filesystemKey).(core.FS)
	if !ok {
		panic("filesystem not found in context")
	}
	return fs
}

// WithSowFS adds SowFS to context.
func WithSowFS(ctx context.Context, sfs sowfs.SowFS) context.Context {
	return context.WithValue(ctx, sowfsKey, sfs)
}

// SowFSFromContext retrieves SowFS from context.
// Returns nil if SowFS is not available (e.g., not in a .sow directory).
func SowFSFromContext(ctx context.Context) sowfs.SowFS {
	sfs, _ := ctx.Value(sowfsKey).(sowfs.SowFS)
	return sfs
}

// WithSow adds the unified Sow instance to context.
func WithSow(ctx context.Context, s *sow.Sow) context.Context {
	return context.WithValue(ctx, sowKey, s)
}

// SowFromContext retrieves the unified Sow instance from context.
// Panics if Sow is not found in context (should always be available).
func SowFromContext(ctx context.Context) *sow.Sow {
	s, ok := ctx.Value(sowKey).(*sow.Sow)
	if !ok {
		panic("sow instance not found in context")
	}
	return s
}
