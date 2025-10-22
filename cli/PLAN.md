# Composable Phases Architecture - MVP Plan

**Status:** In Progress
**Branch:** (current branch)
**Goal:** Refactor project architecture to support multiple project types without command explosion

---

## Problem Statement

The current `sow` architecture has project-specific commands embedded in the CLI:

```bash
sow agent project phase enable discovery
sow agent task add --name "..."
sow agent project review add-report
sow agent project finalize add-document
```

As we add new project types (design projects, spike projects, documentation projects), this approach creates **command explosion**:
- Each project type has custom phases (3-5 phases per type)
- Each phase has custom fields (2-5 fields per phase)
- Each field needs operations (add, update, etc.)

**Result:** 10-15 new commands per project type, scaling to 50-75 commands with 5 project types.

---

## Solution: Composable Phases Architecture

### Core Concepts

1. **CUE-First Schema Design**
   - Phases defined as CUE schemas in `schemas/phases/`
   - Projects compose phases in `schemas/projects/`
   - Go types generated via `cue exp gengotypes`
   - Single source of truth for structure

2. **Phase Library**
   - Reusable phase implementations in `internal/phases/<phase>/`
   - Each phase accepts its generated type from schemas
   - Phases own their prompt templates (colocated)
   - Phases are "lego blocks" that add states to state machine

3. **Meta-Level Chain Builder**
   - Helper function chains phases together
   - Wires up forward transitions automatically
   - Returns phase map for project-specific customizations

4. **Project Types**
   - Compose phases in desired order
   - Add exceptional transitions (e.g., backward loops)
   - Can override prompts by replacing entry actions
   - Own the complete state machine

5. **Implicit Active Phase**
   - Only one phase active at a time (enforced by state machine)
   - CLI commands operate on active phase (no `--phase` flag needed)
   - State machine validates operations based on current phase

### Bounded Command Set

With this architecture, commands **do not multiply** with project types:

```bash
# Phase transitions (4 commands)
sow agent enable <phase>      # Enable optional phase
sow agent skip <phase>        # Skip optional phase
sow agent complete            # Complete active phase
sow agent status              # Show current state

# Operations on active phase (4 commands)
sow agent task add <name>     # Add task (if phase supports)
sow agent artifact add <path> # Add artifact (if phase supports)
sow agent set <field> <value> # Set custom field
sow agent log "..."           # Log action

# Read operations (2 commands)
sow agent info [--phase]      # Show phase details
sow agent task list           # List tasks
```

**Total: ~10 commands for all project types, forever.**

---

## Architecture Details

### CUE Schema Organization

```
cli/schemas/
├── cue.mod/
│
├── common.cue                 # Shared types (#Artifact, #Task, etc.)
│
├── phases/                    # Reusable phase definitions
│   ├── discovery.cue
│   ├── design.cue
│   ├── implementation.cue
│   ├── review.cue
│   └── finalize.cue
│
├── projects/                  # Project type schemas
│   ├── standard.cue           # Composes 5 standard phases
│   └── design.cue             # Future: design project phases
│
└── cue_types_gen.go           # Generated from all .cue files
```

**Example Phase Schema:**

```cue
// schemas/phases/discovery.cue
package schemas

#DiscoveryPhase: {
    // Common phase fields
    status: "pending" | "in_progress" | "completed" | "skipped"
    created_at: string
    started_at: null | string
    completed_at: null | string

    // Discovery-specific fields
    enabled: bool
    discovery_type?: "bug" | "feature" | "docs" | "refactor" | "general"
    artifacts: [...#Artifact]
}
```

**Example Project Schema:**

```cue
// schemas/projects/standard.cue
package schemas

#StandardProjectState: {
    project: {
        type: "standard"
        name: string
        branch: string
        description: string
        github_issue?: int
        created_at: string
        updated_at: string
    }

    statechart: {
        current_state: string
    }

    phases: {
        discovery: #DiscoveryPhase
        design: #DesignPhase
        implementation: #ImplementationPhase
        review: #ReviewPhase
        finalize: #FinalizePhase
    }
}

// Root discriminated union (future)
#ProjectState: #StandardProjectState | #DesignProjectState | ...
```

