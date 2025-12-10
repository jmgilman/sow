package project

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithGuard(t *testing.T) {
	t.Parallel()

	t.Run("sets guard function", func(t *testing.T) {
		t.Parallel()

		guardCalled := false
		guard := func() bool {
			guardCalled = true
			return true
		}

		config := &transitionConfig{}
		opt := WithGuard(guard)
		opt(config)

		assert.NotNil(t, config.guard)
		config.guard()
		assert.True(t, guardCalled)
	})
}

func TestWithGuardDescription(t *testing.T) {
	t.Parallel()

	t.Run("sets guard and description", func(t *testing.T) {
		t.Parallel()

		guardCalled := false
		guard := func() bool {
			guardCalled = true
			return true
		}

		config := &transitionConfig{}
		opt := WithGuardDescription("test description", guard)
		opt(config)

		assert.NotNil(t, config.guard)
		assert.Equal(t, "test description", config.guardDescription)
		config.guard()
		assert.True(t, guardCalled)
	})

	t.Run("empty description is allowed", func(t *testing.T) {
		t.Parallel()

		config := &transitionConfig{}
		opt := WithGuardDescription("", func() bool { return true })
		opt(config)

		assert.Empty(t, config.guardDescription)
		assert.NotNil(t, config.guard)
	})
}

func TestWithOnEntry(t *testing.T) {
	t.Parallel()

	t.Run("sets entry action", func(t *testing.T) {
		t.Parallel()

		entryCalled := false
		action := func(_ context.Context, _ ...any) error {
			entryCalled = true
			return nil
		}

		config := &transitionConfig{}
		opt := WithOnEntry(action)
		opt(config)

		assert.NotNil(t, config.onEntry)
		_ = config.onEntry(context.Background())
		assert.True(t, entryCalled)
	})
}

func TestWithOnExit(t *testing.T) {
	t.Parallel()

	t.Run("sets exit action", func(t *testing.T) {
		t.Parallel()

		exitCalled := false
		action := func(_ context.Context, _ ...any) error {
			exitCalled = true
			return nil
		}

		config := &transitionConfig{}
		opt := WithOnExit(action)
		opt(config)

		assert.NotNil(t, config.onExit)
		_ = config.onExit(context.Background())
		assert.True(t, exitCalled)
	})
}

func TestTransitionOptions_Composition(t *testing.T) {
	t.Parallel()

	t.Run("multiple options can be composed", func(t *testing.T) {
		t.Parallel()

		config := &transitionConfig{}
		opts := []TransitionOption{
			WithGuard(func() bool { return true }),
			WithOnEntry(func(_ context.Context, _ ...any) error { return nil }),
			WithOnExit(func(_ context.Context, _ ...any) error { return nil }),
		}

		for _, opt := range opts {
			opt(config)
		}

		assert.NotNil(t, config.guard)
		assert.NotNil(t, config.onEntry)
		assert.NotNil(t, config.onExit)
	})

	t.Run("later options overwrite earlier ones", func(t *testing.T) {
		t.Parallel()

		config := &transitionConfig{}
		firstGuardCalled := false
		secondGuardCalled := false

		WithGuard(func() bool {
			firstGuardCalled = true
			return false
		})(config)

		WithGuard(func() bool {
			secondGuardCalled = true
			return true
		})(config)

		config.guard()
		assert.False(t, firstGuardCalled, "first guard should not be called")
		assert.True(t, secondGuardCalled, "second guard should be called")
	})
}
