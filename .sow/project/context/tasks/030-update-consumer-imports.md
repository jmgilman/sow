# Update Consumer Import Paths

## Context

This task is part of the `libs/schemas` module migration project. After the `libs/schemas` module is created (Tasks 010-020), this task updates all consumer files to use the new import paths.

There are approximately 75 Go files that import from `cli/schemas`:
- ~56 files import `github.com/jmgilman/sow/cli/schemas/project`
- ~22 files import `github.com/jmgilman/sow/cli/schemas` (root package)
- Some files import both

The import path changes are:
- `github.com/jmgilman/sow/cli/schemas` → `github.com/jmgilman/sow/libs/schemas`
- `github.com/jmgilman/sow/cli/schemas/project` → `github.com/jmgilman/sow/libs/schemas/project`

## Requirements

### 1. Update cli/go.mod

Add dependency on the new libs/schemas module:

```go
require (
    github.com/jmgilman/sow/libs/schemas v0.0.0
    // ... other deps
)

replace github.com/jmgilman/sow/libs/schemas => ../libs/schemas
```

### 2. Update Import Paths

For each consumer file, update import statements:

**Old:**
```go
import (
    "github.com/jmgilman/sow/cli/schemas"
    "github.com/jmgilman/sow/cli/schemas/project"
)
```

**New:**
```go
import (
    "github.com/jmgilman/sow/libs/schemas"
    "github.com/jmgilman/sow/libs/schemas/project"
)
```

### 3. Files to Update

#### Files importing `cli/schemas` (root package - ~22 files):
- `cli/cmd/agent/resume_test.go`
- `cli/cmd/agent/spawn_test.go`
- `cli/cmd/agent/spawn.go`
- `cli/cmd/config/init_test.go`
- `cli/cmd/config/show_test.go`
- `cli/cmd/config/validate_test.go`
- `cli/cmd/config/validate.go`
- `cli/cmd/refs/refs.go`
- `cli/internal/agents/executor_registry.go`
- `cli/internal/refs/file_test.go`
- `cli/internal/refs/file.go`
- `cli/internal/refs/git_test.go`
- `cli/internal/refs/git.go`
- `cli/internal/refs/index_manager.go`
- `cli/internal/refs/manager_test.go`
- `cli/internal/refs/manager.go`
- `cli/internal/refs/ref.go`
- `cli/internal/refs/types.go`
- `cli/internal/sdks/project/state/validate.go`
- `cli/internal/sow/config.go`
- `cli/internal/sow/user_config_test.go`
- `cli/internal/sow/user_config.go`

#### Files importing `cli/schemas/project` (~56 files):
- `cli/cmd/input.go`
- `cli/cmd/output.go`
- `cli/cmd/project/shared_test.go`
- `cli/cmd/project/shared.go`
- `cli/cmd/project/status_test.go`
- `cli/cmd/project/status.go`
- `cli/cmd/task_input.go`
- `cli/cmd/task_output.go`
- `cli/cmd/task.go`
- `cli/internal/cmdutil/artifacts_test.go`
- `cli/internal/cmdutil/fieldpath_test.go`
- `cli/internal/projects/breakdown/breakdown_test.go`
- `cli/internal/projects/breakdown/breakdown.go`
- `cli/internal/projects/breakdown/guards_test.go`
- `cli/internal/projects/breakdown/guards.go`
- `cli/internal/projects/breakdown/integration_test.go`
- `cli/internal/projects/breakdown/prompts_test.go`
- `cli/internal/projects/breakdown/prompts.go`
- `cli/internal/projects/design/design_test.go`
- `cli/internal/projects/design/design.go`
- `cli/internal/projects/design/guards_test.go`
- `cli/internal/projects/design/guards.go`
- `cli/internal/projects/design/integration_test.go`
- `cli/internal/projects/design/prompts_test.go`
- `cli/internal/projects/exploration/exploration_test.go`
- `cli/internal/projects/exploration/exploration.go`
- `cli/internal/projects/exploration/guards_test.go`
- `cli/internal/projects/exploration/guards.go`
- `cli/internal/projects/exploration/integration_test.go`
- `cli/internal/projects/exploration/prompts.go`
- `cli/internal/projects/standard/guards_test.go`
- `cli/internal/projects/standard/lifecycle_test.go`
- `cli/internal/projects/standard/prompts.go`
- `cli/internal/projects/standard/standard.go`
- `cli/internal/sdks/project/branch_test.go`
- `cli/internal/sdks/project/config.go`
- `cli/internal/sdks/project/guard_description_test.go`
- `cli/internal/sdks/project/integration_test.go`
- `cli/internal/sdks/project/machine_test.go`
- `cli/internal/sdks/project/state/artifact.go`
- `cli/internal/sdks/project/state/collections_test.go`
- `cli/internal/sdks/project/state/convert_test.go`
- `cli/internal/sdks/project/state/convert.go`
- `cli/internal/sdks/project/state/integration_test.go`
- `cli/internal/sdks/project/state/loader.go`
- `cli/internal/sdks/project/state/mock_test.go`
- `cli/internal/sdks/project/state/phase.go`
- `cli/internal/sdks/project/state/project_test.go`
- `cli/internal/sdks/project/state/project.go`
- `cli/internal/sdks/project/state/task.go`
- `cli/internal/sdks/project/templates/renderer_test.go`
- `cli/internal/sdks/project/templates/renderer.go`
- `cli/internal/templates/renderer.go`

