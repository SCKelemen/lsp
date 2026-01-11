package examples

import (
	"testing"

	"github.com/SCKelemen/lsp/core"
)

func TestSimpleInlineCompletionProvider(t *testing.T) {
	provider := &SimpleInlineCompletionProvider{}

	tests := []struct {
		name            string
		content         string
		position        core.Position
		wantCompletions bool
		wantContains    string // Check if any completion contains this text
	}{
		{
			name: "fmt.Println suggestion",
			content: `package main

func main() {
	fmt.P
}`,
			position:        core.Position{Line: 3, Character: 6}, // after "fmt.P"
			wantCompletions: true,
			wantContains:    "rintln",
		},
		{
			name: "function definition",
			content: `package main

func `,
			position:        core.Position{Line: 2, Character: 5}, // after "func "
			wantCompletions: true,
			wantContains:    "main",
		},
		{
			name: "error handling",
			content: `package main

func test() error {
	err := doSomething()
	if err`,
			position:        core.Position{Line: 4, Character: 7}, // after "if err"
			wantCompletions: true,
			wantContains:    "!= nil",
		},
		{
			name: "for loop",
			content: `package main

func test() {
	for `,
			position:        core.Position{Line: 3, Character: 5}, // after "for "
			wantCompletions: true,
			wantContains:    "range",
		},
		{
			name: "no completion for non-Go file",
			content: `not go code
test`,
			position:        core.Position{Line: 1, Character: 4},
			wantCompletions: false,
		},
		{
			name: "no completion for empty line",
			content: `package main

`,
			position:        core.Position{Line: 2, Character: 0},
			wantCompletions: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := core.InlineCompletionContext{
				URI:         "test.go",
				Content:     tt.content,
				Position:    tt.position,
				TriggerKind: core.InlineCompletionTriggerKindInvoked,
			}

			result := provider.ProvideInlineCompletions(ctx)

			if tt.wantCompletions {
				if result == nil || len(result.Items) == 0 {
					t.Fatalf("expected completions, got none")
				}

				if tt.wantContains != "" {
					found := false
					for _, item := range result.Items {
						if containsSubstring(item.InsertText, tt.wantContains) ||
							containsSubstring(item.FilterText, tt.wantContains) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected completion containing %q, but found none", tt.wantContains)
					}
				}
			} else {
				if result != nil && len(result.Items) > 0 {
					t.Fatalf("expected no completions, got %d", len(result.Items))
				}
			}
		})
	}
}

func TestAdvancedInlineCompletionProvider_TriggerKind(t *testing.T) {
	provider := &AdvancedInlineCompletionProvider{}

	content := `package main

func main() {
	fmt.P
}`

	position := core.Position{Line: 3, Character: 6}

	tests := []struct {
		name        string
		triggerKind core.InlineCompletionTriggerKind
		expectNil   bool // Conservative mode might return nil
	}{
		{
			name:        "automatic trigger",
			triggerKind: core.InlineCompletionTriggerKindAutomatic,
			expectNil:   true, // Conservative - might not suggest
		},
		{
			name:        "explicit trigger",
			triggerKind: core.InlineCompletionTriggerKindInvoked,
			expectNil:   false, // Aggressive - should suggest
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := core.InlineCompletionContext{
				URI:         "test.go",
				Content:     content,
				Position:    position,
				TriggerKind: tt.triggerKind,
			}

			result := provider.ProvideInlineCompletions(ctx)

			if tt.expectNil && result != nil && len(result.Items) > 0 {
				// It's okay if conservative mode still provides suggestions
				// This test just documents the behavior
				t.Logf("Conservative mode provided %d completions", len(result.Items))
			}

			if !tt.expectNil && (result == nil || len(result.Items) == 0) {
				t.Errorf("Aggressive mode should provide completions")
			}
		})
	}
}

func TestContextAwareInlineCompletionProvider(t *testing.T) {
	provider := &ContextAwareInlineCompletionProvider{}

	tests := []struct {
		name                   string
		content                string
		position               core.Position
		selectedCompletionInfo *core.SelectedCompletionInfo
		wantCompletions        bool
	}{
		{
			name: "with selected completion context",
			content: `package main

func main() {
	fmt.Print
}`,
			position: core.Position{Line: 3, Character: 10},
			selectedCompletionInfo: &core.SelectedCompletionInfo{
				Range: core.Range{
					Start: core.Position{Line: 3, Character: 5},
					End:   core.Position{Line: 3, Character: 10},
				},
				Text: "Print",
			},
			wantCompletions: true,
		},
		{
			name: "without selected completion context",
			content: `package main

func main() {
	fmt.P
}`,
			position:               core.Position{Line: 3, Character: 6},
			selectedCompletionInfo: nil,
			wantCompletions:        true, // Falls back to simple provider
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := core.InlineCompletionContext{
				URI:                    "test.go",
				Content:                tt.content,
				Position:               tt.position,
				TriggerKind:            core.InlineCompletionTriggerKindInvoked,
				SelectedCompletionInfo: tt.selectedCompletionInfo,
			}

			result := provider.ProvideInlineCompletions(ctx)

			hasCompletions := result != nil && len(result.Items) > 0

			if hasCompletions != tt.wantCompletions {
				t.Errorf("expected completions = %v, got %v", tt.wantCompletions, hasCompletions)
			}
		})
	}
}

func containsSubstring(s, substr string) bool {
	return len(substr) > 0 && len(s) > 0 && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		containsInMiddle(s, substr)))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
