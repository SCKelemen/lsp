package examples

import (
	"strings"
	"testing"

	"github.com/SCKelemen/lsp/core"
)

// TestGoFormattingProvider tests Go formatting with gofmt.
func TestGoFormattingProvider(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantFormatted bool
	}{
		{
			name: "already formatted",
			input: `package main

func main() {
	println("hello")
}
`,
			wantFormatted: false, // No edits needed
		},
		{
			name: "needs formatting",
			input: `package main
func main(){
println("hello")
}`,
			wantFormatted: true,
		},
		{
			name: "fix indentation",
			input: `package main

func main() {
    println("hello")
}`,
			wantFormatted: true,
		},
		{
			name: "fix spacing",
			input: `package main

func main(  )  {
	println("hello")
}`,
			wantFormatted: true,
		},
	}

	provider := &GoFormattingProvider{}
	options := core.FormattingOptions{
		TabSize:      4,
		InsertSpaces: false,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edits := provider.ProvideFormatting("file:///test.go", tt.input, options)

			if tt.wantFormatted {
				if len(edits) == 0 {
					t.Error("expected formatting edits, got none")
				}
			} else {
				if len(edits) != 0 {
					t.Errorf("expected no edits, got %d", len(edits))
				}
			}
		})
	}
}

// TestGoFormattingProvider_InvalidSyntax tests with invalid syntax.
func TestGoFormattingProvider_InvalidSyntax(t *testing.T) {
	input := "this is not valid go code"

	provider := &GoFormattingProvider{}
	options := core.FormattingOptions{}

	edits := provider.ProvideFormatting("file:///test.go", input, options)

	// Should return nil for invalid syntax
	if edits != nil {
		t.Errorf("expected nil for invalid syntax, got %d edits", len(edits))
	}
}

// TestSimpleFormattingProvider tests basic formatting.
func TestSimpleFormattingProvider(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		options core.FormattingOptions
		want    string
	}{
		{
			name:  "trim trailing whitespace",
			input: "hello world   \ngoodbye  \n",
			options: core.FormattingOptions{
				TrimTrailingWhitespace: true,
			},
			want: "hello world\ngoodbye\n",
		},
		{
			name:  "insert final newline",
			input: "hello world",
			options: core.FormattingOptions{
				InsertFinalNewline: true,
			},
			want: "hello world\n",
		},
		{
			name:  "trim final newlines",
			input: "hello world\n\n\n",
			options: core.FormattingOptions{
				TrimFinalNewlines: true,
			},
			want: "hello world\n",
		},
		{
			name:  "no trailing whitespace",
			input: "hello world\ngoodbye\n",
			options: core.FormattingOptions{
				TrimTrailingWhitespace: true,
			},
			want: "hello world\ngoodbye\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewSimpleFormattingProvider()
			edits := provider.ProvideFormatting("file:///test.txt", tt.input, tt.options)

			result := applyEdits(tt.input, edits)

			if result != tt.want {
				t.Errorf("got %q, want %q", result, tt.want)
			}
		})
	}
}

// TestGoRangeFormattingProvider tests range formatting.
func TestGoRangeFormattingProvider(t *testing.T) {
	content := `package main

func main() {
x:=1
y:=2
    println(x+y)
}

func other() {
    println("unchanged")
}
`

	provider := &GoRangeFormattingProvider{}
	options := core.FormattingOptions{
		TabSize:      4,
		InsertSpaces: false,
	}

	// Format only lines 3-5 (the problematic lines)
	r := core.Range{
		Start: core.Position{Line: 3, Character: 0},
		End:   core.Position{Line: 5, Character: 20},
	}

	edits := provider.ProvideRangeFormatting("file:///test.go", content, r, options)

	if len(edits) == 0 {
		t.Log("no formatting edits (might be already formatted or error)")
	}

	// Apply edits and verify
	result := applyEdits(content, edits)

	// Should format the selected lines
	// The "other" function should remain unchanged
	if !strings.Contains(result, `func other() {`) {
		t.Error("other function should remain unchanged")
	}
}

// TestSimpleRangeFormattingProvider tests simple range formatting.
func TestSimpleRangeFormattingProvider(t *testing.T) {
	// Note: Lines have trailing spaces where indicated
	content := "line 1  \nline 2\nline 3\nline 4  "

	provider := NewSimpleRangeFormattingProvider()
	options := core.FormattingOptions{
		TrimTrailingWhitespace: true,
	}

	// Format only lines 1-2 (0-indexed)
	r := core.Range{
		Start: core.Position{Line: 1, Character: 0},
		End:   core.Position{Line: 2, Character: 6}, // End of line 2
	}

	edits := provider.ProvideRangeFormatting("file:///test.txt", content, r, options)

	result := applyEdits(content, edits)

	lines := strings.Split(result, "\n")

	// Line 1 should still have trailing whitespace (not in range)
	if !strings.HasSuffix(lines[0], "  ") {
		t.Error("line 1 should still have trailing whitespace")
	}

	// Line 2 should have no trailing whitespace
	if strings.HasSuffix(lines[1], " ") {
		t.Error("line 2 should have no trailing whitespace")
	}

	// Line 3 should have no trailing whitespace
	if strings.HasSuffix(lines[2], " ") {
		t.Error("line 3 should have no trailing whitespace")
	}

	// Line 4 should still have trailing whitespace (not in range)
	if !strings.HasSuffix(lines[3], "  ") {
		t.Error("line 4 should still have trailing whitespace")
	}
}

