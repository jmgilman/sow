package project

import (
	"testing"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
	sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
)

// Phase Options Tests

func TestWithStartState(t *testing.T) {
	config := &PhaseConfig{}
	opt := WithStartState(sdkstate.State("TestState"))
	opt(config)

	if config.startState != sdkstate.State("TestState") {
		t.Errorf("expected startState TestState, got %v", config.startState)
	}
}

func TestWithEndState(t *testing.T) {
	config := &PhaseConfig{}
	opt := WithEndState(sdkstate.State("CompletedState"))
	opt(config)

	if config.endState != sdkstate.State("CompletedState") {
		t.Errorf("expected endState CompletedState, got %v", config.endState)
	}
}

func TestWithInputs(t *testing.T) {
	config := &PhaseConfig{}
	opt := WithInputs("doc", "spec")
	opt(config)

	if len(config.allowedInputTypes) != 2 {
		t.Errorf("expected 2 input types, got %d", len(config.allowedInputTypes))
	}

	if config.allowedInputTypes[0] != "doc" || config.allowedInputTypes[1] != "spec" {
		t.Errorf("expected input types [doc, spec], got %v", config.allowedInputTypes)
	}
}

func TestWithInputsSingleType(t *testing.T) {
	config := &PhaseConfig{}
	opt := WithInputs("doc")
	opt(config)

	if len(config.allowedInputTypes) != 1 {
		t.Errorf("expected 1 input type, got %d", len(config.allowedInputTypes))
	}

	if config.allowedInputTypes[0] != "doc" {
		t.Errorf("expected input type doc, got %v", config.allowedInputTypes[0])
	}
}

func TestWithOutputs(t *testing.T) {
	config := &PhaseConfig{}
	opt := WithOutputs("plan", "code", "tests")
	opt(config)

	if len(config.allowedOutputTypes) != 3 {
		t.Errorf("expected 3 output types, got %d", len(config.allowedOutputTypes))
	}

	expected := []string{"plan", "code", "tests"}
	for i, v := range expected {
		if config.allowedOutputTypes[i] != v {
			t.Errorf("expected output type %s at index %d, got %s", v, i, config.allowedOutputTypes[i])
		}
	}
}

func TestWithOutputsSingleType(t *testing.T) {
	config := &PhaseConfig{}
	opt := WithOutputs("report")
	opt(config)

	if len(config.allowedOutputTypes) != 1 {
		t.Errorf("expected 1 output type, got %d", len(config.allowedOutputTypes))
	}

	if config.allowedOutputTypes[0] != "report" {
		t.Errorf("expected output type report, got %v", config.allowedOutputTypes[0])
	}
}

func TestWithTasks(t *testing.T) {
	config := &PhaseConfig{}

	// Should be false by default
	if config.supportsTasks {
		t.Error("expected supportsTasks to be false by default")
	}

	opt := WithTasks()
	opt(config)

	if !config.supportsTasks {
		t.Error("expected supportsTasks to be true after applying WithTasks()")
	}
}

func TestWithMetadataSchema(t *testing.T) {
	config := &PhaseConfig{}
	schema := "#Phase: { foo: string }"
	opt := WithMetadataSchema(schema)
	opt(config)

	if config.metadataSchema != schema {
		t.Errorf("expected metadataSchema %q, got %q", schema, config.metadataSchema)
	}
}

func TestMultiplePhaseOptions(t *testing.T) {
	config := &PhaseConfig{}

	// Apply multiple options
	WithStartState(sdkstate.State("Planning"))(config)
	WithEndState(sdkstate.State("PlanComplete"))(config)
	WithInputs("requirements")(config)
	WithOutputs("plan", "estimate")(config)
	WithTasks()(config)
	WithMetadataSchema("#Schema")(config)

	// Verify all options were applied
	if config.startState != sdkstate.State("Planning") {
		t.Errorf("expected startState Planning, got %v", config.startState)
	}
	if config.endState != sdkstate.State("PlanComplete") {
		t.Errorf("expected endState PlanComplete, got %v", config.endState)
	}
	if len(config.allowedInputTypes) != 1 || config.allowedInputTypes[0] != "requirements" {
		t.Errorf("expected inputs [requirements], got %v", config.allowedInputTypes)
	}
	if len(config.allowedOutputTypes) != 2 {
		t.Errorf("expected 2 outputs, got %d", len(config.allowedOutputTypes))
	}
	if !config.supportsTasks {
		t.Error("expected supportsTasks to be true")
	}
	if config.metadataSchema != "#Schema" {
		t.Errorf("expected metadataSchema #Schema, got %s", config.metadataSchema)
	}
}

// Transition Options Tests

