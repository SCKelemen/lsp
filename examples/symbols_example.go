package examples

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/SCKelemen/lsp/core"
)

// GoSymbolProvider provides document symbols for Go source files.
type GoSymbolProvider struct{}

func (p *GoSymbolProvider) ProvideDocumentSymbols(uri, content string) []core.DocumentSymbol {
	if !strings.HasSuffix(uri, ".go") {
		return nil
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return nil
	}

	var symbols []core.DocumentSymbol

	for _, decl := range f.Decls {
		if symbol := p.declToSymbol(decl, fset); symbol != nil {
			symbols = append(symbols, *symbol)
		}
	}

	return symbols
}

func (p *GoSymbolProvider) declToSymbol(decl ast.Decl, fset *token.FileSet) *core.DocumentSymbol {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		return p.funcToSymbol(d, fset)
	case *ast.GenDecl:
		return p.genDeclToSymbol(d, fset)
	default:
		return nil
	}
}

func (p *GoSymbolProvider) funcToSymbol(fn *ast.FuncDecl, fset *token.FileSet) *core.DocumentSymbol {
	start := fset.Position(fn.Pos())
	end := fset.Position(fn.End())
	nameStart := fset.Position(fn.Name.Pos())
	nameEnd := fset.Position(fn.Name.End())

	kind := core.SymbolKindFunction
	name := fn.Name.Name
	detail := p.getFunctionSignature(fn)

	if fn.Recv != nil {
		kind = core.SymbolKindMethod
		if len(fn.Recv.List) > 0 {
			recvType := p.exprToString(fn.Recv.List[0].Type)
			detail = "(" + recvType + ") " + detail
		}
	}

	symbol := &core.DocumentSymbol{
		Name:   name,
		Detail: detail,
		Kind:   kind,
		Range: core.Range{
			Start: core.Position{Line: start.Line - 1, Character: start.Column - 1},
			End:   core.Position{Line: end.Line - 1, Character: end.Column - 1},
		},
		SelectionRange: core.Range{
			Start: core.Position{Line: nameStart.Line - 1, Character: nameStart.Column - 1},
			End:   core.Position{Line: nameEnd.Line - 1, Character: nameEnd.Column - 1},
		},
	}

	return symbol
}

func (p *GoSymbolProvider) getFunctionSignature(fn *ast.FuncDecl) string {
	var sig strings.Builder
	sig.WriteString(fn.Name.Name)
	sig.WriteString("(")

	if fn.Type.Params != nil {
		for i, param := range fn.Type.Params.List {
			if i > 0 {
				sig.WriteString(", ")
			}
			for j, name := range param.Names {
				if j > 0 {
					sig.WriteString(", ")
				}
				sig.WriteString(name.Name)
			}
			if len(param.Names) > 0 {
				sig.WriteString(" ")
			}
			sig.WriteString(p.exprToString(param.Type))
		}
	}

	sig.WriteString(")")

	if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
		sig.WriteString(" ")
		if len(fn.Type.Results.List) == 1 && len(fn.Type.Results.List[0].Names) == 0 {
			sig.WriteString(p.exprToString(fn.Type.Results.List[0].Type))
		} else {
			sig.WriteString("(")
			for i, result := range fn.Type.Results.List {
				if i > 0 {
					sig.WriteString(", ")
				}
				sig.WriteString(p.exprToString(result.Type))
			}
			sig.WriteString(")")
		}
	}

	return sig.String()
}

func (p *GoSymbolProvider) genDeclToSymbol(gen *ast.GenDecl, fset *token.FileSet) *core.DocumentSymbol {
	var symbols []core.DocumentSymbol

	for _, spec := range gen.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			symbols = append(symbols, p.typeSpecToSymbols(s, fset)...)
		case *ast.ValueSpec:
			symbols = append(symbols, p.valueSpecToSymbols(s, gen.Tok, fset)...)
		}
	}

	if len(symbols) == 1 {
		return &symbols[0]
	}

	if len(symbols) > 1 {
		start := fset.Position(gen.Pos())
		end := fset.Position(gen.End())

		kind := core.SymbolKindVariable
		name := "declarations"

		switch gen.Tok {
		case token.TYPE:
			kind = core.SymbolKindClass
			name = "types"
		case token.CONST:
			kind = core.SymbolKindConstant
			name = "constants"
		case token.VAR:
			kind = core.SymbolKindVariable
			name = "variables"
		}

		return &core.DocumentSymbol{
			Name: name,
			Kind: kind,
			Range: core.Range{
				Start: core.Position{Line: start.Line - 1, Character: start.Column - 1},
				End:   core.Position{Line: end.Line - 1, Character: end.Column - 1},
			},
			SelectionRange: core.Range{
				Start: core.Position{Line: start.Line - 1, Character: start.Column - 1},
				End:   core.Position{Line: start.Line - 1, Character: start.Column - 1},
			},
			Children: symbols,
		}
	}

	return nil
}

