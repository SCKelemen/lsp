# Writing Validators Guide

This guide shows you how to write validators (diagnostic providers) using this LSP library, from simple standalone validators to full LSP server integration.

**Table of Contents:**
1. [Core Concepts](#core-concepts)
2. [Your First Validator](#your-first-validator)
3. [Testing Validators](#testing-validators)
4. [Registry and Composition](#registry-and-composition)
5. [Code Fixes](#code-fixes)
6. [LSP Server Integration](#lsp-server-integration)

---

## Core Concepts

### What is a Validator?

A validator implements the `DiagnosticProvider` interface to find problems in code:

```go
type DiagnosticProvider interface {
    ProvideDiagnostics(uri, content string) []Diagnostic
}
```

**Key properties:**
- ✅ **Protocol-agnostic**: Uses UTF-8 byte offsets (natural Go string handling)
- ✅ **Reusable**: Same code works in CLI tools and LSP servers
- ✅ **Composable**: Multiple validators can run on the same file
- ✅ **Testable**: Pure functions, easy to unit test

### Core Types

All validators work with these core types (UTF-8 offsets):

```go
// Position in a document (UTF-8 byte offset, zero-based line)
type Position struct {
    Line      int  // Zero-based line number
    Character int  // UTF-8 byte offset within the line
}

// Range in a document
type Range struct {
    Start Position
    End   Position
}

// Diagnostic represents an error, warning, or hint
type Diagnostic struct {
    Range    Range
    Severity *DiagnosticSeverity  // Error, Warning, Info, Hint
    Code     *DiagnosticCode      // Optional diagnostic code
    Source   string               // Name of your validator
    Message  string               // Human-readable message
    Tags     []DiagnosticTag      // Unnecessary, Deprecated
    // ... additional fields
}
```

**Important:** These use UTF-8 byte offsets, not UTF-16 code units. Adapters handle conversion when sending to LSP clients.

---

## Your First Validator

Let's write a simple validator that finds lines that are too long.

### Step 1: Create the Validator

```go
package validators

import "github.com/SCKelemen/lsp/core"

// LineLengthValidator checks for lines that exceed a maximum length.
type LineLengthValidator struct {
    MaxLength int
}

func NewLineLengthValidator(maxLength int) *LineLengthValidator {
    return &LineLengthValidator{MaxLength: maxLength}
}

func (v *LineLengthValidator) ProvideDiagnostics(uri, content string) []core.Diagnostic {
    var diagnostics []core.Diagnostic

    // Split content into lines
    lines := strings.Split(content, "\n")

    for lineNum, line := range lines {
        // Check if line exceeds max length
        if len(line) > v.MaxLength {
            severity := core.SeverityWarning
            code := core.NewStringCode("line-too-long")

            // Create diagnostic at the end of the line
            diagnostics = append(diagnostics, core.Diagnostic{
                Range: core.Range{
                    Start: core.Position{Line: lineNum, Character: v.MaxLength},
                    End:   core.Position{Line: lineNum, Character: len(line)},
                },
                Severity: &severity,
                Code:     &code,
                Source:   "line-length",
                Message:  fmt.Sprintf("Line exceeds %d characters", v.MaxLength),
            })
        }
    }

    return diagnostics
}
```

### Step 2: Use It Directly (No LSP Server)

```go
package main

import (
    "fmt"
    "github.com/SCKelemen/lsp/core"
)

func main() {
    content := `package main

func main() {
    // This is a very long line that exceeds our maximum line length limit and should be flagged
    fmt.Println("Hello, world!")
}
`

    validator := NewLineLengthValidator(80)
    diagnostics := validator.ProvideDiagnostics("file:///main.go", content)

    for _, diag := range diagnostics {
        fmt.Printf("%s:%d:%d: %s: %s\n",
            "main.go",
            diag.Range.Start.Line+1,  // Convert to 1-based for display
            diag.Range.Start.Character+1,
            diag.Severity,
            diag.Message,
        )
    }
}
```

Output:
```
main.go:4:81: warning: Line exceeds 80 characters
```

---

## Testing Validators

Good validators are well-tested. Here's how to write tests:

### Basic Test Pattern

```go
package validators

import (
    "testing"
    "github.com/SCKelemen/lsp/core"
)

func TestLineLengthValidator(t *testing.T) {
    tests := []struct {
        name          string
        content       string
        maxLength     int
        wantCount     int  // Expected number of diagnostics
        wantLine      int  // Expected line number of first diagnostic
        wantMessage   string
    }{
        {
            name:      "no issues",
            content:   "package main\n\nfunc main() {\n}\n",
            maxLength: 80,
            wantCount: 0,
        },
        {
            name:        "one long line",
            content:     "package main\n\n// This is a very long comment that definitely exceeds our maximum character limit\nfunc main() {\n}\n",
            maxLength:   80,
            wantCount:   1,
            wantLine:    2,
            wantMessage: "Line exceeds 80 characters",
        },
        {
            name:      "multiple long lines",
            content:   "// Line 1 is very long and exceeds the limit set by our configuration\n// Line 2 is also very long and exceeds the limit set by our configuration\nfunc main() {\n}\n",
            maxLength: 60,
            wantCount: 2,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            validator := NewLineLengthValidator(tt.maxLength)
            diagnostics := validator.ProvideDiagnostics("file:///test.go", tt.content)

            if len(diagnostics) != tt.wantCount {
                t.Errorf("got %d diagnostics, want %d", len(diagnostics), tt.wantCount)
            }

            if tt.wantCount > 0 {
                first := diagnostics[0]

                if first.Range.Start.Line != tt.wantLine {
                    t.Errorf("first diagnostic at line %d, want line %d",
                        first.Range.Start.Line, tt.wantLine)
                }

                if tt.wantMessage != "" && first.Message != tt.wantMessage {
                    t.Errorf("got message %q, want %q", first.Message, tt.wantMessage)
                }
            }
        })
    }
}
```

### Testing with UTF-8 Multi-byte Characters

```go
func TestLineLengthValidator_Unicode(t *testing.T) {
    content := "// 你好世界 - This line has Chinese characters"

    validator := NewLineLengthValidator(30)
    diagnostics := validator.ProvideDiagnostics("file:///test.go", content)

    if len(diagnostics) != 1 {
        t.Fatalf("expected 1 diagnostic, got %d", len(diagnostics))
    }

    diag := diagnostics[0]

    // Verify the position is at a valid UTF-8 boundary
    if diag.Range.Start.Character < 0 || diag.Range.Start.Character > len(content) {
        t.Errorf("invalid character position: %d", diag.Range.Start.Character)
    }
}
```

---

## Registry and Composition

Run multiple validators on the same file using a registry:

### Creating a Registry

```go
package main

import (
    "github.com/SCKelemen/lsp/core"
)

func main() {
    // Create registry
    registry := core.NewDiagnosticRegistry()

    // Register multiple validators
    registry.Register(NewLineLengthValidator(80))
    registry.Register(NewTabValidator())
    registry.Register(NewTrailingWhitespaceValidator())

    // Run all validators
    content := readFile("main.go")
    diagnostics := registry.ProvideDiagnostics("file:///main.go", content)

    // Process results
    for _, diag := range diagnostics {
        printDiagnostic(diag)
    }
}
```

### Example: Tab Validator

```go
type TabValidator struct{}

func (v *TabValidator) ProvideDiagnostics(uri, content string) []core.Diagnostic {
    var diagnostics []core.Diagnostic
    severity := core.SeverityWarning

    lines := strings.Split(content, "\n")
    for lineNum, line := range lines {
        for i, ch := range line {
            if ch == '\t' {
                code := core.NewStringCode("no-tabs")
                diagnostics = append(diagnostics, core.Diagnostic{
                    Range: core.Range{
                        Start: core.Position{Line: lineNum, Character: i},
                        End:   core.Position{Line: lineNum, Character: i + 1},
                    },
                    Severity: &severity,
                    Code:     &code,
                    Source:   "tabs",
                    Message:  "Use spaces instead of tabs",
                })
            }
        }
    }

    return diagnostics
}
```

### Registry Benefits

✅ **Composable**: Add/remove validators dynamically
✅ **Organized**: Keep validators separate and focused
✅ **Efficient**: Single pass through document (future optimization)

---

## Code Fixes

Provide quick fixes for diagnostics using `CodeFixProvider`:

### Creating a Code Fix Provider

```go
type TabCodeFixProvider struct{}

func (p *TabCodeFixProvider) ProvideCodeFixes(ctx core.CodeFixContext) []core.CodeAction {
    var actions []core.CodeAction

    for _, diag := range ctx.Diagnostics {
        // Only handle "no-tabs" diagnostics
        if diag.Code != nil && diag.Code.StringValue == "no-tabs" {
            // Create a text edit to replace tab with 4 spaces
            edit := core.TextEdit{
                Range:   diag.Range,
                NewText: "    ", // 4 spaces
            }

            // Create workspace edit
            workspaceEdit := &core.WorkspaceEdit{
                Changes: map[string][]core.TextEdit{
                    ctx.URI: {edit},
                },
            }

            // Create code action
            kind := core.CodeActionKindQuickFix
            actions = append(actions, core.CodeAction{
                Title: "Replace tab with spaces",
                Kind:  &kind,
                Edit:  workspaceEdit,
                Diagnostics: []core.Diagnostic{diag},
            })
        }
    }

    return actions
}
```

### Using Code Fixes

```go
func main() {
    content := "func main() {\n\tfmt.Println(\"hello\")\n}\n"
    uri := "file:///main.go"

    // 1. Get diagnostics
    tabValidator := &TabValidator{}
    diagnostics := tabValidator.ProvideDiagnostics(uri, content)

    // 2. Get code fixes for diagnostics
    fixProvider := &TabCodeFixProvider{}
    ctx := core.CodeFixContext{
        URI:         uri,
        Content:     content,
        Diagnostics: diagnostics,
    }
    fixes := fixProvider.ProvideCodeFixes(ctx)

    // 3. Apply first fix
    if len(fixes) > 0 {
        fix := fixes[0]
        if fix.Edit != nil {
            newContent := applyWorkspaceEdit(content, fix.Edit)
            fmt.Println("Fixed content:")
            fmt.Println(newContent)
        }
    }
}

func applyWorkspaceEdit(content string, edit *core.WorkspaceEdit) string {
    // Simple implementation for single file
    // Real implementation would handle multiple files
    for uri, edits := range edit.Changes {
        for _, e := range edits {
            // Apply each text edit
            content = applyTextEdit(content, e)
        }
    }
    return content
}
```

---

## LSP Server Integration

Now let's integrate validators with a full LSP server.

### Step 1: Set Up LSP Server

```go
package main

import (
    "context"
    "github.com/SCKelemen/lsp/core"
    "github.com/SCKelemen/lsp/adapter"
    protocol "github.com/SCKelemen/lsp/protocol"
    "github.com/tliron/glsp"
    protocol_server "github.com/tliron/glsp/protocol"
    "github.com/tliron/glsp/server"
)

type MyServer struct {
    documents       *core.DocumentManager
    diagnosticReg   *core.DiagnosticRegistry
    codeFixReg      *core.CodeFixRegistry
}

func NewMyServer() *MyServer {
    s := &MyServer{
        documents:     core.NewDocumentManager(),
        diagnosticReg: core.NewDiagnosticRegistry(),
        codeFixReg:    core.NewCodeFixRegistry(),
    }

    // Register validators
    s.diagnosticReg.Register(NewLineLengthValidator(80))
    s.diagnosticReg.Register(&TabValidator{})

    // Register fix providers
    s.codeFixReg.Register(&TabCodeFixProvider{})

    return s
}
```

### Step 2: Handle textDocument/didOpen

```go
func (s *MyServer) TextDocumentDidOpen(
    ctx *glsp.Context,
    params *protocol.DidOpenTextDocumentParams,
) error {
    uri := string(params.TextDocument.URI)
    content := params.TextDocument.Text
    version := int(params.TextDocument.Version)

    // Store document
    s.documents.Open(uri, content, version)

    // Run validators (core types with UTF-8)
    coreDiagnostics := s.diagnosticReg.ProvideDiagnostics(uri, content)

    // Convert to protocol (UTF-16)
    protocolDiagnostics := adapter.CoreToProtocolDiagnostics(coreDiagnostics, content)

    // Publish diagnostics to client
    ctx.Notify(protocol_server.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
        URI:         params.TextDocument.URI,
        Version:     &params.TextDocument.Version,
        Diagnostics: protocolDiagnostics,
    })

    return nil
}
```

### Step 3: Handle textDocument/didChange

```go
func (s *MyServer) TextDocumentDidChange(
    ctx *glsp.Context,
    params *protocol.DidChangeTextDocumentParams,
) error {
    uri := string(params.TextDocument.URI)

    // Get current document
    doc, ok := s.documents.Get(uri)
    if !ok {
        return fmt.Errorf("document not found: %s", uri)
    }

    // Apply changes
    for _, change := range params.ContentChanges {
        if change.Range == nil {
            // Full document update
            doc.Content = change.Text
        } else {
            // Incremental update
            // Convert protocol range (UTF-16) to core range (UTF-8)
            coreRange := adapter.ProtocolToCoreRange(*change.Range, doc.Content)

            // Apply edit
            doc.ApplyEdit(coreRange, change.Text)
        }
    }
    doc.Version = int(params.TextDocument.Version)

    // Re-run validators
    coreDiagnostics := s.diagnosticReg.ProvideDiagnostics(uri, doc.Content)

    // Convert and publish
    protocolDiagnostics := adapter.CoreToProtocolDiagnostics(coreDiagnostics, doc.Content)

    ctx.Notify(protocol_server.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
        URI:         params.TextDocument.URI,
        Version:     &params.TextDocument.Version,
        Diagnostics: protocolDiagnostics,
    })

    return nil
}
```

### Step 4: Handle textDocument/codeAction

```go
func (s *MyServer) TextDocumentCodeAction(
    ctx *glsp.Context,
    params *protocol.CodeActionParams,
) ([]protocol.CodeAction, error) {
    uri := string(params.TextDocument.URI)

    // Get document
    doc, ok := s.documents.Get(uri)
    if !ok {
        return nil, fmt.Errorf("document not found: %s", uri)
    }

    // Convert protocol diagnostics to core
    coreDiagnostics := adapter.ProtocolToCoreDiagnostics(params.Context.Diagnostics, doc.Content)

    // Get code fixes using core types
    coreCtx := core.CodeFixContext{
        URI:         uri,
        Content:     doc.Content,
        Diagnostics: coreDiagnostics,
    }
    coreFixes := s.codeFixReg.ProvideCodeFixes(coreCtx)

    // Convert code actions to protocol
    var protocolActions []protocol.CodeAction
    for _, coreFix := range coreFixes {
        protocolAction := adapter.CoreToProtocolCodeAction(coreFix, doc.Content)
        protocolActions = append(protocolActions, protocolAction)
    }

    return protocolActions, nil
}
```

### Step 5: Main Function

```go
func main() {
    // Create server
    myServer := NewMyServer()

    // Create GLSP handler
    handler := protocol_server.NewHandler()

    // Register handlers
    handler.Initialize = myServer.Initialize
    handler.Initialized = myServer.Initialized
    handler.Shutdown = myServer.Shutdown
    handler.SetTrace = myServer.SetTrace

    // Document synchronization
    handler.TextDocumentDidOpen = myServer.TextDocumentDidOpen
    handler.TextDocumentDidChange = myServer.TextDocumentDidChange
    handler.TextDocumentDidClose = myServer.TextDocumentDidClose

    // Language features
    handler.TextDocumentCodeAction = myServer.TextDocumentCodeAction

    // Create server
    srv := server.NewServer(handler, "my-language-server", false)

    // Start server
    srv.RunStdio()
}
```

---

## Complete Example

See `examples/` directory for complete working examples:

- `examples/codefix_provider_example.go` - Full validator with code fixes
- `examples/highlight_example.go` - Document highlighting provider
- `examples/core_handler_example.go` - LSP server integration patterns

---

## Best Practices

### 1. Keep Validators Focused
✅ One validator = one concern (line length, tabs, trailing whitespace)
❌ Don't create mega-validators that check everything

### 2. Use Meaningful Diagnostic Codes
```go
code := core.NewStringCode("line-too-long")  // ✅ Clear
code := core.NewIntCode(1)                    // ❌ Unclear
```

### 3. Provide Helpful Messages
```go
// ✅ Good: Specific and actionable
Message: "Line exceeds 80 characters (currently 95)"

// ❌ Bad: Vague
Message: "Line too long"
```

### 4. Set Appropriate Severity
- `SeverityError`: Code won't compile/run
- `SeverityWarning`: Should fix but not critical
- `SeverityInformation`: FYI, style suggestion
- `SeverityHint`: Subtle improvement

### 5. Handle UTF-8 Correctly
```go
// ✅ Use rune iteration for characters
for i, r := range line {
    // i is byte offset (UTF-8)
    // r is the rune
}

// ❌ Don't use byte indexing for characters
for i := 0; i < len(line); i++ {
    ch := line[i]  // Wrong for multi-byte characters
}
```

### 6. Test with Unicode
Always test validators with:
- CJK characters: `你好世界`
- Emoji: `😀👋🏻`
- Combining marks: `é` (e + combining accent)
- RTL text: `שלום עולם`

### 7. Make Code Fixes Optional
Not every diagnostic needs a code fix. Only provide fixes when:
- The fix is mechanical and obvious
- The fix is always safe to apply
- The fix doesn't require user decision

---

## Summary

**Validator Lifecycle:**

1. **Write validator** implementing `DiagnosticProvider`
2. **Test validator** with unit tests
3. **Register validator** in a registry
4. **Optionally add code fixes** with `CodeFixProvider`
5. **Integrate with LSP server** using adapters

**Key Points:**
- ✅ Validators use core types (UTF-8 offsets)
- ✅ Adapters convert at LSP boundaries (UTF-8 ↔ UTF-16)
- ✅ Same validator works in CLI and LSP server
- ✅ Registry pattern enables composition
- ✅ Test with Unicode content

For more details, see:
- [CORE_TYPES.md](CORE_TYPES.md) - Core types architecture
- [LSP_CAPABILITIES.md](LSP_CAPABILITIES.md) - Supported LSP features
- [README.md](README.md) - Quick start guide
