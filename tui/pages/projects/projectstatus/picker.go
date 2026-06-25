package projectstatus

import (
	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/components/form/selectfield"
)

// statusLabel renders a project status for display in the picker.
func statusLabel(s gtd.ProjectStatus) string {
	switch s {
	case gtd.ProjectStatusOpen:
		return "Open"
	case gtd.ProjectStatusSomeday:
		return "Someday"
	case gtd.ProjectStatusDone:
		return "Done"
	case gtd.ProjectStatusDropped:
		return "Dropped"
	default:
		return string(s)
	}
}

// optionsFor returns the picker options for a project with the given current
// status: the current status first (preselected), then the statuses reachable
// from it. Open reaches someday/done/dropped; someday reaches open/dropped; a
// closed project (done or dropped) reaches only open.
func optionsFor(current gtd.ProjectStatus) []selectfield.Option[gtd.ProjectStatus] {
	var targets []gtd.ProjectStatus
	switch current {
	case gtd.ProjectStatusOpen:
		targets = []gtd.ProjectStatus{gtd.ProjectStatusSomeday, gtd.ProjectStatusDone, gtd.ProjectStatusDropped}
	case gtd.ProjectStatusSomeday:
		targets = []gtd.ProjectStatus{gtd.ProjectStatusOpen, gtd.ProjectStatusDropped}
	default:
		targets = []gtd.ProjectStatus{gtd.ProjectStatusOpen}
	}

	statuses := append([]gtd.ProjectStatus{current}, targets...)
	opts := make([]selectfield.Option[gtd.ProjectStatus], len(statuses))
	for i, s := range statuses {
		opts[i] = selectfield.Option[gtd.ProjectStatus]{Display: statusLabel(s), Value: s}
	}
	return opts
}

// transitionFor maps a (current, target) status pair to the transition that
// effects it. ok is false when target is not a valid distinct target for
// current (including target == current).
func transitionFor(current, target gtd.ProjectStatus) (Transition, bool) {
	if current == target {
		return 0, false
	}
	switch target {
	case gtd.ProjectStatusDone:
		if current == gtd.ProjectStatusOpen {
			return Complete, true
		}
	case gtd.ProjectStatusDropped:
		if current == gtd.ProjectStatusOpen || current == gtd.ProjectStatusSomeday {
			return Drop, true
		}
	case gtd.ProjectStatusSomeday:
		if current == gtd.ProjectStatusOpen {
			return Park, true
		}
	case gtd.ProjectStatusOpen:
		return Reopen, true
	}
	return 0, false
}
