package examples

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/SCKelemen/lsp/core"
)

// GoSignatureHelpProvider provides signature help for Go function calls.
// Shows function signatures and highlights the active parameter as you type.
type GoSignatureHelpProvider struct{}

func (p *GoSignatureHelpProvider) ProvideSignatureHelp(ctx core.SignatureHelpContext) *core.SignatureHelp {
	if !strings.HasSuffix(ctx.URI, ".go") {
		return nil
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", ctx.Content, parser.ParseComments)
	if err != nil {
		return nil
	}

	// Find the function call at the position
	offset := core.PositionToByteOffset(ctx.Content, ctx.Position)
	callExpr := p.findCallExprAtPosition(f, fset, ctx.Content, offset)
	if callExpr == nil {
		return nil
	}

	// Find the function declaration
	var funcName string
	if ident, ok := callExpr.Fun.(*ast.Ident); ok {
		funcName = ident.Name
	} else {
		return nil
	}

	funcDecl := p.findFuncDecl(f, funcName)
	if funcDecl == nil {
		return nil
	}

	// Build signature information
	sigInfo := p.buildSignatureInfo(funcDecl, fset)
	if sigInfo == nil {
		return nil
	}

	// Determine active parameter based on cursor position
	activeParam := p.determineActiveParameter(callExpr, fset, ctx.Content, offset)

	return &core.SignatureHelp{
		Signatures:      []core.SignatureInformation{*sigInfo},
		ActiveSignature: intPtr(0),
		ActiveParameter: intPtr(activeParam),
	}
}

// findCallExprAtPosition finds the function call expression at the given offset
func (p *GoSignatureHelpProvider) findCallExprAtPosition(f *ast.File, fset *token.FileSet, content string, offset int) *ast.CallExpr {
	var result *ast.CallExpr

	ast.Inspect(f, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check if the cursor is within the function call
		callStart := fset.Position(call.Lparen)
		callEnd := fset.Position(call.Rparen)

		startOffset := core.PositionToByteOffset(content, core.Position{
			Line:      callStart.Line - 1,
			Character: callStart.Column - 1,
		})
		endOffset := core.PositionToByteOffset(content, core.Position{
			Line:      callEnd.Line - 1,
			Character: callEnd.Column - 1,
		})

		if startOffset <= offset && offset <= endOffset {
			result = call
			return false
		}

		return true
	})

	return result
}

// findFuncDecl finds the function declaration with the given name
func (p *GoSignatureHelpProvider) findFuncDecl(f *ast.File, name string) *ast.FuncDecl {
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

// buildSignatureInfo creates signature information from a function declaration
func (p *GoSignatureHelpProvider) buildSignatureInfo(funcDecl *ast.FuncDecl, fset *token.FileSet) *core.SignatureInformation {
	if funcDecl.Type == nil {
		return nil
	}

	// Build signature label
	var label strings.Builder
	label.WriteString(funcDecl.Name.Name)
	label.WriteString("(")

	var params []core.ParameterInformation

	if funcDecl.Type.Params != nil {
		for i, field := range funcDecl.Type.Params.List {
		if i > 0 {
			label.WriteString(", ")
		}

		// Get parameter names and type
		var paramNames []string
		if len(field.Names) == 0 {
			// Unnamed parameter
			paramNames = []string{""}
		} else {
			for _, name := range field.Names {
				paramNames = append(paramNames, name.Name)
			}
		}

		// Get type string
		typeStr := p.typeToString(field.Type)

		// Add each parameter
		for j, paramName := range paramNames {
			if j > 0 {
				label.WriteString(", ")
			}

			paramLabel := paramName
			if paramName != "" {
				paramLabel += " "
			}
			paramLabel += typeStr

			label.WriteString(paramLabel)

			params = append(params, core.ParameterInformation{
				Label:         paramLabel,
				Documentation: "", // Could be extracted from comments
			})
		}
		}
	}

	label.WriteString(")")

	// Add return type if present
	if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) > 0 {
		label.WriteString(" ")
		if len(funcDecl.Type.Results.List) == 1 && len(funcDecl.Type.Results.List[0].Names) == 0 {
			label.WriteString(p.typeToString(funcDecl.Type.Results.List[0].Type))
		} else {
			label.WriteString("(")
			for i, field := range funcDecl.Type.Results.List {
				if i > 0 {
					label.WriteString(", ")
				}
				label.WriteString(p.typeToString(field.Type))
			}
			label.WriteString(")")
		}
	}

	return &core.SignatureInformation{
		Label:         label.String(),
		Documentation: p.extractDocumentation(funcDecl),
		Parameters:    params,
	}
}

