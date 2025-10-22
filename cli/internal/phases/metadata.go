package phases

// PhaseMetadata describes a phase's characteristics for validation and introspection.
// This allows the CLI to understand what operations are valid for a given phase.
type PhaseMetadata struct {
	// Name is the phase identifier (e.g., "discovery", "implementation")
	Name string

	// States lists all states that belong to this phase
	States []State

	// SupportsTasks indicates if this phase supports task management
	SupportsTasks bool

	// SupportsArtifacts indicates if this phase supports artifact tracking
	SupportsArtifacts bool

	// CustomFields lists phase-specific fields that can be set via CLI
	CustomFields []FieldDef
}

// FieldDef describes a custom field that can be set on a phase.
type FieldDef struct {
	// Name is the field identifier (e.g., "discovery_type", "architect_used")
	Name string

	// Type indicates the Go type of the field
	Type FieldType

	// Description explains what this field is for (optional, for help text)
	Description string
}

// FieldType represents the type of a custom field.
type FieldType string

const (
	// StringField represents a string field
	StringField FieldType = "string"

	// BoolField represents a boolean field
	BoolField FieldType = "bool"

	// IntField represents an integer field
	IntField FieldType = "int"

	// ArrayField represents an array/slice field
	ArrayField FieldType = "array"
)