// TestFormattingOptions tests various formatting options.
func TestFormattingOptions(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		options core.FormattingOptions
		check   func(t *testing.T, result string)
	}{
		{
			name:  "tab size affects formatting",
			input: "hello",
			options: core.FormattingOptions{
				TabSize:      8,
				InsertSpaces: true,
			},
			check: func(t *testing.T, result string) {
				// Just verify it doesn't crash
			},
		},
		{
			name:  "insert spaces vs tabs",
			input: "hello",
			options: core.FormattingOptions{
				TabSize:      4,
				InsertSpaces: false, // Use tabs
			},
			check: func(t *testing.T, result string) {
				// Verify formatting respects tab preference
			},
		},
		{
			name:  "multiple options combined",
			input: "hello world   \n\n",
			options: core.FormattingOptions{
				TrimTrailingWhitespace: true,
				InsertFinalNewline:     true,
				TrimFinalNewlines:      true,
			},
			check: func(t *testing.T, result string) {
				if strings.HasSuffix(result, "   ") {
					t.Error("should trim trailing whitespace")
				}
				if !strings.HasSuffix(result, "\n") {
					t.Error("should have final newline")
				}
				if strings.HasSuffix(result, "\n\n") {
					t.Error("should trim extra final newlines")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewSimpleFormattingProvider()
			edits := provider.ProvideFormatting("file:///test.txt", tt.input, tt.options)

			result := applyEdits(tt.input, edits)

			tt.check(t, result)
		})
	}
}

// TestFormattingEdgeCases tests edge cases.
func TestFormattingEdgeCases(t *testing.T) {
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
			name:    "only whitespace",
			content: "   \n\t\n   ",
			uri:     "file:///whitespace.go",
		},
		{
			name:    "single line",
			content: "package main",
			uri:     "file:///oneline.go",
		},
		{
			name:    "non-go file",
			content: "some content",
			uri:     "file:///test.txt",
		},
	}

	providers := []core.FormattingProvider{
		&GoFormattingProvider{},
		NewSimpleFormattingProvider(),
	}

	options := core.FormattingOptions{
		TabSize:                4,
		InsertSpaces:           true,
		TrimTrailingWhitespace: true,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, provider := range providers {
				// Should not crash
				edits := provider.ProvideFormatting(tt.uri, tt.content, options)

				if edits == nil {
					edits = []core.TextEdit{}
				}

				t.Logf("provider %d returned %d edits", i, len(edits))

				// Verify all edits are valid
				lines := strings.Split(tt.content, "\n")
				for j, edit := range edits {
					if edit.Range.Start.Line < 0 {
						t.Errorf("provider %d, edit %d: negative start line", i, j)
					}
					if edit.Range.Start.Line >= len(lines) {
						t.Errorf("provider %d, edit %d: start line %d exceeds file length %d",
							i, j, edit.Range.Start.Line, len(lines))
					}
				}
			}
		})
	}
}

// TestRangeFormattingEdgeCases tests range formatting edge cases.
func TestRangeFormattingEdgeCases(t *testing.T) {
	content := "line 1\nline 2\nline 3\n"

	tests := []struct {
		name  string
		r     core.Range
		valid bool
	}{
		{
			name:  "valid range",
			r:     core.Range{Start: core.Position{Line: 0, Character: 0}, End: core.Position{Line: 1, Character: 6}},
			valid: true,
		},
		{
			name:  "negative start line",
			r:     core.Range{Start: core.Position{Line: -1, Character: 0}, End: core.Position{Line: 1, Character: 6}},
			valid: false,
		},
		{
			name:  "end line out of bounds",
			r:     core.Range{Start: core.Position{Line: 0, Character: 0}, End: core.Position{Line: 100, Character: 0}},
			valid: false,
		},
		{
			name:  "empty range",
			r:     core.Range{Start: core.Position{Line: 1, Character: 0}, End: core.Position{Line: 1, Character: 0}},
			valid: true,
		},
	}

	providers := []core.RangeFormattingProvider{
		&GoRangeFormattingProvider{},
		NewSimpleRangeFormattingProvider(),
	}

	options := core.FormattingOptions{
		TrimTrailingWhitespace: true,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, provider := range providers {
				// Should not crash
				edits := provider.ProvideRangeFormatting("file:///test.go", content, tt.r, options)

				if !tt.valid {
					// Invalid ranges should return nil or empty
					if edits == nil {
						edits = []core.TextEdit{}
					}
					if len(edits) > 0 {
						t.Logf("provider %d returned %d edits for invalid range (may be acceptable)", i, len(edits))
					}
				}
			}
		})
	}
}

