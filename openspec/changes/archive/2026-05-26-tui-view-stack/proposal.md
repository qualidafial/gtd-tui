## Why

The current app.Model manages views with a flat tabs array plus a single overlay slot. This design cannot stack overlays (e.g. project list -> project view -> project edit), forces special-case dispatching in app.Update, and couples the overlay lifecycle to app.Model. Upcoming features (project drill-down, task detail views, comments) all need nested navigation. Rather than bolt on ad-hoc nesting for each feature, this change introduces a view stack that handles arbitrary depth cleanly.

## What Changes

### View stack via overlay parent chain

- Replace the `tabs []Screen` + `overlay Screen` fields in app.Model with a single `active Screen` field
- Extract the tab bar and tab-switching logic into a `tabContainer` screen that becomes the home view
- Introduce an `overlay` wrapper that pairs any Screen with a parent reference and an esc-to-pop keybinding
- The "view stack" is implicit in the overlay parent chain, not an explicit data structure managed by app

### Push, pop, and dismiss — all handled by app.Model

- **Push**: a screen sends `screen.Push(child)`, which produces a `PushMsg`. app.Model handles PushMsg by wrapping the current active screen: `screen.Overlay(m.active, msg.Screen)`. The screen that initiated the push does not decide who its parent is — app always uses the current active.
- **Pop via esc**: the overlay wrapper handles esc by sending `screen.Dismiss()`. app.Model handles the resulting DismissMsg.
- **Pop via dismiss**: an inner screen signals "I'm done" by returning `m, screen.Dismiss()`. `Dismiss()` produces a `DismissMsg`. app.Model handles DismissMsg by calling `Pop()` on the active screen if it satisfies the `Popper` interface, returning the parent screen and an `InitCmd`.
- **Popper interface**: `Pop() (Screen, tea.Cmd)` — satisfied by the overlay wrapper and by bespoke screens that hold their own parent reference.

### Init-driven data lifecycle

- `InitCmd` produces an `InitMsg`; app.Update responds by calling `active.Init()` and forwarding the resulting command
- Every screen reloads its own data in `Init()`. This eliminates cross-view coordination messages — when a parent screen regains focus via pop, it reinitializes and fetches fresh state from the database.
- `TasksChangedMsg` is removed. Screens that modify data simply dismiss; the parent (and any ancestor) will reload when re-initialized. Tab switches within `tabContainer` also trigger `Init()` on the newly active tab, so cross-tab staleness resolves naturally.
- The cost is slightly more DB queries on view transitions, but these are local SQLite reads and effectively free.

### Keybindings and help

- app.Model retains the core keybindings that must always work: ctrl+c (quit), ? (toggle help)
- app.Model always renders the help bar using the active view's `KeyMap()`
- The overlay wrapper injects the esc binding into the inner screen's `KeyMap()`
- Tab/shift+tab keybindings move into `tabContainer` — they are only active when `tabContainer` is the topmost view, naturally suppressed when any overlay is pushed

## Out of Scope

- **Project TUI pages.** This change refactors navigation infrastructure only. Project list, project view, and project edit are separate changes that build on this.
- **Query bar / filter changes.** No changes to task query or list filtering.
- **New screens.** No new views are added; existing views (task list, task edit, task status) are migrated to the new pattern.

## Capabilities

### New Capabilities
- `tui-view-stack`: Overlay-based view stack enabling arbitrary nesting of screens with push/dismiss navigation, managed centrally by app.Model

### Modified Capabilities
- `tui-application`: app.Model manages a single active Screen instead of tabs + overlay; handles PushMsg, DismissMsg, and InitMsg centrally; removes cross-view coordination messages in favor of init-driven reloading
- `tui-navigation`: Navigation history is implicit in the overlay parent chain rather than an explicit stack data structure; tab switching moves into tabContainer; tabs re-init on activation

## Impact

- `tui/app.go`: Replace tabs/overlay/activeTab with single `active Screen`; remove tab-switching keybindings (move to tabContainer); remove ShowOverlayMsg/HideOverlayMsg handling; add PushMsg (wrap active in overlay), DismissMsg (pop via Popper interface), InitMsg (call active.Init()) handling; remove TasksChangedMsg broadcast loop
- `tui/components/screen/screen.go`: Remove ShowOverlayMsg/HideOverlayMsg/ShowOverlay/HideOverlay; add PushMsg/Push, DismissMsg/Dismiss, InitMsg/InitCmd; add Popper interface; retain InputCapturer (still needed for app-level ? help toggle suppression during text input)
- New `tui/components/screen/overlay.go`: overlay wrapper with parent reference, esc sends Dismiss(), KeyMap merging with esc binding; satisfies Popper via `Pop() (Screen, tea.Cmd)`; purely structural — forwards all other messages to inner screen
- New `tui/components/tabcontainer/`: extracted tab bar, tab switching (tab/shift+tab keybindings), tab labels, styling; implements Screen; calls Init() on tab activation
- `tui/pages/tasks/tasks.go`: Remove TasksChangedMsg type and TasksChanged() constructor
- `tui/pages/tasks/tasklist/model.go`: Remove TasksChangedMsg/TasksLoadedMsg cross-tab handling; push taskedit/taskstatus via screen.Push(child); remove filterMatches logic (each tab reloads independently on Init)
- `tui/pages/tasks/taskedit/model.go`: Replace HideOverlay() with Dismiss() on save and form abort; implement InputCapturer (form focus state); overlay wrapper handles esc suppression and KeyMap merging automatically
- `tui/pages/tasks/taskstatus/model.go`: Replace HideOverlay() with Dismiss() on confirm and form abort; implement InputCapturer (form focus state); overlay wrapper handles esc suppression and KeyMap merging automatically
