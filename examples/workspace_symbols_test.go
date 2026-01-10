package examples

import (
	"strings"
	"testing"

	"github.com/SCKelemen/lsp/core"
)

// TestSimpleWorkspaceSymbolProvider tests basic workspace symbol search.
func TestSimpleWorkspaceSymbolProvider(t *testing.T) {
	provider := NewSimpleWorkspaceSymbolProvider()

	// Add test symbols
	provider.AddSymbol(core.WorkspaceSymbol{
		Name:          "Server",
		Kind:          core.SymbolKindStruct,
		ContainerName: "main",
		Location:      core.Location{URI: "file:///server.go"},
	})

	provider.AddSymbol(core.WorkspaceSymbol{
		Name:          "StartServer",
		Kind:          core.SymbolKindFunction,
		ContainerName: "main",
		Location:      core.Location{URI: "file:///main.go"},
	})

	provider.AddSymbol(core.WorkspaceSymbol{
		Name:          "Client",
		Kind:          core.SymbolKindStruct,
		ContainerName: "main",
		Location:      core.Location{URI: "file:///client.go"},
	})

	tests := []struct {
		name      string
		query     string
		wantCount int
		wantNames []string
	}{
		{
			name:      "empty query returns all",
			query:     "",
			wantCount: 3,
			wantNames: []string{"Server", "StartServer", "Client"},
		},
		{
			name:      "exact match",
			query:     "Server",
			wantCount: 2, // Server and StartServer
			wantNames: []string{"Server", "StartServer"},
		},
		{
			name:      "partial match",
			query:     "server",
			wantCount: 2, // Case-insensitive
			wantNames: []string{"Server", "StartServer"},
		},
		{
			name:      "case insensitive",
			query:     "CLIENT",
			wantCount: 1,
			wantNames: []string{"Client"},
		},
		{
			name:      "no matches",
			query:     "NonExistent",
			wantCount: 0,
			wantNames: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := provider.ProvideWorkspaceSymbols(tt.query)

			if len(results) != tt.wantCount {
				t.Errorf("got %d results, want %d", len(results), tt.wantCount)
			}

			// Check if expected names are present
			for _, wantName := range tt.wantNames {
				found := false
				for _, result := range results {
					if result.Name == wantName {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected to find symbol %q", wantName)
				}
			}
		})
	}
}

// TestSimpleWorkspaceSymbolProvider_Clear tests clearing the symbol index.
func TestSimpleWorkspaceSymbolProvider_Clear(t *testing.T) {
	provider := NewSimpleWorkspaceSymbolProvider()

	provider.AddSymbol(core.WorkspaceSymbol{
		Name: "Test",
		Kind: core.SymbolKindFunction,
	})

	results := provider.ProvideWorkspaceSymbols("")
	if len(results) != 1 {
		t.Fatalf("expected 1 symbol before clear, got %d", len(results))
	}

	provider.Clear()

	results = provider.ProvideWorkspaceSymbols("")
	if len(results) != 0 {
		t.Errorf("expected 0 symbols after clear, got %d", len(results))
	}
}

