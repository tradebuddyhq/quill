package debugger

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
	colorBgGray = "\033[48;5;236m"
)

// StartREPL launches the interactive debugger REPL.
func StartREPL(quillFile string) {
	dbg := New(quillFile)

	fmt.Printf("%sQuill Debugger%s - %s\n", colorBold, colorReset, quillFile)
	fmt.Printf("Type %shelp%s for available commands.\n\n", colorCyan, colorReset)

	fmt.Print("Starting debugger...")
	if err := dbg.Start(); err != nil {
		fmt.Printf("\n%sError: %s%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
	fmt.Println(" connected.")
	fmt.Println()

	// Show initial context if paused
	if dbg.IsPaused() {
		showSource(dbg, dbg.CurrentLine(), 5)
	}

	// Handle Ctrl+C gracefully
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nStopping debugger...")
		dbg.Stop()
		os.Exit(0)
	}()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Printf("%s(quill-debug) %s", colorGreen, colorReset)
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		cmd := parts[0]
		arg := ""
		if len(parts) > 1 {
			arg = strings.TrimSpace(parts[1])
		}

		switch cmd {
		case "break", "b":
			cmdBreak(dbg, arg)
		case "delete", "d":
			cmdDelete(dbg, arg)
		case "continue", "c":
			cmdContinue(dbg)
		case "step", "s":
			cmdStep(dbg)
		case "into", "i":
			cmdInto(dbg)
		case "out", "o":
			cmdOut(dbg)
		case "print", "p":
			cmdPrint(dbg, arg)
		case "locals", "l":
			cmdLocals(dbg)
		case "stack", "bt":
			cmdStack(dbg)
		case "list", "ls":
			cmdList(dbg, arg)
		case "help", "h":
			cmdHelp()
		case "quit", "q":
			fmt.Println("Stopping debugger...")
			dbg.Stop()
			return
		default:
			fmt.Printf("%sUnknown command: %s%s (type 'help' for commands)\n", colorRed, cmd, colorReset)
		}
	}

	dbg.Stop()
}

func cmdBreak(dbg *Debugger, arg string) {
	if arg == "" {
		// List breakpoints
		if len(dbg.Breakpoints()) == 0 {
			fmt.Println("No breakpoints set.")
			return
		}
		lines := make([]int, 0, len(dbg.Breakpoints()))
		for l := range dbg.Breakpoints() {
			lines = append(lines, l)
		}
		sort.Ints(lines)
		fmt.Println("Breakpoints:")
		for _, l := range lines {
			src := ""
			if l >= 1 && l <= len(dbg.QuillLines()) {
				src = strings.TrimSpace(dbg.QuillLines()[l-1])
			}
			fmt.Printf("  %sline %d%s: %s\n", colorRed, l, colorReset, src)
		}
		return
	}

	lineNum, err := strconv.Atoi(arg)
	if err != nil {
		fmt.Printf("%sUsage: break <line_number>%s\n", colorRed, colorReset)
		return
	}

	if err := dbg.SetBreakpoint(lineNum); err != nil {
		fmt.Printf("%sError: %s%s\n", colorRed, err, colorReset)
		return
	}

	src := ""
	if lineNum >= 1 && lineNum <= len(dbg.QuillLines()) {
		src = strings.TrimSpace(dbg.QuillLines()[lineNum-1])
	}
	fmt.Printf("Breakpoint set at %sline %d%s: %s\n", colorYellow, lineNum, colorReset, src)
}

func cmdDelete(dbg *Debugger, arg string) {
	if arg == "" {
		fmt.Printf("%sUsage: delete <line_number>%s\n", colorRed, colorReset)
		return
	}

	lineNum, err := strconv.Atoi(arg)
	if err != nil {
		fmt.Printf("%sUsage: delete <line_number>%s\n", colorRed, colorReset)
		return
	}

	if err := dbg.RemoveBreakpoint(lineNum); err != nil {
		fmt.Printf("%sError: %s%s\n", colorRed, err, colorReset)
		return
	}

	fmt.Printf("Breakpoint removed from line %d\n", lineNum)
}

