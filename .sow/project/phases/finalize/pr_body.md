# feat(schemas): migrate schemas to standalone libs/schemas module

## Summary

This PR creates a new standalone `libs/schemas` Go module containing all CUE schemas and generated Go types for the sow project. The schemas are migrated from `cli/schemas/` to `libs/schemas/`, enabling reuse outside the CLI and establishing a clean foundation package with no internal dependencies.

This addresses Issue #114 and is the first step in the broader libs/ refactoring effort outlined in the exploration phase.

## Changes

### New Module: `libs/schemas/`

Created a standalone Go module with:
- `go.mod` declaring module `github.com/jmgilman/sow/libs/schemas`
- `cue.mod/module.cue` for CUE language configuration
- `embed.go` for embedding CUE files into the binary
- `doc.go` with comprehensive package documentation
- `README.md` following project documentation standards

### Migrated Schema Files

**Root package schemas:**
- `config.cue` - Repository configuration
- `user_config.cue` - User configuration
- `refs_cache.cue` - Cache metadata
- `refs_committed.cue` - Team-shared refs
- `refs_local.cue` - Local-only refs
- `knowledge_index.cue` - Knowledge tracking

**Project subpackage schemas:**
- `project/project.cue` - ProjectState, StatechartState
- `project/phase.cue` - PhaseState
- `project/task.cue` - TaskState
- `project/artifact.cue` - ArtifactState
- `project/test_helper.cue` - Test helpers

### Consumer Updates

Updated 76 files in `cli/` to use new import paths:
- `github.com/jmgilman/sow/cli/schemas` → `github.com/jmgilman/sow/libs/schemas`
- `github.com/jmgilman/sow/cli/schemas/project` → `github.com/jmgilman/sow/libs/schemas/project`

Added `replace` directive in `cli/go.mod` for local development.

### CI Updates

Updated `.github/workflows/health.yml` to validate the new module:
- Added libs/schemas to test job
- Added libs/schemas to lint job
- Updated CUE validation to use libs/schemas

### Cleanup

Removed old `cli/schemas/` directory after migration.

## Testing

- All libs/schemas tests pass (53+ schema validation tests)
- All CLI tests pass (25 packages)
- golangci-lint passes with 0 issues on both modules
- CUE schema validation passes
- Code generation is idempotent

## Breaking Changes

Import paths have changed:
```go
// Before
import "github.com/jmgilman/sow/cli/schemas"
import "github.com/jmgilman/sow/cli/schemas/project"

// After
import "github.com/jmgilman/sow/libs/schemas"
import "github.com/jmgilman/sow/libs/schemas/project"
```

## Related Issues

Closes #114

---

Generated with [sow](https://github.com/jmgilman/sow)
