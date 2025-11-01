package project

import "strings"

// DetectProjectType determines project type from branch name convention.
// Returns "standard" for unknown prefixes (default).
//
// Branch prefix mapping:
//   - explore/ → exploration
//   - design/ → design
//   - breakdown/ → breakdown
//   - all others → standard (default)
func DetectProjectType(branchName string) string {
	switch {
	case strings.HasPrefix(branchName, "explore/"):
		return "exploration"
	case strings.HasPrefix(branchName, "design/"):
		return "design"
	case strings.HasPrefix(branchName, "breakdown/"):
		return "breakdown"
	default:
		return "standard"
	}
}
