package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	"cuelang.org/go/encoding/yaml"
	"github.com/your-org/sow/internal/schema"
)

// Validator manages CUE validation with schema caching
type Validator struct {
	ctx     *cue.Context
	schemas map[string]cue.Value
	mu      sync.RWMutex
}

// Global validator instance (singleton for performance)
var globalValidator *Validator
var once sync.Once

// getValidator returns the global validator instance (lazy initialization)
func getValidator() *Validator {
	once.Do(func() {
		globalValidator = &Validator{
			ctx:     cuecontext.New(),
			schemas: make(map[string]cue.Value),
		}
	})
	return globalValidator
}

// getSchema loads and caches a CUE schema
func (v *Validator) getSchema(schemaType string) (cue.Value, error) {
	// Check cache first (read lock)
	v.mu.RLock()
	if schema, ok := v.schemas[schemaType]; ok {
		v.mu.RUnlock()
		return schema, nil
	}
	v.mu.RUnlock()

	// Load schema (write lock)
	v.mu.Lock()
	defer v.mu.Unlock()

	// Double-check in case another goroutine loaded it
	if schema, ok := v.schemas[schemaType]; ok {
		return schema, nil
	}

	// Get embedded schema source
	schemaCUE := schema.GetSchema(schemaType)
	if schemaCUE == "" {
		return cue.Value{}, fmt.Errorf("unknown schema type: %s", schemaType)
	}

	// Compile schema
	schemaValue := v.ctx.CompileString(schemaCUE)
	if schemaValue.Err() != nil {
		return cue.Value{}, fmt.Errorf("failed to compile schema %s: %w", schemaType, schemaValue.Err())
	}

	// Extract the specific definition based on schema type
	var definitionName string
	switch schemaType {
	case "project-state":
		definitionName = "#ProjectState"
	case "task-state":
		definitionName = "#TaskState"
	case "sink-index":
		definitionName = "#SinkIndex"
	case "repo-index":
		definitionName = "#RepoIndex"
	case "sow-version":
		definitionName = "#VersionFile"
	default:
		return cue.Value{}, fmt.Errorf("unknown schema type: %s", schemaType)
	}

	// Lookup the definition
	definition := schemaValue.LookupPath(cue.ParsePath(definitionName))
	if definition.Err() != nil {
		return cue.Value{}, fmt.Errorf("failed to find definition %s in schema %s: %w", definitionName, schemaType, definition.Err())
	}

	// Cache it
	v.schemas[schemaType] = definition
	return definition, nil
}

// validate performs validation of a file against a schema
func (v *Validator) validate(filePath string, schemaType string, isJSON bool) error {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Get schema
	schemaValue, err := v.getSchema(schemaType)
	if err != nil {
		return err
	}

	// Parse data based on format
	var dataValue cue.Value
	if isJSON {
		// Parse JSON
		var jsonData interface{}
		if err := json.Unmarshal(data, &jsonData); err != nil {
			return fmt.Errorf("failed to parse JSON in %s: %w", filePath, err)
		}
		dataValue = v.ctx.Encode(jsonData)
	} else {
		// Parse YAML
		yamlData, err := yaml.Extract(filePath, data)
		if err != nil {
			return fmt.Errorf("failed to parse YAML in %s: %w", filePath, err)
		}
		dataValue = v.ctx.BuildFile(yamlData)
	}

	if dataValue.Err() != nil {
		return fmt.Errorf("failed to parse data from %s: %w", filePath, dataValue.Err())
	}

	// Unify data with schema (validation happens here)
	unified := schemaValue.Unify(dataValue)

	// Check for unification errors first
	if unified.Err() != nil {
		return formatValidationError(filePath, unified.Err())
	}

	// Then validate for completeness and concreteness
	if err := unified.Validate(cue.Concrete(true)); err != nil {
		return formatValidationError(filePath, err)
	}

	return nil
}

// formatValidationError formats CUE validation errors into readable messages
func formatValidationError(filePath string, err error) error {
	if err == nil {
		return nil
	}

	// Extract error details from CUE error
	var errMsgs []string
	for _, e := range errors.Errors(err) {
		msg := e.Error()

		// Try to extract field path and error message
		parts := strings.SplitN(msg, ":", 2)
		if len(parts) == 2 {
			path := strings.TrimSpace(parts[0])
			detail := strings.TrimSpace(parts[1])
			errMsgs = append(errMsgs, fmt.Sprintf("  - %s: %s", path, detail))
		} else {
			errMsgs = append(errMsgs, fmt.Sprintf("  - %s", msg))
		}
	}

	if len(errMsgs) == 0 {
		return fmt.Errorf("validation failed for %s: %v", filePath, err)
	}

	return fmt.Errorf("validation failed for %s:\n%s", filePath, strings.Join(errMsgs, "\n"))
}

// ValidateProjectState validates a project state YAML file
func ValidateProjectState(filePath string) error {
	return getValidator().validate(filePath, "project-state", false)
}

// ValidateTaskState validates a task state YAML file
func ValidateTaskState(filePath string) error {
	return getValidator().validate(filePath, "task-state", false)
}

// ValidateSinkIndex validates a sink index JSON file
func ValidateSinkIndex(filePath string) error {
	return getValidator().validate(filePath, "sink-index", true)
}

// ValidateRepoIndex validates a repo index JSON file
func ValidateRepoIndex(filePath string) error {
	return getValidator().validate(filePath, "repo-index", true)
}

// ValidateVersion validates a version YAML file
func ValidateVersion(filePath string) error {
	return getValidator().validate(filePath, "sow-version", false)
}
