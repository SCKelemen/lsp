package examples

import (
	"strings"
	"testing"

	"github.com/SCKelemen/lsp/core"
)

func TestGoRangesFormattingProvider(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		ranges    []core.Range
		wantEdits bool
	}{
		{
			name: "format single range",
			content: `package main

func main() {
x:=1
}`,
			ranges: []core.Range{
				{
					Start: core.Position{Line: 3, Character: 0},
					End:   core.Position{Line: 3, Character: 10},
				},
			},
			wantEdits: true,
		},
		{
			name: "format multiple non-contiguous ranges",
			content: `package main

func test1() {
x:=1
}

func test2() {
y:=2
}`,
			ranges: []core.Range{
				{
					Start: core.Position{Line: 3, Character: 0},
					End:   core.Position{Line: 3, Character: 10},
				},
				{
					Start: core.Position{Line: 7, Character: 0},
					End:   core.Position{Line: 7, Character: 10},
				},
			},
			wantEdits: true,
		},
		{
			name: "no edits needed - already formatted",
			content: `package main

func main() {
	x := 1
}`,
			ranges: []core.Range{
				{
					Start: core.Position{Line: 3, Character: 0},
					End:   core.Position{Line: 3, Character: 10},
				},
			},
			wantEdits: false,
		},
		{
			name:      "empty ranges",
			content:   `package main`,
			ranges:    []core.Range{},
			wantEdits: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &GoRangesFormattingProvider{}
			options := core.FormattingOptions{
				TabSize:      4,
				InsertSpaces: false,
			}

			edits := provider.ProvideRangesFormatting("file:///test.go", tt.content, tt.ranges, options)

			hasEdits := len(edits) > 0
			if hasEdits != tt.wantEdits {
				t.Errorf("got edits = %v, want %v (got %d edits)", hasEdits, tt.wantEdits, len(edits))
			}
		})
	}
}

func TestSimpleRangesFormattingProvider(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		ranges    []core.Range
		wantEdits int
	}{
		{
			name: "trim trailing whitespace in single range",
			content: `package main

func main() {
    x := 1   
}`,
			ranges: []core.Range{
				{
					Start: core.Position{Line: 3, Character: 0},
					End:   core.Position{Line: 3, Character: 50},
				},
			},
			wantEdits: 1,
		},
		{
			name: "trim trailing whitespace in multiple ranges",
			content: `package main

func test1() {
    x := 1   
}

func test2() {
    y := 2   
}`,
			ranges: []core.Range{
				{
					Start: core.Position{Line: 3, Character: 0},
					End:   core.Position{Line: 3, Character: 50},
				},
				{
					Start: core.Position{Line: 7, Character: 0},
					End:   core.Position{Line: 7, Character: 50},
				},
			},
			wantEdits: 2,
		},
		{
			name: "no edits needed",
			content: `package main

func main() {
    x := 1
}`,
			ranges: []core.Range{
				{
					Start: core.Position{Line: 3, Character: 0},
					End:   core.Position{Line: 3, Character: 50},
				},
			},
			wantEdits: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewSimpleRangesFormattingProvider()
			options := core.FormattingOptions{
				TabSize:                4,
				InsertSpaces:           true,
				TrimTrailingWhitespace: true,
			}

			edits := provider.ProvideRangesFormatting("file:///test.go", tt.content, tt.ranges, options)

			if len(edits) != tt.wantEdits {
				t.Errorf("got %d edits, want %d", len(edits), tt.wantEdits)
			}
		})
	}
}

func TestRangesFormatting_NonOverlapping(t *testing.T) {
	content := `package main

func main() {
    line1   
    line2
    line3   
    line4
}
`

	provider := NewSimpleRangesFormattingProvider()
	options := core.FormattingOptions{
		TrimTrailingWhitespace: true,
	}

	// Format lines 3 and 5 (non-overlapping)
	ranges := []core.Range{
		{
			Start: core.Position{Line: 3, Character: 0},
			End:   core.Position{Line: 3, Character: 50},
		},
		{
			Start: core.Position{Line: 5, Character: 0},
			End:   core.Position{Line: 5, Character: 50},
		},
	}

	edits := provider.ProvideRangesFormatting("file:///test.go", content, ranges, options)

	// Should have 2 edits (one for each line with trailing whitespace)
	if len(edits) != 2 {
		t.Errorf("got %d edits, want 2", len(edits))
	}

	// Verify edits are for the correct lines
	editLines := make(map[int]bool)
	for _, edit := range edits {
		editLines[edit.Range.Start.Line] = true
	}

	if !editLines[3] {
		t.Error("expected edit for line 3")
	}
	if !editLines[5] {
		t.Error("expected edit for line 5")
	}
}

