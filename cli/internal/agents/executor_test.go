package agents

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
)

// TestExecutorInterface verifies interface compliance.
func TestExecutorInterface(_ *testing.T) {
	// Compile-time check that MockExecutor implements Executor.
	// This test documents the intent; the actual check is in executor_mock.go.
	var _ Executor = (*MockExecutor)(nil)
}

// TestCommandRunnerInterface verifies interface compliance.
func TestCommandRunnerInterface(_ *testing.T) {
	// Compile-time check that MockCommandRunner implements CommandRunner.
	var _ CommandRunner = (*MockCommandRunner)(nil)
	// Compile-time check that DefaultCommandRunner implements CommandRunner.
	var _ CommandRunner = (*DefaultCommandRunner)(nil)
}

// TestMockExecutor_Defaults tests default behavior when function fields are nil.
func TestMockExecutor_Defaults(t *testing.T) {
	mock := &MockExecutor{}

	t.Run("Name returns empty string by default", func(t *testing.T) {
		name := mock.Name()
		if name != "" {
			t.Errorf("Name() = %q, want empty string", name)
		}
	})

	t.Run("Spawn returns nil by default", func(t *testing.T) {
		err := mock.Spawn(context.Background(), Implementer, "prompt", "session-123")
		if err != nil {
			t.Errorf("Spawn() = %v, want nil", err)
		}
	})

	t.Run("Resume returns nil by default", func(t *testing.T) {
		err := mock.Resume(context.Background(), "session-123", "prompt")
		if err != nil {
			t.Errorf("Resume() = %v, want nil", err)
		}
	})

	t.Run("SupportsResumption returns false by default", func(t *testing.T) {
		supports := mock.SupportsResumption()
		if supports {
			t.Errorf("SupportsResumption() = true, want false")
		}
	})
}

// TestMockExecutor_FunctionFields tests that function fields are called when set.
func TestMockExecutor_FunctionFields(t *testing.T) {
	t.Run("NameFunc is called when set", func(t *testing.T) {
		mock := &MockExecutor{
			NameFunc: func() string { return "test-executor" },
		}
		name := mock.Name()
		if name != "test-executor" {
			t.Errorf("Name() = %q, want %q", name, "test-executor")
		}
	})

	t.Run("SpawnFunc is called when set", func(t *testing.T) {
		var capturedAgent *Agent
		var capturedPrompt string
		var capturedSessionID string
		expectedErr := errors.New("spawn error")

		mock := &MockExecutor{
			SpawnFunc: func(_ context.Context, agent *Agent, prompt string, sessionID string) error {
				capturedAgent = agent
				capturedPrompt = prompt
				capturedSessionID = sessionID
				return expectedErr
			},
		}

		err := mock.Spawn(context.Background(), Implementer, "test-prompt", "session-456")
		if !errors.Is(err, expectedErr) {
			t.Errorf("Spawn() = %v, want %v", err, expectedErr)
		}
		if capturedAgent != Implementer {
			t.Errorf("captured agent = %v, want %v", capturedAgent, Implementer)
		}
		if capturedPrompt != "test-prompt" {
			t.Errorf("captured prompt = %q, want %q", capturedPrompt, "test-prompt")
		}
		if capturedSessionID != "session-456" {
			t.Errorf("captured sessionID = %q, want %q", capturedSessionID, "session-456")
		}
	})

	t.Run("ResumeFunc is called when set", func(t *testing.T) {
		var capturedSessionID string
		var capturedPrompt string
		expectedErr := errors.New("resume error")

		mock := &MockExecutor{
			ResumeFunc: func(_ context.Context, sessionID string, prompt string) error {
				capturedSessionID = sessionID
				capturedPrompt = prompt
				return expectedErr
			},
		}

		err := mock.Resume(context.Background(), "session-789", "resume-prompt")
		if !errors.Is(err, expectedErr) {
			t.Errorf("Resume() = %v, want %v", err, expectedErr)
		}
		if capturedSessionID != "session-789" {
			t.Errorf("captured sessionID = %q, want %q", capturedSessionID, "session-789")
		}
		if capturedPrompt != "resume-prompt" {
			t.Errorf("captured prompt = %q, want %q", capturedPrompt, "resume-prompt")
		}
	})

	t.Run("SupportsResumptionFunc is called when set", func(t *testing.T) {
		mock := &MockExecutor{
			SupportsResumptionFunc: func() bool { return true },
		}
		supports := mock.SupportsResumption()
		if !supports {
			t.Errorf("SupportsResumption() = false, want true")
		}
	})
}

// TestMockCommandRunner_Defaults tests default behavior when function fields are nil.
func TestMockCommandRunner_Defaults(t *testing.T) {
	mock := &MockCommandRunner{}

	t.Run("Run returns nil by default", func(t *testing.T) {
		err := mock.Run(context.Background(), "test-cmd", []string{"arg1"}, nil, "")
		if err != nil {
			t.Errorf("Run() = %v, want nil", err)
		}
	})
}

