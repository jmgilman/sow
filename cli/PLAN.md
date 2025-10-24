# Design Mode Implementation Plan

## Overview

This document describes the implementation of **design mode** for the sow CLI. Design mode is a new operating mode (similar to exploration mode) focused on creating formal design documents from research findings and other inputs.

### Context

- **Exploration mode** (already implemented): Freeform research and discovery
- **Design mode** (this plan): Formalize findings into structured design documents
- **Project mode** (already implemented): Structured implementation with orchestrator + workers

Design mode bridges the gap between exploration and implementation, allowing teams to consolidate research into formal artifacts (ADRs, architecture docs, diagrams, specs) before building.

### Key Characteristics

1. **Branch naming**: Uses `design/*` branches (e.g., `design/auth-system`)
2. **Input tracking**: Maintains index of all input sources (explorations, files, references)
3. **Output tracking**: Tracks each design document with its own target location
4. **Single agent**: Orchestrator works directly with user (no worker agents)
5. **Flexible finalization**: Different documents can go to different locations

---

## Architecture Decisions

### 1. Index Structure

The design index tracks both **inputs** (what informs the design) and **outputs** (what will be produced):

**Rationale**:
- Inputs allow session resumability without reprompting
- Outputs with individual target locations support flexible organization
- Each output can go to a different final location (ADRs to `.sow/knowledge/adrs/`, diagrams to `docs/diagrams/`, etc.)

### 2. No Global Target Location

Unlike the initial proposal, there is **no global target location**. Each output has its own target.

**Rationale**:
- Design sessions typically produce heterogeneous artifacts
- Different document types belong in different locations
- More flexible and realistic for actual usage

### 3. No Automated Finalization Command

The orchestrator agent performs finalization manually (not via a `sow design finalize` command).

**Rationale**:
- Allows orchestrator to validate targets, check files, handle edge cases
- Provides flexibility for custom finalization workflows
- Keeps CLI focused on state management, not business logic

### 4. Glob Patterns and Directories for Inputs

Inputs support paths, glob patterns, and directories.

**Rationale**:
- Reduces manual work when importing multiple files
- Natural UX (e.g., `.sow/exploration/` imports entire directory)
- Matches user mental model

---

## Implementation Checklist

### Phase 1: Schema and Core Logic
- [ ] `schemas/design_index.cue` - CUE schema definition
- [ ] `internal/design/errors.go` - Error definitions
- [ ] `internal/design/index.go` - Index load/save/init operations
- [ ] `internal/design/manager.go` - Input/output management functions
- [ ] `internal/design/manager_test.go` - Unit tests

### Phase 2: CLI Commands
- [ ] `cmd/design.go` - Entry point command
- [ ] `cmd/design/design.go` - Subcommand parent
- [ ] `cmd/design/add_input.go` - Register input sources
- [ ] `cmd/design/remove_input.go` - Remove inputs
- [ ] `cmd/design/add_output.go` - Register output documents
- [ ] `cmd/design/remove_output.go` - Remove outputs
- [ ] `cmd/design/set_output_target.go` - Update output target location
- [ ] `cmd/design/set_status.go` - Update design status
- [ ] `cmd/design/index.go` - Display index

### Phase 3: Prompt System
- [ ] `internal/prompts/templates/modes/design.md` - Orchestrator prompt template
- [ ] `internal/prompts/context.go` - Add DesignContext types
- [ ] `internal/prompts/prompts.go` - Register design prompt

### Phase 4: Integration
- [ ] `cmd/root.go` - Wire into root command
- [ ] Manual testing and validation

**Total: 18 files (3 modified, 15 new)**

---

## Detailed File Specifications

### 1. Schema Definition

**File**: `cli/schemas/design_index.cue`

```cue
package schemas

import "time"

// DesignIndex defines the schema for design mode index files at:
// .sow/design/index.yaml
//
// This tracks input sources and planned outputs for active design work.
#DesignIndex: {
    // Design session metadata
    design: {
        // Topic being designed (human-readable)
        topic: string & !=""

        // Git branch name for this design session
        branch: string & =~"^design/[a-z0-9][a-z0-9-]*[a-z0-9]$"

        // When this design session was created
        created_at: time.Time

        // Design session status
        status: "active" | "in_review" | "completed"
    }

    // Input sources for this design session
    inputs: [...#DesignInput]

    // Output documents produced by this design session
    outputs: [...#DesignOutput]
}

// DesignInput represents an input source for the design process
#DesignInput: {
    // Input type
    type: "exploration" | "file" | "reference" | "url" | "git"

    // Path, glob pattern, directory, or identifier
    path: string & !=""

    // Brief description of what this input provides
    description: string & !=""

    // Optional tags for organization
    tags?: [...string]

    // When this input was added
    added_at: time.Time
}

// DesignOutput represents a design document to be produced
#DesignOutput: {
    // Path relative to .sow/design/
    path: string & !=""

    // Brief description of the document
    description: string & !=""

    // Target location for this specific document when finalized
    target_location: string & !=""

    // Document type (for organization)
    type?: "adr" | "architecture" | "diagram" | "spec" | "other"

    // Optional tags
    tags?: [...string]

    // When this output was added to the index
    added_at: time.Time
}
```

