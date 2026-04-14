package lsp

import (
	"fmt"
	"quill/ast"
	"strings"
)

// HoverProvider computes hover information for a position in a document.
type HoverProvider struct{}

func NewHoverProvider() *HoverProvider {
	return &HoverProvider{}
}

// GetHover returns hover information for the given position.
func (h *HoverProvider) GetHover(doc *Document, pos Position, program *ast.Program) *Hover {
	word := doc.GetWordAtPosition(pos)
	if word == "" {
		return nil
	}

	// Check keywords
	if info, ok := keywordDocs[word]; ok {
		return &Hover{
			Contents: MarkupContent{
				Kind:  "markdown",
				Value: fmt.Sprintf("**%s** (keyword)\n\n%s", word, info),
			},
		}
	}

	// Check stdlib functions
	if info, ok := stdlibDocs[word]; ok {
		return &Hover{
			Contents: MarkupContent{
				Kind:  "markdown",
				Value: fmt.Sprintf("**%s** (stdlib)\n\n%s\n\n%s", word, info.Signature, info.Doc),
			},
		}
	}

	// Check user-defined functions
	if program != nil {
		for _, stmt := range program.Statements {
			if fn, ok := stmt.(*ast.FuncDefinition); ok {
				if fn.Name == word {
					sig := buildFuncSignature(fn)
					return &Hover{
						Contents: MarkupContent{
							Kind:  "markdown",
							Value: fmt.Sprintf("**%s** (function)\n\n```quill\n%s\n```", word, sig),
						},
					}
				}
			}
		}

		// Check user-defined variables
		for _, stmt := range program.Statements {
			if assign, ok := stmt.(*ast.AssignStatement); ok {
				if assign.Name == word {
					typeStr := inferExprTypeLabel(assign.Value)
					return &Hover{
						Contents: MarkupContent{
							Kind:  "markdown",
							Value: fmt.Sprintf("**%s** (variable)\n\nType: `%s`", word, typeStr),
						},
					}
				}
			}
		}

		// Check classes
		for _, stmt := range program.Statements {
			if desc, ok := stmt.(*ast.DescribeStatement); ok {
				if desc.Name == word {
					info := fmt.Sprintf("**%s** (class)", word)
					if desc.Extends != "" {
						info += fmt.Sprintf("\n\nExtends: `%s`", desc.Extends)
					}
					if len(desc.Properties) > 0 {
						info += "\n\nProperties:"
						for _, p := range desc.Properties {
							info += fmt.Sprintf("\n- `%s`", p.Name)
						}
					}
					if len(desc.Methods) > 0 {
						info += "\n\nMethods:"
						for _, m := range desc.Methods {
							info += fmt.Sprintf("\n- `%s`", buildFuncSignature(&m))
						}
					}
					return &Hover{
						Contents: MarkupContent{Kind: "markdown", Value: info},
					}
				}
			}
		}
	}

	return nil
}

func buildFuncSignature(fn *ast.FuncDefinition) string {
	var params []string
	for i, p := range fn.Params {
		if i < len(fn.ParamTypes) && fn.ParamTypes[i] != "" {
			params = append(params, fmt.Sprintf("%s: %s", p, fn.ParamTypes[i]))
		} else {
			params = append(params, p)
		}
	}
	sig := fmt.Sprintf("to %s %s", fn.Name, strings.Join(params, ", "))
	if fn.ReturnType != "" {
		sig += " -> " + fn.ReturnType
	}
	return sig
}

func inferExprTypeLabel(expr ast.Expression) string {
	switch expr.(type) {
	case *ast.StringLiteral:
		return "text"
	case *ast.NumberLiteral:
		return "number"
	case *ast.BoolLiteral:
		return "boolean"
	case *ast.ListLiteral:
		return "list"
	case *ast.ObjectLiteral:
		return "object"
	case *ast.NothingLiteral:
		return "nothing"
	case *ast.LambdaExpr:
		return "function"
	default:
		return "any"
	}
}

// StdlibInfo holds documentation for a stdlib function.
type StdlibInfo struct {
	Signature string
	Doc       string
	RetType   string
}

