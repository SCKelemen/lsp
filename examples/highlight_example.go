package examples

import (
	"strings"
	"unicode"

	"github.com/SCKelemen/lsp/core"
	"github.com/SCKelemen/unicode/uax29"
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

	// Get all word boundaries in the content
	breaks := uax29.FindWordBreaks(ctx.Content)
	if len(breaks) < 2 {
		return nil
	}

	// Find all occurrences of this word using UAX29 word boundaries
	for i := 0; i < len(breaks)-1; i++ {
		start := breaks[i]
		end := breaks[i+1]

		// Extract the word at this boundary
		candidateWord := ctx.Content[start:end]

		// Check if this word matches our target word
		if candidateWord == word {
			startPos := core.ByteOffsetToPosition(ctx.Content, start)
			endPos := core.ByteOffsetToPosition(ctx.Content, end)

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

	// Get all word boundaries in the content
	breaks := uax29.FindWordBreaks(ctx.Content)
	if len(breaks) < 2 {
		return nil
	}

	// Find all occurrences of this word using UAX29 word boundaries
	for i := 0; i < len(breaks)-1; i++ {
		start := breaks[i]
		end := breaks[i+1]

		// Extract the word at this boundary
		candidateWord := ctx.Content[start:end]

		// Check if this word matches our target word
		if candidateWord == word {
			startPos := core.ByteOffsetToPosition(ctx.Content, start)
			endPos := core.ByteOffsetToPosition(ctx.Content, end)

			// Determine if this is a read or write based on context
			kind := determineHighlightKind(ctx.Content, start, end-start)

			highlights = append(highlights, core.DocumentHighlight{
				Range: core.Range{
					Start: startPos,
					End:   endPos,
				},
				Kind: &kind,
			})
		}
	}

	return highlights
}

// Helper: Get the word at a position using Unicode word boundaries
func getWordAtPosition(content string, pos core.Position) string {
	offset := core.PositionToByteOffset(content, pos)
	if offset < 0 || offset >= len(content) {
		return ""
	}

	// Get all word boundaries in the content
	breaks := uax29.FindWordBreaks(content)
	if len(breaks) < 2 {
		return ""
	}

	// Helper function to check if a segment is a valid word
	isValidWord := func(word string) bool {
		if len(strings.TrimSpace(word)) == 0 {
			return false // Just whitespace
		}

		// Check if it contains at least one alphanumeric character
		for _, r := range word {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r > 127 {
				return true
			}
		}
		return false // Just punctuation
	}

	// Find which word boundary the offset falls into
	for i := 0; i < len(breaks)-1; i++ {
		start := breaks[i]
		end := breaks[i+1]

		if offset >= start && offset < end {
			word := content[start:end]
			if isValidWord(word) {
				return word
			}
			// If cursor is on whitespace/punctuation, try the previous word
			if i > 0 {
				prevStart := breaks[i-1]
				prevEnd := breaks[i]
				prevWord := content[prevStart:prevEnd]
				if isValidWord(prevWord) {
					return prevWord
				}
			}
			return ""
		}
	}

	return ""
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
