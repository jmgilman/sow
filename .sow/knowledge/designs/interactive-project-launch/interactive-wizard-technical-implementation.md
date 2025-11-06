# Interactive Wizard Technical Implementation

**Document Type**: Technical Design Specification
**Task**: 010 (Part 2 of 2)
**Author**: Architecture Team
**Date**: 2025-01-06
**Status**: Draft
**Related**: [UX Flow Design](./wizard-ux-flow.md), [huh Library Verification](../../context/huh-library-verification.md)

## Overview

This document specifies the technical implementation of the interactive project wizard. It covers the state machine architecture, huh library integration, validation logic, and testing strategy.

**Companion document**: [UX Flow Design](./wizard-ux-flow.md) covers the user-facing flows, screens, and error messages.

---

## Architecture

### Wizard State Machine

The wizard uses a lightweight state machine to track progress and handle navigation:

```go
type WizardState string

const (
    StateEntry          WizardState = "entry"
    StateCreateSource   WizardState = "create_source"
    StateIssueSelect    WizardState = "issue_select"
    StateTypeSelect     WizardState = "type_select"
    StateNameEntry      WizardState = "name_entry"
    StatePromptEntry    WizardState = "prompt_entry"
    StateProjectSelect  WizardState = "project_select"
    StateContinuePrompt WizardState = "continue_prompt"
    StateComplete       WizardState = "complete"
    StateCancelled      WizardState = "cancelled"
)

type Wizard struct {
    state   WizardState
    ctx     *sow.Context
    choices map[string]interface{}
}

func (w *Wizard) Run() error {
    for w.state != StateComplete && w.state != StateCancelled {
        if err := w.handleState(); err != nil {
            return err
        }
    }

    if w.state == StateCancelled {
        return nil  // User cancelled, no error
    }

    // Finalize: create/continue project and launch Claude
    return w.finalize()
}

func (w *Wizard) handleState() error {
    switch w.state {
    case StateEntry:
        return w.handleEntry()
    case StateCreateSource:
        return w.handleCreateSource()
    case StateIssueSelect:
        return w.handleIssueSelect()
    case StateTypeSelect:
        return w.handleTypeSelect()
    case StateNameEntry:
        return w.handleNameEntry()
    case StatePromptEntry:
        return w.handlePromptEntry()
    case StateProjectSelect:
        return w.handleProjectSelect()
    case StateContinuePrompt:
        return w.handleContinuePrompt()
    }
    return nil
}
```

### Data Flow

```
Entry → choices["action"] = "create" | "continue"

Create Path:
  Source → choices["source"] = "issue" | "branch"
  Issue  → choices["issue"] = IssueInfo
  Type   → choices["type"] = "standard" | "exploration" | "design" | "breakdown"
  Name   → choices["name"] = "user input"
  Prompt → choices["prompt"] = "user input"

Continue Path:
  Project → choices["project"] = ProjectInfo
  Prompt  → choices["prompt"] = "user input"

Finalize → Use choices map to create/continue project
```

---

## huh Library Integration

### Key Capabilities

Based on [huh library verification](../../context/huh-library-verification.md), the library supports:

- ✅ Select prompts with single/multi-selection
- ✅ Input fields with validation and inline errors
- ✅ Text areas with multi-line editing
- ✅ External editor support via **Ctrl+E** (not Ctrl+O)
- ✅ Real-time field updates via `DescriptionFunc`
- ✅ Dynamic forms with conditional fields
- ✅ Built-in help text and keybinding display

### Form Structure Pattern

Each screen is implemented as a huh form with groups:

```go
form := huh.NewForm(
    huh.NewGroup(
        // Fields for this screen
    ),
)

if err := form.Run(); err != nil {
    if errors.Is(err, huh.ErrUserAborted) {
        w.state = StateCancelled
        return nil
    }
    return err
}

// Transition to next state based on user input
```

---

## Implementation Examples

### Example 1: Entry Screen (Select Prompt)

```go
func (w *Wizard) handleEntry() error {
    var action string

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("What would you like to do?").
                Options(
                    huh.NewOption("Create new project", "create"),
                    huh.NewOption("Continue existing project", "continue"),
                    huh.NewOption("Cancel", "cancel"),
                ).
                Value(&action),
        ),
    )

    if err := form.Run(); err != nil {
        return err
    }

    w.choices["action"] = action

    // Transition
    switch action {
    case "create":
        w.state = StateCreateSource
    case "continue":
        w.state = StateProjectSelect
    case "cancel":
        w.state = StateCancelled
    }

    return nil
}
```

