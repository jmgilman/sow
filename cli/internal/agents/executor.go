package agents

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jmgilman/sow/cli/internal/agents/logformat"
)

// Executor invokes agent CLIs and manages sessions.
// Each implementation handles a specific CLI tool (Claude, Cursor, etc.).
// The interface supports spawning new agent sessions and resuming existing ones.
//
// Executor methods block until the subprocess exits. Session state is persisted
// externally (e.g., in task state.yaml files), not in Go memory. This design
// supports the spawn -> block -> exit -> read state workflow.
//
// Example usage:
//
//	func spawnWorker(ctx context.Context, executor Executor, agent *Agent, taskID string) error {
//	    sessionID := uuid.New().String()
//	    prompt := fmt.Sprintf("Execute task %s", taskID)
//	    return executor.Spawn(ctx, agent, prompt, sessionID)
//	}
type Executor interface {
	// Name returns the executor identifier (e.g., "claude-code", "cursor").
	// Used for registry lookup and configuration.
	Name() string

	// Spawn invokes an agent with the given prompt and session ID.
	// Blocks until the subprocess exits.
	// The sessionID should be persisted before calling Spawn.
	// Returns error if subprocess fails or context is cancelled.
	Spawn(ctx context.Context, agent *Agent, prompt string, sessionID string) error

	// Resume continues an existing session with additional prompt.
	// Blocks until the subprocess exits.
	// Returns error if session not found, executor doesn't support resumption,
	// subprocess fails, or context is cancelled.
	Resume(ctx context.Context, sessionID string, prompt string) error

	// SupportsResumption indicates if this executor can resume sessions.
	// Some CLIs may not support session resumption.
	SupportsResumption() bool

	// ValidateAvailability checks if the executor's CLI binary is available on PATH.
	// Returns nil if available, error with actionable message if not.
	// Should be called before Spawn/Resume to fail fast with clear guidance.
	ValidateAvailability() error
}

// CommandRunner abstracts subprocess execution for testability.
// In production, this is backed by os/exec. In tests, it's mocked.
//
// This abstraction allows executor implementations to be tested without
// actually spawning CLI processes.
type CommandRunner interface {
	// Run executes a command with the given arguments and stdin.
	// The outputPath parameter specifies where to save command output.
	// If outputPath is empty, output goes only to stdout/stderr.
	// If outputPath is set, output is tee'd to both terminal and file.
	// Returns error if command fails or context is cancelled.
	Run(ctx context.Context, name string, args []string, stdin io.Reader, outputPath string) error
}

// DefaultCommandRunner implements CommandRunner using os/exec.
// Output is written to a file if outputPath is provided, otherwise discarded.
//
// When outputPath is provided, the runner creates two output files:
//   - {base}.json: Raw stream-json output for machine processing
//   - {base}.log: Human-readable formatted output
//
// For example, if outputPath is "session-123.log", it creates:
//   - session-123.json (raw JSON)
//   - session-123.log (formatted)
type DefaultCommandRunner struct{}

// Run executes a command with the given arguments and stdin.
//
// If outputPath is non-empty, output is written to two files:
//   - Raw JSON to {base}.json
//   - Formatted output to {base}.log (or the original path if not .log)
//
// The output files are created/appended, and the directory is created if needed.
// If outputPath is empty, output is discarded (but stderr is still captured for error messages).
//
// Note: Agent commands are invoked by the orchestrator, not humans, so
// terminal output is not needed. The orchestrator reads state.yaml after
// the subprocess exits to determine the outcome.
func (r *DefaultCommandRunner) Run(ctx context.Context, name string, args []string, stdin io.Reader, outputPath string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = stdin

	// Always capture stderr for error messages
	var stderrBuf bytes.Buffer

	// Track dualWriter for flushing after command completes
	var dualWriter *logformat.DualWriter

	if outputPath != "" {
		// Ensure output directory exists
		outputDir := filepath.Dir(outputPath)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Determine file paths for raw JSON and formatted output
		rawPath, formattedPath := splitOutputPaths(outputPath)

		// Open raw JSON file for appending
		rawFile, err := os.OpenFile(rawPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open raw output file: %w", err)
		}
		defer func() { _ = rawFile.Close() }()

		// Open formatted output file for appending
		formattedFile, err := os.OpenFile(formattedPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open formatted output file: %w", err)
		}
		defer func() { _ = formattedFile.Close() }()

		// Create a dual writer that writes raw JSON and formatted output
		dualWriter = logformat.NewDualWriter(rawFile, formattedFile)

		cmd.Stdout = dualWriter
		// Write stderr to raw file, formatted file, and buffer (for error messages)
		cmd.Stderr = io.MultiWriter(rawFile, formattedFile, &stderrBuf)
	} else {
		// No file output, but still capture stderr for error messages
		cmd.Stderr = &stderrBuf
	}

	runErr := cmd.Run()

	// Flush any buffered formatted output before returning
	if dualWriter != nil {
		_ = dualWriter.Flush()
	}

	if runErr != nil {
		stderrStr := strings.TrimSpace(stderrBuf.String())
		if stderrStr != "" {
			return fmt.Errorf("command %s failed: %w\nstderr: %s", name, runErr, stderrStr)
		}
		return fmt.Errorf("command %s failed: %w", name, runErr)
	}
	return nil
}

// splitOutputPaths determines the raw JSON and formatted output paths.
// If outputPath ends in .log, the raw path uses .json extension.
// Otherwise, raw uses .json suffix and formatted uses .log suffix.
func splitOutputPaths(outputPath string) (rawPath, formattedPath string) {
	ext := filepath.Ext(outputPath)
	base := strings.TrimSuffix(outputPath, ext)

	if ext == ".log" {
		// Replace .log with .json for raw output
		rawPath = base + ".json"
		formattedPath = outputPath
	} else {
		// Add extensions to the base path
		rawPath = outputPath + ".json"
		formattedPath = outputPath + ".log"
	}
	return
}

// Compile-time check that DefaultCommandRunner implements CommandRunner.
var _ CommandRunner = (*DefaultCommandRunner)(nil)
