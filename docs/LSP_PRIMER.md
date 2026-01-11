# LSP Primer: Editor Features → LSP Capabilities

If you're familiar with VS Code but new to LSP servers, here's how editor features map to LSP capabilities.

## Code Intelligence

| You Know This As... | LSP Feature | What It Does |
|---------------------|-------------|--------------|
| Red/yellow squiggles under code | **Diagnostics** (`textDocument/publishDiagnostics`) | Show errors, warnings, and hints |
| Lightbulb with quick fixes | **Code Actions** (`textDocument/codeAction`) | Provide quick fixes and refactorings |
| Auto-complete dropdown | **Completion** (`textDocument/completion`) | Suggest code completions |
| Gray ghost text suggestions | **Inline Completions** (`textDocument/inlineCompletion`) | AI-powered inline code suggestions |
| Parameter hints `(paramName: ...)` | **Signature Help** (`textDocument/signatureHelp`) | Show function parameters while typing |
| Type hints `x: number` | **Inlay Hints** (`textDocument/inlayHint`) | Show inferred types and parameter names |

## Navigation

| You Know This As... | LSP Feature | What It Does |
|---------------------|-------------|--------------|
| Ctrl+Click / F12 to jump to definition | **Go to Definition** (`textDocument/definition`) | Jump to where something is defined |
| Find all references | **References** (`textDocument/references`) | Find all uses of a symbol |
| Hover tooltip with docs | **Hover** (`textDocument/hover`) | Show documentation on hover |
| Breadcrumbs / outline view | **Document Symbols** (`textDocument/documentSymbol`) | Show file structure |
| File symbol search (Ctrl+Shift+O) | **Document Symbols** | Quick navigation within file |
| Workspace symbol search (Ctrl+T) | **Workspace Symbols** (`workspace/symbol`) | Search symbols across workspace |

## Editing

| You Know This As... | LSP Feature | What It Does |
|---------------------|-------------|--------------|
| Format Document | **Formatting** (`textDocument/formatting`) | Auto-format code |
| Format Selection | **Range Formatting** (`textDocument/rangeFormatting`) | Format selected code |
| Rename Symbol (F2) | **Rename** (`textDocument/rename`) | Rename across all files |
| Fold/unfold code regions | **Folding Range** (`textDocument/foldingRange`) | Define collapsible regions |
| Color picker on `#FF0000` | **Document Color** (`textDocument/documentColor`) | Show color decorators |

## Advanced Features

| You Know This As... | LSP Feature | What It Does |
|---------------------|-------------|--------------|
| Syntax highlighting (semantic) | **Semantic Tokens** (`textDocument/semanticTokens`) | Enhanced syntax coloring |
| Code lenses (clickable hints above code) | **Code Lens** (`textDocument/codeLens`) | Inline actionable commands |
| Smart selection expansion | **Selection Range** (`textDocument/selectionRange`) | Expand/shrink selection intelligently |

## Getting Started

**Not sure what to implement?** Start with these four features:

1. **Diagnostics** - Show errors and warnings
2. **Completion** - Basic auto-complete
3. **Hover** - Show documentation
4. **Go to Definition** - Jump to definitions

These four features provide 80% of the value with 20% of the effort.

## Implementation Guides

For detailed implementation guides, see:

- [VALIDATORS.md](VALIDATORS.md) - Diagnostic providers
- [CODE_ACTIONS.md](CODE_ACTIONS.md) - Code action providers
- [NAVIGATION.md](NAVIGATION.md) - Definition & hover providers
- [SYMBOLS.md](SYMBOLS.md) - Document symbol providers
- [FOLDING.md](FOLDING.md) - Folding range providers
- [FORMATTING.md](FORMATTING.md) - Formatting providers

## Complete Feature List

For a comprehensive list of all LSP capabilities and their support status in this library, see [LSP_CAPABILITIES.md](LSP_CAPABILITIES.md).
