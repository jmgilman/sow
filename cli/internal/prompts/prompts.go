package prompts

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"
)

// PromptID uniquely identifies a shared prompt template.
// Project-specific prompts are defined in their respective packages.
type PromptID string

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
	PromptModeExplore   PromptID = "mode.explore"
	PromptModeDesign    PromptID = "mode.design"
	PromptModeBreakdown PromptID = "mode.breakdown"
)

// Guidance prompt IDs - On-demand guidance for specific tasks.
const (
	PromptGuidanceResearch PromptID = "guidance.research"

	// Design guidance prompts.
	PromptGuidanceDesignPRD        PromptID = "guidance.design.prd"
	PromptGuidanceDesignArc42      PromptID = "guidance.design.arc42"
	PromptGuidanceDesignDoc        PromptID = "guidance.design.design-doc"
	PromptGuidanceDesignADR        PromptID = "guidance.design.adr"
	PromptGuidanceDesignC4Diagrams PromptID = "guidance.design.c4-diagrams"
)

// Context represents data needed to render a prompt.
// Implementations must provide a ToMap method that converts
// the context into template-compatible data.
type Context interface {
	ToMap() map[string]interface{}
}

// Registry manages prompt templates with generic key type.
// K must be a comparable type (typically string or a string-based type).
type Registry[K comparable] struct {
	templates map[K]*template.Template
}

// NewRegistry creates a new empty prompt registry with the specified key type.
func NewRegistry[K comparable]() *Registry[K] {
	return &Registry[K]{
		templates: make(map[K]*template.Template),
	}
}

// RegisterFromFS loads and parses a template from the provided embedded filesystem.
// This allows each registry to use its own embed.FS instance.
func (r *Registry[K]) RegisterFromFS(fs embed.FS, id K, path string) error {
	content, err := fs.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", path, err)
	}

	// Create template with custom functions
	tmpl, err := template.New(fmt.Sprint(id)).Funcs(template.FuncMap{
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

// Register loads and parses a template from the default shared embedded filesystem.
// This is a convenience method for the shared prompts registry.
// Project-specific registries should use RegisterFromFS with their own embed.FS.
func (r *Registry[K]) Register(id K, path string) error {
	return r.RegisterFromFS(templatesFS, id, path)
}

// Render renders a prompt template with the given context.
func (r *Registry[K]) Render(id K, ctx Context) (string, error) {
	tmpl, ok := r.templates[id]
	if !ok {
		return "", fmt.Errorf("unknown prompt ID: %v", id)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx.ToMap()); err != nil {
		return "", fmt.Errorf("failed to render prompt %v: %w", id, err)
	}

	return buf.String(), nil
}

// Embed shared prompt templates from the templates/ directory.
// Project-specific templates are embedded in their respective packages.
//
//go:embed templates/greet/*.md templates/greet/states/*.md templates/commands/*.md templates/modes/*.md templates/guidance/*.md templates/guidance/design/*.md
var templatesFS embed.FS

// Default registry, initialized at startup.
// This registry uses PromptID as the key type for shared prompts.
var defaultRegistry *Registry[PromptID]

func init() {
	defaultRegistry = NewRegistry[PromptID]()

	// Map shared prompt IDs to their template files.
	// Project-specific prompts are registered in their own packages.
	promptFiles := map[PromptID]string{
		// Composable greeting system
		PromptGreetBase:          "templates/greet/base.md",
		PromptGreetStateUninit:   "templates/greet/states/uninitialized.md",
		PromptGreetStateOperator: "templates/greet/states/operator.md",
		PromptGreetStateOrch:     "templates/greet/states/orchestrator.md",

		// Entry point command prompts
		PromptCommandNew:      "templates/commands/new.md",
		PromptCommandContinue: "templates/commands/continue.md",

		// Mode prompts
		PromptModeExplore:   "templates/modes/explore.md",
		PromptModeDesign:    "templates/modes/design.md",
		PromptModeBreakdown: "templates/modes/breakdown.md",

		// Guidance prompts
		PromptGuidanceResearch: "templates/guidance/research.md",

		// Design guidance prompts
		PromptGuidanceDesignPRD:        "templates/guidance/design/prd.md",
		PromptGuidanceDesignArc42:      "templates/guidance/design/arc42.md",
		PromptGuidanceDesignDoc:        "templates/guidance/design/design-doc.md",
		PromptGuidanceDesignADR:        "templates/guidance/design/adr.md",
		PromptGuidanceDesignC4Diagrams: "templates/guidance/design/c4-diagrams.md",
	}

	// Load and parse all shared templates
	for id, path := range promptFiles {
		if err := defaultRegistry.Register(id, path); err != nil {
			panic(fmt.Sprintf("failed to register shared prompt %s: %v", id, err))
		}
	}
}

// Render renders a prompt using the default shared prompts registry.
func Render(id PromptID, ctx Context) (string, error) {
	return defaultRegistry.Render(id, ctx)
}