### Example 2: Name Entry with Validation and Preview

```go
func (w *Wizard) handleNameEntry() error {
    var name string
    projectType := w.choices["type"].(string)
    prefix := getTypePrefix(projectType)

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().
                Title("Enter project name:").
                Placeholder("e.g., Web Based Agents").
                Value(&name).
                Validate(func(s string) error {
                    if strings.TrimSpace(s) == "" {
                        return fmt.Errorf("project name cannot be empty")
                    }

                    normalized := normalizeName(s)
                    branchName := fmt.Sprintf("%s/%s", prefix, normalized)

                    if isProtectedBranch(branchName) {
                        return fmt.Errorf("cannot use protected branch name")
                    }

                    return nil
                }),

            // Real-time preview
            huh.NewNote().
                Title("Preview").
                DescriptionFunc(func() string {
                    if name == "" {
                        return fmt.Sprintf("%s/<project-name>", prefix)
                    }
                    normalized := normalizeName(name)
                    return fmt.Sprintf("%s/%s", prefix, normalized)
                }, &name), // Bind to name field for updates
        ),
    )

    if err := form.Run(); err != nil {
        return err
    }

    w.choices["name"] = name
    w.state = StatePromptEntry

    return nil
}
```

### Example 3: Text Area with External Editor

```go
func (w *Wizard) handlePromptEntry() error {
    var prompt string

    // Get context for display
    var contextLine string
    if issue, ok := w.choices["issue"].(IssueInfo); ok {
        contextLine = fmt.Sprintf("Issue: #%d - %s", issue.Number, issue.Title)
    } else {
        projectType := w.choices["type"].(string)
        name := w.choices["name"].(string)
        prefix := getTypePrefix(projectType)
        normalized := normalizeName(name)
        branchName := fmt.Sprintf("%s/%s", prefix, normalized)
        contextLine = fmt.Sprintf("Branch: %s", branchName)
    }

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewText().
                Title("Enter your task or question for Claude (optional):").
                Description(fmt.Sprintf("%s\nPress Ctrl+E to open $EDITOR for multi-line input", contextLine)).
                CharLimit(5000).
                Value(&prompt).
                WithEditor(true), // Enable external editor via Ctrl+E
        ),
    )

    if err := form.Run(); err != nil {
        return err
    }

    w.choices["prompt"] = prompt
    w.state = StateComplete

    return nil
}
```

### Example 4: Project List with Dynamic Options

```go
func (w *Wizard) handleProjectSelect() error {
    // Discover projects
    projects, err := listProjects(w.ctx)
    if err != nil {
        return fmt.Errorf("failed to list projects: %w", err)
    }

    if len(projects) == 0 {
        fmt.Fprintln(os.Stderr, "No existing projects found")
        w.state = StateCancelled
        return nil
    }

    var selectedProject string

    // Build options dynamically
    options := make([]huh.Option[string], 0, len(projects)+1)
    for _, proj := range projects {
        label := formatProjectOption(proj)
        options = append(options, huh.NewOption(label, proj.Branch))
    }
    options = append(options, huh.NewOption("Cancel", "cancel"))

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("Select a project to continue:").
                Options(options...).
                Value(&selectedProject),
        ),
    )

    if err := form.Run(); err != nil {
        return err
    }

    if selectedProject == "cancel" {
        w.state = StateCancelled
        return nil
    }

    // Validate project still exists
    proj, err := validateProject(w.ctx, selectedProject)
    if err != nil {
        // Show error and return to project list
        showError(err.Error())
        return nil // Stay in current state to retry
    }

    w.choices["project"] = proj
    w.state = StateContinuePrompt

    return nil
}

func formatProjectOption(proj ProjectInfo) string {
    // Format: "branch - name [Type: phase, x/y tasks completed]"
    progress := formatProgress(proj)
    return fmt.Sprintf("%s - %s\n[%s]", proj.Branch, proj.Name, progress)
}

func formatProgress(proj ProjectInfo) string {
    // Format: "<Type>: <phase>[, x/y tasks completed]"
    if proj.TasksTotal > 0 {
        return fmt.Sprintf("%s: %s, %d/%d tasks completed",
            proj.Type, proj.Phase, proj.TasksCompleted, proj.TasksTotal)
    }
    return fmt.Sprintf("%s: %s", proj.Type, proj.Phase)
}
```

