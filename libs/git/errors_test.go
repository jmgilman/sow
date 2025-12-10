package git

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrGHNotInstalled_Error(t *testing.T) {
	err := ErrGHNotInstalled{}

	got := err.Error()

	assert.Contains(t, got, "gh")
	assert.Contains(t, got, "not found")
}

func TestErrGHNotAuthenticated_Error(t *testing.T) {
	err := ErrGHNotAuthenticated{}

	got := err.Error()

	assert.Contains(t, got, "not authenticated")
}

func TestErrGHCommand_Error(t *testing.T) {
	t.Run("includes stderr when present", func(t *testing.T) {
		err := ErrGHCommand{
			Command: "issue list",
			Stderr:  "error: rate limit exceeded",
			Err:     errors.New("exit code 1"),
		}

		got := err.Error()

		assert.Contains(t, got, "issue list")
		assert.Contains(t, got, "rate limit exceeded")
	})

	t.Run("includes wrapped error when no stderr", func(t *testing.T) {
		err := ErrGHCommand{
			Command: "pr create",
			Stderr:  "",
			Err:     errors.New("network error"),
		}

		got := err.Error()

		assert.Contains(t, got, "pr create")
		assert.Contains(t, got, "network error")
	})
}

func TestErrGHCommand_Unwrap(t *testing.T) {
	wrappedErr := errors.New("underlying error")
	err := ErrGHCommand{
		Command: "test",
		Err:     wrappedErr,
	}

	got := err.Unwrap()

	assert.Same(t, wrappedErr, got)
}

func TestErrNotGitRepository_Error(t *testing.T) {
	err := ErrNotGitRepository{
		Path: "/some/path",
	}

	got := err.Error()

	assert.Contains(t, got, "/some/path")
	assert.Contains(t, got, "not a git repository")
}

func TestErrBranchExists_Error(t *testing.T) {
	err := ErrBranchExists{
		Branch: "feature/test",
	}

	got := err.Error()

	assert.Contains(t, got, "feature/test")
	assert.Contains(t, got, "already exists")
}

func TestErrorsAs(t *testing.T) {
	t.Run("ErrGHNotInstalled", func(t *testing.T) {
		err := ErrGHNotInstalled{}

		var target ErrGHNotInstalled
		assert.True(t, errors.As(err, &target))
	})

	t.Run("ErrGHNotAuthenticated", func(t *testing.T) {
		err := ErrGHNotAuthenticated{}

		var target ErrGHNotAuthenticated
		assert.True(t, errors.As(err, &target))
	})

	t.Run("ErrGHCommand", func(t *testing.T) {
		err := ErrGHCommand{Command: "test", Stderr: "error"}

		var target ErrGHCommand
		assert.True(t, errors.As(err, &target))
		assert.Equal(t, "test", target.Command)
		assert.Equal(t, "error", target.Stderr)
	})

	t.Run("ErrNotGitRepository", func(t *testing.T) {
		err := ErrNotGitRepository{Path: "/test"}

		var target ErrNotGitRepository
		assert.True(t, errors.As(err, &target))
		assert.Equal(t, "/test", target.Path)
	})

	t.Run("ErrBranchExists", func(t *testing.T) {
		err := ErrBranchExists{Branch: "main"}

		var target ErrBranchExists
		assert.True(t, errors.As(err, &target))
		assert.Equal(t, "main", target.Branch)
	})
}

func TestErrorChaining(t *testing.T) {
	t.Run("ErrGHCommand unwraps to underlying error", func(t *testing.T) {
		underlying := errors.New("network failure")
		err := ErrGHCommand{
			Command: "pr list",
			Err:     underlying,
		}

		assert.True(t, errors.Is(err, underlying))
	})
}
