# CUE Library Reference

**Package**: `github.com/jmgilman/go/cue`

**Purpose**: CUE evaluation and validation capabilities with platform-specific error handling and filesystem abstraction.

---

## Overview

This library wraps `cuelang.org/go` functionality, providing:
- Schema loading from filesystem
- Validation of data against schemas
- Encoding to YAML/JSON
- Decoding to Go structs
- Extensible attribute processing

The library is **stateless and generic** - it works with any CUE schemas and Go struct types without coupling to specific schema packages.

---

## Core Components

### 1. Loader

Manages CUE loading operations from a filesystem.

```go
type Loader struct {
    fs     core.ReadFS
    cueCtx *cue.Context
}

func NewLoader(filesystem core.ReadFS) *Loader
```

**Methods**:
- `LoadFile(ctx, filePath string) (cue.Value, error)` - Load single CUE file
- `LoadPackage(ctx, packagePath string) (cue.Value, error)` - Load all .cue files in directory (non-recursive)
- `LoadModule(ctx, modulePath string) (cue.Value, error)` - Load CUE module with package structure (recursive)
- `LoadBytes(ctx, source []byte, filename string) (cue.Value, error)` - Load from memory
- `Context() *cue.Context` - Access underlying CUE context

**Usage Pattern**:
```go
loader := cue.NewLoader(filesystem)
schema, err := loader.LoadFile(ctx, "schemas/project-state.cue")
```

### 2. Validation

Validates CUE values against schemas.

```go
type ValidationOptions struct {
    Concrete bool  // Require all values fully specified
    Final    bool  // Resolve default values first
    All      bool  // Report all errors (not just first)
}

type ValidationIssue struct {
    Path     []string    // Field path where error occurred
    Message  string      // Human-readable error message
    Position token.Pos   // Source position if available
}
```

**Functions**:
- `Validate(ctx, schema, data cue.Value) error` - Validate with default options
- `ValidateWithOptions(ctx, schema, data cue.Value, opts ValidationOptions) error` - Custom validation
- `ValidateConstraint(ctx, value, constraint cue.Value) error` - Validate individual constraints

**Usage Pattern**:
```go
// Load schema and data
schema, _ := schemaLoader.LoadFile(ctx, "schema.cue")
data, _ := dataLoader.LoadFile(ctx, "config.cue")

// Validate
if err := cue.Validate(ctx, schema, data); err != nil {
    // Detailed validation errors with field paths
    return err
}
```

### 3. Encoding

Encodes CUE values to YAML/JSON.

**Functions**:
- `EncodeYAML(ctx, value cue.Value) ([]byte, error)` - Encode to YAML
- `EncodeJSON(ctx, value cue.Value) ([]byte, error)` - Encode to JSON
- `EncodeYAMLStream(ctx, value cue.Value, w io.Writer) error` - Stream large YAML (>10MB)

**Usage Pattern**:
```go
yaml, err := cue.EncodeYAML(ctx, value)
if err != nil {
    return err
}
```

### 4. Decoding

Decodes CUE values to Go structs.

**Function**:
- `Decode(ctx, value cue.Value, target interface{}) error`

**Usage Pattern**:
```go
type Config struct {
    Name string `json:"name"`
    Port int    `json:"port"`
}

var config Config
if err := cue.Decode(ctx, value, &config); err != nil {
    return err
}
```

### 5. Attribute Processing (Advanced)

Extensible attribute processing for runtime value substitution.

**Package**: `github.com/jmgilman/go/cue/attributes`

```go
type Processor interface {
    Name() string
    Process(ctx context.Context, attr Attribute) (cue.Value, error)
}

type Registry struct { ... }
func NewRegistry() *Registry
func (r *Registry) Register(p Processor) error

type Walker struct { ... }
func NewWalker(registry *Registry, cueCtx *cue.Context) *Walker
func (w *Walker) Walk(ctx context.Context, value cue.Value) (cue.Value, error)
```

---

## Application in sow CLI

### Use Case 1: Schema Embedding and Loading

**File**: `internal/schema/embed.go`

```go
package schema

import (
    _ "embed"
    "github.com/jmgilman/go/cue"
    "github.com/jmgilman/go/fs/billy"
)

//go:embed ../../schemas/cue/project-state.cue
var projectStateCUE []byte

//go:embed ../../schemas/cue/task-state.cue
var taskStateCUE []byte

var (
    loader        *cue.Loader
    projectSchema cue.Value
    taskSchema    cue.Value
)

func init() {
    // Create in-memory filesystem with embedded schemas
    memFS := billy.NewMemory()

    // Write embedded schemas to memory filesystem
    memFS.WriteFile("project-state.cue", projectStateCUE, 0644)
    memFS.WriteFile("task-state.cue", taskStateCUE, 0644)

    // Create loader
    loader = cue.NewLoader(memFS)

    // Load schemas at startup
    ctx := context.Background()
    projectSchema, _ = loader.LoadFile(ctx, "project-state.cue")
    taskSchema, _ = loader.LoadFile(ctx, "task-state.cue")
}

func GetProjectSchema() cue.Value {
    return projectSchema
}

func GetTaskSchema() cue.Value {
    return taskSchema
}
```

