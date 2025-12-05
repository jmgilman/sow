package agents

import (
	"context"
	"fmt"
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
//	executor := NewCursorExecutor(true)
//	err := executor.Spawn(ctx, agents.Implementer, "Execute task", "session-123")
type CursorExecutor struct {
	yoloMode bool // When true, adds yolo mode flag (if cursor supports it in the future)
	runner   CommandRunner
}

// NewCursorExecutor creates a CursorExecutor with the given configuration.
// yoloMode: if true, enables automatic mode (exact flag TBD based on cursor-agent CLI)
//
// This constructor uses DefaultCommandRunner for subprocess execution.
func NewCursorExecutor(yoloMode bool) *CursorExecutor {
	return &CursorExecutor{
		yoloMode: yoloMode,
		runner:   &DefaultCommandRunner{},
	}
}

// NewCursorExecutorWithRunner creates a CursorExecutor with a custom CommandRunner.
// This is primarily for testing to inject mock command execution.
func NewCursorExecutorWithRunner(yoloMode bool, runner CommandRunner) *CursorExecutor {
	return &CursorExecutor{
		yoloMode: yoloMode,
		runner:   runner,
	}
}

// Name returns the executor identifier.
// Used for registry lookup and configuration.
func (e *CursorExecutor) Name() string {
	return "cursor"
}

// Spawn invokes a Cursor agent with the given prompt and session ID.
// It loads the agent's prompt template and combines it with the task prompt.
//
// The method builds CLI arguments as follows:
//   - Base command: cursor-agent agent
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
	args := []string{"agent"}
	if sessionID != "" {
		args = append(args, "--chat-id", sessionID)
	}

	// Execute
	return e.runner.Run(ctx, "cursor-agent", args, strings.NewReader(fullPrompt))
}

// Resume continues an existing Cursor session with additional prompt.
// Uses cursor-agent agent --resume <sessionID> format.
//
// The prompt is passed via stdin.
// Blocks until the subprocess exits.
func (e *CursorExecutor) Resume(ctx context.Context, sessionID string, prompt string) error {
	args := []string{"agent", "--resume", sessionID}
	return e.runner.Run(ctx, "cursor-agent", args, strings.NewReader(prompt))
}

// SupportsResumption indicates that Cursor supports session resumption.
// Cursor Agent CLI supports the --resume flag for continuing sessions.
func (e *CursorExecutor) SupportsResumption() bool {
	return true
}

// Compile-time check that CursorExecutor implements Executor.
var _ Executor = (*CursorExecutor)(nil)
