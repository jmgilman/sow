# Task 010: Schema Extensions and Type Generation

# Schema Extensions and Type Generation

## Overview

Extend CUE schemas to support all three project types (exploration, design, breakdown) by adding four new optional fields, updating the discriminated union, and regenerating Go types.

## Design Reference

**Primary**: `.sow/knowledge/designs/project-modes/core-design.md` - Section "Schema Extensions"
- See "Implementation Details" for exact CUE syntax and field specifications
- See "Backward Compatibility" for migration considerations

## Objectives

1. Add 4 new optional fields to CUE schemas
2. Update discriminated union for all 4 project types
3. Regenerate Go types using `go generate`
4. Write schema validation tests

## Files to Modify

- `cli/schemas/phases/common.cue` - Add 4 new optional fields
- `cli/schemas/projects/projects.cue` - Update discriminated union
- Run `go generate` to regenerate Go types

## Schema Changes

### Change 1: Make Artifact.approved optional
```cue
// cli/schemas/phases/common.cue
#Artifact: {
    approved?: bool @go(,optional=nillable)  // Changed from required bool
    // ... existing fields
}
```

### Change 2: Add Phase.inputs field
```cue
// cli/schemas/phases/common.cue
#Phase: {
    // ... existing fields
    inputs?: [...#Artifact] @go(,optional=nillable)  // NEW
}
```

### Change 3: Add Task.refs field
```cue
// cli/schemas/phases/common.cue
#Task: {
    // ... existing fields
    refs?: [...#Artifact] @go(,optional=nillable)  // NEW
}
```

### Change 4: Add Task.metadata field
```cue
// cli/schemas/phases/common.cue
#Task: {
    // ... existing fields
    metadata?: {[string]: _} @go(,optional=nillable)  // NEW
}
```

### Change 5: Update discriminated union
```cue
// cli/schemas/projects/projects.cue
#ProjectState:
    | #StandardProjectState
    | #ExplorationProjectState
    | #DesignProjectState
    | #BreakdownProjectState
```

## Acceptance Criteria

- [ ] `Artifact.approved` field changed from `bool` to `bool? @go(,optional=nillable)`
- [ ] `Phase.inputs` field added as `[...#Artifact]? @go(,optional=nillable)`
- [ ] `Task.refs` field added as `[...#Artifact]? @go(,optional=nillable)`
- [ ] `Task.metadata` field added as `{[string]: _}? @go(,optional=nillable)`
- [ ] `#ProjectState` discriminated union includes all 4 types (standard, exploration, design, breakdown)
- [ ] `go generate` runs successfully and regenerates Go types
- [ ] CUE validation passes for all schemas
- [ ] Schema validation tests written and passing
- [ ] Tests verify all 4 new fields accept expected data types
- [ ] Tests verify fields are optional (nullable)

## Testing

Write schema validation tests to verify:
- All 4 new fields accept expected data types
- Fields are properly optional (nullable)
- Discriminated union validates correctly for all 4 types
- CUE validation passes

## Important Notes

- All new fields MUST be optional to maintain backward compatibility
- Run `go generate` after CUE changes to regenerate Go types
- Verify CUE validation passes before regeneration
- This is foundation-only work - no implementation of new project types yet

## Dependencies

None - this is the first task and foundation for all others.
