package agents

import (
	"context"
	"fmt"
	"strings"
)

// ClaudeExecutor implements Executor for the Claude Code CLI.
// It spawns claude processes with appropriate flags for agent execution
// and supports session resumption for bidirectional communication.
//
// The executor uses the CommandRunner interface for subprocess execution,
// which allows for testing without actually spawning CLI processes.
//
// Example usage:
//
//	executor := NewClaudeExecutor(true, "sonnet")
//	err := executor.Spawn(ctx, agents.Implementer, "Execute task", sessionID)
//	if err != nil {
//	    return fmt.Errorf("failed to spawn: %w", err)
//	}
type ClaudeExecutor struct {
	yoloMode bool   // When true, adds --dangerously-skip-permissions flag
	model    string // Model to use (e.g., "sonnet", "opus"), empty for default
	runner   CommandRunner
}

// NewClaudeExecutor creates a ClaudeExecutor with the given configuration.
// It uses DefaultCommandRunner for subprocess execution.
//
// Parameters:
//   - yoloMode: if true, skips permission prompts (--dangerously-skip-permissions)
//   - model: model name to use (empty string uses default)
func NewClaudeExecutor(yoloMode bool, model string) *ClaudeExecutor {
	return &ClaudeExecutor{
		yoloMode: yoloMode,
		model:    model,
		runner:   &DefaultCommandRunner{},
	}
}

// NewClaudeExecutorWithRunner creates a ClaudeExecutor with a custom CommandRunner.
// This is primarily for testing to inject mock command execution.
//
// Parameters:
//   - yoloMode: if true, skips permission prompts (--dangerously-skip-permissions)
//   - model: model name to use (empty string uses default)
//   - runner: custom CommandRunner for subprocess execution
func NewClaudeExecutorWithRunner(yoloMode bool, model string, runner CommandRunner) *ClaudeExecutor {
	return &ClaudeExecutor{
		yoloMode: yoloMode,
		model:    model,
		runner:   runner,
	}
}

// Name returns the executor identifier.
// This is used for registry lookup and configuration.
func (e *ClaudeExecutor) Name() string {
	return "claude-code"
}

// Spawn invokes the Claude CLI with the given agent and prompt.
// It loads the agent's prompt template, combines it with the task prompt,
// builds the appropriate CLI arguments, and executes the claude command.
//
// The method blocks until the subprocess exits.
//
// Parameters:
//   - ctx: context for cancellation
//   - agent: agent configuration with prompt template path
//   - prompt: task-specific prompt to append to agent prompt
//   - sessionID: session identifier for session persistence (empty for no session)
//
// Returns error if prompt loading fails or subprocess execution fails.
func (e *ClaudeExecutor) Spawn(ctx context.Context, agent *Agent, prompt string, sessionID string) error {
	// Load agent prompt template
	agentPrompt, err := LoadPrompt(agent.PromptPath)
	if err != nil {
		return fmt.Errorf("failed to load agent prompt: %w", err)
	}

	// Combine prompts
	fullPrompt := agentPrompt + "\n\n" + prompt

	// Build args
	args := []string{}
	if e.yoloMode {
		args = append(args, "--dangerously-skip-permissions")
	}
	if e.model != "" {
		args = append(args, "--model", e.model)
	}
	if sessionID != "" {
		args = append(args, "--session-id", sessionID)
	}

	// Execute
	if err = e.runner.Run(ctx, "claude", args, strings.NewReader(fullPrompt)); err != nil {
		return fmt.Errorf("claude spawn failed: %w", err)
	}
	return nil
}

// Resume continues an existing Claude session with additional prompt.
// It uses the --resume flag to resume the session identified by sessionID.
//
// The method blocks until the subprocess exits.
//
// Parameters:
//   - ctx: context for cancellation
//   - sessionID: session identifier to resume
//   - prompt: additional prompt to send to the session
//
// Returns error if subprocess execution fails.
func (e *ClaudeExecutor) Resume(ctx context.Context, sessionID string, prompt string) error {
	args := []string{"--resume", sessionID}
	if err := e.runner.Run(ctx, "claude", args, strings.NewReader(prompt)); err != nil {
		return fmt.Errorf("claude resume failed: %w", err)
	}
	return nil
}

// SupportsResumption indicates that Claude Code supports session resumption.
// Claude Code's --resume flag allows continuing existing sessions.
func (e *ClaudeExecutor) SupportsResumption() bool {
	return true
}

// Compile-time check that ClaudeExecutor implements Executor.
var _ Executor = (*ClaudeExecutor)(nil)
