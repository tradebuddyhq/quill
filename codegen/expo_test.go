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

func TestExpoUseContext(t *testing.T) {
	src := `context ThemeContext is "light"

component ThemedScreen:
  use context ThemeContext as theme

  to render:
    view:
      text: "Theme"
`
	output, err := compileExpo(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "React.createContext(") {
		t.Error("expected React.createContext call")
	}
	if !strings.Contains(output, "useContext(ThemeContext)") {
		t.Error("expected useContext(ThemeContext)")
	}
	if !strings.Contains(output, "const theme = useContext") {
		t.Error("expected const theme = useContext")
	}
	if !strings.Contains(output, "useContext") {
		t.Error("expected useContext in React imports")
	}
}

func TestExpoUseMemo(t *testing.T) {
	src := `component ListScreen:
  state items are []

  memo total is items.length when [items]

  to render:
    view:
      text: "Total"
`
	output, err := compileExpo(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "useMemo(") {
		t.Error("expected useMemo call")
	}
	if !strings.Contains(output, "[items]") {
		t.Error("expected dependency array [items]")
	}
	if !strings.Contains(output, ", useMemo") {
		t.Error("expected useMemo in React imports")
	}
}

func TestExpoUseCallback(t *testing.T) {
	src := `component CallbackScreen:
  state items are []

  callback handlePress is with item:
    say item
  when [items]

  to render:
    view:
      text: "Callback"
`
	output, err := compileExpo(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "useCallback(") {
		t.Error("expected useCallback call")
	}
	if !strings.Contains(output, "[items]") {
		t.Error("expected dependency array [items]")
	}
	if !strings.Contains(output, ", useCallback") {
		t.Error("expected useCallback in React imports")
	}
}

func TestExpoAlertMapping(t *testing.T) {
	src := `component AlertScreen:
  to showAlert:
    alert("Error", "Something went wrong")

  to render:
    view:
      text: "Alert"
`
	output, err := compileExpo(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Alert.alert(") {
		t.Error("expected Alert.alert() call")
	}
	if !strings.Contains(output, "Alert") {
		t.Error("expected Alert in react-native imports")
	}
}

func TestExpoAsyncStorage(t *testing.T) {
	src := `component StorageScreen:
  to saveData:
    store("key", "value")

  to render:
    view:
      text: "Storage"
`
	output, err := compileExpo(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "AsyncStorage.setItem(") {
		t.Error("expected AsyncStorage.setItem call")
	}
	if !strings.Contains(output, "@react-native-async-storage/async-storage") {
		t.Error("expected AsyncStorage import")
	}
}

func TestExpoESMImports(t *testing.T) {
	src := `use "expo-location" as Location

component LocationScreen:
  to render:
    view:
      text: "Location"
`
	output, err := compileExpo(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "import * as Location from 'expo-location'") {
		t.Error("expected ESM import for expo-location")
	}
}

func TestExpoFlatListRendering(t *testing.T) {
	src := `component ListScreen:
  state items are []

  to render:
    view:
      flatlist data items key id:
        view:
          text: "Item"
`
	output, err := compileExpo(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "<FlatList") {
		t.Error("expected <FlatList")
	}
	if !strings.Contains(output, "FlatList") {
		t.Error("expected FlatList in imports")
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

func TestExpoOtherwiseRender(t *testing.T) {
	src := `component TestScreen:
  state loading is yes

  to render:
    view:
      if loading:
        text: "Loading..."
      otherwise:
        text: "Done!"
`
	output, err := compileExpo(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "?") || !strings.Contains(output, ":") {
		t.Error("expected ternary expression for if/otherwise")
	}
	if !strings.Contains(output, "Loading...") {
		t.Error("expected Loading... text")
	}
	if !strings.Contains(output, "Done!") {
		t.Error("expected Done! text")
	}
}

func TestExpoProviderElement(t *testing.T) {
	src := `context ThemeContext is "light"

component App:
  state theme is "dark"

  to render:
    provide ThemeContext with theme:
      view:
        text: "Hello"
`
	output, err := compileExpo(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "ThemeContext.Provider") {
		t.Error("expected ThemeContext.Provider")
	}
	if !strings.Contains(output, "value={theme}") {
		t.Error("expected value={theme} prop")
	}
}

func TestExpoChildrenRendering(t *testing.T) {
	src := `component Card with children:
  to render:
    view style card:
      children
`
	output, err := compileExpo(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "{ children }") {
		t.Error("expected children in props")
	}
	if !strings.Contains(output, "{children}") {
		t.Error("expected {children} in render")
	}
}

func TestExpoMultipleStyles(t *testing.T) {
	src := `component TestScreen:
  style native:
    container:
      flex is 1
    highlighted:
      background color is "#ff0"
  to render:
    view style [container, highlighted]:
      text: "Multi"
`
	output, err := compileExpo(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "styles.container") || !strings.Contains(output, "styles.highlighted") {
		t.Error("expected multiple style references")
	}
}

func TestExpoAlertWithButtons(t *testing.T) {
	src := `component TestScreen:
  to confirmDelete:
    alert("Delete?", "Are you sure?", [{ text: "Cancel" }, { text: "OK" }])

  to render:
    view:
      text: "Test"
`
	output, err := compileExpo(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Alert.alert(") {
		t.Error("expected Alert.alert")
	}
	if !strings.Contains(output, "Cancel") {
		t.Error("expected Cancel button text")
	}
}

func TestExpoTernaryStyle(t *testing.T) {
	src := `component TestScreen:
  state score is 80

  to render:
    view:
      text: "Score"

  style native:
    indicator:
      color is score greater than 70 ? "#22c55e" : "#ef4444"
      font size is 16
`
	output, err := compileExpo(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "score") {
		t.Error("expected score in style value")
	}
	if !strings.Contains(output, ">") {
		t.Error("expected > comparison operator in style value")
	}
}
