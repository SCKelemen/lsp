package examples

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/SCKelemen/lsp/core"
)

// SimpleHoverProvider provides hover information for Go code.
type SimpleHoverProvider struct{}

func (p *SimpleHoverProvider) ProvideHover(uri, content string, position core.Position) *core.HoverInfo {
	if !strings.HasSuffix(uri, ".go") {
		return nil
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return nil
	}

	// Find node at position
	offset := core.PositionToByteOffset(content, position)
	node := p.findNodeAtOffset(f, fset, offset)
	if node == nil {
		return nil
	}

	return p.generateHover(f, fset, node, content)
}

func (p *SimpleHoverProvider) findNodeAtOffset(f *ast.File, fset *token.FileSet, offset int) ast.Node {
	var found ast.Node
	var foundSize int = -1

	ast.Inspect(f, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		start := fset.Position(n.Pos()).Offset
		end := fset.Position(n.End()).Offset

		if offset >= start && offset < end {
			size := end - start
			// Keep the smallest (most specific) node
			if foundSize == -1 || size < foundSize {
				found = n
				foundSize = size
			}
		}

		return true
	})

	return found
}

func (p *SimpleHoverProvider) generateHover(f *ast.File, fset *token.FileSet, node ast.Node, content string) *core.HoverInfo {
	switch n := node.(type) {
	case *ast.Ident:
		return p.hoverForIdent(f, fset, n)
	case *ast.FuncDecl:
		return p.hoverForFunc(fset, n)
	case *ast.TypeSpec:
		return p.hoverForType(fset, n)
	default:
		return nil
	}
}

func (p *SimpleHoverProvider) hoverForIdent(f *ast.File, fset *token.FileSet, ident *ast.Ident) *core.HoverInfo {
	var hoverText strings.Builder
	hoverText.WriteString("```go\n")

	// Look for the definition
	found := false
	var docText string
	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Name.Name == ident.Name {
				hoverText.WriteString(p.funcSignature(node))
				if node.Doc != nil {
					docText = node.Doc.Text()
				}
				found = true
				return false
			}

		case *ast.TypeSpec:
			if node.Name.Name == ident.Name {
				hoverText.WriteString("type ")
				hoverText.WriteString(node.Name.Name)
				hoverText.WriteString(" ")
				hoverText.WriteString(p.exprString(node.Type))
				found = true
				return false
			}

		case *ast.ValueSpec:
			for _, name := range node.Names {
				if name.Name == ident.Name {
					hoverText.WriteString("var ")
					hoverText.WriteString(name.Name)
					if node.Type != nil {
						hoverText.WriteString(" ")
						hoverText.WriteString(p.exprString(node.Type))
					}
					found = true
					return false
				}
			}
		}
		return true
	})

	if !found {
		return nil
	}

	hoverText.WriteString("\n```")

	// Add documentation if available
	if docText != "" {
		hoverText.WriteString("\n\n")
		hoverText.WriteString(docText)
	}

	identStart := fset.Position(ident.Pos())
	identEnd := fset.Position(ident.End())

	r := core.Range{
		Start: core.Position{Line: identStart.Line - 1, Character: identStart.Column - 1},
		End:   core.Position{Line: identEnd.Line - 1, Character: identEnd.Column - 1},
	}

	return &core.HoverInfo{
		Contents: hoverText.String(),
		Range:    &r,
	}
}

func (p *SimpleHoverProvider) hoverForFunc(fset *token.FileSet, fn *ast.FuncDecl) *core.HoverInfo {
	var hoverText strings.Builder
	hoverText.WriteString("```go\n")
	hoverText.WriteString(p.funcSignature(fn))
	hoverText.WriteString("\n```")

	if fn.Doc != nil {
		hoverText.WriteString("\n\n")
		hoverText.WriteString(fn.Doc.Text())
	}

	funcStart := fset.Position(fn.Pos())
	funcEnd := fset.Position(fn.Name.End())

	r := core.Range{
		Start: core.Position{Line: funcStart.Line - 1, Character: funcStart.Column - 1},
		End:   core.Position{Line: funcEnd.Line - 1, Character: funcEnd.Column - 1},
	}

	return &core.HoverInfo{
		Contents: hoverText.String(),
		Range:    &r,
	}
}

func (p *SimpleHoverProvider) hoverForType(fset *token.FileSet, typeSpec *ast.TypeSpec) *core.HoverInfo {
	var hoverText strings.Builder
	hoverText.WriteString("```go\n")
	hoverText.WriteString("type ")
	hoverText.WriteString(typeSpec.Name.Name)
	hoverText.WriteString(" ")
	hoverText.WriteString(p.exprString(typeSpec.Type))
	hoverText.WriteString("\n```")

	typeStart := fset.Position(typeSpec.Pos())
	typeEnd := fset.Position(typeSpec.End())

	r := core.Range{
		Start: core.Position{Line: typeStart.Line - 1, Character: typeStart.Column - 1},
		End:   core.Position{Line: typeEnd.Line - 1, Character: typeEnd.Column - 1},
	}

	return &core.HoverInfo{
		Contents: hoverText.String(),
		Range:    &r,
	}
}