**Note**: After creating this file, regenerate Go types:
```bash
cd cli/schemas
go generate ./...
```

---

### 2. Error Definitions

**File**: `cli/internal/design/errors.go`

```go
package design

import "errors"

var (
    ErrNoDesign       = errors.New("no active design session")
    ErrDesignExists   = errors.New("design session already exists")
    ErrInputExists    = errors.New("input already exists")
    ErrInputNotFound  = errors.New("input not found")
    ErrOutputExists   = errors.New("output already exists")
    ErrOutputNotFound = errors.New("output not found")
)
```

---

### 3. Index Operations

**File**: `cli/internal/design/index.go`

```go
package design

import (
    "fmt"
    "path/filepath"
    "time"

    "github.com/jmgilman/sow/cli/internal/sow"
    "github.com/jmgilman/sow/cli/schemas"
    "gopkg.in/yaml.v3"
)

const (
    // IndexPath is the path to the design index relative to .sow/.
    IndexPath = "design/index.yaml"
)

// LoadIndex loads the design index from disk.
// Returns ErrNoDesign if design directory doesn't exist.
func LoadIndex(ctx *sow.Context) (*schemas.DesignIndex, error) {
    fs := ctx.FS()
    if fs == nil {
        return nil, sow.ErrNotInitialized
    }

    // Check if design directory exists
    exists, err := fs.Exists("design")
    if err != nil {
        return nil, fmt.Errorf("failed to check design directory: %w", err)
    }
    if !exists {
        return nil, ErrNoDesign
    }

    // Read index file
    data, err := fs.ReadFile(IndexPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read design index: %w", err)
    }

    // Parse YAML
    var index schemas.DesignIndex
    if err := yaml.Unmarshal(data, &index); err != nil {
        return nil, fmt.Errorf("failed to parse design index: %w", err)
    }

    return &index, nil
}

// SaveIndex saves the design index to disk.
func SaveIndex(ctx *sow.Context, index *schemas.DesignIndex) error {
    fs := ctx.FS()
    if fs == nil {
        return sow.ErrNotInitialized
    }

    // Marshal to YAML
    data, err := yaml.Marshal(index)
    if err != nil {
        return fmt.Errorf("failed to marshal design index: %w", err)
    }

    // Write atomically (write to temp file, then rename)
    tmpPath := IndexPath + ".tmp"
    if err := fs.WriteFile(tmpPath, data, 0644); err != nil {
        return fmt.Errorf("failed to write temp index: %w", err)
    }

    if err := fs.Rename(tmpPath, IndexPath); err != nil {
        _ = fs.Remove(tmpPath) // Clean up temp file
        return fmt.Errorf("failed to rename temp index: %w", err)
    }

    return nil
}

// InitDesign creates a new design directory and index.
func InitDesign(ctx *sow.Context, topic, branch string) error {
    fs := ctx.FS()
    if fs == nil {
        return sow.ErrNotInitialized
    }

    // Check if design already exists
    exists, _ := fs.Exists("design")
    if exists {
        return ErrDesignExists
    }

    // Create design directory
    if err := fs.MkdirAll("design", 0755); err != nil {
        return fmt.Errorf("failed to create design directory: %w", err)
    }

    // Create initial index
    index := &schemas.DesignIndex{
        Design: struct {
            Topic      string    `json:"topic"`
            Branch     string    `json:"branch"`
            Created_at time.Time `json:"created_at"`
            Status     string    `json:"status"`
        }{
            Topic:      topic,
            Branch:     branch,
            Created_at: time.Now(),
            Status:     "active",
        },
        Inputs:  []schemas.DesignInput{},
        Outputs: []schemas.DesignOutput{},
    }

    if err := SaveIndex(ctx, index); err != nil {
        // Clean up on failure
        _ = fs.RemoveAll("design")
        return fmt.Errorf("failed to save initial index: %w", err)
    }

    return nil
}

// Exists checks if a design directory exists.
func Exists(ctx *sow.Context) bool {
    fs := ctx.FS()
    if fs == nil {
        return false
    }
    exists, _ := fs.Exists("design")
    return exists
}

// Delete removes the design directory and all its contents.
func Delete(ctx *sow.Context) error {
    fs := ctx.FS()
    if fs == nil {
        return sow.ErrNotInitialized
    }

    exists, _ := fs.Exists("design")
    if !exists {
        return ErrNoDesign
    }

    if err := fs.RemoveAll("design"); err != nil {
        return fmt.Errorf("failed to remove design directory: %w", err)
    }

    return nil
}

// GetFilePath returns the absolute path to a file in the design directory.
func GetFilePath(ctx *sow.Context, relativePath string) string {
    return filepath.Join(ctx.RepoRoot(), ".sow", "design", relativePath)
}
```

---

### 4. Management Functions

**File**: `cli/internal/design/manager.go`

