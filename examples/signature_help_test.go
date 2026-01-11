package examples

import (
	"testing"

	"github.com/SCKelemen/lsp/core"
)

func TestGoSignatureHelpProvider(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		position        core.Position
		wantSignatures  int
		wantLabel       string
		wantActiveParam *int
	}{
		{
			name: "function call with single parameter",
			content: `package main

func Double(x int) int {
	return x * 2
}

func main() {
	y := Double(5)
}`,
			position:        core.Position{Line: 7, Character: 14}, // Inside "Double(5"
			wantSignatures:  1,
			wantLabel:       "Double(x int) int",
			wantActiveParam: intPtr(0),
		},
		{
			name: "function call with multiple parameters - first param",
			content: `package main

func Add(a int, b int) int {
	return a + b
}

func main() {
	result := Add(10, 20)
}`,
			position:        core.Position{Line: 7, Character: 17}, // After "Add(1"
			wantSignatures:  1,
			wantLabel:       "Add(a int, b int) int",
			wantActiveParam: intPtr(0),
		},
		{
			name: "function call with multiple parameters - second param",
			content: `package main

func Add(a int, b int) int {
	return a + b
}

func main() {
	result := Add(10, 20)
}`,
			position:        core.Position{Line: 7, Character: 21}, // After "Add(10, 2"
			wantSignatures:  1,
			wantLabel:       "Add(a int, b int) int",
			wantActiveParam: intPtr(1),
		},
		{
			name: "function with documentation",
			content: `package main

// Multiply multiplies two numbers.
// It returns the product of a and b.
func Multiply(a int, b int) int {
	return a * b
}

func main() {
	result := Multiply(3, 4)
}`,
			position:       core.Position{Line: 9, Character: 22}, // Inside "Multiply(3"
			wantSignatures: 1,
			wantLabel:      "Multiply(a int, b int) int",
		},
		{
			name: "function with no parameters",
			content: `package main

func GetValue() int {
	return 42
}

func main() {
	x := GetValue()
}`,
			position:       core.Position{Line: 7, Character: 14}, // Inside "GetValue()"
			wantSignatures: 1,
			wantLabel:      "GetValue() int",
		},
		{
			name: "function with complex types",
			content: `package main

func Process(data []string, count int, options map[string]bool) error {
	return nil
}

func main() {
	Process(nil, 10, nil)
}`,
			position:       core.Position{Line: 7, Character: 13}, // After "Process(n"
			wantSignatures: 1,
			wantLabel:      "Process(data []string, count int, options map[string]bool) error",
		},
		{
			name: "function with variadic parameters",
			content: `package main

func Sum(values ...int) int {
	total := 0
	for _, v := range values {
		total += v
	}
	return total
}

func main() {
	result := Sum(1, 2, 3)
}`,
			position:       core.Position{Line: 11, Character: 16}, // Inside "Sum(1"
			wantSignatures: 1,
			wantLabel:      "Sum(values ...int) int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &GoSignatureHelpProvider{}
			ctx := core.SignatureHelpContext{
				URI:              "file:///test.go",
				Content:          tt.content,
				Position:         tt.position,
				TriggerCharacter: "(",
			}

			help := provider.ProvideSignatureHelp(ctx)

			if tt.wantSignatures == 0 {
				if help != nil && len(help.Signatures) > 0 {
					t.Errorf("expected no signatures, got %d", len(help.Signatures))
				}
				return
			}

			if help == nil {
				t.Fatal("expected signature help, got nil")
			}

			if len(help.Signatures) != tt.wantSignatures {
				t.Errorf("got %d signatures, want %d", len(help.Signatures), tt.wantSignatures)
			}

			if len(help.Signatures) > 0 {
				sig := help.Signatures[0]

				if sig.Label != tt.wantLabel {
					t.Errorf("got label %q, want %q", sig.Label, tt.wantLabel)
				}

				if tt.wantActiveParam != nil {
					if help.ActiveParameter == nil {
						t.Error("expected active parameter, got nil")
					} else if *help.ActiveParameter != *tt.wantActiveParam {
						t.Errorf("got active parameter %d, want %d", *help.ActiveParameter, *tt.wantActiveParam)
					}
				}
			}
		})
	}
}

func TestGoSignatureHelpProvider_NoHelp(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		position core.Position
	}{
		{
			name: "not in a function call",
			content: `package main

func main() {
	x := 10
}`,
			position: core.Position{Line: 3, Character: 6}, // After "x := 1"
		},
		{
			name: "function not found",
			content: `package main

func main() {
	UnknownFunc()
}`,
			position: core.Position{Line: 3, Character: 15}, // Inside "UnknownFunc("
		},
		{
			name: "non-Go file",
			content: `some text
not go code`,
			position: core.Position{Line: 0, Character: 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &GoSignatureHelpProvider{}

			uri := "file:///test.go"
			if tt.name == "non-Go file" {
				uri = "file:///test.txt"
			}

			ctx := core.SignatureHelpContext{
				URI:      uri,
				Content:  tt.content,
				Position: tt.position,
			}

			help := provider.ProvideSignatureHelp(ctx)

			if help != nil && len(help.Signatures) > 0 {
				t.Errorf("expected no signature help, got %d signatures", len(help.Signatures))
			}
		})
	}
}

func TestGoSignatureHelpProvider_ActiveParameterTracking(t *testing.T) {
	content := `package main

func ThreeParams(a int, b string, c bool) {
}

func main() {
	ThreeParams(1, "hello", true)
}
`

	provider := &GoSignatureHelpProvider{}

	tests := []struct {
		name        string
		position    core.Position
		wantParam   int
	}{
		{
			name:      "first parameter",
			position:  core.Position{Line: 6, Character: 14}, // After "ThreeParams(1"
			wantParam: 0,
		},
		{
			name:      "second parameter",
			position:  core.Position{Line: 6, Character: 20}, // After comma, in "hello"
			wantParam: 1,
		},
		{
			name:      "third parameter",
			position:  core.Position{Line: 6, Character: 28}, // After second comma, in "true"
			wantParam: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := core.SignatureHelpContext{
				URI:      "file:///test.go",
				Content:  content,
				Position: tt.position,
			}

			help := provider.ProvideSignatureHelp(ctx)

			if help == nil || len(help.Signatures) == 0 {
				t.Fatal("expected signature help")
			}

			if help.ActiveParameter == nil {
				t.Fatal("expected active parameter")
			}

			if *help.ActiveParameter != tt.wantParam {
				t.Errorf("got active parameter %d, want %d", *help.ActiveParameter, tt.wantParam)
			}
		})
	}
}

func TestGoSignatureHelpProvider_Documentation(t *testing.T) {
	content := `package main

// Calculator performs arithmetic operations.
// This is a simple calculator function.
func Calculator(x int, y int, op string) int {
	return 0
}

func main() {
	result := Calculator(1, 2, "+")
}
`

	provider := &GoSignatureHelpProvider{}
	ctx := core.SignatureHelpContext{
		URI:      "file:///test.go",
		Content:  content,
		Position: core.Position{Line: 9, Character: 23}, // Inside "Calculator(1"
	}

	help := provider.ProvideSignatureHelp(ctx)

	if help == nil || len(help.Signatures) == 0 {
		t.Fatal("expected signature help")
	}

	sig := help.Signatures[0]

	if sig.Documentation == "" {
		t.Error("expected documentation, got empty string")
	}

	// Check that documentation contains expected content
	if !containsInMiddle(sig.Documentation, "Calculator") {
		t.Errorf("documentation doesn't contain expected text: %q", sig.Documentation)
	}
}
