## 1. keymap package rename

- [x] 1.1 Rename `Chord` struct → `Binding` and `Group = []Chord` → `Group = []Binding` in `keymap.go`
- [x] 1.2 Rename the `Map` interface method `Chords() []Group` → `Keys() []Group`
- [x] 1.3 Update `Binding` literals in `resolve.go`; move the `Resolve` call inside `ShortHelp`/`FullHelp`
- [x] 1.4 Replace "chord" with "binding" in package doc comments (`keymap.go`, `resolve.go`, `render.go`)

## 2. Propagate the method rename

- [x] 2.1 Rename `Chords()` → `Keys()` on every `keymap.Map` implementer: app, overlay, screen KeyMap, tabcontainer, querybar
- [x] 2.2 Rename `Chords()` → `Keys()` on `form.Model`, `form.Field`, and every concrete field
- [x] 2.3 Rename `Chords()` → `Keys()` on the form-based and KeyMap-owning pages
- [x] 2.4 Update call sites to pass `Keys()` directly to `ShortHelp`/`FullHelp`

## 3. Prose and fixtures

- [x] 3.1 Replace "chord" with "binding" in code comments across `tui/`
- [x] 3.2 Rename test-local `chord()` builder → `binding()` and `stubField.chords` → `bindings`
- [x] 3.3 Update the four impacted specs (`keymap-resolution`, `form-field-toolkit`, `tui-application`, `tui-view-stack`) to the new terminology and prose

## 4. Verification

- [x] 4.1 `go vet ./...` and `go test ./...` pass; no "chord" remains in `tui/` or `openspec/specs/`
