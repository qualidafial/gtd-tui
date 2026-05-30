## Context

Charm's `help.Model` renders the contents of a `help.KeyMap` (an interface with `ShortHelp() []key.Binding` and `FullHelp() [][]key.Binding`). Until this change, each `screen.Screen` had a `KeyMap() help.KeyMap` method that returned a constructed value, and each page's private `keyMap` struct carried two extra slots — `nav list.KeyMap` and `editing help.KeyMap` — that the page mutated at render time before returning `KeyMap()`. The result:

- `Model.KeyMap()` was a fresh constructed value on every render, often wrapping the child screen's `KeyMap()` return in a per-screen adapter (overlay had a private `overlayKeyMap`, huh-form screens had private `keyMap`/`formKeyMap`/`emptyKeyMap`).
- Binding fields were private (`m.keys`), so tests reached in via re-declared types and screens couldn't share or compose bindings.
- The `nav`/`editing` slots inside a page's keymap were a layering smell — the page knew the keys belonged to the list/querybar but had to plumb them in itself.

## Goals / Non-Goals

**Goals:**
- One way to expose a screen's help: implement `ShortHelp`/`FullHelp` directly on the model.
- One place per page to define keys: a per-package `keymap.go` with an exported `KeyMap` struct that itself implements `help.KeyMap`.
- Expose `Model.KeyMap` so tests, parents, and overlays can reach the actual binding objects without reflection or re-declaration.
- Compose help across screens via `slices.Concat` on the screen's own `ShortHelp/FullHelp` output, not via a separate `KeyMap()` adapter.

**Non-Goals:**
- Changing which keys do what (other than the `Toggle → ToggleComplete` / `Enter → View` renames).
- Touching the storage / service / domain layers.
- Reorganizing the help bar's text content for the user.

## Decisions

### Embed `help.KeyMap` in `Screen`

`Screen` becomes:
```go
type Screen interface {
    Init() tea.Cmd
    Update(msg tea.Msg) (Screen, tea.Cmd)
    View() string
    help.KeyMap
}
```

This means every screen *is* a `help.KeyMap`. The root `app.Model` can pass `m.active` (a `Screen`) directly to `help.Model.View`. Composition becomes a two-line method on each parent screen:

```go
func (m Model) ShortHelp() []key.Binding {
    return slices.Concat(m.KeyMap.ShortHelp(), m.child.ShortHelp())
}
```

Alternatives considered:
- Keep `KeyMap() help.KeyMap` and just export the inner field. Rejected: the indirection is what made overlay/huh-form screens need adapter types.
- A `Helper` interface returning `*help.Model`. Rejected: that conflates rendering with binding declaration.

### Per-package `KeyMap` struct

Every component or page that owns bindings gets a `keymap.go` defining an exported `KeyMap` with `DefaultKeyMap()` factory and `ShortHelp`/`FullHelp` methods. The Model stores it as `KeyMap KeyMap` (exported field, not a constructor call per render).

Naming convention: `DefaultKeyMap()` (preferred) or `defaultKeyMap()` for packages where the symbol shouldn't be exported (mostly internal — tasklist/projects, where the constructor is only called from `New`). Both forms appear in the staged diff; we accept the inconsistency rather than renaming all callers.

### Drop `nav` and `editing` slots from page keymaps

Page keymaps no longer embed `list.KeyMap` or a swapped-in `help.KeyMap`. The Model handles composition:

```go
// in tasklist Model
func (m Model) ShortHelp() []key.Binding {
    if m.query.CapturingInput() {
        return m.query.ShortHelp()
    }
    return slices.Concat(
        m.KeyMap.ShortHelp(),
        []key.Binding{m.list.KeyMap.CursorUp, m.list.KeyMap.CursorDown},
    )
}
```

This puts the conditional ("query editing replaces page help") where the state lives, instead of pre-stashing it on the keymap struct.

### Rename `Toggle → ToggleComplete`, `projects.Enter → projects.View`

The dynamic `SetHelp` already swaps the visible label between "complete" and "reopen" based on the selected task's status. Naming the binding `ToggleComplete` (not `Toggle`) makes the primary action obvious from the field name; the label is still updated at runtime. `projects.Enter` was confusing because "enter" was both the key and an action verb; `View` describes what the binding does ("view project") and the key remains `enter`.

### Configurable overlay esc binding

`screen.overlay` used to reference a package-level `keyEsc` var. With a per-overlay `KeyMap`, the binding is now an instance field. No caller exercises that flexibility today, but it removes a hidden global and matches the per-page pattern.

### Huh-form screens forward `KeyBinds()`

`projectedit`, `projectpicker`, `projectstatus`, `taskedit`, `taskstatus` are huh forms. Huh exposes the form's active bindings via `form.KeyBinds() []key.Binding`. The model's `ShortHelp` returns that slice; `FullHelp` returns it wrapped in one group. The previous adapter types (`keyMap`, `formKeyMap`, `emptyKeyMap`) existed only because the old `Screen.KeyMap()` shape needed an object; removing the indirection deletes them.

## Risks / Trade-offs

- [Touching every TUI page in one change makes the diff large] → Mitigated by it being a mechanical refactor with no behavior change and tests proving the help bar still composes correctly.
- [Inconsistent `DefaultKeyMap` vs `defaultKeyMap` casing] → Accepted; will be sanded down opportunistically. Not worth the rename churn now.
- [Public `KeyMap` field is mutable] → Acceptable. Bindings already mutate at runtime via `SetEnabled`/`SetHelp`; nothing changes by making the holder exported. Tests benefit.
- [Renamed binding fields (`Toggle`, `Enter`) break any documentation referencing them] → Specs are updated as part of this change; no external consumers.

## Migration Plan

API-only, internal. No data migration. All in-tree callers are updated in the same commit (already staged).