## Why

The keymap toolkit named its wrapper type `Chord` and its `Map` interface method `Chords()`. "Chord" implies a multi-key combination, but these values wrap a single `key.Binding`; the name misleads. The method name also shadowed the meaning of the embedded `key.Binding.Keys()`, so call sites read `m.KeyMap.Chords()` next to `c.Keys()` for related concepts. Renaming the type to `Binding` and the method to `Keys()` aligns the toolkit with the underlying `bubbles/v2/key` vocabulary.

This is a pure rename — no routing, resolution, or help behavior changes.

## What Changes

- Rename the `keymap.Chord` type to `keymap.Binding` (its `Show`/`Vis` fields and live-`key.Binding` embedding are unchanged).
- Rename the `keymap.Map` interface method `Chords() []Group` to `Keys() []Group`, and update `Group = []Chord` to `Group = []Binding`.
- Propagate the method rename to every layer that implements `keymap.Map` (app, overlay, tabcontainer, querybar, form, every concrete field, and the form-based and KeyMap-owning pages).
- Move the `Resolve` call inside `keymap.ShortHelp`/`FullHelp` so callers pass the aggregated `Keys()` directly rather than pre-resolving.
- Update the four impacted spec documents to the new terminology.
- Replace the word "chord" in prose — spec text and code comments — with "binding", and rename the test-local `chord()` builder / `chords` fixture field to match, so no "chord" terminology survives.

## Capabilities

### Modified Capabilities

- `keymap-resolution`: `Chord` type → `Binding`; `Map.Chords()` → `Map.Keys()`; `Group = []Binding`.
- `form-field-toolkit`: `Field`/`Form` expose `Keys() []keymap.Group` instead of `Chords()`.
- `tui-application`: app and per-component KeyMaps expose `Keys()` instead of `Chords()`.
- `tui-view-stack`: the `Screen` contract and overlay aggregation use `Keys()` and `keymap.Binding`.

## Impact

- `tui/internal/keymap`: `keymap.go` (type/interface rename), `resolve.go` (`Binding` literals, `ShortHelp`/`FullHelp` now call `Resolve` internally), tests.
- `tui/app.go`, `tui/keymap.go`, `tui/components/{screen,tabcontainer,querybar,form,...}`, and all `keymap.go` / form-based pages: `Chords()` → `Keys()`.
- Test fixtures: `keymap` package `chord()` builder → `binding()`; `form` package `stubField.chords` → `bindings`.
- No behavior change; `go vet ./...` and `go test ./...` pass.