**Generated Go Types:**

```go
// schemas/cue_types_gen.go (auto-generated)
package schemas

type DiscoveryPhase struct {
    Status         string
    Created_at     time.Time
    Started_at     *string
    Completed_at   *string
    Enabled        bool
    Discovery_type *string
    Artifacts      []Artifact
}

type StandardProjectState struct {
    Project struct {
        Type         string
        Name         string
        Branch       string
        Description  string
        Github_issue *int
        Created_at   time.Time
        Updated_at   time.Time
    }

    Statechart struct {
        Current_state string
    }

    Phases struct {
        Discovery      DiscoveryPhase
        Design         DesignPhase
        Implementation ImplementationPhase
        Review         ReviewPhase
        Finalize       FinalizePhase
    }
}
```

---

### Phase Library Structure

```
internal/phases/
├── discovery/
│   ├── discovery.go           # Phase implementation
│   ├── discovery_test.go
│   └── templates/             # Embedded templates
│       ├── decision.md
│       └── active.md
│
├── design/
│   ├── design.go
│   ├── design_test.go
│   └── templates/
│       ├── decision.md
│       └── active.md
│
├── implementation/
│   ├── implementation.go
│   ├── implementation_test.go
│   └── templates/
│       ├── planning.md
│       └── executing.md
│
├── review/
│   ├── review.go
│   ├── review_test.go
│   └── templates/
│       └── active.md
│
├── finalize/
│   ├── finalize.go
│   ├── finalize_test.go
│   └── templates/
│       ├── documentation.md
│       ├── checks.md
│       └── delete.md
│
├── phase.go                   # Phase interface
├── metadata.go                # PhaseMetadata types
└── builder.go                 # BuildPhaseChain meta-helper
```

**Phase Interface:**

```go
// internal/phases/phase.go
package phases

type Phase interface {
    // Add this phase's states and transitions to the state machine
    // Links to the next phase's entry state
    AddToMachine(sm *StateMachine, nextPhaseEntry State)

    // What state does this phase start at?
    EntryState() State

    // Metadata for validation
    Metadata() PhaseMetadata
}

type PhaseMetadata struct {
    Name              string
    States            []State
    SupportsTasks     bool
    SupportsArtifacts bool
    CustomFields      []FieldDef
}

type FieldDef struct {
    Name string
    Type FieldType  // String, Bool, Int, Array
}
```

**Meta-Level Chain Builder:**

```go
// internal/phases/builder.go
package phases

// BuildPhaseChain wires up a sequence of phases into a state machine.
// Returns a map of phases by name for project-specific customization.
func BuildPhaseChain(sm *StateMachine, phases []Phase) PhaseMap {
    phaseMap := make(PhaseMap)

    // Wire initial transition
    sm.Configure(NoProject).
        Permit(EventProjectInit, phases[0].EntryState())

    // Chain phases together
    for i, phase := range phases {
        phaseMap[phase.Metadata().Name] = phase

        var nextEntry State
        if i < len(phases)-1 {
            nextEntry = phases[i+1].EntryState()
        } else {
            nextEntry = NoProject  // Last phase loops back
        }

        phase.AddToMachine(sm, nextEntry)
    }

    return phaseMap
}

type PhaseMap map[string]Phase
```

---

### Example Phase Implementation