---

## State Detection Logic

### Check Branch State (for Creation)

```go
type BranchState struct {
    BranchExists  bool
    WorktreeExists bool
    ProjectExists bool
}

func checkBranchState(ctx *sow.Context, branchName string) (*BranchState, error) {
    state := &BranchState{}

    // Check if branch exists
    branches, err := ctx.Git().Branches()
    if err != nil {
        return nil, err
    }
    for _, b := range branches {
        if b == branchName {
            state.BranchExists = true
            break
        }
    }

    // Check if worktree exists
    worktreePath := sow.WorktreePath(ctx.RepoRoot(), branchName)
    if _, err := os.Stat(worktreePath); err == nil {
        state.WorktreeExists = true

        // Check if project exists in worktree
        projectPath := filepath.Join(worktreePath, ".sow", "project", "state.yaml")
        if _, err := os.Stat(projectPath); err == nil {
            state.ProjectExists = true
        }
    }

    return state, nil
}

func canCreateProject(state *BranchState) error {
    if state.ProjectExists {
        return fmt.Errorf("project already exists on this branch")
    }
    // Branch and worktree can exist - we'll create/ensure them
    return nil
}
```

### List Existing Projects

```go
type ProjectInfo struct {
    Branch         string
    Name           string
    Type           string
    Phase          string
    TasksCompleted int
    TasksTotal     int
    ModTime        time.Time
}

func listProjects(ctx *sow.Context) ([]ProjectInfo, error) {
    worktreesDir := filepath.Join(ctx.RepoRoot(), ".sow", "worktrees")

    entries, err := os.ReadDir(worktreesDir)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, nil // No worktrees directory = no projects
        }
        return nil, err
    }

    var projects []ProjectInfo
    for _, entry := range entries {
        if !entry.IsDir() {
            continue
        }

        branchName := entry.Name()
        projectPath := filepath.Join(worktreesDir, branchName, ".sow", "project", "state.yaml")

        // Only include if project exists
        info, err := os.Stat(projectPath)
        if err != nil {
            continue // Skip if no project
        }

        // Load project state
        proj, err := state.LoadFromPath(projectPath)
        if err != nil {
            continue // Skip invalid projects
        }

        // Count tasks if applicable
        var tasksCompleted, tasksTotal int
        if phase := proj.GetActivePhase(); phase != nil {
            for _, task := range phase.Tasks {
                tasksTotal++
                if task.Status == "completed" {
                    tasksCompleted++
                }
            }
        }

        projects = append(projects, ProjectInfo{
            Branch:         branchName,
            Name:           proj.Name,
            Type:           proj.Type,
            Phase:          proj.Machine.State(),
            TasksCompleted: tasksCompleted,
            TasksTotal:     tasksTotal,
            ModTime:        info.ModTime(),
        })
    }

    // Sort by most recently modified
    sort.Slice(projects, func(i, j int) bool {
        return projects[i].ModTime.After(projects[j].ModTime)
    })

    return projects, nil
}
```

### Conditional Uncommitted Changes Check

```go
func shouldCheckUncommittedChanges(ctx *sow.Context, targetBranch string) (bool, error) {
    currentBranch, err := ctx.Git().CurrentBranch()
    if err != nil {
        return false, err
    }

    // Only check if we'll need to switch branches
    // (git worktree can't have same branch checked out twice)
    return currentBranch == targetBranch, nil
}

func performUncommittedChangesCheckIfNeeded(ctx *sow.Context, targetBranch string) error {
    shouldCheck, err := shouldCheckUncommittedChanges(ctx, targetBranch)
    if err != nil {
        return err
    }

    if !shouldCheck {
        return nil // No check needed
    }

    // Use existing validation
    if err := sow.CheckUncommittedChanges(ctx); err != nil {
        return fmt.Errorf("repository has uncommitted changes\n\n"+
            "You are currently on branch '%s'.\n"+
            "Creating a worktree requires switching to a different branch first.\n\n"+
            "To fix:\n"+
            "  Commit: git add . && git commit -m \"message\"\n"+
            "  Or stash: git stash",
            targetBranch)
    }

    return nil
}
```

---

## Helper Functions

### Project Type Configuration

