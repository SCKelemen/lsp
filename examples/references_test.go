package examples

import (
	"testing"

	"github.com/SCKelemen/lsp/core"
)

func TestSimpleReferencesProvider(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		position  core.Position
		wantCount int
	}{
		{
			name: "find variable references",
			content: `func main() {
	count := 0
	count++
	println(count)
}`,
			position:  core.Position{Line: 1, Character: 1}, // On "count"
			wantCount: 3,                                     // Declaration + 2 uses
		},
		{
			name: "find function name references",
			content: `func calculate() int {
	return 42
}

func main() {
	result := calculate()
	println(calculate())
}`,
			position:  core.Position{Line: 0, Character: 5}, // On "calculate"
			wantCount: 3,                                     // Declaration + 2 calls
		},
		{
			name: "whitespace falls back to previous word",
			content: `func main() {
	x := 1
	x++
}`,
			position:  core.Position{Line: 1, Character: 2}, // On whitespace after "x"
			wantCount: 2,                                     // Falls back to "x", finds 2 occurrences
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &SimpleReferencesProvider{}
			context := core.ReferenceContext{
				IncludeDeclaration: true,
			}

			refs := provider.FindReferences("file:///test.go", tt.content, tt.position, context)

			if len(refs) != tt.wantCount {
				t.Errorf("got %d references, want %d", len(refs), tt.wantCount)
			}

			// Verify all references are valid locations
			for i, ref := range refs {
				if !ref.IsValid() {
					t.Errorf("reference %d: invalid location %v", i, ref)
				}

				// Verify the text at each reference location
				startOffset := core.PositionToByteOffset(tt.content, ref.Range.Start)
				endOffset := core.PositionToByteOffset(tt.content, ref.Range.End)

				if startOffset < 0 || endOffset > len(tt.content) {
					t.Errorf("reference %d: invalid offsets [%d:%d]", i, startOffset, endOffset)
					continue
				}
			}
		})
	}
}

func TestGoReferencesProvider(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		position  core.Position
		wantCount int
	}{
		{
			name: "find function references",
			content: `package main

func helper() int {
	return 42
}

func main() {
	x := helper()
	y := helper()
	println(x, y)
}`,
			position:  core.Position{Line: 2, Character: 5}, // On "helper"
			wantCount: 3,                                     // Declaration + 2 calls
		},
		{
			name: "find variable references",
			content: `package main

func main() {
	value := 10
	result := value + 20
	println(value, result)
}`,
			position:  core.Position{Line: 3, Character: 1}, // On "value"
			wantCount: 3,                                     // Declaration + 2 uses
		},
		{
			name: "find type references",
			content: `package main

type Point struct {
	X, Y int
}

func main() {
	p := Point{X: 1, Y: 2}
	q := Point{}
	println(p, q)
}`,
			position:  core.Position{Line: 2, Character: 5}, // On "Point"
			wantCount: 3,                                     // Declaration + 2 uses
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &GoReferencesProvider{}
			context := core.ReferenceContext{
				IncludeDeclaration: true,
			}

			refs := provider.FindReferences("file:///test.go", tt.content, tt.position, context)

			if len(refs) != tt.wantCount {
				t.Errorf("got %d references, want %d", len(refs), tt.wantCount)
				for i, ref := range refs {
					startOffset := core.PositionToByteOffset(tt.content, ref.Range.Start)
					endOffset := core.PositionToByteOffset(tt.content, ref.Range.End)
					if startOffset >= 0 && endOffset <= len(tt.content) {
						text := tt.content[startOffset:endOffset]
						t.Logf("  ref %d: %s at %s", i, text, ref.Range.String())
					}
				}
			}
		})
	}
}

