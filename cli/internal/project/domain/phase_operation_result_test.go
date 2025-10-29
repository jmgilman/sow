package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test event constant - use string to avoid import cycles.
const testEvent = "test_event"

func TestNoEvent(t *testing.T) {
	result := NoEvent()

	assert.NotNil(t, result)
	assert.Equal(t, "", string(result.Event))
}

func TestWithEvent(t *testing.T) {
	result := WithEvent(testEvent)

	assert.NotNil(t, result)
	assert.Equal(t, string(testEvent), string(result.Event))
}

func TestPhaseOperationResult_HasEvent(t *testing.T) {
	tests := []struct {
		name     string
		result   *PhaseOperationResult
		hasEvent bool
	}{
		{
			name:     "NoEvent returns empty event",
			result:   NoEvent(),
			hasEvent: false,
		},
		{
			name:     "WithEvent returns specified event",
			result:   WithEvent(testEvent),
			hasEvent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEmpty := tt.result.Event == ""
			assert.Equal(t, !tt.hasEvent, isEmpty)
		})
	}
}
