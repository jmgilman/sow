package state

import (
	"context"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cuepkg "github.com/jmgilman/go/cue"
	"github.com/jmgilman/go/fs/billy"
	"github.com/jmgilman/go/fs/core"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/jmgilman/sow/cli/schemas/project"
)

// projectSchemas holds the loaded CUE value for the project subpackage.
// It is loaded once at package initialization using the embedded schemas.
var projectSchemas cue.Value

// init loads the project subpackage schemas from the embedded filesystem.
// This uses the same embedded schemas as cli/schemas/validator.go for consistency.
func init() {
	// Create in-memory filesystem
	memFS := billy.NewMemory()

	// Copy embedded schemas to in-memory filesystem
	if err := core.CopyFromEmbedFS(schemas.CUESchemas, memFS, "."); err != nil {
		panic(fmt.Errorf("failed to copy embedded schemas: %w", err))
	}

	// Create loader and load the project subpackage
	loader := cuepkg.NewLoader(memFS)
	var err error
	projectSchemas, err = loader.LoadPackage(context.Background(), "project")
	if err != nil {
		panic(fmt.Errorf("failed to load project schemas: %w", err))
	}
}

// validateStructure performs CUE-based structural validation on a ProjectState.
// It validates universal fields (name, type, status, etc.), regex patterns,
// required fields, and collection structures.
//
// This validation is run on both Load() and Save() operations to ensure
// structural integrity of the project state.
func validateStructure(projectState project.ProjectState) error {
	ctx := cuecontext.New()

	// Lookup the ProjectState schema from the loaded package
	schema := projectSchemas.LookupPath(cue.ParsePath("#ProjectState"))
	if schema.Err() != nil {
		return fmt.Errorf("ProjectState schema not found: %w", schema.Err())
	}

	// Encode the project state to a CUE value
	value := ctx.Encode(projectState)
	if value.Err() != nil {
		return fmt.Errorf("failed to encode project state: %w", value.Err())
	}

	// Unify the schema with the value and validate
	result := schema.Unify(value)
	if err := result.Validate(cue.Concrete(true)); err != nil {
		return fmt.Errorf("structural validation failed: %w", err)
	}

	return nil
}

// validateMetadata validates a metadata map against a CUE schema string.
// This is used for project-type-specific metadata validation on phases and tasks.
//
// If cueSchema is empty, validation is skipped (no schema = no validation).
// This validation is only run on Save() operations, not Load(), to allow for
// schema evolution and backward compatibility.
func validateMetadata(metadata map[string]interface{}, cueSchema string) error {
	if cueSchema == "" {
		// No schema provided, skip validation
		return nil
	}

	ctx := cuecontext.New()

	// Compile the metadata schema
	schema := ctx.CompileString(cueSchema)
	if schema.Err() != nil {
		return fmt.Errorf("invalid metadata schema: %w", schema.Err())
	}

	// Encode the metadata map to a CUE value
	value := ctx.Encode(metadata)
	if value.Err() != nil {
		return fmt.Errorf("failed to encode metadata: %w", value.Err())
	}

	// Unify and validate
	result := schema.Unify(value)
	if err := result.Validate(cue.Concrete(true)); err != nil {
		return fmt.Errorf("metadata validation failed: %w", err)
	}

	return nil
}

// validateArtifactTypes checks if artifact types are in the allowed list.
// Empty allowed list means "allow all types" (no validation).
func validateArtifactTypes(
	artifacts []project.ArtifactState,
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
