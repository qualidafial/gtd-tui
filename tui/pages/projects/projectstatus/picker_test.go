package projectstatus

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qualidafial/gtd-tui"
)

func optionValues(current gtd.ProjectStatus) []gtd.ProjectStatus {
	opts := optionsFor(current)
	vals := make([]gtd.ProjectStatus, len(opts))
	for i, o := range opts {
		vals[i] = o.Value
	}
	return vals
}

func TestOptionsFor(t *testing.T) {
	tests := []struct {
		current gtd.ProjectStatus
		want    []gtd.ProjectStatus
	}{
		{gtd.ProjectStatusOpen, []gtd.ProjectStatus{gtd.ProjectStatusOpen, gtd.ProjectStatusSomeday, gtd.ProjectStatusDone, gtd.ProjectStatusDropped}},
		{gtd.ProjectStatusSomeday, []gtd.ProjectStatus{gtd.ProjectStatusSomeday, gtd.ProjectStatusOpen, gtd.ProjectStatusDropped}},
		{gtd.ProjectStatusDone, []gtd.ProjectStatus{gtd.ProjectStatusDone, gtd.ProjectStatusOpen}},
		{gtd.ProjectStatusDropped, []gtd.ProjectStatus{gtd.ProjectStatusDropped, gtd.ProjectStatusOpen}},
	}
	for _, tt := range tests {
		t.Run(string(tt.current), func(t *testing.T) {
			got := optionValues(tt.current)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.current, got[0], "current status should be first/preselected")
		})
	}
}

func TestTransitionFor(t *testing.T) {
	tests := []struct {
		current, target gtd.ProjectStatus
		want            Transition
		ok              bool
	}{
		{gtd.ProjectStatusOpen, gtd.ProjectStatusDone, Complete, true},
		{gtd.ProjectStatusOpen, gtd.ProjectStatusDropped, Drop, true},
		{gtd.ProjectStatusOpen, gtd.ProjectStatusSomeday, Park, true},
		{gtd.ProjectStatusSomeday, gtd.ProjectStatusOpen, Reopen, true},
		{gtd.ProjectStatusSomeday, gtd.ProjectStatusDropped, Drop, true},
		{gtd.ProjectStatusDone, gtd.ProjectStatusOpen, Reopen, true},
		{gtd.ProjectStatusDropped, gtd.ProjectStatusOpen, Reopen, true},
		// No-op / invalid pairs.
		{gtd.ProjectStatusOpen, gtd.ProjectStatusOpen, 0, false},
		{gtd.ProjectStatusDone, gtd.ProjectStatusDropped, 0, false},
	}
	for _, tt := range tests {
		t.Run(string(tt.current)+"->"+string(tt.target), func(t *testing.T) {
			got, ok := transitionFor(tt.current, tt.target)
			assert.Equal(t, tt.ok, ok)
			if tt.ok {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
