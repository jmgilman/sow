# Create libs/schemas Go Module Structure

## Context

This task is part of the `libs/schemas` module migration project. The goal is to move ALL CUE schemas from `cli/schemas/` to a new standalone `libs/schemas/` Go module, enabling reuse outside the CLI and establishing a clean foundation package with no dependencies.

The new `libs/schemas` module will contain:
- Config schemas (`Config`, `UserConfig`)
- Refs schemas (`RefsCommittedIndex`, `RefsLocalIndex`, `RefsCacheIndex`)
- Knowledge index schema (`KnowledgeIndex`)
- Project schemas in a `project/` subpackage

This task creates the module structure with root-level schemas only. Project schemas are handled in a separate task.

## Requirements

### 1. Create Directory Structure

Create the following directories and files:

```
libs/schemas/
├── go.mod                    # New Go module
├── go.sum                    # Dependencies (will be populated)
├── cue.mod/
│   └── module.cue            # CUE module definition
├── embed.go                  # Embed CUE files
├── config.cue                # Config schema (copy from cli/schemas/)
├── user_config.cue           # UserConfig schema (copy from cli/schemas/)
├── refs_cache.cue            # RefsCacheIndex schema (copy from cli/schemas/)
├── refs_committed.cue        # RefsCommittedIndex schema (copy from cli/schemas/)
├── refs_local.cue            # RefsLocalIndex schema (copy from cli/schemas/)
└── knowledge_index.cue       # KnowledgeIndex schema (copy from cli/schemas/)
```

### 2. Create go.mod

```go
module github.com/jmgilman/sow/libs/schemas

go 1.24.0

require (
    cuelang.org/go v0.12.0
)
```

Note: Check the current Go version in `cli/go.mod` and use the same version. CUE dependency version should match `cli/go.mod` as well.

### 3. Create cue.mod/module.cue

```cue
module: "github.com/jmgilman/sow/libs/schemas"
language: {
    version: "v0.13.2"
}
```

Note: Match the CUE language version from `cli/schemas/cue.mod/module.cue`.

### 4. Create embed.go

Create an embed.go file that embeds all CUE files:

```go
package schemas

import "embed"

// CUESchemas embeds all CUE schema files from the schemas package.
// This allows the schemas to be bundled into the binary and loaded at runtime.
// Includes subdirectories and cue.mod for import resolution.
//
//go:embed *.cue project/*.cue cue.mod/module.cue
var CUESchemas embed.FS
```

### 5. Copy CUE Schema Files

Copy the following files from `cli/schemas/` to `libs/schemas/`:
- `config.cue` - Keep as-is
- `user_config.cue` - Keep as-is
- `refs_cache.cue` - Keep as-is
- `refs_committed.cue` - Keep as-is
- `refs_local.cue` - Keep as-is
- `knowledge_index.cue` - Keep as-is

**Important**: Do NOT modify the CUE files' content. Only copy them.

### 6. Generate Go Types

Run CUE code generation to create `cue_types_gen.go`:

```bash
cd libs/schemas
cue exp gengotypes ./...
```

This should generate a `cue_types_gen.go` file with Go types for all root-level schemas.

### 7. Verify Module Compiles

```bash
cd libs/schemas
go mod tidy
go build ./...
```

## Acceptance Criteria

1. [ ] `libs/schemas/` directory exists with all required files
2. [ ] `go.mod` declares module path `github.com/jmgilman/sow/libs/schemas`
3. [ ] `cue.mod/module.cue` declares CUE module with correct path
4. [ ] All 6 CUE files are copied and unmodified from source
5. [ ] `embed.go` correctly embeds CUE files (pattern includes `project/*.cue` for future)
6. [ ] `cue_types_gen.go` is generated with all Go types:
   - `Config`
   - `UserConfig`
   - `RefsCacheIndex`, `CachedRef`, `CacheMetadata`, `GitMetadata`, `FileMetadata`, `CacheUsage`
   - `RefsCommittedIndex`, `Ref`, `RefConfig`
   - `RefsLocalIndex`
   - `KnowledgeIndex`, `ExplorationSummary`, `ArtifactReference`
7. [ ] `go build ./...` succeeds
8. [ ] `go mod tidy` produces clean go.sum

## Technical Details

### Package Name
The package name must be `schemas` to match the import convention:
```go
import "github.com/jmgilman/sow/libs/schemas"
```

### CUE Generation Command
The doc.go file should include a `go:generate` directive:
```go
//go:generate sh -c "cue exp gengotypes ./... && rm -f project/cue_types_gen.go"
```

Note: The `rm` command handles the case where project/ hasn't been migrated yet.

### golangci-lint Considerations
The generated `cue_types_gen.go` uses underscores in field names to match JSON schema. The linter exclusion for revive's var-naming rule already exists in `.golangci.yml`.

## Relevant Inputs

- `cli/schemas/go.mod` - Reference for Go and dependency versions
- `cli/schemas/cue.mod/module.cue` - Reference for CUE language version
- `cli/schemas/embed.go` - Pattern for CUE file embedding
- `cli/schemas/config.cue` - Source file to copy
- `cli/schemas/user_config.cue` - Source file to copy
- `cli/schemas/refs_cache.cue` - Source file to copy
- `cli/schemas/refs_committed.cue` - Source file to copy
- `cli/schemas/refs_local.cue` - Source file to copy
- `cli/schemas/knowledge_index.cue` - Source file to copy
- `cli/schemas/cue_types_gen.go` - Reference for expected generated output
- `.golangci.yml` - Linter configuration (for understanding exclusions)

## Examples

### Expected go.mod content
```go
module github.com/jmgilman/sow/libs/schemas

go 1.24.0

require cuelang.org/go v0.12.0
```

### Expected embed.go content
```go
package schemas

import "embed"

// CUESchemas embeds all CUE schema files from the schemas package.
// This allows the schemas to be bundled into the binary and loaded at runtime.
// Includes subdirectories (project/) and cue.mod for import resolution.
//
//go:embed *.cue project/*.cue cue.mod/module.cue
var CUESchemas embed.FS
```

## Dependencies

None - this is the first task in the migration.

## Constraints

- Do NOT modify the content of CUE schema files (exact copy)
- Do NOT create documentation files yet (separate task)
- Do NOT create the project/ subpackage yet (separate task)
- Package name MUST be `schemas` (not `schema` or anything else)
- CUE version in module.cue must match existing cli/schemas/cue.mod/module.cue
