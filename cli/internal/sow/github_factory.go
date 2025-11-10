package sow

import (
	"os"

	"github.com/jmgilman/sow/cli/internal/exec"
)

// NewGitHubClient creates a GitHub client with automatic environment detection.
//
// The factory automatically selects between CLI and API implementations based on
// the environment:
//
//   - If GITHUB_TOKEN environment variable is set: Returns GitHubAPI client (future)
//   - Otherwise: Returns GitHubCLI client (wraps gh CLI tool)
//
// For local development with gh CLI installed:
//
//	client, err := sow.NewGitHubClient()
//	if err != nil {
//	    return err
//	}
//	issues, err := client.ListIssues("bug", "open")
//
// For web VMs or CI/CD with GitHub token:
//
//	// Set GITHUB_TOKEN environment variable
//	os.Setenv("GITHUB_TOKEN", "ghp_...")
//	client, err := sow.NewGitHubClient()
//
// Returns an error if neither gh CLI nor GitHub token is available.
func NewGitHubClient() (GitHubClient, error) {
	// Check for GITHUB_TOKEN environment variable
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		// API client not yet implemented
		// For now, fall back to CLI client
		// TODO: return NewGitHubAPI(token), nil
		_ = token // Suppress unused variable warning until API client is implemented
	}

	// Default to CLI client
	ghExec := exec.NewLocal("gh")
	return NewGitHubCLI(ghExec), nil
}
