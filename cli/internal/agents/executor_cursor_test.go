package agents

import (
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

// TestCursorExecutor_InterfaceCompliance verifies CursorExecutor implements Executor.
func TestCursorExecutor_InterfaceCompliance(_ *testing.T) {
	// Compile-time check that CursorExecutor implements Executor.
	var _ Executor = (*CursorExecutor)(nil)
}

// TestCursorExecutor_Name verifies Name() returns "cursor".
func TestCursorExecutor_Name(t *testing.T) {
	executor := NewCursorExecutor(false)
	name := executor.Name()
	if name != "cursor" {
		t.Errorf("Name() = %q, want %q", name, "cursor")
	}
}

// TestCursorExecutor_SupportsResumption verifies SupportsResumption() returns true.
func TestCursorExecutor_SupportsResumption(t *testing.T) {
	executor := NewCursorExecutor(false)
	supports := executor.SupportsResumption()
	if !supports {
		t.Errorf("SupportsResumption() = false, want true")
	}
}

// TestCursorExecutor_Spawn_BuildsCorrectArgs verifies Spawn builds correct CLI arguments.
func TestCursorExecutor_Spawn_BuildsCorrectArgs(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		wantArgs  []string
	}{
		{
			name:      "no session ID",
			sessionID: "",
			wantArgs:  []string{"agent"},
		},
		{
			name:      "with session ID",
			sessionID: "abc-123",
			wantArgs:  []string{"agent", "--chat-id", "abc-123"},
		},
		{
			name:      "with UUID session ID",
			sessionID: "550e8400-e29b-41d4-a716-446655440000",
			wantArgs:  []string{"agent", "--chat-id", "550e8400-e29b-41d4-a716-446655440000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &MockCommandRunner{}

			executor := NewCursorExecutorWithRunner(false, runner)
			err := executor.Spawn(context.Background(), Implementer, "test prompt", tt.sessionID)
			if err != nil {
				t.Fatalf("Spawn() unexpected error: %v", err)
			}

			if runner.LastName != "cursor-agent" {
				t.Errorf("command = %q, want %q", runner.LastName, "cursor-agent")
			}

			if !reflect.DeepEqual(runner.LastArgs, tt.wantArgs) {
				t.Errorf("args = %v, want %v", runner.LastArgs, tt.wantArgs)
			}
		})
	}
}

// TestCursorExecutor_Spawn_CombinesPrompts verifies Spawn combines agent and task prompts.
func TestCursorExecutor_Spawn_CombinesPrompts(t *testing.T) {
	runner := &MockCommandRunner{}

	executor := NewCursorExecutorWithRunner(false, runner)
	err := executor.Spawn(context.Background(), Implementer, "Execute task 010", "")
	if err != nil {
		t.Fatalf("Spawn() unexpected error: %v", err)
	}

	// Should contain task prompt
	if !strings.Contains(runner.LastStdin, "Execute task 010") {
		t.Error("stdin should contain task prompt")
	}

	// Should contain something from the agent prompt (loaded via LoadPrompt)
	// The implementer.md template exists and has content
	if len(runner.LastStdin) <= len("Execute task 010") {
		t.Error("stdin should contain agent prompt in addition to task prompt")
	}

	// Verify the prompts are combined with newlines
	if !strings.Contains(runner.LastStdin, "\n\n") {
		t.Error("agent prompt and task prompt should be separated by double newline")
	}
}

// TestCursorExecutor_Spawn_LoadPromptError verifies Spawn returns wrapped error on LoadPrompt failure.
func TestCursorExecutor_Spawn_LoadPromptError(t *testing.T) {
	runner := &MockCommandRunner{}

	// Create an agent with a non-existent prompt path
	badAgent := &Agent{
		Name:         "bad-agent",
		Description:  "Agent with missing prompt",
		Capabilities: "None",
		PromptPath:   "nonexistent-prompt.md",
	}

	executor := NewCursorExecutorWithRunner(false, runner)
	err := executor.Spawn(context.Background(), badAgent, "test prompt", "")

	if err == nil {
		t.Fatal("Spawn() expected error for missing prompt, got nil")
	}

	if !strings.Contains(err.Error(), "failed to load agent prompt") {
		t.Errorf("error should contain 'failed to load agent prompt', got: %v", err)
	}

	if !strings.Contains(err.Error(), "nonexistent-prompt.md") {
		t.Errorf("error should contain prompt path, got: %v", err)
	}
}

// TestCursorExecutor_Spawn_RunnerError verifies Spawn returns wrapped runner errors.
func TestCursorExecutor_Spawn_RunnerError(t *testing.T) {
	expectedErr := errors.New("runner failed")
	runner := &MockCommandRunner{
		RunFunc: func(_ context.Context, _ string, _ []string, _ io.Reader) error {
			return expectedErr
		},
	}

	executor := NewCursorExecutorWithRunner(false, runner)
	err := executor.Spawn(context.Background(), Implementer, "test prompt", "")

	if !errors.Is(err, expectedErr) {
		t.Errorf("Spawn() = %v, want wrapped %v", err, expectedErr)
	}
	if !strings.Contains(err.Error(), "cursor-agent spawn failed") {
		t.Errorf("error should contain context: %v", err)
	}
}

