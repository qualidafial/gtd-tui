## Context

The GTD TUI application has a defined domain model and service layer (foundation specs), but no user interface. Users need a way to interact with the system to perform GTD workflows: capture thoughts to inbox, clarify items, organize tasks and projects, engage with work, and reflect on activity.

The architecture specifies Bubbletea v2 for the TUI, with pages under `tui/pages/`. This change scaffolds the fundamental structure without implementing full functionality for each page.

## Goals / Non-Goals

**Goals:**
- Establish the Bubbletea v2 application skeleton with proper model structure
- Create a working entry point in cmd/gtd/
- Implement page-based navigation with global shortcuts
- Set up placeholder pages for all major GTD workflows
- Define conventions for consistent page structure and styling
- Enable dependency injection for services

**Non-Goals:**
- Full implementation of any page's CRUD operations
- Actual clarify flow UI
- Data display or editing forms
- Persistence or database integration
- Styling beyond basic structure

## Decisions

### Decision: Single root model with embedded page models

The root model in `tui/app.go` will embed all page models and manage routing between them. Pages are not dynamically loaded; they exist for the application lifetime.

**Rationale:** Bubbletea works best with a single model hierarchy. Embedding pages allows the root to dispatch messages and render the active page cleanly. Since all pages are known at compile time and there's no memory concern, preloading them simplifies the design.

**Alternative considered:** Lazy page instantiation. Rejected because the overhead of keeping simple page structs in memory is negligible, and lazy loading adds complexity without benefit.

### Decision: Page interface extending tea.Model

Pages implement a `Page` interface that extends `tea.Model`:

```go
type Page interface {
    tea.Model
    Focus() tea.Cmd    // Called when page becomes active
    Blur()             // Called when page becomes inactive  
    ShortHelp() []key.Binding
    FullHelp() [][]key.Binding
}
```

**Rationale:** Focus/Blur hooks allow pages to refresh data or clean up when activated/deactivated. Help bindings integrate with bubbletea's help component for consistent shortcut display.

**Alternative considered:** Using tea.Model directly without extension. Rejected because Focus/Blur semantics are essential for data refresh, and help bindings are needed for discoverability.

### Decision: Page subpackages under tui/pages/

Each page lives in its own subpackage:
- `tui/pages/inbox/`
- `tui/pages/tasks/`
- `tui/pages/projects/`
- `tui/pages/someday/`
- `tui/pages/references/`
- `tui/pages/meetings/`

**Rationale:** Follows the architecture spec. Subpackages provide isolation, clear boundaries, and room for page-specific components to grow without polluting a single package.

**Alternative considered:** All pages in a single `tui/pages/` package. Rejected because as pages grow with list/detail views, the single package would become unwieldy.

### Decision: Navigation via message passing

Page navigation uses Bubbletea messages. A `NavigateMsg` struct carries the target page identifier. The root model handles this in Update, switching the active page.

```go
type NavigateMsg struct {
    Page PageID
}
```

**Rationale:** Message passing is the Bubbletea idiom. Pages don't need to know about each other; they emit navigation messages and the root routes.

**Alternative considered:** Direct function calls between pages. Rejected as it violates Bubbletea's architecture and creates tight coupling.

### Decision: Navigation history stack in root model

The root model maintains a `[]PageID` history stack. Back navigation pops the stack. Direct navigation pushes to the stack.

**Rationale:** GTD workflows involve drilling into details and returning. A simple stack handles this naturally. The "one pane of glass" principle means users frequently navigate between linked entities.

**Alternative considered:** No history (always return to default page). Rejected because the UX of losing context when backing out is poor.

### Decision: Global key bindings handled in root model

Global navigation keys (go to inbox, tasks, projects, etc.) are handled in the root model's Update method before dispatching to the active page. If a global key is pressed, the root handles it; otherwise, the message is forwarded to the active page.

**Rationale:** Ensures global keys work consistently regardless of page state. Pages don't need to implement global navigation.

**Alternative considered:** Let pages handle all keys and bubble up navigation. Rejected because it duplicates logic and risks inconsistency.

### Decision: Default page determined at startup

On startup, the root model queries the inbox service. If `InboxService.List()` returns items, show inbox; otherwise show tasks.

**Rationale:** Directly implements the GTD principle that inbox is the entry point when items exist (per gtd-workflows spec).

### Decision: Services passed to constructor, stored in root

The root model's constructor accepts all service interfaces:

```go
func NewApp(inbox InboxService, tasks TaskService, projects ProjectService, ...) *App
```

Services are stored in the root and passed to pages as needed.

**Rationale:** Dependency injection enables testing with mocks. Centralizing services in root avoids passing them through multiple layers.

## Risks / Trade-offs

**Risk:** Page models grow large as functionality is added.
**Mitigation:** This scaffold establishes subpackage isolation. Each page can internally decompose into components as needed.

**Risk:** Navigation history stack could grow unbounded in long sessions.
**Mitigation:** Cap stack size (e.g., 50 entries) and drop oldest when exceeded. Unlikely to be an issue in practice.

**Risk:** Global keys could conflict with page-local keys.
**Mitigation:** Use a consistent modifier pattern (e.g., unmodified letters for global nav, or reserve specific keys). Document conventions clearly.

**Risk:** Placeholder pages provide no useful functionality.
**Mitigation:** Expected for a scaffold. Each page displays its name and help text, confirming the navigation works. Real functionality comes in subsequent changes.
