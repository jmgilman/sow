// Package refs implements commands for managing external knowledge and code references.
package refs

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/refs"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewRefsCmd creates the refs command with unified subcommands.
func NewRefsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refs",
		Short: "Manage external knowledge and code references",
		Long: `Manage external references to knowledge and code repositories.

Refs provide access to external documentation, style guides, and code
examples. They are cached locally and symlinked into .sow/refs/ for
easy access by AI agents.

The type of reference is automatically inferred from the URL scheme:
  git+https://github.com/org/repo → git type
  git+ssh://git@github.com/org/repo → git type
  git@github.com:org/repo           → git type (SSH shorthand)
  file:///absolute/path             → file type

Commands:
  add     - Add a new reference
  update  - Update existing references
  remove  - Remove a reference
  list    - List configured references
  status  - Check reference staleness
  init    - Initialize refs after cloning`,
	}

	// Unified subcommands
	cmd.AddCommand(newAddCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newRemoveCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newInitCmd())

	return cmd
}

// refWithSource associates a ref with its source (committed or local).
type refWithSource struct {
	Ref    schemas.Ref
	Source string // "committed" or "local"
}

// printRefsTable prints refs in table format.
func printRefsTable(cmd *cobra.Command, refsList []refWithSource) {
	// Header
	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintf(out, "%-20s %-6s %-10s %-15s %-10s %s\n",
		"ID", "TYPE", "SEMANTIC", "LINK", "SOURCE", "URL")
	_, _ = fmt.Fprintf(out, "%-20s %-6s %-10s %-15s %-10s %s\n",
		strings.Repeat("─", 20),
		strings.Repeat("─", 6),
		strings.Repeat("─", 10),
		strings.Repeat("─", 15),
		strings.Repeat("─", 10),
		strings.Repeat("─", 30),
	)

	// Rows
	for _, rws := range refsList {
		ref := rws.Ref

		// Infer type
		refType, err := refs.InferTypeFromURL(ref.Source)
		if err != nil {
			refType = "unknown"
		}

		// Strip scheme for display
		displayURL := ref.Source
		displayURL = strings.TrimPrefix(displayURL, "git+")
		displayURL = strings.TrimPrefix(displayURL, "file://")

		_, _ = fmt.Fprintf(out, "%-20s %-6s %-10s %-15s %-10s %s\n",
			ref.Id, refType, ref.Semantic, ref.Link, rws.Source, displayURL)

		// Show additional details
		if ref.Config.Branch != "" {
			_, _ = fmt.Fprintf(out, "  └─ branch: %s\n", ref.Config.Branch)
		}
		if ref.Config.Path != "" {
			_, _ = fmt.Fprintf(out, "  └─ path: %s\n", ref.Config.Path)
		}
	}
}

// printRefsJSON prints refs in JSON format.
func printRefsJSON(cmd *cobra.Command, refsList []refWithSource) error {
	data, err := json.MarshalIndent(refsList, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	cmd.Println(string(data))
	return nil
}

// printRefsYAML prints refs in YAML format.
func printRefsYAML(cmd *cobra.Command, refsList []refWithSource) error {
	data, err := yaml.Marshal(refsList)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}
	cmd.Print(string(data))
	return nil
}
