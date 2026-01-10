# LSP Capabilities Support Matrix

This document shows all Language Server Protocol capabilities and their support status in this library.

## Legend

| Symbol | Meaning |
|--------|---------|
| ✅ | Fully supported with core types (UTF-8) |
| 🔧 | Protocol types only (no core types yet) |
| 📋 | Planned/In Progress |
| ❌ | Not implemented |

**Usage Types:**
- **LSP**: Can be used in LSP server with protocol conversion
- **CLI**: Can be used directly in CLI tools without LSP server
- **Both**: Can be used in both LSP servers and CLI tools

---

## General Messages

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `initialize` | 🔧 | LSP | - | - | Server setup |
| `initialized` | 🔧 | LSP | - | - | Notification after init |
| `shutdown` | 🔧 | LSP | - | - | Server shutdown |
| `exit` | 🔧 | LSP | - | - | Server exit |
| `$/cancelRequest` | 🔧 | LSP | - | - | Request cancellation |
| `$/progress` | 🔧 | LSP | - | - | Progress reporting |

---

## Text Document Synchronization

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `textDocument/didOpen` | ✅ | Both | `Position`, `Range` | - | Open document notification |
| `textDocument/didChange` | ✅ | Both | `Position`, `Range`, `TextEdit` | - | Document change notification |
| `textDocument/willSave` | 🔧 | LSP | - | - | Pre-save notification |
| `textDocument/willSaveWaitUntil` | 🔧 | LSP | - | - | Pre-save with edits |
| `textDocument/didSave` | 🔧 | LSP | - | - | Post-save notification |
| `textDocument/didClose` | 🔧 | LSP | - | - | Close document notification |

---

## Language Features

### Diagnostics

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `textDocument/publishDiagnostics` | ✅ | Both | `Diagnostic` | `DiagnosticProvider` | Errors, warnings, hints |
| `textDocument/diagnostic` (pull) | ✅ | Both | `Diagnostic` | `DiagnosticProvider` | LSP 3.17 pull model |
| `workspace/diagnostic` | ✅ | Both | `Diagnostic` | `DiagnosticProvider` | Workspace-wide diagnostics |

### Code Completion

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `textDocument/completion` | ✅ | Both | `CompletionItem`, `CompletionList` | `CompletionProvider` | Code completion suggestions |
| `completionItem/resolve` | ✅ | Both | `CompletionItem` | `CompletionItemResolveProvider` | Lazy load completion details |

### Hover

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `textDocument/hover` | ✅ | Both | `HoverInfo` | `HoverProvider` | Hover information |

### Signature Help

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `textDocument/signatureHelp` | ✅ | Both | `SignatureHelp`, `SignatureInformation` | `SignatureHelpProvider` | Parameter hints |

### Go To

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `textDocument/declaration` | ✅ | Both | `Location`, `LocationLink` | `DefinitionProvider` | Go to declaration |
| `textDocument/definition` | ✅ | Both | `Location`, `LocationLink` | `DefinitionProvider` | Go to definition |
| `textDocument/typeDefinition` | ✅ | Both | `Location`, `LocationLink` | `DefinitionProvider` | Go to type definition |
| `textDocument/implementation` | ✅ | Both | `Location`, `LocationLink` | `DefinitionProvider` | Go to implementation |

### References

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `textDocument/references` | ✅ | Both | `Location`, `ReferenceContext` | `ReferencesProvider` | Find all references |

### Document Symbols

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `textDocument/documentSymbol` | ✅ | Both | `DocumentSymbol` | `DocumentSymbolProvider` | Symbol tree/outline |
| `textDocument/documentHighlight` | ✅ | Both | `DocumentHighlight`, `DocumentHighlightKind` | `DocumentHighlightProvider` | Symbol highlighting |

### Code Actions

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `textDocument/codeAction` | ✅ | Both | `CodeAction` | `CodeFixProvider` | Quick fixes, refactorings |
| `codeAction/resolve` | ✅ | Both | `CodeAction` | `CodeFixProvider` | Lazy load code action details |

### Code Lens

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `textDocument/codeLens` | ✅ | Both | `CodeLens` | `CodeLensProvider` | Inline actionable commands |
| `codeLens/resolve` | ✅ | Both | `CodeLens` | `CodeLensResolveProvider` | Lazy load code lens details |

### Document Links

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `textDocument/documentLink` | ✅ | Both | `DocumentLink` | `DocumentLinkProvider` | Clickable links in document |
| `documentLink/resolve` | ✅ | Both | `DocumentLink` | `DocumentLinkResolveProvider` | Resolve link target |

### Color

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `textDocument/documentColor` | 📋 | Both | - | - | Color decorators |
| `textDocument/colorPresentation` | 📋 | Both | - | - | Color picker formats |

