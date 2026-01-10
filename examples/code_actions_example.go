package examples

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/SCKelemen/lsp/core"
)

// UnusedImportProvider provides code actions to remove unused imports.
// This is a simplified example - production code would use proper type checking.
type UnusedImportProvider struct{}

func (p *UnusedImportProvider) ProvideCodeFixes(ctx core.CodeFixContext) []core.CodeAction {
	var actions []core.CodeAction

	// Only for Go files
	if !strings.HasSuffix(ctx.URI, ".go") {
		return nil
	}

	// Find unused imports
	unused := p.findUnusedImports(ctx.Content)
	if len(unused) == 0 {
		return nil
	}

	// Create action to remove all unused imports
	kind := core.CodeActionKindSourceOrganizeImports
	actions = append(actions, core.CodeAction{
		Title:       fmt.Sprintf("Remove %d unused import(s)", len(unused)),
		Kind:        &kind,
		IsPreferred: true,
		Edit:        p.createRemovalEdit(ctx.URI, ctx.Content, unused),
	})

	// Create individual actions for each unused import
	for _, imp := range unused {
		actions = append(actions, p.createSingleRemovalAction(ctx.URI, ctx.Content, imp))
	}

	return actions
}

type importInfo struct {
	Path      string
	StartLine int
	EndLine   int
}

func (p *UnusedImportProvider) findUnusedImports(content string) []importInfo {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", content, parser.ImportsOnly)
	if err != nil {
		return nil
	}

	var unused []importInfo

	for _, imp := range f.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)

		// Simple heuristic: check if package name appears in code
		pkgName := p.getPackageName(imp, importPath)
		if !p.isImportUsed(content, pkgName) {
			pos := fset.Position(imp.Pos())
			end := fset.Position(imp.End())

			unused = append(unused, importInfo{
				Path:      importPath,
				StartLine: pos.Line - 1, // 0-based
				EndLine:   end.Line - 1,
			})
		}
	}

	return unused
}

func (p *UnusedImportProvider) getPackageName(imp *ast.ImportSpec, path string) string {
	if imp.Name != nil {
		return imp.Name.Name
	}
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

func (p *UnusedImportProvider) isImportUsed(content string, pkgName string) bool {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		// Skip import lines
		if strings.Contains(line, "import") {
			continue
		}
		// Check for package usage
		if strings.Contains(line, pkgName+".") {
			return true
		}
	}
	return false
}

func (p *UnusedImportProvider) createRemovalEdit(uri, content string, unused []importInfo) *core.WorkspaceEdit {
	var edits []core.TextEdit

	lines := strings.Split(content, "\n")

	for _, imp := range unused {
		// Remove the entire line including newline
		if imp.StartLine < len(lines) {
			edits = append(edits, core.TextEdit{
				Range: core.Range{
					Start: core.Position{Line: imp.StartLine, Character: 0},
					End:   core.Position{Line: imp.StartLine + 1, Character: 0},
				},
				NewText: "",
			})
		}
	}

	return &core.WorkspaceEdit{
		Changes: map[string][]core.TextEdit{
			uri: edits,
		},
	}
}

func (p *UnusedImportProvider) createSingleRemovalAction(uri, content string, imp importInfo) core.CodeAction {
	kind := core.CodeActionKindQuickFix

	return core.CodeAction{
		Title: fmt.Sprintf("Remove unused import %q", imp.Path),
		Kind:  &kind,
		Edit: &core.WorkspaceEdit{
			Changes: map[string][]core.TextEdit{
				uri: {
					{
						Range: core.Range{
							Start: core.Position{Line: imp.StartLine, Character: 0},
							End:   core.Position{Line: imp.StartLine + 1, Character: 0},
						},
						NewText: "",
					},
				},
			},
		},
	}
}

// QuickFixProvider provides quick fixes for specific diagnostics.
type QuickFixProvider struct{}

func (p *QuickFixProvider) ProvideCodeFixes(ctx core.CodeFixContext) []core.CodeAction {
	var actions []core.CodeAction

	for _, diag := range ctx.Diagnostics {
		if diag.Code == nil {
			continue
		}

		switch diag.Code.StringValue {
		case "unused-var":
			actions = append(actions, p.fixUnusedVar(ctx, diag)...)
		case "missing-return":
			actions = append(actions, p.fixMissingReturn(ctx, diag)...)
		}
	}

	return actions
}

