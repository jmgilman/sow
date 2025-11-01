# Task Breakdown: Core Infrastructure for Project Types

## Overview

This project implements the foundational infrastructure required to support three specialized project types (exploration, design, breakdown) within the unified project system. The work is decomposed into four sequential tasks following the natural dependency order: schema changes first, then command infrastructure, then type system, and finally integration verification.

## Tasks

### Task 1: Schema Extensions and Type Generation

**Agent**: implementer

**Description**:
Extend CUE schemas to support all three project types by adding four new optional fields, update the discriminated union, and regenerate Go types.

**Design Reference**:
- **Primary**: `.sow/knowledge/designs/project-modes/core-design.md` - Section "Schema Extensions"
- See "Implementation Details" for exact CUE syntax and field specifications
- See "Backward Compatibility" for migration considerations

**Files to modify**:
- `cli/schemas/phases/common.cue` - Add 4 new optional fields
- `cli/schemas/projects/projects.cue` - Update discriminated union
- Run `go generate` to regenerate Go types

**Acceptance Criteria**:
- [ ] `Artifact.approved` field changed from `bool` to `bool? @go(,optional=nillable)`
- [ ] `Phase.inputs` field added as `[...#Artifact]? @go(,optional=nillable)`
- [ ] `Task.refs` field added as `[...#Artifact]? @go(,optional=nillable)`
- [ ] `Task.metadata` field added as `{[string]: _}? @go(,optional=nillable)`
- [ ] `#ProjectState` discriminated union includes all 4 types (standard, exploration, design, breakdown)
- [ ] `go generate` runs successfully and regenerates Go types
- [ ] CUE validation passes for all schemas
- [ ] Schema validation tests written and passing
- [ ] Tests verify all 4 new fields accept expected data types
- [ ] Tests verify fields are optional (nullable)

**Dependencies**: None

---

### Task 2: Intra-Phase State Progression Command

**Agent**: implementer

**Description**:
Implement the `sow agent advance` command infrastructure to support intra-phase state progression. Add `Advance()` method to Phase interface, create the CLI command, and update all existing standard project phases to implement the method.

**Design Reference**:
- **Primary**: `.sow/knowledge/designs/project-modes/core-design.md` - Section "Intra-Phase State Progression"
- See "Command Interface" for CLI implementation details
- See "Phase Interface Extension" for `Advance()` method signature and semantics
- Reference implementation pattern from existing phase operations that return `PhaseOperationResult`

**Files to create**:
- `cli/cmd/agent/advance.go` - New command implementation

**Files to modify**:
- `cli/internal/project/domain/phase.go` - Add `Advance()` method to Phase interface
- `cli/cmd/agent/agent.go` - Register advance command
- All existing phase implementations in `cli/internal/project/standard/` - Implement `Advance()` returning `ErrNotSupported`

**Acceptance Criteria**:
- [ ] `Phase.Advance()` method added to interface with signature `(*PhaseOperationResult, error)`
- [ ] `sow agent advance` command exists and is registered
- [ ] Command loads current project and calls `phase.Advance()`
- [ ] Command handles `ErrNotSupported` with clear message
- [ ] Command fires events from `PhaseOperationResult` when returned
- [ ] Command saves state after successful advance
- [ ] All existing standard project phases implement `Advance()` returning `(nil, project.ErrNotSupported)`
- [ ] Tests verify `ErrNotSupported` handling
- [ ] Tests verify event firing on successful advance
- [ ] Tests verify state persistence after advance

**Dependencies**: Task 1 (Go types must be regenerated)

---

### Task 3: Project Type Detection and Routing System

**Agent**: implementer

**Description**:
Implement project type detection from branch names and routing infrastructure in the loader. Create type detection function and update loader to route based on project type discriminator.

**Design Reference**:
- **Primary**: `.sow/knowledge/designs/project-modes/core-design.md` - Section "Project Type System"
- See "Branch Naming Convention" for branch prefix mapping rules
- See "Type Detection and Routing" for loader implementation pattern
- See "Discriminated Union" for how project types are represented in state

**Files to create**:
- `cli/internal/project/types.go` - Type detection function

**Files to modify**:
- `cli/internal/project/loader/loader.go` - Add routing in `Load()` and type detection in `Create()`

**Acceptance Criteria**:
- [ ] `DetectProjectType(branchName string) string` function exists in `cli/internal/project/types.go`
- [ ] Function correctly maps branch prefixes: `explore/` → "exploration", `design/` → "design", `breakdown/` → "breakdown"
- [ ] Function returns "standard" for unknown prefixes (default)
- [ ] `loader.Load()` routes based on `state.Project.Type` discriminator
- [ ] `loader.Load()` returns helpful error for unimplemented types (exploration, design, breakdown)
- [ ] `loader.Create()` detects type from branch using `DetectProjectType()`
- [ ] `loader.Create()` returns error for unimplemented types with clear message
- [ ] `loader.Create()` allows standard projects to be created normally
- [ ] Type detection tests cover all branch prefix cases
- [ ] Type detection tests verify default behavior
- [ ] Loader routing tests verify standard project loads correctly
- [ ] Loader routing tests verify error messages for unimplemented types

**Dependencies**: Task 1 (discriminated union must exist in schemas)

---

### Task 4: Integration Testing and Backward Compatibility Verification

**Agent**: implementer

**Description**:
Verify that all infrastructure changes work correctly together and that existing standard project functionality remains unchanged. Run comprehensive integration tests to ensure backward compatibility.

**Design Reference**:
- **Primary**: `.sow/knowledge/designs/project-modes/core-design.md` - Section "Testing Strategy"
- See "Backward Compatibility" for specific compatibility guarantees to verify
- See "Migration Path" for understanding how existing projects should continue working
- All previous sections provide context for what infrastructure should be tested

**Testing focus**:
- Existing standard project functionality
- New infrastructure in isolation
- Backward compatibility guarantees

**Acceptance Criteria**:
- [ ] All existing tests pass without modification
- [ ] Can create a new standard project successfully
- [ ] Can load an existing standard project successfully
- [ ] Running `sow agent advance` on standard project phase returns appropriate `ErrNotSupported` message
- [ ] Schema extensions don't break existing projects (optional fields)
- [ ] Integration tests verify end-to-end workflows
- [ ] No breaking changes introduced for existing users

**Dependencies**: Tasks 1, 2, 3 (all infrastructure must be implemented)

---

## Implementation Notes

### Key Constraints
- **Backward compatibility**: All changes must preserve existing standard project functionality
- **Minimal scope**: Foundation only, no implementation of new project types
- **Clear errors**: Unimplemented types return helpful error messages, not failures
- **Schema validation**: All CUE changes must validate before regenerating Go types

### Success Indicators
1. `go generate` completes successfully after schema changes
2. `sow agent advance` command exists and handles errors correctly
3. `DetectProjectType()` correctly maps all branch prefix patterns
4. Loader has routing infrastructure with placeholder errors for unimplemented types
5. All existing tests pass unchanged
6. Standard projects create, load, and operate normally

### Testing Strategy
- Unit tests integrated with each implementation task
- Integration tests verify end-to-end functionality
- Backward compatibility explicitly tested
- Schema validation tests ensure CUE correctness
