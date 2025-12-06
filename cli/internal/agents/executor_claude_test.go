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
	executor := NewClaudeExecutor(false, "", "", nil)
	if got := executor.Name(); got != "claude-code" {
		t.Errorf("Name() = %q, want %q", got, "claude-code")
	}
}

// TestClaudeExecutor_SupportsResumption verifies that Claude supports session resumption.
func TestClaudeExecutor_SupportsResumption(t *testing.T) {
	executor := NewClaudeExecutor(false, "", "", nil)
	if !executor.SupportsResumption() {
		t.Error("SupportsResumption() = false, want true")
	}
}

// TestClaudeExecutor_Spawn_BuildsCorrectArgs verifies that Spawn builds correct CLI arguments
// for various configuration combinations.
// Note: --print, --verbose, --output-format stream-json, and --permission-mode acceptEdits are always included.
func TestClaudeExecutor_Spawn_BuildsCorrectArgs(t *testing.T) {
	tests := []struct {
		name      string
		yoloMode  bool
		model     string
		sessionID string
		wantArgs  []string
	}{
		{
			name:      "minimal - base flags only",
			yoloMode:  false,
			model:     "",
			sessionID: "",
			wantArgs:  []string{"--print", "--verbose", "--output-format", "stream-json", "--permission-mode", "acceptEdits"},
		},
		{
			name:      "yolo mode only",
			yoloMode:  true,
			model:     "",
			sessionID: "",
			wantArgs:  []string{"--print", "--verbose", "--output-format", "stream-json", "--permission-mode", "acceptEdits", "--dangerously-skip-permissions"},
		},
		{
			name:      "with model only",
			yoloMode:  false,
			model:     "sonnet",
			sessionID: "",
			wantArgs:  []string{"--print", "--verbose", "--output-format", "stream-json", "--permission-mode", "acceptEdits", "--model", "sonnet"},
		},
		{
			name:      "with session ID only",
			yoloMode:  false,
			model:     "",
			sessionID: "abc-123",
			wantArgs:  []string{"--print", "--verbose", "--output-format", "stream-json", "--permission-mode", "acceptEdits", "--session-id", "abc-123"},
		},
		{
			name:      "yolo mode and model",
			yoloMode:  true,
			model:     "opus",
			sessionID: "",
			wantArgs:  []string{"--print", "--verbose", "--output-format", "stream-json", "--permission-mode", "acceptEdits", "--dangerously-skip-permissions", "--model", "opus"},
		},
		{
			name:      "all flags combined",
			yoloMode:  true,
			model:     "opus",
			sessionID: "xyz-789",
			wantArgs:  []string{"--print", "--verbose", "--output-format", "stream-json", "--permission-mode", "acceptEdits", "--dangerously-skip-permissions", "--model", "opus", "--session-id", "xyz-789"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &MockCommandRunner{}
			executor := NewClaudeExecutorWithRunner(tt.yoloMode, tt.model, "", nil, runner)

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
	executor := NewClaudeExecutorWithRunner(false, "", "", nil, runner)

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
	executor := NewClaudeExecutorWithRunner(false, "", "", nil, runner)

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

// TestClaudeExecutor_Spawn_RunnerError verifies that runner errors are wrapped with context.
func TestClaudeExecutor_Spawn_RunnerError(t *testing.T) {
	expectedErr := errors.New("runner failed")
	runner := &MockCommandRunner{
		RunFunc: func(_ context.Context, _ string, _ []string, _ io.Reader, _ string) error {
			return expectedErr
		},
	}
	executor := NewClaudeExecutorWithRunner(false, "", "", nil, runner)

	err := executor.Spawn(context.Background(), Implementer, "test", "")
	if !errors.Is(err, expectedErr) {
		t.Errorf("Spawn() error = %v, want wrapped %v", err, expectedErr)
	}
	if !strings.Contains(err.Error(), "claude spawn failed") {
		t.Errorf("error should contain context: %v", err)
	}
}

// TestClaudeExecutor_Resume_BuildsCorrectArgs verifies that Resume builds correct CLI arguments.
// Note: --print, --verbose, --output-format stream-json, and --permission-mode acceptEdits are always included.
func TestClaudeExecutor_Resume_BuildsCorrectArgs(t *testing.T) {
	runner := &MockCommandRunner{}
	executor := NewClaudeExecutorWithRunner(false, "", "", nil, runner)

	sessionID := "session-123"
	err := executor.Resume(context.Background(), sessionID, "feedback prompt")
	if err != nil {
		t.Fatalf("Resume() error = %v, want nil", err)
	}

	if runner.LastName != "claude" {
		t.Errorf("command = %q, want %q", runner.LastName, "claude")
	}

	wantArgs := []string{"--print", "--verbose", "--output-format", "stream-json", "--permission-mode", "acceptEdits", "--resume", sessionID}
	if !reflect.DeepEqual(runner.LastArgs, wantArgs) {
		t.Errorf("args = %v, want %v", runner.LastArgs, wantArgs)
	}
}

// TestClaudeExecutor_Resume_WithYoloMode verifies Resume includes yolo mode flag when enabled.
func TestClaudeExecutor_Resume_WithYoloMode(t *testing.T) {
	runner := &MockCommandRunner{}
	executor := NewClaudeExecutorWithRunner(true, "", "", nil, runner)

	sessionID := "session-123"
	err := executor.Resume(context.Background(), sessionID, "feedback prompt")
	if err != nil {
		t.Fatalf("Resume() error = %v, want nil", err)
	}

	wantArgs := []string{"--print", "--verbose", "--output-format", "stream-json", "--permission-mode", "acceptEdits", "--resume", sessionID, "--dangerously-skip-permissions"}
	if !reflect.DeepEqual(runner.LastArgs, wantArgs) {
		t.Errorf("args = %v, want %v", runner.LastArgs, wantArgs)
	}
}

// TestClaudeExecutor_Resume_PassesPromptViaStdin verifies that Resume passes prompt via stdin.
func TestClaudeExecutor_Resume_PassesPromptViaStdin(t *testing.T) {
	runner := &MockCommandRunner{}
	executor := NewClaudeExecutorWithRunner(false, "", "", nil, runner)

	prompt := "Fix the security issue on line 42"
	err := executor.Resume(context.Background(), "session-123", prompt)
	if err != nil {
		t.Fatalf("Resume() error = %v, want nil", err)
	}

	if runner.LastStdin != prompt {
		t.Errorf("stdin = %q, want %q", runner.LastStdin, prompt)
	}
}

// TestClaudeExecutor_Resume_RunnerError verifies that Resume returns wrapped runner errors.
func TestClaudeExecutor_Resume_RunnerError(t *testing.T) {
	expectedErr := errors.New("resume failed")
	runner := &MockCommandRunner{
		RunFunc: func(_ context.Context, _ string, _ []string, _ io.Reader, _ string) error {
			return expectedErr
		},
	}
	executor := NewClaudeExecutorWithRunner(false, "", "", nil, runner)

	err := executor.Resume(context.Background(), "session-123", "feedback")
	if !errors.Is(err, expectedErr) {
		t.Errorf("Resume() error = %v, want wrapped %v", err, expectedErr)
	}
	if !strings.Contains(err.Error(), "claude resume failed") {
		t.Errorf("error should contain context: %v", err)
	}
}

// TestNewClaudeExecutor_UsesDefaultRunner verifies that NewClaudeExecutor uses DefaultCommandRunner.
func TestNewClaudeExecutor_UsesDefaultRunner(t *testing.T) {
	executor := NewClaudeExecutor(true, "sonnet", "", nil)

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
func TestClaudeExecutor_InterfaceCompliance(_ *testing.T) {
	var _ Executor = (*ClaudeExecutor)(nil)
}

// TestClaudeExecutor_OutputPath verifies outputPath is computed correctly.
func TestClaudeExecutor_OutputPath(t *testing.T) {
	tests := []struct {
		name       string
		outputDir  string
		sessionID  string
		wantPath   string
	}{
		{
			name:      "both set",
			outputDir: "/tmp/agent-outputs",
			sessionID: "session-123",
			wantPath:  "/tmp/agent-outputs/session-123.log",
		},
		{
			name:      "empty outputDir",
			outputDir: "",
			sessionID: "session-123",
			wantPath:  "",
		},
		{
			name:      "empty sessionID",
			outputDir: "/tmp/agent-outputs",
			sessionID: "",
			wantPath:  "",
		},
		{
			name:      "both empty",
			outputDir: "",
			sessionID: "",
			wantPath:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewClaudeExecutor(false, "", tt.outputDir, nil)
			got := executor.outputPath(tt.sessionID)
			if got != tt.wantPath {
				t.Errorf("outputPath(%q) = %q, want %q", tt.sessionID, got, tt.wantPath)
			}
		})
	}
}

// TestClaudeExecutor_Spawn_PassesOutputPath verifies Spawn passes output path to runner.
func TestClaudeExecutor_Spawn_PassesOutputPath(t *testing.T) {
	runner := &MockCommandRunner{}
	executor := NewClaudeExecutorWithRunner(false, "", "/tmp/outputs", nil, runner)

	err := executor.Spawn(context.Background(), Implementer, "test prompt", "my-session")
	if err != nil {
		t.Fatalf("Spawn() error = %v, want nil", err)
	}

	wantPath := "/tmp/outputs/my-session.log"
	if runner.LastOutputPath != wantPath {
		t.Errorf("LastOutputPath = %q, want %q", runner.LastOutputPath, wantPath)
	}
}

// TestClaudeExecutor_Resume_PassesOutputPath verifies Resume passes output path to runner.
func TestClaudeExecutor_Resume_PassesOutputPath(t *testing.T) {
	runner := &MockCommandRunner{}
	executor := NewClaudeExecutorWithRunner(false, "", "/tmp/outputs", nil, runner)

	err := executor.Resume(context.Background(), "my-session", "feedback")
	if err != nil {
		t.Fatalf("Resume() error = %v, want nil", err)
	}

	wantPath := "/tmp/outputs/my-session.log"
	if runner.LastOutputPath != wantPath {
		t.Errorf("LastOutputPath = %q, want %q", runner.LastOutputPath, wantPath)
	}
}

// TestClaudeExecutor_Spawn_WithCustomArgs verifies Spawn appends custom args.
func TestClaudeExecutor_Spawn_WithCustomArgs(t *testing.T) {
	runner := &MockCommandRunner{}
	customArgs := []string{"--custom-flag", "value"}
	executor := NewClaudeExecutorWithRunner(false, "", "", customArgs, runner)

	err := executor.Spawn(context.Background(), Implementer, "test prompt", "")
	if err != nil {
		t.Fatalf("Spawn() error = %v, want nil", err)
	}

	// Verify custom args are at the end
	args := runner.LastArgs
	if len(args) < 2 {
		t.Fatalf("expected at least 2 args, got %d", len(args))
	}

	// Check that custom args are appended at the end
	lastTwo := args[len(args)-2:]
	if lastTwo[0] != "--custom-flag" || lastTwo[1] != "value" {
		t.Errorf("expected custom args at end, got %v", lastTwo)
	}
}

// TestClaudeExecutor_Resume_WithCustomArgs verifies Resume appends custom args.
func TestClaudeExecutor_Resume_WithCustomArgs(t *testing.T) {
	runner := &MockCommandRunner{}
	customArgs := []string{"--extra", "arg"}
	executor := NewClaudeExecutorWithRunner(false, "", "", customArgs, runner)

	err := executor.Resume(context.Background(), "session-123", "feedback")
	if err != nil {
		t.Fatalf("Resume() error = %v, want nil", err)
	}

	// Verify custom args are at the end
	args := runner.LastArgs
	if len(args) < 2 {
		t.Fatalf("expected at least 2 args, got %d", len(args))
	}

	// Check that custom args are appended at the end
	lastTwo := args[len(args)-2:]
	if lastTwo[0] != "--extra" || lastTwo[1] != "arg" {
		t.Errorf("expected custom args at end, got %v", lastTwo)
	}
}

// TestClaudeExecutor_ValidateAvailability verifies the validation method exists.
// Note: We can't easily test the actual exec.LookPath behavior without mocking,
// but we can verify the method exists and returns an error type.
func TestClaudeExecutor_ValidateAvailability(t *testing.T) {
	executor := NewClaudeExecutor(false, "", "", nil)

	// Just verify the method exists and returns either nil or an error
	err := executor.ValidateAvailability()
	// We can't predict if "claude" is on PATH, so just verify no panic
	_ = err
}
