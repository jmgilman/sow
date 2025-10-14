package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/your-org/sow/internal/validation"
)

// NewValidateCmd creates the validate command
func NewValidateCmd() *cobra.Command {
	var fileType string

	cmd := &cobra.Command{
		Use:   "validate [flags] <file-pattern>",
		Short: "Validate file(s) against CUE schemas",
		Long: `Validate sow files against embedded CUE schemas.

Auto-detects file type from path or use --type flag.
Supports glob patterns for multiple files.

Type Detection:
  **/project/state.yaml       → project-state
  **/tasks/*/state.yaml        → task-state
  **/sinks/index.json          → sink-index
  **/repos/index.json          → repo-index
  **/.version                  → sow-version`,
		Example: `  # Auto-detect type
  sow validate .sow/project/state.yaml

  # Explicit type with glob
  sow validate --type task-state '.sow/project/phases/*/tasks/*/state.yaml'

  # Validate all task states
  sow validate --type task-state '.sow/project/phases/implement/tasks/*/state.yaml'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pattern := args[0]
			return runValidate(cmd, pattern, fileType)
		},
	}

	cmd.Flags().StringVar(&fileType, "type", "", "File type: project-state, task-state, sink-index, repo-index, sow-version")

	return cmd
}

func runValidate(cmd *cobra.Command, pattern string, explicitType string) error {
	// Expand glob pattern
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}

	if len(files) == 0 {
		// Try treating it as a literal path
		files = []string{pattern}
	}

	if len(files) == 0 {
		return fmt.Errorf("no files found matching pattern: %s", pattern)
	}

	// Track results
	validCount := 0
	invalidCount := 0
	errors := make(map[string]error)

	// Validate each file
	for _, file := range files {
		// Determine file type
		fileType := explicitType
		if fileType == "" {
			fileType = detectFileType(file)
			if fileType == "" {
				fmt.Fprintf(cmd.OutOrStderr(), "✗ %s: unknown file type (use --type to specify)\n", file)
				invalidCount++
				errors[file] = fmt.Errorf("unknown file type")
				continue
			}
		}

		// Validate using appropriate validator
		var validateErr error
		switch fileType {
		case "project-state":
			validateErr = validation.ValidateProjectState(file)
		case "task-state":
			validateErr = validation.ValidateTaskState(file)
		case "sink-index":
			validateErr = validation.ValidateSinkIndex(file)
		case "repo-index":
			validateErr = validation.ValidateRepoIndex(file)
		case "sow-version":
			validateErr = validation.ValidateVersion(file)
		default:
			fmt.Fprintf(cmd.OutOrStderr(), "✗ %s: unknown type '%s'\n", file, fileType)
			invalidCount++
			errors[file] = fmt.Errorf("unknown type: %s", fileType)
			continue
		}

		if validateErr != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "✗ %s\n", file)
			fmt.Fprintf(cmd.OutOrStderr(), "  %v\n", validateErr)
			invalidCount++
			errors[file] = validateErr
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "✓ %s\n", file)
			validCount++
		}
	}

	// Summary
	fmt.Fprintf(cmd.OutOrStdout(), "\n")
	if invalidCount == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "All %d file(s) valid\n", validCount)
		return nil
	} else {
		fmt.Fprintf(cmd.OutOrStderr(), "%d valid, %d invalid\n", validCount, invalidCount)
		return fmt.Errorf("validation failed for %d file(s)", invalidCount)
	}
}

// detectFileType auto-detects file type from path
func detectFileType(path string) string {
	// Normalize path separators
	normalized := filepath.ToSlash(path)

	// Check patterns
	if strings.Contains(normalized, "/project/state.yaml") {
		return "project-state"
	}
	if strings.Contains(normalized, "/tasks/") && strings.HasSuffix(normalized, "/state.yaml") {
		return "task-state"
	}
	if strings.Contains(normalized, "/sinks/index.json") {
		return "sink-index"
	}
	if strings.Contains(normalized, "/repos/index.json") {
		return "repo-index"
	}
	if strings.HasSuffix(normalized, "/.version") {
		return "sow-version"
	}

	return ""
}