var keywordDocs = map[string]string{
	"is":        "Assignment keyword. Assigns a value to a variable.\n\nExample: `name is \"Alice\"`",
	"are":       "Assignment keyword (plural form). Assigns a value to a variable.\n\nExample: `colors are [\"red\", \"blue\"]`",
	"say":       "Output keyword. Prints a value to the console.\n\nExample: `say \"Hello!\"`",
	"if":        "Conditional keyword. Starts a conditional block.\n\nExample: `if x greater than 5`",
	"otherwise": "Else keyword. Provides an alternative branch.\n\nExample: `otherwise say \"nope\"`",
	"for":       "Loop keyword. Used with `each` for iteration.\n\nExample: `for each item in list`",
	"each":      "Iteration keyword. Used with `for` to iterate over collections.",
	"in":        "Membership keyword. Used in `for each ... in` loops.",
	"to":        "Function definition keyword. Defines a new function.\n\nExample: `to greet name`",
	"give":      "Return keyword (first half). Used as `give back` to return a value.",
	"back":      "Return keyword (second half). Used as `give back` to return a value.\n\nExample: `give back result`",
	"while":     "Loop keyword. Repeats a block while a condition is true.\n\nExample: `while count less than 10`",
	"try":       "Error handling keyword. Starts a try block.\n\nExample: `try`",
	"fails":     "Error handling keyword. Catches errors from a try block.\n\nExample: `fails with error`",
	"match":     "Pattern matching keyword. Matches a value against patterns.\n\nExample: `match color`",
	"when":      "Pattern case keyword. Defines a case in a match block.\n\nExample: `when \"red\"`",
	"define":    "Type definition keyword. Defines an enum/algebraic type.\n\nExample: `define Color`",
	"describe":  "Class definition keyword. Defines a class with properties and methods.\n\nExample: `describe Animal`",
	"use":       "Import keyword. Imports a module or package.\n\nExample: `use \"math\"`",
	"from":      "Import keyword. Selective imports.\n\nExample: `from \"utils\" use helper`",
	"new":       "Constructor keyword. Creates a new instance of a class.\n\nExample: `new Animal \"Rex\"`",
	"my":        "Self-reference keyword. Refers to the current instance in a class.\n\nExample: `my name is name`",
	"await":     "Async keyword. Waits for an async operation to complete.\n\nExample: `await fetch url`",
	"as":        "Alias keyword. Gives an imported module an alias.\n\nExample: `use \"http\" as web`",
	"with":      "Used with `fails with` for error variable binding.",
	"test":      "Test block keyword. Defines a test.\n\nExample: `test \"addition works\"`",
	"expect":    "Assertion keyword. Used in tests to assert conditions.\n\nExample: `expect result equal 42`",
	"and":       "Logical AND operator.\n\nExample: `if x and y`",
	"or":        "Logical OR operator.\n\nExample: `if x or y`",
	"not":       "Logical NOT operator.\n\nExample: `if not done`",
	"greater":   "Comparison keyword. Used as `greater than`.\n\nExample: `if x greater than 5`",
	"less":      "Comparison keyword. Used as `less than`.\n\nExample: `if x less than 5`",
	"than":      "Comparison keyword. Used after `greater` or `less`.",
	"equal":     "Comparison keyword. Tests equality.\n\nExample: `if x equal 5`",
	"contains":  "Membership test. Checks if a collection contains an element.\n\nExample: `if list contains item`",
	"extends":   "Inheritance keyword. A class extends another.\n\nExample: `describe Dog extends Animal`",
	"nothing":   "Null value keyword. Represents the absence of a value.",
	"yes":       "Boolean true literal.",
	"no":        "Boolean false literal.",
	"true":      "Boolean true literal (alias for `yes`).",
	"false":     "Boolean false literal (alias for `no`).",
	"break":     "Loop control. Exits the current loop.\n\nExample: `break`",
	"continue":  "Loop control. Skips to the next iteration.\n\nExample: `continue`",
	"of":        "Type annotation keyword. Used in generics like `list of number`.",
}

