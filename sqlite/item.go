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
)

var itemColumns = []string{
	"id", "title", "description", "created_at", "updated_at",
	"clarified_into_task_id", "clarified_into_project_id", "discarded",
}

// CreateItem inserts a fresh inbox capture. Clarify pointers and the discarded
// flag are not accepted on create — items start in the inbox unconditionally
// and reach terminal state only via the service-layer clarify operations.
func (d *DB) CreateItem(ctx context.Context, item gtd.Item) (gtd.Item, error) {
	now := time.Now().UTC()
	item.CreatedAt = now
	item.UpdatedAt = now

	query, args, err := sq.Insert("items").
		Columns("title", "description", "created_at", "updated_at").
		Values(item.Title, item.Description, item.CreatedAt, item.UpdatedAt).
		ToSql()
	if err != nil {
		return gtd.Item{}, err
	}
	res, err := d.db.ExecContext(ctx, query, args...)
	if err != nil {
		return gtd.Item{}, fmt.Errorf("create item: %w", err)
	}
	item.ID, err = res.LastInsertId()
	if err != nil {
		return gtd.Item{}, fmt.Errorf("create item: %w", err)
	}
	return item, nil
}

// ListItems returns inbox items that have not been clarified or discarded,
// ordered by created_at ASC (FIFO).
func (d *DB) ListItems(ctx context.Context) ([]gtd.Item, error) {
	query, args, err := sq.Select(itemColumns...).
		From("items").
		Where(sq.Eq{
			"clarified_into_task_id":    nil,
			"clarified_into_project_id": nil,
			"discarded":                 0,
		}).
		OrderBy("created_at ASC", "id ASC").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("items: %w", err)
	}
	defer rows.Close()

	var items []gtd.Item
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, fmt.Errorf("items: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (d *DB) GetItem(ctx context.Context, id int64) (gtd.Item, error) {
	query, args, err := sq.Select(itemColumns...).
		From("items").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return gtd.Item{}, err
	}
	item, err := scanItem(d.db.QueryRowContext(ctx, query, args...))
	if errors.Is(err, sql.ErrNoRows) {
		return gtd.Item{}, fmt.Errorf("item %d: not found", id)
	}
	if err != nil {
		return gtd.Item{}, fmt.Errorf("item %d: %w", id, err)
	}
	return item, nil
}

// UpdateItemClarifiedIntoTask stamps the item's ClarifiedIntoTaskID and returns
// the resulting row. The caller is responsible for transaction wrapping and
// pre-checks; the underlying CHECK constraint enforces mutual exclusion as a
// safety net.
func (d *DB) UpdateItemClarifiedIntoTask(ctx context.Context, id, taskID int64) (gtd.Item, error) {
	return d.stampItem(ctx, id, "clarified_into_task_id", taskID)
}

// UpdateItemClarifiedIntoProject stamps the item's ClarifiedIntoProjectID
// (used by both Incubate and ClarifyAsProject) and returns the resulting row.
func (d *DB) UpdateItemClarifiedIntoProject(ctx context.Context, id, projectID int64) (gtd.Item, error) {
	return d.stampItem(ctx, id, "clarified_into_project_id", projectID)
}

// UpdateItemDiscarded marks an item as discarded and returns the resulting row.
func (d *DB) UpdateItemDiscarded(ctx context.Context, id int64) (gtd.Item, error) {
	return d.stampItem(ctx, id, "discarded", 1)
}

func (d *DB) stampItem(ctx context.Context, id int64, column string, value any) (gtd.Item, error) {
	now := time.Now().UTC()
	query, args, err := sq.Update("items").
		Set(column, value).
		Set("updated_at", now).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING " + strings.Join(itemColumns, ", ")).
		ToSql()
	if err != nil {
		return gtd.Item{}, err
	}
	item, err := scanItem(d.db.QueryRowContext(ctx, query, args...))
	if errors.Is(err, sql.ErrNoRows) {
		return gtd.Item{}, fmt.Errorf("item %d: not found", id)
	}
	if err != nil {
		return gtd.Item{}, fmt.Errorf("update item %d: %w", id, err)
	}
	return item, nil
}

func scanItem(s scanner) (gtd.Item, error) {
	var item gtd.Item
	var taskID, projectID sql.NullInt64
	var discarded int
	err := s.Scan(
		&item.ID, &item.Title, &item.Description, &item.CreatedAt, &item.UpdatedAt,
		&taskID, &projectID, &discarded,
	)
	if err != nil {
		return gtd.Item{}, err
	}
	if taskID.Valid {
		item.ClarifiedIntoTaskID = &taskID.Int64
	}
	if projectID.Valid {
		item.ClarifiedIntoProjectID = &projectID.Int64
	}
	item.Discarded = discarded != 0
	return item, nil
}