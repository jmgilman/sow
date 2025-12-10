# Update Consumer Files to Use libs/config

## Context

This task updates all consumer files in the CLI that use config functions to import from the new `libs/config` module instead of `cli/internal/sow`. This is the migration step that switches the codebase to use the new, decoupled config package.

There are 8 consumer files identified:
- 6 files in `cli/cmd/config/` (validate, show, reset, path, init, edit)
- 2 files in `cli/cmd/agent/` (spawn, resume)

The changes are primarily import path changes and minor signature adjustments for functions that now accept explicit parameters instead of Context.

## Requirements

### Import Changes

Update imports in all consumer files:
```go
// Before
import "github.com/jmgilman/sow/cli/internal/sow"

// After
import "github.com/jmgilman/sow/libs/config"
```

### Function Call Changes

#### sow.GetUserConfigPath() -> config.GetUserConfigPath()
Used in: `cli/cmd/config/validate.go`, `show.go`, `reset.go`, `path.go`, `init.go`, `edit.go`

No signature change - just import path.

#### sow.LoadUserConfig() -> config.LoadUserConfig()
Used in: `cli/cmd/agent/spawn.go`, `resume.go`

No signature change - just import path.

#### sow.LoadUserConfigFromPath() -> config.LoadUserConfigFromPath()
Used in: `cli/cmd/config/show.go`

No signature change - just import path.

#### sow.ValidateUserConfig() -> config.ValidateUserConfig()
Used in: `cli/cmd/config/validate.go`

No signature change - just import path.

### Files to Update

1. **cli/cmd/config/validate.go**
   - Change import from `internal/sow` to `libs/config`
   - Change `sow.GetUserConfigPath()` to `config.GetUserConfigPath()`
   - Change `sow.ValidateUserConfig()` to `config.ValidateUserConfig()`

2. **cli/cmd/config/show.go**
   - Change import from `internal/sow` to `libs/config`
   - Change `sow.GetUserConfigPath()` to `config.GetUserConfigPath()`
   - Change `sow.LoadUserConfigFromPath()` to `config.LoadUserConfigFromPath()`

3. **cli/cmd/config/reset.go**
   - Change import from `internal/sow` to `libs/config`
   - Change `sow.GetUserConfigPath()` to `config.GetUserConfigPath()`

4. **cli/cmd/config/path.go**
   - Change import from `internal/sow` to `libs/config`
   - Change `sow.GetUserConfigPath()` to `config.GetUserConfigPath()`

5. **cli/cmd/config/init.go**
   - Change import from `internal/sow` to `libs/config`
   - Change `sow.GetUserConfigPath()` to `config.GetUserConfigPath()`

6. **cli/cmd/config/edit.go**
   - Change import from `internal/sow` to `libs/config`
   - Change `sow.GetUserConfigPath()` to `config.GetUserConfigPath()`

7. **cli/cmd/agent/spawn.go**
   - Change import from `internal/sow` to `libs/config`
   - Change `sow.LoadUserConfig()` to `config.LoadUserConfig()`

8. **cli/cmd/agent/resume.go**
   - Change import from `internal/sow` to `libs/config`
   - Change `sow.LoadUserConfig()` to `config.LoadUserConfig()`

## Acceptance Criteria

1. [ ] All 8 consumer files updated with new imports
2. [ ] All function calls use `config.` prefix instead of `sow.`
3. [ ] All files compile without errors
4. [ ] All existing tests still pass
5. [ ] `go build ./...` succeeds from cli directory
6. [ ] `go test ./...` passes for affected packages
7. [ ] `golangci-lint run` passes

### Verification Steps

After making changes, run:
```bash
cd cli
go mod tidy
go build ./...
go test ./cmd/config/... ./cmd/agent/...
golangci-lint run ./cmd/config/... ./cmd/agent/...
```

## Technical Details

### Import Organization

Imports should be organized per STYLE.md (stdlib, external, internal):

```go
import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "gopkg.in/yaml.v3"

    "github.com/jmgilman/sow/cli/internal/cmdutil"
    "github.com/jmgilman/sow/libs/config"
    "github.com/jmgilman/sow/libs/schemas"
)
```

### Note on cli/internal/sow Import

Some files may still need to import `cli/internal/sow` for other functionality (like `sow.Context`, `sow.ErrNotInitialized`). Only remove the import if no other symbols from that package are used.

### go.mod Changes

The CLI's `go.mod` will need to reference the new module. Add to `cli/go.mod`:

```
require (
    github.com/jmgilman/sow/libs/config v0.0.0
)

replace github.com/jmgilman/sow/libs/config => ../libs/config
```

## Relevant Inputs

- `cli/cmd/config/validate.go` - Uses GetUserConfigPath, ValidateUserConfig
- `cli/cmd/config/show.go` - Uses GetUserConfigPath, LoadUserConfigFromPath
- `cli/cmd/config/reset.go` - Uses GetUserConfigPath
- `cli/cmd/config/path.go` - Uses GetUserConfigPath
- `cli/cmd/config/init.go` - Uses GetUserConfigPath
- `cli/cmd/config/edit.go` - Uses GetUserConfigPath
- `cli/cmd/agent/spawn.go` - Uses LoadUserConfig
- `cli/cmd/agent/resume.go` - Uses LoadUserConfig
- `cli/go.mod` - Module dependencies to update

## Examples

### Before (validate.go)

```go
package config

import (
    "github.com/jmgilman/sow/cli/internal/sow"
    "github.com/jmgilman/sow/libs/schemas"
    // ...
)

func runValidate(cmd *cobra.Command, _ []string) error {
    path, err := sow.GetUserConfigPath()
    // ...
    if err := sow.ValidateUserConfig(&config); err != nil {
    // ...
}
```

### After (validate.go)

```go
package config

import (
    "github.com/jmgilman/sow/libs/config"
    "github.com/jmgilman/sow/libs/schemas"
    // ...
)

func runValidate(cmd *cobra.Command, _ []string) error {
    path, err := config.GetUserConfigPath()
    // ...
    if err := config.ValidateUserConfig(&cfg); err != nil {
    // ...
}
```

### Before (spawn.go)

```go
package agent

import (
    "github.com/jmgilman/sow/cli/internal/sow"
    // ...
)

func runSpawn(cmd *cobra.Command, args []string, ...) error {
    // ...
    userConfig, err := sow.LoadUserConfig()
    // ...
}
```

### After (spawn.go)

```go
package agent

import (
    "github.com/jmgilman/sow/cli/internal/sow"  // May still be needed for other symbols
    "github.com/jmgilman/sow/libs/config"
    // ...
)

func runSpawn(cmd *cobra.Command, args []string, ...) error {
    // ...
    userConfig, err := config.LoadUserConfig()
    // ...
}
```

## Dependencies

- Tasks 010-040 must be completed first (module exists with all functionality)

## Constraints

- Do NOT change any business logic - only imports and function prefixes
- Do NOT remove old code from `cli/internal/sow/` yet - that's the next task
- Ensure tests still pass after migration
- Keep imports organized per STYLE.md
