## 1. Task list (tasklist)

- [x] 1.1 Add a `defaultQuery string` field to `tasklist.Model` and set it from the `query` parameter in `New`
- [x] 1.2 Add a `ResetQuery` binding (`\`, help `"revert filter"`) to `tasklist.KeyMap` in `DefaultKeyMap`
- [x] 1.3 Add `ResetQuery` to `tasklist.KeyMap.Keys()`, grouped next to `FocusQuery`
- [x] 1.4 Handle `ResetQuery` in `Model.Update` (only when the query bar is not capturing input): `SetValue(defaultQuery)`, reparse into `m.filter`, return `loadCmd()`
- [x] 1.5 In `updateKeybindings`, set `ResetQuery` enabled to `m.query.Value() != m.defaultQuery`

## 2. Project list (projectlist)

- [x] 2.1 Add a `ResetQuery` binding (`\`, help `"revert filter"`) to the projects `KeyMap`
- [x] 2.2 Advertise `ResetQuery` in the projects `Keys()` alongside the focus-query binding
- [x] 2.3 Handle `ResetQuery` in `Update` (only when not capturing input): `SetValue(defaultProjectQuery)`, reparse into the filter, reload the list
- [x] 2.4 Enable/disable `ResetQuery` based on `query.Value() != defaultProjectQuery` in the keybinding-reconciliation routine

## 3. Tests

- [x] 3.1 tasklist: pressing `\` after applying a different query reverts text to the default and reloads with the default filter
- [x] 3.2 tasklist: after `\`, focusing/editing then Esc reverts to the default query
- [x] 3.3 tasklist: `\` binding is disabled when the current query equals the default
- [x] 3.4 tasklist: `\` is inert while the query bar is focused (enters the literal character)
- [x] 3.5 tasklist: `\` on an empty-default instance clears to all tasks
- [x] 3.6 projectlist: pressing `\` reverts to `status:open` and reloads
- [x] 3.7 projectlist: `\` binding disabled when query equals `status:open`

## 4. Verification

- [x] 4.1 Run the full test suite (`go test ./...`)
- [x] 4.2 Confirm `\` appears in the help bar only when not at the default, on both the Tasks and Projects tabs
- [x] 4.3 Update README/keybinding docs if they enumerate list shortcuts
