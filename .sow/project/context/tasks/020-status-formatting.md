# Task 020: Status Output Formatting

## Context

The status command needs a clean, readable output format. This task handles the formatting logic, including alignment and progress calculations.

## Goals

- Create consistent, readable terminal output
- Handle variable-length data gracefully (long names, many tasks)
- Calculate and display accurate progress metrics

## Requirements

### Output Formatter

Create formatting helpers in `status.go` (or a separate file if complex):

1. **Phase line formatting**:
   ```
     {phase_name}  [{status}]  {X}/{Y} tasks completed
   ```
   - Pad phase names to consistent width
   - Right-align task counts

2. **Task line formatting**:
   ```
     [{status}]  {id}  {name}
   ```
   - Consistent status width (max status is "needs_review" = 12 chars)
   - ID is always 3 digits

### Progress Calculation

For each phase:
```go
func countTasksByStatus(phase PhaseState) (completed, total int) {
    for _, task := range phase.Tasks {
        total++
        if task.Status == "completed" {
            completed++
        }
    }
    return
}
```

### Phase Order

Phases should display in lifecycle order:
1. `implementation`
2. `review`
3. `finalize`

Filter out phases that don't exist in the project.

### Edge Cases

- **No tasks in phase**: Show `0/0 tasks completed`
- **Long project name**: Display as-is (no truncation needed)
- **Empty phases map**: Show "No phases defined"
- **No tasks in current phase**: Show "No tasks" instead of empty list

## Acceptance Criteria

- [ ] Output is properly aligned and readable
- [ ] Phase names are ordered correctly (implementation, review, finalize)
- [ ] Task counts are accurate
- [ ] Edge cases handled gracefully

## Technical Notes

### Simple Approach (Recommended)

Use `fmt.Fprintf` with width specifiers:
```go
fmt.Printf("  %-15s [%-11s]  %d/%d tasks completed\n",
    phaseName, status, completed, total)
```

### Phase Ordering

```go
phaseOrder := []string{"implementation", "review", "finalize"}
for _, name := range phaseOrder {
    if phase, exists := proj.Phases[name]; exists {
        // format and print
    }
}
```

## Relevant Inputs

- `cli/cmd/project/delete.go` - Output pattern (stderr vs stdout)
- `.sow/project/state.yaml` - Sample state file structure
- `cli/schemas/project/phase.cue` - Phase schema
- `cli/schemas/project/task.cue` - Task schema with status enum
