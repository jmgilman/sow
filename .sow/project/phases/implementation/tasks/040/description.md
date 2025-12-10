# Implement Path Helper Functions

## Context

This task implements path helper functions for the `libs/config` module. These functions compute absolute paths to configuration-defined directories like ADRs and design docs.

The key change from the original implementation is **accepting explicit `repoRoot` instead of `*Context`**. The original functions like `GetADRsPath(ctx *Context, config *schemas.Config)` accepted the CLI's Context type. The new API accepts explicit parameters:

```go
// Before (coupled to Context)
func GetADRsPath(ctx *Context, config *schemas.Config) string

// After (explicit dependencies)
func GetADRsPath(repoRoot string, config *schemas.Config) string
```

This enables the config package to be used outside the CLI context.

## Requirements

Create `libs/config/paths.go` with the following public API:

### GetADRsPath

```go
// GetADRsPath returns the absolute path to the ADRs directory.
// The path is computed as: <repoRoot>/.sow/knowledge/<config.artifacts.adrs>
// If config is nil or the ADRs path is not configured, uses DefaultADRsPath.
func GetADRsPath(repoRoot string, config *schemas.Config) string
```

### GetDesignDocsPath

```go
// GetDesignDocsPath returns the absolute path to the design docs directory.
// The path is computed as: <repoRoot>/.sow/knowledge/<config.artifacts.design_docs>
// If config is nil or the design docs path is not configured, uses DefaultDesignDocsPath.
func GetDesignDocsPath(repoRoot string, config *schemas.Config) string
```

### GetExplorationsPath

```go
// GetExplorationsPath returns the absolute path to the explorations directory.
// This path is not configurable and always uses DefaultExplorationsPath.
// The path is computed as: <repoRoot>/.sow/knowledge/explorations
func GetExplorationsPath(repoRoot string) string
```

### GetKnowledgePath

```go
// GetKnowledgePath returns the absolute path to the knowledge directory.
// The path is computed as: <repoRoot>/.sow/knowledge
func GetKnowledgePath(repoRoot string) string
```

## Acceptance Criteria

1. [ ] `paths.go` implements all four path helper functions
2. [ ] All functions return absolute paths using `filepath.Join`
3. [ ] Functions handle nil config gracefully (use defaults)
4. [ ] Functions handle nil/unset config fields gracefully (use defaults)
5. [ ] All tests pass with proper behavioral coverage
6. [ ] `golangci-lint run` passes

### Test Requirements (TDD - write tests first)

Create `libs/config/paths_test.go` with table-driven tests covering:

1. **GetADRsPath behaviors**:
   - Config with custom ADRs path -> uses custom path
   - Config with nil Artifacts -> uses default
   - Config with nil Adrs field -> uses default
   - nil config -> uses default
   - Various repoRoot values (relative, absolute, trailing slash)

2. **GetDesignDocsPath behaviors**:
   - Config with custom design_docs path -> uses custom path
   - Config with nil Artifacts -> uses default
   - Config with nil Design_docs field -> uses default
   - nil config -> uses default

3. **GetExplorationsPath behaviors**:
   - Returns correct path with repoRoot
   - Handles various repoRoot formats

4. **GetKnowledgePath behaviors**:
   - Returns correct path with repoRoot

5. **Path construction**:
   - All paths are properly joined (no double slashes)
   - Works on Unix path format

## Technical Details

### Import Dependencies

```go
import (
    "path/filepath"

    "github.com/jmgilman/sow/libs/schemas"
)
```

### Path Structure

The knowledge directory structure under `.sow/`:

```
.sow/
├── knowledge/
│   ├── adrs/           # GetADRsPath returns this
│   ├── design/         # GetDesignDocsPath returns this
│   └── explorations/   # GetExplorationsPath returns this
└── config.yaml
```

### Handling Nil Pointers

The Config type has nested pointers that need careful handling:

```go
func GetADRsPath(repoRoot string, config *schemas.Config) string {
    path := DefaultADRsPath
    if config != nil && config.Artifacts != nil && config.Artifacts.Adrs != nil {
        path = *config.Artifacts.Adrs
    }
    return filepath.Join(repoRoot, ".sow", "knowledge", path)
}
```

### filepath.Join Usage

Use `filepath.Join` for all path construction:
- Handles path separators correctly per platform
- Cleans redundant separators and dots
- Example: `filepath.Join("/repo", ".sow", "knowledge", "adrs")` -> `/repo/.sow/knowledge/adrs`

## Relevant Inputs

- `cli/internal/sow/config.go:82-101` - Original path helper implementations
- `libs/config/defaults.go` - Default path constants (from Task 010)
- `.standards/TESTING.md` - Testing requirements
- `.standards/STYLE.md` - Code style requirements

## Examples

### Usage Example

```go
// Load config and get paths
cfg, err := config.LoadRepoConfig(fs)
if err != nil {
    return err
}

repoRoot := "/home/user/myproject"
adrsPath := config.GetADRsPath(repoRoot, cfg)
// -> "/home/user/myproject/.sow/knowledge/adrs" (or custom if configured)

designPath := config.GetDesignDocsPath(repoRoot, cfg)
// -> "/home/user/myproject/.sow/knowledge/design" (or custom if configured)

explorationsPath := config.GetExplorationsPath(repoRoot)
// -> "/home/user/myproject/.sow/knowledge/explorations" (always default)
```

### Test Example

```go
func TestGetADRsPath(t *testing.T) {
    customPath := "custom-adrs"

    tests := []struct {
        name     string
        repoRoot string
        config   *schemas.Config
        want     string
    }{
        {
            name:     "custom path configured",
            repoRoot: "/repo",
            config: &schemas.Config{
                Artifacts: &struct{
                    Adrs        *string `json:"adrs,omitempty"`
                    Design_docs *string `json:"design_docs,omitempty"`
                }{
                    Adrs: &customPath,
                },
            },
            want: "/repo/.sow/knowledge/custom-adrs",
        },
        {
            name:     "nil config uses default",
            repoRoot: "/repo",
            config:   nil,
            want:     "/repo/.sow/knowledge/adrs",
        },
        {
            name:     "nil Artifacts uses default",
            repoRoot: "/repo",
            config:   &schemas.Config{Artifacts: nil},
            want:     "/repo/.sow/knowledge/adrs",
        },
        {
            name:     "nil Adrs field uses default",
            repoRoot: "/repo",
            config: &schemas.Config{
                Artifacts: &struct{
                    Adrs        *string `json:"adrs,omitempty"`
                    Design_docs *string `json:"design_docs,omitempty"`
                }{
                    Adrs: nil,
                },
            },
            want: "/repo/.sow/knowledge/adrs",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := GetADRsPath(tt.repoRoot, tt.config)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

## Dependencies

- Task 010 (module structure) must be completed first - provides defaults.go with constants
- Tasks 020/030 are independent but should be done before this for consistency

## Constraints

- Functions must handle all nil pointer cases gracefully
- Use `filepath.Join` for all path construction
- Functions are pure (no side effects, no filesystem access)
- Functions must be under 80 lines (these will be very short)
- Do NOT validate that paths exist - that's the caller's responsibility
