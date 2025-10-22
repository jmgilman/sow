// Package prompts provides a unified abstraction layer for all internal prompt templates.
//
// It centralizes template loading, parsing, and rendering for both statechart state machine
// prompts and command prompts. All templates are embedded at compile time and pre-parsed
// for performance.
package prompts

import (
	"github.com/jmgilman/sow/cli/schemas"
)

// State constants (matching statechart.State values).
const (
	StateNoProject               = "NoProject"
	StateDiscoveryDecision       = "DiscoveryDecision"
	StateDiscoveryActive         = "DiscoveryActive"
	StateDesignDecision          = "DesignDecision"
	StateDesignActive            = "DesignActive"
	StateImplementationPlanning  = "ImplementationPlanning"
	StateImplementationExecuting = "ImplementationExecuting"
	StateReviewActive            = "ReviewActive"
	StateFinalizeDocumentation   = "FinalizeDocumentation"
	StateFinalizeChecks          = "FinalizeChecks"
	StateFinalizeDelete          = "FinalizeDelete"
)

// GreetContext holds the context for rendering the greeting template.
type GreetContext struct {
	SowInitialized bool
	HasProject     bool
	Project        *ProjectGreetContext
	OpenIssues     int
	GHAvailable    bool
}

// ProjectGreetContext holds project-specific greeting context.
type ProjectGreetContext struct {
	Name            string
	Branch          string
	Description     string
	CurrentPhase    string
	PhaseStatus     string
	TasksTotal      int
	TasksComplete   int
	TasksInProgress int
	TasksPending    int
	TasksAbandoned  int
	CurrentTask     *TaskGreetContext
}

// TaskGreetContext holds current task information.
type TaskGreetContext struct {
	ID   string
	Name string
}

// ToMap converts GreetContext to a map for template rendering.
func (c *GreetContext) ToMap() map[string]interface{} {
	data := make(map[string]interface{})
	data["SowInitialized"] = c.SowInitialized
	data["HasProject"] = c.HasProject
	data["OpenIssues"] = c.OpenIssues
	data["GHAvailable"] = c.GHAvailable

	if c.Project != nil {
		projectMap := make(map[string]interface{})
		projectMap["Name"] = c.Project.Name
		projectMap["Branch"] = c.Project.Branch
		projectMap["Description"] = c.Project.Description
		projectMap["CurrentPhase"] = c.Project.CurrentPhase
		projectMap["PhaseStatus"] = c.Project.PhaseStatus
		projectMap["TasksTotal"] = c.Project.TasksTotal
		projectMap["TasksComplete"] = c.Project.TasksComplete
		projectMap["TasksInProgress"] = c.Project.TasksInProgress
		projectMap["TasksPending"] = c.Project.TasksPending
		projectMap["TasksAbandoned"] = c.Project.TasksAbandoned

		if c.Project.CurrentTask != nil {
			taskMap := make(map[string]interface{})
			taskMap["ID"] = c.Project.CurrentTask.ID
			taskMap["Name"] = c.Project.CurrentTask.Name
			projectMap["CurrentTask"] = taskMap
		}

		data["Project"] = projectMap
	}

	return data
}

// StatechartContext holds state machine prompt context.
type StatechartContext struct {
	State        string // State name (matches statechart.State string values)
	ProjectState *schemas.ProjectState
}

// ToMap converts StatechartContext to a map for template rendering.
// This replicates the logic from statechart/prompts.go prepareTemplateData.
func (c *StatechartContext) ToMap() map[string]interface{} {
	data := make(map[string]interface{})

	if c.ProjectState == nil {
		return data
	}

	// Project metadata (available to all prompts)
	data["ProjectName"] = c.ProjectState.Project.Name
	data["ProjectDescription"] = c.ProjectState.Project.Description
	data["ProjectBranch"] = c.ProjectState.Project.Branch

	// Add phase-specific data
	c.addDiscoveryData(data)
	c.addDesignData(data)
	c.addImplementationData(data)
	c.addReviewData(data)
	c.addFinalizeData(data)

	return data
}

// addDiscoveryData adds discovery phase data to the template data.
func (c *StatechartContext) addDiscoveryData(data map[string]interface{}) {
	if c.State != StateDiscoveryActive && c.State != StateDiscoveryDecision {
		return
	}

	if discoveryType, ok := c.ProjectState.Phases.Discovery.Discovery_type.(string); ok && discoveryType != "" {
		data["DiscoveryType"] = discoveryType
	}

	artifacts := c.ProjectState.Phases.Discovery.Artifacts
	data["ArtifactCount"] = len(artifacts)

	approvedCount := 0
	for _, a := range artifacts {
		if a.Approved {
			approvedCount++
		}
	}
	data["ApprovedCount"] = approvedCount
}

