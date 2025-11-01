# ADR-001: Unified Container Model for CLI Architecture

**Status**: Proposed
**Date**: 2025-11-01
**Deciders**: Architecture Team

## Context

The sow CLI currently organizes commands by container type (project, phase, task) with specialized operations per type. This creates several architectural problems:

**Inconsistent patterns**: The same conceptual operation uses different command names depending on container type. Adding an artifact to a project uses `project add-artifact`, while adding a reference to a task uses `task add-reference`. This inconsistency makes the CLI difficult to learn and forces users to memorize type-specific commands.

**High coupling to project types**: Each project type potentially requires new CLI commands. Adding a new project type with different workflow needs means extending the CLI codebase with new commands. This creates a tight coupling between project type implementations and CLI infrastructure, making the system difficult to extend.

**Large command surface area**: The CLI currently exposes approximately 40 commands across project, phase, and task operations. Each command requires implementation, testing, documentation, and maintenance. This large surface area increases the burden of changes and the likelihood of inconsistencies.

**Unclear data flow**: Artifact management doesn't distinguish between inputs (consumed) and outputs (produced). This makes it difficult to understand data flow through the project lifecycle. Commands like `add-artifact` don't communicate directionality or purpose.

The current architecture makes the CLI rigid, difficult to extend, and costly to maintain. As sow evolves to support multiple project types beyond the standard workflow, these problems will compound.

## Decision

Adopt a unified container model where all entities (projects, phases, tasks) share identical structure and operations.

All containers have four components:
1. **Singular fields** - scalar values (name, status, timestamps, etc.)
2. **Inputs** - artifacts consumed by the container
3. **Outputs** - artifacts produced by the container
4. **Sub-containers** - nested containers (phases in projects, tasks in phases)

Three universal operations work across all container types:
1. **set** - modify any scalar field (`sow <container> set <field-path> <value>`)
2. **input/output management** - add, modify, remove, list artifacts (`sow input add --type <type>`)
3. **advance** - progress state (single command for all transitions)

Project types define semantics (valid artifact types per container) rather than commands. The CLI provides universal operations; project types configure what's valid within that universal framework.

This architecture replaces specialized commands with a small set of universal patterns, reducing command count from ~40 to ~12 while enabling unlimited project type extensibility.

## Consequences

### Positive

- **Extreme extensibility**: New project types require zero CLI changes. Project types only define valid artifact types and state transitions. The same universal commands work for all project types.

- **Consistent patterns**: Learning one container type teaches all others. If you know `sow output add --type task_list`, you automatically know `sow task output add --type modified`. Patterns transfer across the entire system.

- **Self-documenting data flow**: Explicit inputs and outputs make data flow visible in commands and state files. Seeing `input add` vs `output add` immediately communicates directionality and purpose.

- **Minimal maintenance burden**: 12 universal commands instead of 40 specialized ones. Changes to command behavior benefit all container types. Testing surface area reduced by 70%.

- **Predictable interface**: Users can infer command names from structure. Knowing the container model predicts available operations without consulting documentation.

### Negative

- **Breaking change**: Completely incompatible with current CLI. Requires migration of all existing projects and retraining of users. No gradual transition path.

- **Increased verbosity**: Some common operations become more verbose. `project approve-artifact tasks.md` becomes `output set --index 0 approved true`. Trading conciseness for consistency.

- **Abstraction overhead**: The container model adds conceptual overhead. Users must understand the inputs/outputs abstraction before using the CLI effectively. Simpler commands like `approve-artifact` require less explanation.

- **Index management burden**: Users must track artifact indices when updating. `output set --index 0` requires knowing the index. Forgetting the index or using wrong one causes errors.

### Neutral

- **State file format change**: Phases and tasks gain `inputs` and `outputs` arrays. Current `artifacts` and `references` fields migrate to appropriate category. State file size increases ~10-20% but remains manageable.

- **Command count reduction**: From ~40 to ~12 commands is dramatic but most reduced commands were convenience wrappers. Core operations remain; sugar layer removed.

- **Learning curve shifts**: Initial learning curve steeper (must understand container model), but long-term learning curve flatter (patterns transfer). Trade-off between initial and ongoing cognitive load.

## Alternatives Considered

### Option 1: Keep Specialized Commands, Add Convenience Layer