```go
// internal/phases/discovery/discovery.go
package discovery

import (
    "embed"
    "github.com/jmgilman/sow/cli/schemas"
    "github.com/jmgilman/sow/cli/internal/phases"
)

//go:embed templates/*.md
var templates embed.FS

type DiscoveryPhase struct {
    optional bool
    data     *schemas.DiscoveryPhase  // Generated type!
}

func New(optional bool, data *schemas.DiscoveryPhase) *DiscoveryPhase {
    return &DiscoveryPhase{
        optional: optional,
        data:     data,
    }
}

func (p *DiscoveryPhase) EntryState() phases.State {
    return phases.DiscoveryDecision
}

func (p *DiscoveryPhase) AddToMachine(sm *stateless.StateMachine, nextEntry phases.State) {
    if p.optional {
        sm.Configure(phases.DiscoveryDecision).
            Permit(phases.EventEnableDiscovery, phases.DiscoveryActive).
            Permit(phases.EventSkipDiscovery, nextEntry).
            OnEntry(p.onDecisionEntry)
    } else {
        sm.Configure(phases.DiscoveryDecision).
            Permit(phases.EventEnableDiscovery, phases.DiscoveryActive).
            OnEntry(p.onDecisionEntry)
    }

    sm.Configure(phases.DiscoveryActive).
        Permit(phases.EventCompleteDiscovery, nextEntry, artifactsApprovedGuard).
        OnEntry(p.onActiveEntry)
}

func (p *DiscoveryPhase) Metadata() phases.PhaseMetadata {
    return phases.PhaseMetadata{
        Name:              "discovery",
        States:            []phases.State{phases.DiscoveryDecision, phases.DiscoveryActive},
        SupportsTasks:     false,
        SupportsArtifacts: true,
        CustomFields: []phases.FieldDef{
            {Name: "discovery_type", Type: phases.StringField},
        },
    }
}

// Entry actions render prompts
func (p *DiscoveryPhase) onDecisionEntry(ctx context.Context, args ...any) error {
    prompt := p.renderPrompt("decision")
    fmt.Println(prompt)
    return nil
}

func (p *DiscoveryPhase) onActiveEntry(ctx context.Context, args ...any) error {
    prompt := p.renderPrompt("active")
    fmt.Println(prompt)
    return nil
}

// Render prompt using embedded templates and phase data
func (p *DiscoveryPhase) renderPrompt(name string) string {
    content, _ := templates.ReadFile("templates/" + name + ".md")

    tmpl, _ := template.New(name).Parse(string(content))

    var buf bytes.Buffer
    tmpl.Execute(&buf, p.data)  // Just the DiscoveryPhase data

    return buf.String()
}

// Guards
func artifactsApprovedGuard(ctx context.Context, args ...any) bool {
    // Check all artifacts approved
    for _, artifact := range p.data.Artifacts {
        if !artifact.Approved {
            return false
        }
    }
    return true
}
```

**Template Example:**

```markdown
<!-- internal/phases/discovery/templates/decision.md -->

# Discovery Decision

You are deciding whether to enable the Discovery phase for this project.

**Project:** {{.Name}}
**Description:** {{.Description}}

## Discovery Worthiness Rubric

Score each criterion 0-2:
- Context Availability
- Problem Clarity
- Codebase Familiarity
- Research Needs

**Score 0-2:** Skip discovery
**Score 3-5:** Optional (ask human)
**Score 6-8:** Recommended

## Commands

Enable discovery:
  sow agent enable discovery

Skip discovery:
  sow agent skip discovery
```

---

### Project Type Implementation

```
internal/project/types/
├── types.go                   # ProjectType interface
└── standard/
    ├── standard.go            # Standard project implementation
    ├── standard_test.go
    └── templates/             # Optional prompt overrides
        └── implementation_planning.md
```

**ProjectType Interface:**

```go
// internal/project/types/types.go
package types

type ProjectType interface {
    BuildStateMachine() *statechart.Machine
    Phases() map[string]phases.PhaseMetadata
    Type() string
}

// DetectProjectType returns the appropriate project type based on state
func DetectProjectType(state *schemas.ProjectState) ProjectType {
    // For MVP, only StandardProject exists
    // Future: switch on state.Project.Type
    return standard.New(state)
}
```

**Standard Project:**

