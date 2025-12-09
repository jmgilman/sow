package project

// This file ensures all schema definitions are referenced
// so that the "time" import is considered "used" during testing.
_testHelper: {
	_project:    #ProjectState
	_phase:      #PhaseState
	_artifact:   #ArtifactState
	_task:       #TaskState
	_statechart: #StatechartState
}
