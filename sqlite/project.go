package sqlite

// import (
// 	"context"
// 	"database/sql"
// 	"fmt"
// 	"time"

// 	sq "github.com/Masterminds/squirrel"
// 	"github.com/qualidafial/gtd-tui"
// )

// var projectColumns = []string{
// 	"id", "title", "outcome", "description", "status", "due", "created_at", "updated_at",
// }

// func (d *DB) Project(ctx context.Context, id int64) (gtd.Project, error) {
// 	query, args, err := sq.Select(projectColumns...).From("projects").Where(sq.Eq{"id": id}).ToSql()
// 	if err != nil {
// 		return gtd.Project{}, err
// 	}
// 	project, err := scanProject(d.db.QueryRowContext(ctx, query, args...))
// 	if err == sql.ErrNoRows {
// 		return gtd.Project{}, fmt.Errorf("project %d: not found", id)
// 	}
// 	if err != nil {
// 		return gtd.Project{}, fmt.Errorf("project %d: %w", id, err)
// 	}
// 	return project, nil
// }

// func (d *DB) Projects(ctx context.Context, filter gtd.ProjectFilter) ([]gtd.Project, error) {
// 	q := sq.Select(projectColumns...).From("projects")
// 	if filter.Status != nil {
// 		q = q.Where(sq.Eq{"status": string(*filter.Status)})
// 	}
// 	if filter.Query != "" {
// 		pattern := "%" + filter.Query + "%"
// 		q = q.Where(sq.Or{
// 			sq.Like{"title": pattern},
// 			sq.Like{"outcome": pattern},
// 		})
// 	}
// 	q = q.OrderBy("title ASC")

// 	query, args, err := q.ToSql()
// 	if err != nil {
// 		return nil, err
// 	}
// 	rows, err := d.db.QueryContext(ctx, query, args...)
// 	if err != nil {
// 		return nil, fmt.Errorf("projects: %w", err)
// 	}
// 	defer rows.Close()

// 	var projects []gtd.Project
// 	for rows.Next() {
// 		project, err := scanProject(rows)
// 		if err != nil {
// 			return nil, fmt.Errorf("projects: %w", err)
// 		}
// 		projects = append(projects, project)
// 	}
// 	return projects, rows.Err()
// }

// func (d *DB) CreateProject(ctx context.Context, project gtd.Project) (gtd.Project, error) {
// 	now := time.Now().UTC()
// 	project.CreatedAt = now
// 	project.UpdatedAt = now

// 	query, args, err := sq.Insert("projects").
// 		Columns("title", "outcome", "description", "status", "due", "created_at", "updated_at").
// 		Values(project.Title, project.Outcome, project.Description, string(project.Status),
// 			nullTime(project.Due), project.CreatedAt, project.UpdatedAt).
// 		ToSql()
// 	if err != nil {
// 		return gtd.Project{}, err
// 	}
// 	res, err := d.db.ExecContext(ctx, query, args...)
// 	if err != nil {
// 		return gtd.Project{}, fmt.Errorf("create project: %w", err)
// 	}
// 	project.ID, err = res.LastInsertId()
// 	if err != nil {
// 		return gtd.Project{}, fmt.Errorf("create project: %w", err)
// 	}
// 	return project, nil
// }

// func (d *DB) UpdateProject(ctx context.Context, project gtd.Project) (gtd.Project, error) {
// 	project.UpdatedAt = time.Now().UTC()

// 	query, args, err := sq.Update("projects").
// 		Set("title", project.Title).
// 		Set("outcome", project.Outcome).
// 		Set("description", project.Description).
// 		Set("status", string(project.Status)).
// 		Set("due", nullTime(project.Due)).
// 		Set("updated_at", project.UpdatedAt).
// 		Where(sq.Eq{"id": project.ID}).
// 		ToSql()
// 	if err != nil {
// 		return gtd.Project{}, err
// 	}
// 	if _, err := d.db.ExecContext(ctx, query, args...); err != nil {
// 		return gtd.Project{}, fmt.Errorf("update project %d: %w", project.ID, err)
// 	}
// 	return project, nil
// }

// func (d *DB) DeleteProject(ctx context.Context, id int64) error {
// 	query, args, err := sq.Delete("projects").Where(sq.Eq{"id": id}).ToSql()
// 	if err != nil {
// 		return err
// 	}
// 	if _, err := d.db.ExecContext(ctx, query, args...); err != nil {
// 		return fmt.Errorf("delete project %d: %w", id, err)
// 	}
// 	return nil
// }

// func scanProject(s scanner) (gtd.Project, error) {
// 	var project gtd.Project
// 	var due sql.NullTime
// 	err := s.Scan(
// 		&project.ID, &project.Title, &project.Outcome, &project.Description,
// 		&project.Status, &due, &project.CreatedAt, &project.UpdatedAt,
// 	)
// 	if err != nil {
// 		return gtd.Project{}, err
// 	}
// 	if due.Valid {
// 		project.Due = &due.Time
// 	}
// 	return project, nil
// }
