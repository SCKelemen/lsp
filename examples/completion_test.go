package examples

import (
	"strings"
	"testing"

	"github.com/SCKelemen/lsp/core"
)

// TestKeywordCompletionProvider tests keyword completions.
func TestKeywordCompletionProvider(t *testing.T) {
	provider := NewGoKeywordCompletionProvider()

	tests := []struct {
		name           string
		content        string
		position       core.Position
		wantNil        bool
		wantMinItems   int
		checkItems     func(t *testing.T, items []core.CompletionItem)
	}{
		{
			name:         "no prefix returns all keywords",
			content:      "package main\n\n",
			position:     core.Position{Line: 2, Character: 0},
			wantMinItems: 25, // All Go keywords
		},
		{
			name:         "prefix 'f' filters keywords",
			content:      "f",
			position:     core.Position{Line: 0, Character: 1},
			wantMinItems: 1, // At least "for", "func", "fallthrough"
			checkItems: func(t *testing.T, items []core.CompletionItem) {
				for _, item := range items {
					if !strings.HasPrefix(item.Label, "f") {
						t.Errorf("expected items starting with 'f', got %q", item.Label)
					}
				}
			},
		},
		{
			name:         "prefix 'for' matches",
			content:      "for",
			position:     core.Position{Line: 0, Character: 3},
			wantMinItems: 1,
			checkItems: func(t *testing.T, items []core.CompletionItem) {
				found := false
				for _, item := range items {
					if item.Label == "for" {
						found = true
						if item.Kind == nil || *item.Kind != core.CompletionItemKindKeyword {
							t.Error("expected keyword kind")
						}
					}
				}
				if !found {
					t.Error("expected to find 'for' keyword")
				}
			},
		},
		{
			name:         "prefix 'xyz' matches nothing",
			content:      "xyz",
			position:     core.Position{Line: 0, Character: 3},
			wantNil:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := core.CompletionContext{
				URI:         "file:///test.go",
				Content:     tt.content,
				Position:    tt.position,
				TriggerKind: core.CompletionTriggerKindInvoked,
			}

			list := provider.ProvideCompletions(ctx)

			if tt.wantNil {
				if list != nil {
					t.Errorf("expected nil, got %d items", len(list.Items))
				}
				return
			}

			if list == nil {
				t.Fatal("expected completion list, got nil")
			}

			if len(list.Items) < tt.wantMinItems {
				t.Errorf("got %d items, want at least %d", len(list.Items), tt.wantMinItems)
			}

			if tt.checkItems != nil {
				tt.checkItems(t, list.Items)
			}
		})
	}
}

