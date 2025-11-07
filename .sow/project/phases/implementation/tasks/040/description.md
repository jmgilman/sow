# Task 040: Helper Functions and Utilities

## Context

This task implements helper functions that will be used throughout the wizard screens. These utilities provide common functionality for name normalization, project type configuration, branch name preview, error display, and loading indicators.

These helpers are designed to be pure functions (where possible) with well-defined inputs and outputs, making them easy to test and reuse. They implement the business logic that the wizard needs, separated from the UI layer.

**Key Design Decisions**:
- **Name normalization**: Converts user-friendly names into valid git branch names
- **Type configuration**: Maps project types to branch prefixes (feat/, explore/, design/, breakdown/)
- **Error display**: Shows errors using huh forms for consistency with the wizard UI
- **Loading indicators**: Provides spinners for long-running operations

## Requirements

### File: wizard_helpers.go

Create `cli/cmd/project/wizard_helpers.go` with the following functions:

### Function 1: normalizeName

**Purpose**: Convert user-friendly project names into valid git branch names.

**Signature**:
```go
func normalizeName(name string) string
```

**Algorithm** (from technical design lines 571-599):
1. Trim leading/trailing whitespace
2. Convert to lowercase
3. Replace spaces with hyphens
4. Remove invalid characters (keep only: a-z, 0-9, -, _)
5. Collapse multiple consecutive hyphens into single hyphen
6. Remove leading/trailing hyphens
7. Return normalized name

**Example Transformations**:
- "Web Based Agents" → "web-based-agents"
- "API V2" → "api-v2"
- "feature--name" → "feature-name"
- "-leading-trailing-" → "leading-trailing"
- "With!Invalid@Chars#" → "withinvalidchars"
- "UPPERCASE" → "uppercase"
- "  spaces  " → "spaces"

**Edge Cases**:
- Empty string → "" (empty string)
- Only spaces → "" (empty string)
- Only special characters → "" (empty string)
- Unicode characters → removed (only ASCII alphanumeric allowed)

### Function 2: ProjectTypeConfig Type and Data

**Purpose**: Define project type configuration and mappings.

**Type Definition**:
```go
type ProjectTypeConfig struct {
    Prefix      string
    Description string
}
```

**Global Map**:
```go
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
```

**Data Source**: UX design lines 129-135

### Function 3: getTypePrefix

**Purpose**: Get branch prefix for a project type, with fallback.

**Signature**:
```go
func getTypePrefix(projectType string) string
```

**Logic**:
- If projectType exists in map, return its Prefix
- Otherwise, return "feat/" as default
- Always returns a valid prefix (never empty)

**Examples**:
- `getTypePrefix("standard")` → "feat/"
- `getTypePrefix("exploration")` → "explore/"
- `getTypePrefix("unknown")` → "feat/" (fallback)
- `getTypePrefix("")` → "feat/" (fallback)

### Function 4: getTypeOptions

**Purpose**: Convert project types map into huh-compatible options for select prompts.

**Signature**:
```go
func getTypeOptions() []huh.Option[string]
```

**Logic**:
- Iterate over projectTypes map
- Create huh.NewOption() for each type
- Display label: "{Description}" (e.g., "Feature work and bug fixes")
- Value: type name (e.g., "standard")
- Order: standard, exploration, design, breakdown (consistent order)
- Add "Cancel" option at the end

**Return Example**:
```go
[]huh.Option[string]{
    huh.NewOption("Feature work and bug fixes", "standard"),
    huh.NewOption("Research and investigation", "exploration"),
    huh.NewOption("Architecture and design documents", "design"),
    huh.NewOption("Decompose work into tasks", "breakdown"),
    huh.NewOption("Cancel", "cancel"),
}
```

### Function 5: previewBranchName

**Purpose**: Show user what the branch name will be before they commit to it.

**Signature**:
```go
func previewBranchName(projectType, name string) string
```

**Logic**:
- Get prefix for project type
- Normalize the name
- Combine: `prefix + normalizedName`
- Example: `previewBranchName("standard", "Web Based Agents")` → "feat/web-based-agents"

### Function 6: showError

**Purpose**: Display an error message to the user in a formatted way.

**Signature**:
```go
func showError(message string) error
```

**Implementation**:
- Create a huh form with a Note field displaying the error
- Run the form (user presses Enter to acknowledge)
- Return nil (error is shown, not propagated)

**Example Usage**:
```go
if isProtectedBranch(branchName) {
    return showError("Cannot use protected branch name: " + branchName)
}
```

