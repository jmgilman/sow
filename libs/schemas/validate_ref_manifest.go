package schemas

import (
	"context"
	"errors"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cuepkg "github.com/jmgilman/go/cue"
	"github.com/jmgilman/go/fs/billy"
	"github.com/jmgilman/go/fs/core"
)

// ErrRefManifestValidation indicates RefManifest validation failed.
var ErrRefManifestValidation = errors.New("ref manifest validation failed")

// refManifestSchema holds the loaded CUE value for the RefManifest schema.
// It is loaded once at package initialization using the embedded schemas.
var refManifestSchema cue.Value

// init loads the RefManifest schema from the embedded filesystem.
func init() {
	// Create in-memory filesystem
	memFS := billy.NewMemory()

	// Copy embedded schemas to in-memory filesystem
	if err := core.CopyFromEmbedFS(CUESchemas, memFS, "."); err != nil {
		panic(fmt.Errorf("failed to copy embedded schemas: %w", err))
	}

	// Create loader and load the root package
	loader := cuepkg.NewLoader(memFS)
	schemas, err := loader.LoadPackage(context.Background(), ".")
	if err != nil {
		panic(fmt.Errorf("failed to load schemas: %w", err))
	}

	// Look up the RefManifest definition
	refManifestSchema = schemas.LookupPath(cue.ParsePath("#RefManifest"))
	if refManifestSchema.Err() != nil {
		panic(fmt.Errorf("RefManifest schema not found: %w", refManifestSchema.Err()))
	}
}

// ValidateRefManifest validates a RefManifest against the CUE schema.
// It returns an error if the manifest is invalid, with details about
// which field(s) failed validation and why.
func ValidateRefManifest(manifest *RefManifest) error {
	if manifest == nil {
		return fmt.Errorf("%w: manifest is nil", ErrRefManifestValidation)
	}

	ctx := cuecontext.New()

	// Encode the manifest struct to a CUE value
	value := ctx.Encode(manifest)
	if value.Err() != nil {
		return fmt.Errorf("%w: failed to encode manifest: %w", ErrRefManifestValidation, value.Err())
	}

	// Unify the schema with the value
	result := refManifestSchema.Unify(value)

	// Validate with Concrete(true) to ensure all fields have concrete values
	if err := result.Validate(cue.Concrete(true)); err != nil {
		return fmt.Errorf("%w: %w", ErrRefManifestValidation, err)
	}

	return nil
}
