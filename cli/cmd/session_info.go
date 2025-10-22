package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	projectpkg "github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/spf13/cobra"
)

// SessionInfo represents the session information output structure.
type SessionInfo struct {
	Repository  RepositoryInfo   `json:"repository"`
	Context     ContextInfo      `json:"context"`
	Project     *ProjectInfo     `json:"project,omitempty"`
	Statechart  *StatechartInfo  `json:"statechart,omitempty"`
	Versions    VersionInfo      `json:"versions"`
	Available   []string         `json:"available_commands,omitempty"`
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

// StatechartInfo contains statechart state information.
type StatechartInfo struct {
	CurrentState string   `json:"current_state"`
	Permitted    []string `json:"permitted_events,omitempty"`
}

// VersionInfo contains version information.
type VersionInfo struct {
	CLI       string `json:"cli"`
	Structure string `json:"structure"`
	Mismatch  bool   `json:"mismatch"`
}

// NewSessionInfoCmd creates the session-info command.
func NewSessionInfoCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "session-info",
		Short: "Display session context information",
		Long: `Display current session context information.

Shows:
  - Repository root path
  - Current context (task or project)
  - Task details (if in task context)
  - Project details (if project exists)
  - Statechart state (current state, mode, permitted events)
  - CLI version
  - Schema version

By default outputs human-readable text. Use --json for structured JSON output.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSessionInfo(cmd, jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	return cmd
}

func runSessionInfo(cmd *cobra.Command, jsonOutput bool) error {
	// Get Sow from context
	ctx := cmdutil.GetContext(cmd.Context())

	// Check if sow is initialized
	if !ctx.IsInitialized() {
		return sow.ErrNotInitialized
	}

	// Build session info structure
	info := SessionInfo{
		Versions: VersionInfo{
			CLI:       Version,
			Structure: sow.StructureVersion,
			Mismatch:  false, // TODO: implement version comparison if needed
		},
	}

	// Get repository information
	info.Repository.Root = ctx.RepoRoot()
	branch, _ := ctx.Git().CurrentBranch() // Ignore error, just leave empty if fails
	info.Repository.Branch = branch

	// Detect workspace context
	contextType, taskID := sow.DetectContext(ctx.RepoRoot())
	info.Context.Type = contextType
	if contextType == "task" {
		info.Context.TaskID = taskID
	}

	// Get project information if project exists
	proj, err := projectpkg.Load(ctx)
	if err == nil {
		// Project exists - read state
		state := proj.State()

		info.Project = &ProjectInfo{
			Name:        state.Project.Name,
			Branch:      state.Project.Branch,
			Description: state.Project.Description,
		}

		// Determine current phase and status
		currentPhase, status := determineCurrentPhaseAndStatus(state)
		info.Project.Phase = currentPhase
		info.Project.Status = status

		// Load statechart information from project
		machine := proj.Machine()
		currentState := machine.State()

		// Get permitted triggers
		triggers, err := machine.PermittedTriggers()
		if err == nil {
			// Convert triggers to strings
			permittedEvents := make([]string, len(triggers))
			for i, trigger := range triggers {
				permittedEvents[i] = string(trigger)
			}

			info.Statechart = &StatechartInfo{
				CurrentState: string(currentState),
				Permitted:    permittedEvents,
			}
		}
	}
	// If project load fails, it means no project exists - info.Project and info.Statechart remain nil

	// Add available commands based on context
	info.Available = getAvailableCommands(info.Context.Type, info.Project != nil)

	// Output in requested format
	if jsonOutput {
		// JSON output
		output, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal session info: %w", err)
		}
		cmd.Println(string(output))
	} else {
		// Human-readable output
		if info.Project == nil {
			cmd.Println("No active project")
		} else {
			cmd.Printf("Project: %s\n", info.Project.Name)
			cmd.Printf("Description: %s\n", info.Project.Description)
			if info.Project.Phase != "" {
				cmd.Printf("Phase: %s (%s)\n", info.Project.Phase, info.Project.Status)
			}
		}

		cmd.Printf("Repository: %s\n", info.Repository.Root)
		cmd.Printf("Branch: %s\n", info.Repository.Branch)

		switch info.Context.Type {
		case "task":
			cmd.Printf("Context: Task %s\n", info.Context.TaskID)
		case "project":
			cmd.Println("Context: Project")
		}
	}

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
func getAvailableCommands(_ string, hasProject bool) []string {
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
