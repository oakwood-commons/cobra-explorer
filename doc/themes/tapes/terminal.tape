# doc/themes/tapes/terminal.tape — captures doc/themes/terminal.png
#
# The "terminal" theme is transparent and has no background of its own, so it
# is captured against a representative dark VHS background so its contrast
# reads correctly. On your machine it takes on whatever background (and
# opacity) your terminal uses.
#
# Render this single theme (from the repo root):
#   vhs doc/themes/tapes/terminal.tape
# Or re-render every theme at once:
#   doc/themes/render-themes.sh

Output doc/themes/tapes/_scratch.gif

Set Shell "bash"
Set FontSize 16
Set Width 1100
Set Height 640
Set Padding 20
# A dark VHS background so the transparent "terminal" theme reads correctly.
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

Type@60ms "/tmp/demo explore --theme terminal"
Enter
Sleep 3s
Down@250ms 4
Right
Sleep 600ms
Enter
Sleep 900ms
Down@250ms 1
Enter
Sleep 500ms
Type@80ms "9090"
Enter
Sleep 600ms
Tab
Sleep 1s
Screenshot doc/themes/terminal.png
Type "q"
Sleep 500ms
