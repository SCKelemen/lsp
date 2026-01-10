package examples

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/SCKelemen/lsp/core"
)

// GoParameterNameInlayHintsProvider provides parameter name hints for function calls.
// This shows the parameter names inline with function call arguments.
type GoParameterNameInlayHintsProvider struct{}

func (p *GoParameterNameInlayHintsProvider) ProvideInlayHints(uri, content string, rng core.Range) []core.InlayHint {
	if !strings.HasSuffix(uri, ".go") {
		return nil
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return nil
	}

	var hints []core.InlayHint

	// Find all function calls within the range
	ast.Inspect(f, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			hints = append(hints, p.hintsForCall(call, fset, f, content, rng)...)
		}
		return true
	})

	return hints
}

func (p *GoParameterNameInlayHintsProvider) hintsForCall(call *ast.CallExpr, fset *token.FileSet, f *ast.File, content string, rng core.Range) []core.InlayHint {
	var hints []core.InlayHint

	// Get the function being called
	var funcDecl *ast.FuncDecl
	if ident, ok := call.Fun.(*ast.Ident); ok {
		funcDecl = p.findFuncDecl(f, ident.Name)
	}

	if funcDecl == nil || funcDecl.Type.Params == nil {
		return nil
	}

	// Build a flat list of parameter names from the parameter list
	var paramNames []string
	for _, field := range funcDecl.Type.Params.List {
		if len(field.Names) == 0 {
			// Unnamed parameter
			paramNames = append(paramNames, "")
		} else {
			// Named parameters - can be multiple with same type (e.g., a, b int)
			for _, name := range field.Names {
				paramNames = append(paramNames, name.Name)
			}
		}
	}

	// Match arguments to parameter names
	for argIdx, arg := range call.Args {
		if argIdx >= len(paramNames) {
			break
		}

		paramName := paramNames[argIdx]
		if paramName == "" || paramName == "_" {
			continue
		}

		// Get the position of the argument
		argPos := fset.Position(arg.Pos())
		corePos := core.Position{
			Line:      argPos.Line - 1,
			Character: argPos.Column - 1,
		}

		// Check if this position is within the requested range
		if !rng.Contains(corePos) {
			continue
		}

		kind := core.InlayHintKindParameter
		hints = append(hints, core.InlayHint{
			Position:     corePos,
			Label:        paramName + ":",
			Kind:         &kind,
			PaddingRight: true,
		})
	}

	return hints
}

func (p *GoParameterNameInlayHintsProvider) findFuncDecl(f *ast.File, name string) *ast.FuncDecl {
	var found *ast.FuncDecl
	ast.Inspect(f, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			if funcDecl.Name.Name == name {
				found = funcDecl
				return false
			}
		}
		return true
	})
	return found
}

// GoTypeInlayHintsProvider provides type hints for variable declarations.
// This shows inferred types for variables declared with :=.
type GoTypeInlayHintsProvider struct{}

func (p *GoTypeInlayHintsProvider) ProvideInlayHints(uri, content string, rng core.Range) []core.InlayHint {
	if !strings.HasSuffix(uri, ".go") {
		return nil
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return nil
	}

	var hints []core.InlayHint

	// Find all short variable declarations (:=) within the range
	ast.Inspect(f, func(n ast.Node) bool {
		if assign, ok := n.(*ast.AssignStmt); ok && assign.Tok == token.DEFINE {
			hints = append(hints, p.hintsForAssignment(assign, fset, content, rng)...)
		}
		return true
	})

	return hints
}

func (p *GoTypeInlayHintsProvider) hintsForAssignment(assign *ast.AssignStmt, fset *token.FileSet, content string, rng core.Range) []core.InlayHint {
	var hints []core.InlayHint

	for i, lhs := range assign.Lhs {
		if ident, ok := lhs.(*ast.Ident); ok && ident.Name != "_" {
			// Get the position after the identifier
			identEnd := fset.Position(ident.End())
			corePos := core.Position{
				Line:      identEnd.Line - 1,
				Character: identEnd.Column - 1,
			}

			// Check if this position is within the requested range
			if !rng.Contains(corePos) {
				continue
			}

			// Try to infer the type from the right-hand side
			var typeStr string
			if i < len(assign.Rhs) {
				typeStr = p.inferType(assign.Rhs[i])
			}

			if typeStr == "" {
				continue
			}

			kind := core.InlayHintKindType
			hints = append(hints, core.InlayHint{
				Position:    corePos,
				Label:       fmt.Sprintf(": %s", typeStr),
				Kind:        &kind,
				PaddingLeft: false,
			})
		}
	}

	return hints
}

