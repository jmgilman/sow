package taskutil

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/sowfs"
)

// ResolveTaskID resolves a task ID for commands that make it optional.
//
// This helper function implements the task ID resolution logic for commands:
//   1. If providedID is not empty, validate and return it
//   2. If providedID is empty, use InferTaskID() from ContextFS
//   3. Return appropriate errors with helpful messages
//
// Usage in commands:
//
//	taskID, err := taskutil.ResolveTaskID(sowFS, args[0]) // if ID is positional arg
//	taskID, err := taskutil.ResolveTaskID(sowFS, idFlag)  // if ID is a flag
//
// Parameters:
//   - sowFS: The SowFS instance to use for inference
//   - providedID: The task ID provided by the user (empty string if not provided)
//
// Returns:
//   - The resolved task ID (either provided or inferred)
//   - Error if validation fails or inference fails
func ResolveTaskID(sowFS sowfs.SowFS, providedID string) (string, error) {
	// If user provided an ID explicitly, validate and use it
	if providedID != "" {
		// Validation happens in ProjectFS.Task() when we try to access it
		// We just return it here for now
		return providedID, nil
	}

	// No ID provided - try to infer it
	taskID, err := sowFS.Context().InferTaskID()
	if err != nil {
		return "", fmt.Errorf("failed to infer task ID: %w", err)
	}

	return taskID, nil
}

// ResolveTaskIDFromArgs is a convenience wrapper for commands that take task ID as an optional positional argument.
//
// Usage:
//
//	taskID, err := taskutil.ResolveTaskIDFromArgs(sowFS, args)
//
// Parameters:
//   - sowFS: The SowFS instance to use for inference
//   - args: Command arguments (if len > 0, args[0] is used as task ID)
//
// Returns:
//   - The resolved task ID (either from args[0] or inferred)
//   - Error if validation fails or inference fails
func ResolveTaskIDFromArgs(sowFS sowfs.SowFS, args []string) (string, error) {
	providedID := ""
	if len(args) > 0 {
		providedID = args[0]
	}
	return ResolveTaskID(sowFS, providedID)
}