func (p *SimpleHoverProvider) funcSignature(fn *ast.FuncDecl) string {
	var sig strings.Builder

	sig.WriteString("func ")

	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		sig.WriteString("(")
		recv := fn.Recv.List[0]
		if len(recv.Names) > 0 {
			sig.WriteString(recv.Names[0].Name)
			sig.WriteString(" ")
		}
		sig.WriteString(p.exprString(recv.Type))
		sig.WriteString(") ")
	}

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
			sig.WriteString(p.exprString(param.Type))
		}
	}

	sig.WriteString(")")

	if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
		sig.WriteString(" ")
		if len(fn.Type.Results.List) == 1 && len(fn.Type.Results.List[0].Names) == 0 {
			sig.WriteString(p.exprString(fn.Type.Results.List[0].Type))
		} else {
			sig.WriteString("(")
			for i, result := range fn.Type.Results.List {
				if i > 0 {
					sig.WriteString(", ")
				}
				sig.WriteString(p.exprString(result.Type))
			}
			sig.WriteString(")")
		}
	}

	return sig.String()
}

func (p *SimpleHoverProvider) exprString(expr ast.Expr) string {
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
		return "map[" + p.exprString(e.Key) + "]" + p.exprString(e.Value)
	case *ast.SelectorExpr:
		return p.exprString(e.X) + "." + e.Sel.Name
	case *ast.FuncType:
		return "func(...)"
	default:
		return "..."
	}
}

// SimpleDefinitionProvider provides go-to-definition for Go code.
type SimpleDefinitionProvider struct{}

func (p *SimpleDefinitionProvider) ProvideDefinition(uri, content string, position core.Position) []core.Location {
	if !strings.HasSuffix(uri, ".go") {
		return nil
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", content, 0)
	if err != nil {
		return nil
	}

	// Find identifier at position
	offset := core.PositionToByteOffset(content, position)
	ident := p.findIdentAtOffset(f, fset, offset)
	if ident == nil {
		return nil
	}

	// Find definition
	return p.findDefinition(f, fset, ident, uri)
}

func (p *SimpleDefinitionProvider) findIdentAtOffset(f *ast.File, fset *token.FileSet, offset int) *ast.Ident {
	var found *ast.Ident

	ast.Inspect(f, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			start := fset.Position(ident.Pos()).Offset
			end := fset.Position(ident.End()).Offset

			if offset >= start && offset < end {
				found = ident
				return false
			}
		}
		return true
	})

	return found
}

func (p *SimpleDefinitionProvider) findDefinition(f *ast.File, fset *token.FileSet, ident *ast.Ident, uri string) []core.Location {
	var defNode ast.Node

	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Name.Name == ident.Name {
				defNode = node.Name
				return false
			}

		case *ast.TypeSpec:
			if node.Name.Name == ident.Name {
				defNode = node.Name
				return false
			}

		case *ast.ValueSpec:
			for _, name := range node.Names {
				if name.Name == ident.Name {
					defNode = name
					return false
				}
			}
		}
		return true
	})

	if defNode == nil {
		return nil
	}

	defPos := fset.Position(defNode.Pos())
	defEnd := fset.Position(defNode.End())

	return []core.Location{
		{
			URI: uri,
			Range: core.Range{
				Start: core.Position{
					Line:      defPos.Line - 1,
					Character: defPos.Column - 1,
				},
				End: core.Position{
					Line:      defEnd.Line - 1,
					Character: defEnd.Column - 1,
				},
			},
		},
	}
}

// MarkedStringHoverProvider provides hover with marked strings.
type MarkedStringHoverProvider struct{}

func (p *MarkedStringHoverProvider) ProvideHover(uri, content string, position core.Position) *core.HoverInfo {
	// Simple example: hover over specific keywords
	lines := strings.Split(content, "\n")
	if position.Line >= len(lines) {
		return nil
	}

	line := lines[position.Line]
	if position.Character >= len(line) {
		return nil
	}

	// Find word at position
	word := p.getWordAtPosition(line, position.Character)
	if word == "" {
		return nil
	}

	// Provide hover for Go keywords
	hoverText := p.getKeywordHover(word)
	if hoverText == "" {
		return nil
	}

	// Find word range
	wordStart := position.Character
	for wordStart > 0 && p.isIdentChar(line[wordStart-1]) {
		wordStart--
	}

	wordEnd := position.Character
	for wordEnd < len(line) && p.isIdentChar(line[wordEnd]) {
		wordEnd++
	}

	r := core.Range{
		Start: core.Position{Line: position.Line, Character: wordStart},
		End:   core.Position{Line: position.Line, Character: wordEnd},
	}

	return &core.HoverInfo{
		Contents: hoverText,
		Range:    &r,
	}
}

func (p *MarkedStringHoverProvider) getWordAtPosition(line string, pos int) string {
	if pos >= len(line) {
		return ""
	}

	start := pos
	for start > 0 && p.isIdentChar(line[start-1]) {
		start--
	}

	end := pos
	for end < len(line) && p.isIdentChar(line[end]) {
		end++
	}

	if start >= end {
		return ""
	}

	return line[start:end]
}

func (p *MarkedStringHoverProvider) isIdentChar(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_'
}

func (p *MarkedStringHoverProvider) getKeywordHover(word string) string {
	keywords := map[string]string{
		"func":      "```go\nfunc\n```\n\nDefines a function",
		"var":       "```go\nvar\n```\n\nDeclares a variable",
		"const":     "```go\nconst\n```\n\nDeclares a constant",
		"type":      "```go\ntype\n```\n\nDefines a type",
		"interface": "```go\ninterface\n```\n\nDefines an interface type",
		"struct":    "```go\nstruct\n```\n\nDefines a struct type",
		"package":   "```go\npackage\n```\n\nDeclares the package name",
		"import":    "```go\nimport\n```\n\nImports packages",
		"return":    "```go\nreturn\n```\n\nReturns from a function",
		"if":        "```go\nif\n```\n\nConditional statement",
		"for":       "```go\nfor\n```\n\nLoop statement",
	}

	return keywords[word]
}
