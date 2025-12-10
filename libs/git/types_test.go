package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIssue_HasLabel(t *testing.T) {
	tests := []struct {
		name   string
		labels []Label
		search string
		want   bool
	}{
		{
			name:   "returns true when label exists",
			labels: []Label{{Name: "bug"}, {Name: "sow"}},
			search: "sow",
			want:   true,
		},
		{
			name:   "returns false when label does not exist",
			labels: []Label{{Name: "bug"}, {Name: "feature"}},
			search: "sow",
			want:   false,
		},
		{
			name:   "handles empty labels slice",
			labels: []Label{},
			search: "sow",
			want:   false,
		},
		{
			name:   "handles nil labels slice",
			labels: nil,
			search: "sow",
			want:   false,
		},
		{
			name:   "is case-sensitive",
			labels: []Label{{Name: "SOW"}},
			search: "sow",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issue := &Issue{
				Number: 123,
				Title:  "Test Issue",
				Labels: tt.labels,
			}

			got := issue.HasLabel(tt.search)

			assert.Equal(t, tt.want, got)
		})
	}
}
