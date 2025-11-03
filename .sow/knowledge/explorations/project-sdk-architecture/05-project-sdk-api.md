# Project SDK: Public API Design

## Summary

This document defines the public API for the Project SDK - the builder interface that project types use to define their complete configuration including phases, state machines, prompts, and validation rules.

## Objectives

1. **Single builder interface** - Project types define everything through one fluent API
2. **Complete encapsulation** - SDK handles state machine wiring, phase transitions, prompt generation
3. **Type-safe where possible** - Leverage Go's type system while maintaining flexibility
4. **Clear separation** - Phase configuration vs state machine vs prompts

## Critical Architectural Constraint

**Data-Only Child Types:**

Everything below `Project` in the hierarchy MUST be pure data with zero business logic:

- `Phase` - Pure data (status, metadata, artifacts, tasks)
- `Artifact` - Pure data (type, path, approved, metadata)
- `Task` - Pure data (status, inputs, outputs, metadata)

**All orchestration happens at Project level:**

- `project.Advance()` - Fires project state machine events
- `project.AdvanceTask(taskID)` - (future) Fires task state machine events
- `project.AdvanceArtifact(artifactID)` - (future) Fires artifact state machine events

**Why this constraint matters:**

1. **No circular dependencies** - Child types never reference parent
2. **Clean serialization** - Data-only types serialize to/from YAML trivially
3. **Centralized logic** - All state machines managed by Project
4. **Simple testing** - Child types are just structs with fields
5. **No forward references** - Architecture remains clean and maintainable

**Implications:**

- Phase/Task/Artifact are pure data structs with public fields
- Collections (PhaseCollection, ArtifactCollection, TaskCollection) provide structural operations only (Get, Add, Remove)
- State machines are NEVER stored on child types - only on Project
- Guards in task/artifact machines close over project state (no direct reference needed)
- CLI commands use collections for navigation, then direct field access for mutations

## Core Type Structure

```go
// Project - Top-level orchestrator with state machine
type Project struct {
    // Data fields (serialized to/from YAML)
    Name        string
    Type        string
    Description string
    Branch      string
    Phases      PhaseCollection     // Collection type, not map
    Statechart  StatechartState

    // Runtime fields (set during Load, not serialized)
    config  *ProjectTypeConfig     // Reference to type configuration
    machine *statechart.Machine    // State machine instance (built once)
}

// Phase - Pure data, no business logic
type Phase struct {
    Status      string
    Enabled     bool
    CreatedAt   time.Time
    StartedAt   *time.Time
    CompletedAt *time.Time
    Metadata    map[string]interface{}
    Inputs      ArtifactCollection  // Collection type, not slice
    Outputs     ArtifactCollection  // Collection type, not slice
    Tasks       TaskCollection      // Collection type, not slice
}

// Collections provide structural operations (Get, Add, Remove)
// They are type wrappers over basic types (map/slice) and serialize/deserialize transparently
type PhaseCollection map[string]*Phase

func (pc PhaseCollection) Get(name string) (*Phase, error) {
    phase, exists := pc[name]
    if !exists {
        return nil, fmt.Errorf("phase not found: %s", name)
    }
    return phase, nil
}

type ArtifactCollection []Artifact

func (ac ArtifactCollection) Get(index int) (*Artifact, error) {
    if index < 0 || index >= len(ac) {
        return nil, fmt.Errorf("index out of range: %d", index)
    }
    return &ac[index], nil
}

func (ac *ArtifactCollection) Add(artifact Artifact) {
    *ac = append(*ac, artifact)
}

type TaskCollection []Task

func (tc TaskCollection) Get(id string) (*Task, error) {
    for i := range tc {
        if tc[i].ID == id {
            return &tc[i], nil
        }
    }
    return nil, fmt.Errorf("task not found: %s", id)
}

// Note: Collections serialize/deserialize transparently
// - PhaseCollection → map in YAML
// - ArtifactCollection → array in YAML
// - TaskCollection → array in YAML
// Methods (Get, Add, etc.) are available after deserialization

// Artifact - Pure data
type Artifact struct {
    Type      string
    Path      string
    Approved  bool
    CreatedAt time.Time
    Metadata  map[string]interface{}
}

// Task - Pure data
type Task struct {
    ID          string
    Name        string
    Status      string
    Iteration   int
    Metadata    map[string]interface{}
    Inputs      []Artifact
    Outputs     []Artifact
}

// ProjectTypeConfig - Schema/rules for a project type
type ProjectTypeConfig struct {
    name           string
    requiredPhases []string
    phaseConfigs   map[string]*PhaseConfig
    transitions    []TransitionConfig
    prompts        map[State]PromptFunc

    // Only two public methods:
    // - Validate(project *Project) error
    // - BuildMachine(project *Project, initialState State) *statechart.Machine
}
```

