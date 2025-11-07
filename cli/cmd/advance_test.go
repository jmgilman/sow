package cmd

import (
	"testing"
)

// TestAdvanceCommandSignature verifies that the advance command accepts the correct arguments and has the required flags
func TestAdvanceCommandSignature(t *testing.T) {
	cmd := NewAdvanceCmd()

	// Test accepts 0 arguments
	err := cmd.Args(cmd, []string{})
	if err != nil {
		t.Errorf("should accept 0 arguments: %v", err)
	}

	// Test accepts 1 argument
	err = cmd.Args(cmd, []string{"event_name"})
	if err != nil {
		t.Errorf("should accept 1 argument: %v", err)
	}

	// Test rejects 2 arguments
	err = cmd.Args(cmd, []string{"event1", "event2"})
	if err == nil {
		t.Error("should reject 2 arguments")
	}

	// Test --list flag is defined
	listFlag := cmd.Flags().Lookup("list")
	if listFlag == nil {
		t.Error("--list flag not defined")
	}
	if listFlag != nil && listFlag.Value.Type() != "bool" {
		t.Errorf("--list flag should be boolean, got %s", listFlag.Value.Type())
	}

	// Test --dry-run flag is defined
	dryRunFlag := cmd.Flags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Error("--dry-run flag not defined")
	}
	if dryRunFlag != nil && dryRunFlag.Value.Type() != "bool" {
		t.Errorf("--dry-run flag should be boolean, got %s", dryRunFlag.Value.Type())
	}
}

// TestAdvanceFlagValidation verifies mutual exclusivity rules for flags and arguments
func TestAdvanceFlagValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		listFlag    bool
		dryRunFlag  bool
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "no flags, no args (auto mode) - valid",
			args:        []string{},
			listFlag:    false,
			dryRunFlag:  false,
			shouldError: false,
		},
		{
			name:        "no flags, one arg (explicit event mode) - valid",
			args:        []string{"finalize"},
			listFlag:    false,
			dryRunFlag:  false,
			shouldError: false,
		},
		{
			name:        "list flag, no args (discovery mode) - valid",
			args:        []string{},
			listFlag:    true,
			dryRunFlag:  false,
			shouldError: false,
		},
		{
			name:        "dry-run flag, one arg (dry-run mode) - valid",
			args:        []string{"finalize"},
			listFlag:    false,
			dryRunFlag:  true,
			shouldError: false,
		},
		{
			name:        "list flag with event argument - invalid",
			args:        []string{"finalize"},
			listFlag:    true,
			dryRunFlag:  false,
			shouldError: true,
			errorMsg:    "cannot specify event argument with --list flag",
		},
		{
			name:        "dry-run flag without event argument - invalid",
			args:        []string{},
			listFlag:    false,
			dryRunFlag:  true,
			shouldError: true,
			errorMsg:    "--dry-run requires an event argument",
		},
		{
			name:        "both list and dry-run flags - invalid",
			args:        []string{},
			listFlag:    true,
			dryRunFlag:  true,
			shouldError: true,
			errorMsg:    "cannot use --list and --dry-run together",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh command for each test
			cmd := NewAdvanceCmd()

			// Set flags
			cmd.Flags().Set("list", boolToString(tt.listFlag))
			cmd.Flags().Set("dry-run", boolToString(tt.dryRunFlag))

			// Call the validation function that should be in RunE
			err := validateAdvanceFlags(cmd, tt.args)

			if tt.shouldError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("expected error message %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

// Helper function to convert bool to string for flag setting
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
