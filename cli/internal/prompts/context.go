// Package prompts provides a unified abstraction layer for all internal prompt templates.
//
// It centralizes template loading, parsing, and rendering for both statechart state machine
// prompts and command prompts. All templates are embedded at compile time and pre-parsed
// for performance.
package prompts

import (
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/jmgilman/sow/cli/schemas/phases"
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

	// Get discovery_type from metadata
	if c.ProjectState.Phases.Discovery.Metadata != nil {
		if discoveryType, ok := c.ProjectState.Phases.Discovery.Metadata["discovery_type"].(string); ok && discoveryType != "" {
			data["DiscoveryType"] = discoveryType
		}
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
func (c *StatechartContext) addTaskStatusBreakdown(tasks []phases.Task, data map[string]interface{}) {
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

	// Get iteration from metadata (default to 1)
	iteration := int64(1)
	if c.ProjectState.Phases.Review.Metadata != nil {
		if iter, ok := c.ProjectState.Phases.Review.Metadata["iteration"].(int); ok && iter > 0 {
			iteration = int64(iter)
		} else if iter, ok := c.ProjectState.Phases.Review.Metadata["iteration"].(int64); ok && iter > 0 {
			iteration = iter
		}
	}
	data["ReviewIteration"] = iteration

	// Get review artifacts (artifacts with type=review metadata)
	var reviewArtifacts []interface{}
	for _, artifact := range c.ProjectState.Phases.Review.Artifacts {
		if artifact.Metadata != nil {
			if artifactType, ok := artifact.Metadata["type"].(string); ok && artifactType == "review" {
				reviewArtifacts = append(reviewArtifacts, artifact)
			}
		}
	}

	// Previous iteration context
	if iteration > 1 && len(reviewArtifacts) > 0 {
		data["HasPreviousReview"] = true
		// Get the last review artifact's assessment
		lastArtifact := c.ProjectState.Phases.Review.Artifacts[len(c.ProjectState.Phases.Review.Artifacts)-1]
		if lastArtifact.Metadata != nil {
			if assessment, ok := lastArtifact.Metadata["assessment"].(string); ok {
				data["PreviousAssessment"] = assessment
			}
		}
	}
}

// addFinalizeData adds finalize phase data to the template data.
func (c *StatechartContext) addFinalizeData(data map[string]interface{}) {
	if c.State == StateFinalizeDocumentation {
		// Get documentation_updates from metadata
		if c.ProjectState.Phases.Finalize.Metadata != nil {
			if updates, ok := c.ProjectState.Phases.Finalize.Metadata["documentation_updates"].([]interface{}); ok && len(updates) > 0 {
				data["HasDocumentationUpdates"] = true
				data["DocumentationUpdates"] = updates
			} else if update, ok := c.ProjectState.Phases.Finalize.Metadata["documentation_updates"].(string); ok && update != "" {
				// Handle string case
				data["HasDocumentationUpdates"] = true
				data["DocumentationUpdates"] = []string{update}
			}
		}
	}

	if c.State == StateFinalizeChecks {
		data["InFinalizeChecks"] = true
	}

	if c.State == StateFinalizeDelete {
		c.extractFinalizeDeleteData(data)
	}
}

// extractFinalizeDeleteData extracts metadata for the FinalizeDelete state.
func (c *StatechartContext) extractFinalizeDeleteData(data map[string]interface{}) {
	if c.ProjectState.Phases.Finalize.Metadata == nil {
		return
	}

	if deleted, ok := c.ProjectState.Phases.Finalize.Metadata["project_deleted"].(bool); ok {
		data["ProjectDeleted"] = deleted
	}
	if prURL, ok := c.ProjectState.Phases.Finalize.Metadata["pr_url"].(string); ok && prURL != "" {
		data["HasPR"] = true
		data["PRURL"] = prURL
	}
}

// NewProjectContext holds context for the "sow new" command prompt.
type NewProjectContext struct {
	RepoRoot        string
	BranchName      string
	IssueNumber     *int
	IssueTitle      string
	IssueBody       string
	InitialPrompt   string
	StatechartState string
}

// ToMap converts NewProjectContext to a map for template rendering.
func (c *NewProjectContext) ToMap() map[string]interface{} {
	data := make(map[string]interface{})
	data["RepoRoot"] = c.RepoRoot
	data["BranchName"] = c.BranchName
	data["StatechartState"] = c.StatechartState
	data["InitialPrompt"] = c.InitialPrompt

	if c.IssueNumber != nil {
		data["IssueNumber"] = *c.IssueNumber
		data["IssueTitle"] = c.IssueTitle
		data["IssueBody"] = c.IssueBody
	}

	return data
}

// ContinueProjectContext holds context for the "sow continue" command prompt.
type ContinueProjectContext struct {
	// Repository info
	BranchName string

	// Project metadata
	ProjectName        string
	ProjectDescription string
	IssueNumber        *int

	// Current state
	StatechartState string

	// Phase status
	DiscoveryEnabled      bool
	DiscoveryStatus       string
	DesignEnabled         bool
	DesignStatus          string
	ImplementationStatus  string
	ReviewStatus          string
	FinalizeStatus        string

	// Task information
	TasksTotal       int
	TasksCompleted   int
	TasksInProgress  int
	TasksPending     int
	TasksAbandoned   int
	CurrentTaskID    string
	CurrentTaskName  string
	CurrentTaskStatus string

	// Guidance
	StateSpecificGuidance string
	NextActions           string
	CurrentPhaseDescription string
	NextActionSummary     string
}

// ToMap converts ContinueProjectContext to a map for template rendering.
func (c *ContinueProjectContext) ToMap() map[string]interface{} {
	data := make(map[string]interface{})

	// Repository info
	data["BranchName"] = c.BranchName

	// Project metadata
	data["ProjectName"] = c.ProjectName
	data["ProjectDescription"] = c.ProjectDescription
	if c.IssueNumber != nil {
		data["IssueNumber"] = *c.IssueNumber
	}

	// Current state
	data["StatechartState"] = c.StatechartState

	// Phase status
	data["DiscoveryEnabled"] = c.DiscoveryEnabled
	data["DiscoveryStatus"] = c.DiscoveryStatus
	data["DesignEnabled"] = c.DesignEnabled
	data["DesignStatus"] = c.DesignStatus
	data["ImplementationStatus"] = c.ImplementationStatus
	data["ReviewStatus"] = c.ReviewStatus
	data["FinalizeStatus"] = c.FinalizeStatus

	// Task information
	data["TasksTotal"] = c.TasksTotal
	data["TasksCompleted"] = c.TasksCompleted
	data["TasksInProgress"] = c.TasksInProgress
	data["TasksPending"] = c.TasksPending
	data["TasksAbandoned"] = c.TasksAbandoned
	data["CurrentTaskID"] = c.CurrentTaskID
	data["CurrentTaskName"] = c.CurrentTaskName
	data["CurrentTaskStatus"] = c.CurrentTaskStatus

	// Guidance
	data["StateSpecificGuidance"] = c.StateSpecificGuidance
	data["NextActions"] = c.NextActions
	data["CurrentPhaseDescription"] = c.CurrentPhaseDescription
	data["NextActionSummary"] = c.NextActionSummary

	return data
}

// ExplorationContext holds the context for rendering exploration mode prompts.
type ExplorationContext struct {
	Topic  string
	Branch string
	Status string
	Files  []ExplorationFile
}

// ExplorationFile represents a file in the exploration index for templates.
type ExplorationFile struct {
	Path        string
	Description string
	Tags        []string
}

// ToMap converts ExplorationContext to a map for template rendering.
func (c *ExplorationContext) ToMap() map[string]interface{} {
	data := make(map[string]interface{})
	data["Topic"] = c.Topic
	data["Branch"] = c.Branch
	data["Status"] = c.Status

	if len(c.Files) > 0 {
		files := make([]map[string]interface{}, len(c.Files))
		for i, f := range c.Files {
			files[i] = map[string]interface{}{
				"Path":        f.Path,
				"Description": f.Description,
				"Tags":        f.Tags,
			}
		}
		data["Files"] = files
	}

	return data
}

// GuidanceContext holds the context for rendering guidance prompts.
// Currently guidance prompts don't need context, but this allows for future expansion.
type GuidanceContext struct {
	// Future: Could include current exploration info, recent files, etc.
}

// ToMap converts GuidanceContext to a map for template rendering.
func (c *GuidanceContext) ToMap() map[string]interface{} {
	return make(map[string]interface{})
}
