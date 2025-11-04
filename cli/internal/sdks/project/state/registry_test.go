package state

import "testing"

// TestRegisterAddsConfigToRegistry verifies that Register() adds a config to the registry.
func TestRegisterAddsConfigToRegistry(t *testing.T) {
	// Clear registry for test isolation
	Registry = make(map[string]ProjectTypeConfig)

	config := newMockProjectTypeConfig()
	Register("test", config)

	if len(Registry) != 1 {
		t.Errorf("expected registry to have 1 entry, got %d", len(Registry))
	}

	if _, exists := Registry["test"]; !exists {
		t.Error("expected 'test' to be in registry")
	}
}

// TestRegisterStoresConfigUnderCorrectName verifies that Register() stores config under the correct name.
func TestRegisterStoresConfigUnderCorrectName(t *testing.T) {
	// Clear registry for test isolation
	Registry = make(map[string]ProjectTypeConfig)

	config := newMockProjectTypeConfig()
	Register("mytype", config)

	storedConfig, exists := Registry["mytype"]
	if !exists {
		t.Fatal("expected 'mytype' to be in registry")
	}

	if storedConfig != config {
		t.Error("stored config does not match registered config")
	}
}

// TestRegisterDuplicatePanics verifies that Register() panics when registering a duplicate name.
func TestRegisterDuplicatePanics(t *testing.T) {
	// Clear registry for test isolation
	Registry = make(map[string]ProjectTypeConfig)

	config := newMockProjectTypeConfig()
	Register("test", config)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on duplicate registration")
		} else {
			// Verify panic message contains helpful information
			msg, ok := r.(string)
			if !ok {
				t.Error("expected panic message to be a string")
			}
			if msg != "project type already registered: test" {
				t.Errorf("expected panic message to mention 'test', got: %s", msg)
			}
		}
	}()

	Register("test", config) // Should panic
}

// TestRegisterMultipleDifferentTypes verifies that multiple different types can be registered.
func TestRegisterMultipleDifferentTypes(t *testing.T) {
	// Clear registry for test isolation
	Registry = make(map[string]ProjectTypeConfig)

	config1 := newMockProjectTypeConfig()
	config2 := newMockProjectTypeConfig()
	config3 := newMockProjectTypeConfig()

	Register("type1", config1)
	Register("type2", config2)
	Register("type3", config3)

	if len(Registry) != 3 {
		t.Errorf("expected registry to have 3 entries, got %d", len(Registry))
	}

	if Registry["type1"] != config1 {
		t.Error("type1 config mismatch")
	}
	if Registry["type2"] != config2 {
		t.Error("type2 config mismatch")
	}
	if Registry["type3"] != config3 {
		t.Error("type3 config mismatch")
	}
}

// TestGetReturnsConfigForRegisteredType verifies that Get() returns (config, true) for a registered type.
func TestGetReturnsConfigForRegisteredType(t *testing.T) {
	// Clear registry for test isolation
	Registry = make(map[string]ProjectTypeConfig)

	config := newMockProjectTypeConfig()
	Registry["test"] = config

	gotConfig, exists := Get("test")

	if !exists {
		t.Error("expected Get to return true for registered type")
	}

	if gotConfig != config {
		t.Error("expected Get to return the registered config")
	}
}

// TestGetReturnsCorrectConfigForRegisteredType verifies that Get() returns the correct config.
func TestGetReturnsCorrectConfigForRegisteredType(t *testing.T) {
	// Clear registry for test isolation
	Registry = make(map[string]ProjectTypeConfig)

	config1 := newMockProjectTypeConfig()
	config2 := newMockProjectTypeConfig()
	Registry["type1"] = config1
	Registry["type2"] = config2

	gotConfig1, _ := Get("type1")
	gotConfig2, _ := Get("type2")

	if gotConfig1 != config1 {
		t.Error("type1 config mismatch")
	}
	if gotConfig2 != config2 {
		t.Error("type2 config mismatch")
	}
}

// TestGetReturnsNilForUnregisteredType verifies that Get() returns (nil, false) for an unregistered type.
func TestGetReturnsNilForUnregisteredType(t *testing.T) {
	// Clear registry for test isolation
	Registry = make(map[string]ProjectTypeConfig)

	gotConfig, exists := Get("nonexistent")

	if exists {
		t.Error("expected Get to return false for unregistered type")
	}

	if gotConfig != nil {
		t.Error("expected Get to return nil for unregistered type")
	}
}

// TestGetWorksAfterMultipleTypesRegistered verifies that Get() works correctly after multiple types are registered.
func TestGetWorksAfterMultipleTypesRegistered(t *testing.T) {
	// Clear registry for test isolation
	Registry = make(map[string]ProjectTypeConfig)

	config1 := newMockProjectTypeConfig()
	config2 := newMockProjectTypeConfig()
	config3 := newMockProjectTypeConfig()

	Register("type1", config1)
	Register("type2", config2)
	Register("type3", config3)

	// Verify all can be retrieved
	gotConfig1, exists1 := Get("type1")
	gotConfig2, exists2 := Get("type2")
	gotConfig3, exists3 := Get("type3")

	if !exists1 || !exists2 || !exists3 {
		t.Error("expected all registered types to be found")
	}

	if gotConfig1 != config1 || gotConfig2 != config2 || gotConfig3 != config3 {
		t.Error("retrieved configs do not match registered configs")
	}
}

// TestRegisterThenGetIntegration verifies that the integration between Register and Get.
func TestRegisterThenGetIntegration(t *testing.T) {
	// Clear registry for test isolation
	Registry = make(map[string]ProjectTypeConfig)

	config := newMockProjectTypeConfig()
	Register("integration", config)

	gotConfig, exists := Get("integration")

	if !exists {
		t.Fatal("expected to find registered type")
	}

	if gotConfig != config {
		t.Error("retrieved config does not match registered config")
	}

	// Verify it's the exact same instance
	if gotConfig != config {
		t.Error("expected Get to return the same instance that was registered")
	}
}
