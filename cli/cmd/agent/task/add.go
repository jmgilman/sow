package task

import (
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	projectpkg "github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/spf13/cobra"
)

// NewAddCmd creates the task add command.
func NewAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a new task to the implementation phase",
		Long: `Add a new task to the current project's implementation phase.

Creates a new task with gap-numbered ID (010, 020, 030...) and initializes
the task state, description file, and log file. Tasks are atomic units of
work assigned to specific agent types.

The task ID is auto-generated unless specified with --id. Gap numbering
allows insertion of tasks between existing ones if needed.

Requirements:
  - Must be in a sow repository with an active project
  - Project implementation phase must be enabled
  - Task ID must be unique if specified

Example:
  sow task add "Add authentication" --description "Implement JWT auth" --agent implementer
  sow task add "Review auth code" --description "Review JWT implementation" --agent reviewer --dependencies 010`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdd(cmd, args)
		},
	}

	// Flags
	cmd.Flags().StringP("description", "d", "", "Task description (required)")
	cmd.Flags().StringP("agent", "a", "implementer", "Agent type to execute this task")
	cmd.Flags().BoolP("parallel", "p", false, "Whether task can run in parallel with others")
	cmd.Flags().StringSliceP("dependencies", "D", nil, "Task IDs this task depends on (comma-separated)")
	cmd.Flags().String("id", "", "Task ID (auto-generated if not specified)")

	_ = cmd.MarkFlagRequired("description")

	return cmd
}

func runAdd(cmd *cobra.Command, args []string) error {
	name := args[0]
	description, _ := cmd.Flags().GetString("description")
	agent, _ := cmd.Flags().GetString("agent")
	parallel, _ := cmd.Flags().GetBool("parallel")
	dependencies, _ := cmd.Flags().GetStringSlice("dependencies")
	idFlag, _ := cmd.Flags().GetString("id")

	// Trim dependencies
	for i, dep := range dependencies {
		dependencies[i] = strings.TrimSpace(dep)
	}

	// Get Sow from context
	ctx := cmdutil.GetContext(cmd.Context())

	// Get project
	project, err := projectpkg.Load(ctx)
	if err != nil {
		return fmt.Errorf("no active project - run 'sow project init' first")
	}

	// Build task options
	opts := []sow.TaskOption{
		sow.WithAgent(agent),
		sow.WithParallel(parallel),
		sow.WithDescription(description),
	}

	if len(dependencies) > 0 {
		opts = append(opts, sow.WithDependencies(dependencies...))
	}

	if idFlag != "" {
		opts = append(opts, sow.WithID(idFlag))
	}

	// Create task (handles validation, state machine transitions, file creation)
	task, err := project.AddTask(name, opts...)
	if err != nil {
		return err
	}

	// Success
	taskID := task.ID()
	cmd.Printf("\n✓ Created task %s: %s\n", taskID, name)
	cmd.Printf("\nTask Details:\n")
	cmd.Printf("  ID:           %s\n", taskID)
	cmd.Printf("  Name:         %s\n", name)
	cmd.Printf("  Status:       pending\n")
	cmd.Printf("  Agent:        %s\n", agent)
	cmd.Printf("  Parallel:     %v\n", parallel)
	if len(dependencies) > 0 {
		cmd.Printf("  Dependencies: %s\n", strings.Join(dependencies, ", "))
	} else {
		cmd.Printf("  Dependencies: none\n")
	}

	return nil
}
