package learn

import "testing"

func TestLessonCount(t *testing.T) {
	lessons := Lessons()
	if len(lessons) != 10 {
		t.Errorf("expected 10 lessons, got %d", len(lessons))
	}
}

func TestLessonValidation(t *testing.T) {
	lessons := Lessons()

	// Lesson 1: Variables
	if !lessons[0].Validate(`name is "Alice"`) {
		t.Error("lesson 1 should accept: name is \"Alice\"")
	}
	if lessons[0].Validate(`hello`) {
		t.Error("lesson 1 should reject: hello")
	}

	// Lesson 2: Printing
	if !lessons[1].Validate(`say "Hello, World!"`) {
		t.Error("lesson 2 should accept: say \"Hello, World!\"")
	}

	// Lesson 3: Numbers
	if !lessons[2].Validate(`total is 10 + 5`) {
		t.Error("lesson 3 should accept: total is 10 + 5")
	}

	// Lesson 4: Lists
	if !lessons[3].Validate(`colors are ["red", "blue"]`) {
		t.Error("lesson 4 should accept: colors are [\"red\", \"blue\"]")
	}

	// Lesson 5: If
	if !lessons[4].Validate(`if 10 is greater than 5:`) {
		t.Error("lesson 5 should accept: if 10 is greater than 5:")
	}

	// Lesson 6: Functions
	if !lessons[5].Validate(`to greet name:`) {
		t.Error("lesson 6 should accept: to greet name:")
	}

	// Lesson 7: Loops
	if !lessons[6].Validate(`for each item in ["a", "b", "c"]:`) {
		t.Error("lesson 7 should accept: for each item in [\"a\", \"b\", \"c\"]:")
	}

	// Lesson 8: String interpolation
	if !lessons[7].Validate(`say "I am {age} years old"`) {
		t.Error("lesson 8 should accept: say \"I am {age} years old\"")
	}

	// Lesson 9: Fetching data
	if !lessons[8].Validate(`data is await fetchJSON("https://api.example.com")`) {
		t.Error("lesson 9 should accept: data is await fetchJSON(\"https://api.example.com\")")
	}

	// Lesson 10: Putting it together
	if !lessons[9].Validate(`to sayHello name:`) {
		t.Error("lesson 10 should accept: to sayHello name:")
	}
}

func TestMatchPattern(t *testing.T) {
	if !matchPattern("hello world", `^hello`) {
		t.Error("matchPattern should match 'hello' at start")
	}
	if matchPattern("world hello", `^hello`) {
		t.Error("matchPattern should not match 'hello' not at start")
	}
}

func TestAllLessonsHaveContent(t *testing.T) {
	for i, lesson := range Lessons() {
		if lesson.Title == "" {
			t.Errorf("lesson %d has empty title", i+1)
		}
		if lesson.Explanation == "" {
			t.Errorf("lesson %d has empty explanation", i+1)
		}
		if lesson.Prompt == "" {
			t.Errorf("lesson %d has empty prompt", i+1)
		}
		if lesson.Hint == "" {
			t.Errorf("lesson %d has empty hint", i+1)
		}
		if lesson.Success == "" {
			t.Errorf("lesson %d has empty success message", i+1)
		}
		if lesson.Validate == nil {
			t.Errorf("lesson %d has nil validate function", i+1)
		}
	}
}
