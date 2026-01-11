# Code Actions Guide

Complete guide to implementing code action providers in LSP, from quick fixes to refactorings.

**LSP Specification References:**
- [Code Action Request](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#textDocument_codeAction)
- [Code Action Types](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#codeAction)
- [Code Action Kinds](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.16/specification/#codeActionKind)

## Table of Contents

1. [Core Concepts](#core-concepts)
2. [Your First Code Action Provider](#your-first-code-action-provider)
3. [Different Types of Code Actions](#different-types-of-code-actions)
4. [Testing Code Action Providers](#testing-code-action-providers)
5. [Composing Multiple Providers](#composing-multiple-providers)
6. [LSP Server Integration](#lsp-server-integration)

## Core Concepts

### What Are Code Actions?

Code actions are operations that can be performed on code, such as:
- **Quick fixes**: Fix problems identified by diagnostics
- **Refactorings**: Extract functions, rename variables, inline code
- **Source actions**: Organize imports, fix all issues, format code

### Provider Interface

```go
import "github.com/SCKelemen/lsp/core"

// CodeFixProvider provides code actions for a document
type CodeFixProvider interface {
    ProvideCodeFixes(ctx CodeFixContext) []CodeAction
}
```

### Code Fix Context

The context provides everything you need:

```go
type CodeFixContext struct {
    URI         string              // Document URI
    Content     string              // Document content
    Range       Range               // Selected or requested range
    Diagnostics []Diagnostic        // Diagnostics in the range
    Only        []CodeActionKind    // Requested action kinds (empty = all)
    TriggerKind CodeActionTriggerKind // Invoked or Automatic
}
```

### Code Action Structure

```go
type CodeAction struct {
    Title       string              // Human-readable title
    Kind        *CodeActionKind     // quickfix, refactor, source, etc.
    Diagnostics []Diagnostic        // Diagnostics this fixes
    IsPreferred bool                // Preferred action in group
    Disabled    *CodeActionDisabled // Why disabled (if applicable)
    Edit        *WorkspaceEdit      // Changes to apply
    Command     *Command            // Command to execute
    Data        interface{}         // Custom data for resolve
}
```

### Code Action Kinds

```go
const (
    CodeActionKindQuickFix              = "quickfix"
    CodeActionKindRefactor              = "refactor"
    CodeActionKindRefactorExtract       = "refactor.extract"
    CodeActionKindRefactorInline        = "refactor.inline"
    CodeActionKindRefactorRewrite       = "refactor.rewrite"
    CodeActionKindSource                = "source"
    CodeActionKindSourceOrganizeImports = "source.organizeImports"
    CodeActionKindSourceFixAll          = "source.fixAll"
)
```

## Your First Code Action Provider

Let's create a provider that removes unused imports in Go code.

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

// UnusedImportProvider provides code actions to remove unused imports.
type UnusedImportProvider struct{}

func (p *UnusedImportProvider) ProvideCodeFixes(ctx core.CodeFixContext) []core.CodeAction {
    var actions []core.CodeAction

    // Only provide actions for Go files
    if !strings.HasSuffix(ctx.URI, ".go") {
        return nil
    }

    // Find unused imports
    unusedImports := p.findUnusedImports(ctx.Content)
    if len(unusedImports) == 0 {
        return nil
    }

    // Create an action to remove all unused imports
    kind := core.CodeActionKindSourceOrganizeImports
    actions = append(actions, core.CodeAction{
        Title:       "Remove unused imports",
        Kind:        &kind,
        IsPreferred: true,
        Edit:        p.createRemovalEdit(ctx.URI, ctx.Content, unusedImports),
    })

    // Create individual actions for each unused import
    for _, imp := range unusedImports {
        actions = append(actions, p.createSingleRemovalAction(ctx.URI, ctx.Content, imp))
    }

    return actions
}
```

### Step 2: Implement Helper Functions

```go
// importInfo holds information about an import
type importInfo struct {
    Path  string
    Name  string
    Start core.Position
    End   core.Position
}

func (p *UnusedImportProvider) findUnusedImports(content string) []importInfo {
    fset := token.NewFileSet()
    f, err := parser.ParseFile(fset, "", content, parser.ImportsOnly)
    if err != nil {
        return nil
    }

    var unused []importInfo

    // Check each import
    for _, imp := range f.Imports {
        importPath := strings.Trim(imp.Path.Value, `"`)

        // Simple heuristic: check if import path appears in content
        // A real implementation would do proper usage analysis
        if !p.isImportUsed(content, importPath, imp) {
            // Convert token positions to core.Position
            start := fset.Position(imp.Pos())
            end := fset.Position(imp.End())

            unused = append(unused, importInfo{
                Path: importPath,
                Name: p.getImportName(imp, importPath),
                Start: core.Position{
                    Line:      start.Line - 1, // 0-based
                    Character: start.Column - 1,
                },
                End: core.Position{
                    Line:      end.Line - 1,
                    Character: end.Column - 1,
                },
            })
        }
    }

    return unused
}

func (p *UnusedImportProvider) getImportName(imp *ast.ImportSpec, path string) string {
    if imp.Name != nil {
        return imp.Name.Name
    }
    // Get last component of path
    parts := strings.Split(path, "/")
    return parts[len(parts)-1]
}

func (p *UnusedImportProvider) isImportUsed(content string, importPath string, imp *ast.ImportSpec) bool {
    // Get the identifier that would be used in code
    name := p.getImportName(imp, importPath)

    // Simple check: does the identifier appear in the code?
    // A real implementation would use proper AST analysis
    lines := strings.Split(content, "\n")
    for _, line := range lines {
        // Skip import declarations
        if strings.Contains(line, "import") {
            continue
        }
        if strings.Contains(line, name+".") {
            return true
        }
    }

    return false
}
```

### Step 3: Create Edit Functions

```go
func (p *UnusedImportProvider) createRemovalEdit(uri, content string, unused []importInfo) *core.WorkspaceEdit {
    var edits []core.TextEdit

    // Sort in reverse order to avoid position shifts
    // (In real code, you'd implement proper sorting)

    for _, imp := range unused {
        // Find the full line range to remove (including newline)
        lineStart := core.Position{Line: imp.Start.Line, Character: 0}
        lineEnd := core.Position{Line: imp.Start.Line + 1, Character: 0}

        edits = append(edits, core.TextEdit{
            Range: core.Range{
                Start: lineStart,
                End:   lineEnd,
            },
            NewText: "",
        })
    }

    return &core.WorkspaceEdit{
        Changes: map[string][]core.TextEdit{
            uri: edits,
        },
    }
}

func (p *UnusedImportProvider) createSingleRemovalAction(uri, content string, imp importInfo) core.CodeAction {
    kind := core.CodeActionKindQuickFix

    lineStart := core.Position{Line: imp.Start.Line, Character: 0}
    lineEnd := core.Position{Line: imp.Start.Line + 1, Character: 0}

    return core.CodeAction{
        Title: "Remove unused import \"" + imp.Path + "\"",
        Kind:  &kind,
        Edit: &core.WorkspaceEdit{
            Changes: map[string][]core.TextEdit{
                uri: {
                    {
                        Range: core.Range{
                            Start: lineStart,
                            End:   lineEnd,
                        },
                        NewText: "",
                    },
                },
            },
        },
    }
}
```

### Step 4: Use the Provider

```go
func main() {
    content := `package main

import (
    "fmt"
    "strings"
    "unused"
)

func main() {
    fmt.Println(strings.ToUpper("hello"))
}
`

    provider := &UnusedImportProvider{}
    ctx := core.CodeFixContext{
        URI:     "file:///example.go",
        Content: content,
        Range:   core.Range{}, // Full document
    }

    actions := provider.ProvideCodeFixes(ctx)

    for _, action := range actions {
        fmt.Printf("Action: %s (kind: %s)\n", action.Title, *action.Kind)
    }
}
```

## Different Types of Code Actions

### Quick Fixes (Diagnostic-Based)

Quick fixes resolve specific diagnostics:

```go
type SpellingFixProvider struct{}

func (p *SpellingFixProvider) ProvideCodeFixes(ctx core.CodeFixContext) []core.CodeAction {
    var actions []core.CodeAction

    // Only process diagnostics with spelling errors
    for _, diag := range ctx.Diagnostics {
        if diag.Code == nil || diag.Code.StringValue != "spelling-error" {
            continue
        }

        // Get misspelled word
        word := p.extractWord(ctx.Content, diag.Range)

        // Get suggestions
        suggestions := p.getSuggestions(word)

        // Create an action for each suggestion
        for _, suggestion := range suggestions {
            kind := core.CodeActionKindQuickFix
            actions = append(actions, core.CodeAction{
                Title: "Change to \"" + suggestion + "\"",
                Kind:  &kind,
                Diagnostics: []core.Diagnostic{diag},
                Edit: &core.WorkspaceEdit{
                    Changes: map[string][]core.TextEdit{
                        ctx.URI: {
                            {
                                Range:   diag.Range,
                                NewText: suggestion,
                            },
                        },
                    },
                },
            })
        }
    }

    return actions
}

func (p *SpellingFixProvider) extractWord(content string, r core.Range) string {
    lines := strings.Split(content, "\n")
    if r.Start.Line >= len(lines) {
        return ""
    }
    line := lines[r.Start.Line]
    if r.End.Character > len(line) {
        return line[r.Start.Character:]
    }
    return line[r.Start.Character:r.End.Character]
}

func (p *SpellingFixProvider) getSuggestions(word string) []string {
    // In real code, use a spell checker library
    return []string{"suggestion1", "suggestion2"}
}
```

### Refactoring: Extract Function

Extract selected code into a new function:

```go
type ExtractFunctionProvider struct{}

func (p *ExtractFunctionProvider) ProvideCodeFixes(ctx core.CodeFixContext) []core.CodeAction {
    // Only provide extraction if user selected code
    if ctx.Range.Start == ctx.Range.End {
        return nil
    }

    // Check if selection is valid for extraction
    if !p.isValidExtraction(ctx.Content, ctx.Range) {
        return nil
    }

    kind := core.CodeActionKindRefactorExtract

    return []core.CodeAction{
        {
            Title: "Extract function",
            Kind:  &kind,
            Edit:  p.createExtractionEdit(ctx.URI, ctx.Content, ctx.Range),
        },
    }
}

func (p *ExtractFunctionProvider) isValidExtraction(content string, r core.Range) bool {
    // Validate that the selection is complete statements
    // This is simplified - real implementation needs proper AST analysis
    return r.Start.Line < r.End.Line || r.End.Character > r.Start.Character
}

func (p *ExtractFunctionProvider) createExtractionEdit(uri, content string, r core.Range) *core.WorkspaceEdit {
    // Extract the selected code
    selected := p.extractText(content, r)

    // Generate function name
    funcName := "extracted"

    // Create the new function
    newFunc := "\nfunc " + funcName + "() {\n\t" + strings.TrimSpace(selected) + "\n}\n"

    // Find insertion point (before current function)
    insertPos := p.findInsertionPoint(content, r.Start)

    return &core.WorkspaceEdit{
        Changes: map[string][]core.TextEdit{
            uri: {
                // Insert new function
                {
                    Range: core.Range{
                        Start: insertPos,
                        End:   insertPos,
                    },
                    NewText: newFunc,
                },
                // Replace selection with function call
                {
                    Range:   r,
                    NewText: "\t" + funcName + "()",
                },
            },
        },
    }
}

func (p *ExtractFunctionProvider) extractText(content string, r core.Range) string {
    lines := strings.Split(content, "\n")
    if r.Start.Line == r.End.Line {
        return lines[r.Start.Line][r.Start.Character:r.End.Character]
    }

    // Multi-line extraction
    var result strings.Builder
    for i := r.Start.Line; i <= r.End.Line; i++ {
        if i >= len(lines) {
            break
        }
        line := lines[i]

        if i == r.Start.Line {
            result.WriteString(line[r.Start.Character:])
        } else if i == r.End.Line {
            result.WriteString("\n")
            result.WriteString(line[:r.End.Character])
        } else {
            result.WriteString("\n")
            result.WriteString(line)
        }
    }

    return result.String()
}

func (p *ExtractFunctionProvider) findInsertionPoint(content string, pos core.Position) core.Position {
    // Find the start of the current function
    // Simplified: just insert before current position
    return core.Position{Line: pos.Line, Character: 0}
}
```

### Source Actions: Organize Imports

```go
type OrganizeImportsProvider struct{}

func (p *OrganizeImportsProvider) ProvideCodeFixes(ctx core.CodeFixContext) []core.CodeAction {
    // Only for Go files
    if !strings.HasSuffix(ctx.URI, ".go") {
        return nil
    }

    // Check if organize imports was requested
    if len(ctx.Only) > 0 {
        found := false
        for _, kind := range ctx.Only {
            if kind == core.CodeActionKindSourceOrganizeImports {
                found = true
                break
            }
        }
        if !found {
            return nil
        }
    }

    kind := core.CodeActionKindSourceOrganizeImports

    return []core.CodeAction{
        {
            Title: "Organize Imports",
            Kind:  &kind,
            Edit:  p.organizeImports(ctx.URI, ctx.Content),
        },
    }
}

func (p *OrganizeImportsProvider) organizeImports(uri, content string) *core.WorkspaceEdit {
    fset := token.NewFileSet()
    f, err := parser.ParseFile(fset, "", content, parser.ImportsOnly)
    if err != nil {
        return nil
    }

    if len(f.Imports) == 0 {
        return nil
    }

    // Collect imports
    var stdImports, thirdPartyImports []string

    for _, imp := range f.Imports {
        path := strings.Trim(imp.Path.Value, `"`)

        // Categorize import
        if p.isStdLib(path) {
            stdImports = append(stdImports, imp.Path.Value)
        } else {
            thirdPartyImports = append(thirdPartyImports, imp.Path.Value)
        }
    }

    // Sort each group
    sort.Strings(stdImports)
    sort.Strings(thirdPartyImports)

    // Build new import block
    var newImports strings.Builder
    newImports.WriteString("import (\n")

    for _, imp := range stdImports {
        newImports.WriteString("\t" + imp + "\n")
    }

    if len(stdImports) > 0 && len(thirdPartyImports) > 0 {
        newImports.WriteString("\n")
    }

    for _, imp := range thirdPartyImports {
        newImports.WriteString("\t" + imp + "\n")
    }

    newImports.WriteString(")\n")

    // Find import block range
    importRange := p.findImportRange(content, f)

    return &core.WorkspaceEdit{
        Changes: map[string][]core.TextEdit{
            uri: {
                {
                    Range:   importRange,
                    NewText: newImports.String(),
                },
            },
        },
    }
}

func (p *OrganizeImportsProvider) isStdLib(path string) bool {
    // Simple heuristic: stdlib packages don't have dots
    return !strings.Contains(path, ".")
}

func (p *OrganizeImportsProvider) findImportRange(content string, f *ast.File) core.Range {
    // Find the entire import block
    // Simplified implementation
    lines := strings.Split(content, "\n")

    var start, end core.Position
    inImport := false

    for i, line := range lines {
        if strings.Contains(line, "import") {
            if !inImport {
                start = core.Position{Line: i, Character: 0}
                inImport = true
            }
        } else if inImport && strings.Contains(line, ")") {
            end = core.Position{Line: i + 1, Character: 0}
            break
        }
    }

    return core.Range{Start: start, End: end}
}
```

## Testing Code Action Providers

### Basic Test Structure

```go
func TestUnusedImportProvider(t *testing.T) {
    tests := []struct {
        name      string
        content   string
        wantCount int
        wantKinds []core.CodeActionKind
    }{
        {
            name: "no unused imports",
            content: `package main

import "fmt"

func main() {
    fmt.Println("hello")
}`,
            wantCount: 0,
        },
        {
            name: "one unused import",
            content: `package main

import (
    "fmt"
    "strings"
)

func main() {
    fmt.Println("hello")
}`,
            wantCount: 2, // One "remove all" + one individual
            wantKinds: []core.CodeActionKind{
                core.CodeActionKindSourceOrganizeImports,
                core.CodeActionKindQuickFix,
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            provider := &UnusedImportProvider{}
            ctx := core.CodeFixContext{
                URI:     "file:///test.go",
                Content: tt.content,
                Range:   core.Range{},
            }

            actions := provider.ProvideCodeFixes(ctx)

            if len(actions) != tt.wantCount {
                t.Errorf("got %d actions, want %d", len(actions), tt.wantCount)
            }

            if tt.wantKinds != nil {
                for i, action := range actions {
                    if i < len(tt.wantKinds) {
                        if action.Kind == nil || *action.Kind != tt.wantKinds[i] {
                            t.Errorf("action %d: got kind %v, want %v",
                                i, action.Kind, tt.wantKinds[i])
                        }
                    }
                }
            }
        })
    }
}
```

### Testing Edit Application

```go
func TestExtractFunctionEdit(t *testing.T) {
    content := `package main

func main() {
    x := 1
    y := 2
    sum := x + y
    println(sum)
}
`

    provider := &ExtractFunctionProvider{}

    // Select the calculation lines
    selectedRange := core.Range{
        Start: core.Position{Line: 3, Character: 4},
        End:   core.Position{Line: 5, Character: 16},
    }

    ctx := core.CodeFixContext{
        URI:     "file:///test.go",
        Content: content,
        Range:   selectedRange,
    }

    actions := provider.ProvideCodeFixes(ctx)

    if len(actions) == 0 {
        t.Fatal("expected at least one action")
    }

    action := actions[0]

    // Apply the edit
    newContent := applyWorkspaceEdit(content, action.Edit)

    // Verify the function was created
    if !strings.Contains(newContent, "func extracted()") {
        t.Error("extracted function not found in result")
    }

    // Verify the function call was inserted
    if !strings.Contains(newContent, "extracted()") {
        t.Error("function call not found in result")
    }
}

// Helper to apply workspace edits (simplified)
func applyWorkspaceEdit(content string, edit *core.WorkspaceEdit) string {
    for _, edits := range edit.Changes {
        for _, e := range edits {
            content = applyTextEdit(content, e)
        }
    }
    return content
}

func applyTextEdit(content string, edit core.TextEdit) string {
    lines := strings.Split(content, "\n")

    if edit.Range.Start.Line == edit.Range.End.Line {
        line := lines[edit.Range.Start.Line]
        newLine := line[:edit.Range.Start.Character] +
                   edit.NewText +
                   line[edit.Range.End.Character:]
        lines[edit.Range.Start.Line] = newLine
        return strings.Join(lines, "\n")
    }

    // Multi-line edit
    before := lines[edit.Range.Start.Line][:edit.Range.Start.Character]
    after := lines[edit.Range.End.Line][edit.Range.End.Character:]
    lines[edit.Range.Start.Line] = before + edit.NewText + after
    lines = append(lines[:edit.Range.Start.Line+1], lines[edit.Range.End.Line+1:]...)

    return strings.Join(lines, "\n")
}
```

## Composing Multiple Providers

### Using the Registry

```go
// Create a registry for all code action providers
registry := core.NewCodeFixRegistry()

// Register different types of providers
registry.Register(&UnusedImportProvider{})
registry.Register(&SpellingFixProvider{})
registry.Register(&ExtractFunctionProvider{})
registry.Register(&OrganizeImportsProvider{})

// Get all applicable actions
ctx := core.CodeFixContext{
    URI:         "file:///example.go",
    Content:     content,
    Range:       selectedRange,
    Diagnostics: diagnostics,
}

actions := registry.ProvideCodeFixes(ctx)

// Actions from all providers are combined
for _, action := range actions {
    fmt.Printf("%s (%s)\n", action.Title, *action.Kind)
}
```

### Filtering by Kind

```go
// Client requests only quick fixes
ctx := core.CodeFixContext{
    URI:     "file:///example.go",
    Content: content,
    Range:   selectedRange,
    Only:    []core.CodeActionKind{core.CodeActionKindQuickFix},
}

actions := registry.ProvideCodeFixes(ctx)

// Only quick fix providers should respond
// Providers should check ctx.Only and filter appropriately
```

## LSP Server Integration

### Complete Handler Example

```go
import (
    "github.com/SCKelemen/lsp"
    "github.com/SCKelemen/lsp/adapter"
    "github.com/SCKelemen/lsp/core"
    protocol "github.com/SCKelemen/lsp/protocol"
)

type MyServer struct {
    documents  *core.DocumentManager
    codeActions *core.CodeFixRegistry
}

func NewMyServer() *MyServer {
    s := &MyServer{
        documents:  core.NewDocumentManager(),
        codeActions: core.NewCodeFixRegistry(),
    }

    // Register all code action providers
    s.codeActions.Register(&UnusedImportProvider{})
    s.codeActions.Register(&SpellingFixProvider{})
    s.codeActions.Register(&ExtractFunctionProvider{})
    s.codeActions.Register(&OrganizeImportsProvider{})

    return s
}

func (s *MyServer) TextDocumentCodeAction(
    context *glsp.Context,
    params *protocol.CodeActionParams,
) ([]protocol.CodeAction, error) {
    uri := string(params.TextDocument.URI)
    content := s.documents.GetContent(uri)

    if content == "" {
        return nil, nil
    }

    // Convert protocol range to core range (UTF-16 to UTF-8)
    coreRange := adapter.ProtocolToCoreRange(params.Range, content)

    // Convert protocol diagnostics to core diagnostics
    coreDiags := adapter.ProtocolToCoreDiagnostics(params.Context.Diagnostics, content)

    // Convert requested kinds
    var only []core.CodeActionKind
    if params.Context.Only != nil {
        for _, k := range params.Context.Only {
            only = append(only, core.CodeActionKind(k))
        }
    }

    // Create context for providers
    ctx := core.CodeFixContext{
        URI:         uri,
        Content:     content,
        Range:       coreRange,
        Diagnostics: coreDiags,
        Only:        only,
        TriggerKind: core.CodeActionTriggerKind(params.Context.TriggerKind),
    }

    // Get actions from all providers
    coreActions := s.codeActions.ProvideCodeFixes(ctx)

    // Convert back to protocol
    var protocolActions []protocol.CodeAction
    for _, action := range coreActions {
        protocolAction := s.coreToProtocolCodeAction(action, content)
        protocolActions = append(protocolActions, protocolAction)
    }

    return protocolActions, nil
}

func (s *MyServer) coreToProtocolCodeAction(
    action core.CodeAction,
    content string,
) protocol.CodeAction {
    result := protocol.CodeAction{
        Title:       action.Title,
        IsPreferred: action.IsPreferred,
    }

    if action.Kind != nil {
        kind := protocol.CodeActionKind(*action.Kind)
        result.Kind = &kind
    }

    if action.Disabled != nil {
        result.Disabled = &protocol.CodeActionDisabled{
            Reason: action.Disabled.Reason,
        }
    }

    // Convert diagnostics
    if len(action.Diagnostics) > 0 {
        result.Diagnostics = adapter.CoreToProtocolDiagnostics(
            action.Diagnostics,
            content,
        )
    }

    // Convert workspace edit
    if action.Edit != nil {
        protocolEdit := s.coreToProtocolWorkspaceEdit(action.Edit, content)
        result.Edit = &protocolEdit
    }

    // Convert command if present
    if action.Command != nil {
        result.Command = &protocol.Command{
            Title:     action.Command.Title,
            Command:   action.Command.Command,
            Arguments: action.Command.Arguments,
        }
    }

    return result
}

func (s *MyServer) coreToProtocolWorkspaceEdit(
    edit *core.WorkspaceEdit,
    content string,
) protocol.WorkspaceEdit {
    result := protocol.WorkspaceEdit{}

    if len(edit.Changes) > 0 {
        result.Changes = make(map[protocol.DocumentURI][]protocol.TextEdit)

        for uri, edits := range edit.Changes {
            var protocolEdits []protocol.TextEdit
            for _, e := range edits {
                protocolEdits = append(protocolEdits, protocol.TextEdit{
                    Range:   adapter.CoreToProtocolRange(e.Range, content),
                    NewText: e.NewText,
                })
            }
            result.Changes[protocol.DocumentURI(uri)] = protocolEdits
        }
    }

    return result
}
```

### Server Capabilities

Advertise code action support in initialization:

```go
func (s *MyServer) Initialize(
    context *glsp.Context,
    params *protocol.InitializeParams,
) (interface{}, error) {
    capabilities := protocol.ServerCapabilities{
        CodeActionProvider: &protocol.CodeActionOptions{
            CodeActionKinds: []protocol.CodeActionKind{
                protocol.CodeActionKindQuickFix,
                protocol.CodeActionKindRefactor,
                protocol.CodeActionKindRefactorExtract,
                protocol.CodeActionKindSource,
                protocol.CodeActionKindSourceOrganizeImports,
            },
            ResolveProvider: false, // Set true if supporting resolve
        },
    }

    return protocol.InitializeResult{
        Capabilities: capabilities,
    }, nil
}
```

## Summary

You now know how to:

1. ✅ Implement code action providers with the `CodeFixProvider` interface
2. ✅ Create different types of actions (quick fixes, refactorings, source actions)
3. ✅ Work with workspace edits to modify code
4. ✅ Test code action providers thoroughly
5. ✅ Compose multiple providers with registries
6. ✅ Integrate providers into an LSP server with proper UTF-8/UTF-16 conversion

**Key Points:**
- Use appropriate `CodeActionKind` for each action type
- Quick fixes should reference the diagnostics they resolve
- Refactorings work on selected ranges
- Source actions are typically document-wide operations
- Always test edit application to ensure correctness
- Use registries to combine multiple providers

**Next Steps:**
- See [VALIDATORS.md](VALIDATORS.md) for diagnostic providers that feed into quick fixes
- Check `examples/` for complete working code
- Read [CORE_TYPES.md](CORE_TYPES.md) for UTF-8/UTF-16 conversion details