func TestRangesFormatting_EmptyRanges(t *testing.T) {
	content := `package main

func main() {
    x := 1
}
`

	provider := NewSimpleRangesFormattingProvider()
	options := core.FormattingOptions{
		TrimTrailingWhitespace: true,
	}

	edits := provider.ProvideRangesFormatting("file:///test.go", content, []core.Range{}, options)

	if len(edits) != 0 {
		t.Errorf("expected no edits for empty ranges, got %d", len(edits))
	}
}

func TestRangesFormatting_InvalidRange(t *testing.T) {
	content := `package main

func main() {
    x := 1
}
`

	provider := NewSimpleRangesFormattingProvider()
	options := core.FormattingOptions{
		TrimTrailingWhitespace: true,
	}

	// Invalid range (beyond document bounds)
	ranges := []core.Range{
		{
			Start: core.Position{Line: 100, Character: 0},
			End:   core.Position{Line: 150, Character: 0},
		},
	}

	edits := provider.ProvideRangesFormatting("file:///test.go", content, ranges, options)

	// Should handle gracefully with no edits
	if len(edits) != 0 {
		t.Errorf("expected no edits for invalid range, got %d", len(edits))
	}
}

func TestRangesFormatting_NonGoFile(t *testing.T) {
	content := `some text file
with trailing spaces
more text
`

	provider := &GoRangesFormattingProvider{}
	options := core.FormattingOptions{
		TrimTrailingWhitespace: true,
	}

	ranges := []core.Range{
		{
			Start: core.Position{Line: 1, Character: 0},
			End:   core.Position{Line: 1, Character: 50},
		},
	}

	edits := provider.ProvideRangesFormatting("file:///test.txt", content, ranges, options)

	// Go provider should return nil for non-Go files
	if edits != nil && len(edits) > 0 {
		t.Errorf("expected no edits for non-Go file, got %d", len(edits))
	}
}

func TestRangesFormatting_Integration(t *testing.T) {
	content := `package main

func test1() {
    x := 1
    y := 2
}

func test2() {
    a := 3
    b := 4
}
`

	provider := NewSimpleRangesFormattingProvider()
	options := core.FormattingOptions{
		TabSize:                4,
		InsertSpaces:           true,
		TrimTrailingWhitespace: true,
	}

	// Format specific lines in both functions
	ranges := []core.Range{
		{
			Start: core.Position{Line: 3, Character: 0},
			End:   core.Position{Line: 3, Character: 50},
		},
		{
			Start: core.Position{Line: 8, Character: 0},
			End:   core.Position{Line: 8, Character: 50},
		},
	}

	edits := provider.ProvideRangesFormatting("file:///test.go", content, ranges, options)

	// Content has no trailing whitespace, so no edits expected
	// This test verifies that the provider doesn't generate unnecessary edits
	if len(edits) != 0 {
		t.Errorf("got %d edits, want 0 (no trailing whitespace present)", len(edits))
		for i, edit := range edits {
			t.Logf("Edit %d: line %d, char %d-%d", i, edit.Range.Start.Line, edit.Range.Start.Character, edit.Range.End.Character)
		}
	}
}

func applyEditsToContent(content string, edits []core.TextEdit) string {
	if len(edits) == 0 {
		return content
	}

	lines := strings.Split(content, "\n")

	// Sort edits by position (reverse order for safe application)
	for i := len(edits) - 1; i >= 0; i-- {
		edit := edits[i]
		if edit.Range.Start.Line == edit.Range.End.Line {
			// Single line edit
			if edit.Range.Start.Line < len(lines) {
				line := lines[edit.Range.Start.Line]
				newLine := line[:edit.Range.Start.Character] +
					edit.NewText +
					line[edit.Range.End.Character:]
				lines[edit.Range.Start.Line] = newLine
			}
		}
	}

	return strings.Join(lines, "\n")
}
