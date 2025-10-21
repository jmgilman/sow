package statechart

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"

	"github.com/jmgilman/sow/cli/schemas"
)

// Embed all prompt templates from the prompts/ directory
//
//go:embed prompts/*.md
var promptsFS embed.FS

// Parsed templates, initialized once at startup.
var templates map[State]*template.Template

func init() {
	templates = make(map[State]*template.Template)

	// Map states to their template files
	stateFiles := map[State]string{
		NoProject:               "prompts/no_project.md",
		DiscoveryDecision:       "prompts/discovery_decision.md",
		DiscoveryActive:         "prompts/discovery_active.md",
		DesignDecision:          "prompts/design_decision.md",
		DesignActive:            "prompts/design_active.md",
		ImplementationPlanning:  "prompts/implementation_planning.md",
		ImplementationExecuting: "prompts/implementation_executing.md",
		ReviewActive:            "prompts/review_active.md",
		FinalizeDocumentation:   "prompts/finalize_documentation.md",
		FinalizeChecks:          "prompts/finalize_checks.md",
		FinalizeDelete:          "prompts/finalize_delete.md",
	}

	// Load and parse all templates
	for state, file := range stateFiles {
		content, err := promptsFS.ReadFile(file)
		if err != nil {
			panic(fmt.Sprintf("failed to read template %s: %v", file, err))
		}

		tmpl, err := template.New(string(state)).Parse(string(content))
		if err != nil {
			panic(fmt.Sprintf("failed to parse template %s: %v", file, err))
		}

		templates[state] = tmpl
	}
}

// PromptContext contains all information needed to generate contextual prompts.
type PromptContext struct {
	State        State
	ProjectState *schemas.ProjectState
}

// GeneratePrompt generates a contextual prompt for the current state using templates.
func GeneratePrompt(ctx PromptContext) string {
	tmpl, ok := templates[ctx.State]
	if !ok {
		return fmt.Sprintf("Unknown state: %s", ctx.State)
	}

	// Prepare data for the template
	data := prepareTemplateData(ctx)

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Sprintf("Error rendering prompt for state %s: %v", ctx.State, err)
	}

	return buf.String()
}

// prepareTemplateData extracts relevant data from ProjectState for template rendering.
func prepareTemplateData(ctx PromptContext) map[string]interface{} {
	data := make(map[string]interface{})

	if ctx.ProjectState == nil {
		return data
	}

	// Project metadata (available to all prompts)
	data["ProjectName"] = ctx.ProjectState.Project.Name
	data["ProjectDescription"] = ctx.ProjectState.Project.Description
	data["ProjectBranch"] = ctx.ProjectState.Project.Branch

	// Add phase-specific data
	addDiscoveryData(ctx, data)
	addDesignData(ctx, data)
	addImplementationData(ctx, data)
	addReviewData(ctx, data)
	addFinalizeData(ctx, data)

	return data
}

// addDiscoveryData adds discovery phase data to the template data.
func addDiscoveryData(ctx PromptContext, data map[string]interface{}) {
	if ctx.State != DiscoveryActive && ctx.State != DiscoveryDecision {
		return
	}

	if discoveryType, ok := ctx.ProjectState.Phases.Discovery.Discovery_type.(string); ok && discoveryType != "" {
		data["DiscoveryType"] = discoveryType
	}

	artifacts := ctx.ProjectState.Phases.Discovery.Artifacts
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
func addDesignData(ctx PromptContext, data map[string]interface{}) {
	if ctx.State != DesignActive && ctx.State != DesignDecision {
		return
	}

	artifacts := ctx.ProjectState.Phases.Design.Artifacts
	data["ArtifactCount"] = len(artifacts)

	approvedCount := 0
	for _, a := range artifacts {
		if a.Approved {
			approvedCount++
		}
	}
	data["ApprovedCount"] = approvedCount

	// Check if discovery phase was completed
	hasDiscovery := ctx.ProjectState.Phases.Discovery.Status == "completed"
	data["HasDiscovery"] = hasDiscovery
	if hasDiscovery {
		data["DiscoveryArtifactCount"] = len(ctx.ProjectState.Phases.Discovery.Artifacts)
	}
}

// addImplementationData adds implementation phase data to the template data.
func addImplementationData(ctx PromptContext, data map[string]interface{}) {
	if ctx.State != ImplementationPlanning && ctx.State != ImplementationExecuting {
		return
	}

	tasks := ctx.ProjectState.Phases.Implementation.Tasks
	data["TaskTotal"] = len(tasks)

	// Check available inputs
	hasDiscovery := ctx.ProjectState.Phases.Discovery.Status == "completed"
	hasDesign := ctx.ProjectState.Phases.Design.Status == "completed"
	data["HasDiscovery"] = hasDiscovery
	data["HasDesign"] = hasDesign

	if hasDiscovery {
		data["DiscoveryArtifactCount"] = len(ctx.ProjectState.Phases.Discovery.Artifacts)
	}
	if hasDesign {
		data["DesignArtifactCount"] = len(ctx.ProjectState.Phases.Design.Artifacts)
	}

	// Task status breakdown (for executing state)
	if ctx.State == ImplementationExecuting {
		addTaskStatusBreakdown(tasks, data)
	}
}

// addTaskStatusBreakdown adds task status counts to the template data.
func addTaskStatusBreakdown(tasks []schemas.Task, data map[string]interface{}) {
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
func addReviewData(ctx PromptContext, data map[string]interface{}) {
	if ctx.State != ReviewActive {
		return
	}

	iteration := ctx.ProjectState.Phases.Review.Iteration
	if iteration == 0 {
		iteration = 1 // Default to 1 if not set
	}
	data["ReviewIteration"] = iteration

	// Previous iteration context
	if iteration > 1 && len(ctx.ProjectState.Phases.Review.Reports) > 0 {
		data["HasPreviousReview"] = true
		prevReport := ctx.ProjectState.Phases.Review.Reports[len(ctx.ProjectState.Phases.Review.Reports)-1]
		data["PreviousAssessment"] = prevReport.Assessment
	}
}

// addFinalizeData adds finalize phase data to the template data.
func addFinalizeData(ctx PromptContext, data map[string]interface{}) {
	if ctx.State == FinalizeDocumentation {
		if updates, ok := ctx.ProjectState.Phases.Finalize.Documentation_updates.([]interface{}); ok && len(updates) > 0 {
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

	if ctx.State == FinalizeChecks {
		data["InFinalizeChecks"] = true
	}

	if ctx.State == FinalizeDelete {
		data["ProjectDeleted"] = ctx.ProjectState.Phases.Finalize.Project_deleted
		if prURL, ok := ctx.ProjectState.Phases.Finalize.Pr_url.(string); ok && prURL != "" {
			data["HasPR"] = true
			data["PRURL"] = prURL
		}
	}
}
