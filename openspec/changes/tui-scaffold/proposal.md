## Why

The domain model and service layer are defined, but users have no way to interact with the GTD system. A terminal UI is needed to enable the core GTD workflows: capture, clarify, organize, engage, and reflect. The Bubbletea v2 framework provides the foundation for a responsive, keyboard-driven TUI that embodies the "low ceremony, one pane of glass" principles.

## What Changes

- Add `cmd/gtd/` entry point that initializes the TUI application
- Add `tui/` package with Bubbletea v2 application scaffold
- Add `tui/pages/` structure for page-based navigation
- Implement root model that manages page routing and global state
- Add placeholder pages for each major GTD workflow (inbox, tasks, projects, etc.)
- Wire up keyboard navigation between pages
- Establish TUI conventions for consistent look and feel

## Capabilities

### New Capabilities

- `tui-application`: Root TUI application structure using Bubbletea v2, including the main model, initialization, and page routing
- `tui-navigation`: Keyboard-driven navigation between pages and views, supporting the "one pane of glass" principle
- `tui-pages`: Page structure for GTD workflow views (inbox, tasks, projects, someday, references, meetings)

### Modified Capabilities

## Impact

- **New packages**: `cmd/gtd/`, `tui/`, `tui/pages/`
- **Dependencies**: `github.com/charmbracelet/bubbletea/v2` and related Charm libraries
- **Entry point**: Creates the main executable that users will run
- **Foundation specs**: References architecture and gtd-workflows specs from the foundation change
