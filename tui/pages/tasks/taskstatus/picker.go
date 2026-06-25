package taskstatus

import (
	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/components/form/selectfield"
)

// statusLabel renders a task status for display in the picker.
func statusLabel(s gtd.TaskStatus) string {
	switch s {
	case gtd.TaskStatusOpen:
		return "Open"
	case gtd.TaskStatusDone:
		return "Done"
	case gtd.TaskStatusDropped:
		return "Dropped"
	default:
		return string(s)
	}
}

// optionsFor returns the picker options for a task with the given current
// status: the current status first (so it is preselected), followed by the
// statuses reachable from it. Open reaches done and dropped; a closed task
// (done or dropped) reaches only open.
func optionsFor(current gtd.TaskStatus) []selectfield.Option[gtd.TaskStatus] {
	var targets []gtd.TaskStatus
	switch current {
	case gtd.TaskStatusOpen:
		targets = []gtd.TaskStatus{gtd.TaskStatusDone, gtd.TaskStatusDropped}
	default:
		targets = []gtd.TaskStatus{gtd.TaskStatusOpen}
	}

	statuses := append([]gtd.TaskStatus{current}, targets...)
	opts := make([]selectfield.Option[gtd.TaskStatus], len(statuses))
	for i, s := range statuses {
		opts[i] = selectfield.Option[gtd.TaskStatus]{Display: statusLabel(s), Value: s}
	}
	return opts
}

// transitionFor maps a (current, target) status pair to the transition that
// effects it. ok is false when target is not a valid distinct target for
// current (including target == current).
func transitionFor(current, target gtd.TaskStatus) (Transition, bool) {
	if current == target {
		return 0, false
	}
	switch target {
	case gtd.TaskStatusDone:
		if current == gtd.TaskStatusOpen {
			return Complete, true
		}
	case gtd.TaskStatusDropped:
		if current == gtd.TaskStatusOpen {
			return Drop, true
		}
	case gtd.TaskStatusOpen:
		return Reopen, true
	}
	return 0, false
}
