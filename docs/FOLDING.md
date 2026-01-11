# Folding Ranges Guide

Complete guide to implementing folding range providers for collapsible code regions.

**LSP Specification References:**
- [Folding Range Request](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#textDocument_foldingRange)
- [Folding Range Type](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#foldingRange)

## Table of Contents

1. [Core Concepts](#core-concepts)
2. [Your First Folding Provider](#your-first-folding-provider)
3. [Different Folding Strategies](#different-folding-strategies)
4. [Testing Folding Providers](#testing-folding-providers)
5. [LSP Server Integration](#lsp-server-integration)

## Core Concepts

### What Are Folding Ranges?

Folding ranges identify sections of code that can be collapsed in the editor, such as:
- **Functions and methods**: Collapse function bodies
- **Comments**: Collapse multi-line comments
- **Imports**: Collapse import blocks
- **Regions**: Collapse marked regions (#region/#endregion)
- **Control structures**: Collapse if/for/while blocks

### Provider Interface

```go
import "github.com/SCKelemen/lsp/core"

// FoldingRangeProvider provides folding ranges for a document
type FoldingRangeProvider interface {
    ProvideFoldingRanges(uri, content string) []FoldingRange
}
```

### Folding Range Structure

```go
type FoldingRange struct {
    StartLine      int                // Zero-based start line
    StartCharacter *int               // Optional UTF-8 byte offset on start line
    EndLine        int                // Zero-based end line
    EndCharacter   *int               // Optional UTF-8 byte offset on end line
    Kind           *FoldingRangeKind  // comment, imports, region, or nil
}
```

### Folding Range Kinds

```go
const (
    FoldingRangeKindComment = "comment"
    FoldingRangeKindImports = "imports"
    FoldingRangeKindRegion  = "region"
)
```

**Note**: When `Kind` is nil, it defaults to a generic code fold (like functions or blocks).

## Your First Folding Provider

Let's create a provider that identifies foldable regions in Go code.

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

// GoFoldingProvider provides folding ranges for Go source code.
type GoFoldingProvider struct{}

func (p *GoFoldingProvider) ProvideFoldingRanges(uri, content string) []core.FoldingRange {
    // Only handle Go files
    if !strings.HasSuffix(uri, ".go") {
        return nil
    }

    var ranges []core.FoldingRange

    // Parse the file
    fset := token.NewFileSet()
    f, err := parser.ParseFile(fset, "", content, parser.ParseComments)
    if err != nil {
        return nil
    }

    // Add import folding
    if importRange := p.getImportFolding(f, fset); importRange != nil {
        ranges = append(ranges, *importRange)
    }

    // Add comment folding
    ranges = append(ranges, p.getCommentFolding(f, fset)...)

    // Add function folding
    ranges = append(ranges, p.getFunctionFolding(f, fset)...)

    return ranges
}
```

### Step 2: Implement Import Folding

```go
func (p *GoFoldingProvider) getImportFolding(f *ast.File, fset *token.FileSet) *core.FoldingRange {
    if len(f.Imports) < 2 {
        // Don't fold single imports
        return nil
    }

    // Get first and last import positions
    first := fset.Position(f.Imports[0].Pos())
    last := fset.Position(f.Imports[len(f.Imports)-1].End())

    // Only fold if imports span multiple lines
    if first.Line == last.Line {
        return nil
    }

    kind := core.FoldingRangeKindImports

    return &core.FoldingRange{
        StartLine: first.Line - 1, // Convert to 0-based
        EndLine:   last.Line - 1,
        Kind:      &kind,
    }
}
```

### Step 3: Implement Comment Folding

```go
func (p *GoFoldingProvider) getCommentFolding(f *ast.File, fset *token.FileSet) []core.FoldingRange {
    var ranges []core.FoldingRange

    kind := core.FoldingRangeKindComment

    for _, cg := range f.Comments {
        // Only fold multi-line comments
        if len(cg.List) < 2 {
            continue
        }

        start := fset.Position(cg.Pos())
        end := fset.Position(cg.End())

        // Only fold if comment spans multiple lines
        if start.Line == end.Line {
            continue
        }

        ranges = append(ranges, core.FoldingRange{
            StartLine: start.Line - 1,
            EndLine:   end.Line - 1,
            Kind:      &kind,
        })
    }

    return ranges
}
```

### Step 4: Implement Function Folding

```go
func (p *GoFoldingProvider) getFunctionFolding(f *ast.File, fset *token.FileSet) []core.FoldingRange {
    var ranges []core.FoldingRange

    // Walk the AST looking for functions
    ast.Inspect(f, func(n ast.Node) bool {
        fn, ok := n.(*ast.FuncDecl)
        if !ok {
            return true
        }

        // Only fold if function has a body
        if fn.Body == nil {
            return true
        }

        // Get the brace positions
        start := fset.Position(fn.Body.Lbrace)
        end := fset.Position(fn.Body.Rbrace)

        // Only fold multi-line functions
        if start.Line == end.Line {
            return true
        }

        // Start folding after the opening brace
        startLine := start.Line - 1
        startChar := start.Column - 1
        endLine := end.Line - 1

        ranges = append(ranges, core.FoldingRange{
            StartLine:      startLine,
            StartCharacter: &startChar,
            EndLine:        endLine,
        })

        return true
    })

    return ranges
}
```

### Step 5: Use the Provider

```go
func main() {
    content := `package main

import (
    "fmt"
    "os"
)

// This is a multi-line comment
// that should be foldable
// across several lines

func main() {
    fmt.Println("Hello")
    if true {
        fmt.Println("World")
    }
}

func helper() {
    // Single line comment
    os.Exit(0)
}
`

    provider := &GoFoldingProvider{}
    ranges := provider.ProvideFoldingRanges("file:///example.go", content)

    for _, r := range ranges {
        kindStr := "code"
        if r.Kind != nil {
            kindStr = string(*r.Kind)
        }
        fmt.Printf("Fold [%d:%d] (%s)\n", r.StartLine+1, r.EndLine+1, kindStr)
    }
}
```

## Different Folding Strategies

### Brace-Based Folding (Generic)

For languages with brace-delimited blocks:

```go
type BraceFoldingProvider struct{}

func (p *BraceFoldingProvider) ProvideFoldingRanges(uri, content string) []core.FoldingRange {
    var ranges []core.FoldingRange

    lines := strings.Split(content, "\n")
    stack := []int{} // Stack of opening brace line numbers

    for lineNum, line := range lines {
        // Track opening braces
        for charPos, ch := range line {
            if ch == '{' {
                stack = append(stack, lineNum)
            } else if ch == '}' {
                if len(stack) > 0 {
                    // Pop from stack
                    startLine := stack[len(stack)-1]
                    stack = stack[:len(stack)-1]

                    // Only fold multi-line blocks
                    if lineNum > startLine {
                        ranges = append(ranges, core.FoldingRange{
                            StartLine: startLine,
                            EndLine:   lineNum,
                        })
                    }
                }
            }
        }
    }

    return ranges
}
```

### Indentation-Based Folding (Python, YAML)

For languages that use indentation:

```go
type IndentFoldingProvider struct {
    TabSize int // Number of spaces per indentation level
}

func (p *IndentFoldingProvider) ProvideFoldingRanges(uri, content string) []core.FoldingRange {
    if p.TabSize == 0 {
        p.TabSize = 4
    }

    var ranges []core.FoldingRange

    lines := strings.Split(content, "\n")
    stack := []indentBlock{} // Stack of indent blocks

    for lineNum, line := range lines {
        // Skip blank lines
        if strings.TrimSpace(line) == "" {
            continue
        }

        // Calculate indentation level
        indent := p.getIndentLevel(line)

        // Pop blocks with greater or equal indentation
        for len(stack) > 0 && stack[len(stack)-1].indent >= indent {
            block := stack[len(stack)-1]
            stack = stack[:len(stack)-1]

            // Create folding range if it spans multiple lines
            if lineNum-1 > block.startLine {
                ranges = append(ranges, core.FoldingRange{
                    StartLine: block.startLine,
                    EndLine:   lineNum - 1,
                })
            }
        }

        // Push new block
        stack = append(stack, indentBlock{
            startLine: lineNum,
            indent:    indent,
        })
    }

    // Handle remaining blocks
    endLine := len(lines) - 1
    for len(stack) > 0 {
        block := stack[len(stack)-1]
        stack = stack[:len(stack)-1]

        if endLine > block.startLine {
            ranges = append(ranges, core.FoldingRange{
                StartLine: block.startLine,
                EndLine:   endLine,
            })
        }
    }

    return ranges
}

type indentBlock struct {
    startLine int
    indent    int
}

func (p *IndentFoldingProvider) getIndentLevel(line string) int {
    spaces := 0
    for _, ch := range line {
        if ch == ' ' {
            spaces++
        } else if ch == '\t' {
            spaces += p.TabSize
        } else {
            break
        }
    }
    return spaces / p.TabSize
}
```

### Region-Based Folding

Support explicit region markers:

```go
type RegionFoldingProvider struct {
    StartMarker string // e.g., "#region" or "// region"
    EndMarker   string // e.g., "#endregion" or "// endregion"
}

func (p *RegionFoldingProvider) ProvideFoldingRanges(uri, content string) []core.FoldingRange {
    var ranges []core.FoldingRange

    lines := strings.Split(content, "\n")
    stack := []int{} // Stack of region start line numbers

    kind := core.FoldingRangeKindRegion

    for lineNum, line := range lines {
        trimmed := strings.TrimSpace(line)

        if strings.Contains(trimmed, p.StartMarker) {
            stack = append(stack, lineNum)
        } else if strings.Contains(trimmed, p.EndMarker) {
            if len(stack) > 0 {
                startLine := stack[len(stack)-1]
                stack = stack[:len(stack)-1]

                ranges = append(ranges, core.FoldingRange{
                    StartLine: startLine,
                    EndLine:   lineNum,
                    Kind:      &kind,
                })
            }
        }
    }

    return ranges
}
```

### Composite Provider

Combine multiple folding strategies:

```go
type CompositeFoldingProvider struct {
    providers []core.FoldingRangeProvider
}

func NewCompositeFoldingProvider() *CompositeFoldingProvider {
    return &CompositeFoldingProvider{
        providers: []core.FoldingRangeProvider{
            &GoFoldingProvider{},
            &RegionFoldingProvider{
                StartMarker: "// region",
                EndMarker:   "// endregion",
            },
        },
    }
}

func (p *CompositeFoldingProvider) ProvideFoldingRanges(uri, content string) []core.FoldingRange {
    var allRanges []core.FoldingRange

    // Collect ranges from all providers
    for _, provider := range p.providers {
        if ranges := provider.ProvideFoldingRanges(uri, content); len(ranges) > 0 {
            allRanges = append(allRanges, ranges...)
        }
    }

    // Remove duplicates and overlaps
    return p.deduplicateRanges(allRanges)
}

func (p *CompositeFoldingProvider) deduplicateRanges(ranges []core.FoldingRange) []core.FoldingRange {
    if len(ranges) == 0 {
        return ranges
    }

    // Sort by start line
    sort.Slice(ranges, func(i, j int) bool {
        if ranges[i].StartLine == ranges[j].StartLine {
            return ranges[i].EndLine < ranges[j].EndLine
        }
        return ranges[i].StartLine < ranges[j].StartLine
    })

    // Remove exact duplicates
    result := []core.FoldingRange{ranges[0]}

    for i := 1; i < len(ranges); i++ {
        curr := ranges[i]
        prev := result[len(result)-1]

        // Skip if identical
        if curr.StartLine == prev.StartLine && curr.EndLine == prev.EndLine {
            continue
        }

        result = append(result, curr)
    }

    return result
}
```

## Testing Folding Providers

### Basic Test Structure

```go
func TestGoFoldingProvider(t *testing.T) {
    tests := []struct {
        name      string
        content   string
        wantCount int
        wantKinds []string
    }{
        {
            name: "simple function",
            content: `package main

func main() {
    println("hello")
}`,
            wantCount: 1,
            wantKinds: []string{"code"},
        },
        {
            name: "imports block",
            content: `package main

import (
    "fmt"
    "os"
)

func main() {}`,
            wantCount: 1,
            wantKinds: []string{"imports"},
        },
        {
            name: "multi-line comment",
            content: `package main

// This is a comment
// spanning multiple
// lines

func main() {}`,
            wantCount: 2, // Comment + function
            wantKinds: []string{"comment", "code"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            provider := &GoFoldingProvider{}
            ranges := provider.ProvideFoldingRanges("file:///test.go", tt.content)

            if len(ranges) != tt.wantCount {
                t.Errorf("got %d ranges, want %d", len(ranges), tt.wantCount)
                for i, r := range ranges {
                    kind := "code"
                    if r.Kind != nil {
                        kind = string(*r.Kind)
                    }
                    t.Logf("  range %d: lines %d-%d (%s)", i, r.StartLine, r.EndLine, kind)
                }
            }

            // Verify kinds if specified
            if tt.wantKinds != nil {
                for i, r := range ranges {
                    if i >= len(tt.wantKinds) {
                        break
                    }

                    expectedKind := tt.wantKinds[i]
                    actualKind := "code"
                    if r.Kind != nil {
                        actualKind = string(*r.Kind)
                    }

                    if actualKind != expectedKind {
                        t.Errorf("range %d: got kind %q, want %q", i, actualKind, expectedKind)
                    }
                }
            }
        })
    }
}
```

### Testing Line Ranges

```go
func TestFoldingRangeAccuracy(t *testing.T) {
    content := `package main

import (
    "fmt"
)

func main() {
    fmt.Println("hello")
}
`

    provider := &GoFoldingProvider{}
    ranges := provider.ProvideFoldingRanges("file:///test.go", content)

    // Find the function fold
    var fnFold *core.FoldingRange
    for i := range ranges {
        if ranges[i].Kind == nil {
            fnFold = &ranges[i]
            break
        }
    }

    if fnFold == nil {
        t.Fatal("function fold not found")
    }

    // Function should start at line 6 (func main() {)
    // and end at line 8 (})
    if fnFold.StartLine != 6 {
        t.Errorf("start line = %d, want 6", fnFold.StartLine)
    }

    if fnFold.EndLine != 8 {
        t.Errorf("end line = %d, want 8", fnFold.EndLine)
    }
}
```

### Testing No-Fold Scenarios

```go
func TestNoFoldingForSingleLine(t *testing.T) {
    tests := []struct {
        name    string
        content string
    }{
        {
            name: "single line function",
            content: `package main

func main() { println("hello") }`,
        },
        {
            name: "single import",
            content: `package main

import "fmt"

func main() {}`,
        },
        {
            name: "single line comment",
            content: `package main

// Single comment

func main() {}`,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            provider := &GoFoldingProvider{}
            ranges := provider.ProvideFoldingRanges("file:///test.go", tt.content)

            // Should have at most 1 range (the main function if multi-line)
            if len(ranges) > 1 {
                t.Errorf("got %d ranges, expected 0 or 1", len(ranges))
            }
        })
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
    documents *core.DocumentManager
    folding   core.FoldingRangeProvider
}

func NewMyServer() *MyServer {
    return &MyServer{
        documents: core.NewDocumentManager(),
        folding: &CompositeFoldingProvider{
            providers: []core.FoldingRangeProvider{
                &GoFoldingProvider{},
                &RegionFoldingProvider{
                    StartMarker: "// region",
                    EndMarker:   "// endregion",
                },
            },
        },
    }
}

func (s *MyServer) TextDocumentFoldingRange(
    context *lsp.Context,
    params *protocol.FoldingRangeParams,
) ([]protocol.FoldingRange, error) {
    uri := string(params.TextDocument.URI)
    content := s.documents.GetContent(uri)

    if content == "" {
        return nil, nil
    }

    // Get folding ranges using core types
    coreRanges := s.folding.ProvideFoldingRanges(uri, content)

    // Convert to protocol ranges
    var protocolRanges []protocol.FoldingRange
    for _, r := range coreRanges {
        pr := protocol.FoldingRange{
            StartLine: protocol.UInteger(r.StartLine),
            EndLine:   protocol.UInteger(r.EndLine),
        }

        if r.StartCharacter != nil {
            // Convert UTF-8 byte offset to UTF-16 code units
            lines := strings.Split(content, "\n")
            if r.StartLine < len(lines) {
                lineContent := lines[r.StartLine]
                utf16Offset := core.UTF8ToUTF16Offset(lineContent, 0, *r.StartCharacter)
                startChar := protocol.UInteger(utf16Offset)
                pr.StartCharacter = &startChar
            }
        }

        if r.EndCharacter != nil {
            // Convert UTF-8 byte offset to UTF-16 code units
            lines := strings.Split(content, "\n")
            if r.EndLine < len(lines) {
                lineContent := lines[r.EndLine]
                utf16Offset := core.UTF8ToUTF16Offset(lineContent, 0, *r.EndCharacter)
                endChar := protocol.UInteger(utf16Offset)
                pr.EndCharacter = &endChar
            }
        }

        if r.Kind != nil {
            kind := protocol.FoldingRangeKind(*r.Kind)
            pr.Kind = &kind
        }

        protocolRanges = append(protocolRanges, pr)
    }

    return protocolRanges, nil
}
```

### Server Capabilities

Advertise folding range support:

```go
func (s *MyServer) Initialize(
    context *lsp.Context,
    params *protocol.InitializeParams,
) (interface{}, error) {
    capabilities := protocol.ServerCapabilities{
        FoldingRangeProvider: true,
    }

    return protocol.InitializeResult{
        Capabilities: capabilities,
    }, nil
}
```

## Summary

You now know how to:

1. ✅ Implement folding range providers with the `FoldingRangeProvider` interface
2. ✅ Create different folding strategies (brace-based, indentation-based, region-based)
3. ✅ Combine multiple providers for comprehensive folding support
4. ✅ Test folding providers with various code structures
5. ✅ Integrate folding into an LSP server with UTF-8/UTF-16 conversion

**Key Points:**
- Folding ranges use line-based positions (start/end line)
- Optional character offsets provide more precise folding
- Use appropriate `Kind` values (comment, imports, region)
- Only fold multi-line regions
- Avoid overlapping or duplicate ranges

**Next Steps:**
- See [CODE_ACTIONS.md](CODE_ACTIONS.md) for code action providers
- See [VALIDATORS.md](VALIDATORS.md) for diagnostic providers
- Check `examples/` for complete working examples
- Read [CORE_TYPES.md](CORE_TYPES.md) for UTF-8/UTF-16 details
