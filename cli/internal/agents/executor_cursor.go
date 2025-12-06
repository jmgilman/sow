package agents

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// CursorExecutor implements Executor for the Cursor Agent CLI.
// It spawns cursor-agent processes with appropriate flags for agent execution
// and supports session resumption for bidirectional communication.
//
// The Cursor Agent CLI uses different flags than Claude Code:
//   - Command: cursor-agent agent (not cursor)
//   - New session: --chat-id <sessionID> (not --session-id)
//   - Resume: --resume <sessionID>
//
// Example usage:
//
//	executor := NewCursorExecutor(true, ".sow/project/agent-outputs")
//	err := executor.Spawn(ctx, agents.Implementer, "Execute task", "session-123")
type CursorExecutor struct {
	yoloMode   bool     // When true, adds --force flag (skip permission prompts)
	outputDir  string   // Directory for output logs (empty disables logging)
	customArgs []string // Additional CLI arguments from user config
	runner     CommandRunner
}

// NewCursorExecutor creates a CursorExecutor with the given configuration.
//
// Parameters:
//   - yoloMode: if true, adds --force flag to skip permission prompts
//   - outputDir: directory for output logs (empty string disables output logging)
//   - customArgs: additional CLI arguments from user config (can be nil)
//
// This constructor uses DefaultCommandRunner for subprocess execution.
func NewCursorExecutor(yoloMode bool, outputDir string, customArgs []string) *CursorExecutor {
	return &CursorExecutor{
		yoloMode:   yoloMode,
		outputDir:  outputDir,
		customArgs: customArgs,
		runner:     &DefaultCommandRunner{},
	}
}

// NewCursorExecutorWithRunner creates a CursorExecutor with a custom CommandRunner.
// This is primarily for testing to inject mock command execution.
//
// Parameters:
//   - yoloMode: if true, adds --force flag to skip permission prompts
//   - outputDir: directory for output logs (empty string disables output logging)
//   - customArgs: additional CLI arguments from user config (can be nil)
//   - runner: custom CommandRunner for subprocess execution
func NewCursorExecutorWithRunner(yoloMode bool, outputDir string, customArgs []string, runner CommandRunner) *CursorExecutor {
	return &CursorExecutor{
		yoloMode:   yoloMode,
		outputDir:  outputDir,
		customArgs: customArgs,
		runner:     runner,
	}
}

// Name returns the executor identifier.
// Used for registry lookup and configuration.
func (e *CursorExecutor) Name() string {
	return "cursor"
}

// outputPath returns the path for saving output logs for a given session.
// Returns empty string if outputDir is not configured.
func (e *CursorExecutor) outputPath(sessionID string) string {
	if e.outputDir == "" || sessionID == "" {
		return ""
	}
	return filepath.Join(e.outputDir, sessionID+".log")
}

// Spawn invokes a Cursor agent with the given prompt and session ID.
// It loads the agent's prompt template and combines it with the task prompt.
//
// The method builds CLI arguments as follows:
//   - Base command: cursor-agent agent
//   - --print: non-interactive mode (always used)
//   - --force: skip permission prompts (only when yoloMode is true)
//   - If sessionID is not empty: --chat-id <sessionID>
//
// The combined prompt is passed via stdin.
// Blocks until the subprocess exits.
func (e *CursorExecutor) Spawn(ctx context.Context, agent *Agent, prompt string, sessionID string) error {
	// Load agent prompt template
	agentPrompt, err := LoadPrompt(agent.PromptPath)
	if err != nil {
		return fmt.Errorf("failed to load agent prompt: %w", err)
	}

	// Combine prompts
	fullPrompt := agentPrompt + "\n\n" + prompt

	// Build args - cursor-agent uses subcommand "agent"
	// Use stream-json for real-time output to log files
	args := []string{"agent", "--print", "--output-format", "stream-json"}

	// Add --force when yoloMode enabled (Cursor's equivalent of skip-permissions)
	if e.yoloMode {
		args = append(args, "--force")
	}

	if sessionID != "" {
		args = append(args, "--chat-id", sessionID)
	}

	// Append custom arguments from user config
	if len(e.customArgs) > 0 {
		args = append(args, e.customArgs...)
	}

	// Execute with output capture
	if err = e.runner.Run(ctx, "cursor-agent", args, strings.NewReader(fullPrompt), e.outputPath(sessionID)); err != nil {
		return fmt.Errorf("cursor-agent spawn failed: %w", err)
	}
	return nil
}

// Resume continues an existing Cursor session with additional prompt.
// Uses cursor-agent agent --resume <sessionID> format.
//
// The --print flag is always used for non-interactive mode.
// The --force flag is added when yoloMode is enabled.
//
// The prompt is passed via stdin.
// Blocks until the subprocess exits.
func (e *CursorExecutor) Resume(ctx context.Context, sessionID string, prompt string) error {
	// Use stream-json for real-time output to log files
	args := []string{"agent", "--print", "--output-format", "stream-json", "--resume", sessionID}

	// Add --force when yoloMode enabled (Cursor's equivalent of skip-permissions)
	if e.yoloMode {
		args = append(args, "--force")
	}

	// Append custom arguments from user config
	if len(e.customArgs) > 0 {
		args = append(args, e.customArgs...)
	}

	// Execute with output capture (appends to same session log)
	if err := e.runner.Run(ctx, "cursor-agent", args, strings.NewReader(prompt), e.outputPath(sessionID)); err != nil {
		return fmt.Errorf("cursor-agent resume failed: %w", err)
	}
	return nil
}

// SupportsResumption indicates that Cursor supports session resumption.
// Cursor Agent CLI supports the --resume flag for continuing sessions.
func (e *CursorExecutor) SupportsResumption() bool {
	return true
}

// ValidateAvailability checks if the cursor-agent CLI binary is available on PATH.
// Returns nil if available, error with installation guidance if not.
func (e *CursorExecutor) ValidateAvailability() error {
	_, err := exec.LookPath("cursor-agent")
	if err != nil {
		return fmt.Errorf("cursor-agent CLI not found on PATH: %w\n\nInstall from: https://cursor.com/cli", err)
	}
	return nil
}

// Compile-time check that CursorExecutor implements Executor.
var _ Executor = (*CursorExecutor)(nil)
