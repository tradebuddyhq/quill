package repl

// Windows: raw mode not implemented, falls back to bufio.Scanner
func enableRawMode() bool {
	return false
}

func disableRawMode() {}
