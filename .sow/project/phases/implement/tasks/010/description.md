# Task: Set up Go project structure with CUE embedding

## Objective

Initialize the Go project with proper module structure, dependencies, and CUE schema embedding setup.

## Context

Starting from scratch. Need to create go.mod, set up directory structure, add dependencies, and configure CUE schema embedding.

## Requirements

1. **Initialize Go module**:
   - Create `go.mod` with appropriate module path
   - Set Go version (1.21+)

2. **Create directory structure**:
   - `cmd/sow/` - Main CLI entry point
   - `internal/` - Internal packages (commands, validation, etc.)
   - `pkg/` - Public packages if needed
   - Follow design from task design/020

3. **Add dependencies**:
   - CUE Go libraries (`cuelang.org/go/cue`)
   - CLI framework (cobra or similar)
   - Any other required dependencies

4. **Set up CUE schema embedding**:
   - Use `go:embed` directive to embed `schemas/cue/`
   - Create loader for embedded schemas
   - Verify schemas are accessible at runtime

5. **Create initial main.go**:
   - Basic CLI setup
   - Version command
   - Help text

6. **Verify build**:
   - `go build` succeeds
   - Binary runs and shows help/version

## References

- `.sow/knowledge/adrs/` - CLI architecture decisions from design phase
- `schemas/cue/` - CUE schemas created in design/010

## Deliverables

- [ ] go.mod initialized
- [ ] Directory structure created
- [ ] Dependencies added
- [ ] CUE schemas embedded
- [ ] main.go with basic CLI setup
- [ ] Successful build verification
