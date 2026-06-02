## Context

Six overlays (`itemcapture`, `taskedit`, `projectedit`, `taskstatus`, `projectstatus`, `projectpicker`) and one shared widget (`tui/components/date`) currently lean on `charm.land/huh/v2` for field rendering, focus management, and form-state orchestration. Each overlay carries the same boilerplate: build `[]huh.Field`, wrap in `NewGroup`, wrap in `NewForm`, install a custom `KeyMap`, then translate `Form.State == StateCompleted` into a save and `StateAborted` into a dismiss.

`huh.Form`'s message-pump model is what makes our submit-from-anywhere slow (see superseded `submit-form-shortcut`). The same model is also why every overlay has a "save error standoff" branch that pokes `form.State = StateNormal` after a failure — we're constantly working around an opaque state machine.

We already use `bubbles/v2`. `textinput` and `textarea` are mature, fast, and have no opinions about form-level orchestration. We're not replacing huh's *primitives* (those are good); we're replacing its *orchestration* (which is bad for our use case).

## Goals / Non-Goals

**Goals:**
- Remove `charm.land/huh/v2` from `go.mod`.
- Synchronous, sub-millisecond `Submit` from any field (target: no perceptible latency, including on `taskedit`'s 5 fields).
- One small `formfield` package owns all field-and-form mechanics. Overlays become readable: build fields, wire keymap, render — no `Form.State` casts, no `f, ok := form.(*huh.Form)`.
- Preserve overall behavior and traversal: field order, validation rules, esc/tab/ctrl+s semantics, error-on-failure feedback. Exact rendering (label styling, focus indicator, error placement, spacing) is not required to match huh's output — we can pick what looks right.

**Non-Goals:**
- Redesigning any overlay's UX. This is a refactor, not a rework.
- Building a general-purpose form library. The toolkit covers exactly what our overlays need; nothing more.
- Animations, theming engines, or other huh features we don't use.
- Migrating tests away from the existing real-stack/screentest patterns. They keep working; they're the safety net.

## Decisions

### Decision 1: New package `tui/components/form` with per-field subpackages

Naming follows bubbles conventions:

- Top level: `tui/components/form` exporting `form.Model`, `form.New(fields ...Field) *Model`, `form.Field` interface, `form.KeyMap`, `form.SubmittedMsg`.
- One subpackage per concrete field type: `inputfield`, `textfield`, `selectfield`, `datefield` (replacing today's `tui/components/date`), `savefield`. Each subpackage's primary constructor is `New(...)` returning a `*Model` that implements `form.Field`.
- Inside callers that import `tui/components/form`, the form value is conventionally named `f` (since `form` shadows the package); inside the `form` package itself the receiver is `m` (matches `Model`).

```go
// package form

// Values is an immutable snapshot of every field's current Value() at the
// moment the form rendered / navigated / validated. Visibility predicates
// receive a Values and decide whether their field should participate.
type Values interface {
    Get(key string) any
}

type Field interface {
    Key() string                                // unique within a form
    Focus() (Field, tea.Cmd)
    Blur() Field
    Focused() bool
    Visible(Values) bool                        // pure: depends only on the snapshot
    Update(tea.Msg) (Field, tea.Cmd)
    View() string                               // label + cached-error styling
    Value() any                                 // concrete fields choose the type
    Validate() (Field, error)                   // runs validator, caches result in returned field
    SetWidth(int) Field                         // notify field of available width
    ShortHelp() []key.Binding
    FullHelp() [][]key.Binding
}

type Model struct {
    fields []Field
    focus  int
    KeyMap KeyMap                               // Next, Prev, Save
}

// New returns a Model by value. It panics if any two fields share a Key —
// uniqueness is a construction-time invariant, not a runtime check.
func New(fields ...Field) Model

func (m Model) Init() tea.Cmd
func (m Model) Update(tea.Msg) (Model, tea.Cmd)
func (m Model) View() string
func (m Model) Focused() Field
func (m Model) Next() (Model, tea.Cmd)          // validate current, skip hidden, focus next visible
func (m Model) Prev() (Model, tea.Cmd)
func (m Model) Submit() (Model, ok bool, cmd tea.Cmd) // validate visible fields in order;
                                                       // on first failure, focus that field, ok=false
func (m Model) ShortHelp() []key.Binding        // form keys + focused field's keys
func (m Model) FullHelp() [][]key.Binding

// emitted by Update on a successful Submit so the parent screen can perform its
// save side effect without inspecting Model internals
type SubmittedMsg struct{}
```

Everything is value semantics end-to-end. The form holds `[]Field` by value; every mutating operation returns a new `Model`. Visibility predicates receive a fresh `Values` snapshot per call and never hold a reference to any field instance — the huh `.Value(&m.field)` aliasing trap is structurally impossible.

Inner bubbles widgets (`textinput.Model`, `textarea.Model`, `list.Model`) are still allowed to expose pointer-receiver methods for housekeeping. That's fine: a field's `Update` owns its receiver as a local value and may call `(&m.textinput).SetCursor(0)` etc. inside that scope before returning the new field value. The constraint is that those pointer methods only fire from within `Field.Update` — never stashed in a closure or shared.

`Submit` is the linchpin: a synchronous for-loop over fields calling `Validate()`. No message dispatch, no async cmds, no blink-timer overhead. On first failure, focus jumps to the offending field; on success, the surrounding overlay does whatever it does today on `StateCompleted`.

Both `form.Model` and `form.Field` implement `bubbles/v2/help.Model`. Each screen forwards help wiring through the form with `return f.ShortHelp()` — the form composes its KeyMap with the focused field's bindings, so the footer stays accurate as focus moves without per-screen `append([]key.Binding{...}, ...)` boilerplate.

**Concrete field subpackages**:

- `tui/components/form/inputfield` — wraps `bubbles/v2/textinput`. `inputfield.New(label string, opts ...Option) *Model`. Optional `WithValidator(func(string) error)`.
- `tui/components/form/textfield` — wraps `bubbles/v2/textarea`. `textfield.New(label string, opts ...Option) *Model`. `Alt+Enter`/`Ctrl+J` newline; everything else passes through.
- `tui/components/form/selectfield` — generic `selectfield.New[T](label string, options []Option[T], opts ...Option) *Model[T]` backed by `bubbles/v2/list.Model`. Inherits `/`-to-filter from the list component as-is.
- `tui/components/form/radiofield` — generic `radiofield.New[T](label string, options []Option[T], opts ...Option) *Model[T]`. Equivalent to `huh.NewSelect().Inline(true)` — inline rendering for binary or small-N choices. The selected option highlights in a bright accent color when the field is focused; `left`/`right` arrow keys move between choices. No filtering, no scroll. Used for yes/no, task-vs-project, self-vs-delegated, and similar binary decisions in the clarify wizard.
- `tui/components/form/datefield` — replaces today's `tui/components/date`. Same `bubbles/textinput` + `naturaltime` core; reimplements `form.Field` instead of `huh.Field`. On `Blur()` snaps the displayed text to the parsed absolute (e.g. "tomorrow at 3pm" → `2026-06-02 15:00`) so the user sees what's actually saved.
- `tui/components/form/savefield` — terminal `[ Save ]` button. `savefield.New() *Model` (default label "Save"; `WithLabel("Update")` etc. for variants). Validates always-passes; on `Enter` while focused it emits the same submit signal the form's `Save` key emits, routing through the normal `form.Submit()` path. Rendering is a focusable button styled with the form's focus color.

### Decision 2: Keymap lives on the Form, not the fields

`huh` gives each field its own `KeyMap`. We instead put navigation/submit keys on `form.Model` and let fields only consume keys that mean something *to them* (typing, cursor movement, selection toggles). This means:

- `Tab`/`Shift+Tab` always navigates fields, regardless of which field is focused.
- `Enter` no longer overloads meaning by field position. `textfield` ignores plain enter (newline is `alt+enter`); `inputfield` advances to next field; `selectfield` commits the selection (and moves on); `savefield` triggers submit. Each is a deliberate, field-local choice — never position-dependent.
- `Ctrl+S` saves from anywhere, handled by `form.Update`.
- `Esc` cancels — overlays decide what to do (dismiss, clear error, etc.).

### Decision 3: Trailing `savefield` instead of a yes/no confirm

Overlays that need an explicit confirmation step (today: `taskstatus`, `projectstatus`, plus the edit overlays where users might want a deliberate commit moment) end with a `savefield`. `Enter` on the save button submits the form via the normal `form.Submit()` path; `Esc` always cancels. There is no separate "Confirm" yes/no widget — yes is "press Enter on the Save button or anywhere via Ctrl+S," no is "press Esc."

The non-obvious benefit: a trailing button gives natural-date fields (the `datefield` Due / Defer Until / status When) a chance to *render* their parsed absolute form before commit. When the user types `tomorrow at 3pm` and Tabs away to land on `[ Save ]`, the date field blurs, snaps its displayed text to `2026-06-02 15:00`, and the user gets a visible confirmation of what's about to be saved. Without the trailing button, the only way to verify the parse would be to tab past it and back.

This subsumes the original Decision 3 (`NewText` newline handling) — `textfield` exposes the newline binding via its own `ShortHelp()`, which the form picks up automatically.

### Decision 4: Dynamic field visibility via `Field.Visible(Values) bool`

The clarify-wizard work in `submit-form-shortcut` (and the now-living `tui/pages/inbox/clarify`) required a lot of custom code to walk through a progressive question tree, because `huh.Form` has no concept of conditional fields. We bake the missing concept into `form.Field`: every field implements `Visible(Values) bool`, where `Values` is an immutable snapshot of every field's current `Value()`. Default implementation returns `true`. Concrete fields take an `Option` (`WithVisible(func(form.Values) bool)`) to wire conditional behavior:

```go
form.New(
    radiofield.New("kind", "Kind", []Option[string]{{"task", "Task"}, {"project", "Project"}}),
    inputfield.New("assignee", "Assignee", inputfield.WithVisible(func(v form.Values) bool {
        return v.Get("kind") == "task"
    })),
)
```

The form evaluates visibility *progressively* in declaration order: when checking whether field `i` is visible, the snapshot it sees contains only the values of visible fields with index `j < i`. A field cannot see its own value, and a hidden field's value never appears in any later field's snapshot. This forward-only rule rules out circular dependencies by construction — visibility can only flow downstream — and matches how the clarify wizard's question tree actually works.

The form re-evaluates predicates everywhere a field would otherwise participate:

- **Rendering**: hidden fields contribute nothing to `View()`.
- **Navigation**: `Next()`/`Prev()` skip hidden fields. If focus is on a field that becomes hidden between updates, focus advances to the next visible field on the following `Update`.
- **Submit**: `Submit()` does not call `Validate()` on hidden fields and does not treat their cached errors as failures.
- **Help**: `Form.ShortHelp()` only considers the focused (visible) field's bindings.

The huh aliasing trap (`.Value(&m.field)` pinning a pointer to an obsolete copy) is structurally impossible here: predicates depend on a `Values` interface produced fresh on every check, never on a field instance. Field constructors return `Field` by value, and the form holds them by value.

**Mandatory `Key() string`**: every field carries a unique key, supplied as the required first argument to its `New(...)`. `form.New` panics if it sees a duplicate — this is a construction-time invariant, not a runtime check. Keys are how predicates name the fields they depend on.

**Typing at the boundary**: `Values.Get(key) any` is type-erased for now. Predicates do their own assertions (`v.Get("kind") == "task"`). Typed accessors can be layered on later (`Values.String(key)`, `Values.Bool(key)`) once usage patterns crystallize.

### Decision 4a: Rendering goes through a viewport

`form.Model` embeds a `bubbles/v2/viewport.Model`. The viewport's content is rebuilt on every Update from the joined views of the currently-visible fields, and the form scrolls to keep the focused field's first line in view (`viewport.EnsureVisible`). `Init` returns `tea.RequestWindowSize` so the first render has accurate dimensions. Width changes propagate to every field via a `SetWidth(int) Field` method on the `Field` interface, so wrapped/clamped widgets (textinput, textarea, list, datefield) lay out against the real frame.

Why baked-in rather than left to the overlay: every overlay that hosts a form-with-five-fields-or-more would otherwise need the same scrolling code. Putting it in `form.Model` makes overlays stop caring about content height.

**Rejected alternatives:**
- *`Visible() bool` paramless (with closure-captured pointer to other fields).* This is the huh trap — closures pin a specific instance, which goes stale the moment the form returns a new `Model` from `Update`.
- *Form-level visibility registry (`form.When(cond, fields...)` wrapper).* Workable, but moves the predicate away from the field it gates. Worth layering on top of `Visible(Values)` later as syntactic sugar (`form.When(form.StringEq("kind", "task"), inputfield.New(...))`), but the underlying mechanism is still per-field `Visible(Values)`.
- *Generic `Values.Get[T](key) T` accessors.* Forces callers through type parameters at construction sites that don't otherwise need them. Skip until there's a concrete pain point.

### Decision 5: Validation is synchronous; validators are pure, Validate caches

The caller-supplied validator function (passed via `WithValidator(func(string) error)`) MUST be a pure function of the field's value — no side effects, deterministic. This matches what overlay validators already are (length checks, presence checks).

`Field.Validate()` invokes the validator and returns a *new* `Field` whose `View()` renders the result. The returned field carries the cached error; subsequent `Validate()` calls without a value change return the same error. This sidesteps the "pure function that caches" contradiction: the validator is pure, but `Field.Validate` is allowed to thread its cache through the returned value.

The cached error clears automatically on the next `Update` that changes the field's value, so the user isn't shown a stale error after they've begun typing the fix. Fields MUST NOT render validation errors live (as a pure function of their current value) — errors appear only after an explicit `Validate` call, driven by `Submit` or by `Next` gating tab navigation. This gives the huh-style deferred-error UX (no noisy red text while the user is still typing the first character).

### Decision 6: Migration order

Overlays migrate one at a time, each in its own commit:

1. `itemcapture` — smallest (2 fields + `savefield`), proves the toolkit on a real overlay.
2. `taskstatus`, `projectstatus` — `datefield` + `savefield`; proves the trailing-save pattern.
3. `projectpicker` — proves `selectfield[T]`.
4. `projectedit` — 4 fields with `datefield` + `savefield`.
5. `taskedit` — 5 fields + `savefield`; the original motivator.
6. **Clarify wizard rewrite** — collapse `tui/pages/inbox/clarify/**` into a single `form.Model` over all possible fields, with `radiofield` for binary choices (`task | project`, `self | delegated`) and `Visible()` predicates wiring the progressive reveal. Proves `radiofield` and the visibility model. Largest payoff — this is the overlay the visibility feature was added for.
7. Delete `tui/components/formx`, `tui/components/date`, and remove the `huh` import everywhere it lingers.
8. Remove `charm.land/huh/v2` from `go.mod` and `go mod tidy`.

Each migration step is independently mergeable. Tests are the contract: existing real-stack overlay tests must continue passing without modification (except for any direct `huh.Form` casts they make in assertions).

### Decision 7: No separate `Error()` accessor on the Field interface

In huh, `field.Error()` is a public accessor on the field. We don't need that on the interface: `View()` is the only consumer of the cached error, and `View()` is implemented on the same type that owns the cache — it reads its own state directly. Dropping `Error()` from the `Field` interface removes a redundant surface and avoids the awkward "pure function that caches" contradiction that an `Error()` next to a pure `Validate()` implied.

The cache itself still exists internally on concrete fields; it's just not exposed as part of the contract. Overlay save-error standoff branches that previously checked `Error() != nil` instead use the `error` returned from `form.Submit` (which is the form's authoritative validation result anyway).

## Risks / Trade-offs

- **Blast radius** → 6 overlays + date widget + new package + dependency drop. Mitigated by per-overlay commits and existing real-stack tests.
- **Bubbles textarea quirks** → less feature-rich than huh's `Text` (no built-in suggestions, no character counter). We don't use those features, but verify during migration.
- **Custom Select / Confirm** → small new code, but exactly the API we want. No huh-specific edge cases (e.g. huh's `Select` filtering keymap is intricate and we don't need it).
- **`datefield` rework** → the existing `tui/components/date` widget already wraps a textinput and parses with `naturaltime`. Move it to `tui/components/form/datefield`, swap the huh interface for `form.Field`, and add the on-blur snap-to-absolute behavior that the trailing `savefield` makes useful.
- **Help-footer fidelity** → `huh.Form.KeyBinds()` returns a curated list per field type. We replace with `Form.ShortHelp()` / `FullHelp()` (which already composes form-level keys with the focused field's), so screens get reactive bindings for free. Verify each overlay's footer still reads sensibly after migration.

## Migration Plan

Per-overlay commits as listed in Decision 6. The `submit-form-shortcut` change is left in-place but marked superseded; its code (`formx` package + intercepts) survives until the cleanup step of this migration deletes it.

Rollback: each per-overlay commit can be reverted independently; the `formfield` package can stay in tree if needed and `huh` re-added.

## Open Questions

- ~~`Confirm` shape~~ — resolved: trailing `savefield` (Decision 3); no yes/no widget.
- ~~`Select` filtering~~ — resolved: backed by `bubbles/v2/list.Model`, filtering inherited as-is.
- ~~Naming~~ — resolved: `tui/components/form` with field subpackages (`inputfield`, `textfield`, `selectfield`, `datefield`, `savefield`).
