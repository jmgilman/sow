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

// Parsed templates, initialized once at startup
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

	// Discovery phase data
	if ctx.State == DiscoveryActive {
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

	// Design phase data
	if ctx.State == DesignActive {
		artifacts := ctx.ProjectState.Phases.Design.Artifacts
		data["ArtifactCount"] = len(artifacts)

		approvedCount := 0
		for _, a := range artifacts {
			if a.Approved {
				approvedCount++
			}
		}
		data["ApprovedCount"] = approvedCount
	}

	// Implementation phase data
	if ctx.State == ImplementationExecuting {
		tasks := ctx.ProjectState.Phases.Implementation.Tasks
		data["TaskTotal"] = len(tasks)

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
	}

	// Review phase data
	if ctx.State == ReviewActive {
		iteration := ctx.ProjectState.Phases.Review.Iteration
		if iteration == 0 {
			iteration = 1 // Default to 1 if not set
		}
		data["ReviewIteration"] = iteration
	}

	return data
}