func (p *GoTypeInlayHintsProvider) inferType(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.BasicLit:
		switch e.Kind {
		case token.INT:
			return "int"
		case token.FLOAT:
			return "float64"
		case token.STRING:
			return "string"
		case token.CHAR:
			return "rune"
		}
	case *ast.CompositeLit:
		if e.Type != nil {
			return p.exprString(e.Type)
		}
	case *ast.FuncLit:
		return "func"
	case *ast.UnaryExpr:
		if e.Op == token.AND {
			return "*" + p.inferType(e.X)
		}
	case *ast.CallExpr:
		// Could be a function call, type conversion, or composite literal
		if ident, ok := e.Fun.(*ast.Ident); ok {
			// Simple heuristic: if it starts with uppercase, it's probably a type
			if len(ident.Name) > 0 && ident.Name[0] >= 'A' && ident.Name[0] <= 'Z' {
				return ident.Name
			}
		}
	}
	return ""
}

func (p *GoTypeInlayHintsProvider) exprString(expr ast.Expr) string {
	if expr == nil {
		return ""
	}

	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + p.exprString(e.X)
	case *ast.ArrayType:
		return "[]" + p.exprString(e.Elt)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", p.exprString(e.Key), p.exprString(e.Value))
	case *ast.SelectorExpr:
		return p.exprString(e.X) + "." + e.Sel.Name
	default:
		return ""
	}
}

// CompositeInlayHintsProvider combines multiple inlay hint providers.
type CompositeInlayHintsProvider struct {
	Providers []InlayHintsProvider
}

type InlayHintsProvider interface {
	ProvideInlayHints(uri, content string, rng core.Range) []core.InlayHint
}

func NewCompositeInlayHintsProvider(providers ...InlayHintsProvider) *CompositeInlayHintsProvider {
	return &CompositeInlayHintsProvider{
		Providers: providers,
	}
}

func (p *CompositeInlayHintsProvider) ProvideInlayHints(uri, content string, rng core.Range) []core.InlayHint {
	var allHints []core.InlayHint

	for _, provider := range p.Providers {
		hints := provider.ProvideInlayHints(uri, content, rng)
		allHints = append(allHints, hints...)
	}

	return allHints
}

// Example usage in CLI tool
func CLIInlayHintsExample() {
	content := `package main

func add(a, b int) int {
	return a + b
}

func main() {
	x := 42
	y := add(10, 20)
	z := "hello"
}
`

	// Create composite provider with both parameter and type hints
	provider := NewCompositeInlayHintsProvider(
		&GoParameterNameInlayHintsProvider{},
		&GoTypeInlayHintsProvider{},
	)

	// Get hints for the entire document
	rng := core.Range{
		Start: core.Position{Line: 0, Character: 0},
		End:   core.Position{Line: 100, Character: 0},
	}

	hints := provider.ProvideInlayHints("file:///main.go", content, rng)

	println(fmt.Sprintf("Found %d inlay hints:", len(hints)))
	for _, hint := range hints {
		kindStr := "unknown"
		if hint.Kind != nil {
			switch *hint.Kind {
			case core.InlayHintKindType:
				kindStr = "type"
			case core.InlayHintKindParameter:
				kindStr = "parameter"
			}
		}
		println(fmt.Sprintf("  %s: %s (%s)", hint.Position.String(), hint.Label, kindStr))
	}
}

// Example usage in LSP server
// func (s *Server) TextDocumentInlayHint(
// 	ctx *glsp.Context,
// 	params *protocol.InlayHintParams,
// ) ([]protocol.InlayHint, error) {
// 	uri := string(params.TextDocument.URI)
// 	content := s.documents.GetContent(uri)
//
// 	// Convert protocol range to core range
// 	coreRange := adapter_3_16.ProtocolToCoreRange(params.Range, content)
//
// 	// Use provider with core types
// 	coreHints := s.inlayHintsProvider.ProvideInlayHints(uri, content, coreRange)
//
// 	// Convert back to protocol
// 	return adapter_3_16.CoreToProtocolInlayHints(coreHints, content), nil
// }
