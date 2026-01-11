# LSP - Language Server Protocol Library for Go

A Go library for implementing Language Server Protocol (LSP) servers with protocol-agnostic core types optimized for CLI tools and server implementations.

## Key Features

- **Protocol-Agnostic Core Types**: Work with UTF-8 byte offsets naturally in Go
- **CLI-First Design**: Use LSP data structures in CLI tools without protocol overhead
- **Reusable Providers**: Write business logic once, use in both CLI and LSP server
- **Automatic UTF-16 Conversion**: Adapters handle protocol conversion at boundaries per LSP spec requirements
- **Full LSP Support**: Implements LSP 3.16, 3.17, and 3.18

## LSP Primer: Editor Features → LSP Capabilities

If you're familiar with VS Code but new to LSP servers, here's how editor features map to LSP capabilities:

### Code Intelligence

| You Know This As... | LSP Feature | What It Does |
|---------------------|-------------|--------------|
| Red/yellow squiggles under code | **Diagnostics** (`textDocument/publishDiagnostics`) | Show errors, warnings, and hints |
| Lightbulb with quick fixes | **Code Actions** (`textDocument/codeAction`) | Provide quick fixes and refactorings |
| Auto-complete dropdown | **Completion** (`textDocument/completion`) | Suggest code completions |
| Gray ghost text suggestions | **Inline Completions** (`textDocument/inlineCompletion`) | AI-powered inline code suggestions |
| Parameter hints `(paramName: ...)` | **Signature Help** (`textDocument/signatureHelp`) | Show function parameters while typing |
| Type hints `x: number` | **Inlay Hints** (`textDocument/inlayHint`) | Show inferred types and parameter names |

### Navigation

| You Know This As... | LSP Feature | What It Does |
|---------------------|-------------|--------------|
| Ctrl+Click / F12 to jump to definition | **Go to Definition** (`textDocument/definition`) | Jump to where something is defined |
| Find all references | **References** (`textDocument/references`) | Find all uses of a symbol |
| Hover tooltip with docs | **Hover** (`textDocument/hover`) | Show documentation on hover |
| Breadcrumbs / outline view | **Document Symbols** (`textDocument/documentSymbol`) | Show file structure |
| File symbol search (Ctrl+Shift+O) | **Document Symbols** | Quick navigation within file |
| Workspace symbol search (Ctrl+T) | **Workspace Symbols** (`workspace/symbol`) | Search symbols across workspace |

### Editing

| You Know This As... | LSP Feature | What It Does |
|---------------------|-------------|--------------|
| Format Document | **Formatting** (`textDocument/formatting`) | Auto-format code |
| Format Selection | **Range Formatting** (`textDocument/rangeFormatting`) | Format selected code |
| Rename Symbol (F2) | **Rename** (`textDocument/rename`) | Rename across all files |
| Fold/unfold code regions | **Folding Range** (`textDocument/foldingRange`) | Define collapsible regions |
| Color picker on `#FF0000` | **Document Color** (`textDocument/documentColor`) | Show color decorators |

### Advanced Features

| You Know This As... | LSP Feature | What It Does |
|---------------------|-------------|--------------|
| Syntax highlighting (semantic) | **Semantic Tokens** (`textDocument/semanticTokens`) | Enhanced syntax coloring |
| Code lenses (clickable hints above code) | **Code Lens** (`textDocument/codeLens`) | Inline actionable commands |
| Smart selection expansion | **Selection Range** (`textDocument/selectionRange`) | Expand/shrink selection intelligently |

**Not sure what to implement?** Start with:
1. **Diagnostics** - Show errors and warnings
2. **Completion** - Basic auto-complete
3. **Hover** - Show documentation
4. **Go to Definition** - Jump to definitions

These four features provide 80% of the value with 20% of the effort.

## Quick Start

### 1. CLI Tool (No LSP Server)

```go
import "github.com/SCKelemen/lsp/core"

func lintFile(content string) []core.Diagnostic {
    var diagnostics []core.Diagnostic
    severity := core.SeverityWarning

    // Work naturally with UTF-8 byte offsets
    for i := 0; i < len(content); i++ {
        if content[i] == '\t' {
            pos := core.ByteOffsetToPosition(content, i)
            diagnostics = append(diagnostics, core.Diagnostic{
                Range: core.Range{
                    Start: pos,
                    End:   core.Position{Line: pos.Line, Character: pos.Character + 1},
                },
                Severity: &severity,
                Message:  "Use spaces instead of tabs",
            })
        }
    }
    return diagnostics
}
```

