package cmd

import (
	"context"

	"github.com/jmgilman/go/fs/core"
)

// Context keys for adapters.
type contextKey string

const (
	filesystemKey contextKey = "filesystem"
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