## Proposed Builder API

**Used once at app startup to register project type:**

```go
func init() {
    project.Register("standard", NewStandardProjectConfig())
}

func NewStandardProjectConfig() *sdk.ProjectTypeConfig {
    return sdk.NewProjectTypeConfigBuilder("standard").

        // ===== PHASE CONFIGURATION =====

        WithPhase("planning",
            WithStartState(PlanningActive),
            WithEndState(PlanningActive),
            WithInputs("context"),               // Allowed input types
            WithOutputs("task_list"),            // Allowed output types
        ).

        WithPhase("implementation",
            WithStartState(ImplementationPlanning),
            WithEndState(ImplementationExecuting),
            WithInputs(),                        // Empty = allow all types
            WithOutputs(),
            WithTasks(),
            WithMetadata(map[string]sdk.FieldSpec{
                "tasks_approved": {Type: "bool"},
            }),
        ).

        WithPhase("review",
            WithStartState(ReviewActive),
            WithEndState(ReviewActive),
            WithInputs(),
            WithOutputs("review"),
        ).

        WithPhase("finalize",
            WithStartState(FinalizeDocumentation),
            WithEndState(FinalizeDelete),
            WithMetadata(map[string]sdk.FieldSpec{
                "project_deleted": {Type: "bool"},
            }),
        ).

        // ===== STATE MACHINE =====

        SetInitialState(NoProject).

        AddTransition(
            NoProject,
            PlanningActive,
            EventProjectInit,
        ).

        AddTransition(
            PlanningActive,
            ImplementationPlanning,
            EventCompletePlanning,
            sdk.WithGuard(func(p *Project) bool {
                return p.PhaseOutputApproved("planning", "task_list")
            }),
        ).

        AddTransition(
            ImplementationPlanning,
            ImplementationExecuting,
            EventTasksApproved,
            sdk.WithGuard(func(p *Project) bool {
                return p.PhaseMetadataBool("implementation", "tasks_approved")
            }),
        ).

        AddTransition(
            ImplementationExecuting,
            ReviewActive,
            EventAllTasksComplete,
            sdk.WithGuard(func(p *Project) bool {
                return p.AllTasksComplete()
            }),
        ).

        AddTransition(
            ReviewActive,
            FinalizeDocumentation,
            EventReviewPass,
            sdk.WithGuard(func(p *Project) bool {
                return p.ReviewPassed()
            }),
        ).

        AddTransition(
            ReviewActive,
            ImplementationPlanning,
            EventReviewFail,
            sdk.WithGuard(func(p *Project) bool {
                return p.ReviewFailed()
            }),
        ).

        // Finalize substates
        AddTransition(
            FinalizeDocumentation,
            FinalizeChecks,
            EventDocumentationDone,
        ).

        AddTransition(
            FinalizeChecks,
            FinalizeDelete,
            EventChecksDone,
        ).

        AddTransition(
            FinalizeDelete,
            NoProject,
            EventProjectDelete,
            sdk.WithGuard(func(p *Project) bool {
                return p.PhaseMetadataBool("finalize", "project_deleted")
            }),
        ).

        // ===== PROMPTS =====

        WithPrompt(PlanningActive, func(p *Project) string {
            return "Planning phase active. Create and approve task list."
        }).

        WithPrompt(ImplementationPlanning, func(p *Project) string {
            return "Review and approve implementation tasks."
        }).

        WithPrompt(ImplementationExecuting, func(p *Project) string {
            return "Execute implementation tasks."
        }).

        WithPrompt(ReviewActive, func(p *Project) string {
            return "Review implementation and provide assessment."
        }).

        WithPrompt(FinalizeDocumentation, func(p *Project) string {
            return "Update documentation."
        }).

        WithPrompt(FinalizeChecks, func(p *Project) string {
            return "Run tests and linters."
        }).

        WithPrompt(FinalizeDelete, func(p *Project) string {
            return "Clean up project directory."
        }).

        // ===== BUILD =====

        Build()
}
```

## API Quick Reference

