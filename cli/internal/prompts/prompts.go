// Package prompts provides embedded shared prompt templates.
//
// This is a pure data layer - it only embeds markdown template files.
// Use the templates package (internal/templates) to render these templates.
//
// Example usage:
//
//	import (
//	    "github.com/jmgilman/sow/cli/internal/prompts"
//	    "github.com/jmgilman/sow/cli/internal/templates"
//	)
//
//	output, err := templates.Render(prompts.FS, "templates/guidance/research.md", nil)
package prompts

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"
)

// FS contains all shared prompt templates.
// Templates are organized under templates/ directory:
//
//   - templates/guidance/        - On-demand guidance for specific tasks
//   - templates/guidance/design/ - Design artifact templates
//   - templates/guidance/implementer/ - Implementation guidance
//   - templates/guidance/reviewer/    - Code review guidance
//   - templates/commands/        - Command-specific prompts
//   - templates/greet/           - Greeting and mode selection prompts
//   - templates/modes/           - Mode-specific entry prompts
//
//go:embed templates
var FS embed.FS

// PromptID uniquely identifies a shared prompt template.
// This type is kept for backward compatibility with the legacy rendering system.
// New code should use the templates package directly with file paths.
type PromptID string

// Legacy prompt IDs for backward compatibility.
const (
	PromptGreetBase          PromptID = "greet.base"
	PromptGreetStateUninit   PromptID = "greet.state.uninitialized"
	PromptGreetStateOperator PromptID = "greet.state.operator"
	PromptGreetStateOrch     PromptID = "greet.state.orchestrator"
	PromptGreetOrchestrator  PromptID = "greet.orchestrator"
	PromptCommandNew         PromptID = "command.new"
	PromptCommandContinue    PromptID = "command.continue"
	PromptModeExplore        PromptID = "mode.explore"
	PromptModeDesign         PromptID = "mode.design"
	PromptModeBreakdown      PromptID = "mode.breakdown"
)

// Context represents data needed to render a prompt (legacy).
type Context interface {
	ToMap() map[string]interface{}
}

// Render is a legacy wrapper that maps PromptIDs to template paths.
// New code should use templates.Render() directly with file paths.
//
// Deprecated: Use templates.Render(prompts.FS, path, data) instead.
func Render(id PromptID, ctx Context) (string, error) {
	// Map legacy IDs to template paths
	pathMap := map[PromptID]string{
		PromptGreetBase:          "templates/greet/base.md",
		PromptGreetStateUninit:   "templates/greet/states/uninitialized.md",
		PromptGreetStateOperator: "templates/greet/states/operator.md",
		PromptGreetStateOrch:     "templates/greet/states/orchestrator.md",
		PromptGreetOrchestrator:  "templates/greet/orchestrator.md",
		PromptCommandNew:         "templates/commands/new.md",
		PromptCommandContinue:    "templates/commands/continue.md",
		PromptModeExplore:        "templates/modes/explore.md",
		PromptModeDesign:         "templates/modes/design.md",
		PromptModeBreakdown:      "templates/modes/breakdown.md",
	}

	path, ok := pathMap[id]
	if !ok {
		return "", fmt.Errorf("unknown prompt ID: %v", id)
	}

	// Import templates package to use Render
	// This is done at function scope to avoid import cycles
	return renderLegacy(FS, path, ctx)
}

// renderLegacy handles legacy rendering with Context interface.
func renderLegacy(embedFS embed.FS, path string, ctx Context) (string, error) {
	content, err := embedFS.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read template %s: %w", path, err)
	}

	// Use template helper functions for consistent rendering
	// This is necessary for templates that use functions like {{join}}
	tmpl, err := template.New(path).Funcs(legacyFuncMap()).Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", path, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx.ToMap()); err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", path, err)
	}

	return buf.String(), nil
}

// legacyFuncMap returns template functions for legacy rendering.
// This provides basic string operations needed by mode templates.
func legacyFuncMap() template.FuncMap {
	return template.FuncMap{
		// String operations - join a slice of strings with a separator
		// Usage: {{join ", " .Tags}}
		"join": func(sep string, elems []string) string {
			return strings.Join(elems, sep)
		},
	}
}