```go
// internal/project/types/standard/standard.go
package standard

import (
    "github.com/jmgilman/sow/cli/schemas"
    "github.com/jmgilman/sow/cli/internal/phases"
    "github.com/jmgilman/sow/cli/internal/phases/discovery"
    "github.com/jmgilman/sow/cli/internal/phases/design"
    "github.com/jmgilman/sow/cli/internal/phases/implementation"
    "github.com/jmgilman/sow/cli/internal/phases/review"
    "github.com/jmgilman/sow/cli/internal/phases/finalize"
)

type StandardProject struct {
    state *schemas.StandardProjectState
}

func New(state *schemas.StandardProjectState) *StandardProject {
    return &StandardProject{state: state}
}

func (p *StandardProject) Type() string {
    return "standard"
}

func (p *StandardProject) BuildStateMachine() *statechart.Machine {
    sm := stateless.NewStateMachine(phases.NoProject)

    // Compose phases with their data from state
    phaseList := []phases.Phase{
        discovery.New(true, &p.state.Phases.Discovery),
        design.New(true, &p.state.Phases.Design),
        implementation.New(&p.state.Phases.Implementation),
        review.New(&p.state.Phases.Review),
        finalize.New(&p.state.Phases.Finalize),
    }

    // Build forward chain
    phaseMap := phases.BuildPhaseChain(sm, phaseList)

    // Add exceptional backward transition
    // Review can fail and loop back to implementation
    implPhase := phaseMap["implementation"]
    sm.Configure(phases.ReviewActive).
        Permit(phases.EventReviewFail, implPhase.EntryState(), reviewFailedGuard)

    return statechart.NewMachineFromSM(sm, p.state)
}

func (p *StandardProject) Phases() map[string]phases.PhaseMetadata {
    return map[string]phases.PhaseMetadata{
        "discovery":      discovery.New(true, nil).Metadata(),
        "design":         design.New(true, nil).Metadata(),
        "implementation": implementation.New(nil).Metadata(),
        "review":         review.New(nil).Metadata(),
        "finalize":       finalize.New(nil).Metadata(),
    }
}

func reviewFailedGuard(ctx context.Context, args ...any) bool {
    // Check latest review report has assessment="fail"
    // (simplified - actual implementation would check reports)
    return true
}
```

**Prompt Override (Optional):**

If a project wants custom prompts, it overrides entry actions:

```go
func (p *StandardProject) BuildStateMachine() *statechart.Machine {
    sm := stateless.NewStateMachine(phases.NoProject)

    phaseList := []phases.Phase{ /* ... */ }
    phases.BuildPhaseChain(sm, phaseList)

    // Override implementation planning prompt
    sm.Configure(phases.ImplementationPlanning).
        OnEntry(p.customImplementationPrompt)

    return statechart.NewMachineFromSM(sm, p.state)
}

func (p *StandardProject) customImplementationPrompt(ctx context.Context, args ...any) error {
    // Render custom template or use raw string
    prompt := p.renderCustomPrompt()
    fmt.Println(prompt)
    return nil
}
```

---

## Prompting Architecture

### Design Principles

1. **Prompts live with phases** - Templates colocated in phase packages
2. **Phases receive only their data** - No cross-phase dependencies
3. **Templates use generated types** - Type-safe, matches CUE schema
4. **Projects can override** - Replace entry actions entirely
5. **Simple rendering** - Go templates, no complex framework

### Template Organization

```
internal/phases/discovery/templates/
├── decision.md       # DiscoveryDecision state prompt
└── active.md         # DiscoveryActive state prompt

internal/project/types/standard/templates/
└── implementation_planning.md   # Optional override
```

### How It Works

1. **Phase defines templates** - Embedded via `//go:embed`
2. **Phase receives its data** - Constructor gets `*schemas.DiscoveryPhase`
3. **Entry action renders** - `tmpl.Execute(&buf, p.data)`
4. **Templates access phase fields** - `{{.Status}}`, `{{.Enabled}}`, `{{.Artifacts}}`
5. **Projects can override** - Replace entry action with custom function

### Migration Strategy

**Existing templates:**
```
internal/prompts/templates/statechart/
├── discovery_decision.md
├── discovery_active.md
└── ...
```

**Move to phases:**
```bash
# Discovery
mkdir -p internal/phases/discovery/templates
mv internal/prompts/templates/statechart/discovery_decision.md \
   internal/phases/discovery/templates/decision.md
mv internal/prompts/templates/statechart/discovery_active.md \
   internal/phases/discovery/templates/active.md

# Design
mkdir -p internal/phases/design/templates
mv internal/prompts/templates/statechart/design_decision.md \
   internal/phases/design/templates/decision.md
mv internal/prompts/templates/statechart/design_active.md \
   internal/phases/design/templates/active.md

# ... etc for all phases
```

**Template content works as-is** - just needs field names to match generated types.

---

## MVP Scope

