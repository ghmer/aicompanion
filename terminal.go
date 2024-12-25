package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// ANSI color codes
type TermColor string

const (
	// Standard Colors
	Black   TermColor = "\033[30m"
	Red     TermColor = "\033[31m"
	Green   TermColor = "\033[32m"
	Yellow  TermColor = "\033[33m"
	Blue    TermColor = "\033[34m"
	Magenta TermColor = "\033[35m"
	Cyan    TermColor = "\033[36m"
	White   TermColor = "\033[37m"

	// Bright Colors
	BrightBlack   TermColor = "\033[90m"
	BrightRed     TermColor = "\033[91m"
	BrightGreen   TermColor = "\033[92m"
	BrightYellow  TermColor = "\033[93m"
	BrightBlue    TermColor = "\033[94m"
	BrightMagenta TermColor = "\033[95m"
	BrightCyan    TermColor = "\033[96m"
	BrightWhite   TermColor = "\033[97m"

	// Background Colors
	BgBlack         TermColor = "\033[40m"
	BgRed           TermColor = "\033[41m"
	BgGreen         TermColor = "\033[42m"
	BgYellow        TermColor = "\033[43m"
	BgBlue          TermColor = "\033[44m"
	BgMagenta       TermColor = "\033[45m"
	BgCyan          TermColor = "\033[46m"
	BgWhite         TermColor = "\033[47m"
	BgBrightBlack   TermColor = "\033[100m"
	BgBrightRed     TermColor = "\033[101m"
	BgBrightGreen   TermColor = "\033[102m"
	BgBrightYellow  TermColor = "\033[103m"
	BgBrightBlue    TermColor = "\033[104m"
	BgBrightMagenta TermColor = "\033[105m"
	BgBrightCyan    TermColor = "\033[106m"
	BgBrightWhite   TermColor = "\033[107m"

	// Text Styles
	Bold         TermColor = "\033[1m"
	Underline    TermColor = "\033[4m"
	Reset        TermColor = "\033[0m"
	ClearConsole TermColor = "\033[H\033[2J"
)

// printLine fills the terminal line with a specified character or a default '=' character.
func printLine(char rune) {
	width := 80 // Default width fallback
	if w, ok := os.LookupEnv("COLUMNS"); ok {
		if parsedWidth, err := fmt.Sscanf(w, "%d", &width); err == nil && parsedWidth > 0 {
			width = parsedWidth
		}
	} else {
		if detectedWidth, _, err := term.GetSize(syscall.Stdin); err == nil {
			width = detectedWidth
		}
	}

	if char == 0 {
		char = '='
	}
	line := strings.Repeat(string(char), width)
	fmt.Println(line)
}

// clearConsole clears the console screen.
func clearConsole() {
	fmt.Print(ClearConsole)
}
