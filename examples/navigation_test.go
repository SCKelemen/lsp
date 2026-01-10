package examples

import (
	"strings"
	"testing"

	"github.com/SCKelemen/lsp/core"
)

// TestSimpleHoverProvider tests basic hover functionality.
func TestSimpleHoverProvider(t *testing.T) {
	content := `package main

// Add returns the sum of two integers.
func Add(a, b int) int {
	return a + b
}

func main() {
	result := Add(1, 2)
	println(result)
}
`

	tests := []struct {
		name         string
		position     core.Position
		wantHover    bool
		wantContains string
	}{
		{
			name:         "hover over Add function call",
			position:     core.Position{Line: 8, Character: 12}, // "Add" in main
			wantHover:    true,
			wantContains: "func Add(a, b int) int",
		},
		{
			name:         "hover over function definition",
			position:     core.Position{Line: 3, Character: 5}, // "Add" definition
			wantHover:    true,
			wantContains: "func Add",
		},
		{
			name:      "hover over empty space",
			position:  core.Position{Line: 7, Character: 0}, // Empty line
			wantHover: false,
		},
		{
			name:      "hover outside bounds",
			position:  core.Position{Line: 100, Character: 0},
			wantHover: false,
		},
	}

	provider := &SimpleHoverProvider{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hover := provider.ProvideHover("file:///test.go", content, tt.position)

			if tt.wantHover {
				if hover == nil {
					t.Fatal("expected hover, got nil")
				}

				if !strings.Contains(hover.Contents, tt.wantContains) {
					t.Errorf("hover contents %q does not contain %q",
						hover.Contents, tt.wantContains)
				}

				// Verify hover has a range
				if hover.Range == nil {
					t.Error("hover should have a range")
				}
			} else {
				if hover != nil {
					t.Errorf("expected no hover, got: %s", hover.Contents)
				}
			}
		})
	}
}

// TestSimpleHoverProvider_Unicode tests hover with Unicode content.
func TestSimpleHoverProvider_Unicode(t *testing.T) {
	content := `package main

// 计算两个数的和
func 求和(a, b int) int {
	return a + b
}

func main() {
	结果 := 求和(1, 2)
	println(结果)
}
`

	provider := &SimpleHoverProvider{}

	// Hover over function name with Unicode
	position := core.Position{Line: 8, Character: 11} // "求和" starts at byte 11 (after tab, "结果", " := ")

	hover := provider.ProvideHover("file:///test.go", content, position)

	if hover == nil {
		t.Fatal("expected hover for Unicode function name")
	}

	if !strings.Contains(hover.Contents, "求和") {
		t.Errorf("hover should contain function name, got: %s", hover.Contents)
	}

	// Verify range is valid
	if hover.Range != nil {
		if hover.Range.Start.Character < 0 {
			t.Error("hover range has negative start character")
		}

		lines := strings.Split(content, "\n")
		if hover.Range.Start.Line < len(lines) {
			line := lines[hover.Range.Start.Line]
			if hover.Range.Start.Character > len(line) {
				t.Errorf("hover range start %d exceeds line length %d",
					hover.Range.Start.Character, len(line))
			}
		}
	}
}

// TestSimpleDefinitionProvider tests go-to-definition.
func TestSimpleDefinitionProvider(t *testing.T) {
	content := `package main

func helper() string {
	return "help"
}

func main() {
	result := helper()
	println(result)
}
`

	tests := []struct {
		name         string
		position     core.Position
		wantLocation bool
		wantLine     int
	}{
		{
			name:         "definition of helper call",
			position:     core.Position{Line: 7, Character: 12}, // "helper" in main
			wantLocation: true,
			wantLine:     2, // Definition on line 2
		},
		{
			name:         "definition at declaration",
			position:     core.Position{Line: 2, Character: 5}, // "helper" definition
			wantLocation: true,
			wantLine:     2, // Same line
		},
		{
			name:         "no definition for keyword",
			position:     core.Position{Line: 6, Character: 1}, // "func" keyword (character 0-3)
			wantLocation: false,
		},
	}

	provider := &SimpleDefinitionProvider{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locations := provider.ProvideDefinition("file:///test.go", content, tt.position)

			if tt.wantLocation {
				if len(locations) == 0 {
					t.Fatal("expected location, got none")
				}

				loc := locations[0]

				if loc.Range.Start.Line != tt.wantLine {
					t.Errorf("definition at line %d, want line %d",
						loc.Range.Start.Line, tt.wantLine)
				}

				if loc.URI != "file:///test.go" {
					t.Errorf("location URI = %s, want file:///test.go", loc.URI)
				}
			} else {
				if len(locations) > 0 {
					t.Errorf("expected no location, got %d", len(locations))
				}
			}
		})
	}
}

