# Integrate Selected Files into Project Initialization

## Context

This task integrates the selected knowledge files from the wizard into the project initialization process. When users select knowledge files during project creation, those files need to be registered as artifacts in the project state so the orchestrator can use them as context.

The project initialization happens in `shared.go` where `initializeProject` creates a new project in a worktree. This function currently handles GitHub issue artifacts (if an issue is provided). This task extends it to also handle knowledge file artifacts.

The overall project goal is to enable knowledge file selection during project creation. This specific task focuses on:
1. Modifying `initializeProject` to accept knowledge files parameter
2. Creating artifact entries for knowledge files
3. Registering artifacts in the appropriate phase
4. Passing selected files from wizard to initialization
5. Ensuring proper relative paths and metadata

This builds on Task 020 which captures the user's file selections in the wizard state.

## Requirements

### 1. Modify initializeProject Function Signature

In `cli/cmd/project/shared.go`:
- Add a new parameter `knowledgeFiles []string` to the `initializeProject` function
- Update the function signature to:
  ```go
  func initializeProject(
      ctx *sow.Context,
      branch string,
      description string,
      issue *sow.Issue,
      knowledgeFiles []string,  // NEW parameter
  ) (*state.Project, error)
  ```
- The parameter should be last to maintain backward compatibility with existing patterns

### 2. Implement Knowledge File Artifact Creation

In the `initializeProject` function body:
- After handling issue artifacts (existing code), add logic to process knowledge files
- For each knowledge file:
  - Create an `ArtifactState` with type `"reference"`
  - Set path relative to project directory: `../../knowledge/<filename>`
  - Mark as approved (auto-approved like issue artifacts)
  - Add metadata: source as "user_selected", description as "Knowledge file selected during project creation"
- Add all knowledge file artifacts to the appropriate phase inputs

### 3. Determine Target Phase for Knowledge Inputs

Implement logic to determine which phase should receive knowledge file inputs:
- For standard projects: use "implementation" phase (first phase)
- For other project types: use "implementation" as default
- This can be made more sophisticated in future enhancements

Create a helper function:
```go
func determineKnowledgeInputPhase(projectType string) string
```

### 4. Update finalizeCreation to Pass Knowledge Files

In `cli/cmd/project/wizard_state.go`, update the `finalizeCreation` method:
- Extract knowledge files from `w.choices["knowledge_files"]`
- Handle type assertion safely (cast to `[]string`)
- Pass knowledge files to `initializeProject` call
- Handle case where key doesn't exist or is empty (pass empty slice)

### 5. Handle Edge Cases

Implement proper handling for:
- Empty knowledge files slice (no selection): valid, skip artifact creation
- Knowledge files key missing from choices: treat as empty selection
- Invalid type in choices: log warning, treat as empty selection
- Relative path construction: ensure paths work from project directory to knowledge files

## Acceptance Criteria

### Functional Requirements

1. **Function signature updated**:
   - `initializeProject` accepts `knowledgeFiles []string` parameter
   - All callers updated to pass the new parameter
   - Existing behavior preserved when empty slice passed

2. **Artifact creation works**:
   - Each selected file becomes an `ArtifactState`
   - Artifact type is "reference" (knowledge files are references, not generated)
   - Paths are correct relative to project directory
   - All artifacts marked as approved
   - Metadata includes source and description

3. **Integration with project state**:
   - Artifacts added to correct phase inputs
   - Phase determined by helper function
   - Multiple artifacts supported (can select many files)
   - Artifacts don't conflict with existing issue artifacts

4. **Wizard integration**:
   - `finalizeCreation` extracts knowledge files from choices
   - Type assertion handled safely
   - Empty selections handled gracefully
   - Knowledge files passed to `initializeProject`

5. **Edge cases handled**:
   - Empty selection (0 files): no artifacts created, no error
   - Missing choices key: treated as empty selection
   - Invalid type in choices: logged and treated as empty
   - Relative path construction works correctly

### Test Requirements (TDD Approach)

Following the project's TDD methodology, write tests FIRST, then implement:

