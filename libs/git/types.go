package git

// Label represents a GitHub label.
type Label struct {
	Name string `json:"name"`
}

// Issue represents a GitHub issue.
type Issue struct {
	Number int     `json:"number"`
	Title  string  `json:"title"`
	Body   string  `json:"body"`
	State  string  `json:"state"`
	URL    string  `json:"url"`
	Labels []Label `json:"labels"`
}

// HasLabel checks if an issue has a specific label.
func (i *Issue) HasLabel(label string) bool {
	for _, l := range i.Labels {
		if l.Name == label {
			return true
		}
	}
	return false
}

// LinkedBranch represents a branch linked to an issue.
type LinkedBranch struct {
	Name string
	URL  string
}
