package examples

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strings"

	"github.com/SCKelemen/lsp/core"
)

// TestRunnerCodeLensProvider provides "Run Test" code lenses for test functions.
// This is commonly used in test files to provide quick actions for running tests.
type TestRunnerCodeLensProvider struct{}

func (p *TestRunnerCodeLensProvider) ProvideCodeLenses(ctx core.CodeLensContext) []core.CodeLens {
	if !strings.HasSuffix(ctx.URI, "_test.go") {
		return nil
	}

	var lenses []core.CodeLens

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", ctx.Content, parser.ParseComments)
	if err != nil {
		return nil
	}

	// Find test functions
	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			if strings.HasPrefix(fn.Name.Name, "Test") {
				// Create a code lens at the function declaration
				pos := fset.Position(fn.Pos())
				nameEnd := fset.Position(fn.Name.End())

				command := &core.Command{
					Title:     fmt.Sprintf("▶ Run %s", fn.Name.Name),
					Command:   "go.test.run",
					Arguments: []interface{}{fn.Name.Name},
				}

				lenses = append(lenses, core.CodeLens{
					Range: core.Range{
						Start: core.Position{Line: pos.Line - 1, Character: pos.Column - 1},
						End:   core.Position{Line: nameEnd.Line - 1, Character: nameEnd.Column - 1},
					},
					Command: command,
				})

				// Also add a "Debug Test" lens
				debugCommand := &core.Command{
					Title:     fmt.Sprintf("🐛 Debug %s", fn.Name.Name),
					Command:   "go.test.debug",
					Arguments: []interface{}{fn.Name.Name},
				}

				lenses = append(lenses, core.CodeLens{
					Range: core.Range{
						Start: core.Position{Line: pos.Line - 1, Character: pos.Column - 1},
						End:   core.Position{Line: nameEnd.Line - 1, Character: nameEnd.Column - 1},
					},
					Command: debugCommand,
				})
			}
		}
	}

	return lenses
}

// ReferenceCountCodeLensProvider shows reference counts for symbols.
// This is a simplified example - a real implementation would use a symbol index.
type ReferenceCountCodeLensProvider struct {
	// ReferenceCounter is a function that counts references to a symbol
	ReferenceCounter func(uri, symbolName string) int
}

func (p *ReferenceCountCodeLensProvider) ProvideCodeLenses(ctx core.CodeLensContext) []core.CodeLens {
	if !strings.HasSuffix(ctx.URI, ".go") {
		return nil
	}

	var lenses []core.CodeLens

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", ctx.Content, parser.ParseComments)
	if err != nil {
		return nil
	}

	// Add reference counts for functions and types
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name.Name == "_" || !ast.IsExported(d.Name.Name) {
				continue
			}

			// Count references (simplified - just count occurrences)
			count := strings.Count(ctx.Content, d.Name.Name) - 1 // -1 for declaration

			if p.ReferenceCounter != nil {
				count = p.ReferenceCounter(ctx.URI, d.Name.Name)
			}

			pos := fset.Position(d.Name.Pos())
			nameEnd := fset.Position(d.Name.End())

			command := &core.Command{
				Title:     fmt.Sprintf("%d references", count),
				Command:   "editor.action.showReferences",
				Arguments: []interface{}{ctx.URI, d.Name.Name},
			}

			lenses = append(lenses, core.CodeLens{
				Range: core.Range{
					Start: core.Position{Line: pos.Line - 1, Character: pos.Column - 1},
					End:   core.Position{Line: nameEnd.Line - 1, Character: nameEnd.Column - 1},
				},
				Command: command,
			})

		case *ast.GenDecl:
			for _, spec := range d.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok {
					if ts.Name.Name == "_" || !ast.IsExported(ts.Name.Name) {
						continue
					}

					count := strings.Count(ctx.Content, ts.Name.Name) - 1

					if p.ReferenceCounter != nil {
						count = p.ReferenceCounter(ctx.URI, ts.Name.Name)
					}

					pos := fset.Position(ts.Name.Pos())
					nameEnd := fset.Position(ts.Name.End())

					command := &core.Command{
						Title:     fmt.Sprintf("%d references", count),
						Command:   "editor.action.showReferences",
						Arguments: []interface{}{ctx.URI, ts.Name.Name},
					}

					lenses = append(lenses, core.CodeLens{
						Range: core.Range{
							Start: core.Position{Line: pos.Line - 1, Character: pos.Column - 1},
							End:   core.Position{Line: nameEnd.Line - 1, Character: nameEnd.Column - 1},
						},
						Command: command,
					})
				}
			}
		}
	}

	return lenses
}

// TODOCodeLensProvider shows actionable items for TODO comments.
// This helps developers track and manage TODO items in code.
type TODOCodeLensProvider struct{}

func (p *TODOCodeLensProvider) ProvideCodeLenses(ctx core.CodeLensContext) []core.CodeLens {
	var lenses []core.CodeLens

	// Find TODO comments
	todoRegex := regexp.MustCompile(`//\s*(TODO|FIXME|HACK|XXX)(?:\(([^)]+)\))?:\s*(.+)`)
	matches := todoRegex.FindAllStringSubmatchIndex(ctx.Content, -1)

	for _, match := range matches {
		lineStart := match[0]
		lineEnd := match[1]

		// Extract the TODO type, author, and message
		todoType := ctx.Content[match[2]:match[3]]  // TODO, FIXME, etc.
		var author string
		if match[4] != -1 && match[5] != -1 {
			author = ctx.Content[match[4]:match[5]]
		}
		message := ctx.Content[match[6]:match[7]]

		startPos := core.ByteOffsetToPosition(ctx.Content, lineStart)
		endPos := core.ByteOffsetToPosition(ctx.Content, lineEnd)

		// Create command based on TODO type
		var title string
		if author != "" {
			title = fmt.Sprintf("📝 %s by %s: %s", todoType, author, truncate(message, 50))
		} else {
			title = fmt.Sprintf("📝 %s: %s", todoType, truncate(message, 50))
		}

		command := &core.Command{
			Title:   title,
			Command: "todo.show",
			Arguments: []interface{}{
				map[string]interface{}{
					"type":    todoType,
					"author":  author,
					"message": message,
					"file":    ctx.URI,
					"line":    startPos.Line,
				},
			},
		}

		lenses = append(lenses, core.CodeLens{
			Range: core.Range{
				Start: startPos,
				End:   endPos,
			},
			Command: command,
		})
	}

	return lenses
}