var stdlibDocs = map[string]StdlibInfo{
	// Text functions
	"length":        {Signature: "length(value) -> number", Doc: "Returns the length of a text string or list.", RetType: "number"},
	"toText":        {Signature: "toText(value) -> text", Doc: "Converts any value to its text representation.", RetType: "text"},
	"toNumber":      {Signature: "toNumber(value) -> number", Doc: "Converts a text string to a number.", RetType: "number"},
	"trim":          {Signature: "trim(text) -> text", Doc: "Removes whitespace from both ends of a string.", RetType: "text"},
	"upper":         {Signature: "upper(text) -> text", Doc: "Converts text to uppercase.", RetType: "text"},
	"lower":         {Signature: "lower(text) -> text", Doc: "Converts text to lowercase.", RetType: "text"},
	"capitalize":    {Signature: "capitalize(text) -> text", Doc: "Capitalizes the first letter of text.", RetType: "text"},
	"split":         {Signature: "split(text, separator) -> list", Doc: "Splits text into a list by separator.", RetType: "list"},
	"join":          {Signature: "join(list, separator) -> text", Doc: "Joins a list into text with a separator.", RetType: "text"},
	"replace_text":  {Signature: "replace_text(text, old, new) -> text", Doc: "Replaces occurrences of old with new in text.", RetType: "text"},
	"includes":      {Signature: "includes(text, search) -> boolean", Doc: "Checks if text contains the search string.", RetType: "boolean"},
	"startsWith":    {Signature: "startsWith(text, prefix) -> boolean", Doc: "Checks if text starts with prefix.", RetType: "boolean"},
	"endsWith":      {Signature: "endsWith(text, suffix) -> boolean", Doc: "Checks if text ends with suffix.", RetType: "boolean"},
	"indexOf":       {Signature: "indexOf(text, search) -> number", Doc: "Returns the index of search in text, or -1.", RetType: "number"},
	"truncate":      {Signature: "truncate(text, maxLength) -> text", Doc: "Truncates text to maxLength characters.", RetType: "text"},
	"words":         {Signature: "words(text) -> list", Doc: "Splits text into a list of words.", RetType: "list"},
	"lines":         {Signature: "lines(text) -> list", Doc: "Splits text into a list of lines.", RetType: "list"},
	"regex":          {Signature: "regex(pattern, flags?) -> RegExp", Doc: "Creates a RegExp object. Flags: 'g' (global), 'i' (case-insensitive), 'm' (multiline).", RetType: "object"},
	"matches":        {Signature: "matches(text, pattern, flags?) -> list", Doc: "Returns all regex matches in text. Default flag is 'g'.", RetType: "list"},
	"matchGroups":    {Signature: "matchGroups(text, pattern, flags?) -> list", Doc: "Returns a list of capture group arrays for each match.", RetType: "list"},
	"matchFirst":     {Signature: "matchFirst(text, pattern, flags?) -> object|nothing", Doc: "Returns first match with {match, groups, index} or nothing.", RetType: "object"},
	"matchesPattern": {Signature: "matchesPattern(text, pattern, flags?) -> boolean", Doc: "Tests if text matches a regex pattern.", RetType: "boolean"},
	"encodeBase64":  {Signature: "encodeBase64(text) -> text", Doc: "Encodes text as Base64.", RetType: "text"},
	"decodeBase64":  {Signature: "decodeBase64(text) -> text", Doc: "Decodes Base64 text.", RetType: "text"},
	"encodeURL":     {Signature: "encodeURL(text) -> text", Doc: "URL-encodes text.", RetType: "text"},
	"decodeURL":     {Signature: "decodeURL(text) -> text", Doc: "URL-decodes text.", RetType: "text"},

	// Number functions
	"round":     {Signature: "round(number) -> number", Doc: "Rounds a number to the nearest integer.", RetType: "number"},
	"floor":     {Signature: "floor(number) -> number", Doc: "Rounds down to the nearest integer.", RetType: "number"},
	"ceil":      {Signature: "ceil(number) -> number", Doc: "Rounds up to the nearest integer.", RetType: "number"},
	"abs":       {Signature: "abs(number) -> number", Doc: "Returns the absolute value.", RetType: "number"},
	"random":    {Signature: "random() -> number", Doc: "Returns a random number between 0 and 1.", RetType: "number"},
	"randomInt":    {Signature: "randomInt(min, max) -> number", Doc: "Returns a random integer between min and max.", RetType: "number"},
	"formatNumber": {Signature: "formatNumber(n, decimals?) -> text", Doc: "Formats a number with locale separators or fixed decimal places.", RetType: "text"},
	"toFixed":      {Signature: "toFixed(n, digits) -> number", Doc: "Rounds a number to a fixed number of decimal places.", RetType: "number"},
	"percent":      {Signature: "percent(n, decimals?) -> text", Doc: "Formats a decimal as a percentage string (e.g. 0.85 -> '85.0%').", RetType: "text"},
	"currency":     {Signature: "currency(n, symbol?) -> text", Doc: "Formats a number as currency with commas (default '$').", RetType: "text"},
	"ordinal":      {Signature: "ordinal(n) -> text", Doc: "Returns a number with its ordinal suffix (1st, 2nd, 3rd, etc.).", RetType: "text"},
	"clamp":        {Signature: "clamp(n, min, max) -> number", Doc: "Clamps a number between min and max.", RetType: "number"},
	"lerp":         {Signature: "lerp(a, b, t) -> number", Doc: "Linear interpolation between a and b by factor t (0-1).", RetType: "number"},
	"mapRange":     {Signature: "mapRange(n, inMin, inMax, outMin, outMax) -> number", Doc: "Maps a number from one range to another.", RetType: "number"},

	// List functions
	"push":     {Signature: "push(list, item) -> list", Doc: "Adds an item to the end of a list.", RetType: "list"},
	"sort":     {Signature: "sort(list) -> list", Doc: "Sorts a list in ascending order.", RetType: "list"},
	"sortBy":   {Signature: "sortBy(list, field) -> list", Doc: "Sorts a list of objects by a field name in ascending order.", RetType: "list"},
	"reverse":  {Signature: "reverse(list) -> list", Doc: "Reverses a list.", RetType: "list"},
	"unique":   {Signature: "unique(list) -> list", Doc: "Returns unique elements of a list.", RetType: "list"},
	"filter":   {Signature: "filter(list, fn) -> list", Doc: "Filters a list by a predicate function.", RetType: "list"},
	"map_list": {Signature: "map_list(list, fn) -> list", Doc: "Transforms each element of a list.", RetType: "list"},
	"flat":     {Signature: "flat(list) -> list", Doc: "Flattens nested lists into a single list.", RetType: "list"},
	"concat":   {Signature: "concat(list1, list2) -> list", Doc: "Concatenates two lists.", RetType: "list"},
	"slice":    {Signature: "slice(list, start, end) -> list", Doc: "Returns a slice of a list.", RetType: "list"},
	"range":    {Signature: "range(start, end) -> list", Doc: "Creates a list of numbers from start to end.", RetType: "list"},
	"some":     {Signature: "some(list, fn) -> boolean", Doc: "Returns true if any element satisfies the predicate.", RetType: "boolean"},
	"every":    {Signature: "every(list, fn) -> boolean", Doc: "Returns true if all elements satisfy the predicate.", RetType: "boolean"},

	// Object functions
	"keys":        {Signature: "keys(object) -> list", Doc: "Returns the keys of an object.", RetType: "list"},
	"values":      {Signature: "values(object) -> list", Doc: "Returns the values of an object.", RetType: "list"},
	"entries":     {Signature: "entries(object) -> list", Doc: "Returns key-value pairs as a list.", RetType: "list"},
	"fromEntries": {Signature: "fromEntries(list) -> object", Doc: "Creates an object from key-value pairs.", RetType: "object"},
	"hasKey":      {Signature: "hasKey(object, key) -> boolean", Doc: "Checks if an object has the given key.", RetType: "boolean"},
	"merge":       {Signature: "merge(obj1, obj2) -> object", Doc: "Merges two objects.", RetType: "object"},
	"pick":        {Signature: "pick(object, keys) -> object", Doc: "Returns an object with only the specified keys.", RetType: "object"},
	"omit":        {Signature: "omit(object, keys) -> object", Doc: "Returns an object without the specified keys.", RetType: "object"},
	"deepCopy":    {Signature: "deepCopy(value) -> object", Doc: "Creates a deep copy of a value.", RetType: "object"},

	// Type checking functions
	"typeOf":     {Signature: "typeOf(value) -> text", Doc: "Returns the type of a value as text.", RetType: "text"},
	"isText":     {Signature: "isText(value) -> boolean", Doc: "Returns true if the value is text.", RetType: "boolean"},
	"isNumber":   {Signature: "isNumber(value) -> boolean", Doc: "Returns true if the value is a number.", RetType: "boolean"},
	"isList":     {Signature: "isList(value) -> boolean", Doc: "Returns true if the value is a list.", RetType: "boolean"},
	"isObject":   {Signature: "isObject(value) -> boolean", Doc: "Returns true if the value is an object.", RetType: "boolean"},
	"isNothing":  {Signature: "isNothing(value) -> boolean", Doc: "Returns true if the value is nothing.", RetType: "boolean"},
	"isFunction": {Signature: "isFunction(value) -> boolean", Doc: "Returns true if the value is a function.", RetType: "boolean"},

	// JSON functions
	"parseJSON": {Signature: "parseJSON(text) -> object", Doc: "Parses a JSON string into an object.", RetType: "object"},

	// Date/time functions
	"now":        {Signature: "now() -> text", Doc: "Returns the current date and time as text.", RetType: "text"},
	"today":      {Signature: "today() -> text", Doc: "Returns today's date as text.", RetType: "text"},
	"timestamp":  {Signature: "timestamp() -> number", Doc: "Returns the current Unix timestamp in milliseconds.", RetType: "number"},
	"formatDate":    {Signature: "formatDate(date, format) -> text", Doc: "Formats a date string using local time. Use formatDateUTC() for server-safe UTC output.", RetType: "text"},
	"formatDateUTC": {Signature: "formatDateUTC(date, format) -> text", Doc: "Formats a date string using UTC. Use this on servers to avoid timezone issues.", RetType: "text"},
	"diffDays":   {Signature: "diffDays(date1, date2) -> number", Doc: "Returns the difference in days between two dates.", RetType: "number"},

	// Utility functions
	"uuid":     {Signature: "uuid() -> text", Doc: "Generates a random UUID.", RetType: "text"},
	"hash":     {Signature: "hash(text) -> text", Doc: "Returns a hash of the text.", RetType: "text"},
	"exists":   {Signature: "exists(value) -> boolean", Doc: "Returns true if value is not nothing.", RetType: "boolean"},
	"platform": {Signature: "platform() -> text", Doc: "Returns the current platform name.", RetType: "text"},

	// Text functions (additional)
	"padStart":       {Signature: "padStart(text, length, char) -> text", Doc: "Pads text at the start to reach the given length.", RetType: "text"},
	"padEnd":         {Signature: "padEnd(text, length, char) -> text", Doc: "Pads text at the end to reach the given length.", RetType: "text"},
	"repeat":         {Signature: "repeat(text, count) -> text", Doc: "Repeats text a given number of times.", RetType: "text"},
	"contains":       {Signature: "contains(collection, item) -> boolean", Doc: "Checks if a string or list contains the given item.", RetType: "boolean"},
	"replacePattern": {Signature: "replacePattern(text, pattern, replacement, flags?) -> text", Doc: "Replaces regex matches in text. Default flag is 'g'.", RetType: "text"},
	"toJSON":         {Signature: "toJSON(value) -> text", Doc: "Converts a value to a JSON string.", RetType: "text"},

	// List functions (additional)
	"pop":       {Signature: "pop(list) -> any", Doc: "Removes and returns the last item from a list.", RetType: "any"},
	"find":      {Signature: "find(list, fn) -> any", Doc: "Returns the first element that satisfies the predicate.", RetType: "any"},
	"reduce":    {Signature: "reduce(list, fn, initial) -> any", Doc: "Reduces a list to a single value using a function.", RetType: "any"},
	"zip":       {Signature: "zip(list1, list2) -> list", Doc: "Combines two lists into a list of pairs.", RetType: "list"},
	"countWhere": {Signature: "countWhere(list, fn) -> number", Doc: "Counts elements that satisfy the predicate.", RetType: "number"},
	"groupBy":   {Signature: "groupBy(list, fn) -> object", Doc: "Groups list elements by a key function.", RetType: "object"},
	"sum":       {Signature: "sum(list) -> number", Doc: "Returns the sum of all numbers in a list.", RetType: "number"},
	"smallest":  {Signature: "smallest(list) -> number", Doc: "Returns the smallest value in a list.", RetType: "number"},
	"largest":   {Signature: "largest(list) -> number", Doc: "Returns the largest value in a list.", RetType: "number"},

	// Date/time functions (additional)
	"addDays": {Signature: "addDays(date, days) -> text", Doc: "Adds days to a date and returns the result.", RetType: "text"},

	// Async functions
	"wait":     {Signature: "wait(ms) -> Promise", Doc: "Pauses execution for the given milliseconds. Use with `await`.", RetType: "Promise"},
	"delay":    {Signature: "delay(ms) -> Promise", Doc: "Returns a promise that resolves after the given milliseconds.", RetType: "Promise"},
	"parallel": {Signature: "parallel(fns) -> list", Doc: "Runs multiple async functions in parallel and returns all results.", RetType: "list"},
	"race":     {Signature: "race(fns) -> any", Doc: "Runs multiple async functions and returns the first to finish.", RetType: "any"},

	// Network functions
	"fetchURL":  {Signature: "fetchURL(url) -> text", Doc: "Fetches a URL and returns the response as text.", RetType: "text"},
	"fetchJSON": {Signature: "fetchJSON(url) -> object", Doc: "Fetches a URL and returns the response as parsed JSON.", RetType: "object"},
	"postJSON":    {Signature: "postJSON(url, data) -> object", Doc: "Sends a POST request with JSON data and returns the response.", RetType: "object"},
	"putJSON":     {Signature: "putJSON(url, data) -> object", Doc: "Sends a PUT request with JSON data and returns the response.", RetType: "object"},
	"patchJSON":   {Signature: "patchJSON(url, data) -> object", Doc: "Sends a PATCH request with JSON data and returns the response.", RetType: "object"},
	"deleteJSON":  {Signature: "deleteJSON(url, body?) -> object|text", Doc: "Sends a DELETE request. Optional JSON body. Returns parsed response.", RetType: "object"},

	// File/IO functions
	"read":       {Signature: "read(prompt) -> text", Doc: "Reads input from the user.", RetType: "text"},
	"write":      {Signature: "write(path, content)", Doc: "Writes content to a file, creating or overwriting it.", RetType: ""},
	"append_file": {Signature: "append_file(path, content)", Doc: "Appends content to a file.", RetType: ""},
	"readJSON":   {Signature: "readJSON(path) -> object", Doc: "Reads and parses a JSON file.", RetType: "object"},
	"writeJSON":  {Signature: "writeJSON(path, data)", Doc: "Writes data to a file as formatted JSON.", RetType: ""},
	"readLines":  {Signature: "readLines(path) -> list", Doc: "Reads a file and returns a list of lines.", RetType: "list"},
	"writeLines": {Signature: "writeLines(path, lines)", Doc: "Writes a list of lines to a file.", RetType: ""},
	"fileExists": {Signature: "fileExists(path) -> boolean", Doc: "Checks if a file exists.", RetType: "boolean"},
	"fileInfo":   {Signature: "fileInfo(path) -> object", Doc: "Returns information about a file (size, modified date, etc.).", RetType: "object"},
	"deleteFile": {Signature: "deleteFile(path)", Doc: "Deletes a file.", RetType: ""},
	"copyFile":   {Signature: "copyFile(source, dest)", Doc: "Copies a file from source to destination.", RetType: ""},
	"moveFile":   {Signature: "moveFile(source, dest)", Doc: "Moves or renames a file.", RetType: ""},
	"listFiles":  {Signature: "listFiles(dir) -> list", Doc: "Lists files in a directory.", RetType: "list"},
	"listFilesDeep": {Signature: "listFilesDeep(dir) -> list", Doc: "Lists files recursively in a directory.", RetType: "list"},
	"makeDir":    {Signature: "makeDir(path)", Doc: "Creates a directory (and parent directories).", RetType: ""},
	"watchFiles": {Signature: "watchFiles(path, callback)", Doc: "Watches a file or directory for changes.", RetType: ""},

	// Path functions
	"currentDir":    {Signature: "currentDir() -> text", Doc: "Returns the current working directory.", RetType: "text"},
	"homePath":      {Signature: "homePath() -> text", Doc: "Returns the user's home directory path.", RetType: "text"},
	"joinPath":      {Signature: "joinPath(...parts) -> text", Doc: "Joins path segments into a single path.", RetType: "text"},
	"fileName":      {Signature: "fileName(path) -> text", Doc: "Returns the file name from a path.", RetType: "text"},
	"fileExtension": {Signature: "fileExtension(path) -> text", Doc: "Returns the file extension from a path.", RetType: "text"},
	"parentDir":     {Signature: "parentDir(path) -> text", Doc: "Returns the parent directory of a path.", RetType: "text"},

	// System functions
	"env":        {Signature: "env(name, fallback?) -> text", Doc: "Returns an environment variable value. Falls back to the second argument or empty string.", RetType: "text"},
	"setEnv":     {Signature: "setEnv(name, value)", Doc: "Sets an environment variable.", RetType: ""},
	"args":       {Signature: "args() -> list", Doc: "Returns command-line arguments.", RetType: "list"},
	"memory":     {Signature: "memory() -> object", Doc: "Returns memory usage information.", RetType: "object"},
	"cpuCount":   {Signature: "cpuCount() -> number", Doc: "Returns the number of CPU cores.", RetType: "number"},
	"run":        {Signature: "run(command) -> text", Doc: "Runs a shell command and returns the output.", RetType: "text"},
	"runAsync":   {Signature: "runAsync(command) -> Promise", Doc: "Runs a shell command asynchronously.", RetType: "Promise"},
	"exit":       {Signature: "exit(code)", Doc: "Exits the program with the given status code.", RetType: ""},
	"prompt":     {Signature: "prompt(message) -> text", Doc: "Displays a prompt and reads user input.", RetType: "text"},

	// Server functions
	"createServer":  {Signature: "createServer(options) -> object", Doc: "Creates an HTTP server.", RetType: "object"},
	"serveStatic":   {Signature: "serveStatic(dir) -> function", Doc: "Creates middleware to serve static files.", RetType: "function"},
	"template":      {Signature: "template(text, data) -> text", Doc: "Renders a template string with data.", RetType: "text"},

	// Database functions
	"openDB":   {Signature: "openDB(path) -> object", Doc: "Opens a SQLite database connection.", RetType: "object"},
	"query":    {Signature: "query(db, sql, params) -> list", Doc: "Executes a SQL query and returns results.", RetType: "list"},
	"execute":  {Signature: "execute(db, sql, params)", Doc: "Executes a SQL statement (INSERT, UPDATE, DELETE).", RetType: ""},
	"closeDB":  {Signature: "closeDB(db)", Doc: "Closes a database connection.", RetType: ""},

	// Browser DOM functions
	"select":        {Signature: "select(selector) -> element", Doc: "Selects a DOM element by CSS selector.", RetType: "element"},
	"selectAll":     {Signature: "selectAll(selector) -> list", Doc: "Selects all matching DOM elements.", RetType: "list"},
	"setText":       {Signature: "setText(element, text)", Doc: "Sets the text content of an element.", RetType: ""},
	"getText":       {Signature: "getText(element) -> text", Doc: "Gets the text content of an element.", RetType: "text"},
	"setHTML":       {Signature: "setHTML(element, html)", Doc: "Sets the inner HTML of an element. **Warning:** Never pass user input directly — use sanitizeHTML() first to prevent XSS.", RetType: ""},
	"sanitizeHTML":  {Signature: "sanitizeHTML(text) -> text", Doc: "Escapes HTML special characters to prevent XSS. Use this before passing user input to setHTML().", RetType: "text"},
	"getHTML":       {Signature: "getHTML(element) -> text", Doc: "Gets the inner HTML of an element.", RetType: "text"},
	"setValue":      {Signature: "setValue(element, value)", Doc: "Sets the value of a form element.", RetType: ""},
	"getValue":      {Signature: "getValue(element) -> text", Doc: "Gets the value of a form element.", RetType: "text"},
	"setAttribute":  {Signature: "setAttribute(element, attr, value)", Doc: "Sets an attribute on an element.", RetType: ""},
	"getAttribute":  {Signature: "getAttribute(element, attr) -> text", Doc: "Gets an attribute from an element.", RetType: "text"},
	"addClass":      {Signature: "addClass(element, ...classes)", Doc: "Adds CSS classes to an element.", RetType: ""},
	"removeClass":   {Signature: "removeClass(element, ...classes)", Doc: "Removes CSS classes from an element.", RetType: ""},
	"toggleClass":   {Signature: "toggleClass(element, class)", Doc: "Toggles a CSS class on an element.", RetType: ""},
	"hasClass":      {Signature: "hasClass(element, class) -> boolean", Doc: "Checks if an element has a CSS class.", RetType: "boolean"},
	"setStyle":      {Signature: "setStyle(element, prop, value)", Doc: "Sets a CSS style property on an element.", RetType: ""},
	"getStyle":      {Signature: "getStyle(element, prop) -> text", Doc: "Gets a computed CSS style property.", RetType: "text"},
	"hide":          {Signature: "hide(element)", Doc: "Hides an element (display: none).", RetType: ""},
	"show":          {Signature: "show(element)", Doc: "Shows a hidden element.", RetType: ""},
	"onClick":       {Signature: "onClick(element, fn)", Doc: "Adds a click event listener.", RetType: ""},
	"onInput":       {Signature: "onInput(element, fn)", Doc: "Adds an input event listener.", RetType: ""},
	"onChange":       {Signature: "onChange(element, fn)", Doc: "Adds a change event listener.", RetType: ""},
	"onSubmit":      {Signature: "onSubmit(element, fn)", Doc: "Adds a submit event listener (prevents default).", RetType: ""},
	"onKeyPress":    {Signature: "onKeyPress(element, fn)", Doc: "Adds a keydown event listener.", RetType: ""},
	"onLoad":        {Signature: "onLoad(fn)", Doc: "Runs a function when the page loads.", RetType: ""},
	"onScroll":      {Signature: "onScroll(fn)", Doc: "Adds a scroll event listener.", RetType: ""},
	"createElement": {Signature: "createElement(tag, text) -> element", Doc: "Creates a new DOM element.", RetType: "element"},
	"append":        {Signature: "append(parent, child) -> element", Doc: "Appends a child element to a parent.", RetType: "element"},
	"prepend":       {Signature: "prepend(parent, child) -> element", Doc: "Prepends a child element to a parent.", RetType: "element"},
	"removeElement": {Signature: "removeElement(element)", Doc: "Removes an element from the DOM.", RetType: ""},
	"cloneElement":  {Signature: "cloneElement(element) -> element", Doc: "Creates a deep clone of an element.", RetType: "element"},
	"goTo":          {Signature: "goTo(url)", Doc: "Navigates to a URL.", RetType: ""},
	"reload":        {Signature: "reload()", Doc: "Reloads the current page.", RetType: ""},
	"currentURL":    {Signature: "currentURL() -> text", Doc: "Returns the current page URL.", RetType: "text"},
	"getParam":      {Signature: "getParam(name) -> text", Doc: "Gets a URL query parameter by name.", RetType: "text"},
	"save":          {Signature: "save(key, value)", Doc: "Saves a value to localStorage.", RetType: ""},
	"load":          {Signature: "load(key) -> any", Doc: "Loads a value from localStorage.", RetType: "any"},
	"removeData":    {Signature: "removeData(key)", Doc: "Removes a value from localStorage.", RetType: ""},
	"clearData":     {Signature: "clearData()", Doc: "Clears all localStorage data.", RetType: ""},
	"after":         {Signature: "after(ms, fn) -> number", Doc: "Runs a function after a delay. Returns a timer ID.", RetType: "number"},
	"stopTimer":     {Signature: "stopTimer(id)", Doc: "Stops a timer created by after() or every().", RetType: ""},
	"say":           {Signature: "say(value)", Doc: "Prints a value to the console.", RetType: ""},

	// CLI functions
	"arg":      {Signature: "arg(index) -> text", Doc: "Returns a command-line argument by index.", RetType: "text"},
	"flag":     {Signature: "flag(name) -> text", Doc: "Returns the value of a command-line flag.", RetType: "text"},
	"hasFlag":  {Signature: "hasFlag(name) -> boolean", Doc: "Checks if a command-line flag is present.", RetType: "boolean"},
	"exitWith": {Signature: "exitWith(code)", Doc: "Exits the program with the given status code.", RetType: ""},

	// AI / Vectors
	"embed":            {Signature: "embed(text, provider) -> list", Doc: "Generates a vector embedding for text.", RetType: "list"},
	"cosineSimilarity": {Signature: "cosineSimilarity(vec1, vec2) -> number", Doc: "Computes cosine similarity between two vectors.", RetType: "number"},
	"createVectorStore": {Signature: "createVectorStore() -> object", Doc: "Creates an in-memory vector store for similarity search.", RetType: "object"},
	"createAgent":      {Signature: "createAgent(config) -> object", Doc: "Creates an AI agent with tools and instructions.", RetType: "object"},

	// Crypto functions
	"hmac":         {Signature: "hmac(key, data, algorithm) -> text", Doc: "Computes an HMAC digest.", RetType: "text"},
	"encrypt":      {Signature: "encrypt(data, key) -> text", Doc: "Encrypts data with a key using AES.", RetType: "text"},
	"decrypt":      {Signature: "decrypt(data, key) -> text", Doc: "Decrypts AES-encrypted data with a key.", RetType: "text"},
	"randomBytes":  {Signature: "randomBytes(length) -> text", Doc: "Generates cryptographically secure random bytes.", RetType: "text"},
	"bcryptHash":   {Signature: "bcryptHash(password) -> text", Doc: "Hashes a password using bcrypt.", RetType: "text"},
	"bcryptVerify": {Signature: "bcryptVerify(password, hash) -> boolean", Doc: "Verifies a password against a bcrypt hash.", RetType: "boolean"},

	// Validation functions
	"required":      {Signature: "required(field) -> rule", Doc: "Creates a validation rule that requires a field.", RetType: "rule"},
	"minLength":     {Signature: "minLength(n) -> rule", Doc: "Creates a validation rule for minimum text length.", RetType: "rule"},
	"maxLength":     {Signature: "maxLength(n) -> rule", Doc: "Creates a validation rule for maximum text length.", RetType: "rule"},
	"min":           {Signature: "min(n) -> rule", Doc: "Creates a validation rule for minimum number value.", RetType: "rule"},
	"max":           {Signature: "max(n) -> rule", Doc: "Creates a validation rule for maximum number value.", RetType: "rule"},
	"email":         {Signature: "email() -> rule", Doc: "Creates a validation rule for email format.", RetType: "rule"},
	"url":           {Signature: "url() -> rule", Doc: "Creates a validation rule for URL format.", RetType: "rule"},
	"validateData":  {Signature: "validateData(data, rules) -> object", Doc: "Validates data against a set of rules. Returns {valid, errors}.", RetType: "object"},

	// Document processing
	"chunk":           {Signature: "chunk(text, size) -> list", Doc: "Splits text into chunks of the given size.", RetType: "list"},
	"extract":         {Signature: "extract(text, pattern) -> list", Doc: "Extracts matches from text using a pattern.", RetType: "list"},
	"splitSentences":  {Signature: "splitSentences(text) -> list", Doc: "Splits text into sentences.", RetType: "list"},
	"splitParagraphs": {Signature: "splitParagraphs(text) -> list", Doc: "Splits text into paragraphs.", RetType: "list"},

	// HTML template helpers
	"escapeHTML": {Signature: "escapeHTML(text) -> text", Doc: "Escapes HTML special characters.", RetType: "text"},
	"page":       {Signature: "page(title, ...content) -> text", Doc: "Creates a full HTML page.", RetType: "text"},
	"layout":     {Signature: "layout(name, ...content) -> text", Doc: "Creates a named layout template.", RetType: "text"},
}
