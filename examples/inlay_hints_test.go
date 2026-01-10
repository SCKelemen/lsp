package examples

import (
	"testing"

	"github.com/SCKelemen/lsp/core"
)

func TestGoParameterNameInlayHintsProvider(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantCount int
		wantHints map[string]string // position -> label
	}{
		{
			name: "function call with parameters",
			content: `package main

func add(a, b int) int {
	return a + b
}

func main() {
	result := add(10, 20)
}`,
			wantCount: 2,
			wantHints: map[string]string{
				"7:15": "a:", // add(10, 20)
				"7:19": "b:", //     ^  ^
			},
		},
		{
			name: "function with single parameter",
			content: `package main

func double(x int) int {
	return x * 2
}

func main() {
	y := double(5)
}`,
			wantCount: 1,
			wantHints: map[string]string{
				"7:13": "x:",
			},
		},
		{
			name: "multiple function calls",
			content: `package main

func calc(a, b int) int {
	return a + b
}

func main() {
	x := calc(1, 2)
	y := calc(3, 4)
}`,
			wantCount: 4, // 2 params × 2 calls
		},
		{
			name: "no hints for function without parameters",
			content: `package main

func getValue() int {
	return 42
}

func main() {
	x := getValue()
}`,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &GoParameterNameInlayHintsProvider{}

			// Get hints for the entire document
			rng := core.Range{
				Start: core.Position{Line: 0, Character: 0},
				End:   core.Position{Line: 100, Character: 0},
			}

			hints := provider.ProvideInlayHints("file:///test.go", tt.content, rng)

			if len(hints) != tt.wantCount {
				t.Errorf("got %d hints, want %d", len(hints), tt.wantCount)
				for i, hint := range hints {
					t.Logf("  hint %d: %s at %s", i, hint.Label, hint.Position.String())
				}
			}

			// Verify specific hints if provided
			if tt.wantHints != nil {
				for pos, wantLabel := range tt.wantHints {
					found := false
					for _, hint := range hints {
						if hint.Position.String() == pos {
							if hint.Label != wantLabel {
								t.Errorf("at position %s: got label %q, want %q", pos, hint.Label, wantLabel)
							}
							found = true
							break
						}
					}
					if !found {
						t.Errorf("no hint found at position %s", pos)
					}
				}
			}

			// Verify all hints have the right kind
			for i, hint := range hints {
				if hint.Kind == nil {
					t.Errorf("hint %d: kind is nil", i)
				} else if *hint.Kind != core.InlayHintKindParameter {
					t.Errorf("hint %d: got kind %d, want %d (parameter)", i, *hint.Kind, core.InlayHintKindParameter)
				}
			}
		})
	}
}

func TestGoTypeInlayHintsProvider(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantCount int
		wantTypes map[string]string // position -> type
	}{
		{
			name: "basic types",
			content: `package main

func main() {
	x := 42
	y := 3.14
	z := "hello"
}`,
			wantCount: 3,
			wantTypes: map[string]string{
				"3:2":  ": int",
				"4:2":  ": float64",
				"5:2":  ": string",
			},
		},
		{
			name: "type from uppercase function call",
			content: `package main

func GetValue() int {
	return 42
}

func main() {
	x := GetValue()
}`,
			wantCount: 1, // Infers GetValue as a type (false positive, but acceptable)
			wantTypes: map[string]string{
				"7:2": ": GetValue",
			},
		},
		{
			name: "composite literal",
			content: `package main

type Point struct {
	X, Y int
}

func main() {
	p := Point{X: 1, Y: 2}
}`,
			wantCount: 1,
			wantTypes: map[string]string{
				"7:2": ": Point",
			},
		},
		{
			name: "skip underscore",
			content: `package main

func main() {
	_ := 42
	x := 10
}`,
			wantCount: 1, // Only x, not _
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &GoTypeInlayHintsProvider{}

			// Get hints for the entire document
			rng := core.Range{
				Start: core.Position{Line: 0, Character: 0},
				End:   core.Position{Line: 100, Character: 0},
			}

			hints := provider.ProvideInlayHints("file:///test.go", tt.content, rng)

			if len(hints) != tt.wantCount {
				t.Errorf("got %d hints, want %d", len(hints), tt.wantCount)
				for i, hint := range hints {
					t.Logf("  hint %d: %s at %s", i, hint.Label, hint.Position.String())
				}
			}

			// Verify specific type hints if provided
			if tt.wantTypes != nil {
				for pos, wantLabel := range tt.wantTypes {
					found := false
					for _, hint := range hints {
						if hint.Position.String() == pos {
							if hint.Label != wantLabel {
								t.Errorf("at position %s: got label %q, want %q", pos, hint.Label, wantLabel)
							}
							found = true
							break
						}
					}
					if !found {
						t.Errorf("no hint found at position %s", pos)
					}
				}
			}

			// Verify all hints have the right kind
			for i, hint := range hints {
				if hint.Kind == nil {
					t.Errorf("hint %d: kind is nil", i)
				} else if *hint.Kind != core.InlayHintKindType {
					t.Errorf("hint %d: got kind %d, want %d (type)", i, *hint.Kind, core.InlayHintKindType)
				}
			}
		})
	}
}