```go
// Entry point
sdk.NewProjectTypeConfigBuilder(name string) *ProjectTypeConfigBuilder

// Phase configuration
.WithPhase(name string, opts ...PhaseOpt) *ProjectTypeConfigBuilder
    WithStartState(state State) PhaseOpt
    WithEndState(state State) PhaseOpt
    WithInputs(allowedTypes ...string) PhaseOpt      // Empty = allow all
    WithOutputs(allowedTypes ...string) PhaseOpt     // Empty = allow all
    WithTasks() PhaseOpt
    WithMetadata(fields map[string]FieldSpec) PhaseOpt

// State machine
.SetInitialState(state State) *ProjectTypeConfigBuilder
.AddTransition(from, to State, event Event, opts ...TransitionOption) *ProjectTypeConfigBuilder
    sdk.WithGuard(guardFunc func(*Project) bool) TransitionOption
    sdk.WithOnEntry(action func(*Project) error) TransitionOption
    sdk.WithOnExit(action func(*Project) error) TransitionOption

// Prompts
.WithPrompt(state State, generator func(*Project) string) *ProjectTypeConfigBuilder

// Build
.Build() *sdk.ProjectTypeConfig
```

---

## Field Specification

```go
type FieldSpec struct {
    Type     string  // "string", "int", "bool", "float"
    Validate string  // go-playground/validator tags
}
```

**Examples:**

```go
// Boolean field
FieldSpec{Type: "bool"}

// Integer with minimum
FieldSpec{Type: "int", Validate: "gte=1"}

// String enum
FieldSpec{Type: "string", Validate: "oneof=pass fail"}

// Required string
FieldSpec{Type: "string", Validate: "required"}
```

---

## State and Event Types

```go
type State string
type Event string
```

Project types define their own state and event constants:

```go
const (
    NoProject State = "NoProject"  // Shared across all project types

    // Standard project states
    PlanningActive            State = "PlanningActive"
    ImplementationPlanning    State = "ImplementationPlanning"
    ImplementationExecuting   State = "ImplementationExecuting"
    ReviewActive              State = "ReviewActive"
    FinalizeDocumentation     State = "FinalizeDocumentation"
    FinalizeChecks            State = "FinalizeChecks"
    FinalizeDelete            State = "FinalizeDelete"
)

const (
    // Standard project events
    EventProjectInit        Event = "EventProjectInit"
    EventCompletePlanning   Event = "EventCompletePlanning"
    EventTasksApproved      Event = "EventTasksApproved"
    EventAllTasksComplete   Event = "EventAllTasksComplete"
    EventReviewPass         Event = "EventReviewPass"
    EventReviewFail         Event = "EventReviewFail"
    EventDocumentationDone  Event = "EventDocumentationDone"
    EventChecksDone         Event = "EventChecksDone"
    EventProjectDelete      Event = "EventProjectDelete"
)
```

---

## Helper Methods on Project

Project provides helper methods for common guard patterns:

```go
func (p *Project) PhaseOutputApproved(phaseName, outputType string) bool

func (p *Project) PhaseMetadataBool(phaseName, key string) bool

func (p *Project) AllTasksComplete() bool

func (p *Project) ReviewPassed() bool

func (p *Project) ReviewFailed() bool
```

---

## Builder Internals

### ProjectTypeConfigBuilder Structure

```go
type ProjectTypeConfigBuilder struct {
    name           string
    initialState   State
    phases         map[string]*PhaseConfig
    transitions    []TransitionConfig
    prompts        map[State]PromptFunc
}

type PhaseConfig struct {
    name              string
    startState        State
    endState          State
    inputsEnabled     bool
    outputsEnabled    bool
    tasksEnabled      bool
    allowedInputTypes []string  // Empty = allow all
    allowedOutputTypes []string  // Empty = allow all
    metadata          map[string]FieldSpec
}

type TransitionConfig struct {
    from      State
    to        State
    event     Event
    guardFunc func(*Project) bool  // Template function, bound later
    onEntry   []func(*Project) error
    onExit    []func(*Project) error
}

type PromptFunc func(*Project) string
```

### Builder Method Implementations

**NewProjectTypeConfigBuilder:**
```go
func NewProjectTypeConfigBuilder(name string) *ProjectTypeConfigBuilder {
    return &ProjectTypeConfigBuilder{
        name:        name,
        phases:      make(map[string]*PhaseConfig),
        transitions: make([]TransitionConfig, 0),
        prompts:     make(map[State]PromptFunc),
    }
}
```