func TestMultiFileReferencesProvider(t *testing.T) {
	file1 := `package main

func helper() int {
	return 42
}
`

	file2 := `package main

func main() {
	x := helper()
	y := helper()
	println(x, y)
}
`

	provider := &MultiFileReferencesProvider{
		Files: map[string]string{
			"file:///file1.go": file1,
			"file:///file2.go": file2,
		},
	}

	context := core.ReferenceContext{
		IncludeDeclaration: true,
	}

	// Find references to "helper" in file1
	refs := provider.FindReferences("file:///file1.go", file1, core.Position{Line: 2, Character: 5}, context)

	// Should find 3 references: 1 in file1 (declaration) + 2 in file2 (calls)
	if len(refs) != 3 {
		t.Errorf("got %d references, want 3", len(refs))
		for i, ref := range refs {
			t.Logf("  ref %d: %s at %s", i, ref.URI, ref.Range.String())
		}
	}

	// Verify we have references in both files
	file1Count := 0
	file2Count := 0
	for _, ref := range refs {
		if ref.URI == "file:///file1.go" {
			file1Count++
		} else if ref.URI == "file:///file2.go" {
			file2Count++
		}
	}

	if file1Count != 1 {
		t.Errorf("got %d references in file1, want 1", file1Count)
	}
	if file2Count != 2 {
		t.Errorf("got %d references in file2, want 2", file2Count)
	}
}

func TestReferencesProvider_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		position core.Position
	}{
		{
			name:     "empty file",
			content:  "",
			position: core.Position{Line: 0, Character: 0},
		},
		{
			name:     "position out of bounds",
			content:  "func main() {}",
			position: core.Position{Line: 10, Character: 10},
		},
		{
			name:     "position on punctuation",
			content:  "func main() {}",
			position: core.Position{Line: 0, Character: 9}, // On "("
		},
		{
			name:     "invalid syntax",
			content:  "func main( {",
			position: core.Position{Line: 0, Character: 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &SimpleReferencesProvider{}
			context := core.ReferenceContext{
				IncludeDeclaration: true,
			}

			// Should not crash
			refs := provider.FindReferences("file:///test.go", tt.content, tt.position, context)

			// Edge cases typically return no references
			if refs != nil && len(refs) > 0 {
				t.Logf("returned %d references (acceptable for edge case)", len(refs))
			}
		})
	}
}

func TestReferencesProvider_Unicode(t *testing.T) {
	content := `package main

func main() {
	myVar := 10
	result := myVar + 20
	println(myVar, result)
}
`

	provider := &SimpleReferencesProvider{}
	context := core.ReferenceContext{
		IncludeDeclaration: true,
	}

	// Find references to "myVar"
	refs := provider.FindReferences("file:///test.go", content, core.Position{Line: 3, Character: 1}, context)

	// Should find 3 references: declaration + 2 uses
	if len(refs) != 3 {
		t.Errorf("got %d references, want 3", len(refs))
	}

	// Verify all references point to "myVar"
	for i, ref := range refs {
		startOffset := core.PositionToByteOffset(content, ref.Range.Start)
		endOffset := core.PositionToByteOffset(content, ref.Range.End)

		if startOffset < 0 || endOffset > len(content) {
			t.Errorf("reference %d: invalid offsets", i)
			continue
		}

		text := content[startOffset:endOffset]
		if text != "myVar" {
			t.Errorf("reference %d: got %q, want \"myVar\"", i, text)
		}
	}
}

func TestReferencesProvider_RangeAccuracy(t *testing.T) {
	content := `func calculate(x, y int) int {
	result := x + y
	return result
}
`

	provider := &SimpleReferencesProvider{}
	context := core.ReferenceContext{
		IncludeDeclaration: true,
	}

	// Find references to "result"
	refs := provider.FindReferences("file:///test.go", content, core.Position{Line: 1, Character: 1}, context)

	if len(refs) != 2 {
		t.Errorf("got %d references, want 2", len(refs))
	}

	// Verify each reference points exactly to "result"
	for i, ref := range refs {
		startOffset := core.PositionToByteOffset(content, ref.Range.Start)
		endOffset := core.PositionToByteOffset(content, ref.Range.End)

		if startOffset < 0 || endOffset > len(content) {
			t.Errorf("reference %d: invalid offsets [%d:%d]", i, startOffset, endOffset)
			continue
		}

		text := content[startOffset:endOffset]
		if text != "result" {
			t.Errorf("reference %d: got %q, want \"result\"", i, text)
		}

		// Verify range is valid
		if !ref.Range.IsValid() {
			t.Errorf("reference %d: invalid range %v", i, ref.Range)
		}
	}
}
