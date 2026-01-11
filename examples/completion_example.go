package examples

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/SCKelemen/lsp/core"
	"github.com/SCKelemen/unicode/uax29"
)

// KeywordCompletionProvider provides keyword completions for a language.
// This is useful for simple languages or configuration files.
type KeywordCompletionProvider struct {
	// Keywords are the language keywords to complete
	Keywords []string

	// TriggerCharacters are characters that trigger completion
	TriggerCharacters []string
}

func NewGoKeywordCompletionProvider() *KeywordCompletionProvider {
	return &KeywordCompletionProvider{
		Keywords: []string{
			"break", "case", "chan", "const", "continue",
			"default", "defer", "else", "fallthrough", "for",
			"func", "go", "goto", "if", "import",
			"interface", "map", "package", "range", "return",
			"select", "struct", "switch", "type", "var",
		},
		TriggerCharacters: []string{},
	}
}

func (p *KeywordCompletionProvider) ProvideCompletions(ctx core.CompletionContext) *core.CompletionList {
	// Get the word being typed using Unicode word boundaries
	offset := core.PositionToByteOffset(ctx.Content, ctx.Position)
	if offset < 0 {
		return nil
	}

	// Get all word boundaries
	breaks := uax29.FindWordBreaks(ctx.Content)

	// Find which word we're in
	prefix := ""
	if len(breaks) > 0 {
		for i := 0; i < len(breaks)-1; i++ {
			start := breaks[i]
			end := breaks[i+1]

			// Check if cursor is within or at the end of this word
			if offset >= start && offset <= end {
				prefix = ctx.Content[start:offset]
				break
			}
		}
	}

	// Treat whitespace-only prefix as empty (show all completions)
	prefix = strings.TrimSpace(prefix)

	// Filter keywords by prefix
	var items []core.CompletionItem
	for _, keyword := range p.Keywords {
		if prefix == "" || strings.HasPrefix(keyword, prefix) {
			kind := core.CompletionItemKindKeyword
			items = append(items, core.CompletionItem{
				Label:  keyword,
				Kind:   &kind,
				Detail: "keyword",
				InsertText: keyword,
			})
		}
	}

	if len(items) == 0 {
		return nil
	}

	return &core.CompletionList{
		IsIncomplete: false,
		Items:        items,
	}
}

// SnippetCompletionProvider provides snippet completions.
// Snippets are templates with placeholders that can be filled in.
type SnippetCompletionProvider struct {
	Snippets []Snippet
}

type Snippet struct {
	Prefix      string
	Label       string
	Description string
	Body        string
}

func NewGoSnippetProvider() *SnippetCompletionProvider {
	return &SnippetCompletionProvider{
		Snippets: []Snippet{
			{
				Prefix:      "for",
				Label:       "for loop",
				Description: "For loop with index",
				Body:        "for ${1:i} := ${2:0}; $1 < ${3:count}; $1++ {\n\t${4:// body}\n}",
			},
			{
				Prefix:      "forr",
				Label:       "for range",
				Description: "For range loop",
				Body:        "for ${1:i}, ${2:v} := range ${3:collection} {\n\t${4:// body}\n}",
			},
			{
				Prefix:      "func",
				Label:       "function",
				Description: "Function declaration",
				Body:        "func ${1:name}(${2:params}) ${3:returnType} {\n\t${4:// body}\n}",
			},
			{
				Prefix:      "if",
				Label:       "if statement",
				Description: "If statement",
				Body:        "if ${1:condition} {\n\t${2:// body}\n}",
			},
			{
				Prefix:      "ife",
				Label:       "if else",
				Description: "If-else statement",
				Body:        "if ${1:condition} {\n\t${2:// true}\n} else {\n\t${3:// false}\n}",
			},
		},
	}
}

func (p *SnippetCompletionProvider) ProvideCompletions(ctx core.CompletionContext) *core.CompletionList {
	// Get the word being typed using Unicode word boundaries
	offset := core.PositionToByteOffset(ctx.Content, ctx.Position)
	if offset < 0 {
		return nil
	}

	// Get all word boundaries
	breaks := uax29.FindWordBreaks(ctx.Content)

	// Find which word we're in
	prefix := ""
	if len(breaks) > 0 {
		for i := 0; i < len(breaks)-1; i++ {
			start := breaks[i]
			end := breaks[i+1]

			// Check if cursor is within or at the end of this word
			if offset >= start && offset <= end {
				prefix = ctx.Content[start:offset]
				break
			}
		}
	}

	// Treat whitespace-only prefix as empty (show all completions)
	prefix = strings.TrimSpace(prefix)

	// Filter snippets by prefix
	var items []core.CompletionItem
	for _, snippet := range p.Snippets {
		if prefix == "" || strings.HasPrefix(snippet.Prefix, prefix) {
			kind := core.CompletionItemKindSnippet
			format := core.InsertTextFormatSnippet

			items = append(items, core.CompletionItem{
				Label:            snippet.Label,
				Kind:             &kind,
				Detail:           snippet.Description,
				InsertText:       snippet.Body,
				InsertTextFormat: &format,
				Documentation:    fmt.Sprintf("Snippet: %s\n\n%s", snippet.Prefix, snippet.Description),
			})
		}
	}

	if len(items) == 0 {
		return nil
	}

	return &core.CompletionList{
		IsIncomplete: false,
		Items:        items,
	}
}

