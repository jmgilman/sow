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
	executor := NewCursorExecutor(false, "", nil)
	name := executor.Name()
	if name != "cursor" {
		t.Errorf("Name() = %q, want %q", name, "cursor")
	}
}

// TestCursorExecutor_SupportsResumption verifies SupportsResumption() returns true.
func TestCursorExecutor_SupportsResumption(t *testing.T) {
	executor := NewCursorExecutor(false, "", nil)
	supports := executor.SupportsResumption()
	if !supports {
		t.Errorf("SupportsResumption() = false, want true")
	}
}

// TestCursorExecutor_Spawn_BuildsCorrectArgs verifies Spawn builds correct CLI arguments.
// Note: --print and --output-format stream-json are always included for non-interactive streaming mode.
// --force is only included when yoloMode is enabled.
func TestCursorExecutor_Spawn_BuildsCorrectArgs(t *testing.T) {
	tests := []struct {
		name      string
		yoloMode  bool
		sessionID string
		wantArgs  []string
	}{
		{
			name:      "no session ID, no yolo mode",
			yoloMode:  false,
			sessionID: "",
			wantArgs:  []string{"agent", "--print", "--output-format", "stream-json"},
		},
		{
			name:      "with session ID, no yolo mode",
			yoloMode:  false,
			sessionID: "abc-123",
			wantArgs:  []string{"agent", "--print", "--output-format", "stream-json", "--chat-id", "abc-123"},
		},
		{
			name:      "with UUID session ID, no yolo mode",
			yoloMode:  false,
			sessionID: "550e8400-e29b-41d4-a716-446655440000",
			wantArgs:  []string{"agent", "--print", "--output-format", "stream-json", "--chat-id", "550e8400-e29b-41d4-a716-446655440000"},
		},
		{
			name:      "yolo mode enabled, no session ID",
			yoloMode:  true,
			sessionID: "",
			wantArgs:  []string{"agent", "--print", "--output-format", "stream-json", "--force"},
		},
		{
			name:      "yolo mode enabled, with session ID",
			yoloMode:  true,
			sessionID: "abc-123",
			wantArgs:  []string{"agent", "--print", "--output-format", "stream-json", "--force", "--chat-id", "abc-123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &MockCommandRunner{}

			executor := NewCursorExecutorWithRunner(tt.yoloMode, "", nil, runner)
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

	executor := NewCursorExecutorWithRunner(false, "", nil, runner)
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

	executor := NewCursorExecutorWithRunner(false, "", nil, runner)
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
		RunFunc: func(_ context.Context, _ string, _ []string, _ io.Reader, _ string) error {
			return expectedErr
		},
	}

	executor := NewCursorExecutorWithRunner(false, "", nil, runner)
	err := executor.Spawn(context.Background(), Implementer, "test prompt", "")

	if !errors.Is(err, expectedErr) {
		t.Errorf("Spawn() = %v, want wrapped %v", err, expectedErr)
	}
	if !strings.Contains(err.Error(), "cursor-agent spawn failed") {
		t.Errorf("error should contain context: %v", err)
	}
}

// TestCursorExecutor_Resume_BuildsCorrectArgs verifies Resume builds correct CLI arguments.
// Note: --print and --output-format stream-json are always included for non-interactive streaming mode.
// --force is only included when yoloMode is enabled.
func TestCursorExecutor_Resume_BuildsCorrectArgs(t *testing.T) {
	runner := &MockCommandRunner{}

	executor := NewCursorExecutorWithRunner(false, "", nil, runner)
	err := executor.Resume(context.Background(), "session-123", "resume feedback")
	if err != nil {
		t.Fatalf("Resume() unexpected error: %v", err)
	}

	if runner.LastName != "cursor-agent" {
		t.Errorf("command = %q, want %q", runner.LastName, "cursor-agent")
	}

	wantArgs := []string{"agent", "--print", "--output-format", "stream-json", "--resume", "session-123"}
	if !reflect.DeepEqual(runner.LastArgs, wantArgs) {
		t.Errorf("args = %v, want %v", runner.LastArgs, wantArgs)
	}
}

// TestCursorExecutor_Resume_WithYoloMode verifies Resume includes --force when yoloMode is enabled.
func TestCursorExecutor_Resume_WithYoloMode(t *testing.T) {
	runner := &MockCommandRunner{}

	executor := NewCursorExecutorWithRunner(true, "", nil, runner)
	err := executor.Resume(context.Background(), "session-123", "resume feedback")
	if err != nil {
		t.Fatalf("Resume() unexpected error: %v", err)
	}

	wantArgs := []string{"agent", "--print", "--output-format", "stream-json", "--resume", "session-123", "--force"}
	if !reflect.DeepEqual(runner.LastArgs, wantArgs) {
		t.Errorf("args = %v, want %v", runner.LastArgs, wantArgs)
	}
}

// TestCursorExecutor_Resume_PassesPromptViaStdin verifies Resume passes prompt to stdin.
func TestCursorExecutor_Resume_PassesPromptViaStdin(t *testing.T) {
	runner := &MockCommandRunner{}

	executor := NewCursorExecutorWithRunner(false, "", nil, runner)
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
		RunFunc: func(_ context.Context, _ string, _ []string, _ io.Reader, _ string) error {
			return expectedErr
		},
	}

	executor := NewCursorExecutorWithRunner(false, "", nil, runner)
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
	executor := NewCursorExecutor(false, "", nil)

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

	executor := NewCursorExecutorWithRunner(true, "", nil, customRunner)

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

// TestCursorExecutor_YoloModeAffectsForceFlag verifies yoloMode controls --force flag.
func TestCursorExecutor_YoloModeAffectsForceFlag(t *testing.T) {
	// Test with yoloMode true - should include --force
	runnerYolo := &MockCommandRunner{}
	executorYolo := NewCursorExecutorWithRunner(true, "", nil, runnerYolo)
	if err := executorYolo.Spawn(context.Background(), Implementer, "test", ""); err != nil {
		t.Fatalf("Spawn() unexpected error: %v", err)
	}

	hasForce := false
	for _, arg := range runnerYolo.LastArgs {
		if arg == "--force" {
			hasForce = true
			break
		}
	}
	if !hasForce {
		t.Errorf("yoloMode=true should include --force, got args: %v", runnerYolo.LastArgs)
	}

	// Test with yoloMode false - should NOT include --force
	runnerNoYolo := &MockCommandRunner{}
	executorNoYolo := NewCursorExecutorWithRunner(false, "", nil, runnerNoYolo)
	if err := executorNoYolo.Spawn(context.Background(), Implementer, "test", ""); err != nil {
		t.Fatalf("Spawn() unexpected error: %v", err)
	}

	hasForce = false
	for _, arg := range runnerNoYolo.LastArgs {
		if arg == "--force" {
			hasForce = true
			break
		}
	}
	if hasForce {
		t.Errorf("yoloMode=false should NOT include --force, got args: %v", runnerNoYolo.LastArgs)
	}
}

// TestCursorExecutor_OutputPath verifies outputPath is computed correctly.
func TestCursorExecutor_OutputPath(t *testing.T) {
	tests := []struct {
		name      string
		outputDir string
		sessionID string
		wantPath  string
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
			executor := NewCursorExecutor(false, tt.outputDir, nil)
			got := executor.outputPath(tt.sessionID)
			if got != tt.wantPath {
				t.Errorf("outputPath(%q) = %q, want %q", tt.sessionID, got, tt.wantPath)
			}
		})
	}
}