### Use Case 2: Validation

**File**: `internal/schema/validate.go`

```go
package schema

import (
    "context"
    "github.com/jmgilman/go/cue"
    "github.com/jmgilman/go/fs/core"
)

func ValidateProjectState(fs core.ReadFS, filePath string) error {
    ctx := context.Background()

    // Load data file from actual filesystem
    dataLoader := cue.NewLoader(fs)
    data, err := dataLoader.LoadFile(ctx, filePath)
    if err != nil {
        return err
    }

    // Validate against embedded schema
    return cue.Validate(ctx, GetProjectSchema(), data)
}

func ValidateTaskState(fs core.ReadFS, filePath string) error {
    ctx := context.Background()

    dataLoader := cue.NewLoader(fs)
    data, err := dataLoader.LoadFile(ctx, filePath)
    if err != nil {
        return err
    }

    return cue.Validate(ctx, GetTaskSchema(), data)
}
```

### Use Case 3: Decode to Go Structs

**File**: `internal/schema/decode.go`

```go
package schema

import (
    "context"
    "github.com/jmgilman/go/cue"
    "github.com/jmgilman/go/fs/core"
)

type ProjectState struct {
    Project struct {
        Name        string `json:"name"`
        Branch      string `json:"branch"`
        Description string `json:"description"`
    } `json:"project"`
    Phases map[string]Phase `json:"phases"`
}

type Phase struct {
    Enabled   bool   `json:"enabled"`
    Status    string `json:"status"`
    CreatedAt string `json:"created_at"`
}

func LoadProjectState(fs core.ReadFS, filePath string) (*ProjectState, error) {
    ctx := context.Background()

    // Load and validate
    dataLoader := cue.NewLoader(fs)
    data, err := dataLoader.LoadFile(ctx, filePath)
    if err != nil {
        return nil, err
    }

    if err := cue.Validate(ctx, GetProjectSchema(), data); err != nil {
        return nil, err
    }

    // Decode to struct
    var state ProjectState
    if err := cue.Decode(ctx, data, &state); err != nil {
        return nil, err
    }

    return &state, nil
}
```

### Use Case 4: Schema Inspection (for `sow schema` commands)

**File**: `internal/commands/schema.go`

```go
package commands

import (
    "context"
    "fmt"
    "github.com/jmgilman/go/cue"
    "github.com/spf13/cobra"
)

func NewSchemaCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "schema",
        Short: "View and validate CUE schemas",
    }

    cmd.AddCommand(&cobra.Command{
        Use:   "show <type>",
        Short: "Display specific schema",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            schemaType := args[0]

            var schema cue.Value
            switch schemaType {
            case "project":
                schema = GetProjectSchema()
            case "task":
                schema = GetTaskSchema()
            default:
                return fmt.Errorf("unknown schema type: %s", schemaType)
            }

            // Encode schema to YAML for display
            ctx := context.Background()
            yaml, err := cue.EncodeYAML(ctx, schema)
            if err != nil {
                return err
            }

            fmt.Println(string(yaml))
            return nil
        },
    })

    return cmd
}
```

---

## Key Benefits for sow CLI

1. **Zero External Dependencies**: Schemas embedded in binary via `go:embed` + in-memory FS
2. **Strong Validation**: Detailed field-level error messages with paths
3. **Type Safety**: Go structs validated against schemas at runtime
4. **Performance**: Schemas loaded once at init time (~100ms validation)
5. **Flexibility**: Can validate files from any filesystem (local, memory, test fixtures)

---

## Error Handling

All errors are wrapped with `errors.PlatformError`. Error codes:
- `CodeCUELoadFailed` - Module/file loading failures
- `CodeCUEBuildFailed` - Build/evaluation failures
- `CodeCUEValidationFailed` - Validation failures (includes field paths)
- `CodeCUEDecodeFailed` - Decoding failures
- `CodeCUEEncodeFailed` - Encoding failures

---

## Performance Considerations

- **Module loading** can be expensive - cache `cue.Value` results if needed
- **Validation** is typically fast (<100ms) but complex schemas may take longer
- **Large outputs**: Use `EncodeYAMLStream()` for manifests >10MB
- **Context**: Use `context.WithTimeout()` to set time limits

---

## Related Documentation

- [cuelang.org/go documentation](https://pkg.go.dev/cuelang.org/go/cue)
- Platform errors: `github.com/jmgilman/go/errors`
- Filesystem abstraction: `github.com/jmgilman/go/fs/core`
