# Inbox: capture a new item, then clarify an item into a task.
# Regenerate with: ./demo/generate.sh  (or `vhs demo/inbox.tape` with GTD_DB set).
Output demo/inbox.gif
Require gtd

Set Shell bash
Set FontSize 18
Set Width 1200
Set Height 720
Set Padding 24
Set Theme "Catppuccin Mocha"

# Launch the app off-camera so the recording opens on the Inbox tab.
Hide
Type "gtd" Enter
Sleep 1.5s
Show

Sleep 1s

# Capture a new item (title + description), save with ctrl+s.
Type "+"
Sleep 600ms
Type "Schedule dentist for the kids"
Sleep 600ms
Tab
Type "Both due for a 6-month cleaning"
Sleep 800ms
Ctrl+S
Sleep 1.2s

# Clarify the top item: walk the decision tree, create a next action.
Enter
Sleep 1.5s
Down
Sleep 700ms
Down
Sleep 700ms
Down
Sleep 700ms
Ctrl+S
Sleep 1.5s

# Quit.
Ctrl+C
