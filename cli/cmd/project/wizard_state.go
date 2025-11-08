package project

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/jmgilman/sow/cli/internal/sow"
)

// WizardState represents the current state of the project wizard.
type WizardState string

// Wizard states define the flow through project creation/continuation.
const (
	StateEntry          WizardState = "entry"
	StateCreateSource   WizardState = "create_source"
	StateIssueSelect    WizardState = "issue_select"
	StateTypeSelect     WizardState = "type_select"
	StateNameEntry      WizardState = "name_entry"
	StatePromptEntry    WizardState = "prompt_entry"
	StateProjectSelect  WizardState = "project_select"
	StateContinuePrompt WizardState = "continue_prompt"
	StateComplete       WizardState = "complete"
	StateCancelled      WizardState = "cancelled"
)

// Wizard manages the interactive project creation/continuation workflow.
type Wizard struct {
	state       WizardState
	ctx         *sow.Context
	choices     map[string]interface{}
	claudeFlags []string
}

// NewWizard creates a new wizard instance.
func NewWizard(ctx *sow.Context, claudeFlags []string) *Wizard {
	return &Wizard{
		state:       StateEntry,
		ctx:         ctx,
		choices:     make(map[string]interface{}),
		claudeFlags: claudeFlags,
	}
}

// Run executes the wizard state machine loop.
func (w *Wizard) Run() error {
	for w.state != StateComplete && w.state != StateCancelled {
		if err := w.handleState(); err != nil {
			return err
		}
	}

	if w.state == StateCancelled {
		return nil // User cancelled, not an error
	}

	return w.finalize()
}

// handleState dispatches to the appropriate handler based on current state.
func (w *Wizard) handleState() error {
	switch w.state {
	case StateEntry:
		return w.handleEntry()
	case StateCreateSource:
		return w.handleCreateSource()
	case StateIssueSelect:
		return w.handleIssueSelect()
	case StateTypeSelect:
		return w.handleTypeSelect()
	case StateNameEntry:
		return w.handleNameEntry()
	case StatePromptEntry:
		return w.handlePromptEntry()
	case StateProjectSelect:
		return w.handleProjectSelect()
	case StateContinuePrompt:
		return w.handleContinuePrompt()
	default:
		return fmt.Errorf("unknown state: %s", w.state)
	}
}

// handleEntry shows the main entry screen with create/continue/cancel options.
func (w *Wizard) handleEntry() error {
	var action string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What would you like to do?").
				Options(
					huh.NewOption("Create new project", "create"),
					huh.NewOption("Continue existing project", "continue"),
					huh.NewOption("Cancel", "cancel"),
				).
				Value(&action),
		),
	)

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			w.state = StateCancelled
			return nil
		}
		return fmt.Errorf("entry screen error: %w", err)
	}

	w.choices["action"] = action

	switch action {
	case "create":
		w.state = StateCreateSource
	case "continue":
		w.state = StateProjectSelect
	case "cancel":
		w.state = StateCancelled
	}

	return nil
}

// handleCreateSource shows options for creating a project (stub for now).
func (w *Wizard) handleCreateSource() error {
	fmt.Println("Create source screen (stub)")
	w.state = StateComplete
	return nil
}

// handleIssueSelect allows selecting a GitHub issue (stub for now).
func (w *Wizard) handleIssueSelect() error {
	fmt.Println("Issue select screen (stub)")
	w.state = StateComplete
	return nil
}

// handleTypeSelect allows selecting project type (stub for now).
func (w *Wizard) handleTypeSelect() error {
	fmt.Println("Type select screen (stub)")
	w.state = StateComplete
	return nil
}

// handleNameEntry allows entering project name (stub for now).
func (w *Wizard) handleNameEntry() error {
	fmt.Println("Name entry screen (stub)")
	w.state = StateComplete
	return nil
}

// handlePromptEntry allows entering initial prompt (stub for now).
func (w *Wizard) handlePromptEntry() error {
	fmt.Println("Prompt entry screen (stub)")
	w.state = StateComplete
	return nil
}

// handleProjectSelect allows selecting existing project to continue (stub for now).
func (w *Wizard) handleProjectSelect() error {
	fmt.Println("Project select screen (stub)")
	w.state = StateComplete
	return nil
}

// handleContinuePrompt allows entering additional prompt for continuing (stub for now).
func (w *Wizard) handleContinuePrompt() error {
	fmt.Println("Continue prompt screen (stub)")
	w.state = StateComplete
	return nil
}

// finalize creates/continues the project and launches Claude (stub for now).
func (w *Wizard) finalize() error {
	fmt.Println("Finalize: create/continue project and launch Claude")
	fmt.Printf("Choices: %+v\n", w.choices)
	return nil
}