**Goal:** Reproduce current standard project behavior using composable phases architecture.

### What We're Building

1. **Schema Reorganization** (`schemas/`)
   - Move phase schemas to `schemas/phases/`
   - Create `schemas/projects/standard.cue`
   - Regenerate Go types with `cue exp gengotypes`

2. **Phase Library Package** (`internal/phases/`)
   - Base types and interfaces (`phase.go`, `metadata.go`, `builder.go`)
   - 5 standard project phase packages:
     - `discovery/` - Simple decision pattern
     - `design/` - Simple decision pattern
     - `implementation/` - Dual-state pattern (Planning → Executing)
     - `review/` - Simple active pattern with backward capability
     - `finalize/` - Multi-stage pattern (Documentation → Checks → Delete)
   - Each phase includes templates embedded

3. **Project Types Package** (`internal/project/types/`)
   - ProjectType interface
   - Standard project implementation
   - Type detection helper

4. **Updated Project Package** (`internal/project/`)
   - Create/Load uses project types
   - Routes to appropriate BuildStateMachine()
   - No breaking changes to existing API

5. **Prompt Migration**
   - Move templates from `internal/prompts/templates/statechart/` to phases
   - Update template field names to match generated types
   - Remove old centralized prompt infrastructure (or deprecate)

### What We're NOT Building (Yet)

- Design project type (future milestone)
- Command refactoring to implicit phase (future milestone)
- New CLI commands (use existing to validate)
- Schema migration for old state files (assume clean start for MVP)

### Success Criteria

✅ All existing unit tests pass (or updated appropriately)
✅ State machine produces same transitions as current implementation
✅ Can create, continue, and complete a standard project via CLI
✅ Phase metadata correctly identifies supported operations
✅ Prompts render correctly with phase data
✅ No breaking changes to state.yaml format (beyond type field)
✅ Code is well-documented with clear interfaces

---

## Implementation Plan

### Phase 1: Schema Reorganization

**Goal:** Restructure CUE schemas for composability, regenerate Go types.

**Tasks:**

1. **Create new schema structure**
   ```bash
   mkdir schemas/phases
   mkdir schemas/projects
   ```

2. **Extract phase schemas**
   - Create `schemas/phases/discovery.cue` (extract from current)
   - Create `schemas/phases/design.cue`
   - Create `schemas/phases/implementation.cue`
   - Create `schemas/phases/review.cue`
   - Create `schemas/phases/finalize.cue`
   - Each defines `#<Phase>Phase` type with all fields

3. **Create standard project schema**
   - Create `schemas/projects/standard.cue`
   - Compose: `phases: { discovery: #DiscoveryPhase, ... }`
   - Add `project.type: "standard"` field
   - Define `#StandardProjectState` root type

4. **Update common types**
   - Keep `schemas/common.cue` with `#Artifact`, `#Task`, etc.
   - Ensure all phases can reference common types

5. **Regenerate Go types**
   ```bash
   cd schemas
   cue exp gengotypes ./... > cue_types_gen.go
   ```

6. **Verify generated types**
   - Check `StandardProjectState` struct exists
   - Check phase types (`DiscoveryPhase`, `ImplementationPhase`, etc.)
   - Verify field types match expectations

**Validation:**
- CUE validation passes: `cue vet`
- Go builds: `go build ./schemas`
- Generated types have correct structure
- No breaking changes to existing state files (add type field as optional/default)

---

### Phase 2: Phase Library Foundation

**Goal:** Create phase package infrastructure and implement first phase.

**Tasks:**

1. **Create package structure**
   ```bash
   mkdir -p internal/phases/discovery/templates
   mkdir -p internal/phases/design/templates
   mkdir -p internal/phases/implementation/templates
   mkdir -p internal/phases/review/templates
   mkdir -p internal/phases/finalize/templates
   ```

2. **Define core types**
   - `internal/phases/phase.go` - Phase interface
   - `internal/phases/metadata.go` - PhaseMetadata, FieldDef types
   - `internal/phases/states.go` - State constants (DiscoveryDecision, etc.)
   - `internal/phases/events.go` - Event constants

