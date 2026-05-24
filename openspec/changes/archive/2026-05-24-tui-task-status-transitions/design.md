## Context

The task list (`tui/pages/tasks/tasklist`) wires `+`/`enter`/`delete`/`shift+arrows`/`/` but only `delete` touches status, via the drop-only `taskdelete` confirm overlay. The service interface already exposes `CompleteTask`, `DropTask`, and `ReopenTask`, each tested. This change is pure TUI wiring plus generalizing the one existing overlay. Comment-on-transition is a near-future feature, so the overlay must be the single place a comment field can later be added.

## Goals / Non-Goals

**Goals:**
- Reach all three status transitions from the task list with confirmation.
- `space` toggles (pending→done; done/dropped→pending); `delete` drops.
- One parameterized confirm overlay for all three transitions, replacing `taskdelete`.
- Contextual `space` help label driven by the selected task's status.

**Non-Goals:**
- No comment field yet — only the seam for it.
- No domain or service changes.
- No visual treatment for transitioned tasks beyond them dropping out of a non-matching filter.

## Decisions

**Generalize `taskdelete` into a transition-parameterized overlay.** Introduce a `Transition` value (Complete / Drop / Reopen) passed to the overlay constructor. A small table maps each transition to its confirm title, description, affirmative label, and the service method to call. Rationale: three near-identical confirm overlays would triplicate the future comment-field work; the proposal already commits to comments-on-transition, so a single overlay is the shape the feature is growing toward. This is consolidation, not speculative abstraction — the overlay already exists; we are parameterizing it, not inventing a framework.

- *Alternative considered:* keep `taskdelete` and add `taskcomplete`/`taskreopen`. Rejected — more code and three places to thread comments through later.
- The package may be renamed to reflect the broader role (e.g. `taskstatus`/`tasktransition`); the implementer chooses. The overlay's external contract is `New(task, svc, transition)`.

**`space` resolves to a transition at keypress time from the selected task's status.** pending→Complete; done→Reopen; dropped→Reopen. `delete`→Drop for pending/done; no-op for dropped. The tasklist model inspects `selectedItem.task.Status` to pick the transition and open the overlay. Rationale: the binding is a toggle, so the action is inherently status-dependent; computing it at the call site keeps the overlay dumb (it just performs the transition it was handed).

**Contextual help label.** The `space` binding's help text is computed when the keymap is built from the currently-selected task's status: `complete` if pending, else `reopen`. The tasklist already constructs its `KeyMap` per render (`KeyMap()` in model.go), so the selected status can be threaded in. The static `KeyEdit`/`KeyDelete` bindings are unaffected.

**Reuse the existing reload path.** On confirm, the overlay fires the service method then emits `tasks.TasksChanged()`, exactly as `taskdelete` does today. The tasklist's existing `TasksChangedMsg` handler reloads via the active filter, so a transitioned task that no longer matches simply drops out — satisfying the "disappear" requirement with no new code.

## Risks / Trade-offs

- **`space` is a common list key.** The bubbles list default delegate may bind `space` (e.g. for selection/pagination). → Confirm the list doesn't consume `space`; the tasklist already intercepts keys before delegating to `m.list.Update`, so handling `space` in the model's keypress switch (before the fall-through) avoids the conflict. Verify in manual testing.
- **No-op `delete` on a dropped task could feel broken.** → It is intentionally inert and not advertised in the help bar for dropped tasks; acceptable since dropped is terminal-ish and `space` reopens.
- **Help label depends on selection.** With no task selected (empty list) the `space` label is undefined. → Fall back to a neutral label (e.g. `toggle`) or omit the binding when no item is selected.