**WithPhase:**
```go
func (b *ProjectTypeConfigBuilder) WithPhase(name string, opts ...PhaseOpt) *ProjectTypeConfigBuilder {
    // Create config with defaults
    config := &PhaseConfig{
        name:              name,
        inputsEnabled:     false,
        outputsEnabled:    false,
        tasksEnabled:      false,
        allowedInputTypes: nil,  // nil = not enabled
        allowedOutputTypes: nil,
        metadata:          make(map[string]FieldSpec),
    }

    // Apply all options
    for _, opt := range opts {
        opt(config)
    }

    // Store in builder
    b.phases[name] = config

    return b
}
```

**SetInitialState:**
```go
func (b *ProjectTypeConfigBuilder) SetInitialState(state State) *ProjectTypeConfigBuilder {
    b.initialState = state
    return b
}
```

**AddTransition:**
```go
func (b *ProjectTypeConfigBuilder) AddTransition(
    from, to State,
    event Event,
    opts ...TransitionOption,
) *ProjectTypeConfigBuilder {
    // Create transition config
    config := &TransitionConfig{
        from:    from,
        to:      to,
        event:   event,
        onEntry: make([]func(*Project) error, 0),
        onExit:  make([]func(*Project) error, 0),
    }

    // Apply options (guards, actions)
    for _, opt := range opts {
        opt(config)
    }

    // Store transition template
    b.transitions = append(b.transitions, *config)

    return b
}

// TransitionOption implementations
type TransitionOption func(*TransitionConfig)

func WithGuard(guardFunc func(*Project) bool) TransitionOption {
    return func(tc *TransitionConfig) {
        tc.guardFunc = guardFunc
    }
}

func WithOnEntry(action func(*Project) error) TransitionOption {
    return func(tc *TransitionConfig) {
        tc.onEntry = append(tc.onEntry, action)
    }
}

func WithOnExit(action func(*Project) error) TransitionOption {
    return func(tc *TransitionConfig) {
        tc.onExit = append(tc.onExit, action)
    }
}
```

**WithPrompt:**
```go
func (b *ProjectTypeConfigBuilder) WithPrompt(
    state State,
    generator func(*Project) string,
) *ProjectTypeConfigBuilder {
    b.prompts[state] = generator
    return b
}
```

### Build() Implementation

**Build() assembles the final ProjectTypeConfig:**

```go
func (b *ProjectTypeConfigBuilder) Build() *ProjectTypeConfig {
    // Validate configuration
    if err := b.validate(); err != nil {
        panic(fmt.Sprintf("invalid project type config: %v", err))
    }

    return &ProjectTypeConfig{
        name:           b.name,
        initialState:   b.initialState,
        phases:         b.phases,
        transitions:    b.transitions,
        prompts:        b.prompts,
    }
}

func (b *ProjectTypeConfigBuilder) validate() error {
    // Ensure initial state is set
    if b.initialState == "" {
        return errors.New("initial state not set")
    }

    // Ensure at least one phase
    if len(b.phases) == 0 {
        return errors.New("no phases defined")
    }

    // Validate each phase has start/end states
    for name, phase := range b.phases {
        if phase.startState == "" {
            return fmt.Errorf("phase %s missing start state", name)
        }
        if phase.endState == "" {
            return fmt.Errorf("phase %s missing end state", name)
        }
    }

    // Validate transitions reference valid states
    usedStates := make(map[State]bool)
    for _, tc := range b.transitions {
        usedStates[tc.from] = true
        usedStates[tc.to] = true
    }

    // Ensure all phase states are used in transitions
    for _, phase := range b.phases {
        if !usedStates[phase.startState] && phase.startState != NoProject {
            return fmt.Errorf("phase %s start state %s not used in any transition",
                phase.name, phase.startState)
        }
    }

    return nil
}
```

### ProjectTypeConfig Public Methods

The built `ProjectTypeConfig` exposes only two public methods:

**1. Validate(project *Project) error**

