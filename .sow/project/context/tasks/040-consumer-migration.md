# Migrate Consumers to libs/exec

## Context

This task is part of creating the `libs/exec` module. Tasks 010-030 created the new module with interface, implementation, and generated mock. This task migrates all consumer files from importing `cli/internal/exec` to `github.com/jmgilman/sow/libs/exec`.

There are 8 consumer files that currently import from `cli/internal/exec`:

1. `cli/internal/sow/github_cli.go`
2. `cli/internal/sow/github_cli_test.go`
3. `cli/internal/sow/github_factory.go`
4. `cli/cmd/project/wizard_state.go`
5. `cli/cmd/project/shared.go`
6. `cli/cmd/issue/show.go`
7. `cli/cmd/issue/list.go`
8. `cli/cmd/issue/check.go`

## Requirements

### 1. Update Import Paths

For each consumer file, change:
```go
import "github.com/jmgilman/sow/cli/internal/exec"
```

To:
```go
import "github.com/jmgilman/sow/libs/exec"
```

For files using the aliased import:
```go
import sowexec "github.com/jmgilman/sow/cli/internal/exec"
```

Change to:
```go
import "github.com/jmgilman/sow/libs/exec"
```

And update all references from `sowexec.NewLocal` to `exec.NewLocalExecutor`.

### 2. Update Constructor Calls

The constructor name has changed from `NewLocal` to `NewLocalExecutor`:

**Before:**
```go
ghExec := exec.NewLocal("gh")
```

**After:**
```go
ghExec := exec.NewLocalExecutor("gh")
```

### 3. Update Test Files Using MockExecutor

For test files that use `MockExecutor`, update to use the generated mock from `mocks` subpackage:

**Before:**
```go
import "github.com/jmgilman/sow/cli/internal/exec"

mock := &exec.MockExecutor{
    RunFunc: func(args ...string) (string, string, error) {
        return "output", "", nil
    },
}
```

**After:**
```go
import (
    "github.com/jmgilman/sow/libs/exec"
    "github.com/jmgilman/sow/libs/exec/mocks"
)

mock := &mocks.ExecutorMock{
    RunFunc: func(args ...string) (string, string, error) {
        return "output", "", nil
    },
}
```

### 4. Update cli/go.mod

Add the new module as a dependency in `cli/go.mod`:

```go
require github.com/jmgilman/sow/libs/exec v0.0.0
```

And add a replace directive for local development:

```go
replace github.com/jmgilman/sow/libs/exec => ../libs/exec
```

### 5. Verify All Consumers Compile

After migration, verify:
```bash
cd cli && go build ./...
```

## Acceptance Criteria

1. [ ] All 8 consumer files updated with new import path
2. [ ] All `exec.NewLocal` calls changed to `exec.NewLocalExecutor`
3. [ ] Test files using `MockExecutor` updated to use `mocks.ExecutorMock`
4. [ ] `cli/go.mod` includes dependency on `libs/exec` with replace directive
5. [ ] `cd cli && go build ./...` succeeds
6. [ ] `cd cli && go test ./...` passes
7. [ ] Linting passes: `golangci-lint run` shows no issues
8. [ ] Import aliases removed where no longer needed

## Technical Details

### Consumer File Details

#### cli/internal/sow/github_cli.go

**Current import:**
```go
import (
    "github.com/jmgilman/sow/cli/internal/exec"
)
```

**Updated import:**
```go
import (
    "github.com/jmgilman/sow/libs/exec"
)
```

**No constructor changes needed** - this file uses the interface type `exec.Executor`, which is the same.

#### cli/internal/sow/github_cli_test.go

**Current:**
```go
import (
    "github.com/jmgilman/sow/cli/internal/exec"
)

mock := &exec.MockExecutor{...}
```

**Updated:**
```go
import (
    "github.com/jmgilman/sow/libs/exec/mocks"
)

mock := &mocks.ExecutorMock{...}
```

#### cli/internal/sow/github_factory.go

**Current:**
```go
import (
    "github.com/jmgilman/sow/cli/internal/exec"
)

ghExec := exec.NewLocal("gh")
```