3. **Implement BuildPhaseChain**
   - `internal/phases/builder.go` - Meta-level helper
   - Input: []Phase
   - Output: PhaseMap
   - Wires NoProject → phases[0]
   - Chains phases together
   - Unit test with mock phases

4. **Implement Discovery phase**
   - `internal/phases/discovery/discovery.go`
   - Constructor: `New(optional bool, data *schemas.DiscoveryPhase)`
   - Implement Phase interface
   - AddToMachine with decision/active states
   - Entry actions render prompts
   - Guards for artifact approval

5. **Migrate Discovery templates**
   ```bash
   cp internal/prompts/templates/statechart/discovery_decision.md \
      internal/phases/discovery/templates/decision.md
   cp internal/prompts/templates/statechart/discovery_active.md \
      internal/phases/discovery/templates/active.md
   ```
   - Update template field references to match schemas.DiscoveryPhase
   - Embed templates with `//go:embed templates/*.md`

6. **Test Discovery phase**
   - Unit test: AddToMachine configures states correctly
   - Unit test: Guards work (artifacts approved)
   - Unit test: Prompts render with test data
   - Integration test: Fire events, verify state changes

**Validation:**
- Discovery phase unit tests pass
- Templates render with mock data
- State machine transitions work
- BuildPhaseChain works with Discovery phase

---

### Phase 3: Complete Phase Library

**Goal:** Implement remaining 4 phases.

**Tasks:**

1. **Implement Design phase**
   - Similar to Discovery (simple decision pattern)
   - Copy templates, update field references
   - Unit tests

2. **Implement Implementation phase**
   - Dual-state pattern (Planning → Executing)
   - AddToMachine with both substates
   - Guards: hasTasksGuard, allTasksCompleteGuard
   - Templates: planning.md, executing.md
   - Unit tests

3. **Implement Review phase**
   - Simple active pattern
   - Note: backward transition added by project, not phase
   - Guards: latestReviewApproved
   - Template: active.md
   - Unit tests

4. **Implement Finalize phase**
   - Multi-stage pattern (Documentation → Checks → Delete)
   - Three states, three templates
   - Guards: documentationAssessed, checksAssessed, projectDeleted
   - Templates: documentation.md, checks.md, delete.md
   - Unit tests

5. **Update all templates**
   - Ensure field names match generated types
   - Test rendering with mock phase data
   - Verify prompts make sense in context

**Validation:**
- All 5 phases implemented
- All unit tests pass
- All templates render correctly
- Can chain all 5 phases together in test

---

### Phase 4: Project Types Package

**Goal:** Create project type abstraction and StandardProject implementation.

**Tasks:**

1. **Create types package**
   ```bash
   mkdir -p internal/project/types/standard
   ```

2. **Define ProjectType interface**
   - `internal/project/types/types.go`
   - Interface: BuildStateMachine(), Phases(), Type()
   - DetectProjectType helper function

3. **Implement StandardProject**
   - `internal/project/types/standard/standard.go`
   - Constructor: `New(state *schemas.StandardProjectState)`
   - BuildStateMachine:
     - Create phases with state data
     - BuildPhaseChain
     - Add backward transition (Review → Implementation)
   - Phases: return metadata from all 5 phases
   - Type: return "standard"

4. **Test StandardProject**
   - Unit test: BuildStateMachine creates all states
   - Unit test: Forward transitions exist
   - Unit test: Backward transition exists (Review → Implementation)
   - Unit test: Guards are attached
   - Integration test: Full lifecycle walk-through

**Validation:**
- StandardProject.BuildStateMachine() produces working state machine
- All states configured
- All transitions work
- Prompts render at each state
- Guards prevent invalid transitions

---

### Phase 5: Integration with Project Package

**Goal:** Wire new architecture into existing project package.

**Tasks:**

1. **Update project.Create()**
   - Add `project.type = "standard"` to state initialization
   - Use `types.DetectProjectType(state)` to get project type
   - Call `projectType.BuildStateMachine()` instead of current approach
   - Ensure backward compatibility

2. **Update project.Load()**
   - Read state from disk (includes project.type field)
   - Use `types.DetectProjectType(state)`
   - Build machine via project type
   - Return project with new machine

