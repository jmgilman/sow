package cmdutil

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/jmgilman/sow/cli/internal/sdks/project/state"
)

// ParseFieldPath splits a field path into segments using dot notation.
// Example: "metadata.assessment" → ["metadata", "assessment"].
func ParseFieldPath(path string) []string {
	if path == "" {
		return []string{""}
	}
	return strings.Split(path, ".")
}

// IsMetadataPath checks if the field path routes to the metadata map.
// Returns true if the first segment is "metadata" and there are at least 2 segments.
func IsMetadataPath(segments []string) bool {
	return len(segments) >= 2 && segments[0] == "metadata"
}

// ConvertValue converts a string value to the appropriate type (bool, int, or string).
// Supports: "true"/"false" → bool, numeric strings → int, everything else → string.
func ConvertValue(value string) interface{} {
	// Try bool conversion
	if value == "true" {
		return true
	}
	if value == "false" {
		return false
	}

	// Try int conversion
	if intVal, err := strconv.Atoi(value); err == nil {
		return intVal
	}

	// Default to string
	return value
}

// SetField sets a field value on a target struct using a field path.
// Supports direct fields and metadata paths (metadata.key or metadata.key.subkey).
// The value string is automatically converted to the appropriate type.
func SetField(target interface{}, fieldPath string, value string) error {
	if target == nil {
		return fmt.Errorf("cannot set field on nil target")
	}

	if fieldPath == "" {
		return fmt.Errorf("field path cannot be empty")
	}

	segments := ParseFieldPath(fieldPath)

	// Check for invalid "metadata" without subfield
	if len(segments) == 1 && segments[0] == "metadata" {
		return fmt.Errorf("invalid metadata path: must specify a key after 'metadata'")
	}

	// Handle metadata paths
	if IsMetadataPath(segments) {
		return setMetadataField(target, segments[1:], value)
	}

	// Handle direct fields
	return setDirectField(target, segments[0], value)
}

// GetField retrieves a field value from a target struct using a field path.
// Supports direct fields and metadata paths (metadata.key or metadata.key.subkey).
func GetField(target interface{}, fieldPath string) (interface{}, error) {
	if target == nil {
		return nil, fmt.Errorf("cannot get field from nil target")
	}

	if fieldPath == "" {
		return nil, fmt.Errorf("field path cannot be empty")
	}

	segments := ParseFieldPath(fieldPath)

	// Handle metadata paths
	if IsMetadataPath(segments) {
		return getMetadataField(target, segments[1:])
	}

	// Handle direct fields
	return getDirectField(target, segments[0])
}

// setDirectField sets a direct field on the target struct.
func setDirectField(target interface{}, fieldName string, value string) error {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	if v.IsNil() {
		return fmt.Errorf("cannot set field on nil target")
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a struct pointer")
	}

	// Handle wrapped state types - look for embedded struct
	if v.NumField() > 0 {
		firstField := v.Field(0)
		if firstField.Type().Kind() == reflect.Struct && firstField.Type().Name() != "" {
			// Try the embedded struct
			v = firstField
		}
	}

	field := v.FieldByName(capitalizeFirst(fieldName))
	if !field.IsValid() {
		return fmt.Errorf("unknown field: %s", fieldName)
	}

	if !field.CanSet() {
		return fmt.Errorf("cannot set field: %s", fieldName)
	}

	return setFieldValue(field, value)
}

// getDirectField gets a direct field from the target struct.
func getDirectField(target interface{}, fieldName string) (interface{}, error) {
	v := reflect.ValueOf(target)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("target must be a struct or struct pointer")
	}

	// Handle wrapped state types - look for embedded struct
	if v.NumField() > 0 {
		firstField := v.Field(0)
		if firstField.Type().Kind() == reflect.Struct && firstField.Type().Name() != "" {
			// Try the embedded struct
			v = firstField
		}
	}

	field := v.FieldByName(capitalizeFirst(fieldName))
	if !field.IsValid() {
		return nil, fmt.Errorf("unknown field: %s", fieldName)
	}

	return field.Interface(), nil
}

// setMetadataField sets a metadata field, creating nested maps as needed.
func setMetadataField(target interface{}, segments []string, value string) error {
	if len(segments) == 0 {
		return fmt.Errorf("invalid metadata path: must have at least one key after 'metadata'")
	}

	// Get the metadata map
	metadataField, err := getMetadataMap(target)
	if err != nil {
		return err
	}

	// Navigate/create nested structure
	currentMap := metadataField
	for i := 0; i < len(segments)-1; i++ {
		key := segments[i]
		nextVal, exists := currentMap[key]

		if !exists {
			// Create new nested map
			newMap := make(map[string]interface{})
			currentMap[key] = newMap
			currentMap = newMap
		} else {
			// Try to use existing map
			nextMap, ok := nextVal.(map[string]interface{})
			if !ok {
				return fmt.Errorf("metadata path conflict: %s is not a map", key)
			}
			currentMap = nextMap
		}
	}

	// Set the final value with type conversion
	finalKey := segments[len(segments)-1]
	currentMap[finalKey] = ConvertValue(value)

	return nil
}

