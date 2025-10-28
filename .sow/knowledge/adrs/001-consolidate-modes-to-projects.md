# ADR-001: Consolidate Operating Modes into Project Types

**Status**: Proposed
**Date**: 2025-10-28
**Deciders**: Josh Gilman, Architecture Team
**Technical Story**: [modes-to-projects exploration](../.sow/knowledge/explorations/modes-to-projects-2025-01.md)

## Context

The sow framework currently implements three distinct operating modes—exploration, design, and breakdown—as separate systems with their own CLI commands, directory structures, and index schemas. While this separation provided initial flexibility, it creates several problems:

**Conceptual fragmentation**: Users must learn three separate systems with different mental models. Exploration uses topics/files/journal, design uses inputs/outputs, breakdown uses work units. Each has unique workflows and conventions.

**Code duplication**: Mode implementations share common patterns (state tracking, artifact management, logging) but duplicate these across separate codebases. Changes require updating multiple systems.

**Workflow inflexibility**: Modes operate independently. A typical workflow (explore → design → breakdown → implement) requires manual context transfer between systems. No programmatic state transitions exist.

**Limited context engineering**: Each mode loads a single large prompt at initialization. Long sessions cause prompt "forgetting" as conversation context grows. State changes don't trigger focused re-prompting.

Comprehensive exploration research (January 2025) revealed that all three modes map cleanly to the existing project phase system with six targeted schema improvements. No semantic mismatches exist. All features preserve naturally.

The decision: consolidate modes into specialized project types while maintaining their characteristic flexibility.

## Decision

Implement exploration, design, and breakdown as distinct project types within the unified project system, replacing the standalone mode implementations.

**Key architectural changes**:

1. **Unified storage model**: All work uses `.sow/project/` structure. No more separate `.sow/exploration/`, `.sow/design/`, `.sow/breakdown/` directories.

2. **Project type discrimination**: Branch prefix determines project type. `explore/` → ExplorationProjectState, `design/` → DesignProjectState, `breakdown/` → BreakdownProjectState. Default branches use StandardProjectState.

3. **Phase-specific state machines**: Remove constrained `Phase.status` enum. Each project type defines meaningful phase states. Exploration phases use "gathering | researching | summarizing | completed" instead of generic "pending | in_progress | completed".

4. **Schema extensions**: Six minimal schema additions enable clean translation:
   - `Artifact.approved?: bool` - supports both exploration files (no approval) and design outputs (approval required)
   - `Phase.inputs?: [...#Artifact]` - tracks sources informing the phase
   - `Task.refs?: [...#Artifact]` - exploration topic references
   - `Task.metadata?: {[string]: _}` - breakdown GitHub metadata, extensible for future needs
   - Logging for journaling - reuse existing infrastructure with `action=journal`
   - `Phase.status: string` - unconstrained, project types define valid states

5. **State-driven prompting**: Phase state transitions trigger focused orchestrator prompts. Instead of single large prompt that fades, continual re-prompting provides current context throughout session lifecycle.

This preserves mode flexibility (exploration/design remain loosely structured) while gaining project system benefits (zero-context resumability, state machines, unified tooling).

## Consequences

### Positive

- **Unified mental model**: Single "project" concept replaces three separate systems. Easier to learn, explain, document.

- **Code consolidation**: Shared patterns (state tracking, artifact management, logging) implemented once. Changes benefit all project types.

- **Workflow automation**: State machines enable programmatic transitions. Exploration → design → breakdown flows become automated with guard checks.

- **Improved context engineering**: State-driven prompting keeps orchestrator focused. State changes trigger appropriate guidance without relying on conversation history.

- **Extensibility foundation**: Adding new project types (refactoring, debugging, performance optimization) follows established pattern.

- **Simplified CLI**: Mode-specific command namespaces (`sow exploration add-topic`) consolidate to single `sow project` command with context-aware wizard. No flags needed—wizard handles project creation, resumption, and type selection.

### Negative

- **Breaking change**: Existing modes incompatible with new project structure. Active mode sessions must be restarted. Acceptable since framework not publicly used yet.

