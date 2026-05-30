## 1. Screen interface

- [x] 1.1 Embed `help.KeyMap` in `screen.Screen` (replace `KeyMap() help.KeyMap` method)
- [x] 1.2 Add `tui/components/screen/keymap.go` defining `screen.KeyMap{Back}` with `ShortHelp/FullHelp` and `DefaultKeyMap()`
- [x] 1.3 Refactor `screen.overlay` to carry a `KeyMap` field, use `o.KeyMap.Back` for esc, and expose `ShortHelp/FullHelp` directly via `slices.Concat`

## 2. App + cross-cutting components

- [x] 2.1 Add `tui/keymap.go` defining `tui.KeyMap{Quit, Help}` with `ShortHelp/FullHelp` and `DefaultKeyMap()`
- [x] 2.2 Update `tui/app.go` to hold `KeyMap tui.KeyMap`, route Quit/Help via `m.KeyMap`, implement `ShortHelp/FullHelp` by concatenating with the active screen, and disable `Help` while the active screen captures input
- [x] 2.3 Update `tui/app_test.go` `stubScreen` to satisfy the new interface (drop `KeyMap()`; add `ShortHelp/FullHelp`)
- [x] 2.4 Rename `tui/components/tabcontainer/tabcontainer.go` → `model.go`; add `tui/components/tabcontainer/keymap.go` with `KeyMap{Next, Prev}`; expose `Model.KeyMap`
- [x] 2.5 Move `KeyMap` declaration out of `tui/components/querybar/querybar.go` into `tui/components/querybar/keymap.go`; add `Model.ShortHelp/FullHelp` forwarding to the KeyMap

## 3. Pages — list-style

- [x] 3.1 `tui/pages/tasks/tasklist`: rename `keyMap → KeyMap`, drop `nav` / `editing` fields, rename `Toggle → ToggleComplete`, expose `Model.KeyMap`, implement `Model.ShortHelp/FullHelp` composing the keymap with `m.list.KeyMap` nav and `m.query` editing
- [x] 3.2 `tui/pages/tasks/tasklist/render.go`: update delegate to use exported `KeyMap`
- [x] 3.3 `tui/pages/tasks/tasklist/model_test.go`: update bindings access (`m.KeyMap.MoveUp`) and `New` signature
- [x] 3.4 `tui/pages/projects`: same shape — rename `keyMap → KeyMap`, `Toggle → ToggleComplete`, `Enter → View`, drop `nav/editing` slots, expose `Model.KeyMap`, implement `Model.ShortHelp/FullHelp`
- [x] 3.5 `tui/pages/projects/render.go` + `model_test.go`: update for new keymap shape

## 4. Pages — overlay / detail

- [x] 4.1 Add `tui/pages/projects/projectview/keymap.go` defining `KeyMap{Edit}`; update `projectview/model.go` to use `m.KeyMap.Edit` and implement `ShortHelp/FullHelp` concatenating with the inner task list

## 5. Pages — huh-form

- [x] 5.1 `tui/pages/tasks/taskedit/model.go`: replace `KeyMap()` method with `ShortHelp/FullHelp` returning `form.KeyBinds()`; delete `tui/pages/tasks/taskedit/keymap.go`
- [x] 5.2 `tui/pages/tasks/taskstatus/model.go`: same; delete `taskstatus/keymap.go`
- [x] 5.3 `tui/pages/projects/projectedit/model.go`: replace `KeyMap()` + inline `keyMap` adapter with direct `ShortHelp/FullHelp`
- [x] 5.4 `tui/pages/projects/projectpicker/model.go`: replace `KeyMap()` + `formKeyMap`/`emptyKeyMap` adapters with direct `ShortHelp/FullHelp`
- [x] 5.5 `tui/pages/projects/projectstatus/model.go`: replace `KeyMap()` + `keyMap` adapter with direct `ShortHelp/FullHelp`

## 6. Verification

- [x] 6.1 `go build ./...` clean
- [x] 6.2 `go test ./...` green
- [x] 6.3 Manual TUI check: help bar still shows app + active-screen bindings, esc still dismisses overlays, `?` still toggles full help and is suppressed in huh forms