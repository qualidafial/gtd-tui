package gtd

import (
	"context"
	"time"
)

type ProjectStatus string

const (
	ProjectStatusOpen    ProjectStatus = "open"
	ProjectStatusSomeday ProjectStatus = "someday"
	ProjectStatusDone    ProjectStatus = "done"
	ProjectStatusDropped ProjectStatus = "dropped"
)

type Project struct {
	ID          int64
	Title       string
	Outcome     string
	Description string
	Due         *time.Time
	Status      ProjectStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
	// StatusChangedAt records when the project last entered its current status:
	// equal to CreatedAt on creation (the transition into open), then
	// overwritten by the supplied instant on every Complete/Drop/Park/Reopen.
	StatusChangedAt time.Time
}

type ProjectTaskCounts struct {
	Complete int // done tasks (non-dropped, non-pending)
	Total    int // non-dropped tasks
}

type ProjectFilter struct {
	Status *ProjectStatus
	Search []string
}

func (f ProjectFilter) WithStatus(s ProjectStatus) ProjectFilter {
	f.Status = &s
	return f
}

type ProjectService interface {
	GetProject(ctx context.Context, id int64) (Project, error)
	ListProjects(ctx context.Context, filter ProjectFilter) ([]Project, error)
	CreateProject(ctx context.Context, project Project) (Project, error)
	UpdateProject(ctx context.Context, project Project) (Project, error)
	// CompleteProject transitions the project to done. When cascade is true,
	// pending tasks are marked done; when false, they are detached (ProjectID
	// set to nil). The at instant stamps the project's StatusChangedAt and any
	// cascaded task's StatusChangedAt.
	CompleteProject(ctx context.Context, id int64, cascade bool, at time.Time) (Project, error)
	// DropProject transitions the project to dropped, with the same cascade
	// semantics as CompleteProject.
	DropProject(ctx context.Context, id int64, cascade bool, at time.Time) (Project, error)
	// ParkProject transitions the project to someday without changing task
	// statuses; tasks are filtered from default views by query logic.
	ParkProject(ctx context.Context, id int64, at time.Time) (Project, error)
	// ReopenProject restores a someday/done/dropped project to open without
	// changing task statuses. Mirrors ReopenTask.
	ReopenProject(ctx context.Context, id int64, at time.Time) (Project, error)
	// MoveProjectUp / MoveProjectDown shift a project one slot within projects
	// of the same status that also match filter. The moving project's status
	// group is always the universe; filter narrows further.
	MoveProjectUp(ctx context.Context, id int64, filter ProjectFilter) error
	MoveProjectDown(ctx context.Context, id int64, filter ProjectFilter) error
	CountTasksByProjects(ctx context.Context, projectIDs []int64) (map[int64]ProjectTaskCounts, error)
	// ConvertTaskToProject promotes a standalone task into a new open project,
	// keeping the task as the project's first action. Project Title/Description
	// default from the task when empty; the task is re-parented and its
	// Title/Description replaced with the reframed values. Returns the created
	// project and the updated task.
	ConvertTaskToProject(ctx context.Context, taskID int64, project Project, reframed Task) (Project, Task, error)
	// ConvertProjectToTask collapses an empty open project into a standalone
	// task, inheriting Title/Description/Due and folding any Outcome into the
	// Description, then deletes the project. Rejects non-open or non-empty
	// projects. Returns the created task.
	ConvertProjectToTask(ctx context.Context, projectID int64) (Task, error)
	// LinkTaskToProject re-parents a standalone task into a project, placing it
	// at the bottom of the project's task ordering. Rejects non-standalone tasks
	// and invalid projects. Returns the updated task.
	LinkTaskToProject(ctx context.Context, taskID, projectID int64) (Task, error)
}