### Formatting

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `textDocument/formatting` | ✅ | Both | `TextEdit`, `FormattingOptions` | `FormattingProvider` | Format entire document |
| `textDocument/rangeFormatting` | ✅ | Both | `TextEdit`, `FormattingOptions` | `RangeFormattingProvider` | Format selection |
| `textDocument/onTypeFormatting` | 📋 | Both | `TextEdit` | - | Format on type (e.g., after ;) |

### Rename

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `textDocument/rename` | ✅ | Both | `WorkspaceEdit` | `RenameProvider` | Rename symbol |
| `textDocument/prepareRename` | ✅ | Both | `Range` | `PrepareRenameProvider` | Validate rename position |

### Folding

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `textDocument/foldingRange` | ✅ | Both | `FoldingRange` | `FoldingRangeProvider` | Code folding regions |

### Selection Range

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `textDocument/selectionRange` | 📋 | Both | - | - | Smart selection expansion |

### Call Hierarchy

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `textDocument/prepareCallHierarchy` | 📋 | Both | - | - | Prepare call hierarchy |
| `callHierarchy/incomingCalls` | 📋 | Both | - | - | Find callers |
| `callHierarchy/outgoingCalls` | 📋 | Both | - | - | Find callees |

### Type Hierarchy

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `textDocument/prepareTypeHierarchy` | 📋 | Both | - | - | Prepare type hierarchy |
| `typeHierarchy/supertypes` | 📋 | Both | - | - | Find supertypes |
| `typeHierarchy/subtypes` | 📋 | Both | - | - | Find subtypes |

### Semantic Tokens

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `textDocument/semanticTokens/full` | 📋 | LSP | - | - | Full semantic tokens |
| `textDocument/semanticTokens/full/delta` | 📋 | LSP | - | - | Incremental semantic tokens |
| `textDocument/semanticTokens/range` | 📋 | LSP | - | - | Range semantic tokens |

### Linked Editing

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `textDocument/linkedEditingRange` | 📋 | Both | - | - | Linked editing ranges |

### Moniker

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `textDocument/moniker` | 📋 | Both | - | - | Symbol moniker |

### Inlay Hint

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `textDocument/inlayHint` | ✅ | Both | `InlayHint`, `InlayHintKind` | `InlayHintsProvider` | Inline hints (types, params) |
| `inlayHint/resolve` | 🔧 | LSP | `InlayHint` | - | Resolve inlay hint details |

### Inline Value

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `textDocument/inlineValue` | 📋 | LSP | - | - | Inline values during debug |

---

## Workspace Features

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `workspace/symbol` | ✅ | Both | `WorkspaceSymbol` | `WorkspaceSymbolProvider` | Workspace-wide symbol search |
| `workspace/executeCommand` | 🔧 | LSP | `Command` | - | Execute custom command |
| `workspace/applyEdit` | ✅ | Both | `WorkspaceEdit` | - | Apply workspace edit |
| `workspace/willCreateFiles` | ✅ | Both | `CreateFile` | - | Pre-create notification |
| `workspace/didCreateFiles` | ✅ | Both | `CreateFile` | - | Post-create notification |
| `workspace/willRenameFiles` | ✅ | Both | `RenameFile` | - | Pre-rename notification |
| `workspace/didRenameFiles` | ✅ | Both | `RenameFile` | - | Post-rename notification |
| `workspace/willDeleteFiles` | ✅ | Both | `DeleteFile` | - | Pre-delete notification |
| `workspace/didDeleteFiles` | ✅ | Both | `DeleteFile` | - | Post-delete notification |
| `workspace/didChangeWatchedFiles` | 🔧 | LSP | - | - | File system change notification |
| `workspace/didChangeWorkspaceFolders` | 🔧 | LSP | - | - | Workspace folder changes |
| `workspace/didChangeConfiguration` | 🔧 | LSP | - | - | Configuration changes |

---

## Window Features

| Capability | Status | Usage | Core Type | Provider Interface | Notes |
|------------|--------|-------|-----------|-------------------|-------|
| `window/showMessage` | 🔧 | LSP | - | - | Show message to user |
| `window/showMessageRequest` | 🔧 | LSP | - | - | Show message with actions |
| `window/logMessage` | 🔧 | LSP | - | - | Log message |
| `window/workDoneProgress/create` | 🔧 | LSP | - | - | Create progress indicator |
| `window/workDoneProgress/cancel` | 🔧 | LSP | - | - | Cancel progress |

---

## Summary Statistics

### Implementation Status
- ✅ **Fully Supported with Core Types**: 33 capabilities
- 🔧 **Protocol Types Only**: 15 capabilities
- 📋 **Planned**: 12 capabilities
- ❌ **Not Implemented**: 0 capabilities