// addDesignData adds design phase data to the template data.
func (c *StatechartContext) addDesignData(data map[string]interface{}) {
	if c.State != StateDesignActive && c.State != StateDesignDecision {
		return
	}

	artifacts := c.ProjectState.Phases.Design.Artifacts
	data["ArtifactCount"] = len(artifacts)

	approvedCount := 0
	for _, a := range artifacts {
		if a.Approved {
			approvedCount++
		}
	}
	data["ApprovedCount"] = approvedCount

	// Check if discovery phase was completed
	hasDiscovery := c.ProjectState.Phases.Discovery.Status == "completed"
	data["HasDiscovery"] = hasDiscovery
	if hasDiscovery {
		data["DiscoveryArtifactCount"] = len(c.ProjectState.Phases.Discovery.Artifacts)
	}
}

// addImplementationData adds implementation phase data to the template data.
func (c *StatechartContext) addImplementationData(data map[string]interface{}) {
	if c.State != StateImplementationPlanning && c.State != StateImplementationExecuting {
		return
	}

	tasks := c.ProjectState.Phases.Implementation.Tasks
	data["TaskTotal"] = len(tasks)

	// Check available inputs
	hasDiscovery := c.ProjectState.Phases.Discovery.Status == "completed"
	hasDesign := c.ProjectState.Phases.Design.Status == "completed"
	data["HasDiscovery"] = hasDiscovery
	data["HasDesign"] = hasDesign

	if hasDiscovery {
		data["DiscoveryArtifactCount"] = len(c.ProjectState.Phases.Discovery.Artifacts)
	}
	if hasDesign {
		data["DesignArtifactCount"] = len(c.ProjectState.Phases.Design.Artifacts)
	}

	// Task status breakdown (for executing state)
	if c.State == StateImplementationExecuting {
		c.addTaskStatusBreakdown(tasks, data)
	}
}

// addTaskStatusBreakdown adds task status counts to the template data.
func (c *StatechartContext) addTaskStatusBreakdown(tasks []schemas.Task, data map[string]interface{}) {
	completed := 0
	inProgress := 0
	pending := 0

	for _, t := range tasks {
		switch t.Status {
		case "completed":
			completed++
		case "in_progress":
			inProgress++
		case "pending":
			pending++
		}
	}

	data["TaskCompleted"] = completed
	data["TaskInProgress"] = inProgress
	data["TaskPending"] = pending
	data["Tasks"] = tasks
}

// addReviewData adds review phase data to the template data.
func (c *StatechartContext) addReviewData(data map[string]interface{}) {
	if c.State != StateReviewActive {
		return
	}

	iteration := c.ProjectState.Phases.Review.Iteration
	if iteration == 0 {
		iteration = 1 // Default to 1 if not set
	}
	data["ReviewIteration"] = iteration

	// Previous iteration context
	if iteration > 1 && len(c.ProjectState.Phases.Review.Reports) > 0 {
		data["HasPreviousReview"] = true
		prevReport := c.ProjectState.Phases.Review.Reports[len(c.ProjectState.Phases.Review.Reports)-1]
		data["PreviousAssessment"] = prevReport.Assessment
	}
}

// addFinalizeData adds finalize phase data to the template data.
func (c *StatechartContext) addFinalizeData(data map[string]interface{}) {
	if c.State == StateFinalizeDocumentation {
		if updates, ok := c.ProjectState.Phases.Finalize.Documentation_updates.([]interface{}); ok && len(updates) > 0 {
			// Convert to string slice
			strUpdates := make([]string, 0, len(updates))
			for _, u := range updates {
				if s, ok := u.(string); ok {
					strUpdates = append(strUpdates, s)
				}
			}
			data["HasDocumentationUpdates"] = len(strUpdates) > 0
			data["DocumentationUpdates"] = strUpdates
		}
	}

	if c.State == StateFinalizeChecks {
		data["InFinalizeChecks"] = true
	}

	if c.State == StateFinalizeDelete {
		data["ProjectDeleted"] = c.ProjectState.Phases.Finalize.Project_deleted
		if prURL, ok := c.ProjectState.Phases.Finalize.Pr_url.(string); ok && prURL != "" {
			data["HasPR"] = true
			data["PRURL"] = prURL
		}
	}
}