// TestGoWorkspaceSymbolProvider tests Go-specific workspace symbol provider.
func TestGoWorkspaceSymbolProvider(t *testing.T) {
	provider := NewGoWorkspaceSymbolProvider("/workspace")

	// Index some Go files
	file1 := `package main

func StartServer(port int) error {
	return nil
}

type Server struct {
	Port int
}

func (s *Server) HandleRequest() {}
`

	file2 := `package main

const MaxConnections = 100

var globalCounter int

type Config struct {
	Host string
}
`

	provider.IndexFile("file:///main.go", file1)
	provider.IndexFile("file:///config.go", file2)

	tests := []struct {
		name      string
		query     string
		wantCount int
		checkName string
		checkKind core.SymbolKind
	}{
		{
			name:      "find function",
			query:     "StartServer",
			wantCount: 1,
			checkName: "StartServer",
			checkKind: core.SymbolKindFunction,
		},
		{
			name:      "find struct",
			query:     "Server",
			wantCount: 2, // Matches both "Server" and "StartServer"
			checkName: "Server",
			checkKind: core.SymbolKindStruct,
		},
		{
			name:      "find method",
			query:     "HandleRequest",
			wantCount: 1,
			checkName: "HandleRequest",
			checkKind: core.SymbolKindMethod,
		},
		{
			name:      "find constant",
			query:     "MaxConnections",
			wantCount: 1,
			checkName: "MaxConnections",
			checkKind: core.SymbolKindConstant,
		},
		{
			name:      "find variable",
			query:     "globalCounter",
			wantCount: 1,
			checkName: "globalCounter",
			checkKind: core.SymbolKindVariable,
		},
		{
			name:      "partial match finds multiple",
			query:     "Config",
			wantCount: 1,
			checkName: "Config",
			checkKind: core.SymbolKindStruct,
		},
		{
			name:      "empty query returns all",
			query:     "",
			wantCount: 6, // StartServer, Server, HandleRequest, MaxConnections, globalCounter, Config
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := provider.ProvideWorkspaceSymbols(tt.query)

			if len(results) != tt.wantCount {
				t.Errorf("got %d results, want %d", len(results), tt.wantCount)
			}

			if tt.checkName != "" && len(results) > 0 {
				found := false
				for _, result := range results {
					if result.Name == tt.checkName {
						found = true
						if result.Kind != tt.checkKind {
							t.Errorf("symbol %q has kind %v, want %v", tt.checkName, result.Kind, tt.checkKind)
						}
						break
					}
				}
				if !found {
					t.Errorf("expected to find symbol %q", tt.checkName)
				}
			}
		})
	}
}

// TestGoWorkspaceSymbolProvider_IndexFile tests file indexing.
func TestGoWorkspaceSymbolProvider_IndexFile(t *testing.T) {
	provider := NewGoWorkspaceSymbolProvider("/workspace")

	content := `package main

func Helper() string {
	return "help"
}
`

	provider.IndexFile("file:///helper.go", content)

	results := provider.ProvideWorkspaceSymbols("Helper")

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	symbol := results[0]

	if symbol.Name != "Helper" {
		t.Errorf("got name %q, want \"Helper\"", symbol.Name)
	}

	if symbol.Kind != core.SymbolKindFunction {
		t.Errorf("got kind %v, want Function", symbol.Kind)
	}

	if symbol.Location.URI != "file:///helper.go" {
		t.Errorf("got URI %q, want \"file:///helper.go\"", symbol.Location.URI)
	}

	// Re-index with different content
	newContent := `package main

type Helper struct{}
`

	provider.IndexFile("file:///helper.go", newContent)

	results = provider.ProvideWorkspaceSymbols("Helper")

	if len(results) != 1 {
		t.Fatalf("expected 1 result after re-index, got %d", len(results))
	}

	// Should now be a struct, not a function
	if results[0].Kind != core.SymbolKindStruct {
		t.Errorf("after re-index, got kind %v, want Struct", results[0].Kind)
	}
}

// TestGoWorkspaceSymbolProvider_InvalidSyntax tests handling of invalid syntax.
func TestGoWorkspaceSymbolProvider_InvalidSyntax(t *testing.T) {
	provider := NewGoWorkspaceSymbolProvider("/workspace")

	// Index valid file first
	provider.IndexFile("file:///test.go", `package main
func Valid() {}`)

	results := provider.ProvideWorkspaceSymbols("")
	if len(results) != 1 {
		t.Fatalf("expected 1 symbol, got %d", len(results))
	}

	// Index file with invalid syntax
	provider.IndexFile("file:///test.go", `this is not valid go code`)

	// Should clear symbols for that file
	results = provider.ProvideWorkspaceSymbols("")
	if len(results) != 0 {
		t.Errorf("expected 0 symbols after invalid syntax, got %d", len(results))
	}
}