// TestMockCommandRunner_FunctionFields tests that function fields are called when set.
func TestMockCommandRunner_FunctionFields(t *testing.T) {
	t.Run("RunFunc is called when set", func(t *testing.T) {
		var capturedName string
		var capturedArgs []string
		expectedErr := errors.New("run error")

		mock := &MockCommandRunner{
			RunFunc: func(_ context.Context, name string, args []string, _ io.Reader, _ string) error {
				capturedName = name
				capturedArgs = args
				return expectedErr
			},
		}

		err := mock.Run(context.Background(), "my-cmd", []string{"--flag", "value"}, nil, "")
		if !errors.Is(err, expectedErr) {
			t.Errorf("Run() = %v, want %v", err, expectedErr)
		}
		if capturedName != "my-cmd" {
			t.Errorf("captured name = %q, want %q", capturedName, "my-cmd")
		}
		if len(capturedArgs) != 2 || capturedArgs[0] != "--flag" || capturedArgs[1] != "value" {
			t.Errorf("captured args = %v, want [--flag value]", capturedArgs)
		}
	})

	t.Run("RunFunc receives stdin", func(t *testing.T) {
		var capturedStdin io.Reader

		mock := &MockCommandRunner{
			RunFunc: func(_ context.Context, _ string, _ []string, stdin io.Reader, _ string) error {
				capturedStdin = stdin
				return nil
			},
		}

		stdinContent := bytes.NewBufferString("input data")
		err := mock.Run(context.Background(), "cmd", nil, stdinContent, "")
		if err != nil {
			t.Errorf("Run() = %v, want nil", err)
		}
		if capturedStdin != stdinContent {
			t.Errorf("stdin was not passed correctly")
		}
	})
}

// TestMockCommandRunner_CapturesCall tests that call parameters are captured for verification.
func TestMockCommandRunner_CapturesCall(t *testing.T) {
	mock := &MockCommandRunner{}

	stdinContent := bytes.NewBufferString("test input")
	err := mock.Run(context.Background(), "captured-cmd", []string{"arg1", "arg2"}, stdinContent, "/path/to/output.log")
	if err != nil {
		t.Errorf("Run() = %v, want nil", err)
	}

	if mock.LastName != "captured-cmd" {
		t.Errorf("LastName = %q, want %q", mock.LastName, "captured-cmd")
	}
	if len(mock.LastArgs) != 2 || mock.LastArgs[0] != "arg1" || mock.LastArgs[1] != "arg2" {
		t.Errorf("LastArgs = %v, want [arg1 arg2]", mock.LastArgs)
	}
	if mock.LastStdin != "test input" {
		t.Errorf("LastStdin = %q, want %q", mock.LastStdin, "test input")
	}
	if mock.LastOutputPath != "/path/to/output.log" {
		t.Errorf("LastOutputPath = %q, want %q", mock.LastOutputPath, "/path/to/output.log")
	}
}

// TestMockCommandRunner_CapturesCallWithNilStdin tests stdin capture with nil input.
func TestMockCommandRunner_CapturesCallWithNilStdin(t *testing.T) {
	mock := &MockCommandRunner{}

	err := mock.Run(context.Background(), "cmd", nil, nil, "")
	if err != nil {
		t.Errorf("Run() = %v, want nil", err)
	}

	if mock.LastStdin != "" {
		t.Errorf("LastStdin = %q, want empty string", mock.LastStdin)
	}
}

// TestDefaultCommandRunner_StructExists verifies DefaultCommandRunner is usable.
func TestDefaultCommandRunner_StructExists(t *testing.T) {
	// This test verifies the struct exists and can be instantiated.
	// We don't actually run commands in unit tests.
	runner := &DefaultCommandRunner{}
	_ = runner // Use runner to avoid unused variable warning

	// The fact that we can instantiate it is sufficient for this test.
	// Compile-time interface check is in executor.go.
	t.Log("DefaultCommandRunner instantiated successfully")
}

// TestSplitOutputPaths tests the output path splitting logic.
func TestSplitOutputPaths(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		wantRaw          string
		wantFormatted    string
	}{
		{
			name:          ".log extension is replaced with .json for raw",
			input:         "/path/to/session-123.log",
			wantRaw:       "/path/to/session-123.json",
			wantFormatted: "/path/to/session-123.log",
		},
		{
			name:          "other extension gets both suffixes",
			input:         "/path/to/output.txt",
			wantRaw:       "/path/to/output.txt.json",
			wantFormatted: "/path/to/output.txt.log",
		},
		{
			name:          "no extension gets both suffixes",
			input:         "/path/to/output",
			wantRaw:       "/path/to/output.json",
			wantFormatted: "/path/to/output.log",
		},
		{
			name:          "simple filename with .log",
			input:         "session.log",
			wantRaw:       "session.json",
			wantFormatted: "session.log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRaw, gotFormatted := splitOutputPaths(tt.input)
			if gotRaw != tt.wantRaw {
				t.Errorf("splitOutputPaths(%q) raw = %q, want %q", tt.input, gotRaw, tt.wantRaw)
			}
			if gotFormatted != tt.wantFormatted {
				t.Errorf("splitOutputPaths(%q) formatted = %q, want %q", tt.input, gotFormatted, tt.wantFormatted)
			}
		})
	}
}
