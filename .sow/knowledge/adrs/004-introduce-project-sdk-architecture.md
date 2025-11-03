# 004. Introduce Project SDK Architecture

**Status**: Proposed

**Date**: 2025-11-02

---

## Context

The current project architecture in `sow` hard-codes project types (standard, design, explore) directly into the CLI implementation. Each project type requires modifying core project management code, making it difficult to add new project types or customize existing ones. This tight coupling prevents users from defining domain-specific workflows and blocks the implementation of the unified command hierarchy design, which requires universal Project/Phase/Artifact/Task types that work across all project types.

Additionally, state management is scattered: task state lives in separate `state.yaml` files within task directories, requiring multiple file operations and complex synchronization. We use CUE schemas for other state files (breakdown_index, design_index) but the exploration for Project SDK initially proposed removing CUE to reduce complexity. However, this would create inconsistency across our state file validation strategy.

The project state machine currently has a gap: the `Advance()` method is project-type-specific (e.g., determining review outcomes based on metadata), yet there's no mechanism to configure this behavior per project type. Finally, all our existing schemas use CUE with code generation via `cue exp gengotypes`, and abandoning this would require maintaining two parallel validation approaches.

## Decision

We will introduce a **Project SDK** that enables defining project types through a fluent builder API while maintaining CUE schema validation for consistency. This is implemented as a major refactor in parallel packages (`internal/sdks/project`, `internal/sdks/state`) to allow incremental migration from `internal/project`.

**Key architectural choices:**

1. **Keep CUE validation** for consistency with other state files, using two-tier validation: universal CUE schemas for common fields, embedded project-type-specific schemas for metadata validation
2. **Wrapper pattern** embedding CUE-generated types in runtime types (Project, Phase, Artifact, Task) with collection APIs and state machine integration
3. **Consolidate task state** into the project state file (single atomic save, simpler synchronization) while keeping task directories for description.md, log.md, and feedback/
4. **OnAdvance configuration** allowing project types to define state-specific event determination logic (simple states return single event, complex states examine project state)
5. **Conversion layer** for Load/Save that unmarshals YAML into CUE types, validates, then converts to wrapper types (and reverse for Save)

This enables the unified command hierarchy (universal commands working across all project types), allows users to define custom project types without modifying core code, and maintains validation consistency across all state files.

## Consequences

### Positive

- Project types defined through declarative builder API without modifying core code
- Unified command hierarchy enabled (single set of commands for all project types)
- Type-safe fluent API for navigation (Phases.Get(), Outputs.Get(), Tasks.Get())
- Consistent CUE validation across all state files (breakdown_index, design_index, project state)
- Single atomic save operation (all task state in one file)
- Extensible through metadata maps with embedded validation schemas
- State machine transitions configurable per project type via OnAdvance
- Parallel implementation path allows incremental migration with zero downtime

### Negative

- Conversion overhead on every Load/Save operation (field copying between CUE and wrapper types)
- Breaking change to existing project state format (acceptable given early development stage)
- CUE knowledge required for teams wanting to define custom project types (mitigated by comprehensive examples)
- Additional complexity from maintaining both CUE types and wrapper types
- Metadata access through type assertion (10% of access) vs type-safe common fields (90%)

### Neutral

- Project types registered at startup via init() functions in type packages
- Workers and orchestrators use identical Project API regardless of type
- Task directories still exist but contain fewer files (no state.yaml)
- Machine state synchronized from wrapper type state before Save()
- Validation happens at both structural (CUE) and semantic (metadata schema) levels

## References

- [Project SDK Implementation Design](../designs/project-sdk-implementation.md) - Comprehensive design with complete schemas, API documentation, and implementation examples
- [Unified Command Hierarchy Design](../designs/command-hierarchy-design.md) - This SDK enables the unified command hierarchy by providing universal types
- [Project SDK API Specification](../explorations/project-sdk-architecture/05-project-sdk-api.md) - Original API exploration and findings
- [Project SDK Architecture Summary](../explorations/project-sdk-architecture/summary.md) - Exploration phase summary and key insights
- Current implementation: `cli/internal/project/`
- Current schemas: `cli/schemas/projects/`
- State machine: `cli/internal/project/statechart/`
