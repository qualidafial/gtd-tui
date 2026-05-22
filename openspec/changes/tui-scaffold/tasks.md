## 1. Project Setup

- [ ] 1.1 Add Bubbletea v2 and related Charm dependencies to go.mod
- [ ] 1.2 Create cmd/gtd/ directory structure
- [ ] 1.3 Create tui/ package directory structure
- [ ] 1.4 Create tui/pages/ subdirectories for each page (inbox, tasks, projects, someday, references, meetings)

## 2. Core Application Structure

- [ ] 2.1 Define Page interface in tui/page.go with Focus, Blur, and help binding methods
- [ ] 2.2 Define PageID type and constants for each page (Inbox, Tasks, Projects, Someday, References, Meetings)
- [ ] 2.3 Define NavigateMsg type for page navigation
- [ ] 2.4 Create root App model in tui/app.go with page embedding and routing

## 3. Root Model Implementation

- [ ] 3.1 Implement App constructor accepting service interfaces
- [ ] 3.2 Implement Init method that determines default page based on inbox state
- [ ] 3.3 Implement Update method with global key handling and page dispatch
- [ ] 3.4 Implement View method that renders active page with consistent chrome
- [ ] 3.5 Implement navigation history stack with push/pop for back navigation
- [ ] 3.6 Handle WindowSizeMsg to track and propagate terminal dimensions

## 4. Placeholder Pages

- [ ] 4.1 Implement inbox page placeholder in tui/pages/inbox/ with basic structure
- [ ] 4.2 Implement tasks page placeholder in tui/pages/tasks/ with basic structure
- [ ] 4.3 Implement projects page placeholder in tui/pages/projects/ with basic structure
- [ ] 4.4 Implement someday page placeholder in tui/pages/someday/ with basic structure
- [ ] 4.5 Implement references page placeholder in tui/pages/references/ with basic structure
- [ ] 4.6 Implement meetings page placeholder in tui/pages/meetings/ with basic structure

## 5. Navigation

- [ ] 5.1 Implement global navigation key bindings in root model (i=inbox, t=tasks, p=projects, etc.)
- [ ] 5.2 Implement back navigation with Escape key
- [ ] 5.3 Wire up page Focus/Blur calls on navigation transitions
- [ ] 5.4 Add help key binding to show available shortcuts

## 6. Entry Point

- [ ] 6.1 Create cmd/gtd/main.go with tea.Program initialization
- [ ] 6.2 Wire service stubs or nil services for initial testing
- [ ] 6.3 Handle program cleanup and terminal restore on exit
- [ ] 6.4 Implement quit handling for q and Ctrl+C

## 7. Styling and Chrome

- [ ] 7.1 Define shared styles in tui/styles.go using lipgloss
- [ ] 7.2 Create header component showing current page name
- [ ] 7.3 Create footer component showing available key bindings
- [ ] 7.4 Apply consistent layout structure across all pages

## 8. Verification

- [ ] 8.1 Verify application starts and displays default page
- [ ] 8.2 Verify global navigation keys switch between all pages
- [ ] 8.3 Verify back navigation returns to previous page
- [ ] 8.4 Verify quit keys exit the application cleanly
- [ ] 8.5 Verify terminal resize updates layout
