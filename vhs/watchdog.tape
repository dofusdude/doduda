Output out/watchdog.gif

Require echo

Set Shell "bash"
Set FontSize 21
Set Width 1100
Set Height 350
Set Theme "Mirage"

Set FontFamily "Fira Code"

Set TypingSpeed 0.05

Type "doduda listen \" Sleep 200ms Enter
Type "  --hook https://httpbin.org/post \" Sleep 200ms Enter
Type "  --interval 5 \" Sleep 200ms Enter
Type "  --body ./webhook-body.json \" Sleep 200ms Enter
Type "  --auth-header 'Bearer 12345'" Sleep 700ms Enter

Sleep 4
