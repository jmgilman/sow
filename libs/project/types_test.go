package project

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestState_String(t *testing.T) {
	tests := []struct {
		name  string
		state State
		want  string
	}{
		{name: "NoProject constant", state: NoProject, want: "NoProject"},
		{name: "custom state", state: State("PlanningActive"), want: "PlanningActive"},
		{name: "empty state", state: State(""), want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.state.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEvent_String(t *testing.T) {
	tests := []struct {
		name  string
		event Event
		want  string
	}{
		{name: "custom event", event: Event("AdvancePlanning"), want: "AdvancePlanning"},
		{name: "empty event", event: Event(""), want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.event.String()
			assert.Equal(t, tt.want, got)
		})
	}
}
