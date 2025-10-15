# /phase:discovery - Discovery Phase Router

**Purpose**: Categorize discovery work and delegate to type-specific command
**Mode**: Router only (automatic categorization and delegation)

---

## Workflow

### 1. Update Phase Status

Update `.sow/project/state.yaml`:
```yaml
phases:
  discovery:
    status: in_progress
    started_at: [ISO 8601 timestamp]
```

Commit state change.

### 2. Categorize Work Type

Read `project.description` from state file.

**Categorization rules** (check in order):

| Keywords in Description | Category | Command |
|------------------------|----------|---------|
| "bug", "issue", "broken", "error", "doesn't work", "not working", "fails" | bug | `/phase:discovery:bug` |
| "docs out of date", "documentation gap", "code doesn't match docs", "update docs", "documentation" | docs | `/phase:discovery:docs` |
| "refactor", "messy", "clean up", "2000 lines", "reorganize", "restructure", "too complex" | refactor | `/phase:discovery:refactor` |
| "add", "new feature", "implement", "build", "create" | feature | `/phase:discovery:feature` |
| Multiple categories or no clear match | general | `/phase:discovery:general` |

**Case-insensitive matching**

### 3. Update Discovery Type

Update state with categorized type:
```yaml
phases:
  discovery:
    discovery_type: [bug|feature|docs|refactor|general]
```

Commit state change.

### 4. Invoke Type-Specific Command

**Output**:
```
Discovery phase started (type: [category])
Delegating to specialized discovery workflow...
```

**Invoke**: `/phase:discovery:[category]`

---

## Categorization Examples

```
"Fix login bug after password reset" → bug
"Add comprehensive notification system" → feature
"Code doesn't match docs/api-spec.md" → docs
"Refactor auth module - it's 2000 lines" → refactor
"Understand current system and improve docs" → general (multiple types)
"Investigate performance issues" → bug
"Build new dashboard component" → feature
"Clean up messy user service" → refactor
```

---

## Notes

- **No human interaction**: Categorization is automatic
- **Fallback to general**: When uncertain, use general category
- **State tracks type**: `discovery_type` field preserves categorization
- **Type-specific commands**: Each category has focused instructions for that work type