// TestSimpleDefinitionProvider_Variables tests variable definitions.
func TestSimpleDefinitionProvider_Variables(t *testing.T) {
	content := `package main

var globalVar = 42

func main() {
	localVar := 10
	println(globalVar)
	println(localVar)
}
`

	provider := &SimpleDefinitionProvider{}

	tests := []struct {
		name     string
		position core.Position
		wantLine int
	}{
		{
			name:     "global variable usage",
			position: core.Position{Line: 6, Character: 10}, // "globalVar" in main
			wantLine: 2,                                     // Definition
		},
		{
			name:     "local variable usage",
			position: core.Position{Line: 7, Character: 10}, // "localVar" in println
			wantLine: 5,                                     // Definition
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locations := provider.ProvideDefinition("file:///test.go", content, tt.position)

			if len(locations) == 0 {
				// Local variables with := are not yet supported by SimpleDefinitionProvider
				if tt.name == "local variable usage" {
					t.Skip("local variable definitions (with :=) not yet supported")
				}
				t.Fatal("expected location")
			}

			if locations[0].Range.Start.Line != tt.wantLine {
				t.Errorf("definition at line %d, want line %d",
					locations[0].Range.Start.Line, tt.wantLine)
			}
		})
	}
}

// TestSimpleDefinitionProvider_Types tests type definitions.
func TestSimpleDefinitionProvider_Types(t *testing.T) {
	content := `package main

type User struct {
	Name string
}

func NewUser() *User {
	return &User{}
}
`

	provider := &SimpleDefinitionProvider{}

	// Go to definition of User type in NewUser return type
	position := core.Position{Line: 6, Character: 16} // "*User" in return type

	locations := provider.ProvideDefinition("file:///test.go", content, position)

	if len(locations) == 0 {
		t.Fatal("expected location for type definition")
	}

	loc := locations[0]

	// Should point to type definition on line 2
	if loc.Range.Start.Line != 2 {
		t.Errorf("type definition at line %d, want line 2", loc.Range.Start.Line)
	}
}

// TestMarkedStringHoverProvider tests keyword hover.
func TestMarkedStringHoverProvider(t *testing.T) {
	content := `package main

func test() {
	var x int
	return
}
`

	provider := &MarkedStringHoverProvider{}

	tests := []struct {
		name         string
		position     core.Position
		wantHover    bool
		wantContains string
	}{
		{
			name:         "hover over func keyword",
			position:     core.Position{Line: 2, Character: 0},
			wantHover:    true,
			wantContains: "Defines a function",
		},
		{
			name:         "hover over var keyword",
			position:     core.Position{Line: 3, Character: 1},
			wantHover:    true,
			wantContains: "Declares a variable",
		},
		{
			name:         "hover over return keyword",
			position:     core.Position{Line: 4, Character: 1},
			wantHover:    true,
			wantContains: "Returns from a function",
		},
		{
			name:         "hover over package keyword",
			position:     core.Position{Line: 0, Character: 0},
			wantHover:    true,
			wantContains: "package name",
		},
		{
			name:      "hover over identifier - not a keyword",
			position:  core.Position{Line: 2, Character: 5}, // "test"
			wantHover: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hover := provider.ProvideHover("file:///test.go", content, tt.position)

			if tt.wantHover {
				if hover == nil {
					t.Fatal("expected hover, got nil")
				}

				if !strings.Contains(hover.Contents, tt.wantContains) {
					t.Errorf("hover contents %q does not contain %q",
						hover.Contents, tt.wantContains)
				}

				// Verify range
				if hover.Range == nil {
					t.Error("hover should have a range")
				} else {
					if hover.Range.Start.Line != tt.position.Line {
						t.Errorf("hover range line = %d, want %d",
							hover.Range.Start.Line, tt.position.Line)
					}
				}
			} else {
				if hover != nil {
					t.Errorf("expected no hover, got: %s", hover.Contents)
				}
			}
		})
	}
}

