package examples

import (
	"fmt"
	"strings"

	"github.com/SCKelemen/lsp/core"
)

// This example shows a complete validator implementation from start to finish:
// 1. Writing the validator
// 2. Writing the code fix provider
// 3. Using them together
// 4. Testing the results

// ===========================
// Part 1: The Validator
// ===========================

// LineLengthValidator checks for lines that exceed a maximum length.
// This is a common code style check in many projects.
type LineLengthValidator struct {
	MaxLength int
}

func NewLineLengthValidator(maxLength int) *LineLengthValidator {
	return &LineLengthValidator{MaxLength: maxLength}
}

func (v *LineLengthValidator) ProvideDiagnostics(uri, content string) []core.Diagnostic {
	var diagnostics []core.Diagnostic

	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		// Get the actual length in UTF-8 bytes
		lineLen := len(line)

		if lineLen > v.MaxLength {
			severity := core.SeverityWarning
			code := core.NewStringCode("line-too-long")

			// Create diagnostic highlighting the excess characters
			diagnostics = append(diagnostics, core.Diagnostic{
				Range: core.Range{
					Start: core.Position{Line: lineNum, Character: v.MaxLength},
					End:   core.Position{Line: lineNum, Character: lineLen},
				},
				Severity: &severity,
				Code:     &code,
				Source:   "line-length",
				Message:  fmt.Sprintf("Line exceeds %d characters (currently %d)", v.MaxLength, lineLen),
			})
		}
	}

	return diagnostics
}

// ===========================
// Part 2: The Code Fix Provider
// ===========================

// LineLengthCodeFixProvider provides fixes for line length issues.
// Note: This is a simplified example. A real implementation would need
// smarter line breaking (respecting words, operators, etc.)
type LineLengthCodeFixProvider struct{}

func (p *LineLengthCodeFixProvider) ProvideCodeFixes(ctx core.CodeFixContext) []core.CodeAction {
	var actions []core.CodeAction

	for _, diag := range ctx.Diagnostics {
		// Only handle line-too-long diagnostics
		if diag.Code != nil && diag.Code.StringValue == "line-too-long" {
			// For this example, we'll just suggest adding a comment
			// A real implementation would do smart line breaking

			lineNum := diag.Range.Start.Line
			lines := strings.Split(ctx.Content, "\n")

			if lineNum < len(lines) {
				line := lines[lineNum]

				// Create an edit that adds a // TODO comment at the start of the line
				edit := core.TextEdit{
					Range: core.Range{
						Start: core.Position{Line: lineNum, Character: 0},
						End:   core.Position{Line: lineNum, Character: 0},
					},
					NewText: "// TODO: Break this long line\n",
				}

				workspaceEdit := &core.WorkspaceEdit{
					Changes: map[string][]core.TextEdit{
						ctx.URI: {edit},
					},
				}

				kind := core.CodeActionKindQuickFix
				actions = append(actions, core.CodeAction{
					Title:       "Add TODO comment for long line",
					Kind:        &kind,
					Edit:        workspaceEdit,
					Diagnostics: []core.Diagnostic{diag},
				})

				// Another option: disable the warning for this line
				disableEdit := core.TextEdit{
					Range: core.Range{
						Start: core.Position{Line: lineNum, Character: len(line)},
						End:   core.Position{Line: lineNum, Character: len(line)},
					},
					NewText: "  // nolint:line-too-long",
				}

				disableWorkspaceEdit := &core.WorkspaceEdit{
					Changes: map[string][]core.TextEdit{
						ctx.URI: {disableEdit},
					},
				}

				actions = append(actions, core.CodeAction{
					Title:       "Disable line length check for this line",
					Kind:        &kind,
					Edit:        disableWorkspaceEdit,
					Diagnostics: []core.Diagnostic{diag},
				})
			}
		}
	}

	return actions
}

// ===========================
// Part 3: Using the Validator
// ===========================

