# Task 030: Project Type Detection and Routing System

# Project Type Detection and Routing System

## Overview

Implement project type detection from branch names and routing infrastructure in the loader. Create type detection function and update loader to route based on project type discriminator.

## Design Reference

**Primary**: `.sow/knowledge/designs/project-modes/core-design.md` - Section "Project Type System"
- See "Branch Naming Convention" for branch prefix mapping rules
- See "Type Detection and Routing" for loader implementation pattern
- See "Discriminated Union" for how project types are represented in state

## Objectives

1. Create `DetectProjectType()` function for branch name mapping
2. Update loader with type routing logic in `Load()`
3. Update loader with type detection in `Create()`
4. Implement helpful errors for unimplemented types
5. Write type detection and loader routing tests

## Files to Create

- `cli/internal/project/types.go` - Type detection function

## Files to Modify

- `cli/internal/project/loader/loader.go` - Add routing in `Load()` and type detection in `Create()`

## Implementation Details

### 1. Create Type Detection Function

```go
// cli/internal/project/types.go
package project

import "strings"

// DetectProjectType determines project type from branch name convention.
// Returns "standard" for unknown prefixes (default).
func DetectProjectType(branchName string) string {
    switch {
    case strings.HasPrefix(branchName, "explore/"):
        return "exploration"
    case strings.HasPrefix(branchName, "design/"):
        return "design"
    case strings.HasPrefix(branchName, "breakdown/"):
        return "breakdown"
    default:
        return "standard"
    }
}
```

### 2. Update Loader Load() Method

```go
// cli/internal/project/loader/loader.go

func Load(ctx *sow.Context) (domain.Project, error) {
    // ... existing existence check

    // Load state from disk
    state, _, err := statechart.LoadProjectState(ctx.FS())
    if err != nil {
        return nil, fmt.Errorf("failed to load project state: %w", err)
    }

    // Route based on project type
    switch state.Project.Type {
    case "standard":
        return standard.New((*projects.StandardProjectState)(state), ctx), nil
    case "exploration":
        // NOTE: This will fail until exploration implementation
        return nil, fmt.Errorf("exploration project type not yet implemented")
    case "design":
        return nil, fmt.Errorf("design project type not yet implemented")
    case "breakdown":
        return nil, fmt.Errorf("breakdown project type not yet implemented")
    default:
        return nil, fmt.Errorf("unknown project type: %s", state.Project.Type)
    }
}
```

### 3. Update Loader Create() Method

```go
// cli/internal/project/loader/loader.go

func Create(ctx *sow.Context, name, description string) (domain.Project, error) {
    // ... existing validation

    // Detect type from branch name
    branch, err := ctx.Git().CurrentBranch()
    if err != nil {
        return nil, fmt.Errorf("failed to get branch: %w", err)
    }

    projectType := project.DetectProjectType(branch)

    // For now, only support standard type until other types implemented
    if projectType != "standard" {
        return nil, fmt.Errorf(
            "project type %s not yet implemented - detected from branch %s",
            projectType, branch,
        )
    }

    // Create type-specific initial state
    state := statechart.NewProjectState(name, description, branch)
    return standard.New(state, ctx), nil
}
```

## Acceptance Criteria

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

## Testing

### Type Detection Tests

```go
func TestDetectProjectType(t *testing.T) {
    tests := []struct {
        branch   string
        expected string
    }{
        {"explore/auth", "exploration"},
        {"design/api", "design"},
        {"breakdown/features", "breakdown"},
        {"feat/new-thing", "standard"},
        {"main", "standard"},
    }

    for _, tt := range tests {
        got := project.DetectProjectType(tt.branch)
        assert.Equal(t, tt.expected, got)
    }
}
```

### Loader Routing Tests

- Test standard project loads correctly
- Test error messages for unimplemented types
- Test unknown type handling

## Important Notes

- This is infrastructure only - new project types are not implemented yet
- Error messages for unimplemented types should be helpful and clear
- Standard projects must continue to work unchanged
- Unknown branch prefixes default to "standard" type (backward compatible)

## Dependencies

Task 010 (discriminated union must exist in schemas)