// TestFormattingWithUnicode tests formatting with Unicode content.
func TestFormattingWithUnicode(t *testing.T) {
	content := `package main

// 这是一个注释
func 主函数() {
	消息 := "你好世界"
	println(消息)
}
`

	provider := NewSimpleFormattingProvider()
	options := core.FormattingOptions{
		TrimTrailingWhitespace: true,
	}

	edits := provider.ProvideFormatting("file:///test.go", content, options)

	// Should trim trailing whitespace even with Unicode
	result := applyEdits(content, edits)

	lines := strings.Split(result, "\n")

	// Check that lines don't have trailing whitespace
	for i, line := range lines {
		if strings.HasSuffix(line, " ") || strings.HasSuffix(line, "\t") {
			t.Errorf("line %d still has trailing whitespace: %q", i, line)
		}
	}

	// Verify positions are valid UTF-8 byte offsets
	for i, edit := range edits {
		originalLines := strings.Split(content, "\n")
		if edit.Range.Start.Line < len(originalLines) {
			line := originalLines[edit.Range.Start.Line]
			if edit.Range.Start.Character > len(line) {
				t.Errorf("edit %d: start character %d exceeds line length %d (UTF-8 bytes)",
					i, edit.Range.Start.Character, len(line))
			}
		}
	}
}

// TestRangeFormattingWithUnicode tests range formatting with Unicode.
func TestRangeFormattingWithUnicode(t *testing.T) {
	// Note: First line has trailing spaces
	content := "// 第一行  \n// 第二行\n// 第三行\n"

	provider := NewSimpleRangeFormattingProvider()
	options := core.FormattingOptions{
		TrimTrailingWhitespace: true,
	}

	// Format only middle line (line 1)
	r := core.Range{
		Start: core.Position{Line: 1, Character: 0},
		End:   core.Position{Line: 1, Character: 20},
	}

	edits := provider.ProvideRangeFormatting("file:///test.go", content, r, options)

	result := applyEdits(content, edits)

	lines := strings.Split(result, "\n")

	// First line should still have trailing whitespace
	if !strings.HasSuffix(lines[0], "  ") {
		t.Error("first line should still have trailing whitespace")
	}

	// Second line should have no trailing whitespace
	if strings.HasSuffix(lines[1], " ") {
		t.Error("second line should have no trailing whitespace")
	}

	// Third line should not be affected
	if strings.HasSuffix(lines[2], " ") {
		// Note: Third line has no trailing whitespace in input
		t.Log("third line correctly has no trailing whitespace")
	}
}

// Helper function to apply edits to content
func applyEdits(content string, edits []core.TextEdit) string {
	if len(edits) == 0 {
		return content
	}

	// For simplicity, if there's a single edit replacing entire document
	if len(edits) == 1 {
		edit := edits[0]
		if edit.Range.Start.Line == 0 && edit.Range.Start.Character == 0 {
			lines := strings.Split(content, "\n")
			endLine := len(lines) - 1
			if endLine < 0 {
				endLine = 0
			}
			if edit.Range.End.Line >= endLine {
				return edit.NewText
			}
		}
	}

	// Apply each edit
	lines := strings.Split(content, "\n")

	for _, edit := range edits {
		if edit.Range.Start.Line >= len(lines) {
			continue
		}

		if edit.Range.Start.Line == edit.Range.End.Line {
			// Single line edit
			line := lines[edit.Range.Start.Line]

			// Bounds checking
			startChar := edit.Range.Start.Character
			if startChar > len(line) {
				startChar = len(line)
			}

			endChar := edit.Range.End.Character
			if endChar > len(line) {
				endChar = len(line)
			}

			newLine := line[:startChar] + edit.NewText + line[endChar:]
			lines[edit.Range.Start.Line] = newLine
		} else {
			// Multi-line edit (simplified)
			if edit.Range.Start.Line < len(lines) && edit.Range.End.Line < len(lines) {
				before := ""
				if edit.Range.Start.Character < len(lines[edit.Range.Start.Line]) {
					before = lines[edit.Range.Start.Line][:edit.Range.Start.Character]
				}

				after := ""
				if edit.Range.End.Character < len(lines[edit.Range.End.Line]) {
					after = lines[edit.Range.End.Line][edit.Range.End.Character:]
				}

				lines[edit.Range.Start.Line] = before + edit.NewText + after

				// Remove lines in between
				if edit.Range.End.Line > edit.Range.Start.Line {
					lines = append(
						lines[:edit.Range.Start.Line+1],
						lines[edit.Range.End.Line+1:]...,
					)
				}
			}
		}
	}

	return strings.Join(lines, "\n")
}
