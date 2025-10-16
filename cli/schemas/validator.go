package schemas

import (
	"context"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cueyaml "cuelang.org/go/encoding/yaml"
	cuepkg "github.com/jmgilman/go/cue"
	"github.com/jmgilman/go/fs/billy"
	"github.com/jmgilman/go/fs/core"
)

// CUEValidator provides validation for all schema types.
// It loads the embedded CUE schemas once and reuses them for all validations.
type CUEValidator struct {
	ctx *cue.Context
	pkg cue.Value
}

// NewCUEValidator creates a new validator by loading the embedded CUE schemas.
// Returns an error if the schemas cannot be loaded or compiled.
func NewCUEValidator() (*CUEValidator, error) {
	// Create CUE context
	cueCtx := cuecontext.New()

	// Convert embed.FS to our custom fs.FS
	memFS := billy.NewMemory()
	if err := core.CopyFromEmbedFS(CUESchemas, memFS, "."); err != nil {
		return nil, fmt.Errorf("failed to copy embedded schemas: %w", err)
	}

	// Create loader from filesystem
	loader := cuepkg.NewLoader(memFS)

	// Load CUE package from filesystem
	cueValue, err := loader.LoadPackage(context.Background(), ".")
	if err != nil {
		return nil, fmt.Errorf("failed to load CUE schemas: %w", err)
	}

	return &CUEValidator{
		ctx: cueCtx,
		pkg: cueValue,
	}, nil
}

// ValidateProjectState validates project state YAML against the ProjectState schema.
func (v *CUEValidator) ValidateProjectState(data []byte) error {
	// Lookup the ProjectState definition
	schema := v.pkg.LookupPath(cue.ParsePath("#ProjectState"))
	if !schema.Exists() {
		return fmt.Errorf("ProjectState schema not found")
	}

	// Validate YAML against schema
	if err := cueyaml.Validate(data, schema); err != nil {
		return fmt.Errorf("project state validation failed: %w", err)
	}

	return nil
}

// ValidateTaskState validates task state YAML against the TaskState schema.
func (v *CUEValidator) ValidateTaskState(data []byte) error {
	// Lookup the TaskState definition
	schema := v.pkg.LookupPath(cue.ParsePath("#TaskState"))
	if !schema.Exists() {
		return fmt.Errorf("TaskState schema not found")
	}

	// Validate YAML against schema
	if err := cueyaml.Validate(data, schema); err != nil {
		return fmt.Errorf("task state validation failed: %w", err)
	}

	return nil
}

// ValidateRefsCommittedIndex validates committed refs index JSON against the RefsCommittedIndex schema.
func (v *CUEValidator) ValidateRefsCommittedIndex(data []byte) error {
	// Lookup the RefsCommittedIndex definition
	schema := v.pkg.LookupPath(cue.ParsePath("#RefsCommittedIndex"))
	if !schema.Exists() {
		return fmt.Errorf("RefsCommittedIndex schema not found")
	}

	// Validate JSON against schema (yaml.Validate works for JSON too)
	if err := cueyaml.Validate(data, schema); err != nil {
		return fmt.Errorf("committed refs index validation failed: %w", err)
	}

	return nil
}

// ValidateRefsLocalIndex validates local refs index JSON against the RefsLocalIndex schema.
func (v *CUEValidator) ValidateRefsLocalIndex(data []byte) error {
	// Lookup the RefsLocalIndex definition
	schema := v.pkg.LookupPath(cue.ParsePath("#RefsLocalIndex"))
	if !schema.Exists() {
		return fmt.Errorf("RefsLocalIndex schema not found")
	}

	// Validate JSON against schema (yaml.Validate works for JSON too)
	if err := cueyaml.Validate(data, schema); err != nil {
		return fmt.Errorf("local refs index validation failed: %w", err)
	}

	return nil
}