```go
type ProjectTypeConfig struct {
    Prefix      string
    Description string
}

var projectTypes = map[string]ProjectTypeConfig{
    "standard": {
        Prefix:      "feat/",
        Description: "Feature work and bug fixes",
    },
    "exploration": {
        Prefix:      "explore/",
        Description: "Research and investigation",
    },
    "design": {
        Prefix:      "design/",
        Description: "Architecture and design documents",
    },
    "breakdown": {
        Prefix:      "breakdown/",
        Description: "Decompose work into tasks",
    },
}

func getTypePrefix(projectType string) string {
    if config, ok := projectTypes[projectType]; ok {
        return config.Prefix
    }
    return "feat/" // Default to standard
}

func getTypeOptions() []huh.Option[string] {
    return []huh.Option[string]{
        huh.NewOption("Standard - Feature work and bug fixes", "standard"),
        huh.NewOption("Exploration - Research and investigation", "exploration"),
        huh.NewOption("Design - Architecture and design documents", "design"),
        huh.NewOption("Breakdown - Decompose work into tasks", "breakdown"),
        huh.NewOption("Cancel", "cancel"),
    }
}
```

### Name Normalization

```go
func normalizeName(name string) string {
    // Trim whitespace
    name = strings.TrimSpace(name)

    // Convert to lowercase
    name = strings.ToLower(name)

    // Replace spaces with hyphens
    name = strings.ReplaceAll(name, " ", "-")

    // Remove invalid characters (keep only a-z, 0-9, -, _)
    var result strings.Builder
    for _, r := range name {
        if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
            result.WriteRune(r)
        }
    }
    name = result.String()

    // Collapse multiple consecutive hyphens
    for strings.Contains(name, "--") {
        name = strings.ReplaceAll(name, "--", "-")
    }

    // Remove leading/trailing hyphens
    name = strings.Trim(name, "-")

    return name
}
```

### Branch Validation

```go
func isProtectedBranch(branchName string) bool {
    return branchName == "main" || branchName == "master"
}

func isValidBranchName(name string) error {
    if name == "" {
        return fmt.Errorf("branch name cannot be empty")
    }

    if isProtectedBranch(name) {
        return fmt.Errorf("cannot use protected branch name")
    }

    // Git ref name validation
    invalidPatterns := []string{
        "..",
        "//",
        "~",
        "^",
        ":",
        "?",
        "*",
        "[",
    }

    for _, pattern := range invalidPatterns {
        if strings.Contains(name, pattern) {
            return fmt.Errorf("invalid characters in branch name")
        }
    }

    if strings.HasPrefix(name, "/") || strings.HasSuffix(name, "/") {
        return fmt.Errorf("branch name cannot start or end with /")
    }

    if strings.Contains(name, " ") {
        return fmt.Errorf("branch name cannot contain spaces")
    }

    return nil
}
```

---

## Error Handling

### Error Display Pattern

```go
func showError(message string) {
    // Use huh's Confirm for simple acknowledgment
    _ = huh.NewForm(
        huh.NewGroup(
            huh.NewNote().
                Title("Error").
                Description(message),
            huh.NewConfirm().
                Title("Press Enter to continue").
                Affirmative("Continue").
                Negative(""), // Hide negative option
        ),
    ).Run()
}

func showErrorWithOptions(message string, options map[string]string) (string, error) {
    var choice string

    opts := make([]huh.Option[string], 0, len(options))
    for label, value := range options {
        opts = append(opts, huh.NewOption(label, value))
    }

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewNote().
                Title("Error").
                Description(message),
            huh.NewSelect[string]().
                Title("What would you like to do?").
                Options(opts...).
                Value(&choice),
        ),
    )

    if err := form.Run(); err != nil {
        return "", err
    }

    return choice, nil
}
```

---

## Testing Strategy

### Unit Tests

**Name Normalization**:
```go
func TestNormalizeName(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"Web Based Agents", "web-based-agents"},
        {"API V2", "api-v2"},
        {"feature--name", "feature-name"},
        {"-leading-trailing-", "leading-trailing"},
        {"With!Invalid@Chars#", "withinvalidchars"},
    }

    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            result := normalizeName(tt.input)
            if result != tt.expected {
                t.Errorf("normalizeName(%q) = %q, want %q", tt.input, result, tt.expected)
            }
        })
    }
}
```