- **Increased schema complexity**: Six new optional fields add surface area. More documentation required. More validation rules to maintain.

- **State machine implementation cost**: Each project type requires state machine definition. Orchestrator must interpret and enforce state transitions.

- **Prompt engineering complexity**: State-specific prompts multiply maintenance. Each project type × phase × state combination needs crafted guidance.

### Neutral

- **Branch naming convention dependency**: Automatic type detection relies on branch prefixes. Teams with different conventions must configure mapping or use manual type specification.

- **Backward compatibility not maintained**: Clean break from mode system. No parallel support for legacy modes.

## Alternatives Considered

### Option 1: Keep Modes as Separate Systems

**Description**: Maintain current architecture with three independent mode systems.

**Pros**:
- No breaking changes
- No migration required
- Each mode remains maximally flexible
- Lower implementation cost

**Cons**:
- Code duplication persists
- No workflow automation
- Mental model fragmentation continues
- Context engineering limitations remain
- Each new mode type requires full implementation

**Why not chosen**: Technical debt continues accumulating. Long-term maintenance burden outweighs migration cost. Opportunity for state machine workflows and improved context engineering too valuable to defer.

### Option 2: Mode Plugins with Shared Core

**Description**: Extract common functionality to shared library, implement modes as plugins.

**Pros**:
- Reduces code duplication
- Maintains mode independence
- Non-breaking (can be gradual migration)
- Plugin architecture enables third-party modes

**Cons**:
- Doesn't address mental model fragmentation
- Plugin boundaries complicate state machine implementation
- Still separate directory structures and CLI namespaces
- No unified workflow automation
- More complex architecture (core + plugin system)

**Why not chosen**: Adds architectural complexity without solving core problems. Modes still separate systems. Doesn't enable state machine workflows. Plugin system overkill for three internal modes.

### Option 3: Single Project Type with Mode Flag

**Description**: Keep single project type, add `mode: "exploration" | "design" | "breakdown"` flag to control behavior.

**Pros**:
- Unified storage model
- Simpler than project type discrimination
- Easier migration path
- Less schema complexity

**Cons**:
- Mode-specific behavior via conditionals throughout codebase
- Doesn't leverage CUE schema discrimination
- Less type safety (single schema for all modes)
- State machine definitions less clear
- Harder to add new modes (must modify core project schema)

**Why not chosen**: Discriminated union approach (project types) provides better type safety, clearer separation of concerns, and easier extensibility. CUE schemas naturally express mode-specific constraints. Adding new project types doesn't require modifying existing schemas.

## Implementation Notes

**Implementation approach** (breaking change, no migration):
1. Create project type schemas: `ExplorationProjectState`, `DesignProjectState`, `BreakdownProjectState`
2. Implement six schema improvements in base phase/task/artifact definitions
3. Update CLI to detect project type from branch prefix
4. Remove mode-specific commands, directories, and code
5. Update documentation to reflect unified project system

**Justification for breaking change**: Framework currently in pre-public phase with no external users. Clean break enables faster implementation without migration tooling or deprecation period overhead. Active mode sessions (if any exist) simply restart as projects.

**Branch prefix detection and wizard workflow**:
- `explore/` → `ExplorationProjectState`
- `design/` → `DesignProjectState`
- `breakdown/` → `BreakdownProjectState`
- Other → `StandardProjectState` (default)

User runs `sow project`. CLI detects context (existing project or branch prefix), launches wizard to confirm type or allow selection. No flags required—wizard handles all interaction.

**State machine implementation**: Define state machines declaratively in project type schemas. CLI validates state transitions. Orchestrator loads state-specific prompts via `sow prompt project/<type>/<state>`.

## References

- **Exploration findings**: `.sow/knowledge/explorations/modes-to-projects-2025-01.md` - Comprehensive research documenting translation mappings, schema improvements, feature preservation analysis, and state machine design
- **Implementation design**: `.sow/knowledge/designs/project-modes-design.md` (pending) - Detailed implementation approach
- **Schema definitions**: `schemas/phases/common.cue`, `schemas/projects/*.cue`
