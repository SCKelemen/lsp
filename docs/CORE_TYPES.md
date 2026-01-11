# Core Types Architecture

This LSP library provides protocol-agnostic core types that use UTF-8 byte offsets for natural Go string handling. This makes it easy to use LSP data structures in CLI tools and other contexts without the overhead of UTF-16 conversion.

**LSP Specification References:**
- LSP 3.16: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/
- LSP 3.17: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/
- Position Encoding (UTF-16): https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#textDocuments

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     Your Business Logic                      │
│              (Uses core types with UTF-8 offsets)            │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ├─ CLI Tools: Use core types directly
                       │
                       └─ LSP Server: Convert at boundaries
                                ↓
                  ┌─────────────────────────┐
                  │   Adapter Packages      │
                  │  (UTF-8 ↔ UTF-16)      │
                  └─────────────────────────┘
                                ↓
                  ┌─────────────────────────┐
                  │   Protocol Types        │
                  │   (JSON-RPC/LSP)        │
                  └─────────────────────────┘
```

## Key Packages

### `core/`
Protocol-agnostic types using UTF-8 byte offsets:
- **Position**: Line and UTF-8 byte offset ([LSP Spec](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#position))
- **Range**: Start and end positions ([LSP Spec](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#range))
- **Location**: URI and range ([LSP Spec](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#location))
- **Diagnostic**: Error/warning with range, severity, code, etc. ([LSP Spec](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#diagnostic))
- **DocumentManager**: Utility for managing documents in memory

### `adapter/`
Convert between core types (UTF-8) and protocol types (UTF-16):
- Position conversions (UTF-8 ↔ UTF-16 code units)
- Range conversions
- Diagnostic conversions
- Batch conversion helpers
- Full support for LSP 3.16, 3.17, and 3.18 features

**Important:** The LSP specification requires UTF-16 code units for character offsets, not UTF-8 bytes.
Our adapters handle this conversion automatically at API boundaries.

## Usage Patterns

### Pattern 1: CLI Tools (No LSP Server)

CLI tools can use core types directly without any protocol conversion:

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
                Source:   "style-checker",
            })
        }
    }

    return diagnostics
}

func main() {
    content := readFile("example.txt")
    diags := lintFile(content)

    // Print diagnostics
    for _, diag := range diags {
        fmt.Printf("%s:%s: %s\n",
            "example.txt",
            diag.Range.Start,
            diag.Message)
    }
}
```

### Pattern 2: LSP Server Handlers

LSP server handlers convert at the boundaries:

```go
import (
    "github.com/SCKelemen/lsp"
    "github.com/SCKelemen/lsp/adapter"
    "github.com/SCKelemen/lsp/core"
    protocol "github.com/SCKelemen/lsp/protocol"
)

type MyServer struct {
    documents *core.DocumentManager
}

func (s *MyServer) TextDocumentDidOpen(
    context *glsp.Context,
    params *protocol.DidOpenTextDocumentParams,
) error {
    // Extract from protocol types
    uri := string(params.TextDocument.URI)
    content := params.TextDocument.Text
    version := int(params.TextDocument.Version)

    // Store using core types
    s.documents.Open(uri, content, version)

    // Business logic with core types
    coreDiagnostics := s.validateDocument(uri, content)

    // Convert back to protocol types
    protocolDiags := adapter.CoreToProtocolDiagnostics(
        coreDiagnostics,
        content,
    )

    // Send to client
    context.Notify(
        protocol.MethodTextDocumentPublishDiagnostics,
        protocol.PublishDiagnosticsParams{
            URI:         params.TextDocument.URI,
            Diagnostics: protocolDiags,
        },
    )

    return nil
}

// Business logic uses core types with UTF-8 offsets
func (s *MyServer) validateDocument(uri, content string) []core.Diagnostic {
    // Your validation logic here
    // Work naturally with UTF-8 byte offsets
    var diagnostics []core.Diagnostic
    // ... validation logic ...
    return diagnostics
}
```

### Pattern 3: Position-Based Operations

When working with positions from the client:

