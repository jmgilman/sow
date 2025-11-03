// Package modes provides a unified framework for different sow operating modes.
// This includes exploration mode, design mode, and any future modes.
package modes

import (
	"github.com/jmgilman/sow/cli/internal/prompts"
	"github.com/jmgilman/sow/cli/internal/sow"
)

// Mode defines the interface that all sow modes must implement.
// This allows for shared branch handling, index management, and prompt generation.
type Mode interface {
	// Name returns the mode name (e.g., "exploration", "design")
	Name() string

	// BranchPrefix returns the expected branch prefix (e.g., "explore/", "design/")
	BranchPrefix() string

	// DirectoryName returns the directory name under .sow/ (e.g., "exploration", "design")
	DirectoryName() string

	// IndexPath returns the path to the index file relative to .sow/
	IndexPath() string

	// PromptID returns the prompt template ID for this mode
	PromptID() prompts.PromptID

	// ValidStatuses returns the valid status values for this mode
	ValidStatuses() []string
}

// ModeInfo contains information about a mode session.
type ModeInfo struct {
	Branch          string
	Topic           string
	ShouldCreateNew bool
}

// RunResult encapsulates the result of running a mode.
type RunResult struct {
	SelectedBranch  string
	Topic           string
	Prompt          string
	ShouldCreateNew bool
}

// ExtractionResult contains the result of extracting a topic from a branch name.
type ExtractionResult struct {
	Topic string
	Found bool
}

// ExtractTopicFromBranch extracts the topic from a branch name using the mode's prefix.
// If the branch starts with the mode's prefix, it strips that prefix.
// Otherwise, it uses the full branch name as the topic.
func ExtractTopicFromBranch(mode Mode, branchName string) string {
	prefix := mode.BranchPrefix()
	if len(branchName) > len(prefix) && branchName[:len(prefix)] == prefix {
		return branchName[len(prefix):]
	}
	return branchName
}

// ExistsFunc is a function that checks if a mode exists in the current context.
type ExistsFunc func(ctx *sow.Context) bool
