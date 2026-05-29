## Context

The task list has a query bar baked directly into `tasklist/model.go` — it owns a `textinput.Model`, debounce logic, parse-error state, and a 3-line rendering area (prompt + `^^^` underline + error message). The project list has no filtering at all. Error messages are scattered: each screen manages its own `err` field and renders errors in different locations.

The `textinput.Model` from Charm bubbles does not support per-character styling — its `Styles.Text` applies uniformly to all text. Highlighting an offending range inline requires either post-processing the ANSI output or rendering the text manually outside the textinput's `View()`.

## Goals / Non-Goals

**Goals:**
- Reusable query bar component shared between task list and project list.
- Single-line query bar at all times — inline error highlighting via color/underline on the offending range.
- Project query parser supporting `status:` (open/someday/done/dropped) and free-text search on title/outcome.
- `ProjectFilter.Search` field with SQLite `LIKE` matching, following the task search pattern.
- Centralized error display in `tui.Model` — the help bar row shows errors instead of help bindings, with an esc-to-clear hint.
- Screens stop managing their own error display; they return `error` as `tea.Msg` and the app handles presentation.

**Non-Goals:**
- Date predicates for project queries (no `due:`, `defer:`, `ready:`). Projects lack defer/ready and due filtering can come later.
- Assignee filter for projects (projects don't have assignees).
- Changing the help bar component itself — we replace its content when an error is active, not modify the bubbles help model.

## Decisions

### Query bar as `tui/components/querybar/` package
A new `querybar.Model` encapsulates: `textinput.Model`, debounce sequence, parse-error state, focused/blurred state, a `ValidateFunc func(string) *ParseError` callback, and a configurable debounce interval. The component exposes `Focus()`, `Blur()`, `Value()`, `SetValue()`, `View()`, and `CapturingInput()`.

When the user presses enter and validation succeeds, the querybar's `Update` returns an `querybar.ApplyMsg{Query string}` via cmd. The parent handles this message by parsing the query into its typed filter (e.g. `taskquery.Parse()` → `TaskFilter`, `projectquery.Parse()` → `ProjectFilter`) and reloading. The validate function returns nil or a `*ParseError`; the parent re-parses on apply to get the filter value. This avoids generics and keeps the querybar unaware of filter types.

When validation fails on enter or debounce, the querybar stores the `*ParseError` for inline highlighting and returns a cmd yielding it (as an `error`) for the app error bar.

**Alternative**: Generic `querybar.Model[F any]` with a `ParseFunc func(string) (F, error)` — rejected because it adds generics complexity for marginal benefit; parsing is cheap enough to do twice.

**Alternative**: Keep query bar inline in each list — rejected because the task list and project list need identical behavior and the code is non-trivial (debounce, error state, focus management).

### Single-line rendering with inline error highlight
The query bar always renders as one line. When a `ParseError` is present, the component post-processes the `textinput.View()` output using `ansi.Cut` from `github.com/charmbracelet/x/ansi` to slice the rendered text into three segments around the error's `[Start, End)` range (offset by prompt width): before, offending, and after. The offending segment is re-styled with red foreground and underline, then all three are concatenated. This preserves the textinput's own ANSI styling (cursor, prompt) while adding error highlighting. When blurred with no error, the textinput renders normally as a single line. When focused with no error, it delegates to `textinput.View()` unchanged.

**Alternative**: Build the view manually bypassing textinput — rejected because it duplicates cursor handling and prompt styling logic.

**Alternative**: Keep the `^^^` underline on a second row — rejected per user requirement to always be single-line.

### Parse error message goes to app error bar
The query bar component stores the `ParseError` for inline highlighting, but the error *message* is communicated via a `tea.Cmd` returning an `error`. The app's centralized error bar displays it. The query bar does not render the message text itself.

### `projectquery` package at `internal/projectquery/`
Mirrors `internal/taskquery/`: tokenize, split key:value, recognized keys map. Supports `status:` (open/someday/done/dropped) and free-text search tokens. Returns a `gtd.ProjectFilter`. Same `ParseError` type with `Start`/`End` rune offsets for range highlighting.

**Decision**: Define `querybar.ParseError` in the querybar package with `Message`, `Start`, `End` fields. Both `taskquery` and `projectquery` return `*querybar.ParseError`. This gives one type for the component to work with and eliminates the current `taskquery.ParseError` (replaced by the shared type).

### `ProjectFilter.Search []string` and SQLite implementation
Add `Search []string` to `ProjectFilter`. In `sqlite/project.go`, for each search term apply `(lower(title) LIKE ? OR lower(outcome) LIKE ? OR lower(description) LIKE ?)` using `likeContains`. Same pattern as task search.

### Centralized error bar in `tui.Model`
`tui.Model` gains an `err error` field. In `Update`:
- `case error:` → set `m.err = msg`, return nil cmd.
- On any `tea.KeyPressMsg`, if `m.err != nil` and key is esc → clear `m.err`, return nil. If `m.err != nil` and key is not esc → clear `m.err`, then continue processing the key normally (errors are transient, any action dismisses them, but esc is the explicit clear).

Any keypress clears the error bar and passes through to the active screen — the key is never consumed. The error is a transient notification; any user action dismisses it. This matches the projectlist pattern (which currently clears `m.err` on any keypress) and avoids special-casing esc at the app level.

Wait — screens currently handle `case error:` in their own Update. If the app intercepts `error` first, screens never see them. This is correct for ambient errors. For save-error overlays (taskedit, projectedit), the error arrives as a typed message (`taskSavedMsg{err}`, `projectSavedMsg{err}`), not as a bare `error`. Those overlays set their internal error state to block re-fire, then return a `tea.Cmd` that yields an `error` for the app to display. So: overlay receives `savedMsg`, sets blocking state, returns `cmd` that yields `error` → app catches it and displays.

**Change for save-error overlays**: Instead of storing the error and rendering it, they store only a `blocked bool` (or keep the `err` field for blocking purposes but don't render). After receiving a failed save message, they return an error cmd: `func() tea.Msg { return fmt.Errorf("save failed: %w", msg.err) }`. The app displays it. Esc in the overlay clears the block and resumes the form (overlay still handles this internally). Esc at the app level also clears the error bar display.

For this to work cleanly, the overlay's esc handling must come before the app's esc-to-clear-error handling. Since the app dispatches to the active screen first (the overlay), and the overlay consumes esc when it's in error-blocked state, the app never sees that esc. The next esc (when the overlay is back in normal state) either aborts the form or dismisses the overlay — also consumed before the app. So the app's esc-to-clear only fires when no screen consumes it. That's fine — the error bar message persists until replaced or until an esc reaches the app.

Actually, reconsidering: the app should clear `m.err` on esc *before* dispatching to screens, because the error bar is app-level UI. But then the screen also sees esc and might dismiss. The user's intent on esc is: (1) if error showing, clear it; (2) if no error, dismiss/cancel. So: on any keypress, app clears `m.err` if set, then forwards the key to the active screen unconditionally.

### Error bar rendering
In `renderFooter()`: if `m.err != nil`, render the error text in red/bold instead of the help bar.

## Risks / Trade-offs

**[Trade-off]** Moving `ParseError` to the querybar package creates a dependency from `internal/taskquery` and `internal/projectquery` on a TUI component. → Acceptable: the querybar package is lightweight (no bubbletea dependency in the ParseError type itself), and the alternative (a separate `internal/queryerr` package) adds indirection for one struct.

**[Trade-off]** The app intercepts all `error` messages, so screens can no longer handle errors locally. Any screen that needs to react to an error (e.g., save-error blocking) must use a typed message instead. → This is already the case for save-error overlays (they use `taskSavedMsg`, etc.), so no breakage. Screens that had `case error:` for ambient errors just remove that case.
