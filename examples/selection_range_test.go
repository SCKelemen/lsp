package examples

import (
	"testing"

	"github.com/SCKelemen/lsp/core"
)

func TestSelectionRangeProvider(t *testing.T) {
	provider := &GoSelectionRangeProvider{}

	tests := []struct {
		name     string
		content  string
		position core.Position
		wantMin  int // minimum expected hierarchy depth
	}{
		{
			name: "identifier in function",
			content: `package main

func example() {
	x := 42
}`,
			position: core.Position{Line: 3, Character: 1}, // on 'x'
			wantMin:  2,                                    // at least identifier -> statement
		},
		{
			name: "function call",
			content: `package main

func main() {
	println("hello")
}`,
			position: core.Position{Line: 3, Character: 2}, // on 'println'
			wantMin:  2,
		},
		{
			name: "struct field",
			content: `package main

type Person struct {
	Name string
}`,
			position: core.Position{Line: 3, Character: 1}, // on 'Name'
			wantMin:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ranges := provider.ProvideSelectionRanges("test.go", tt.content, []core.Position{tt.position})

			if len(ranges) != 1 {
				t.Fatalf("expected 1 selection range, got %d", len(ranges))
			}

			// Count hierarchy depth
			depth := 0
			current := &ranges[0]
			for current != nil {
				depth++
				current = current.Parent
			}

			if depth < tt.wantMin {
				t.Errorf("expected at least %d levels in hierarchy, got %d", tt.wantMin, depth)
			}

			// Verify hierarchy is properly nested (each parent should contain the child)
			current = &ranges[0]
			for current != nil && current.Parent != nil {
				parent := current.Parent
				if !rangeContains(parent.Range, current.Range) {
					t.Errorf("parent range %+v does not contain child range %+v", parent.Range, current.Range)
				}
				current = parent
			}
		})
	}
}

func TestSelectionRangeNonGoFile(t *testing.T) {
	provider := &GoSelectionRangeProvider{}
	content := "not go code"
	position := core.Position{Line: 0, Character: 0}

	ranges := provider.ProvideSelectionRanges("test.txt", content, []core.Position{position})

	if ranges != nil {
		t.Errorf("expected nil for non-go file, got %d ranges", len(ranges))
	}
}

func TestSelectionRangeMultiplePositions(t *testing.T) {
	provider := &GoSelectionRangeProvider{}
	content := `package main

func main() {
	x := 1
	y := 2
}`

	positions := []core.Position{
		{Line: 3, Character: 1}, // on 'x'
		{Line: 4, Character: 1}, // on 'y'
	}

	ranges := provider.ProvideSelectionRanges("test.go", content, positions)

	if len(ranges) != 2 {
		t.Fatalf("expected 2 selection ranges, got %d", len(ranges))
	}

	// Both should have some hierarchy
	for i, r := range ranges {
		if r.Parent == nil {
			t.Errorf("selection range %d has no parent", i)
		}
	}
}

func rangeContains(outer, inner core.Range) bool {
	// Outer starts before or at inner start
	if outer.Start.Line > inner.Start.Line {
		return false
	}
	if outer.Start.Line == inner.Start.Line && outer.Start.Character > inner.Start.Character {
		return false
	}

	// Outer ends after or at inner end
	if outer.End.Line < inner.End.Line {
		return false
	}
	if outer.End.Line == inner.End.Line && outer.End.Character < inner.End.Character {
		return false
	}

	return true
}
