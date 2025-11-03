# Task 060: Two-Tier Validation Implementation

# Task 060: Two-Tier Validation Implementation

## Objective

Implement the two-tier validation system: CUE validates structure (already done in unit-002), and runtime validates metadata against embedded schemas. This enables project types to define custom metadata schemas that are validated on Save().

## Context

**Design Reference:** See `.sow/knowledge/designs/project-sdk-implementation.md`:
- Section "Validation Implementation" (lines 764-831) for complete implementation
- Section "Cross-Cutting Concerns - Validation Strategy" (lines 1273-1294) for validation approach
- Section "Metadata Schema Management" (lines 1295-1340) for schema embedding patterns

**Two-Tier Validation:**
1. **Structure Validation (CUE)** - Already implemented in unit-002, validates universal fields
2. **Metadata Validation (Runtime CUE)** - This task, validates project-type-specific metadata

**Existing Code:**
- `cli/internal/sdks/project/state/registry.go` has stub `Validate()` that returns nil
- `cli/internal/sdks/project/state/validate.go` has `validateStructure()` for CUE validation
- Structure validation already called in Load() and Save()

**Prerequisite:** Task 020 completed (ProjectTypeConfig with phaseConfigs)

**What This Task Builds:**
Runtime validation of phase metadata against embedded CUE schemas defined in project type configs.

## Requirements

### 1. Validate Method on ProjectTypeConfig

Modify `cli/internal/sdks/project/state/registry.go`:

Replace stub:
```go
func (ptc *ProjectTypeConfig) Validate(_ *Project) error {
    return nil
}
```

With full implementation:
```go
// Validate validates project state against project type configuration.
//
// Performs two-tier validation:
//  1. Artifact type validation - Checks inputs/outputs against allowed types
//  2. Metadata validation - Validates metadata against embedded CUE schemas
//
// Returns error describing first validation failure found.
func (ptc *ProjectTypeConfig) Validate(project *Project) error {
    // Validate each phase
    for phaseName, phaseConfig := range ptc.phaseConfigs {
        phase, exists := project.Phases[phaseName]
        if !exists {
            // Phase not in state - skip (may be optional/future phase)
            continue
        }

        // Validate artifact types
        if err := validateArtifactTypes(
            phase.Inputs,
            phaseConfig.allowedInputTypes,
            phaseName,
            "input",
        ); err != nil {
            return err
        }

        if err := validateArtifactTypes(
            phase.Outputs,
            phaseConfig.allowedOutputTypes,
            phaseName,
            "output",
        ); err != nil {
            return err
        }

        // Validate metadata against embedded schema
        if phaseConfig.metadataSchema != "" {
            if err := validateMetadata(
                phase.Metadata,
                phaseConfig.metadataSchema,
            ); err != nil {
                return fmt.Errorf("phase %s metadata: %w", phaseName, err)
            }
        } else if len(phase.Metadata) > 0 {
            return fmt.Errorf("phase %s does not support metadata", phaseName)
        }
    }

    return nil
}
```

### 2. Artifact Type Validation

Add to `cli/internal/sdks/project/state/validate.go`:

```go
// validateArtifactTypes checks if artifact types are in the allowed list.
// Empty allowed list means "allow all types" (no validation).
func validateArtifactTypes(
    artifacts []Artifact,
    allowedTypes []string,
    phaseName string,
    category string, // "input" or "output"
) error {
    // Empty allowed list = allow all
    if len(allowedTypes) == 0 {
        return nil
    }

    // Build set for O(1) lookup
    allowed := make(map[string]bool)
    for _, t := range allowedTypes {
        allowed[t] = true
    }

    // Check each artifact
    for _, artifact := range artifacts {
        if !allowed[artifact.Type] {
            return fmt.Errorf(
                "phase %s: %s artifact type %q not allowed (allowed: %v)",
                phaseName,
                category,
                artifact.Type,
                allowedTypes,
            )
        }
    }

    return nil
}
```

### 3. Metadata Schema Validation

Add to `cli/internal/sdks/project/state/validate.go`:

```go
// validateMetadata validates metadata against an embedded CUE schema.
// Returns error if schema is invalid or metadata doesn't conform.
func validateMetadata(metadata map[string]interface{}, cueSchema string) error {
    ctx := cuecontext.New()

    // Compile embedded schema
    schema := ctx.CompileString(cueSchema)
    if schema.Err() != nil {
        return fmt.Errorf("invalid schema: %w", schema.Err())
    }

    // Encode metadata
    value := ctx.Encode(metadata)

    // Unify and validate
    result := schema.Unify(value)
    if err := result.Validate(cue.Concrete(true)); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    return nil
}
```

Required imports for validate.go:
```go
import (
    "fmt"
    "cuelang.org/go/cue"
    "cuelang.org/go/cue/cuecontext"
)
```

## Files to Modify

1. `cli/internal/sdks/project/state/registry.go` - Implement Validate() method
2. `cli/internal/sdks/project/state/validate.go` - Add helper functions
3. `cli/internal/sdks/project/state/validate_test.go` - Add metadata validation tests

## Testing Requirements (TDD)

Extend `cli/internal/sdks/project/state/validate_test.go`:

