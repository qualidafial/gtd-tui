package gtd

// import (
// 	"context"
// 	"time"
// )

// type ProjectStatus string

// const (
// 	ProjectStatusActive   ProjectStatus = "active"
// 	ProjectStatusDeferred ProjectStatus = "deferred"
// 	ProjectStatusSomeday  ProjectStatus = "someday"
// 	ProjectStatusDone     ProjectStatus = "done"
// 	ProjectStatusDropped  ProjectStatus = "dropped"
// )

// type Project struct {
// 	ID          int64
// 	Title       string
// 	Outcome     string
// 	Description string
// 	Status      ProjectStatus
// 	Due         *time.Time
// 	CreatedAt   time.Time
// 	UpdatedAt   time.Time
// }

// type ProjectFilter struct {
// 	Status     *ProjectStatus
// 	ProjectIDs []int64
// 	TaskIDs    []int64
// 	Query      string
// }

// type ProjectService interface {
// 	Project(ctx context.Context, id int64) (Project, error)
// 	Projects(ctx context.Context, filter ProjectFilter) ([]Project, error)
// 	CreateProject(ctx context.Context, project Project) (Project, error)
// 	UpdateProject(ctx context.Context, project Project) (Project, error)
// 	DeleteProject(ctx context.Context, id int64) error
// }
