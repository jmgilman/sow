package agent

import (
	"fmt"
	"strconv"

	"github.com/jmgilman/sow/cli/internal/cmdutil"
	"github.com/jmgilman/sow/cli/internal/phases"
	projectpkg "github.com/jmgilman/sow/cli/internal/project"
	"github.com/jmgilman/sow/cli/schemas"
	"github.com/spf13/cobra"
)

// NewSetCmd creates the command to set a custom field on the active phase.
//
// Usage:
//
//	sow agent set <field> <value>
//
// Sets a custom field on the currently active phase.
func NewSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <field> <value>",
		Short: "Set a custom field on the active phase",
		Long: `Set a custom field on the currently active phase.

Each phase has different custom fields that can be set. Use 'sow agent info' to see
available fields for the current phase.

The value is parsed based on the field type:
  - String fields: any text value
  - Bool fields: true/false
  - Int fields: numeric value

Example:
  # Set discovery type
  sow agent set discovery_type feature

  # Set tasks approved flag
  sow agent set tasks_approved true

  # Set review iteration
  sow agent set iteration 2`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fieldName := args[0]
			valueStr := args[1]

			// Get context
			ctx := cmdutil.GetContext(cmd.Context())

			// Load project
			project, err := projectpkg.Load(ctx)
			if err != nil {
				return fmt.Errorf("no active project - run 'sow agent project init' first")
			}

			// Determine active phase
			state := project.State()
			activePhase, phaseStatus := projectpkg.DetermineActivePhase(state)

			if activePhase == "unknown" {
				return fmt.Errorf("no active phase found")
			}

			// Validate we're in an active state (not a decision state)
			// Only optional phases (discovery, design) have decision states
			if phaseStatus == "pending" && (activePhase == "discovery" || activePhase == "design") {
				return fmt.Errorf("phase %s is in decision state - enable it first", activePhase)
			}

			// Get phase metadata to validate field exists
			metadata, err := projectpkg.GetPhaseMetadata(activePhase)
			if err != nil {
				return fmt.Errorf("failed to get phase metadata: %w", err)
			}

			// Find the field
			var fieldDef *phases.FieldDef
			for _, field := range metadata.CustomFields {
				if field.Name == fieldName {
					fieldDef = &field
					break
				}
			}

			if fieldDef == nil {
				return fmt.Errorf("field %s not found in %s phase - use 'sow agent info' to see available fields", fieldName, activePhase)
			}

			// Set the field based on phase and type
			if err := setField(state, activePhase, fieldName, valueStr, fieldDef.Type); err != nil {
				return fmt.Errorf("failed to set field: %w", err)
			}

			// Save the state
			if err := project.Machine().Save(); err != nil {
				return fmt.Errorf("failed to save state: %w", err)
			}

			cmd.Printf("\n✓ Set %s = %s on %s phase\n", fieldName, valueStr, activePhase)
			return nil
		},
	}

	return cmd
}

// setField sets a field value on the appropriate phase in the state.
func setField(state *schemas.ProjectState, phaseName, fieldName, valueStr string, fieldType phases.FieldType) error {
	// Parse value based on type
	var value interface{}
	var err error

	switch fieldType {
	case phases.StringField:
		value = valueStr

	case phases.BoolField:
		value, err = strconv.ParseBool(valueStr)
		if err != nil {
			return fmt.Errorf("invalid boolean value %q - use true or false", valueStr)
		}

	case phases.IntField:
		intVal, err := strconv.Atoi(valueStr)
		if err != nil {
			return fmt.Errorf("invalid integer value %q", valueStr)
		}
		value = intVal

	default:
		return fmt.Errorf("unsupported field type: %s", fieldType)
	}

	// Set the field on the appropriate phase
	switch phaseName {
	case "discovery":
		return setDiscoveryField(state, fieldName, value)
	case "design":
		return setDesignField(state, fieldName, value)
	case "implementation":
		return setImplementationField(state, fieldName, value)
	case "review":
		return setReviewField(state, fieldName, value)
	case "finalize":
		return setFinalizeField(state, fieldName, value)
	default:
		return fmt.Errorf("unknown phase: %s", phaseName)
	}
}

func setDiscoveryField(state *schemas.ProjectState, fieldName string, value interface{}) error {
	switch fieldName {
	case "discovery_type":
		strVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("discovery_type must be a string")
		}
		state.Phases.Discovery.Discovery_type = &strVal
	default:
		return fmt.Errorf("unknown field: %s", fieldName)
	}
	return nil
}

func setDesignField(state *schemas.ProjectState, fieldName string, value interface{}) error {
	switch fieldName {
	case "architect_used":
		boolVal, ok := value.(bool)
		if !ok {
			return fmt.Errorf("architect_used must be a boolean")
		}
		state.Phases.Design.Architect_used = &boolVal
	default:
		return fmt.Errorf("unknown field: %s", fieldName)
	}
	return nil
}

func setImplementationField(state *schemas.ProjectState, fieldName string, value interface{}) error {
	switch fieldName {
	case "planner_used":
		boolVal, ok := value.(bool)
		if !ok {
			return fmt.Errorf("planner_used must be a boolean")
		}
		state.Phases.Implementation.Planner_used = &boolVal
	case "tasks_approved":
		boolVal, ok := value.(bool)
		if !ok {
			return fmt.Errorf("tasks_approved must be a boolean")
		}
		state.Phases.Implementation.Tasks_approved = boolVal
	default:
		return fmt.Errorf("unknown field: %s", fieldName)
	}
	return nil
}

func setReviewField(state *schemas.ProjectState, fieldName string, value interface{}) error {
	switch fieldName {
	case "iteration":
		intVal, ok := value.(int)
		if !ok {
			return fmt.Errorf("iteration must be an integer")
		}
		state.Phases.Review.Iteration = int64(intVal)
	default:
		return fmt.Errorf("unknown field: %s", fieldName)
	}
	return nil
}

func setFinalizeField(state *schemas.ProjectState, fieldName string, value interface{}) error {
	switch fieldName {
	case "project_deleted":
		boolVal, ok := value.(bool)
		if !ok {
			return fmt.Errorf("project_deleted must be a boolean")
		}
		state.Phases.Finalize.Project_deleted = boolVal
	case "pr_url":
		strVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("pr_url must be a string")
		}
		state.Phases.Finalize.Pr_url = &strVal
	default:
		return fmt.Errorf("unknown field: %s", fieldName)
	}
	return nil
}