// TestCursorExecutor_Resume_BuildsCorrectArgs verifies Resume builds correct CLI arguments.
func TestCursorExecutor_Resume_BuildsCorrectArgs(t *testing.T) {
	runner := &MockCommandRunner{}

	executor := NewCursorExecutorWithRunner(false, runner)
	err := executor.Resume(context.Background(), "session-123", "resume feedback")
	if err != nil {
		t.Fatalf("Resume() unexpected error: %v", err)
	}

	if runner.LastName != "cursor-agent" {
		t.Errorf("command = %q, want %q", runner.LastName, "cursor-agent")
	}

	wantArgs := []string{"agent", "--resume", "session-123"}
	if !reflect.DeepEqual(runner.LastArgs, wantArgs) {
		t.Errorf("args = %v, want %v", runner.LastArgs, wantArgs)
	}
}

// TestCursorExecutor_Resume_PassesPromptViaStdin verifies Resume passes prompt to stdin.
func TestCursorExecutor_Resume_PassesPromptViaStdin(t *testing.T) {
	runner := &MockCommandRunner{}

	executor := NewCursorExecutorWithRunner(false, runner)
	err := executor.Resume(context.Background(), "session-123", "feedback prompt text")
	if err != nil {
		t.Fatalf("Resume() unexpected error: %v", err)
	}

	if runner.LastStdin != "feedback prompt text" {
		t.Errorf("stdin = %q, want %q", runner.LastStdin, "feedback prompt text")
	}
}

// TestCursorExecutor_Resume_RunnerError verifies Resume returns wrapped runner errors.
func TestCursorExecutor_Resume_RunnerError(t *testing.T) {
	expectedErr := errors.New("resume runner failed")
	runner := &MockCommandRunner{
		RunFunc: func(_ context.Context, _ string, _ []string, _ io.Reader) error {
			return expectedErr
		},
	}

	executor := NewCursorExecutorWithRunner(false, runner)
	err := executor.Resume(context.Background(), "session-123", "feedback")

	if !errors.Is(err, expectedErr) {
		t.Errorf("Resume() = %v, want wrapped %v", err, expectedErr)
	}
	if !strings.Contains(err.Error(), "cursor-agent resume failed") {
		t.Errorf("error should contain context: %v", err)
	}
}

// TestNewCursorExecutor_UsesDefaultRunner verifies NewCursorExecutor uses DefaultCommandRunner.
func TestNewCursorExecutor_UsesDefaultRunner(t *testing.T) {
	executor := NewCursorExecutor(false)

	// We can't directly check the type of runner since it's private,
	// but we can verify the executor was created successfully.
	if executor == nil {
		t.Fatal("NewCursorExecutor() returned nil")
	}

	// Verify it implements Executor by checking Name works
	if executor.Name() != "cursor" {
		t.Errorf("Name() = %q, want %q", executor.Name(), "cursor")
	}
}

// TestNewCursorExecutorWithRunner_AcceptsCustomRunner verifies custom runner injection.
func TestNewCursorExecutorWithRunner_AcceptsCustomRunner(t *testing.T) {
	customRunner := &MockCommandRunner{}

	executor := NewCursorExecutorWithRunner(true, customRunner)

	if executor == nil {
		t.Fatal("NewCursorExecutorWithRunner() returned nil")
	}

	// Verify the custom runner is used by calling Spawn
	err := executor.Spawn(context.Background(), Implementer, "test", "")
	if err != nil {
		t.Fatalf("Spawn() unexpected error: %v", err)
	}

	// The mock runner should have been called
	if customRunner.LastName != "cursor-agent" {
		t.Errorf("custom runner not used, LastName = %q, want %q", customRunner.LastName, "cursor-agent")
	}
}

// TestCursorExecutor_YoloModeStored verifies yoloMode is stored (for future use).
func TestCursorExecutor_YoloModeStored(t *testing.T) {
	// Create with yoloMode true
	executor := NewCursorExecutorWithRunner(true, &MockCommandRunner{})
	if executor == nil {
		t.Fatal("NewCursorExecutorWithRunner() returned nil")
	}

	// Create with yoloMode false
	executorNoYolo := NewCursorExecutorWithRunner(false, &MockCommandRunner{})
	if executorNoYolo == nil {
		t.Fatal("NewCursorExecutorWithRunner() returned nil for yoloMode=false")
	}

	// Both should work - yoloMode is stored but not currently used
	// This test just verifies the constructors accept the flag
}