3. **Update Machine wrapper** (if needed)
   - Currently in `internal/project/statechart/machine.go`
   - May need helper: `DetermineActivePhase(state State) string`
   - Ensure compatibility with phase-based machines

4. **Add state migration** (optional)
   - For existing state files without `type` field
   - Default to "standard" if missing
   - Can be simple check in Load()

5. **Update project package tests**
   - Fix any tests broken by changes
   - Add tests for project type detection
   - Add tests for state migration (if implemented)

**Validation:**
- project.Create() works, produces standard project
- project.Load() loads existing projects correctly
- State machine behaves identically to old implementation
- All project package tests pass

---

### Phase 6: End-to-End Validation

**Goal:** Verify full workflow via existing CLI commands.

**Tasks:**

1. **Manual E2E test script**
   ```bash
   #!/bin/bash
   set -e

   # Setup
   cd /tmp
   mkdir test-sow && cd test-sow
   git init
   sow init
   git checkout -b test/composable-phases

   # Create project
   sow new "Test composable phases"

   # Verify state
   cat .sow/project/state.yaml | grep "type: standard"

   # Skip optional phases
   sow agent project phase skip discovery
   sow agent project phase skip design

   # Implementation
   sow agent task add "Task 1"
   sow agent task add "Task 2"
   sow agent project phase set implementation tasks_approved true

   # Complete tasks
   sow agent task update --id 010 --field status --value completed
   sow agent task update --id 020 --field status --value completed

   # Review pass
   sow agent project review add-report phases/review/report.md --assessment pass

   # Finalize
   sow agent project finalize add-document docs/README.md
   sow agent project delete

   echo "✅ E2E test passed"
   ```

2. **Run existing test suite**
   ```bash
   cd cli
   go test ./...
   ```

3. **Test edge cases**
   - Skip both discovery and design
   - Review fails, loops back to implementation
   - Multiple review iterations
   - Finalize with no documentation updates

4. **Verify prompts**
   - Check each state displays correct prompt
   - Verify prompt references correct commands
   - Ensure context data renders properly

**Validation:**
- E2E script completes successfully
- All unit/integration tests pass
- Edge cases work correctly
- Prompts display and are helpful
- No regressions from old behavior

---

### Phase 7: Documentation and Cleanup

**Goal:** Document new architecture, clean up old code.

**Tasks:**

1. **Add package documentation**
   - `internal/phases/README.md` - Architecture, how to add phases
   - `internal/project/types/README.md` - How to add project types
   - Godoc for all exported types

2. **Update main docs**
   - `docs/ARCHITECTURE.md` - Add composable phases section
   - `docs/PROJECT_LIFECYCLE.md` - Add project types concept
   - Add diagrams showing phase composition

3. **Add code examples**
   - Example: Creating a custom phase
   - Example: Creating a new project type
   - Example: Overriding phase prompts

4. **Clean up deprecated code**
   - Remove old state machine code (if replaced)
   - Remove old prompt infrastructure (if not used)
   - Update imports across codebase

5. **Run linters and formatters**
   ```bash
   just lint
   just fmt
   ```

**Validation:**
- All documentation is accurate
- Code examples compile and run
- No deprecated code remains
- Linters pass
- CI passes (tests, linting, build)

---

## Testing Strategy

### Unit Tests

**Phase Package:**
- `phase_test.go` - Mock phases, interface contracts
- `builder_test.go` - BuildPhaseChain with various inputs
- `discovery/discovery_test.go` - Discovery phase in isolation
- `design/design_test.go` - Design phase in isolation
- `implementation/implementation_test.go` - Implementation dual-state
- `review/review_test.go` - Review with backward capability
- `finalize/finalize_test.go` - Finalize multi-stage

**Project Types:**
- `standard/standard_test.go` - StandardProject state machine
- Verify all states configured
- Verify all transitions
- Verify guards attached
- Verify metadata accurate

**Integration:**
- `project/project_test.go` - Create/Load with new architecture
- Full lifecycle test
- Backward transition test
- Phase skip tests

### Regression Testing

All existing tests must pass:
```bash
cd cli
go test ./...
```

### Manual E2E Testing

See Phase 6 implementation plan for E2E test script.

---

## Migration Path for Future Project Types

