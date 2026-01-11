# Navigation Providers Guide

Complete guide to implementing navigation features: go-to-definition, hover information, and more.

**LSP Specification References:**
- [Go to Definition](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#textDocument_definition)
- [Hover](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#textDocument_hover)
- [References](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#textDocument_references)

## Table of Contents

1. [Core Concepts](#core-concepts)
2. [Definition Provider](#definition-provider)
3. [Hover Provider](#hover-provider)
4. [Testing Navigation Providers](#testing-navigation-providers)
5. [LSP Server Integration](#lsp-server-integration)

## Core Concepts

### What Are Navigation Providers?

Navigation providers enable code intelligence features:
- **Go to Definition**: Jump to where a symbol is defined
- **Hover**: Show information about a symbol
- **Find References**: Find all usages of a symbol
- **Type Definition**: Go to type definition
- **Implementation**: Find implementations of an interface

### Provider Interfaces

```go
import "github.com/SCKelemen/lsp/core"

// DefinitionProvider provides go-to-definition locations
type DefinitionProvider interface {
    ProvideDefinition(uri, content string, position Position) []Location
}

// HoverProvider provides hover information
type HoverProvider interface {
    ProvideHover(uri, content string, position Position) *HoverInfo
}
```

### Core Types

```go
// Location represents a location in a document
type Location struct {
    URI   string
    Range Range
}

// HoverInfo contains hover information
type HoverInfo struct {
    Contents string  // Markdown or plain text
    Range    *Range  // Range to which hover applies
}
```

## Definition Provider

### Your First Definition Provider

Let's create a provider for Go import statements:

```go
package mypackage

import (
    "go/ast"
    "go/parser"
    "go/token"
    "os"
    "path/filepath"
    "strings"

    "github.com/SCKelemen/lsp/core"
)

// GoDefinitionProvider provides go-to-definition for Go code.
type GoDefinitionProvider struct {
    workspace string // Root workspace directory
}

func NewGoDefinitionProvider(workspace string) *GoDefinitionProvider {
    return &GoDefinitionProvider{workspace: workspace}
}

func (p *GoDefinitionProvider) ProvideDefinition(
    uri, content string,
    position core.Position,
) []core.Location {
    // Only handle Go files
    if !strings.HasSuffix(uri, ".go") {
        return nil
    }

    // Parse the file
    fset := token.NewFileSet()
    f, err := parser.ParseFile(fset, "", content, 0)
    if err != nil {
        return nil
    }

    // Find the identifier at the position
    ident := p.findIdentAtPosition(f, fset, position, content)
    if ident == nil {
        return nil
    }

    // Find the definition
    return p.findDefinition(f, fset, ident, uri, content)
}
```

### Finding Identifier at Position

```go
func (p *GoDefinitionProvider) findIdentAtPosition(
    f *ast.File,
    fset *token.FileSet,
    pos core.Position,
    content string,
) *ast.Ident {
    var found *ast.Ident

    // Convert position to token.Pos
    targetPos := p.corePositionToTokenPos(fset, pos, content)

    ast.Inspect(f, func(n ast.Node) bool {
        if ident, ok := n.(*ast.Ident); ok {
            identPos := fset.Position(ident.Pos())
            identEnd := fset.Position(ident.End())

            // Check if target position is within identifier
            if targetPos >= ident.Pos() && targetPos < ident.End() {
                found = ident
                return false
            }
        }
        return true
    })

    return found
}

func (p *GoDefinitionProvider) corePositionToTokenPos(
    fset *token.FileSet,
    pos core.Position,
    content string,
) token.Pos {
    // Get byte offset
    offset := core.PositionToByteOffset(content, pos)

    // Create a file for position calculation
    file := fset.AddFile("", -1, len(content))
    file.SetLinesForContent([]byte(content))

    return file.Pos(offset)
}
```

### Finding Definition Location

```go
func (p *GoDefinitionProvider) findDefinition(
    f *ast.File,
    fset *token.FileSet,
    ident *ast.Ident,
    uri string,
    content string,
) []core.Location {
    // Check if it's a type, function, variable, etc.

    // Strategy 1: Look for definition in same file
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

    if defNode != nil {
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

    // Strategy 2: Check if it's an imported package
    for _, imp := range f.Imports {
        if imp.Name != nil && imp.Name.Name == ident.Name {
            // Find the package definition
            return p.findPackageDefinition(imp, ident.Name)
        }

        // Handle imports without alias
        importPath := strings.Trim(imp.Path.Value, `"`)
        pkgName := filepath.Base(importPath)
        if pkgName == ident.Name {
            return p.findPackageDefinition(imp, pkgName)
        }
    }

    return nil
}

func (p *GoDefinitionProvider) findPackageDefinition(
    imp *ast.ImportSpec,
    pkgName string,
) []core.Location {
    // Get import path
    importPath := strings.Trim(imp.Path.Value, `"`)

    // Try to find the package in workspace or GOPATH
    // This is simplified - real implementation would use go/packages
    packageDir := p.resolveImportPath(importPath)
    if packageDir == "" {
        return nil
    }

    // Find the first .go file in the package
    files, err := os.ReadDir(packageDir)
    if err != nil {
        return nil
    }

    for _, file := range files {
        if !file.IsDir() && strings.HasSuffix(file.Name(), ".go") {
            filePath := filepath.Join(packageDir, file.Name())

            return []core.Location{
                {
                    URI: "file://" + filePath,
                    Range: core.Range{
                        Start: core.Position{Line: 0, Character: 0},
                        End:   core.Position{Line: 0, Character: 0},
                    },
                },
            }
        }
    }

    return nil
}

func (p *GoDefinitionProvider) resolveImportPath(importPath string) string {
    // Simplified: check workspace first
    packageDir := filepath.Join(p.workspace, importPath)
    if stat, err := os.Stat(packageDir); err == nil && stat.IsDir() {
        return packageDir
    }

    // Check GOPATH
    gopath := os.Getenv("GOPATH")
    if gopath != "" {
        packageDir = filepath.Join(gopath, "src", importPath)
        if stat, err := os.Stat(packageDir); err == nil && stat.IsDir() {
            return packageDir
        }
    }

    return ""
}
```

## Hover Provider

### Your First Hover Provider

Provide information when hovering over symbols:

```go
type GoHoverProvider struct{}

func (p *GoHoverProvider) ProvideHover(
    uri, content string,
    position core.Position,
) *core.HoverInfo {
    // Only handle Go files
    if !strings.HasSuffix(uri, ".go") {
        return nil
    }

    // Parse the file
    fset := token.NewFileSet()
    f, err := parser.ParseFile(fset, "", content, parser.ParseComments)
    if err != nil {
        return nil
    }

    // Find what's at the position
    ident := p.findNodeAtPosition(f, fset, position, content)
    if ident == nil {
        return nil
    }

    // Generate hover content
    return p.generateHoverInfo(f, fset, ident, content)
}
```

### Finding Node at Position

```go
func (p *GoHoverProvider) findNodeAtPosition(
    f *ast.File,
    fset *token.FileSet,
    pos core.Position,
    content string,
) ast.Node {
    offset := core.PositionToByteOffset(content, pos)

    var found ast.Node

    ast.Inspect(f, func(n ast.Node) bool {
        if n == nil {
            return false
        }

        nodeStart := fset.Position(n.Pos())
        nodeEnd := fset.Position(n.End())

        startOffset := core.PositionToByteOffset(content, core.Position{
            Line:      nodeStart.Line - 1,
            Character: nodeStart.Column - 1,
        })

        endOffset := core.PositionToByteOffset(content, core.Position{
            Line:      nodeEnd.Line - 1,
            Character: nodeEnd.Column - 1,
        })

        if offset >= startOffset && offset < endOffset {
            found = n
        }

        return true
    })

    return found
}
```

### Generating Hover Content

```go
func (p *GoHoverProvider) generateHoverInfo(
    f *ast.File,
    fset *token.FileSet,
    node ast.Node,
    content string,
) *core.HoverInfo {
    switch n := node.(type) {
    case *ast.Ident:
        return p.hoverForIdent(f, fset, n, content)
    case *ast.FuncDecl:
        return p.hoverForFunc(f, fset, n, content)
    case *ast.TypeSpec:
        return p.hoverForType(f, fset, n, content)
    default:
        return nil
    }
}

func (p *GoHoverProvider) hoverForIdent(
    f *ast.File,
    fset *token.FileSet,
    ident *ast.Ident,
    content string,
) *core.HoverInfo {
    // Find the declaration
    var hoverText strings.Builder
    hoverText.WriteString("```go\n")

    // Look for the definition
    ast.Inspect(f, func(n ast.Node) bool {
        switch node := n.(type) {
        case *ast.FuncDecl:
            if node.Name.Name == ident.Name {
                hoverText.WriteString(p.funcSignature(node))
                return false
            }

        case *ast.TypeSpec:
            if node.Name.Name == ident.Name {
                hoverText.WriteString("type ")
                hoverText.WriteString(node.Name.Name)
                hoverText.WriteString(" ")
                hoverText.WriteString(p.exprString(node.Type))
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
                    return false
                }
            }
        }
        return true
    })

    hoverText.WriteString("\n```")

    // Add documentation if available
    if doc := p.findDocComment(f, ident.Name); doc != "" {
        hoverText.WriteString("\n\n")
        hoverText.WriteString(doc)
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

func (p *GoHoverProvider) hoverForFunc(
    f *ast.File,
    fset *token.FileSet,
    fn *ast.FuncDecl,
    content string,
) *core.HoverInfo {
    var hoverText strings.Builder
    hoverText.WriteString("```go\n")
    hoverText.WriteString(p.funcSignature(fn))
    hoverText.WriteString("\n```")

    // Add function documentation
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

func (p *GoHoverProvider) hoverForType(
    f *ast.File,
    fset *token.FileSet,
    typeSpec *ast.TypeSpec,
    content string,
) *core.HoverInfo {
    var hoverText strings.Builder
    hoverText.WriteString("```go\n")
    hoverText.WriteString("type ")
    hoverText.WriteString(typeSpec.Name.Name)
    hoverText.WriteString(" ")

    // Show type definition
    switch t := typeSpec.Type.(type) {
    case *ast.StructType:
        hoverText.WriteString("struct {\n")
        if t.Fields != nil {
            for _, field := range t.Fields.List {
                hoverText.WriteString("    ")
                for i, name := range field.Names {
                    if i > 0 {
                        hoverText.WriteString(", ")
                    }
                    hoverText.WriteString(name.Name)
                }
                hoverText.WriteString(" ")
                hoverText.WriteString(p.exprString(field.Type))
                hoverText.WriteString("\n")
            }
        }
        hoverText.WriteString("}")

    case *ast.InterfaceType:
        hoverText.WriteString("interface {\n")
        if t.Methods != nil {
            for _, method := range t.Methods.List {
                hoverText.WriteString("    ")
                if len(method.Names) > 0 {
                    hoverText.WriteString(method.Names[0].Name)
                    hoverText.WriteString(p.exprString(method.Type))
                }
                hoverText.WriteString("\n")
            }
        }
        hoverText.WriteString("}")

    default:
        hoverText.WriteString(p.exprString(typeSpec.Type))
    }

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
```

### Helper Functions

```go
func (p *GoHoverProvider) funcSignature(fn *ast.FuncDecl) string {
    var sig strings.Builder

    sig.WriteString("func ")

    // Add receiver for methods
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

    // Parameters
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

    // Return type
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

func (p *GoHoverProvider) exprString(expr ast.Expr) string {
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
        var sig strings.Builder
        sig.WriteString("func(")
        if e.Params != nil {
            for i, param := range e.Params.List {
                if i > 0 {
                    sig.WriteString(", ")
                }
                sig.WriteString(p.exprString(param.Type))
            }
        }
        sig.WriteString(")")
        if e.Results != nil && len(e.Results.List) > 0 {
            sig.WriteString(" ")
            if len(e.Results.List) == 1 {
                sig.WriteString(p.exprString(e.Results.List[0].Type))
            } else {
                sig.WriteString("(")
                for i, r := range e.Results.List {
                    if i > 0 {
                        sig.WriteString(", ")
                    }
                    sig.WriteString(p.exprString(r.Type))
                }
                sig.WriteString(")")
            }
        }
        return sig.String()
    default:
        return "..."
    }
}

func (p *GoHoverProvider) findDocComment(f *ast.File, name string) string {
    // Look for documentation comment
    for _, decl := range f.Decls {
        switch d := decl.(type) {
        case *ast.FuncDecl:
            if d.Name.Name == name && d.Doc != nil {
                return d.Doc.Text()
            }
        case *ast.GenDecl:
            for _, spec := range d.Specs {
                if ts, ok := spec.(*ast.TypeSpec); ok {
                    if ts.Name.Name == name && ts.Doc != nil {
                        return ts.Doc.Text()
                    }
                }
            }
        }
    }
    return ""
}
```

## Testing Navigation Providers

### Testing Definition Provider

```go
func TestGoDefinitionProvider(t *testing.T) {
    content := `package main

func helper() string {
    return "hello"
}

func main() {
    result := helper()
    println(result)
}
`

    provider := NewGoDefinitionProvider("/workspace")

    // Test finding definition of "helper" call
    position := core.Position{Line: 7, Character: 15} // Position of "helper"

    locations := provider.ProvideDefinition("file:///test.go", content, position)

    if len(locations) != 1 {
        t.Fatalf("expected 1 location, got %d", len(locations))
    }

    loc := locations[0]

    // Should point to function definition on line 2
    if loc.Range.Start.Line != 2 {
        t.Errorf("definition line = %d, want 2", loc.Range.Start.Line)
    }

    if loc.URI != "file:///test.go" {
        t.Errorf("definition URI = %s, want file:///test.go", loc.URI)
    }
}
```

### Testing Hover Provider

```go
func TestGoHoverProvider(t *testing.T) {
    content := `package main

// Add returns the sum of two integers.
func Add(a, b int) int {
    return a + b
}

func main() {
    result := Add(1, 2)
    println(result)
}
`

    provider := &GoHoverProvider{}

    // Test hovering over "Add" function call
    position := core.Position{Line: 8, Character: 15}

    hover := provider.ProvideHover("file:///test.go", content, position)

    if hover == nil {
        t.Fatal("expected hover info, got nil")
    }

    // Should contain function signature
    if !strings.Contains(hover.Contents, "func Add(a, b int) int") {
        t.Errorf("hover contents missing signature: %s", hover.Contents)
    }

    // Should contain documentation
    if !strings.Contains(hover.Contents, "returns the sum") {
        t.Errorf("hover contents missing documentation: %s", hover.Contents)
    }

    // Should have a range
    if hover.Range == nil {
        t.Error("hover range is nil")
    }
}
```

## LSP Server Integration

### Handler Implementation

```go
import (
    "github.com/SCKelemen/lsp"
    "github.com/SCKelemen/lsp/adapter"
    "github.com/SCKelemen/lsp/core"
    protocol "github.com/SCKelemen/lsp/protocol"
)

type MyServer struct {
    documents  *core.DocumentManager
    definition core.DefinitionProvider
    hover      core.HoverProvider
}

func NewMyServer(workspace string) *MyServer {
    return &MyServer{
        documents:  core.NewDocumentManager(),
        definition: NewGoDefinitionProvider(workspace),
        hover:      &GoHoverProvider{},
    }
}

func (s *MyServer) TextDocumentDefinition(
    context *lsp.Context,
    params *protocol.DefinitionParams,
) ([]protocol.Location, error) {
    uri := string(params.TextDocument.URI)
    content := s.documents.GetContent(uri)

    if content == "" {
        return nil, nil
    }

    // Convert protocol position to core position
    corePos := adapter.ProtocolToCorePosition(params.Position, content)

    // Get definitions
    coreLocations := s.definition.ProvideDefinition(uri, content, corePos)

    // Convert to protocol locations
    var protocolLocations []protocol.Location
    for _, loc := range coreLocations {
        // Get content for the target file
        targetContent := s.documents.GetContent(loc.URI)
        if targetContent == "" {
            // If not in memory, read from disk
            targetContent = content // Fallback to same file
        }

        protocolLocations = append(protocolLocations, protocol.Location{
            URI:   protocol.DocumentURI(loc.URI),
            Range: adapter.CoreToProtocolRange(loc.Range, targetContent),
        })
    }

    return protocolLocations, nil
}

func (s *MyServer) TextDocumentHover(
    context *lsp.Context,
    params *protocol.HoverParams,
) (*protocol.Hover, error) {
    uri := string(params.TextDocument.URI)
    content := s.documents.GetContent(uri)

    if content == "" {
        return nil, nil
    }

    // Convert protocol position to core position
    corePos := adapter.ProtocolToCorePosition(params.Position, content)

    // Get hover info
    hover := s.hover.ProvideHover(uri, content, corePos)
    if hover == nil {
        return nil, nil
    }

    // Convert to protocol hover
    result := &protocol.Hover{
        Contents: protocol.MarkupContent{
            Kind:  protocol.MarkupKindMarkdown,
            Value: hover.Contents,
        },
    }

    if hover.Range != nil {
        r := adapter.CoreToProtocolRange(*hover.Range, content)
        result.Range = &r
    }

    return result, nil
}
```

### Server Capabilities

```go
func (s *MyServer) Initialize(
    context *lsp.Context,
    params *protocol.InitializeParams,
) (interface{}, error) {
    capabilities := protocol.ServerCapabilities{
        DefinitionProvider: true,
        HoverProvider:      true,
    }

    return protocol.InitializeResult{
        Capabilities: capabilities,
    }, nil
}
```

## Summary

You now know how to:

1. ✅ Implement definition providers with the `DefinitionProvider` interface
2. ✅ Implement hover providers with the `HoverProvider` interface
3. ✅ Parse code and find symbols at specific positions
4. ✅ Generate rich hover content with markdown formatting
5. ✅ Test navigation providers thoroughly
6. ✅ Integrate navigation features into an LSP server

**Key Points:**
- Convert positions accurately between core and protocol types
- Support cross-file navigation for go-to-definition
- Use markdown for rich hover content with syntax highlighting
- Include documentation comments in hover information
- Test with various cursor positions and code structures

**Next Steps:**
- See [CODE_ACTIONS.md](CODE_ACTIONS.md) for code action providers
- See [SYMBOLS.md](SYMBOLS.md) for document symbol providers
- Check `examples/` for complete working examples
- Read [CORE_TYPES.md](CORE_TYPES.md) for UTF-8/UTF-16 details
