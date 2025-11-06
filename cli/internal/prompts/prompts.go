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
	"embed"
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
