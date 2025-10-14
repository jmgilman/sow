package config

// Version info injected via ldflags at build time
var (
	// Version is the semantic version of the CLI
	Version = "dev"

	// BuildDate is the ISO 8601 timestamp when the binary was built
	BuildDate = "unknown"

	// Commit is the git commit hash
	Commit = "none"
)
