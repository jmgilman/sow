package design

import (
	"fmt"
	"time"

	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
)

// AddInput adds an input to the design index.
func AddInput(ctx *sow.Context, inputType, path, description string, tags []string) error {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return err
	}

	// Check if input already exists
	for _, input := range index.Inputs {
		if input.Path == path {
			return ErrInputExists
		}
	}

	// Add input
	input := schemas.DesignInput{
		Type:        inputType,
		Path:        path,
		Description: description,
		Tags:        tags,
		Added_at:    time.Now(),
	}
	index.Inputs = append(index.Inputs, input)

	// Save index
	if err := SaveIndex(ctx, index); err != nil {
		return err
	}

	// Log the action
	LogInputAdded(ctx, inputType, path, description, tags)

	return nil
}

// RemoveInput removes an input from the design index.
func RemoveInput(ctx *sow.Context, path string) error {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return err
	}

	// Find and remove input
	found := false
	newInputs := make([]schemas.DesignInput, 0, len(index.Inputs))
	for _, input := range index.Inputs {
		if input.Path == path {
			found = true
			continue
		}
		newInputs = append(newInputs, input)
	}

	if !found {
		return ErrInputNotFound
	}

	index.Inputs = newInputs

	// Save index
	if err := SaveIndex(ctx, index); err != nil {
		return err
	}

	// Log the action
	LogInputRemoved(ctx, path)

	return nil
}

// GetInput retrieves an input from the design index.
func GetInput(ctx *sow.Context, path string) (*schemas.DesignInput, error) {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return nil, err
	}

	// Find input
	for _, input := range index.Inputs {
		if input.Path == path {
			return &input, nil
		}
	}

	return nil, ErrInputNotFound
}

// ListInputs returns all inputs in the design index.
func ListInputs(ctx *sow.Context) ([]schemas.DesignInput, error) {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return nil, err
	}

	return index.Inputs, nil
}

// AddOutput adds an output to the design index.
func AddOutput(ctx *sow.Context, path, description, targetLocation, docType string, tags []string) error {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return err
	}

	// Check if output already exists
	for _, output := range index.Outputs {
		if output.Path == path {
			return ErrOutputExists
		}
	}

	// Add output
	output := schemas.DesignOutput{
		Path:            path,
		Description:     description,
		Target_location: targetLocation,
		Type:            docType,
		Tags:            tags,
		Added_at:        time.Now(),
	}
	index.Outputs = append(index.Outputs, output)

	// Save index
	if err := SaveIndex(ctx, index); err != nil {
		return err
	}

	// Log the action
	LogOutputAdded(ctx, path, description, targetLocation, docType, tags)

	return nil
}

// RemoveOutput removes an output from the design index.
func RemoveOutput(ctx *sow.Context, path string) error {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return err
	}

	// Find and remove output
	found := false
	newOutputs := make([]schemas.DesignOutput, 0, len(index.Outputs))
	for _, output := range index.Outputs {
		if output.Path == path {
			found = true
			continue
		}
		newOutputs = append(newOutputs, output)
	}

	if !found {
		return ErrOutputNotFound
	}

	index.Outputs = newOutputs

	// Save index
	if err := SaveIndex(ctx, index); err != nil {
		return err
	}

	// Log the action
	LogOutputRemoved(ctx, path)

	return nil
}

// UpdateOutputTarget updates an output's target location.
func UpdateOutputTarget(ctx *sow.Context, path, targetLocation string) error {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return err
	}

	// Find and update output
	found := false
	for i, output := range index.Outputs {
		if output.Path == path {
			index.Outputs[i].Target_location = targetLocation
			found = true
			break
		}
	}

	if !found {
		return ErrOutputNotFound
	}

	// Save index
	if err := SaveIndex(ctx, index); err != nil {
		return err
	}

	// Log the action
	LogOutputTargetSet(ctx, path, targetLocation)

	return nil
}

// GetOutput retrieves an output from the design index.
func GetOutput(ctx *sow.Context, path string) (*schemas.DesignOutput, error) {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return nil, err
	}

	// Find output
	for _, output := range index.Outputs {
		if output.Path == path {
			return &output, nil
		}
	}

	return nil, ErrOutputNotFound
}

// ListOutputs returns all outputs in the design index.
func ListOutputs(ctx *sow.Context) ([]schemas.DesignOutput, error) {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return nil, err
	}

	return index.Outputs, nil
}

// UpdateStatus updates the design session status.
func UpdateStatus(ctx *sow.Context, status string) error {
	// Validate status
	validStatuses := map[string]bool{
		"active":    true,
		"in_review": true,
		"completed": true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s (must be active, in_review, or completed)", status)
	}

	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return err
	}

	// Store old status for logging
	oldStatus := index.Design.Status

	// Update status
	index.Design.Status = status

	// Save index
	if err := SaveIndex(ctx, index); err != nil {
		return err
	}

	// Log the action
	LogStatusChanged(ctx, oldStatus, status)

	return nil
}
