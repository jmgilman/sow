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

### Work with Refs Indexes

```go
import "github.com/jmgilman/sow/libs/schemas"

// Committed refs for team-shared references
committed := schemas.RefsCommittedIndex{
    Version: "1.0.0",
    Refs: []schemas.Ref{{
        Id:       "abc123",
        Source:   "git+https://github.com/org/repo",
        Semantic: "code",
        Link:     "my-ref",
    }},
}

// Cache index for local cache metadata
cache := schemas.RefsCacheIndex{
    Version: "1.0.0",
    Refs:    []schemas.CachedRef{},
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
fs := schemas.CUESchemas
// See CUE documentation for validation examples
```

## Links

- [CUE Language](https://cuelang.org/)
- [CUE Go Integration](https://cuelang.org/docs/integrations/go/)
