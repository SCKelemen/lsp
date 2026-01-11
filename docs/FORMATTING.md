# Formatting Providers Guide

Complete guide to implementing document and range formatting providers.

**LSP Specification References:**
- [Document Formatting](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#textDocument_formatting)
- [Range Formatting](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#textDocument_rangeFormatting)
- [On Type Formatting](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#textDocument_onTypeFormatting)

## Table of Contents

1. [Core Concepts](#core-concepts)
2. [Document Formatting Provider](#document-formatting-provider)
3. [Range Formatting Provider](#range-formatting-provider)
4. [Testing Formatting Providers](#testing-formatting-providers)
5. [LSP Server Integration](#lsp-server-integration)

## Core Concepts

### What Are Formatting Providers?

Formatting providers automatically format code according to style rules:
- **Document Formatting**: Format entire document
- **Range Formatting**: Format selected range
- **On-Type Formatting**: Format as user types

### Provider Interfaces

```go
import "github.com/SCKelemen/lsp/core"

// FormattingProvider provides document formatting
type FormattingProvider interface {
    ProvideFormatting(uri, content string, options FormattingOptions) []TextEdit
}

// RangeFormattingProvider provides range formatting
type RangeFormattingProvider interface {
    ProvideRangeFormatting(uri, content string, r Range, options FormattingOptions) []TextEdit
}
```

### Formatting Options

```go
type FormattingOptions struct {
    TabSize                int  // Size of a tab in spaces
    InsertSpaces           bool // Use spaces instead of tabs
    TrimTrailingWhitespace bool // Remove trailing whitespace
    InsertFinalNewline     bool // Ensure file ends with newline
    TrimFinalNewlines      bool // Remove extra final newlines
}
```

### Text Edit

```go
type TextEdit struct {
    Range   Range  // Range to replace (UTF-8 offsets)
    NewText string // Replacement text
}
```

## Document Formatting Provider

### Your First Formatting Provider

Let's create a simple formatter for Go code that fixes indentation:

```go
package mypackage

import (
    "go/format"
    "strings"

    "github.com/SCKelemen/lsp/core"
)

// GoFormattingProvider formats Go source code using gofmt.
type GoFormattingProvider struct{}

func (p *GoFormattingProvider) ProvideFormatting(
    uri, content string,
    options core.FormattingOptions,
) []core.TextEdit {
    // Only handle Go files
    if !strings.HasSuffix(uri, ".go") {
        return nil
    }

    // Format using gofmt
    formatted, err := format.Source([]byte(content))
    if err != nil {
        // If formatting fails, return no edits
        return nil
    }

    // If content is already formatted, return no edits
    if string(formatted) == content {
        return nil
    }

    // Return a single edit that replaces entire document
    lines := strings.Split(content, "\n")
    endLine := len(lines) - 1
    endChar := len(lines[endLine])

    return []core.TextEdit{
        {
            Range: core.Range{
                Start: core.Position{Line: 0, Character: 0},
                End:   core.Position{Line: endLine, Character: endChar},
            },
            NewText: string(formatted),
        },
    }
}
```

### Custom Formatting Provider

Create a custom formatter with specific rules:

```go
type CustomFormattingProvider struct {
    IndentSize int
    UseSpaces  bool
}

func NewCustomFormattingProvider() *CustomFormattingProvider {
    return &CustomFormattingProvider{
        IndentSize: 4,
        UseSpaces:  true,
    }
}

func (p *CustomFormattingProvider) ProvideFormatting(
    uri, content string,
    options core.FormattingOptions,
) []core.TextEdit {
    // Use options if provided
    if options.TabSize > 0 {
        p.IndentSize = options.TabSize
    }
    p.UseSpaces = options.InsertSpaces

    var edits []core.TextEdit

    // Apply formatting rules
    edits = append(edits, p.fixIndentation(content, options)...)
    edits = append(edits, p.fixTrailingWhitespace(content, options)...)
    edits = append(edits, p.fixFinalNewline(content, options)...)

    return edits
}
```

### Fixing Indentation

```go
func (p *CustomFormattingProvider) fixIndentation(
    content string,
    options core.FormattingOptions,
) []core.TextEdit {
    var edits []core.TextEdit
    lines := strings.Split(content, "\n")

    indentChar := "\t"
    if p.UseSpaces {
        indentChar = strings.Repeat(" ", p.IndentSize)
    }

    currentIndent := 0

    for lineNum, line := range lines {
        trimmed := strings.TrimSpace(line)

        // Skip empty lines
        if trimmed == "" {
            continue
        }

        // Calculate expected indentation based on braces
        if strings.HasSuffix(trimmed, "{") {
            // Opening brace - indent next line
            defer func(indent int) { currentIndent = indent + 1 }(currentIndent)
        }
        if strings.HasPrefix(trimmed, "}") {
            // Closing brace - dedent this line
            currentIndent = max(0, currentIndent-1)
        }

        // Calculate expected indentation
        expectedIndent := strings.Repeat(indentChar, currentIndent)
        expectedLine := expectedIndent + trimmed

        // If line doesn't match expected indentation, create edit
        if line != expectedLine {
            edits = append(edits, core.TextEdit{
                Range: core.Range{
                    Start: core.Position{Line: lineNum, Character: 0},
                    End:   core.Position{Line: lineNum, Character: len(line)},
                },
                NewText: expectedLine,
            })
        }
    }

    return edits
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}
```

### Fixing Trailing Whitespace

```go
func (p *CustomFormattingProvider) fixTrailingWhitespace(
    content string,
    options core.FormattingOptions,
) []core.TextEdit {
    if !options.TrimTrailingWhitespace {
        return nil
    }

    var edits []core.TextEdit
    lines := strings.Split(content, "\n")

    for lineNum, line := range lines {
        trimmed := strings.TrimRight(line, " \t")

        if len(trimmed) != len(line) {
            edits = append(edits, core.TextEdit{
                Range: core.Range{
                    Start: core.Position{Line: lineNum, Character: len(trimmed)},
                    End:   core.Position{Line: lineNum, Character: len(line)},
                },
                NewText: "",
            })
        }
    }

    return edits
}
```

### Fixing Final Newline

```go
func (p *CustomFormattingProvider) fixFinalNewline(
    content string,
    options core.FormattingOptions,
) []core.TextEdit {
    var edits []core.TextEdit

    lines := strings.Split(content, "\n")
    lastLine := lines[len(lines)-1]

    if options.InsertFinalNewline && lastLine != "" {
        // Add final newline if missing
        edits = append(edits, core.TextEdit{
            Range: core.Range{
                Start: core.Position{Line: len(lines) - 1, Character: len(lastLine)},
                End:   core.Position{Line: len(lines) - 1, Character: len(lastLine)},
            },
            NewText: "\n",
        })
    }

    if options.TrimFinalNewlines {
        // Remove extra newlines at end
        emptyLines := 0
        for i := len(lines) - 1; i >= 0; i-- {
            if strings.TrimSpace(lines[i]) == "" {
                emptyLines++
            } else {
                break
            }
        }

        if emptyLines > 1 {
            // Keep only one empty line at end
            startLine := len(lines) - emptyLines
            endLine := len(lines) - 2

            edits = append(edits, core.TextEdit{
                Range: core.Range{
                    Start: core.Position{Line: startLine, Character: 0},
                    End:   core.Position{Line: endLine, Character: len(lines[endLine])},
                },
                NewText: "",
            })
        }
    }

    return edits
}
```

## Range Formatting Provider

### Formatting Selected Range

Format only a selected portion of the document:

```go
type GoRangeFormattingProvider struct{}

func (p *GoRangeFormattingProvider) ProvideRangeFormatting(
    uri, content string,
    r core.Range,
    options core.FormattingOptions,
) []core.TextEdit {
    // Only handle Go files
    if !strings.HasSuffix(uri, ".go") {
        return nil
    }

    // Extract the selected range
    lines := strings.Split(content, "\n")

    // Get complete lines covering the range
    startLine := r.Start.Line
    endLine := r.End.Line

    if startLine < 0 || endLine >= len(lines) {
        return nil
    }

    // Extract range content (complete lines)
    var rangeContent strings.Builder
    for i := startLine; i <= endLine; i++ {
        rangeContent.WriteString(lines[i])
        if i < endLine {
            rangeContent.WriteString("\n")
        }
    }

    // Try to make it a valid Go snippet for formatting
    // This is simplified - real implementation would be smarter
    snippet := rangeContent.String()

    // Wrap in a function if needed to make it parseable
    wrappedSnippet := "package main\nfunc _() {\n" + snippet + "\n}"

    formatted, err := format.Source([]byte(wrappedSnippet))
    if err != nil {
        // If wrapping didn't work, try direct formatting
        formatted, err = format.Source([]byte(snippet))
        if err != nil {
            return nil
        }
    } else {
        // Extract the formatted snippet from the wrapper
        formattedStr := string(formatted)

        // Remove package and function wrapper
        start := strings.Index(formattedStr, "func _() {\n") + len("func _() {\n")
        end := strings.LastIndex(formattedStr, "\n}")

        if start > 0 && end > start {
            formatted = []byte(formattedStr[start:end])
        }
    }

    // If content unchanged, return no edits
    if string(formatted) == snippet {
        return nil
    }

    // Create edit for the range
    return []core.TextEdit{
        {
            Range: core.Range{
                Start: core.Position{Line: startLine, Character: 0},
                End:   core.Position{Line: endLine, Character: len(lines[endLine])},
            },
            NewText: string(formatted),
        },
    }
}
```

### Custom Range Formatting

```go
type CustomRangeFormattingProvider struct {
    IndentSize int
    UseSpaces  bool
}

func (p *CustomRangeFormattingProvider) ProvideRangeFormatting(
    uri, content string,
    r core.Range,
    options core.FormattingOptions,
) []core.TextEdit {
    if options.TabSize > 0 {
        p.IndentSize = options.TabSize
    }
    p.UseSpaces = options.InsertSpaces

    lines := strings.Split(content, "\n")

    // Validate range
    if r.Start.Line < 0 || r.End.Line >= len(lines) {
        return nil
    }

    var edits []core.TextEdit

    // Format each line in range
    for lineNum := r.Start.Line; lineNum <= r.End.Line; lineNum++ {
        line := lines[lineNum]

        // Skip empty lines
        if strings.TrimSpace(line) == "" {
            continue
        }

        // Fix indentation for this line
        trimmed := strings.TrimLeft(line, " \t")
        currentIndent := len(line) - len(trimmed)

        // Calculate how many indent levels based on current indentation
        indentLevels := currentIndent / p.IndentSize

        // Reconstruct with correct indent
        indentChar := "\t"
        if p.UseSpaces {
            indentChar = strings.Repeat(" ", p.IndentSize)
        }

        newIndent := strings.Repeat(indentChar, indentLevels)
        newLine := newIndent + trimmed

        // Trim trailing whitespace if requested
        if options.TrimTrailingWhitespace {
            newLine = strings.TrimRight(newLine, " \t")
        }

        if newLine != line {
            edits = append(edits, core.TextEdit{
                Range: core.Range{
                    Start: core.Position{Line: lineNum, Character: 0},
                    End:   core.Position{Line: lineNum, Character: len(line)},
                },
                NewText: newLine,
            })
        }
    }

    return edits
}
```

## Testing Formatting Providers

### Testing Document Formatting

```go
func TestGoFormattingProvider(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name: "already formatted",
            input: `package main

func main() {
	println("hello")
}
`,
            want: `package main

func main() {
	println("hello")
}
`,
        },
        {
            name: "needs formatting",
            input: `package main
func main(){
println("hello")
}`,
            want: `package main

func main() {
	println("hello")
}
`,
        },
        {
            name: "fix indentation",
            input: `package main

func main() {
    println("hello")
}`,
            want: `package main

func main() {
	println("hello")
}
`,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            provider := &GoFormattingProvider{}
            options := core.FormattingOptions{
                TabSize:      4,
                InsertSpaces: false,
            }

            edits := provider.ProvideFormatting("file:///test.go", tt.input, options)

            // Apply edits
            result := applyEdits(tt.input, edits)

            if result != tt.want {
                t.Errorf("formatting result:\n%s\n\nwant:\n%s", result, tt.want)
            }
        })
    }
}

func applyEdits(content string, edits []core.TextEdit) string {
    if len(edits) == 0 {
        return content
    }

    // For simplicity, if there's only one edit replacing entire document
    if len(edits) == 1 {
        edit := edits[0]
        if edit.Range.Start.Line == 0 && edit.Range.Start.Character == 0 {
            return edit.NewText
        }
    }

    // Apply each edit
    lines := strings.Split(content, "\n")

    for _, edit := range edits {
        if edit.Range.Start.Line == edit.Range.End.Line {
            // Single line edit
            line := lines[edit.Range.Start.Line]
            newLine := line[:edit.Range.Start.Character] +
                edit.NewText +
                line[edit.Range.End.Character:]
            lines[edit.Range.Start.Line] = newLine
        } else {
            // Multi-line edit (simplified)
            before := lines[edit.Range.Start.Line][:edit.Range.Start.Character]
            after := lines[edit.Range.End.Line][edit.Range.End.Character:]
            lines[edit.Range.Start.Line] = before + edit.NewText + after

            // Remove lines in between
            if edit.Range.End.Line > edit.Range.Start.Line {
                lines = append(
                    lines[:edit.Range.Start.Line+1],
                    lines[edit.Range.End.Line+1:]...,
                )
            }
        }
    }

    return strings.Join(lines, "\n")
}
```

### Testing Range Formatting

```go
func TestGoRangeFormattingProvider(t *testing.T) {
    content := `package main

func main() {
x:=1
y:=2
    println(x+y)
}

func other() {
    println("unchanged")
}
`

    provider := &GoRangeFormattingProvider{}
    options := core.FormattingOptions{
        TabSize:      4,
        InsertSpaces: false,
    }

    // Format only lines 3-5 (the function body)
    r := core.Range{
        Start: core.Position{Line: 3, Character: 0},
        End:   core.Position{Line: 5, Character: 20},
    }

    edits := provider.ProvideRangeFormatting("file:///test.go", content, r, options)

    if len(edits) == 0 {
        t.Fatal("expected formatting edits")
    }

    result := applyEdits(content, edits)

    // Check that the formatted lines have correct indentation
    lines := strings.Split(result, "\n")

    // Line with x:=1 should be formatted with proper indentation
    if !strings.Contains(lines[3], "\tx := 1") {
        t.Errorf("line 3 not formatted correctly: %q", lines[3])
    }

    // Line with "other" function should be unchanged
    if !strings.Contains(result, `func other() {`) {
        t.Error("other function was modified")
    }
}
```

### Testing Formatting Options

```go
func TestFormattingOptions(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        options core.FormattingOptions
        want    string
    }{
        {
            name:  "trim trailing whitespace",
            input: "hello world   \ngoodbye  \n",
            options: core.FormattingOptions{
                TrimTrailingWhitespace: true,
            },
            want: "hello world\ngoodbye\n",
        },
        {
            name:  "insert final newline",
            input: "hello world",
            options: core.FormattingOptions{
                InsertFinalNewline: true,
            },
            want: "hello world\n",
        },
        {
            name:  "trim final newlines",
            input: "hello world\n\n\n",
            options: core.FormattingOptions{
                TrimFinalNewlines: true,
            },
            want: "hello world\n",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            provider := NewCustomFormattingProvider()
            edits := provider.ProvideFormatting("file:///test.txt", tt.input, tt.options)

            result := applyEdits(tt.input, edits)

            if result != tt.want {
                t.Errorf("got %q, want %q", result, tt.want)
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
    documents       *core.DocumentManager
    formatting      core.FormattingProvider
    rangeFormatting core.RangeFormattingProvider
}

func NewMyServer() *MyServer {
    return &MyServer{
        documents:       core.NewDocumentManager(),
        formatting:      &GoFormattingProvider{},
        rangeFormatting: &GoRangeFormattingProvider{},
    }
}

func (s *MyServer) TextDocumentFormatting(
    context *lsp.Context,
    params *protocol.DocumentFormattingParams,
) ([]protocol.TextEdit, error) {
    uri := string(params.TextDocument.URI)
    content := s.documents.GetContent(uri)

    if content == "" {
        return nil, nil
    }

    // Convert formatting options
    coreOptions := core.FormattingOptions{
        TabSize:                int(params.Options.TabSize),
        InsertSpaces:           params.Options.InsertSpaces,
        TrimTrailingWhitespace: params.Options.TrimTrailingWhitespace,
        InsertFinalNewline:     params.Options.InsertFinalNewline,
        TrimFinalNewlines:      params.Options.TrimFinalNewlines,
    }

    // Get formatting edits
    coreEdits := s.formatting.ProvideFormatting(uri, content, coreOptions)

    // Convert to protocol edits
    var protocolEdits []protocol.TextEdit
    for _, edit := range coreEdits {
        protocolEdits = append(protocolEdits, protocol.TextEdit{
            Range:   adapter.CoreToProtocolRange(edit.Range, content),
            NewText: edit.NewText,
        })
    }

    return protocolEdits, nil
}

func (s *MyServer) TextDocumentRangeFormatting(
    context *lsp.Context,
    params *protocol.DocumentRangeFormattingParams,
) ([]protocol.TextEdit, error) {
    uri := string(params.TextDocument.URI)
    content := s.documents.GetContent(uri)

    if content == "" {
        return nil, nil
    }

    // Convert range and options
    coreRange := adapter.ProtocolToCoreRange(params.Range, content)
    coreOptions := core.FormattingOptions{
        TabSize:                int(params.Options.TabSize),
        InsertSpaces:           params.Options.InsertSpaces,
        TrimTrailingWhitespace: params.Options.TrimTrailingWhitespace,
        InsertFinalNewline:     params.Options.InsertFinalNewline,
        TrimFinalNewlines:      params.Options.TrimFinalNewlines,
    }

    // Get range formatting edits
    coreEdits := s.rangeFormatting.ProvideRangeFormatting(uri, content, coreRange, coreOptions)

    // Convert to protocol edits
    var protocolEdits []protocol.TextEdit
    for _, edit := range coreEdits {
        protocolEdits = append(protocolEdits, protocol.TextEdit{
            Range:   adapter.CoreToProtocolRange(edit.Range, content),
            NewText: edit.NewText,
        })
    }

    return protocolEdits, nil
}
```

### Server Capabilities

```go
func (s *MyServer) Initialize(
    context *lsp.Context,
    params *protocol.InitializeParams,
) (interface{}, error) {
    capabilities := protocol.ServerCapabilities{
        DocumentFormattingProvider:      true,
        DocumentRangeFormattingProvider: true,
    }

    return protocol.InitializeResult{
        Capabilities: capabilities,
    }, nil
}
```

## Summary

You now know how to:

1. ✅ Implement formatting providers with `FormattingProvider` and `RangeFormattingProvider` interfaces
2. ✅ Use formatting options (tab size, spaces, trailing whitespace, final newlines)
3. ✅ Create text edits to transform document content
4. ✅ Format entire documents or selected ranges
5. ✅ Test formatting providers thoroughly
6. ✅ Integrate formatting into an LSP server

**Key Points:**
- Return minimal edits when possible (don't replace entire document if not needed)
- Respect formatting options provided by the client
- Handle edge cases (empty files, invalid ranges, parse errors)
- Use language-specific formatters when available (gofmt, prettier, etc.)
- Test with various formatting options and edge cases

**Next Steps:**
- See [CODE_ACTIONS.md](CODE_ACTIONS.md) for code action providers
- See [VALIDATORS.md](VALIDATORS.md) for diagnostic providers
- Check `examples/` for complete working examples
- Read [CORE_TYPES.md](CORE_TYPES.md) for UTF-8/UTF-16 details
