# Task 010: Add huh Library Dependency

## Context

This task is part of building an interactive wizard to replace the existing flag-based `sow project new` and `sow project continue` commands. The wizard will provide a guided, terminal-based UI for creating and continuing projects.

The charmbracelet/huh library provides terminal UI components (select prompts, text inputs, forms) that have been verified to support all requirements in the design documents. This task adds the library as a dependency to the project.

**Design verification**: The huh library capabilities have been thoroughly verified in `.sow/knowledge/designs/huh-library-verification.md` and confirmed to support:
- Select prompts with options
- Text input with validation
- Text areas with external editor support (Ctrl+E)
- Loading spinners
- Error display with forms

## Requirements

Add the charmbracelet/huh library and its spinner subpackage to the project's Go dependencies:

1. **Add main huh package**: `github.com/charmbracelet/huh`
   - Use latest stable version available
   - This provides form components, select prompts, input fields, text areas

2. **Add spinner subpackage**: `github.com/charmbracelet/huh/spinner`
   - Provides loading indicators for long-running operations
   - Used to display progress during git operations, GitHub API calls, etc.

3. **Update go.mod and go.sum**:
   - Run `go get` to fetch dependencies
   - Run `go mod tidy` to clean up
   - Verify no dependency conflicts

4. **Verify installation**:
   - Ensure the library can be imported
   - Check that no conflicts exist with existing dependencies

## Acceptance Criteria

- [ ] `cli/go.mod` contains `github.com/charmbracelet/huh` with a version number
- [ ] `cli/go.mod` contains `github.com/charmbracelet/huh/spinner` (or it's included transitively)
- [ ] `cli/go.sum` is updated with checksums for the new dependencies
- [ ] `go mod tidy` runs without errors from the `cli/` directory
- [ ] No dependency conflicts reported by Go
- [ ] A simple test import compiles successfully:
  ```go
  import (
      "github.com/charmbracelet/huh"
      "github.com/charmbracelet/huh/spinner"
  )
  ```

## Technical Details

**Go Module Location**: `/cli/go.mod`
- The project uses Go modules with the module path `github.com/jmgilman/sow/cli`
- Current Go version: 1.25.3

**Installation Commands**:
```bash
cd cli/
go get github.com/charmbracelet/huh
go get github.com/charmbracelet/huh/spinner
go mod tidy
```

**Testing Installation**:
Create a temporary test file to verify imports work:
```go
package main

import (
    "github.com/charmbracelet/huh"
    "github.com/charmbracelet/huh/spinner"
)

func main() {
    // Just verify imports compile
    _ = huh.NewForm()
    _ = spinner.New()
}
```

Then compile it:
```bash
go build -o /tmp/test-huh
```

If it compiles successfully, delete the test file.

## Relevant Inputs

- `cli/go.mod` - Current Go module file where dependency will be added
- `.sow/knowledge/designs/huh-library-verification.md` - Verification that huh supports all required features
- `.sow/knowledge/designs/interactive-wizard-technical-implementation.md` - Technical design showing how huh will be used

## Examples

**Example go.mod entry after installation**:
```go
require (
    github.com/charmbracelet/huh v0.6.0 // or whatever latest version is
    github.com/spf13/cobra v1.9.1
    github.com/stretchr/testify v1.11.1
    // ... other dependencies
)
```

**Example successful output**:
```bash
$ cd cli/
$ go get github.com/charmbracelet/huh
go: downloading github.com/charmbracelet/huh v0.6.0
go: added github.com/charmbracelet/huh v0.6.0

$ go mod tidy
# no errors

$ go build ./...
# successful build
```

## Dependencies

None - this is the first task and has no dependencies on other tasks.

## Constraints

- **No code changes**: Only modify `go.mod` and `go.sum` - do not write any Go code yet
- **Use latest stable**: Prefer the latest stable release of huh, not pre-release versions
- **Verify compatibility**: Ensure the library version is compatible with Go 1.25.3
- **Clean module**: Run `go mod tidy` to ensure the module file is clean

## Testing Requirements

**Manual Verification**:
1. Verify `go.mod` contains the huh dependency
2. Verify `go mod tidy` runs without errors
3. Verify a test file with huh imports compiles successfully
4. Verify existing tests still pass: `go test ./...`

**No unit tests needed** for this task as it only adds a dependency.

## Implementation Notes

**Why huh?**
The library has been verified to support all wizard requirements:
- Arrow key navigation for select prompts
- Enter to select, Esc to cancel
- Input validation with inline errors
- External editor support for long text (Ctrl+E, not Ctrl+O as originally designed)
- Loading spinners for async operations
- Clean, professional terminal UI

**Version Selection**:
Use the latest stable version. As of the design documents, the library is actively maintained and stable.

**Spinner Subpackage**:
The spinner package may be included automatically with the main huh package, but explicitly getting it ensures it's available.

## Success Indicators

After completing this task:
1. The huh library is available for import in all Go files
2. No compilation errors related to dependencies
3. Subsequent tasks can begin implementing wizard screens using huh components
4. The foundation is set for all interactive UI work
