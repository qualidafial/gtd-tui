package gtd

// IsStandalone reports whether the task belongs to no project. Standalone tasks
// are the only ones eligible to be promoted into a project (ConvertTaskToProject)
// or linked into one (LinkTaskToProject).
func IsStandalone(t Task) bool {
	return t.ProjectID == nil
}

// CanConvertProjectToTask reports whether a project may collapse into a single
// standalone task: it must be open and hold zero tasks of any status (the only
// lossless case). taskCount is the number of tasks attached to the project
// across every status.
func CanConvertProjectToTask(p Project, taskCount int) bool {
	return p.Status == ProjectStatusOpen && taskCount == 0
}