**Description**: Maintain current specialized commands (`project add-artifact`, `task add-reference`) but add universal operations as alternative interface. Both patterns coexist.

**Pros**:
- Backward compatible - existing workflows unaffected
- Gradual migration path - users adopt universal commands over time
- Familiarity preserved - no forced retraining

**Cons**:
- Two ways to do everything - confusing for new users and maintainers
- Double maintenance burden - every change requires updating both patterns
- Technical debt accumulates - eventually requires deprecation cycle anyway
- Doesn't solve extensibility problem - new project types still need new commands

**Why not chosen**: Fails to address core architectural problems. Complexity doubles instead of reducing. Eventually requires the same breaking change but with accumulated technical debt.

### Option 2: RPC-Style API with Generic Operations

**Description**: Single `sow exec` command that takes container type, operation, and arguments. Example: `sow exec project add-artifact --path tasks.md --type task_list`.

**Pros**:
- Ultimate flexibility - one command handles everything
- Easy to extend - new operations just add parameters
- Programmatically friendly - structured parameters

**Cons**:
- Poor usability - extremely verbose for common operations
- No discoverability - must consult documentation for every operation
- Loses semantic meaning - everything looks like `exec <args>`
- No type safety - parameters validated at runtime only

**Why not chosen**: Trades usability for generality. The verbosity and lack of discoverability make it unsuitable for interactive CLI use. Better suited for programmatic access than human interface.

### Option 3: GraphQL-Style Query Language

**Description**: Query language for reading state and mutation language for modifications. Example: `sow query "project.phases.planning.outputs[type=task_list]"` or `sow mutate "project.phases.planning.outputs[0].approved = true"`.

**Pros**:
- Extremely powerful querying capabilities
- Single unified language for all operations
- Familiar to developers who use GraphQL

**Cons**:
- Massive complexity for simple operations
- Requires query language parser implementation
- Difficult to provide clear error messages for syntax errors
- Over-engineered for the use case - most operations are simple CRUD

**Why not chosen**: Complexity vastly exceeds benefit. Simple operations like approving artifacts shouldn't require learning a query language. The power isn't needed for typical sow workflows.

### Option 4: Preserve Specialized Commands, Document Patterns

**Description**: Keep current specialized commands but document the patterns more clearly. Add consistency where possible without architectural changes.

**Pros**:
- No breaking changes required
- Preserves existing user knowledge
- Minimal implementation effort

**Cons**:
- Doesn't solve extensibility problem at all
- Maintains high maintenance burden
- Inconsistencies remain by design
- Each new project type still requires CLI changes

**Why not chosen**: Addresses symptoms rather than root cause. Fails to achieve goal of project type extensibility. The architectural problems persist and will worsen as more project types are added.

## Implementation Notes

**Migration strategy**:
- State file migration tool converts current schema to unified schema
- `artifacts` → `outputs` (most cases) or `inputs` (context from previous phases)
- `references` → `inputs` with type "reference"
- `files_modified` → `outputs` with type "modified"
- `feedback` → `inputs` with type "feedback"

**Rollout approach**:
- Breaking change accepted (no backward compatibility)
- Version 2.0 signals architectural shift
- Migration guide documents command mappings
- Existing projects require migration before v2.0 adoption

**Key changes**:
- Phase schema: Add `inputs: [Artifact]` and `outputs: [Artifact]` arrays
- Task schema: Replace `references`, `files_modified`, `feedback` with `inputs` and `outputs`
- Artifact schema: Universal schema with `type`, `path`, `approved`, `created_at`, `metadata`
- CLI commands: Implement universal operations, remove specialized commands
- State progression: Consolidate to single `advance` command

## References

- [Design: State Progression via Advance Command](./advance-command-design.md) - Technical details for `advance` command implementation
- [Design: Unified Command Hierarchy](./command-hierarchy-design.md) - Complete command specification and artifact types
- [Exploration: Simplified Command Hierarchy](../../knowledge/explorations/simplified_command_hierarchy.md) - Research findings and complete lifecycle examples
- [Arc42 Section 5: Building Block View](../../docs/architecture/05-building-blocks.md) - Current architecture
- [Arc42 Section 8: Cross-cutting Concepts](../../docs/architecture/08-crosscutting-concepts.md) - State management patterns
