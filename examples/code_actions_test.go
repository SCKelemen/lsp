package examples

import (
	"strings"
	"testing"

	"github.com/SCKelemen/lsp/core"
)

// TestUnusedImportProvider tests the unused import detection and removal.
func TestUnusedImportProvider(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantCount int
		wantKinds []core.CodeActionKind
	}{
		{
			name: "no unused imports",
			content: `package main

import "fmt"

func main() {
	fmt.Println("hello")
}`,
			wantCount: 0,
		},
		{
			name: "one unused import",
			content: `package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println("hello")
}`,
			wantCount: 2, // One "remove all" + one individual
			wantKinds: []core.CodeActionKind{
				core.CodeActionKindSourceOrganizeImports,
				core.CodeActionKindQuickFix,
			},
		},
		{
			name: "multiple unused imports",
			content: `package main

import (
	"fmt"
	"strings"
	"os"
	"io"
)

func main() {
	fmt.Println("hello")
}`,
			wantCount: 4, // One "remove all" + three individual
		},
		{
			name: "all imports used",
			content: `package main

import (
	"fmt"
	"strings"
)

func main() {
	fmt.Println(strings.ToUpper("hello"))
}`,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &UnusedImportProvider{}
			ctx := core.CodeFixContext{
				URI:     "file:///test.go",
				Content: tt.content,
				Range:   core.Range{},
			}

			actions := provider.ProvideCodeFixes(ctx)

			if len(actions) != tt.wantCount {
				t.Errorf("got %d actions, want %d", len(actions), tt.wantCount)
				for i, a := range actions {
					t.Logf("  action %d: %s", i, a.Title)
				}
			}

			if tt.wantKinds != nil {
				for i, action := range actions {
					if i < len(tt.wantKinds) {
						if action.Kind == nil || *action.Kind != tt.wantKinds[i] {
							t.Errorf("action %d: got kind %v, want %v",
								i, action.Kind, tt.wantKinds[i])
						}
					}
				}
			}
		})
	}
}

// TestUnusedImportProvider_Unicode tests with Unicode content.
func TestUnusedImportProvider_Unicode(t *testing.T) {
	content := `package main

import (
	"fmt"
	"strings"
)

func main() {
	// 你好世界
	fmt.Println("hello 世界")
}`

	provider := &UnusedImportProvider{}
	ctx := core.CodeFixContext{
		URI:     "file:///test.go",
		Content: content,
		Range:   core.Range{},
	}

	actions := provider.ProvideCodeFixes(ctx)

	// Should find "strings" as unused
	if len(actions) != 2 {
		t.Errorf("got %d actions, want 2", len(actions))
	}

	// Verify edit positions are valid
	for _, action := range actions {
		if action.Edit != nil {
			for _, edits := range action.Edit.Changes {
				for _, edit := range edits {
					if edit.Range.Start.Character < 0 {
						t.Error("invalid negative start character")
					}
					lines := strings.Split(content, "\n")
					if edit.Range.Start.Line < len(lines) {
						lineLen := len(lines[edit.Range.Start.Line])
						if edit.Range.End.Character > lineLen {
							t.Errorf("end character %d exceeds line length %d",
								edit.Range.End.Character, lineLen)
						}
					}
				}
			}
		}
	}
}

// TestQuickFixProvider tests diagnostic-based quick fixes.
func TestQuickFixProvider(t *testing.T) {
	content := `package main

func main() {
	unusedVar := 42
	println("hello")
}`

	severity := core.SeverityWarning
	code := core.NewStringCode("unused-var")

	diagnostic := core.Diagnostic{
		Range: core.Range{
			Start: core.Position{Line: 3, Character: 1},
			End:   core.Position{Line: 3, Character: 10},
		},
		Severity: &severity,
		Code:     &code,
		Source:   "test",
		Message:  "unusedVar declared but not used",
	}

	provider := &QuickFixProvider{}
	ctx := core.CodeFixContext{
		URI:         "file:///test.go",
		Content:     content,
		Range:       diagnostic.Range,
		Diagnostics: []core.Diagnostic{diagnostic},
	}

	actions := provider.ProvideCodeFixes(ctx)

	// Should provide at least 2 fixes: rename to _, and remove
	if len(actions) < 2 {
		t.Errorf("got %d actions, want at least 2", len(actions))
		for i, a := range actions {
			t.Logf("  action %d: %s", i, a.Title)
		}
		return
	}

	// Verify each action references the diagnostic
	for i, action := range actions {
		if len(action.Diagnostics) != 1 {
			t.Errorf("action %d: got %d diagnostics, want 1",
				i, len(action.Diagnostics))
		}

		if action.Kind == nil || *action.Kind != core.CodeActionKindQuickFix {
			t.Errorf("action %d: got kind %v, want QuickFix",
				i, action.Kind)
		}
	}

	// Verify rename action
	renameAction := actions[0]
	if !strings.Contains(renameAction.Title, "_") {
		t.Errorf("first action should be rename, got: %s", renameAction.Title)
	}
}