// SymbolCompletionProvider provides completions based on symbols in scope.
// This uses AST parsing to find available identifiers.
type SymbolCompletionProvider struct{}

func (p *SymbolCompletionProvider) ProvideCompletions(ctx core.CompletionContext) *core.CompletionList {
	if !strings.HasSuffix(ctx.URI, ".go") {
		return nil
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", ctx.Content, parser.ParseComments)
	if err != nil {
		// If parsing fails, try partial completion
		return nil
	}

	// Get the word being typed using Unicode word boundaries
	offset := core.PositionToByteOffset(ctx.Content, ctx.Position)
	if offset < 0 {
		return nil
	}

	// Get all word boundaries
	breaks := uax29.FindWordBreaks(ctx.Content)

	// Find which word we're in
	prefix := ""
	if len(breaks) > 0 {
		for i := 0; i < len(breaks)-1; i++ {
			start := breaks[i]
			end := breaks[i+1]

			// Check if cursor is within or at the end of this word
			if offset >= start && offset <= end {
				prefix = strings.ToLower(ctx.Content[start:offset])
				break
			}
		}
	}

	// Treat whitespace-only prefix as empty (show all completions)
	prefix = strings.TrimSpace(prefix)

	// Collect all identifiers in scope
	symbols := make(map[string]core.CompletionItemKind)

	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Name.Name != "_" {
				symbols[node.Name.Name] = core.CompletionItemKindFunction
			}
		case *ast.TypeSpec:
			if node.Name.Name != "_" {
				switch node.Type.(type) {
				case *ast.StructType:
					symbols[node.Name.Name] = core.CompletionItemKindStruct
				case *ast.InterfaceType:
					symbols[node.Name.Name] = core.CompletionItemKindInterface
				default:
					symbols[node.Name.Name] = core.CompletionItemKindClass
				}
			}
		case *ast.ValueSpec:
			for _, name := range node.Names {
				if name.Name != "_" {
					symbols[name.Name] = core.CompletionItemKindVariable
				}
			}
		}
		return true
	})

	// Filter by prefix
	var items []core.CompletionItem
	for name, kind := range symbols {
		if prefix == "" || strings.HasPrefix(strings.ToLower(name), prefix) {
			kindCopy := kind
			items = append(items, core.CompletionItem{
				Label: name,
				Kind:  &kindCopy,
			})
		}
	}

	if len(items) == 0 {
		return nil
	}

	return &core.CompletionList{
		IsIncomplete: false,
		Items:        items,
	}
}

// ImportCompletionProvider provides completions for import statements.
type ImportCompletionProvider struct {
	// AvailablePackages is a list of available packages
	AvailablePackages []string
}

func NewGoImportCompletionProvider() *ImportCompletionProvider {
	return &ImportCompletionProvider{
		AvailablePackages: []string{
			"fmt", "strings", "strconv", "io", "os",
			"path/filepath", "net/http", "encoding/json",
			"context", "sync", "time", "errors",
		},
	}
}

func (p *ImportCompletionProvider) ProvideCompletions(ctx core.CompletionContext) *core.CompletionList {
	// Check if we're inside an import statement
	offset := core.PositionToByteOffset(ctx.Content, ctx.Position)
	if offset < 0 {
		return nil
	}

	// Simple heuristic: check if "import" appears before the cursor
	beforeCursor := ctx.Content[:offset]
	if !strings.Contains(beforeCursor, "import") {
		return nil
	}

	// Find the current word
	start := offset
	for start > 0 && ctx.Content[start-1] != '"' && ctx.Content[start-1] != '\n' {
		start--
	}

	prefix := ""
	if start < offset {
		prefix = ctx.Content[start:offset]
	}

	// Filter packages by prefix
	var items []core.CompletionItem
	for _, pkg := range p.AvailablePackages {
		if prefix == "" || strings.HasPrefix(pkg, prefix) || strings.Contains(pkg, prefix) {
			kind := core.CompletionItemKindModule
			items = append(items, core.CompletionItem{
				Label:      pkg,
				Kind:       &kind,
				Detail:     "package",
				InsertText: pkg,
			})
		}
	}

	if len(items) == 0 {
		return nil
	}

	return &core.CompletionList{
		IsIncomplete: true, // More packages may be available
		Items:        items,
	}
}

// LazyCompletionProvider demonstrates lazy resolution of completion items.
// The initial items have minimal information, and details are resolved on demand.
type LazyCompletionProvider struct {
	BaseProvider core.CompletionProvider
	ResolveFunc  func(item core.CompletionItem) core.CompletionItem
}

