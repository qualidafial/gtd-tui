## Context

The TUI currently manages views through app.Model, which holds a `tabs []screen.Screen` array, an `activeTab` index, and a single `overlay screen.Screen` slot. Screens push overlays via `ShowOverlayMsg` and pop via `HideOverlayMsg`, both handled as special cases in app.Update. Cross-view data coordination uses `TasksChangedMsg` broadcast to all tabs.

This approach cannot stack overlays, forces app.Model to mediate every view transition, and requires explicit coordination messages to keep views in sync after mutations.

The existing Screen interface:
```go
type Screen interface {
    Init() tea.Cmd
    Update(msg tea.Msg) (Screen, tea.Cmd)
    View() string
    KeyMap() help.KeyMap
}
```

Files involved: `tui/app.go` (root model), `tui/components/screen/screen.go` (Screen interface + overlay messages), `tui/pages/tasks/tasklist/model.go` (list with query bar), `tui/pages/tasks/taskedit/model.go` (huh form overlay), `tui/pages/tasks/taskstatus/model.go` (confirm dialog overlay), `tui/pages/tasks/tasks.go` (TasksChangedMsg).

## Goals / Non-Goals

**Goals:**
- Enable arbitrary overlay stacking (e.g. project list -> project view -> project edit)
- Centralize all view transition logic in app.Model (push, pop, init)
- Replace cross-view coordination messages with init-driven reloading
- Establish a reusable pattern for all future push/pop navigation

**Non-Goals:**
- Adding new screens (projects, inbox, etc.) — this is infrastructure only
- Changing the query bar, list rendering, or task filtering behavior
- Animated transitions between views

## Decisions

### app.Model as the central view manager

All view transitions — push, pop, init — are handled by app.Model. Screens never create overlays directly. They send commands (`screen.Push`, `screen.Dismiss`), and app handles the structural decisions.

```go
// app.Update:
case screen.PushMsg:
    m.active = screen.Overlay(m.active, msg.Screen)
    return m, m.active.Init()

case screen.DismissMsg:
    if popper, ok := m.active.(screen.Popper); ok {
        m.active = popper.Pop()
        return m, m.active.Init()
    }
    return m, nil

case screen.InitMsg:
    return m, m.active.Init()
```

This means a tab inside tabContainer that wants to push an edit screen just sends `screen.Push(child)`. App wraps the current `active` (which is the tabContainer) as the parent. When the overlay pops, it returns to the tabContainer — not to the bare tab. The tab never needs to know who its container is.

**Alternative considered:** Each container (tabContainer, overlay) handles PushMsg from its children. Rejected because it distributes view management across multiple types, and the tab-as-parent problem requires containers to intercept and re-wrap — complexity without benefit.

### Overlay wrapper with parent reference

The overlay is a thin struct in the screen package that wraps any Screen with a parent Screen reference and an esc keybinding. The "view stack" is the parent chain — not an explicit data structure.

```go
// tui/components/screen/overlay.go
type overlay struct {
    inner  Screen
    parent Screen
}

func Overlay(parent, child Screen) Screen {
    return overlay{inner: child, parent: parent}
}

func (o overlay) Pop() Screen {
    return o.parent
}
```

The overlay is purely structural. It forwards all messages to its inner screen, with one exception: when the inner screen is not capturing input and esc is pressed, the overlay sends `screen.Dismiss()`. It does not handle PushMsg or DismissMsg — those propagate back to app via the command channel.

```go
func (o overlay) Update(msg tea.Msg) (Screen, tea.Cmd) {
    if msg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(msg, keyEsc) {
        if !CapturingInput(o.inner) {
            return o, Dismiss()
        }
    }
    inner, cmd := o.inner.Update(msg)
    o.inner = inner
    return o, cmd
}
```

**Alternative considered:** Explicit stack in app.Model (`views []Screen` with push/pop). Rejected — the parent chain achieves the same result without app managing indices, and each overlay naturally captures its own parent.

**Alternative considered:** Replacing the root `tea.Model` entirely on push (each overlay becomes the `tea.Model`). Rejected because it requires every overlay to handle ctrl+c/quit and help rendering independently, duplicating app-level concerns.

### InputCapturer delegation through overlays

The overlay delegates `InputCapturer` to its inner screen. This solves two problems:

1. **esc handling**: When the inner screen is capturing input (e.g. huh form in edit mode), the overlay suppresses its own esc and forwards to the inner screen. The form gets esc, aborts, and the screen sends `Dismiss()`. When the inner screen is not capturing, the overlay handles esc directly.

2. **? handling**: app.Model suppresses `?` (help toggle) when `CapturingInput(m.active)` returns true. The overlay propagates this from its inner screen, so `?` reaches text inputs in huh forms instead of toggling help.

```go
func (o overlay) CapturingInput() bool {
    return CapturingInput(o.inner)
}
```

This means huh form screens (taskedit, taskstatus) work correctly inside overlays without any special handling. They implement InputCapturer (reporting true when the form is focused), and the overlay + app suppress their own keybindings accordingly. No "bespoke screen" concept is needed — all screens use `screen.Push()` and get wrapped in overlays uniformly.

The flow for a huh form screen:
1. Form is active (CapturingInput = true) → overlay suppresses esc, app suppresses ? → form receives all keys
2. User presses esc → form aborts (StateAborted) → screen sends `Dismiss()`
3. User completes form → screen saves, sends `Dismiss()`
4. App receives DismissMsg → pops overlay → inits parent

