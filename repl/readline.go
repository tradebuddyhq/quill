package repl

import (
	"bufio"
	"fmt"
	"os"
)

// readline provides line editing with arrow-key history for the REPL.
type readline struct {
	history []string
	rawMode bool
	scanner *bufio.Scanner
}

func newReadline() *readline {
	rl := &readline{}
	rl.rawMode = enableRawMode()
	if !rl.rawMode {
		rl.scanner = bufio.NewScanner(os.Stdin)
	}
	return rl
}

func (rl *readline) close() {
	if rl.rawMode {
		disableRawMode()
	}
}

func (rl *readline) readLine(prompt string) (string, bool) {
	if !rl.rawMode {
		return rl.readLineFallback(prompt)
	}

	fmt.Print(prompt)

	buf := make([]byte, 0, 256)
	pos := 0
	histIdx := len(rl.history)
	var saved []byte

	for {
		var b [1]byte
		n, err := os.Stdin.Read(b[:])
		if n == 0 || err != nil {
			fmt.Println()
			return "", false
		}

		ch := b[0]

		switch {
		case ch == 3: // Ctrl+C
			fmt.Println()
			return "exit", true

		case ch == 4: // Ctrl+D
			if len(buf) == 0 {
				fmt.Println()
				return "", false
			}

		case ch == 13 || ch == 10: // Enter
			fmt.Println()
			line := string(buf)
			if line != "" {
				rl.history = append(rl.history, line)
			}
			return line, true

		case ch == 127 || ch == 8: // Backspace
			if pos > 0 {
				buf = append(buf[:pos-1], buf[pos:]...)
				pos--
				rl.refreshLine(prompt, buf, pos)
			}

		case ch == 27: // Escape sequence (arrow keys)
			var seq [2]byte
			os.Stdin.Read(seq[:1])
			if seq[0] == '[' {
				os.Stdin.Read(seq[1:])
				switch seq[1] {
				case 'A': // Up arrow
					if histIdx > 0 {
						if histIdx == len(rl.history) {
							saved = make([]byte, len(buf))
							copy(saved, buf)
						}
						histIdx--
						buf = []byte(rl.history[histIdx])
						pos = len(buf)
						rl.refreshLine(prompt, buf, pos)
					}
				case 'B': // Down arrow
					if histIdx < len(rl.history) {
						histIdx++
						if histIdx == len(rl.history) {
							if saved != nil {
								buf = saved
							} else {
								buf = nil
							}
						} else {
							buf = []byte(rl.history[histIdx])
						}
						pos = len(buf)
						rl.refreshLine(prompt, buf, pos)
					}
				case 'C': // Right arrow
					if pos < len(buf) {
						pos++
						fmt.Print("\x1b[C")
					}
				case 'D': // Left arrow
					if pos > 0 {
						pos--
						fmt.Print("\x1b[D")
					}
				}
			}

		case ch == 1: // Ctrl+A - beginning of line
			pos = 0
			rl.refreshLine(prompt, buf, pos)

		case ch == 5: // Ctrl+E - end of line
			pos = len(buf)
			rl.refreshLine(prompt, buf, pos)

		case ch == 21: // Ctrl+U - clear line
			buf = nil
			pos = 0
			rl.refreshLine(prompt, buf, pos)

		case ch >= 32: // Printable character
			if pos == len(buf) {
				buf = append(buf, ch)
			} else {
				buf = append(buf[:pos+1], buf[pos:]...)
				buf[pos] = ch
			}
			pos++
			rl.refreshLine(prompt, buf, pos)
		}
	}
}

func (rl *readline) refreshLine(prompt string, buf []byte, pos int) {
	fmt.Printf("\r\x1b[K%s%s", prompt, string(buf))
	if pos < len(buf) {
		fmt.Printf("\x1b[%dD", len(buf)-pos)
	}
}

func (rl *readline) readLineFallback(prompt string) (string, bool) {
	fmt.Print(prompt)
	if !rl.scanner.Scan() {
		return "", false
	}
	return rl.scanner.Text(), true
}
