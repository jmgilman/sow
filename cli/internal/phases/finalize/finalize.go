// Package finalize implements the Finalize phase of the project lifecycle.
//
// The Finalize phase has three stages:
// 1. Documentation - Update/move documentation
// 2. Checks - Run tests, linters, validation
// 3. Delete - Delete project folder and loop back to NoProject
package finalize

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"text/template"

	"github.com/jmgilman/sow/cli/internal/phases"
	"github.com/jmgilman/sow/cli/internal/project/statechart"
	phasesSchema "github.com/jmgilman/sow/cli/schemas/phases"
	"github.com/qmuntal/stateless"
)

//go:embed templates/*.md
var templates embed.FS

// FinalizePhase implements the Phase interface for the finalize phase.
type FinalizePhase struct {
	data    *phasesSchema.FinalizePhase // Phase data from project state
	project phases.ProjectInfo          // Minimal project info for templates
}

// New creates a new Finalize phase instance.
//
// Parameters:
//   - data: Pointer to the FinalizePhase data from project state
//   - project: Basic project information for template rendering
func New(data *phasesSchema.FinalizePhase, project phases.ProjectInfo) *FinalizePhase {
	return &FinalizePhase{
		data:    data,
		project: project,
	}
}

// EntryState returns the state where this phase begins (FinalizeDocumentation).
func (p *FinalizePhase) EntryState() statechart.State {
	return statechart.FinalizeDocumentation
}

// AddToMachine configures the finalize phase states in the state machine.
//
// The finalize phase has three states:
// 1. FinalizeDocumentation - Update docs, move artifacts
// 2. FinalizeChecks - Run tests, linters, final validation
// 3. FinalizeDelete - Delete project folder
//
// Transitions:
// - FinalizeDocumentation → FinalizeChecks (EventDocumentationDone, guard: always true)
// - FinalizeChecks → FinalizeDelete (EventChecksDone, guard: always true)
// - FinalizeDelete → nextPhaseEntry (EventProjectDelete, guard: project_deleted flag)
func (p *FinalizePhase) AddToMachine(sm *stateless.StateMachine, nextPhaseEntry statechart.State) {
	// Configure documentation state
	sm.Configure(statechart.FinalizeDocumentation).
		Permit(statechart.EventDocumentationDone, statechart.FinalizeChecks, p.documentationAssessedGuard).
		OnEntry(p.onDocumentationEntry)

	// Configure checks state
	sm.Configure(statechart.FinalizeChecks).
		Permit(statechart.EventChecksDone, statechart.FinalizeDelete, p.checksAssessedGuard).
		OnEntry(p.onChecksEntry)

	// Configure delete state
	sm.Configure(statechart.FinalizeDelete).
		Permit(statechart.EventProjectDelete, nextPhaseEntry, p.projectDeletedGuard).
		OnEntry(p.onDeleteEntry)
}

// Metadata returns phase metadata for CLI validation and introspection.
func (p *FinalizePhase) Metadata() phases.PhaseMetadata {
	return phases.PhaseMetadata{
		Name: "finalize",
		States: []statechart.State{
			statechart.FinalizeDocumentation,
			statechart.FinalizeChecks,
			statechart.FinalizeDelete,
		},
		SupportsTasks:     false,
		SupportsArtifacts: false,
		CustomFields: []phases.FieldDef{
			{
				Name:        "project_deleted",
				Type:        phases.BoolField,
				Description: "Critical gate: must be true before phase completion",
			},
			{
				Name:        "pr_url",
				Type:        phases.StringField,
				Description: "Pull request URL created during finalization",
			},
		},
	}
}

// Entry Actions

// onDocumentationEntry renders and displays the documentation prompt.
func (p *FinalizePhase) onDocumentationEntry(_ context.Context, _ ...any) error {
	prompt := p.renderPrompt("documentation")
	fmt.Println(prompt)
	return nil
}

// onChecksEntry renders and displays the checks prompt.
func (p *FinalizePhase) onChecksEntry(_ context.Context, _ ...any) error {
	prompt := p.renderPrompt("checks")
	fmt.Println(prompt)
	return nil
}

// onDeleteEntry renders and displays the delete prompt.
func (p *FinalizePhase) onDeleteEntry(_ context.Context, _ ...any) error {
	prompt := p.renderPrompt("delete")
	fmt.Println(prompt)
	return nil
}

// Guards

// documentationAssessedGuard checks if documentation work has been assessed.
// For now, this always returns true - the act of calling the command IS the signal.
func (p *FinalizePhase) documentationAssessedGuard(_ context.Context, _ ...any) bool {
	// Always allow transition - the command itself is the validation
	return true
}

// checksAssessedGuard checks if final checks have been assessed and handled.
// For now, this always returns true - checks are considered assessed if documentation is done.
func (p *FinalizePhase) checksAssessedGuard(_ context.Context, _ ...any) bool {
	// Always allow transition - the command itself is the validation
	return true
}

// projectDeletedGuard checks if the project folder has been deleted.
func (p *FinalizePhase) projectDeletedGuard(_ context.Context, _ ...any) bool {
	if p.data == nil {
		return false
	}
	return p.data.Project_deleted
}

// Template Rendering

// renderPrompt loads and renders a template with phase data.
func (p *FinalizePhase) renderPrompt(name string) string {
	// Load template
	content, err := templates.ReadFile("templates/" + name + ".md")
	if err != nil {
		return fmt.Sprintf("Error loading template %s: %v", name, err)
	}

	// Parse template
	tmpl, err := template.New(name).Parse(string(content))
	if err != nil {
		return fmt.Sprintf("Error parsing template %s: %v", name, err)
	}

	// Prepare template data
	data := p.prepareTemplateData()

	// Render template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Sprintf("Error rendering template %s: %v", name, err)
	}

	return buf.String()
}

// prepareTemplateData creates a map of data for template rendering.
func (p *FinalizePhase) prepareTemplateData() map[string]interface{} {
	data := make(map[string]interface{})

	// Project info (always available)
	data["ProjectName"] = p.project.Name
	data["ProjectDescription"] = p.project.Description
	data["ProjectBranch"] = p.project.Branch

	// Phase data (if available)
	if p.data != nil {
		// Documentation updates
		if len(p.data.Documentation_updates) > 0 {
			data["HasDocumentationUpdates"] = true
			data["DocumentationUpdates"] = p.data.Documentation_updates
		} else {
			data["HasDocumentationUpdates"] = false
		}

		// Artifacts moved
		if len(p.data.Artifacts_moved) > 0 {
			data["HasArtifactsMoved"] = true
			data["ArtifactsMoved"] = p.data.Artifacts_moved
		} else {
			data["HasArtifactsMoved"] = false
		}

		// PR URL
		if p.data.Pr_url != nil && *p.data.Pr_url != "" {
			data["HasPRUrl"] = true
			data["PRUrl"] = *p.data.Pr_url
		} else {
			data["HasPRUrl"] = false
		}

		// Project deleted flag
		data["ProjectDeleted"] = p.data.Project_deleted
	}

	return data
}