func cmdContinue(dbg *Debugger) {
	if !dbg.IsPaused() {
		fmt.Println("Program is not paused.")
		return
	}

	fmt.Println("Continuing...")
	if err := dbg.Continue(); err != nil {
		fmt.Printf("%sError: %s%s\n", colorRed, err, colorReset)
		return
	}

	if dbg.IsPaused() {
		fmt.Printf("Paused at %sline %d%s\n", colorYellow, dbg.CurrentLine(), colorReset)
		showSource(dbg, dbg.CurrentLine(), 5)
	} else {
		fmt.Println("Program finished.")
	}
}

func cmdStep(dbg *Debugger) {
	if !dbg.IsPaused() {
		fmt.Println("Program is not paused.")
		return
	}

	if err := dbg.StepOver(); err != nil {
		fmt.Printf("%sError: %s%s\n", colorRed, err, colorReset)
		return
	}

	if dbg.IsPaused() {
		showSource(dbg, dbg.CurrentLine(), 3)
	} else {
		fmt.Println("Program finished.")
	}
}

func cmdInto(dbg *Debugger) {
	if !dbg.IsPaused() {
		fmt.Println("Program is not paused.")
		return
	}

	if err := dbg.StepInto(); err != nil {
		fmt.Printf("%sError: %s%s\n", colorRed, err, colorReset)
		return
	}

	if dbg.IsPaused() {
		showSource(dbg, dbg.CurrentLine(), 3)
	} else {
		fmt.Println("Program finished.")
	}
}

func cmdOut(dbg *Debugger) {
	if !dbg.IsPaused() {
		fmt.Println("Program is not paused.")
		return
	}

	if err := dbg.StepOut(); err != nil {
		fmt.Printf("%sError: %s%s\n", colorRed, err, colorReset)
		return
	}

	if dbg.IsPaused() {
		showSource(dbg, dbg.CurrentLine(), 3)
	} else {
		fmt.Println("Program finished.")
	}
}

func cmdPrint(dbg *Debugger, expr string) {
	if expr == "" {
		fmt.Printf("%sUsage: print <expression>%s\n", colorRed, colorReset)
		return
	}

	if !dbg.IsPaused() {
		fmt.Println("Program is not paused.")
		return
	}

	result, err := dbg.Evaluate(expr)
	if err != nil {
		fmt.Printf("%sError: %s%s\n", colorRed, err, colorReset)
		return
	}

	fmt.Printf("%s = %s%s%s\n", expr, colorCyan, result, colorReset)
}

