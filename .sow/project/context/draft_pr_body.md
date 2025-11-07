# CLI Enhanced Advance Command and Standard Project Refactoring

## Intent

This PR enhances the `sow advance` command to support explicit event selection, discovery (`--list`), and dry-run validation (`--dry-run`), enabling orchestrators to handle intent-based branching scenarios. It also refactors the standard project's ReviewActive state to use the new AddBranch API, replacing the workaround pattern with declarative branching.

## Status

🚧 **Draft** - Implementation in progress

## Progress

- [x] Planning phase completed (8 tasks defined)
- [ ] Implementation phase
  - [ ] CLI command infrastructure
  - [ ] Auto-determination mode
  - [ ] List mode
  - [ ] Dry-run mode
  - [ ] Explicit event mode
  - [ ] Standard project descriptions
  - [ ] ReviewActive refactoring
  - [ ] Integration testing
- [ ] Review phase
- [ ] Final checks

## Components

1. **CLI Enhancement**: Four operation modes for `sow advance`
   - Auto-determination (backward compatible)
   - Discovery mode (`--list`)
   - Dry-run validation (`--dry-run [event]`)
   - Explicit event execution (`sow advance [event]`)

2. **Standard Project Refactoring**: Demonstrates proper branching patterns
   - Add descriptions to all transitions
   - Refactor ReviewActive with AddBranch API
   - Reference implementation for future project types

---

_This PR body will be updated with full implementation details and test results before marking ready for review._

🤖 Generated with [sow](https://github.com/jmgilman/sow)
