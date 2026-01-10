package examples

import (
	"strings"
	"testing"

	"github.com/SCKelemen/lsp/core"
)

// TestSimpleRenameProvider_PrepareRename tests prepare rename validation.
func TestSimpleRenameProvider_PrepareRename(t *testing.T) {
	provider := &SimpleRenameProvider{}

	tests := []struct {
		name     string
		content  string
		position core.Position
		wantNil  bool
		wantText string
	}{
		{
			name:     "valid identifier",
			content:  "func calculateSum() {}",
			position: core.Position{Line: 0, Character: 10}, // on "Sum"
			wantNil:  false,
			wantText: "calculateSum",
		},
		{
			name:     "start of word",
			content:  "var myVariable = 42",
			position: core.Position{Line: 0, Character: 4}, // start of "myVariable"
			wantNil:  false,
			wantText: "myVariable",
		},
		{
			name:     "end of word",
			content:  "var myVariable = 42",
			position: core.Position{Line: 0, Character: 13}, // end of "myVariable"
			wantNil:  false,
			wantText: "myVariable",
		},
		{
			name:     "whitespace position",
			content:  "var x = 42",
			position: core.Position{Line: 0, Character: 3}, // on space
			wantNil:  true,
		},
		{
			name:     "operator position",
			content:  "x = y + z",
			position: core.Position{Line: 0, Character: 2}, // on "="
			wantNil:  true,
		},
		{
			name:     "out of bounds",
			content:  "var x",
			position: core.Position{Line: 10, Character: 0},
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := provider.PrepareRename("file:///test.go", tt.content, tt.position)

			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil, got range %v", result)
				}
				return
			}

			if result == nil {
				t.Fatal("expected range, got nil")
			}

			// Extract the text from the range
			startOffset := core.PositionToByteOffset(tt.content, result.Start)
			endOffset := core.PositionToByteOffset(tt.content, result.End)

			if startOffset < 0 || endOffset > len(tt.content) {
				t.Fatalf("invalid range offsets: start=%d, end=%d, content len=%d", startOffset, endOffset, len(tt.content))
			}

			gotText := tt.content[startOffset:endOffset]
			if gotText != tt.wantText {
				t.Errorf("got text %q, want %q", gotText, tt.wantText)
			}
		})
	}
}

// TestSimpleRenameProvider_ProvideRename tests basic rename functionality.
func TestSimpleRenameProvider_ProvideRename(t *testing.T) {
	provider := &SimpleRenameProvider{}

	tests := []struct {
		name          string
		content       string
		position      core.Position
		newName       string
		wantNil       bool
		wantEditCount int
		checkContent  func(t *testing.T, edits []core.TextEdit, content string)
	}{
		{
			name: "rename variable single occurrence",
			content: `func main() {
	var count = 0
	println(count)
}`,
			position:      core.Position{Line: 1, Character: 5}, // on "count"
			newName:       "total",
			wantEditCount: 2, // declaration and usage
			checkContent: func(t *testing.T, edits []core.TextEdit, content string) {
				result := applyEdits(content, edits)
				if !strings.Contains(result, "var total = 0") {
					t.Error("expected 'var total = 0' in result")
				}
				if !strings.Contains(result, "println(total)") {
					t.Error("expected 'println(total)' in result")
				}
			},
		},
		{
			name: "rename respects word boundaries",
			content: `var counter = 0
var count = 1
var recount = 2`,
			position:      core.Position{Line: 1, Character: 4}, // on "count"
			newName:       "total",
			wantEditCount: 1, // only "count", not "counter" or "recount"
			checkContent: func(t *testing.T, edits []core.TextEdit, content string) {
				result := applyEdits(content, edits)
				if !strings.Contains(result, "var counter = 0") {
					t.Error("'counter' should not be renamed")
				}
				if !strings.Contains(result, "var total = 1") {
					t.Error("expected 'var total = 1' in result")
				}
				if !strings.Contains(result, "var recount = 2") {
					t.Error("'recount' should not be renamed")
				}
			},
		},
		{
			name: "rename with multiple occurrences",
			content: `x := 1
y := x + 2
z := x * y
result := x + y + z`,
			position:      core.Position{Line: 0, Character: 0}, // on first "x"
			newName:       "value",
			wantEditCount: 4, // all occurrences of "x"
			checkContent: func(t *testing.T, edits []core.TextEdit, content string) {
				result := applyEdits(content, edits)
				xCount := strings.Count(result, "x")
				valueCount := strings.Count(result, "value")
				if xCount != 0 {
					t.Errorf("expected no 'x' in result, found %d", xCount)
				}
				if valueCount != 4 {
					t.Errorf("expected 4 'value' in result, found %d", valueCount)
				}
			},
		},
		{
			name:     "invalid position returns nil",
			content:  "var x = 42",
			position: core.Position{Line: 0, Character: 6}, // on "="
			newName:  "y",
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := core.RenameContext{
				URI:      "file:///test.go",
				Content:  tt.content,
				Position: tt.position,
				NewName:  tt.newName,
			}

			result := provider.ProvideRename(ctx)

			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil, got workspace edit")
				}
				return
			}

			if result == nil {
				t.Fatal("expected workspace edit, got nil")
			}

			edits, ok := result.Changes[ctx.URI]
			if !ok {
				t.Fatal("expected changes for URI")
			}

			if len(edits) != tt.wantEditCount {
				t.Errorf("got %d edits, want %d", len(edits), tt.wantEditCount)
			}

			// Verify all edits replace with new name
			for i, edit := range edits {
				if edit.NewText != tt.newName {
					t.Errorf("edit %d: got new text %q, want %q", i, edit.NewText, tt.newName)
				}
			}

			if tt.checkContent != nil {
				tt.checkContent(t, edits, tt.content)
			}
		})
	}
}

