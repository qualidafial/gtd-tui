## Why

The GTD clarify wizard can turn an inbox item into a standalone next-action, but it cannot attach that action to a project the user already has. This is one of the most common clarify outcomes — "this is the next step for a project that already exists." The `inbox-page` spec already mandates a project picker in the single-task branch (per-task block step 5, "Existing-project attach uses the project picker"), but the wizard implementation never built it.

The service layer is already ready: `InboxService.ClarifyAsTask` accepts and validates `task.ProjectID` (scenario "ClarifyAsTask with project"). So this change is wizard-layer only — surface a project selector in the single-task branch and pass the chosen ID through to the existing operation.

While here, fix a documentation drift: the `project-picker-overlay` spec still describes the picker as a `huh.Select`, but `huh` was removed and replaced by the in-house `selectfield` form component.

## What Changes

- Thread `ProjectService` into the clarify wizard (`tui.New` already holds it → `inbox.New` → `clarify.New`).
- The clarify wizard loads open projects on `Init` before presenting the single-task block (a brief loading phase, mirroring the `projectpicker` overlay's `ready` pattern).
- The single-task branch gains a **Project** `selectfield` that lists open projects with a `(none)` option, defaulting to `(none)` = standalone. Selecting a project sets `task.ProjectID`; the wizard then calls the unchanged `ClarifyAsTask`.
- The picker shows for every single task, including the sub-2-minute do-it-now path.
- The single-task path remains single: after attaching, the wizard commits one task and dismisses (or runs the do-it-now prompt) — it does NOT enter the project per-task loop.
- **Spec refinement:** replace the single-task branch's "Belongs to an existing project? (Yes/No → picker)" gate with a single optional select defaulting to `(none)`.
- **Drift fix:** update `project-picker-overlay` to describe the `selectfield`-based picker instead of `huh.Select`.

No domain, SQLite, or `InboxService` changes.

## Capabilities

### Modified Capabilities

- `inbox-page`: the single-task per-task block exposes an optional open-project select (default `(none)`); the wizard loads open projects before the block renders. Replaces the prior Yes/No-gated picker description.
- `project-picker-overlay`: re-describe the picker in terms of the in-house `selectfield` component (with `(none)` / `WithInitialValue` / submit-on-enter) instead of the removed `huh.Select`. Behavior is unchanged.

## Impact

- **TUI only**:
  - `tui/app.go`: pass `projectSvc` to `inbox.New`.
  - `tui/pages/inbox/model.go`: add `projectSvc` field; thread to `clarify.New`.
  - `tui/pages/inbox/clarify/model.go`: add `projectSvc` field and a `ready`/loading phase; embed a `selectfield` "Project" in the single-task block of the initial form; `taskFromVals` reads the project value (defaulting to standalone when the key is absent, e.g. in the loop form).
- **No service / domain / SQLite changes** — `ClarifyAsTask` already validates `ProjectID`.
- **Dependencies**: reuses the existing `tui/components/form/selectfield` component (same one the `projectpicker` overlay uses).