// TestCursorExecutor_Spawn_PassesOutputPath verifies Spawn passes output path to runner.
func TestCursorExecutor_Spawn_PassesOutputPath(t *testing.T) {
	runner := &MockCommandRunner{}
	executor := NewCursorExecutorWithRunner(false, "/tmp/outputs", nil, runner)

	err := executor.Spawn(context.Background(), Implementer, "test prompt", "my-session")
	if err != nil {
		t.Fatalf("Spawn() error = %v, want nil", err)
	}

	wantPath := "/tmp/outputs/my-session.log"
	if runner.LastOutputPath != wantPath {
		t.Errorf("LastOutputPath = %q, want %q", runner.LastOutputPath, wantPath)
	}
}

// TestCursorExecutor_Resume_PassesOutputPath verifies Resume passes output path to runner.
func TestCursorExecutor_Resume_PassesOutputPath(t *testing.T) {
	runner := &MockCommandRunner{}
	executor := NewCursorExecutorWithRunner(false, "/tmp/outputs", nil, runner)

	err := executor.Resume(context.Background(), "my-session", "feedback")
	if err != nil {
		t.Fatalf("Resume() error = %v, want nil", err)
	}

	wantPath := "/tmp/outputs/my-session.log"
	if runner.LastOutputPath != wantPath {
		t.Errorf("LastOutputPath = %q, want %q", runner.LastOutputPath, wantPath)
	}
}

// TestCursorExecutor_Spawn_WithCustomArgs verifies Spawn appends custom args.
func TestCursorExecutor_Spawn_WithCustomArgs(t *testing.T) {
	runner := &MockCommandRunner{}
	customArgs := []string{"--custom-flag", "value"}
	executor := NewCursorExecutorWithRunner(false, "", customArgs, runner)

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

// TestCursorExecutor_Resume_WithCustomArgs verifies Resume appends custom args.
func TestCursorExecutor_Resume_WithCustomArgs(t *testing.T) {
	runner := &MockCommandRunner{}
	customArgs := []string{"--extra", "arg"}
	executor := NewCursorExecutorWithRunner(false, "", customArgs, runner)

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

// TestCursorExecutor_ValidateAvailability verifies the validation method exists.
func TestCursorExecutor_ValidateAvailability(t *testing.T) {
	executor := NewCursorExecutor(false, "", nil)

	// Just verify the method exists and returns either nil or an error
	err := executor.ValidateAvailability()
	// We can't predict if "cursor-agent" is on PATH, so just verify no panic
	_ = err
}
