package config

import "errors"

// Sentinel errors for configuration loading failures.
var (
	// ErrConfigNotFound is returned when the configuration file does not exist.
	ErrConfigNotFound = errors.New("config file not found")

	// ErrInvalidConfig is returned when the configuration is structurally invalid
	// or fails validation rules.
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrInvalidYAML is returned when the configuration file contains invalid YAML syntax.
	ErrInvalidYAML = errors.New("invalid YAML syntax")
)
