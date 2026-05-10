package tui

import "fmt"

// applyTerminalTheme sends ANSI OSC escape sequences to change the terminal's global background and foreground.
// OSC 10 changes foreground, OSC 11 changes background.
// OSC 110 and 111 reset them to default.
func applyTerminalTheme(isDark bool) {
	if isDark {
		// Default colors
		fmt.Print("\033]110\007\033]111\007")
	} else {
		// Light theme colors
		fmt.Print("\033]10;#111111\007\033]11;#FAFAFA\007")
	}
}