// getMetadataField retrieves a metadata field, navigating nested maps.
func getMetadataField(target interface{}, segments []string) (interface{}, error) {
	if len(segments) == 0 {
		return nil, fmt.Errorf("invalid metadata path: must have at least one key after 'metadata'")
	}

	// Get the metadata map
	metadataField, err := getMetadataMap(target)
	if err != nil {
		return nil, err
	}

	// Navigate nested structure
	currentVal := interface{}(metadataField)
	for _, key := range segments {
		currentMap, ok := currentVal.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("metadata path not found: expected map at %s", key)
		}

		nextVal, exists := currentMap[key]
		if !exists {
			return nil, fmt.Errorf("metadata key not found: %s", key)
		}

		currentVal = nextVal
	}

	return currentVal, nil
}

// getMetadataMap gets the Metadata map from a target struct, initializing if needed.
func getMetadataMap(target interface{}) (map[string]interface{}, error) {
	v := reflect.ValueOf(target)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("target must be a struct or struct pointer")
	}

	// Handle wrapped state types - look for embedded struct
	if v.NumField() > 0 {
		firstField := v.Field(0)
		if firstField.Type().Kind() == reflect.Struct && firstField.Type().Name() != "" {
			// Use the embedded struct
			v = firstField
		}
	}

	metadataField := v.FieldByName("Metadata")
	if !metadataField.IsValid() {
		return nil, fmt.Errorf("target does not have a Metadata field")
	}

	// Initialize metadata map if nil
	if metadataField.IsNil() {
		newMap := make(map[string]interface{})
		metadataField.Set(reflect.ValueOf(newMap))
		return newMap, nil
	}

	metadataMap, ok := metadataField.Interface().(map[string]interface{})
	if !ok {
		// Try to convert map[string]any to map[string]interface{}
		metadataAny, ok := metadataField.Interface().(map[string]any)
		if !ok {
			return nil, fmt.Errorf("metadata field is not a map")
		}
		// Convert to map[string]interface{}
		metadataMap = make(map[string]interface{})
		for k, v := range metadataAny {
			metadataMap[k] = v
		}
	}

	return metadataMap, nil
}

// setFieldValue sets a reflect.Value to the converted value.
func setFieldValue(field reflect.Value, value string) error {
	converted := ConvertValue(value)

	switch field.Kind() {
	case reflect.Bool:
		if b, ok := converted.(bool); ok {
			field.SetBool(b)
		} else {
			return fmt.Errorf("cannot convert %v to bool", converted)
		}
	case reflect.Int, reflect.Int64:
		if i, ok := converted.(int); ok {
			field.SetInt(int64(i))
		} else {
			return fmt.Errorf("cannot convert %v to int", converted)
		}
	case reflect.String:
		if s, ok := converted.(string); ok {
			field.SetString(s)
		} else {
			return fmt.Errorf("cannot convert %v to string", converted)
		}
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}

	return nil
}

// capitalizeFirst converts a field name to Go exported name format
// Example: "approved" → "Approved", "created_at" → "Created_at".
func capitalizeFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// Known direct fields by type (for validation/documentation purposes).
var knownFields = map[string][]string{
	"Artifact": {"type", "path", "approved", "created_at"},
	"Phase":    {"status", "enabled", "created_at", "started_at", "completed_at"},
	"Task":     {"id", "name", "phase", "status", "iteration", "assigned_agent", "created_at", "started_at", "updated_at", "completed_at"},
	"Project":  {"name", "type", "branch", "description", "created_at", "updated_at"},
}

// IsKnownField checks if a field name is a known direct field for the given type.
func IsKnownField(typeName string, fieldName string) bool {
	fields, ok := knownFields[typeName]
	if !ok {
		return false
	}

	for _, f := range fields {
		if f == fieldName {
			return true
		}
	}
	return false
}

// GetTypeName returns the type name of a state object.
func GetTypeName(target interface{}) string {
	v := reflect.ValueOf(target)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()

	// Check for wrapped state types
	switch target.(type) {
	case *state.Artifact, state.Artifact:
		return "Artifact"
	case *state.Task, state.Task:
		return "Task"
	case *state.Phase, state.Phase:
		return "Phase"
	default:
		return t.Name()
	}
}
