package taskstatus

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qualidafial/gtd-tui"
)

func optionValues(current gtd.TaskStatus) []gtd.TaskStatus {
	opts := optionsFor(current)
	vals := make([]gtd.TaskStatus, len(opts))
	for i, o := range opts {
		vals[i] = o.Value
	}
	return vals
}

func TestOptionsFor(t *testing.T) {
	tests := []struct {
		current gtd.TaskStatus
		want    []gtd.TaskStatus
	}{
		{gtd.TaskStatusOpen, []gtd.TaskStatus{gtd.TaskStatusOpen, gtd.TaskStatusDone, gtd.TaskStatusDropped}},
		{gtd.TaskStatusDone, []gtd.TaskStatus{gtd.TaskStatusDone, gtd.TaskStatusOpen}},
		{gtd.TaskStatusDropped, []gtd.TaskStatus{gtd.TaskStatusDropped, gtd.TaskStatusOpen}},
	}
	for _, tt := range tests {
		t.Run(string(tt.current), func(t *testing.T) {
			got := optionValues(tt.current)
			assert.Equal(t, tt.want, got)
			// The current status is always first so the picker preselects it.
			assert.Equal(t, tt.current, got[0], "current status should be first/preselected")
		})
	}
}

func TestTransitionFor(t *testing.T) {
	tests := []struct {
		current, target gtd.TaskStatus
		want            Transition
		ok              bool
	}{
		{gtd.TaskStatusOpen, gtd.TaskStatusDone, Complete, true},
		{gtd.TaskStatusOpen, gtd.TaskStatusDropped, Drop, true},
		{gtd.TaskStatusDone, gtd.TaskStatusOpen, Reopen, true},
		{gtd.TaskStatusDropped, gtd.TaskStatusOpen, Reopen, true},
		// No-op / invalid pairs.
		{gtd.TaskStatusOpen, gtd.TaskStatusOpen, 0, false},
		{gtd.TaskStatusDone, gtd.TaskStatusDropped, 0, false},
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