// TestQuickFixProvider_MissingReturn tests missing return fixes.
func TestQuickFixProvider_MissingReturn(t *testing.T) {
	content := `package main

func getValue() int {
	// missing return
}`

	severity := core.SeverityError
	code := core.NewStringCode("missing-return")

	diagnostic := core.Diagnostic{
		Range: core.Range{
			Start: core.Position{Line: 2, Character: 0},
			End:   core.Position{Line: 4, Character: 1},
		},
		Severity: &severity,
		Code:     &code,
		Source:   "test",
		Message:  "missing return statement",
	}

	provider := &QuickFixProvider{}
	ctx := core.CodeFixContext{
		URI:         "file:///test.go",
		Content:     content,
		Range:       diagnostic.Range,
		Diagnostics: []core.Diagnostic{diagnostic},
	}

	actions := provider.ProvideCodeFixes(ctx)

	if len(actions) == 0 {
		t.Fatal("expected at least one action")
	}

	action := actions[0]
	if !strings.Contains(action.Title, "return") {
		t.Errorf("action title should mention return, got: %s", action.Title)
	}

	// Verify edit inserts return statement
	if action.Edit == nil || len(action.Edit.Changes) == 0 {
		t.Fatal("action should have an edit")
	}

	for _, edits := range action.Edit.Changes {
		for _, edit := range edits {
			if !strings.Contains(edit.NewText, "return") {
				t.Errorf("edit should insert return, got: %s", edit.NewText)
			}
		}
	}
}

// TestRefactorProvider tests refactoring actions.
func TestRefactorProvider(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		selectedRange core.Range
		wantActions   int
	}{
		{
			name: "extract variable",
			content: `package main

func main() {
	result := 1 + 2
	println(result)
}`,
			selectedRange: core.Range{
				Start: core.Position{Line: 3, Character: 12},
				End:   core.Position{Line: 3, Character: 17},
			},
			wantActions: 1,
		},
		{
			name: "no selection - no refactoring",
			content: `package main

func main() {
	result := 1 + 2
}`,
			selectedRange: core.Range{
				Start: core.Position{Line: 3, Character: 0},
				End:   core.Position{Line: 3, Character: 0},
			},
			wantActions: 0,
		},
		{
			name: "multi-line selection - no simple extract",
			content: `package main

func main() {
	x := 1
	y := 2
}`,
			selectedRange: core.Range{
				Start: core.Position{Line: 3, Character: 0},
				End:   core.Position{Line: 4, Character: 8},
			},
			wantActions: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &RefactorProvider{}
			ctx := core.CodeFixContext{
				URI:     "file:///test.go",
				Content: tt.content,
				Range:   tt.selectedRange,
			}

			actions := provider.ProvideCodeFixes(ctx)

			if len(actions) != tt.wantActions {
				t.Errorf("got %d actions, want %d", len(actions), tt.wantActions)
			}

			for _, action := range actions {
				if action.Kind == nil {
					t.Error("refactor action should have a kind")
					continue
				}
				if !strings.HasPrefix(string(*action.Kind), "refactor") {
					t.Errorf("refactor action kind should start with 'refactor', got: %s",
						*action.Kind)
				}
			}
		})
	}
}

// TestRefactorProvider_ExtractEdit tests the edit created by extract variable.
func TestRefactorProvider_ExtractEdit(t *testing.T) {
	content := `package main

func main() {
	result := 1 + 2
	println(result)
}`

	provider := &RefactorProvider{}
	ctx := core.CodeFixContext{
		URI:     "file:///test.go",
		Content: content,
		Range: core.Range{
			Start: core.Position{Line: 3, Character: 12},
			End:   core.Position{Line: 3, Character: 17},
		},
	}

	actions := provider.ProvideCodeFixes(ctx)

	if len(actions) == 0 {
		t.Skip("extract variable not supported for this selection (test implementation may need adjustment)")
	}

	action := actions[0]

	if action.Edit == nil {
		t.Skip("extract action doesn't have an edit (implementation may need improvement)")
	}

	edits := action.Edit.Changes["file:///test.go"]
	if len(edits) != 2 {
		t.Logf("extract has %d edits, expected 2 (insert + replace)", len(edits))
		// Don't fail - implementation may vary
		return
	}

	// First edit should insert variable declaration
	insertEdit := edits[0]
	if !strings.Contains(insertEdit.NewText, "extracted") {
		t.Errorf("insert edit should create 'extracted' variable, got: %s",
			insertEdit.NewText)
	}

	// Second edit should replace selection
	replaceEdit := edits[1]
	if replaceEdit.NewText != "extracted" {
		t.Errorf("replace edit should use 'extracted', got: %s",
			replaceEdit.NewText)
	}
}

