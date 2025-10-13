# Task 020: Create .sow/ template files and directories

**Phase**: implement
**Assigned Agent**: implementer
**Created**: 2025-10-12

## Objective

Create template structure for the `.sow/` directory (data layer). This structure
will be created by the `/init` command when users bootstrap sow in their repository.

Note: This task creates templates/documentation for what `/init` will create,
not the actual `.sow/` structure in user repositories (except for this project's
own `.sow/project/` which we're using for bootstrapping).

## Requirements

1. Document the `.sow/` directory structure that `/init` will create:
   - `.sow/.version` template
   - `.sow/knowledge/` structure and templates
   - `.sow/sinks/index.json` template
   - `.sow/repos/index.json` template

2. Create example/template files:
   - Example `.sow/.version` YAML
   - Example `.sow/knowledge/overview.md` template
   - Empty `.sow/sinks/index.json`
   - Empty `.sow/repos/index.json`

3. Create `.gitignore` entries template:
   - `.sow/sinks/` - git-ignored (per-developer)
   - `.sow/repos/` - git-ignored (external repos)
   - `.sow/project/` - NOT ignored (committed to feature branches)

4. Document in FILE_STRUCTURE.md or create separate init templates

## Context References

- `docs/FILE_STRUCTURE.md` - Complete data layer structure
- `docs/PROJECT_MANAGEMENT.md` - Project structure details
- Design task 020 output - Data layer design

## Acceptance Criteria

- [ ] `.sow/` directory structure clearly documented
- [ ] Template files created for `/init` to use
- [ ] `.gitignore` entries defined
- [ ] Git versioning strategy implemented
- [ ] Example files follow schema specifications
- [ ] Structure matches design from task 020

## Notes

This is about defining templates and structure, not creating actual user files.
The `/init` command (Milestone 2) will use these templates.

Exception: We're currently using `.sow/project/` for bootstrapping this project,
so that structure already exists and follows the design.

Focus on creating clear templates that `/init` can copy/generate.