### 2. Reusable Provider

```go
// Write once, use in CLI and LSP server
type TabFixProvider struct{}

func (p *TabFixProvider) ProvideCodeFixes(ctx core.CodeFixContext) []core.CodeAction {
    // Your fix logic using UTF-8 offsets
    return actions
}

// Use in CLI tool
registry := core.NewCodeFixRegistry()
registry.Register(&TabFixProvider{})
fixes := registry.ProvideCodeFixes(ctx)

// Same provider works in LSP server - just convert at boundaries
```

### 3. LSP Server Handler

```go
import (
    "github.com/SCKelemen/lsp"
    "github.com/SCKelemen/lsp/adapter"
    "github.com/SCKelemen/lsp/core"
    protocol "github.com/SCKelemen/lsp/protocol"
)

func (s *Server) TextDocumentDidOpen(
    context *glsp.Context,
    params *protocol.DidOpenTextDocumentParams,
) error {
    uri := string(params.TextDocument.URI)
    content := params.TextDocument.Text

    // Business logic with core types (UTF-8)
    diagnostics := s.validate(uri, content)

    // Convert to protocol at boundary
    protocolDiags := adapter.CoreToProtocolDiagnostics(diagnostics, content)

    // Send to client
    context.Notify(...)
    return nil
}
```

## 📚 Comprehensive Feature Guides

Complete guides for implementing LSP features from scratch to production:

### Core Features

➡️ **[VALIDATORS.md](VALIDATORS.md)** - Diagnostic Providers
- Writing validators that detect code issues
- Testing with Unicode content
- Composing multiple validators
- Integration with LSP servers

➡️ **[CODE_ACTIONS.md](CODE_ACTIONS.md)** - Code Action Providers
- Quick fixes for diagnostics
- Refactoring actions (extract, inline, rewrite)
- Source actions (organize imports, fix all)
- Workspace edits and multiple file changes

### Navigation & Information

➡️ **[NAVIGATION.md](NAVIGATION.md)** - Definition & Hover Providers
- Go-to-definition across files
- Hover information with markdown
- Finding references and implementations
- Type definitions and declarations

➡️ **[SYMBOLS.md](SYMBOLS.md)** - Document Symbol Providers
- Hierarchical document structure
- Outline and breadcrumb navigation
- Symbol kinds (functions, classes, variables)
- Tree-based symbol representation

### Editor Features

➡️ **[FOLDING.md](FOLDING.md)** - Folding Range Providers
- Collapsible code regions
- Multiple folding strategies (braces, indentation, regions)
- Comments, imports, and function folding
- Language-agnostic folding patterns

➡️ **[FORMATTING.md](FORMATTING.md)** - Formatting Providers
- Document-wide formatting
- Range formatting for selections
- Formatting options (tabs, spaces, whitespace)
- Integration with language formatters

**Each guide includes:**
- ✅ Step-by-step implementation
- ✅ Complete working examples
- ✅ Comprehensive test patterns
- ✅ LSP server integration
- ✅ UTF-8/UTF-16 conversion details

## Architecture

```
┌─────────────────────────────────────┐
│      Your Business Logic            │
│   (core types, UTF-8 offsets)       │
└────────────┬────────────────────────┘
             │
             ├─ CLI Tools: Direct use
             │
             └─ LSP Server: Convert at boundaries
                      ↓
          ┌───────────────────────┐
          │   Adapter Packages    │
          │   (UTF-8 ↔ UTF-16)   │
          └───────────────────────┘
                      ↓
          ┌───────────────────────┐
          │   Protocol Types      │
          │   (JSON-RPC/LSP)      │
          └───────────────────────┘
```

## Why Core Types?

**LSP uses UTF-16** (per specification), but **Go strings are UTF-8**. Core types use UTF-8 for natural Go string handling:

```go
content := "hello 世界"  // 世界 is 3 bytes in UTF-8, 1 code unit in UTF-16

// Core types: direct indexing
pos := core.ByteOffsetToPosition(content, 8)  // Natural UTF-8 offset

// Protocol types: need conversion
protocolPos := adapter.CoreToProtocolPosition(pos, content)
```

## Core Packages