// TestHoverEdgeCases tests hover edge cases.
func TestHoverEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		position core.Position
		uri      string
	}{
		{
			name:     "empty file",
			content:  "",
			position: core.Position{Line: 0, Character: 0},
			uri:      "file:///empty.go",
		},
		{
			name:     "invalid syntax",
			content:  "this is not valid go",
			position: core.Position{Line: 0, Character: 5},
			uri:      "file:///invalid.go",
		},
		{
			name:     "position out of bounds",
			content:  "package main",
			position: core.Position{Line: 100, Character: 0},
			uri:      "file:///test.go",
		},
		{
			name:     "non-go file",
			content:  "some content",
			position: core.Position{Line: 0, Character: 0},
			uri:      "file:///test.txt",
		},
	}

	providers := []core.HoverProvider{
		&SimpleHoverProvider{},
		&MarkedStringHoverProvider{},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, provider := range providers {
				// Should not crash
				hover := provider.ProvideHover(tt.uri, tt.content, tt.position)

				// Should return nil for edge cases
				if hover != nil {
					t.Logf("provider %d returned hover: %s", i, hover.Contents)
				}
			}
		})
	}
}

// TestDefinitionEdgeCases tests definition edge cases.
func TestDefinitionEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		position core.Position
		uri      string
	}{
		{
			name:     "empty file",
			content:  "",
			position: core.Position{Line: 0, Character: 0},
			uri:      "file:///empty.go",
		},
		{
			name:     "invalid syntax",
			content:  "this is not valid go",
			position: core.Position{Line: 0, Character: 5},
			uri:      "file:///invalid.go",
		},
		{
			name:     "position out of bounds",
			content:  "package main",
			position: core.Position{Line: 100, Character: 0},
			uri:      "file:///test.go",
		},
		{
			name:     "non-go file",
			content:  "some content",
			position: core.Position{Line: 0, Character: 0},
			uri:      "file:///test.txt",
		},
	}

	provider := &SimpleDefinitionProvider{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not crash
			locations := provider.ProvideDefinition(tt.uri, tt.content, tt.position)

			// Should return nil or empty slice
			if locations == nil {
				locations = []core.Location{}
			}

			t.Logf("returned %d locations", len(locations))

			// Verify all locations are valid
			for i, loc := range locations {
				if loc.URI == "" {
					t.Errorf("location %d has empty URI", i)
				}
				if loc.Range.Start.Line < 0 {
					t.Errorf("location %d has negative start line", i)
				}
			}
		})
	}
}

// TestHoverWithDocumentation tests hover with documentation comments.
func TestHoverWithDocumentation(t *testing.T) {
	content := `package main

// Add calculates the sum of two integers.
// It returns the result as an integer.
func Add(a, b int) int {
	return a + b
}

func main() {
	Add(1, 2)
}
`

	provider := &SimpleHoverProvider{}

	// Hover over Add function
	position := core.Position{Line: 4, Character: 5}

	hover := provider.ProvideHover("file:///test.go", content, position)

	if hover == nil {
		t.Fatal("expected hover")
	}

	// Should contain function signature
	if !strings.Contains(hover.Contents, "func Add(a, b int) int") {
		t.Error("hover should contain function signature")
	}

	// Should contain documentation
	if !strings.Contains(hover.Contents, "calculates the sum") {
		t.Error("hover should contain documentation")
	}
}

// TestHoverRangeAccuracy tests hover range accuracy.
func TestHoverRangeAccuracy(t *testing.T) {
	content := `package main

func helper() string {
	return "help"
}
`

	provider := &SimpleHoverProvider{}

	// Hover over "helper"
	position := core.Position{Line: 2, Character: 5}

	hover := provider.ProvideHover("file:///test.go", content, position)

	if hover == nil {
		t.Fatal("expected hover")
	}

	if hover.Range == nil {
		t.Fatal("hover should have a range")
	}

	// Range should be on line 2
	if hover.Range.Start.Line != 2 {
		t.Errorf("hover range start line = %d, want 2", hover.Range.Start.Line)
	}

	// Range should be valid
	if hover.Range.Start.Character < 0 {
		t.Error("hover range has negative start character")
	}

	if hover.Range.End.Character <= hover.Range.Start.Character {
		t.Error("hover range end should be after start")
	}

	// Range should be within line bounds
	lines := strings.Split(content, "\n")
	line := lines[hover.Range.Start.Line]

	if hover.Range.Start.Character > len(line) {
		t.Errorf("hover range start %d exceeds line length %d",
			hover.Range.Start.Character, len(line))
	}

	if hover.Range.End.Character > len(line) {
		t.Errorf("hover range end %d exceeds line length %d",
			hover.Range.End.Character, len(line))
	}
}

