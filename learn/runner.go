package learn

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

// matchPattern checks if input matches a regex pattern.
func matchPattern(input, pattern string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(strings.TrimSpace(input))
}

// Run starts the interactive tutorial.
func Run() {
	lessons := Lessons()
	reader := bufio.NewReader(os.Stdin)

	clearScreen()
	printWelcome()

	for i, lesson := range lessons {
		printLesson(i+1, len(lessons), lesson)

		for {
			fmt.Printf("%s  quill> %s", colorCyan, colorReset)
			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("\nGoodbye!")
				return
			}
			input = strings.TrimSpace(input)

			if input == "quit" || input == "exit" {
				fmt.Printf("\n%sProgress saved: lesson %d of %d. Run 'quill learn' to continue anytime.%s\n", colorDim, i+1, len(lessons), colorReset)
				return
			}

			if input == "skip" {
				fmt.Printf("%s  Skipping...%s\n\n", colorDim, colorReset)
				break
			}

			if input == "hint" {
				fmt.Printf("\n  %s%s%s\n\n", colorYellow, lesson.Hint, colorReset)
				continue
			}

			if input == "" {
				continue
			}

			if lesson.Validate(input) {
				fmt.Printf("\n  %s%s%s\n", colorGreen, lesson.Success, colorReset)
				if i < len(lessons)-1 {
					fmt.Printf("  %sPress Enter to continue...%s", colorDim, colorReset)
					reader.ReadString('\n')
				}
				fmt.Println()
				break
			} else {
				fmt.Printf("  %sNot quite. Type 'hint' for help, or try again.%s\n\n", colorRed, colorReset)
			}
		}
	}

	printComplete()
}

func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

func printWelcome() {
	fmt.Println()
	fmt.Printf("  %s%sWelcome to Quill!%s\n", colorBold, colorCyan, colorReset)
	fmt.Printf("  %sLet's write your first program - right here in the terminal.%s\n", colorDim, colorReset)
	fmt.Println()
	fmt.Printf("  %sCommands:%s  hint  - get help  |  skip  - skip lesson  |  exit  - quit\n", colorDim, colorReset)
	fmt.Println()
	fmt.Println("  " + strings.Repeat("─", 56))
	fmt.Println()
}

func printLesson(num, total int, lesson Lesson) {
	fmt.Printf("  %s%sLesson %d of %d: %s%s\n", colorBold, colorCyan, num, total, lesson.Title, colorReset)
	fmt.Println()

	// Print explanation with indentation
	for _, line := range strings.Split(lesson.Explanation, "\n") {
		fmt.Printf("  %s\n", line)
	}
	fmt.Println()
	fmt.Printf("  %s%s%s%s\n\n", colorBold, colorYellow, lesson.Prompt, colorReset)
}

func printComplete() {
	fmt.Println("  " + strings.Repeat("─", 56))
	fmt.Println()
	fmt.Printf("  %s%sYou've completed all 10 lessons!%s\n", colorBold, colorGreen, colorReset)
	fmt.Println()
	fmt.Println("  What's next:")
	fmt.Printf("  %s•%s  quill init            - start a real project\n", colorCyan, colorReset)
	fmt.Printf("  %s•%s  quill run hello.quill  - run a program\n", colorCyan, colorReset)
	fmt.Printf("  %s•%s  quill repl             - free-form playground\n", colorCyan, colorReset)
	fmt.Printf("  %s•%s  quill.tradebuddy.dev   - full documentation\n", colorCyan, colorReset)
	fmt.Println()
	fmt.Printf("  %sHappy coding!%s\n\n", colorBold, colorReset)
}