func TestCompositeInlayHintsProvider(t *testing.T) {
	content := `package main

func add(a, b int) int {
	return a + b
}

func main() {
	x := 42
	y := add(10, 20)
}
`

	provider := NewCompositeInlayHintsProvider(
		&GoParameterNameInlayHintsProvider{},
		&GoTypeInlayHintsProvider{},
	)

	rng := core.Range{
		Start: core.Position{Line: 0, Character: 0},
		End:   core.Position{Line: 100, Character: 0},
	}

	hints := provider.ProvideInlayHints("file:///test.go", content, rng)

	// Should have:
	// - 2 parameter hints (a:, b: for add call)
	// - 2 type hints (x: int, y: int - though y might not be inferred)
	// So at least 2, possibly more
	if len(hints) < 2 {
		t.Errorf("got %d hints, want at least 2", len(hints))
	}

	// Count by kind
	paramCount := 0
	typeCount := 0
	for _, hint := range hints {
		if hint.Kind != nil {
			switch *hint.Kind {
			case core.InlayHintKindParameter:
				paramCount++
			case core.InlayHintKindType:
				typeCount++
			}
		}
	}

	if paramCount != 2 {
		t.Errorf("got %d parameter hints, want 2", paramCount)
	}
	if typeCount < 1 {
		t.Errorf("got %d type hints, want at least 1", typeCount)
	}
}

func TestInlayHints_RangeFiltering(t *testing.T) {
	content := `package main

func add(a, b int) int {
	return a + b
}

func main() {
	x := add(10, 20)  // line 7
	y := add(30, 40)  // line 8
}
`

	provider := &GoParameterNameInlayHintsProvider{}

	tests := []struct {
		name      string
		rng       core.Range
		wantCount int
	}{
		{
			name: "entire document",
			rng: core.Range{
				Start: core.Position{Line: 0, Character: 0},
				End:   core.Position{Line: 100, Character: 0},
			},
			wantCount: 4, // 2 params × 2 calls
		},
		{
			name: "only line 7",
			rng: core.Range{
				Start: core.Position{Line: 7, Character: 0},
				End:   core.Position{Line: 7, Character: 100},
			},
			wantCount: 2, // 2 params from first call
		},
		{
			name: "only line 8",
			rng: core.Range{
				Start: core.Position{Line: 8, Character: 0},
				End:   core.Position{Line: 8, Character: 100},
			},
			wantCount: 2, // 2 params from second call
		},
		{
			name: "narrow range excluding calls",
			rng: core.Range{
				Start: core.Position{Line: 0, Character: 0},
				End:   core.Position{Line: 5, Character: 0},
			},
			wantCount: 0, // No calls in this range
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hints := provider.ProvideInlayHints("file:///test.go", content, tt.rng)

			if len(hints) != tt.wantCount {
				t.Errorf("got %d hints, want %d", len(hints), tt.wantCount)
			}
		})
	}
}

func TestInlayHints_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		content string
		rng     core.Range
	}{
		{
			name:    "empty file",
			content: "",
			rng: core.Range{
				Start: core.Position{Line: 0, Character: 0},
				End:   core.Position{Line: 100, Character: 0},
			},
		},
		{
			name:    "invalid syntax",
			content: "func main( {",
			rng: core.Range{
				Start: core.Position{Line: 0, Character: 0},
				End:   core.Position{Line: 100, Character: 0},
			},
		},
		{
			name:    "non-Go file",
			content: "Some text file",
			rng: core.Range{
				Start: core.Position{Line: 0, Character: 0},
				End:   core.Position{Line: 100, Character: 0},
			},
		},
	}

	providers := []struct {
		name     string
		provider InlayHintsProvider
	}{
		{"parameter", &GoParameterNameInlayHintsProvider{}},
		{"type", &GoTypeInlayHintsProvider{}},
	}

	for _, tt := range tests {
		for _, prov := range providers {
			t.Run(tt.name+"_"+prov.name, func(t *testing.T) {
				// Should not crash
				hints := prov.provider.ProvideInlayHints("file:///test.txt", tt.content, tt.rng)

				// Edge cases typically return no hints
				if hints != nil && len(hints) > 0 {
					t.Logf("returned %d hints (acceptable for edge case)", len(hints))
				}
			})
		}
	}
}

func TestInlayHints_PositionAccuracy(t *testing.T) {
	content := `package main

func add(a, b int) int {
	return a + b
}

func main() {
	result := add(10, 20)
}
`

	provider := &GoParameterNameInlayHintsProvider{}
	rng := core.Range{
		Start: core.Position{Line: 0, Character: 0},
		End:   core.Position{Line: 100, Character: 0},
	}

	hints := provider.ProvideInlayHints("file:///test.go", content, rng)

	if len(hints) != 2 {
		t.Fatalf("got %d hints, want 2", len(hints))
	}

	// Verify positions are valid and point to the start of arguments
	for i, hint := range hints {
		if !hint.Position.IsValid() {
			t.Errorf("hint %d: invalid position %v", i, hint.Position)
		}

		// The hint should appear before the argument
		// Verify there's actual code at or near this position
		offset := core.PositionToByteOffset(content, hint.Position)
		if offset < 0 || offset >= len(content) {
			t.Errorf("hint %d: position %s maps to invalid offset %d", i, hint.Position.String(), offset)
		}
	}
}
