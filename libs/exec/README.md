# sow Exec

Command execution abstraction providing a clean, testable interface for running external commands.

## Quick Start

```go
import "github.com/jmgilman/sow/libs/exec"

// Create an executor for the gh CLI
gh := exec.NewLocal("gh")

// Check if command exists
if !gh.Exists() {
    return fmt.Errorf("gh CLI not found")
}

// Run a command
stdout, stderr, err := gh.Run("issue", "list", "--label", "bug")
if err != nil {
    return fmt.Errorf("gh issue list failed: %s", stderr)
}
```

## Usage

### Create an Executor

```go
// Create executors for different commands
gh := exec.NewLocal("gh")
git := exec.NewLocal("git")
kubectl := exec.NewLocal("kubectl")
```

### Run Commands

```go
// Run with output capture
stdout, stderr, err := gh.Run("pr", "list", "--state", "open")

// Run without output (only check success/failure)
err := git.RunSilent("fetch", "--all")
```

### Use Context for Timeouts

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

stdout, stderr, err := gh.RunContext(ctx, "pr", "create", "--title", "My PR")
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        return fmt.Errorf("command timed out")
    }
    return fmt.Errorf("command failed: %w", err)
}
```

### Check Command Existence

```go
gh := exec.NewLocal("gh")
if !gh.Exists() {
    return fmt.Errorf("gh CLI is required but not installed")
}
```

## Testing

Use mocks to test code that depends on the Executor interface.

### Using Generated Mocks

```go
import (
    "testing"
    "go.uber.org/mock/gomock"
    "github.com/jmgilman/sow/libs/exec"
)

func TestMyService(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mock := exec.NewMockExecutor(ctrl)
    mock.EXPECT().Exists().Return(true)
    mock.EXPECT().Run("issue", "list").Return(`[{"number": 1}]`, "", nil)

    service := NewMyService(mock)
    // ... test service
}
```

### Using Custom Mocks

```go
type mockExecutor struct {
    command   string
    exists    bool
    runOutput string
    runErr    error
}

func (m *mockExecutor) Command() string { return m.command }
func (m *mockExecutor) Exists() bool    { return m.exists }
func (m *mockExecutor) Run(args ...string) (string, string, error) {
    return m.runOutput, "", m.runErr
}
// ... implement remaining methods

func TestWithCustomMock(t *testing.T) {
    mock := &mockExecutor{
        command:   "gh",
        exists:    true,
        runOutput: `{"number": 123}`,
    }
    service := NewMyService(mock)
    // ... test service
}
```

## Links

- [Go Package Documentation](https://pkg.go.dev/github.com/jmgilman/sow/libs/exec)
