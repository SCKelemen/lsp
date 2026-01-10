package examples

import (
	"strings"
	"testing"

	"github.com/SCKelemen/lsp/core"
)

// TestGoSymbolProvider tests basic symbol extraction.
func TestGoSymbolProvider(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantCount int
		wantNames []string
		wantKinds []core.SymbolKind
	}{
		{
			name: "simple function",
			content: `package main

func main() {
	println("hello")
}`,
			wantCount: 1,
			wantNames: []string{"main"},
			wantKinds: []core.SymbolKind{core.SymbolKindFunction},
		},
		{
			name: "multiple functions",
			content: `package main

func helper() string {
	return "help"
}

func main() {
	println(helper())
}`,
			wantCount: 2,
			wantNames: []string{"helper", "main"},
			wantKinds: []core.SymbolKind{
				core.SymbolKindFunction,
				core.SymbolKindFunction,
			},
		},
		{
			name: "struct with methods",
			content: `package main

type User struct {
	Name string
}

func (u *User) String() string {
	return u.Name
}`,
			wantCount: 2,
			wantNames: []string{"User", "String"},
			wantKinds: []core.SymbolKind{
				core.SymbolKindStruct,
				core.SymbolKindMethod,
			},
		},
		{
			name: "interface",
			content: `package main

type Reader interface {
	Read() error
}`,
			wantCount: 1,
			wantNames: []string{"Reader"},
			wantKinds: []core.SymbolKind{core.SymbolKindInterface},
		},
		{
			name: "constants and variables",
			content: `package main

const Version = "1.0.0"

var Config = struct{}{}`,
			wantCount: 2,
			wantNames: []string{"Version", "Config"},
			wantKinds: []core.SymbolKind{
				core.SymbolKindConstant,
				core.SymbolKindVariable,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &GoSymbolProvider{}
			symbols := provider.ProvideDocumentSymbols("file:///test.go", tt.content)

			if len(symbols) != tt.wantCount {
				t.Errorf("got %d symbols, want %d", len(symbols), tt.wantCount)
				for i, s := range symbols {
					t.Logf("  symbol %d: %s (kind=%d)", i, s.Name, s.Kind)
				}
				return
			}

			for i, symbol := range symbols {
				if i >= len(tt.wantNames) {
					break
				}

				if symbol.Name != tt.wantNames[i] {
					t.Errorf("symbol %d: got name %q, want %q",
						i, symbol.Name, tt.wantNames[i])
				}

				if i < len(tt.wantKinds) && symbol.Kind != tt.wantKinds[i] {
					t.Errorf("symbol %d: got kind %d, want %d",
						i, symbol.Kind, tt.wantKinds[i])
				}
			}
		})
	}
}

// TestGoSymbolProvider_Hierarchy tests hierarchical symbols.
func TestGoSymbolProvider_Hierarchy(t *testing.T) {
	content := `package main

type User struct {
	Name  string
	Email string
}

func (u *User) String() string {
	return u.Name
}
`

	provider := &GoSymbolProvider{}
	symbols := provider.ProvideDocumentSymbols("file:///test.go", content)

	// Find User struct
	var userSymbol *core.DocumentSymbol
	for i := range symbols {
		if symbols[i].Name == "User" {
			userSymbol = &symbols[i]
			break
		}
	}

	if userSymbol == nil {
		t.Fatal("User symbol not found")
	}

	// Check struct has field children
	if len(userSymbol.Children) != 2 {
		t.Errorf("User struct has %d children, want 2", len(userSymbol.Children))
	}

	// Verify field names
	expectedFields := []string{"Name", "Email"}
	for i, child := range userSymbol.Children {
		if i >= len(expectedFields) {
			break
		}
		if child.Name != expectedFields[i] {
			t.Errorf("field %d: got %q, want %q", i, child.Name, expectedFields[i])
		}
		if child.Kind != core.SymbolKindField {
			t.Errorf("field %d: got kind %d, want %d",
				i, child.Kind, core.SymbolKindField)
		}
	}
}

