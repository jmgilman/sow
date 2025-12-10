package exec

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalExecutor_Command(t *testing.T) {
	e := NewLocalExecutor("echo")
	assert.Equal(t, "echo", e.Command())
}

func TestLocalExecutor_Exists(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    bool
	}{
		{name: "echo exists", command: "echo", want: true},
		{name: "true exists", command: "true", want: true},
		{name: "sh exists", command: "sh", want: true},
		{name: "nonexistent command", command: "definitely-not-a-command-12345", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewLocalExecutor(tt.command)
			got := e.Exists()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLocalExecutor_Run(t *testing.T) {
	t.Run("captures stdout", func(t *testing.T) {
		e := NewLocalExecutor("echo")
		stdout, stderr, err := e.Run("hello")

		require.NoError(t, err)
		assert.Equal(t, "hello\n", stdout)
		assert.Empty(t, stderr)
	})

	t.Run("captures stderr", func(t *testing.T) {
		e := NewLocalExecutor("sh")
		stdout, stderr, err := e.Run("-c", "echo error >&2")

		require.NoError(t, err)
		assert.Empty(t, stdout)
		assert.Equal(t, "error\n", stderr)
	})

	t.Run("returns error for failed command", func(t *testing.T) {
		e := NewLocalExecutor("false")
		_, _, err := e.Run()

		assert.Error(t, err)
	})

	t.Run("returns both stdout and stderr", func(t *testing.T) {
		e := NewLocalExecutor("sh")
		stdout, stderr, err := e.Run("-c", "echo out; echo err >&2")

		require.NoError(t, err)
		assert.Equal(t, "out\n", stdout)
		assert.Equal(t, "err\n", stderr)
	})

	t.Run("handles empty args", func(t *testing.T) {
		e := NewLocalExecutor("true")
		_, _, err := e.Run()

		require.NoError(t, err)
	})

	t.Run("handles args with spaces", func(t *testing.T) {
		e := NewLocalExecutor("echo")
		stdout, _, err := e.Run("hello world")

		require.NoError(t, err)
		assert.Equal(t, "hello world\n", stdout)
	})
}

func TestLocalExecutor_RunContext(t *testing.T) {
	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		e := NewLocalExecutor("sleep")
		_, _, err := e.RunContext(ctx, "10")

		assert.Error(t, err)
	})

	t.Run("respects context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		e := NewLocalExecutor("sleep")
		_, _, err := e.RunContext(ctx, "10")

		assert.Error(t, err)
	})

	t.Run("completes before timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		e := NewLocalExecutor("echo")
		stdout, _, err := e.RunContext(ctx, "fast")

		require.NoError(t, err)
		assert.Equal(t, "fast\n", stdout)
	})
}

func TestLocalExecutor_RunSilent(t *testing.T) {
	t.Run("returns nil on success", func(t *testing.T) {
		e := NewLocalExecutor("true")
		err := e.RunSilent()

		assert.NoError(t, err)
	})

	t.Run("returns error on failure", func(t *testing.T) {
		e := NewLocalExecutor("false")
		err := e.RunSilent()

		assert.Error(t, err)
	})

	t.Run("discards output", func(t *testing.T) {
		e := NewLocalExecutor("echo")
		err := e.RunSilent("hello")

		// If output was captured and returned, this test wouldn't fail
		// We just verify the method works - output is discarded
		assert.NoError(t, err)
	})
}

func TestLocalExecutor_RunSilentContext(t *testing.T) {
	t.Run("returns nil on success", func(t *testing.T) {
		e := NewLocalExecutor("true")
		err := e.RunSilentContext(context.Background())

		assert.NoError(t, err)
	})

	t.Run("returns error on failure", func(t *testing.T) {
		e := NewLocalExecutor("false")
		err := e.RunSilentContext(context.Background())

		assert.Error(t, err)
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		e := NewLocalExecutor("sleep")
		err := e.RunSilentContext(ctx, "10")

		assert.Error(t, err)
	})
}

func TestLocalExecutor_MustExist(t *testing.T) {
	t.Run("does not panic for existing command", func(t *testing.T) {
		e := NewLocalExecutor("echo")

		// Should not panic
		assert.NotPanics(t, func() {
			e.MustExist()
		})
	})

	t.Run("panics for nonexistent command", func(t *testing.T) {
		e := NewLocalExecutor("definitely-not-a-command-12345")

		assert.Panics(t, func() {
			e.MustExist()
		})
	})
}

// TestLocalExecutor_ImplementsExecutor verifies compile-time interface compliance.
// The actual check is in local.go via: var _ Executor = (*LocalExecutor)(nil).
func TestLocalExecutor_ImplementsExecutor(_ *testing.T) {
	var _ Executor = (*LocalExecutor)(nil)
}
