package agents

import (
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

// TestClaudeExecutor_Name verifies that Name() returns the correct identifier.
func TestClaudeExecutor_Name(t *testing.T) {
	executor := NewClaudeExecutor(false, "")
	if got := executor.Name(); got != "claude-code" {
		t.Errorf("Name() = %q, want %q", got, "claude-code")
	}
}

// TestClaudeExecutor_SupportsResumption verifies that Claude supports session resumption.
func TestClaudeExecutor_SupportsResumption(t *testing.T) {
	executor := NewClaudeExecutor(false, "")
	if !executor.SupportsResumption() {
		t.Error("SupportsResumption() = false, want true")
	}
}

// TestClaudeExecutor_Spawn_BuildsCorrectArgs verifies that Spawn builds correct CLI arguments
// for various configuration combinations.
func TestClaudeExecutor_Spawn_BuildsCorrectArgs(t *testing.T) {
	tests := []struct {
		name      string
		yoloMode  bool
		model     string
		sessionID string
		wantArgs  []string
	}{
		{
			name:      "minimal - no flags",
			yoloMode:  false,
			model:     "",
			sessionID: "",
			wantArgs:  []string{},
		},
		{
			name:      "yolo mode only",
			yoloMode:  true,
			model:     "",
			sessionID: "",
			wantArgs:  []string{"--dangerously-skip-permissions"},
		},
		{
			name:      "with model only",
			yoloMode:  false,
			model:     "sonnet",
			sessionID: "",
			wantArgs:  []string{"--model", "sonnet"},
		},
		{
			name:      "with session ID only",
			yoloMode:  false,
			model:     "",
			sessionID: "abc-123",
			wantArgs:  []string{"--session-id", "abc-123"},
		},
		{
			name:      "yolo mode and model",
			yoloMode:  true,
			model:     "opus",
			sessionID: "",
			wantArgs:  []string{"--dangerously-skip-permissions", "--model", "opus"},
		},
		{
			name:      "all flags combined",
			yoloMode:  true,
			model:     "opus",
			sessionID: "xyz-789",
			wantArgs:  []string{"--dangerously-skip-permissions", "--model", "opus", "--session-id", "xyz-789"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &MockCommandRunner{}
			executor := NewClaudeExecutorWithRunner(tt.yoloMode, tt.model, runner)

			err := executor.Spawn(context.Background(), Implementer, "test task prompt", tt.sessionID)
			if err != nil {
				t.Fatalf("Spawn() error = %v, want nil", err)
			}

			if runner.LastName != "claude" {
				t.Errorf("command = %q, want %q", runner.LastName, "claude")
			}

			if !reflect.DeepEqual(runner.LastArgs, tt.wantArgs) {
				t.Errorf("args = %v, want %v", runner.LastArgs, tt.wantArgs)
			}
		})
	}
}

// TestClaudeExecutor_Spawn_CombinesPrompts verifies that agent prompt and task prompt are combined.
func TestClaudeExecutor_Spawn_CombinesPrompts(t *testing.T) {
	runner := &MockCommandRunner{}
	executor := NewClaudeExecutorWithRunner(false, "", runner)

	taskPrompt := "Execute task 010 with specific instructions"
	err := executor.Spawn(context.Background(), Implementer, taskPrompt, "")
	if err != nil {
		t.Fatalf("Spawn() error = %v, want nil", err)
	}

	// Verify stdin contains task prompt
	if !strings.Contains(runner.LastStdin, taskPrompt) {
		t.Error("stdin should contain task prompt")
	}

	// Verify stdin contains agent prompt (from template)
	// Agent prompt is loaded from templates/implementer.md
	// Just verify the combined prompt is longer than just the task prompt
	if len(runner.LastStdin) <= len(taskPrompt)+10 {
		t.Error("stdin should contain agent prompt before task prompt")
	}

	// Verify prompts are separated by newlines
	if !strings.Contains(runner.LastStdin, "\n\n") {
		t.Error("agent prompt and task prompt should be separated by newlines")
	}
}

// TestClaudeExecutor_Spawn_LoadPromptError verifies that LoadPrompt errors are wrapped with context.
func TestClaudeExecutor_Spawn_LoadPromptError(t *testing.T) {
	runner := &MockCommandRunner{}
	executor := NewClaudeExecutorWithRunner(false, "", runner)

	// Create an agent with invalid prompt path
	badAgent := &Agent{
		Name:       "bad-agent",
		PromptPath: "nonexistent.md",
	}

	err := executor.Spawn(context.Background(), badAgent, "test", "")
	if err == nil {
		t.Fatal("Spawn() should return error for invalid prompt path")
	}

	// Verify error message contains context
	if !strings.Contains(err.Error(), "failed to load agent prompt") {
		t.Errorf("error = %q, should contain 'failed to load agent prompt'", err.Error())
	}
}

// TestClaudeExecutor_Spawn_RunnerError verifies that runner errors are returned as-is.
func TestClaudeExecutor_Spawn_RunnerError(t *testing.T) {
	expectedErr := errors.New("runner failed")
	runner := &MockCommandRunner{
		RunFunc: func(ctx context.Context, name string, args []string, stdin io.Reader) error {
			return expectedErr
		},
	}
	executor := NewClaudeExecutorWithRunner(false, "", runner)

	err := executor.Spawn(context.Background(), Implementer, "test", "")
	if err != expectedErr {
		t.Errorf("Spawn() error = %v, want %v", err, expectedErr)
	}
}

// TestClaudeExecutor_Resume_BuildsCorrectArgs verifies that Resume builds correct CLI arguments.
func TestClaudeExecutor_Resume_BuildsCorrectArgs(t *testing.T) {
	runner := &MockCommandRunner{}
	executor := NewClaudeExecutorWithRunner(false, "", runner)

	sessionID := "session-123"
	err := executor.Resume(context.Background(), sessionID, "feedback prompt")
	if err != nil {
		t.Fatalf("Resume() error = %v, want nil", err)
	}

	if runner.LastName != "claude" {
		t.Errorf("command = %q, want %q", runner.LastName, "claude")
	}

	wantArgs := []string{"--resume", sessionID}
	if !reflect.DeepEqual(runner.LastArgs, wantArgs) {
		t.Errorf("args = %v, want %v", runner.LastArgs, wantArgs)
	}
}

// TestClaudeExecutor_Resume_PassesPromptViaStdin verifies that Resume passes prompt via stdin.
func TestClaudeExecutor_Resume_PassesPromptViaStdin(t *testing.T) {
	runner := &MockCommandRunner{}
	executor := NewClaudeExecutorWithRunner(false, "", runner)

	prompt := "Fix the security issue on line 42"
	err := executor.Resume(context.Background(), "session-123", prompt)
	if err != nil {
		t.Fatalf("Resume() error = %v, want nil", err)
	}

	if runner.LastStdin != prompt {
		t.Errorf("stdin = %q, want %q", runner.LastStdin, prompt)
	}
}

// TestClaudeExecutor_Resume_RunnerError verifies that Resume returns runner errors.
func TestClaudeExecutor_Resume_RunnerError(t *testing.T) {
	expectedErr := errors.New("resume failed")
	runner := &MockCommandRunner{
		RunFunc: func(ctx context.Context, name string, args []string, stdin io.Reader) error {
			return expectedErr
		},
	}
	executor := NewClaudeExecutorWithRunner(false, "", runner)

	err := executor.Resume(context.Background(), "session-123", "feedback")
	if err != expectedErr {
		t.Errorf("Resume() error = %v, want %v", err, expectedErr)
	}
}

// TestNewClaudeExecutor_UsesDefaultRunner verifies that NewClaudeExecutor uses DefaultCommandRunner.
func TestNewClaudeExecutor_UsesDefaultRunner(t *testing.T) {
	executor := NewClaudeExecutor(true, "sonnet")

	// We can only verify the type by checking it's not nil
	// and that the configuration is set correctly
	if executor.yoloMode != true {
		t.Error("yoloMode should be true")
	}
	if executor.model != "sonnet" {
		t.Errorf("model = %q, want %q", executor.model, "sonnet")
	}
	if executor.runner == nil {
		t.Error("runner should not be nil")
	}

	// Verify runner is DefaultCommandRunner (by type assertion)
	if _, ok := executor.runner.(*DefaultCommandRunner); !ok {
		t.Error("runner should be *DefaultCommandRunner")
	}
}

// TestClaudeExecutor_InterfaceCompliance verifies compile-time interface compliance.
// This is a compile-time check - if ClaudeExecutor doesn't implement Executor,
// this file won't compile.
func TestClaudeExecutor_InterfaceCompliance(t *testing.T) {
	var _ Executor = (*ClaudeExecutor)(nil)
}
