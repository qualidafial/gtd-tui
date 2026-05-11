package service

import (
	"context"

	gtd "github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/sqlite"
)

type ProjectService struct {
	db *sqlite.DB
}

func NewProjectService(db *sqlite.DB) *ProjectService {
	return &ProjectService{db: db}
}

func (s *ProjectService) Project(ctx context.Context, id int64) (gtd.Project, error) {
	return s.db.Project(ctx, id)
}

func (s *ProjectService) Projects(ctx context.Context, filter gtd.ProjectFilter) ([]gtd.Project, error) {
	return s.db.Projects(ctx, filter)
}

func (s *ProjectService) CreateProject(ctx context.Context, project gtd.Project) (gtd.Project, error) {
	return s.db.CreateProject(ctx, project)
}

func (s *ProjectService) UpdateProject(ctx context.Context, project gtd.Project) (gtd.Project, error) {
	return s.db.UpdateProject(ctx, project)
}

func (s *ProjectService) DeleteProject(ctx context.Context, id int64) error {
	return s.db.DeleteProject(ctx, id)
}

var _ gtd.ProjectService = (*ProjectService)(nil)
