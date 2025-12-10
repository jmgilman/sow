// Package templates provides utilities for rendering templates from embedded filesystems.
// It provides a unified interface for rendering both project-specific templates (with project state)
// and general-purpose templates (guidance, commands, etc.).
package templates

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/jmgilman/sow/libs/project/state"
	projschema "github.com/jmgilman/sow/libs/schemas/project"
)

// Render renders a template from an embedded filesystem with optional data context.
// The path is relative to the embedded filesystem root.
//
// Data can be:
//   - *state.Project for project-specific templates (enables project helper functions)
//   - nil for simple templates that don't need context
//   - any other type for custom contexts
//
// Example usage with project state:
//
//	//go:embed templates/*.md
//	var templatesFS embed.FS
//
//	func generatePrompt(p *state.Project) string {
//	    output, err := templates.Render(templatesFS, "templates/planning.md", p)
//	    if err != nil {
//	        return fmt.Sprintf("Error: %v", err)
//	    }
//	    return output
//	}
//
// Example usage without context:
//
//	output, err := templates.Render(promptsFS, "templates/guidance/research.md", nil)
//
// Example template with project state:
//
//	# Project: {{.Name}}
//	Branch: {{.Branch}}
//
//	{{$planning := phase . "planning"}}
//	{{if hasApprovedOutput $planning "task_list"}}
//	✓ Task list approved
//	{{end}}
func Render(embedFS embed.FS, path string, data interface{}) (string, error) {
	// Read template from embedded filesystem
	content, err := embedFS.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read template %s: %w", path, err)
	}

	// Parse template with helper functions
	tmpl, err := template.New(path).Funcs(DefaultFuncMap()).Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", path, err)
	}

	// Handle nil data by providing empty context
	ctx := data
	if ctx == nil {
		ctx = make(map[string]interface{})
	}

	// Execute template with data
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", path, err)
	}

	return buf.String(), nil
}

// ListTemplates lists all .md templates in the given directory of the embedded filesystem.
// Returns paths relative to baseDir with .md suffix removed.
//
// This is useful for discovering available templates at runtime, such as for --list flags
// or error messages.
//
// Example:
//
//	templates/guidance/research.md -> guidance/research
//	templates/design/adr.md -> design/adr
func ListTemplates(embedFS embed.FS, baseDir string) ([]string, error) {
	var templates []string

	err := fs.WalkDir(embedFS, baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only include .md files
		if filepath.Ext(path) != ".md" {
			return nil
		}

		// Remove baseDir prefix and .md suffix
		// "templates/guidance/research.md" -> "guidance/research"
		relPath := strings.TrimPrefix(path, baseDir+"/")
		relPath = strings.TrimSuffix(relPath, ".md")

		templates = append(templates, relPath)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", baseDir, err)
	}

	// Sort for consistent output
	sort.Strings(templates)

	return templates, nil
}

// DefaultFuncMap returns the standard template functions available to all templates.
// These helpers simplify common operations, especially on project state.
//
// Note: Project-specific helpers (phase, hasApprovedOutput, etc.) require the data
// to be *state.Project. They will return errors if used with other data types.
func DefaultFuncMap() template.FuncMap {
	return template.FuncMap{
		// String operations
		"join": func(sep string, elems []string) string {
			return strings.Join(elems, sep)
		},

		// Phase lookup helper (requires *state.Project)
		// Usage: {{$planning := phase . "planning"}}
		"phase": func(p *state.Project, name string) (projschema.PhaseState, error) {
			phase, exists := p.Phases[name]
			if !exists {
				return projschema.PhaseState{}, fmt.Errorf("phase %s not found", name)
			}
			return phase, nil
		},

		// Count outputs of a specific type in a phase
		// Usage: {{countOutputs $planning "task_list"}}
		"countOutputs": func(phase projschema.PhaseState, outputType string) int {
			count := 0
			for _, output := range phase.Outputs {
				if output.Type == outputType {
					count++
				}
			}
			return count
		},

		// Check if a phase has an approved output of a specific type
		// Usage: {{if hasApprovedOutput $planning "task_list"}}...{{end}}
		"hasApprovedOutput": func(phase projschema.PhaseState, outputType string) bool {
			for _, output := range phase.Outputs {
				if output.Type == outputType && output.Approved {
					return true
				}
			}
			return false
		},

		// Count tasks by status
		// Usage: {{countTasksByStatus $impl "completed"}}
		"countTasksByStatus": func(phase projschema.PhaseState, status string) int {
			count := 0
			for _, task := range phase.Tasks {
				if task.Status == status {
					count++
				}
			}
			return count
		},

		// Get metadata value from phase
		// Usage: {{phaseMetadata $impl "tasks_approved"}}
		"phaseMetadata": func(phase projschema.PhaseState, key string) interface{} {
			if phase.Metadata == nil {
				return nil
			}
			return phase.Metadata[key]
		},
	}
}
