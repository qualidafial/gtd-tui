package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/qualidafial/gtd-tui"
)

var taskColumns = []string{
	"t.id", "t.title", "t.description", "t.status",
	"t.due", "t.defer_until", "t.created_at", "t.updated_at",
}

func (d *DB) Task(ctx context.Context, id int64) (gtd.Task, error) {
	query, args, err := sq.Select(taskColumns...).From("tasks t").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return gtd.Task{}, err
	}
	task, err := scanTask(d.db.QueryRowContext(ctx, query, args...))
	if err == sql.ErrNoRows {
		return gtd.Task{}, fmt.Errorf("task %d: not found", id)
	}
	if err != nil {
		return gtd.Task{}, fmt.Errorf("task %d: %w", id, err)
	}
	return task, nil
}

func (d *DB) Tasks(ctx context.Context, filter gtd.TaskFilter) ([]gtd.Task, error) {
	q := sq.Select(taskColumns...).
		From("tasks t")
	if filter.Statuses != nil {
		q = q.Where(sq.Eq{"status": filter.Statuses})
	}
	// if len(filter.ProjectIDs) > 0 {
	// 	q = q.Join("project_tasks pt ON tasks.id = pt.task_id")
	// 	q = q.Where(sq.Eq{"pt.project_id": filter.ProjectIDs})
	// }
	if len(filter.TaskIDs) > 0 {
		q = q.Where(sq.Eq{"t.task_id": filter.TaskIDs})
	}
	q = q.OrderBy("t.created_at ASC")

	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("tasks: %w", err)
	}
	defer rows.Close()

	var tasks []gtd.Task
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("tasks: %w", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (d *DB) CreateTask(ctx context.Context, task gtd.Task) (gtd.Task, error) {
	now := time.Now().UTC()
	task.CreatedAt = now
	task.UpdatedAt = now

	query, args, err := sq.Insert("tasks").
		Columns("title", "description", "status", "due", "defer_until", "created_at", "updated_at").
		Values(task.Title, task.Description, string(task.Status),
			nullTime(task.Due), nullTime(task.DeferUntil),
			task.CreatedAt, task.UpdatedAt).
		ToSql()
	if err != nil {
		return gtd.Task{}, err
	}
	res, err := d.db.ExecContext(ctx, query, args...)
	if err != nil {
		return gtd.Task{}, fmt.Errorf("create task: %w", err)
	}
	task.ID, err = res.LastInsertId()
	if err != nil {
		return gtd.Task{}, fmt.Errorf("create task: %w", err)
	}
	return task, nil
}

func (d *DB) UpdateTask(ctx context.Context, task gtd.Task) (gtd.Task, error) {
	task.UpdatedAt = time.Now().UTC()

	query, args, err := sq.Update("tasks").
		Set("title", task.Title).
		Set("description", task.Description).
		Set("status", string(task.Status)).
		Set("due", nullTime(task.Due)).
		Set("defer_until", nullTime(task.DeferUntil)).
		Set("updated_at", task.UpdatedAt).
		Where(sq.Eq{"id": task.ID}).
		ToSql()
	if err != nil {
		return gtd.Task{}, err
	}
	if _, err := d.db.ExecContext(ctx, query, args...); err != nil {
		return gtd.Task{}, fmt.Errorf("update task %d: %w", task.ID, err)
	}
	return task, nil
}

func (d *DB) DropTask(ctx context.Context, id int64) (gtd.Task, error) {
	now := time.Now().UTC()
	query, args, err := sq.Update("tasks").
		Set("status", string(gtd.TaskStatusDropped)).
		Set("updated_at", now).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING id, title, description, status, due, defer_until, created_at, updated_at").
		ToSql()
	if err != nil {
		return gtd.Task{}, err
	}
	task, err := scanTask(d.db.QueryRowContext(ctx, query, args...))
	if err == sql.ErrNoRows {
		return gtd.Task{}, fmt.Errorf("drop task %d: not found", id)
	}
	if err != nil {
		return gtd.Task{}, fmt.Errorf("drop task %d: %w", id, err)
	}
	return task, nil
}

func (d *DB) DeleteTask(ctx context.Context, id int64) error {
	query, args, err := sq.Delete("tasks").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return err
	}
	if _, err := d.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("delete task %d: %w", id, err)
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTask(s scanner) (gtd.Task, error) {
	var task gtd.Task
	var due, deferUntil sql.NullTime
	err := s.Scan(
		&task.ID, &task.Title, &task.Description, &task.Status,
		&due, &deferUntil, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		return gtd.Task{}, err
	}
	if due.Valid {
		task.Due = &due.Time
	}
	if deferUntil.Valid {
		task.DeferUntil = &deferUntil.Time
	}
	return task, nil
}

func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t.UTC(), Valid: true}
}