func (p *LazyCompletionProvider) ProvideCompletions(ctx core.CompletionContext) *core.CompletionList {
	// Get completions from base provider
	list := p.BaseProvider.ProvideCompletions(ctx)
	if list == nil {
		return nil
	}

	// Strip documentation and details (will be resolved later)
	for i := range list.Items {
		list.Items[i].Documentation = ""
		list.Items[i].Detail = ""
		list.Items[i].Data = map[string]interface{}{
			"label": list.Items[i].Label,
			"kind":  list.Items[i].Kind,
		}
	}

	return list
}

func (p *LazyCompletionProvider) ResolveCompletionItem(item core.CompletionItem) core.CompletionItem {
	if p.ResolveFunc != nil {
		return p.ResolveFunc(item)
	}

	// Default resolution: add generic documentation
	if data, ok := item.Data.(map[string]interface{}); ok {
		if label, ok := data["label"].(string); ok {
			item.Documentation = fmt.Sprintf("Documentation for %s", label)
			item.Detail = fmt.Sprintf("Details about %s", label)
		}
	}

	return item
}

// CompositeCompletionProvider combines multiple completion providers.
type CompositeCompletionProvider struct {
	Providers []core.CompletionProvider
}

func NewCompositeCompletionProvider(providers ...core.CompletionProvider) *CompositeCompletionProvider {
	return &CompositeCompletionProvider{
		Providers: providers,
	}
}

func (p *CompositeCompletionProvider) ProvideCompletions(ctx core.CompletionContext) *core.CompletionList {
	var allItems []core.CompletionItem
	isIncomplete := false

	for _, provider := range p.Providers {
		list := provider.ProvideCompletions(ctx)
		if list != nil {
			allItems = append(allItems, list.Items...)
			if list.IsIncomplete {
				isIncomplete = true
			}
		}
	}

	if len(allItems) == 0 {
		return nil
	}

	return &core.CompletionList{
		IsIncomplete: isIncomplete,
		Items:        allItems,
	}
}

// Example usage in CLI tool
func CLICompletionExample() {
	content := `package main

import "fmt"

func main() {
	fo
}
`

	provider := NewCompositeCompletionProvider(
		NewGoKeywordCompletionProvider(),
		NewGoSnippetProvider(),
		&SymbolCompletionProvider{},
	)

	ctx := core.CompletionContext{
		URI:              "file:///main.go",
		Content:          content,
		Position:         core.Position{Line: 5, Character: 4}, // After "fo"
		TriggerKind:      core.CompletionTriggerKindInvoked,
		TriggerCharacter: "",
	}

	list := provider.ProvideCompletions(ctx)

	if list == nil {
		println("No completions available")
		return
	}

	println(fmt.Sprintf("Found %d completions (incomplete: %v):", len(list.Items), list.IsIncomplete))
	for i, item := range list.Items {
		kindStr := "unknown"
		if item.Kind != nil {
			kindStr = fmt.Sprintf("%d", *item.Kind)
		}
		println(fmt.Sprintf("  %d. %s (%s) - %s", i+1, item.Label, kindStr, item.Detail))
	}
}

// Example usage in LSP server
// func (s *Server) TextDocumentCompletion(
// 	ctx *lsp.Context,
// 	params *protocol.CompletionParams,
// ) (*protocol.CompletionList, error) {
// 	uri := string(params.TextDocument.URI)
// 	content := s.documents.GetContent(uri)
//
// 	// Convert protocol position to core position
// 	corePos := adapter_3_16.ProtocolToCorePosition(params.Position, content)
//
// 	// Use provider with core types
// 	coreCtx := core.CompletionContext{
// 		URI:              uri,
// 		Content:          content,
// 		Position:         corePos,
// 		TriggerKind:      core.CompletionTriggerKind(params.Context.TriggerKind),
// 		TriggerCharacter: params.Context.TriggerCharacter,
// 	}
//
// 	coreList := s.completionProvider.ProvideCompletions(coreCtx)
// 	if coreList == nil {
// 		return nil, nil
// 	}
//
// 	// Convert back to protocol
// 	return adapter_3_16.CoreToProtocolCompletionList(*coreList, content), nil
// }
//
// func (s *Server) CompletionItemResolve(
// 	ctx *lsp.Context,
// 	params *protocol.CompletionItem,
// ) (*protocol.CompletionItem, error) {
// 	// Convert protocol to core
// 	coreItem := adapter_3_16.ProtocolToCoreCompletionItem(*params)
//
// 	// Resolve with provider
// 	resolved := s.completionResolver.ResolveCompletionItem(coreItem)
//
// 	// Convert back to protocol
// 	protocolItem := adapter_3_16.CoreToProtocolCompletionItem(resolved)
// 	return &protocolItem, nil
// }