```go
package design

import (
    "fmt"
    "time"

    "github.com/jmgilman/sow/cli/internal/sow"
    "github.com/jmgilman/sow/cli/schemas"
)

// AddInput adds an input to the design index.
func AddInput(ctx *sow.Context, inputType, path, description string, tags []string) error {
    // Load current index
    index, err := LoadIndex(ctx)
    if err != nil {
        return err
    }

    // Check if input already exists
    for _, input := range index.Inputs {
        if input.Path == path {
            return ErrInputExists
        }
    }

    // Add input
    input := schemas.DesignInput{
        Type:        inputType,
        Path:        path,
        Description: description,
        Tags:        tags,
        Added_at:    time.Now(),
    }
    index.Inputs = append(index.Inputs, input)

    // Save index
    if err := SaveIndex(ctx, index); err != nil {
        return err
    }

    return nil
}

// RemoveInput removes an input from the design index.
func RemoveInput(ctx *sow.Context, path string) error {
    // Load current index
    index, err := LoadIndex(ctx)
    if err != nil {
        return err
    }

    // Find and remove input
    found := false
    newInputs := make([]schemas.DesignInput, 0, len(index.Inputs))
    for _, input := range index.Inputs {
        if input.Path == path {
            found = true
            continue
        }
        newInputs = append(newInputs, input)
    }

    if !found {
        return ErrInputNotFound
    }

    index.Inputs = newInputs

    // Save index
    if err := SaveIndex(ctx, index); err != nil {
        return err
    }

    return nil
}

// GetInput retrieves an input from the design index.
func GetInput(ctx *sow.Context, path string) (*schemas.DesignInput, error) {
    // Load current index
    index, err := LoadIndex(ctx)
    if err != nil {
        return nil, err
    }

    // Find input
    for _, input := range index.Inputs {
        if input.Path == path {
            return &input, nil
        }
    }

    return nil, ErrInputNotFound
}

// ListInputs returns all inputs in the design index.
func ListInputs(ctx *sow.Context) ([]schemas.DesignInput, error) {
    // Load current index
    index, err := LoadIndex(ctx)
    if err != nil {
        return nil, err
    }

    return index.Inputs, nil
}

// AddOutput adds an output to the design index.
func AddOutput(ctx *sow.Context, path, description, targetLocation, docType string, tags []string) error {
    // Load current index
    index, err := LoadIndex(ctx)
    if err != nil {
        return err
    }

    // Check if output already exists
    for _, output := range index.Outputs {
        if output.Path == path {
            return ErrOutputExists
        }
    }

    // Add output
    output := schemas.DesignOutput{
        Path:            path,
        Description:     description,
        Target_location: targetLocation,
        Type:            docType,
        Tags:            tags,
        Added_at:        time.Now(),
    }
    index.Outputs = append(index.Outputs, output)

    // Save index
    if err := SaveIndex(ctx, index); err != nil {
        return err
    }

    return nil
}

// RemoveOutput removes an output from the design index.
func RemoveOutput(ctx *sow.Context, path string) error {
    // Load current index
    index, err := LoadIndex(ctx)
    if err != nil {
        return err
    }

    // Find and remove output
    found := false
    newOutputs := make([]schemas.DesignOutput, 0, len(index.Outputs))
    for _, output := range index.Outputs {
        if output.Path == path {
            found = true
            continue
        }
        newOutputs = append(newOutputs, output)
    }

    if !found {
        return ErrOutputNotFound
    }

    index.Outputs = newOutputs

    // Save index
    if err := SaveIndex(ctx, index); err != nil {
        return err
    }

    return nil
}

// UpdateOutputTarget updates an output's target location.
func UpdateOutputTarget(ctx *sow.Context, path, targetLocation string) error {
    // Load current index
    index, err := LoadIndex(ctx)
    if err != nil {
        return err
    }

    // Find and update output
    found := false
    for i, output := range index.Outputs {
        if output.Path == path {
            index.Outputs[i].Target_location = targetLocation
            found = true
            break
        }
    }

    if !found {
        return ErrOutputNotFound
    }

    // Save index
    if err := SaveIndex(ctx, index); err != nil {
        return err
    }

    return nil
}

// GetOutput retrieves an output from the design index.
func GetOutput(ctx *sow.Context, path string) (*schemas.DesignOutput, error) {
    // Load current index
    index, err := LoadIndex(ctx)
    if err != nil {
        return nil, err
    }

    // Find output
    for _, output := range index.Outputs {
        if output.Path == path {
            return &output, nil
        }
    }

    return nil, ErrOutputNotFound
}

// ListOutputs returns all outputs in the design index.
func ListOutputs(ctx *sow.Context) ([]schemas.DesignOutput, error) {
    // Load current index
    index, err := LoadIndex(ctx)
    if err != nil {
        return nil, err
    }

    return index.Outputs, nil
}

// UpdateStatus updates the design session status.
func UpdateStatus(ctx *sow.Context, status string) error {
    // Validate status
    validStatuses := map[string]bool{
        "active":     true,
        "in_review":  true,
        "completed":  true,
    }
    if !validStatuses[status] {
        return fmt.Errorf("invalid status: %s (must be active, in_review, or completed)", status)
    }

    // Load current index
    index, err := LoadIndex(ctx)
    if err != nil {
        return err
    }

    // Update status
    index.Design.Status = status

    // Save index
    if err := SaveIndex(ctx, index); err != nil {
        return err
    }

    return nil
}
```

