package prompts

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"
)

// PromptID uniquely identifies a prompt template.
type PromptID string

// Statechart prompt IDs - map to state machine states.
const (
	PromptNoProject               PromptID = "statechart.no_project"
	PromptPlanningActive          PromptID = "statechart.planning_active"
	PromptImplementationPlanning  PromptID = "statechart.implementation_planning"
	PromptImplementationExecuting PromptID = "statechart.implementation_executing"
	PromptReviewActive            PromptID = "statechart.review_active"
	PromptFinalizeDocumentation   PromptID = "statechart.finalize_documentation"
	PromptFinalizeChecks          PromptID = "statechart.finalize_checks"
	PromptFinalizeDelete          PromptID = "statechart.finalize_delete"
)

// Command prompt IDs - Composable greeting system.
const (
	PromptGreetBase          PromptID = "greet.base"
	PromptGreetStateUninit   PromptID = "greet.state.uninitialized"
	PromptGreetStateOperator PromptID = "greet.state.operator"
	PromptGreetStateOrch     PromptID = "greet.state.orchestrator"
)

// Command prompt IDs - Entry point prompts for CLI commands.
const (
	PromptCommandNew      PromptID = "command.new"
	PromptCommandContinue PromptID = "command.continue"
)

// Mode prompt IDs - Entry points for different modes.
const (
	PromptModeExplore PromptID = "mode.explore"
)

// Guidance prompt IDs - On-demand guidance for specific tasks.
const (
	PromptGuidanceResearch PromptID = "guidance.research"
)

// Context represents data needed to render a prompt.
// Implementations must provide a ToMap method that converts
// the context into template-compatible data.
type Context interface {
	ToMap() map[string]interface{}
}

// Registry manages all prompt templates.
type Registry struct {
	templates map[PromptID]*template.Template
}

// NewRegistry creates a new empty prompt registry.
func NewRegistry() *Registry {
	return &Registry{
		templates: make(map[PromptID]*template.Template),
	}
}

// Register loads and parses a template from the embedded filesystem.
func (r *Registry) Register(id PromptID, path string) error {
	content, err := templatesFS.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", path, err)
	}

	// Create template with custom functions
	tmpl, err := template.New(string(id)).Funcs(template.FuncMap{
		"join": func(sep string, elems []string) string {
			return strings.Join(elems, sep)
		},
	}).Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", path, err)
	}

	r.templates[id] = tmpl
	return nil
}

// Render renders a prompt template with the given context.
func (r *Registry) Render(id PromptID, ctx Context) (string, error) {
	tmpl, ok := r.templates[id]
	if !ok {
		return "", fmt.Errorf("unknown prompt ID: %s", id)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx.ToMap()); err != nil {
		return "", fmt.Errorf("failed to render prompt %s: %w", id, err)
	}

	return buf.String(), nil
}

// Embed all prompt templates from the templates/ directory
//
//go:embed templates/**/*.md templates/greet/*.md templates/greet/states/*.md templates/commands/*.md templates/modes/*.md templates/guidance/*.md
var templatesFS embed.FS

// Default registry, initialized at startup.
var defaultRegistry *Registry

func init() {
	defaultRegistry = NewRegistry()

	// Map prompt IDs to their template files
	promptFiles := map[PromptID]string{
		// Statechart prompts
		PromptNoProject:               "templates/statechart/no_project.md",
		PromptPlanningActive:          "templates/statechart/planning_active.md",
		PromptImplementationPlanning:  "templates/statechart/implementation_planning.md",
		PromptImplementationExecuting: "templates/statechart/implementation_executing.md",
		PromptReviewActive:            "templates/statechart/review_active.md",
		PromptFinalizeDocumentation:   "templates/statechart/finalize_documentation.md",
		PromptFinalizeChecks:          "templates/statechart/finalize_checks.md",
		PromptFinalizeDelete:          "templates/statechart/finalize_delete.md",

		// Composable greeting system
		PromptGreetBase:          "templates/greet/base.md",
		PromptGreetStateUninit:   "templates/greet/states/uninitialized.md",
		PromptGreetStateOperator: "templates/greet/states/operator.md",
		PromptGreetStateOrch:     "templates/greet/states/orchestrator.md",

		// Entry point command prompts
		PromptCommandNew:      "templates/commands/new.md",
		PromptCommandContinue: "templates/commands/continue.md",

		// Mode prompts
		PromptModeExplore: "templates/modes/explore.md",

		// Guidance prompts
		PromptGuidanceResearch: "templates/guidance/research.md",
	}

	// Load and parse all templates
	for id, path := range promptFiles {
		if err := defaultRegistry.Register(id, path); err != nil {
			panic(fmt.Sprintf("failed to register prompt %s: %v", id, err))
		}
	}
}

// Render renders a prompt using the default registry.
func Render(id PromptID, ctx Context) (string, error) {
	return defaultRegistry.Render(id, ctx)
}
