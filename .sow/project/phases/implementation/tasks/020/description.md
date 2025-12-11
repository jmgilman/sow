# Create Validation Function and Comprehensive Tests

## Context

This task is part of Work Unit 002, which creates the schema and validation infrastructure for `.sow-ref.yaml` manifest files. The previous task (010) created the CUE schema at `libs/schemas/ref_manifest.cue` and generated Go types.

This task implements:
1. A validation function that validates Go structs against the CUE schema
2. Comprehensive unit tests covering valid and invalid manifests
3. Tests for all validation constraints (regex patterns, enum values, required fields)

The validation function will be consumed by future work units:
- Work Unit 004 (Packaging): Validates `.sow-ref.yaml` before OCI push
- Work Unit 005 (Inspection): Validates remote manifest before displaying
- Work Unit 006 (Installation): Validates manifest after extraction

## Requirements

### Validation Function

Create a validation function in `libs/schemas/validate_ref_manifest.go` that:

1. **Loads the RefManifest schema** from the embedded CUE schemas
2. **Validates a RefManifest struct** against the schema
3. **Returns detailed error messages** for validation failures

**Function signature:**
```go
// ValidateRefManifest validates a RefManifest against the CUE schema.
// It returns an error if the manifest is invalid, with details about
// which field(s) failed validation and why.
func ValidateRefManifest(manifest *RefManifest) error
```

**Implementation approach (following existing patterns):**

1. Use `sync.Once` or package-level `init()` to load schemas once:
   ```go
   var refManifestSchema cue.Value

   func init() {
       // Load schema from embedded FS
       memFS := billy.NewMemory()
       if err := core.CopyFromEmbedFS(CUESchemas, memFS, "."); err != nil {
           panic(fmt.Errorf("failed to copy embedded schemas: %w", err))
       }
       loader := cuepkg.NewLoader(memFS)
       schemas, err := loader.LoadPackage(context.Background(), ".")
       if err != nil {
           panic(fmt.Errorf("failed to load schemas: %w", err))
       }
       refManifestSchema = schemas.LookupPath(cue.ParsePath("#RefManifest"))
       if refManifestSchema.Err() != nil {
           panic(fmt.Errorf("RefManifest schema not found: %w", refManifestSchema.Err()))
       }
   }
   ```

2. Encode the input struct to CUE value
3. Unify with the schema
4. Validate with `cue.Concrete(true)`
5. Return wrapped error with field context

**Error handling:**
- Define a sentinel error: `var ErrRefManifestValidation = errors.New("ref manifest validation failed")`
- Wrap CUE errors with context: `fmt.Errorf("%w: %w", ErrRefManifestValidation, err)`
- Include field paths in error messages when possible

### Test File

Create comprehensive tests in `libs/schemas/validate_ref_manifest_test.go`:

**Test categories:**

1. **Valid manifest tests:**
   - Minimal valid manifest (only required fields)
   - Full manifest with all optional sections
   - All 11 classification types individually
   - Edge cases: various valid link formats

2. **Invalid manifest tests - missing required fields:**
   - Missing `schema_version`
   - Missing `ref` section
   - Missing `ref.title`
   - Missing `ref.link`
   - Missing `content` section
   - Missing `content.description`
   - Empty `classifications` array
   - Empty `tags` array

3. **Invalid manifest tests - format violations:**
   - Invalid `schema_version` format (not semver)
   - Invalid `link` format - uppercase letters
   - Invalid `link` format - starts with hyphen
   - Invalid `link` format - ends with hyphen
   - Invalid `link` format - special characters
   - Invalid `link` format - spaces
   - Invalid classification type (typo like "guidlines")

4. **Error message quality tests:**
   - Verify error contains field path
   - Verify error wraps `ErrRefManifestValidation`

### Test Patterns

Follow the existing test patterns from `libs/schemas/project/schemas_test.go` and `libs/project/state/validate_test.go`:

- Use `testify/assert` and `testify/require`
- Use table-driven tests for multiple cases
- Use descriptive test names
- Test both positive (valid) and negative (invalid) cases

## Acceptance Criteria

1. `libs/schemas/validate_ref_manifest.go` exists with:
   - `ErrRefManifestValidation` sentinel error
   - `ValidateRefManifest(*RefManifest) error` function
   - Proper doc comments

2. `libs/schemas/validate_ref_manifest_test.go` exists with comprehensive tests