**UI Example**:
```
╔══════════════════════════════════════════════════════════╗
║                         Error                            ║
╚══════════════════════════════════════════════════════════╝

Cannot use protected branch name: main

[Press Enter to continue]
```

### Function 7: withSpinner

**Purpose**: Wrap long-running operations with a loading spinner.

**Signature**:
```go
func withSpinner(title string, action func() error) error
```

**Implementation** (from technical design lines 865-881):
```go
func withSpinner(title string, action func() error) error {
    var err error

    _ = spinner.New().
        Title(title).
        Action(func() {
            err = action()
        }).
        Run()

    return err
}
```

**Example Usage**:
```go
err := withSpinner("Fetching GitHub issues...", func() error {
    issues, err = sow.ListIssues(ctx)
    return err
})
```

**UI Example**:
```
⠋ Fetching GitHub issues...
```

## Acceptance Criteria

### Name Normalization
- [ ] `normalizeName()` handles all test cases from requirements
- [ ] Whitespace is trimmed
- [ ] Converts to lowercase
- [ ] Spaces become hyphens
- [ ] Invalid characters removed
- [ ] Consecutive hyphens collapsed
- [ ] Leading/trailing hyphens removed
- [ ] Empty input returns empty string

### Project Type Configuration
- [ ] `projectTypes` map contains all four types
- [ ] Each type has correct prefix (feat/, explore/, design/, breakdown/)
- [ ] Each type has correct description
- [ ] `getTypePrefix()` returns correct prefix for valid types
- [ ] `getTypePrefix()` returns "feat/" for invalid types
- [ ] `getTypeOptions()` returns options in consistent order
- [ ] `getTypeOptions()` includes Cancel option

### Branch Preview
- [ ] `previewBranchName()` combines prefix and normalized name correctly
- [ ] Handles all project types correctly
- [ ] Normalizes name before combining

