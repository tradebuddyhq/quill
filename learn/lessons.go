package learn

// Lesson represents a single interactive tutorial step.
type Lesson struct {
	Title       string
	Explanation string
	Prompt      string   // What to tell the user to type
	Validate    func(input string) bool
	Hint        string // Shown when validation fails
	Success     string // Shown when validation passes
}

// Lessons returns the full list of interactive tutorial lessons.
func Lessons() []Lesson {
	return []Lesson{
		{
			Title:       "Variables",
			Explanation: "A variable is like a labeled box that holds a value.\nIn Quill, you create one with:  name is value",
			Prompt:      `Try it: type   name is "your name"   (use any name you like)`,
			Validate: func(input string) bool {
				return matchPattern(input, `^\w+ is ".+"$`)
			},
			Hint:    `Use the format:  name is "Alice"  — a word, then "is", then a value in quotes.`,
			Success: "You just created a variable! Quill remembers it for you.",
		},
		{
			Title:       "Printing",
			Explanation: `To show something on screen, use "say".`,
			Prompt:      `Try it: type   say "Hello, World!"`,
			Validate: func(input string) bool {
				return matchPattern(input, `^say ".+"$`)
			},
			Hint:    `Type: say "Hello, World!"  — the word say, then your message in quotes.`,
			Success: "That's how you print in Quill. No parentheses needed!",
		},
		{
			Title:       "Numbers & Math",
			Explanation: "Variables can hold numbers too. You can do math with them.\nQuill supports +, -, *, and /.",
			Prompt:      "Try it: type   total is 10 + 5",
			Validate: func(input string) bool {
				return matchPattern(input, `^\w+ is \d+\s*[+\-*/]\s*\d+$`)
			},
			Hint:    "Use the format:  total is 10 + 5  — a name, 'is', then a math expression.",
			Success: "Nice! You can use +, -, *, / just like a calculator.",
		},
		{
			Title:       "Lists",
			Explanation: "A list holds multiple values inside square brackets.\nYou separate items with commas.",
			Prompt:      `Try it: type   colors are ["red", "blue", "green"]`,
			Validate: func(input string) bool {
				return matchPattern(input, `^\w+ (is|are) \[.+\]$`)
			},
			Hint:    `Use the format:  colors are ["red", "blue", "green"]  — square brackets with items inside.`,
			Success: "Lists are one of the most useful things in programming. You'll use them everywhere.",
		},
		{
			Title:       "If / Otherwise",
			Explanation: "You can make decisions with 'if' and 'otherwise'.\n\n  if age is greater than 18:\n    say \"Welcome!\"\n  otherwise:\n    say \"Too young.\"",
			Prompt:      "Try it: type   if 10 is greater than 5:",
			Validate: func(input string) bool {
				return matchPattern(input, `^if .+(is greater than|is less than|is|is not) .+:$`)
			},
			Hint:    "Type a condition with a colon at the end:  if 10 is greater than 5:",
			Success: "Conditions let your program make decisions. The colon starts a block.",
		},
		{
			Title:       "Functions",
			Explanation: "A function is a reusable block of code.\nIn Quill, you define one with 'to':\n\n  to greet name:\n    say \"Hello, {name}!\"",
			Prompt:      "Try it: type   to greet name:",
			Validate: func(input string) bool {
				return matchPattern(input, `^to \w+.*:$`)
			},
			Hint:    "Use the format:  to greet name:  — the word 'to', a function name, optional parameters, and a colon.",
			Success: "Functions let you write code once and reuse it. The colon means 'here's what it does'.",
		},
		{
			Title:       "Loops",
			Explanation: "To repeat something for each item in a list, use 'for each':\n\n  for each color in colors:\n    say color",
			Prompt:      `Try it: type   for each item in ["a", "b", "c"]:`,
			Validate: func(input string) bool {
				return matchPattern(input, `^for each \w+ in .+:$`)
			},
			Hint:    `Type:  for each item in ["a", "b", "c"]:  — 'for each', a variable name, 'in', a list, and a colon.`,
			Success: "Loops save you from writing the same code over and over.",
		},
		{
			Title:       "String Interpolation",
			Explanation: "You can put variables inside strings using curly braces:\n\n  name is \"World\"\n  say \"Hello, {name}!\"",
			Prompt:      `Try it: type   say "I am {age} years old"   (or use any variable name)`,
			Validate: func(input string) bool {
				return matchPattern(input, `^say ".*\{.+\}.*"$`)
			},
			Hint:    `Put a variable name inside curly braces within quotes:  say "Hello, {name}!"`,
			Success: "String interpolation makes it easy to build dynamic text. No concatenation needed!",
		},
		{
			Title:       "Fetching Data",
			Explanation: "Quill has built-in functions for working with the web.\nfetchJSON grabs data from any API:\n\n  data is await fetchJSON(\"https://api.example.com/users\")",
			Prompt:      `Try it: type   data is await fetchJSON("https://api.example.com")   (any URL works)`,
			Validate: func(input string) bool {
				return matchPattern(input, `^\w+ is await fetchJSON\(.+\)$`)
			},
			Hint:    `Use:  data is await fetchJSON("https://api.example.com")  — 'await' because it talks to the internet.`,
			Success: "That's it — one line to fetch data from the web. No imports, no setup.",
		},
		{
			Title:       "Putting It Together",
			Explanation: "Let's combine what you've learned!\nWrite a function that takes a name and says hello to them.",
			Prompt:      `Try it: type   to sayHello name:`,
			Validate: func(input string) bool {
				return matchPattern(input, `^to \w+ \w+.*:$`)
			},
			Hint:    "Define a function with a parameter:  to sayHello name:",
			Success: "You've completed the Quill tutorial! You know variables, functions, loops, conditions, and more.",
		},
	}
}
