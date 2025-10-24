// Package design provides design mode functionality for creating formal design documents.
package design

import "errors"

var (
	// ErrNoDesign indicates no design session exists in the current context.
	ErrNoDesign = errors.New("no active design session")

	// ErrDesignExists indicates a design session already exists.
	ErrDesignExists = errors.New("design session already exists")

	// ErrInputExists indicates an input already exists in the design index.
	ErrInputExists = errors.New("input already exists")

	// ErrInputNotFound indicates an input is not in the design index.
	ErrInputNotFound = errors.New("input not found")

	// ErrOutputExists indicates an output already exists in the design index.
	ErrOutputExists = errors.New("output already exists")

	// ErrOutputNotFound indicates an output is not in the design index.
	ErrOutputNotFound = errors.New("output not found")
)
