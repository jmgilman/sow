# Issue #114: libs/schemas module

**URL**: https://github.com/jmgilman/sow/issues/114
**State**: OPEN

## Description

# Work Unit 001: libs/schemas Module

Create a new `libs/schemas` Go module containing ALL CUE schemas and generated Go types for the sow project.

## Objective

Move the entire `cli/schemas/` package to `libs/schemas/`, creating a standalone Go module with clean API design, full standards compliance, and proper documentation.

**This is NOT just a file move.** This is an opportunity to clean up, standardize, and create a well-designed public API.

## Scope

### What's Moving
- Project schemas: `ProjectState`, `PhaseState`, `TaskState`, `ArtifactState`
- Config schemas: `Config`, `UserConfig`
- Refs schemas: `RefsCommittedIndex`, `RefsLocalIndex`, `RefsCache`
- Knowledge schemas: `KnowledgeIndex`
- All CUE files and embedded resources
- Code generation infrastructure

### Target Structure
```
libs/schemas/
├── go.mod
├── go.sum
├── README.md                 # Per READMES.md standard
├── doc.go                    # Package documentation
├── embed.go                  # CUE file embedding
├── cue.mod/module.cue        # CUE module definition
│
├── config.cue                # Repo config schema
├── user_config.cue           # User config schema
├── refs_cache.cue            # Refs cache schema
├── refs_committed.cue        # Refs committed index
├── refs_local.cue            # Refs local index
├── knowledge_index.cue       # Knowledge index schema
│
├── project/                  # Project-related schemas
│   ├── project.cue           # ProjectState
│   ├── phase.cue             # PhaseState
│   ├── task.cue              # TaskState
│   ├── artifact.cue          # ArtifactState
│   └── cue_types_gen.go      # Generated Go types
│
└── cue_types_gen.go          # Generated: Config, UserConfig, Refs*, KnowledgeIndex
```

## Standards Requirements

### Go Code Standards (STYLE.md)
- All generated and hand-written Go code must pass golangci-lint
- Error handling with proper wrapping (`%w`)
- No global mutable state
- Accept interfaces, return concrete types
- Proper naming conventions (MixedCaps, acronyms uppercase)

### Testing Standards (TESTING.md)
- Behavioral test coverage for any validation logic
- Table-driven tests with `t.Run()`
- Use `testify/assert` and `testify/require`
- No external dependencies in unit tests

### README Standards (READMES.md)
- Overview (what/why/for whom)
- Quick Start (import and basic usage)
- Usage examples for common tasks
- Configuration (code generation setup)
- Links to CUE documentation

### Linting
- Must pass `golangci-lint run` with project's `.golangci.yml`
- Proper error wrapping for external errors
- No unchecked errors

## API Design Requirements

### Clean Public Interface
- Export only what consumers need
- Consistent naming across all schema types
- Consider helper functions for common operations (e.g., schema validation)

### CUE Integration
- Clean embedding of CUE files
- Documented code generation process
- Reproducible builds (`go generate` must be idempotent)

## Consumer Impact

~69 files currently import from `cli/schemas/`. After this work:
- All imports change from `github.com/jmgilman/sow/cli/schemas` to `github.com/jmgilman/sow/libs/schemas`
- Import paths for project types: `github.com/jmgilman/sow/libs/schemas/project`

## Dependencies

None - this is a foundation package with no internal dependencies.

## Acceptance Criteria

1. [ ] New `libs/schemas` Go module exists and compiles
2. [ ] All CUE schemas moved and organized
3. [ ] Code generation works (`go generate ./...`)
4. [ ] All tests pass with proper coverage
5. [ ] `golangci-lint run` passes with no issues
6. [ ] README.md follows READMES.md standard
7. [ ] Package documentation in doc.go
8. [ ] All 69 consumer files updated to new import paths
9. [ ] Old `cli/schemas/` removed (except any CLI-specific items)
10. [ ] No regression in existing functionality

## Out of Scope

- Adding new schemas
- Changing schema structure
- Adding new validation logic beyond what exists