// RunCompleteValidatorExample demonstrates the full validator workflow.
func RunCompleteValidatorExample() {
	fmt.Println("=== Complete Validator Example ===")
	fmt.Println()

	// Sample code with some long lines
	content := `package main

import (
	"fmt"
)

func main() {
	// This is a very long comment that definitely exceeds the maximum line length limit we've set up
	name := "Alice"
	greeting := "Hello, " + name + "!"
	fmt.Println(greeting)

	// Another problematic line: this comment also exceeds our configured maximum and should be flagged by the validator
}
`

	uri := "file:///example.go"

	// Step 1: Create and run the validator
	fmt.Println("Step 1: Running validator...")
	validator := NewLineLengthValidator(80)
	diagnostics := validator.ProvideDiagnostics(uri, content)

	fmt.Printf("Found %d diagnostic(s):\n\n", len(diagnostics))

	for i, diag := range diagnostics {
		fmt.Printf("Diagnostic %d:\n", i+1)
		fmt.Printf("  Location: Line %d, Column %d-%d\n",
			diag.Range.Start.Line+1, // Convert to 1-based for display
			diag.Range.Start.Character+1,
			diag.Range.End.Character+1)
		fmt.Printf("  Severity: %s\n", getSeverityString(diag.Severity))
		fmt.Printf("  Code: %s\n", diag.Code.StringValue)
		fmt.Printf("  Source: %s\n", diag.Source)
		fmt.Printf("  Message: %s\n", diag.Message)
		fmt.Print("\n")
	}

	// Step 2: Get code fixes for the diagnostics
	fmt.Println("Step 2: Getting code fixes...")
	fixProvider := &LineLengthCodeFixProvider{}
	ctx := core.CodeFixContext{
		URI:         uri,
		Content:     content,
		Diagnostics: diagnostics,
	}
	fixes := fixProvider.ProvideCodeFixes(ctx)

	fmt.Printf("Found %d code fix(es):\n\n", len(fixes))

	for i, fix := range fixes {
		fmt.Printf("Fix %d:\n", i+1)
		fmt.Printf("  Title: %s\n", fix.Title)
		fmt.Printf("  Kind: %s\n", *fix.Kind)
		fmt.Printf("  Edits: %d change(s)\n", len(fix.Edit.Changes[uri]))
		fmt.Print("\n")
	}

	// Step 3: Apply the first fix (if available)
	if len(fixes) > 0 {
		fmt.Println("Step 3: Applying first fix...")
		firstFix := fixes[0]

		// Apply the edit (simplified - real implementation would handle multiple files)
		newContent := applyWorkspaceEdit(content, firstFix.Edit)

		fmt.Println("Content after applying fix:")
		fmt.Println("---")
		fmt.Println(newContent)
		fmt.Println("---")
	}

	// Step 4: Show how to use with a registry
	fmt.Println("\nStep 4: Using with a registry...")
	registry := core.NewDiagnosticRegistry()
	registry.Register(validator)
	// You could register more validators here

	allDiagnostics := registry.ProvideDiagnostics(uri, content)
	fmt.Printf("Registry found %d total diagnostic(s)\n", len(allDiagnostics))
}

// ===========================
// Helper Functions
// ===========================

func getSeverityString(severity *core.DiagnosticSeverity) string {
	if severity == nil {
		return "unknown"
	}
	switch *severity {
	case core.SeverityError:
		return "error"
	case core.SeverityWarning:
		return "warning"
	case core.SeverityInformation:
		return "info"
	case core.SeverityHint:
		return "hint"
	default:
		return "unknown"
	}
}

func applyWorkspaceEdit(content string, edit *core.WorkspaceEdit) string {
	// Simplified implementation for single file with single edit
	// Real implementation would:
	// 1. Handle multiple files
	// 2. Sort edits by position (reverse order for correct application)
	// 3. Apply each edit properly accounting for position shifts

	for _, edits := range edit.Changes {
		for _, e := range edits {
			content = applyTextEdit(content, e)
		}
	}
	return content
}

func applyTextEdit(content string, edit core.TextEdit) string {
	// Convert content to lines for easier manipulation
	lines := strings.Split(content, "\n")

	startLine := edit.Range.Start.Line
	startChar := edit.Range.Start.Character
	endLine := edit.Range.End.Line
	endChar := edit.Range.End.Character

	// Handle simple case: single line edit
	if startLine == endLine {
		line := lines[startLine]
		newLine := line[:startChar] + edit.NewText + line[endChar:]
		lines[startLine] = newLine
		return strings.Join(lines, "\n")
	}

	// Multi-line edit (more complex, simplified here)
	// Replace everything from start to end
	before := lines[startLine][:startChar]
	after := lines[endLine][endChar:]
	lines[startLine] = before + edit.NewText + after

	// Remove lines in between
	lines = append(lines[:startLine+1], lines[endLine+1:]...)

	return strings.Join(lines, "\n")
}

// ===========================
// Example CLI Usage
// ===========================

// ExampleCLIUsage shows how to use the validator as a standalone CLI tool.
func ExampleCLIUsage() {
	fmt.Println()
	fmt.Println("=== CLI Usage Example ===")
	fmt.Println()

	content := `package main

func main() {
	veryLongVariableName := "This is a string that contains a lot of characters and will definitely exceed our line limit"
	println(veryLongVariableName)
}
`

	validator := NewLineLengthValidator(100)
	diagnostics := validator.ProvideDiagnostics("file:///main.go", content)

	// Format output like a typical linter
	for _, diag := range diagnostics {
		fmt.Printf("main.go:%d:%d: %s: %s [%s]\n",
			diag.Range.Start.Line+1,
			diag.Range.Start.Character+1,
			getSeverityString(diag.Severity),
			diag.Message,
			diag.Code.StringValue,
		)
	}

	if len(diagnostics) == 0 {
		fmt.Println("No issues found!")
	}
}