**Updated:**
```go
import (
    "github.com/jmgilman/sow/libs/exec"
)

ghExec := exec.NewLocalExecutor("gh")
```

#### cli/cmd/project/wizard_state.go and cli/cmd/project/shared.go

**Current (aliased):**
```go
import sowexec "github.com/jmgilman/sow/cli/internal/exec"

ghExec := sowexec.NewLocal("gh")
```

**Updated (no alias needed):**
```go
import "github.com/jmgilman/sow/libs/exec"

ghExec := exec.NewLocalExecutor("gh")
```

#### cli/cmd/issue/*.go (show.go, list.go, check.go)

**Current:**
```go
import "github.com/jmgilman/sow/cli/internal/exec"

ghExec := exec.NewLocal("gh")
```

**Updated:**
```go
import "github.com/jmgilman/sow/libs/exec"

ghExec := exec.NewLocalExecutor("gh")
```

### cli/go.mod Changes

Add at the end of the require block:
```go
require github.com/jmgilman/sow/libs/exec v0.0.0
```

Add replace directive:
```go
replace github.com/jmgilman/sow/libs/exec => ../libs/exec
```

### Import Grouping (per STYLE.md)

Imports should be grouped: stdlib, external, internal. The `libs/exec` import goes in the internal group:

```go
import (
    "context"
    "fmt"

    "github.com/spf13/cobra"

    "github.com/jmgilman/sow/libs/exec"
    "github.com/jmgilman/sow/cli/internal/sow"
)
```

## Relevant Inputs

- `cli/internal/sow/github_cli.go` - Consumer using exec.Executor interface
- `cli/internal/sow/github_cli_test.go` - Test using MockExecutor
- `cli/internal/sow/github_factory.go` - Consumer using exec.NewLocal
- `cli/cmd/project/wizard_state.go` - Consumer with aliased import
- `cli/cmd/project/shared.go` - Consumer with aliased import
- `cli/cmd/issue/show.go` - Consumer using exec.NewLocal
- `cli/cmd/issue/list.go` - Consumer using exec.NewLocal
- `cli/cmd/issue/check.go` - Consumer using exec.NewLocal
- `libs/exec/executor.go` - New interface definition
- `libs/exec/local.go` - NewLocalExecutor constructor
- `libs/exec/mocks/executor.go` - Generated mock
- `cli/go.mod` - Module file to update

## Examples

### Before/After for github_factory.go

**Before:**
```go
package sow

import (
    "os"

    "github.com/jmgilman/sow/cli/internal/exec"
)

func NewGitHubClient() (GitHubClient, error) {
    if token := os.Getenv("GITHUB_TOKEN"); token != "" {
        _ = token
    }

    ghExec := exec.NewLocal("gh")
    return NewGitHubCLI(ghExec), nil
}
```

**After:**
```go
package sow

import (
    "os"

    "github.com/jmgilman/sow/libs/exec"
)

func NewGitHubClient() (GitHubClient, error) {
    if token := os.Getenv("GITHUB_TOKEN"); token != "" {
        _ = token
    }

    ghExec := exec.NewLocalExecutor("gh")
    return NewGitHubCLI(ghExec), nil
}
```

### Before/After for github_cli_test.go mock usage

**Before:**
```go
mock := &exec.MockExecutor{
    ExistsFunc: func() bool { return true },
    RunFunc: func(args ...string) (string, string, error) {
        return `{"number": 123}`, "", nil
    },
}
```

**After:**
```go
mock := &mocks.ExecutorMock{
    ExistsFunc: func() bool { return true },
    RunFunc: func(args ...string) (string, string, error) {
        return `{"number": 123}`, "", nil
    },
}
```

## Dependencies

- Tasks 010, 020, and 030 must be completed first
- The new `libs/exec` module must exist and compile

## Constraints

- **Do not remove old cli/internal/exec/ yet** - Cleanup is task 050
- **Maintain existing behavior** - This is purely an import path change
- **Keep tests passing** - All existing tests must continue to pass
- **Update all consumers at once** - Don't leave the codebase in a split state
- Must pass `golangci-lint run` with the project's `.golangci.yml` configuration
