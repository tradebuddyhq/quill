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
	"matches":       {Signature: "matches(text, pattern) -> list", Doc: "Returns regex matches in text.", RetType: "list"},
	"matchesPattern": {Signature: "matchesPattern(text, pattern) -> boolean", Doc: "Tests if text matches a regex pattern.", RetType: "boolean"},
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
	"randomInt": {Signature: "randomInt(min, max) -> number", Doc: "Returns a random integer between min and max.", RetType: "number"},

	// List functions
	"push":     {Signature: "push(list, item) -> list", Doc: "Adds an item to the end of a list.", RetType: "list"},
	"sort":     {Signature: "sort(list) -> list", Doc: "Sorts a list.", RetType: "list"},
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
	"formatDate": {Signature: "formatDate(date, format) -> text", Doc: "Formats a date string.", RetType: "text"},
	"diffDays":   {Signature: "diffDays(date1, date2) -> number", Doc: "Returns the difference in days between two dates.", RetType: "number"},

	// Utility functions
	"uuid":     {Signature: "uuid() -> text", Doc: "Generates a random UUID.", RetType: "text"},
	"hash":     {Signature: "hash(text) -> text", Doc: "Returns a hash of the text.", RetType: "text"},
	"exists":   {Signature: "exists(value) -> boolean", Doc: "Returns true if value is not nothing.", RetType: "boolean"},
	"platform": {Signature: "platform() -> text", Doc: "Returns the current platform name.", RetType: "text"},

	// File/IO functions
	"read":       {Signature: "read(prompt) -> text", Doc: "Reads input from the user.", RetType: "text"},
	"readJSON":   {Signature: "readJSON(path) -> object", Doc: "Reads and parses a JSON file.", RetType: "object"},
	"fileExists": {Signature: "fileExists(path) -> boolean", Doc: "Checks if a file exists.", RetType: "boolean"},
	"fileInfo":   {Signature: "fileInfo(path) -> object", Doc: "Returns information about a file.", RetType: "object"},
	"env":        {Signature: "env(name) -> text", Doc: "Returns an environment variable value.", RetType: "text"},
	"args":       {Signature: "args() -> list", Doc: "Returns command-line arguments.", RetType: "list"},
	"memory":     {Signature: "memory() -> object", Doc: "Returns memory usage information.", RetType: "object"},
}
