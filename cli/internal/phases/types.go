package phases

// ProjectInfo holds minimal project information needed for template rendering.
//
// This is passed to phases during construction so they can render templates
// with project context. Phases only receive this minimal info, not the entire
// project state, maintaining clean separation of concerns.
type ProjectInfo struct {
	Name        string
	Description string
	Branch      string
}
