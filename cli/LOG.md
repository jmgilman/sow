# Composable Phases Architecture - Decision Log

**Status:** In Progress
**Related:** [PLAN.md](./PLAN.md)

This log tracks important decisions, breaking changes, and deviations from the original plan during implementation of the Composable Phases Architecture MVP.

---

## 2025-01-22: Phase 1 Complete - Breaking Change: Optional Fields Now Use Pointers

### Decision

We fixed the CUE schema anti-pattern of using `*null | Type` for optional fields, replacing it with proper optional field syntax `field?: Type @go(,optional=nillable)`. This generates Go pointers instead of `any` types.

### Breaking Change

**Before (bad):**
```cue
pr_url: *null | string
discovery_type: *null | "bug" | "feature" | ...
```

Generated:
```go
Pr_url any `json:"pr_url"`                    // Requires type assertion
Discovery_type any `json:"discovery_type"`    // Requires type assertion
```

**After (good):**
```cue
pr_url?: string @go(,optional=nillable)
discovery_type?: "bug" | "feature" | ... @go(,optional=nillable)
```

Generated:
```go
Pr_url *string `json:"pr_url,omitempty"`                // Proper nil check
Discovery_type *string `json:"discovery_type,omitempty"` // Proper nil check
```

### Impact

This breaks existing code that uses type assertions on `any` fields:

**Files with errors:**
- `internal/prompts/context.go` (lines 127, 192, 242, 261)
- Potentially other files that access optional fields

**Example fix needed:**
```go
// Old (with any):
if discoveryType, ok := state.Phases.Discovery.Discovery_type.(string); ok && discoveryType != "" {
    data["DiscoveryType"] = discoveryType
}

// New (with *string):
if state.Phases.Discovery.Discovery_type != nil && *state.Phases.Discovery.Discovery_type != "" {
    data["DiscoveryType"] = *state.Phases.Discovery.Discovery_type
}
```

### Rationale

**Why we're accepting this:**

1. **Type Safety**: Compile-time type checking instead of runtime type assertions
2. **Nil Semantics**: Standard Go idiom for optional values (check `!= nil`)
3. **Better Ergonomics**: IDE autocomplete works, no guessing about types
4. **Correctness**: The old pattern was objectively wrong in CUE

**Why now:**

This is the right time to make this change because:
- We're in the middle of a major refactoring already
- The MVP hasn't been released yet
- Fixing this later would be more disruptive
- It aligns with the goal of clean architecture from the start

### Status

✅ **ACCEPTED** - We will fix consuming code as part of Phase 5 (Integration) or as we encounter errors during development.

### Affected Fields

All optional fields across the schema:

**Phase fields:**
- `started_at?`, `completed_at?` → `*time.Time`
- `discovery_type?` → `*string`
- `architect_used?`, `planner_used?` → `*bool`
- `pr_url?` → `*string`
- `dependencies?` → `[]string` (slices are naturally nillable)
- `documentation_updates?`, `artifacts_moved?`, `pending_task_additions?` → slices

**Project fields:**
- `github_issue?` → `*int64`

**Task state fields:**
- `started_at?`, `completed_at?` → `*time.Time`

---

---

## 2025-01-22: Phase 1 Complete - Type Alias Issues Resolved

### Problem

After completing Phase 1 (schema reorganization), the codebase failed to compile due to two intertwined issues:

1. **Type alias incompatibility**: Go's type system treats type aliases as distinct types. The pattern `type Task phases.Task` created incompatible types - code couldn't pass `schemas.Artifact` where `phases.Artifact` was expected, even though they were structurally identical.

2. **Pointer field breakage**: Optional fields changed from `any` to proper pointers (`*string`, `*time.Time`, `*int64`), breaking existing code that used type assertions or direct string assignments.

### Impact

**Compilation errors in 12 files:**
- 9 files with type alias issues (43 occurrences of `schemas.{Type}` needing update)
- 7 files with pointer field issues (~20 occurrences of incorrect assignments/dereferencing)

**Example errors:**
```
cannot use []phases.Task as []schemas.Task
invalid operation: Discovery_type (*string) is not an interface
```

### Resolution

**1. Removed type aliases (schemas/cue_types_gen.go)**
- Deleted 9 broken type aliases for phase and common types
- Kept only `type ProjectState projects.ProjectState` for backward compatibility

**2. Updated imports (9 files)**
- Added `import "github.com/jmgilman/sow/cli/schemas/phases"` where needed
- Added `import "github.com/jmgilman/sow/cli/schemas/projects"` where needed

**3. Updated type references (43 occurrences)**
- `schemas.Task` → `phases.Task`
- `schemas.Artifact` → `phases.Artifact`
- `schemas.ReviewReport` → `phases.ReviewReport`
- `schemas.{Phase}Phase` → `phases.{Phase}Phase`

**4. Fixed pointer assignments (~20 occurrences)**

**String to pointer:**
```go
// Before (WRONG):
state.Phases.Discovery.Discovery_type = "bug"
state.Phases.Discovery.Started_at = now.Format(time.RFC3339)

// After (CORRECT):
discoveryType := "bug"
state.Phases.Discovery.Discovery_type = &discoveryType
startedAt := now
state.Phases.Discovery.Started_at = &startedAt
```

**Pointer dereferencing:**
```go
// Before (WRONG):
if discoveryType, ok := field.(string); ok && discoveryType != "" {
    data["DiscoveryType"] = discoveryType
}

// After (CORRECT):
if field != nil && *field != "" {
    data["DiscoveryType"] = *field
}
```

**int to int64 conversion:**
```go
// Before (WRONG):
state.Project.Github_issue = &issueNumber  // issueNumber is int

// After (CORRECT):
issueNum64 := int64(issueNumber)
state.Project.Github_issue = &issueNum64
```

### Outcome

✅ **All compilation errors resolved**
✅ **All unit tests pass** (statechart, prompts, refs, sow, schemas)
✅ **Build succeeds**: `go build ./...` completes without errors

**Files modified:** 12 files across internal/, cmd/, and schemas/
**Total fixes:** ~70 changes (type references, imports, pointer handling)

### Lessons Learned

1. **Type aliases don't provide backward compatibility in Go** - distinct types even if structurally identical
2. **Pointer fields are better than `any`** - compile-time safety vs runtime type assertions
3. **CUE's `field?: Type @go(,optional=nillable)` generates proper Go pointers** - much better than `*null | Type` pattern

### Status

✅ **RESOLVED** - Phase 1 complete with clean compilation and passing tests. Ready for Phase 2 (Phase Library Foundation).

---

## Next Decisions

_(Future decisions will be logged here as they arise during MVP development)_
