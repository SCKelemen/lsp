package examples

import (
	"go/format"
	"strings"

	"github.com/SCKelemen/lsp/core"
)

// GoRangesFormattingProvider formats multiple ranges in Go source code.
// This is useful for formatting discontinuous selections simultaneously.
// LSP 3.18 feature.
type GoRangesFormattingProvider struct{}

func (p *GoRangesFormattingProvider) ProvideRangesFormatting(uri, content string, ranges []core.Range, options core.FormattingOptions) []core.TextEdit {
	if !strings.HasSuffix(uri, ".go") {
		return nil
	}

	if len(ranges) == 0 {
		return nil
	}

	var allEdits []core.TextEdit

	// Format each range independently
	for _, rng := range ranges {
		edits := p.formatRange(uri, content, rng, options)
		allEdits = append(allEdits, edits...)
	}

	return allEdits
}

func (p *GoRangesFormattingProvider) formatRange(uri, content string, rng core.Range, options core.FormattingOptions) []core.TextEdit {
	lines := strings.Split(content, "\n")

	// Validate range
	if rng.Start.Line < 0 || rng.End.Line >= len(lines) {
		return nil
	}

	// Extract the range content (complete lines)
	var rangeContent strings.Builder
	for i := rng.Start.Line; i <= rng.End.Line; i++ {
		rangeContent.WriteString(lines[i])
		if i < rng.End.Line {
			rangeContent.WriteString("\n")
		}
	}

	snippet := rangeContent.String()

	// Try to format as a complete Go file first
	formatted, err := format.Source([]byte(snippet))
	if err != nil {
		// If that fails, try wrapping in a function
		wrappedSnippet := "package main\nfunc _() {\n" + snippet + "\n}"
		formatted, err = format.Source([]byte(wrappedSnippet))
		if err != nil {
			// If still fails, return no edits
			return nil
		}

		// Extract the formatted snippet from the wrapper
		formattedStr := string(formatted)
		start := strings.Index(formattedStr, "func _() {\n") + len("func _() {\n")
		end := strings.LastIndex(formattedStr, "\n}")
		if start > 0 && end > start {
			formatted = []byte(formattedStr[start:end])
		}
	}

	// If content unchanged, return no edits
	if string(formatted) == snippet {
		return nil
	}

	// Create edit for this range
	return []core.TextEdit{
		{
			Range: core.Range{
				Start: core.Position{Line: rng.Start.Line, Character: 0},
				End:   core.Position{Line: rng.End.Line, Character: len(lines[rng.End.Line])},
			},
			NewText: string(formatted),
		},
	}
}

// SimpleRangesFormattingProvider provides basic formatting for multiple ranges.
type SimpleRangesFormattingProvider struct {
	TabSize      int
	InsertSpaces bool
}

func NewSimpleRangesFormattingProvider() *SimpleRangesFormattingProvider {
	return &SimpleRangesFormattingProvider{
		TabSize:      4,
		InsertSpaces: true,
	}
}

func (p *SimpleRangesFormattingProvider) ProvideRangesFormatting(uri, content string, ranges []core.Range, options core.FormattingOptions) []core.TextEdit {
	if len(ranges) == 0 {
		return nil
	}

	// Use options if provided
	if options.TabSize > 0 {
		p.TabSize = options.TabSize
	}
	p.InsertSpaces = options.InsertSpaces

	var allEdits []core.TextEdit

	// Process each range
	for _, rng := range ranges {
		edits := p.formatRange(content, rng, options)
		allEdits = append(allEdits, edits...)
	}

	return allEdits
}

func (p *SimpleRangesFormattingProvider) formatRange(content string, rng core.Range, options core.FormattingOptions) []core.TextEdit {
	var edits []core.TextEdit

	lines := strings.Split(content, "\n")

	// Validate range
	if rng.Start.Line < 0 || rng.End.Line >= len(lines) {
		return nil
	}

	// Format each line in the range
	for lineNum := rng.Start.Line; lineNum <= rng.End.Line; lineNum++ {
		line := lines[lineNum]

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		trimmed := strings.TrimRight(line, " \t")

		// Apply formatting
		if options.TrimTrailingWhitespace && len(trimmed) != len(line) {
			edits = append(edits, core.TextEdit{
				Range: core.Range{
					Start: core.Position{Line: lineNum, Character: len(trimmed)},
					End:   core.Position{Line: lineNum, Character: len(line)},
				},
				NewText: "",
			})
		}
	}

	return edits
}

// Example: Format multiple non-contiguous ranges
func CLIRangesFormattingExample() {
	content := `package main

func main() {
    x := 1
    y := 2

    // Some code in between
    z := 3

    a := 4
    b := 5
}
`

	provider := NewSimpleRangesFormattingProvider()

	// Format two separate ranges: lines 3-4 and lines 9-10
	ranges := []core.Range{
		{
			Start: core.Position{Line: 3, Character: 0},
			End:   core.Position{Line: 4, Character: 50},
		},
		{
			Start: core.Position{Line: 9, Character: 0},
			End:   core.Position{Line: 10, Character: 50},
		},
	}

	options := core.FormattingOptions{
		TabSize:                4,
		InsertSpaces:           true,
		TrimTrailingWhitespace: true,
	}

	edits := provider.ProvideRangesFormatting("file:///main.go", content, ranges, options)

	println("Formatting edits for multiple ranges:", len(edits))
	for i, edit := range edits {
		println("Edit", i, "at line", edit.Range.Start.Line)
	}
}

// Example usage in LSP server
// func (s *Server) TextDocumentRangesFormatting(
// 	ctx *lsp.Context,
// 	params *protocol.DocumentRangesFormattingParams,
// ) ([]protocol.TextEdit, error) {
// 	uri := string(params.TextDocument.URI)
// 	content := s.documents.GetContent(uri)
//
// 	if content == "" {
// 		return nil, nil
// 	}
//
// 	// Convert protocol ranges to core ranges
// 	var coreRanges []core.Range
// 	for _, protocolRange := range params.Ranges {
// 		coreRanges = append(coreRanges, adapter.ProtocolToCoreRange(protocolRange, content))
// 	}
//
// 	// Convert formatting options
// 	coreOptions := core.FormattingOptions{
// 		TabSize:                int(params.Options.TabSize),
// 		InsertSpaces:           params.Options.InsertSpaces,
// 		TrimTrailingWhitespace: params.Options.TrimTrailingWhitespace,
// 		InsertFinalNewline:     params.Options.InsertFinalNewline,
// 		TrimFinalNewlines:      params.Options.TrimFinalNewlines,
// 	}
//
// 	// Get formatting edits for all ranges
// 	coreEdits := s.rangesFormatting.ProvideRangesFormatting(uri, content, coreRanges, coreOptions)
//
// 	// Convert to protocol edits
// 	var protocolEdits []protocol.TextEdit
// 	for _, edit := range coreEdits {
// 		protocolEdits = append(protocolEdits, protocol.TextEdit{
// 			Range:   adapter.CoreToProtocolRange(edit.Range, content),
// 			NewText: edit.NewText,
// 		})
// 	}
//
// 	return protocolEdits, nil
// }
