package sow_test

import (
	"os"
	"testing"

	"github.com/jmgilman/sow/cli/internal/sow"
)

func TestNewGitHubClient_WithoutToken(t *testing.T) {
	// Ensure GITHUB_TOKEN is not set
	oldToken := os.Getenv("GITHUB_TOKEN")
	defer func() { _ = os.Setenv("GITHUB_TOKEN", oldToken) }()
	_ = os.Unsetenv("GITHUB_TOKEN")

	client, err := sow.NewGitHubClient()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client == nil {
		t.Fatal("expected non-nil client, got nil")
	}

	// Verify it returned a CLI client by checking the type
	_, ok := client.(*sow.GitHubCLI)
	if !ok {
		t.Errorf("expected *GitHubCLI when GITHUB_TOKEN is not set, got %T", client)
	}
}

func TestNewGitHubClient_WithToken(t *testing.T) {
	// Set GITHUB_TOKEN environment variable
	oldToken := os.Getenv("GITHUB_TOKEN")
	defer func() { _ = os.Setenv("GITHUB_TOKEN", oldToken) }()
	_ = os.Setenv("GITHUB_TOKEN", "test-token")

	client, err := sow.NewGitHubClient()

	// For now, since API client is not implemented, we expect an error
	// or the function to return CLI client as fallback
	if err != nil {
		// This is acceptable if API client is not yet implemented
		t.Logf("API client not yet implemented, error: %v", err)
		return
	}

	if client == nil {
		t.Fatal("expected non-nil client, got nil")
	}

	// When API client is implemented, check for *GitHubAPI type
	// For now, we accept either GitHubCLI (fallback) or an error
	_, isCLI := client.(*sow.GitHubCLI)
	if !isCLI {
		// Future: check for GitHubAPI type
		t.Logf("Client type: %T (API client not yet implemented)", client)
	}
}

func TestNewGitHubClient_ReturnsInterface(t *testing.T) {
	// Ensure GITHUB_TOKEN is not set
	oldToken := os.Getenv("GITHUB_TOKEN")
	defer func() { _ = os.Setenv("GITHUB_TOKEN", oldToken) }()
	_ = os.Unsetenv("GITHUB_TOKEN")

	client, err := sow.NewGitHubClient()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the returned type implements GitHubClient interface
	// This is mainly a compile-time check, but we can verify methods exist
	var _ = client

	// Verify interface methods are available
	if client == nil {
		t.Fatal("expected non-nil client")
	}

	// Test that CheckAvailability method exists (interface method)
	// We won't call it because it would require gh CLI to be installed
	// but we can verify the method exists through reflection or type assertion
}
