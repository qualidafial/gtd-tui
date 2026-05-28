## Why

Projects can be created with a title-only quick-create overlay, but there is no way to edit a project's attributes after creation. Users cannot fix typos or fill in outcome/description/due without direct DB access. The quick-create overlay is also too limited — new projects should get the same full-field form as editing, unifying creation and editing into one component (mirroring how task edit handles both create and update).

## What Changes

- Add a `projectedit` overlay package (`tui/pages/projects/projectedit/`) that presents a huh form for project fields: Title (required), Outcome (required), Description, and Due.
- Display a read-only header above the form (for existing projects) showing Project ID, Status, Created, and Updated. No header for new projects.
- Replace the existing `projectcreate` quick-create overlay — the `+` key in project list now opens the same editor with an empty project.
- Wire the edit overlay into `projectview` (`e` key) and `projectlist` (`e` key on selected project).
- Support save-error surfacing with esc-to-retry, matching the task edit pattern.
- Delete the `projectcreate` package.

## Capabilities

### New Capabilities
- `project-edit-ui`: Defines the unified project editor overlay — form fields, read-only header, save/cancel behavior, error handling, and create-vs-update semantics.

### Modified Capabilities
- `project-view-screen`: Add an `e` key binding to open the project edit overlay from the project view, and reload the project header on dismiss.
- `project-list-ui`: Add an `e` key binding to open the project edit overlay for the selected project. Change `+`/`insert` to open the project editor (replacing `projectcreate`).

## Impact

- New TUI package: `tui/pages/projects/projectedit/`
- Deleted: `tui/pages/projects/projectcreate/` (replaced by projectedit)
- Modified: `tui/pages/projects/projectview/model.go` (edit key binding, reload on dismiss)
- Modified: `tui/pages/projects/projectlist.go` (edit key binding, replace create reference)
- Modified: `tui/pages/projects/keymap.go` (new edit key)
- Uses existing `ProjectService.CreateProject` and `UpdateProject` — no domain/service/sqlite changes needed.