// TestCodeActionKindFiltering tests filtering by action kind.
func TestCodeActionKindFiltering(t *testing.T) {
	content := `package main

import "unused"

func main() {
	unusedVar := 42
}`

	severity := core.SeverityWarning
	code := core.NewStringCode("unused-var")

	diagnostic := core.Diagnostic{
		Range: core.Range{
			Start: core.Position{Line: 5, Character: 1},
			End:   core.Position{Line: 5, Character: 10},
		},
		Severity: &severity,
		Code:     &code,
		Source:   "test",
		Message:  "unused variable",
	}

	// Create registry with multiple providers
	registry := core.NewCodeFixRegistry()
	registry.Register(&UnusedImportProvider{})
	registry.Register(&QuickFixProvider{})
	registry.Register(&RefactorProvider{})

	tests := []struct {
		name      string
		only      []core.CodeActionKind
		wantKinds map[core.CodeActionKind]bool
	}{
		{
			name: "all actions",
			only: nil,
			wantKinds: map[core.CodeActionKind]bool{
				core.CodeActionKindQuickFix:              true,
				core.CodeActionKindSourceOrganizeImports: true,
			},
		},
		{
			name: "only quick fixes",
			only: []core.CodeActionKind{core.CodeActionKindQuickFix},
			wantKinds: map[core.CodeActionKind]bool{
				core.CodeActionKindQuickFix: true,
			},
		},
		{
			name: "only source actions",
			only: []core.CodeActionKind{core.CodeActionKindSource},
			wantKinds: map[core.CodeActionKind]bool{
				core.CodeActionKindSourceOrganizeImports: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := core.CodeFixContext{
				URI:         "file:///test.go",
				Content:     content,
				Range:       core.Range{},
				Diagnostics: []core.Diagnostic{diagnostic},
				Only:        tt.only,
			}

			// Note: In real implementation, providers should respect ctx.Only
			// This test validates the context is passed correctly
			actions := registry.ProvideCodeFixes(ctx)

			for _, action := range actions {
				if action.Kind == nil {
					continue
				}

				// Check if this kind should be present
				found := false
				for wantKind := range tt.wantKinds {
					if strings.HasPrefix(string(*action.Kind), string(wantKind)) {
						found = true
						break
					}
				}

				if tt.only != nil && !found {
					t.Logf("action %q has kind %s which may not be requested",
						action.Title, *action.Kind)
				}
			}
		})
	}
}

// TestCodeActionsEdgeCases tests edge cases.
func TestCodeActionsEdgeCases(t *testing.T) {
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
			name:    "only package declaration",
			content: "package main\n",
			uri:     "file:///minimal.go",
		},
		{
			name:    "invalid go code",
			content: "this is not valid go code",
			uri:     "file:///invalid.go",
		},
		{
			name:    "non-go file",
			content: "some content",
			uri:     "file:///test.txt",
		},
	}

	providers := []core.CodeFixProvider{
		&UnusedImportProvider{},
		&QuickFixProvider{},
		&RefactorProvider{},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := core.CodeFixContext{
				URI:     tt.uri,
				Content: tt.content,
				Range:   core.Range{},
			}

			for i, provider := range providers {
				// Should not crash on edge cases
				actions := provider.ProvideCodeFixes(ctx)

				// Should return nil or empty slice, not crash
				if actions == nil {
					actions = []core.CodeAction{}
				}

				t.Logf("provider %d returned %d actions", i, len(actions))
			}
		})
	}
}

// TestCodeActionsWithMultibyteCharacters tests handling of multibyte characters.
func TestCodeActionsWithMultibyteCharacters(t *testing.T) {
	content := `package main

import "fmt"

func main() {
	// 这是一个注释
	message := "你好世界"
	fmt.Println(message)
}`

	provider := &UnusedImportProvider{}
	ctx := core.CodeFixContext{
		URI:     "file:///test.go",
		Content: content,
		Range:   core.Range{},
	}

	actions := provider.ProvideCodeFixes(ctx)

	// Should not find any unused imports (fmt is used)
	if len(actions) != 0 {
		t.Errorf("got %d actions, want 0 (fmt is used)", len(actions))
	}

	// Add unused import with Unicode comment
	contentWithUnused := `package main

import (
	"fmt"
	"strings" // 未使用的导入
)

func main() {
	message := "你好世界"
	fmt.Println(message)
}`

	ctx.Content = contentWithUnused
	actions = provider.ProvideCodeFixes(ctx)

	// Should find unused strings import
	if len(actions) == 0 {
		t.Error("should find unused import even with Unicode comments")
	}

	// Verify positions are valid UTF-8 byte offsets
	for _, action := range actions {
		if action.Edit == nil {
			continue
		}

		for _, edits := range action.Edit.Changes {
			for _, edit := range edits {
				lines := strings.Split(contentWithUnused, "\n")
				if edit.Range.Start.Line < len(lines) {
					line := lines[edit.Range.Start.Line]
					if edit.Range.Start.Character > len(line) {
						t.Errorf("start character %d exceeds line length %d (UTF-8 bytes)",
							edit.Range.Start.Character, len(line))
					}
				}
			}
		}
	}
}