// TestSnippetCompletionProvider tests snippet completions.
func TestSnippetCompletionProvider(t *testing.T) {
	provider := NewGoSnippetProvider()

	tests := []struct {
		name         string
		content      string
		position     core.Position
		wantNil      bool
		wantMinItems int
		checkItems   func(t *testing.T, items []core.CompletionItem)
	}{
		{
			name:         "no prefix returns all snippets",
			content:      "",
			position:     core.Position{Line: 0, Character: 0},
			wantMinItems: 5, // All defined snippets
		},
		{
			name:         "prefix 'for' filters snippets",
			content:      "for",
			position:     core.Position{Line: 0, Character: 3},
			wantMinItems: 2, // "for" and "forr"
			checkItems: func(t *testing.T, items []core.CompletionItem) {
				foundFor := false
				foundForr := false
				for _, item := range items {
					if item.Kind == nil || *item.Kind != core.CompletionItemKindSnippet {
						t.Error("expected snippet kind")
					}
					if item.InsertTextFormat == nil || *item.InsertTextFormat != core.InsertTextFormatSnippet {
						t.Error("expected snippet insert format")
					}
					if strings.Contains(item.Label, "for loop") {
						foundFor = true
					}
					if strings.Contains(item.Label, "for range") {
						foundForr = true
					}
				}
				if !foundFor || !foundForr {
					t.Error("expected to find both 'for' and 'forr' snippets")
				}
			},
		},
		{
			name:         "snippet contains placeholders",
			content:      "if",
			position:     core.Position{Line: 0, Character: 2},
			wantMinItems: 1,
			checkItems: func(t *testing.T, items []core.CompletionItem) {
				for _, item := range items {
					if strings.Contains(item.Label, "if") {
						if !strings.Contains(item.InsertText, "${") {
							t.Error("expected snippet to contain placeholders")
						}
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := core.CompletionContext{
				URI:         "file:///test.go",
				Content:     tt.content,
				Position:    tt.position,
				TriggerKind: core.CompletionTriggerKindInvoked,
			}

			list := provider.ProvideCompletions(ctx)

			if tt.wantNil {
				if list != nil {
					t.Errorf("expected nil, got %d items", len(list.Items))
				}
				return
			}

			if list == nil {
				t.Fatal("expected completion list, got nil")
			}

			if len(list.Items) < tt.wantMinItems {
				t.Errorf("got %d items, want at least %d", len(list.Items), tt.wantMinItems)
			}

			if tt.checkItems != nil {
				tt.checkItems(t, list.Items)
			}
		})
	}
}

// TestSymbolCompletionProvider tests symbol-based completions.
func TestSymbolCompletionProvider(t *testing.T) {
	provider := &SymbolCompletionProvider{}

	tests := []struct {
		name         string
		content      string
		position     core.Position
		wantNil      bool
		wantMinItems int
		checkItems   func(t *testing.T, items []core.CompletionItem)
	}{
		{
			name: "completes function names",
			content: `package main

func CalculateSum(a, b int) int {
	return a + b
}

func main() {
	Cal
}`,
			position:     core.Position{Line: 7, Character: 5}, // After "Cal"
			wantMinItems: 1,
			checkItems: func(t *testing.T, items []core.CompletionItem) {
				found := false
				for _, item := range items {
					if item.Label == "CalculateSum" {
						found = true
						if item.Kind == nil || *item.Kind != core.CompletionItemKindFunction {
							t.Error("expected function kind")
						}
					}
				}
				if !found {
					t.Error("expected to find 'CalculateSum'")
				}
			},
		},
		{
			name: "completes type names",
			content: `package main

type Config struct {
	Value int
}

func main() {
	var c Co
}`,
			position:     core.Position{Line: 7, Character: 10}, // After "Co"
			wantMinItems: 1,
			checkItems: func(t *testing.T, items []core.CompletionItem) {
				found := false
				for _, item := range items {
					if item.Label == "Config" {
						found = true
						if item.Kind == nil || *item.Kind != core.CompletionItemKindStruct {
							t.Error("expected struct kind")
						}
					}
				}
				if !found {
					t.Error("expected to find 'Config'")
				}
			},
		},
		{
			name: "completes variable names",
			content: `package main

var myVariable = 42

func main() {
	println(my)
}`,
			position:     core.Position{Line: 5, Character: 10}, // After "my" in println
			wantMinItems: 1, // myVariable
			checkItems: func(t *testing.T, items []core.CompletionItem) {
				found := false
				for _, item := range items {
					if item.Label == "myVariable" {
						found = true
						if item.Kind == nil || *item.Kind != core.CompletionItemKindVariable {
							t.Error("expected variable kind for myVariable")
						}
					}
				}
				if !found {
					t.Error("expected to find 'myVariable'")
				}
			},
		},
		{
			name: "case insensitive matching",
			content: `package main

func MyFunction() {}

func main() {
	myf
}`,
			position:     core.Position{Line: 5, Character: 5}, // After "myf"
			wantMinItems: 1,
			checkItems: func(t *testing.T, items []core.CompletionItem) {
				found := false
				for _, item := range items {
					if item.Label == "MyFunction" {
						found = true
					}
				}
				if !found {
					t.Error("expected case-insensitive match for 'MyFunction'")
				}
			},
		},
		{
			name:     "non-Go file returns nil",
			content:  "function test() {}",
			position: core.Position{Line: 0, Character: 5},
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri := "file:///test.go"
			if tt.wantNil && tt.name == "non-Go file returns nil" {
				uri = "file:///test.js"
			}

			ctx := core.CompletionContext{
				URI:         uri,
				Content:     tt.content,
				Position:    tt.position,
				TriggerKind: core.CompletionTriggerKindInvoked,
			}

			list := provider.ProvideCompletions(ctx)

			if tt.wantNil {
				if list != nil {
					t.Errorf("expected nil, got %d items", len(list.Items))
				}
				return
			}

			if list == nil {
				t.Fatal("expected completion list, got nil")
			}

			if len(list.Items) < tt.wantMinItems {
				t.Errorf("got %d items, want at least %d", len(list.Items), tt.wantMinItems)
			}

			if tt.checkItems != nil {
				tt.checkItems(t, list.Items)
			}
		})
	}
}

// TestImportCompletionProvider tests import completions.
func TestImportCompletionProvider(t *testing.T) {
	provider := NewGoImportCompletionProvider()

	tests := []struct {
		name         string
		content      string
		position     core.Position
		wantNil      bool
		wantMinItems int
		checkItems   func(t *testing.T, items []core.CompletionItem)
	}{
		{
			name: "completes inside import block",
			content: `package main

import (
	"f
)`,
			position:     core.Position{Line: 3, Character: 3}, // After "f
			wantMinItems: 1,
			checkItems: func(t *testing.T, items []core.CompletionItem) {
				foundFmt := false
				for _, item := range items {
					if item.Label == "fmt" {
						foundFmt = true
						if item.Kind == nil || *item.Kind != core.CompletionItemKindModule {
							t.Error("expected module kind")
						}
					}
				}
				if !foundFmt {
					t.Error("expected to find 'fmt' package")
				}
			},
		},
		{
			name: "completes with partial path",
			content: `package main

import (
	"path/
)`,
			position:     core.Position{Line: 3, Character: 7}, // After "path/
			wantMinItems: 1,
			checkItems: func(t *testing.T, items []core.CompletionItem) {
				found := false
				for _, item := range items {
					if strings.Contains(item.Label, "filepath") {
						found = true
					}
				}
				if !found {
					t.Error("expected to find package with 'filepath'")
				}
			},
		},
		{
			name: "list is incomplete for imports",
			content: `package main

import (
	"
)`,
			position:     core.Position{Line: 3, Character: 2},
			wantMinItems: 1,
			checkItems: func(t *testing.T, items []core.CompletionItem) {
				// Import lists should be marked as incomplete
				// (checked in the list, not individual items)
			},
		},
		{
			name:     "no completion outside import",
			content:  "package main\n\nfunc main() {\n\tf\n}",
			position: core.Position{Line: 3, Character: 3},
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := core.CompletionContext{
				URI:         "file:///test.go",
				Content:     tt.content,
				Position:    tt.position,
				TriggerKind: core.CompletionTriggerKindInvoked,
			}

			list := provider.ProvideCompletions(ctx)

			if tt.wantNil {
				if list != nil {
					t.Errorf("expected nil, got %d items", len(list.Items))
				}
				return
			}

			if list == nil {
				t.Fatal("expected completion list, got nil")
			}

			if len(list.Items) < tt.wantMinItems {
				t.Errorf("got %d items, want at least %d", len(list.Items), tt.wantMinItems)
			}

			// Import lists should be marked as incomplete
			if tt.name == "list is incomplete for imports" && !list.IsIncomplete {
				t.Error("expected IsIncomplete to be true for import completions")
			}

			if tt.checkItems != nil {
				tt.checkItems(t, list.Items)
			}
		})
	}
}

