# Task 090: Register Exploration Package in Application

## Context

This task registers the exploration package with the main application so the project type is available at runtime. Package registration happens via Go's `init()` mechanism - when a package is imported, its `init()` function runs and registers the project type with the global registry.

The standard project type is already registered in `cli/cmd/root.go` using a blank import. The exploration package needs the same treatment.

## Requirements

### Update root.go

File: `cli/cmd/root.go`

Add blank import for exploration package:

1. **Locate the imports section** with the standard project import (around line 18):
   ```go
   _ "github.com/jmgilman/sow/cli/internal/projects/standard"
   ```

2. **Add exploration package import** on the next line:
   ```go
   _ "github.com/jmgilman/sow/cli/internal/projects/standard"
   _ "github.com/jmgilman/sow/cli/internal/projects/exploration"
   ```

3. **Verify import grouping** follows Go conventions (grouped with other blank imports)

### Verify Registration

After updating:

1. **Build the application**:
   ```bash
   cd cli
   go build
   ```

2. **Verify no compilation errors**

3. **Check registry at runtime** (optional manual test):
   - Create a test to verify "exploration" is in the registry
   - Or use the application to create an exploration project

## Acceptance Criteria

- [ ] Blank import added to `cli/cmd/root.go`
- [ ] Import uses correct package path
- [ ] Import grouped appropriately with other blank imports
- [ ] Application builds successfully
- [ ] No import cycle errors
- [ ] No compilation errors
- [ ] Exploration package `init()` will run at startup
- [ ] Project type will be available in registry

## Technical Details

### Blank Import Pattern

Go's blank identifier `_` allows importing a package solely for its side effects (the `init()` function):

```go
import _ "package/path"
```

This:
1. Runs the package's `init()` function
2. Does not require using any exported identifiers from the package
3. Prevents "imported and not used" compiler errors

### Package Initialization Order

Go initializes packages in dependency order:
1. All imported packages are initialized first
2. Package-level variables are initialized
3. `init()` functions are called
4. `main()` function runs (if main package)

For sow:
1. `cli/internal/projects/exploration` package loads
2. `init()` calls `state.Register("exploration", config)`
3. Config is stored in global registry
4. Application can now create exploration projects

### Import Cycles

Import cycles (A imports B, B imports A) cause compilation failures. The projects package structure avoids this:
- `cli/cmd/root.go` imports `cli/internal/projects/exploration`
- `exploration` imports SDK packages
- SDK packages don't import project packages
- No cycle exists

### Registry Pattern

The registry is a global map in `cli/internal/sdks/project/state/registry.go`:
```go
var Registry = make(map[string]ProjectTypeConfig)
```

Registration happens via:
```go
func Register(typeName string, config ProjectTypeConfig) {
    if _, exists := Registry[typeName]; exists {
        panic(fmt.Sprintf("project type already registered: %s", typeName))
    }
    Registry[typeName] = config
}
```

Projects retrieve configs via:
```go
config, exists := state.Registry["exploration"]
```

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/cmd/root.go` - File to update (line 18)
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/internal/sdks/project/state/registry.go` - Registry implementation
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/cli/internal/projects/standard/standard.go` - Reference registration
- `/Users/josh/code/sow/.sow/worktrees/36-exploration-project-type-implementation/.sow/project/context/issue-36.md` - Requirements

## Examples

### Standard Project Registration (Reference)

From `cli/cmd/root.go` (around line 18):

```go
import (
    "context"
    "fmt"
    "os"

    // ... other imports

    _ "github.com/jmgilman/sow/cli/internal/projects/standard"
)
```

### Package init() Pattern (Reference)

From `cli/internal/projects/standard/standard.go`:

```go
package standard

import (
    "github.com/jmgilman/sow/cli/internal/sdks/project"
    "github.com/jmgilman/sow/cli/internal/sdks/project/state"
    // ... other imports
)

// init registers the standard project type on package load.
func init() {
    state.Register("standard", NewStandardProjectConfig())
}
```

### Expected Result

After registration, the exploration package will be available:

```go
// In application code
config, exists := state.Registry["exploration"]
if !exists {
    return fmt.Errorf("exploration project type not registered")
}

// Use config to create/load exploration projects
```

## Dependencies

- Task 010 (Package structure) - Created exploration package with init()
- All previous tasks - Complete implementation must exist for registration to work
- This task enables the application to use the exploration project type
- Must be completed before integration testing

## Constraints

- Must use blank import pattern (underscore prefix)
- Import path must match actual package location
- Cannot create import cycles
- Must be in correct import grouping (with other blank imports)
- Package name "exploration" must match directory name
- Registry allows only one registration per type name (will panic on duplicate)