```go
func (ptc *ProjectTypeConfig) Validate(project *Project) error {
    // Validate required phases exist
    for name := range ptc.phases {
        if _, exists := project.Phases[name]; !exists {
            return fmt.Errorf("missing required phase: %s", name)
        }
    }

    // Validate no unexpected phases
    for name := range project.Phases {
        if _, expected := ptc.phases[name]; !expected {
            return fmt.Errorf("unexpected phase: %s", name)
        }
    }

    // Validate each phase
    for name, phaseConfig := range ptc.phases {
        phase := project.Phases[name]

        // Validate inputs if not enabled
        if !phaseConfig.inputsEnabled && len(phase.Inputs) > 0 {
            return fmt.Errorf("phase %s has inputs but inputs not enabled", name)
        }

        // Validate outputs if not enabled
        if !phaseConfig.outputsEnabled && len(phase.Outputs) > 0 {
            return fmt.Errorf("phase %s has outputs but outputs not enabled", name)
        }

        // Validate tasks if not enabled
        if !phaseConfig.tasksEnabled && len(phase.Tasks) > 0 {
            return fmt.Errorf("phase %s has tasks but tasks not enabled", name)
        }

        // Validate artifact types
        if phaseConfig.inputsEnabled && len(phaseConfig.allowedInputTypes) > 0 {
            for _, artifact := range phase.Inputs {
                if !contains(phaseConfig.allowedInputTypes, artifact.Type) {
                    return fmt.Errorf("phase %s has invalid input type: %s", name, artifact.Type)
                }
            }
        }

        if phaseConfig.outputsEnabled && len(phaseConfig.allowedOutputTypes) > 0 {
            for _, artifact := range phase.Outputs {
                if !contains(phaseConfig.allowedOutputTypes, artifact.Type) {
                    return fmt.Errorf("phase %s has invalid output type: %s", name, artifact.Type)
                }
            }
        }

        // Validate metadata fields using go-playground/validator
        validator := validator.New()
        for fieldName, value := range phase.Metadata {
            spec, exists := phaseConfig.metadata[fieldName]
            if !exists {
                return fmt.Errorf("phase %s has unexpected metadata field: %s", name, fieldName)
            }

            // Type check
            if err := spec.ValidateType(value); err != nil {
                return fmt.Errorf("phase %s metadata %s: %w", name, fieldName, err)
            }

            // Constraint check
            if spec.Validate != "" {
                if err := validator.Var(value, spec.Validate); err != nil {
                    return fmt.Errorf("phase %s metadata %s validation failed: %w",
                        name, fieldName, err)
                }
            }
        }
    }

    return nil
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
```

**2. BuildMachine(project *Project, initialState State) *statechart.Machine**

This is where the magic happens - binding template guards/actions to the specific project instance:

```go
func (ptc *ProjectTypeConfig) BuildMachine(
    project *Project,
    initialState State,
) *statechart.Machine {
    // Create statechart builder with initial state
    builder := statechart.NewBuilder(initialState, nil, nil)

    // Add each configured transition
    for _, tc := range ptc.transitions {
        var opts []statechart.TransitionOption

        // Bind guard template to this project instance
        if tc.guardFunc != nil {
            // Create closure that captures project
            boundGuard := func() bool {
                return tc.guardFunc(project)  // project captured in closure
            }
            opts = append(opts, statechart.WithGuard(boundGuard))
        }

        // Bind exit actions
        for _, exitAction := range tc.onExit {
            // Capture action in closure
            action := exitAction
            boundExit := func(_ context.Context, _ ...any) error {
                return action(project)  // project captured in closure
            }
            opts = append(opts, statechart.WithOnExit(boundExit))
        }

        // Bind entry actions
        for _, entryAction := range tc.onEntry {
            // Capture action in closure
            action := entryAction
            boundEntry := func(_ context.Context, _ ...any) error {
                return action(project)  // project captured in closure
            }
            opts = append(opts, statechart.WithOnEntry(boundEntry))
        }

        // Add transition with bound guards/actions
        builder.AddTransition(tc.from, tc.to, tc.event, opts...)
    }

    // Add automatic phase status transitions
    // These are injected based on phase start/end states
    for _, phaseConfig := range ptc.phases {
        // On entering start state, set phase to in_progress
        builder.ConfigureState(phaseConfig.startState).
            OnEntry(func(_ context.Context, _ ...any) error {
                phase := project.Phases[phaseConfig.name]
                phase.Status = "in_progress"
                if phase.StartedAt == nil {
                    now := time.Now()
                    phase.StartedAt = &now
                }
                return nil
            })

        // On exiting end state, set phase to completed
        builder.ConfigureState(phaseConfig.endState).
            OnExit(func(_ context.Context, _ ...any) error {
                phase := project.Phases[phaseConfig.name]
                phase.Status = "completed"
                now := time.Now()
                phase.CompletedAt = &now
                return nil
            })
    }

    // Add prompt generation for each state
    for state, promptFunc := range ptc.prompts {
        builder.ConfigureState(state).
            OnEntry(func(_ context.Context, _ ...any) error {
                prompt := promptFunc(project)
                fmt.Println(prompt)
                return nil
            })
    }

    return builder.Build()
}
```

