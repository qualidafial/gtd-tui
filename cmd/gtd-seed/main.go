// Command gtd-seed populates a fresh demo database with representative inbox
// items, tasks, and projects. It exists to give the VHS demo recordings
// (see demo/) realistic, deterministic content. The target path comes from
// GTD_DB or the first positional argument; an existing file at that path is
// removed first so each run starts clean.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	gtd "github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/service"
	"github.com/qualidafial/gtd-tui/sqlite"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	path := os.Getenv("GTD_DB")
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	if path == "" {
		return fmt.Errorf("usage: gtd-seed <db-path> (or set GTD_DB)")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}
	// Start clean so the seed is deterministic across runs.
	for _, suffix := range []string{"", "-wal", "-shm"} {
		if err := os.Remove(path + suffix); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove %s: %w", path+suffix, err)
		}
	}

	db, err := sqlite.Open(path)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	ctx := context.Background()
	tasks := service.NewTaskService(db)
	projects := service.NewProjectService(db)
	inbox := service.NewInboxService(db)

	day := 24 * time.Hour

	// Unprocessed inbox captures — the clarify queue.
	for _, it := range []gtd.Item{
		{Title: "Email from Dana re: Q3 offsite venue"},
		{Title: "Idea: weekly review template in the wiki"},
		{Title: "Replace the smoke detector battery"},
		{Title: "Look into standing desk options"},
	} {
		if _, err := inbox.Create(ctx, it); err != nil {
			return fmt.Errorf("create item %q: %w", it.Title, err)
		}
	}

	// A project with linked next actions, in display order.
	shed, err := projects.CreateProject(ctx, gtd.Project{
		Title:   "Build a backyard shed",
		Outcome: "A weatherproof 8x10 shed with shelving installed",
	})
	if err != nil {
		return fmt.Errorf("create project shed: %w", err)
	}
	for _, t := range []gtd.Task{
		{Title: "Finalize shed dimensions and site", Status: gtd.TaskStatusOpen, ProjectID: &shed.ID},
		{Title: "Get two lumber quotes", Status: gtd.TaskStatusOpen, ProjectID: &shed.ID, Due: at(2 * day)},
		{Title: "Pull permit from the county", Status: gtd.TaskStatusOpen, ProjectID: &shed.ID, DeferUntil: at(3 * day)},
	} {
		if _, err := tasks.CreateTask(ctx, t); err != nil {
			return fmt.Errorf("create shed task %q: %w", t.Title, err)
		}
	}

	// A parked (someday) project to show the status filter.
	if _, err := projects.CreateProject(ctx, gtd.Project{
		Title:   "Learn to sail",
		Outcome: "Comfortable single-handing a small dinghy",
		Status:  gtd.ProjectStatusSomeday,
	}); err != nil {
		return fmt.Errorf("create project sail: %w", err)
	}

	// Standalone next actions, including a delegated and an overdue one.
	for _, t := range []gtd.Task{
		{Title: "Renew passport", Status: gtd.TaskStatusOpen, Due: at(-2 * day)},
		{Title: "Draft the sprint retro notes", Status: gtd.TaskStatusOpen, Assignee: new("Priya")},
		{Title: "Book dentist cleaning", Status: gtd.TaskStatusOpen},
		{Title: "Review the budget spreadsheet", Status: gtd.TaskStatusOpen, Due: at(5 * day)},
	} {
		if _, err := tasks.CreateTask(ctx, t); err != nil {
			return fmt.Errorf("create task %q: %w", t.Title, err)
		}
	}

	// A completed task, transitioned through the real status API so its
	// StatusChangedAt reflects an actual completion.
	plants, err := tasks.CreateTask(ctx, gtd.Task{Title: "Water the plants", Status: gtd.TaskStatusOpen})
	if err != nil {
		return fmt.Errorf("create task plants: %w", err)
	}
	if _, err := tasks.CompleteTask(ctx, plants.ID, time.Now()); err != nil {
		return fmt.Errorf("complete task plants: %w", err)
	}

	fmt.Fprintf(os.Stderr, "seeded demo database at %s\n", path)
	return nil
}

func at(d time.Duration) *time.Time {
	return new(time.Now().Add(d))
}
