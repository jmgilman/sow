package project

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/spf13/cobra"
)

// NewInitCmd creates the project init command.
func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Initialize a new project",
		Long: `Initialize a new project on the current branch.

Creates the initial project state with all 5 phases. By default, only the
required phases (Implementation, Review, Finalize) are enabled. The truth
table in /project:new will later modify phase enablement based on requirements.

GitHub Issue Integration:
  When using --issue, the project is linked to a GitHub issue. The CLI will:
  - Fetch the issue and validate it has the 'sow' label
  - Check if the issue already has a linked branch (fails if found)
  - Create a branch via 'gh issue develop' (auto-named or custom via --branch-name)
  - Initialize the project with issue details
  - Link the project to the issue

Requirements:
  - Must be in a sow repository (.sow directory exists)
  - Must be on a feature branch (not main or master)
  - No existing project can be present
  - If using --issue: gh CLI must be installed and authenticated

Examples:
  # Create project manually
  sow project init my-feature --description "Add authentication feature"

  # Create project from GitHub issue
  sow project init --issue 123

  # Create project from issue with custom branch name
  sow project init --issue 123 --branch-name custom-branch`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd, args)
		},
	}

	// Flags
	cmd.Flags().StringP("description", "d", "", "Project description (required unless using --issue)")
	cmd.Flags().IntP("issue", "i", 0, "GitHub issue number to link this project to")
	cmd.Flags().String("branch-name", "", "Custom branch name (only with --issue)")

	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	issueNumber, _ := cmd.Flags().GetInt("issue")
	description, _ := cmd.Flags().GetString("description")
	branchName, _ := cmd.Flags().GetString("branch-name")

	// Get Sow from context
	sowInstance := cmdutil.SowFromContext(cmd.Context())

	var project *sow.Project
	var err error

	if issueNumber > 0 {
		// Create project from GitHub issue
		if branchName != "" && len(args) > 0 {
			return fmt.Errorf("cannot specify both <name> and --branch-name when using --issue")
		}

		if description != "" {
			_, _ = fmt.Fprintf(cmd.OutOrStderr(), "ℹ️  Note: --description is ignored when using --issue (description will be taken from issue)\n\n")
		}

		project, err = sowInstance.CreateProjectFromIssue(issueNumber, branchName)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "\n✓ Initialized project '%s' on branch '%s' (linked to issue #%d)\n",
			project.Name(), project.Branch(), issueNumber)
	} else {
		// Create project manually
		if len(args) == 0 {
			return fmt.Errorf("project name is required when not using --issue")
		}

		if description == "" {
			return fmt.Errorf("--description is required when not using --issue")
		}

		if branchName != "" {
			return fmt.Errorf("--branch-name can only be used with --issue")
		}

		name := args[0]
		project, err = sowInstance.CreateProject(name, description)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "\n✓ Initialized project '%s' on branch '%s'\n",
			project.Name(), project.Branch())
	}

	return nil
}