---

### 5. Unit Tests

**File**: `cli/internal/design/manager_test.go`

```go
package design

import (
    "testing"

    "github.com/jmgilman/sow/cli/internal/sow"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// Test helper to create a temporary context
func createTestContext(t *testing.T) (*sow.Context, func()) {
    // Implementation similar to exploration/manager_test.go
    // Creates temp directory, initializes sow, returns context and cleanup function
    // TODO: Implement based on existing test patterns
    return nil, func() {}
}

func TestInitDesign(t *testing.T) {
    ctx, cleanup := createTestContext(t)
    defer cleanup()

    err := InitDesign(ctx, "test-topic", "design/test")
    require.NoError(t, err)

    // Verify design exists
    assert.True(t, Exists(ctx))

    // Verify index can be loaded
    index, err := LoadIndex(ctx)
    require.NoError(t, err)
    assert.Equal(t, "test-topic", index.Design.Topic)
    assert.Equal(t, "design/test", index.Design.Branch)
    assert.Equal(t, "active", index.Design.Status)
    assert.Empty(t, index.Inputs)
    assert.Empty(t, index.Outputs)
}

func TestAddInput(t *testing.T) {
    ctx, cleanup := createTestContext(t)
    defer cleanup()

    InitDesign(ctx, "test", "design/test")

    err := AddInput(ctx, "file", "test.md", "Test file", []string{"tag1"})
    require.NoError(t, err)

    // Verify input was added
    inputs, err := ListInputs(ctx)
    require.NoError(t, err)
    assert.Len(t, inputs, 1)
    assert.Equal(t, "file", inputs[0].Type)
    assert.Equal(t, "test.md", inputs[0].Path)

    // Test duplicate
    err = AddInput(ctx, "file", "test.md", "Duplicate", nil)
    assert.ErrorIs(t, err, ErrInputExists)
}

func TestAddOutput(t *testing.T) {
    ctx, cleanup := createTestContext(t)
    defer cleanup()

    InitDesign(ctx, "test", "design/test")

    err := AddOutput(ctx, "adr-001.md", "Test ADR", ".sow/knowledge/adrs/", "adr", []string{"tag1"})
    require.NoError(t, err)

    // Verify output was added
    outputs, err := ListOutputs(ctx)
    require.NoError(t, err)
    assert.Len(t, outputs, 1)
    assert.Equal(t, "adr-001.md", outputs[0].Path)
    assert.Equal(t, ".sow/knowledge/adrs/", outputs[0].Target_location)

    // Test duplicate
    err = AddOutput(ctx, "adr-001.md", "Duplicate", ".sow/", "adr", nil)
    assert.ErrorIs(t, err, ErrOutputExists)
}

func TestUpdateOutputTarget(t *testing.T) {
    ctx, cleanup := createTestContext(t)
    defer cleanup()

    InitDesign(ctx, "test", "design/test")
    AddOutput(ctx, "adr-001.md", "Test ADR", ".sow/knowledge/adrs/", "adr", nil)

    err := UpdateOutputTarget(ctx, "adr-001.md", ".sow/knowledge/decisions/")
    require.NoError(t, err)

    // Verify target was updated
    output, err := GetOutput(ctx, "adr-001.md")
    require.NoError(t, err)
    assert.Equal(t, ".sow/knowledge/decisions/", output.Target_location)
}

func TestUpdateStatus(t *testing.T) {
    ctx, cleanup := createTestContext(t)
    defer cleanup()

    InitDesign(ctx, "test", "design/test")

    err := UpdateStatus(ctx, "in_review")
    require.NoError(t, err)

    index, err := LoadIndex(ctx)
    require.NoError(t, err)
    assert.Equal(t, "in_review", index.Design.Status)

    // Test invalid status
    err = UpdateStatus(ctx, "invalid")
    assert.Error(t, err)
}
```

---

### 6. Entry Point Command

**File**: `cli/cmd/design.go`

```go
package cmd

import (
    "fmt"
    "os"
    "strings"

    gogit "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing"
    "github.com/jmgilman/sow/cli/internal/cmdutil"
    "github.com/jmgilman/sow/cli/internal/design"
    "github.com/jmgilman/sow/cli/internal/prompts"
    "github.com/jmgilman/sow/cli/internal/sow"
    "github.com/spf13/cobra"
)

// NewDesignCmd creates the design command.
func NewDesignCmd() *cobra.Command {
    var branchName string

    cmd := &cobra.Command{
        Use:   "design [prompt]",
        Short: "Start or resume a design mode session",
        Long: `Start or resume design mode for creating formal design documents.

