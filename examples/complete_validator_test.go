package examples

import (
	"strings"
	"testing"

	"github.com/SCKelemen/lsp/core"
)

// TestLineLengthValidator tests the LineLengthValidator.
func TestLineLengthValidator(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		maxLength     int
		wantCount     int
		wantLine      int
		wantMessage   string
	}{
		{
			name:      "no issues - short lines",
			content:   "package main\n\nfunc main() {\n}\n",
			maxLength: 80,
			wantCount: 0,
		},
		{
			name:        "one long line",
			content:     "package main\n\n// This is a very long comment that definitely exceeds our maximum character limit\nfunc main() {\n}\n",
			maxLength:   80,
			wantCount:   1,
			wantLine:    2,
			wantMessage: "Line exceeds 80 characters",
		},
		{
			name:      "multiple long lines",
			content:   "// Line 1 is very long and exceeds the limit set by our configuration\n// Line 2 is also very long and exceeds the limit set by our configuration\nfunc main() {\n}\n",
			maxLength: 60,
			wantCount: 2,
		},
		{
			name:      "exactly at limit - should not trigger",
			content:   strings.Repeat("a", 80) + "\n",
			maxLength: 80,
			wantCount: 0,
		},
		{
			name:      "one character over limit",
			content:   strings.Repeat("a", 81) + "\n",
			maxLength: 80,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewLineLengthValidator(tt.maxLength)
			diagnostics := validator.ProvideDiagnostics("file:///test.go", tt.content)

			if len(diagnostics) != tt.wantCount {
				t.Errorf("got %d diagnostics, want %d", len(diagnostics), tt.wantCount)
				for i, d := range diagnostics {
					t.Logf("  diagnostic %d: line %d, message: %s", i, d.Range.Start.Line, d.Message)
				}
			}

			if tt.wantCount > 0 {
				first := diagnostics[0]

				if first.Range.Start.Line != tt.wantLine {
					t.Errorf("first diagnostic at line %d, want line %d",
						first.Range.Start.Line, tt.wantLine)
				}

				if tt.wantMessage != "" && !strings.Contains(first.Message, tt.wantMessage) {
					t.Errorf("message %q does not contain %q", first.Message, tt.wantMessage)
				}

				// Verify diagnostic properties
				if first.Source != "line-length" {
					t.Errorf("source = %q, want %q", first.Source, "line-length")
				}

				if first.Code == nil || first.Code.StringValue != "line-too-long" {
					t.Errorf("code = %v, want 'line-too-long'", first.Code)
				}

				if first.Severity == nil || *first.Severity != core.SeverityWarning {
					t.Errorf("severity = %v, want Warning", first.Severity)
				}
			}
		})
	}
}

// TestLineLengthValidator_Unicode tests the validator with Unicode content.
func TestLineLengthValidator_Unicode(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		maxLength int
		wantCount int
	}{
		{
			name:      "Chinese characters within limit",
			content:   "// 你好世界",
			maxLength: 50,
			wantCount: 0,
		},
		{
			name:      "Chinese characters exceed limit",
			content:   "// " + strings.Repeat("你好世界", 20), // 80 characters (20 * 4 bytes each)
			maxLength: 50,
			wantCount: 1,
		},
		{
			name:      "Emoji within limit",
			content:   "// Hello 😀 World",
			maxLength: 50,
			wantCount: 0,
		},
		{
			name:      "Mixed ASCII and Unicode",
			content:   "// Hello 世界 " + strings.Repeat("x", 100),
			maxLength: 80,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewLineLengthValidator(tt.maxLength)
			diagnostics := validator.ProvideDiagnostics("file:///test.go", tt.content)

			if len(diagnostics) != tt.wantCount {
				t.Errorf("got %d diagnostics, want %d", len(diagnostics), tt.wantCount)
			}

			// Verify positions are at valid UTF-8 boundaries
			for _, diag := range diagnostics {
				if diag.Range.Start.Character < 0 || diag.Range.Start.Character > len(tt.content) {
					t.Errorf("invalid start position: %d (content length: %d)",
						diag.Range.Start.Character, len(tt.content))
				}
				if diag.Range.End.Character < 0 || diag.Range.End.Character > len(tt.content) {
					t.Errorf("invalid end position: %d (content length: %d)",
						diag.Range.End.Character, len(tt.content))
				}
			}
		})
	}
}

