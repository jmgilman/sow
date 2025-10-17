// Package refs implements commands for managing external knowledge and code references.
package refs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jmgilman/sow/cli/internal/refs"
	"github.com/jmgilman/sow/cli/internal/sowfs"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewRefsCmd creates the refs command with unified subcommands.
func NewRefsCmd(sowFSFromContext func(context.Context) sowfs.SowFS) *cobra.Command {
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
	cmd.AddCommand(newAddCmd(sowFSFromContext))
	cmd.AddCommand(newUpdateCmd(sowFSFromContext))
	cmd.AddCommand(newRemoveCmd(sowFSFromContext))
	cmd.AddCommand(newListCmd(sowFSFromContext))
	cmd.AddCommand(newStatusCmd(sowFSFromContext))
	cmd.AddCommand(newInitCmd(sowFSFromContext))

	return cmd
}

// refWithSource associates a ref with its source (committed or local).
type refWithSource struct {
	Ref    schemas.Ref
	Source string // "committed" or "local"
}

// generateIDFromURL creates a ref ID from a URL and type.
func generateIDFromURL(url, typeName string) string {
	switch typeName {
	case "git":
		// Remove scheme prefix
		url = strings.TrimPrefix(url, "git+https://")
		url = strings.TrimPrefix(url, "git+ssh://")
		url = strings.TrimPrefix(url, "git+http://")
		url = strings.TrimPrefix(url, "git@")

		// Remove domain (take last 2 path components)
		parts := strings.Split(url, "/")
		if len(parts) >= 2 {
			url = strings.Join(parts[len(parts)-2:], "-")
		}

		// Remove .git suffix
		url = strings.TrimSuffix(url, ".git")

	case "file":
		// Get base directory name
		url = strings.TrimPrefix(url, "file://")
		parts := strings.Split(strings.TrimSuffix(url, "/"), "/")
		if len(parts) > 0 {
			url = parts[len(parts)-1]
		}
	}

	// Convert to kebab-case
	url = strings.ToLower(url)
	url = strings.ReplaceAll(url, "/", "-")
	url = strings.ReplaceAll(url, "_", "-")
	url = strings.ReplaceAll(url, ":", "-")

	return url
}

// filterRefs filters refs based on type, semantic, and tags.
func filterRefs(_ context.Context, refsList []refWithSource, typeFilter, semanticFilter string, tagsFilter []string) []refWithSource {
	var filtered []refWithSource

	for _, rws := range refsList {
		ref := rws.Ref

		// Filter by type
		if typeFilter != "" {
			refType, err := refs.InferTypeFromURL(ref.Source)
			if err != nil || refType != typeFilter {
				continue
			}
		}

		// Filter by semantic
		if semanticFilter != "" && ref.Semantic != semanticFilter {
			continue
		}

		// Filter by tags
		if len(tagsFilter) > 0 {
			hasAllTags := true
			for _, filterTag := range tagsFilter {
				found := false
				for _, refTag := range ref.Tags {
					if refTag == filterTag {
						found = true
						break
					}
				}
				if !found {
					hasAllTags = false
					break
				}
			}
			if !hasAllTags {
				continue
			}
		}

		filtered = append(filtered, rws)
	}

	return filtered
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
