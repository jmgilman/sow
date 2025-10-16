package refs

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// gitSSHShorthandRegex matches git SSH shorthand format: git@host:path.
var gitSSHShorthandRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+@[a-zA-Z0-9._-]+:[^/]+`)

// InferTypeFromScheme infers the ref type from a URL scheme.
//
// Scheme patterns:
//   - git+https, git+ssh, git@ → "git"
//   - file → "file"
//   - web+https, web+http → "web" (future)
func InferTypeFromScheme(scheme string) string {
	// Handle compound schemes (e.g., "git+https")
	if strings.Contains(scheme, "+") {
		parts := strings.Split(scheme, "+")
		return parts[0]
	}

	// Handle direct schemes
	switch scheme {
	case "file":
		return "file"
	case "git":
		return "git"
	case "https", "http", "ssh":
		// These might be git URLs without the git+ prefix
		// The caller should check the full URL context
		return "unknown"
	default:
		return "unknown"
	}
}

// InferTypeFromURL infers the ref type from a complete URL.
//
// Handles:
//   - git+https://github.com/org/repo
//   - git+ssh://git@github.com/org/repo
//   - git@github.com:org/repo (auto-converted to git+ssh://)
//   - file:///absolute/path
func InferTypeFromURL(rawURL string) (string, error) {
	// Check for git SSH shorthand first (git@host:path)
	if gitSSHShorthandRegex.MatchString(rawURL) {
		return "git", nil
	}

	// Parse as URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Infer from scheme
	typeName := InferTypeFromScheme(u.Scheme)
	if typeName == "unknown" {
		return "", fmt.Errorf("cannot infer type from URL: %s", rawURL)
	}

	return typeName, nil
}

// NormalizeGitURL normalizes git URLs to standard format.
//
// Conversions:
//   - git@github.com:org/repo → git+ssh://git@github.com/org/repo
//   - https://github.com/org/repo → git+https://github.com/org/repo (if not already prefixed)
//   - ssh://git@github.com/org/repo → git+ssh://git@github.com/org/repo (if not already prefixed)
func NormalizeGitURL(rawURL string) (string, error) {
	// Check for git SSH shorthand (git@host:path)
	if gitSSHShorthandRegex.MatchString(rawURL) {
		// Convert: git@github.com:org/repo → git+ssh://git@github.com/org/repo
		parts := strings.Split(rawURL, ":")
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid git SSH shorthand: %s", rawURL)
		}
		return fmt.Sprintf("git+ssh://%s/%s", parts[0], parts[1]), nil
	}

	// Parse as URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// If already has git+ prefix, return as-is
	if strings.HasPrefix(u.Scheme, "git+") {
		return rawURL, nil
	}

	// Add git+ prefix based on scheme
	switch u.Scheme {
	case "https", "http":
		u.Scheme = "git+" + u.Scheme
		return u.String(), nil
	case "ssh":
		u.Scheme = "git+ssh"
		return u.String(), nil
	case "file":
		// Local git repository accessed via file:// protocol
		u.Scheme = "git+file"
		return u.String(), nil
	case "git":
		// Already git protocol, just return
		return rawURL, nil
	default:
		return "", fmt.Errorf("cannot normalize non-git URL: %s", rawURL)
	}
}

// ParseGitURL parses a git URL and extracts components.
//
// Returns the normalized URL, transport (https/ssh), and error.
func ParseGitURL(rawURL string) (normalizedURL, transport string, err error) {
	// Normalize first
	normalized, err := NormalizeGitURL(rawURL)
	if err != nil {
		return "", "", err
	}

	// Parse normalized URL
	u, err := url.Parse(normalized)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse normalized URL: %w", err)
	}

	// Extract transport from compound scheme
	if strings.HasPrefix(u.Scheme, "git+") {
		transport = strings.TrimPrefix(u.Scheme, "git+")
	} else {
		transport = u.Scheme
	}

	return normalized, transport, nil
}

// ValidateFileURL validates a file:// URL.
//
// Requirements:
//   - Must start with file:///
//   - Must be absolute path (starts with /)
func ValidateFileURL(rawURL string) error {
	if !strings.HasPrefix(rawURL, "file:///") {
		return fmt.Errorf("file URLs must start with file:/// (absolute path)")
	}

	// Parse to validate format
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid file URL: %w", err)
	}

	// Check path is absolute
	if !strings.HasPrefix(u.Path, "/") {
		return fmt.Errorf("file URL must contain absolute path")
	}

	return nil
}

// FileURLToPath converts a file:// URL to a filesystem path.
//
// Example: file:///Users/josh/docs → /Users/josh/docs.
func FileURLToPath(rawURL string) (string, error) {
	if err := ValidateFileURL(rawURL); err != nil {
		return "", err
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse file URL: %w", err)
	}

	return u.Path, nil
}

// PathToFileURL converts a filesystem path to a file:// URL.
//
// Path must be absolute (start with /).
func PathToFileURL(path string) (string, error) {
	if !strings.HasPrefix(path, "/") {
		return "", fmt.Errorf("path must be absolute: %s", path)
	}

	return "file://" + path, nil
}
