## Context

Each list screen seeds its query bar with a construction-time default query
(`status:open ready:now` for the Tasks tab, `status:open` for the Projects tab,
`""` for the project-view task list). The `querybar` component tracks only two
strings: the live input value and `appliedQuery` (the last committed query). Esc
reverts the live value to `appliedQuery`. The default is consumed once via
`SetValue` at construction and then discarded — there is no way to return to it
without retyping by hand.

The task list (`tasklist.Model`) keeps only the parsed `filter`, not the seed
string; the project list already has a `defaultProjectQuery` const. The reset
must fire while the query bar is blurred, so it is driven by the parent screen's
keymap, not by the querybar's focused-key handling.

## Goals / Non-Goals

**Goals:**
- One keystroke (`\`) restores a screen's default filter while browsing the list.
- The binding is hidden and inert when the current query already equals the
  default.
- After a reset, Esc behaves consistently (reverts to the now-applied default).
- No changes to the shared `querybar` component.

**Non-Goals:**
- A reset gesture while the query bar is focused/editing (Esc already covers
  "discard live edit"; `\` while editing is a literal character).
- A configurable or per-user default query.
- Any change to the default query values themselves.

## Decisions

**Reset is parent-driven, not a querybar feature.**
The reset key is only meaningful while the query bar is blurred. The querybar's
`Update` ignores keys unless focused, so a querybar-owned binding could never
fire in the intended context. The parent screen owns a `ResetQuery` binding in
its `KeyMap`, handles it in `Update`, and advertises it via `Keys()` — exactly
like the existing `FocusQuery` (`/`) binding. Alternative considered: add a
`Reset()` method and `defaultQuery` storage to the querybar. Rejected: it splits
ownership of a binding the querybar can't actually consume and adds surface area
for zero reuse benefit, since each parent must reparse and reload anyway.

**Reuse `querybar.SetValue` for the reset.**
`SetValue(s)` already sets both the input text and `appliedQuery = s`. So the
handler is: `m.query.SetValue(defaultQuery)`, reparse the default into the typed
filter, and issue the existing reload command. This automatically yields the
"Esc reverts to default" behavior because `appliedQuery` now holds the default —
no new querybar state required.

**tasklist retains the seed string in a new `defaultQuery` field.**
`tasklist.Model` currently discards the seed query after parsing it in `New`.
Add a `defaultQuery string` field set from the existing `query` parameter. No
call-site change in `app.go` — the seed is already passed. The project list reads
the existing `defaultProjectQuery` const directly.

**Enable/disable via the existing per-state reconciliation.**
The `\` binding is toggled with `SetEnabled(m.query.Value() != m.defaultQuery)`
in the same routine that reconciles the other selection-dependent bindings
(`updateKeybindings` on the task list; the equivalent on the project list).
Disabled bindings are both hidden from help and rejected by `key.Matches`, so a
single flag governs visibility and inertness — matching the codebase's existing
"one source of truth for is-this-action-available" pattern.

## Risks / Trade-offs

- **`\` already bound elsewhere** → Confirmed free in both list keymaps and the
  global/keymap-resolution layer. No conflict.
- **Comparison uses raw strings, not normalized filters** → `m.query.Value()`
  vs. `defaultQuery` is a literal string compare. After a reset, the value is set
  to the exact default string, so equality holds; after a manual edit back to the
  default text it also holds. Whitespace-only differences would read as "not at
  default" and merely re-enable a no-op `\`, which is harmless. Acceptable.
- **Empty-default screens** (project-view task list) → `\` resets to `""` (all
  tasks). Correct for an empty default; the binding is simply disabled whenever
  the live query is already empty.
