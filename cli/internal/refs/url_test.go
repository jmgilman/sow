package refs

import (
	"testing"
)

func TestInferTypeFromScheme(t *testing.T) {
	tests := []struct {
		name     string
		scheme   string
		expected string
	}{
		{"git+https", "git+https", "git"},
		{"git+ssh", "git+ssh", "git"},
		{"file", "file", "file"},
		{"git", "git", "git"},
		{"https alone", "https", "unknown"},
		{"ssh alone", "ssh", "unknown"},
		{"unknown", "ftp", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InferTypeFromScheme(tt.scheme)
			if result != tt.expected {
				t.Errorf("InferTypeFromScheme(%q) = %q, want %q", tt.scheme, result, tt.expected)
			}
		})
	}
}

func TestInferTypeFromURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantType  string
		wantError bool
	}{
		{
			name:      "git+https URL",
			url:       "git+https://github.com/org/repo",
			wantType:  "git",
			wantError: false,
		},
		{
			name:      "git+ssh URL",
			url:       "git+ssh://git@github.com/org/repo",
			wantType:  "git",
			wantError: false,
		},
		{
			name:      "git SSH shorthand",
			url:       "git@github.com:org/repo",
			wantType:  "git",
			wantError: false,
		},
		{
			name:      "file URL",
			url:       "file:///Users/josh/docs",
			wantType:  "file",
			wantError: false,
		},
		{
			name:      "https without git prefix",
			url:       "https://github.com/org/repo",
			wantType:  "",
			wantError: true,
		},
		{
			name:      "invalid URL",
			url:       "not a url",
			wantType:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, err := InferTypeFromURL(tt.url)
			if (err != nil) != tt.wantError {
				t.Errorf("InferTypeFromURL() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if gotType != tt.wantType {
				t.Errorf("InferTypeFromURL() = %q, want %q", gotType, tt.wantType)
			}
		})
	}
}

func TestNormalizeGitURL(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  string
		wantError bool
	}{
		{
			name:      "SSH shorthand",
			input:     "git@github.com:org/repo",
			expected:  "git+ssh://git@github.com/org/repo",
			wantError: false,
		},
		{
			name:      "SSH shorthand with dots in username",
			input:     "git@git.example.com:org/repo",
			expected:  "git+ssh://git@git.example.com/org/repo",
			wantError: false,
		},
		{
			name:      "HTTPS without prefix",
			input:     "https://github.com/org/repo",
			expected:  "git+https://github.com/org/repo",
			wantError: false,
		},
		{
			name:      "HTTP without prefix",
			input:     "http://github.com/org/repo",
			expected:  "git+http://github.com/org/repo",
			wantError: false,
		},
		{
			name:      "SSH without prefix",
			input:     "ssh://git@github.com/org/repo",
			expected:  "git+ssh://git@github.com/org/repo",
			wantError: false,
		},
		{
			name:      "Already normalized git+https",
			input:     "git+https://github.com/org/repo",
			expected:  "git+https://github.com/org/repo",
			wantError: false,
		},
		{
			name:      "Already normalized git+ssh",
			input:     "git+ssh://git@github.com/org/repo",
			expected:  "git+ssh://git@github.com/org/repo",
			wantError: false,
		},
		{
			name:      "Invalid SSH shorthand",
			input:     "git@github.com",
			expected:  "",
			wantError: true,
		},
		{
			name:      "Local git repo via file URL",
			input:     "file:///some/path",
			expected:  "git+file:///some/path",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeGitURL(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("NormalizeGitURL() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if result != tt.expected {
				t.Errorf("NormalizeGitURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseGitURL(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantURL       string
		wantTransport string
		wantError     bool
	}{
		{
			name:          "SSH shorthand",
			input:         "git@github.com:org/repo",
			wantURL:       "git+ssh://git@github.com/org/repo",
			wantTransport: "ssh",
			wantError:     false,
		},
		{
			name:          "HTTPS URL",
			input:         "https://github.com/org/repo",
			wantURL:       "git+https://github.com/org/repo",
			wantTransport: "https",
			wantError:     false,
		},
		{
			name:          "SSH URL",
			input:         "ssh://git@github.com/org/repo",
			wantURL:       "git+ssh://git@github.com/org/repo",
			wantTransport: "ssh",
			wantError:     false,
		},
		{
			name:          "Already normalized",
			input:         "git+https://github.com/org/repo",
			wantURL:       "git+https://github.com/org/repo",
			wantTransport: "https",
			wantError:     false,
		},
		{
			name:          "Local git repo via file URL",
			input:         "file:///path/to/repo",
			wantURL:       "git+file:///path/to/repo",
			wantTransport: "file",
			wantError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, gotTransport, err := ParseGitURL(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("ParseGitURL() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if gotURL != tt.wantURL {
				t.Errorf("ParseGitURL() URL = %q, want %q", gotURL, tt.wantURL)
			}
			if gotTransport != tt.wantTransport {
				t.Errorf("ParseGitURL() transport = %q, want %q", gotTransport, tt.wantTransport)
			}
		})
	}
}

func TestValidateFileURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantError bool
	}{
		{
			name:      "valid absolute path",
			url:       "file:///Users/josh/docs",
			wantError: false,
		},
		{
			name:      "valid absolute path (root)",
			url:       "file:///",
			wantError: false,
		},
		{
			name:      "missing triple slash",
			url:       "file://relative/path",
			wantError: true,
		},
		{
			name:      "missing file scheme",
			url:       "/Users/josh/docs",
			wantError: true,
		},
		{
			name:      "relative path",
			url:       "file://docs",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFileURL(tt.url)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateFileURL(%q) error = %v, wantError %v", tt.url, err, tt.wantError)
			}
		})
	}
}

func TestFileURLToPath(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantPath  string
		wantError bool
	}{
		{
			name:      "valid URL",
			url:       "file:///Users/josh/docs",
			wantPath:  "/Users/josh/docs",
			wantError: false,
		},
		{
			name:      "root path",
			url:       "file:///",
			wantPath:  "/",
			wantError: false,
		},
		{
			name:      "invalid URL",
			url:       "file://relative",
			wantPath:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, err := FileURLToPath(tt.url)
			if (err != nil) != tt.wantError {
				t.Errorf("FileURLToPath() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if gotPath != tt.wantPath {
				t.Errorf("FileURLToPath() = %q, want %q", gotPath, tt.wantPath)
			}
		})
	}
}

func TestPathToFileURL(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantURL   string
		wantError bool
	}{
		{
			name:      "absolute path",
			path:      "/Users/josh/docs",
			wantURL:   "file:///Users/josh/docs",
			wantError: false,
		},
		{
			name:      "root path",
			path:      "/",
			wantURL:   "file:///",
			wantError: false,
		},
		{
			name:      "relative path",
			path:      "docs",
			wantURL:   "",
			wantError: true,
		},
		{
			name:      "relative path with dot",
			path:      "./docs",
			wantURL:   "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotURL, err := PathToFileURL(tt.path)
			if (err != nil) != tt.wantError {
				t.Errorf("PathToFileURL() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if gotURL != tt.wantURL {
				t.Errorf("PathToFileURL() = %q, want %q", gotURL, tt.wantURL)
			}
		})
	}
}