Design mode provides a guided environment for:
- Synthesizing research findings into formal documents
- Creating ADRs (Architecture Decision Records)
- Documenting architecture and system design
- Producing diagrams and specifications
- Collaborating with stakeholders on design decisions

This command handles design lifecycle based on context:

No arguments:
  - Uses current branch
  - If design exists: continues it
  - If not: creates new design (validates not on protected branch)

With [prompt]:
  - Provides initial context to the orchestrator
  - Useful for scoping the design topic

With --branch:
  - Checks out the branch (creates if doesn't exist)
  - If design exists: continues it
  - If not: creates new design

Directory Structure:
  - .sow/design/              Design workspace
  - .sow/design/index.yaml    Input/output index

The design index tracks:
- Input sources (explorations, files, references) that inform the design
- Planned outputs (documents) with their target locations

Examples:
  sow design                                      # Continue or start in current branch
  sow design "Create auth system architecture"   # Start with context
  sow design --branch design/auth-system          # Work on specific branch`,
        Args: cobra.MaximumNArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            initialPrompt := ""
            if len(args) > 0 {
                initialPrompt = args[0]
            }
            return runDesign(cmd, branchName, initialPrompt)
        },
    }

    cmd.Flags().StringVar(&branchName, "branch", "", "Branch to work on (creates if doesn't exist)")

    return cmd
}

func runDesign(cmd *cobra.Command, branchName, initialPrompt string) error {
    ctx := cmdutil.GetContext(cmd.Context())

    // Require sow to be initialized
    if !ctx.IsInitialized() {
        fmt.Fprintln(os.Stderr, "Error: sow not initialized in this repository")
        fmt.Fprintln(os.Stderr, "Run: sow init")
        return fmt.Errorf("not initialized")
    }

    var selectedBranch string
    var topic string
    var shouldCreateNew bool
    var err error

    if branchName != "" {
        // Scenario: --branch flag provided
        selectedBranch, topic, shouldCreateNew, err = handleDesignBranchScenario(ctx, branchName)
        if err != nil {
            return err
        }
    } else {
        // Scenario: No flags (current branch)
        selectedBranch, topic, shouldCreateNew, err = handleDesignCurrentBranchScenario(ctx)
        if err != nil {
            return err
        }
    }

    // At this point we know:
    // - selectedBranch: which branch we're on
    // - topic: the design topic
    // - shouldCreateNew: whether to create new or continue

    if shouldCreateNew {
        // Create new design
        if err := design.InitDesign(ctx, topic, selectedBranch); err != nil {
            return fmt.Errorf("failed to initialize design: %w", err)
        }
        cmd.Printf("\n✓ Created new design session: %s\n", topic)
        cmd.Printf("  Branch: %s\n", selectedBranch)
    } else {
        // Continue existing design
        index, err := design.LoadIndex(ctx)
        if err != nil {
            return fmt.Errorf("failed to load design: %w", err)
        }
        cmd.Printf("\n✓ Resuming design: %s\n", index.Design.Topic)
        cmd.Printf("  Branch: %s\n", index.Design.Branch)
        cmd.Printf("  Status: %s\n", index.Design.Status)
        cmd.Printf("  Inputs:  %d\n", len(index.Inputs))
        cmd.Printf("  Outputs: %d\n", len(index.Outputs))

        // Use the topic from the existing design
        topic = index.Design.Topic
    }

    // Generate design mode prompt
    designPrompt, err := generateDesignPrompt(ctx, topic, selectedBranch, initialPrompt)
    if err != nil {
        return fmt.Errorf("failed to generate design prompt: %w", err)
    }

    return launchClaudeCode(cmd, ctx, designPrompt)
}

// handleDesignBranchScenario handles the --branch flag scenario.
// Returns: (branchName, topic, shouldCreateNew, error).
func handleDesignBranchScenario(ctx *sow.Context, branchName string) (string, string, bool, error) {
    git := ctx.Git()

    // Check if branch exists locally
    branches, err := git.Branches()
    if err != nil {
        return "", "", false, fmt.Errorf("failed to list branches: %w", err)
    }

    branchExists := false
    for _, b := range branches {
        if b == branchName {
            branchExists = true
            break
        }
    }

    if branchExists {
        // Checkout existing branch
        if err := git.CheckoutBranch(branchName); err != nil {
            return "", "", false, fmt.Errorf("failed to checkout branch %s: %w", branchName, err)
        }
    } else {
        // Create new branch
        // First check we're on a safe branch to create from
        currentBranch, err := git.CurrentBranch()
        if err != nil {
            return "", "", false, fmt.Errorf("failed to get current branch: %w", err)
        }

        if !git.IsProtectedBranch(currentBranch) {
            return "", "", false, fmt.Errorf("cannot create branch %s from %s - please checkout main/master first", branchName, currentBranch)
        }

        // Create and checkout new branch
        if err := createBranch(git, branchName); err != nil {
            return "", "", false, fmt.Errorf("failed to create branch %s: %w", branchName, err)
        }
    }

    // Extract topic from branch name
    // If it starts with design/, strip that prefix, otherwise use the whole name
    topic := branchName
    if strings.HasPrefix(branchName, "design/") {
        topic = strings.TrimPrefix(branchName, "design/")
    }

    // Check if design exists in this branch
    designExists := design.Exists(ctx)

    return branchName, topic, !designExists, nil
}

