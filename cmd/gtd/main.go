package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	"github.com/qualidafial/gtd-tui/service"
	"github.com/qualidafial/gtd-tui/sqlite"
	"github.com/qualidafial/gtd-tui/tui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	dbPath, err := dbPath()
	if err != nil {
		return fmt.Errorf("db path: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	db, err := sqlite.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	projects := service.NewProjectService(db)
	tasks := service.NewTaskService(db)
	inbox := service.NewInboxService(db)

	m := tui.New(
		tasks,
		projects,
		inbox,
	)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("run: %w", err)
	}
	return nil
}

// dbPath returns the SQLite database path. GTD_DB overrides the default
// location, letting demos and tests run against an isolated database without
// clobbering the user's real one.
func dbPath() (string, error) {
	if p := os.Getenv("GTD_DB"); p != "" {
		return p, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "gtd", "gtd.db"), nil
}
