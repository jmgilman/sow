# Add README.md and doc.go Documentation

## Context

This task is part of the `libs/schemas` module migration project. After the module is created and consumers are updated (Tasks 010-030), this task adds proper documentation following the project's standards.

The project has specific documentation standards defined in `.standards/READMES.md` that must be followed. Key requirements:
- README provides overview, quick start, and usage examples
- README links to deeper docs rather than including everything
- `doc.go` provides package-level GoDoc documentation with `go:generate` directive

## Requirements

### 1. Create libs/schemas/README.md

Create a README following the READMES.md standard structure:

1. **Overview** (1-3 sentences: what/why/for whom)
2. **Quick Start** (copy-paste import and basic usage)
3. **Usage** (2-4 common tasks)
4. **Configuration** (code generation setup)
5. **Links** (to CUE documentation, related packages)

### 2. Create libs/schemas/doc.go

Create package documentation with:
- Package overview
- `go:generate` directive for CUE code generation
- Links to README for detailed documentation

### 3. Verify Documentation Quality

- All exported types should have GoDoc comments (generated types have them)
- README should be concise and actionable
- Links should be accurate

## Acceptance Criteria

1. [ ] `libs/schemas/README.md` exists and follows READMES.md structure
2. [ ] `libs/schemas/doc.go` exists with package comment and generate directive
3. [ ] README includes:
   - Clear overview of what the package provides
   - Quick start with import path and example
   - Usage examples for common operations
   - Code generation instructions
4. [ ] doc.go includes:
   - Package-level documentation
   - `//go:generate` directive for CUE types
   - Generation strategy explanation
5. [ ] `go doc github.com/jmgilman/sow/libs/schemas` shows correct package docs
6. [ ] `go doc github.com/jmgilman/sow/libs/schemas/project` shows correct subpackage docs

## Technical Details

### README Structure

The README should follow the project standard (from .standards/READMES.md):
- Lead with value in first line
- Provide 60-second proof (quick start)
- Use active voice and imperative verbs
- Keep runnable commands with prerequisites stated

### doc.go Structure

```go
//go:generate sh -c "cue exp gengotypes ./... && rm -f project/cue_types_gen.go || true"

// Package schemas provides CUE schema definitions and generated Go types
// for all sow state and configuration files.
//
// This package contains schemas for:
//   - Configuration: Config, UserConfig
//   - Refs indexes: RefsCommittedIndex, RefsLocalIndex, RefsCacheIndex
//   - Knowledge: KnowledgeIndex
//
// Project-specific schemas are in the project subpackage.
//
// Generation:
//   Run `go generate ./...` to regenerate Go types from CUE schemas.
//
// See README.md for detailed usage examples.
package schemas
```

### generate Directive Note

The rm command in the generate directive handles edge cases where CUE might generate files we don't want. The `|| true` prevents errors if the file doesn't exist.

## Relevant Inputs

- `.standards/READMES.md` - README structure and style requirements
- `.standards/STYLE.md` - Go documentation standards (section: Comments & Documentation)
- `cli/schemas/README.md` - Existing README to reference for content (but restructure to match standard)
- `cli/schemas/doc.go` - Existing doc.go to reference for generate directive

## Examples

### README.md Content

```markdown
# sow Schemas

CUE schema definitions and generated Go types for sow configuration and state files.

## Quick Start

```go
import (
    "github.com/jmgilman/sow/libs/schemas"
    "github.com/jmgilman/sow/libs/schemas/project"
)

// Use configuration types
var cfg schemas.Config

// Use project state types
var state project.ProjectState
```

## Usage

### Load Configuration

```go
import "github.com/jmgilman/sow/libs/schemas"

// Config for .sow/config.yaml
cfg := schemas.Config{
    Artifacts: &struct{
        Adrs        *string `json:"adrs,omitempty"`
        Design_docs *string `json:"design_docs,omitempty"`
    }{
        Adrs: ptr(".sow/knowledge/adrs"),
    },
}
```

### Create Project State

```go
import (
    "time"
    "github.com/jmgilman/sow/libs/schemas/project"
)