### Error Display
- [ ] `showError()` displays message in huh form
- [ ] User can acknowledge error with Enter
- [ ] Returns nil (doesn't propagate error)

### Loading Indicator
- [ ] `withSpinner()` displays spinner during action
- [ ] Spinner shows provided title
- [ ] Spinner disappears when action completes
- [ ] Errors from action are propagated correctly

## Relevant Inputs

- `.sow/knowledge/designs/interactive-wizard-technical-implementation.md` - Helper function implementations (lines 523-646, 865-881)
- `.sow/knowledge/designs/interactive-wizard-ux-flow.md` - Type configuration data (lines 129-135)
- `.sow/knowledge/designs/huh-library-verification.md` - How to use huh forms and spinners
- `.sow/project/context/issue-68.md` - Reference implementations (lines 674-728)
- `cli/internal/projects/standard/standard.go` - Example of project type (for understanding the registry)
- `cli/internal/projects/exploration/exploration.go` - Example of project type
- `cli/internal/projects/design/design.go` - Example of project type
- `cli/internal/projects/breakdown/breakdown.go` - Example of project type

## Examples

### Example 1: Name Normalization

```go
func TestNormalization() {
    testCases := []struct {
        input    string
        expected string
    }{
        {"Web Based Agents", "web-based-agents"},
        {"API V2", "api-v2"},
        {"feature--name", "feature-name"},
        {"-leading-trailing-", "leading-trailing"},
        {"With!Invalid@Chars#", "withinvalidchars"},
        {"UPPERCASE", "uppercase"},
        {"  spaces  ", "spaces"},
        {"", ""},
        {"   ", ""},
        {"!!!@@@###", ""},
    }

    for _, tc := range testCases {
        result := normalizeName(tc.input)
        if result != tc.expected {
            t.Errorf("normalizeName(%q) = %q; want %q", tc.input, result, tc.expected)
        }
    }
}
```

### Example 2: Branch Preview

```go
// User input: "Add JWT Auth"
// Selected type: "standard"

preview := previewBranchName("standard", "Add JWT Auth")
// preview = "feat/add-jwt-auth"

// Can show to user before they confirm:
fmt.Printf("Branch will be created: %s\n", preview)
```

### Example 3: Using withSpinner

```go
var issues []*sow.Issue
err := withSpinner("Fetching GitHub issues...", func() error {
    var err error
    issues, err = sow.ListIssues(ctx, "sow")
    return err
})
if err != nil {
    return fmt.Errorf("failed to fetch issues: %w", err)
}
// issues now populated, spinner gone
```

### Example 4: Error Display

```go
func (w *Wizard) handleNameEntry() error {
    var name string

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().
                Title("Enter project name:").
                Value(&name),
        ),
    )

    if err := form.Run(); err != nil {
        // ... handle abort
    }

    normalized := normalizeName(name)
    if normalized == "" {
        return showError("Project name contains no valid characters")
    }

    // Continue with valid name...
}
```

## Dependencies

- **Task 010**: Huh library must be installed
- **Task 030**: Wizard foundation should exist (though helpers are independent)

## Constraints

- **Pure functions**: Name normalization should be pure (no side effects)
- **Consistent formatting**: Follow existing Go code style
- **No external dependencies**: Only use standard library + huh
- **Deterministic**: Same input always produces same output
- **Well-documented**: Each function needs clear godoc comments

## Testing Requirements

### Unit Tests

Create `cli/cmd/project/wizard_helpers_test.go`:

**Test: normalizeName**
```go
func TestNormalizeName(t *testing.T) {
    // Test all examples from requirements
    // Test edge cases: empty, spaces, special chars
    // Test Unicode characters (should be removed)
    // Test very long names (should work)
}

func TestNormalizeName_EdgeCases(t *testing.T) {
    // Empty string
    // Only spaces
    // Only special characters
    // Mixed valid/invalid
}
```

**Test: getTypePrefix**
```go
func TestGetTypePrefix(t *testing.T) {
    // Test all four valid types
    // Test unknown type (should return "feat/")
    // Test empty string (should return "feat/")
}
```

**Test: getTypeOptions**
```go
func TestGetTypeOptions(t *testing.T) {
    // Verify returns 5 options (4 types + cancel)
    // Verify order is consistent
    // Verify each option has correct label and value
}
```

**Test: previewBranchName**
```go
func TestPreviewBranchName(t *testing.T) {
    // Test each project type
    // Test name normalization happens
    // Test prefix + name combination
}
```

**Test: withSpinner**
```go
func TestWithSpinner_PropagatesError(t *testing.T) {
    // Action returns error
    // Verify error is propagated
}

func TestWithSpinner_ReturnsNilOnSuccess(t *testing.T) {
    // Action succeeds
    // Verify nil returned
}
```

**Note**: Testing `showError()` may be difficult due to huh UI. Consider integration testing or mocking.

### Manual Testing

**Test Name Normalization**:
```bash
# Create test program that calls normalizeName() with various inputs
# Verify output matches expected
```

**Test Error Display**:
```bash
# Trigger an error in the wizard
# Verify error displays correctly
# Verify user can acknowledge and continue
```

**Test Spinner**:
```bash
# Trigger a long-running operation
# Verify spinner appears
# Verify spinner shows correct title
# Verify spinner disappears when done
```

## Implementation Notes

### Name Normalization Algorithm

The algorithm must be strict to ensure valid git branch names:
- Git branch names cannot start with `.`
- Git branch names cannot contain `..`, `\`, spaces, `^`, `~`, `:`, `?`, `*`, `[`, `@{`
- We restrict to alphanumeric, hyphens, and underscores for simplicity

### Project Type Configuration

The project types map is hardcoded because:
1. These are the four project types currently supported
2. They're defined in the codebase already (standard, exploration, design, breakdown)
3. The wizard needs consistent branch prefixes
4. Future work could make this dynamic by querying the registry

### Error Display Strategy

We use huh forms for errors to maintain consistency:
- User stays in the terminal UI
- No context switching
- Clear visual formatting
- User explicitly acknowledges (Enter key)

### Spinner Usage Guidelines

From the design:
- **Use** for operations > 500ms (GitHub API, git operations, file scans)
- **Don't use** for user input (they control timing)
- **Don't use** for quick operations (< 500ms)

### Why These Helpers?

Each helper solves a specific problem:
- **normalizeName**: Users type friendly names, git needs valid branch names
- **getTypePrefix**: Project types determine branch prefix conventions
- **previewBranchName**: Users should see final branch name before committing
- **showError**: Errors need consistent, user-friendly display
- **withSpinner**: Long operations need visual feedback

## Success Indicators

After completing this task:
1. All helper functions implemented and tested
2. Name normalization handles all edge cases correctly
3. Project type configuration is complete and accessible
4. Error display works consistently across wizard
5. Loading indicators work for long operations
6. Helper functions are pure and reusable
7. Test coverage > 90% for helper functions
8. Documentation is clear and complete
