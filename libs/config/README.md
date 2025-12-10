# sow Config

Configuration loading for sow repositories and user settings.

## Quick Start

```go
import (
    "github.com/jmgilman/go/fs/billy"
    "github.com/jmgilman/sow/libs/config"
)

// Create a local filesystem
fs := billy.NewLocal()

// Load repository configuration
cfg, err := config.LoadRepoConfig(fs)
if err != nil {
    return err
}
```

## Usage

### Load Repo Config from Filesystem

```go
// Using a filesystem interface for testability
fs := billy.NewLocal()
cfg, err := config.LoadRepoConfig(fs)
if err != nil {
    return fmt.Errorf("load config: %w", err)
}
```

### Load Repo Config from Bytes

```go
// More flexible - works with any data source
data := []byte(`
artifacts:
  adrs: custom-adrs
  design_docs: docs/design
`)
cfg, err := config.LoadRepoConfigFromBytes(data)
```

### Load User Config

```go
// Create a local filesystem
fs := billy.NewLocal()

// Load from standard location (~/.config/sow/config.yaml)
userCfg, err := config.LoadUserConfig(fs)
if err != nil {
    return fmt.Errorf("load user config: %w", err)
}

// Load from a specific path
userCfg, err := config.LoadUserConfigFromPath(fs, "/path/to/config.yaml")

// Get the config file path
path, err := config.GetUserConfigPath()
```

### Testing with In-Memory Filesystem

```go
// Use billy.NewMemory() for tests
fs := billy.NewMemory()

// Set up test config
_ = fs.WriteFile("config.yaml", []byte(`
artifacts:
  adrs: test-adrs
`), 0644)

// Load config in test
cfg, err := config.LoadRepoConfig(fs)
```

### Get Path Helpers

```go
// Get absolute paths to artifact directories
adrsPath := config.GetADRsPath(repoRoot, cfg)
designPath := config.GetDesignDocsPath(repoRoot, cfg)
explorationsPath := config.GetExplorationsPath(repoRoot)
```

## Configuration

### Environment Variables

| Variable | Description |
|----------|-------------|
| `XDG_CONFIG_HOME` | Override default config directory location |

### Config File Locations

| Type | Location |
|------|----------|
| Repository | `.sow/config.yaml` (relative to repo root) |
| User (Linux/Mac) | `~/.config/sow/config.yaml` |
| User (Windows) | `%APPDATA%\sow\config.yaml` |

## Links

- [Go Package Documentation](https://pkg.go.dev/github.com/jmgilman/sow/libs/config)