// TestGoSymbolProvider_InterfaceMethods tests interface method extraction.
func TestGoSymbolProvider_InterfaceMethods(t *testing.T) {
	content := `package main

type Reader interface {
	Read(p []byte) (n int, err error)
	Close() error
}
`

	provider := &GoSymbolProvider{}
	symbols := provider.ProvideDocumentSymbols("file:///test.go", content)

	if len(symbols) != 1 {
		t.Fatalf("got %d symbols, want 1", len(symbols))
	}

	readerSymbol := symbols[0]

	if readerSymbol.Kind != core.SymbolKindInterface {
		t.Errorf("got kind %d, want Interface", readerSymbol.Kind)
	}

	// Check interface has method children
	if len(readerSymbol.Children) != 2 {
		t.Errorf("Reader interface has %d children, want 2", len(readerSymbol.Children))
		return
	}

	// Verify method names
	expectedMethods := []string{"Read", "Close"}
	for i, child := range readerSymbol.Children {
		if i >= len(expectedMethods) {
			break
		}
		if child.Name != expectedMethods[i] {
			t.Errorf("method %d: got %q, want %q", i, child.Name, expectedMethods[i])
		}
		if child.Kind != core.SymbolKindMethod {
			t.Errorf("method %d: got kind %d, want Method",
				i, child.Kind)
		}
	}
}

// TestGoSymbolProvider_FunctionSignatures tests function detail extraction.
func TestGoSymbolProvider_FunctionSignatures(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantDetails []string
	}{
		{
			name: "simple function",
			content: `package main

func add(a, b int) int {
	return a + b
}`,
			wantDetails: []string{"add(a, b int) int"},
		},
		{
			name: "no parameters",
			content: `package main

func getValue() string {
	return "value"
}`,
			wantDetails: []string{"getValue() string"},
		},
		{
			name: "multiple return values",
			content: `package main

func divide(a, b int) (int, error) {
	return a / b, nil
}`,
			wantDetails: []string{"divide(a, b int) (int, error)"},
		},
		{
			name: "method with receiver",
			content: `package main

type Calculator struct{}

func (c *Calculator) Add(a, b int) int {
	return a + b
}`,
			wantDetails: []string{"Calculator", "(*Calculator) Add(a, b int) int"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &GoSymbolProvider{}
			symbols := provider.ProvideDocumentSymbols("file:///test.go", tt.content)

			for i, symbol := range symbols {
				if i >= len(tt.wantDetails) {
					break
				}

				if !strings.Contains(symbol.Detail, tt.wantDetails[i]) && symbol.Detail != tt.wantDetails[i] {
					// For struct types, detail might just be the name
					if symbol.Kind == core.SymbolKindStruct && symbol.Name == tt.wantDetails[i] {
						continue
					}
					t.Errorf("symbol %d (%s): got detail %q, want to contain %q",
						i, symbol.Name, symbol.Detail, tt.wantDetails[i])
				}
			}
		})
	}
}

// TestGoSymbolProvider_RangeAccuracy tests position accuracy.
func TestGoSymbolProvider_RangeAccuracy(t *testing.T) {
	content := `package main

func helper() {
	println("hello")
}
`

	provider := &GoSymbolProvider{}
	symbols := provider.ProvideDocumentSymbols("file:///test.go", content)

	if len(symbols) == 0 {
		t.Fatal("expected at least one symbol")
	}

	symbol := symbols[0]

	// Function should start at line 2 (0-based)
	if symbol.Range.Start.Line != 2 {
		t.Errorf("function start line = %d, want 2", symbol.Range.Start.Line)
	}

	// Selection range should point to function name
	if symbol.SelectionRange.Start.Line != 2 {
		t.Errorf("selection start line = %d, want 2", symbol.SelectionRange.Start.Line)
	}

	// Selection range should be narrower than full range
	if symbol.SelectionRange.End.Line > symbol.Range.End.Line {
		t.Error("selection range should be within full range")
	}
}