func (p *QuickFixProvider) fixUnusedVar(ctx core.CodeFixContext, diag core.Diagnostic) []core.CodeAction {
	var actions []core.CodeAction

	// Get the variable name
	varName := p.extractText(ctx.Content, diag.Range)

	kind := core.CodeActionKindQuickFix

	// Action 1: Prefix with underscore
	actions = append(actions, core.CodeAction{
		Title:       fmt.Sprintf("Rename to _%s", varName),
		Kind:        &kind,
		Diagnostics: []core.Diagnostic{diag},
		Edit: &core.WorkspaceEdit{
			Changes: map[string][]core.TextEdit{
				ctx.URI: {
					{
						Range:   diag.Range,
						NewText: "_" + varName,
					},
				},
			},
		},
	})

	// Action 2: Delete the line
	lines := strings.Split(ctx.Content, "\n")
	if diag.Range.Start.Line < len(lines) {
		actions = append(actions, core.CodeAction{
			Title:       "Remove unused variable",
			Kind:        &kind,
			Diagnostics: []core.Diagnostic{diag},
			Edit: &core.WorkspaceEdit{
				Changes: map[string][]core.TextEdit{
					ctx.URI: {
						{
							Range: core.Range{
								Start: core.Position{Line: diag.Range.Start.Line, Character: 0},
								End:   core.Position{Line: diag.Range.Start.Line + 1, Character: 0},
							},
							NewText: "",
						},
					},
				},
			},
		})
	}

	return actions
}

func (p *QuickFixProvider) fixMissingReturn(ctx core.CodeFixContext, diag core.Diagnostic) []core.CodeAction {
	kind := core.CodeActionKindQuickFix

	return []core.CodeAction{
		{
			Title:       "Add return statement",
			Kind:        &kind,
			Diagnostics: []core.Diagnostic{diag},
			Edit: &core.WorkspaceEdit{
				Changes: map[string][]core.TextEdit{
					ctx.URI: {
						{
							Range: core.Range{
								Start: diag.Range.End,
								End:   diag.Range.End,
							},
							NewText: "\n\treturn nil",
						},
					},
				},
			},
		},
	}
}

func (p *QuickFixProvider) extractText(content string, r core.Range) string {
	lines := strings.Split(content, "\n")
	if r.Start.Line >= len(lines) {
		return ""
	}

	line := lines[r.Start.Line]
	if r.Start.Line == r.End.Line {
		if r.End.Character <= len(line) {
			return line[r.Start.Character:r.End.Character]
		}
		return line[r.Start.Character:]
	}

	// Multi-line (simplified)
	return line[r.Start.Character:]
}

// RefactorProvider provides refactoring code actions.
type RefactorProvider struct{}

func (p *RefactorProvider) ProvideCodeFixes(ctx core.CodeFixContext) []core.CodeAction {
	// Only provide refactorings for selected ranges
	if ctx.Range.Start == ctx.Range.End {
		return nil
	}

	// Check if refactor actions were requested
	if len(ctx.Only) > 0 {
		hasRefactor := false
		for _, kind := range ctx.Only {
			if strings.HasPrefix(string(kind), "refactor") {
				hasRefactor = true
				break
			}
		}
		if !hasRefactor {
			return nil
		}
	}

	var actions []core.CodeAction

	// Extract to variable
	if p.canExtractVariable(ctx.Content, ctx.Range) {
		kind := core.CodeActionKindRefactorExtract
		actions = append(actions, core.CodeAction{
			Title: "Extract to variable",
			Kind:  &kind,
			Edit:  p.createExtractVariableEdit(ctx.URI, ctx.Content, ctx.Range),
		})
	}

	return actions
}

func (p *RefactorProvider) canExtractVariable(content string, r core.Range) bool {
	// Simple check: range is within a single line and not empty
	return r.Start.Line == r.End.Line && r.End.Character > r.Start.Character
}

func (p *RefactorProvider) createExtractVariableEdit(uri, content string, r core.Range) *core.WorkspaceEdit {
	lines := strings.Split(content, "\n")
	if r.Start.Line >= len(lines) {
		return nil
	}

	line := lines[r.Start.Line]
	if r.End.Character > len(line) {
		return nil
	}

	// Extract selected text
	selected := line[r.Start.Character:r.End.Character]

	// Get indentation
	indent := ""
	for _, ch := range line {
		if ch == ' ' || ch == '\t' {
			indent += string(ch)
		} else {
			break
		}
	}

	// Create new variable declaration
	newVar := fmt.Sprintf("%sextracted := %s\n", indent, selected)

	return &core.WorkspaceEdit{
		Changes: map[string][]core.TextEdit{
			uri: {
				// Insert variable declaration
				{
					Range: core.Range{
						Start: core.Position{Line: r.Start.Line, Character: 0},
						End:   core.Position{Line: r.Start.Line, Character: 0},
					},
					NewText: newVar,
				},
				// Replace selection with variable name
				{
					Range:   r,
					NewText: "extracted",
				},
			},
		},
	}
}
