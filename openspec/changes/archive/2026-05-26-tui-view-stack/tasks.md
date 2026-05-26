## 1. Screen Package — Messages and Interfaces

- [x] 1.1 Add PushMsg struct (carries a Screen) and Push(child Screen) tea.Cmd to screen package
- [x] 1.2 Add DismissMsg struct and Dismiss() tea.Cmd to screen package
- [x] 1.3 Add InitMsg struct and InitCmd() tea.Cmd to screen package
- [x] 1.4 Add Popper interface: Pop() Screen
- [x] 1.5 Remove ShowOverlayMsg, HideOverlayMsg, ShowOverlay(), HideOverlay() from screen package

## 2. Overlay Wrapper

- [x] 2.1 Create tui/components/screen/overlay.go with unexported overlay struct (inner Screen, parent Screen)
- [x] 2.2 Implement Overlay(parent, child Screen) Screen constructor
- [x] 2.3 Implement overlay.Pop() Screen (satisfies Popper)
- [x] 2.4 Implement overlay.Update: forward messages to inner; send Dismiss() on esc when inner is not capturing input
- [x] 2.5 Implement overlay.Init: delegate to inner.Init()
- [x] 2.6 Implement overlay.View: delegate to inner.View()
- [x] 2.7 Implement overlay.KeyMap: merge inner KeyMap with esc binding via overlayKeyMap
- [x] 2.8 Implement overlay.CapturingInput: delegate to inner via screen.CapturingInput()

## 3. Tab Container

- [x] 3.1 Create tui/components/tabcontainer/ package with Model struct (tabs []Screen, labels []string, activeTab int, width, height)
- [x] 3.2 Implement New() constructor accepting screens and labels
- [x] 3.3 Implement Init: call Init() on active tab
- [x] 3.4 Implement Update: handle tab/shift+tab (scoped keybindings), forward other messages to active tab
- [x] 3.5 Implement tab switching: call Init() on newly active tab when switching
- [x] 3.6 Implement View: render tab bar (logo, labels, active indicator) + active tab View()
- [x] 3.7 Implement KeyMap: merge active tab KeyMap with tab/shift+tab bindings
- [x] 3.8 Implement CapturingInput: delegate to active tab via screen.CapturingInput()
- [x] 3.9 Handle tea.WindowSizeMsg: subtract tab bar height, forward adjusted dimensions to active tab

## 4. Refactor app.Model

- [x] 4.1 Replace tabs/activeTab/overlay fields with single `active Screen` field
- [x] 4.2 Initialize active as tabcontainer.New() in tui.New(), passing screens and labels
- [x] 4.3 Move tab bar rendering (logo, tab labels, styling) from app.go to tabcontainer
- [x] 4.4 Remove tab/shift+tab keybinding handling from app.Update (now in tabContainer)
- [x] 4.5 Remove ShowOverlayMsg/HideOverlayMsg case branches from app.Update
- [x] 4.6 Remove TasksChangedMsg/TasksLoadedMsg broadcast loop from app.Update
- [x] 4.7 Add PushMsg handler: m.active = screen.Overlay(m.active, msg.Screen); return m, m.active.Init()
- [x] 4.8 Add DismissMsg handler: pop via Popper interface; return m, m.active.Init()
- [x] 4.9 Add InitMsg handler: return m, m.active.Init()
- [x] 4.10 Update ? keybinding: suppress when CapturingInput(m.active) returns true
- [x] 4.11 Update WindowSizeMsg: forward to m.active instead of iterating tabs
- [x] 4.12 Update View: render m.active.View() + help footer (no overlay branch)
- [x] 4.13 Update help rendering: use m.active.KeyMap() instead of mergedKeyMap with overlay/appKeys booleans
- [x] 4.14 Remove InputCapturer check for tab suppression (no longer needed — tabs are in tabContainer)

## 5. Migrate tasklist

- [x] 5.1 Replace screen.ShowOverlay(taskedit.New(...)) with screen.Push(taskedit.New(...))
- [x] 5.2 Replace screen.ShowOverlay(taskstatus.New(...)) with screen.Push(taskstatus.New(...))
- [x] 5.3 Remove TasksChangedMsg handler from tasklist.Update (parent reloads on init)
- [x] 5.4 Remove TasksLoadedMsg filter-matching logic (filterMatches func) — each tab reloads independently
- [x] 5.5 Remove tasksReorderedMsg cross-tab filtering (each tab is independent)

## 6. Migrate taskedit

- [x] 6.1 Replace screen.HideOverlay() with screen.Dismiss() on successful save (taskSavedMsg with no error)
- [x] 6.2 Replace screen.HideOverlay() with screen.Dismiss() on form abort (StateAborted)
- [x] 6.3 Remove tasks.TasksChanged() from save/abort command batches (parent reloads on init)
- [x] 6.4 Implement CapturingInput() — return true when form is in StateNormal (editing)
- [x] 6.5 Remove import of tui/pages/tasks package (TasksChanged no longer needed)

## 7. Migrate taskstatus

- [x] 7.1 Replace screen.HideOverlay() with screen.Dismiss() on successful transition (taskTransitionedMsg with no error)
- [x] 7.2 Replace screen.HideOverlay() with screen.Dismiss() on form abort (StateAborted) and cancel (confirm=false)
- [x] 7.3 Remove tasks.TasksChanged() from transition/abort command batches (parent reloads on init)
- [x] 7.4 Implement CapturingInput() — return true when form is in StateNormal (editing)
- [x] 7.5 Remove import of tui/pages/tasks package (TasksChanged no longer needed)

## 8. Cleanup

- [x] 8.1 Remove TasksChangedMsg and TasksChanged() from tui/pages/tasks/tasks.go (or remove file if empty)
- [x] 8.2 Remove mergedKeyMap type from app.go (replaced by active.KeyMap() + app globals)
- [x] 8.3 Update cmd/gtd/main.go if tui.New() signature changed
- [x] 8.4 Run go build ./... to verify compilation
- [x] 8.5 Run go test ./... to verify all tests pass
- [x] 8.6 Manual TUI testing: verify tab switching, task edit push/dismiss, task status push/dismiss, ? toggle, esc at root, ctrl+c quit, query bar ? and esc handling
