package breakdown

import (
	"fmt"
	"time"

	"github.com/jmgilman/sow/cli/internal/sow"
	"github.com/jmgilman/sow/cli/schemas"
)

// AddInput adds an input to the breakdown index.
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
	input := schemas.BreakdownInput{
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

	return nil
}

// RemoveInput removes an input from the breakdown index.
func RemoveInput(ctx *sow.Context, path string) error {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return err
	}

	// Find and remove input
	found := false
	newInputs := make([]schemas.BreakdownInput, 0, len(index.Inputs))
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

	return nil
}

// ListInputs returns all inputs in the breakdown index.
func ListInputs(ctx *sow.Context) ([]schemas.BreakdownInput, error) {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return nil, err
	}

	return index.Inputs, nil
}

// AddWorkUnit adds a work unit to the breakdown index.
func AddWorkUnit(ctx *sow.Context, id, title, description string, dependsOn []string) error {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return err
	}

	// Check if work unit already exists
	for _, unit := range index.Work_units {
		if unit.Id == id {
			return ErrWorkUnitExists
		}
	}

	// Add work unit
	now := time.Now()
	unit := schemas.BreakdownWorkUnit{
		Id:          id,
		Title:       title,
		Description: description,
		Status:      "proposed",
		Depends_on:  dependsOn,
		Created_at:  now,
		Updated_at:  now,
	}
	index.Work_units = append(index.Work_units, unit)

	// Save index
	if err := SaveIndex(ctx, index); err != nil {
		return err
	}

	return nil
}

// UpdateWorkUnit updates a work unit's metadata in the breakdown index.
func UpdateWorkUnit(ctx *sow.Context, id string, title, description *string, dependsOn []string) error {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return err
	}

	// Find and update work unit
	found := false
	for i, unit := range index.Work_units {
		if unit.Id == id {
			if title != nil {
				index.Work_units[i].Title = *title
			}
			if description != nil {
				index.Work_units[i].Description = *description
			}
			if dependsOn != nil {
				index.Work_units[i].Depends_on = dependsOn
			}
			index.Work_units[i].Updated_at = time.Now()
			found = true
			break
		}
	}

	if !found {
		return ErrWorkUnitNotFound
	}

	// Save index
	if err := SaveIndex(ctx, index); err != nil {
		return err
	}

	return nil
}

// RemoveWorkUnit removes a work unit from the breakdown index.
func RemoveWorkUnit(ctx *sow.Context, id string) error {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return err
	}

	// Find and remove work unit
	found := false
	newUnits := make([]schemas.BreakdownWorkUnit, 0, len(index.Work_units))
	for _, unit := range index.Work_units {
		if unit.Id == id {
			found = true
			continue
		}
		newUnits = append(newUnits, unit)
	}

	if !found {
		return ErrWorkUnitNotFound
	}

	index.Work_units = newUnits

	// Save index
	if err := SaveIndex(ctx, index); err != nil {
		return err
	}

	return nil
}

// GetWorkUnit retrieves a work unit from the breakdown index.
func GetWorkUnit(ctx *sow.Context, id string) (*schemas.BreakdownWorkUnit, error) {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return nil, err
	}

	// Find work unit
	for _, unit := range index.Work_units {
		if unit.Id == id {
			return &unit, nil
		}
	}

	return nil, ErrWorkUnitNotFound
}

// ListWorkUnits returns all work units in the breakdown index.
func ListWorkUnits(ctx *sow.Context) ([]schemas.BreakdownWorkUnit, error) {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return nil, err
	}

	return index.Work_units, nil
}

// SetWorkUnitDocumentPath sets the document path for a work unit and updates status to document_created.
func SetWorkUnitDocumentPath(ctx *sow.Context, id, documentPath string) error {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return err
	}

	// Find and update work unit
	found := false
	for i, unit := range index.Work_units {
		if unit.Id == id {
			index.Work_units[i].Document_path = documentPath
			if index.Work_units[i].Status == "proposed" {
				index.Work_units[i].Status = "document_created"
			}
			index.Work_units[i].Updated_at = time.Now()
			found = true
			break
		}
	}

	if !found {
		return ErrWorkUnitNotFound
	}

	// Save index
	if err := SaveIndex(ctx, index); err != nil {
		return err
	}

	return nil
}

// ApproveWorkUnit marks a work unit as approved for publishing.
func ApproveWorkUnit(ctx *sow.Context, id string) error {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return err
	}

	// Find and approve work unit
	found := false
	for i, unit := range index.Work_units {
		if unit.Id == id {
			if unit.Status == "published" {
				return ErrAlreadyPublished
			}
			index.Work_units[i].Status = "approved"
			index.Work_units[i].Updated_at = time.Now()
			found = true
			break
		}
	}

	if !found {
		return ErrWorkUnitNotFound
	}

	// Save index
	if err := SaveIndex(ctx, index); err != nil {
		return err
	}

	return nil
}

// PublishWorkUnit publishes a work unit as a GitHub issue.
func PublishWorkUnit(ctx *sow.Context, id string, issueURL string, issueNumber int64) error {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return err
	}

	// Find and update work unit
	found := false
	for i, unit := range index.Work_units {
		if unit.Id == id {
			if unit.Status == "published" {
				return ErrAlreadyPublished
			}
			if unit.Status != "approved" {
				return ErrNotApproved
			}
			index.Work_units[i].Status = "published"
			index.Work_units[i].Github_issue_url = issueURL
			index.Work_units[i].Github_issue_number = issueNumber
			index.Work_units[i].Updated_at = time.Now()
			found = true
			break
		}
	}

	if !found {
		return ErrWorkUnitNotFound
	}

	// Save index
	if err := SaveIndex(ctx, index); err != nil {
		return err
	}

	return nil
}

// GetUnpublishedUnits returns all work units that are approved but not yet published.
func GetUnpublishedUnits(ctx *sow.Context) ([]schemas.BreakdownWorkUnit, error) {
	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return nil, err
	}

	var unpublished []schemas.BreakdownWorkUnit
	for _, unit := range index.Work_units {
		if unit.Status == "approved" {
			unpublished = append(unpublished, unit)
		}
	}

	return unpublished, nil
}

// UpdateStatus updates the breakdown session status.
func UpdateStatus(ctx *sow.Context, status string) error {
	// Validate status
	validStatuses := map[string]bool{
		"active":    true,
		"completed": true,
		"abandoned": true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s (must be active, completed, or abandoned)", status)
	}

	// Load current index
	index, err := LoadIndex(ctx)
	if err != nil {
		return err
	}

	// Update status
	index.Breakdown.Status = status

	// Save index
	if err := SaveIndex(ctx, index); err != nil {
		return err
	}

	return nil
}
