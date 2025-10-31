package agent

import (
	"fmt"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/project/loader"
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
  sow agent init my-feature --description "Add authentication feature"

  # Create project from GitHub issue
  sow agent init --issue 123

  # Create project from issue with custom branch name
  sow agent init --issue 123 --branch-name custom-branch`,
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

	// Get context (require .sow to exist)
	ctx, err := cmdutil.RequireInitialized(cmd.Context())
	if err != nil {
		return err
	}

	if issueNumber > 0 {
		return initFromIssue(cmd, args, ctx, issueNumber, description, branchName)
	}
	return initManual(cmd, args, ctx, description, branchName)
}

// initFromIssue creates a project from a GitHub issue.
func initFromIssue(cmd *cobra.Command, args []string, ctx *sow.Context, issueNumber int, description, branchName string) error {
	if branchName != "" && len(args) > 0 {
		return fmt.Errorf("cannot specify both <name> and --branch-name when using --issue")
	}

	if description != "" {
		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "ℹ️  Note: --description is ignored when using --issue (description will be taken from issue)\n\n")
	}

	proj, err := loader.CreateFromIssue(ctx, issueNumber, branchName)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.OutOrStderr(), "\n✓ Initialized project '%s' on branch '%s' (linked to issue #%d)\n",
		proj.Name(), proj.Branch(), issueNumber)
	return nil
}

// initManual creates a project manually.
func initManual(cmd *cobra.Command, args []string, ctx *sow.Context, description, branchName string) error {
	if len(args) == 0 {
		return fmt.Errorf("project name is required when not using --issue")
	}

	if description == "" {
		return fmt.Errorf("--description is required when not using --issue")
	}

	if branchName != "" {
		return fmt.Errorf("--branch-name can only be used with --issue")
	}

	// Validate not on protected branch
	currentBranch, err := ctx.Git().CurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	if ctx.Git().IsProtectedBranch(currentBranch) {
		return fmt.Errorf("cannot create project on protected branch '%s' - use a feature branch", currentBranch)
	}

	name := args[0]
	proj, err := loader.Create(ctx, name, description)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(cmd.OutOrStderr(), "\n✓ Initialized project '%s' on branch '%s'\n",
		proj.Name(), proj.Branch())
	return nil
}
