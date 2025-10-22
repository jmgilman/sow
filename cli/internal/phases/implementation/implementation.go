// Package implementation implements the Implementation phase of the project lifecycle.
//
// The Implementation phase is a required phase with dual states:
// 1. Planning - Create and approve task breakdown
// 2. Executing - Execute tasks autonomously
package implementation

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

// ImplementationPhase implements the Phase interface for the implementation phase.
type ImplementationPhase struct {
	data    *phasesSchema.ImplementationPhase // Phase data from project state
	project ProjectInfo                       // Minimal project info for templates
}

// ProjectInfo holds minimal project information needed for template rendering.
type ProjectInfo struct {
	Name        string
	Description string
	Branch      string
}

// New creates a new Implementation phase instance.
//
// Parameters:
//   - data: Pointer to the ImplementationPhase data from project state
//   - project: Basic project information for template rendering
func New(data *phasesSchema.ImplementationPhase, project ProjectInfo) *ImplementationPhase {
	return &ImplementationPhase{
		data:    data,
		project: project,
	}
}

// EntryState returns the state where this phase begins (ImplementationPlanning).
func (p *ImplementationPhase) EntryState() phases.State {
	return phases.ImplementationPlanning
}

// AddToMachine configures the implementation phase states in the state machine.
//
// The implementation phase has two states:
// 1. ImplementationPlanning - Create task breakdown, get human approval
// 2. ImplementationExecuting - Execute approved tasks autonomously
//
// Transitions:
// - ImplementationPlanning → ImplementationExecuting (EventTaskCreated, guard: has at least one task)
// - ImplementationPlanning → ImplementationExecuting (EventTasksApproved, guard: tasks approved)
// - ImplementationExecuting → nextPhaseEntry (EventAllTasksComplete, guard: all tasks complete)
func (p *ImplementationPhase) AddToMachine(sm *stateless.StateMachine, nextPhaseEntry phases.State) {
	// Configure planning state
	sm.Configure(phases.ImplementationPlanning).
		Permit(phases.EventTaskCreated, phases.ImplementationExecuting, p.hasAtLeastOneTaskGuard).
		Permit(phases.EventTasksApproved, phases.ImplementationExecuting, p.tasksApprovedGuard).
		OnEntry(p.onPlanningEntry)

	// Configure executing state
	sm.Configure(phases.ImplementationExecuting).
		Permit(phases.EventAllTasksComplete, nextPhaseEntry, p.allTasksCompleteGuard).
		OnEntry(p.onExecutingEntry)
}

// Metadata returns phase metadata for CLI validation and introspection.
func (p *ImplementationPhase) Metadata() phases.PhaseMetadata {
	return phases.PhaseMetadata{
		Name:          "implementation",
		States:        []phases.State{phases.ImplementationPlanning, phases.ImplementationExecuting},
		SupportsTasks: true,
		SupportsArtifacts: false,
		CustomFields: []phases.FieldDef{
			{
				Name:        "planner_used",
				Type:        phases.BoolField,
				Description: "Whether the planner agent was used for task breakdown",
			},
			{
				Name:        "tasks_approved",
				Type:        phases.BoolField,
				Description: "Whether the task plan has been approved by human",
			},
		},
	}
}

// Entry Actions

// onPlanningEntry renders and displays the planning prompt.
func (p *ImplementationPhase) onPlanningEntry(_ context.Context, _ ...any) error {
	prompt := p.renderPrompt("planning")
	fmt.Println(prompt)
	return nil
}

// onExecutingEntry renders and displays the executing prompt.
func (p *ImplementationPhase) onExecutingEntry(_ context.Context, _ ...any) error {
	prompt := p.renderPrompt("executing")
	fmt.Println(prompt)
	return nil
}

// Guards

// hasAtLeastOneTaskGuard checks if at least one task has been created.
func (p *ImplementationPhase) hasAtLeastOneTaskGuard(_ context.Context, _ ...any) bool {
	if p.data == nil {
		return false
	}
	return len(p.data.Tasks) >= 1
}

// tasksApprovedGuard checks if task plan has been approved by human.
func (p *ImplementationPhase) tasksApprovedGuard(_ context.Context, _ ...any) bool {
	if p.data == nil {
		return false
	}
	return p.data.Tasks_approved && len(p.data.Tasks) >= 1
}

// allTasksCompleteGuard checks if all tasks are completed or abandoned.
func (p *ImplementationPhase) allTasksCompleteGuard(_ context.Context, _ ...any) bool {
	if p.data == nil {
		return false
	}

	tasks := p.data.Tasks
	if len(tasks) == 0 {
		return false
	}

	for _, t := range tasks {
		if t.Status != "completed" && t.Status != "abandoned" {
			return false
		}
	}

	return true
}

// Template Rendering

// renderPrompt loads and renders a template with phase data.
func (p *ImplementationPhase) renderPrompt(name string) string {
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
func (p *ImplementationPhase) prepareTemplateData() map[string]interface{} {
	data := make(map[string]interface{})

	// Project info (always available)
	data["ProjectName"] = p.project.Name
	data["ProjectDescription"] = p.project.Description
	data["ProjectBranch"] = p.project.Branch

	// Phase data (if available)
	if p.data != nil {
		// Planner used (optional field)
		if p.data.Planner_used != nil {
			data["PlannerUsed"] = *p.data.Planner_used
		}

		// Tasks
		data["TaskTotal"] = len(p.data.Tasks)
		data["Tasks"] = p.data.Tasks

		// Task status breakdown
		completed := 0
		inProgress := 0
		pending := 0
		abandoned := 0

		for _, t := range p.data.Tasks {
			switch t.Status {
			case "completed":
				completed++
			case "in_progress":
				inProgress++
			case "pending":
				pending++
			case "abandoned":
				abandoned++
			}
		}

		data["TaskCompleted"] = completed
		data["TaskInProgress"] = inProgress
		data["TaskPending"] = pending
		data["TaskAbandoned"] = abandoned

		// Tasks approved flag
		data["TasksApproved"] = p.data.Tasks_approved
	}

	return data
}