// typeToString converts an AST type expression to a string
func (p *GoSignatureHelpProvider) typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + p.typeToString(t.X)
	case *ast.ArrayType:
		return "[]" + p.typeToString(t.Elt)
	case *ast.Ellipsis:
		return "..." + p.typeToString(t.Elt)
	case *ast.MapType:
		return "map[" + p.typeToString(t.Key) + "]" + p.typeToString(t.Value)
	case *ast.SelectorExpr:
		return p.typeToString(t.X) + "." + t.Sel.Name
	case *ast.FuncType:
		return "func(...)"
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StructType:
		return "struct{...}"
	case *ast.ChanType:
		return "chan " + p.typeToString(t.Value)
	default:
		return "unknown"
	}
}

// extractDocumentation extracts documentation from function comments
func (p *GoSignatureHelpProvider) extractDocumentation(funcDecl *ast.FuncDecl) string {
	if funcDecl.Doc == nil {
		return ""
	}

	var doc strings.Builder
	for _, comment := range funcDecl.Doc.List {
		text := strings.TrimPrefix(comment.Text, "//")
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
		text = strings.TrimSpace(text)
		if doc.Len() > 0 {
			doc.WriteString("\n")
		}
		doc.WriteString(text)
	}

	return doc.String()
}

// determineActiveParameter figures out which parameter the cursor is on
func (p *GoSignatureHelpProvider) determineActiveParameter(call *ast.CallExpr, fset *token.FileSet, content string, offset int) int {
	if len(call.Args) == 0 {
		return 0
	}

	// Count commas before the cursor position to determine parameter index
	paramIndex := 0

	for i, arg := range call.Args {
		argStart := fset.Position(arg.Pos())
		startOffset := core.PositionToByteOffset(content, core.Position{
			Line:      argStart.Line - 1,
			Character: argStart.Column - 1,
		})

		if offset < startOffset {
			return paramIndex
		}

		// If we're past this argument, move to next parameter
		argEnd := fset.Position(arg.End())
		endOffset := core.PositionToByteOffset(content, core.Position{
			Line:      argEnd.Line - 1,
			Character: argEnd.Column - 1,
		})

		if offset >= startOffset && offset <= endOffset {
			return i
		}

		paramIndex = i + 1
	}

	return paramIndex
}

func intPtr(i int) *int {
	return &i
}

// Example usage in CLI tool
func CLISignatureHelpExample() {
	content := `package main

// Add adds two integers and returns the result.
func Add(a int, b int) int {
	return a + b
}

func main() {
	result := Add(10, 20)
}
`

	provider := &GoSignatureHelpProvider{}

	// Request signature help at the position after "Add("
	ctx := core.SignatureHelpContext{
		URI:     "file:///main.go",
		Content: content,
		Position: core.Position{
			Line:      8,
			Character: 17, // After "Add(1"
		},
		TriggerCharacter: "(",
	}

	help := provider.ProvideSignatureHelp(ctx)
	if help != nil && len(help.Signatures) > 0 {
		sig := help.Signatures[0]
		println("Signature:", sig.Label)
		println("Documentation:", sig.Documentation)
		if help.ActiveParameter != nil {
			println("Active Parameter:", *help.ActiveParameter)
		}
	}
}