func TestWithGuard(t *testing.T) {
	config := &TransitionConfig{}

	guardFunc := func(_ *state.Project) bool {
		return true
	}

	opt := WithGuard("test guard", guardFunc)
	opt(config)

	if config.guardTemplate.Func == nil {
		t.Error("expected guardTemplate.Func to be set")
	}

	if config.guardTemplate.Description != "test guard" {
		t.Errorf("expected description 'test guard', got %q", config.guardTemplate.Description)
	}

	// Test that the guard function works
	if !config.guardTemplate.Func(nil) {
		t.Error("expected guard function to return true")
	}
}

func TestWithOnEntry(t *testing.T) {
	config := &TransitionConfig{}

	called := false
	entryAction := func(_ *state.Project) error {
		called = true
		return nil
	}

	opt := WithOnEntry(entryAction)
	opt(config)

	if config.onEntry == nil {
		t.Error("expected onEntry to be set")
	}

	// Test that the action can be called
	_ = config.onEntry(nil)
	if !called {
		t.Error("expected onEntry action to be called")
	}
}

func TestWithOnExit(t *testing.T) {
	config := &TransitionConfig{}

	called := false
	exitAction := func(_ *state.Project) error {
		called = true
		return nil
	}

	opt := WithOnExit(exitAction)
	opt(config)

	if config.onExit == nil {
		t.Error("expected onExit to be set")
	}

	// Test that the action can be called
	_ = config.onExit(nil)
	if !called {
		t.Error("expected onExit action to be called")
	}
}

func TestMultipleTransitionOptions(t *testing.T) {
	config := &TransitionConfig{}

	guardCalled := false
	entryCalled := false
	exitCalled := false

	guardFunc := func(_ *state.Project) bool {
		guardCalled = true
		return true
	}

	entryAction := func(_ *state.Project) error {
		entryCalled = true
		return nil
	}

	exitAction := func(_ *state.Project) error {
		exitCalled = true
		return nil
	}

	// Apply multiple options
	WithGuard("test guard", guardFunc)(config)
	WithOnEntry(entryAction)(config)
	WithOnExit(exitAction)(config)

	// Verify all options were applied
	if config.guardTemplate.Func == nil {
		t.Error("expected guardTemplate.Func to be set")
	}
	if config.onEntry == nil {
		t.Error("expected onEntry to be set")
	}
	if config.onExit == nil {
		t.Error("expected onExit to be set")
	}

	// Test that all functions work
	config.guardTemplate.Func(nil)
	_ = config.onEntry(nil)
	_ = config.onExit(nil)

	if !guardCalled || !entryCalled || !exitCalled {
		t.Error("expected all transition functions to be called")
	}
}

func TestWithFailedPhase(t *testing.T) {
	config := &TransitionConfig{}

	opt := WithFailedPhase("review")
	opt(config)

	if config.failedPhase != "review" {
		t.Errorf("expected failedPhase 'review', got %q", config.failedPhase)
	}
}

func TestWithDescription(t *testing.T) {
	t.Run("sets description on transition config", func(t *testing.T) {
		config := &TransitionConfig{}
		opt := WithDescription("Test transition description")

		opt(config)

		if config.description != "Test transition description" {
			t.Errorf("expected description 'Test transition description', got %q", config.description)
		}
	})

	t.Run("works with other options", func(t *testing.T) {
		config := &TransitionConfig{}
		guardFunc := func(_ *state.Project) bool { return true }

		// Apply multiple options
		WithGuard("test guard", guardFunc)(config)
		WithDescription("Test description")(config)

		// Verify both options were applied
		if config.guardTemplate.Description != "test guard" {
			t.Errorf("expected guard description 'test guard', got %q", config.guardTemplate.Description)
		}
		if config.description != "Test description" {
			t.Errorf("expected description 'Test description', got %q", config.description)
		}
		if config.guardTemplate.Func == nil {
			t.Error("expected guardTemplate.Func to be set")
		}
	})

	t.Run("empty description is allowed", func(t *testing.T) {
		config := &TransitionConfig{}
		opt := WithDescription("")

		opt(config)

		// Empty string should be set without error
		if config.description != "" {
			t.Errorf("expected empty description, got %q", config.description)
		}
	})
}

// Test that options can be applied in any order.
func TestOptionsCanBeAppliedInAnyOrder(t *testing.T) {
	// Test phase options in different order
	config1 := &PhaseConfig{}
	WithStartState(sdkstate.State("A"))(config1)
	WithEndState(sdkstate.State("B"))(config1)
	WithTasks()(config1)

	config2 := &PhaseConfig{}
	WithTasks()(config2)
	WithEndState(sdkstate.State("B"))(config2)
	WithStartState(sdkstate.State("A"))(config2)

	// Both should have same result
	if config1.startState != config2.startState {
		t.Error("options order affected startState")
	}
	if config1.endState != config2.endState {
		t.Error("options order affected endState")
	}
	if config1.supportsTasks != config2.supportsTasks {
		t.Error("options order affected supportsTasks")
	}
}
