// Package breakdown provides functionality for breakdown mode.
package breakdown

import "errors"

var (
	// ErrNoBreakdown is returned when no breakdown session exists.
	ErrNoBreakdown = errors.New("no active breakdown session")

	// ErrBreakdownExists is returned when trying to create a breakdown that already exists.
	ErrBreakdownExists = errors.New("breakdown session already exists")

	// ErrInputExists is returned when trying to add an input that already exists.
	ErrInputExists = errors.New("input already exists in breakdown index")

	// ErrInputNotFound is returned when an input cannot be found.
	ErrInputNotFound = errors.New("input not found in breakdown index")

	// ErrWorkUnitExists is returned when trying to add a work unit that already exists.
	ErrWorkUnitExists = errors.New("work unit already exists in breakdown index")

	// ErrWorkUnitNotFound is returned when a work unit cannot be found.
	ErrWorkUnitNotFound = errors.New("work unit not found in breakdown index")

	// ErrAlreadyPublished is returned when trying to publish an already published work unit.
	ErrAlreadyPublished = errors.New("work unit already published")

	// ErrNotApproved is returned when trying to publish a work unit that hasn't been approved.
	ErrNotApproved = errors.New("work unit not approved for publishing")
)