### Popper interface

app.Model needs to pop the active screen on DismissMsg. The overlay wrapper satisfies `Popper`.

```go
type Popper interface {
    Pop() Screen
}
```

If the active screen doesn't satisfy Popper (e.g. tabContainer at the root), DismissMsg is a no-op.

### Dismiss via command, not nil return

An inner screen signals "I'm done" by returning `m, screen.Dismiss()`. `Dismiss()` produces a `DismissMsg` that app handles by popping.

```go
func Dismiss() tea.Cmd {
    return func() tea.Msg { return DismissMsg{} }
}
```

The inner screen always returns a valid model from Update — no nil returns anywhere in the system.

**Alternative considered:** Inner returns nil model, container detects and pops. Rejected — nil models are a footgun with struct value semantics. Any accidental nil propagation panics on the next View() call.

### Push via command

A screen sends `screen.Push(child)` to show a new view. `Push()` produces a `PushMsg` that app handles by wrapping the current active.

```go
func Push(child Screen) tea.Cmd {
    return func() tea.Msg { return PushMsg{Screen: child} }
}
```

The screen that initiated the push does not decide who the parent is. App always uses `m.active` as the parent. This solves the tab-in-container problem: a tasklist inside tabContainer sends Push, app wraps tabContainer (not the tasklist) as the parent, and popping returns to tabContainer with tabs intact.

### Init-driven data lifecycle

Every screen reloads its own data in `Init()`. App calls `Init()` on the new active after every push or pop.

This eliminates `TasksChangedMsg`. When taskedit saves and dismisses, app pops, the parent re-inits, and reloads from the database. Cross-tab staleness resolves the same way: tabContainer calls `Init()` on the newly active tab when switching.

**Trade-off:** Slightly more DB queries on every view transition. Acceptable — these are local SQLite reads that take <1ms.

### tabContainer as a Screen

Extract tab bar logic from app.Model into `tui/components/tabcontainer/`. The tab container implements Screen, owns tab/shift+tab keybindings and tab label rendering, and manages an array of child screens. Tab switches call `Init()` on the newly active tab.

```go
type Model struct {
    tabs      []screen.Screen
    labels    []string
    activeTab int
    width     int
    height    int
}
```

app.Model's `active` field starts as the tabContainer. Tab-switching keybindings are scoped to the tabContainer — when any overlay is pushed on top, tab switching is naturally suppressed because messages only reach the topmost view.

tabContainer also delegates InputCapturer to its active tab. When the tasklist's query bar is focused, `CapturingInput` propagates up through tabContainer to the overlay (if any) to app, suppressing `?` and (if overlaid) esc at each level.

### app.Model becomes minimal

After extraction, app.Model holds:
- `active Screen` — the topmost view (tabContainer or an overlay)
- `help help.Model` — renders `active.KeyMap()` plus ctrl+c and ?
- `width, height` — cached window dimensions, forwarded to active view

app.Update handles:
- `tea.WindowSizeMsg` — cache dimensions, forward to active
- `tea.KeyPressMsg` for ctrl+c (quit) and ? (toggle help, suppressed when CapturingInput)
- `PushMsg` — wrap active in overlay, init the child
- `DismissMsg` — pop via Popper, init the parent
- `InitMsg` — call `active.Init()`
- All other messages — forward to `active.Update()`

### Overlay KeyMap merging

The overlay wrapper injects an esc binding into the inner screen's KeyMap:

```go
func (o overlay) KeyMap() help.KeyMap {
    return overlayKeyMap{
        inner: o.inner.KeyMap(),
    }
}
```

The overlayKeyMap prepends esc to the inner's ShortHelp/FullHelp. The overlay wrapper handles the esc mechanics and KeyMap merging so the inner screen doesn't need to define its own esc binding.

For huh form screens that already show esc in their own KeyMap (as "back" or "cancel"), the overlay's esc appears as a duplicate. This is harmless — the form's esc label is more specific and appears first. If it becomes confusing, the inner screen can suppress the duplicate by not including esc in its own KeyMap, relying on the overlay's.

### DismissMsg propagation through overlay chain

When an inner screen sends `Dismiss()`, app pops one level. If the parent also needs to dismiss (e.g. save cascading up), the parent would need to send its own `Dismiss()` in its `Init()`. In practice this isn't needed — a save-and-dismiss pops one level, and the parent reloads on init.

## Risks / Trade-offs

**Two-tick dismiss**: `Dismiss()` is a command, so the pop happens on the next message cycle — not synchronously. The inner screen renders one more frame after requesting dismiss. In practice this is imperceptible (<16ms).

**Init() must be idempotent and cheap**: Every push/pop calls Init() on the new top view. Screens must not accumulate state across Init calls (e.g. appending to a list instead of replacing it). The existing tasklist.Init() already returns a load command that replaces items on arrival, so this is safe today.

**Parent staleness during overlay**: While an overlay is active, the parent screen is frozen — it doesn't receive messages or update. This is fine because the parent reloads on pop. But if a background process sends a message meant for a buried view, it's lost. Currently there are no such background processes; if added later, the app could forward specific message types down the chain.

**InputCapturer chain**: InputCapturer must propagate correctly through every layer: overlay delegates to inner, tabContainer delegates to active tab. A missing delegation at any layer causes keybinding conflicts. This is straightforward to implement and test, but must be verified for each new container type.
