package refs

import (
	"context"
	"testing"
)

func TestGetType(t *testing.T) {
	tests := []struct {
		name      string
		typeName  string
		wantError bool
	}{
		{
			name:      "git type exists",
			typeName:  "git",
			wantError: false,
		},
		{
			name:      "file type exists",
			typeName:  "file",
			wantError: false,
		},
		{
			name:      "unknown type",
			typeName:  "unknown",
			wantError: true,
		},
		{
			name:      "empty string",
			typeName:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, err := GetType(tt.typeName)
			if (err != nil) != tt.wantError {
				t.Errorf("GetType(%q) error = %v, wantError %v", tt.typeName, err, tt.wantError)
				return
			}
			if !tt.wantError && gotType == nil {
				t.Errorf("GetType(%q) returned nil type", tt.typeName)
			}
			if !tt.wantError && gotType.Name() != tt.typeName {
				t.Errorf("GetType(%q).Name() = %q, want %q", tt.typeName, gotType.Name(), tt.typeName)
			}
		})
	}
}

func TestAllTypes(t *testing.T) {
	types := AllTypes()

	// Should have at least git and file types
	if len(types) < 2 {
		t.Errorf("AllTypes() returned %d types, want at least 2", len(types))
	}

	// Verify git and file are present
	foundGit := false
	foundFile := false
	for _, typ := range types {
		switch typ.Name() {
		case "git":
			foundGit = true
		case "file":
			foundFile = true
		}
	}

	if !foundGit {
		t.Error("AllTypes() missing git type")
	}
	if !foundFile {
		t.Error("AllTypes() missing file type")
	}
}

func TestEnabledTypes(t *testing.T) {
	ctx := context.Background()
	types, err := EnabledTypes(ctx)
	if err != nil {
		t.Fatalf("EnabledTypes() error = %v", err)
	}

	// File type should always be enabled
	foundFile := false
	for _, typ := range types {
		if typ.Name() == "file" {
			foundFile = true
			break
		}
	}

	if !foundFile {
		t.Error("EnabledTypes() missing file type (should always be enabled)")
	}

	// Git type enabled depends on system (git binary in PATH)
	// We can't reliably test this, but we can verify the function runs
}

func TestDisabledTypes(t *testing.T) {
	ctx := context.Background()
	types, err := DisabledTypes(ctx)
	if err != nil {
		t.Fatalf("DisabledTypes() error = %v", err)
	}

	// File type should never be disabled
	for _, typ := range types {
		if typ.Name() == "file" {
			t.Error("DisabledTypes() includes file type (should never be disabled)")
		}
	}
}

func TestTypeForScheme(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		scheme    string
		wantType  string
		wantError bool
	}{
		{
			name:      "git scheme",
			scheme:    "git",
			wantType:  "git",
			wantError: false,
		},
		{
			name:      "file scheme",
			scheme:    "file",
			wantType:  "file",
			wantError: false,
		},
		{
			name:      "unknown scheme",
			scheme:    "unknown",
			wantType:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, err := TypeForScheme(ctx, tt.scheme)
			if (err != nil) != tt.wantError {
				t.Errorf("TypeForScheme(%q) error = %v, wantError %v", tt.scheme, err, tt.wantError)
				return
			}
			if !tt.wantError && gotType.Name() != tt.wantType {
				t.Errorf("TypeForScheme(%q).Name() = %q, want %q", tt.scheme, gotType.Name(), tt.wantType)
			}
		})
	}
}