func cmdLocals(dbg *Debugger) {
	if !dbg.IsPaused() {
		fmt.Println("Program is not paused.")
		return
	}

	locals, err := dbg.GetLocals()
	if err != nil {
		fmt.Printf("%sError: %s%s\n", colorRed, err, colorReset)
		return
	}

	if len(locals) == 0 {
		fmt.Println("No local variables.")
		return
	}

	// Sort for consistent display
	names := make([]string, 0, len(locals))
	for name := range locals {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Println("Local variables:")
	for _, name := range names {
		fmt.Printf("  %s%s%s = %s%s%s\n", colorBold, name, colorReset, colorCyan, locals[name], colorReset)
	}
}

func cmdStack(dbg *Debugger) {
	if !dbg.IsPaused() {
		fmt.Println("Program is not paused.")
		return
	}

	frames, err := dbg.GetCallStack()
	if err != nil {
		fmt.Printf("%sError: %s%s\n", colorRed, err, colorReset)
		return
	}

	if len(frames) == 0 {
		fmt.Println("Empty call stack.")
		return
	}

	fmt.Println("Call stack:")
	for i, f := range frames {
		name := f.FunctionName
		if name == "" {
			name = "<top-level>"
		}
		marker := "  "
		if i == 0 {
			marker = fmt.Sprintf("%s> %s", colorGreen, colorReset)
		}
		if f.QuillLine > 0 {
			fmt.Printf("%s#%d  %s%s%s at line %d\n", marker, i, colorBold, name, colorReset, f.QuillLine)
		} else {
			fmt.Printf("%s#%d  %s%s%s (runtime)\n", marker, i, colorBold, name, colorReset)
		}
	}
}

func cmdList(dbg *Debugger, arg string) {
	line := dbg.CurrentLine()
	context := 10

	if arg != "" {
		if n, err := strconv.Atoi(arg); err == nil {
			line = n
		}
	}

	showSource(dbg, line, context)
}

func cmdHelp() {
	fmt.Println("Commands:")
	fmt.Printf("  %sbreak%s <line>    %sb%s <line>    Set a breakpoint (no arg: list breakpoints)\n", colorBold, colorReset, colorBold, colorReset)
	fmt.Printf("  %sdelete%s <line>   %sd%s <line>    Remove a breakpoint\n", colorBold, colorReset, colorBold, colorReset)
	fmt.Printf("  %scontinue%s        %sc%s            Continue execution\n", colorBold, colorReset, colorBold, colorReset)
	fmt.Printf("  %sstep%s            %ss%s            Step over (next line)\n", colorBold, colorReset, colorBold, colorReset)
	fmt.Printf("  %sinto%s            %si%s            Step into function call\n", colorBold, colorReset, colorBold, colorReset)
	fmt.Printf("  %sout%s             %so%s            Step out of function\n", colorBold, colorReset, colorBold, colorReset)
	fmt.Printf("  %sprint%s <expr>    %sp%s <expr>    Evaluate and print expression\n", colorBold, colorReset, colorBold, colorReset)
	fmt.Printf("  %slocals%s          %sl%s            Show local variables\n", colorBold, colorReset, colorBold, colorReset)
	fmt.Printf("  %sstack%s           %sbt%s           Show call stack\n", colorBold, colorReset, colorBold, colorReset)
	fmt.Printf("  %slist%s [line]     %sls%s [line]   Show source around line\n", colorBold, colorReset, colorBold, colorReset)
	fmt.Printf("  %shelp%s            %sh%s            Show this help\n", colorBold, colorReset, colorBold, colorReset)
	fmt.Printf("  %squit%s            %sq%s            Stop debugging and exit\n", colorBold, colorReset, colorBold, colorReset)
}

// showSource displays Quill source lines around the given line, highlighting
// the current line and marking breakpoints.
func showSource(dbg *Debugger, centerLine int, context int) {
	lines := dbg.QuillLines()
	if len(lines) == 0 || centerLine < 1 {
		return
	}

	start := centerLine - context
	if start < 1 {
		start = 1
	}
	end := centerLine + context
	if end > len(lines) {
		end = len(lines)
	}

	fmt.Println()
	for i := start; i <= end; i++ {
		lineText := lines[i-1]

		// Markers
		bpMarker := " "
		if _, hasBP := dbg.Breakpoints()[i]; hasBP {
			bpMarker = fmt.Sprintf("%s*%s", colorRed, colorReset)
		}

		arrow := "  "
		if i == dbg.CurrentLine() {
			arrow = fmt.Sprintf("%s->%s", colorGreen, colorReset)
		}

		lineNum := fmt.Sprintf("%s%4d%s", colorGray, i, colorReset)

		if i == dbg.CurrentLine() {
			fmt.Printf(" %s %s %s %s%s%s%s\n", bpMarker, arrow, lineNum, colorBgGray, colorBold, lineText, colorReset)
		} else {
			fmt.Printf(" %s %s %s %s\n", bpMarker, arrow, lineNum, lineText)
		}
	}
	fmt.Println()
}

// ParseREPLCommand parses a REPL input line into command and argument.
// Exported for testing.
func ParseREPLCommand(input string) (cmd string, arg string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", ""
	}
	parts := strings.SplitN(input, " ", 2)
	cmd = parts[0]
	if len(parts) > 1 {
		arg = strings.TrimSpace(parts[1])
	}
	return cmd, arg
}

// ResolveCommandAlias normalizes a command alias to its canonical name.
// Exported for testing.
func ResolveCommandAlias(cmd string) string {
	switch cmd {
	case "b":
		return "break"
	case "d":
		return "delete"
	case "c":
		return "continue"
	case "s":
		return "step"
	case "i":
		return "into"
	case "o":
		return "out"
	case "p":
		return "print"
	case "l":
		return "locals"
	case "bt":
		return "stack"
	case "ls":
		return "list"
	case "h":
		return "help"
	case "q":
		return "quit"
	default:
		return cmd
	}
}
