package examples

import (
	"strings"
	"unicode"

	"github.com/SCKelemen/lsp/core"
)

// SimpleHighlightProvider highlights all occurrences of a word in a document.
// This is a basic example that highlights based on exact word matching.
type SimpleHighlightProvider struct{}

func (p *SimpleHighlightProvider) ProvideDocumentHighlights(ctx core.DocumentHighlightContext) []core.DocumentHighlight {
	var highlights []core.DocumentHighlight

	// Get the word at the cursor position
	word := getWordAtPosition(ctx.Content, ctx.Position)
	if word == "" {
		return nil
	}

	// Find all occurrences of this word
	content := ctx.Content
	wordLen := len(word)

	for i := 0; i <= len(content)-wordLen; i++ {
		// Check if we have a word match
		if content[i:i+wordLen] == word {
			// Check word boundaries (don't match partial words)
			beforeOk := i == 0 || !isWordChar(rune(content[i-1]))
			afterOk := i+wordLen >= len(content) || !isWordChar(rune(content[i+wordLen]))

			if beforeOk && afterOk {
				startPos := core.ByteOffsetToPosition(content, i)
				endPos := core.ByteOffsetToPosition(content, i+wordLen)

				// Default to Text highlighting
				kind := core.DocumentHighlightKindText

				highlights = append(highlights, core.DocumentHighlight{
					Range: core.Range{
						Start: startPos,
						End:   endPos,
					},
					Kind: &kind,
				})
			}
		}
	}

	return highlights
}

// VariableHighlightProvider highlights variables with Read/Write distinction.
// This is a more advanced example that distinguishes between reads and writes.
type VariableHighlightProvider struct{}

func (p *VariableHighlightProvider) ProvideDocumentHighlights(ctx core.DocumentHighlightContext) []core.DocumentHighlight {
	var highlights []core.DocumentHighlight

	// Get the word at the cursor position
	word := getWordAtPosition(ctx.Content, ctx.Position)
	if word == "" {
		return nil
	}

	content := ctx.Content
	wordLen := len(word)

	for i := 0; i <= len(content)-wordLen; i++ {
		if content[i:i+wordLen] == word {
			// Check word boundaries
			beforeOk := i == 0 || !isWordChar(rune(content[i-1]))
			afterOk := i+wordLen >= len(content) || !isWordChar(rune(content[i+wordLen]))

			if beforeOk && afterOk {
				startPos := core.ByteOffsetToPosition(content, i)
				endPos := core.ByteOffsetToPosition(content, i+wordLen)

				// Determine if this is a read or write based on context
				kind := determineHighlightKind(content, i, wordLen)

				highlights = append(highlights, core.DocumentHighlight{
					Range: core.Range{
						Start: startPos,
						End:   endPos,
					},
					Kind: &kind,
				})
			}
		}
	}

	return highlights
}

// Helper: Get the word at a position
func getWordAtPosition(content string, pos core.Position) string {
	offset := core.PositionToByteOffset(content, pos)
	if offset >= len(content) {
		return ""
	}

	// Find word boundaries
	start := offset
	for start > 0 && isWordChar(rune(content[start-1])) {
		start--
	}

	end := offset
	for end < len(content) && isWordChar(rune(content[end])) {
		end++
	}

	if start >= end {
		return ""
	}

	return content[start:end]
}

// Helper: Check if character is part of a word
func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

// Helper: Determine if this is a read or write based on surrounding context
func determineHighlightKind(content string, offset, wordLen int) core.DocumentHighlightKind {
	// Look ahead to see if there's an assignment operator
	afterWord := offset + wordLen

	// Skip whitespace
	for afterWord < len(content) && unicode.IsSpace(rune(content[afterWord])) {
		afterWord++
	}

	// Check for assignment operators
	if afterWord < len(content) {
		switch {
		case content[afterWord] == '=':
			// Check it's not == or !=
			if afterWord+1 < len(content) && content[afterWord+1] != '=' {
				return core.DocumentHighlightKindWrite
			}
		case afterWord+1 < len(content) &&
			(strings.HasPrefix(content[afterWord:], "++") ||
				strings.HasPrefix(content[afterWord:], "--") ||
				strings.HasPrefix(content[afterWord:], "+=") ||
				strings.HasPrefix(content[afterWord:], "-=") ||
				strings.HasPrefix(content[afterWord:], "*=") ||
				strings.HasPrefix(content[afterWord:], "/=")):
			return core.DocumentHighlightKindWrite
		}
	}

	// Default to read
	return core.DocumentHighlightKindRead
}

// Example usage in CLI tool
func CLIHighlightExample() {
	content := `func example() {
    count := 0
    count++
    count = count + 5
    println(count)
}`

	// User clicks on "count" at line 1, character 4
	ctx := core.DocumentHighlightContext{
		URI:      "file:///example.go",
		Content:  content,
		Position: core.Position{Line: 1, Character: 4},
	}

	provider := &VariableHighlightProvider{}
	highlights := provider.ProvideDocumentHighlights(ctx)

	println("Found", len(highlights), "highlights:")
	for _, h := range highlights {
		kindStr := "Text"
		if h.Kind != nil {
			switch *h.Kind {
			case core.DocumentHighlightKindRead:
				kindStr = "Read"
			case core.DocumentHighlightKindWrite:
				kindStr = "Write"
			}
		}
		println("  -", h.Range.String(), kindStr)
	}
}

// Example usage in LSP server
// func (s *Server) TextDocumentDocumentHighlight(
// 	ctx *glsp.Context,
// 	params *protocol.DocumentHighlightParams,
// ) ([]protocol.DocumentHighlight, error) {
// 	uri := string(params.TextDocument.URI)
// 	content := s.documents.GetContent(uri)
//
// 	// Convert protocol position to core position
// 	corePos := adapter_3_16.ProtocolToCorePosition(params.Position, content)
//
// 	// Use provider with core types
// 	highlightCtx := core.DocumentHighlightContext{
// 		URI:      uri,
// 		Content:  content,
// 		Position: corePos,
// 	}
// 	coreHighlights := s.highlightProvider.ProvideDocumentHighlights(highlightCtx)
//
// 	// Convert back to protocol
// 	return adapter_3_16.CoreToProtocolDocumentHighlights(coreHighlights, content), nil
// }