### Adding Design Project Type

1. **Create phase schemas**
   - `schemas/phases/research.cue`
   - `schemas/phases/design_exploration.cue`
   - `schemas/phases/decomposition.cue`
   - `schemas/phases/cleanup.cue`

2. **Create project schema**
   - `schemas/projects/design.cue`
   - Compose: `phases: { research: #ResearchPhase, ... }`
   - Add to `#ProjectState` discriminated union

3. **Regenerate types**
   ```bash
   cue exp gengotypes ./... > cue_types_gen.go
   ```

4. **Implement phases**
   - `internal/phases/research/`
   - `internal/phases/design_exploration/`
   - `internal/phases/decomposition/`
   - `internal/phases/cleanup/`

5. **Implement project type**
   ```go
   // internal/project/types/design/design.go
   type DesignProject struct {
       state *schemas.DesignProjectState
   }

   func (p *DesignProject) BuildStateMachine() *statechart.Machine {
       phaseList := []phases.Phase{
           research.New(true, &p.state.Phases.Research),
           design_exploration.New(true, &p.state.Phases.DesignExploration),
           decomposition.New(&p.state.Phases.Decomposition),
           cleanup.New(&p.state.Phases.Cleanup),
       }

       phases.BuildPhaseChain(sm, phaseList)
       return statechart.NewMachineFromSM(sm, p.state)
   }
   ```

6. **Update type detection**
   ```go
   func DetectProjectType(state *schemas.ProjectState) ProjectType {
       switch state.Project.Type {
       case "standard":
           return standard.New(state.(*schemas.StandardProjectState))
       case "design":
           return design.New(state.(*schemas.DesignProjectState))
       default:
           return standard.New(state.(*schemas.StandardProjectState))
       }
   }
   ```

7. **No CLI changes needed!** Same commands work for all types.

**Estimated effort:** 1-2 days per project type after MVP complete.

---

## Risk Mitigation

**Risk: Breaking Existing Projects**
- Mitigation: Add `project.type` as optional field (defaults to "standard")
- Mitigation: All current tests must pass
- Mitigation: Manual E2E validation before merge

**Risk: State Machine Complexity**
- Mitigation: Phases isolated, easy to reason about
- Mitigation: Unit tests for each phase independently
- Mitigation: Clear documentation and diagrams

**Risk: Template Migration Issues**
- Mitigation: Move templates incrementally
- Mitigation: Test rendering at each step
- Mitigation: Keep old templates until verified

**Risk: Performance Regression**
- Mitigation: State machine construction is one-time (during load)
- Mitigation: Benchmark if concerned
- Mitigation: No runtime performance impact expected

---

## Definition of Done

- [ ] Schema reorganization complete (phases/, projects/)
- [ ] Go types regenerated from CUE
- [ ] All 5 standard phases implemented
- [ ] Meta-level BuildPhaseChain working
- [ ] StandardProject type implemented
- [ ] Project package integrated (Create/Load use types)
- [ ] All templates migrated to phase packages
- [ ] All existing unit tests pass (or updated)
- [ ] New unit tests for phases, project types (>80% coverage)
- [ ] Manual E2E test script passes
- [ ] Documentation updated (README, architecture docs)
- [ ] No breaking changes to state files (beyond type field)
- [ ] Code reviewed and approved
- [ ] Ready to merge

---

## Next Steps After MVP

1. **Command Refactoring** - Implement implicit active phase (remove --phase flags)
2. **Design Project Type** - Add exploratory project workflow
3. **CLI Simplification** - Consolidate using bounded command set
4. **Prompt Refinement** - Optimize for new architecture
5. **Advanced Features** - Phase-specific validation, hooks, etc.

---

## References

- Current state machine: `internal/project/statechart/machine.go`
- Current project package: `internal/project/project.go`
- Current schemas: `schemas/*.cue`
- Current prompts: `internal/prompts/templates/statechart/`
- Generated types: `schemas/cue_types_gen.go`

---

**Document Version:** 2.0
**Last Updated:** 2025-01-XX
**Authors:** Design discussion between human and Claude
**Changes from v1.0:** Added CUE-first architecture, simplified prompting, corrected phase/project structure
