## Context

The task edit overlay (`taskedit`) establishes the pattern for form-based editing in the TUI: a huh form overlay pushed via the view stack, with a read-only metadata header for existing entities, save-error surfacing, and esc-to-retry. Projects currently have a minimal `projectcreate` overlay that only accepts a title via a raw `textinput`. There is no way to edit projects after creation.

The `Project` domain type has four user-editable fields: Title, Outcome, Description, and Due. `ProjectService.CreateProject` and `UpdateProject` already exist and handle persistence.

## Goals / Non-Goals

**Goals:**
- Unified project editor overlay for both create and edit, mirroring the taskedit pattern.
- Outcome is required (GTD: defining the desired outcome is essential to project clarification).
- Replace `projectcreate` with the new editor — `+` in project list opens the editor with an empty project.
- Entry points from both project list (`e` key) and project view (`e` key).
- Project view reloads its header after the editor dismisses.

**Non-Goals:**
- Status editing from the editor — status transitions remain in the dedicated confirmation overlays.
- Inline editing of project fields in the project view header.
- Description field in the project view header (already excluded by spec).

## Decisions

### Package location: `tui/pages/projects/projectedit/`
Follows the existing convention (`projectcreate/`, `projectstatus/`, `projectview/`). Single `model.go` file.

**Alternative**: Extending `projectcreate` — rejected because the quick-create pattern (raw textinput, no form) is too different from a full huh form. Cleaner to replace than adapt.

### Huh form with four fields: Title, Outcome, Description, Due
- Title: `huh.NewInput()` with non-empty validation.
- Outcome: `huh.NewInput()` with non-empty validation.
- Description: `huh.NewText()` (multi-line, optional).
- Due: `date.NewField()` (shared date field, optional).

This matches the taskedit field pattern. No Assignee or DeferUntil since those are task-specific.

### Read-only header for existing projects
Shows Project ID, Status (with relative time from StatusChangedAt), Created, Updated. Same styling as taskedit (`metaLabelStyle`/`metaValueStyle`). Hidden for new projects (ID == 0).

### Create vs update: branch on ID == 0
Same pattern as taskedit: `ID == 0` means create, otherwise update. The constructor takes a `gtd.Project` value — callers pass a zero-value project for create.

### Delete `projectcreate` package
After `projectedit` handles both flows, `projectcreate` is dead code. Remove the package and update imports.

### Create flow: dismiss then push project view
After a successful create, the editor uses `tea.Sequence` to dismiss itself, then push the project view for the newly created project. This lets the user immediately start adding tasks — a natural GTD workflow (clarify project → define next actions). The sequence ensures dismiss completes before push. For testability, the sequence cmd can be cast to `[]tea.Cmd` to inspect individual commands.

The editor needs a way to construct the project view. It receives a `ViewFactory` function from the caller (project list passes a closure that builds a `projectview.Model`). On update, the factory is unnecessary and not called — only dismiss is sent.

### Project view reload on dismiss
When the editor dismisses from project view, the view needs to reload the project to reflect title/outcome/due changes in the header. The project view will re-fetch the project on `screen.ReloadMsg` (sent after the editor dismisses), matching how task list reloads work.

## Risks / Trade-offs

**[Risk]** Project view holds a snapshot of the project struct; after editing, the header is stale until reload. → Mitigation: dismiss triggers reload via the existing screen reload mechanism. The staleness window is sub-frame.

**[Trade-off]** Outcome is required for new projects but existing projects in the DB may have empty outcomes. → The editor will enforce validation on save, so editing an old project with an empty outcome will require filling it in. This is intentional — it nudges users to define outcomes per GTD principles.