// TestGoWorkspaceSymbolProvider_MethodContainer tests method container names.
func TestGoWorkspaceSymbolProvider_MethodContainer(t *testing.T) {
	provider := NewGoWorkspaceSymbolProvider("/workspace")

	content := `package main

type Server struct{}

func (s *Server) Start() error {
	return nil
}

func (s Server) Stop() error {
	return nil
}
`

	provider.IndexFile("file:///server.go", content)

	results := provider.ProvideWorkspaceSymbols("")

	// Should have: Server (struct), Start (method), Stop (method)
	if len(results) != 3 {
		t.Fatalf("expected 3 symbols, got %d", len(results))
	}

	// Check methods have correct container
	for _, result := range results {
		if result.Kind == core.SymbolKindMethod {
			if result.ContainerName != "Server" {
				t.Errorf("method %q has container %q, want \"Server\"", result.Name, result.ContainerName)
			}
		}
	}
}

// TestGoWorkspaceSymbolProvider_InterfaceType tests interface detection.
func TestGoWorkspaceSymbolProvider_InterfaceType(t *testing.T) {
	provider := NewGoWorkspaceSymbolProvider("/workspace")

	content := `package main

type Handler interface {
	Handle() error
}
`

	provider.IndexFile("file:///handler.go", content)

	results := provider.ProvideWorkspaceSymbols("Handler")

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Kind != core.SymbolKindInterface {
		t.Errorf("got kind %v, want Interface", results[0].Kind)
	}
}

// TestGoWorkspaceSymbolProvider_NonGoFile tests that non-Go files are skipped.
func TestGoWorkspaceSymbolProvider_NonGoFile(t *testing.T) {
	provider := NewGoWorkspaceSymbolProvider("/workspace")

	content := `func NotGo() {}`

	provider.IndexFile("file:///test.txt", content)
	provider.IndexFile("file:///test.js", content)
	provider.IndexFile("file:///README.md", content)

	results := provider.ProvideWorkspaceSymbols("")

	if len(results) != 0 {
		t.Errorf("expected 0 symbols for non-Go files, got %d", len(results))
	}
}

// TestWorkspaceSymbol_EdgeCases tests edge cases.
func TestWorkspaceSymbol_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "empty file",
			content: "",
		},
		{
			name:    "only package declaration",
			content: "package main",
		},
		{
			name:    "only comments",
			content: "// This is a comment\n/* Multi-line\ncomment */",
		},
		{
			name:    "underscore names",
			content: `package main
func _() {}
type _ struct{}
const _ = 0`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewGoWorkspaceSymbolProvider("/workspace")

			// Should not crash
			provider.IndexFile("file:///test.go", tt.content)

			results := provider.ProvideWorkspaceSymbols("")

			// Underscore names should be skipped
			for _, result := range results {
				if result.Name == "_" {
					t.Error("should not include underscore names")
				}
			}

			t.Logf("returned %d symbols", len(results))
		})
	}
}

// TestWorkspaceSymbol_LocationAccuracy tests that locations are accurate.
func TestWorkspaceSymbol_LocationAccuracy(t *testing.T) {
	provider := NewGoWorkspaceSymbolProvider("/workspace")

	content := `package main

func Helper() string {
	return "help"
}

type Server struct {
	Port int
}
`

	provider.IndexFile("file:///test.go", content)

	results := provider.ProvideWorkspaceSymbols("")

	for _, result := range results {
		// Verify URI
		if result.Location.URI != "file:///test.go" {
			t.Errorf("symbol %q has URI %q, want \"file:///test.go\"", result.Name, result.Location.URI)
		}

		// Verify range is valid
		r := result.Location.Range
		if r.Start.Line < 0 || r.Start.Character < 0 {
			t.Errorf("symbol %q has negative position", result.Name)
		}

		if r.End.Line < r.Start.Line {
			t.Errorf("symbol %q has end line before start line", result.Name)
		}

		if r.End.Line == r.Start.Line && r.End.Character <= r.Start.Character {
			t.Errorf("symbol %q has end character not after start character", result.Name)
		}

		// Verify range extracts the symbol name
		startOffset := core.PositionToByteOffset(content, r.Start)
		endOffset := core.PositionToByteOffset(content, r.End)

		if startOffset >= 0 && endOffset <= len(content) && startOffset < endOffset {
			extractedName := content[startOffset:endOffset]
			if extractedName != result.Name {
				t.Errorf("range extracts %q but symbol name is %q", extractedName, result.Name)
			}
		}
	}
}