// TestLineLengthCodeFixProvider tests the code fix provider.
func TestLineLengthCodeFixProvider(t *testing.T) {
	content := "package main\n\n// This is a very long comment that definitely exceeds our maximum line length limit\n"
	uri := "file:///test.go"

	// Get diagnostics first
	validator := NewLineLengthValidator(80)
	diagnostics := validator.ProvideDiagnostics(uri, content)

	if len(diagnostics) == 0 {
		t.Fatal("expected at least one diagnostic")
	}

	// Get code fixes
	provider := &LineLengthCodeFixProvider{}
	ctx := core.CodeFixContext{
		URI:         uri,
		Content:     content,
		Diagnostics: diagnostics,
	}
	fixes := provider.ProvideCodeFixes(ctx)

	// Should provide at least one fix
	if len(fixes) == 0 {
		t.Fatal("expected at least one code fix")
	}

	// Verify first fix
	fix := fixes[0]

	if fix.Title == "" {
		t.Error("fix should have a title")
	}

	if fix.Kind == nil || *fix.Kind != core.CodeActionKindQuickFix {
		t.Errorf("fix kind = %v, want QuickFix", fix.Kind)
	}

	if fix.Edit == nil {
		t.Fatal("fix should have an edit")
	}

	if len(fix.Edit.Changes) == 0 {
		t.Error("fix edit should have changes")
	}

	// Verify fix is associated with diagnostic
	if len(fix.Diagnostics) != 1 {
		t.Errorf("fix should be associated with 1 diagnostic, got %d", len(fix.Diagnostics))
	}
}

// TestApplyTextEdit tests the text edit application logic.
func TestApplyTextEdit(t *testing.T) {
	tests := []struct {
		name    string
		content string
		edit    core.TextEdit
		want    string
	}{
		{
			name:    "insert at start of line",
			content: "Hello world\n",
			edit: core.TextEdit{
				Range: core.Range{
					Start: core.Position{Line: 0, Character: 0},
					End:   core.Position{Line: 0, Character: 0},
				},
				NewText: "// ",
			},
			want: "// Hello world\n",
		},
		{
			name:    "insert at end of line",
			content: "Hello world\n",
			edit: core.TextEdit{
				Range: core.Range{
					Start: core.Position{Line: 0, Character: 11},
					End:   core.Position{Line: 0, Character: 11},
				},
				NewText: "!",
			},
			want: "Hello world!\n",
		},
		{
			name:    "replace text in middle",
			content: "Hello world\n",
			edit: core.TextEdit{
				Range: core.Range{
					Start: core.Position{Line: 0, Character: 6},
					End:   core.Position{Line: 0, Character: 11},
				},
				NewText: "there",
			},
			want: "Hello there\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyTextEdit(tt.content, tt.edit)
			if got != tt.want {
				t.Errorf("applyTextEdit() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestRegistryIntegration tests using the validator with a registry.
func TestRegistryIntegration(t *testing.T) {
	content := "package main\n\n// This is a very long comment that exceeds the limit\n"
	uri := "file:///test.go"

	// Create registry and add validator
	registry := core.NewDiagnosticRegistry()
	registry.Register(NewLineLengthValidator(50))

	// Get diagnostics
	diagnostics := registry.ProvideDiagnostics(uri, content)

	if len(diagnostics) == 0 {
		t.Error("registry should return diagnostics from registered validators")
	}

	// Verify diagnostic has correct properties
	for _, diag := range diagnostics {
		if diag.Source != "line-length" {
			t.Errorf("unexpected diagnostic source: %s", diag.Source)
		}
	}
}

// BenchmarkLineLengthValidator benchmarks the validator performance.
func BenchmarkLineLengthValidator(b *testing.B) {
	// Generate test content with various line lengths
	var content strings.Builder
	for i := 0; i < 100; i++ {
		if i%3 == 0 {
			// Long line
			content.WriteString(strings.Repeat("x", 120))
		} else {
			// Normal line
			content.WriteString(strings.Repeat("x", 60))
		}
		content.WriteString("\n")
	}

	validator := NewLineLengthValidator(80)
	uri := "file:///test.go"
	testContent := content.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ProvideDiagnostics(uri, testContent)
	}
}
