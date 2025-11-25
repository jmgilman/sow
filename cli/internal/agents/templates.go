package agents

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed templates/*
var templatesFS embed.FS

// LoadPrompt loads an agent's prompt template from the embedded filesystem.
// The promptPath is relative to the templates/ directory.
//
// Example:
//
//	content, err := LoadPrompt("implementer.md")
func LoadPrompt(promptPath string) (string, error) {
	data, err := fs.ReadFile(templatesFS, "templates/"+promptPath)
	if err != nil {
		return "", fmt.Errorf("failed to load prompt %s: %w", promptPath, err)
	}
	return string(data), nil
}
