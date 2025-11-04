# Task 010: Create Package Structure and Copy Templates

## Overview

Create the new `internal/projects/standard/` package directory structure and physically copy all template files from the existing standard project implementation. This task sets up the foundation for the SDK-based migration while preserving all existing template content.

**Critical**: The old `internal/project/standard/` package must remain completely untouched. We are creating a parallel implementation that will coexist until Unit 5 migrates CLI commands.

## Context

**Design Reference**: See `.sow/knowledge/designs/project-sdk-implementation.md` (lines 1395-1436) for package structure details.

**Why This Matters**: The templates must be copied (not referenced) because the old `internal/project/` package will be deleted in Unit 6. The new package must be self-contained with its own templates.

## Requirements

### Directory Structure

Create the following directory structure:

```
cli/internal/projects/standard/
├── cue/                    # Metadata schemas (Task 3)
├── templates/              # Prompt templates (this task)
├── states.go              # State constants (Task 2)
├── events.go              # Event constants (Task 2)
├── metadata.go            # Schema embeddings (Task 3)
├── guards.go              # Guard functions (Task 5)
├── guards_test.go         # Guard tests (Task 5)
├── prompts.go             # Prompt functions (Task 4)
├── standard.go            # SDK configuration (Task 6)
└── lifecycle_test.go      # Integration tests (Task 6)
```

### Template Files to Copy

Copy **all 8 template files** from `cli/internal/project/standard/templates/` to `cli/internal/projects/standard/templates/`:

1. `planning_active.md`
2. `implementation_planning.md`
3. `implementation_executing.md`
4. `review_active.md`
5. `finalize_documentation.md`
6. `finalize_checks.md`
7. `finalize_delete.md`
8. `no_project.md`

**Important**: Copy files exactly as-is with no modifications. These are comprehensive prompt templates (100+ lines each) containing detailed instructions for each project state.

## Acceptance Criteria

- [ ] Directory `cli/internal/projects/standard/` created
- [ ] Subdirectory `cli/internal/projects/standard/cue/` created
- [ ] Subdirectory `cli/internal/projects/standard/templates/` created
- [ ] All 8 template files copied from `cli/internal/project/standard/templates/` to `cli/internal/projects/standard/templates/`
- [ ] Copied templates are byte-for-byte identical to originals
- [ ] Old package `cli/internal/project/standard/` completely untouched (verify with `git diff`)
- [ ] No other files created (just directories and templates)

## Validation Commands

```bash
# Verify directory structure
ls -la cli/internal/projects/standard/
ls -la cli/internal/projects/standard/cue/
ls -la cli/internal/projects/standard/templates/

# Verify all 8 templates copied
ls cli/internal/projects/standard/templates/ | wc -l  # Should be 8

# Verify old package untouched
git diff cli/internal/project/standard/

# Verify templates are identical
diff cli/internal/project/standard/templates/planning_active.md \
     cli/internal/projects/standard/templates/planning_active.md
# Repeat for all 8 files
```

## Notes

- This task is purely structural - no code, no tests, just directories and file copies
- Templates are markdown files with embedded template syntax for dynamic content
- The templates will be embedded in Task 4 using `//go:embed` directives
- This task has no dependencies and can be completed first

## Standards

- Use `cp` or equivalent to ensure exact file copies
- Preserve file permissions and timestamps where possible
- Verify copies are identical (no accidental edits)
