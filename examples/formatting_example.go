package examples

import (
	"go/format"
	"strings"

	"github.com/SCKelemen/lsp/core"
)

// GoFormattingProvider formats Go source code using gofmt.
type GoFormattingProvider struct{}

func (p *GoFormattingProvider) ProvideFormatting(uri, content string, options core.FormattingOptions) []core.TextEdit {
	if !strings.HasSuffix(uri, ".go") {
		return nil
	}

	// Format using gofmt
	formatted, err := format.Source([]byte(content))
	if err != nil {
		return nil
	}

	// If content is already formatted, return no edits
	if string(formatted) == content {
		return nil
	}

	// Return a single edit replacing entire document
	lines := strings.Split(content, "\n")
	endLine := len(lines) - 1
	endChar := 0
	if endLine >= 0 && endLine < len(lines) {
		endChar = len(lines[endLine])
	}

	return []core.TextEdit{
		{
			Range: core.Range{
				Start: core.Position{Line: 0, Character: 0},
				End:   core.Position{Line: endLine, Character: endChar},
			},
			NewText: string(formatted),
		},
	}
}

// SimpleFormattingProvider provides basic formatting without using go/format.
type SimpleFormattingProvider struct {
	TabSize      int
	InsertSpaces bool
}

func NewSimpleFormattingProvider() *SimpleFormattingProvider {
	return &SimpleFormattingProvider{
		TabSize:      4,
		InsertSpaces: true,
	}
}

func (p *SimpleFormattingProvider) ProvideFormatting(uri, content string, options core.FormattingOptions) []core.TextEdit {
	// Use options if provided
	if options.TabSize > 0 {
		p.TabSize = options.TabSize
	}
	p.InsertSpaces = options.InsertSpaces

	var edits []core.TextEdit

	// Apply formatting rules
	edits = append(edits, p.fixTrailingWhitespace(content, options)...)
	edits = append(edits, p.fixFinalNewline(content, options)...)

	return edits
}

func (p *SimpleFormattingProvider) fixTrailingWhitespace(content string, options core.FormattingOptions) []core.TextEdit {
	if !options.TrimTrailingWhitespace {
		return nil
	}

	var edits []core.TextEdit
	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		trimmed := strings.TrimRight(line, " \t")

		if len(trimmed) != len(line) {
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

func (p *SimpleFormattingProvider) fixFinalNewline(content string, options core.FormattingOptions) []core.TextEdit {
	var edits []core.TextEdit

	lines := strings.Split(content, "\n")
	lastLine := lines[len(lines)-1]

	if options.InsertFinalNewline && lastLine != "" {
		// Add final newline if missing
		edits = append(edits, core.TextEdit{
			Range: core.Range{
				Start: core.Position{Line: len(lines) - 1, Character: len(lastLine)},
				End:   core.Position{Line: len(lines) - 1, Character: len(lastLine)},
			},
			NewText: "\n",
		})
	}

	if options.TrimFinalNewlines {
		// Count empty lines at end
		emptyLines := 0
		for i := len(lines) - 1; i >= 0; i-- {
			if strings.TrimSpace(lines[i]) == "" {
				emptyLines++
			} else {
				break
			}
		}

		// Keep at most one final newline (which means one empty line at the end after split)
		// For input "hello\n\n\n", lines=["hello","","",""], we want ["hello",""]
		// Remove from the END of the line we want to keep to the end of the last line
		if emptyLines > 1 {
			// Last line to keep (one empty line)
			lastLineToKeep := len(lines) - emptyLines
			lastLine := len(lines) - 1

			// Create edit that removes from end of lastLineToKeep to end of file
			edits = append(edits, core.TextEdit{
				Range: core.Range{
					Start: core.Position{Line: lastLineToKeep, Character: len(lines[lastLineToKeep])},
					End:   core.Position{Line: lastLine, Character: len(lines[lastLine])},
				},
				NewText: "",
			})
		}
	}

	return edits
}

// GoRangeFormattingProvider formats a selected range in Go code.
type GoRangeFormattingProvider struct{}

func (p *GoRangeFormattingProvider) ProvideRangeFormatting(
	uri, content string,
	r core.Range,
	options core.FormattingOptions,
) []core.TextEdit {
	if !strings.HasSuffix(uri, ".go") {
		return nil
	}

	lines := strings.Split(content, "\n")

	// Validate range
	if r.Start.Line < 0 || r.End.Line >= len(lines) {
		return nil
	}

	// Extract range content (complete lines)
	var rangeContent strings.Builder
	for i := r.Start.Line; i <= r.End.Line; i++ {
		rangeContent.WriteString(lines[i])
		if i < r.End.Line {
			rangeContent.WriteString("\n")
		}
	}

	snippet := rangeContent.String()

	// Try to format - wrap in function to make it valid
	wrappedSnippet := "package main\nfunc _() {\n" + snippet + "\n}"

	formatted, err := format.Source([]byte(wrappedSnippet))
	if err != nil {
		// If wrapping didn't work, try direct formatting
		formatted, err = format.Source([]byte(snippet))
		if err != nil {
			return nil
		}
	} else {
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

	// Create edit for the range
	return []core.TextEdit{
		{
			Range: core.Range{
				Start: core.Position{Line: r.Start.Line, Character: 0},
				End:   core.Position{Line: r.End.Line, Character: len(lines[r.End.Line])},
			},
			NewText: string(formatted),
		},
	}
}

// SimpleRangeFormattingProvider provides simple range formatting.
type SimpleRangeFormattingProvider struct {
	TabSize      int
	InsertSpaces bool
}

func NewSimpleRangeFormattingProvider() *SimpleRangeFormattingProvider {
	return &SimpleRangeFormattingProvider{
		TabSize:      4,
		InsertSpaces: true,
	}
}

func (p *SimpleRangeFormattingProvider) ProvideRangeFormatting(
	uri, content string,
	r core.Range,
	options core.FormattingOptions,
) []core.TextEdit {
	if options.TabSize > 0 {
		p.TabSize = options.TabSize
	}
	p.InsertSpaces = options.InsertSpaces

	lines := strings.Split(content, "\n")

	// Validate range
	if r.Start.Line < 0 || r.End.Line >= len(lines) {
		return nil
	}

	var edits []core.TextEdit

	// Format each line in range
	for lineNum := r.Start.Line; lineNum <= r.End.Line; lineNum++ {
		line := lines[lineNum]

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Trim trailing whitespace if requested
		newLine := line
		if options.TrimTrailingWhitespace {
			newLine = strings.TrimRight(newLine, " \t")
		}

		if newLine != line {
			edits = append(edits, core.TextEdit{
				Range: core.Range{
					Start: core.Position{Line: lineNum, Character: 0},
					End:   core.Position{Line: lineNum, Character: len(line)},
				},
				NewText: newLine,
			})
		}
	}

	return edits
}
