// Package templates provides utilities for rendering prompt templates
// with project state context. It wraps Go's text/template with
// project-specific helper functions.
package templates

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	projschema "github.com/jmgilman/sow/cli/schemas/project"
)

// Render renders a template from an embedded filesystem with project state as context.
// The path is relative to the embedded filesystem root.
//
// Templates have access to all fields and methods of *state.Project, plus helper
// functions from DefaultFuncMap() for common operations like phase lookup and
// artifact filtering.
//
// Example usage:
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
// Example template:
//
//	# Project: {{.Name}}
//	Branch: {{.Branch}}
//
//	{{$planning := phase . "planning"}}
//	{{if hasApprovedOutput $planning "task_list"}}
//	✓ Task list approved
//	{{end}}
func Render(fs embed.FS, path string, project *state.Project) (string, error) {
	// Read template from embedded filesystem
	content, err := fs.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read template %s: %w", path, err)
	}

	// Parse template with helper functions
	tmpl, err := template.New(path).Funcs(DefaultFuncMap()).Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", path, err)
	}

	// Execute template with project as data
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, project); err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", path, err)
	}

	return buf.String(), nil
}

// DefaultFuncMap returns the standard template functions available to all templates.
// These helpers simplify common operations on project state.
func DefaultFuncMap() template.FuncMap {
	return template.FuncMap{
		// String operations
		"join": func(sep string, elems []string) string {
			return strings.Join(elems, sep)
		},

		// Phase lookup helper
		// Usage: {{$planning := phase . "planning"}}
		"phase": func(p *state.Project, name string) (map[string]projschema.PhaseState, error) {
			phase, exists := p.Phases[name]
			if !exists {
				return nil, fmt.Errorf("phase %s not found", name)
			}
			// Return as map for template compatibility
			return map[string]projschema.PhaseState{name: phase}, nil
		},

		// Count outputs of a specific type in a phase
		// Usage: {{countOutputs $planning "task_list"}}
		"countOutputs": func(phases map[string]projschema.PhaseState, phaseName string, outputType string) int {
			phase, exists := phases[phaseName]
			if !exists {
				return 0
			}
			count := 0
			for _, output := range phase.Outputs {
				if output.Type == outputType {
					count++
				}
			}
			return count
		},

		// Check if a phase has an approved output of a specific type
		// Usage: {{if hasApprovedOutput $planning "planning" "task_list"}}...{{end}}
		"hasApprovedOutput": func(phases map[string]projschema.PhaseState, phaseName string, outputType string) bool {
			phase, exists := phases[phaseName]
			if !exists {
				return false
			}
			for _, output := range phase.Outputs {
				if output.Type == outputType && output.Approved {
					return true
				}
			}
			return false
		},

		// Count tasks by status
		// Usage: {{countTasksByStatus $impl "implementation" "completed"}}
		"countTasksByStatus": func(phases map[string]projschema.PhaseState, phaseName string, status string) int {
			phase, exists := phases[phaseName]
			if !exists {
				return 0
			}
			count := 0
			for _, task := range phase.Tasks {
				if task.Status == status {
					count++
				}
			}
			return count
		},

		// Get metadata value from phase
		// Usage: {{phaseMetadata $impl "implementation" "tasks_approved"}}
		"phaseMetadata": func(phases map[string]projschema.PhaseState, phaseName string, key string) interface{} {
			phase, exists := phases[phaseName]
			if !exists || phase.Metadata == nil {
				return nil
			}
			return phase.Metadata[key]
		},
	}
}