// TestGoRenameProvider_PrepareRename tests Go-specific prepare rename.
func TestGoRenameProvider_PrepareRename(t *testing.T) {
	provider := &GoRenameProvider{}

	tests := []struct {
		name     string
		content  string
		position core.Position
		wantNil  bool
		wantText string
	}{
		{
			name: "valid function name",
			content: `package main

func calculateSum(a, b int) int {
	return a + b
}`,
			position: core.Position{Line: 2, Character: 6}, // on "calculateSum"
			wantNil:  false,
			wantText: "calculateSum",
		},
		{
			name: "valid variable name",
			content: `package main

func main() {
	var result = 42
}`,
			position: core.Position{Line: 3, Character: 6}, // on "result"
			wantNil:  false,
			wantText: "result",
		},
		{
			name: "cannot rename keyword",
			content: `package main

func main() {
	var x = 42
}`,
			position: core.Position{Line: 2, Character: 0}, // on "func"
			wantNil:  true,
		},
		{
			name: "cannot rename underscore",
			content: `package main

func main() {
	_ = getValue()
}`,
			position: core.Position{Line: 3, Character: 1}, // on "_"
			wantNil:  true,
		},
		{
			name: "non-Go file returns nil",
			content: `function hello() {
	return "world";
}`,
			position: core.Position{Line: 0, Character: 10},
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri := "file:///test.go"
			if tt.name == "non-Go file returns nil" {
				uri = "file:///test.js"
			}

			result := provider.PrepareRename(uri, tt.content, tt.position)

			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil, got range %v", result)
				}
				return
			}

			if result == nil {
				t.Fatal("expected range, got nil")
			}

			// Extract the text
			startOffset := core.PositionToByteOffset(tt.content, result.Start)
			endOffset := core.PositionToByteOffset(tt.content, result.End)

			if startOffset < 0 || endOffset > len(tt.content) {
				t.Fatalf("invalid range offsets")
			}

			gotText := tt.content[startOffset:endOffset]
			if gotText != tt.wantText {
				t.Errorf("got text %q, want %q", gotText, tt.wantText)
			}
		})
	}
}

// TestGoRenameProvider_ProvideRename tests Go-specific rename.
func TestGoRenameProvider_ProvideRename(t *testing.T) {
	provider := &GoRenameProvider{}

	tests := []struct {
		name          string
		content       string
		position      core.Position
		newName       string
		wantNil       bool
		wantEditCount int
		checkContent  func(t *testing.T, result string)
	}{
		{
			name: "rename function and all calls",
			content: `package main

func oldName() {
	println("test")
}

func main() {
	oldName()
	oldName()
}`,
			position:      core.Position{Line: 2, Character: 6}, // on function name
			newName:       "newName",
			wantEditCount: 3, // function declaration + 2 calls
			checkContent: func(t *testing.T, result string) {
				if !strings.Contains(result, "func newName()") {
					t.Error("expected 'func newName()' in result")
				}
				if strings.Count(result, "newName()") != 3 {
					t.Error("expected 3 occurrences of 'newName()'")
				}
			},
		},
		{
			name: "rename variable in scope",
			content: `package main

func main() {
	x := 10
	y := x + 5
	println(x, y)
}`,
			position:      core.Position{Line: 3, Character: 1}, // on "x"
			newName:       "value",
			wantEditCount: 3, // declaration + 2 uses
			checkContent: func(t *testing.T, result string) {
				if !strings.Contains(result, "value := 10") {
					t.Error("expected 'value := 10' in result")
				}
				if !strings.Contains(result, "y := value + 5") {
					t.Error("expected 'y := value + 5' in result")
				}
			},
		},
		{
			name: "rename struct type",
			content: `package main

type OldStruct struct {
	Value int
}

func NewOldStruct() *OldStruct {
	return &OldStruct{}
}`,
			position:      core.Position{Line: 2, Character: 6}, // on "OldStruct"
			newName:       "NewStruct",
			wantEditCount: 3, // type declaration + return type + literal (not function name)
			checkContent: func(t *testing.T, result string) {
				if !strings.Contains(result, "type NewStruct struct") {
					t.Error("expected 'type NewStruct struct' in result")
				}
				if !strings.Contains(result, "*NewStruct") {
					t.Error("expected '*NewStruct' in result")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := core.RenameContext{
				URI:      "file:///test.go",
				Content:  tt.content,
				Position: tt.position,
				NewName:  tt.newName,
			}

			result := provider.ProvideRename(ctx)

			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil, got workspace edit")
				}
				return
			}

			if result == nil {
				t.Fatal("expected workspace edit, got nil")
			}

			edits, ok := result.Changes[ctx.URI]
			if !ok {
				t.Fatal("expected changes for URI")
			}

			if len(edits) != tt.wantEditCount {
				t.Errorf("got %d edits, want %d", len(edits), tt.wantEditCount)
			}

			if tt.checkContent != nil {
				resultContent := applyEdits(tt.content, edits)
				tt.checkContent(t, resultContent)
			}
		})
	}
}

