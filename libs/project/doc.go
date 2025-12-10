// Package project provides a project SDK for defining project types with phases,
// transitions, and state machines.
//
// This package provides core type definitions for building project-based workflows
// with state machine semantics. It defines the fundamental types used throughout
// the project lifecycle system, including states, events, guards, and actions.
//
// # Key Concepts
//
// States represent positions in the project lifecycle state machine. All project
// types share the NoProject state for when no active project exists:
//
//	const (
//	    NoProject project.State = "NoProject"
//	    Planning  project.State = "Planning"
//	    Active    project.State = "Active"
//	)
//
// Events are triggers that cause state transitions. They are fired when advancing
// through the project lifecycle:
//
//	const (
//	    EventInit     project.Event = "Init"
//	    EventAdvance  project.Event = "Advance"
//	    EventComplete project.Event = "Complete"
//	)
//
// Guards are condition functions that determine if a transition is allowed:
//
//	guard := project.Guard(func() bool {
//	    return planningComplete
//	})
//
// GuardTemplate provides a reusable guard pattern that can be bound to a project
// instance via closure:
//
//	template := project.GuardTemplate{
//	    Description: "planning must be complete",
//	    Func: func(p *state.Project) bool {
//	        return p.PlanningComplete()
//	    },
//	}
//
// Actions are functions that mutate project state during transitions:
//
//	action := project.Action(func(p *state.Project) error {
//	    p.SetStatus("active")
//	    return nil
//	})
//
// PromptFunc generates contextual prompts for states during transitions:
//
//	promptFunc := project.PromptFunc(func(s project.State) string {
//	    switch s {
//	    case Planning:
//	        return "Create your project plan"
//	    default:
//	        return ""
//	    }
//	})
//
// # Subpackages
//
// The state subpackage provides project state types and persistence:
//
//	import "github.com/jmgilman/sow/libs/project/state"
//
// It contains:
//   - Project wrapper type with runtime behavior
//   - Phase, Task, and Artifact types
//   - Collection types for phases, tasks, and artifacts
//   - Backend interface for storage abstraction
//   - YAML and memory backend implementations
//   - Load/Save operations with CUE validation
//
// # Example Usage
//
// Define states and events for a custom project type:
//
//	const (
//	    StateIdle     project.State = "Idle"
//	    StateActive   project.State = "Active"
//	    StateComplete project.State = "Complete"
//
//	    EventStart  project.Event = "Start"
//	    EventFinish project.Event = "Finish"
//	)
//
// Create a prompt generator:
//
//	prompts := project.PromptFunc(func(s project.State) string {
//	    prompts := map[project.State]string{
//	        StateActive:   "Work on your tasks",
//	        StateComplete: "Review and finalize",
//	    }
//	    return prompts[s]
//	})
package project