```go
func (s *MyServer) TextDocumentHover(
    context *glsp.Context,
    params *protocol.HoverParams,
) (*protocol.Hover, error) {
    uri := string(params.TextDocument.URI)
    content := s.documents.GetContent(uri)

    // Convert protocol position (UTF-16) to core position (UTF-8)
    corePos := adapter.ProtocolToCorePosition(params.Position, content)

    // Business logic with UTF-8 offsets
    hoverInfo := s.getHoverInfo(uri, content, corePos)
    if hoverInfo == nil {
        return nil, nil
    }

    // Convert range back to protocol
    protocolRange := adapter.CoreToProtocolRange(hoverInfo.Range, content)

    return &protocol.Hover{
        Contents: hoverInfo.Contents,
        Range:    &protocolRange,
    }, nil
}
```

## UTF-8 vs UTF-16

### Why Core Types Use UTF-8

- **Natural Go strings**: Go strings are UTF-8 encoded
- **Direct indexing**: `content[i]` works without conversion
- **Standard library**: All Go string functions work naturally
- **CLI tools**: No conversion overhead for non-LSP usage

### When Conversion Happens

- **LSP protocol**: Uses UTF-16 code units per specification
- **Conversion points**: Only at handler boundaries (protocol → core → protocol)
- **Multi-byte characters**: Automatically handled by adapters
  - Emoji (😀): 4 UTF-8 bytes → 2 UTF-16 code units
  - Chinese (你): 3 UTF-8 bytes → 1 UTF-16 code unit

### Character Boundary Handling

When a position lands in the middle of a multi-byte character, conversions round down to the character start:

```go
content := "hello 你 world"  // 你 is bytes 6-8

// If UTF-16 offset 7 (middle of 你) is converted to UTF-8:
pos := adapter.ProtocolToCorePosition(
    protocol.Position{Line: 0, Character: 7},
    content,
)
// Result: Position{Line: 0, Character: 6} (start of 你)
```

## Document Manager

The `core.DocumentManager` helps manage documents in memory:

```go
docs := core.NewDocumentManager()

// Open documents
docs.Open("file:///example.txt", "content", 1)

// Get content
content := docs.GetContent("file:///example.txt")

// Apply edits
docs.ApplyEdit(
    "file:///example.txt",
    core.Range{
        Start: core.Position{Line: 0, Character: 0},
        End:   core.Position{Line: 0, Character: 5},
    },
    "replacement",
)

// Close documents
docs.Close("file:///example.txt")
```

## Best Practices

1. **Use core types in business logic**: Keep UTF-8 throughout your code
2. **Convert at boundaries**: Only convert to/from protocol types at handler boundaries
3. **Store document content**: Use `DocumentManager` or store content to enable conversions
4. **Validate ranges**: Use `Range.IsValid()` to check range validity
5. **Thread safety**: `DocumentManager` is thread-safe for concurrent access

## Migration from Protocol Types

If you have existing code using protocol types:

1. Identify business logic (keep separate from handlers)
2. Change business logic to use core types
3. Add adapter conversions at handler entry/exit
4. Store document content for position conversions

Example:

```go
// Before: Business logic mixed with protocol types
func (s *Server) validate(params *protocol.TextDocumentParams) {
    // Logic using protocol.Position, protocol.Range, etc.
}

// After: Business logic uses core types
func (s *Server) validateCore(uri string, content string, pos core.Position) []core.Diagnostic {
    // Logic using core.Position, core.Range, etc.
}

// Handler converts at boundary
func (s *Server) TextDocumentHandler(ctx *glsp.Context, params *protocol.TextDocumentParams) error {
    uri := string(params.TextDocument.URI)
    content := s.documents.GetContent(uri)
    corePos := adapter.ProtocolToCorePosition(params.Position, content)

    coreDiags := s.validateCore(uri, content, corePos)
    protocolDiags := adapter.CoreToProtocolDiagnostics(coreDiags, content)

    // Send to client...
}
```

## Examples

See `examples/core_handler_example.go` for complete working examples.

## Testing

All core types and adapters include comprehensive tests:
- `core/encoding_test.go`: UTF-8/UTF-16 conversion tests
- `adapter/position_test.go`: Position and range conversion tests
- Round-trip tests ensure conversion accuracy
