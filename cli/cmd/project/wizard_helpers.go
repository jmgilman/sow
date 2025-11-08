package project

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
)

// ProjectTypeConfig defines the configuration for a project type.
type ProjectTypeConfig struct {
	Prefix      string
	Description string
}

// projectTypes maps project type names to their configuration.
// These are the four project types currently supported by sow.
var projectTypes = map[string]ProjectTypeConfig{
	"standard": {
		Prefix:      "feat/",
		Description: "Feature work and bug fixes",
	},
	"exploration": {
		Prefix:      "explore/",
		Description: "Research and investigation",
	},
	"design": {
		Prefix:      "design/",
		Description: "Architecture and design documents",
	},
	"breakdown": {
		Prefix:      "breakdown/",
		Description: "Decompose work into tasks",
	},
}

// normalizeName converts user-friendly project names into valid git branch names.
//
// The function applies the following transformations:
//  1. Trim leading/trailing whitespace
//  2. Convert to lowercase
//  3. Replace spaces with hyphens
//  4. Remove invalid characters (keep only: a-z, 0-9, -, _)
//  5. Collapse multiple consecutive hyphens into single hyphen
//  6. Remove leading/trailing hyphens
//  7. Return normalized name
//
// Example transformations:
//   - "Web Based Agents" → "web-based-agents"
//   - "API V2" → "api-v2"
//   - "feature--name" → "feature-name"
//   - "-leading-trailing-" → "leading-trailing"
//   - "With!Invalid@Chars#" → "withinvalidchars"
//   - "UPPERCASE" → "uppercase"
//   - "  spaces  " → "spaces"
//
// Edge cases:
//   - Empty string → "" (empty string)
//   - Only spaces → "" (empty string)
//   - Only special characters → "" (empty string)
//   - Unicode characters → removed (only ASCII alphanumeric allowed)
func normalizeName(name string) string {
	// 1. Trim leading/trailing whitespace
	name = strings.TrimSpace(name)

	// 2. Convert to lowercase
	name = strings.ToLower(name)

	// 3. Replace spaces with hyphens
	name = strings.ReplaceAll(name, " ", "-")

	// 4. Remove invalid characters (keep only: a-z, 0-9, -, _)
	// This regex matches anything that is NOT alphanumeric, hyphen, or underscore
	invalidCharsRegex := regexp.MustCompile(`[^a-z0-9\-_]+`)
	name = invalidCharsRegex.ReplaceAllString(name, "")

	// 5. Collapse multiple consecutive hyphens into single hyphen
	multipleHyphensRegex := regexp.MustCompile(`-+`)
	name = multipleHyphensRegex.ReplaceAllString(name, "-")

	// 6. Remove leading/trailing hyphens
	name = strings.Trim(name, "-")

	// 7. Return normalized name
	return name
}

// getTypePrefix returns the branch prefix for a given project type.
// If the project type is not recognized, returns "feat/" as the default.
//
// Examples:
//   - getTypePrefix("standard") → "feat/"
//   - getTypePrefix("exploration") → "explore/"
//   - getTypePrefix("unknown") → "feat/" (fallback)
//   - getTypePrefix("") → "feat/" (fallback)
func getTypePrefix(projectType string) string {
	if config, exists := projectTypes[projectType]; exists {
		return config.Prefix
	}
	return "feat/" // Default fallback
}

// getTypeOptions converts the projectTypes map into huh-compatible options
// for select prompts.
//
// The options are returned in a consistent order:
//  1. standard
//  2. exploration
//  3. design
//  4. breakdown
//  5. cancel
//
// Each option displays the type's description as the label and uses
// the type name as the value.
//
// Returns a slice of huh.Option[string] ready to use in a select prompt.
func getTypeOptions() []huh.Option[string] {
	// Return options in consistent order
	return []huh.Option[string]{
		huh.NewOption(projectTypes["standard"].Description, "standard"),
		huh.NewOption(projectTypes["exploration"].Description, "exploration"),
		huh.NewOption(projectTypes["design"].Description, "design"),
		huh.NewOption(projectTypes["breakdown"].Description, "breakdown"),
		huh.NewOption("Cancel", "cancel"),
	}
}

// previewBranchName shows what the branch name will be for a given project type and name.
// It combines the project type's prefix with the normalized name.
//
// Example:
//   - previewBranchName("standard", "Web Based Agents") → "feat/web-based-agents"
//   - previewBranchName("exploration", "API Research") → "explore/api-research"
func previewBranchName(projectType, name string) string {
	prefix := getTypePrefix(projectType)
	normalizedName := normalizeName(name)
	return prefix + normalizedName
}

// showError displays an error message to the user in a formatted way using huh forms.
// The user must press Enter to acknowledge the error.
//
// Returns nil after the error is shown (error is not propagated).
//
// Example usage:
//
//	if isProtectedBranch(branchName) {
//	    return showError("Cannot use protected branch name: " + branchName)
//	}
func showError(message string) error {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Error").
				Description(message),
		),
	)

	// Run the form (user presses Enter to acknowledge)
	_ = form.Run()

	// Return nil - error is shown, not propagated
	return nil
}

// withSpinner wraps a long-running operation with a loading spinner.
// The spinner displays the provided title while the action is running.
//
// If the action returns an error, it is propagated to the caller.
// If the action succeeds, nil is returned.
//
// Example usage:
//
//	var issues []*sow.Issue
//	err := withSpinner("Fetching GitHub issues...", func() error {
//	    var err error
//	    issues, err = sow.ListIssues(ctx)
//	    return err
//	})
//	if err != nil {
//	    return fmt.Errorf("failed to fetch issues: %w", err)
//	}
func withSpinner(title string, action func() error) error {
	var err error

	_ = spinner.New().
		Title(title).
		Action(func() {
			err = action()
		}).
		Run()

	return err
}
