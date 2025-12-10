package project

import (
	"context"
)

// TransitionOption configures a transition.
type TransitionOption func(*transitionConfig)

// transitionConfig holds internal configuration for a transition.
type transitionConfig struct {
	guard            Guard
	guardDescription string
	onEntry          func(context.Context, ...any) error
	onExit           func(context.Context, ...any) error
}

// WithGuard sets a guard function that must return true for the transition to proceed.
func WithGuard(guard Guard) TransitionOption {
	return func(c *transitionConfig) {
		c.guard = guard
	}
}

// WithGuardDescription sets a guard with a human-readable description.
// The description is used in error messages when the guard fails.
func WithGuardDescription(description string, guard Guard) TransitionOption {
	return func(c *transitionConfig) {
		c.guard = guard
		c.guardDescription = description
	}
}

// WithOnEntry sets an action to run when entering the target state.
func WithOnEntry(action func(context.Context, ...any) error) TransitionOption {
	return func(c *transitionConfig) {
		c.onEntry = action
	}
}

// WithOnExit sets an action to run when leaving the source state.
func WithOnExit(action func(context.Context, ...any) error) TransitionOption {
	return func(c *transitionConfig) {
		c.onExit = action
	}
}