**Artifact Type Validation Tests:**
- validateArtifactTypes() allows artifacts when type in allowed list
- validateArtifactTypes() rejects artifacts when type not in allowed list
- validateArtifactTypes() allows all types when allowed list empty
- validateArtifactTypes() error includes phase name, category, and invalid type
- Multiple artifacts validated correctly

**Metadata Validation Tests:**
- validateMetadata() passes for valid metadata
- validateMetadata() fails for invalid metadata
- validateMetadata() error includes validation details
- validateMetadata() handles schema compilation errors
- Empty metadata passes validation

**Validate Method Tests:**
- Validate() validates all phases
- Validate() skips phases not in project state
- Validate() validates input artifact types
- Validate() validates output artifact types
- Validate() validates metadata when schema present
- Validate() rejects metadata when no schema defined
- Validate() passes when metadata absent and no schema
- Validate() error includes phase name
- Multiple validation errors (first one returned)

**Test Pattern - Artifact Type Validation:**
```go
func TestValidateArtifactTypesAllowed(t *testing.T) {
    artifacts := []Artifact{
        {Type: "task_list"},
        {Type: "design"},
    }
    allowed := []string{"task_list", "design", "review"}

    err := validateArtifactTypes(artifacts, allowed, "planning", "output")
    if err != nil {
        t.Errorf("expected no error, got %v", err)
    }
}

func TestValidateArtifactTypesRejects(t *testing.T) {
    artifacts := []Artifact{
        {Type: "invalid_type"},
    }
    allowed := []string{"task_list", "design"}

    err := validateArtifactTypes(artifacts, allowed, "planning", "output")
    if err == nil {
        t.Error("expected error for invalid artifact type")
    }
    if !strings.Contains(err.Error(), "invalid_type") {
        t.Errorf("error should mention invalid type, got: %v", err)
    }
}
```

**Test Pattern - Metadata Validation:**
```go
func TestValidateMetadataValid(t *testing.T) {
    schema := `{
        tasks_approved?: bool
        complexity?: "low" | "medium" | "high"
    }`

    metadata := map[string]interface{}{
        "tasks_approved": true,
        "complexity":     "high",
    }

    err := validateMetadata(metadata, schema)
    if err != nil {
        t.Errorf("expected valid metadata to pass, got %v", err)
    }
}

func TestValidateMetadataInvalid(t *testing.T) {
    schema := `{
        complexity: "low" | "medium" | "high"
    }`

    metadata := map[string]interface{}{
        "complexity": "invalid", // Not in enum
    }

    err := validateMetadata(metadata, schema)
    if err == nil {
        t.Error("expected error for invalid metadata")
    }
}
```

**Integration Test:**
```go
func TestValidateFullProjectType(t *testing.T) {
    config := &ProjectTypeConfig{
        phaseConfigs: map[string]*PhaseConfig{
            "planning": {
                allowedOutputTypes: []string{"task_list"},
                metadataSchema: `{
                    complexity?: "low" | "medium" | "high"
                }`,
            },
        },
    }

    project := &Project{
        ProjectState: schemas.ProjectState{
            Phases: map[string]schemas.PhaseState{
                "planning": {
                    Outputs: []schemas.ArtifactState{
                        {Type: "task_list"},
                    },
                    Metadata: map[string]interface{}{
                        "complexity": "medium",
                    },
                },
            },
        },
    }

    err := config.Validate(project)
    if err != nil {
        t.Errorf("expected valid project to pass validation, got %v", err)
    }
}
```

## Acceptance Criteria

- [ ] Validate() iterates over all phase configs
- [ ] Validate() skips phases not in project state
- [ ] validateArtifactTypes() checks inputs against allowed input types
- [ ] validateArtifactTypes() checks outputs against allowed output types
- [ ] Empty allowed types list means "allow all" (no validation)
- [ ] validateMetadata() compiles embedded CUE schema
- [ ] validateMetadata() encodes metadata and validates against schema
- [ ] Validate() rejects metadata when no schema defined for phase
- [ ] Validate() allows missing metadata when no schema defined
- [ ] Validation errors include phase name and clear message
- [ ] Schema compilation errors reported clearly
- [ ] All tests pass (100% coverage of validation behavior)
- [ ] Code compiles without errors

## Dependencies

**Required:** Task 020 (ProjectTypeConfig with phaseConfigs)

## Technical Notes

- CUE imports: `cuelang.org/go/cue` and `cuelang.org/go/cue/cuecontext`
- Use `cue.Concrete(true)` for strict validation (all values must be concrete)
- Empty allowed types slice = allow all types (common for flexible phases)
- Metadata validation is optional (only if schema present)
- Validation happens on Save() and Load() (already wired in unit-002)
- First error stops validation (fail-fast approach)

**CUE Validation Pattern:**
```go
ctx := cuecontext.New()
schema := ctx.CompileString(schemaString)
value := ctx.Encode(data)
result := schema.Unify(value)
err := result.Validate(cue.Concrete(true))
```

**Example Metadata Schema:**
```cue
{
    tasks_approved?: bool
    complexity?: "low" | "medium" | "high"
    notes?: string
}
```

The `?` suffix means optional fields. Without it, fields are required.

## Estimated Time

2.5 hours