// TestWorkspaceSymbol_ContainerNames tests container name assignment.
func TestWorkspaceSymbol_ContainerNames(t *testing.T) {
	provider := NewGoWorkspaceSymbolProvider("/workspace")

	content := `package mypackage

func TopLevelFunc() {}

type MyStruct struct{}

func (m *MyStruct) MyMethod() {}

const MyConst = 42
`

	provider.IndexFile("file:///test.go", content)

	results := provider.ProvideWorkspaceSymbols("")

	expectations := map[string]string{
		"TopLevelFunc": "mypackage",
		"MyStruct":     "mypackage",
		"MyMethod":     "MyStruct",
		"MyConst":      "mypackage",
	}

	for name, expectedContainer := range expectations {
		found := false
		for _, result := range results {
			if result.Name == name {
				found = true
				if result.ContainerName != expectedContainer {
					t.Errorf("symbol %q has container %q, want %q", name, result.ContainerName, expectedContainer)
				}
				break
			}
		}
		if !found {
			t.Errorf("expected to find symbol %q", name)
		}
	}
}

// TestWorkspaceSymbol_MultipleFiles tests searching across multiple files.
func TestWorkspaceSymbol_MultipleFiles(t *testing.T) {
	provider := NewGoWorkspaceSymbolProvider("/workspace")

	file1 := `package main
func File1Func() {}`

	file2 := `package main
func File2Func() {}`

	file3 := `package util
func UtilFunc() {}`

	provider.IndexFile("file:///file1.go", file1)
	provider.IndexFile("file:///file2.go", file2)
	provider.IndexFile("file:///util.go", file3)

	// Search all
	results := provider.ProvideWorkspaceSymbols("")
	if len(results) != 3 {
		t.Errorf("expected 3 symbols total, got %d", len(results))
	}

	// Search specific
	results = provider.ProvideWorkspaceSymbols("File1")
	if len(results) != 1 {
		t.Errorf("expected 1 symbol matching 'File1', got %d", len(results))
	}

	// Verify symbol is from correct file
	if len(results) > 0 && results[0].Location.URI != "file:///file1.go" {
		t.Errorf("symbol from wrong file: %q", results[0].Location.URI)
	}
}

// TestWorkspaceSymbol_Unicode tests handling of Unicode symbol names.
func TestWorkspaceSymbol_Unicode(t *testing.T) {
	provider := NewGoWorkspaceSymbolProvider("/workspace")

	content := `package main

func 问候() string {
	return "你好"
}

type 配置 struct {
	值 string
}
`

	provider.IndexFile("file:///test.go", content)

	// Search for Unicode function name
	results := provider.ProvideWorkspaceSymbols("问候")

	if len(results) != 1 {
		t.Errorf("expected 1 result for Unicode search, got %d", len(results))
	}

	if len(results) > 0 && results[0].Name != "问候" {
		t.Errorf("got name %q, want \"问候\"", results[0].Name)
	}

	// Search for Unicode type name
	results = provider.ProvideWorkspaceSymbols("配置")

	if len(results) != 1 {
		t.Errorf("expected 1 result for Unicode type search, got %d", len(results))
	}

	// Verify ranges are valid UTF-8 byte offsets
	for _, result := range results {
		r := result.Location.Range
		lines := strings.Split(content, "\n")
		if r.Start.Line < len(lines) {
			line := lines[r.Start.Line]
			if r.Start.Character > len(line) {
				t.Errorf("symbol %q: start character %d exceeds line length %d (UTF-8 bytes)",
					result.Name, r.Start.Character, len(line))
			}
		}
	}
}