3. All tests pass: `go test ./libs/schemas/...`

4. Tests cover all validation constraints:
   - `schema_version` semver format
   - `ref.link` kebab-case regex
   - `content.classifications` non-empty array
   - `content.tags` non-empty array
   - All 11 classification type enum values
   - Required vs optional field handling

5. Code passes linting: `golangci-lint run ./libs/schemas/...`

**TDD Approach:**
- Write test cases for valid manifests first
- Write test cases for each invalid scenario
- Implement validation function to make tests pass
- Refine error messages based on test assertions

## Technical Details

### Existing Validation Pattern Reference

From `libs/project/state/validate.go`:

```go
var projectSchemas cue.Value

func init() {
    memFS := billy.NewMemory()
    if err := core.CopyFromEmbedFS(schemas.CUESchemas, memFS, "."); err != nil {
        panic(fmt.Errorf("failed to copy embedded schemas: %w", err))
    }
    loader := cuepkg.NewLoader(memFS)
    var err error
    projectSchemas, err = loader.LoadPackage(context.Background(), "project")
    if err != nil {
        panic(fmt.Errorf("failed to load project schemas: %w", err))
    }
}

func validateStructure(projectState *project.ProjectState) error {
    ctx := cuecontext.New()
    schema := projectSchemas.LookupPath(cue.ParsePath("#ProjectState"))
    if schema.Err() != nil {
        return fmt.Errorf("ProjectState schema not found: %w", schema.Err())
    }
    value := ctx.Encode(projectState)
    if value.Err() != nil {
        return fmt.Errorf("failed to encode project state: %w", value.Err())
    }
    result := schema.Unify(value)
    if err := result.Validate(cue.Concrete(true)); err != nil {
        return fmt.Errorf("%w: %w", ErrValidationFailed, err)
    }
    return nil
}
```

### Test Pattern Reference

From `libs/project/state/validate_test.go`:

```go
func TestValidateStructure_ValidState(t *testing.T) {
    state := &project.ProjectState{
        // ... valid state ...
    }
    err := validateStructure(state)
    assert.NoError(t, err)
}

func TestValidateStructure_MissingRequiredFields(t *testing.T) {
    tests := []struct {
        name  string
        state *project.ProjectState
    }{
        {
            name: "empty name",
            state: &project.ProjectState{
                Name: "",
                // ... other fields ...
            },
        },
        // ... more cases ...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateStructure(tt.state)
            assert.Error(t, err)
            assert.True(t, errors.Is(err, ErrValidationFailed))
        })
    }
}
```

### Required Imports

```go
import (
    "context"
    "errors"
    "fmt"
    "sync"

    "cuelang.org/go/cue"
    "cuelang.org/go/cue/cuecontext"
    cuepkg "github.com/jmgilman/go/cue"
    "github.com/jmgilman/go/fs/billy"
    "github.com/jmgilman/go/fs/core"
)
```

For tests:
```go
import (
    "errors"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)
```

### File Organization

```
libs/schemas/
├── ref_manifest.cue              # Created in task 010
├── validate_ref_manifest.go      # New file (this task)
├── validate_ref_manifest_test.go # New file (this task)
├── cue_types_gen.go              # Updated by task 010
└── ... existing files ...
```

## Relevant Inputs

- `libs/schemas/ref_manifest.cue` - The CUE schema to validate against (from task 010)
- `libs/schemas/cue_types_gen.go` - Generated Go types including RefManifest
- `libs/project/state/validate.go` - Existing validation pattern to follow
- `libs/project/state/validate_test.go` - Existing test patterns
- `libs/schemas/project/schemas_test.go` - Schema test patterns
- `.standards/TESTING.md` - Test conventions
- `.standards/STYLE.md` - Code style conventions
- `.sow/project/context/issue-124.md` - Full requirements including test cases

## Examples

### Valid Manifest for Testing