### `core/`
Protocol-agnostic types using UTF-8:
- **types.go**: Position, Range, Location, Diagnostic
- **language_features.go**: FoldingRange, TextEdit, DocumentSymbol, CodeAction, WorkspaceEdit
- **codefix.go**: Provider interfaces (CodeFixProvider, DiagnosticProvider, etc.)
- **document.go**: DocumentManager for managing documents in memory
- **encoding.go**: UTF-8 ↔ UTF-16 conversion utilities

### `adapter/`
Convert between core (UTF-8) and protocol (UTF-16) types, supporting LSP 3.16, 3.17, and 3.18 features

### `examples/`
Complete working examples for CLI tools and LSP servers

## Provider Interfaces

Write providers that work with core types:

```go
type DiagnosticProvider interface {
    ProvideDiagnostics(uri, content string) []Diagnostic
}

type CodeFixProvider interface {
    ProvideCodeFixes(ctx CodeFixContext) []CodeAction
}

type FoldingRangeProvider interface {
    ProvideFoldingRanges(uri, content string) []FoldingRange
}

// ... and more (hover, definition, formatting, etc.)
```

**→ See comprehensive feature guides above for complete implementation details**

## Traditional LSP Server (Original Functionality)

This library still supports traditional LSP server usage:

```go
package main

import (
	"github.com/SCKelemen/lsp"
	protocol "github.com/SCKelemen/lsp/protocol"
	"github.com/SCKelemen/lsp/server"
	"github.com/tliron/commonlog"
	_ "github.com/tliron/commonlog/simple"
)

const lsName = "my language"

var (
	version string = "0.0.1"
	handler protocol.Handler
)

func main() {
	commonlog.Configure(1, nil)

	handler = protocol.Handler{
		Initialize:  initialize,
		Initialized: initialized,
		Shutdown:    shutdown,
		SetTrace:    setTrace,
	}

	server := server.NewServer(&handler, lsName, false)
	server.RunStdio()
}

func initialize(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
	capabilities := handler.CreateServerCapabilities()
	return protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    lsName,
			Version: &version,
		},
	}, nil
}

func initialized(context *glsp.Context, params *protocol.InitializedParams) error {
	return nil
}

func shutdown(context *glsp.Context) error {
	protocol.SetTraceValue(protocol.TraceValueOff)
	return nil
}

func setTrace(context *glsp.Context, params *protocol.SetTraceParams) error {
	protocol.SetTraceValue(params.Value)
	return nil
}
```

## Documentation

### Feature Implementation Guides
- **[VALIDATORS.md](VALIDATORS.md)** - Diagnostic providers and validators
- **[CODE_ACTIONS.md](CODE_ACTIONS.md)** - Quick fixes, refactorings, and source actions
- **[NAVIGATION.md](NAVIGATION.md)** - Go-to-definition and hover information
- **[SYMBOLS.md](SYMBOLS.md)** - Document symbols and outline
- **[FOLDING.md](FOLDING.md)** - Folding ranges for code regions
- **[FORMATTING.md](FORMATTING.md)** - Document and range formatting

### Architecture & Reference
- **[CORE_TYPES.md](CORE_TYPES.md)** - Core types architecture and UTF-8/UTF-16 conversion
- **[LSP_CAPABILITIES.md](LSP_CAPABILITIES.md)** - All LSP capabilities with support status
- **[examples/](examples/)** - Complete working examples with tests
- **[core/](core/)** - Core type definitions and utilities

## Testing

```bash
# Run all tests
go test ./core/... ./adapter/... ./examples/...

# Test specific package
go test ./core/... -v

# Run adapter tests
go test ./adapter/... -v

# Build examples
go build ./examples/...
```

## License

BearWare 1.0

Copyright (c) 2025 Samuel Kelemen

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

## Attribution

**Forked from [tliron/glsp](https://github.com/tliron/glsp)** by Tal Liron.

Major architectural improvements in this fork:
- Added core types with UTF-8 byte offsets
- Created adapter packages for protocol conversion
- Added provider interfaces for reusable business logic
- Comprehensive documentation and examples for CLI and server usage
- Document manager for stateful document tracking
- Full support for LSP 3.16, 3.17, and 3.18

## References

- [LSP 3.16 Specification](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/)
- [LSP 3.17 Specification](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/)
- [LSP 3.18 Specification](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.18/specification/)