**Branch Validation**:
```go
func TestIsValidBranchName(t *testing.T) {
    tests := []struct {
        name    string
        wantErr bool
    }{
        {"feat/valid-name", false},
        {"explore/test", false},
        {"main", true},  // Protected
        {"master", true}, // Protected
        {"has spaces", true},
        {"has..dots", true},
        {"has//slashes", true},
        {"/leading-slash", true},
        {"trailing-slash/", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := isValidBranchName(tt.name)
            if (err != nil) != tt.wantErr {
                t.Errorf("isValidBranchName(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
            }
        })
    }
}
```

### Integration Tests

**Wizard State Machine**:
```go
func TestWizardCreateFlow(t *testing.T) {
    w := &Wizard{
        state:   StateEntry,
        ctx:     testContext(),
        choices: make(map[string]interface{}),
    }

    // Simulate user selections
    w.choices["action"] = "create"
    w.state = StateCreateSource

    w.choices["source"] = "branch"
    w.state = StateTypeSelect

    w.choices["type"] = "exploration"
    w.state = StateNameEntry

    w.choices["name"] = "Test Project"
    w.state = StatePromptEntry

    w.choices["prompt"] = "Test prompt"
    w.state = StateComplete

    // Verify final state
    assert.Equal(t, StateComplete, w.state)
    assert.Equal(t, "exploration", w.choices["type"])
    assert.Equal(t, "Test Project", w.choices["name"])
}
```

### Manual Testing Checklist

- [ ] Test all paths (create from issue, create from branch, continue)
- [ ] Test all project types
- [ ] Test validation on all input fields
- [ ] Test external editor (Ctrl+E)
- [ ] Test error recovery (back navigation)
- [ ] Test with uncommitted changes
- [ ] Test with existing projects/branches
- [ ] Test GitHub issue integration
- [ ] Test on different terminal sizes
- [ ] Test with/without `gh` CLI installed

---

## Implementation Checklist

### Phase 1: Core Structure
- [ ] Wizard state machine implementation
- [ ] State handler skeleton
- [ ] Entry screen (create/continue/cancel)
- [ ] Cancel handling (Ctrl+C, Esc)

### Phase 2: Create from Branch
- [ ] Project type selection screen
- [ ] Project name entry with validation
- [ ] Real-time branch preview
- [ ] Name normalization function
- [ ] Branch state checking
- [ ] Initial prompt entry with Ctrl+E support

### Phase 3: Create from Issue
- [ ] GitHub CLI integration
- [ ] Issue listing screen
- [ ] Linked branch detection
- [ ] Issue selection
- [ ] Type selection for issues
- [ ] Branch creation via `gh issue develop`

### Phase 4: Continue Existing
- [ ] Worktree directory scanning
- [ ] Project state loading
- [ ] Project list screen with progress
- [ ] Project selection
- [ ] Continuation prompt entry

### Phase 5: Finalization
- [ ] Conditional uncommitted changes check
- [ ] Branch/worktree creation
- [ ] Project initialization
- [ ] 3-layer prompt generation
- [ ] Claude launch

### Phase 6: Error Handling
- [ ] All error messages implemented
- [ ] Error display functions
- [ ] Recovery flows
- [ ] Validation on all inputs

### Phase 7: Polish
- [ ] Loading indicators for async operations
- [ ] Help text and hints
- [ ] Success messages
- [ ] Testing on various terminals

---

## Performance Considerations

### Async Operations

Use spinner for long-running operations:

```go
import "github.com/charmbracelet/huh/spinner"

func fetchIssuesWithSpinner(ctx context.Context) ([]IssueInfo, error) {
    var issues []IssueInfo
    var err error

    _ = spinner.New().
        Title("Fetching issues from GitHub...").
        Action(func() {
            issues, err = fetchIssues(ctx)
        }).
        Run()

    return issues, err
}
```

### Caching

Consider caching frequently accessed data:
- GitHub issues (cache for session)
- Project list (refresh on demand)
- Branch list (refresh when needed)

---

## References

- **[UX Flow Design](./wizard-ux-flow.md)**: User-facing flows and screens
- **[huh Library Verification](../../context/huh-library-verification.md)**: Library capability verification
- **charmbracelet/huh**: https://github.com/charmbracelet/huh
- **huh Examples**: https://github.com/charmbracelet/huh/tree/main/examples
- **Bubble Tea**: https://github.com/charmbracelet/bubbletea (huh's foundation)
