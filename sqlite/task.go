package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/internal/orderkey"
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
	// Open tasks sort by order_key; closed (done/dropped) tasks have a
	// NULL order_key and fall through to updated_at descending.
	q = q.OrderBy(
		"t.order_key ASC NULLS LAST",
		"t.updated_at DESC",
		"t.id ASC",
	)

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

	var key any
	if !isClosedStatus(task.Status) {
		k, err := d.nextOrderKey(ctx, task.Status)
		if err != nil {
			return gtd.Task{}, fmt.Errorf("create task: %w", err)
		}
		key = k
	}

	query, args, err := sq.Insert("tasks").
		Columns("title", "description", "status", "due", "defer_until", "created_at", "updated_at", "order_key").
		Values(task.Title, task.Description, string(task.Status),
			nullTime(task.Due), nullTime(task.DeferUntil),
			task.CreatedAt, task.UpdatedAt, key).
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

	err := d.RunTx(ctx, func(ctx context.Context, tx *DB) error {
		currentStatus, _, err := tx.taskStatusAndKey(ctx, task.ID)
		if err != nil {
			return err
		}

		update := sq.Update("tasks").
			Set("title", task.Title).
			Set("description", task.Description).
			Set("status", string(task.Status)).
			Set("due", nullTime(task.Due)).
			Set("defer_until", nullTime(task.DeferUntil)).
			Set("updated_at", task.UpdatedAt).
			Where(sq.Eq{"id": task.ID})

		if task.Status != currentStatus {
			if isClosedStatus(task.Status) {
				update = update.Set("order_key", nil)
			} else {
				key, err := tx.nextOrderKey(ctx, task.Status)
				if err != nil {
					return err
				}
				update = update.Set("order_key", key)
			}
		}

		query, args, err := update.ToSql()
		if err != nil {
			return err
		}
		if _, err := tx.db.ExecContext(ctx, query, args...); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return gtd.Task{}, fmt.Errorf("update task %d: %w", task.ID, err)
	}
	return task, nil
}

func (d *DB) DropTask(ctx context.Context, id int64) (gtd.Task, error) {
	now := time.Now().UTC()
	query, args, err := sq.Update("tasks").
		Set("status", string(gtd.TaskStatusDropped)).
		Set("updated_at", now).
		Set("order_key", nil).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING id, title, description, status, due, defer_until, created_at, updated_at").
		ToSql()
	if err != nil {
		return gtd.Task{}, err
	}
	task, err := scanTask(d.db.QueryRowContext(ctx, query, args...))
	if errors.Is(err, sql.ErrNoRows) {
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

// MoveUp shifts the task one slot earlier within its status group.
// No-op when already at the top.
func (d *DB) MoveUp(ctx context.Context, id int64) error {
	return d.RunTx(ctx, func(ctx context.Context, tx *DB) error {
		return tx.shiftTask(ctx, id, -1)
	})
}

// MoveDown shifts the task one slot later within its status group.
// No-op when already at the bottom.
func (d *DB) MoveDown(ctx context.Context, id int64) error {
	return d.RunTx(ctx, func(ctx context.Context, tx *DB) error {
		return tx.shiftTask(ctx, id, +1)
	})
}

// shiftTask moves id by delta slots within its status group. The fast
// path swaps a single key via orderkey.Between; on exhaustion the whole
// status group is renumbered with id placed at its new index.
func (d *DB) shiftTask(ctx context.Context, id int64, delta int) error {
	status, key, err := d.taskStatusAndKey(ctx, id)
	if err != nil {
		return err
	}
	if isClosedStatus(status) {
		return fmt.Errorf("task %d: cannot reorder %s tasks", id, status)
	}
	others, err := d.statusOrder(ctx, status, id)
	if err != nil {
		return err
	}

	// Position of id among all peers = count of peers with smaller key.
	pos := 0
	for _, o := range others {
		if o.key < key {
			pos++
		}
	}
	newPos := pos + delta
	if newPos < 0 || newPos > len(others) {
		return nil // already at the edge
	}

	var prevKey, nextKey string
	if newPos > 0 {
		prevKey = others[newPos-1].key
	}
	if newPos < len(others) {
		nextKey = others[newPos].key
	}

	if newKey, ok := orderkey.Between(prevKey, nextKey); ok {
		return d.setOrderKey(ctx, id, newKey)
	}

	keys := orderkey.Renumber(len(others) + 1)
	for i, t := range others[:newPos] {
		if err := d.setOrderKey(ctx, t.id, keys[i]); err != nil {
			return err
		}
	}
	if err := d.setOrderKey(ctx, id, keys[newPos]); err != nil {
		return err
	}
	for i, t := range others[newPos:] {
		if err := d.setOrderKey(ctx, t.id, keys[newPos+1+i]); err != nil {
			return err
		}
	}
	return nil
}

func isClosedStatus(s gtd.TaskStatus) bool {
	return s == gtd.TaskStatusDone || s == gtd.TaskStatusDropped
}

type statusEntry struct {
	id  int64
	key string
}

// statusOrder returns id/order_key pairs for every task in the given
// status, ordered by order_key, excluding excludeID when non-zero.
func (d *DB) statusOrder(ctx context.Context, status gtd.TaskStatus, excludeID int64) ([]statusEntry, error) {
	q := sq.Select("id", "order_key").
		From("tasks").
		Where(sq.Eq{"status": string(status)}).
		OrderBy("order_key ASC", "id ASC")
	if excludeID != 0 {
		q = q.Where(sq.NotEq{"id": excludeID})
	}
	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []statusEntry
	for rows.Next() {
		var e statusEntry
		if err := rows.Scan(&e.id, &e.key); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (d *DB) nextOrderKey(ctx context.Context, status gtd.TaskStatus) (string, error) {
	var maxKey sql.NullString
	query, args, err := sq.Select("MAX(order_key)").From("tasks").
		Where(sq.Eq{"status": string(status)}).ToSql()
	if err != nil {
		return "", err
	}
	if err := d.db.QueryRowContext(ctx, query, args...).Scan(&maxKey); err != nil {
		return "", err
	}
	key, ok := orderkey.Between(maxKey.String, "")
	if !ok {
		return "", fmt.Errorf("order keys exhausted for status %q", status)
	}
	return key, nil
}

func (d *DB) taskStatusAndKey(ctx context.Context, id int64) (gtd.TaskStatus, string, error) {
	query, args, err := sq.Select("status", "order_key").From("tasks").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return "", "", err
	}
	var status string
	var key sql.NullString
	err = d.db.QueryRowContext(ctx, query, args...).Scan(&status, &key)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", fmt.Errorf("task %d: not found", id)
	}
	if err != nil {
		return "", "", err
	}
	return gtd.TaskStatus(status), key.String, nil
}

func (d *DB) setOrderKey(ctx context.Context, id int64, key string) error {
	query, args, err := sq.Update("tasks").
		Set("order_key", key).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}
	res, err := d.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("task %d: not found", id)
	}
	return nil
}

func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t.UTC(), Valid: true}
}
