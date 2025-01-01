package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

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
	ClearLine    TermColor = "\x1B[2K\r"
	ClearConsole TermColor = "\033[H\033[2J"
)

// SpinningCharacter represents a character that is being spun.
type SpinningCharacter struct {
	ch         rune
	timeout    int
	resetcount int
	done       bool
}

// NewSpinningCharacter returns a new instance of CharacterSpinning.
func NewSpinningCharacter(ch rune, timeout, resetcount int) *SpinningCharacter {
	return &SpinningCharacter{
		ch:         ch,
		timeout:    timeout,
		resetcount: resetcount,
		done:       false,
	}
}

// StartSpinning starts spinning the character.
func (cs *SpinningCharacter) StartSpinning(ctx context.Context) {
	go func() {
		var count int
		for {
			select {
			case <-ctx.Done(): // Stop spinning when context is canceled
				return
			default:
				fmt.Printf("%s\r*AI is thinking*>%s %s", Yellow, Reset, string(cs.ch))
				time.Sleep(time.Duration(cs.timeout) * time.Millisecond)

				if count%cs.resetcount == 0 {
					// Cycle through characters
					switch cs.ch {
					case '~':
						cs.ch = '!'
					case '!':
						cs.ch = '.'
					case '.':
						cs.ch = '-'
					case '-':
						cs.ch = '@'
					default:
						cs.ch = '~'
					}
				}

				count += 1
			}
		}
	}()
}

// gets the width of the current terminal
func getTerminalWidth(defaultWidth int) int {
	if w, ok := os.LookupEnv("COLUMNS"); ok {
		if parsedWidth, err := fmt.Sscanf(w, "%d", &defaultWidth); err == nil && parsedWidth > 0 {
			return parsedWidth
		}
	}
	if detectedWidth, _, err := term.GetSize(syscall.Stdin); err == nil {
		return detectedWidth
	}
	return defaultWidth
}

// printLine fills the terminal line with a specified character or a default '=' character.
func printLine(char rune) {
	width := getTerminalWidth(80)
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