### Key Insights

**1. Templates vs Instances:**
- Builder stores **template functions**: `func(*Project) bool`
- BuildMachine creates **bound closures**: `func() bool` that capture specific project
- Same template used for all projects of this type

**2. Automatic Phase Wiring:**
- SDK injects OnEntry/OnExit for phase status updates
- Project types don't manually wire phase transitions
- Based on StartState/EndState configuration

**3. Closure Binding:**
- Guards defined once with project parameter: `func(p *Project) bool`
- BuildMachine wraps them: `func() bool { return guardFunc(project) }`
- `project` variable captured in closure at machine creation time
- Guards always see live state of that specific project instance

**4. Validation:**
- Build-time: Validates configuration structure (phases defined, states used)
- Runtime: Validates project instances match config (required phases, allowed types, metadata)

---

## Load() and Save() Implementation

### Load() - Deserialize and Initialize

```go
// Registry is populated at app startup
var Registry = map[string]*ProjectTypeConfig{
    "standard":    NewStandardProjectConfig(),
    "exploration": NewExplorationProjectConfig(),
}

func Load(ctx context.Context) (*Project, error) {
    // 1. Read YAML file
    path := filepath.Join(ctx.WorkingDir(), ".sow/project/state.yaml")
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, ErrNoProject
        }
        return nil, fmt.Errorf("failed to read state file: %w", err)
    }

    // 2. Deserialize into Project instance
    var project Project
    if err := yaml.Unmarshal(data, &project); err != nil {
        return nil, fmt.Errorf("failed to unmarshal state: %w", err)
    }

    // 3. Lookup and attach type config
    config, exists := Registry[project.Type]
    if !exists {
        return nil, fmt.Errorf("unknown project type: %s", project.Type)
    }
    project.config = config

    // 4. Build state machine once, initialized with current state
    initialState := State(project.Statechart.CurrentState)
    project.machine = config.BuildMachine(&project, initialState)

    // 5. Validate project instance against config
    if err := config.Validate(&project); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    return &project, nil
}
```

**Key points:**
- Deserializes universal `Project` struct from YAML
- Collections (PhaseCollection, ArtifactCollection, TaskCollection) deserialize transparently - they're just type wrappers over map/slice
- Type config looked up from registry by `project.Type` field
- State machine built fresh, initialized with saved state
- Validation ensures instance matches type rules
- Collection methods (Get, Add, etc.) available immediately after deserialization

### Save() - Serialize to Disk

```go
func (p *Project) Save() error {
    // 1. Sync statechart state with machine's current state
    if p.machine != nil {
        p.Statechart.CurrentState = p.machine.State().String()
        p.Statechart.UpdatedAt = time.Now()
    }

    // 2. Validate before saving
    if err := p.config.Validate(p); err != nil {
        return fmt.Errorf("validation failed before save: %w", err)
    }

    // 3. Marshal to YAML
    data, err := yaml.Marshal(p)
    if err != nil {
        return fmt.Errorf("failed to marshal state: %w", err)
    }

    // 4. Write to file atomically
    path := filepath.Join(p.workingDir, ".sow/project/state.yaml")
    tmpPath := path + ".tmp"

    if err := os.WriteFile(tmpPath, data, 0644); err != nil {
        return fmt.Errorf("failed to write temp file: %w", err)
    }

    if err := os.Rename(tmpPath, path); err != nil {
        return fmt.Errorf("failed to rename temp file: %w", err)
    }

    return nil
}
```

**Key points:**
- Syncs `Statechart.CurrentState` from machine's internal state
- Validates before saving (catches bugs early)
- Atomic write via temp file + rename
- Only data fields serialized (config and machine are runtime-only)
- Collections serialize as their underlying types (map/slice) - no special handling needed

---

## Complete CLI Integration Examples

### Example 1: Simple Data Mutation (output set)

**Command:**
```bash
sow output set --index 0 approved true
```

