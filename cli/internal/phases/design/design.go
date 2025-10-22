// Package design implements the Design phase of the project lifecycle.
//
// The Design phase is an optional phase that helps create architectural decisions
// and design artifacts before implementation. It operates in "subservient mode",
// where the orchestrator acts as an assistant to the human.
package design

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"text/template"

	"github.com/jmgilman/sow/cli/internal/phases"
	phasesSchema "github.com/jmgilman/sow/cli/schemas/phases"
	"github.com/qmuntal/stateless"
)

//go:embed templates/*.md
var templates embed.FS

// DesignPhase implements the Phase interface for the design phase.
type DesignPhase struct {
	optional bool                     // Whether this phase can be skipped
	data     *phasesSchema.DesignPhase // Phase data from project state
	project  ProjectInfo              // Minimal project info for templates
}

// ProjectInfo holds minimal project information needed for template rendering.
type ProjectInfo struct {
	Name        string
	Description string
	Branch      string
}

// New creates a new Design phase instance.
//
// Parameters:
//   - optional: If true, the phase can be skipped via EventSkipDesign
//   - data: Pointer to the DesignPhase data from project state
//   - project: Basic project information for template rendering
func New(optional bool, data *phasesSchema.DesignPhase, project ProjectInfo) *DesignPhase {
	return &DesignPhase{
		optional: optional,
		data:     data,
		project:  project,
	}
}

// EntryState returns the state where this phase begins (DesignDecision).
func (p *DesignPhase) EntryState() phases.State {
	return phases.DesignDecision
}

// AddToMachine configures the design phase states in the state machine.
//
// The design phase has two states:
// 1. DesignDecision - Decide whether to enable or skip design
// 2. DesignActive - Perform design work, create artifacts (ADRs, design docs)
//
// Transitions:
// - DesignDecision → DesignActive (EventEnableDesign)
// - DesignDecision → nextPhaseEntry (EventSkipDesign, if optional)
// - DesignActive → nextPhaseEntry (EventCompleteDesign, guard: artifacts approved)
func (p *DesignPhase) AddToMachine(sm *stateless.StateMachine, nextPhaseEntry phases.State) {
	// Configure decision state
	decisionConfig := sm.Configure(phases.DesignDecision).
		Permit(phases.EventEnableDesign, phases.DesignActive).
		OnEntry(p.onDecisionEntry)

	// If optional, allow skipping
	if p.optional {
		decisionConfig.Permit(phases.EventSkipDesign, nextPhaseEntry)
	}

	// Configure active state
	sm.Configure(phases.DesignActive).
		Permit(phases.EventCompleteDesign, nextPhaseEntry, p.artifactsApprovedGuard).
		OnEntry(p.onActiveEntry)
}

// Metadata returns phase metadata for CLI validation and introspection.
func (p *DesignPhase) Metadata() phases.PhaseMetadata {
	return phases.PhaseMetadata{
		Name:              "design",
		States:            []phases.State{phases.DesignDecision, phases.DesignActive},
		SupportsTasks:     false,
		SupportsArtifacts: true,
		CustomFields: []phases.FieldDef{
			{
				Name:        "architect_used",
				Type:        phases.BoolField,
				Description: "Whether the architect agent was used for this phase",
			},
		},
	}
}

// Entry Actions

// onDecisionEntry renders and displays the decision prompt.
func (p *DesignPhase) onDecisionEntry(_ context.Context, _ ...any) error {
	prompt := p.renderPrompt("decision")
	fmt.Println(prompt)
	return nil
}

// onActiveEntry renders and displays the active phase prompt.
func (p *DesignPhase) onActiveEntry(_ context.Context, _ ...any) error {
	prompt := p.renderPrompt("active")
	fmt.Println(prompt)
	return nil
}

// Guards

// artifactsApprovedGuard checks if all artifacts are approved (or no artifacts exist).
func (p *DesignPhase) artifactsApprovedGuard(_ context.Context, _ ...any) bool {
	if p.data == nil {
		return false
	}

	// No artifacts is fine - design can complete without creating any
	if len(p.data.Artifacts) == 0 {
		return true
	}

	// All artifacts must be approved
	for _, artifact := range p.data.Artifacts {
		if !artifact.Approved {
			return false
		}
	}

	return true
}

// Template Rendering

// renderPrompt loads and renders a template with phase data.
func (p *DesignPhase) renderPrompt(name string) string {
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
func (p *DesignPhase) prepareTemplateData() map[string]interface{} {
	data := make(map[string]interface{})

	// Project info (always available)
	data["ProjectName"] = p.project.Name
	data["ProjectDescription"] = p.project.Description
	data["ProjectBranch"] = p.project.Branch

	// Phase data (if available)
	if p.data != nil {
		// Architect used (optional field)
		if p.data.Architect_used != nil {
			data["ArchitectUsed"] = *p.data.Architect_used
		}

		// Artifact counts
		data["ArtifactCount"] = len(p.data.Artifacts)

		approvedCount := 0
		for _, a := range p.data.Artifacts {
			if a.Approved {
				approvedCount++
			}
		}
		data["ApprovedCount"] = approvedCount
	}

	return data
}
