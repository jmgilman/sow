package design

import "errors"

var (
	ErrNoDesign       = errors.New("no active design session")
	ErrDesignExists   = errors.New("design session already exists")
	ErrInputExists    = errors.New("input already exists")
	ErrInputNotFound  = errors.New("input not found")
	ErrOutputExists   = errors.New("output already exists")
	ErrOutputNotFound = errors.New("output not found")
)
