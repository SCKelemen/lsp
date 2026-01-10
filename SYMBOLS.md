# Document Symbols Guide

Complete guide to implementing document symbol providers for code outline and navigation.

**LSP Specification References:**
- [Document Symbols Request](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#textDocument_documentSymbol)
- [Document Symbol Type](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#documentSymbol)
- [Symbol Kind](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#symbolKind)

## Table of Contents

1. [Core Concepts](#core-concepts)
2. [Your First Symbol Provider](#your-first-symbol-provider)
3. [Hierarchical Symbols](#hierarchical-symbols)
4. [Testing Symbol Providers](#testing-symbol-providers)
5. [LSP Server Integration](#lsp-server-integration)

## Core Concepts

### What Are Document Symbols?

Document symbols represent the structure of a document, providing:
- **Outline view**: Show document structure in a tree
- **Breadcrumb navigation**: Display current location
- **Quick navigation**: Jump to symbols
- **Code structure**: Functions, classes, variables, etc.

### Provider Interface

```go
import "github.com/SCKelemen/lsp/core"

// DocumentSymbolProvider provides document symbols
type DocumentSymbolProvider interface {
    ProvideDocumentSymbols(uri, content string) []DocumentSymbol
}
```

### Document Symbol Structure

```go
type DocumentSymbol struct {
    Name           string           // Symbol name
    Detail         string           // Additional detail (e.g., signature)
    Kind           SymbolKind       // Type of symbol
    Tags           []SymbolTag      // Tags (e.g., deprecated)
    Deprecated     bool             // Whether deprecated
    Range          Range            // Full range including comments
    SelectionRange Range            // Range of the symbol name
    Children       []DocumentSymbol // Child symbols
}
```

### Symbol Kinds

```go
const (
    SymbolKindFile          = 1
    SymbolKindModule        = 2
    SymbolKindNamespace     = 3
    SymbolKindPackage       = 4
    SymbolKindClass         = 5
    SymbolKindMethod        = 6
    SymbolKindProperty      = 7
    SymbolKindField         = 8
    SymbolKindConstructor   = 9
    SymbolKindEnum          = 10
    SymbolKindInterface     = 11
    SymbolKindFunction      = 12
    SymbolKindVariable      = 13
    SymbolKindConstant      = 14
    SymbolKindString        = 15
    SymbolKindNumber        = 16
    SymbolKindBoolean       = 17
    SymbolKindArray         = 18
    SymbolKindObject        = 19
    SymbolKindKey           = 20
    SymbolKindNull          = 21
    SymbolKindEnumMember    = 22
    SymbolKindStruct        = 23
    SymbolKindEvent         = 24
    SymbolKindOperator      = 25
    SymbolKindTypeParameter = 26
)
```

## Your First Symbol Provider

Let's create a provider for Go source code.

### Step 1: Define the Provider

```go
package mypackage

import (
    "go/ast"
    "go/parser"
    "go/token"
    "strings"

    "github.com/SCKelemen/lsp/core"
)

// GoSymbolProvider provides symbols for Go source files.
type GoSymbolProvider struct{}

func (p *GoSymbolProvider) ProvideDocumentSymbols(uri, content string) []core.DocumentSymbol {
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

    var symbols []core.DocumentSymbol

    // Add package-level declarations
    for _, decl := range f.Decls {
        if symbol := p.declToSymbol(decl, fset, content); symbol != nil {
            symbols = append(symbols, *symbol)
        }
    }

    return symbols
}
```

### Step 2: Convert Declarations to Symbols

```go
func (p *GoSymbolProvider) declToSymbol(
    decl ast.Decl,
    fset *token.FileSet,
    content string,
) *core.DocumentSymbol {
    switch d := decl.(type) {
    case *ast.FuncDecl:
        return p.funcToSymbol(d, fset, content)
    case *ast.GenDecl:
        return p.genDeclToSymbol(d, fset, content)
    default:
        return nil
    }
}
```

### Step 3: Handle Function Declarations

```go
func (p *GoSymbolProvider) funcToSymbol(
    fn *ast.FuncDecl,
    fset *token.FileSet,
    content string,
) *core.DocumentSymbol {
    // Get positions
    start := fset.Position(fn.Pos())
    end := fset.Position(fn.End())
    nameStart := fset.Position(fn.Name.Pos())
    nameEnd := fset.Position(fn.Name.End())

    // Determine if it's a method or function
    kind := core.SymbolKindFunction
    name := fn.Name.Name
    detail := p.getFunctionSignature(fn)

    if fn.Recv != nil {
        // It's a method
        kind = core.SymbolKindMethod

        // Add receiver type to detail
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

    // Add function parameters and local variables as children
    // (optional - for more detailed outline)
    if fn.Type.Params != nil {
        for _, field := range fn.Type.Params.List {
            for _, name := range field.Names {
                symbol.Children = append(symbol.Children, core.DocumentSymbol{
                    Name:   name.Name,
                    Detail: p.exprToString(field.Type),
                    Kind:   core.SymbolKindVariable,
                    Range:  p.nodeToRange(name, fset),
                    SelectionRange: p.nodeToRange(name, fset),
                })
            }
        }
    }

    return symbol
}

func (p *GoSymbolProvider) getFunctionSignature(fn *ast.FuncDecl) string {
    var sig strings.Builder
    sig.WriteString(fn.Name.Name)
    sig.WriteString("(")

    // Parameters
    if fn.Type.Params != nil {
        for i, field := range fn.Type.Params.List {
            if i > 0 {
                sig.WriteString(", ")
            }
            for j, name := range field.Names {
                if j > 0 {
                    sig.WriteString(", ")
                }
                sig.WriteString(name.Name)
            }
            sig.WriteString(" ")
            sig.WriteString(p.exprToString(field.Type))
        }
    }

    sig.WriteString(")")

    // Return type
    if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
        sig.WriteString(" ")
        if len(fn.Type.Results.List) == 1 && len(fn.Type.Results.List[0].Names) == 0 {
            sig.WriteString(p.exprToString(fn.Type.Results.List[0].Type))
        } else {
            sig.WriteString("(")
            for i, field := range fn.Type.Results.List {
                if i > 0 {
                    sig.WriteString(", ")
                }
                sig.WriteString(p.exprToString(field.Type))
            }
            sig.WriteString(")")
        }
    }

    return sig.String()
}
```

### Step 4: Handle General Declarations

```go
func (p *GoSymbolProvider) genDeclToSymbol(
    gen *ast.GenDecl,
    fset *token.FileSet,
    content string,
) *core.DocumentSymbol {
    // Get the declaration range
    start := fset.Position(gen.Pos())
    end := fset.Position(gen.End())

    var symbols []core.DocumentSymbol

    // Process each specification
    for _, spec := range gen.Specs {
        switch s := spec.(type) {
        case *ast.TypeSpec:
            symbols = append(symbols, p.typeSpecToSymbol(s, gen.Tok, fset)...)
        case *ast.ValueSpec:
            symbols = append(symbols, p.valueSpecToSymbol(s, gen.Tok, fset)...)
        }
    }

    // If only one symbol, return it directly
    if len(symbols) == 1 {
        return &symbols[0]
    }

    // If multiple symbols in a group, create a container
    if len(symbols) > 1 {
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

func (p *GoSymbolProvider) typeSpecToSymbol(
    spec *ast.TypeSpec,
    tok token.Token,
    fset *token.FileSet,
) []core.DocumentSymbol {
    nameStart := fset.Position(spec.Name.Pos())
    nameEnd := fset.Position(spec.Name.End())
    typeStart := fset.Position(spec.Pos())
    typeEnd := fset.Position(spec.End())

    kind := core.SymbolKindClass

    // Determine specific kind based on type
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

func (p *GoSymbolProvider) valueSpecToSymbol(
    spec *ast.ValueSpec,
    tok token.Token,
    fset *token.FileSet,
) []core.DocumentSymbol {
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
```

### Step 5: Helper Functions

```go
func (p *GoSymbolProvider) getStructFields(
    st *ast.StructType,
    fset *token.FileSet,
) []core.DocumentSymbol {
    var fields []core.DocumentSymbol

    if st.Fields == nil {
        return fields
    }

    for _, field := range st.Fields.List {
        for _, name := range field.Names {
            fields = append(fields, core.DocumentSymbol{
                Name:           name.Name,
                Detail:         p.exprToString(field.Type),
                Kind:           core.SymbolKindField,
                Range:          p.nodeToRange(field, fset),
                SelectionRange: p.nodeToRange(name, fset),
            })
        }
    }

    return fields
}

func (p *GoSymbolProvider) getInterfaceMethods(
    it *ast.InterfaceType,
    fset *token.FileSet,
) []core.DocumentSymbol {
    var methods []core.DocumentSymbol

    if it.Methods == nil {
        return methods
    }

    for _, method := range it.Methods.List {
        for _, name := range method.Names {
            methods = append(methods, core.DocumentSymbol{
                Name:           name.Name,
                Detail:         p.exprToString(method.Type),
                Kind:           core.SymbolKindMethod,
                Range:          p.nodeToRange(method, fset),
                SelectionRange: p.nodeToRange(name, fset),
            })
        }
    }

    return methods
}

func (p *GoSymbolProvider) exprToString(expr ast.Expr) string {
    if expr == nil {
        return ""
    }

    // Simple string representation
    // In real code, use go/types for accurate type strings
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
    default:
        return "..."
    }
}

func (p *GoSymbolProvider) nodeToRange(node ast.Node, fset *token.FileSet) core.Range {
    start := fset.Position(node.Pos())
    end := fset.Position(node.End())

    return core.Range{
        Start: core.Position{Line: start.Line - 1, Character: start.Column - 1},
        End:   core.Position{Line: end.Line - 1, Character: end.Column - 1},
    }
}
```

## Hierarchical Symbols

### Understanding Symbol Hierarchy

Document symbols can be nested to represent code structure:

```
Package
├── Import declarations
├── Type: User (struct)
│   ├── Field: Name (string)
│   ├── Field: Email (string)
│   └── Method: String() string
├── Function: NewUser(name, email string) *User
└── Function: main()
```

### Example: Markdown Symbols

```go
type MarkdownSymbolProvider struct{}

func (p *MarkdownSymbolProvider) ProvideDocumentSymbols(uri, content string) []core.DocumentSymbol {
    if !strings.HasSuffix(uri, ".md") {
        return nil
    }

    var symbols []core.DocumentSymbol
    stack := []*core.DocumentSymbol{} // Stack for hierarchy

    lines := strings.Split(content, "\n")

    for lineNum, line := range lines {
        // Check if line is a heading
        if !strings.HasPrefix(line, "#") {
            continue
        }

        // Determine heading level
        level := 0
        for _, ch := range line {
            if ch == '#' {
                level++
            } else {
                break
            }
        }

        // Extract heading text
        title := strings.TrimSpace(line[level:])

        // Create symbol
        symbol := core.DocumentSymbol{
            Name:   title,
            Detail: strings.Repeat("#", level),
            Kind:   core.SymbolKindString, // or custom kind
            Range: core.Range{
                Start: core.Position{Line: lineNum, Character: 0},
                End:   core.Position{Line: lineNum, Character: len(line)},
            },
            SelectionRange: core.Range{
                Start: core.Position{Line: lineNum, Character: level + 1},
                End:   core.Position{Line: lineNum, Character: len(line)},
            },
        }

        // Pop stack until we find the parent level
        for len(stack) > 0 && len(stack) >= level {
            stack = stack[:len(stack)-1]
        }

        if len(stack) == 0 {
            // Top-level symbol
            symbols = append(symbols, symbol)
            stack = append(stack, &symbols[len(symbols)-1])
        } else {
            // Child symbol
            parent := stack[len(stack)-1]
            parent.Children = append(parent.Children, symbol)
            stack = append(stack, &parent.Children[len(parent.Children)-1])
        }
    }

    return symbols
}
```

### Example: JSON Symbols

```go
type JSONSymbolProvider struct{}

func (p *JSONSymbolProvider) ProvideDocumentSymbols(uri, content string) []core.DocumentSymbol {
    if !strings.HasSuffix(uri, ".json") {
        return nil
    }

    var data interface{}
    if err := json.Unmarshal([]byte(content), &data); err != nil {
        return nil
    }

    return p.valueToSymbols(data, "", content, 0)
}

func (p *JSONSymbolProvider) valueToSymbols(
    value interface{},
    name string,
    content string,
    offset int,
) []core.DocumentSymbol {
    var symbols []core.DocumentSymbol

    switch v := value.(type) {
    case map[string]interface{}:
        kind := core.SymbolKindObject
        symbol := core.DocumentSymbol{
            Name:   name,
            Kind:   kind,
            Range:  p.findRangeInContent(content, name, offset),
            SelectionRange: p.findRangeInContent(content, name, offset),
        }

        // Add object properties as children
        for key, val := range v {
            children := p.valueToSymbols(val, key, content, offset)
            symbol.Children = append(symbol.Children, children...)
        }

        symbols = append(symbols, symbol)

    case []interface{}:
        kind := core.SymbolKindArray
        symbol := core.DocumentSymbol{
            Name:   name,
            Detail: fmt.Sprintf("[%d items]", len(v)),
            Kind:   kind,
            Range:  p.findRangeInContent(content, name, offset),
            SelectionRange: p.findRangeInContent(content, name, offset),
        }

        // Optionally add array items as children
        for i, item := range v {
            itemName := fmt.Sprintf("[%d]", i)
            children := p.valueToSymbols(item, itemName, content, offset)
            symbol.Children = append(symbol.Children, children...)
        }

        symbols = append(symbols, symbol)

    case string:
        symbols = append(symbols, core.DocumentSymbol{
            Name:   name,
            Detail: fmt.Sprintf("\"%s\"", v),
            Kind:   core.SymbolKindString,
            Range:  p.findRangeInContent(content, name, offset),
            SelectionRange: p.findRangeInContent(content, name, offset),
        })

    case float64:
        symbols = append(symbols, core.DocumentSymbol{
            Name:   name,
            Detail: fmt.Sprintf("%v", v),
            Kind:   core.SymbolKindNumber,
            Range:  p.findRangeInContent(content, name, offset),
            SelectionRange: p.findRangeInContent(content, name, offset),
        })

    case bool:
        symbols = append(symbols, core.DocumentSymbol{
            Name:   name,
            Detail: fmt.Sprintf("%v", v),
            Kind:   core.SymbolKindBoolean,
            Range:  p.findRangeInContent(content, name, offset),
            SelectionRange: p.findRangeInContent(content, name, offset),
        })

    case nil:
        symbols = append(symbols, core.DocumentSymbol{
            Name:   name,
            Detail: "null",
            Kind:   core.SymbolKindNull,
            Range:  p.findRangeInContent(content, name, offset),
            SelectionRange: p.findRangeInContent(content, name, offset),
        })
    }

    return symbols
}

func (p *JSONSymbolProvider) findRangeInContent(content, name string, offset int) core.Range {
    // Simplified: find name in content starting from offset
    // Real implementation would need proper JSON parsing with positions
    idx := strings.Index(content[offset:], name)
    if idx == -1 {
        return core.Range{}
    }

    // Convert byte offset to line/character
    lineNum := strings.Count(content[:offset+idx], "\n")
    lineStart := strings.LastIndex(content[:offset+idx], "\n") + 1
    charOffset := offset + idx - lineStart

    return core.Range{
        Start: core.Position{Line: lineNum, Character: charOffset},
        End:   core.Position{Line: lineNum, Character: charOffset + len(name)},
    }
}
```

## Testing Symbol Providers

### Basic Test Structure

```go
func TestGoSymbolProvider(t *testing.T) {
    tests := []struct {
        name          string
        content       string
        wantCount     int
        wantNames     []string
        wantKinds     []core.SymbolKind
    }{
        {
            name: "simple function",
            content: `package main

func main() {
    println("hello")
}`,
            wantCount: 1,
            wantNames: []string{"main"},
            wantKinds: []core.SymbolKind{core.SymbolKindFunction},
        },
        {
            name: "struct with methods",
            content: `package main

type User struct {
    Name string
}

func (u *User) String() string {
    return u.Name
}`,
            wantCount: 2, // User struct + String method
            wantNames: []string{"User", "String"},
            wantKinds: []core.SymbolKind{
                core.SymbolKindStruct,
                core.SymbolKindMethod,
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            provider := &GoSymbolProvider{}
            symbols := provider.ProvideDocumentSymbols("file:///test.go", tt.content)

            if len(symbols) != tt.wantCount {
                t.Errorf("got %d symbols, want %d", len(symbols), tt.wantCount)
                for i, s := range symbols {
                    t.Logf("  symbol %d: %s (%d)", i, s.Name, s.Kind)
                }
            }

            for i, symbol := range symbols {
                if i >= len(tt.wantNames) {
                    break
                }

                if symbol.Name != tt.wantNames[i] {
                    t.Errorf("symbol %d: got name %q, want %q",
                        i, symbol.Name, tt.wantNames[i])
                }

                if i < len(tt.wantKinds) && symbol.Kind != tt.wantKinds[i] {
                    t.Errorf("symbol %d: got kind %d, want %d",
                        i, symbol.Kind, tt.wantKinds[i])
                }
            }
        })
    }
}
```

### Testing Symbol Hierarchy

```go
func TestSymbolHierarchy(t *testing.T) {
    content := `package main

type User struct {
    Name  string
    Email string
}

func (u *User) String() string {
    return u.Name
}
`

    provider := &GoSymbolProvider{}
    symbols := provider.ProvideDocumentSymbols("file:///test.go", content)

    // Find User struct
    var userSymbol *core.DocumentSymbol
    for i := range symbols {
        if symbols[i].Name == "User" {
            userSymbol = &symbols[i]
            break
        }
    }

    if userSymbol == nil {
        t.Fatal("User symbol not found")
    }

    // Check that struct has field children
    if len(userSymbol.Children) != 2 {
        t.Errorf("User struct has %d children, want 2", len(userSymbol.Children))
    }

    // Verify field names
    expectedFields := []string{"Name", "Email"}
    for i, child := range userSymbol.Children {
        if i >= len(expectedFields) {
            break
        }
        if child.Name != expectedFields[i] {
            t.Errorf("field %d: got %q, want %q", i, child.Name, expectedFields[i])
        }
        if child.Kind != core.SymbolKindField {
            t.Errorf("field %d: got kind %d, want %d", i, child.Kind, core.SymbolKindField)
        }
    }
}
```

## LSP Server Integration

### Handler Implementation

```go
import (
    "github.com/SCKelemen/lsp"
    "github.com/SCKelemen/lsp/adapter_3_16"
    "github.com/SCKelemen/lsp/core"
    protocol "github.com/SCKelemen/lsp/protocol_3_16"
)

type MyServer struct {
    documents *core.DocumentManager
    symbols   core.DocumentSymbolProvider
}

func NewMyServer() *MyServer {
    return &MyServer{
        documents: core.NewDocumentManager(),
        symbols:   &GoSymbolProvider{},
    }
}

func (s *MyServer) TextDocumentDocumentSymbol(
    context *glsp.Context,
    params *protocol.DocumentSymbolParams,
) ([]interface{}, error) {
    uri := string(params.TextDocument.URI)
    content := s.documents.GetContent(uri)

    if content == "" {
        return nil, nil
    }

    // Get symbols using core types
    coreSymbols := s.symbols.ProvideDocumentSymbols(uri, content)

    // Convert to protocol symbols
    var result []interface{}
    for _, symbol := range coreSymbols {
        protocolSymbol := s.coreToProtocolSymbol(symbol, content)
        result = append(result, protocolSymbol)
    }

    return result, nil
}

func (s *MyServer) coreToProtocolSymbol(
    symbol core.DocumentSymbol,
    content string,
) protocol.DocumentSymbol {
    result := protocol.DocumentSymbol{
        Name:           symbol.Name,
        Detail:         symbol.Detail,
        Kind:           protocol.SymbolKind(symbol.Kind),
        Deprecated:     symbol.Deprecated,
        Range:          adapter_3_16.CoreToProtocolRange(symbol.Range, content),
        SelectionRange: adapter_3_16.CoreToProtocolRange(symbol.SelectionRange, content),
    }

    if len(symbol.Tags) > 0 {
        for _, tag := range symbol.Tags {
            result.Tags = append(result.Tags, protocol.SymbolTag(tag))
        }
    }

    // Convert children recursively
    if len(symbol.Children) > 0 {
        for _, child := range symbol.Children {
            protocolChild := s.coreToProtocolSymbol(child, content)
            result.Children = append(result.Children, protocolChild)
        }
    }

    return result
}
```

### Server Capabilities

```go
func (s *MyServer) Initialize(
    context *glsp.Context,
    params *protocol.InitializeParams,
) (interface{}, error) {
    capabilities := protocol.ServerCapabilities{
        DocumentSymbolProvider: true,
    }

    return protocol.InitializeResult{
        Capabilities: capabilities,
    }, nil
}
```

## Summary

You now know how to:

1. ✅ Implement document symbol providers with the `DocumentSymbolProvider` interface
2. ✅ Create hierarchical symbol trees with parent-child relationships
3. ✅ Use appropriate symbol kinds for different code elements
4. ✅ Provide detailed information (signatures, types) for symbols
5. ✅ Test symbol providers thoroughly
6. ✅ Integrate symbol providers into an LSP server

**Key Points:**
- Use `Range` for the full symbol extent (including comments/whitespace)
- Use `SelectionRange` for the symbol name (what to highlight)
- Build hierarchy with `Children` for nested symbols
- Choose appropriate `SymbolKind` for each symbol type
- Support both flat and hierarchical symbol representations

**Next Steps:**
- See [CODE_ACTIONS.md](CODE_ACTIONS.md) for code action providers
- See [NAVIGATION.md](NAVIGATION.md) for definition and hover providers
- Check `examples/` for complete working examples
- Read [CORE_TYPES.md](CORE_TYPES.md) for UTF-8/UTF-16 details
