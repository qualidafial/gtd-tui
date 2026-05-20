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
	dataDir, err := dataDir()
	if err != nil {
		return fmt.Errorf("data dir: %w", err)
	}

	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	db, err := sqlite.Open(filepath.Join(dataDir, "gtd.db"))
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	// projects := service.NewProjectService(db)
	tasks := service.NewTaskService(db)
	// projectTasks := service.NewProjectTaskService(db)

	m := tui.New(
		// projects,
		tasks,
		// projectTasks,
	)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("run: %w", err)
	}
	return nil
}

func dataDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "gtd"), nil
}