// TestGoSymbolProvider_Unicode tests with Unicode content.
func TestGoSymbolProvider_Unicode(t *testing.T) {
	content := `package main

// 这是一个函数
func 函数名() string {
	return "你好世界"
}

type 用户 struct {
	名字 string
}
`

	provider := &GoSymbolProvider{}
	symbols := provider.ProvideDocumentSymbols("file:///test.go", content)

	if len(symbols) < 2 {
		t.Errorf("got %d symbols, want at least 2", len(symbols))
	}

	// Verify all ranges are valid
	lines := strings.Split(content, "\n")
	for i, symbol := range symbols {
		// Check range validity
		if symbol.Range.Start.Line < 0 || symbol.Range.Start.Line >= len(lines) {
			t.Errorf("symbol %d: invalid start line %d", i, symbol.Range.Start.Line)
		}
		if symbol.Range.End.Line < 0 || symbol.Range.End.Line >= len(lines) {
			t.Errorf("symbol %d: invalid end line %d", i, symbol.Range.End.Line)
		}

		// Check character offsets are valid UTF-8 byte offsets
		startLine := lines[symbol.Range.Start.Line]
		if symbol.Range.Start.Character < 0 || symbol.Range.Start.Character > len(startLine) {
			t.Errorf("symbol %d: start character %d exceeds line length %d",
				i, symbol.Range.Start.Character, len(startLine))
		}

		// Check selection range
		if symbol.SelectionRange.Start.Line < 0 || symbol.SelectionRange.Start.Line >= len(lines) {
			t.Errorf("symbol %d: invalid selection start line %d",
				i, symbol.SelectionRange.Start.Line)
		}

		selLine := lines[symbol.SelectionRange.Start.Line]
		if symbol.SelectionRange.Start.Character < 0 || symbol.SelectionRange.Start.Character > len(selLine) {
			t.Errorf("symbol %d: selection start character %d exceeds line length %d",
				i, symbol.SelectionRange.Start.Character, len(selLine))
		}
	}
}

// TestGoSymbolProvider_EdgeCases tests edge cases.
func TestGoSymbolProvider_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		content string
		uri     string
	}{
		{
			name:    "empty file",
			content: "",
			uri:     "file:///empty.go",
		},
		{
			name:    "only package declaration",
			content: "package main",
			uri:     "file:///minimal.go",
		},
		{
			name:    "invalid syntax",
			content: "this is not valid go",
			uri:     "file:///invalid.go",
		},
		{
			name:    "non-go file",
			content: "some content",
			uri:     "file:///test.txt",
		},
		{
			name:    "only comments",
			content: "// Just a comment\n// Another comment",
			uri:     "file:///comments.go",
		},
	}

	provider := &GoSymbolProvider{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not crash on edge cases
			symbols := provider.ProvideDocumentSymbols(tt.uri, tt.content)

			// Should return nil or empty slice
			if symbols == nil {
				symbols = []core.DocumentSymbol{}
			}

			t.Logf("returned %d symbols", len(symbols))

			// Verify all returned symbols are valid
			for i, symbol := range symbols {
				if symbol.Name == "" {
					t.Errorf("symbol %d has empty name", i)
				}
				if symbol.Range.Start.Line < 0 {
					t.Errorf("symbol %d has negative start line", i)
				}
				if symbol.Range.End.Line < symbol.Range.Start.Line {
					t.Errorf("symbol %d: end line before start line", i)
				}
			}
		})
	}
}