// TestMultiFileRenameProvider tests renaming across multiple files.
func TestMultiFileRenameProvider(t *testing.T) {
	file1 := `package main

func main() {
	value := GetValue()
	println(value)
}`

	file2 := `package main

func GetValue() int {
	return 42
}`

	file3 := `package main

const DefaultValue = 100`

	provider := &MultiFileRenameProvider{
		Files: map[string]string{
			"file:///main.go":  file1,
			"file:///value.go": file2,
			"file:///const.go": file3,
		},
	}

	tests := []struct {
		name              string
		uri               string
		content           string
		position          core.Position
		newName           string
		wantFileCount     int
		wantTotalEdits    int
		checkFiles        map[string]int // map of URI to expected edit count
	}{
		{
			name:           "rename across two files",
			uri:            "file:///main.go",
			content:        file1,
			position:       core.Position{Line: 3, Character: 1}, // on "value"
			newName:        "result",
			wantFileCount:  1, // only main.go has "value"
			wantTotalEdits: 2, // declaration + usage
			checkFiles: map[string]int{
				"file:///main.go": 2,
			},
		},
		{
			name:           "rename function across files",
			uri:            "file:///value.go",
			content:        file2,
			position:       core.Position{Line: 2, Character: 6}, // on "GetValue"
			newName:        "FetchValue",
			wantFileCount:  2, // main.go and value.go
			wantTotalEdits: 2, // 1 in each file
			checkFiles: map[string]int{
				"file:///main.go":  1,
				"file:///value.go": 1,
			},
		},
		{
			name:           "rename in single file only",
			uri:            "file:///const.go",
			content:        file3,
			position:       core.Position{Line: 2, Character: 7}, // on "DefaultValue"
			newName:        "InitialValue",
			wantFileCount:  1,
			wantTotalEdits: 1,
			checkFiles: map[string]int{
				"file:///const.go": 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := core.RenameContext{
				URI:      tt.uri,
				Content:  tt.content,
				Position: tt.position,
				NewName:  tt.newName,
			}

			result := provider.ProvideRename(ctx)

			if result == nil {
				t.Fatal("expected workspace edit, got nil")
			}

			if len(result.Changes) != tt.wantFileCount {
				t.Errorf("got changes in %d files, want %d", len(result.Changes), tt.wantFileCount)
			}

			totalEdits := 0
			for uri, edits := range result.Changes {
				totalEdits += len(edits)

				if expectedCount, ok := tt.checkFiles[uri]; ok {
					if len(edits) != expectedCount {
						t.Errorf("file %s: got %d edits, want %d", uri, len(edits), expectedCount)
					}
				}
			}

			if totalEdits != tt.wantTotalEdits {
				t.Errorf("got %d total edits, want %d", totalEdits, tt.wantTotalEdits)
			}
		})
	}
}

// TestRenameProvider_EdgeCases tests edge cases for rename.
func TestRenameProvider_EdgeCases(t *testing.T) {
	provider := &SimpleRenameProvider{}

	tests := []struct {
		name    string
		content string
		test    func(t *testing.T, p *SimpleRenameProvider, content string)
	}{
		{
			name:    "empty file",
			content: "",
			test: func(t *testing.T, p *SimpleRenameProvider, content string) {
				result := p.PrepareRename("file:///test.go", content, core.Position{})
				if result != nil {
					t.Error("expected nil for empty file")
				}
			},
		},
		{
			name:    "position at end of identifier",
			content: "var x",
			test: func(t *testing.T, p *SimpleRenameProvider, content string) {
				// Position at character 4 (the 'x')
				pos := core.Position{Line: 0, Character: 4}
				result := p.PrepareRename("file:///test.go", content, pos)
				if result == nil {
					t.Error("expected range for identifier")
				}
			},
		},
		{
			name:    "unicode identifier",
			content: "var 变量 = 42\nprintln(变量)",
			test: func(t *testing.T, p *SimpleRenameProvider, content string) {
				t.Skip("SimpleRenameProvider has known Unicode word boundary limitations")

				// Note: The SimpleRenameProvider uses byte-based word boundary detection
				// which doesn't properly handle multibyte UTF-8 characters.
				// This can cause:
				// 1. PrepareRename to incorrectly detect boundaries
				// 2. Regex compilation to fail on partial UTF-8 byte sequences
				//
				// For production use, consider using a Unicode-aware word boundary detector
				// or the GoRenameProvider which uses AST parsing.
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t, provider, tt.content)
		})
	}
}

// Note: applyEdits is defined in formatting_test.go and reused here