// handleDesignCurrentBranchScenario handles the no-flags scenario (current branch).
// Returns: (branchName, topic, shouldCreateNew, error).
func handleDesignCurrentBranchScenario(ctx *sow.Context) (string, string, bool, error) {
    git := ctx.Git()

    currentBranch, err := git.CurrentBranch()
    if err != nil {
        return "", "", false, fmt.Errorf("failed to get current branch: %w", err)
    }

    // Check if design exists
    designExists := design.Exists(ctx)

    if !designExists {
        // Validate we're not on a protected branch before creating new
        if git.IsProtectedBranch(currentBranch) {
            return "", "", false, fmt.Errorf("cannot create design on protected branch '%s' - create a branch first", currentBranch)
        }
    }

    // Extract topic from branch name
    topic := currentBranch
    if strings.HasPrefix(currentBranch, "design/") {
        topic = strings.TrimPrefix(currentBranch, "design/")
    }

    return currentBranch, topic, !designExists, nil
}

// createBranch creates a new branch and checks it out.
// Reuses implementation from explore.go
func createBranch(git *sow.Git, branchName string) error {
    // Use underlying go-git to create branch
    wt, err := git.Repository().Underlying().Worktree()
    if err != nil {
        return fmt.Errorf("failed to get worktree: %w", err)
    }

    // Get current HEAD
    head, err := git.Repository().Underlying().Head()
    if err != nil {
        return fmt.Errorf("failed to get HEAD: %w", err)
    }

    // Create branch reference
    branchRef := "refs/heads/" + branchName
    if err := git.Repository().Underlying().Storer.SetReference(
        plumbing.NewHashReference(plumbing.ReferenceName(branchRef), head.Hash()),
    ); err != nil {
        return fmt.Errorf("failed to create branch reference: %w", err)
    }

    // Checkout the new branch
    if err := wt.Checkout(&gogit.CheckoutOptions{
        Branch: plumbing.ReferenceName(branchRef),
    }); err != nil {
        return fmt.Errorf("failed to checkout new branch: %w", err)
    }

    return nil
}