### Usage Breakdown
- **CLI + LSP (Both)**: 33 capabilities
- **LSP Only**: 15 capabilities
- **CLI Only**: 0 capabilities

### Core Types Available
- ✅ `Position` (UTF-8 offsets)
- ✅ `Range` (UTF-8 offsets)
- ✅ `Location` (UTF-8 offsets)
- ✅ `LocationLink` (UTF-8 offsets)
- ✅ `Diagnostic` (all severity levels and tags)
- ✅ `TextEdit` (UTF-8 offsets)
- ✅ `WorkspaceEdit` (create/rename/delete files)
- ✅ `DocumentSymbol` (hierarchical with UTF-8 offsets)
- ✅ `DocumentHighlight` (text, read, write kinds)
- ✅ `CodeAction` (quick fixes, refactorings)
- ✅ `FoldingRange` (comment, imports, region)
- ✅ `CompletionItem` (full LSP completion support)
- ✅ `CompletionList` (with incomplete flag)
- ✅ `SignatureHelp` (parameter hints)
- ✅ `FormattingOptions` (type-safe struct)
- ✅ `HoverInfo` (hover documentation)
- ✅ `CodeLens` (inline commands)
- ✅ `Command` (executable commands)
- ✅ `InlayHint` (parameter names, inferred types)
- ✅ `DocumentLink` (clickable URIs in documents)
- ✅ `WorkspaceSymbol` (workspace-wide symbol search)
- ✅ `ReferenceContext` (reference search options)

### Provider Interfaces Available
- ✅ `DiagnosticProvider`
- ✅ `CodeFixProvider`
- ✅ `FoldingRangeProvider`
- ✅ `DocumentSymbolProvider`
- ✅ `DocumentHighlightProvider`
- ✅ `DefinitionProvider`
- ✅ `HoverProvider`
- ✅ `FormattingProvider`
- ✅ `RangeFormattingProvider`
- ✅ `CompletionProvider`
- ✅ `CompletionItemResolveProvider`
- ✅ `SignatureHelpProvider`
- ✅ `RenameProvider`
- ✅ `PrepareRenameProvider`
- ✅ `CodeLensProvider`
- ✅ `CodeLensResolveProvider`
- ✅ `InlayHintsProvider`
- ✅ `ReferencesProvider`
- ✅ `DocumentLinkProvider`
- ✅ `DocumentLinkResolveProvider`
- ✅ `WorkspaceSymbolProvider`

---

## Key Features

### ✅ What Works Now

#### For CLI Tools (No LSP Server Needed)
- **Diagnostics**: Find errors, warnings in files
- **Code Fixes**: Generate fixes for diagnostics
- **Formatting**: Format entire documents or ranges
- **Completion**: Generate code completion suggestions
- **Symbol Navigation**: Build symbol trees, find definitions
- **Folding**: Detect foldable regions
- **Document Management**: Track document state with `DocumentManager`

#### For LSP Servers
- All of the above **plus**:
- Automatic UTF-16 ↔ UTF-8 conversion at boundaries
- Full protocol message handling
- Client-server communication
- Progress reporting
- Configuration management

### 🎯 Core Design Principles

1. **UTF-8 First**: All core types use UTF-8 byte offsets for natural Go string handling
2. **Reusable Logic**: Write providers once, use in CLI and LSP
3. **Clean Boundaries**: Protocol conversion only at handler edges
4. **Type Safety**: Strongly-typed core structs instead of `map[string]any`
5. **Provider Pattern**: Composable providers with registry support

### 📋 Planned Features

The following features are planned for future releases:
- Color decorators
- Selection range expansion
- Call hierarchy
- Type hierarchy
- Semantic tokens
- On-type formatting
- Linked editing range

These will follow the same pattern: core types with UTF-8 offsets, provider interfaces, and adapters for protocol conversion.

---

## Usage Examples

### Using in CLI Tool
```go
// No LSP server needed - work directly with core types
registry := core.NewDiagnosticRegistry()
registry.Register(&MyLinter{})

diagnostics := registry.ProvideDiagnostics(uri, content)
// Use diagnostics directly with UTF-8 offsets
```

### Using in LSP Server
```go
// Convert at boundaries only
func (s *Server) TextDocumentDidOpen(ctx *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
    uri := string(params.TextDocument.URI)
    content := params.TextDocument.Text

    // Use same provider as CLI
    diagnostics := s.registry.ProvideDiagnostics(uri, content)

    // Convert to protocol at boundary
    protocolDiags := adapter_3_16.CoreToProtocolDiagnostics(diagnostics, content)

    // Send to client
    ctx.Notify(...)
    return nil
}
```

---

## See Also

- [CORE_TYPES.md](CORE_TYPES.md) - Detailed architecture guide
- [README.md](README.md) - Getting started
- [examples/](examples/) - Complete working examples
