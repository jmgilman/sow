package git

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		contains []string
	}{
		{
			name:     "ErrGHNotInstalled includes gh and not found",
			err:      ErrGHNotInstalled{},
			contains: []string{"gh", "not found"},
		},
		{
			name:     "ErrGHNotAuthenticated includes not authenticated",
			err:      ErrGHNotAuthenticated{},
			contains: []string{"not authenticated"},
		},
		{
			name: "ErrGHCommand includes command and stderr when present",
			err: ErrGHCommand{
				Command: "issue list",
				Stderr:  "error: rate limit exceeded",
				Err:     errors.New("exit code 1"),
			},
			contains: []string{"issue list", "rate limit exceeded"},
		},
		{
			name: "ErrGHCommand includes command and wrapped error when no stderr",
			err: ErrGHCommand{
				Command: "pr create",
				Stderr:  "",
				Err:     errors.New("network error"),
			},
			contains: []string{"pr create", "network error"},
		},
		{
			name:     "ErrNotGitRepository includes path and not a git repository",
			err:      ErrNotGitRepository{Path: "/some/path"},
			contains: []string{"/some/path", "not a git repository"},
		},
		{
			name:     "ErrBranchExists includes branch and already exists",
			err:      ErrBranchExists{Branch: "feature/test"},
			contains: []string{"feature/test", "already exists"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()

			for _, want := range tt.contains {
				assert.Contains(t, got, want)
			}
		})
	}
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

func TestErrorsAs(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		validate func(t *testing.T, err error)
	}{
		{
			name: "ErrGHNotInstalled",
			err:  ErrGHNotInstalled{},
			validate: func(t *testing.T, err error) {
				var target ErrGHNotInstalled
				assert.True(t, errors.As(err, &target))
			},
		},
		{
			name: "ErrGHNotAuthenticated",
			err:  ErrGHNotAuthenticated{},
			validate: func(t *testing.T, err error) {
				var target ErrGHNotAuthenticated
				assert.True(t, errors.As(err, &target))
			},
		},
		{
			name: "ErrGHCommand preserves fields",
			err:  ErrGHCommand{Command: "test", Stderr: "error"},
			validate: func(t *testing.T, err error) {
				var target ErrGHCommand
				assert.True(t, errors.As(err, &target))
				assert.Equal(t, "test", target.Command)
				assert.Equal(t, "error", target.Stderr)
			},
		},
		{
			name: "ErrNotGitRepository preserves path",
			err:  ErrNotGitRepository{Path: "/test"},
			validate: func(t *testing.T, err error) {
				var target ErrNotGitRepository
				assert.True(t, errors.As(err, &target))
				assert.Equal(t, "/test", target.Path)
			},
		},
		{
			name: "ErrBranchExists preserves branch",
			err:  ErrBranchExists{Branch: "main"},
			validate: func(t *testing.T, err error) {
				var target ErrBranchExists
				assert.True(t, errors.As(err, &target))
				assert.Equal(t, "main", target.Branch)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.err)
		})
	}
}

func TestErrGHCommand_ErrorChaining(t *testing.T) {
	underlying := errors.New("network failure")
	err := ErrGHCommand{
		Command: "pr list",
		Err:     underlying,
	}

	assert.True(t, errors.Is(err, underlying))
}
