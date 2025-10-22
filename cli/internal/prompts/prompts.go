package prompts

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

// PromptID uniquely identifies a prompt template.
type PromptID string

// Statechart prompt IDs - map to state machine states.
const (
	PromptNoProject               PromptID = "statechart.no_project"
	PromptDiscoveryDecision       PromptID = "statechart.discovery_decision"
	PromptDiscoveryActive         PromptID = "statechart.discovery_active"
	PromptDesignDecision          PromptID = "statechart.design_decision"
	PromptDesignActive            PromptID = "statechart.design_active"
	PromptImplementationPlanning  PromptID = "statechart.implementation_planning"
	PromptImplementationExecuting PromptID = "statechart.implementation_executing"
	PromptReviewActive            PromptID = "statechart.review_active"
	PromptFinalizeDocumentation   PromptID = "statechart.finalize_documentation"
	PromptFinalizeChecks          PromptID = "statechart.finalize_checks"
	PromptFinalizeDelete          PromptID = "statechart.finalize_delete"
)

// Command prompt IDs.
const (
	PromptGreetStandard     PromptID = "command.greet_standard"
	PromptGreetOperator     PromptID = "command.greet_operator"
	PromptGreetOrchestrator PromptID = "command.greet_orchestrator"
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

	tmpl, err := template.New(string(id)).Parse(string(content))
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
//go:embed templates/**/*.md templates/**/*.tmpl
var templatesFS embed.FS

// Default registry, initialized at startup.
var defaultRegistry *Registry

func init() {
	defaultRegistry = NewRegistry()

	// Map prompt IDs to their template files
	promptFiles := map[PromptID]string{
		// Statechart prompts
		PromptNoProject:               "templates/statechart/no_project.md",
		PromptDiscoveryDecision:       "templates/statechart/discovery_decision.md",
		PromptDiscoveryActive:         "templates/statechart/discovery_active.md",
		PromptDesignDecision:          "templates/statechart/design_decision.md",
		PromptDesignActive:            "templates/statechart/design_active.md",
		PromptImplementationPlanning:  "templates/statechart/implementation_planning.md",
		PromptImplementationExecuting: "templates/statechart/implementation_executing.md",
		PromptReviewActive:            "templates/statechart/review_active.md",
		PromptFinalizeDocumentation:   "templates/statechart/finalize_documentation.md",
		PromptFinalizeChecks:          "templates/statechart/finalize_checks.md",
		PromptFinalizeDelete:          "templates/statechart/finalize_delete.md",

		// Command prompts
		PromptGreetStandard:     "templates/commands/greet_standard.tmpl",
		PromptGreetOperator:     "templates/commands/greet_operator.tmpl",
		PromptGreetOrchestrator: "templates/commands/greet_orchestrator.tmpl",
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