// TestLazyCompletionProvider tests lazy resolution.
func TestLazyCompletionProvider(t *testing.T) {
	baseProvider := NewGoKeywordCompletionProvider()

	tests := []struct {
		name          string
		resolveFunc   func(item core.CompletionItem) core.CompletionItem
		checkResolved func(t *testing.T, resolved core.CompletionItem)
	}{
		{
			name: "default resolution adds documentation",
			checkResolved: func(t *testing.T, resolved core.CompletionItem) {
				if resolved.Documentation == "" {
					t.Error("expected documentation after resolution")
				}
				if resolved.Detail == "" {
					t.Error("expected detail after resolution")
				}
			},
		},
		{
			name: "custom resolution function",
			resolveFunc: func(item core.CompletionItem) core.CompletionItem {
				item.Documentation = "Custom documentation"
				item.Detail = "Custom detail"
				return item
			},
			checkResolved: func(t *testing.T, resolved core.CompletionItem) {
				if resolved.Documentation != "Custom documentation" {
					t.Errorf("expected custom documentation, got %q", resolved.Documentation)
				}
				if resolved.Detail != "Custom detail" {
					t.Errorf("expected custom detail, got %q", resolved.Detail)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &LazyCompletionProvider{
				BaseProvider: baseProvider,
				ResolveFunc:  tt.resolveFunc,
			}

			ctx := core.CompletionContext{
				URI:         "file:///test.go",
				Content:     "f",
				Position:    core.Position{Line: 0, Character: 1},
				TriggerKind: core.CompletionTriggerKindInvoked,
			}

			list := provider.ProvideCompletions(ctx)

			if list == nil || len(list.Items) == 0 {
				t.Fatal("expected completion items")
			}

			// Check that initial items have no documentation
			firstItem := list.Items[0]
			if firstItem.Documentation != "" {
				t.Error("expected empty documentation before resolution")
			}
			if firstItem.Detail != "" {
				t.Error("expected empty detail before resolution")
			}
			if firstItem.Data == nil {
				t.Error("expected data to be set")
			}

			// Resolve the item
			resolved := provider.ResolveCompletionItem(firstItem)

			if tt.checkResolved != nil {
				tt.checkResolved(t, resolved)
			}
		})
	}
}

// TestCompositeCompletionProvider tests combining multiple providers.
func TestCompositeCompletionProvider(t *testing.T) {
	content := `package main

func CalculateSum(a, b int) int {
	return a + b
}

func main() {
	fo
}
`

	provider := NewCompositeCompletionProvider(
		NewGoKeywordCompletionProvider(),
		NewGoSnippetProvider(),
		&SymbolCompletionProvider{},
	)

	ctx := core.CompletionContext{
		URI:         "file:///test.go",
		Content:     content,
		Position:    core.Position{Line: 7, Character: 4}, // After "fo"
		TriggerKind: core.CompletionTriggerKindInvoked,
	}

	list := provider.ProvideCompletions(ctx)

	if list == nil {
		t.Fatal("expected completion list")
	}

	// Should have completions from multiple providers:
	// - Keywords: "for"
	// - Snippets: "for loop", "for range"
	// - Symbols: (none match "fo")
	if len(list.Items) < 3 {
		t.Errorf("got %d items, want at least 3", len(list.Items))
	}

	// Check that we have items from different providers
	hasKeyword := false
	hasSnippet := false

	for _, item := range list.Items {
		if item.Kind != nil {
			if *item.Kind == core.CompletionItemKindKeyword {
				hasKeyword = true
			}
			if *item.Kind == core.CompletionItemKindSnippet {
				hasSnippet = true
			}
		}
	}

	if !hasKeyword {
		t.Error("expected keyword completion")
	}
	if !hasSnippet {
		t.Error("expected snippet completion")
	}
}

// TestCompletion_EdgeCases tests edge cases for completions.
func TestCompletion_EdgeCases(t *testing.T) {
	provider := NewGoKeywordCompletionProvider()

	tests := []struct {
		name     string
		content  string
		position core.Position
		test     func(t *testing.T, list *core.CompletionList)
	}{
		{
			name:     "empty file",
			content:  "",
			position: core.Position{Line: 0, Character: 0},
			test: func(t *testing.T, list *core.CompletionList) {
				if list == nil {
					t.Error("expected completions for empty file")
				}
			},
		},
		{
			name:     "position out of bounds",
			content:  "package main",
			position: core.Position{Line: 10, Character: 0},
			test: func(t *testing.T, list *core.CompletionList) {
				if list != nil {
					t.Error("expected nil for out of bounds position")
				}
			},
		},
		{
			name:     "trigger character completion",
			content:  "obj.",
			position: core.Position{Line: 0, Character: 4},
			test: func(t *testing.T, list *core.CompletionList) {
				// Keyword provider doesn't handle dot trigger,
				// but should return nil gracefully
				if list != nil {
					t.Log("Keyword provider returned completions (may be valid)")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := core.CompletionContext{
				URI:         "file:///test.go",
				Content:     tt.content,
				Position:    tt.position,
				TriggerKind: core.CompletionTriggerKindInvoked,
			}

			list := provider.ProvideCompletions(ctx)
			tt.test(t, list)
		})
	}
}