// generateDesignPrompt creates the design mode prompt with context.
func generateDesignPrompt(sowCtx *sow.Context, topic, branch, initialPrompt string) (string, error) {
    // Load design index if it exists
    var inputs []prompts.DesignInput
    var outputs []prompts.DesignOutput
    status := "active"

    if design.Exists(sowCtx) {
        index, err := design.LoadIndex(sowCtx)
        if err != nil {
            return "", fmt.Errorf("failed to load design index: %w", err)
        }

        // Convert schema inputs to prompt inputs
        for _, input := range index.Inputs {
            inputs = append(inputs, prompts.DesignInput{
                Type:        input.Type,
                Path:        input.Path,
                Description: input.Description,
                Tags:        input.Tags,
            })
        }

        // Convert schema outputs to prompt outputs
        for _, output := range index.Outputs {
            outputs = append(outputs, prompts.DesignOutput{
                Path:           output.Path,
                Description:    output.Description,
                TargetLocation: output.Target_location,
                Type:           output.Type,
                Tags:           output.Tags,
            })
        }

        status = index.Design.Status
    }

    // Create design context
    ctx := &prompts.DesignContext{
        Topic:         topic,
        Branch:        branch,
        Status:        status,
        Inputs:        inputs,
        Outputs:       outputs,
        InitialPrompt: initialPrompt,
    }

    // Render prompt
    prompt, err := prompts.Render(prompts.PromptModeDesign, ctx)
    if err != nil {
        return "", fmt.Errorf("failed to render design prompt: %w", err)
    }

    return prompt, nil
}
```

---

### 7-14. Management Commands

All management commands follow similar patterns. See existing `cmd/exploration/*.go` files for reference structure.

**File**: `cli/cmd/design/design.go`
- Parent command for subcommands
- Pattern: Copy from `cmd/exploration/exploration.go`

**File**: `cli/cmd/design/add_input.go`
- Registers input sources
- Flags: `--type`, `--description`, `--tags`
- Pattern: Similar to `cmd/exploration/add_file.go`

**File**: `cli/cmd/design/remove_input.go`
- Removes input by path
- Pattern: Similar to `cmd/exploration/remove_file.go`

**File**: `cli/cmd/design/add_output.go`
- Registers planned output documents
- Flags: `--description`, `--target`, `--type`, `--tags`
- Validates target is not empty

**File**: `cli/cmd/design/remove_output.go`
- Removes output by path

**File**: `cli/cmd/design/set_output_target.go`
- Updates output's target location
- Args: `<path> <new-target>`

**File**: `cli/cmd/design/set_status.go`
- Updates design status
- Args: `<status>` (active, in_review, completed)

**File**: `cli/cmd/design/index.go`
- Displays complete index (inputs + outputs)
- Format: Structured, readable output

---

### 15. Prompt Template

**File**: `cli/internal/prompts/templates/modes/design.md`

See detailed template in "Revised Plan - Key Changes" section above (section 4).

Key sections:
1. Role definition
2. Workspace info
3. Input sources list
4. Planned outputs list
5. Output management commands
6. Design workflow (4 phases)
7. Best practices
8. Getting started guidance

---

### 16. Prompt Context

**File**: `cli/internal/prompts/context.go` (additions)

Add to existing file:

```go
// DesignContext holds the context for rendering design mode prompts.
type DesignContext struct {
    Topic         string
    Branch        string
    Status        string
    Inputs        []DesignInput
    Outputs       []DesignOutput
    InitialPrompt string
}

// DesignInput represents an input for template rendering.
type DesignInput struct {
    Type        string
    Path        string
    Description string
    Tags        []string
}

// DesignOutput represents an output for template rendering.
type DesignOutput struct {
    Path           string
    Description    string
    TargetLocation string
    Type           string
    Tags           []string
}

// ToMap converts DesignContext to a map for template rendering.
func (c *DesignContext) ToMap() map[string]interface{} {
    data := make(map[string]interface{})
    data["Topic"] = c.Topic
    data["Branch"] = c.Branch
    data["Status"] = c.Status
    data["InitialPrompt"] = c.InitialPrompt

    if len(c.Inputs) > 0 {
        inputs := make([]map[string]interface{}, len(c.Inputs))
        for i, input := range c.Inputs {
            inputs[i] = map[string]interface{}{
                "Type":        input.Type,
                "Path":        input.Path,
                "Description": input.Description,
                "Tags":        input.Tags,
            }
        }
        data["Inputs"] = inputs
    }

    if len(c.Outputs) > 0 {
        outputs := make([]map[string]interface{}, len(c.Outputs))
        for i, output := range c.Outputs {
            outputs[i] = map[string]interface{}{
                "Path":           output.Path,
                "Description":    output.Description,
                "TargetLocation": output.TargetLocation,
                "Type":           output.Type,
                "Tags":           output.Tags,
            }
        }
        data["Outputs"] = outputs
    }

    return data
}
```

---

### 17. Prompt Registry

**File**: `cli/internal/prompts/prompts.go` (additions)

Add constant:
```go
const (
    // ... existing constants
    PromptModeDesign PromptID = "mode.design"
)
```

Register in `init()`:
```go
promptFiles := map[PromptID]string{
    // ... existing entries
    PromptModeDesign: "templates/modes/design.md",
}
```

Ensure embed directive includes design template:
```go
//go:embed templates/**/*.md templates/greet/*.md templates/greet/states/*.md templates/commands/*.md templates/modes/*.md templates/guidance/*.md
var templatesFS embed.FS
```

---

### 18. Root Command Integration

**File**: `cli/cmd/root.go` (additions)

Add import:
```go
"github.com/jmgilman/sow/cli/cmd/design"
```

Add commands in `NewRootCmd()`:
```go
// Add subcommands
cmd.AddCommand(NewInitCmd())
cmd.AddCommand(NewValidateCmd())
cmd.AddCommand(NewStartCmd())
cmd.AddCommand(NewProjectCmd())
cmd.AddCommand(NewExploreCmd())
cmd.AddCommand(NewDesignCmd())  // <-- Add this
cmd.AddCommand(NewPromptCmd())
cmd.AddCommand(exploration.NewExplorationCmd())
cmd.AddCommand(design.NewDesignCmd())  // <-- Add this (for subcommands)
cmd.AddCommand(issue.NewIssueCmd())
cmd.AddCommand(refs.NewRefsCmd())
cmd.AddCommand(agent.NewAgentCmd())
```

---

## Testing Plan

### Unit Tests

Run existing test suite:
```bash
cd cli
go test ./internal/design/...
```

Expected coverage: >85%

### Integration Test

Manual test script:

```bash
# Setup
cd /tmp && mkdir test-design-mode && cd test-design-mode
git init
git config user.name "Test User"
git config user.email "test@example.com"
sow init

# Create exploration first (to have inputs)
git checkout -b explore/auth
sow explore "OAuth research"
echo "# OAuth Research" > .sow/exploration/oauth.md
sow exploration add-file oauth.md --description "OAuth findings" --tags "oauth"
git add .sow/
git commit -m "Add exploration"

# Start design session
git checkout -b design/auth-system
sow design

# Verify design was created
ls -la .sow/design/
cat .sow/design/index.yaml

# Add inputs
sow design add-input .sow/exploration/ \
  --type exploration \
  --description "OAuth research findings" \
  --tags "research,oauth"

sow design add-input "docs/*.md" \
  --type file \
  --description "Existing docs" \
  --tags "current-state"

# Add outputs
sow design add-output adr-001-oauth.md \
  --description "OAuth decision" \
  --target .sow/knowledge/adrs/ \
  --type adr \
  --tags "oauth,decision"

sow design add-output architecture.md \
  --description "System architecture" \
  --target .sow/knowledge/architecture/ \
  --type architecture

# View index
sow design index

# Update status
sow design set-status in_review

# Verify status updated
sow design index | grep Status

# Update output target
sow design set-output-target adr-001-oauth.md .sow/knowledge/decisions/

# Verify target updated
sow design index | grep "adr-001-oauth.md"

# Test resumption
sow design  # Should resume existing session

# Clean up
cd /tmp && rm -rf test-design-mode
```

### Edge Cases to Test

1. **Branch protection**: Cannot create design on main/master
2. **Duplicate inputs**: Adding same path twice
3. **Duplicate outputs**: Adding same path twice
4. **Missing design**: Commands fail gracefully when no design exists
5. **Invalid status**: Reject invalid status values
6. **Empty target**: Reject empty target location for outputs
7. **Resume across sessions**: Index persists and loads correctly

---

## Usage Examples

### Basic Workflow

```bash
# 1. Start design session
git checkout -b design/auth-system
sow design "Create authentication system design"

# 2. Register inputs (in Claude Code session)
sow design add-input .sow/exploration/auth-research.md \
  --type exploration \
  --description "OAuth vs JWT research" \
  --tags "oauth,jwt"

# 3. Work with user to identify needed docs, register outputs
sow design add-output adr-001-oauth.md \
  --description "Decision to use OAuth 2.0" \
  --target .sow/knowledge/adrs/ \
  --type adr

sow design add-output architecture-overview.md \
  --description "High-level system architecture" \
  --target .sow/knowledge/architecture/ \
  --type architecture

# 4. Create documents (orchestrator does this)
# Files created in .sow/design/

# 5. Mark for review
sow design set-status in_review

# 6. After approval, finalize (orchestrator does this):
# - Move files to target locations
# - Create PR
# - Clean up .sow/design/
```

### Resuming Session

```bash
# Simply run design command again
sow design

# Claude Code launches with full context:
# - All inputs loaded
# - All outputs with targets
# - Current status
```

### Multiple Target Locations

```bash
# ADR goes to knowledge base
sow design add-output adr-001.md \
  --target .sow/knowledge/adrs/ \
  --type adr

# Diagram goes to docs
sow design add-output auth-flow.mmd \
  --target docs/diagrams/ \
  --type diagram

# Spec goes to different location
sow design add-output api-spec.md \
  --target docs/api/ \
  --type spec
```

---

## Future Enhancements (Out of Scope for MVP)

1. **Exploration migration**: `sow design --from-exploration explore/auth`
2. **Template support**: `sow design --template adr` (pre-populate outputs)
3. **Validation**: Check that output paths exist before finalization
4. **Auto-discovery**: Scan `.sow/design/` for unregistered files
5. **Dependency tracking**: Link inputs to specific outputs
6. **Version history**: Track iterations of design documents
7. **Export**: Generate summary reports from index

---

## Notes for Implementation

### Code Style

- Follow existing patterns from `exploration` package
- Use same error handling approach (sentinel errors)
- Match command structure and flags style
- Maintain consistent output formatting

### Dependencies

All dependencies already exist:
- `github.com/spf13/cobra` - CLI framework
- `gopkg.in/yaml.v3` - YAML parsing
- `github.com/go-git/go-git/v5` - Git operations

### Testing

- Reuse test helpers from other packages
- Create temp directories for integration tests
- Clean up after tests
- Use `testify` for assertions

### Documentation

After implementation, update:
- `README.md` - Add design mode overview
- `.claude/CLAUDE.md` - Update with design mode instructions
- Create `docs/design-mode.md` - Detailed usage guide

---

## Success Criteria

Implementation is complete when:

- [ ] All 18 files created/modified
- [ ] Unit tests pass with >85% coverage
- [ ] Integration test script passes
- [ ] `sow design` command works end-to-end
- [ ] All management commands functional
- [ ] Prompt template renders correctly
- [ ] Index persists and loads correctly
- [ ] Branch protection enforced
- [ ] Glob patterns work for inputs
- [ ] Multiple target locations supported for outputs

---

## Implementation Order

Recommended sequence:

1. **Schema first**: Create CUE schema and generate Go types
2. **Core logic**: Implement index operations and manager functions
3. **Test core**: Write and verify unit tests
4. **Entry command**: Implement `sow design` command
5. **Management commands**: Implement all subcommands
6. **Prompt system**: Create template and wire up context
7. **Integration**: Wire into root command
8. **Test end-to-end**: Run integration test script
9. **Polish**: Fix any issues, improve error messages

---

## Questions During Implementation

If anything is unclear:

1. **Check exploration mode**: Most patterns can be copied from there
2. **Look at project.go**: For branch handling and git operations
3. **Review prompts package**: For template and context patterns
4. **Ask for clarification**: If fundamentally unclear

This plan should provide sufficient context for any agent to pick up and implement design mode. Good luck!