func (p *GoSymbolProvider) typeSpecToSymbols(spec *ast.TypeSpec, fset *token.FileSet) []core.DocumentSymbol {
	nameStart := fset.Position(spec.Name.Pos())
	nameEnd := fset.Position(spec.Name.End())
	typeStart := fset.Position(spec.Pos())
	typeEnd := fset.Position(spec.End())

	kind := core.SymbolKindClass

	switch spec.Type.(type) {
	case *ast.StructType:
		kind = core.SymbolKindStruct
	case *ast.InterfaceType:
		kind = core.SymbolKindInterface
	}

	symbol := core.DocumentSymbol{
		Name:   spec.Name.Name,
		Detail: p.exprToString(spec.Type),
		Kind:   kind,
		Range: core.Range{
			Start: core.Position{Line: typeStart.Line - 1, Character: typeStart.Column - 1},
			End:   core.Position{Line: typeEnd.Line - 1, Character: typeEnd.Column - 1},
		},
		SelectionRange: core.Range{
			Start: core.Position{Line: nameStart.Line - 1, Character: nameStart.Column - 1},
			End:   core.Position{Line: nameEnd.Line - 1, Character: nameEnd.Column - 1},
		},
	}

	// Add struct fields or interface methods as children
	switch t := spec.Type.(type) {
	case *ast.StructType:
		symbol.Children = p.getStructFields(t, fset)
	case *ast.InterfaceType:
		symbol.Children = p.getInterfaceMethods(t, fset)
	}

	return []core.DocumentSymbol{symbol}
}

func (p *GoSymbolProvider) valueSpecToSymbols(spec *ast.ValueSpec, tok token.Token, fset *token.FileSet) []core.DocumentSymbol {
	var symbols []core.DocumentSymbol

	kind := core.SymbolKindVariable
	if tok == token.CONST {
		kind = core.SymbolKindConstant
	}

	for i, name := range spec.Names {
		nameStart := fset.Position(name.Pos())
		nameEnd := fset.Position(name.End())
		valueStart := fset.Position(spec.Pos())
		valueEnd := fset.Position(spec.End())

		detail := ""
		if spec.Type != nil {
			detail = p.exprToString(spec.Type)
		} else if i < len(spec.Values) {
			detail = p.exprToString(spec.Values[i])
		}

		symbols = append(symbols, core.DocumentSymbol{
			Name:   name.Name,
			Detail: detail,
			Kind:   kind,
			Range: core.Range{
				Start: core.Position{Line: valueStart.Line - 1, Character: valueStart.Column - 1},
				End:   core.Position{Line: valueEnd.Line - 1, Character: valueEnd.Column - 1},
			},
			SelectionRange: core.Range{
				Start: core.Position{Line: nameStart.Line - 1, Character: nameStart.Column - 1},
				End:   core.Position{Line: nameEnd.Line - 1, Character: nameEnd.Column - 1},
			},
		})
	}

	return symbols
}

func (p *GoSymbolProvider) getStructFields(st *ast.StructType, fset *token.FileSet) []core.DocumentSymbol {
	var fields []core.DocumentSymbol

	if st.Fields == nil {
		return fields
	}

	for _, field := range st.Fields.List {
		for _, name := range field.Names {
			nameStart := fset.Position(name.Pos())
			nameEnd := fset.Position(name.End())
			fieldStart := fset.Position(field.Pos())
			fieldEnd := fset.Position(field.End())

			fields = append(fields, core.DocumentSymbol{
				Name:   name.Name,
				Detail: p.exprToString(field.Type),
				Kind:   core.SymbolKindField,
				Range: core.Range{
					Start: core.Position{Line: fieldStart.Line - 1, Character: fieldStart.Column - 1},
					End:   core.Position{Line: fieldEnd.Line - 1, Character: fieldEnd.Column - 1},
				},
				SelectionRange: core.Range{
					Start: core.Position{Line: nameStart.Line - 1, Character: nameStart.Column - 1},
					End:   core.Position{Line: nameEnd.Line - 1, Character: nameEnd.Column - 1},
				},
			})
		}
	}

	return fields
}

func (p *GoSymbolProvider) getInterfaceMethods(it *ast.InterfaceType, fset *token.FileSet) []core.DocumentSymbol {
	var methods []core.DocumentSymbol

	if it.Methods == nil {
		return methods
	}

	for _, method := range it.Methods.List {
		for _, name := range method.Names {
			nameStart := fset.Position(name.Pos())
			nameEnd := fset.Position(name.End())
			methodStart := fset.Position(method.Pos())
			methodEnd := fset.Position(method.End())

			methods = append(methods, core.DocumentSymbol{
				Name:   name.Name,
				Detail: p.exprToString(method.Type),
				Kind:   core.SymbolKindMethod,
				Range: core.Range{
					Start: core.Position{Line: methodStart.Line - 1, Character: methodStart.Column - 1},
					End:   core.Position{Line: methodEnd.Line - 1, Character: methodEnd.Column - 1},
				},
				SelectionRange: core.Range{
					Start: core.Position{Line: nameStart.Line - 1, Character: nameStart.Column - 1},
					End:   core.Position{Line: nameEnd.Line - 1, Character: nameEnd.Column - 1},
				},
			})
		}
	}

	return methods
}

func (p *GoSymbolProvider) exprToString(expr ast.Expr) string {
	if expr == nil {
		return ""
	}

	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + p.exprToString(e.X)
	case *ast.ArrayType:
		return "[]" + p.exprToString(e.Elt)
	case *ast.MapType:
		return "map[" + p.exprToString(e.Key) + "]" + p.exprToString(e.Value)
	case *ast.SelectorExpr:
		return p.exprToString(e.X) + "." + e.Sel.Name
	case *ast.StructType:
		return "struct{...}"
	case *ast.InterfaceType:
		return "interface{...}"
	case *ast.FuncType:
		return "func(...)"
	default:
		return "..."
	}
}