```go
func validMinimalManifest() *RefManifest {
    return &RefManifest{
        Schema_version: "1.0.0",
        Ref: RefIdentification{
            Title: "Go Team Standards",
            Link:  "go-standards",
        },
        Content: RefContent{
            Description: "Team Go coding conventions.",
            Classifications: []RefClassification{
                {Type: "guidelines"},
            },
            Tags: []string{"golang"},
        },
    }
}

func validFullManifest() *RefManifest {
    return &RefManifest{
        Schema_version: "1.0.0",
        Ref: RefIdentification{
            Title: "Go Team Standards",
            Link:  "go-standards",
        },
        Content: RefContent{
            Description: "Team Go coding conventions and best practices.",
            Summary:     stringPtr("Complete reference for Go development."),
            Classifications: []RefClassification{
                {Type: "guidelines", Description: stringPtr("Coding standards")},
                {Type: "code-examples"},
            },
            Tags: []string{"golang", "conventions", "testing"},
        },
        Provenance: &RefProvenance{
            Authors: []string{"Platform Team"},
            Source:  stringPtr("https://github.com/myorg/team-docs"),
            License: stringPtr("MIT"),
        },
        Packaging: &RefPackaging{
            Exclude: []string{"*.draft.md", ".DS_Store"},
        },
        Hints: &RefHints{
            Suggested_queries: []string{"error handling patterns"},
            Primary_files:     []string{"README.md"},
        },
        Metadata: map[string]any{
            "team": "platform",
        },
    }
}
```

### Invalid Manifest Examples

```go
// Missing required field
invalidMissingLink := &RefManifest{
    Schema_version: "1.0.0",
    Ref: RefIdentification{
        Title: "Test",
        // Link is missing
    },
    Content: RefContent{...},
}

// Invalid link format
invalidLinkFormat := &RefManifest{
    Schema_version: "1.0.0",
    Ref: RefIdentification{
        Title: "Test",
        Link:  "Go-Standards", // Uppercase not allowed
    },
    Content: RefContent{...},
}

// Invalid classification type
invalidClassificationType := &RefManifest{
    Schema_version: "1.0.0",
    Ref: RefIdentification{...},
    Content: RefContent{
        Description: "Test",
        Classifications: []RefClassification{
            {Type: "guidlines"}, // Typo
        },
        Tags: []string{"test"},
    },
}
```

### Test Structure Example

```go
func TestValidateRefManifest_ValidCases(t *testing.T) {
    tests := []struct {
        name     string
        manifest *RefManifest
    }{
        {
            name:     "minimal valid manifest",
            manifest: validMinimalManifest(),
        },
        {
            name:     "full valid manifest",
            manifest: validFullManifest(),
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateRefManifest(tt.manifest)
            assert.NoError(t, err)
        })
    }
}

func TestValidateRefManifest_AllClassificationTypes(t *testing.T) {
    types := []string{
        "tutorial", "api-reference", "guidelines", "architecture",
        "runbook", "specification", "reference", "code-examples",
        "code-templates", "code-library", "uncategorized",
    }

    for _, typ := range types {
        t.Run(typ, func(t *testing.T) {
            manifest := validMinimalManifest()
            manifest.Content.Classifications = []RefClassification{
                {Type: typ},
            }
            err := ValidateRefManifest(manifest)
            assert.NoError(t, err, "classification type %q should be valid", typ)
        })
    }
}

func TestValidateRefManifest_InvalidLinkFormats(t *testing.T) {
    tests := []struct {
        name string
        link string
    }{
        {"uppercase", "Go-Standards"},
        {"starts with hyphen", "-go-standards"},
        {"ends with hyphen", "go-standards-"},
        {"spaces", "go standards"},
        {"special chars", "go_standards"},
        {"empty", ""},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            manifest := validMinimalManifest()
            manifest.Ref.Link = tt.link
            err := ValidateRefManifest(manifest)
            require.Error(t, err)
            assert.True(t, errors.Is(err, ErrRefManifestValidation))
        })
    }
}
```

## Dependencies

- Task 010 must be completed first (CUE schema and generated types)
- The following imports are required:
  - `cuelang.org/go/cue`
  - `cuelang.org/go/cue/cuecontext`
  - `github.com/jmgilman/go/cue`
  - `github.com/jmgilman/go/fs/billy`
  - `github.com/jmgilman/go/fs/core`

## Constraints

- Follow existing validation patterns in `libs/project/state/validate.go`
- Use `testify/assert` and `testify/require` for assertions
- Use table-driven tests as per TESTING.md
- All tests must pass with race detector: `go test -race ./libs/schemas/...`
- Code must pass `golangci-lint run ./libs/schemas/...`
- Do not log errors - return them with context
- Schema loading should happen once at init time
- Keep validation function in same package as schema (not a subpackage)