### 4. Run go mod tidy

After updating all imports:

```bash
cd cli
go mod tidy
```

### 5. Verify Build and Tests

```bash
cd cli
go build ./...
go test ./...
```

## Acceptance Criteria

1. [ ] `cli/go.mod` includes `replace` directive for `libs/schemas`
2. [ ] All ~75 consumer files have updated import paths
3. [ ] No remaining references to `cli/schemas` in import statements
4. [ ] `go build ./...` succeeds in `cli/` directory
5. [ ] `go test ./...` succeeds in `cli/` directory
6. [ ] `golangci-lint run` passes (new issues only, per config)

### Verification Commands
```bash
# Check no old imports remain
grep -r '"github.com/jmgilman/sow/cli/schemas"' cli/ --include="*.go" | grep -v "^cli/schemas/"

# Should return empty - no matches outside cli/schemas/

# Verify all tests pass
cd cli && go test ./...

# Verify lint passes
cd cli && golangci-lint run
```

## Technical Details

### Import Alias Handling
Some files may use import aliases. Preserve them:

```go
// Before
import (
    "github.com/jmgilman/sow/cli/schemas/project"
)

// After (same usage, different path)
import (
    "github.com/jmgilman/sow/libs/schemas/project"
)
```

### Replace Directive
The `replace` directive is needed because `libs/schemas` is not published yet:

```go
replace github.com/jmgilman/sow/libs/schemas => ../libs/schemas
```

This allows local development. The directive will be removed when the module is published.

### gci Import Ordering
The project uses `gci` for import ordering (see `.golangci.yml`). After updating imports, run:

```bash
golangci-lint run --fix
```

This ensures imports are properly grouped:
1. Standard library
2. External packages
3. `github.com/jmgilman/sow` packages
4. Local module packages

## Relevant Inputs

- `libs/schemas/` - New module created in Tasks 010-020
- `cli/go.mod` - CLI module configuration to update
- `.golangci.yml` - Linter configuration (for gci settings)
- All consumer files listed above

## Examples

### go.mod Update
```go
module github.com/jmgilman/sow/cli

go 1.24.0

require (
    github.com/jmgilman/sow/libs/schemas v0.0.0
    // ... existing dependencies
)

replace github.com/jmgilman/sow/libs/schemas => ../libs/schemas
```

### Import Update Example
**cli/internal/sow/config.go (before):**
```go
import (
    "fmt"
    "path/filepath"

    "github.com/jmgilman/sow/cli/schemas"
    "gopkg.in/yaml.v3"
)
```

**cli/internal/sow/config.go (after):**
```go
import (
    "fmt"
    "path/filepath"

    "github.com/jmgilman/sow/libs/schemas"

    "gopkg.in/yaml.v3"
)
```

### Import Update for Project Package
**cli/internal/sdks/project/state/project.go (before):**
```go
import (
    "fmt"

    sdkstate "github.com/jmgilman/sow/cli/internal/sdks/state"
    "github.com/jmgilman/sow/cli/internal/sow"
    "github.com/jmgilman/sow/cli/schemas/project"
)
```

**cli/internal/sdks/project/state/project.go (after):**
```go
import (
    "fmt"

    "github.com/jmgilman/sow/cli/internal/sdks/state"
    "github.com/jmgilman/sow/cli/internal/sow"

    "github.com/jmgilman/sow/libs/schemas/project"
)
```

## Dependencies

- Task 010: Create libs/schemas Go module structure
- Task 020: Migrate project schemas to libs/schemas/project

Both must be completed before this task can begin.

## Constraints

- Do NOT modify any Go code logic - only import paths
- Do NOT remove the old `cli/schemas/` directory yet (separate task)
- Preserve any import aliases that exist
- Use sed, goimports, or manual editing - whatever ensures correctness
- The replace directive MUST use relative path `../libs/schemas`
