// Package discovery implements the Discovery phase of the project lifecycle.
//
// The Discovery phase is an optional phase that helps gather context and research
// before proceeding to design or implementation. It operates in "subservient mode",
// where the orchestrator acts as an assistant to the human.
package discovery

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

// DiscoveryPhase implements the Phase interface for the discovery phase.
//
//nolint:revive // DiscoveryPhase naming is intentional for clarity in phase package
type DiscoveryPhase struct {
	optional bool                         // Whether this phase can be skipped
	data     *phasesSchema.DiscoveryPhase // Phase data from project state
	project  phases.ProjectInfo           // Minimal project info for templates
}

// New creates a new Discovery phase instance.
//
// Parameters:
//   - optional: If true, the phase can be skipped via EventSkipDiscovery
//   - data: Pointer to the DiscoveryPhase data from project state
//   - project: Basic project information for template rendering
func New(optional bool, data *phasesSchema.DiscoveryPhase, project phases.ProjectInfo) *DiscoveryPhase {
	return &DiscoveryPhase{
		optional: optional,
		data:     data,
		project:  project,
	}
}

// EntryState returns the state where this phase begins (DiscoveryDecision).
func (p *DiscoveryPhase) EntryState() statechart.State {
	return statechart.DiscoveryDecision
}

// AddToMachine configures the discovery phase states in the state machine.
//
// The discovery phase has two states:
// 1. DiscoveryDecision - Decide whether to enable or skip discovery
// 2. DiscoveryActive - Perform discovery work, gather artifacts
//
// Transitions:
// - DiscoveryDecision → DiscoveryActive (EventEnableDiscovery)
// - DiscoveryDecision → nextPhaseEntry (EventSkipDiscovery, if optional).
// - DiscoveryActive → nextPhaseEntry (EventCompleteDiscovery, guard: artifacts approved).
func (p *DiscoveryPhase) AddToMachine(sm *stateless.StateMachine, nextPhaseEntry statechart.State) {
	// Configure decision state
	decisionConfig := sm.Configure(statechart.DiscoveryDecision).
		Permit(statechart.EventEnableDiscovery, statechart.DiscoveryActive).
		OnEntry(p.onDecisionEntry)

	// If optional, allow skipping
	if p.optional {
		decisionConfig.Permit(statechart.EventSkipDiscovery, nextPhaseEntry)
	}

	// Configure active state
	sm.Configure(statechart.DiscoveryActive).
		Permit(statechart.EventCompleteDiscovery, nextPhaseEntry, p.artifactsApprovedGuard).
		OnEntry(p.onActiveEntry)
}

// Metadata returns phase metadata for CLI validation and introspection.
func (p *DiscoveryPhase) Metadata() phases.PhaseMetadata {
	return phases.PhaseMetadata{
		Name:              "discovery",
		States:            []statechart.State{statechart.DiscoveryDecision, statechart.DiscoveryActive},
		SupportsTasks:     false,
		SupportsArtifacts: true,
		CustomFields: []phases.FieldDef{
			{
				Name:        "discovery_type",
				Type:        phases.StringField,
				Description: "Type of discovery work (bug, feature, docs, refactor, general)",
			},
		},
	}
}

// Entry Actions

// onDecisionEntry renders and displays the decision prompt.
func (p *DiscoveryPhase) onDecisionEntry(_ context.Context, _ ...any) error {
	prompt := p.renderPrompt("decision")
	fmt.Println(prompt)
	return nil
}

// onActiveEntry renders and displays the active phase prompt.
func (p *DiscoveryPhase) onActiveEntry(_ context.Context, _ ...any) error {
	prompt := p.renderPrompt("active")
	fmt.Println(prompt)
	return nil
}

// Guards

// artifactsApprovedGuard checks if all artifacts are approved (or no artifacts exist).
func (p *DiscoveryPhase) artifactsApprovedGuard(_ context.Context, _ ...any) bool {
	if p.data == nil {
		return false
	}

	// No artifacts is fine - discovery can complete without creating any
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
func (p *DiscoveryPhase) renderPrompt(name string) string {
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
// This includes both phase-specific data and project information.
func (p *DiscoveryPhase) prepareTemplateData() map[string]interface{} {
	data := make(map[string]interface{})

	// Project info (always available)
	data["ProjectName"] = p.project.Name
	data["ProjectDescription"] = p.project.Description
	data["ProjectBranch"] = p.project.Branch

	// Phase data (if available)
	if p.data != nil {
		// Discovery type (optional field)
		if p.data.Discovery_type != nil {
			data["DiscoveryType"] = *p.data.Discovery_type
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
