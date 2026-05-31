package gtd

import (
	"context"
	"time"
)

// Item is an unprocessed inbox capture. It is the entry point for the GTD
// clarify workflow: every Item is eventually clarified into a Task or Project,
// or discarded. At most one of ClarifiedIntoTaskID / ClarifiedIntoProjectID is
// non-nil, and Discarded is false whenever either is set. The Incubate and
// ClarifyAsProject operations both target ClarifiedIntoProjectID; the project's
// status distinguishes the two outcomes.
type Item struct {
	ID                     int64
	Title                  string
	Description            string
	CreatedAt              time.Time
	UpdatedAt              time.Time
	ClarifiedIntoTaskID    *int64
	ClarifiedIntoProjectID *int64
	Discarded              bool
}

// InboxService is the full inbox surface: capture (Create/List/Get) plus the
// four clarify operations. The concrete implementation lives in the service
// package; the SQLite layer satisfies the capture-side methods and the
// service-layer wraps them with the cross-store transactional clarify ops.
type InboxService interface {
	Create(ctx context.Context, item Item) (Item, error)
	List(ctx context.Context) ([]Item, error)
	Get(ctx context.Context, id int64) (Item, error)

	Discard(ctx context.Context, itemID int64) (Item, error)
	Incubate(ctx context.Context, itemID int64, project Project) (Project, Item, error)
	ClarifyAsTask(ctx context.Context, itemID int64, task Task) (Task, Item, error)
	ClarifyAsProject(ctx context.Context, itemID int64, project Project, firstTask Task) (Project, Task, Item, error)
}