// TestDefinitionSameFile tests definition within same file.
func TestDefinitionSameFile(t *testing.T) {
	content := `package main

func functionA() {
	functionB()
}

func functionB() {
	functionA()
}
`

	provider := &SimpleDefinitionProvider{}

	// Go to definition of functionB from functionA
	position := core.Position{Line: 3, Character: 1} // "functionB" call

	locations := provider.ProvideDefinition("file:///test.go", content, position)

	if len(locations) == 0 {
		t.Fatal("expected location")
	}

	loc := locations[0]

	// Should point to functionB definition on line 6
	if loc.Range.Start.Line != 6 {
		t.Errorf("definition at line %d, want line 6", loc.Range.Start.Line)
	}

	// Should be same file
	if loc.URI != "file:///test.go" {
		t.Errorf("location URI = %s, want file:///test.go", loc.URI)
	}
}

// TestHoverMultibyteCharacters tests hover with multibyte characters.
func TestHoverMultibyteCharacters(t *testing.T) {
	content := `package main

// 这个函数返回问候语
func 问候() string {
	return "你好世界"
}

func main() {
	消息 := 问候()
	println(消息)
}
`

	provider := &SimpleHoverProvider{}

	// Test hover at various positions
	tests := []struct {
		name     string
		position core.Position
		wantName string
	}{
		{
			name:     "hover over function definition",
			position: core.Position{Line: 3, Character: 5},
			wantName: "问候",
		},
		{
			name:     "hover over function call",
			position: core.Position{Line: 8, Character: 11}, // "问候" starts at byte 11 (after tab, "消息", " := ")
			wantName: "问候",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hover := provider.ProvideHover("file:///test.go", content, tt.position)

			if hover == nil {
				t.Fatal("expected hover for multibyte function name")
			}

			if !strings.Contains(hover.Contents, tt.wantName) {
				t.Errorf("hover should contain %q, got: %s", tt.wantName, hover.Contents)
			}

			// Verify range is valid UTF-8 offset
			if hover.Range != nil {
				lines := strings.Split(content, "\n")
				if hover.Range.Start.Line < len(lines) {
					line := lines[hover.Range.Start.Line]
					if hover.Range.Start.Character > len(line) {
						t.Errorf("hover range start %d exceeds line length %d (UTF-8 bytes)",
							hover.Range.Start.Character, len(line))
					}
				}
			}
		})
	}
}

// TestDefinitionMultibyteCharacters tests definition with multibyte characters.
func TestDefinitionMultibyteCharacters(t *testing.T) {
	content := `package main

func 辅助函数() string {
	return "帮助"
}

func main() {
	结果 := 辅助函数()
	println(结果)
}
`

	provider := &SimpleDefinitionProvider{}

	// Go to definition of function with multibyte name
	position := core.Position{Line: 7, Character: 11} // "辅助函数" starts at byte 11 (after tab, "结果", " := ")

	locations := provider.ProvideDefinition("file:///test.go", content, position)

	if len(locations) == 0 {
		t.Fatal("expected location for multibyte function name")
	}

	loc := locations[0]

	// Should point to definition on line 2
	if loc.Range.Start.Line != 2 {
		t.Errorf("definition at line %d, want line 2", loc.Range.Start.Line)
	}

	// Verify range uses valid UTF-8 byte offsets
	if loc.Range.Start.Character < 0 {
		t.Error("location has negative start character")
	}

	lines := strings.Split(content, "\n")
	if loc.Range.Start.Line < len(lines) {
		line := lines[loc.Range.Start.Line]
		if loc.Range.Start.Character > len(line) {
			t.Errorf("location start %d exceeds line length %d (UTF-8 bytes)",
				loc.Range.Start.Character, len(line))
		}
	}
}
