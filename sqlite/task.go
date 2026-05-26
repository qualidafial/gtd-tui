package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/internal/orderkey"
)

var taskColumns = []string{
	"t.id", "t.title", "t.description", "t.kind", "t.status", "t.assignee", "t.project_id",
	"t.due", "t.defer_until", "t.created_at", "t.updated_at", "t.status_changed_at",
}

func (d *DB) GetTask(ctx context.Context, id int64) (gtd.Task, error) {
	query, args, err := sq.Select(taskColumns...).From("tasks t").Where(sq.Eq{"t.id": id}).ToSql()
	if err != nil {
		return gtd.Task{}, err
	}
	task, err := scanTask(d.db.QueryRowContext(ctx, query, args...))
	if errors.Is(err, sql.ErrNoRows) {
		return gtd.Task{}, fmt.Errorf("task %d: not found", id)
	}
	if err != nil {
		return gtd.Task{}, fmt.Errorf("task %d: %w", id, err)
	}
	return task, nil
}

func (d *DB) ListTasks(ctx context.Context, filter gtd.TaskFilter) ([]gtd.Task, error) {
	q := sq.Select(taskColumns...).From("tasks t")

	if filter.Status != nil {
		q = q.Where(sq.Eq{"t.status": string(*filter.Status)})
	}
	if filter.Kind != nil {
		q = q.Where(sq.Eq{"t.kind": string(*filter.Kind)})
	}
	if len(filter.TaskIDs) > 0 {
		q = q.Where(sq.Eq{"t.id": filter.TaskIDs})
	}
	if filter.ProjectID != nil {
		q = q.Where(sq.Eq{"t.project_id": *filter.ProjectID})
	}
	if !filter.IncludeSomedayProjects {
		q = q.LeftJoin("projects p ON p.id = t.project_id").
			Where("(p.status IS NULL OR p.status != ?)", string(gtd.ProjectStatusSomeday))
	}
	if filter.Assignee != nil {
		q = q.Where("lower(t.assignee) LIKE ?", likeContains(*filter.Assignee))
	}
	for _, term := range filter.Search {
		pattern := likeContains(term)
		q = q.Where("(lower(t.title) LIKE ? OR lower(t.description) LIKE ? OR lower(t.assignee) LIKE ?)",
			pattern, pattern, pattern)
	}
	q = applyDatePredicate(q, "t.due", filter.Due)
	q = applyDatePredicate(q, "t.defer_until", filter.Ready)
	q = applyDatePredicate(q, "t.defer_until", filter.Defer)

	// Pending tasks sort by order_key; closed (done/dropped) tasks have a
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

// likeContains builds a case-insensitive LIKE pattern matching term as a
// substring. The column side is lowercased by the caller; term is lowercased here.
func likeContains(term string) string {
	return "%" + strings.ToLower(term) + "%"
}

// applyDatePredicate adds the SQL constraint for p against column, if non-nil.
func applyDatePredicate(q sq.SelectBuilder, column string, p *gtd.DatePredicate) sq.SelectBuilder {
	if p == nil {
		return q
	}
	switch p.Kind {
	case gtd.OnOrBefore:
		return q.Where(fmt.Sprintf("(%s IS NOT NULL AND %s <= ?)", column, column), p.Time.UTC())
	case gtd.AvailableAsOf:
		return q.Where(fmt.Sprintf("(%s IS NULL OR %s <= ?)", column, column), p.Time.UTC())
	case gtd.After:
		return q.Where(fmt.Sprintf("%s > ?", column), p.Time.UTC())
	case gtd.IsNull:
		return q.Where(fmt.Sprintf("%s IS NULL", column))
	case gtd.IsNotNull:
		return q.Where(fmt.Sprintf("%s IS NOT NULL", column))
	}
	return q
}

func (d *DB) CreateTask(ctx context.Context, task gtd.Task) (gtd.Task, error) {
	now := time.Now().UTC()
	task.CreatedAt = now
	task.UpdatedAt = now
	task.StatusChangedAt = now

	var key any
	if !isClosedStatus(task.Status) {
		k, err := d.nextOrderKey(ctx)
		if err != nil {
			return gtd.Task{}, fmt.Errorf("create task: %w", err)
		}
		key = k
	}

	query, args, err := sq.Insert("tasks").
		Columns("title", "description", "kind", "status", "assignee", "project_id", "due", "defer_until", "created_at", "updated_at", "status_changed_at", "order_key").
		Values(task.Title, task.Description, string(task.Kind), string(task.Status), task.Assignee,
			nullInt64(task.ProjectID), nullTime(task.Due), nullTime(task.DeferUntil),
			task.CreatedAt, task.UpdatedAt, task.StatusChangedAt, key).
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
		current, err := tx.GetTask(ctx, task.ID)
		if err != nil {
			return err
		}
		if task.Status != current.Status {
			return fmt.Errorf("task %d: UpdateTask cannot change status; use CompleteTask/DropTask/ReopenTask", task.ID)
		}

		query, args, err := sq.Update("tasks").
			Set("title", task.Title).
			Set("description", task.Description).
			Set("kind", string(task.Kind)).
			Set("assignee", task.Assignee).
			Set("project_id", nullInt64(task.ProjectID)).
			Set("due", nullTime(task.Due)).
			Set("defer_until", nullTime(task.DeferUntil)).
			Set("updated_at", task.UpdatedAt).
			Where(sq.Eq{"id": task.ID}).
			ToSql()
		if err != nil {
			return err
		}
		_, err = tx.db.ExecContext(ctx, query, args...)
		return err
	})
	if err != nil {
		return gtd.Task{}, fmt.Errorf("update task %d: %w", task.ID, err)
	}
	return task, nil
}

