// Package review implements the Review phase of the project lifecycle.
//
// The Review phase validates implementation work and can loop back to
// implementation if issues are found. It operates in "autonomous mode".
package review

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

// ReviewPhase implements the Phase interface for the review phase.
type ReviewPhase struct {
	data    *phasesSchema.ReviewPhase // Phase data from project state
	project phases.ProjectInfo        // Minimal project info for templates
}

// New creates a new Review phase instance.
//
// Parameters:
//   - data: Pointer to the ReviewPhase data from project state
//   - project: Basic project information for template rendering
func New(data *phasesSchema.ReviewPhase, project phases.ProjectInfo) *ReviewPhase {
	return &ReviewPhase{
		data:    data,
		project: project,
	}
}

// EntryState returns the state where this phase begins (ReviewActive).
func (p *ReviewPhase) EntryState() statechart.State {
	return statechart.ReviewActive
}

// AddToMachine configures the review phase states in the state machine.
//
// The review phase has one state:
// - ReviewActive - Perform review, create report
//
// Transitions:
// - ReviewActive → nextPhaseEntry (EventReviewPass, guard: latest review approved)
//
// Note: The backward transition (EventReviewFail → ImplementationPlanning) is NOT
// configured here - it's added by the project type as an exceptional transition.
func (p *ReviewPhase) AddToMachine(sm *stateless.StateMachine, nextPhaseEntry statechart.State) {
	sm.Configure(statechart.ReviewActive).
		Permit(statechart.EventReviewPass, nextPhaseEntry, p.latestReviewApprovedGuard).
		OnEntry(p.onActiveEntry)
}

// Metadata returns phase metadata for CLI validation and introspection.
func (p *ReviewPhase) Metadata() phases.PhaseMetadata {
	return phases.PhaseMetadata{
		Name:              "review",
		States:            []statechart.State{statechart.ReviewActive},
		SupportsTasks:     false,
		SupportsArtifacts: false,
		CustomFields: []phases.FieldDef{
			{
				Name:        "iteration",
				Type:        phases.IntField,
				Description: "Current review iteration (increments on loop-back)",
			},
		},
	}
}

// Entry Actions

// onActiveEntry renders and displays the active phase prompt.
func (p *ReviewPhase) onActiveEntry(_ context.Context, _ ...any) error {
	prompt := p.renderPrompt("active")
	fmt.Println(prompt)
	return nil
}

// Guards

// latestReviewApprovedGuard checks if the most recent review report is approved.
// This guard is used for the forward transition (EventReviewPass).
// The backward transition guard (EventReviewFail) is configured by the project type.
func (p *ReviewPhase) latestReviewApprovedGuard(_ context.Context, _ ...any) bool {
	if p.data == nil {
		return false
	}

	reports := p.data.Reports
	if len(reports) == 0 {
		return false
	}

	latest := reports[len(reports)-1]
	return latest.Approved
}

// LatestReviewFailedGuard checks if the most recent review report indicates failure.
// This is exported for use by project types to configure the backward transition.
func (p *ReviewPhase) LatestReviewFailedGuard(_ context.Context, _ ...any) bool {
	if p.data == nil {
		return false
	}

	reports := p.data.Reports
	if len(reports) == 0 {
		return false
	}

	latest := reports[len(reports)-1]
	// A report is considered failed if it's not approved and assessment is "fail"
	return !latest.Approved && latest.Assessment == "fail"
}

// Template Rendering

// renderPrompt loads and renders a template with phase data.
func (p *ReviewPhase) renderPrompt(name string) string {
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
func (p *ReviewPhase) prepareTemplateData() map[string]interface{} {
	data := make(map[string]interface{})

	// Project info (always available)
	data["ProjectName"] = p.project.Name
	data["ProjectDescription"] = p.project.Description
	data["ProjectBranch"] = p.project.Branch

	// Phase data (if available)
	if p.data != nil {
		iteration := p.data.Iteration
		if iteration == 0 {
			iteration = 1 // Default to 1 if not set
		}
		data["ReviewIteration"] = iteration

		// Previous iteration context
		if iteration > 1 && len(p.data.Reports) > 0 {
			data["HasPreviousReview"] = true
			prevReport := p.data.Reports[len(p.data.Reports)-1]
			data["PreviousAssessment"] = prevReport.Assessment
		} else {
			data["HasPreviousReview"] = false
		}

		// Report count
		data["ReportCount"] = len(p.data.Reports)
	}

	return data
}