state := project.ProjectState{
    Name:       "my-feature",
    Type:       "standard",
    Branch:     "feat/my-feature",
    Created_at: time.Now(),
    Updated_at: time.Now(),
    Phases:     map[string]project.PhaseState{},
    Statechart: project.StatechartState{
        Current_state: "NoProject",
        Updated_at:    time.Now(),
    },
}
```

## Types

### Root Package (`schemas`)
- `Config` - Repository configuration (.sow/config.yaml)
- `UserConfig` - User configuration (~/.config/sow/config.yaml)
- `KnowledgeIndex` - Knowledge tracking (.sow/knowledge/index.yaml)
- `RefsCommittedIndex` - Shared refs (.sow/refs/index.json)
- `RefsLocalIndex` - Local refs (.sow/refs/index.local.json)
- `RefsCacheIndex` - Cache metadata (~/.cache/sow/index.json)

### Project Package (`schemas/project`)
- `ProjectState` - Complete project state
- `PhaseState` - Phase state within project
- `TaskState` - Task state within phase
- `ArtifactState` - Artifact metadata

## Code Generation

Regenerate Go types from CUE schemas:

```bash
cd libs/schemas
go generate ./...
```

This runs `cue exp gengotypes ./...` to regenerate `cue_types_gen.go` files.

## Validation

Use the embedded CUE schemas for runtime validation:

```go
import (
    "cuelang.org/go/cue/cuecontext"
    "github.com/jmgilman/sow/libs/schemas"
)

// Load embedded schemas
ctx := cuecontext.New()
// See CUE documentation for validation examples
```

## Links

- [CUE Language](https://cuelang.org/)
- [CUE Go Integration](https://cuelang.org/docs/integrations/go/)
```

### doc.go Content

```go
//go:generate sh -c "cue exp gengotypes ./..."

// Package schemas provides CUE schema definitions and generated Go types
// for all sow configuration, index, and project state files.
//
// This package is the foundation schema layer with no internal dependencies.
// All sow state files use these schemas for validation and type safety.
//
// # Schema Types
//
// Root package types:
//   - [Config]: Repository configuration at .sow/config.yaml
//   - [UserConfig]: User configuration at ~/.config/sow/config.yaml
//   - [KnowledgeIndex]: Knowledge tracking at .sow/knowledge/index.yaml
//   - [RefsCommittedIndex]: Team-shared refs at .sow/refs/index.json
//   - [RefsLocalIndex]: Local-only refs at .sow/refs/index.local.json
//   - [RefsCacheIndex]: Cache metadata at ~/.cache/sow/index.json
//
// The [project] subpackage contains project lifecycle types:
//   - [project.ProjectState]: Complete project state
//   - [project.PhaseState]: Phase state within project
//   - [project.TaskState]: Task state within phase
//   - [project.ArtifactState]: Artifact metadata
//
// # Code Generation
//
// Go types are generated from CUE schemas using:
//
//	go generate ./...
//
// This runs cue exp gengotypes to regenerate cue_types_gen.go files.
// Generated types use json tags matching CUE field names.
//
// # CUE Schemas
//
// CUE schema files are embedded via [CUESchemas] for runtime access.
// Use cuelang.org/go/cue for schema validation.
//
// See README.md for usage examples.
package schemas
```

## Dependencies

- Task 010: Create libs/schemas Go module structure
- Task 020: Migrate project schemas to libs/schemas/project
- Task 030: Update consumer import paths

All must be completed first (documentation should reflect working code).

## Constraints

- README must follow `.standards/READMES.md` structure exactly
- doc.go must include `go:generate` directive
- Do NOT include full API reference in README (link to godoc instead)
- Do NOT use emojis in documentation
- Use active voice and imperative verbs
- Code examples must be syntactically correct and runnable
