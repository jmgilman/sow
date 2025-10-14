# Task: Design CLI architecture

## Objective

Design the Go CLI architecture including project structure, command routing, CUE schema embedding, and module organization.

## Context

Building a CLI from scratch that will embed CUE schemas and provide multiple commands (init, validate, schema, log, session-info, sinks, repos). Need clear architecture before implementation.

## Requirements

1. **Design Go project structure**:
   - Directory layout (cmd/, internal/, pkg/)
   - Module organization
   - Where CUE schemas will be embedded
   - Configuration management

2. **Design command routing**:
   - Command structure (cobra or similar)
   - Subcommand organization
   - Flag handling
   - Auto-detection for task vs project context

3. **Design schema embedding strategy**:
   - Using `go:embed` directive
   - Schema loading mechanism
   - Version alignment (CLI version = schema version)

4. **Design core abstractions**:
   - Validation engine interface
   - File system operations
   - Context detection
   - Error handling

5. **Create architecture decision record (ADR)**:
   - Document key architectural decisions
   - Place in `.sow/knowledge/adrs/`
   - Follow ADR template

## References

- `docs/CLI_REFERENCE.md` - Complete CLI command specifications
- `docs/ARCHITECTURE.md` - System architecture
- `ROADMAP.md` - Milestone 1 requirements

## Deliverables

- [ ] ADR documenting CLI architecture
- [ ] Project structure diagram or description
- [ ] Command routing design
- [ ] Schema embedding strategy
- [ ] Core abstractions defined
