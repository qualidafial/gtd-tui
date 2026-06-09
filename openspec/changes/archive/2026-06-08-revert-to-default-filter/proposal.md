## Why

Once a user types and applies a filter in a list's query bar, the screen's
construction-time default query is lost — esc only reverts to the last *applied*
query, never the default. Returning to the default view (e.g. `status:open
ready:now` on the Tasks tab) means manually deleting the current filter and
retyping the seed by hand. A single-key "revert to default" gesture restores the
default view instantly.

## What Changes

- Add a `\` keybinding to the task list and project list screens that, while the
  query bar is blurred (list focused), resets the filter to that screen's
  construction-time default query, reparses it, and reloads the list.
- The binding mirrors `/` (focus filter): `/` opens the filter, `\` clears it
  back to the default.
- Reset reuses `querybar.SetValue`, which records the new value as the applied
  query — so a subsequent esc reverts to the default too, keeping the
  applied/default relationship consistent.
- The `\` binding is hidden and inert when the current query already equals the
  default (no redundant help entry, no no-op reload).
- No change to the `querybar` component itself: the reset is handled entirely by
  the parent screens while the query bar is blurred, so it does not touch the
  component's focused-key handling.

Defaults per screen (unchanged, just now reachable via `\`):
- Tasks tab: `status:open ready:now`
- Projects tab: `status:open` (existing `defaultProjectQuery` const)
- Project-view task list: `""` (all tasks)

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `task-list-query-ui`: adds a "revert to default filter" binding (`\`) and the
  requirement that the task list retains its seed query as the reset target.
- `project-list-ui`: adds the same "revert to default filter" binding (`\`) on
  the project list, resetting to `defaultProjectQuery`.

## Impact

- `tui/pages/tasks/tasklist/keymap.go`: new `ResetQuery` binding.
- `tui/pages/tasks/tasklist/model.go`: new `defaultQuery` field (the tasklist
  currently discards the seed string, keeping only the parsed filter), `\`
  handler in `Update`, enable/disable in `updateKeybindings`, inclusion in
  `Keys()`.
- `tui/pages/projects/projectlist.go`: new `ResetQuery` binding sourced from the
  existing `defaultProjectQuery` const, handler, and enable/disable logic.
- No call-site changes in `tui/app.go` — both screens already receive their seed
  query through `New`.
- No changes to `tui/components/querybar`.
