package examples

import (
	"strings"

	"github.com/SCKelemen/lsp/core"
)

// SimpleInlineCompletionProvider provides basic inline completions.
// In a real implementation, this would typically call an AI model or
// use sophisticated code analysis. This example shows simple pattern-based
// suggestions for demonstration purposes.
type SimpleInlineCompletionProvider struct{}

func (p *SimpleInlineCompletionProvider) ProvideInlineCompletions(ctx core.InlineCompletionContext) *core.InlineCompletionList {
	if !strings.HasSuffix(ctx.URI, ".go") {
		return nil
	}

	// Get the current line up to the cursor
	lines := strings.Split(ctx.Content, "\n")
	if ctx.Position.Line >= len(lines) {
		return nil
	}

	currentLine := lines[ctx.Position.Line]
	if ctx.Position.Character > len(currentLine) {
		return nil
	}

	prefix := currentLine[:ctx.Position.Character]

	// Try different completion strategies
	var items []core.InlineCompletionItem

	// Strategy 1: Common Go patterns
	if item := p.tryCommonPatterns(prefix, ctx); item != nil {
		items = append(items, *item)
	}

	// Strategy 2: Error handling
	if item := p.tryErrorHandling(prefix, ctx); item != nil {
		items = append(items, *item)
	}

	// Strategy 3: Loop patterns
	if item := p.tryLoopPatterns(prefix, ctx); item != nil {
		items = append(items, *item)
	}

	if len(items) == 0 {
		return nil
	}

	return &core.InlineCompletionList{
		Items: items,
	}
}

// tryCommonPatterns suggests completions for common Go patterns
func (p *SimpleInlineCompletionProvider) tryCommonPatterns(prefix string, ctx core.InlineCompletionContext) *core.InlineCompletionItem {
	// fmt.Println completion
	if strings.HasSuffix(prefix, "fmt.P") {
		return &core.InlineCompletionItem{
			InsertText: `rintln("TODO")`,
			FilterText: "Println",
		}
	}

	// Function definition
	if strings.HasSuffix(prefix, "func ") {
		return &core.InlineCompletionItem{
			InsertText: `main() {
	// TODO
}`,
			FilterText: "main",
		}
	}

	// Struct definition
	if strings.HasSuffix(prefix, "type ") {
		return &core.InlineCompletionItem{
			InsertText: `MyStruct struct {
	// TODO: add fields
}`,
			FilterText: "struct",
		}
	}

	return nil
}

// tryErrorHandling suggests error handling patterns
func (p *SimpleInlineCompletionProvider) tryErrorHandling(prefix string, ctx core.InlineCompletionContext) *core.InlineCompletionItem {
	// After variable assignment, suggest error check
	if strings.Contains(prefix, ":= ") && strings.HasSuffix(strings.TrimSpace(prefix), "err") {
		// Suggest error handling on the next line
		return &core.InlineCompletionItem{
			InsertText: `
	if err != nil {
		return err
	}`,
			FilterText: "if err",
		}
	}

	// If writing "if err"
	if strings.HasSuffix(prefix, "if err") {
		return &core.InlineCompletionItem{
			InsertText: ` != nil {
		return err
	}`,
			FilterText: "if err != nil",
		}
	}

	return nil
}

// tryLoopPatterns suggests loop patterns
func (p *SimpleInlineCompletionProvider) tryLoopPatterns(prefix string, ctx core.InlineCompletionContext) *core.InlineCompletionItem {
	// Range loop
	if strings.HasSuffix(prefix, "for ") {
		return &core.InlineCompletionItem{
			InsertText: `i, v := range items {
		// TODO
	}`,
			FilterText: "range",
		}
	}

	// Classic for loop
	if strings.HasSuffix(prefix, "for i") {
		return &core.InlineCompletionItem{
			InsertText: ` := 0; i < len(items); i++ {
		// TODO
	}`,
			FilterText: "for i :=",
		}
	}

	return nil
}

// AdvancedInlineCompletionProvider demonstrates more sophisticated completion logic
// This would typically integrate with an AI model or advanced static analysis
type AdvancedInlineCompletionProvider struct {
	// In production, this might include:
	// - AI model client
	// - Context cache
	// - User preferences
	// - Project-specific patterns
}

func (p *AdvancedInlineCompletionProvider) ProvideInlineCompletions(ctx core.InlineCompletionContext) *core.InlineCompletionList {
	// Check if this was triggered automatically or explicitly
	if ctx.TriggerKind == core.InlineCompletionTriggerKindAutomatic {
		// For automatic triggers, be more conservative
		return p.provideConservativeCompletions(ctx)
	}

	// For explicit triggers (e.g., Ctrl+Space), be more aggressive
	return p.provideAggressiveCompletions(ctx)
}

func (p *AdvancedInlineCompletionProvider) provideConservativeCompletions(ctx core.InlineCompletionContext) *core.InlineCompletionList {
	// Only suggest for clear patterns
	lines := strings.Split(ctx.Content, "\n")
	if ctx.Position.Line >= len(lines) {
		return nil
	}

	currentLine := lines[ctx.Position.Line]
	if ctx.Position.Character > len(currentLine) {
		return nil
	}

	prefix := strings.TrimSpace(currentLine[:ctx.Position.Character])

	// Only suggest if there's a clear intent
	if len(prefix) < 3 {
		return nil
	}

	// In production, this would call an AI model
	// For now, return nil to indicate no suggestions
	return nil
}

func (p *AdvancedInlineCompletionProvider) provideAggressiveCompletions(ctx core.InlineCompletionContext) *core.InlineCompletionList {
	// Could provide multiple suggestions ranked by confidence
	// For now, delegate to simple provider logic
	simple := &SimpleInlineCompletionProvider{}
	return simple.ProvideInlineCompletions(ctx)
}

// ContextAwareInlineCompletionProvider uses context from selected completions
type ContextAwareInlineCompletionProvider struct{}

func (p *ContextAwareInlineCompletionProvider) ProvideInlineCompletions(ctx core.InlineCompletionContext) *core.InlineCompletionList {
	// If there's a selected completion in the autocomplete widget,
	// we can use that context to provide better inline suggestions
	if ctx.SelectedCompletionInfo != nil {
		return p.provideWithCompletionContext(ctx)
	}

	// Otherwise, provide standalone suggestions
	simple := &SimpleInlineCompletionProvider{}
	return simple.ProvideInlineCompletions(ctx)
}

func (p *ContextAwareInlineCompletionProvider) provideWithCompletionContext(ctx core.InlineCompletionContext) *core.InlineCompletionList {
	// The selected completion gives us information about what the user
	// is likely trying to type. We can use this to provide better suggestions.

	selectedText := ctx.SelectedCompletionInfo.Text

	// For example, if they selected a function name, we might suggest
	// the full function call with common parameters
	if strings.Contains(selectedText, "Print") {
		return &core.InlineCompletionList{
			Items: []core.InlineCompletionItem{
				{
					InsertText: `ln("Hello, World!")`,
					FilterText: "Println",
				},
			},
		}
	}

	return nil
}