// LazyCodeLensProvider demonstrates lazy resolution of code lenses.
// The initial code lens has no command, and it's resolved later.
type LazyCodeLensProvider struct {
	// ResolveFunc is called to resolve the command for a code lens
	ResolveFunc func(lens core.CodeLens) core.CodeLens
}

func (p *LazyCodeLensProvider) ProvideCodeLenses(ctx core.CodeLensContext) []core.CodeLens {
	if !strings.HasSuffix(ctx.URI, ".go") {
		return nil
	}

	var lenses []core.CodeLens

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", ctx.Content, parser.ParseComments)
	if err != nil {
		return nil
	}

	// Add unresolved code lenses for exported functions
	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			if ast.IsExported(fn.Name.Name) {
				pos := fset.Position(fn.Name.Pos())
				nameEnd := fset.Position(fn.Name.End())

				// Create lens without command (will be resolved later)
				lenses = append(lenses, core.CodeLens{
					Range: core.Range{
						Start: core.Position{Line: pos.Line - 1, Character: pos.Column - 1},
						End:   core.Position{Line: nameEnd.Line - 1, Character: nameEnd.Column - 1},
					},
					Command: nil, // No command yet
					Data: map[string]interface{}{
						"uri":      ctx.URI,
						"function": fn.Name.Name,
						"exported": true,
					},
				})
			}
		}
	}

	return lenses
}

func (p *LazyCodeLensProvider) ResolveCodeLens(lens core.CodeLens) core.CodeLens {
	if p.ResolveFunc != nil {
		return p.ResolveFunc(lens)
	}

	// Default resolution: add a generic command
	if data, ok := lens.Data.(map[string]interface{}); ok {
		if funcName, ok := data["function"].(string); ok {
			lens.Command = &core.Command{
				Title:     fmt.Sprintf("📊 Analyze %s", funcName),
				Command:   "code.analyze",
				Arguments: []interface{}{funcName},
			}
		}
	}

	return lens
}

// CompositeCodeLensProvider combines multiple code lens providers.
type CompositeCodeLensProvider struct {
	Providers []core.CodeLensProvider
}

func NewCompositeCodeLensProvider(providers ...core.CodeLensProvider) *CompositeCodeLensProvider {
	return &CompositeCodeLensProvider{
		Providers: providers,
	}
}

func (p *CompositeCodeLensProvider) ProvideCodeLenses(ctx core.CodeLensContext) []core.CodeLens {
	var allLenses []core.CodeLens

	for _, provider := range p.Providers {
		lenses := provider.ProvideCodeLenses(ctx)
		if lenses != nil {
			allLenses = append(allLenses, lenses...)
		}
	}

	return allLenses
}

// truncate truncates a string to a maximum length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Example usage in CLI tool
func CLICodeLensExample() {
	content := `package main

import "testing"

// TODO(alice): Implement better error handling
func TestCalculateSum(t *testing.T) {
	result := CalculateSum(2, 3)
	if result != 5 {
		t.Errorf("expected 5, got %d", result)
	}
}

// FIXME: This test is flaky
func TestDivide(t *testing.T) {
	// Test code here
}

func CalculateSum(a, b int) int {
	// TODO: Add validation
	return a + b
}
`

	provider := NewCompositeCodeLensProvider(
		&TestRunnerCodeLensProvider{},
		&TODOCodeLensProvider{},
	)

	ctx := core.CodeLensContext{
		URI:     "file:///test_test.go",
		Content: content,
	}

	lenses := provider.ProvideCodeLenses(ctx)

	println("Found", len(lenses), "code lenses:")
	for i, lens := range lenses {
		println(fmt.Sprintf("  %d. %s at %s", i+1, lens.Command.Title, lens.Range.String()))
	}
}

// Example usage in LSP server
// func (s *Server) TextDocumentCodeLens(
// 	ctx *glsp.Context,
// 	params *protocol.CodeLensParams,
// ) ([]protocol.CodeLens, error) {
// 	uri := string(params.TextDocument.URI)
// 	content := s.documents.GetContent(uri)
//
// 	// Use provider with core types
// 	coreCtx := core.CodeLensContext{
// 		URI:     uri,
// 		Content: content,
// 	}
//
// 	coreLenses := s.codeLensProvider.ProvideCodeLenses(coreCtx)
//
// 	// Convert back to protocol
// 	return adapter_3_16.CoreToProtocolCodeLenses(coreLenses), nil
// }
//
// func (s *Server) CodeLensResolve(
// 	ctx *glsp.Context,
// 	params *protocol.CodeLens,
// ) (*protocol.CodeLens, error) {
// 	// Convert protocol to core
// 	coreLens := adapter_3_16.ProtocolToCoreCodeLens(*params)
//
// 	// Resolve with provider
// 	resolved := s.codeLensResolver.ResolveCodeLens(coreLens)
//
// 	// Convert back to protocol
// 	protocolLens := adapter_3_16.CoreToProtocolCodeLens(resolved)
// 	return &protocolLens, nil
// }
