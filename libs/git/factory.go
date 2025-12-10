package git

import (
	"os"

	"github.com/jmgilman/sow/libs/exec"
)

// NewGitHubClient creates a GitHub client with automatic environment detection.
//
// Currently returns GitHubCLI. Future: could return GitHubAPI if GITHUB_TOKEN is set.
func NewGitHubClient() (GitHubClient, error) {
	// Check for GITHUB_TOKEN environment variable
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		// API client not yet implemented
		// For now, fall back to CLI client
		// TODO: return NewGitHubAPI(token), nil
		_ = token // Suppress unused variable warning until API client is implemented
	}

	// Default to CLI client
	ghExec := exec.NewLocalExecutor("gh")
	return NewGitHubCLI(ghExec), nil
}