1. **Unit tests** in `shared_test.go` (or create if doesn't exist):
   - Test `initializeProject` with empty knowledge files
   - Test `initializeProject` with single knowledge file
   - Test `initializeProject` with multiple knowledge files
   - Test artifact structure (type, path, approved, metadata)
   - Test phase assignment
   - Test helper function `determineKnowledgeInputPhase`

2. **Integration tests** in `wizard_integration_test.go`:
   - Test end-to-end flow: wizard selection → project initialization
   - Test that selected files appear in project state
   - Test paths are correct
   - Test artifacts are accessible
   - Test with both issue and knowledge files (combined scenario)

3. **Test coverage**:
   - All new code paths must be tested
   - Edge cases must have explicit tests
   - Artifact metadata validation

### Code Quality

- Follow existing artifact creation patterns (see issue artifact creation)
- Add comprehensive godoc comments
- Use existing error handling patterns
- Match code style of surrounding functions
- Ensure no linting errors

## Technical Details

### Artifact Structure for Knowledge Files

Knowledge files should be created as "reference" type artifacts:

```go
artifact := projschema.ArtifactState{
    Type:       "reference", // Knowledge files are references
    Path:       filepath.Join("../../knowledge", file), // Relative to project dir
    Approved:   true, // Auto-approved
    Created_at: time.Now(),
    Metadata: map[string]interface{}{
        "source":      "user_selected",
        "description": "Knowledge file selected during project creation",
    },
}
```

The path `../../knowledge/<file>` works because:
- Artifacts are referenced from `.sow/project/state.yaml`
- Knowledge files are at `.sow/knowledge/<file>`
- So we need to go up two levels: `../../knowledge/<file>`

### Implementation Pattern in initializeProject

```go
func initializeProject(
    ctx *sow.Context,
    branch string,
    description string,
    issue *sow.Issue,
    knowledgeFiles []string,
) (*state.Project, error) {
    // ... existing setup code ...

    // Prepare initial inputs
    var initialInputs map[string][]projschema.ArtifactState

    // Add issue input if provided (EXISTING CODE)
    if issue != nil {
        // ... existing issue handling ...
        initialInputs = map[string][]projschema.ArtifactState{
            "implementation": {issueArtifact},
        }
    }

    // NEW: Add knowledge file inputs
    if len(knowledgeFiles) > 0 {
        if initialInputs == nil {
            initialInputs = make(map[string][]projschema.ArtifactState)
        }

        knowledgeArtifacts := make([]projschema.ArtifactState, 0, len(knowledgeFiles))
        for _, file := range knowledgeFiles {
            artifact := projschema.ArtifactState{
                Type:       "reference",
                Path:       filepath.Join("../../knowledge", file),
                Approved:   true,
                Created_at: time.Now(),
                Metadata: map[string]interface{}{
                    "source":      "user_selected",
                    "description": "Knowledge file selected during project creation",
                },
            }
            knowledgeArtifacts = append(knowledgeArtifacts, artifact)
        }

        // Determine target phase
        targetPhase := determineKnowledgeInputPhase("standard") // TODO: get from project type

        if _, exists := initialInputs[targetPhase]; !exists {
            initialInputs[targetPhase] = []projschema.ArtifactState{}
        }
        initialInputs[targetPhase] = append(initialInputs[targetPhase], knowledgeArtifacts...)
    }

    // ... existing project creation code ...
    proj, err := state.Create(ctx, branch, description, initialInputs)
    // ... rest of function ...
}

// Helper function to determine which phase gets knowledge inputs
func determineKnowledgeInputPhase(projectType string) string {
    // For now, use "implementation" as default (first phase for standard projects)
    // This could be enhanced to detect project type and use appropriate phase
    return "implementation"
}
```

### Update in finalizeCreation

```go
func (w *Wizard) finalizeCreation() error {
    // Extract wizard choices
    name, ok := w.choices["name"].(string)
    if !ok {
        return fmt.Errorf("name choice not set or invalid")
    }
    branch, ok := w.choices["branch"].(string)
    if !ok {
        return fmt.Errorf("branch choice not set or invalid")
    }
    initialPrompt := ""
    if prompt, ok := w.choices["prompt"].(string); ok {
        initialPrompt = prompt
    }

    // Extract issue if present (EXISTING CODE)
    var issue *sow.Issue
    if issueData, ok := w.choices["issue"].(*sow.Issue); ok {
        issue = issueData
    }

    // NEW: Extract knowledge files
    var knowledgeFiles []string
    if files, ok := w.choices["knowledge_files"].([]string); ok {
        knowledgeFiles = files
    }

    // ... existing worktree setup ...

    // Initialize project WITH knowledge files
    project, err := initializeProject(worktreeCtx, branch, name, issue, knowledgeFiles)
    if err != nil {
        return fmt.Errorf("failed to initialize project: %w", err)
    }

    // ... rest of existing logic ...
}
```

### Test Examples

```go
func TestInitializeProject_WithKnowledgeFiles(t *testing.T) {
    ctx := setupTestContext(t)
    knowledgeFiles := []string{
        "designs/api.md",
        "adrs/001-architecture.md",
    }

    proj, err := initializeProject(ctx, "feat/test", "Test Project", nil, knowledgeFiles)

    require.NoError(t, err)
    require.NotNil(t, proj)

    // Verify artifacts were created
    phase := proj.Phases["implementation"]
    require.Len(t, phase.Inputs, 2)

    // Verify first artifact
    artifact := phase.Inputs[0]
    assert.Equal(t, "reference", artifact.Type)
    assert.Equal(t, "../../knowledge/designs/api.md", artifact.Path)
    assert.True(t, artifact.Approved)
    assert.Equal(t, "user_selected", artifact.Metadata["source"])
}

func TestInitializeProject_EmptyKnowledgeFiles(t *testing.T) {
    ctx := setupTestContext(t)

    proj, err := initializeProject(ctx, "feat/test", "Test Project", nil, []string{})

    require.NoError(t, err)
    require.NotNil(t, proj)

    // Verify no artifacts were created
    phase := proj.Phases["implementation"]
    assert.Len(t, phase.Inputs, 0)
}

func TestInitializeProject_WithIssueAndKnowledgeFiles(t *testing.T) {
    ctx := setupTestContext(t)
    issue := &sow.Issue{Number: 123, Title: "Test", Body: "Test issue"}
    knowledgeFiles := []string{"designs/api.md"}

    proj, err := initializeProject(ctx, "feat/test", "Test", issue, knowledgeFiles)

    require.NoError(t, err)

    // Verify both issue and knowledge artifacts exist
    phase := proj.Phases["implementation"]
    assert.Len(t, phase.Inputs, 2) // 1 issue + 1 knowledge file

    // Find artifacts by type
    hasIssue := false
    hasKnowledge := false
    for _, artifact := range phase.Inputs {
        if artifact.Type == "github_issue" {
            hasIssue = true
        }
        if artifact.Type == "reference" {
            hasKnowledge = true
        }
    }
    assert.True(t, hasIssue, "Should have issue artifact")
    assert.True(t, hasKnowledge, "Should have knowledge artifact")
}
```

## Relevant Inputs

- `/Users/josh/code/sow/.sow/worktrees/feat/wizard-select-files/cli/cmd/project/shared.go` - Contains initializeProject function to modify
- `/Users/josh/code/sow/.sow/worktrees/feat/wizard-select-files/cli/cmd/project/wizard_state.go` - Contains finalizeCreation to update
- `/Users/josh/code/sow/.sow/worktrees/feat/wizard-select-files/cli/schemas/project/cue_types_gen.go` - Contains ArtifactState structure definition
- `/Users/josh/code/sow/.sow/worktrees/feat/wizard-select-files/.sow/knowledge/designs/file-selector-wizard.md` - Design document with artifact specifications

## Examples

### Artifact Metadata Structure

```go
Metadata: map[string]interface{}{
    "source":      "user_selected",
    "description": "Knowledge file selected during project creation",
}
```

### Relative Path Construction

From `.sow/project/state.yaml` perspective:
```
.sow/
├── knowledge/
│   ├── designs/
│   │   └── api.md          ← Target file
│   └── adrs/
│       └── 001-arch.md
└── project/
    └── state.yaml          ← Reference point
```

Path from state.yaml to api.md: `../../knowledge/designs/api.md`

### Combined Artifacts (Issue + Knowledge)

When both issue and knowledge files are provided:

```go
initialInputs = map[string][]projschema.ArtifactState{
    "implementation": {
        // Issue artifact
        {
            Type: "github_issue",
            Path: "context/issue-123.md",
            Approved: true,
            Metadata: {...},
        },
        // Knowledge file artifacts
        {
            Type: "reference",
            Path: "../../knowledge/designs/api.md",
            Approved: true,
            Metadata: {...},
        },
        {
            Type: "reference",
            Path: "../../knowledge/adrs/001-arch.md",
            Approved: true,
            Metadata: {...},
        },
    },
}
```

## Dependencies

**Depends on Task 020**:
- Requires `handleFileSelect` to capture selections in `w.choices["knowledge_files"]`
- Requires selections to be stored as `[]string`

## Constraints

### Path Handling

- All paths must be relative to `.sow/project/` directory
- Use `filepath.Join` for cross-platform compatibility
- Path structure: `../../knowledge/<relative-file-path>`
- Paths must work when project state is loaded

### Artifact Structure

- Type must be "reference" (not "input" or "output")
- Approved must be true (auto-approved)
- Created_at must be set to current time
- Metadata must include source and description

### Backward Compatibility

- Empty knowledge files parameter should work (no artifacts created)
- Existing projects without knowledge files should continue to work
- Issue artifacts should not be affected by knowledge file additions

### Error Handling

- Invalid type assertions should be logged but not fail
- Empty selections should be handled gracefully
- Path construction errors should be reported clearly

## Notes

- Artifact type "reference" indicates the file is not generated by the project
- The orchestrator will be able to access these files as task inputs
- Multiple knowledge files can be selected and all will be registered
- This task completes the integration - no UI changes needed after this
- Future enhancement: support per-phase knowledge file selection
