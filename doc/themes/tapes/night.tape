# doc/themes/tapes/night.tape — captures doc/themes/night.png
#
# Render this single theme (from the repo root):
#   vhs doc/themes/tapes/night.tape
# Or re-render every theme at once:
#   doc/themes/render-themes.sh

Output doc/themes/tapes/_scratch.gif

Set Shell "bash"
Set FontSize 16
Set Width 1100
Set Height 640
Set Padding 20
Set Theme "Catppuccin Mocha"

# Hidden warm-up: build the demo binary so the launch is instant, and
# launch/quit it once so the alt-screen renderer is primed before the capture
# (the very first TUI launch can otherwise render blank).
Hide
Type "go build -o /tmp/demo ./examples/basic"
Enter
Sleep 6s
Type "clear"
Enter
Sleep 800ms
Type "/tmp/demo explore"
Enter
Sleep 3s
Type "q"
Sleep 1s
Type "clear"
Enter
Sleep 800ms
Show

Type@60ms "/tmp/demo explore --theme night"
Enter
Sleep 3s
Down@250ms 4
Right
Sleep 400ms
Enter
Sleep 500ms
Down@250ms 1
Enter
Sleep 300ms
Type@80ms "9090"
Enter
Sleep 400ms
Tab
Sleep 700ms
Screenshot doc/themes/night.png
Type "q"
Sleep 400ms