**Implementation:**
```go
func runOutputSet(cmd *cobra.Command, args []string) error {
    // 1. Parse arguments
    ctx := cmdutil.GetContext(cmd.Context())
    index := viper.GetInt("index")
    phaseName := viper.GetString("phase")  // Defaults to current phase
    field := args[0]      // "approved"
    value := parseBool(args[1])  // "true" → bool

    // 2. Load project
    project, err := loader.Load(ctx)
    if err != nil {
        return fmt.Errorf("failed to load project: %w", err)
    }

    // 3. Use collections for structural navigation
    phase, err := project.Phases.Get(phaseName)
    if err != nil {
        return err  // "phase not found: planning"
    }

    output, err := phase.Outputs.Get(index)
    if err != nil {
        return err  // "index out of range: 5"
    }

    // 4. Direct field mutation - simple, type-safe
    switch field {
    case "approved":
        output.Approved = value
    case "type":
        output.Type = value
    default:
        return fmt.Errorf("unknown field: %s (use 'metadata set' for custom fields)", field)
    }

    // 5. Save at end
    if err := project.Save(); err != nil {
        return fmt.Errorf("failed to save: %w", err)
    }

    cmd.Println("✓ Output updated")
    return nil
}
```

**For metadata, use dedicated command:**
```bash
sow output metadata set --index 0 assessment pass
```

```go
func runOutputMetadataSet(cmd *cobra.Command, args []string) error {
    ctx := cmdutil.GetContext(cmd.Context())
    index := viper.GetInt("index")
    phaseName := viper.GetString("phase")
    field := args[0]      // "assessment"
    value := parseValue(args[1])  // "pass" (string, int, bool, etc.)

    project, err := loader.Load(ctx)
    if err != nil {
        return fmt.Errorf("failed to load project: %w", err)
    }

    // Use collections for navigation
    phase, err := project.Phases.Get(phaseName)
    if err != nil {
        return err
    }

    output, err := phase.Outputs.Get(index)
    if err != nil {
        return err
    }

    // Direct metadata mutation
    if output.Metadata == nil {
        output.Metadata = make(map[string]interface{})
    }
    output.Metadata[field] = value

    if err := project.Save(); err != nil {
        return fmt.Errorf("failed to save: %w", err)
    }

    cmd.Println("✓ Output metadata updated")
    return nil
}
```

**Flow:**
1. Load project (deserialize + attach config + build machine + validate)
2. Use collections for structural navigation (`Phases.Get`, `Outputs.Get`)
3. Direct field mutation (simple, type-safe for common fields)
4. Save (sync state + validate + serialize)
5. No state machine involved

**Benefits:**
- Command code is thin - just coordination
- Collections handle bounds checking and lookups
- Direct field access - idiomatic Go, type-safe
- Clear separation: common fields vs metadata (separate commands)
- No "magic" routing - explicit commands
- Collections are reusable (Add, Get, Remove operations)

---

### Example 2: State Machine Orchestration (advance)

**Command:**
```bash
sow advance
```

**Implementation:**
```go
func runAdvance(cmd *cobra.Command, args []string) error {
    // 1. Get context
    ctx := cmdutil.GetContext(cmd.Context())

    // 2. Load project
    project, err := loader.Load(ctx)
    if err != nil {
        return fmt.Errorf("failed to load project: %w", err)
    }

    // 3. Call project.Advance() - uses state machine
    if err := project.Advance(); err != nil {
        return fmt.Errorf("failed to advance: %w", err)
    }

    cmd.Println("✓ Advanced to next state")
    return nil
}
```

