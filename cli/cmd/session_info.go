package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/sowfs"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/spf13/cobra"
)

// SessionInfo represents the session information output structure.
type SessionInfo struct {
	Repository  RepositoryInfo  `json:"repository"`
	Context     ContextInfo     `json:"context"`
	Project     *ProjectInfo    `json:"project,omitempty"`
	Versions    VersionInfo     `json:"versions"`
	Available   []string        `json:"available_commands,omitempty"`
}

// RepositoryInfo contains git repository information.
type RepositoryInfo struct {
	Root   string `json:"root"`
	Branch string `json:"branch,omitempty"`
}

// ContextInfo describes the current workspace context.
type ContextInfo struct {
	Type   string `json:"type"` // "none", "project", "task"
	TaskID string `json:"task_id,omitempty"`
}

// ProjectInfo contains active project details.
type ProjectInfo struct {
	Name        string `json:"name"`
	Branch      string `json:"branch"`
	Description string `json:"description"`
	Phase       string `json:"current_phase,omitempty"`
	Status      string `json:"status,omitempty"`
}

// VersionInfo contains version information.
type VersionInfo struct {
	CLI       string `json:"cli"`
	Structure string `json:"structure"`
	Mismatch  bool   `json:"mismatch"`
}

// NewSessionInfoCmd creates the session-info command.
func NewSessionInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session-info",
		Short: "Display session context information",
		Long: `Display current session context information.

Shows:
  - Repository root path
  - Current context (task or project)
  - Task details (if in task context)
  - Project details (if project exists)
  - CLI version
  - Schema version

Output is JSON for easy consumption by agents and tools.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSessionInfo(cmd)
		},
	}

	return cmd
}

func runSessionInfo(cmd *cobra.Command) error {
	// Get SowFS from context (may be nil if not in .sow directory)
	sowFS := SowFSFromContext(cmd.Context())
	if sowFS == nil {
		return fmt.Errorf("not in a sow repository - run 'sow init' first")
	}

	// Build session info structure
	info := SessionInfo{
		Versions: VersionInfo{
			CLI:       Version,
			Structure: sowfs.CurrentVersion,
			Mismatch:  false, // TODO: implement version comparison if needed
		},
	}

	// Get repository information
	info.Repository.Root = sowFS.RepoRoot()

	// Get current git branch
	repo, err := sowFS.Repo()
	if err != nil {
		return fmt.Errorf("failed to access git repository: %w", err)
	}

	branch, err := repo.CurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	info.Repository.Branch = branch

	// Detect workspace context
	wsContext, err := sowFS.Context().Detect()
	if err != nil {
		return fmt.Errorf("failed to detect context: %w", err)
	}

	info.Context.Type = wsContext.Type.String()
	if wsContext.Type == sowfs.ContextTask {
		info.Context.TaskID = wsContext.TaskID
	}

	// Get project information if project exists
	projectFS, err := sowFS.Project()
	if err == nil {
		// Project exists - read state
		state, err := projectFS.State()
		if err != nil {
			return fmt.Errorf("failed to read project state: %w", err)
		}

		info.Project = &ProjectInfo{
			Name:        state.Project.Name,
			Branch:      state.Project.Branch,
			Description: state.Project.Description,
		}

		// Determine current phase and status
		currentPhase, status := determineCurrentPhaseAndStatus(state)
		info.Project.Phase = currentPhase
		info.Project.Status = status
	} else if err != sowfs.ErrProjectNotFound {
		// Unexpected error (not just "no project")
		return fmt.Errorf("failed to check project: %w", err)
	}

	// Add available commands based on context
	info.Available = getAvailableCommands(info.Context.Type, info.Project != nil)

	// Marshal to JSON with indentation
	output, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session info: %w", err)
	}

	cmd.Println(string(output))
	return nil
}

// determineCurrentPhaseAndStatus analyzes project state to determine current phase and status.
func determineCurrentPhaseAndStatus(state *schemas.ProjectState) (string, string) {
	// Check phases in order: discovery, design, implementation, review, finalize
	phases := []struct {
		name   string
		status string
	}{
		{"discovery", state.Phases.Discovery.Status},
		{"design", state.Phases.Design.Status},
		{"implementation", state.Phases.Implementation.Status},
		{"review", state.Phases.Review.Status},
		{"finalize", state.Phases.Finalize.Status},
	}

	// Find the first non-completed phase
	for _, phase := range phases {
		if phase.status == "in_progress" {
			return phase.name, "in_progress"
		}
		if phase.status == "pending" {
			return phase.name, "pending"
		}
	}

	// All phases completed
	return "finalize", "completed"
}

// getAvailableCommands returns a list of relevant commands based on context.
func getAvailableCommands(contextType string, hasProject bool) []string {
	commands := []string{
		"sow validate",
		"sow refs",
	}

	if hasProject {
		commands = append(commands,
			"sow log",
			"sow session-info",
		)
	} else {
		commands = append(commands,
			"sow init", // Suggest init if no project
		)
	}

	return commands
}
