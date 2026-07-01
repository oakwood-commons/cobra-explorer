Output doc/themes/_scratch.gif

Set Shell "bash"
Set FontSize 16
Set Width 1100
Set Height 640
Set Padding 20
# A light VHS background so the transparent "terminal-light" theme reads
# correctly against a light terminal.
Set Theme "Catppuccin Latte"

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

Type@60ms "/tmp/demo explore --theme terminal-light"
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
Screenshot doc/themes/terminal-light.png
Type "q"
Sleep 500ms
