package mocks

import (
	"context"
	"errors"
	"testing"

	"github.com/jmgilman/sow/libs/exec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExecutorMock_ImplementsExecutor verifies the mock implements the interface.
// This is a compile-time check - if it compiles, the mock is valid.
func TestExecutorMock_ImplementsExecutor(_ *testing.T) {
	var _ exec.Executor = &ExecutorMock{}
}

func TestExecutorMock_BasicUsage(t *testing.T) {
	t.Run("Command returns configured value", func(t *testing.T) {
		mock := &ExecutorMock{
			CommandFunc: func() string { return "gh" },
		}

		result := mock.Command()

		assert.Equal(t, "gh", result)
	})

	t.Run("Exists returns configured value", func(t *testing.T) {
		mock := &ExecutorMock{
			ExistsFunc: func() bool { return true },
		}

		result := mock.Exists()

		assert.True(t, result)
	})

	t.Run("Run returns configured values", func(t *testing.T) {
		mock := &ExecutorMock{
			RunFunc: func(_ ...string) (string, string, error) {
				return `{"number": 123}`, "", nil
			},
		}

		stdout, stderr, err := mock.Run("pr", "view")

		require.NoError(t, err)
		assert.Equal(t, `{"number": 123}`, stdout)
		assert.Empty(t, stderr)
	})

	t.Run("Run returns configured error", func(t *testing.T) {
		expectedErr := errors.New("command failed")
		mock := &ExecutorMock{
			RunFunc: func(_ ...string) (string, string, error) {
				return "", "error output", expectedErr
			},
		}

		stdout, stderr, err := mock.Run("bad", "command")

		assert.Equal(t, expectedErr, err)
		assert.Empty(t, stdout)
		assert.Equal(t, "error output", stderr)
	})

	t.Run("RunContext works with context", func(t *testing.T) {
		mock := &ExecutorMock{
			RunContextFunc: func(_ context.Context, _ ...string) (string, string, error) {
				return "output", "", nil
			},
		}

		stdout, stderr, err := mock.RunContext(context.Background(), "arg1")

		require.NoError(t, err)
		assert.Equal(t, "output", stdout)
		assert.Empty(t, stderr)
	})

	t.Run("RunSilent returns configured error", func(t *testing.T) {
		mock := &ExecutorMock{
			RunSilentFunc: func(_ ...string) error {
				return nil
			},
		}

		err := mock.RunSilent("arg1")

		assert.NoError(t, err)
	})

	t.Run("RunSilentContext works with context", func(t *testing.T) {
		mock := &ExecutorMock{
			RunSilentContextFunc: func(_ context.Context, _ ...string) error {
				return nil
			},
		}

		err := mock.RunSilentContext(context.Background(), "arg1")

		assert.NoError(t, err)
	})
}

func TestExecutorMock_PanicsWhenNotConfigured(t *testing.T) {
	t.Run("Command panics when not configured", func(t *testing.T) {
		mock := &ExecutorMock{}

		assert.Panics(t, func() {
			mock.Command()
		})
	})

	t.Run("Exists panics when not configured", func(t *testing.T) {
		mock := &ExecutorMock{}

		assert.Panics(t, func() {
			mock.Exists()
		})
	})

	t.Run("Run panics when not configured", func(t *testing.T) {
		mock := &ExecutorMock{}

		assert.Panics(t, func() {
			_, _, _ = mock.Run("arg")
		})
	})
}
