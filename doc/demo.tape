# doc/demo.tape — VHS script for cobra-explorer's README demo.
#
# Prerequisites:
#   brew install vhs   (or: go install github.com/charmbracelet/vhs@latest)
#   Go toolchain on PATH
#
# Render (from the repo root):
#   vhs doc/demo.tape        # writes doc/demo.gif
#
# Self-contained: the demo CLI is launched with `go run`, so no manual
# build step and no environment-specific binary path is required.
#
# Key bindings used below (see internal/model + internal/tree):
#   j/k or Down/Up   move within a zone
#   l/Right, h/Left  expand / collapse a tree node
#   Enter            select a runnable command / confirm a flag edit
#   Tab / Shift+Tab  cycle focus between Tree, Description, Flags, Command zones
#   Space            toggle a bool flag
#   c                copy the built command (Command zone)
#   q                quit

Output doc/demo.gif

Set Shell "bash"
Set FontSize 16
Set Width 1100
Set Height 600
Set Padding 20
Set Theme "Catppuccin Mocha"

# Hidden warm-up: populate Go's build cache so the visible launch is instant.
Hide
Type "go build -o /dev/null ./examples/basic"
Enter
Sleep 5s
Type "clear"
Enter
Sleep 500ms
Show

# Launch the interactive explorer (works anywhere the Go toolchain is installed).
Type@150ms "go run ./examples/basic explore"
Enter
Sleep 1.5s

# Browse the command tree: move down to "run" and expand it.
Down@600ms 4
Sleep 800ms
Right
Sleep 1s

# Step into the subcommands and select "run server".
Enter
Sleep 1.2s

# Focus is now on the Flags zone. Edit the --port flag.
Down@600ms 1
Enter
Sleep 600ms
Type@150ms "9090"
Enter
Sleep 1.5s

# Move to the --tls bool flag and toggle it on.
Down@500ms 1
Space
Sleep 1s

# Jump to the Command zone and copy the built command.
Tab
Sleep 800ms
Type@150ms "c"
Sleep 2s

# Quit.
Type "q"
Sleep 500ms

