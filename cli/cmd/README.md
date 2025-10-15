# sow CLI Commands

This package contains all CLI commands for the sow tool, organized by domain.

## Structure

```
cli/cmd/
├── context.go        - Context adapter utilities (filesystem, etc.)
├── root.go           - Root command with adapter initialization
├── init.go           - Init command
├── validate.go       - Validate command
├── schema.go         - Schema commands (list, show, export)
├── log.go            - Log command
├── session_info.go   - Session info command
├── refs.go           - Refs commands (add, init, status, update, list, remove)
└── cache.go          - Cache commands (status, prune, clear)
```

## Command Hierarchy

```
sow
├── init                  - Initialize .sow/ structure
├── validate [file...]    - Validate files against schemas
├── schema
│   ├── list             - List all schemas
│   ├── show <schema>    - Display specific schema
│   └── export <schema>  - Export schema to file
├── log                  - Create log entry
├── session-info         - Display session context
├── refs
│   ├── add <source>     - Add new reference
│   ├── init             - Initialize refs structure
│   ├── status [ref-id]  - Show refs status
│   ├── update [ref-id]  - Update refs from remote
│   ├── list             - List all refs
│   └── remove <ref-id>  - Remove reference
└── cache
    ├── status           - Show cache status
    ├── prune            - Remove unused cache
    └── clear            - Clear entire cache
```

## Adapter Pattern

Commands follow the adapter pattern for dependency injection:

### 1. Root Command Initializes Adapters

```go
// root.go
func NewRootCmd() *cobra.Command {
    cmd := &cobra.Command{
        PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
            // Initialize adapters
            fs := billy.NewLocal()

            // Add to context
            ctx := WithFilesystem(cmd.Context(), fs)
            cmd.SetContext(ctx)

            return nil
        },
    }
    // ...
}
```

### 2. Commands Retrieve Adapters from Context

```go
// init.go
func NewInitCmd() *cobra.Command {
    cmd := &cobra.Command{
        RunE: func(cmd *cobra.Command, args []string) error {
            // Get filesystem from context
            fs := FilesystemFromContext(cmd.Context())

            // Use filesystem adapter
            fs.MkdirAll(".sow/knowledge", 0755)

            return nil
        },
    }
    return cmd
}
```

### 3. Context Utilities

```go
// context.go

// Add adapter to context
func WithFilesystem(ctx context.Context, fs core.FS) context.Context

// Retrieve adapter from context
func FilesystemFromContext(ctx context.Context) core.FS
```

## Benefits of Adapter Pattern

1. **Testability**: Commands can be tested with memory filesystem
2. **Flexibility**: Easy to swap implementations (local, memory, remote)
3. **Separation of Concerns**: Commands don't know about filesystem implementation
4. **Dependency Injection**: Dependencies injected at runtime, not compile-time
5. **Consistent Interface**: All commands use same pattern

## Testing Pattern

```go
func TestInitCommand(t *testing.T) {
    // Create test filesystem
    fs := billy.NewMemory()

    // Create context with test filesystem
    ctx := WithFilesystem(context.Background(), fs)

    // Create command
    cmd := NewInitCmd()
    cmd.SetContext(ctx)

    // Execute
    err := cmd.Execute()

    // Assert
    assert.NoError(t, err)
    exists, _ := fs.Exists(".sow/knowledge")
    assert.True(t, exists)
}
```

## Implementation Status

All commands are **templated** with:
- ✅ Proper help text
- ✅ Flag definitions
- ✅ Subcommand structure
- ✅ Adapter pattern setup
- ⏳ Implementation (TODO)

Next step: Implement command logic using:
- CUE schemas for validation
- Filesystem adapter for I/O
- Git library for refs management
- State management with generated types

## Adding New Commands

1. Create new file: `cli/cmd/mycommand.go`
2. Define command with help text and flags
3. Retrieve adapters from context
4. Register in `root.go`: `cmd.AddCommand(NewMyCmd())`

Example:

```go
package cmd

import "github.com/spf13/cobra"

func NewMyCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "my",
        Short: "My command description",
        RunE: func(cmd *cobra.Command, args []string) error {
            fs := FilesystemFromContext(cmd.Context())
            // Implementation...
            return nil
        },
    }

    cmd.Flags().String("option", "", "Option description")
    return cmd
}
```