// TestGoSymbolProvider_ComplexStructure tests complex nested structures.
func TestGoSymbolProvider_ComplexStructure(t *testing.T) {
	content := `package main

const (
	Version = "1.0.0"
	Author  = "Test"
)

type Server struct {
	Host string
	Port int
}

func (s *Server) Start() error {
	return nil
}

func (s *Server) Stop() error {
	return nil
}

func NewServer(host string, port int) *Server {
	return &Server{Host: host, Port: port}
}
`

	provider := &GoSymbolProvider{}
	symbols := provider.ProvideDocumentSymbols("file:///test.go", content)

	// Should have: constants group, Server struct, 3 functions (2 methods + 1 function)
	if len(symbols) < 4 {
		t.Errorf("got %d symbols, want at least 4", len(symbols))
	}

	// Find Server struct and verify it has fields
	var serverSymbol *core.DocumentSymbol
	for i := range symbols {
		if symbols[i].Name == "Server" && symbols[i].Kind == core.SymbolKindStruct {
			serverSymbol = &symbols[i]
			break
		}
	}

	if serverSymbol == nil {
		t.Error("Server struct not found")
		return
	}

	if len(serverSymbol.Children) != 2 {
		t.Errorf("Server struct has %d fields, want 2", len(serverSymbol.Children))
	}

	// Verify methods are present
	methodCount := 0
	for _, symbol := range symbols {
		if symbol.Kind == core.SymbolKindMethod {
			methodCount++
		}
	}

	if methodCount < 2 {
		t.Errorf("got %d methods, want at least 2", methodCount)
	}
}

// TestGoSymbolProvider_SelectionRange tests selection range accuracy.
func TestGoSymbolProvider_SelectionRange(t *testing.T) {
	content := `package main

func calculateSum(a, b int) int {
	return a + b
}
`

	provider := &GoSymbolProvider{}
	symbols := provider.ProvideDocumentSymbols("file:///test.go", content)

	if len(symbols) == 0 {
		t.Fatal("expected at least one symbol")
	}

	symbol := symbols[0]

	// Selection range should be on the same line as function declaration
	if symbol.SelectionRange.Start.Line != 2 {
		t.Errorf("selection range start line = %d, want 2",
			symbol.SelectionRange.Start.Line)
	}

	// Selection range should be smaller than full range
	selectionSize := symbol.SelectionRange.End.Character - symbol.SelectionRange.Start.Character
	fullRangeLines := symbol.Range.End.Line - symbol.Range.Start.Line

	if selectionSize <= 0 {
		t.Error("selection range should have positive size")
	}

	// Selection should be on single line for function name
	if symbol.SelectionRange.Start.Line != symbol.SelectionRange.End.Line {
		t.Error("selection range should be on single line for function name")
	}

	// Full range should span multiple lines for multi-line function
	if fullRangeLines < 2 {
		t.Errorf("full range should span multiple lines, got %d", fullRangeLines)
	}
}

// TestGoSymbolProvider_EmptyStruct tests empty struct handling.
func TestGoSymbolProvider_EmptyStruct(t *testing.T) {
	content := `package main

type Empty struct{}

type EmptyInterface interface{}
`

	provider := &GoSymbolProvider{}
	symbols := provider.ProvideDocumentSymbols("file:///test.go", content)

	if len(symbols) != 2 {
		t.Errorf("got %d symbols, want 2", len(symbols))
	}

	for i, symbol := range symbols {
		// Empty structs/interfaces should have no children
		if len(symbol.Children) != 0 {
			t.Errorf("symbol %d (%s): got %d children, want 0",
				i, symbol.Name, len(symbol.Children))
		}
	}
}

// TestGoSymbolProvider_MethodReceivers tests method receiver detection.
func TestGoSymbolProvider_MethodReceivers(t *testing.T) {
	content := `package main

type Counter struct {
	count int
}

func (c Counter) Value() int {
	return c.count
}

func (c *Counter) Increment() {
	c.count++
}
`

	provider := &GoSymbolProvider{}
	symbols := provider.ProvideDocumentSymbols("file:///test.go", content)

	// Should find struct + 2 methods
	methodCount := 0
	for _, symbol := range symbols {
		if symbol.Kind == core.SymbolKindMethod {
			methodCount++
			// Method detail should include receiver type
			if !strings.Contains(symbol.Detail, "Counter") {
				t.Errorf("method %s detail should include receiver type, got: %s",
					symbol.Name, symbol.Detail)
			}
		}
	}

	if methodCount != 2 {
		t.Errorf("got %d methods, want 2", methodCount)
	}
}
