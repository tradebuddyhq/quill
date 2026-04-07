package codegen

import (
	"quill/lexer"
	"quill/parser"
	"strings"
	"testing"
)

func compileExpo(input string) (string, error) {
	l := lexer.New(input)
	tokens, err := l.Tokenize()
	if err != nil {
		return "", err
	}
	p := parser.New(tokens)
	prog, err := p.Parse()
	if err != nil {
		return "", err
	}
	gen := NewExpo()
	return gen.Generate(prog), nil
}

func TestExpoBasicComponent(t *testing.T) {
	src := `component HomeScreen:
  state count is 0

  to increment:
    count is count + 1

  to render:
    view:
      text: "Hello"
`
	output, err := compileExpo(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "export default function HomeScreen") {
		t.Error("expected export default function HomeScreen")
	}
	if !strings.Contains(output, "useState(0)") {
		t.Error("expected useState(0)")
	}
	if !strings.Contains(output, "setCount(") {
		t.Error("expected setCount setter call")
	}
	if !strings.Contains(output, "<View>") {
		t.Error("expected <View> element")
	}
	if !strings.Contains(output, "<Text>") {
		t.Error("expected <Text> element")
	}
	if !strings.Contains(output, "import React") {
		t.Error("expected React import")
	}
}

func TestExpoComponentWithProps(t *testing.T) {
	src := `component ProfileScreen with navigation route:
  to render:
    view:
      text: "Profile"
`
	output, err := compileExpo(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "function ProfileScreen({ navigation, route })") {
		t.Errorf("expected destructured props, got:\n%s", output)
	}
}

func TestExpoEffect(t *testing.T) {
	src := `component DataScreen:
  state data is nothing

  use effect when [data]:
    say "changed"

  to render:
    view:
      text: "Data"
`
	output, err := compileExpo(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "useEffect") {
		t.Error("expected useEffect")
	}
	if !strings.Contains(output, "[data]") {
		t.Error("expected dependency array [data]")
	}
}

func TestExpoEffectEmptyDeps(t *testing.T) {
	src := `component MountScreen:
  use effect:
    say "mounted"

  to render:
    view:
      text: "Loaded"
`
	output, err := compileExpo(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "useEffect") {
		t.Error("expected useEffect")
	}
	if !strings.Contains(output, ", [])") {
		t.Error("expected empty dependency array")
	}
}

func TestExpoElementMapping(t *testing.T) {
	src := `component TestScreen:
  to render:
    scroll:
      view:
        text: "Hello"
        button onPress doSomething:
          text: "Tap"
`
	output, err := compileExpo(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "<ScrollView>") {
		t.Error("expected <ScrollView>")
	}
	if !strings.Contains(output, "<TouchableOpacity") {
		t.Error("expected <TouchableOpacity> for button")
	}
	if !strings.Contains(output, "onPress") {
		t.Error("expected onPress handler")
	}
}

func TestExpoNavigateStatement(t *testing.T) {
	src := `component NavScreen with navigation:
  to goHome:
    navigate to "Home"

  to goDetails:
    navigate to "Details" with { id: 42 }

  to render:
    view:
      text: "Nav"
`
	output, err := compileExpo(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, `navigation.navigate("Home")`) {
		t.Error("expected navigation.navigate for simple navigate")
	}
	if !strings.Contains(output, `navigation.navigate("Details"`) {
		t.Error("expected navigation.navigate with params")
	}
}

func TestExpoImports(t *testing.T) {
	src := `component TestScreen:
  to render:
    view:
      text: "Hello"
`
	output, err := compileExpo(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "import React") {
		t.Error("expected React import")
	}
	if !strings.Contains(output, "from 'react-native'") {
		t.Error("expected react-native import")
	}
	if !strings.Contains(output, "View") {
		t.Error("expected View in imports")
	}
	if !strings.Contains(output, "Text") {
		t.Error("expected Text in imports")
	}
}

func TestExpoModeFlag(t *testing.T) {
	gen := NewExpo()
	if !gen.expoMode {
		t.Error("expected expoMode to be true")
	}
}

func TestMapRNTag(t *testing.T) {
	tests := map[string]string{
		"view":      "View",
		"text":      "Text",
		"scroll":    "ScrollView",
		"image":     "Image",
		"input":     "TextInput",
		"button":    "TouchableOpacity",
		"touchable": "TouchableOpacity",
		"safearea":  "SafeAreaView",
		"modal":     "Modal",
	}
	for input, expected := range tests {
		result := mapRNTag(input)
		if result != expected {
			t.Errorf("mapRNTag(%q) = %q, want %q", input, result, expected)
		}
	}
}