**Project.Advance() implementation:**
```go
func (p *Project) Advance() error {
    // 1. Determine next event based on current state and conditions
    event, err := p.determineNextEvent()
    if err != nil {
        return err
    }

    // 2. Check if event can fire (guard validation)
    can, err := p.machine.CanFire(event)
    if err != nil {
        return fmt.Errorf("failed to check transition: %w", err)
    }
    if !can {
        return ErrCannotAdvance
    }

    // 3. Fire event - state machine executes:
    //    - Exit actions on current state
    //    - Automatic phase status update (if leaving phase end state)
    //    - Transition to new state
    //    - Automatic phase status update (if entering phase start state)
    //    - Entry actions on new state
    //    - Prompt generation
    if err := p.machine.Fire(event); err != nil {
        return fmt.Errorf("failed to fire event: %w", err)
    }

    // 4. Save updated state
    if err := p.Save(); err != nil {
        return fmt.Errorf("failed to save state after advance: %w", err)
    }

    return nil
}

func (p *Project) determineNextEvent() (Event, error) {
    // Get current state from machine
    currentState := p.machine.State()

    // Determine event based on current state
    // This is project-type-specific logic
    switch currentState {
    case PlanningActive:
        return EventCompletePlanning, nil
    case ImplementationPlanning:
        return EventTasksApproved, nil
    case ImplementationExecuting:
        return EventAllTasksComplete, nil
    case ReviewActive:
        // Check if review passed or failed
        if p.ReviewPassed() {
            return EventReviewPass, nil
        }
        return EventReviewFail, nil
    case FinalizeDocumentation:
        return EventDocumentationDone, nil
    case FinalizeChecks:
        return EventChecksDone, nil
    case FinalizeDelete:
        return EventProjectDelete, nil
    default:
        return "", fmt.Errorf("no event for state: %s", currentState)
    }
}
```

**Flow:**
1. Load project (machine already built and initialized)
2. Determine next event based on current state
3. Fire event on machine:
   - Guards check (using bound closures with project state)
   - Exit actions run (mutate project)
   - Phase status automatically updated
   - Entry actions run (mutate project)
   - Prompts displayed
4. Save (machine's internal state synced to project.Statechart.CurrentState)

---

## Complete End-to-End Example

**Scenario:** User advances from Planning to Implementation

**1. Initial state (planning/tasks.md approved):**
```yaml
name: "add-auth"
type: "standard"
phases:
  planning:
    status: "in_progress"
    outputs:
      - type: "task_list"
        path: "planning/tasks.md"
        approved: true
  implementation:
    status: "pending"
statechart:
  current_state: "PlanningActive"
```

**2. User runs:** `sow advance`

**3. Load() executes:**
- Reads YAML → `Project` instance
- Attaches `Registry["standard"]` config
- Builds machine initialized with `PlanningActive`
- Validates

**4. project.Advance() executes:**
- `determineNextEvent()` → returns `EventCompletePlanning`
- `machine.CanFire(EventCompletePlanning)` → checks guard:
  ```go
  func(p *Project) bool {
      return p.PhaseOutputApproved("planning", "task_list")
  }
  ```
  → Guard returns `true` (task_list is approved)
- `machine.Fire(EventCompletePlanning)`:
  - **Exit PlanningActive:**
    - Automatic: `planning.Status = "completed"`, `planning.CompletedAt = now`
  - **Transition:** `PlanningActive` → `ImplementationPlanning`
  - **Enter ImplementationPlanning:**
    - Automatic: `implementation.Status = "in_progress"`, `implementation.StartedAt = now`
    - Prompt: "Review and approve implementation tasks."

**5. project.Save() executes:**
- Syncs: `Statechart.CurrentState = "ImplementationPlanning"`
- Validates
- Serializes to YAML

**6. Final state:**
```yaml
name: "add-auth"
type: "standard"
phases:
  planning:
    status: "completed"
    completed_at: "2025-01-15T10:30:00Z"
    outputs:
      - type: "task_list"
        path: "planning/tasks.md"
        approved: true
  implementation:
    status: "in_progress"
    started_at: "2025-01-15T10:30:00Z"
statechart:
  current_state: "ImplementationPlanning"
  updated_at: "2025-01-15T10:30:00Z"
```

**7. User sees:**
```
Review and approve implementation tasks.
✓ Advanced to next state
```

---

## Summary

**The complete flow:**

1. **App startup:** Register project type configs
2. **CLI command:** Load project (deserialize + attach config + build machine)
3. **Most commands:** Mutate data directly, call `Save()`
4. **Advance command:** Call `project.Advance()` which uses state machine
5. **State machine:** Guards check conditions, actions mutate project, phases auto-update
6. **Save:** Sync machine state, validate, serialize

**Key architectural points:**
- Pure data types (Phase/Artifact/Task) - no logic
- Project orchestrates everything - owns state machine
- ProjectTypeConfig defines rules - registered once, used many times
- State machine built per project instance - closures capture project
- Automatic phase status management - no manual wiring needed
- Single save point - end of command execution

---

## Open Questions

1. Should phase metadata be defined inline or separately?
2. Do we need a way to define task metadata schemas?
3. How do we handle artifact metadata validation?
4. Should prompts be required for all states or optional?
5. How do we validate that StartState/EndState are used in transitions?