func (d *DB) CompleteTask(ctx context.Context, id int64, at time.Time) (gtd.Task, error) {
	return d.transitionTask(ctx, id, at, gtd.TaskStatusDone, gtd.TaskStatusPending)
}

func (d *DB) DropTask(ctx context.Context, id int64, at time.Time) (gtd.Task, error) {
	return d.transitionTask(ctx, id, at, gtd.TaskStatusDropped, gtd.TaskStatusPending)
}

func (d *DB) ReopenTask(ctx context.Context, id int64, at time.Time) (gtd.Task, error) {
	return d.transitionTask(ctx, id, at, gtd.TaskStatusPending, gtd.TaskStatusDone, gtd.TaskStatusDropped)
}

// transitionTask atomically validates the current status is one of allowedFrom,
// then sets it to newStatus. updated_at tracks record time (now); the supplied
// at is the event time stored in status_changed_at. Clears order_key for closed
// transitions; assigns a fresh key when reopening to pending.
func (d *DB) transitionTask(ctx context.Context, id int64, at time.Time, newStatus gtd.TaskStatus, allowedFrom ...gtd.TaskStatus) (gtd.Task, error) {
	var task gtd.Task
	err := d.RunTx(ctx, func(ctx context.Context, tx *DB) error {
		current, err := tx.GetTask(ctx, id)
		if err != nil {
			return err
		}
		var allowed bool
		for _, s := range allowedFrom {
			if current.Status == s {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("task %d: cannot transition from %s to %s", id, current.Status, newStatus)
		}

		now := time.Now().UTC()
		statusChangedAt := at.UTC()
		update := sq.Update("tasks").
			Set("status", string(newStatus)).
			Set("updated_at", now).
			Set("status_changed_at", statusChangedAt).
			Where(sq.Eq{"id": id})

		if isClosedStatus(newStatus) {
			update = update.Set("order_key", nil)
		} else {
			key, err := tx.nextOrderKey(ctx)
			if err != nil {
				return err
			}
			update = update.Set("order_key", key)
		}

		query, args, err := update.ToSql()
		if err != nil {
			return err
		}
		if _, err := tx.db.ExecContext(ctx, query, args...); err != nil {
			return err
		}

		current.Status = newStatus
		current.UpdatedAt = now
		current.StatusChangedAt = statusChangedAt
		task = current
		return nil
	})
	if err != nil {
		return gtd.Task{}, fmt.Errorf("transition task %d: %w", id, err)
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
	var projectID sql.NullInt64
	err := s.Scan(
		&task.ID, &task.Title, &task.Description, &task.Kind, &task.Status, &task.Assignee, &projectID,
		&due, &deferUntil, &task.CreatedAt, &task.UpdatedAt, &task.StatusChangedAt,
	)
	if err != nil {
		return gtd.Task{}, err
	}
	if projectID.Valid {
		task.ProjectID = &projectID.Int64
	}
	if due.Valid {
		task.Due = &due.Time
	}
	if deferUntil.Valid {
		task.DeferUntil = &deferUntil.Time
	}
	return task, nil
}

func nullInt64(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}

// MoveUp shifts the task one slot earlier within pending tasks.
// No-op when already at the top.
func (d *DB) MoveUp(ctx context.Context, id int64) error {
	return d.RunTx(ctx, func(ctx context.Context, tx *DB) error {
		return tx.shiftTask(ctx, id, -1)
	})
}

// MoveDown shifts the task one slot later within pending tasks.
// No-op when already at the bottom.
func (d *DB) MoveDown(ctx context.Context, id int64) error {
	return d.RunTx(ctx, func(ctx context.Context, tx *DB) error {
		return tx.shiftTask(ctx, id, +1)
	})
}

// shiftTask moves id by delta slots within pending tasks. The fast path swaps
// a single key via orderkey.Between; on exhaustion the whole group is renumbered.
func (d *DB) shiftTask(ctx context.Context, id int64, delta int) error {
	task, err := d.GetTask(ctx, id)
	if err != nil {
		return err
	}
	if isClosedStatus(task.Status) {
		return fmt.Errorf("task %d: cannot reorder %s tasks", id, task.Status)
	}

	key, err := d.taskOrderKey(ctx, id)
	if err != nil {
		return err
	}
	others, err := d.pendingOrder(ctx, id)
	if err != nil {
		return err
	}

	pos := 0
	for _, o := range others {
		if o.key < key {
			pos++
		}
	}
	newPos := pos + delta
	if newPos < 0 || newPos > len(others) {
		return nil
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

// pendingOrder returns id/order_key pairs for all pending tasks ordered by
// order_key, excluding excludeID.
func (d *DB) pendingOrder(ctx context.Context, excludeID int64) ([]statusEntry, error) {
	q := sq.Select("id", "order_key").
		From("tasks").
		Where(sq.Eq{"status": string(gtd.TaskStatusPending)}).
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

func (d *DB) nextOrderKey(ctx context.Context) (string, error) {
	var maxKey sql.NullString
	query, args, err := sq.Select("MAX(order_key)").From("tasks").
		Where(sq.Eq{"status": string(gtd.TaskStatusPending)}).ToSql()
	if err != nil {
		return "", err
	}
	if err := d.db.QueryRowContext(ctx, query, args...).Scan(&maxKey); err != nil {
		return "", err
	}
	key, ok := orderkey.Between(maxKey.String, "")
	if !ok {
		return "", fmt.Errorf("order keys exhausted")
	}
	return key, nil
}

func (d *DB) taskOrderKey(ctx context.Context, id int64) (string, error) {
	query, args, err := sq.Select("order_key").From("tasks").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return "", err
	}
	var key sql.NullString
	err = d.db.QueryRowContext(ctx, query, args...).Scan(&key)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("task %d: not found", id)
	}
	if err != nil {
		return "", err
	}
	return key.String, nil
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
